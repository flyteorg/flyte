package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

type HandlerFactory interface {
	GetHandler(kind v1alpha1.NodeKind) (NodeHandler, error)
	Setup(ctx context.Context, executor Node, setup SetupContext) error
}
