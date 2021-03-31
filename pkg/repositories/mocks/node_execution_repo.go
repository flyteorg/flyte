package mocks

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

type CreateNodeExecutionFunc func(ctx context.Context, event *models.NodeExecutionEvent, input *models.NodeExecution) error
type UpdateNodeExecutionFunc func(ctx context.Context, event *models.NodeExecutionEvent, nodeExecution *models.NodeExecution) error
type GetNodeExecutionFunc func(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error)
type ListNodeExecutionFunc func(ctx context.Context, input interfaces.ListResourceInput) (
	interfaces.NodeExecutionCollectionOutput, error)
type ListNodeExecutionEventFunc func(ctx context.Context, input interfaces.ListResourceInput) (
	interfaces.NodeExecutionEventCollectionOutput, error)

type MockNodeExecutionRepo struct {
	createFunction    CreateNodeExecutionFunc
	updateFunction    UpdateNodeExecutionFunc
	getFunction       GetNodeExecutionFunc
	listFunction      ListNodeExecutionFunc
	listEventFunction ListNodeExecutionEventFunc
	ExistsFunction    func(ctx context.Context, input interfaces.NodeExecutionResource) (bool, error)
}

func (r *MockNodeExecutionRepo) Create(ctx context.Context, event *models.NodeExecutionEvent, input *models.NodeExecution) error {
	if r.createFunction != nil {
		return r.createFunction(ctx, event, input)
	}
	return nil
}

func (r *MockNodeExecutionRepo) SetCreateCallback(createFunction CreateNodeExecutionFunc) {
	r.createFunction = createFunction
}

func (r *MockNodeExecutionRepo) Update(ctx context.Context, event *models.NodeExecutionEvent, nodeExecution *models.NodeExecution) error {
	if r.updateFunction != nil {
		return r.updateFunction(ctx, event, nodeExecution)
	}
	return nil
}

func (r *MockNodeExecutionRepo) SetUpdateCallback(updateFunction UpdateNodeExecutionFunc) {
	r.updateFunction = updateFunction
}

func (r *MockNodeExecutionRepo) Get(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
	if r.getFunction != nil {
		return r.getFunction(ctx, input)
	}
	return models.NodeExecution{}, nil
}

func (r *MockNodeExecutionRepo) SetGetCallback(getFunction GetNodeExecutionFunc) {
	r.getFunction = getFunction
}

func (r *MockNodeExecutionRepo) List(ctx context.Context, input interfaces.ListResourceInput) (
	interfaces.NodeExecutionCollectionOutput, error) {
	if r.listFunction != nil {
		return r.listFunction(ctx, input)
	}
	return interfaces.NodeExecutionCollectionOutput{}, nil
}

func (r *MockNodeExecutionRepo) SetListCallback(listFunction ListNodeExecutionFunc) {
	r.listFunction = listFunction
}

func (r *MockNodeExecutionRepo) ListEvents(ctx context.Context, input interfaces.ListResourceInput) (
	interfaces.NodeExecutionEventCollectionOutput, error) {
	if r.listFunction != nil {
		return r.listEventFunction(ctx, input)
	}
	return interfaces.NodeExecutionEventCollectionOutput{}, nil
}

func (r *MockNodeExecutionRepo) SetListEventCallback(listEventFunction ListNodeExecutionEventFunc) {
	r.listEventFunction = listEventFunction
}

func (r *MockNodeExecutionRepo) Exists(ctx context.Context, input interfaces.NodeExecutionResource) (bool, error) {
	if r.ExistsFunction != nil {
		return r.ExistsFunction(ctx, input)
	}
	return true, nil
}

func NewMockNodeExecutionRepo() interfaces.NodeExecutionRepoInterface {
	return &MockNodeExecutionRepo{}
}
