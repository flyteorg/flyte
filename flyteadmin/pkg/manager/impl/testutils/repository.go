package testutils

import (
	"context"

	repositoryInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	repositoryMocks "github.com/flyteorg/flyteadmin/pkg/repositories/mocks"
)

func GetRepoWithDefaultProjectAndErr(err error) repositoryInterfaces.Repository {
	repo := repositoryMocks.NewMockRepository()
	repo.ProjectRepo().(*repositoryMocks.MockProjectRepo).GetFunction = func(
		ctx context.Context, projectID string) (models.Project, error) {
		activeState := int32(admin.Project_ACTIVE)
		return models.Project{State: &activeState}, err
	}
	return repo
}

func GetRepoWithDefaultProject() repositoryInterfaces.Repository {
	repo := repositoryMocks.NewMockRepository()
	repo.ProjectRepo().(*repositoryMocks.MockProjectRepo).GetFunction = func(
		ctx context.Context, projectID string) (models.Project, error) {
		activeState := int32(admin.Project_ACTIVE)
		return models.Project{State: &activeState}, nil
	}
	return repo
}
