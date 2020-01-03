package mocks

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/repositories/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
)

type CreateOrUpdateWorkflowAttributesFunction func(ctx context.Context, input models.WorkflowAttributes) error
type GetWorkflowAttributesFunction func(ctx context.Context, project, domain, workflow, resource string) (
	models.WorkflowAttributes, error)
type UpdateWorkflowAttributesFunction func(ctx context.Context, input models.WorkflowAttributes) error

type MockWorkflowAttributesRepo struct {
	CreateOrUpdateFunction CreateOrUpdateWorkflowAttributesFunction
	GetFunction            GetWorkflowAttributesFunction
}

func (r *MockWorkflowAttributesRepo) CreateOrUpdate(ctx context.Context, input models.WorkflowAttributes) error {
	if r.CreateOrUpdateFunction != nil {
		return r.CreateOrUpdateFunction(ctx, input)
	}
	return nil
}

func (r *MockWorkflowAttributesRepo) Get(ctx context.Context, project, domain, workflow, resource string) (
	models.WorkflowAttributes, error) {
	if r.GetFunction != nil {
		return r.GetFunction(ctx, project, domain, workflow, resource)
	}
	return models.WorkflowAttributes{}, nil
}

func NewMockWorkflowAttributesRepo() interfaces.WorkflowAttributesRepoInterface {
	return &MockWorkflowAttributesRepo{}
}
