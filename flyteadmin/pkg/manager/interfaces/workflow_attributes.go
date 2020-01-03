package interfaces

import (
	"context"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

// Interface for managing project, domain and workflow -specific attributes.
type WorkflowAttributesInterface interface {
	UpdateWorkflowAttributes(ctx context.Context, request admin.WorkflowAttributesUpdateRequest) (
		*admin.WorkflowAttributesUpdateResponse, error)
}
