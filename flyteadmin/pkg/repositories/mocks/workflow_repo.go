// Mock implementation of a workflow repo to be used for tests.
package mocks

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

type CreateWorkflowFunc func(input models.Workflow, descriptionEntity *models.DescriptionEntity) error
type GetWorkflowFunc func(id *core.Identifier) (models.Workflow, error)
type ListWorkflowFunc func(input interfaces.ListResourceInput) (interfaces.WorkflowCollectionOutput, error)
type ListIdentifiersFunc func(input interfaces.ListResourceInput) (interfaces.WorkflowCollectionOutput, error)

type MockWorkflowRepo struct {
	createFunction      CreateWorkflowFunc
	getFunction         GetWorkflowFunc
	listFunction        ListWorkflowFunc
	listIdentifiersFunc ListIdentifiersFunc
}

func (r *MockWorkflowRepo) Create(ctx context.Context, id *core.Identifier, input models.Workflow, descriptionEntity *models.DescriptionEntity) error {
	if r.createFunction != nil {
		return r.createFunction(input, descriptionEntity)
	}
	return nil
}

func (r *MockWorkflowRepo) SetCreateCallback(createFunction CreateWorkflowFunc) {
	r.createFunction = createFunction
}

func (r *MockWorkflowRepo) Get(ctx context.Context, id *core.Identifier) (models.Workflow, error) {
	if r.getFunction != nil {
		return r.getFunction(id)
	}
	return models.Workflow{
		WorkflowKey: models.WorkflowKey{
			Project: id.Project,
			Domain:  id.Domain,
			Name:    id.Name,
			Version: id.Version,
		},
	}, nil
}

func (r *MockWorkflowRepo) SetGetCallback(getFunction GetWorkflowFunc) {
	r.getFunction = getFunction
}

func (r *MockWorkflowRepo) List(
	ctx context.Context, input interfaces.ListResourceInput) (interfaces.WorkflowCollectionOutput, error) {
	if r.listFunction != nil {
		return r.listFunction(input)
	}
	return interfaces.WorkflowCollectionOutput{}, nil
}

func (r *MockWorkflowRepo) SetListCallback(listFunction ListWorkflowFunc) {
	r.listFunction = listFunction
}

func (r *MockWorkflowRepo) SetListIdentifiersFunc(fn ListIdentifiersFunc) {
	r.listIdentifiersFunc = fn
}

func (r *MockWorkflowRepo) ListIdentifiers(ctx context.Context, input interfaces.ListResourceInput) (
	interfaces.WorkflowCollectionOutput, error) {

	if r.listIdentifiersFunc != nil {
		return r.listIdentifiersFunc(input)
	}

	return interfaces.WorkflowCollectionOutput{}, nil
}

func NewMockWorkflowRepo() interfaces.WorkflowRepoInterface {
	return &MockWorkflowRepo{}
}
