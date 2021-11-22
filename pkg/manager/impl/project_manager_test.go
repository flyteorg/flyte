package impl

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flyteadmin/pkg/common"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	repositoryMocks "github.com/flyteorg/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	runtimeMocks "github.com/flyteorg/flyteadmin/pkg/runtime/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
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

func testListProjects(request admin.ProjectListRequest, token string, orderExpr string, queryExpr *common.GormQueryExpr, t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	repository.ProjectRepo().(*repositoryMocks.MockProjectRepo).ListProjectsFunction = func(
		ctx context.Context, input interfaces.ListResourceInput) ([]models.Project, error) {
		if len(input.InlineFilters) != 0 {
			q, _ := input.InlineFilters[0].GetGormQueryExpr()
			assert.Equal(t, *queryExpr, q)
		}
		assert.Equal(t, orderExpr, input.SortParameter.GetGormOrderExpr())
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
	resp, err := projectManager.ListProjects(context.Background(), request)
	assert.NoError(t, err)

	assert.Len(t, resp.Projects, 1)
	assert.Equal(t, token, resp.GetToken())
	assert.Len(t, resp.Projects[0].Domains, 4)
	for _, domain := range resp.Projects[0].Domains {
		assert.Contains(t, testDomainsForProjManager, domain.Id)
	}
}

func TestListProjects_NoFilters_LimitOne(t *testing.T) {
	testListProjects(admin.ProjectListRequest{
		Token: "1",
		Limit: 1,
	}, "2", "identifier asc", nil, t)
}

func TestListProjects_HighLimit_SortBy_Filter(t *testing.T) {
	testListProjects(admin.ProjectListRequest{
		Token:   "1",
		Limit:   999,
		Filters: "eq(project.name,foo)",
		SortBy: &admin.Sort{
			Key:       "name",
			Direction: admin.Sort_DESCENDING,
		},
	}, "", "name desc", &common.GormQueryExpr{
		Query: "name = ?",
		Args:  "foo",
	}, t)
}

func TestListProjects_NoToken_NoLimit(t *testing.T) {
	testListProjects(admin.ProjectListRequest{}, "", "identifier asc", nil, t)
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
	assert.EqualError(t, err, "invalid label value [#badlabel]: [a valid label must be an empty string or consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyValue',  or 'my_value',  or '12345', regex used for validation is '(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?')]")
}

func TestProjectManager_UpdateProject(t *testing.T) {
	mockRepository := repositoryMocks.NewMockRepository()
	var updateFuncCalled bool
	labels := admin.Labels{
		Values: map[string]string{
			"foo": "#badlabel",
			"bar": "baz",
		},
	}
	labelsBytes, _ := proto.Marshal(&labels)
	mockRepository.ProjectRepo().(*repositoryMocks.MockProjectRepo).GetFunction = func(
		ctx context.Context, projectID string) (models.Project, error) {

		return models.Project{Identifier: "project-id",
			Name:        "old-project-name",
			Description: "old-project-description", Labels: labelsBytes}, nil
	}
	mockRepository.ProjectRepo().(*repositoryMocks.MockProjectRepo).UpdateProjectFunction = func(
		ctx context.Context, projectUpdate models.Project) error {
		updateFuncCalled = true
		assert.Equal(t, "project-id", projectUpdate.Identifier)
		assert.Equal(t, "new-project-name", projectUpdate.Name)
		assert.Equal(t, "new-project-description", projectUpdate.Description)
		assert.Nil(t, projectUpdate.Labels)
		assert.Equal(t, int32(admin.Project_ACTIVE), *projectUpdate.State)
		return nil
	}
	projectManager := NewProjectManager(mockRepository,
		runtimeMocks.NewMockConfigurationProvider(
			getMockApplicationConfigForProjectManagerTest(), nil, nil, nil, nil, nil))
	_, err := projectManager.UpdateProject(context.Background(), admin.Project{
		Id:          "project-id",
		Name:        "new-project-name",
		Description: "new-project-description",
		State:       admin.Project_ACTIVE,
	})
	assert.Nil(t, err)
	assert.True(t, updateFuncCalled)
}

func TestProjectManager_UpdateProject_ErrorDueToProjectNotFound(t *testing.T) {
	mockRepository := repositoryMocks.NewMockRepository()
	mockRepository.ProjectRepo().(*repositoryMocks.MockProjectRepo).GetFunction = func(
		ctx context.Context, projectID string) (models.Project, error) {
		return models.Project{}, errors.New(projectID + " not found")
	}
	mockRepository.ProjectRepo().(*repositoryMocks.MockProjectRepo).UpdateProjectFunction = func(
		ctx context.Context, projectUpdate models.Project) error {
		assert.Fail(t, "No calls to UpdateProject were expected")
		return nil
	}
	projectManager := NewProjectManager(mockRepository,
		runtimeMocks.NewMockConfigurationProvider(
			getMockApplicationConfigForProjectManagerTest(), nil, nil, nil, nil, nil))
	_, err := projectManager.UpdateProject(context.Background(), admin.Project{
		Id:          "not-found-project-id",
		Name:        "not-found-project-name",
		Description: "not-found-project-description",
	})
	assert.Equal(t, errors.New("not-found-project-id not found"), err)
}

func TestProjectManager_UpdateProject_ErrorDueToInvalidProjectName(t *testing.T) {
	mockRepository := repositoryMocks.NewMockRepository()
	mockRepository.ProjectRepo().(*repositoryMocks.MockProjectRepo).GetFunction = func(
		ctx context.Context, projectID string) (models.Project, error) {
		return models.Project{Identifier: "project-id", Name: "old-project-name", Description: "old-project-description"}, nil
	}
	mockRepository.ProjectRepo().(*repositoryMocks.MockProjectRepo).UpdateProjectFunction = func(
		ctx context.Context, projectUpdate models.Project) error {
		assert.Fail(t, "No calls to UpdateProject were expected")
		return nil
	}
	projectManager := NewProjectManager(mockRepository,
		runtimeMocks.NewMockConfigurationProvider(
			getMockApplicationConfigForProjectManagerTest(), nil, nil, nil, nil, nil))
	_, err := projectManager.UpdateProject(context.Background(), admin.Project{
		Id:   "project-id",
		Name: "longnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamel",
	})
	assert.EqualError(t, err, "project_name cannot exceed 64 characters")
}
