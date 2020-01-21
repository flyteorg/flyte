package mocks

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/repositories/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
)

type CreateOrUpdateResourceFunction func(ctx context.Context, input models.Resource) error
type GetResourceFunction func(ctx context.Context, ID interfaces.ResourceID) (
	models.Resource, error)
type DeleteResourceFunction func(ctx context.Context, ID interfaces.ResourceID) error

type MockResourceRepo struct {
	CreateOrUpdateFunction CreateOrUpdateResourceFunction
	GetFunction            GetResourceFunction
	DeleteFunction         DeleteResourceFunction
}

func (r *MockResourceRepo) CreateOrUpdate(ctx context.Context, input models.Resource) error {
	if r.CreateOrUpdateFunction != nil {
		return r.CreateOrUpdateFunction(ctx, input)
	}
	return nil
}

func (r *MockResourceRepo) Get(ctx context.Context, ID interfaces.ResourceID) (
	models.Resource, error) {
	if r.GetFunction != nil {
		return r.GetFunction(ctx, ID)
	}
	return models.Resource{}, nil
}

func (r *MockResourceRepo) GetRaw(ctx context.Context, ID interfaces.ResourceID) (
	models.Resource, error) {
	if r.GetFunction != nil {
		return r.GetFunction(ctx, ID)
	}
	return models.Resource{}, nil
}

func (r *MockResourceRepo) Delete(ctx context.Context, ID interfaces.ResourceID) error {
	if r.DeleteFunction != nil {
		return r.DeleteFunction(ctx, ID)
	}
	return nil
}

func NewMockResourceRepo() interfaces.ResourceRepoInterface {
	return &MockResourceRepo{}
}
