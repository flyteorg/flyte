package k8s

import (
	"context"

	"github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array"
	arrayCore "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/core"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"

	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	var externalResources []*core.ExternalResource

	switch p, version := pluginState.GetPhase(); p {
	case arrayCore.PhaseStart:
		nextState, err = array.DetermineDiscoverability(ctx, tCtx, pluginConfig.MaxArrayJobSize, pluginState)

		nextPhase, _ := nextState.GetPhase()
		if err == nil && nextPhase != arrayCore.PhaseStart && nextPhase != arrayCore.PhasePermanentFailure {
			// we wait to transition out of PhaseStart to InitializeExternalResources because then the array
			// job configuration has then been validated and all of the metadata necessary to report subtask
			// status (ie. cache hit / etc) is available.
			externalResources, err = arrayCore.InitializeExternalResources(ctx, tCtx, nextState,
				func(tCtx core.TaskExecutionContext, childIndex int) string {
					subTaskExecutionID := NewSubTaskExecutionID(tCtx.TaskExecutionMetadata().GetTaskExecutionID(), childIndex, 0)
					return subTaskExecutionID.GetGeneratedName()
				},
			)
		}

	case arrayCore.PhasePreLaunch:
		nextState = pluginState.SetPhase(arrayCore.PhaseLaunch, version+1).SetReason("Nothing to do in PreLaunch phase.")

	case arrayCore.PhaseWaitingForResources:
		fallthrough

	case arrayCore.PhaseLaunch:
		// In order to maintain backwards compatibility with the state transitions
		// in the aws batch plugin. Forward to PhaseCheckingSubTasksExecutions where the launching
		// is actually occurring.
		nextState = pluginState.SetPhase(arrayCore.PhaseCheckingSubTaskExecutions, version+1).SetReason("Nothing to do in Launch phase.")
		err = nil

	case arrayCore.PhaseCheckingSubTaskExecutions:
		nextState, externalResources, err = LaunchAndCheckSubTasksState(ctx, tCtx, e.kubeClient, pluginConfig,
			tCtx.DataStore(), tCtx.OutputWriter().GetOutputPrefixPath(), tCtx.OutputWriter().GetRawOutputPrefix(), pluginState)

	case arrayCore.PhaseAssembleFinalOutput:
		nextState, err = array.AssembleFinalOutputs(ctx, e.outputsAssembler, tCtx, arrayCore.PhaseSuccess, version+1, pluginState)

	case arrayCore.PhaseWriteToDiscoveryThenFail:
		nextState, externalResources, err = array.WriteToDiscovery(ctx, tCtx, pluginState, arrayCore.PhaseAssembleFinalError, version+1)

	case arrayCore.PhaseWriteToDiscovery:
		nextState, externalResources, err = array.WriteToDiscovery(ctx, tCtx, pluginState, arrayCore.PhaseAssembleFinalOutput, version+1)

	case arrayCore.PhaseAssembleFinalError:
		nextState, err = array.AssembleFinalOutputs(ctx, e.errorAssembler, tCtx, arrayCore.PhasePermanentFailure, version+1, pluginState)

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
	phaseInfo, err := arrayCore.MapArrayStateToPluginPhase(ctx, nextState, nil, externalResources)
	if err != nil {
		return core.UnknownTransition, err
	}

	return core.DoTransition(phaseInfo), nil
}

func (e Executor) Abort(ctx context.Context, tCtx core.TaskExecutionContext) error {
	pluginState := &arrayCore.State{}
	if _, err := tCtx.PluginStateReader().Get(pluginState); err != nil {
		return errors.Wrapf(errors.CorruptedPluginState, err, "Failed to read unmarshal custom state")
	}

	return TerminateSubTasks(ctx, tCtx, e.kubeClient, GetConfig(), abortSubtask, pluginState)
}

func (e Executor) Finalize(ctx context.Context, tCtx core.TaskExecutionContext) error {
	pluginState := &arrayCore.State{}
	if _, err := tCtx.PluginStateReader().Get(pluginState); err != nil {
		return errors.Wrapf(errors.CorruptedPluginState, err, "Failed to read unmarshal custom state")
	}

	return TerminateSubTasks(ctx, tCtx, e.kubeClient, GetConfig(), finalizeSubtask, pluginState)
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
