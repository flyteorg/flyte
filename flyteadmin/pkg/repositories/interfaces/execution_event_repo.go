package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

//go:generate mockery -name=ExecutionEventRepoInterface -output=../mocks -case=underscore

type ExecutionEventRepoInterface interface {
	// Inserts a workflow execution event into the database store.
	Create(ctx context.Context, id *core.WorkflowExecutionIdentifier, input models.ExecutionEvent) error
}
