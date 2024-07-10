package mocks

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
)

type CreateExecutionFunc func(ctx context.Context, input models.Execution) error
type UpdateExecutionFunc func(ctx context.Context, execution models.Execution) error
type GetExecutionFunc func(ctx context.Context, input interfaces.Identifier) (models.Execution, error)
type ListExecutionFunc func(ctx context.Context, input interfaces.ListResourceInput) (
	interfaces.ExecutionCollectionOutput, error)
type CountExecutionFunc func(ctx context.Context, input interfaces.CountResourceInput) (int64, error)
type CountByPhaseExecutionFunc func(ctx context.Context, input interfaces.CountResourceInput) (interfaces.ExecutionCountsByPhaseOutput, error)

type MockExecutionRepo struct {
	createFunction       CreateExecutionFunc
	updateFunction       UpdateExecutionFunc
	getFunction          GetExecutionFunc
	listFunction         ListExecutionFunc
	countFunction        CountExecutionFunc
	countByPhaseFunction CountByPhaseExecutionFunc
}

func (r *MockExecutionRepo) Create(ctx context.Context, input models.Execution, _ []*models.ExecutionTag) error {
	if r.createFunction != nil {
		return r.createFunction(ctx, input)
	}
	return nil
}

func (r *MockExecutionRepo) SetCreateCallback(createFunction CreateExecutionFunc) {
	r.createFunction = createFunction
}

func (r *MockExecutionRepo) Update(ctx context.Context, execution models.Execution) error {
	if r.updateFunction != nil {
		return r.updateFunction(ctx, execution)
	}
	return nil
}

func (r *MockExecutionRepo) SetUpdateCallback(updateFunction UpdateExecutionFunc) {
	r.updateFunction = updateFunction
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

func (r *MockExecutionRepo) Count(ctx context.Context, input interfaces.CountResourceInput) (int64, error) {
	if r.countFunction != nil {
		return r.countFunction(ctx, input)
	}
	return 0, nil
}

func (r *MockExecutionRepo) SetCountCallback(countFunction CountExecutionFunc) {
	r.countFunction = countFunction
}

func (r *MockExecutionRepo) CountByPhase(ctx context.Context, input interfaces.CountResourceInput) (interfaces.ExecutionCountsByPhaseOutput, error) {
	if r.countByPhaseFunction != nil {
		return r.countByPhaseFunction(ctx, input)
	}
	return interfaces.ExecutionCountsByPhaseOutput{}, nil
}

func (r *MockExecutionRepo) SetCountByPhaseCallback(countByPhaseFunction CountByPhaseExecutionFunc) {
	r.countByPhaseFunction = countByPhaseFunction
}

func (r *MockExecutionRepo) FindFirstStatusUpdatesCheckpoint(ctx context.Context, cluster string) (uint, error) {
	return 0, nil
}

func (r *MockExecutionRepo) FindNextStatusUpdatesCheckpoint(ctx context.Context, cluster string, checkpoint uint) (uint, error) {
	return 0, nil
}

func (r *MockExecutionRepo) FindStatusUpdates(ctx context.Context, cluster string, checkpoint uint, limit, offset int) ([]interfaces.ExecutionStatus, error) {
	return nil, nil
}

func NewMockExecutionRepo() interfaces.ExecutionRepoInterface {
	return &MockExecutionRepo{}
}
