package interfaces

import (
	"context"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

// Interface for managing project, domain and workflow -specific attributes.
type WorkflowAttributesInterface interface {
	UpdateWorkflowAttributes(ctx context.Context, request admin.WorkflowAttributesUpdateRequest) (
		*admin.WorkflowAttributesUpdateResponse, error)
	GetWorkflowAttributes(ctx context.Context, request admin.WorkflowAttributesGetRequest) (
		*admin.WorkflowAttributesGetResponse, error)
	DeleteWorkflowAttributes(ctx context.Context, request admin.WorkflowAttributesDeleteRequest) (
		*admin.WorkflowAttributesDeleteResponse, error)
}
