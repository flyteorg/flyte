package impl

import (
	"context"
	"errors"
	"testing"

	"github.com/lyft/flyteadmin/pkg/common"
	"github.com/lyft/flyteadmin/pkg/manager/impl/testutils"
	repositoryMocks "github.com/lyft/flyteadmin/pkg/repositories/mocks"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	runtimeInterfaces "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	runtimeMocks "github.com/lyft/flyteadmin/pkg/runtime/mocks"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
)

var mockProjectConfigProvider = runtimeMocks.NewMockConfigurationProvider(
	testutils.GetApplicationConfigWithDefaultDomains(), nil, nil, nil, nil, nil)

var testDomainsForProjManager = []string{"domain", "development", "staging", "production"}

func getMockApplicationConfigForProjectManagerTest() runtimeInterfaces.ApplicationConfiguration {
	mockApplicationConfig := runtimeMocks.MockApplicationProvider{}
	mockApplicationConfig.SetDomainsConfig(runtimeInterfaces.DomainsConfig{
		{
			ID:   "development",
			Name: "development",
		},
		{
			ID:   "staging",
			Name: "staging",
		},
		{
			ID:   "production",
			Name: "production",
		},
		{
			ID:   "domain",
			Name: "domain",
		},
	})
	return &mockApplicationConfig
}

func TestListProjects(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	repository.ProjectRepo().(*repositoryMocks.MockProjectRepo).ListProjectsFunction = func(
		ctx context.Context, parameter common.SortParameter) ([]models.Project, error) {
		activeState := int32(admin.Project_ACTIVE)
		return []models.Project{
			{
				Identifier:  "project",
				Name:        "project",
				Description: "project_description",
				State:       &activeState,
			},
		}, nil
	}

	projectManager := NewProjectManager(repository, mockProjectConfigProvider)
	resp, err := projectManager.ListProjects(context.Background(), admin.ProjectListRequest{})
	assert.NoError(t, err)

	assert.Len(t, resp.Projects, 1)
	assert.Len(t, resp.Projects[0].Domains, 4)
	for _, domain := range resp.Projects[0].Domains {
		assert.Contains(t, testDomainsForProjManager, domain.Id)
	}
}

func TestProjectManager_CreateProject(t *testing.T) {
	mockRepository := repositoryMocks.NewMockRepository()
	var createFuncCalled bool
	mockRepository.ProjectRepo().(*repositoryMocks.MockProjectRepo).CreateFunction = func(
		ctx context.Context, namespace models.Project) error {
		createFuncCalled = true
		assert.Equal(t, "flyte-project-id", namespace.Identifier)
		assert.Equal(t, "flyte-project-name", namespace.Name)
		assert.Equal(t, "flyte-project-description", namespace.Description)
		return nil
	}
	projectManager := NewProjectManager(mockRepository,
		runtimeMocks.NewMockConfigurationProvider(
			getMockApplicationConfigForProjectManagerTest(), nil, nil, nil, nil, nil))
	_, err := projectManager.CreateProject(context.Background(), admin.ProjectRegisterRequest{
		Project: &admin.Project{
			Id:          "flyte-project-id",
			Name:        "flyte-project-name",
			Description: "flyte-project-description",
		},
	})
	assert.Nil(t, err)
	assert.True(t, createFuncCalled)
}

func TestProjectManager_CreateProjectError(t *testing.T) {
	mockRepository := repositoryMocks.NewMockRepository()
	mockRepository.ProjectRepo().(*repositoryMocks.MockProjectRepo).CreateFunction = func(
		ctx context.Context, namespace models.Project) error {
		return errors.New("uh oh")
	}
	projectManager := NewProjectManager(mockRepository,
		runtimeMocks.NewMockConfigurationProvider(
			getMockApplicationConfigForProjectManagerTest(), nil, nil, nil, nil, nil))
	_, err := projectManager.CreateProject(context.Background(), admin.ProjectRegisterRequest{
		Project: &admin.Project{
			Id:          "flyte-project-id",
			Name:        "flyte-project-name",
			Description: "flyte-project-description",
		},
	})
	assert.EqualError(t, err, "uh oh")

	_, err = projectManager.CreateProject(context.Background(), admin.ProjectRegisterRequest{
		Project: &admin.Project{
			Id:          "flyte-project-id",
			Name:        "flyte-project-name",
			Description: "flyte-project-description",
			Domains: []*admin.Domain{
				{
					Id: "i-shouldn't-be-here",
				},
			},
		},
	})
	assert.EqualError(t, err, "Domains are currently only set system wide. Please retry without domains included in your request.")
}

func TestProjectManager_CreateProjectErrorDueToBadLabels(t *testing.T) {
	mockRepository := repositoryMocks.NewMockRepository()
	mockRepository.ProjectRepo().(*repositoryMocks.MockProjectRepo).CreateFunction = func(
		ctx context.Context, namespace models.Project) error {
		return errors.New("uh oh")
	}
	projectManager := NewProjectManager(mockRepository,
		runtimeMocks.NewMockConfigurationProvider(
			getMockApplicationConfigForProjectManagerTest(), nil, nil, nil, nil, nil))
	_, err := projectManager.CreateProject(context.Background(), admin.ProjectRegisterRequest{
		Project: &admin.Project{
			Id:          "flyte-project-id",
			Name:        "flyte-project-name",
			Description: "flyte-project-description",
		},
	})
	assert.EqualError(t, err, "uh oh")

	_, err = projectManager.CreateProject(context.Background(), admin.ProjectRegisterRequest{
		Project: &admin.Project{
			Id:          "flyte-project-id",
			Name:        "flyte-project-name",
			Description: "flyte-project-description",
			Labels: &admin.Labels{
				Values: map[string]string{
					"foo": "#badlabel",
					"bar": "baz",
				},
			},
		},
	})
	assert.EqualError(t, err, "invalid label value [#badlabel]: [a DNS-1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character (e.g. 'my-name',  or '123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?')]")
}
