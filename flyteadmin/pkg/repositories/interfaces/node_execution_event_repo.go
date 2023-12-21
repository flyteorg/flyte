package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
)

//go:generate mockery -name=NodeExecutionEventRepoInterface -output=../mocks -case=underscore

type NodeExecutionEventRepoInterface interface {
	// Inserts a node execution event into the database store.
	Create(ctx context.Context, id *core.NodeExecutionIdentifier, input models.NodeExecutionEvent) error
}
