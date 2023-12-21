// Mock implementation of a workflow repo to be used for tests.
package mocks

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

type GetNamedEntityFunc func(id *admin.NamedEntityIdentifier, resourceType core.ResourceType) (models.NamedEntity, error)
type ListNamedEntityFunc func(input interfaces.ListNamedEntityInput) (interfaces.NamedEntityCollectionOutput, error)
type UpdateNamedEntityFunc func(input models.NamedEntity) error

type MockNamedEntityRepo struct {
	getFunction    GetNamedEntityFunc
	listFunction   ListNamedEntityFunc
	updateFunction UpdateNamedEntityFunc
}

func (r *MockNamedEntityRepo) Update(ctx context.Context, id *admin.NamedEntityIdentifier, NamedEntity models.NamedEntity) error {
	if r.updateFunction != nil {
		return r.updateFunction(NamedEntity)
	}
	return nil
}

func (r *MockNamedEntityRepo) Get(
	ctx context.Context, id *admin.NamedEntityIdentifier, resourceType core.ResourceType) (models.NamedEntity, error) {
	if r.getFunction != nil {
		return r.getFunction(id, resourceType)
	}
	return models.NamedEntity{
		NamedEntityKey: models.NamedEntityKey{
			ResourceType: resourceType,
			Project:      id.Project,
			Domain:       id.Domain,
			Name:         id.Name,
		},
		NamedEntityMetadataFields: models.NamedEntityMetadataFields{
			Description: "",
		},
	}, nil
}

func (r *MockNamedEntityRepo) List(ctx context.Context, input interfaces.ListNamedEntityInput) (interfaces.NamedEntityCollectionOutput, error) {
	if r.listFunction != nil {
		return r.listFunction(input)
	}
	return interfaces.NamedEntityCollectionOutput{}, nil
}

func (r *MockNamedEntityRepo) SetGetCallback(getFunction GetNamedEntityFunc) {
	r.getFunction = getFunction
}

func (r *MockNamedEntityRepo) SetListCallback(listFunction ListNamedEntityFunc) {
	r.listFunction = listFunction
}

func (r *MockNamedEntityRepo) SetUpdateCallback(updateFunction UpdateNamedEntityFunc) {
	r.updateFunction = updateFunction
}

func NewMockNamedEntityRepo() interfaces.NamedEntityRepoInterface {
	return &MockNamedEntityRepo{}
}
