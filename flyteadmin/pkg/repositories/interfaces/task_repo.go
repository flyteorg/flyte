package interfaces

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

// Defines the interface for interacting with Task models.
type TaskRepoInterface interface {
	// Inserts a task model into the database store.
	Create(ctx context.Context, input models.Task, descriptionEntity *models.DescriptionEntity) error
	// Returns a matching task if it exists.
	Get(ctx context.Context, input Identifier) (models.Task, error)
	// Returns task revisions matching query parameters. A limit must be provided for the results page size.
	List(ctx context.Context, input ListResourceInput) (TaskCollectionOutput, error)
	// Returns tasks with only the project, name, and domain filled in.
	// A limit must be provided.
	ListTaskIdentifiers(ctx context.Context, input ListResourceInput) (TaskCollectionOutput, error)
}

// Response format for a query on tasks.
type TaskCollectionOutput struct {
	Tasks []models.Task
}
