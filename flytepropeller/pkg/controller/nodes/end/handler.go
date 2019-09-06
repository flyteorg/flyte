package end

import (
	"context"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/errors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/storage"
)

type endHandler struct {
	store storage.ProtobufStore
}

func (e *endHandler) Initialize(ctx context.Context) error {
	return nil
}

func (e *endHandler) StartNode(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode, nodeInputs *handler.Data) (handler.Status, error) {
	if nodeInputs != nil {
		logger.Debugf(ctx, "Workflow has outputs. Storing them.")
		nodeStatus := w.GetNodeExecutionStatus(node.GetID())
		o := v1alpha1.GetOutputsFile(nodeStatus.GetDataDir())
		so := storage.Options{}
		if err := e.store.WriteProtobuf(ctx, o, so, nodeInputs); err != nil {
			logger.Errorf(ctx, "Failed to store workflow outputs. Error [%s]", err)
			return handler.StatusUndefined, errors.Wrapf(errors.CausedByError, node.GetID(), err, "Failed to store workflow outputs, as end-node")
		}
	}
	logger.Debugf(ctx, "End node success")
	return handler.StatusSuccess, nil
}

func (e *endHandler) CheckNodeStatus(ctx context.Context, g v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode, nodeStatus v1alpha1.ExecutableNodeStatus) (handler.Status, error) {
	return handler.StatusSuccess, nil
}

func (e *endHandler) HandleFailingNode(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode) (handler.Status, error) {
	return handler.StatusFailed(errors.Errorf(errors.IllegalStateError, node.GetID(), "End node cannot enter a failing state")), nil
}

func (e *endHandler) AbortNode(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode) error {
	return nil
}

func New(store storage.ProtobufStore) handler.IFace {
	return &endHandler{
		store: store,
	}
}
