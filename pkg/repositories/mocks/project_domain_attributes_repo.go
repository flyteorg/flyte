package mocks

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/repositories/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
)

type CreateOrUpdateProjectDomainAttributesFunction func(ctx context.Context, input models.ProjectDomainAttributes) error
type GetProjectDomainAttributesFunction func(ctx context.Context, project, domain, resource string) (models.ProjectDomainAttributes, error)
type DeleteProjectDomainAttributesFunction func(ctx context.Context, project, domain, resource string) error

type MockProjectDomainAttributesRepo struct {
	CreateOrUpdateFunction CreateOrUpdateProjectDomainAttributesFunction
	GetFunction            GetProjectDomainAttributesFunction
	DeleteFunction         DeleteProjectDomainAttributesFunction
}

func (r *MockProjectDomainAttributesRepo) CreateOrUpdate(ctx context.Context, input models.ProjectDomainAttributes) error {
	if r.CreateOrUpdateFunction != nil {
		return r.CreateOrUpdateFunction(ctx, input)
	}
	return nil
}

func (r *MockProjectDomainAttributesRepo) Get(ctx context.Context, project, domain, resource string) (
	models.ProjectDomainAttributes, error) {
	if r.GetFunction != nil {
		return r.GetFunction(ctx, project, domain, resource)
	}
	return models.ProjectDomainAttributes{}, nil
}

func (r *MockProjectDomainAttributesRepo) Delete(ctx context.Context, project, domain, resource string) error {
	if r.DeleteFunction != nil {
		return r.DeleteFunction(ctx, project, domain, resource)
	}
	return nil
}

func NewMockProjectDomainAttributesRepo() interfaces.ProjectDomainAttributesRepoInterface {
	return &MockProjectDomainAttributesRepo{}
}
