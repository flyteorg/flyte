package mocks

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

type CreateTaskExecutionFunc func(ctx context.Context, input models.TaskExecution) error
type GetTaskExecutionFunc func(ctx context.Context, id *core.TaskExecutionIdentifier) (models.TaskExecution, error)
type UpdateTaskExecutionFunc func(ctx context.Context, execution models.TaskExecution) error
type ListTaskExecutionFunc func(ctx context.Context, input interfaces.ListResourceInput) (interfaces.TaskExecutionCollectionOutput, error)
type CountTaskExecutionFunc func(ctx context.Context, input interfaces.CountResourceInput) (int64, error)

type MockTaskExecutionRepo struct {
	createFunction CreateTaskExecutionFunc
	getFunction    GetTaskExecutionFunc
	updateFunction UpdateTaskExecutionFunc
	listFunction   ListTaskExecutionFunc
	countFunction  CountTaskExecutionFunc
}

func (r *MockTaskExecutionRepo) Create(ctx context.Context, id *core.TaskExecutionIdentifier, input models.TaskExecution) error {
	if r.createFunction != nil {
		return r.createFunction(ctx, input)
	}
	return nil
}

func (r *MockTaskExecutionRepo) SetCreateCallback(createFunction CreateTaskExecutionFunc) {
	r.createFunction = createFunction
}

func (r *MockTaskExecutionRepo) Get(ctx context.Context, id *core.TaskExecutionIdentifier) (models.TaskExecution, error) {
	if r.getFunction != nil {
		return r.getFunction(ctx, id)
	}
	return models.TaskExecution{}, nil
}

func (r *MockTaskExecutionRepo) SetGetCallback(getFunction GetTaskExecutionFunc) {
	r.getFunction = getFunction
}

func (r *MockTaskExecutionRepo) Update(ctx context.Context, id *core.TaskExecutionIdentifier, execution models.TaskExecution) error {
	if r.updateFunction != nil {
		return r.updateFunction(ctx, execution)
	}
	return nil
}

func (r *MockTaskExecutionRepo) SetUpdateCallback(updateFunction UpdateTaskExecutionFunc) {
	r.updateFunction = updateFunction
}

func (r *MockTaskExecutionRepo) List(ctx context.Context, input interfaces.ListResourceInput) (interfaces.TaskExecutionCollectionOutput, error) {
	if r.listFunction != nil {
		return r.listFunction(ctx, input)
	}
	return interfaces.TaskExecutionCollectionOutput{}, nil
}

func (r *MockTaskExecutionRepo) SetListCallback(listFunction ListTaskExecutionFunc) {
	r.listFunction = listFunction
}

func (r *MockTaskExecutionRepo) Count(ctx context.Context, input interfaces.CountResourceInput) (int64, error) {
	if r.countFunction != nil {
		return r.countFunction(ctx, input)
	}
	return 0, nil
}

func (r *MockTaskExecutionRepo) SetCountCallback(countFunction CountTaskExecutionFunc) {
	r.countFunction = countFunction
}

func NewMockTaskExecutionRepo() interfaces.TaskExecutionRepoInterface {
	return &MockTaskExecutionRepo{}
}
