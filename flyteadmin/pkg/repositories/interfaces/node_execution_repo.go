package interfaces

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

// Defines the interface for interacting with node execution models.
type NodeExecutionRepoInterface interface {
	// Create a new node execution model and the first event that triggers it into the database store.
	Create(ctx context.Context, execution *models.NodeExecution) error
	// Update an existing node execution in the database store with all non-empty fields in the input.
	Update(ctx context.Context, execution *models.NodeExecution) error
	// Get returns a matching execution if it exists.
	Get(ctx context.Context, input NodeExecutionResource) (models.NodeExecution, error)
	// GetWithChildren returns a matching execution with preloaded child node executions. This should only be called for legacy node executions
	// which were created with eventVersion == 0
	GetWithChildren(ctx context.Context, input NodeExecutionResource) (models.NodeExecution, error)
	// List returns node executions matching query parameters. A limit must be provided for the results page size.
	List(ctx context.Context, input ListResourceInput) (NodeExecutionCollectionOutput, error)
	// Exists returns whether a matching execution exists.
	Exists(ctx context.Context, input NodeExecutionResource) (bool, error)
	// Returns count of node executions matching query parameters.
	Count(ctx context.Context, input CountResourceInput) (int64, error)
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
