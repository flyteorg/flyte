package mocks

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

type CreateNodeExecutionFunc func(ctx context.Context, input *models.NodeExecution) error
type UpdateNodeExecutionFunc func(ctx context.Context, nodeExecution *models.NodeExecution) error
type GetNodeExecutionFunc func(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error)
type ListNodeExecutionFunc func(ctx context.Context, input interfaces.ListResourceInput) (
	interfaces.NodeExecutionCollectionOutput, error)
type ExistsNodeExecutionFunc func(ctx context.Context, input interfaces.NodeExecutionResource) (bool, error)
type CountNodeExecutionFunc func(ctx context.Context, input interfaces.CountResourceInput) (int64, error)

type MockNodeExecutionRepo struct {
	createFunction          CreateNodeExecutionFunc
	updateFunction          UpdateNodeExecutionFunc
	getFunction             GetNodeExecutionFunc
	getWithChildrenFunction GetNodeExecutionFunc
	listFunction            ListNodeExecutionFunc
	existsFunction          ExistsNodeExecutionFunc
	countFunction           CountNodeExecutionFunc
}

func (r *MockNodeExecutionRepo) Create(ctx context.Context, input *models.NodeExecution) error {
	if r.createFunction != nil {
		return r.createFunction(ctx, input)
	}
	return nil
}

func (r *MockNodeExecutionRepo) SetCreateCallback(createFunction CreateNodeExecutionFunc) {
	r.createFunction = createFunction
}

func (r *MockNodeExecutionRepo) Update(ctx context.Context, nodeExecution *models.NodeExecution) error {
	if r.updateFunction != nil {
		return r.updateFunction(ctx, nodeExecution)
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

func (r *MockNodeExecutionRepo) GetWithChildren(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
	if r.getWithChildrenFunction != nil {
		return r.getWithChildrenFunction(ctx, input)
	}
	return models.NodeExecution{}, nil
}

func (r *MockNodeExecutionRepo) SetGetWithChildrenCallback(getWithChildrenFunction GetNodeExecutionFunc) {
	r.getWithChildrenFunction = getWithChildrenFunction
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

func (r *MockNodeExecutionRepo) Exists(ctx context.Context, input interfaces.NodeExecutionResource) (bool, error) {
	if r.existsFunction != nil {
		return r.existsFunction(ctx, input)
	}
	return true, nil
}

func (r *MockNodeExecutionRepo) SetExistsCallback(existsFunction ExistsNodeExecutionFunc) {
	r.existsFunction = existsFunction
}

func (r *MockNodeExecutionRepo) Count(ctx context.Context, input interfaces.CountResourceInput) (int64, error) {
	if r.countFunction != nil {
		return r.countFunction(ctx, input)
	}
	return 0, nil
}

func (r *MockNodeExecutionRepo) SetCountCallback(countFunction CountNodeExecutionFunc) {
	r.countFunction = countFunction
}

func NewMockNodeExecutionRepo() interfaces.NodeExecutionRepoInterface {
	return &MockNodeExecutionRepo{}
}
