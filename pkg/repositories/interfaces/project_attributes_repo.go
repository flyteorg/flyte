package interfaces

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/repositories/models"
)

type ProjectAttributesRepoInterface interface {
	// Inserts or updates an existing ProjectAttributes model into the database store.
	CreateOrUpdate(ctx context.Context, input models.ProjectAttributes) error
	// Returns a matching ProjectAttributes model when it exists.
	Get(ctx context.Context, project, resource string) (models.ProjectAttributes, error)
	// Deletes a matching ProjectAttributes model when it exists.
	Delete(ctx context.Context, project, resource string) error
}
