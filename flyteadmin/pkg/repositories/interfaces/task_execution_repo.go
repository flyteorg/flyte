package interfaces

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

// Defines the interface for interacting with task execution models.
type TaskExecutionRepoInterface interface {
	// Inserts a task execution model into the database store.
	Create(ctx context.Context, input models.TaskExecution) error
	// Updates an existing task execution in the database store with all non-empty fields in the input.
	Update(ctx context.Context, execution models.TaskExecution) error
	// Returns a matching execution if it exists.
	Get(ctx context.Context, input GetTaskExecutionInput) (models.TaskExecution, error)
	// Returns task executions matching query parameters. A limit must be provided for the results page size.
	List(ctx context.Context, input ListResourceInput) (TaskExecutionCollectionOutput, error)
	// Returns count of task executions matching query parameters.
	Count(ctx context.Context, input CountResourceInput) (int64, error)
}

type GetTaskExecutionInput struct {
	TaskExecutionID core.TaskExecutionIdentifier
}

// Response format for a query on task executions.
type TaskExecutionCollectionOutput struct {
	TaskExecutions []models.TaskExecution
}
