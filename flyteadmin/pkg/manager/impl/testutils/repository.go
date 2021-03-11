package testutils

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/flyteorg/flyteadmin/pkg/repositories"
	repositoryMocks "github.com/flyteorg/flyteadmin/pkg/repositories/mocks"
)

func GetRepoWithDefaultProjectAndErr(err error) repositories.RepositoryInterface {
	repo := repositoryMocks.NewMockRepository()
	repo.ProjectRepo().(*repositoryMocks.MockProjectRepo).GetFunction = func(
		ctx context.Context, projectID string) (models.Project, error) {
		activeState := int32(admin.Project_ACTIVE)
		return models.Project{State: &activeState}, err
	}
	return repo
}

func GetRepoWithDefaultProject() repositories.RepositoryInterface {
	repo := repositoryMocks.NewMockRepository()
	repo.ProjectRepo().(*repositoryMocks.MockProjectRepo).GetFunction = func(
		ctx context.Context, projectID string) (models.Project, error) {
		activeState := int32(admin.Project_ACTIVE)
		return models.Project{State: &activeState}, nil
	}
	return repo
}
