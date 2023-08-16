package end

import (
	"context"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/storage"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/errors"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/interfaces"
)

type endHandler struct {
}

func (e endHandler) FinalizeRequired() bool {
	return false
}

func (e endHandler) Setup(ctx context.Context, setupContext interfaces.SetupContext) error {
	return nil
}

func (e endHandler) Handle(ctx context.Context, executionContext interfaces.NodeExecutionContext) (handler.Transition, error) {
	inputs, err := executionContext.InputReader().Get(ctx)
	if err != nil {
		return handler.UnknownTransition, err
	}
	if inputs != nil {
		logger.Debugf(ctx, "Workflow has outputs. Storing them.")
		// TODO we should use OutputWriter here
		o := v1alpha1.GetOutputsFile(executionContext.NodeStatus().GetOutputDir())
		so := storage.Options{}
		if err := executionContext.DataStore().WriteProtobuf(ctx, o, so, inputs); err != nil {
			logger.Errorf(ctx, "Failed to store workflow outputs. Error [%s]", err)
			return handler.UnknownTransition, errors.Wrapf(errors.CausedByError, executionContext.NodeID(), err, "Failed to store workflow outputs, as end-node")
		}
	}
	logger.Debugf(ctx, "End node success")
	return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoSuccess(nil)), nil
}

func (e endHandler) Abort(_ context.Context, _ interfaces.NodeExecutionContext, _ string) error {
	return nil
}

func (e endHandler) Finalize(_ context.Context, _ interfaces.NodeExecutionContext) error {
	return nil
}

func New() interfaces.NodeHandler {
	return &endHandler{}
}
