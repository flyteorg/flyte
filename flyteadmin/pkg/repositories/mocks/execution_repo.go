package mocks

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

type CreateExecutionFunc func(ctx context.Context, input models.Execution) error
type UpdateFunc func(ctx context.Context, event models.ExecutionEvent, execution models.Execution) error
type UpdateExecutionFunc func(ctx context.Context, execution models.Execution) error
type GetExecutionFunc func(ctx context.Context, input interfaces.Identifier) (models.Execution, error)
type ListExecutionFunc func(ctx context.Context, input interfaces.ListResourceInput) (
	interfaces.ExecutionCollectionOutput, error)

type MockExecutionRepo struct {
	createFunction      CreateExecutionFunc
	updateFunction      UpdateFunc
	updateExecutionFunc UpdateExecutionFunc
	getFunction         GetExecutionFunc
	listFunction        ListExecutionFunc
	ExistsFunction      func(ctx context.Context, input interfaces.Identifier) (bool, error)
}

func (r *MockExecutionRepo) Create(ctx context.Context, input models.Execution) error {
	if r.createFunction != nil {
		return r.createFunction(ctx, input)
	}
	return nil
}

func (r *MockExecutionRepo) SetCreateCallback(createFunction CreateExecutionFunc) {
	r.createFunction = createFunction
}

func (r *MockExecutionRepo) Update(ctx context.Context, event models.ExecutionEvent, execution models.Execution) error {
	if r.updateFunction != nil {
		return r.updateFunction(ctx, event, execution)
	}
	return nil
}

func (r *MockExecutionRepo) SetUpdateCallback(updateFunction UpdateFunc) {
	r.updateFunction = updateFunction
}

func (r *MockExecutionRepo) UpdateExecution(ctx context.Context, execution models.Execution) error {
	if r.updateExecutionFunc != nil {
		return r.updateExecutionFunc(ctx, execution)
	}
	return nil
}

func (r *MockExecutionRepo) SetUpdateExecutionCallback(updateExecutionFunc UpdateExecutionFunc) {
	r.updateExecutionFunc = updateExecutionFunc
}

func (r *MockExecutionRepo) Get(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
	if r.getFunction != nil {
		return r.getFunction(ctx, input)
	}
	return models.Execution{}, nil
}

func (r *MockExecutionRepo) SetGetCallback(getFunction GetExecutionFunc) {
	r.getFunction = getFunction
}

func (r *MockExecutionRepo) List(ctx context.Context, input interfaces.ListResourceInput) (
	interfaces.ExecutionCollectionOutput, error) {
	if r.listFunction != nil {
		return r.listFunction(ctx, input)
	}
	return interfaces.ExecutionCollectionOutput{}, nil
}

func (r *MockExecutionRepo) SetListCallback(listFunction ListExecutionFunc) {
	r.listFunction = listFunction
}

func (r *MockExecutionRepo) Exists(ctx context.Context, input interfaces.Identifier) (bool, error) {
	if r.ExistsFunction != nil {
		return r.ExistsFunction(ctx, input)
	}
	return true, nil
}

func NewMockExecutionRepo() interfaces.ExecutionRepoInterface {
	return &MockExecutionRepo{}
}
