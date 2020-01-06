package interfaces

import (
	"context"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

// Interface for managing projects and domain -specific attributes.
type ProjectDomainAttributesInterface interface {
	UpdateProjectDomainAttributes(ctx context.Context, request admin.ProjectDomainAttributesUpdateRequest) (
		*admin.ProjectDomainAttributesUpdateResponse, error)
	GetProjectDomainAttributes(ctx context.Context, request admin.ProjectDomainAttributesGetRequest) (
		*admin.ProjectDomainAttributesGetResponse, error)
	DeleteProjectDomainAttributes(ctx context.Context, request admin.ProjectDomainAttributesDeleteRequest) (
		*admin.ProjectDomainAttributesDeleteResponse, error)
}
