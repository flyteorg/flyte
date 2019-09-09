package qubole

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"strconv"
	"time"

	eventErrors "github.com/lyft/flyteidl/clients/go/events/errors"
	"github.com/lyft/flyteplugins/go/tasks/v1/events"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/plugins"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/promutils/labeled"
	utils2 "github.com/lyft/flytestdlib/utils"

	tasksV1 "github.com/lyft/flyteplugins/go/tasks/v1"
	"github.com/lyft/flyteplugins/go/tasks/v1/errors"
	"github.com/lyft/flyteplugins/go/tasks/v1/qubole/client"
	"github.com/lyft/flyteplugins/go/tasks/v1/qubole/config"
	"github.com/lyft/flyteplugins/go/tasks/v1/resourcemanager"
	"github.com/lyft/flyteplugins/go/tasks/v1/types"
	"github.com/lyft/flyteplugins/go/tasks/v1/utils"
)

const ResyncDuration = 30 * time.Second
const hiveExecutorId = "hiveExecutor"
const hiveTaskType = "hive"

type HiveExecutor struct {
	types.OutputsResolver
	recorder        types.EventRecorder
	id              string
	secretsManager  SecretsManager
	executionsCache utils2.AutoRefreshCache
	metrics         HiveExecutorMetrics
	quboleClient    client.QuboleClient
	redisClient     *redis.Client
	resourceManager resourcemanager.ResourceManager
	executionBuffer resourcemanager.ExecutionLooksideBuffer
}

type HiveExecutorMetrics struct {
	Scope                 promutils.Scope
	ReleaseResourceFailed labeled.Counter
	AllocationGranted     labeled.Counter
	AllocationNotGranted  labeled.Counter
}

func (h HiveExecutor) GetID() types.TaskExecutorName {
	return h.id
}

func (h HiveExecutor) GetProperties() types.ExecutorProperties {
	return types.ExecutorProperties{
		RequiresFinalizer: true,
	}
}

func getHiveExecutorMetrics(scope promutils.Scope) HiveExecutorMetrics {
	return HiveExecutorMetrics{
		Scope: scope,
		ReleaseResourceFailed: labeled.NewCounter("released_resource_failed",
			"Error releasing allocation token", scope),
		AllocationGranted: labeled.NewCounter("allocation_granted",
			"Allocation request granted", scope),
		AllocationNotGranted: labeled.NewCounter("allocation_not_granted",
			"Allocation request did not fail but not granted", scope),
	}
}

// This runs once, after the constructor (since the constructor is called in the package init)
func (h *HiveExecutor) Initialize(ctx context.Context, param types.ExecutorInitializationParameters) error {
	// Make sure we can get the Qubole token, and set up metrics so we have the scope
	h.metrics = getHiveExecutorMetrics(param.MetricsScope)
	_, err := h.secretsManager.GetToken()
	if err != nil {
		return err
	}

	h.executionsCache, err = utils2.NewAutoRefreshCache(h.SyncQuboleQuery,
		utils2.NewRateLimiter("qubole-api-updater", 5, 15),
		ResyncDuration, config.GetQuboleConfig().LruCacheSize, param.MetricsScope.NewSubScope(hiveTaskType))
	if err != nil {
		return err
	}

	// Create Redis client
	redisHost := config.GetQuboleConfig().RedisHostPath
	redisPassword := config.GetQuboleConfig().RedisHostKey
	redisMaxRetries := config.GetQuboleConfig().RedisMaxRetries
	redisClient, err := resourcemanager.NewRedisClient(ctx, redisHost, redisPassword, redisMaxRetries)
	if err != nil {
		return err
	}
	h.redisClient = redisClient

	// Assign the resource manager here.  We do it here instead of the constructor because we need to pass in metrics
	resourceManager, err := resourcemanager.GetResourceManagerByType(ctx, config.GetQuboleConfig().ResourceManagerType,
		param.MetricsScope, h.redisClient)
	if err != nil {
		return err
	}
	h.resourceManager = resourceManager

	// Create a lookaside buffer in Redis to hold the command IDs created by Qubole
	expiryDuration := config.GetQuboleConfig().LookasideExpirySeconds.Duration
	h.executionBuffer = resourcemanager.NewRedisLookasideBuffer(ctx, h.redisClient,
		config.GetQuboleConfig().LookasideBufferPrefix, expiryDuration)

	h.recorder = param.EventRecorder

	h.executionsCache.Start(ctx)

	return nil
}

func (h HiveExecutor) getUniqueCacheKey(taskCtx types.TaskContext, idx int) string {
	// The cache will be holding all hive jobs across the engine, so it's imperative that the id of the cache
	// items be unique.  It should be unique across tasks, nodes, retries, etc. It also needs to be deterministic
	// so that we know what to look for in independent calls of CheckTaskStatus
	// Appending the index of the query should be sufficient.
	return fmt.Sprintf("%s_%d", taskCtx.GetTaskExecutionID().GetGeneratedName(), idx)
}

// This function is only ever called once, assuming it doesn't return in error.
// Essentially, what this function does is translate the task's custom field into the TaskContext's CustomState
// that's stored back into etcd
func (h HiveExecutor) StartTask(ctx context.Context, taskCtx types.TaskContext, task *core.TaskTemplate,
	inputs *core.LiteralMap) (types.TaskStatus, error) {
	// Fill in authorization stuff here in the future
	hiveJob := plugins.QuboleHiveJob{}
	err := utils.UnmarshalStruct(task.GetCustom(), &hiveJob)
	if err != nil {
		return types.TaskStatusPermanentFailure(errors.Errorf(errors.BadTaskSpecification,
			"Invalid Job Specification in task: [%v]. Err: [%v]", task.GetCustom(), err)), nil
	}

	// TODO: Asserts around queries, like len > 0 or something.

	// This custom state object will be passed back to us when the CheckTaskStatus call is received.
	customState := make(map[string]interface{})

	// Iterate through the queries that we'll need  to run, and create custom objects for them. We don't even
	// need to look at the query right now.  We won't attempt to run the query yet.
	for idx, q := range hiveJob.QueryCollection.Queries {
		fullFlyteKey := h.getUniqueCacheKey(taskCtx, idx)
		wrappedHiveJob := constructQuboleWorkItem(fullFlyteKey, "", QuboleWorkNotStarted)

		// Merge custom object Tags with labels to form Tags
		tags := hiveJob.Tags
		for k, v := range taskCtx.GetLabels() {
			tags = append(tags, fmt.Sprintf("%s:%s", k, v))
		}
		tags = append(tags, fmt.Sprintf("ns:%s", taskCtx.GetNamespace()))
		wrappedHiveJob.Tags = tags
		wrappedHiveJob.ClusterLabel = hiveJob.ClusterLabel
		wrappedHiveJob.Query = q.Query
		wrappedHiveJob.TimeoutSec = q.TimeoutSec

		customState[fullFlyteKey] = wrappedHiveJob
	}

	// This Phase represents the phase of the entire Job, of all the queries.
	// The Queued state is only ever used at the very beginning here.  The first CheckTaskStatus
	// call made on the object will return the running state.
	status := types.TaskStatus{
		Phase:        types.TaskPhaseQueued,
		PhaseVersion: 0,
		State:        customState,
	}

	return status, nil
}

func (h HiveExecutor) convertCustomStateToQuboleWorkItems(customState map[string]interface{}) (
	map[string]QuboleWorkItem, error) {

	m := make(map[string]QuboleWorkItem, len(customState))
	for k, v := range customState {
		// Cast the corresponding custom object
		item, err := InterfaceConverter(v)
		if err != nil {
			return map[string]QuboleWorkItem{}, err
		}
		m[k] = item
	}
	return m, nil
}

func (h HiveExecutor) CheckTaskStatus(ctx context.Context, taskCtx types.TaskContext, _ *core.TaskTemplate) (
	types.TaskStatus, error) {
	// Get the custom task information, and the custom state information
	customState := taskCtx.GetCustomState()
	logger.Infof(ctx, "Checking status for task execution [%s] Phase [%v] length of custom [%d]",
		taskCtx.GetTaskExecutionID().GetGeneratedName(), taskCtx.GetPhase(), len(customState))
	quboleApiKey, _ := h.secretsManager.GetToken()

	// Loop through all the queries and do whatever needs to be done
	// Also accumulate the new CustomState while iterating
	var newItems = make(map[string]interface{})
	quboleAttempts := 0
	quboleFailures := 0
	for workCacheKey, v := range customState {
		// Cast the corresponding custom object
		item, err := InterfaceConverter(v)
		if err != nil {
			logger.Errorf(ctx, "Error converting old state into an object for key %s", workCacheKey)
			return types.TaskStatusUndefined, err
		}
		logger.Debugf(ctx, "CheckTaskStatus, customState iteration - key [%s] id [%s] status [%s]",
			item.UniqueWorkCacheKey, item.CommandId, item.Status)

		// This copies the items in the cache into new objects.  It's important to leave the initial custom state
		// untouched because we compare new to old later for eventing.
		// This if block handles transitions from NotStarted to Running - i.e. attempt to create the query on Qubole
		// What happens if a job has ten queries and 8 of them launch successfully, but two of fail because
		// of a Qubole error that has nothing to do with the user's code, temporary Qubole flakiness for instance
		// To resolve this, we keep track of the errors in launching Qubole commands
		//   - if all calls fail, then we return a system level error
		//   - if only some calls fail, then that means we've updated the custom state with new Qubole command IDs
		//     so we shouldn't waste those, return as normal so that they get recorded
		if item.Status == QuboleWorkNotStarted {
			foundCommandId, err := h.executionBuffer.RetrieveExecution(ctx, workCacheKey)
			if err != nil {
				if err != resourcemanager.ExecutionNotFoundError {
					logger.Errorf(ctx, "Unable to retrieve from cache for %s", workCacheKey)
					return types.TaskStatusUndefined, errors.Wrapf(errors.DownstreamSystemError, err,
						"unable to retrieve from cache for %s", workCacheKey)
				}
				// Get an allocation token
				logger.Infof(ctx, "Attempting to get allocation token for %s", workCacheKey)
				allocationStatus, err := h.resourceManager.AllocateResource(ctx, taskCtx.GetNamespace(), workCacheKey)
				if err != nil {
					logger.Errorf(ctx, "Resource manager broke for [%s] key [%s], owner [%s]",
						taskCtx.GetTaskExecutionID().GetID(), workCacheKey, taskCtx.GetOwnerReference())
					return types.TaskStatusUndefined, err
				}
				logger.Infof(ctx, "Allocation result for [%s] is [%s]", workCacheKey, allocationStatus)

				// If successfully got an allocation token then kick off the query, and try to progress the job state
				// if no token was granted, we stay in the NotStarted state.
				if allocationStatus == resourcemanager.AllocationStatusGranted {
					h.metrics.AllocationGranted.Inc(ctx)

					quboleAttempts++
					// Note that the query itself doesn't live in the work item object that's cached.  That would take
					// up too much room in the workflow CRD - instead we iterate through the task's custom field
					// each time.
					cmdDetails, err := h.quboleClient.ExecuteHiveCommand(ctx, item.Query, item.TimeoutSec,
						item.ClusterLabel, quboleApiKey, item.Tags)
					if err != nil {
						// If we failed, we'll keep the NotStarted state
						logger.Warnf(ctx, "Error creating Qubole query for %s", item.UniqueWorkCacheKey)
						quboleFailures++
						// Deallocate token if Qubole API returns in error.
						err := h.resourceManager.ReleaseResource(ctx, taskCtx.GetNamespace(), workCacheKey)
						if err != nil {
							h.metrics.ReleaseResourceFailed.Inc(ctx)
						}
					} else {
						commandId := strconv.FormatInt(cmdDetails.ID, 10)
						logger.Infof(ctx, "Created Qubole ID %s for %s", commandId, workCacheKey)
						item.CommandId = commandId
						item.JobUri = cmdDetails.JobUri
						item.Status = QuboleWorkRunning
						item.Query = "" // Clear the query to save space in etcd once we've successfully launched
						err := h.executionBuffer.ConfirmExecution(ctx, workCacheKey, commandId)
						if err != nil {
							logger.Errorf(ctx, "Unable to record execution for %s", workCacheKey)
							return types.TaskStatusUndefined, errors.Wrapf(errors.DownstreamSystemError, err,
								"unable to record execution for %s", workCacheKey)
						}
					}
				} else {
					h.metrics.AllocationNotGranted.Inc(ctx)
					logger.Infof(ctx, "Unable to get allocation token for %s skipping...", workCacheKey)
				}
			} else {
				// If found, this means that we've previously kicked off this execution, but CheckTaskStatus
				// has been called with stale context.
				logger.Infof(ctx, "Unstarted Qubole work found in buffer for %s, setting to running with ID %s",
					workCacheKey, foundCommandId)
				item.CommandId = foundCommandId
				item.Status = QuboleWorkRunning
				item.Query = "" // Clear the query to save space in etcd once we've successfully launched
			}
		}

		// Add to the cache iff the item from the taskContext has a command ID (ie, has already been launched on Qubole)
		// Then check and update the item if necessary.
		if item.CommandId != "" {
			logger.Debugf(ctx, "Calling GetOrCreate for [%s] id [%s] status [%s]",
				item.UniqueWorkCacheKey, item.CommandId, item.Status)
			cc, err := h.executionsCache.GetOrCreate(item)
			if err != nil {
				// This means that our cache has fundamentally broken... return a system error
				logger.Errorf(ctx, "Cache is broken on execution [%s] cache key [%s], owner [%s]",
					taskCtx.GetTaskExecutionID().GetID(), workCacheKey, taskCtx.GetOwnerReference())
				return types.TaskStatusUndefined, err
			}

			cachedItem := cc.(QuboleWorkItem)
			logger.Debugf(ctx, "Finished GetOrCreate - cache key [%s]->[%s] status [%s]->[%s]",
				item.UniqueWorkCacheKey, cachedItem.UniqueWorkCacheKey, item.Status, cachedItem.Status)

			// TODO: Remove this sanity check if still here by late July 2019
			// This is a sanity check - get the item back immediately
			sanityCheck := h.executionsCache.Get(item.UniqueWorkCacheKey)
			if sanityCheck == nil {
				// This means that our cache has fundamentally broken... return a system error
				logger.Errorf(ctx, "Cache is b0rked!!!  Unless there are a lot of evictions happening, a GetOrCreate"+
					" has failed to actually create!!!  Cache key [%s], owner [%s]",
					workCacheKey, taskCtx.GetOwnerReference())
			} else {
				sanityCheckCast := sanityCheck.(QuboleWorkItem)
				logger.Debugf(ctx, "Immediate cache write check worked [%s] status [%s]",
					sanityCheckCast.UniqueWorkCacheKey, sanityCheckCast.Status)
			}
			// Handle all transitions after the initial one - If the one from the cache has a higher value,
			// that means our loop has done something, and we should update the new custom state to reflect that.
			if cachedItem.Status > item.Status {
				item.Status = cachedItem.Status
			}

			// Always copy the number of update retries
			item.Retries = cachedItem.Retries
		}

		// Always add the potentially modified item back to the new list so that it again can be persisted
		// into etcd
		newItems[workCacheKey] = item
	}

	// If all creation attempts fail, then report a system error
	if quboleFailures > 0 && quboleAttempts == quboleFailures {
		err := errors.Errorf(errors.DownstreamSystemError, "All %d Hive creation attempts failed for %s",
			quboleFailures, taskCtx.GetTaskExecutionID().GetGeneratedName())
		logger.Error(ctx, err)
		return types.TaskStatusUndefined, err
	}

	// Otherwise, look through the current state of things and decide what's up.
	newStatus := h.TranslateCurrentState(newItems)
	newStatus.PhaseVersion = taskCtx.GetPhaseVersion()

	// Determine whether or not to send an event.  If the phase has changed, then we definitely want to, if not,
	// we need to compare all the individual items to see if any were updated.
	var sendEvent = false
	if taskCtx.GetPhase() != newStatus.Phase {
		newStatus.PhaseVersion = 0
		sendEvent = true
	} else {
		oldItems, err := h.convertCustomStateToQuboleWorkItems(customState)
		if err != nil {
			// This error condition should not trigger because the exact same thing should've been done earlier
			logger.Errorf(ctx, "Error converting custom state %v", err)
			return types.TaskStatusUndefined, err
		}
		if !workItemMapsAreEqual(oldItems, newItems) {
			// If any of the items we updated, we also need to increment the version in order for admin to record it
			newStatus.PhaseVersion++
			sendEvent = true
		}
	}
	if sendEvent {
		info, err := constructEventInfoFromQuboleWorkItems(taskCtx, newStatus.State)
		if err != nil {
			logger.Errorf(ctx, "Error constructing event info for %s",
				taskCtx.GetTaskExecutionID().GetGeneratedName())
			return types.TaskStatusUndefined, err
		}

		ev := events.CreateEvent(taskCtx, newStatus, info)

		err = h.recorder.RecordTaskEvent(ctx, ev)
		if err != nil && eventErrors.IsEventAlreadyInTerminalStateError(err) {
			return types.TaskStatusPermanentFailure(errors.Wrapf(errors.TaskEventRecordingFailed, err,
				"failed to record task event. state mis-match between Propeller %v and Control Plane.", &ev.Phase)), nil
		} else if err != nil {
			return types.TaskStatusUndefined, errors.Wrapf(errors.TaskEventRecordingFailed, err,
				"failed to record task event")
		}
	}
	logger.Debugf(ctx, "Task [%s] phase [%s]->[%s] phase version [%d]->[%d] sending event: %s",
		taskCtx.GetTaskExecutionID().GetGeneratedName(), taskCtx.GetPhase(), newStatus.Phase, taskCtx.GetPhaseVersion(),
		newStatus.PhaseVersion, sendEvent)

	return newStatus, nil
}

// This translates a series of QuboleWorkItem statuses into what it means for the task as a whole
func (h HiveExecutor) TranslateCurrentState(state map[string]interface{}) types.TaskStatus {
	succeeded := 0
	failed := 0
	total := len(state)
	status := types.TaskStatus{
		State: state,
	}

	for _, k := range state {
		workItem := k.(QuboleWorkItem)
		if workItem.Status == QuboleWorkSucceeded {
			succeeded++
		} else if workItem.Status == QuboleWorkFailed {
			failed++
		}
	}

	if succeeded == total {
		status.Phase = types.TaskPhaseSucceeded
	} else if failed > 0 {
		status.Phase = types.TaskPhaseRetryableFailure
		status.Err = errors.Errorf(errors.DownstreamSystemError, "Qubole job failed")
	} else {
		status.Phase = types.TaskPhaseRunning
	}

	return status
}

// Loop through all the queries in the task, if there are any in a non-terminal state, then
// submit the request to terminate the Qubole query.  If there are any problems with anything, then return
// an error
func (h HiveExecutor) KillTask(ctx context.Context, taskCtx types.TaskContext, reason string) error {
	// Is it ever possible to get a CheckTaskStatus call for a task while this function is running?
	// Or immediately after this function runs?
	customState := taskCtx.GetCustomState()
	logger.Infof(ctx, "Kill task called on [%s] with [%d] customs", taskCtx.GetTaskExecutionID().GetGeneratedName(),
		len(customState))
	quboleApiKey, _ := h.secretsManager.GetToken()

	var callsWithErrors = make([]string, 0, len(customState))

	for key, value := range customState {
		work, err := InterfaceConverter(value)
		if err != nil {
			logger.Errorf(ctx, "Error converting old state into an object for key %s", work.UniqueWorkCacheKey)
			return err
		}
		logger.Debugf(ctx, "KillTask processing custom item key [%s] id [%s] on cluster [%s]",
			work.UniqueWorkCacheKey, work.CommandId, work.ClusterLabel)

		status, err := h.quboleClient.GetCommandStatus(ctx, work.CommandId, quboleApiKey)
		if err != nil {
			logger.Errorf(ctx, "Problem getting command status while terminating %s %s %v",
				work.CommandId, taskCtx.GetTaskExecutionID().GetGeneratedName(), err)
			callsWithErrors = append(callsWithErrors, work.CommandId)
			continue
		}

		if !QuboleWorkIsTerminalState(QuboleStatusToWorkItemStatus(status)) {
			logger.Debugf(ctx, "Terminating cache item [%s] id [%s] status [%s]",
				work.UniqueWorkCacheKey, work.CommandId, work.Status)

			err := h.quboleClient.KillCommand(ctx, work.CommandId, quboleApiKey)
			if err != nil {
				logger.Errorf(ctx, "Error stopping Qubole command in termination sequence %s from %s with %v",
					work.CommandId, key, err)
				callsWithErrors = append(callsWithErrors, work.CommandId)
				continue
			}
			err = h.resourceManager.ReleaseResource(ctx, "", work.UniqueWorkCacheKey)
			if err != nil {
				logger.Errorf(ctx, "Failed to release resource [%s]", work.UniqueWorkCacheKey)
				h.metrics.ReleaseResourceFailed.Inc(ctx)
			}
			logger.Debugf(ctx, "Finished terminating cache item [%s] id [%s]",
				work.UniqueWorkCacheKey, work.CommandId)
		} else {
			logger.Debugf(ctx, "Custom work in terminal state [%s] id [%s] status [%s]",
				work.UniqueWorkCacheKey, work.CommandId, work.Status)

			// This is idempotent anyways, just be tripley safe we're not leaking resources
			err := h.resourceManager.ReleaseResource(ctx, "", work.UniqueWorkCacheKey)
			if err != nil {
				logger.Errorf(ctx, "Failed to release resource [%s]", work.UniqueWorkCacheKey)
				h.metrics.ReleaseResourceFailed.Inc(ctx)
			}
		}
	}

	if len(callsWithErrors) > 0 {
		return errors.Errorf(errors.DownstreamSystemError, "%d errors found for Qubole commands %v",
			len(callsWithErrors), callsWithErrors)
	}

	return nil
}

// This should do minimal work - basically grab an updated status from the Qubole API and store it in the cache
// All other handling should be in the synchronous loop.
func (h *HiveExecutor) SyncQuboleQuery(ctx context.Context, obj utils2.CacheItem) (
	utils2.CacheItem, utils2.CacheSyncAction, error) {

	workItem := obj.(QuboleWorkItem)

	// TODO: Remove this if block if still here by late July 2019.  This should not happen any more ever.
	if workItem.CommandId == "" {
		logger.Debugf(ctx, "Sync loop - CommandID is blank for [%s] skipping", workItem.UniqueWorkCacheKey)
		// No need to do anything if the work hasn't been kicked off yet
		return workItem, utils2.Unchanged, nil
	}

	logger.Debugf(ctx, "Sync loop - processing Hive job [%s] - cache key [%s]",
		workItem.CommandId, workItem.UniqueWorkCacheKey)

	quboleApiKey, _ := h.secretsManager.GetToken()

	if QuboleWorkIsTerminalState(workItem.Status) {
		// Release again - this is idempotent anyways, shouldn't be a huge deal to be on the safe side and release
		// many times.
		logger.Debugf(ctx, "Sync loop - Qubole id [%s] in terminal state, re-releasing cache key [%s]",
			workItem.CommandId, workItem.UniqueWorkCacheKey)

		err := h.resourceManager.ReleaseResource(ctx, "", workItem.UniqueWorkCacheKey)
		if err != nil {
			h.metrics.ReleaseResourceFailed.Inc(ctx)
		}
		return workItem, utils2.Unchanged, nil
	}

	// Get an updated status from Qubole
	logger.Debugf(ctx, "Querying Qubole for %s - %s", workItem.CommandId, workItem.UniqueWorkCacheKey)
	commandStatus, err := h.quboleClient.GetCommandStatus(ctx, workItem.CommandId, quboleApiKey)
	if err != nil {
		logger.Errorf(ctx, "Error from Qubole command %s", workItem.CommandId)
		workItem.Retries++
		// Make sure we don't return nil for the first argument, because that deletes it from the cache.
		return workItem, utils2.Update, err
	}
	workItemStatus := QuboleStatusToWorkItemStatus(commandStatus)

	// Careful how we call this, don't want to ever go backwards, unless it's unknown
	if workItemStatus > workItem.Status || workItemStatus == QuboleWorkUnknown {
		workItem.Status = workItemStatus
		logger.Infof(ctx, "Moving Qubole work %s %s from %s to %s", workItem.CommandId, workItem.UniqueWorkCacheKey,
			workItem.Status, workItemStatus)

		if QuboleWorkIsTerminalState(workItem.Status) {
			err := h.resourceManager.ReleaseResource(ctx, "", workItem.UniqueWorkCacheKey)
			if err != nil {
				h.metrics.ReleaseResourceFailed.Inc(ctx)
			}
		}

		return workItem, utils2.Update, nil
	}

	return workItem, utils2.Unchanged, nil
}

func NewHiveTaskExecutorWithCache(ctx context.Context) (*HiveExecutor, error) {
	hiveExecutor := HiveExecutor{
		id:             hiveExecutorId,
		secretsManager: NewSecretsManager(),
		quboleClient:   client.NewQuboleClient(),
	}

	return &hiveExecutor, nil
}

func NewHiveTaskExecutor(ctx context.Context, executorId string, executorClient client.QuboleClient) (*HiveExecutor, error) {
	hiveExecutor := HiveExecutor{
		id:             executorId,
		secretsManager: NewSecretsManager(),
		quboleClient:   executorClient,
	}

	return &hiveExecutor, nil
}

func init() {
	tasksV1.RegisterLoader(func(ctx context.Context) error {
		hiveExecutor, err := NewHiveTaskExecutorWithCache(ctx)
		if err != nil {
			return err
		}

		return tasksV1.RegisterForTaskTypes(hiveExecutor, hiveTaskType)
	})
}
