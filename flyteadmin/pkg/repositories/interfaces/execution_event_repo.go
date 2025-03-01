package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
)

//go:generate mockery --name=ExecutionEventRepoInterface --output=../mocks --case=underscore --with-expecter

type ExecutionEventRepoInterface interface {
	// Inserts a workflow execution event into the database store.
	Create(ctx context.Context, input models.ExecutionEvent) error
}
