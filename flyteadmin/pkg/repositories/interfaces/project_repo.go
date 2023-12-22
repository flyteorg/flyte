package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

type ProjectRepoInterface interface {
	// Inserts a namespace model into the database store.
	Create(ctx context.Context, id *admin.ProjectIdentifier, project models.Project) error
	// Returns a matching project when it exists.
	Get(ctx context.Context, id *admin.ProjectIdentifier) (models.Project, error)
	// Returns projects matching query parameters.
	List(ctx context.Context, input ListResourceInput) ([]models.Project, error)
	// Given a project that exists in the DB and a partial set of fields to update
	// as a second project (projectUpdate), updates the original project which already
	// exists in the DB.
	UpdateProject(ctx context.Context, id *admin.ProjectIdentifier, projectUpdate models.Project) error
}
