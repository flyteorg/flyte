package testutils

import (
	"context"

	repositoryInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	repositoryMocks "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

func GetRepoWithDefaultProjectAndErr(err error) repositoryInterfaces.Repository {
	repo := repositoryMocks.NewMockRepository()
	repo.ProjectRepo().(*repositoryMocks.MockProjectRepo).GetFunction = func(
		ctx context.Context, identifier *admin.ProjectIdentifier) (models.Project, error) {
		activeState := int32(admin.Project_ACTIVE)
		return models.Project{State: &activeState}, err
	}
	return repo
}

func GetRepoWithDefaultProject() repositoryInterfaces.Repository {
	repo := repositoryMocks.NewMockRepository()
	repo.ProjectRepo().(*repositoryMocks.MockProjectRepo).GetFunction = func(
		ctx context.Context, identifier *admin.ProjectIdentifier) (models.Project, error) {
		activeState := int32(admin.Project_ACTIVE)
		return models.Project{State: &activeState}, nil
	}
	return repo
}
