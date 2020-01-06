package mocks

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/repositories/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
)

type CreateOrUpdateProjectAttributesFunction func(ctx context.Context, input models.ProjectAttributes) error
type GetProjectAttributesFunction func(ctx context.Context, project, resource string) (models.ProjectAttributes, error)
type DeleteProjectAttributesFunction func(ctx context.Context, project, resource string) error

type MockProjectAttributesRepo struct {
	CreateOrUpdateFunction CreateOrUpdateProjectAttributesFunction
	GetFunction            GetProjectAttributesFunction
	DeleteFunction         DeleteProjectAttributesFunction
}

func (r *MockProjectAttributesRepo) CreateOrUpdate(ctx context.Context, input models.ProjectAttributes) error {
	if r.CreateOrUpdateFunction != nil {
		return r.CreateOrUpdateFunction(ctx, input)
	}
	return nil
}

func (r *MockProjectAttributesRepo) Get(ctx context.Context, project, resource string) (
	models.ProjectAttributes, error) {
	if r.GetFunction != nil {
		return r.GetFunction(ctx, project, resource)
	}
	return models.ProjectAttributes{}, nil
}

func (r *MockProjectAttributesRepo) Delete(ctx context.Context, project, resource string) error {
	if r.DeleteFunction != nil {
		return r.DeleteFunction(ctx, project, resource)
	}
	return nil
}

func NewMockProjectAttributesRepo() interfaces.ProjectAttributesRepoInterface {
	return &MockProjectAttributesRepo{}
}
