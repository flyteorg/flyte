package hive

import (
	"context"

	"github.com/flyteorg/flytestdlib/cache"

	"github.com/flyteorg/flyteplugins/go/tasks/errors"
	pluginMachinery "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/hive/client"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/hive/config"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
)

// This is the name of this plugin effectively. In Flyte plugin configuration, use this string to enable this plugin.
const quboleHiveExecutorID = "qubole-hive-executor"

// Version of the custom state this plugin stores.  Useful for backwards compatibility if you one day need to update
// the structure of the stored state
const pluginStateVersion = 0

const hiveTaskType = "hive" // This needs to match the type defined in Flytekit constants.py

const DefaultClusterPrimaryLabel = "default"

type QuboleHiveExecutor struct {
	id              string
	metrics         QuboleHiveExecutorMetrics
	quboleClient    client.QuboleClient
	executionsCache cache.AutoRefresh
	cfg             *config.Config
}

func (q QuboleHiveExecutor) GetID() string {
	return q.id
}

func (q QuboleHiveExecutor) Handle(ctx context.Context, tCtx core.TaskExecutionContext) (core.Transition, error) {
	incomingState := ExecutionState{}

	// We assume here that the first time this function is called, the custom state we get back is whatever we passed in,
	// namely the zero-value of our struct.
	if _, err := tCtx.PluginStateReader().Get(&incomingState); err != nil {
		logger.Errorf(ctx, "Plugin %s failed to unmarshal custom state when handling [%s] [%s]",
			q.id, tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), err)
		return core.UnknownTransition, errors.Wrapf(errors.CorruptedPluginState, err,
			"Failed to unmarshal custom state in Handle")
	}

	// Do what needs to be done, and give this function everything it needs to do its job properly
	// TODO: Play around with making this return a transition directly. How will that pattern affect the multi-Qubole plugin
	outgoingState, transformError := HandleExecutionState(ctx, tCtx, incomingState, q.quboleClient, q.executionsCache, q.cfg, q.metrics)

	// Return if there was an error
	if transformError != nil {
		return core.UnknownTransition, transformError
	}

	// If no error, then infer the new Phase from the various states
	phaseInfo := MapExecutionStateToPhaseInfo(outgoingState, q.quboleClient)

	if err := tCtx.PluginStateWriter().Put(pluginStateVersion, outgoingState); err != nil {
		return core.UnknownTransition, err
	}

	return core.DoTransition(phaseInfo), nil
}

func (q QuboleHiveExecutor) Abort(ctx context.Context, tCtx core.TaskExecutionContext) error {
	incomingState := ExecutionState{}
	if _, err := tCtx.PluginStateReader().Get(&incomingState); err != nil {
		logger.Errorf(ctx, "Plugin %s failed to unmarshal custom state in Finalize [%s] Err [%s]",
			q.id, tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), err)
		return errors.Wrapf(errors.CorruptedPluginState, err, "Failed to unmarshal custom state in Finalize")
	}

	key, err := tCtx.SecretManager().Get(ctx, q.cfg.TokenKey)
	if err != nil {
		logger.Errorf(ctx, "Error reading token in Finalize [%s]", err)
		return err
	}

	return Abort(ctx, tCtx, incomingState, q.quboleClient, key)
}

func (q QuboleHiveExecutor) Finalize(ctx context.Context, tCtx core.TaskExecutionContext) error {
	incomingState := ExecutionState{}
	if _, err := tCtx.PluginStateReader().Get(&incomingState); err != nil {
		logger.Errorf(ctx, "Plugin %s failed to unmarshal custom state in Finalize [%s] Err [%s]",
			q.id, tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), err)
		return errors.Wrapf(errors.CorruptedPluginState, err, "Failed to unmarshal custom state in Finalize")
	}

	return Finalize(ctx, tCtx, incomingState, q.metrics)
}

func (q QuboleHiveExecutor) GetProperties() core.PluginProperties {
	return core.PluginProperties{}
}

func QuboleHiveExecutorLoader(ctx context.Context, iCtx core.SetupContext) (core.Plugin, error) {
	cfg := config.GetQuboleConfig()
	return InitializeHiveExecutor(ctx, iCtx, cfg, BuildResourceConfig(cfg.ClusterConfigs), client.NewQuboleClient(cfg))
}

func BuildResourceConfig(cfg []config.ClusterConfig) map[core.ResourceNamespace]int {
	resourceConfig := make(map[core.ResourceNamespace]int, len(cfg))

	for _, clusterCfg := range cfg {
		resourceConfig[core.ResourceNamespace(clusterCfg.PrimaryLabel)] = clusterCfg.Limit
	}
	return resourceConfig
}

func InitializeHiveExecutor(ctx context.Context, iCtx core.SetupContext, cfg *config.Config, resourceConfig map[core.ResourceNamespace]int,
	quboleClient client.QuboleClient) (core.Plugin, error) {
	logger.Infof(ctx, "Initializing a Hive executor with a resource config [%v]", resourceConfig)
	q, err := NewQuboleHiveExecutor(ctx, cfg, quboleClient, iCtx.SecretManager(), iCtx.MetricsScope())
	if err != nil {
		logger.Errorf(ctx, "Failed to create a new QuboleHiveExecutor due to error: [%v]", err)
		return nil, err
	}

	for clusterPrimaryLabel, clusterLimit := range resourceConfig {
		logger.Infof(ctx, "Registering resource quota ([%v]) and namespace quota cap ([%v]) for cluster [%v]", clusterPrimaryLabel)
		if err := iCtx.ResourceRegistrar().RegisterResourceQuota(ctx, clusterPrimaryLabel, clusterLimit); err != nil {
			logger.Errorf(ctx, "Resource quota registration for [%v] failed due to error [%v]", clusterPrimaryLabel, err)
			return nil, err
		}
	}

	return q, nil
}

// type PluginLoader func(ctx context.Context, iCtx SetupContext) (Plugin, error)
func NewQuboleHiveExecutor(ctx context.Context, cfg *config.Config, quboleClient client.QuboleClient, secretManager core.SecretManager, scope promutils.Scope) (QuboleHiveExecutor, error) {
	executionsAutoRefreshCache, err := NewQuboleHiveExecutionsCache(ctx, quboleClient, secretManager, cfg, scope.NewSubScope(hiveTaskType))
	if err != nil {
		logger.Errorf(ctx, "Failed to create AutoRefreshCache in QuboleHiveExecutor Setup. Error: %v", err)
		return QuboleHiveExecutor{}, err
	}

	err = executionsAutoRefreshCache.Start(ctx)
	if err != nil {
		logger.Errorf(ctx, "Failed to start AutoRefreshCache. Error: %v", err)
	}

	return QuboleHiveExecutor{
		id:              quboleHiveExecutorID,
		cfg:             cfg,
		metrics:         getQuboleHiveExecutorMetrics(scope.NewSubScope("hive")),
		quboleClient:    quboleClient,
		executionsCache: executionsAutoRefreshCache,
	}, nil
}

func init() {
	pluginMachinery.PluginRegistry().RegisterCorePlugin(
		core.PluginEntry{
			ID:                  quboleHiveExecutorID,
			RegisteredTaskTypes: []core.TaskType{hiveTaskType},
			LoadPlugin:          QuboleHiveExecutorLoader,
			IsDefault:           false,
		})
}
