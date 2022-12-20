package presto

import (
	"context"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/template"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"

	"k8s.io/apimachinery/pkg/util/rand"

	"fmt"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/presto/client"

	"time"

	"github.com/flyteorg/flytestdlib/cache"

	idlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/presto/config"

	"github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flytestdlib/logger"

	pb "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

type ExecutionPhase int

const (
	PhaseNotStarted ExecutionPhase = iota
	PhaseQueued                    // resource manager token gotten
	PhaseSubmitted                 // Sent off to Presto
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
	case PhaseQuerySucceeded:
		return "PhaseQuerySucceeded"
	case PhaseQueryFailed:
		return "PhaseQueryFailed"
	}
	return "Bad Presto execution phase"
}

type ExecutionState struct {
	CurrentPhase  ExecutionPhase
	PreviousPhase ExecutionPhase

	// This will store the command ID from Presto
	CommandID string `json:"commandId,omitempty"`

	// This will have the nextUri from Presto which is used to advance the query forward
	URI string `json:"uri,omitempty"`

	// This is the current Presto query (out of 5) needed to complete a Presto task
	CurrentPrestoQuery Query `json:"currentPrestoQuery,omitempty"`

	// This is an id to keep track of the current query. Every query's id should be unique for caching purposes
	CurrentPrestoQueryUUID string `json:"currentPrestoQueryUUID,omitempty"`

	// Keeps track of which Presto query we are on. Its values range from 0-4 for the 5 queries that are needed
	QueryCount int `json:"queryCount,omitempty"`

	// This number keeps track of the number of failures within the sync function. Without this, what happens in
	// the sync function is entirely opaque. Note that this field is completely orthogonal to Flyte system/node/task
	// level retries, just errors from hitting the Presto API, inside the sync loop
	SyncFailureCount int `json:"syncFailureCount,omitempty"`

	// In kicking off the Presto command, this is the number of failures
	CreationFailureCount int `json:"creationFailureCount,omitempty"`

	// The time the execution first requests for an allocation token
	AllocationTokenRequestStartTime time.Time `json:"allocationTokenRequestStartTime,omitempty"`
}

type Query struct {
	Statement         string                   `json:"statement,omitempty"`
	ExecuteArgs       client.PrestoExecuteArgs `json:"executeArgs,omitempty"`
	TempTableName     string                   `json:"tempTableName,omitempty"`
	ExternalTableName string                   `json:"externalTableName,omitempty"`
	ExternalLocation  string                   `json:"externalLocation"`
}

const PrestoSource = "flyte"

// This is the main state iteration
func HandleExecutionState(
	ctx context.Context,
	tCtx core.TaskExecutionContext,
	currentState ExecutionState,
	prestoClient client.PrestoClient,
	executionsCache cache.AutoRefresh,
	metrics ExecutorMetrics) (ExecutionState, error) {

	var transformError error
	var newState ExecutionState

	switch currentState.CurrentPhase {
	case PhaseNotStarted:
		newState, transformError = GetAllocationToken(ctx, tCtx, currentState, metrics)

	case PhaseQueued:
		prestoQuery, err := GetNextQuery(ctx, tCtx, currentState)
		if err != nil {
			return ExecutionState{}, err
		}
		currentState.CurrentPrestoQuery = prestoQuery
		newState, transformError = KickOffQuery(ctx, tCtx, currentState, prestoClient, executionsCache)

	case PhaseSubmitted:
		newState, transformError = MonitorQuery(ctx, tCtx, currentState, executionsCache)

	case PhaseQuerySucceeded:
		if currentState.QueryCount < 1 {
			// If there are still Presto statements to execute, increment the query count, reset the phase to 'queued'
			// and continue executing the remaining statements. In this case, we won't request another allocation token
			// as the 2 statements that get executed are all considered to be part of the same "query"
			currentState.PreviousPhase = currentState.CurrentPhase
			currentState.CurrentPhase = PhaseQueued
		} else {
			//currentState.Phase = PhaseQuerySucceeded
			currentState.PreviousPhase = currentState.CurrentPhase
			transformError = writeOutput(ctx, tCtx, currentState.CurrentPrestoQuery.ExternalLocation)
		}
		currentState.QueryCount++
		newState = currentState

	case PhaseQueryFailed:
		newState = currentState
		transformError = nil
	}

	return newState, transformError
}

func GetAllocationToken(
	ctx context.Context,
	tCtx core.TaskExecutionContext,
	currentState ExecutionState,
	metric ExecutorMetrics) (ExecutionState, error) {

	newState := ExecutionState{}
	uniqueID := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()

	routingGroup, err := composeResourceNamespaceWithRoutingGroup(ctx, tCtx)
	if err != nil {
		return newState, errors.Wrapf(errors.ResourceManagerFailure, err, "Error getting query info when requesting allocation token %s", uniqueID)
	}

	resourceConstraintsSpec := createResourceConstraintsSpec(ctx, tCtx, routingGroup)

	allocationStatus, err := tCtx.ResourceManager().AllocateResource(ctx, routingGroup, uniqueID, resourceConstraintsSpec)
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

	if allocationStatus == core.AllocationStatusGranted {
		metric.AllocationGranted.Inc(ctx)
		newState.CurrentPhase = PhaseQueued
	} else if allocationStatus == core.AllocationStatusExhausted {
		metric.AllocationNotGranted.Inc(ctx)
		newState.CurrentPhase = PhaseNotStarted
	} else if allocationStatus == core.AllocationStatusNamespaceQuotaExceeded {
		metric.AllocationNotGranted.Inc(ctx)
		newState.CurrentPhase = PhaseNotStarted
	} else {
		return newState, errors.Errorf(errors.ResourceManagerFailure, "Got bad allocation result [%s] for token [%s]",
			allocationStatus, uniqueID)
	}

	return newState, nil
}

func composeResourceNamespaceWithRoutingGroup(ctx context.Context, tCtx core.TaskExecutionContext) (core.ResourceNamespace, error) {
	routingGroup, _, _, _, err := GetQueryInfo(ctx, tCtx)
	if err != nil {
		return "", err
	}
	clusterPrimaryLabel := resolveRoutingGroup(ctx, routingGroup, config.GetPrestoConfig())
	return core.ResourceNamespace(clusterPrimaryLabel), nil
}

// This function is the link between the output written by the SDK, and the execution side. It extracts the query
// out of the task template.
func GetQueryInfo(ctx context.Context, tCtx core.TaskExecutionContext) (string, string, string, string, error) {
	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return "", "", "", "", err
	}

	prestoQuery := plugins.PrestoQuery{}
	if err := utils.UnmarshalStruct(taskTemplate.GetCustom(), &prestoQuery); err != nil {
		return "", "", "", "", err
	}

	if err := validatePrestoStatement(prestoQuery); err != nil {
		return "", "", "", "", err
	}

	outputs, err := template.Render(ctx, []string{
		prestoQuery.RoutingGroup,
		prestoQuery.Catalog,
		prestoQuery.Schema,
		prestoQuery.Statement,
	}, template.Parameters{
		TaskExecMetadata: tCtx.TaskExecutionMetadata(),
		Inputs:           tCtx.InputReader(),
		OutputPath:       tCtx.OutputWriter(),
		Task:             tCtx.TaskReader(),
	})
	if err != nil {
		return "", "", "", "", err
	}

	routingGroup := outputs[0]
	catalog := outputs[1]
	schema := outputs[2]
	statement := outputs[3]

	logger.Debugf(ctx, "QueryInfo: query: [%v], routingGroup: [%v], catalog: [%v], schema: [%v]", statement, routingGroup, catalog, schema)
	return routingGroup, catalog, schema, statement, err
}

func validatePrestoStatement(prestoJob plugins.PrestoQuery) error {
	if prestoJob.Statement == "" {
		return errors.Errorf(errors.BadTaskSpecification,
			"Query could not be found. Please ensure that you are at least on Flytekit version 0.3.0 or later.")
	}
	return nil
}

func resolveRoutingGroup(ctx context.Context, routingGroup string, prestoCfg *config.Config) string {
	if routingGroup == "" {
		logger.Debugf(ctx, "Input routing group is an empty string; falling back to using the default routing group [%v]", prestoCfg.DefaultRoutingGroup)
		return prestoCfg.DefaultRoutingGroup
	}

	for _, routingGroupCfg := range prestoCfg.RoutingGroupConfigs {
		if routingGroup == routingGroupCfg.Name {
			logger.Debugf(ctx, "Found the Presto routing group: [%v]", routingGroupCfg.Name)
			return routingGroup
		}
	}

	logger.Debugf(ctx, "Cannot find the routing group [%v] in configmap; "+
		"falling back to using the default routing group [%v]", routingGroup, prestoCfg.DefaultRoutingGroup)
	return prestoCfg.DefaultRoutingGroup
}

func createResourceConstraintsSpec(ctx context.Context, _ core.TaskExecutionContext, routingGroup core.ResourceNamespace) core.ResourceConstraintsSpec {
	cfg := config.GetPrestoConfig()
	constraintsSpec := core.ResourceConstraintsSpec{
		ProjectScopeResourceConstraint:   nil,
		NamespaceScopeResourceConstraint: nil,
	}
	if cfg.RoutingGroupConfigs == nil {
		logger.Infof(ctx, "No routing group config is found. Returning an empty resource constraints spec")
		return constraintsSpec
	}
	for _, routingGroupCfg := range cfg.RoutingGroupConfigs {
		if routingGroupCfg.Name == string(routingGroup) {
			constraintsSpec.ProjectScopeResourceConstraint = &core.ResourceConstraint{Value: int64(float64(routingGroupCfg.Limit) * routingGroupCfg.ProjectScopeQuotaProportionCap)}
			constraintsSpec.NamespaceScopeResourceConstraint = &core.ResourceConstraint{Value: int64(float64(routingGroupCfg.Limit) * routingGroupCfg.NamespaceScopeQuotaProportionCap)}
			break
		}
	}
	logger.Infof(ctx, "Created a resource constraints spec: [%v]", constraintsSpec)
	return constraintsSpec
}

func GetNextQuery(
	ctx context.Context,
	tCtx core.TaskExecutionContext,
	currentState ExecutionState) (Query, error) {

	switch currentState.QueryCount {
	case 0:
		prestoCfg := config.GetPrestoConfig()
		tempTableName := rand.String(32)
		routingGroup, catalog, schema, statement, err := GetQueryInfo(ctx, tCtx)
		if err != nil {
			return Query{}, err
		}
		var user = getUser(ctx, prestoCfg.DefaultUser)

		if prestoCfg.UseNamespaceAsUser {
			user = tCtx.TaskExecutionMetadata().GetNamespace()
		}

		externalLocation, err := tCtx.DataStore().ConstructReference(ctx, tCtx.OutputWriter().GetRawOutputPrefix(), "")
		if err != nil {
			return Query{}, err
		}

		queryWrapTemplate := `
			CREATE TABLE hive.flyte_temporary_tables."%s_temp"
			WITH (format = 'PARQUET', external_location = '%s')
			AS (%s)
		`

		statement = fmt.Sprintf(queryWrapTemplate, tempTableName, externalLocation, statement)

		prestoQuery := Query{
			Statement: statement,
			ExecuteArgs: client.PrestoExecuteArgs{
				RoutingGroup: resolveRoutingGroup(ctx, routingGroup, prestoCfg),
				Catalog:      catalog,
				Schema:       schema,
				Source:       PrestoSource,
				User:         user,
			},
			TempTableName:     tempTableName + "_temp",
			ExternalTableName: tempTableName + "_external",
			ExternalLocation:  externalLocation.String(),
		}

		return prestoQuery, nil

	case 1:
		statement := fmt.Sprintf(`DROP TABLE hive.flyte_temporary_tables."%s"`, currentState.CurrentPrestoQuery.TempTableName)
		currentState.CurrentPrestoQuery.Statement = statement
		return currentState.CurrentPrestoQuery, nil

	default:
		return currentState.CurrentPrestoQuery, nil
	}
}

func getUser(ctx context.Context, defaultUser string) string {
	principalContextUser := ctx.Value("principal")
	if principalContextUser != nil {
		return fmt.Sprintf("%v", principalContextUser)
	}
	return defaultUser
}

func KickOffQuery(
	ctx context.Context,
	tCtx core.TaskExecutionContext,
	currentState ExecutionState,
	prestoClient client.PrestoClient,
	cache cache.AutoRefresh) (ExecutionState, error) {

	// For the caching id, we can't rely simply on the task execution id since we have to run 5 consecutive queries and
	// the ids used for each of these has to be unique. Because of this, we append a random postfix to the task
	// execution id.
	uniqueID := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName() + "_" + rand.String(32)

	statement := currentState.CurrentPrestoQuery.Statement
	executeArgs := currentState.CurrentPrestoQuery.ExecuteArgs

	response, err := prestoClient.ExecuteCommand(ctx, statement, executeArgs)
	if err != nil {
		// If we failed, we'll keep the NotStarted state
		currentState.CreationFailureCount = currentState.CreationFailureCount + 1
		logger.Warnf(ctx, "Error creating Presto query for %s, failure counts %d. Error: %s", uniqueID, currentState.CreationFailureCount, err)
	} else {
		// If we succeed, then store the command id returned from Presto, and update our state. Also, add to the
		// AutoRefreshCache so we start getting updates for its status.
		commandID := response.ID
		logger.Infof(ctx, "Created Presto ID [%s] for token %s", commandID, uniqueID)
		currentState.CommandID = commandID
		currentState.PreviousPhase = currentState.CurrentPhase
		currentState.CurrentPhase = PhaseSubmitted
		currentState.URI = response.NextURI
		currentState.CurrentPrestoQueryUUID = uniqueID

		executionStateCacheItem := ExecutionStateCacheItem{
			ExecutionState: currentState,
			Identifier:     uniqueID,
		}

		// The first time we put it in the cache, we know it won't have succeeded so we don't need to look at it
		_, err := cache.GetOrCreate(uniqueID, executionStateCacheItem)
		if err != nil {
			// This means that our cache has fundamentally broken... return a system error
			logger.Errorf(ctx, "Cache failed to GetOrCreate for execution [%s] cache key [%s], owner [%s]. Error %s",
				tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID(), uniqueID,
				tCtx.TaskExecutionMetadata().GetOwnerReference(), err)
			return currentState, err
		}
	}

	return currentState, nil
}

func MonitorQuery(
	ctx context.Context,
	tCtx core.TaskExecutionContext,
	currentState ExecutionState,
	cache cache.AutoRefresh) (ExecutionState, error) {

	uniqueQueryID := currentState.CurrentPrestoQueryUUID
	executionStateCacheItem := ExecutionStateCacheItem{
		ExecutionState: currentState,
		Identifier:     uniqueQueryID,
	}

	cachedItem, err := cache.GetOrCreate(uniqueQueryID, executionStateCacheItem)
	if err != nil {
		// This means that our cache has fundamentally broken... return a system error
		logger.Errorf(ctx, "Cache is broken on execution [%s] cache key [%s], owner [%s]. Error %s",
			tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID(), uniqueQueryID,
			tCtx.TaskExecutionMetadata().GetOwnerReference(), err)
		return currentState, errors.Wrapf(errors.CacheFailed, err, "Error when GetOrCreate while monitoring")
	}

	cachedExecutionState, ok := cachedItem.(ExecutionStateCacheItem)
	if !ok {
		logger.Errorf(ctx, "Error casting cache object into ExecutionState")
		return currentState, errors.Errorf(errors.CacheFailed, "Failed to cast [%v]", cachedItem)
	}

	// If there were updates made to the state, we'll have picked them up automatically. Nothing more to do.
	return cachedExecutionState.ExecutionState, nil
}

func writeOutput(ctx context.Context, tCtx core.TaskExecutionContext, externalLocation string) error {
	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return err
	}

	results := taskTemplate.Interface.Outputs.Variables["results"]

	return tCtx.OutputWriter().Put(ctx, ioutils.NewInMemoryOutputReader(
		&pb.LiteralMap{
			Literals: map[string]*pb.Literal{
				"results": {
					Value: &pb.Literal_Scalar{
						Scalar: &pb.Scalar{Value: &pb.Scalar_Schema{
							Schema: &pb.Schema{
								Uri:  externalLocation,
								Type: results.GetType().GetSchema(),
							},
						},
						},
					},
				},
			},
		}, nil, nil))
}

// The 'PhaseInfoRunning' occurs 15 times (3 for each of the 5 Presto queries that get run for every Presto task) which
// are differentiated by the version (1-15)
func MapExecutionStateToPhaseInfo(state ExecutionState) core.PhaseInfo {
	var phaseInfo core.PhaseInfo
	t := time.Now()

	//switch state.Phase {
	switch state.CurrentPhase {
	case PhaseNotStarted:
		phaseInfo = core.PhaseInfoNotReady(t, core.DefaultPhaseVersion, "Haven't received allocation token")
	case PhaseQueued:
		if state.CreationFailureCount > 5 {
			phaseInfo = core.PhaseInfoRetryableFailure("PrestoFailure", "Too many creation attempts", nil)
		} else {
			phaseInfo = core.PhaseInfoRunning(uint32(3*state.QueryCount+1), ConstructTaskInfo(state))
		}
	case PhaseSubmitted:
		phaseInfo = core.PhaseInfoRunning(uint32(3*state.QueryCount+2), ConstructTaskInfo(state))
	case PhaseQuerySucceeded:
		if state.QueryCount < 5 {
			phaseInfo = core.PhaseInfoRunning(uint32(3*state.QueryCount+3), ConstructTaskInfo(state))
		} else {
			phaseInfo = core.PhaseInfoSuccess(ConstructTaskInfo(state))
		}
	case PhaseQueryFailed:
		phaseInfo = core.PhaseInfoRetryableFailure(errors.DownstreamSystemError, "Query failed", ConstructTaskInfo(state))
	}

	return phaseInfo
}

func ConstructTaskInfo(e ExecutionState) *core.TaskInfo {
	logs := make([]*idlCore.TaskLog, 0, 1)
	t := time.Now()
	if e.CommandID != "" {
		logs = append(logs, ConstructTaskLog(e))
		return &core.TaskInfo{
			Logs:       logs,
			OccurredAt: &t,
			ExternalResources: []*core.ExternalResource{
				{
					ExternalID: e.CommandID,
				},
			},
		}
	}

	return nil
}

func ConstructTaskLog(e ExecutionState) *idlCore.TaskLog {
	return &idlCore.TaskLog{
		Name:          fmt.Sprintf("Status: %s [%s]", e.PreviousPhase, e.CommandID),
		MessageFormat: idlCore.TaskLog_UNKNOWN,
		Uri:           e.URI,
	}
}

func Abort(ctx context.Context, currentState ExecutionState, client client.PrestoClient) error {
	// Cancel Presto query if non-terminal state
	if !InTerminalState(currentState) && currentState.CommandID != "" {
		err := client.KillCommand(ctx, currentState.CommandID)
		if err != nil {
			logger.Errorf(ctx, "Error terminating Presto command in Finalize [%s]", err)
			return err
		}
	}
	return nil
}

func Finalize(ctx context.Context, tCtx core.TaskExecutionContext, _ ExecutionState, metrics ExecutorMetrics) error {
	// Release allocation token
	uniqueID := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()
	routingGroup, err := composeResourceNamespaceWithRoutingGroup(ctx, tCtx)
	if err != nil {
		return errors.Wrapf(errors.ResourceManagerFailure, err, "Error getting query info when releasing allocation token %s", uniqueID)
	}

	err = tCtx.ResourceManager().ReleaseResource(ctx, routingGroup, uniqueID)

	if err != nil {
		metrics.ResourceReleaseFailed.Inc(ctx)
		logger.Errorf(ctx, "Error releasing allocation token [%s] in Finalize [%s]", uniqueID, err)
		return err
	}
	metrics.ResourceReleased.Inc(ctx)
	return nil
}

func InTerminalState(e ExecutionState) bool {
	return e.CurrentPhase == PhaseQuerySucceeded || e.CurrentPhase == PhaseQueryFailed
}

func IsNotYetSubmitted(e ExecutionState) bool {
	if e.CurrentPhase == PhaseNotStarted || e.CurrentPhase == PhaseQueued {
		return true
	}
	return false
}
