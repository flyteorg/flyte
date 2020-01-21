package interfaces

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/repositories/models"
)

type ResourceRepoInterface interface {
	// Inserts or updates an existing Type model into the database store.
	CreateOrUpdate(ctx context.Context, input models.Resource) error
	// Returns a matching Type model based on hierarchical resolution.
	Get(ctx context.Context, ID ResourceID) (models.Resource, error)
	// Returns a matching Type model.
	GetRaw(ctx context.Context, ID ResourceID) (models.Resource, error)
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
