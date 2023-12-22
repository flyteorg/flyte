// Mock implementation of a workflow repo to be used for tests.
package mocks

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

type CreateDescriptionEntityFunc func(input models.DescriptionEntity) error
type GetDescriptionEntityFunc func(id *core.Identifier) (models.DescriptionEntity, error)
type ListDescriptionEntityFunc func(entity common.Entity, input interfaces.ListResourceInput) (interfaces.DescriptionEntityCollectionOutput, error)

type MockDescriptionEntityRepo struct {
	createFunction CreateDescriptionEntityFunc
	getFunction    GetDescriptionEntityFunc
	listFunction   ListDescriptionEntityFunc
}

func (r *MockDescriptionEntityRepo) Get(
	ctx context.Context, id *core.Identifier) (models.DescriptionEntity, error) {
	if r.getFunction != nil {
		return r.getFunction(id)
	}
	return models.DescriptionEntity{
		DescriptionEntityKey: models.DescriptionEntityKey{
			ResourceType: id.ResourceType,
			Project:      id.Project,
			Domain:       id.Domain,
			Name:         id.Name,
			Version:      id.Version,
		},
		ShortDescription: "hello world",
	}, nil
}

func (r *MockDescriptionEntityRepo) List(ctx context.Context, entity common.Entity, input interfaces.ListResourceInput) (interfaces.DescriptionEntityCollectionOutput, error) {
	if r.listFunction != nil {
		return r.listFunction(entity, input)
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
