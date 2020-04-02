package interfaces

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/repositories/models"
)

// Defines the interface for interacting with Workflow models.
type WorkflowRepoInterface interface {
	// Inserts a workflow model into the database store.
	Create(ctx context.Context, input models.Workflow) error
	// Returns a matching workflow if it exists.
	Get(ctx context.Context, input GetResourceInput) (models.Workflow, error)
	// Returns workflow revisions matching query parameters. A limit must be provided for the results page size.
	List(ctx context.Context, input ListResourceInput) (WorkflowCollectionOutput, error)
	ListIdentifiers(ctx context.Context, input ListResourceInput) (WorkflowCollectionOutput, error)
	// Updates an existing workflow in the database store.
	Update(ctx context.Context, input models.Workflow) error
}

// Response format for a query on workflows.
type WorkflowCollectionOutput struct {
	Workflows []models.Workflow
}
