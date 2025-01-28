package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

//go:generate mockery-v2 --name=WorkflowInterface --output=../mocks --case=underscore --with-expecter

// Interface for managing Flyte Workflows
type WorkflowInterface interface {
	CreateWorkflow(ctx context.Context, request *admin.WorkflowCreateRequest) (*admin.WorkflowCreateResponse, error)
	GetWorkflow(ctx context.Context, request *admin.ObjectGetRequest) (*admin.Workflow, error)
	ListWorkflows(ctx context.Context, request *admin.ResourceListRequest) (*admin.WorkflowList, error)
	ListWorkflowIdentifiers(ctx context.Context, request *admin.NamedEntityIdentifierListRequest) (
		*admin.NamedEntityIdentifierList, error)
}
