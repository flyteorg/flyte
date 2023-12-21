// Mock implementation of a task repo to be used for tests.
package mocks

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

type CreateTaskFunc func(input models.Task, descriptionEntity *models.DescriptionEntity) error
type GetTaskFunc func(id *core.Identifier) (models.Task, error)
type ListTaskFunc func(input interfaces.ListResourceInput) (interfaces.TaskCollectionOutput, error)
type ListTaskIdentifiersFunc func(input interfaces.ListResourceInput) (interfaces.TaskCollectionOutput, error)

type MockTaskRepo struct {
	createFunction            CreateTaskFunc
	getFunction               GetTaskFunc
	listFunction              ListTaskFunc
	listUniqueTaskIdsFunction ListTaskIdentifiersFunc
}

func (r *MockTaskRepo) Create(ctx context.Context, id *core.Identifier, input models.Task, descriptionEntity *models.DescriptionEntity) error {
	if r.createFunction != nil {
		return r.createFunction(input, descriptionEntity)
	}
	return nil
}

func (r *MockTaskRepo) SetCreateCallback(createFunction CreateTaskFunc) {
	r.createFunction = createFunction
}

func (r *MockTaskRepo) Get(ctx context.Context, id *core.Identifier) (models.Task, error) {
	if r.getFunction != nil {
		return r.getFunction(id)
	}
	return models.Task{
		TaskKey: models.TaskKey{
			Project: id.Project,
			Domain:  id.Domain,
			Name:    id.Name,
			Version: id.Version,
		},
	}, nil
}

func (r *MockTaskRepo) SetGetCallback(getFunction GetTaskFunc) {
	r.getFunction = getFunction
}

func (r *MockTaskRepo) List(
	ctx context.Context, input interfaces.ListResourceInput) (interfaces.TaskCollectionOutput, error) {
	if r.listFunction != nil {
		return r.listFunction(input)
	}
	return interfaces.TaskCollectionOutput{}, nil
}

func (r *MockTaskRepo) SetListCallback(listFunction ListTaskFunc) {
	r.listFunction = listFunction
}

func (r *MockTaskRepo) ListTaskIdentifiers(ctx context.Context, input interfaces.ListResourceInput) (
	interfaces.TaskCollectionOutput, error) {

	if r.listUniqueTaskIdsFunction != nil {
		return r.listUniqueTaskIdsFunction(input)
	}
	return interfaces.TaskCollectionOutput{}, nil
}

func (r *MockTaskRepo) SetListTaskIdentifiersCallback(listFunction ListTaskIdentifiersFunc) {
	r.listUniqueTaskIdsFunction = listFunction
}

func NewMockTaskRepo() interfaces.TaskRepoInterface {
	return &MockTaskRepo{}
}
