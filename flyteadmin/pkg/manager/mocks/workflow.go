package mocks

import (
	"context"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

type CreateWorkflowFunc func(ctx context.Context, request admin.WorkflowCreateRequest) (*admin.WorkflowCreateResponse, error)

type MockWorkflowManager struct {
	createWorkflowFunc CreateWorkflowFunc
}

func (r *MockWorkflowManager) SetCreateCallback(createFunction CreateWorkflowFunc) {
	r.createWorkflowFunc = createFunction
}

func (r *MockWorkflowManager) CreateWorkflow(
	ctx context.Context,
	request admin.WorkflowCreateRequest) (*admin.WorkflowCreateResponse, error) {
	if r.createWorkflowFunc != nil {
		return r.createWorkflowFunc(ctx, request)
	}
	return nil, nil
}

func (r *MockWorkflowManager) ListWorkflows(ctx context.Context,
	request admin.ResourceListRequest) (*admin.WorkflowList, error) {
	return nil, nil
}

func (r *MockWorkflowManager) GetWorkflow(
	ctx context.Context, request admin.ObjectGetRequest) (*admin.Workflow, error) {
	return nil, nil
}

func (r *MockWorkflowManager) ListWorkflowIdentifiers(ctx context.Context, request admin.NamedEntityIdentifierListRequest) (
	*admin.NamedEntityIdentifierList, error) {
	return nil, nil
}

func (r *MockWorkflowManager) UpdateWorkflow(ctx context.Context, request admin.WorkflowUpdateRequest) (
	*admin.WorkflowUpdateResponse, error) {
	return nil, nil
}
