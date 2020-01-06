package interfaces

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/repositories/models"
)

type WorkflowAttributesRepoInterface interface {
	// Inserts or updates an existing WorkflowAttributes model into the database store.
	CreateOrUpdate(ctx context.Context, input models.WorkflowAttributes) error
	// Returns a matching WorkflowAttributes model when it exists.
	Get(ctx context.Context, project, domain, workflow, resource string) (models.WorkflowAttributes, error)
	// Deletes a matching ProjectDomainAttributes model when it exists.
	Delete(ctx context.Context, project, domain, workflow, resource string) error
}
