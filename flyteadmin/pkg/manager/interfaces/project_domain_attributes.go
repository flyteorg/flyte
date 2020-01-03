package interfaces

import (
	"context"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

// Interface for managing projects and domain -specific attributes.
type ProjectDomainAttributesInterface interface {
	UpdateProjectDomainAttributes(ctx context.Context, request admin.ProjectDomainAttributesUpdateRequest) (
		*admin.ProjectDomainAttributesUpdateResponse, error)
}
