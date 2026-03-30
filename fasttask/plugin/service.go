package plugin

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/types"

	idlCore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/transformers/k8s"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"

	"github.com/unionai/flyte/fasttask/plugin/interfaces"
	"github.com/unionai/flyte/fasttask/plugin/pb"
)

const (
	maxPendingOwnersPerQueue = 4096
	queueIdLabel             = "queue_id"

	cleanedUpTasksTTL        = 5 * time.Minute
	cleanedUpTasksGCInterval = 1 * time.Minute
)

type fastTaskExecutionMetric struct {
	prom  *prometheus.CounterVec
	cache *cache.Cache
}

const fastTaskMetricKeySeparator = ":"

func newFastTaskExecutionMetric(scope promutils.Scope, cfg *FastTaskExecutionMetricConfig) fastTaskExecutionMetric {
	prom := scope.MustNewCounterVec("fast_task_execution_duration",
		"Cumulative duration of all fast task executions within a given execution",
		"org", "project", "domain", "execution_id", "namespace", "pod")
	c := cache.New(cfg.TTL.Duration, cfg.CleanupInterval.Duration)
	c.OnEvicted(func(key string, _ interface{}) {
		splt := strings.SplitN(key, fastTaskMetricKeySeparator, 6)
		if len(splt) != 4 {
			return
		}
		org, project, domain, execName, namespace, pod := splt[0], splt[1], splt[2], splt[3], splt[4], splt[5]
		prom.DeleteLabelValues(org, project, domain, execName, namespace, pod)
	})
	return fastTaskExecutionMetric{
		prom:  prom,
		cache: c,
	}
}

func (m *fastTaskExecutionMetric) add(execID *pb.ExecutionIdentifier, namespace, pod string, duration time.Duration) {
	org, project, domain, execName := execID.GetOrg(), execID.GetProject(), execID.GetDomain(), execID.GetName()
	m.prom.WithLabelValues(org, project, domain, execName, namespace, pod).Add(duration.Seconds())
	key := strings.Join([]string{org, project, domain, execName, namespace, pod}, fastTaskMetricKeySeparator)
	m.cache.Set(key, nil, cache.DefaultExpiration)
}

// serviceMetrics defines a collection of metrics for the fasttask service.
type serviceMetrics struct {
	taskNoWorkersAvailable    prometheus.Counter
	taskNoCapacityAvailable   prometheus.Counter
	taskAssigned              prometheus.Counter
	queues                    *prometheus.Desc
	workers                   *prometheus.Desc
	executions                *prometheus.Desc
	executionsLimit           *prometheus.Desc
	backlog                   *prometheus.Desc
	backlogLimit              *prometheus.Desc
	workerConnectionErrors    *prometheus.CounterVec
	enqueueOwnerFailure       prometheus.Counter
	pendingOwnerQueueFull     prometheus.Counter
	fastTaskExecutionDuration fastTaskExecutionMetric
}

// newServiceMetrics creates a new serviceMetrics with the given scope.
func newServiceMetrics(scope promutils.Scope, cfg *Config) serviceMetrics {
	return serviceMetrics{
		taskNoWorkersAvailable:    scope.MustNewCounter("task_no_workers_available", "Count of task assignment attempts with no workers available"),
		taskNoCapacityAvailable:   scope.MustNewCounter("task_no_capacity_available", "Count of task assignment attempts with no capacity available"),
		taskAssigned:              scope.MustNewCounter("task_assigned", "Count of task assignments"),
		queues:                    prometheus.NewDesc(scope.NewScopedMetricName("queues"), "Current number of queues", nil, nil),
		workers:                   prometheus.NewDesc(scope.NewScopedMetricName("workers"), "Current number of workers per queue", []string{queueIdLabel}, nil),
		executions:                prometheus.NewDesc(scope.NewScopedMetricName("executions"), "Current number of task executions per queue", []string{queueIdLabel}, nil),
		executionsLimit:           prometheus.NewDesc(scope.NewScopedMetricName("executions_limit"), "Total executions limit per queue", []string{queueIdLabel}, nil),
		backlog:                   prometheus.NewDesc(scope.NewScopedMetricName("backlog"), "Current number of backlogged tasks per queue", []string{queueIdLabel}, nil),
		backlogLimit:              prometheus.NewDesc(scope.NewScopedMetricName("backlog_limit"), "Total backlog limit per queue", []string{queueIdLabel}, nil),
		enqueueOwnerFailure:       scope.MustNewCounter("enqueue_owner_failure", "Count of tasks that failed to enqueue a workflow to be re-evaluated"),
		pendingOwnerQueueFull:     scope.MustNewCounter("pending_owner_queue_full", "Count of tasks that could not add pending owner due to queue being full"),
		workerConnectionErrors:    scope.MustNewCounterVec("worker_connection_errors", "Count of errors encountered during worker connection communications", "error_type"),
		fastTaskExecutionDuration: newFastTaskExecutionMetric(scope, &cfg.FastTaskExecutionMetric),
	}
}

type fastTaskServiceImpl struct {
	pb.UnimplementedFastTaskServer
	enqueueOwner core.EnqueueOwner

	builder            interfaces.EnvironmentBuilder
	store              interfaces.EnvironmentStore
	taskStatusChannels *sync.Map    // map[taskID]chan interfaces.WorkerTaskStatus
	cleanedUpTasks     *cache.Cache // TTL-based tombstones for DELETE re-send after Cleanup
	metrics            serviceMetrics

	// A map of pending owners by queue. When a new worker becomes available, use this to enqueue owners for reevaluation.
	// Note, this is an optimistic approach and may not include all pending owners.
	pendingTaskOwners     map[string]map[string]map[string]string // map[queueID]map[taskID]enqueueLabels
	pendingTaskOwnersLock sync.RWMutex

	// Tasks assigned to workers but not yet confirmed by a heartbeat status report.
	// OfferTask adds entries; processHeartbeatTaskStatuses and Cleanup remove them.
	pendingReservations       map[string]map[string]map[string]struct{} // map[queueID]map[workerID]map[taskID]
	pendingReservationsByTask map[string]pendingReservationInfo         // map[taskID] -> location
	pendingReservationsLock   sync.Mutex
}

// pendingReservationInfo records where a task's pending reservation lives
// in the forward map so that idempotent retries can find it in O(1).
type pendingReservationInfo struct {
	queueID  string
	workerID string
}

var _ interfaces.FastTaskService = (*fastTaskServiceImpl)(nil)
var _ pb.FastTaskServer = (*fastTaskServiceImpl)(nil)

func (f *fastTaskServiceImpl) Describe(ch chan<- *prometheus.Desc) {
	ch <- f.metrics.queues
	ch <- f.metrics.workers
	ch <- f.metrics.executions
	ch <- f.metrics.executionsLimit
	ch <- f.metrics.backlog
	ch <- f.metrics.backlogLimit
}

// Collect emits snapshot metrics for the fasttask service. This is useful to periodically capture the state of queues and workers.
func (f *fastTaskServiceImpl) Collect(ch chan<- prometheus.Metric) {
	logger.Info(context.Background(), "Collecting fasttask service metrics")

	environments := f.store.List()

	ch <- prometheus.MustNewConstMetric(f.metrics.queues, prometheus.GaugeValue, float64(len(environments)))
	for _, env := range environments {
		executions := int32(0)
		executionsLimit := int32(0)
		backlog := int32(0)
		backlogLimit := int32(0)
		totalWorkers := 0
		env.RangeWorkers(func(workerID string, worker interfaces.Worker) bool {
			executions += worker.Capacity().GetExecutionCount()
			executionsLimit += worker.Capacity().GetExecutionLimit()
			backlog += worker.Capacity().GetBacklogCount()
			backlogLimit += worker.Capacity().GetBacklogLimit()

			totalWorkers++
			return true
		})
		ch <- prometheus.MustNewConstMetric(f.metrics.workers, prometheus.GaugeValue, float64(totalWorkers), env.EnvID().String())
		ch <- prometheus.MustNewConstMetric(f.metrics.executionsLimit, prometheus.GaugeValue, float64(executionsLimit), env.EnvID().String())
		ch <- prometheus.MustNewConstMetric(f.metrics.executions, prometheus.GaugeValue, float64(executions), env.EnvID().String())
		ch <- prometheus.MustNewConstMetric(f.metrics.backlog, prometheus.GaugeValue, float64(backlog), env.EnvID().String())
		ch <- prometheus.MustNewConstMetric(f.metrics.backlogLimit, prometheus.GaugeValue, float64(backlogLimit), env.EnvID().String())
	}
}

// workerTaskStatus represents the status of a task as reported by a worker.
type workerTaskStatus struct {
	workerID   string
	taskStatus *pb.TaskStatus
}

func (f *fastTaskServiceImpl) addPendingReservation(queueID, workerID, taskID string) {
	// caller must hold pendingReservationsLock
	queuePending, ok := f.pendingReservations[queueID]
	if !ok {
		queuePending = make(map[string]map[string]struct{})
		f.pendingReservations[queueID] = queuePending
	}
	workerPending, ok := queuePending[workerID]
	if !ok {
		workerPending = make(map[string]struct{})
		queuePending[workerID] = workerPending
	}
	workerPending[taskID] = struct{}{}
	f.pendingReservationsByTask[taskID] = pendingReservationInfo{queueID: queueID, workerID: workerID}
}

// removeReservationLocked removes a single reservation from both maps.
// Caller must hold pendingReservationsLock.
func (f *fastTaskServiceImpl) removeReservationLocked(queueID, workerID, taskID string) {
	if queuePending, ok := f.pendingReservations[queueID]; ok {
		if workerPending, ok := queuePending[workerID]; ok {
			delete(workerPending, taskID)
			if len(workerPending) == 0 {
				delete(queuePending, workerID)
			}
		}
		if len(queuePending) == 0 {
			delete(f.pendingReservations, queueID)
		}
	}
	delete(f.pendingReservationsByTask, taskID)
}

func (f *fastTaskServiceImpl) removePendingReservation(queueID, workerID, taskID string) {
	f.pendingReservationsLock.Lock()
	defer f.pendingReservationsLock.Unlock()
	f.removeReservationLocked(queueID, workerID, taskID)
}

// AddPendingOwner adds to the pending owners list for the queue, if not already full.
// Returns true if the task was newly added, false if it was already pending or queue is full.
func (f *fastTaskServiceImpl) AddPendingOwner(queueID, taskID string, enqueueLabels map[string]string) bool {
	f.pendingTaskOwnersLock.Lock()
	defer f.pendingTaskOwnersLock.Unlock()

	owners, exists := f.pendingTaskOwners[queueID]
	if !exists {
		owners = make(map[string]map[string]string)
		f.pendingTaskOwners[queueID] = owners
	}

	if len(owners) >= maxPendingOwnersPerQueue {
		f.metrics.pendingOwnerQueueFull.Inc()
		return false
	}

	if _, alreadyExists := owners[taskID]; alreadyExists {
		return false
	}

	owners[taskID] = enqueueLabels
	return true
}

// RemovePendingOwner removes the pending owner from the list if still there
func (f *fastTaskServiceImpl) RemovePendingOwner(queueID, taskID string) {
	f.pendingTaskOwnersLock.Lock()
	defer f.pendingTaskOwnersLock.Unlock()

	owners, exists := f.pendingTaskOwners[queueID]
	if !exists {
		return
	}

	delete(owners, taskID)
	if len(owners) == 0 {
		delete(f.pendingTaskOwners, queueID)
	}
}

// EnqueuePendingOwners removes the specified number of pending owners for the queue and enqueues
// them for reevaluation.
func (f *fastTaskServiceImpl) EnqueuePendingOwners(queueID string, count int) {
	f.pendingTaskOwnersLock.Lock()
	defer f.pendingTaskOwnersLock.Unlock()

	owners, exists := f.pendingTaskOwners[queueID]
	if !exists {
		return
	}

	remainingCount := count
	deleteOwners := make([]string, 0)
	enqueued := make(map[string]bool)
	for owner, ownerLabels := range owners {
		// decrement count and track owner as a deletion candidate
		if remainingCount <= 0 {
			break
		}

		remainingCount--
		deleteOwners = append(deleteOwners, owner)

		// avoid duplicate enqueues over the same labels (ie. v1 workflows)
		hashedLabels := hashMapValues(ownerLabels)
		if _, ok := enqueued[hashedLabels]; ok {
			continue
		}

		// enqueue owner and track labels as enqueued
		if err := f.enqueueOwner(ownerLabels); err != nil {
			logger.Warnf(context.Background(), "failed to enqueue owner %v: %+v", ownerLabels, err)
			f.metrics.enqueueOwnerFailure.Inc()
		}

		enqueued[hashedLabels] = true
	}

	// cleanup processed owners
	for _, owner := range deleteOwners {
		delete(owners, owner)
	}

	if len(owners) == 0 {
		delete(f.pendingTaskOwners, queueID)
	}
}

// CheckStatus checks the status of a task on a specific queue and worker.
func (f *fastTaskServiceImpl) CheckStatus(ctx context.Context, taskID, queueID, workerID string) (interfaces.TaskStatus, error) {
	taskStatusChannelResult, exists := f.taskStatusChannels.Load(taskID)
	if !exists {
		// if this plugin restarts then TaskContexts may not exist for tasks that are still active. we can
		// create a TaskContext here because we ensure it will be cleaned up when the task completes.
		f.taskStatusChannels.Store(taskID, make(chan *workerTaskStatus, GetConfig().TaskStatusBufferSize))
		return interfaces.TaskStatus{}, fmt.Errorf("task context not found: %w", taskContextNotFoundError)
	}

	taskStatusChannel := taskStatusChannelResult.(chan *workerTaskStatus)

	var latestWorkerTaskStatus *workerTaskStatus
Loop:
	for {
		select {
		case x := <-taskStatusChannel:
			// ensure we retrieve the latest status from the worker that is currently assigned to the task
			if x.workerID == workerID {
				latestWorkerTaskStatus = x
			}
		default:
			break Loop
		}
	}

	if latestWorkerTaskStatus == nil {
		return interfaces.TaskStatus{}, fmt.Errorf("unable to find task status update: %w", statusUpdateNotFoundError)
	}

	taskStatus := latestWorkerTaskStatus.taskStatus
	phase := core.Phase(taskStatus.GetPhase())

	// if not completed need to send ACK on taskID to worker
	if phase != core.PhaseSuccess && phase != core.PhaseRetryableFailure {
		if env := f.store.Get(queueID); env != nil {
			if worker := env.GetWorker(workerID); worker != nil {
				if worker.State() == interfaces.HEALTHY {
					worker.EnqueueHeartbeatResponse(&pb.HeartbeatResponse{
						TaskId:    taskID,
						Operation: pb.HeartbeatResponse_ACK,
					})
					worker.SetLastAccessedAt(time.Now().Unix())
				}
			}
		}
	}

	return interfaces.TaskStatus{
		Phase:         phase,
		Reason:        taskStatus.GetReason(),
		TaskDuration:  taskStatus.GetTaskDuration().AsDuration(),
		SystemFailure: taskStatus.GetSystemFailure(),
	}, nil
}

// Cleanup is used to indicate a task is no longer being tracked by the worker and delete the
// associated task context.
func (f *fastTaskServiceImpl) Cleanup(ctx context.Context, taskID, queueID, workerID string) error {
	// Record that this task was explicitly cleaned up so that
	// processHeartbeatTaskStatuses can re-send DELETE if the worker still
	// reports it. Must be set before deleting the channel (the gate check).
	f.cleanedUpTasks.Set(taskID, struct{}{}, cache.DefaultExpiration)

	// Close the gate first — prevents processHeartbeatTaskStatuses from
	// re-registering this task after we unregister it below.
	f.taskStatusChannels.Delete(taskID)

	if env := f.store.Get(queueID); env != nil {
		if worker := env.GetWorker(workerID); worker != nil {
			worker.EnqueueHeartbeatResponse(&pb.HeartbeatResponse{
				TaskId:    taskID,
				Operation: pb.HeartbeatResponse_DELETE,
			})
		}
		env.UnregisterTask(taskID)
	}

	// remove pending owner and reservation
	f.RemovePendingOwner(queueID, taskID)
	f.removePendingReservation(queueID, workerID, taskID)

	return nil
}

// Heartbeat is a gRPC stream that manages the heartbeat of a fasttask worker. This includes
// receiving task status updates and sending task assignments.
func (f *fastTaskServiceImpl) Heartbeat(stream pb.FastTask_HeartbeatServer) error {
	workerID := ""

	// recv initial heartbeat request
	heartbeatRequest, err := stream.Recv()
	if heartbeatRequest != nil {
		workerID = heartbeatRequest.GetWorkerId()
	}

	if err == io.EOF || heartbeatRequest == nil {
		// EOF is the error that's returned when the client closes the stream
		logger.Debugf(context.Background(), "heartbeat stream closed for worker %s", workerID)
		return nil
	} else if err != nil {
		f.metrics.workerConnectionErrors.WithLabelValues("initial_connection").Inc()
		// probably never seen this.
		return err
	}

	logger.Debugf(context.Background(), "received initial heartbeat for worker %s", workerID)

	// connect this worker to an existing environment. we wait (up to 30s) for the environment to
	// be created from either (1) a new task or (2) orphan detection to ensure the environment
	// contains necessary metadata to `scaleDown`.
	var env interfaces.Environment
	// builder creates environments. builder may not create the environment in the store, before the
	// replica comes up and starts to heartbeat.
	for i := 0; i < 600; i++ {
		// queueid is the actor environment id, proj/domain etc.
		env = f.store.Get(heartbeatRequest.GetQueueId())
		if env != nil {
			break
		}

		select {
		case <-stream.Context().Done():
			return nil
		case <-time.After(50 * time.Millisecond):
			// continue
		}
	}

	if env == nil {
		// if a replica listens on a queue baz that doesn't exist, then it should get an error.
		return fmt.Errorf("environment %s not found", heartbeatRequest.GetQueueId())
	}

	worker := env.GetOrCreateWorker(heartbeatRequest.GetWorkerId())
	worker.SetState(interfaces.HEALTHY, "")

	capacity := heartbeatRequest.GetCapacity()
	worker.SetCapacity(capacity)
	if executionCapacity := capacity.GetExecutionLimit() - capacity.GetExecutionCount(); executionCapacity > 0 {
		f.EnqueuePendingOwners(heartbeatRequest.GetQueueId(), int(executionCapacity))
	}

	// if the worker disconnects, we transition state to `ORPHANED` to ensure no future tasks are
	// assigned. if the worker reconnects, we will transition back to `HEALTHY`; otherwise, the
	// worker will be cleaned up by the `scaleDown` operation.
	defer func() {
		worker.SetState(interfaces.ORPHANED, "")
	}()

	// start go routine to handle heartbeat responses
	go func() {
		for {
			select {
			case message := <-worker.Responses():
				if err := stream.Send(message); err != nil {
					// Errors in sending should surface itself as other problems elsewhere (like grace period timeouts, etc.)
					f.metrics.workerConnectionErrors.WithLabelValues("send").Inc()
					logger.Warnf(context.Background(), "failed to send heartbeat response %+v", message)
				}
			case <-stream.Context().Done():
				// this is when the connection is dropped.
				return
			}
		}
	}()

	// process task statuses from the initial heartbeat (e.g. final heartbeat on SIGTERM)
	f.processHeartbeatTaskStatuses(heartbeatRequest, env, worker)

	// handle heartbeat requests
	for {
		heartbeatRequest, err := stream.Recv()
		if err == io.EOF || heartbeatRequest == nil {
			logger.Debugf(context.Background(), "heartbeat stream closed for worker %s", workerID)
			break
		} else if err != nil {
			f.metrics.workerConnectionErrors.WithLabelValues("receive").Inc()
			logger.Warnf(context.Background(), "failed to recv heartbeat request %+v", err)
			continue
		}

		// update worker capacity
		capacity := heartbeatRequest.GetCapacity()
		worker.SetCapacity(capacity)
		if executionCapacity := capacity.GetExecutionLimit() - capacity.GetExecutionCount(); executionCapacity > 0 {
			f.EnqueuePendingOwners(heartbeatRequest.GetQueueId(), int(executionCapacity))
		}

		f.processHeartbeatTaskStatuses(heartbeatRequest, env, worker)
	}

	return nil
}

func (f *fastTaskServiceImpl) processHeartbeatTaskStatuses(heartbeatRequest *pb.HeartbeatRequest, env interfaces.Environment, worker interfaces.Worker) {
	for _, taskStatus := range heartbeatRequest.GetTaskStatuses() {
		taskID := taskStatus.GetTaskId()

		// Only register demand and forward status when the channel exists.
		// Missing channel means either Cleanup ran (re-registering would leak
		// demand permanently) or plugin restarted (CheckStatus will recreate
		// the channel on the next evaluation).
		taskStatusChannelResult, channelExists := f.taskStatusChannels.Load(taskID)
		if channelExists {
			env.RegisterTask(taskID)

			// Double-check: if Cleanup deleted the channel between our Load
			// and RegisterTask, undo the registration to prevent demand
			// inflation. Cleanup deletes the channel before calling
			// UnregisterTask, so a missing channel here means Cleanup is
			// running concurrently.
			if _, stillExists := f.taskStatusChannels.Load(taskID); !stillExists {
				env.UnregisterTask(taskID)
				channelExists = false
			}
		}

		if channelExists {
			// Per-task confirmation: the worker reported a status for this task,
			// so it has received the ASSIGN. Remove the pending reservation.
			f.removePendingReservation(heartbeatRequest.GetQueueId(), worker.ID(), taskID)

			taskStatusChannel := taskStatusChannelResult.(chan *workerTaskStatus)
			taskStatusChannel <- &workerTaskStatus{
				workerID:   worker.ID(),
				taskStatus: taskStatus,
			}
		} else if _, cleanedUp := f.cleanedUpTasks.Get(taskID); cleanedUp {
			// Cleanup explicitly ran for this task but the worker is still
			// reporting it — the original DELETE was likely lost. Re-send.
			worker.EnqueueHeartbeatResponse(&pb.HeartbeatResponse{
				TaskId:    taskID,
				Operation: pb.HeartbeatResponse_DELETE,
			})
		}

		// Metrics and fast-feedback run unconditionally — after Cleanup this
		// may produce a benign duplicate reevaluation at the Flyte level;
		// after restart it preserves observability and fast completion feedback.
		execID := taskStatus.GetExecId()
		if taskStatus.GetTaskDuration().AsDuration() > 0 {
			f.metrics.fastTaskExecutionDuration.add(execID,
				taskStatus.GetNamespace(),
				heartbeatRequest.GetWorkerId(),
				taskStatus.GetTaskDuration().AsDuration())
		}

		// if taskStatus is complete then enqueueOwner for fast feedback
		phase := core.Phase(taskStatus.GetPhase())
		if phase == core.PhaseSuccess || phase == core.PhaseRetryableFailure {
			labels := make(map[string]string)

			// for backwards compatibility, add workflow id
			namespacedName := types.NamespacedName{
				Namespace: taskStatus.GetNamespace(),
				Name:      taskStatus.GetWorkflowId(),
			}
			labels[k8s.WorkflowID] = namespacedName.String()

			for label, value := range taskStatus.GetEnqueueLabels() {
				labels[label] = value
			}

			if err := f.enqueueOwner(labels); err != nil {
				f.metrics.enqueueOwnerFailure.Inc()
				logger.Warnf(context.Background(), "failed to enqueue owner for task %s: %+v", taskID, err)
			}
		}
	}
}

func (f *fastTaskServiceImpl) OfferTaskToEnvironment(ctx context.Context, execID *idlCore.WorkflowExecutionIdentifier, environmentID, taskID, namespace, workflowID string, cmd []string, envVars map[string]string, enqueueLabels map[string]string) (interfaces.Worker, error) {
	// retrieve environment
	env := f.store.Get(environmentID)
	if env == nil {
		f.metrics.taskNoWorkersAvailable.Inc()
		return nil, fmt.Errorf("environment '%s' not found", environmentID)
	}

	// track task for demand-based scaling - register when attempting to offer
	// so demand includes tasks waiting for capacity
	env.RegisterTask(taskID)

	// Select a worker and reserve a slot under the pending reservations lock.
	// The lock serializes concurrent OfferTask calls, preventing two goroutines
	// from both reading the same capacity and double-booking a slot.
	f.pendingReservationsLock.Lock()

	// Idempotent retry: if this task already has a reservation, reuse the same
	// worker. This handles the case where OfferTask succeeded but the caller
	// failed before persisting state and is now retrying with the same taskID.
	var worker interfaces.Worker
	if info, exists := f.pendingReservationsByTask[taskID]; exists && info.queueID == environmentID {
		if w := env.GetWorker(info.workerID); w != nil && w.State() == interfaces.HEALTHY {
			worker = w
		} else {
			// Worker gone or unhealthy — remove stale reservation, fall through to re-select
			f.removeReservationLocked(info.queueID, info.workerID, taskID)
		}
	}

	if worker == nil {
		// identify preferred (ie. has capacity) and acceptable (ie. has backlog capacity) worker(s)
		queuePending := f.pendingReservations[environmentID]
		preferredWorkers := make([]interfaces.Worker, 0)
		acceptableWorkers := make([]interfaces.Worker, 0)
		env.RangeWorkers(func(workerID string, w interfaces.Worker) bool {
			if w.State() != interfaces.HEALTHY {
				return true
			}
			cap := w.Capacity()
			pendingCount := int32(len(queuePending[workerID]))
			if cap.GetExecutionCount()+pendingCount < cap.GetExecutionLimit() {
				preferredWorkers = append(preferredWorkers, w)
			} else {
				totalUsed := cap.GetExecutionCount() + cap.GetBacklogCount() + pendingCount
				totalCap := cap.GetExecutionLimit() + cap.GetBacklogLimit()
				if totalUsed < totalCap {
					acceptableWorkers = append(acceptableWorkers, w)
				}
			}
			return true
		})

		// we sort workers by ID to ensure determinism in the selection process. this is important to
		// (1) assign to the same worker in case of failures between assignment and state persistence
		// and (2) allow the maximum number of workers to be cleaned up if unused. each worker
		// maintains a `lastAccessedAt` timestamp to determine when it was last used. by assigning
		// tasks to workers alphabetically, we can ensure that workers are allowed to become stale.
		if len(preferredWorkers) > 0 {
			sort.Slice(preferredWorkers, func(i, j int) bool {
				return preferredWorkers[i].ID() < preferredWorkers[j].ID()
			})

			for _, w := range preferredWorkers {
				cap := w.Capacity()
				pendingCount := int32(len(queuePending[w.ID()]))
				if cap.GetExecutionCount()+pendingCount >= cap.GetExecutionLimit() {
					continue
				}
				f.addPendingReservation(environmentID, w.ID(), taskID)
				worker = w
				break
			}
		}

		if worker == nil && len(acceptableWorkers) > 0 {
			sort.Slice(acceptableWorkers, func(i, j int) bool {
				return acceptableWorkers[i].ID() < acceptableWorkers[j].ID()
			})

			for _, w := range acceptableWorkers {
				cap := w.Capacity()
				pendingCount := int32(len(queuePending[w.ID()]))
				totalUsed := cap.GetExecutionCount() + cap.GetBacklogCount() + pendingCount
				totalCap := cap.GetExecutionLimit() + cap.GetBacklogLimit()
				if totalUsed >= totalCap {
					continue
				}
				f.addPendingReservation(environmentID, w.ID(), taskID)
				worker = w
				break
			}
		}
	}

	f.pendingReservationsLock.Unlock()

	if worker == nil {
		f.metrics.taskNoCapacityAvailable.Inc()
		return nil, noCapacityAvailableError
	}

	// add task to worker
	f.metrics.taskAssigned.Inc()
	f.taskStatusChannels.Store(taskID, make(chan *workerTaskStatus, GetConfig().TaskStatusBufferSize))

	worker.EnqueueHeartbeatResponse(&pb.HeartbeatResponse{
		TaskId: taskID,
		ExecId: &pb.ExecutionIdentifier{
			Org:     execID.GetOrg(),
			Project: execID.GetProject(),
			Domain:  execID.GetDomain(),
			Name:    execID.GetName(),
		},
		Namespace:     namespace,
		Cmd:           cmd,
		EnvVars:       envVars,
		Operation:     pb.HeartbeatResponse_ASSIGN,
		EnqueueLabels: enqueueLabels,
	})
	worker.SetLastAccessedAt(time.Now().Unix())

	return worker, nil
}

func newFastTaskService(enqueueOwner core.EnqueueOwner, builder interfaces.EnvironmentBuilder, store interfaces.EnvironmentStore, scope promutils.Scope, cfg *Config) *fastTaskServiceImpl {
	svc := &fastTaskServiceImpl{
		enqueueOwner:        enqueueOwner,
		builder:             builder,
		pendingTaskOwners:   make(map[string]map[string]map[string]string),
		pendingReservations:       make(map[string]map[string]map[string]struct{}),
		pendingReservationsByTask: make(map[string]pendingReservationInfo),
		store:               store,
		taskStatusChannels:  &sync.Map{},
		cleanedUpTasks:      cache.New(cleanedUpTasksTTL, cleanedUpTasksGCInterval),
		metrics:             newServiceMetrics(scope, cfg),
	}
	prometheus.MustRegister(svc)
	return svc
}
