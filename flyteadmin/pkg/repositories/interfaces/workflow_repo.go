package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

// Defines the interface for interacting with Workflow models.
type WorkflowRepoInterface interface {
	// Inserts a workflow model into the database store.
	Create(ctx context.Context, id *core.Identifier, input models.Workflow, descriptionEntity *models.DescriptionEntity) error
	// Returns a matching workflow if it exists.
	Get(ctx context.Context, id *core.Identifier) (models.Workflow, error)
	// Returns workflow revisions matching query parameters. A limit must be provided for the results page size.
	List(ctx context.Context, input ListResourceInput) (WorkflowCollectionOutput, error)
	ListIdentifiers(ctx context.Context, input ListResourceInput) (WorkflowCollectionOutput, error)
}

// Response format for a query on workflows.
type WorkflowCollectionOutput struct {
	Workflows []models.Workflow
}
