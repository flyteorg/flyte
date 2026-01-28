package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
)


type ExecutionEventRepoInterface interface {
	// Inserts a workflow execution event into the database store.
	Create(ctx context.Context, input models.ExecutionEvent) error
}
