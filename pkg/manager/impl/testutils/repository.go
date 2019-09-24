package testutils

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/repositories/models"

	"github.com/lyft/flyteadmin/pkg/repositories"
	repositoryMocks "github.com/lyft/flyteadmin/pkg/repositories/mocks"
)

func GetRepoWithDefaultProjectAndErr(err error) repositories.RepositoryInterface {
	repo := repositoryMocks.NewMockRepository()
	repo.ProjectRepo().(*repositoryMocks.MockProjectRepo).GetFunction = func(
		ctx context.Context, projectID string) (models.Project, error) {
		return models.Project{}, err
	}
	return repo
}

func GetRepoWithDefaultProject() repositories.RepositoryInterface {
	repo := repositoryMocks.NewMockRepository()
	repo.ProjectRepo().(*repositoryMocks.MockProjectRepo).GetFunction = func(
		ctx context.Context, projectID string) (models.Project, error) {
		return models.Project{}, nil
	}
	return repo
}
