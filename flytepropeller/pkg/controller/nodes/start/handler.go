package start

import (
	"context"

	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/interfaces"
)

type startHandler struct {
}

func (s startHandler) FinalizeRequired() bool {
	return false
}

func (s startHandler) Setup(ctx context.Context, setupContext interfaces.SetupContext) error {
	return nil
}

func (s startHandler) Handle(ctx context.Context, executionContext interfaces.NodeExecutionContext) (handler.Transition, error) {
	return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoSuccess(&handler.ExecutionInfo{})), nil
}

func (s startHandler) Abort(ctx context.Context, executionContext interfaces.NodeExecutionContext, reason string) error {
	return nil
}

func (s startHandler) Finalize(ctx context.Context, executionContext interfaces.NodeExecutionContext) error {
	return nil
}

func New() interfaces.NodeHandler {
	return &startHandler{}
}
