package fakeplugins

import (
	"context"
	pluginCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/io"
	"time"
)

// PhaseToStringMap maps core.Phase to phase string
var PhaseToStringMap = map[pluginCore.Phase]string{
	pluginCore.PhaseQueued:           "PHASE_QUEUED",
	pluginCore.PhaseInitializing:     "PHASE_INITIALIZING",
	pluginCore.PhaseRunning:          "PHASE_RUNNING",
	pluginCore.PhaseSuccess:          "PHASE_SUCCEEDED",
	pluginCore.PhaseRetryableFailure: "PHASE_FAILED",
	pluginCore.PhaseAborted:          "PHASE_ABORTED",
}

type NextPhaseState struct {
	Phase        pluginCore.Phase
	PhaseVersion uint32
	TaskInfo     *pluginCore.TaskInfo
	TaskErr      *io.ExecutionError
	DeckExists   bool
	OutputExists bool
	OrError      bool
}

type NextPhaseStatePlugin struct {
	id    string
	props pluginCore.PluginProperties
}

func (n NextPhaseStatePlugin) GetID() string {
	return n.id
}

func (n NextPhaseStatePlugin) GetProperties() pluginCore.PluginProperties {
	return n.props
}

func (n NextPhaseStatePlugin) Handle(ctx context.Context, tCtx pluginCore.TaskExecutionContext) (pluginCore.Transition, error) {
	// Get current plugin phase
	s := &NextPhaseState{}
	phaseVersion, err := tCtx.PluginStateReader().Get(s)
	if err != nil {
		return pluginCore.Transition{}, err
	}

	switch pluginCore.Phase(phaseVersion) {
	case pluginCore.PhaseSuccess, pluginCore.PhasePermanentFailure, pluginCore.PhaseRetryableFailure:
		return pluginCore.Transition{}, nil
	case pluginCore.PhaseInitializing:
		return pluginCore.DoTransition(pluginCore.PhaseInfoRunning(s.PhaseVersion, s.TaskInfo)), nil
	case pluginCore.PhaseQueued:
		return pluginCore.DoTransition(pluginCore.PhaseInfoInitializing(time.Now(), s.PhaseVersion, "initializing", s.TaskInfo)), nil
	case pluginCore.PhaseRunning:
		return pluginCore.DoTransition(pluginCore.PhaseInfoSuccess(s.TaskInfo)), nil
	case pluginCore.PhaseNotReady:
		return pluginCore.DoTransition(pluginCore.PhaseInfoQueued(time.Now(), s.PhaseVersion, "Queueing")), nil
	default:
		return pluginCore.Transition{}, nil
	}
}

func (n NextPhaseStatePlugin) Abort(ctx context.Context, tCtx pluginCore.TaskExecutionContext) error {
	return nil
}

func (n NextPhaseStatePlugin) Finalize(ctx context.Context, tCtx pluginCore.TaskExecutionContext) error {
	return nil
}

func NewPhaseBasedPlugin() NextPhaseStatePlugin {
	return NextPhaseStatePlugin{
		id:    "next_phase_plugin",
		props: pluginCore.PluginProperties{},
	}
}
