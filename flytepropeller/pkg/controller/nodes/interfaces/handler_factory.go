package interfaces

import (
	"context"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

//go:generate mockery -name HandlerFactory -case=underscore

type HandlerFactory interface {
	GetHandler(kind v1alpha1.NodeKind) (NodeHandler, error)
	Setup(ctx context.Context, executor Node, setup SetupContext) error
}
