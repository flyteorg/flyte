package ext

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
)

//go:generate mockery -all -case=underscore

// AdminDeleterExtInterface Interface for exposing the update capabilities from the admin
type AdminDeleterExtInterface interface {
	AdminServiceClient() service.AdminServiceClient

	// DeleteWorkflowAttributes deletes workflow attributes within a project, domain for a particular matchable resource
	DeleteWorkflowAttributes(ctx context.Context, project, domain, name string, rsType admin.MatchableResource) error

	// DeleteProjectDomainAttributes deletes project domain attributes for a particular matchable resource
	DeleteProjectDomainAttributes(ctx context.Context, project, domain string, rsType admin.MatchableResource) error

	// DeleteProjectAttributes deletes project attributes for a particular matchable resource
	DeleteProjectAttributes(ctx context.Context, project string, rsType admin.MatchableResource) error
}

// AdminDeleterExtClient is used for interacting with extended features used for deleting/archiving data in admin service
type AdminDeleterExtClient struct {
	AdminClient service.AdminServiceClient
}

func (a *AdminDeleterExtClient) AdminServiceClient() service.AdminServiceClient {
	if a == nil {
		return nil
	}
	return a.AdminClient
}
