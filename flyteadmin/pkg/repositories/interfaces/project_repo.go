package interfaces

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

type ProjectRepoInterface interface {
	// Inserts a namespace model into the database store.
	Create(ctx context.Context, project models.Project) error
	// Returns a matching project when it exists.
	Get(ctx context.Context, projectID string) (models.Project, error)
	// Returns projects matching query parameters.
	List(ctx context.Context, input ListResourceInput) ([]models.Project, error)
	// Given a project that exists in the DB and a partial set of fields to update
	// as a second project (projectUpdate), updates the original project which already
	// exists in the DB.
	UpdateProject(ctx context.Context, projectUpdate models.Project) error
}
