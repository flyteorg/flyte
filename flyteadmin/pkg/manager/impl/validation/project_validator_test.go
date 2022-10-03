package validation

import (
	"context"
	"errors"
	"strconv"
	"testing"

	flyteAdminErrors "github.com/flyteorg/flyteadmin/pkg/errors"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"
	repositoryMocks "github.com/flyteorg/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
)

func TestValidateProjectRegisterRequest_ValidRequest(t *testing.T) {
	assert.Nil(t, ValidateProjectRegisterRequest(admin.ProjectRegisterRequest{
		Project: &admin.Project{
			Id:   "proj",
			Name: "proj",
		},
	}))
}

func TestValidateProjectRegisterRequest(t *testing.T) {
	type testValue struct {
		request       admin.ProjectRegisterRequest
		expectedError string
	}
	testValues := []testValue{
		{
			request:       admin.ProjectRegisterRequest{},
			expectedError: "missing project",
		},
		{
			request: admin.ProjectRegisterRequest{
				Project: &admin.Project{
					Name: "proj",
					Domains: []*admin.Domain{
						{
							Id:   "foo",
							Name: "foo",
						},
					},
				},
			},
			expectedError: "missing project_id",
		},
		{
			request: admin.ProjectRegisterRequest{
				Project: &admin.Project{
					Id:   "%)(*&",
					Name: "proj",
				},
			},
			expectedError: "invalid project id [%)(*&]: [a lowercase RFC 1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character (e.g. 'my-name',  or '123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?')]",
		},
		{
			request: admin.ProjectRegisterRequest{
				Project: &admin.Project{
					Id: "proj",
				},
			},
			expectedError: "missing project_name",
		},
		{
			request: admin.ProjectRegisterRequest{
				Project: &admin.Project{
					Id:   "proj",
					Name: "longnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamel",
				},
			},
			expectedError: "project_name cannot exceed 64 characters",
		},
		{
			request: admin.ProjectRegisterRequest{
				Project: &admin.Project{
					Id:   "proj",
					Name: "proj",
					Domains: []*admin.Domain{
						{
							Id:   "foo",
							Name: "foo",
						},
						{
							Id: "foo",
						},
					},
				},
			},
			expectedError: "Domains are currently only set system wide. Please retry without domains included in your request.",
		},
		{
			request: admin.ProjectRegisterRequest{
				Project: &admin.Project{
					Id:   "proj",
					Name: "name",
					// 301 character string
					Description: "longnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongn",
				},
			},
			expectedError: "project_description cannot exceed 300 characters",
		},
		{
			request: admin.ProjectRegisterRequest{
				Project: &admin.Project{
					Id:   "proj",
					Name: "name",
					Labels: &admin.Labels{
						Values: map[string]string{
							"#badkey": "foo",
							"bar":     "baz",
						},
					},
				},
			},
			expectedError: "invalid label key [#badkey]: [name part must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyName',  or 'my.name',  or '123-abc', regex used for validation is '([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]')]",
		},
		{
			request: admin.ProjectRegisterRequest{
				Project: &admin.Project{
					Id:   "proj",
					Name: "name",
					Labels: &admin.Labels{
						Values: map[string]string{
							"foo": ".bad-label-value",
							"bar": "baz",
						},
					},
				},
			},
			expectedError: "invalid label value [.bad-label-value]: [a valid label must be an empty string or consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyValue',  or 'my_value',  or '12345', regex used for validation is '(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?')]",
		},
	}

	for _, val := range testValues {
		t.Run(val.expectedError, func(t *testing.T) {
			assert.EqualError(t, ValidateProjectRegisterRequest(val.request), val.expectedError)
		})
	}
}

func TestValidateProject_ValidProject(t *testing.T) {
	assert.Nil(t, ValidateProject(admin.Project{
		Id:          "proj",
		Description: "An amazing description for this project",
		State:       admin.Project_ARCHIVED,
		Labels: &admin.Labels{
			Values: map[string]string{
				"foo":                "bar",
				"example.com/my-key": "my-value",
			},
		},
	}))
}

func TestValidateProject(t *testing.T) {
	type testValue struct {
		project       admin.Project
		expectedError string
	}
	testValues := []testValue{
		{
			project: admin.Project{
				Name: "proj",
				Domains: []*admin.Domain{
					{
						Id:   "foo",
						Name: "foo",
					},
				},
			},
			expectedError: "missing project_id",
		},
		{
			project: admin.Project{
				Id:   "%)(*&",
				Name: "proj",
			},
			expectedError: "invalid project id [%)(*&]: [a lowercase RFC 1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character (e.g. 'my-name',  or '123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?')]",
		},
		{
			project: admin.Project{
				Id:   "proj",
				Name: "longnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamel",
			},
			expectedError: "project_name cannot exceed 64 characters",
		},
		{
			project: admin.Project{
				Id:   "proj",
				Name: "proj",
				Domains: []*admin.Domain{
					{
						Id:   "foo",
						Name: "foo",
					},
					{
						Id: "foo",
					},
				},
			},
			expectedError: "Domains are currently only set system wide. Please retry without domains included in your request.",
		},
		{
			project: admin.Project{
				Id:   "proj",
				Name: "name",
				// 301 character string
				Description: "longnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongnamelongn",
			},
			expectedError: "project_description cannot exceed 300 characters",
		},
		{
			project: admin.Project{
				Id:   "proj",
				Name: "name",
				Labels: &admin.Labels{
					Values: createLabelsMap(17),
				},
			},
			expectedError: "labels map cannot exceed 16 entries",
		},
		{
			project: admin.Project{
				Id:   "proj",
				Name: "name",
				Labels: &admin.Labels{
					Values: map[string]string{
						"#badkey": "foo",
						"bar":     "baz",
					},
				},
			},
			expectedError: "invalid label key [#badkey]: [name part must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyName',  or 'my.name',  or '123-abc', regex used for validation is '([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]')]",
		},
	}

	for _, val := range testValues {
		t.Run(val.expectedError, func(t *testing.T) {
			assert.EqualError(t, ValidateProject(val.project), val.expectedError)
		})
	}
}

func createLabelsMap(size int) map[string]string {
	result := make(map[string]string, size)
	for i := 0; i < size; i++ {
		result["key-"+strconv.Itoa(i)] = "value"
	}
	return result
}

func TestValidateProjectAndDomain(t *testing.T) {
	mockRepo := repositoryMocks.NewMockRepository()
	mockRepo.ProjectRepo().(*repositoryMocks.MockProjectRepo).GetFunction = func(
		ctx context.Context, projectID string) (models.Project, error) {
		assert.Equal(t, projectID, "flyte-project-id")
		activeState := int32(admin.Project_ACTIVE)
		return models.Project{State: &activeState}, nil
	}
	err := ValidateProjectAndDomain(context.Background(), mockRepo, testutils.GetApplicationConfigWithDefaultDomains(),
		"flyte-project-id", "domain")
	assert.Nil(t, err)
}

func TestValidateProjectAndDomainArchivedProject(t *testing.T) {
	mockRepo := repositoryMocks.NewMockRepository()
	mockRepo.ProjectRepo().(*repositoryMocks.MockProjectRepo).GetFunction = func(
		ctx context.Context, projectID string) (models.Project, error) {
		archivedState := int32(admin.Project_ARCHIVED)
		return models.Project{State: &archivedState}, nil
	}

	err := ValidateProjectAndDomain(context.Background(), mockRepo, testutils.GetApplicationConfigWithDefaultDomains(),
		"flyte-project-id", "domain")
	assert.EqualError(t, err,
		"project [flyte-project-id] is not active")
}

func TestValidateProjectAndDomainError(t *testing.T) {
	mockRepo := repositoryMocks.NewMockRepository()
	mockRepo.ProjectRepo().(*repositoryMocks.MockProjectRepo).GetFunction = func(
		ctx context.Context, projectID string) (models.Project, error) {
		return models.Project{}, errors.New("foo")
	}

	err := ValidateProjectAndDomain(context.Background(), mockRepo, testutils.GetApplicationConfigWithDefaultDomains(),
		"flyte-project-id", "domain")
	assert.EqualError(t, err,
		"failed to validate that project [flyte-project-id] and domain [domain] are registered, err: [foo]")
}

func TestValidateProjectAndDomainNotFound(t *testing.T) {
	mockRepo := repositoryMocks.NewMockRepository()
	mockRepo.ProjectRepo().(*repositoryMocks.MockProjectRepo).GetFunction = func(
		ctx context.Context, projectID string) (models.Project, error) {
		return models.Project{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound, "project [%s] not found", projectID)
	}
	err := ValidateProjectAndDomain(context.Background(), mockRepo, testutils.GetApplicationConfigWithDefaultDomains(),
		"flyte-project", "domain")
	assert.EqualError(t, err, "failed to validate that project [flyte-project] and domain [domain] are registered, err: [project [flyte-project] not found]")
}

func TestValidateProjectDb(t *testing.T) {
	mockRepo := repositoryMocks.NewMockRepository()
	t.Run("base case", func(t *testing.T) {
		mockRepo.ProjectRepo().(*repositoryMocks.MockProjectRepo).GetFunction = func(
			ctx context.Context, projectID string) (models.Project, error) {
			assert.Equal(t, projectID, "flyte-project-id")
			activeState := int32(admin.Project_ACTIVE)
			return models.Project{State: &activeState}, nil
		}
		err := ValidateProjectForUpdate(context.Background(), mockRepo, "flyte-project-id")

		assert.Nil(t, err)
	})

	t.Run("error getting", func(t *testing.T) {
		mockRepo.ProjectRepo().(*repositoryMocks.MockProjectRepo).GetFunction = func(
			ctx context.Context, projectID string) (models.Project, error) {

			return models.Project{}, errors.New("missing")
		}
		err := ValidateProjectForUpdate(context.Background(), mockRepo, "flyte-project-id")
		assert.Error(t, err)
	})

	t.Run("error archived", func(t *testing.T) {
		mockRepo.ProjectRepo().(*repositoryMocks.MockProjectRepo).GetFunction = func(
			ctx context.Context, projectID string) (models.Project, error) {
			state := int32(admin.Project_ARCHIVED)
			return models.Project{State: &state}, nil
		}
		err := ValidateProjectForUpdate(context.Background(), mockRepo, "flyte-project-id")
		assert.Error(t, err)
	})
}

func TestValidateProjectExistsDb(t *testing.T) {
	mockRepo := repositoryMocks.NewMockRepository()
	t.Run("base case", func(t *testing.T) {
		mockRepo.ProjectRepo().(*repositoryMocks.MockProjectRepo).GetFunction = func(
			ctx context.Context, projectID string) (models.Project, error) {
			assert.Equal(t, projectID, "flyte-project-id")
			activeState := int32(admin.Project_ACTIVE)
			return models.Project{State: &activeState}, nil
		}
		err := ValidateProjectExists(context.Background(), mockRepo, "flyte-project-id")

		assert.Nil(t, err)
	})

	t.Run("error getting", func(t *testing.T) {
		mockRepo.ProjectRepo().(*repositoryMocks.MockProjectRepo).GetFunction = func(
			ctx context.Context, projectID string) (models.Project, error) {

			return models.Project{}, errors.New("missing")
		}
		err := ValidateProjectExists(context.Background(), mockRepo, "flyte-project-id")
		assert.Error(t, err)
	})
}
