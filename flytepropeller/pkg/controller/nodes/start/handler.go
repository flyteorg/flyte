package start

import (
	"context"

	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
)

type startHandler struct {
}

func (s startHandler) FinalizeRequired() bool {
	return false
}

func (s startHandler) Setup(context.Context, interfaces.SetupContext) error {
	return nil
}

func (s startHandler) Handle(context.Context, interfaces.NodeExecutionContext) (handler.Transition, error) {
	return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoSuccess(&handler.ExecutionInfo{})), nil
}

func (s startHandler) Abort(context.Context, interfaces.NodeExecutionContext, string) error {
	return nil
}

func (s startHandler) Finalize(context.Context, interfaces.NodeExecutionContext) error {
	return nil
}

func New() interfaces.NodeHandler {
	return &startHandler{}
}
