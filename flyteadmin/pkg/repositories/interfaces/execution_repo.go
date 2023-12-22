package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

// Defines the interface for interacting with workflow execution models.
type ExecutionRepoInterface interface {
	// Inserts a workflow execution model into the database store.
	Create(ctx context.Context, id *core.WorkflowExecutionIdentifier, input models.Execution) error
	// This updates only an existing execution model with all non-empty fields in the input.
	Update(ctx context.Context, id *core.WorkflowExecutionIdentifier, execution models.Execution) error
	// Returns a matching execution if it exists.
	Get(ctx context.Context, id *core.WorkflowExecutionIdentifier) (models.Execution, error)
	// Returns executions matching query parameters. A limit must be provided for the results page size.
	List(ctx context.Context, input ListResourceInput) (ExecutionCollectionOutput, error)
	// Returns count of executions matching query parameters.
	Count(ctx context.Context, input CountResourceInput) (int64, error)
}

// Response format for a query on workflows.
type ExecutionCollectionOutput struct {
	Executions []models.Execution
}
