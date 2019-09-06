package start

import (
	"context"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/errors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
	"github.com/lyft/flytestdlib/storage"
)

type startHandler struct {
	store *storage.DataStore
}

func (s startHandler) Initialize(ctx context.Context) error {
	return nil
}

func (s *startHandler) StartNode(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode, nodeInputs *handler.Data) (handler.Status, error) {
	return handler.StatusSuccess, nil
}

func (s *startHandler) CheckNodeStatus(ctx context.Context, g v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode, nodeStatus v1alpha1.ExecutableNodeStatus) (handler.Status, error) {
	return handler.StatusSuccess, nil
}

func (s *startHandler) HandleFailingNode(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode) (handler.Status, error) {
	return handler.StatusFailed(errors.Errorf(errors.IllegalStateError, node.GetID(), "start node cannot enter a failing state")), nil
}

func (s *startHandler) AbortNode(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode) error {
	return nil
}

func New(store *storage.DataStore) handler.IFace {
	return &startHandler{
		store: store,
	}
}
