package k8s

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	idlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array"
	arrayCore "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/core"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery"

	"github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
)

const executorName = "k8s-array"
const arrayTaskType = "container_array"
const pluginStateVersion = 0

type Executor struct {
	kubeClient       core.KubeClient
	outputsAssembler array.OutputAssembler
	errorAssembler   array.OutputAssembler
}

type KubeClientObj struct {
	client client.Client
}

func (k KubeClientObj) GetClient() client.Client {
	return k.client
}

func (k KubeClientObj) GetCache() cache.Cache {
	return nil
}

func NewKubeClientObj(c client.Client) core.KubeClient {
	return &KubeClientObj{
		client: c,
	}
}

func NewExecutor(kubeClient core.KubeClient, cfg *Config, scope promutils.Scope) (Executor, error) {
	outputAssembler, err := array.NewOutputAssembler(cfg.OutputAssembler, scope.NewSubScope("output_assembler"))
	if err != nil {
		return Executor{}, err
	}

	errorAssembler, err := array.NewErrorAssembler(cfg.MaxErrorStringLength, cfg.ErrorAssembler, scope.NewSubScope("error_assembler"))
	if err != nil {
		return Executor{}, err
	}

	return Executor{
		kubeClient:       kubeClient,
		outputsAssembler: outputAssembler,
		errorAssembler:   errorAssembler,
	}, nil
}

func (e Executor) GetID() string {
	return executorName
}

func (Executor) GetProperties() core.PluginProperties {
	return core.PluginProperties{
		DisableNodeLevelCaching: true,
	}
}

func (e Executor) Handle(ctx context.Context, tCtx core.TaskExecutionContext) (core.Transition, error) {
	pluginConfig := GetConfig()

	pluginState := &arrayCore.State{}
	if _, err := tCtx.PluginStateReader().Get(pluginState); err != nil {
		return core.UnknownTransition, errors.Wrapf(errors.CorruptedPluginState, err, "Failed to read unmarshal custom state")
	}

	var nextState *arrayCore.State
	var err error
	var logLinks []*idlCore.TaskLog
	var subTaskIDs []*string

	switch p, _ := pluginState.GetPhase(); p {
	case arrayCore.PhaseStart:
		nextState, err = array.DetermineDiscoverability(ctx, tCtx, pluginState)

	case arrayCore.PhasePreLaunch:
		nextState = pluginState.SetPhase(arrayCore.PhaseLaunch, core.DefaultPhaseVersion).SetReason("Nothing to do in PreLaunch phase.")
		err = nil

	case arrayCore.PhaseWaitingForResources:
		fallthrough

	case arrayCore.PhaseLaunch:
		// In order to maintain backwards compatibility with the state transitions
		// in the aws batch plugin. Forward to PhaseCheckingSubTasksExecutions where the launching
		// is actually occurring.
		nextState = pluginState.SetPhase(arrayCore.PhaseCheckingSubTaskExecutions, core.DefaultPhaseVersion).SetReason("Nothing to do in Launch phase.")
		err = nil

	case arrayCore.PhaseCheckingSubTaskExecutions:

		nextState, logLinks, subTaskIDs, err = LaunchAndCheckSubTasksState(ctx, tCtx, e.kubeClient, pluginConfig,
			tCtx.DataStore(), tCtx.OutputWriter().GetOutputPrefixPath(), tCtx.OutputWriter().GetRawOutputPrefix(), pluginState)

	case arrayCore.PhaseAssembleFinalOutput:
		nextState, err = array.AssembleFinalOutputs(ctx, e.outputsAssembler, tCtx, arrayCore.PhaseSuccess, pluginState)

	case arrayCore.PhaseWriteToDiscoveryThenFail:
		nextState, err = array.WriteToDiscovery(ctx, tCtx, pluginState, arrayCore.PhaseAssembleFinalError)

	case arrayCore.PhaseWriteToDiscovery:
		nextState, err = array.WriteToDiscovery(ctx, tCtx, pluginState, arrayCore.PhaseAssembleFinalOutput)

	case arrayCore.PhaseAssembleFinalError:
		nextState, err = array.AssembleFinalOutputs(ctx, e.errorAssembler, tCtx, arrayCore.PhasePermanentFailure, pluginState)

	default:
		nextState = pluginState
		err = nil
	}
	if err != nil {
		return core.UnknownTransition, err
	}

	if err := tCtx.PluginStateWriter().Put(pluginStateVersion, nextState); err != nil {
		return core.UnknownTransition, err
	}

	// Determine transition information from the state
	phaseInfo, err := arrayCore.MapArrayStateToPluginPhase(ctx, nextState, logLinks, subTaskIDs)
	if err != nil {
		return core.UnknownTransition, err
	}

	return core.DoTransitionType(core.TransitionTypeBarrier, phaseInfo), nil
}

func (e Executor) Abort(ctx context.Context, tCtx core.TaskExecutionContext) error {
	return nil
}

func (e Executor) Finalize(ctx context.Context, tCtx core.TaskExecutionContext) error {
	pluginConfig := GetConfig()

	pluginState := &arrayCore.State{}
	if _, err := tCtx.PluginStateReader().Get(pluginState); err != nil {
		return errors.Wrapf(errors.CorruptedPluginState, err, "Failed to read unmarshal custom state")
	}

	return TerminateSubTasks(ctx, tCtx, e.kubeClient, pluginConfig, pluginState)
}

func (e Executor) Start(ctx context.Context) error {
	if err := e.outputsAssembler.Start(ctx); err != nil {
		return err
	}

	if err := e.errorAssembler.Start(ctx); err != nil {
		return err
	}

	return nil
}

func init() {
	pluginmachinery.PluginRegistry().RegisterCorePlugin(
		core.PluginEntry{
			ID:                  executorName,
			RegisteredTaskTypes: []core.TaskType{arrayTaskType},
			LoadPlugin:          GetNewExecutorPlugin,
			IsDefault:           false,
		})
}

func GetNewExecutorPlugin(ctx context.Context, iCtx core.SetupContext) (core.Plugin, error) {
	var kubeClient core.KubeClient
	remoteClusterConfig := GetConfig().RemoteClusterConfig
	if remoteClusterConfig.Enabled {
		client, err := GetK8sClient(remoteClusterConfig)
		if err != nil {
			return nil, err
		}
		kubeClient = NewKubeClientObj(client)
	} else {
		kubeClient = iCtx.KubeClient()
	}
	exec, err := NewExecutor(kubeClient, GetConfig(), iCtx.MetricsScope())
	if err != nil {
		return nil, err
	}

	if err = exec.Start(ctx); err != nil {
		return nil, err
	}

	resourceConfig := GetConfig().ResourceConfig
	if IsResourceConfigSet(resourceConfig) {
		primaryLabel := resourceConfig.PrimaryLabel
		limit := resourceConfig.Limit
		if err := iCtx.ResourceRegistrar().RegisterResourceQuota(ctx, core.ResourceNamespace(primaryLabel), limit); err != nil {
			logger.Errorf(ctx, "Token Resource registration for [%v] failed due to error [%v]", primaryLabel, err)
			return nil, err
		}
	}

	return exec, nil
}
