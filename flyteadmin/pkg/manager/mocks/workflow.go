package mocks

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

type CreateWorkflowFunc func(ctx context.Context, request admin.WorkflowCreateRequest) (*admin.WorkflowCreateResponse, error)
type GetWorkflowFunc func(ctx context.Context, request admin.ObjectGetRequest) (*admin.Workflow, error)

type MockWorkflowManager struct {
	createWorkflowFunc CreateWorkflowFunc
	getWorkflowFunc    GetWorkflowFunc
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

func (r *MockWorkflowManager) SetGetCallback(getFunction GetWorkflowFunc) {
	r.getWorkflowFunc = getFunction
}

func (r *MockWorkflowManager) GetWorkflow(
	ctx context.Context, request admin.ObjectGetRequest) (*admin.Workflow, error) {
	if r.getWorkflowFunc != nil {
		return r.getWorkflowFunc(ctx, request)
	}
	return nil, nil
}

func (r *MockWorkflowManager) ListWorkflowIdentifiers(ctx context.Context, request admin.NamedEntityIdentifierListRequest) (
	*admin.NamedEntityIdentifierList, error) {
	return nil, nil
}
