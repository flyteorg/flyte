package interfaces

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/repositories/models"
)

type ProjectDomainRepoInterface interface {
	// Inserts or updates an existing ProjectDomain model into the database store.
	CreateOrUpdate(ctx context.Context, input models.ProjectDomain) error
	// Returns a matching project when it exists.
	Get(ctx context.Context, project, domain string) (models.ProjectDomain, error)
}
