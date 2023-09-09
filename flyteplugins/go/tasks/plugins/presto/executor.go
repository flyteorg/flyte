package presto

import (
	"context"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/presto/client"

	"github.com/flyteorg/flytestdlib/cache"

	"github.com/flyteorg/flyteplugins/go/tasks/errors"
	pluginMachinery "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/presto/config"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
)

// This is the name of this plugin effectively. In Flyte plugin configuration, use this string to enable this plugin.
const prestoPluginID = "presto"

// Version of the custom state this plugin stores.  Useful for backwards compatibility if you one day need to update
// the structure of the stored state
const pluginStateVersion = 0

const prestoTaskType = "presto" // This needs to match the type defined in Flytekit constants.py

type Executor struct {
	id              string
	metrics         ExecutorMetrics
	prestoClient    client.PrestoClient
	executionsCache cache.AutoRefresh
	cfg             *config.Config
}

func (p Executor) GetID() string {
	return p.id
}

func (p Executor) Handle(ctx context.Context, tCtx core.TaskExecutionContext) (core.Transition, error) {
	incomingState := ExecutionState{}

	// We assume here that the first time this function is called, the custom state we get back is whatever we passed in,
	// namely the zero-value of our struct.
	if _, err := tCtx.PluginStateReader().Get(&incomingState); err != nil {
		logger.Errorf(ctx, "Plugin %s failed to unmarshal custom state when handling [%s] [%s]",
			p.id, tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), err)
		return core.UnknownTransition, errors.Wrapf(errors.CorruptedPluginState, err,
			"Failed to unmarshal custom state in Handle")
	}

	// Do what needs to be done, and give this function everything it needs to do its job properly
	outgoingState, transformError := HandleExecutionState(ctx, tCtx, incomingState, p.prestoClient, p.executionsCache, p.metrics)

	// Return if there was an error
	if transformError != nil {
		return core.UnknownTransition, transformError
	}

	// If no error, then infer the new Phase from the various states
	phaseInfo := MapExecutionStateToPhaseInfo(outgoingState)

	if err := tCtx.PluginStateWriter().Put(pluginStateVersion, outgoingState); err != nil {
		return core.UnknownTransition, err
	}

	return core.DoTransition(phaseInfo), nil
}

func (p Executor) Abort(ctx context.Context, tCtx core.TaskExecutionContext) error {
	incomingState := ExecutionState{}
	if _, err := tCtx.PluginStateReader().Get(&incomingState); err != nil {
		logger.Errorf(ctx, "Plugin %s failed to unmarshal custom state in Finalize [%s] Err [%s]",
			p.id, tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), err)
		return errors.Wrapf(errors.CorruptedPluginState, err, "Failed to unmarshal custom state in Finalize")
	}

	return Abort(ctx, incomingState, p.prestoClient)
}

func (p Executor) Finalize(ctx context.Context, tCtx core.TaskExecutionContext) error {
	incomingState := ExecutionState{}
	if _, err := tCtx.PluginStateReader().Get(&incomingState); err != nil {
		logger.Errorf(ctx, "Plugin %s failed to unmarshal custom state in Finalize [%s] Err [%s]",
			p.id, tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), err)
		return errors.Wrapf(errors.CorruptedPluginState, err, "Failed to unmarshal custom state in Finalize")
	}

	return Finalize(ctx, tCtx, incomingState, p.metrics)
}

func (p Executor) GetProperties() core.PluginProperties {
	return core.PluginProperties{}
}

func ExecutorLoader(ctx context.Context, iCtx core.SetupContext) (core.Plugin, error) {
	cfg := config.GetPrestoConfig()
	return InitializePrestoExecutor(ctx, iCtx, cfg, client.NewNoopPrestoClient(cfg))
}

func InitializePrestoExecutor(
	ctx context.Context,
	iCtx core.SetupContext,
	cfg *config.Config,
	prestoClient client.PrestoClient) (core.Plugin, error) {
	logger.Infof(ctx, "Initializing a Presto executo")
	q, err := NewPrestoExecutor(ctx, cfg, prestoClient, iCtx.MetricsScope())
	if err != nil {
		logger.Errorf(ctx, "Failed to create a new Executor due to error: [%v]", err)
		return nil, err
	}

	for _, routingGroup := range cfg.RoutingGroupConfigs {
		logger.Infof(ctx, "Registering resource quota for routing group [%v]", routingGroup.Name)
		if err := iCtx.ResourceRegistrar().RegisterResourceQuota(ctx, core.ResourceNamespace(routingGroup.Name), routingGroup.Limit); err != nil {
			logger.Errorf(ctx, "Resource quota registration for [%v] failed due to error [%v]", routingGroup.Name, err)
			return nil, err
		}
	}

	return q, nil
}

func NewPrestoExecutor(
	ctx context.Context,
	cfg *config.Config,
	prestoClient client.PrestoClient,
	scope promutils.Scope) (Executor, error) {
	subScope := scope.NewSubScope(prestoTaskType)
	executionsAutoRefreshCache, err := NewPrestoExecutionsCache(ctx, prestoClient, cfg, subScope)
	if err != nil {
		logger.Errorf(ctx, "Failed to create AutoRefreshCache in Executor Setup. Error: %v", err)
		return Executor{}, err
	}

	err = executionsAutoRefreshCache.Start(ctx)
	if err != nil {
		logger.Errorf(ctx, "Failed to start AutoRefreshCache. Error: %v", err)
	}

	return Executor{
		id:              prestoPluginID,
		cfg:             cfg,
		metrics:         getPrestoExecutorMetrics(subScope),
		prestoClient:    prestoClient,
		executionsCache: executionsAutoRefreshCache,
	}, nil
}

func init() {
	pluginMachinery.PluginRegistry().RegisterCorePlugin(
		core.PluginEntry{
			ID:                  prestoPluginID,
			RegisteredTaskTypes: []core.TaskType{prestoTaskType},
			LoadPlugin:          ExecutorLoader,
			IsDefault:           false,
		})
}
