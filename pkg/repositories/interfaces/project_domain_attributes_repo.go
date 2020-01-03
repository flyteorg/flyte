package interfaces

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/repositories/models"
)

type ProjectDomainAttributesRepoInterface interface {
	// Inserts or updates an existing ProjectDomainAttributes model into the database store.
	CreateOrUpdate(ctx context.Context, input models.ProjectDomainAttributes) error
	// Returns a matching ProjectDomainAttributes model when it exists.
	Get(ctx context.Context, project, domain, resource string) (models.ProjectDomainAttributes, error)
}
