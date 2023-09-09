// Mock implementation of a workflow repo to be used for tests.
package mocks

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

type CreateDescriptionEntityFunc func(input models.DescriptionEntity) error
type GetDescriptionEntityFunc func(input interfaces.GetDescriptionEntityInput) (models.DescriptionEntity, error)
type ListDescriptionEntityFunc func(input interfaces.ListResourceInput) (interfaces.DescriptionEntityCollectionOutput, error)

type MockDescriptionEntityRepo struct {
	createFunction CreateDescriptionEntityFunc
	getFunction    GetDescriptionEntityFunc
	listFunction   ListDescriptionEntityFunc
}

func (r *MockDescriptionEntityRepo) Create(ctx context.Context, DescriptionEntity models.DescriptionEntity) (uint, error) {
	if r.createFunction != nil {
		return 1, r.createFunction(DescriptionEntity)
	}
	return 1, nil
}

func (r *MockDescriptionEntityRepo) Get(
	ctx context.Context, input interfaces.GetDescriptionEntityInput) (models.DescriptionEntity, error) {
	if r.getFunction != nil {
		return r.getFunction(input)
	}
	return models.DescriptionEntity{
		DescriptionEntityKey: models.DescriptionEntityKey{
			ResourceType: input.ResourceType,
			Project:      input.Project,
			Domain:       input.Domain,
			Name:         input.Name,
			Version:      input.Version,
		},
		ShortDescription: "hello world",
	}, nil
}

func (r *MockDescriptionEntityRepo) List(ctx context.Context, input interfaces.ListResourceInput) (interfaces.DescriptionEntityCollectionOutput, error) {
	if r.listFunction != nil {
		return r.listFunction(input)
	}
	return interfaces.DescriptionEntityCollectionOutput{}, nil
}

func (r *MockDescriptionEntityRepo) SetCreateCallback(createFunction CreateDescriptionEntityFunc) {
	r.createFunction = createFunction
}

func (r *MockDescriptionEntityRepo) SetGetCallback(getFunction GetDescriptionEntityFunc) {
	r.getFunction = getFunction
}

func (r *MockDescriptionEntityRepo) SetListCallback(listFunction ListDescriptionEntityFunc) {
	r.listFunction = listFunction
}

func NewMockDescriptionEntityRepo() interfaces.DescriptionEntityRepoInterface {
	return &MockDescriptionEntityRepo{}
}
