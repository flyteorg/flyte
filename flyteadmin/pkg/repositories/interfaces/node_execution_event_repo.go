package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
)

//go:generate mockery-v2 --name=NodeExecutionEventRepoInterface --output=../mocks --case=underscore --with-expecter

type NodeExecutionEventRepoInterface interface {
	// Inserts a node execution event into the database store.
	Create(ctx context.Context, input models.NodeExecutionEvent) error
}
