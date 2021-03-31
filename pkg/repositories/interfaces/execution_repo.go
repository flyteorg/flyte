package interfaces

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

// Defines the interface for interacting with workflow execution models.
type ExecutionRepoInterface interface {
	// Inserts a workflow execution model into the database store.
	Create(ctx context.Context, input models.Execution) error
	// Updates an existing execution in the database store with all non-empty fields in the input.
	// This execution and event correspond to entire graph (workflow) executions.
	Update(ctx context.Context, event models.ExecutionEvent, execution models.Execution) error
	// This updates only an existing execution model with all non-empty fields in the input.
	UpdateExecution(ctx context.Context, execution models.Execution) error
	// Returns a matching execution if it exists.
	Get(ctx context.Context, input Identifier) (models.Execution, error)
	// Returns executions matching query parameters. A limit must be provided for the results page size.
	List(ctx context.Context, input ListResourceInput) (ExecutionCollectionOutput, error)
	// Returns a matching execution if it exists.
	Exists(ctx context.Context, input Identifier) (bool, error)
}

// Response format for a query on workflows.
type ExecutionCollectionOutput struct {
	Executions []models.Execution
}
