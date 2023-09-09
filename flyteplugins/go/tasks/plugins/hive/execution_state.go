package hive

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/template"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"

	"github.com/flyteorg/flytestdlib/cache"
	"github.com/flyteorg/flytestdlib/contextutils"

	idlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/hive/config"

	"github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/hive/client"
	"github.com/flyteorg/flytestdlib/logger"
)

type ExecutionPhase int

const (
	PhaseNotStarted ExecutionPhase = iota
	PhaseQueued                    // resource manager token gotten
	PhaseSubmitted                 // Sent off to Qubole
	PhaseWriteOutputFile
	PhaseQuerySucceeded
	PhaseQueryFailed
)

func (p ExecutionPhase) String() string {
	switch p {
	case PhaseNotStarted:
		return "PhaseNotStarted"
	case PhaseQueued:
		return "PhaseQueued"
	case PhaseSubmitted:
		return "PhaseSubmitted"
	case PhaseWriteOutputFile:
		return "PhaseWriteOutputFile"
	case PhaseQuerySucceeded:
		return "PhaseQuerySucceeded"
	case PhaseQueryFailed:
		return "PhaseQueryFailed"
	}
	return "Bad Qubole execution phase"
}

type ExecutionState struct {
	Phase ExecutionPhase

	// This will store the command ID from Qubole
	CommandID string `json:"command_id,omitempty"`
	URI       string `json:"uri,omitempty"`

	// This number keeps track of the number of failures within the sync function. Without this, what happens in
	// the sync function is entirely opaque. Note that this field is completely orthogonal to Flyte system/node/task
	// level retries, just errors from hitting the Qubole API, inside the sync loop
	SyncFailureCount int `json:"sync_failure_count,omitempty"`

	// In kicking off the Qubole command, this is the number of failures
	CreationFailureCount int `json:"creation_failure_count,omitempty"`

	// The time the execution first requests for an allocation token
	AllocationTokenRequestStartTime time.Time `json:"allocation_token_request_start_time,omitempty"`
}

// This is the main state iteration
func HandleExecutionState(ctx context.Context, tCtx core.TaskExecutionContext, currentState ExecutionState, quboleClient client.QuboleClient,
	executionsCache cache.AutoRefresh, cfg *config.Config, metrics QuboleHiveExecutorMetrics) (ExecutionState, error) {

	var transformError error
	var newState ExecutionState

	switch currentState.Phase {
	case PhaseNotStarted:
		newState, transformError = GetAllocationToken(ctx, tCtx, currentState, metrics)

	case PhaseQueued:
		newState, transformError = KickOffQuery(ctx, tCtx, currentState, quboleClient, executionsCache, cfg)

	case PhaseSubmitted:
		newState, transformError = MonitorQuery(ctx, tCtx, currentState, executionsCache)

	case PhaseWriteOutputFile:
		newState, transformError = WriteOutputs(ctx, tCtx, currentState)

	case PhaseQuerySucceeded:
		newState = currentState
		transformError = nil

	case PhaseQueryFailed:
		newState = currentState
		transformError = nil
	}

	return newState, transformError
}

func MapExecutionStateToPhaseInfo(state ExecutionState, _ client.QuboleClient) core.PhaseInfo {
	var phaseInfo core.PhaseInfo
	t := time.Now()

	switch state.Phase {
	case PhaseNotStarted:
		phaseInfo = core.PhaseInfoNotReady(t, core.DefaultPhaseVersion, "Haven't received allocation token")
	case PhaseQueued:
		// TODO: Turn into config
		if state.CreationFailureCount > 5 {
			phaseInfo = core.PhaseInfoSystemRetryableFailure("QuboleFailure", "Too many creation attempts", nil)
		} else {
			phaseInfo = core.PhaseInfoQueued(t, uint32(state.CreationFailureCount), "Waiting for Qubole launch")
		}
	case PhaseSubmitted:
		phaseInfo = core.PhaseInfoRunning(core.DefaultPhaseVersion, ConstructTaskInfo(state))

	case PhaseWriteOutputFile:
		phaseInfo = core.PhaseInfoRunning(core.DefaultPhaseVersion+1, ConstructTaskInfo(state))

	case PhaseQuerySucceeded:
		phaseInfo = core.PhaseInfoSuccess(ConstructTaskInfo(state))

	case PhaseQueryFailed:
		phaseInfo = core.PhaseInfoRetryableFailure(errors.DownstreamSystemError, "Query failed", ConstructTaskInfo(state))
	}

	return phaseInfo
}

func ConstructTaskLog(e ExecutionState) *idlCore.TaskLog {
	return &idlCore.TaskLog{
		Name:          fmt.Sprintf("Status: %s [%s]", e.Phase, e.CommandID),
		MessageFormat: idlCore.TaskLog_UNKNOWN,
		Uri:           e.URI,
	}
}

func ConstructTaskInfo(e ExecutionState) *core.TaskInfo {
	logs := make([]*idlCore.TaskLog, 0, 1)
	t := time.Now()

	externalResources := []*core.ExternalResource{
		{
			ExternalID: e.CommandID,
		},
	}

	if e.CommandID != "" {
		logs = append(logs, ConstructTaskLog(e))
		return &core.TaskInfo{
			Logs:              logs,
			OccurredAt:        &t,
			ExternalResources: externalResources,
		}
	}

	return nil
}

func composeResourceNamespaceWithClusterPrimaryLabel(ctx context.Context, tCtx core.TaskExecutionContext) (core.ResourceNamespace, error) {
	_, clusterLabelOverride, _, _, _, err := GetQueryInfo(ctx, tCtx)
	if err != nil {
		return "", err
	}
	clusterPrimaryLabel := getClusterPrimaryLabel(ctx, tCtx, clusterLabelOverride)
	return core.ResourceNamespace(clusterPrimaryLabel), nil
}

func createResourceConstraintsSpec(ctx context.Context, _ core.TaskExecutionContext, targetClusterPrimaryLabel core.ResourceNamespace) core.ResourceConstraintsSpec {
	cfg := config.GetQuboleConfig()
	constraintsSpec := core.ResourceConstraintsSpec{
		ProjectScopeResourceConstraint:   nil,
		NamespaceScopeResourceConstraint: nil,
	}
	if cfg.ClusterConfigs == nil {
		logger.Infof(ctx, "No cluster config is found. Returning an empty resource constraints spec")
		return constraintsSpec
	}
	for _, cluster := range cfg.ClusterConfigs {
		if cluster.PrimaryLabel == string(targetClusterPrimaryLabel) {
			constraintsSpec.ProjectScopeResourceConstraint = &core.ResourceConstraint{Value: int64(float64(cluster.Limit) * cluster.ProjectScopeQuotaProportionCap)}
			constraintsSpec.NamespaceScopeResourceConstraint = &core.ResourceConstraint{Value: int64(float64(cluster.Limit) * cluster.NamespaceScopeQuotaProportionCap)}
			break
		}
	}
	logger.Infof(ctx, "Created a resource constraints spec: [%v]", constraintsSpec)
	return constraintsSpec
}

func GetAllocationToken(ctx context.Context, tCtx core.TaskExecutionContext, currentState ExecutionState, metric QuboleHiveExecutorMetrics) (ExecutionState, error) {
	newState := ExecutionState{}
	uniqueID := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()

	clusterPrimaryLabel, err := composeResourceNamespaceWithClusterPrimaryLabel(ctx, tCtx)
	if err != nil {
		return newState, errors.Wrapf(errors.ResourceManagerFailure, err, "Error getting query info when requesting allocation token %s", uniqueID)
	}

	resourceConstraintsSpec := createResourceConstraintsSpec(ctx, tCtx, clusterPrimaryLabel)

	allocationStatus, err := tCtx.ResourceManager().AllocateResource(ctx, clusterPrimaryLabel, uniqueID, resourceConstraintsSpec)
	if err != nil {
		logger.Errorf(ctx, "Resource manager failed for TaskExecId [%s] token [%s]. error %s",
			tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID(), uniqueID, err)
		return newState, errors.Wrapf(errors.ResourceManagerFailure, err, "Error requesting allocation token %s", uniqueID)
	}
	logger.Infof(ctx, "Allocation result for [%s] is [%s]", uniqueID, allocationStatus)

	// Emitting the duration this execution has been waiting for a token allocation
	if currentState.AllocationTokenRequestStartTime.IsZero() {
		newState.AllocationTokenRequestStartTime = time.Now()
	} else {
		newState.AllocationTokenRequestStartTime = currentState.AllocationTokenRequestStartTime
	}
	waitTime := time.Since(newState.AllocationTokenRequestStartTime)
	metric.ResourceWaitTime.Observe(waitTime.Seconds())

	if allocationStatus == core.AllocationStatusGranted {
		metric.AllocationGranted.Inc(ctx)
		newState.Phase = PhaseQueued
	} else if allocationStatus == core.AllocationStatusExhausted {
		metric.AllocationNotGranted.Inc(ctx)
		newState.Phase = PhaseNotStarted
	} else if allocationStatus == core.AllocationStatusNamespaceQuotaExceeded {
		metric.AllocationNotGranted.Inc(ctx)
		newState.Phase = PhaseNotStarted
	} else {
		return newState, errors.Errorf(errors.ResourceManagerFailure, "Got bad allocation result [%s] for token [%s]",
			allocationStatus, uniqueID)
	}

	return newState, nil
}

func validateQuboleHiveJob(hiveJob plugins.QuboleHiveJob) error {
	if hiveJob.Query == nil {
		return errors.Errorf(errors.BadTaskSpecification,
			"Query could not be found. Please ensure that you are at least on Flytekit version 0.3.0 or later.")
	}
	return nil
}

// This function is the link between the output written by the SDK, and the execution side. It extracts the query
// out of the task template.
func GetQueryInfo(ctx context.Context, tCtx core.TaskExecutionContext) (
	formattedQuery string, cluster string, tags []string, timeoutSec uint32, taskName string, err error) {

	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return "", "", []string{}, 0, "", err
	}

	hiveJob := plugins.QuboleHiveJob{}
	err = utils.UnmarshalStruct(taskTemplate.GetCustom(), &hiveJob)
	if err != nil {
		return "", "", []string{}, 0, "", err
	}

	if err := validateQuboleHiveJob(hiveJob); err != nil {
		return "", "", []string{}, 0, "", err
	}

	query := hiveJob.Query.GetQuery()

	outputs, err := template.Render(ctx, []string{query},
		template.Parameters{
			TaskExecMetadata: tCtx.TaskExecutionMetadata(),
			Inputs:           tCtx.InputReader(),
			OutputPath:       tCtx.OutputWriter(),
			Task:             tCtx.TaskReader(),
		})
	if err != nil {
		return "", "", []string{}, 0, "", err
	}
	formattedQuery = outputs[0]

	cluster = hiveJob.ClusterLabel
	timeoutSec = hiveJob.Query.TimeoutSec
	taskName = taskTemplate.Id.Name
	tags = hiveJob.Tags
	tags = append(tags, fmt.Sprintf("ns:%s", tCtx.TaskExecutionMetadata().GetNamespace()))
	for k, v := range tCtx.TaskExecutionMetadata().GetLabels() {
		tags = append(tags, fmt.Sprintf("%s:%s", k, v))
	}
	logger.Debugf(ctx, "QueryInfo: original query [%s], query: [%s], cluster: [%s], timeoutSec: [%d], tags: [%v]",
		query, formattedQuery, cluster, timeoutSec, tags)

	return formattedQuery, cluster, tags, timeoutSec, taskName, err
}

func mapLabelToPrimaryLabel(ctx context.Context, quboleCfg *config.Config, label string) (primaryLabel string, found bool) {
	primaryLabel = quboleCfg.DefaultClusterLabel
	found = false

	if label == "" {
		logger.Debugf(ctx, "Input cluster label is an empty string; falling back to using the default primary label [%v]", label, primaryLabel)
		return
	}

	// Using a linear search because N is small and because of ClusterConfig's struct definition
	// which is determined specifically for the readability of the corresponding configmap yaml file
	for _, clusterCfg := range quboleCfg.ClusterConfigs {
		for _, l := range clusterCfg.Labels {
			if label != "" && l == label {
				logger.Debugf(ctx, "Found the primary label [%v] for label [%v]", clusterCfg.PrimaryLabel, label)
				primaryLabel, found = clusterCfg.PrimaryLabel, true
				break
			}
		}
	}

	if !found {
		logger.Debugf(ctx, "Cannot find the primary cluster label for label [%v] in configmap; "+
			"falling back to using the default primary label [%v]", label, primaryLabel)
	}

	return primaryLabel, found
}

func mapProjectDomainToDestinationClusterLabel(ctx context.Context, tCtx core.TaskExecutionContext, quboleCfg *config.Config) (string, bool) {
	tExecID := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID()
	project := tExecID.NodeExecutionId.GetExecutionId().GetProject()
	domain := tExecID.NodeExecutionId.GetExecutionId().GetDomain()
	logger.Debugf(ctx, "No clusterLabelOverride. Finding the pre-defined cluster label for (project: %v, domain: %v)", project, domain)
	// Using a linear search because N is small
	for _, m := range quboleCfg.DestinationClusterConfigs {
		if project == m.Project && domain == m.Domain {
			logger.Debugf(ctx, "Found the pre-defined cluster label [%v] for (project: %v, domain: %v)", m.ClusterLabel, project, domain)
			return m.ClusterLabel, true
		}
	}

	// This function finds the label, not primary label, so in the case where no mapping is found, this function should return an empty string
	return "", false
}

func getClusterPrimaryLabel(ctx context.Context, tCtx core.TaskExecutionContext, clusterLabelOverride string) string {
	cfg := config.GetQuboleConfig()

	// If override is not empty and if it has a mapping, we return the mapped primary label
	if clusterLabelOverride != "" {
		if primaryLabel, found := mapLabelToPrimaryLabel(ctx, cfg, clusterLabelOverride); found {
			return primaryLabel
		}
	}

	// If override is empty or if the override does not have a mapping, we return the primary label mapped using (project, domain)
	if clusterLabel, found := mapProjectDomainToDestinationClusterLabel(ctx, tCtx, cfg); found {
		primaryLabel, _ := mapLabelToPrimaryLabel(ctx, cfg, clusterLabel)
		return primaryLabel
	}

	// Else we return the default primary label
	return cfg.DefaultClusterLabel
}

func KickOffQuery(ctx context.Context, tCtx core.TaskExecutionContext, currentState ExecutionState, quboleClient client.QuboleClient,
	cache cache.AutoRefresh, cfg *config.Config) (ExecutionState, error) {

	uniqueID := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()
	apiKey, err := tCtx.SecretManager().Get(ctx, cfg.TokenKey)
	if err != nil {
		return currentState, errors.Wrapf(errors.RuntimeFailure, err, "Failed to read token from secrets manager")
	}

	query, clusterLabelOverride, tags, timeoutSec, taskName, err := GetQueryInfo(ctx, tCtx)
	if err != nil {
		return currentState, err
	}

	clusterPrimaryLabel := getClusterPrimaryLabel(ctx, tCtx, clusterLabelOverride)

	taskExecutionIdentifier := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID()
	commandMetadata := client.CommandMetadata{TaskName: taskName,
		Domain:              taskExecutionIdentifier.GetNodeExecutionId().GetExecutionId().GetDomain(),
		Project:             taskExecutionIdentifier.GetNodeExecutionId().GetExecutionId().GetProject(),
		Labels:              tCtx.TaskExecutionMetadata().GetLabels(),
		AttemptNumber:       taskExecutionIdentifier.GetRetryAttempt(),
		MaxAttempts:         tCtx.TaskExecutionMetadata().GetMaxAttempts(),
		WorkflowExecutionID: taskExecutionIdentifier.GetNodeExecutionId().GetExecutionId().GetName(),
		WorkflowID:          contextutils.Value(ctx, contextutils.WorkflowIDKey),
	}

	cmdDetails, err := quboleClient.ExecuteHiveCommand(ctx, query, timeoutSec,
		clusterPrimaryLabel, apiKey, tags, commandMetadata)
	if err != nil {
		// If we failed, we'll keep the NotStarted state
		currentState.CreationFailureCount = currentState.CreationFailureCount + 1
		logger.Warnf(ctx, "Error creating Qubole query for %s, failure counts %d. Error: %s", uniqueID, currentState.CreationFailureCount, err)
	} else {
		// If we succeed, then store the command id returned from Qubole, and update our state. Also, add to the
		// AutoRefreshCache so we start getting updates.
		commandID := strconv.FormatInt(cmdDetails.ID, 10)
		logger.Infof(ctx, "Created Qubole ID [%s] for token %s", commandID, uniqueID)
		currentState.CommandID = commandID
		currentState.Phase = PhaseSubmitted
		currentState.URI = cmdDetails.URI.String()

		executionStateCacheItem := ExecutionStateCacheItem{
			ExecutionState: currentState,
			Identifier:     uniqueID,
		}

		// The first time we put it in the cache, we know it won't have succeeded so we don't need to look at it
		_, err := cache.GetOrCreate(uniqueID, executionStateCacheItem)
		if err != nil {
			// This means that our cache has fundamentally broken... return a system error
			logger.Errorf(ctx, "Cache failed to GetOrCreate for execution [%s] cache key [%s], owner [%s]. Error %s",
				taskExecutionIdentifier, uniqueID,
				tCtx.TaskExecutionMetadata().GetOwnerReference(), err)
			return currentState, err
		}
	}

	return currentState, nil
}

func MonitorQuery(ctx context.Context, tCtx core.TaskExecutionContext, currentState ExecutionState, cache cache.AutoRefresh) (
	ExecutionState, error) {

	uniqueID := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()
	executionStateCacheItem := ExecutionStateCacheItem{
		ExecutionState: currentState,
		Identifier:     uniqueID,
	}

	cachedItem, err := cache.GetOrCreate(uniqueID, executionStateCacheItem)
	if err != nil {
		// This means that our cache has fundamentally broken... return a system error
		logger.Errorf(ctx, "Cache is broken on execution [%s] cache key [%s], owner [%s]. Error %s",
			tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID(), uniqueID,
			tCtx.TaskExecutionMetadata().GetOwnerReference(), err)
		return currentState, errors.Wrapf(errors.CacheFailed, err, "Error when GetOrCreate while monitoring")
	}

	cachedExecutionState, ok := cachedItem.(ExecutionStateCacheItem)
	if !ok {
		logger.Errorf(ctx, "Error casting cache object into ExecutionState")
		return currentState, errors.Errorf(errors.CacheFailed, "Failed to cast [%v]", cachedItem)
	}

	// TODO: Add a couple of debug lines here - did it change or did it not?

	// If there were updates made to the state, we'll have picked them up automatically. Nothing more to do.
	return cachedExecutionState.ExecutionState, nil
}

func Abort(ctx context.Context, tCtx core.TaskExecutionContext, currentState ExecutionState, qubole client.QuboleClient, apiKey string) error {
	// Cancel Qubole query if non-terminal state
	if !InTerminalState(currentState) && currentState.CommandID != "" {
		err := qubole.KillCommand(ctx, currentState.CommandID, apiKey)
		if err != nil {
			logger.Errorf(ctx, "Error terminating Qubole command in Finalize [%s]", err)
			return err
		}
	}
	return nil
}

func Finalize(ctx context.Context, tCtx core.TaskExecutionContext, _ ExecutionState, metrics QuboleHiveExecutorMetrics) error {
	// Release allocation token
	uniqueID := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()
	clusterPrimaryLabel, err := composeResourceNamespaceWithClusterPrimaryLabel(ctx, tCtx)
	if err != nil {
		return errors.Wrapf(errors.ResourceManagerFailure, err, "Error getting query info when releasing allocation token %s", uniqueID)
	}

	err = tCtx.ResourceManager().ReleaseResource(ctx, clusterPrimaryLabel, uniqueID)

	if err != nil {
		metrics.ResourceReleaseFailed.Inc(ctx)
		logger.Errorf(ctx, "Error releasing allocation token [%s] in Finalize [%s]", uniqueID, err)
		return err
	}
	metrics.ResourceReleased.Inc(ctx)
	return nil
}

func InTerminalState(e ExecutionState) bool {
	return e.Phase == PhaseQuerySucceeded || e.Phase == PhaseQueryFailed
}

func IsNotYetSubmitted(e ExecutionState) bool {
	if e.Phase == PhaseNotStarted || e.Phase == PhaseQueued {
		return true
	}
	return false
}

func WriteOutputs(ctx context.Context, tCtx core.TaskExecutionContext, currentState ExecutionState) (
	ExecutionState, error) {

	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		logger.Errorf(ctx, "Error reading task template: [%s]", err)
		return currentState, err
	}

	externalLocation := tCtx.OutputWriter().GetRawOutputPrefix()
	outputs := taskTemplate.Interface.Outputs.GetVariables()
	if len(outputs) != 0 && len(outputs) != 1 {
		return currentState, errors.Errorf(errors.BadTaskSpecification, "Hive tasks must have zero or one output: [%d] found", len(outputs))
	}
	if len(outputs) == 1 {
		if results, ok := outputs["results"]; ok {
			if results.GetType().GetSchema() == nil {
				return currentState, errors.Errorf(errors.BadTaskSpecification, "A non-SchemaType was found [%v]", results.GetType())
			}
			logger.Debugf(ctx, "Writing outputs file for Hive task at [%s]", tCtx.OutputWriter().GetOutputPrefixPath())
			err = tCtx.OutputWriter().Put(ctx, ioutils.NewInMemoryOutputReader(
				&idlCore.LiteralMap{
					Literals: map[string]*idlCore.Literal{
						"results": {
							Value: &idlCore.Literal_Scalar{
								Scalar: &idlCore.Scalar{Value: &idlCore.Scalar_Schema{
									Schema: &idlCore.Schema{
										Uri:  externalLocation.String(),
										Type: results.GetType().GetSchema(),
									},
								},
								},
							},
						},
					},
				}, nil, nil))
			if err != nil {
				logger.Errorf(ctx, "Error writing outputs file: [%s]", err)
				return currentState, err
			}
		} else {
			logger.Errorf(ctx, "Wrong name for output [%s]", err)
			return currentState, errors.Errorf(errors.BadTaskSpecification, "One output found but wrong name [%s]", outputs)
		}
	}

	logger.Debugf(ctx, "Moving hive task to succeeded")
	currentState.Phase = PhaseQuerySucceeded
	return currentState, nil
}
