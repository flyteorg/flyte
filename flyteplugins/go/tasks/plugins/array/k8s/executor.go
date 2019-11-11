package k8s

import (
	"context"

	idlCore "github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/lyft/flyteplugins/go/tasks/plugins/array"
	arrayCore "github.com/lyft/flyteplugins/go/tasks/plugins/array/core"
	"github.com/lyft/flytestdlib/promutils"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery"

	"github.com/lyft/flyteplugins/go/tasks/errors"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
)

const executorName = "k8s-array"
const arrayTaskType = "container_array"
const pluginStateVersion = 0

type Executor struct {
	kubeClient       core.KubeClient
	outputsAssembler array.OutputAssembler
	errorAssembler   array.OutputAssembler
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

	switch p, _ := pluginState.GetPhase(); p {
	case arrayCore.PhaseStart:
		nextState, err = array.DetermineDiscoverability(ctx, tCtx, pluginState)

	case arrayCore.PhasePreLaunch:
		nextState = pluginState.SetPhase(arrayCore.PhaseLaunch, core.DefaultPhaseVersion).SetReason("Nothing to do in PreLaunch phase.")
		err = nil

	case arrayCore.PhaseWaitingForResources:
		fallthrough

	case arrayCore.PhaseLaunch:
		nextState, err = LaunchSubTasks(ctx, tCtx, e.kubeClient, pluginConfig, pluginState)

	case arrayCore.PhaseCheckingSubTaskExecutions:
		nextState, logLinks, err = CheckSubTasksState(ctx, tCtx, e.kubeClient, tCtx.DataStore(),
			tCtx.OutputWriter().GetOutputPrefixPath(), pluginState)

	case arrayCore.PhaseAssembleFinalOutput:
		nextState, err = array.AssembleFinalOutputs(ctx, e.outputsAssembler, tCtx, arrayCore.PhaseSuccess, pluginState)

	case arrayCore.PhaseWriteToDiscoveryThenFail:
		nextState, err = array.WriteToDiscovery(ctx, tCtx, pluginState, arrayCore.PhaseAssembleFinalError)

	case arrayCore.PhaseWriteToDiscovery:
		nextState, err = array.WriteToDiscovery(ctx, tCtx, pluginState, arrayCore.PhaseAssembleFinalOutput)

	case arrayCore.PhaseAssembleFinalError:
		nextState, err = array.AssembleFinalOutputs(ctx, e.errorAssembler, tCtx, arrayCore.PhaseRetryableFailure, pluginState)

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
	phaseInfo, err := arrayCore.MapArrayStateToPluginPhase(ctx, nextState, logLinks)
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

	return TerminateSubTasks(ctx, tCtx.TaskExecutionMetadata(), e.kubeClient, pluginConfig.MaxErrorStringLength,
		pluginState)
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
	exec, err := NewExecutor(iCtx.KubeClient(), GetConfig(), iCtx.MetricsScope())
	if err != nil {
		return nil, err
	}

	if err = exec.Start(ctx); err != nil {
		return nil, err
	}

	return exec, nil
}
