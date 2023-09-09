package impl

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/common"
	adminErrors "github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	repositoryMocks "github.com/flyteorg/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	runtimeMocks "github.com/flyteorg/flyteadmin/pkg/runtime/mocks"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"github.com/golang/protobuf/proto"

	workflowengine "github.com/flyteorg/flyteadmin/pkg/workflowengine/interfaces"
	workflowMocks "github.com/flyteorg/flyteadmin/pkg/workflowengine/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	mockScope "github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

// Static values for test
const projectValue = "foo"
const domainValue = "bar"
const nameValue = "baz"
const limit = 100

func getMockConfigForTaskTest() runtimeInterfaces.Configuration {
	mockConfig := runtimeMocks.NewMockConfigurationProvider(
		testutils.GetApplicationConfigWithDefaultDomains(), nil, nil, runtimeMocks.NewMockTaskResourceConfiguration(
			runtimeInterfaces.TaskResourceSet{}, runtimeInterfaces.TaskResourceSet{}), runtimeMocks.NewMockWhitelistConfiguration(), nil)
	return mockConfig
}

var taskIdentifier = core.Identifier{
	ResourceType: core.ResourceType_TASK,
	Project:      "project",
	Domain:       "domain",
	Name:         "name",
	Version:      "version",
}

func init() {
	labeled.SetMetricKeys(common.RuntimeTypeKey, common.RuntimeVersionKey, common.ErrorKindKey)
}

func getMockTaskCompiler() workflowengine.Compiler {
	mockCompiler := workflowMocks.NewMockCompiler()
	mockCompiler.(*workflowMocks.MockCompiler).AddCompileTaskCallback(
		func(task *core.TaskTemplate) (*core.CompiledTask, error) {
			return &core.CompiledTask{
				Template: task,
			}, nil
		})
	return mockCompiler
}

func getMockTaskRepository() interfaces.Repository {
	return repositoryMocks.NewMockRepository()
}

func TestCreateTask(t *testing.T) {
	mockRepository := getMockTaskRepository()
	mockRepository.TaskRepo().(*repositoryMocks.MockTaskRepo).SetGetCallback(
		func(input interfaces.Identifier) (models.Task, error) {
			return models.Task{}, errors.New("foo")
		})
	var createCalled bool
	mockRepository.TaskRepo().(*repositoryMocks.MockTaskRepo).SetCreateCallback(func(input models.Task, descriptionEntity *models.DescriptionEntity) error {
		assert.Equal(t, []byte{
			0xbf, 0x79, 0x61, 0x1c, 0xf5, 0xc1, 0xfb, 0x4c, 0xf8, 0xf4, 0xc4, 0x53, 0x5f, 0x8f, 0x73, 0xe2, 0x26, 0x5a,
			0x18, 0x4a, 0xb7, 0x66, 0x98, 0x3c, 0xab, 0x2, 0x6c, 0x9, 0x9b, 0x90, 0xec, 0x8f}, input.Digest)
		createCalled = true
		return nil
	})
	mockRepository.DescriptionEntityRepo().(*repositoryMocks.MockDescriptionEntityRepo).SetGetCallback(
		func(input interfaces.GetDescriptionEntityInput) (models.DescriptionEntity, error) {
			return models.DescriptionEntity{}, adminErrors.NewFlyteAdminErrorf(codes.NotFound, "NotFound")
		})
	taskManager := NewTaskManager(mockRepository, getMockConfigForTaskTest(), getMockTaskCompiler(),
		mockScope.NewTestScope())
	request := testutils.GetValidTaskRequest()
	response, err := taskManager.CreateTask(context.Background(), request)
	assert.NoError(t, err)

	assert.Equal(t, &admin.TaskCreateResponse{}, response)
	assert.True(t, createCalled)

	request.Spec.Description = nil
	response, err = taskManager.CreateTask(context.Background(), request)
	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestCreateTask_ValidationError(t *testing.T) {
	mockRepository := getMockTaskRepository()
	taskManager := NewTaskManager(mockRepository, getMockConfigForTaskTest(), getMockTaskCompiler(),
		mockScope.NewTestScope())
	request := testutils.GetValidTaskRequest()
	request.Id = nil
	response, err := taskManager.CreateTask(context.Background(), request)
	assert.EqualError(t, err, "missing id")
	assert.Nil(t, response)
}

func TestCreateTask_CompilerError(t *testing.T) {
	mockCompiler := workflowMocks.NewMockCompiler()
	expectedErr := errors.New("expected error")
	mockCompiler.(*workflowMocks.MockCompiler).AddCompileTaskCallback(
		func(task *core.TaskTemplate) (*core.CompiledTask, error) {
			return nil, expectedErr
		})
	mockRepository := getMockTaskRepository()
	taskManager := NewTaskManager(mockRepository, getMockConfigForTaskTest(), mockCompiler,
		mockScope.NewTestScope())
	request := testutils.GetValidTaskRequest()
	response, err := taskManager.CreateTask(context.Background(), request)
	assert.EqualError(t, err, expectedErr.Error())
	assert.Nil(t, response)
}

func TestCreateTask_DatabaseError(t *testing.T) {
	repository := getMockTaskRepository()
	repository.TaskRepo().(*repositoryMocks.MockTaskRepo).SetGetCallback(
		func(input interfaces.Identifier) (models.Task, error) {
			return models.Task{}, errors.New("foo")
		})
	expectedErr := errors.New("expected error")
	taskCreateFunc := func(input models.Task, descriptionEntity *models.DescriptionEntity) error {
		return expectedErr
	}

	repository.TaskRepo().(*repositoryMocks.MockTaskRepo).SetCreateCallback(taskCreateFunc)
	taskManager := NewTaskManager(repository, getMockConfigForTaskTest(), getMockTaskCompiler(), mockScope.NewTestScope())
	request := testutils.GetValidTaskRequest()
	response, err := taskManager.CreateTask(context.Background(), request)
	assert.EqualError(t, err, expectedErr.Error())
	assert.Nil(t, response)
}

func TestGetTask(t *testing.T) {
	repository := getMockTaskRepository()
	taskGetFunc := func(input interfaces.Identifier) (models.Task, error) {
		assert.Equal(t, "project", input.Project)
		assert.Equal(t, "domain", input.Domain)
		assert.Equal(t, "name", input.Name)
		assert.Equal(t, "version", input.Version)
		return models.Task{
			BaseModel: models.BaseModel{
				CreatedAt: testutils.MockCreatedAtValue,
			},
			TaskKey: models.TaskKey{
				Project: input.Project,
				Domain:  input.Domain,
				Name:    input.Name,
				Version: input.Version,
			},
			Closure: testutils.GetTaskClosureBytes(),
		}, nil
	}
	repository.TaskRepo().(*repositoryMocks.MockTaskRepo).SetGetCallback(taskGetFunc)
	taskManager := NewTaskManager(repository, getMockConfigForTaskTest(), getMockTaskCompiler(), mockScope.NewTestScope())

	task, err := taskManager.GetTask(context.Background(), admin.ObjectGetRequest{
		Id: &taskIdentifier,
	})
	assert.NoError(t, err)
	assert.Equal(t, "project", task.Id.Project)
	assert.Equal(t, "domain", task.Id.Domain)
	assert.Equal(t, "name", task.Id.Name)
	assert.Equal(t, "version", task.Id.Version)
	assert.True(t, proto.Equal(testutils.GetTaskClosure(), task.Closure))
}

func TestGetTask_DatabaseError(t *testing.T) {
	repository := getMockTaskRepository()
	expectedErr := errors.New("expected error")
	taskGetFunc := func(input interfaces.Identifier) (models.Task, error) {
		return models.Task{}, expectedErr
	}
	repository.TaskRepo().(*repositoryMocks.MockTaskRepo).SetGetCallback(taskGetFunc)
	taskManager := NewTaskManager(repository, getMockConfigForTaskTest(), getMockTaskCompiler(), mockScope.NewTestScope())
	task, err := taskManager.GetTask(context.Background(), admin.ObjectGetRequest{
		Id: &taskIdentifier,
	})
	assert.Nil(t, task)
	assert.EqualError(t, err, expectedErr.Error())
}

func TestGetTask_TransformerError(t *testing.T) {
	repository := getMockTaskRepository()
	taskGetFunc := func(input interfaces.Identifier) (models.Task, error) {
		assert.Equal(t, "project", input.Project)
		assert.Equal(t, "domain", input.Domain)
		assert.Equal(t, "name", input.Name)
		assert.Equal(t, "version", input.Version)
		return models.Task{
			TaskKey: models.TaskKey{
				Project: input.Project,
				Domain:  input.Domain,
				Name:    input.Name,
				Version: input.Version,
			},
			Closure: []byte("invalid task spec"),
		}, nil
	}
	repository.TaskRepo().(*repositoryMocks.MockTaskRepo).SetGetCallback(taskGetFunc)
	taskManager := NewTaskManager(repository, getMockConfigForTaskTest(), getMockTaskCompiler(), mockScope.NewTestScope())

	task, err := taskManager.GetTask(context.Background(), admin.ObjectGetRequest{
		Id: &taskIdentifier,
	})
	assert.Nil(t, task)
	assert.Equal(t, codes.Internal, err.(adminErrors.FlyteAdminError).Code())
}

func TestListTasks(t *testing.T) {
	repository := getMockTaskRepository()
	taskListFunc := func(input interfaces.ListResourceInput) (interfaces.TaskCollectionOutput, error) {
		var projectFilter, domainFilter, nameFilter bool
		for _, filter := range input.InlineFilters {
			assert.Equal(t, common.Task, filter.GetEntity())
			queryExpr, _ := filter.GetGormQueryExpr()
			if queryExpr.Args == projectValue && queryExpr.Query == testutils.ProjectQueryPattern {
				projectFilter = true
			}
			if queryExpr.Args == domainValue && queryExpr.Query == testutils.DomainQueryPattern {
				domainFilter = true
			}
			if queryExpr.Args == nameValue && queryExpr.Query == testutils.NameQueryPattern {
				nameFilter = true
			}
		}
		assert.True(t, projectFilter, "Missing project equality filter")
		assert.True(t, domainFilter, "Missing domain equality filter")
		assert.True(t, nameFilter, "Missing name equality filter")
		assert.Equal(t, 2, input.Limit)
		assert.Equal(t, "domain asc", input.SortParameter.GetGormOrderExpr())
		return interfaces.TaskCollectionOutput{
			Tasks: []models.Task{
				{
					BaseModel: models.BaseModel{
						CreatedAt: testutils.MockCreatedAtValue,
					},
					TaskKey: models.TaskKey{
						Project: projectValue,
						Domain:  domainValue,
						Name:    nameValue,
						Version: "version 0",
					},
				},
				{
					BaseModel: models.BaseModel{
						CreatedAt: testutils.MockCreatedAtValue,
					},
					TaskKey: models.TaskKey{
						Project: projectValue,
						Domain:  domainValue,
						Name:    nameValue,
						Version: "version 1",
					},
				},
			},
		}, nil
	}
	repository.TaskRepo().(*repositoryMocks.MockTaskRepo).SetListCallback(taskListFunc)
	taskManager := NewTaskManager(repository, getMockConfigForTaskTest(), getMockTaskCompiler(), mockScope.NewTestScope())

	taskList, err := taskManager.ListTasks(context.Background(), admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: projectValue,
			Domain:  domainValue,
			Name:    nameValue,
		},
		Limit: 2,
		SortBy: &admin.Sort{
			Direction: admin.Sort_ASCENDING,
			Key:       "domain",
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, taskList)
	assert.Len(t, taskList.Tasks, 2)

	for idx, task := range taskList.Tasks {
		assert.Equal(t, projectValue, task.Id.Project)
		assert.Equal(t, domainValue, task.Id.Domain)
		assert.Equal(t, nameValue, task.Id.Name)
		assert.Equal(t, fmt.Sprintf("version %v", idx), task.Id.Version)
		assert.True(t, proto.Equal(&admin.TaskClosure{
			CreatedAt: testutils.MockCreatedAtProto,
		}, task.Closure))
	}
	assert.Equal(t, "2", taskList.Token)
}

func TestListTasks_MissingParameters(t *testing.T) {
	repository := getMockTaskRepository()
	taskManager := NewTaskManager(repository, getMockConfigForTaskTest(), getMockTaskCompiler(), mockScope.NewTestScope())
	_, err := taskManager.ListTasks(context.Background(), admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Domain: domainValue,
			Name:   nameValue,
		},
		Limit: limit,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, err.(adminErrors.FlyteAdminError).Code())

	_, err = taskManager.ListTasks(context.Background(), admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: projectValue,
			Name:    nameValue,
		},
		Limit: limit,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, err.(adminErrors.FlyteAdminError).Code())
}

func TestListTasks_DatabaseError(t *testing.T) {
	repository := getMockTaskRepository()
	expectedErr := errors.New("expected error")
	taskListFunc := func(input interfaces.ListResourceInput) (interfaces.TaskCollectionOutput, error) {
		return interfaces.TaskCollectionOutput{}, expectedErr
	}

	repository.TaskRepo().(*repositoryMocks.MockTaskRepo).SetListCallback(taskListFunc)
	taskManager := NewTaskManager(repository, getMockConfigForTaskTest(), getMockTaskCompiler(), mockScope.NewTestScope())
	_, err := taskManager.ListTasks(context.Background(), admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: projectValue,
			Domain:  domainValue,
			Name:    nameValue,
		},
		Limit: limit,
	})
	assert.EqualError(t, err, expectedErr.Error())
}

func TestListUniqueTaskIdentifiers(t *testing.T) {
	repository := getMockTaskRepository()
	taskManager := NewTaskManager(repository, getMockConfigForTaskTest(), getMockTaskCompiler(), mockScope.NewTestScope())

	listFunc := func(input interfaces.ListResourceInput) (interfaces.TaskCollectionOutput, error) {
		// Test that parameters are being passed in
		assert.Equal(t, 100, input.Limit)
		assert.Len(t, input.InlineFilters, 2)
		for idx, filter := range input.InlineFilters {
			assert.Equal(t, common.Task, filter.GetEntity())
			query, _ := filter.GetGormQueryExpr()
			if idx == 0 {
				assert.Equal(t, testutils.ProjectQueryPattern, query.Query)
				assert.Equal(t, "foo", query.Args)
			} else {
				assert.Equal(t, testutils.DomainQueryPattern, query.Query)
				assert.Equal(t, "bar", query.Args)
			}
		}
		assert.Equal(t, 10, input.Offset)
		assert.Equal(t, "domain asc", input.SortParameter.GetGormOrderExpr())

		return interfaces.TaskCollectionOutput{
			Tasks: []models.Task{
				{
					TaskKey: models.TaskKey{
						Name:    "name",
						Project: "my project",
						Domain:  "staging",
					},
				},
				{
					TaskKey: models.TaskKey{
						Name:    "name2",
						Project: "my project2",
						Domain:  "staging",
					},
				},
			},
		}, nil
	}

	repository.TaskRepo().(*repositoryMocks.MockTaskRepo).SetListTaskIdentifiersCallback(listFunc)

	resp, err := taskManager.ListUniqueTaskIdentifiers(context.Background(), admin.NamedEntityIdentifierListRequest{
		Project: projectValue,
		Domain:  domainValue,
		Limit:   limit,
		Token:   "10",
		SortBy: &admin.Sort{
			Direction: admin.Sort_ASCENDING,
			Key:       "domain",
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, 2, len(resp.Entities))
	assert.Empty(t, resp.Token)
}
