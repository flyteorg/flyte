package fakeplugins

import (
	"context"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	pluginCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/io"
)

// Phase constants
const (
	PhaseQueued       = "PHASE_QUEUED"
	PhaseInitializing = "PHASE_INITIALIZING"
	PhaseRunning      = "PHASE_RUNNING"
	PhaseSucceeded    = "PHASE_SUCCEEDED"
	PhaseFailed       = "PHASE_FAILED"
	PhaseAborted      = "PHASE_ABORTED"
)

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

func (n NextPhaseStatePlugin) Handle(ctx context.Context, phase string) (string, error) {
	logger := log.FromContext(ctx)
	var nextPhase string
	switch phase {
	case "":
		// New TaskAction - transition to Queued
		nextPhase = PhaseQueued
		logger.Info("New TaskAction detected, transitioning to Queued",
			"name", taskAction.Name, "action", actionSpec.ActionId.Name)

	case PhaseQueued:
		// Queued → Initializing
		nextPhase = PhaseInitializing
		logger.Info("Transitioning from Queued to Initializing",
			"name", taskAction.Name, "action", actionSpec.ActionId.Name)

	case PhaseInitializing:
		// Initializing → Running
		nextPhase = PhaseRunning
		logger.Info("Transitioning from Initializing to Running",
			"name", taskAction.Name, "action", actionSpec.ActionId.Name)

	case PhaseRunning:
		// Running → Succeeded (simulated execution)
		nextPhase = PhaseSucceeded
		logger.Info("Transitioning from Running to Succeeded",
			"name", taskAction.Name, "action", actionSpec.ActionId.Name)

	case PhaseSucceeded, PhaseFailed, PhaseAborted:
		// Terminal states - no further transitions
		logger.Info("TaskAction in terminal state",
			"name", taskAction.Name, "phase", currentPhase)
		return ctrl.Result{}, nil

	default:
		logger.Info("Unknown phase, resetting to Queued",
			"name", taskAction.Name, "phase", currentPhase)
		nextPhase = PhaseQueued
	}
	return nextPhase, nil

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
