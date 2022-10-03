package interfaces

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

type ResourceRepoInterface interface {
	// Inserts or updates an existing Type model into the database store.
	CreateOrUpdate(ctx context.Context, input models.Resource) error
	// Returns a matching Type model based on hierarchical resolution.
	Get(ctx context.Context, ID ResourceID) (models.Resource, error)
	// Returns a matching Type model.
	GetRaw(ctx context.Context, ID ResourceID) (models.Resource, error)
	// GetProjectLevel returns the Project level resource entry, if any, even if there is a higher
	// specificity resource.
	GetProjectLevel(ctx context.Context, ID ResourceID) (models.Resource, error)
	// Lists all resources
	ListAll(ctx context.Context, resourceType string) ([]models.Resource, error)
	// Deletes a matching Type model when it exists.
	Delete(ctx context.Context, ID ResourceID) error
}

type ResourceID struct {
	Project      string
	Domain       string
	Workflow     string
	LaunchPlan   string
	ResourceType string
}
