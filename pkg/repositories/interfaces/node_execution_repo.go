package interfaces

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

// Defines the interface for interacting with node execution models.
type NodeExecutionRepoInterface interface {
	// Inserts a new node execution model and the first event that triggers it into the database store.
	Create(ctx context.Context, execution *models.NodeExecution) error
	// Updates an existing node execution in the database store with all non-empty fields in the input.
	Update(ctx context.Context, execution *models.NodeExecution) error
	// Returns a matching execution if it exists.
	Get(ctx context.Context, input NodeExecutionResource) (models.NodeExecution, error)
	// Returns node executions matching query parameters. A limit must be provided for the results page size.
	List(ctx context.Context, input ListResourceInput) (NodeExecutionCollectionOutput, error)
	// Return node execution events matching query parameters. A limit must be provided for the results page size.
	ListEvents(ctx context.Context, input ListResourceInput) (NodeExecutionEventCollectionOutput, error)
	// Returns whether a matching execution  exists.
	Exists(ctx context.Context, input NodeExecutionResource) (bool, error)
}

type NodeExecutionResource struct {
	NodeExecutionIdentifier core.NodeExecutionIdentifier
}

// Response format for a query on node executions.
type NodeExecutionCollectionOutput struct {
	NodeExecutions []models.NodeExecution
}

// Response format for a query on node execution events.
type NodeExecutionEventCollectionOutput struct {
	NodeExecutionEvents []models.NodeExecutionEvent
}
