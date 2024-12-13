package ext

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
)

//go:generate mockery -all -case=underscore

// AdminUpdaterExtInterface Interface for exposing the update capabilities from the admin
type AdminUpdaterExtInterface interface {
	AdminServiceClient() service.AdminServiceClient

	// UpdateWorkflowAttributes updates workflow attributes within a project, domain for a particular matchable resource
	UpdateWorkflowAttributes(ctx context.Context, project, domain, name string, matchingAttr *admin.MatchingAttributes) error

	// UpdateProjectDomainAttributes updates project domain attributes for a particular matchable resource
	UpdateProjectDomainAttributes(ctx context.Context, project, domain string, matchingAttr *admin.MatchingAttributes) error

	// UpdateProjectAttributes updates project attributes for a particular matchable resource
	UpdateProjectAttributes(ctx context.Context, project string, matchingAttr *admin.MatchingAttributes) error
}

// AdminUpdaterExtClient is used for interacting with extended features used for updating data in admin service
type AdminUpdaterExtClient struct {
	AdminClient service.AdminServiceClient
}

func (a *AdminUpdaterExtClient) AdminServiceClient() service.AdminServiceClient {
	if a == nil {
		return nil
	}
	return a.AdminClient
}
