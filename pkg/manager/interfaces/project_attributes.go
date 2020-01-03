package interfaces

import (
	"context"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

// Interface for managing project-specific attributes.
type ProjectAttributesInterface interface {
	UpdateProjectAttributes(ctx context.Context, request admin.ProjectAttributesUpdateRequest) (
		*admin.ProjectAttributesUpdateResponse, error)
}
