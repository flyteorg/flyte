package mocks

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/repositories/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
)

type CreateOrUpdateProjectDomainFunction func(ctx context.Context, input models.ProjectDomain) error
type GetProjectDomainFunction func(ctx context.Context, project, domain string) (models.ProjectDomain, error)
type UpdateProjectDomainFunction func(ctx context.Context, input models.ProjectDomain) error

type MockProjectDomainRepo struct {
	CreateOrUpdateFunction CreateOrUpdateProjectDomainFunction
	GetFunction            GetProjectDomainFunction
}

func (r *MockProjectDomainRepo) CreateOrUpdate(ctx context.Context, input models.ProjectDomain) error {
	if r.CreateOrUpdateFunction != nil {
		return r.CreateOrUpdateFunction(ctx, input)
	}
	return nil
}

func (r *MockProjectDomainRepo) Get(ctx context.Context, project, domain string) (models.ProjectDomain, error) {
	if r.GetFunction != nil {
		return r.GetFunction(ctx, project, domain)
	}
	return models.ProjectDomain{}, nil
}

func NewMockProjectDomainRepo() interfaces.ProjectDomainRepoInterface {
	return &MockProjectDomainRepo{}
}
