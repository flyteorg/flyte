package mocks

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

type CreateProjectFunction func(ctx context.Context, project models.Project) error
type GetProjectFunction func(ctx context.Context, projectID string) (models.Project, error)
type ListProjectsFunction func(ctx context.Context, input interfaces.ListResourceInput) ([]models.Project, error)
type UpdateProjectFunction func(ctx context.Context, projectUpdate models.Project) error

type MockProjectRepo struct {
	CreateFunction        CreateProjectFunction
	GetFunction           GetProjectFunction
	ListProjectsFunction  ListProjectsFunction
	UpdateProjectFunction UpdateProjectFunction
}

func (r *MockProjectRepo) Create(ctx context.Context, project models.Project) error {
	if r.CreateFunction != nil {
		return r.CreateFunction(ctx, project)
	}
	return nil
}

func (r *MockProjectRepo) Get(ctx context.Context, projectID string) (models.Project, error) {
	if r.GetFunction != nil {
		return r.GetFunction(ctx, projectID)
	}
	activeState := int32(admin.Project_ACTIVE)
	return models.Project{
		Identifier: projectID,
		State:      &activeState,
	}, nil
}

func (r *MockProjectRepo) List(ctx context.Context, input interfaces.ListResourceInput) ([]models.Project, error) {
	if r.ListProjectsFunction != nil {
		return r.ListProjectsFunction(ctx, input)
	}
	return make([]models.Project, 0), nil
}

func (r *MockProjectRepo) UpdateProject(ctx context.Context, projectUpdate models.Project) error {
	if r.UpdateProjectFunction != nil {
		return r.UpdateProjectFunction(ctx, projectUpdate)
	}
	return nil
}

func NewMockProjectRepo() interfaces.ProjectRepoInterface {
	return &MockProjectRepo{}
}
