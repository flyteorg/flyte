package interfaces

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/common"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
)

type ProjectRepoInterface interface {
	// Inserts a namespace model into the database store.
	Create(ctx context.Context, project models.Project) error
	// Returns a matching project when it exists.
	Get(ctx context.Context, projectID string) (models.Project, error)
	// Lists unique projects registered as namespaces
	ListAll(ctx context.Context, sortParameter common.SortParameter) ([]models.Project, error)
}
