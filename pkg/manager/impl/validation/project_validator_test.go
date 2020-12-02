package validation

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/lyft/flyteadmin/pkg/manager/impl/testutils"
	repositoryMocks "github.com/lyft/flyteadmin/pkg/repositories/mocks"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
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
			expectedError: "invalid project id [%)(*&]: [a DNS-1123 label must consist of lower case alphanumeric " +
				"characters or '-', and must start and end with an alphanumeric character (e.g. 'my-name',  or " +
				"'123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?')]",
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
			expectedError: "invalid label key [#badkey]: [a DNS-1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character (e.g. 'my-name',  or '123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?')]",
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
			expectedError: "invalid label value [.bad-label-value]: [a DNS-1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character (e.g. 'my-name',  or '123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?')]",
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
				"foo": "bar",
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
			expectedError: "invalid project id [%)(*&]: [a DNS-1123 label must consist of lower case alphanumeric " +
				"characters or '-', and must start and end with an alphanumeric character (e.g. 'my-name',  or " +
				"'123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?')]",
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
			expectedError: "invalid label key [#badkey]: [a DNS-1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character (e.g. 'my-name',  or '123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?')]",
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
