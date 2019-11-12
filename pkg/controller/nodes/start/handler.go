package start

import (
	"context"

	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
)

type startHandler struct {
}

func (s startHandler) FinalizeRequired() bool {
	return false
}

func (s startHandler) Setup(ctx context.Context, setupContext handler.SetupContext) error {
	return nil
}

func (s startHandler) Handle(ctx context.Context, executionContext handler.NodeExecutionContext) (handler.Transition, error) {
	return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoSuccess(&handler.ExecutionInfo{})), nil
}

func (s startHandler) Abort(ctx context.Context, executionContext handler.NodeExecutionContext, reason string) error {
	return nil
}

func (s startHandler) Finalize(ctx context.Context, executionContext handler.NodeExecutionContext) error {
	return nil
}

func New() handler.Node {
	return &startHandler{}
}
