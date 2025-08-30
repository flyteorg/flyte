package impl

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	adminErrors "github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/testutils"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	repositoryMocks "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	runtimeMocks "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/mocks"
	workflowengine "github.com/flyteorg/flyte/flyteadmin/pkg/workflowengine/interfaces"
	workflowMocks "github.com/flyteorg/flyte/flyteadmin/pkg/workflowengine/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	mockScope "github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
)

// Static values for test
const projectValue = "foo"
const domainValue = "bar"
const nameValue = "baz"
const limit = 100

func getMockConfigForTaskTest() runtimeInterfaces.Configuration {
	whitelistConfiguration := &runtimeMocks.WhitelistConfiguration{}
	whitelistConfiguration.EXPECT().GetTaskTypeWhitelist().Return(map[string][]runtimeInterfaces.WhitelistScope{})
	mockConfig := runtimeMocks.NewMockConfigurationProvider(
		testutils.GetApplicationConfigWithDefaultDomains(), nil, nil, runtimeMocks.NewMockTaskResourceConfiguration(
			runtimeInterfaces.TaskResourceSet{}, runtimeInterfaces.TaskResourceSet{}), whitelistConfiguration, nil)
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
	descriptionEntityRepo := mockRepository.DescriptionEntityRepo().(*repositoryMocks.DescriptionEntityRepoInterface)
	descriptionEntityRepo.EXPECT().
		Get(mock.Anything, mock.Anything).
		Return(models.DescriptionEntity{}, adminErrors.NewFlyteAdminErrorf(codes.NotFound, "NotFound"))
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

func TestCreateTask_DuplicateTaskRegistration(t *testing.T) {
	mockRepository := getMockTaskRepository()
	mockRepository.TaskRepo().(*repositoryMocks.MockTaskRepo).SetGetCallback(
		func(input interfaces.Identifier) (models.Task, error) {
			return models.Task{
				TaskKey: models.TaskKey{
					Project: taskIdentifier.GetProject(),
					Domain:  taskIdentifier.GetDomain(),
					Name:    taskIdentifier.GetName(),
					Version: taskIdentifier.GetVersion(),
				},
				Digest: []byte{
					0xbf, 0x79, 0x61, 0x1c, 0xf5, 0xc1, 0xfb, 0x4c, 0xf8, 0xf4, 0xc4, 0x53, 0x5f, 0x8f, 0x73, 0xe2, 0x26, 0x5a,
					0x18, 0x4a, 0xb7, 0x66, 0x98, 0x3c, 0xab, 0x2, 0x6c, 0x9, 0x9b, 0x90, 0xec, 0x8f},
			}, nil
		})
	mockRepository.TaskRepo().(*repositoryMocks.MockTaskRepo).SetCreateCallback(func(input models.Task, descriptionEntity *models.DescriptionEntity) error {
		return adminErrors.NewFlyteAdminErrorf(codes.AlreadyExists, "task already exists")
	})
	taskManager := NewTaskManager(mockRepository, getMockConfigForTaskTest(), getMockTaskCompiler(),
		mockScope.NewTestScope())
	request := testutils.GetValidTaskRequest()
	_, err := taskManager.CreateTask(context.Background(), request)
	assert.Error(t, err)
	flyteErr, ok := err.(adminErrors.FlyteAdminError)
	assert.True(t, ok, "Error should be of type FlyteAdminError")
	assert.Equal(t, codes.AlreadyExists, flyteErr.Code(), "Error code should be AlreadyExists")
	assert.Contains(t, flyteErr.Error(), "task with identical structure already exists")
	differentTemplate := &core.TaskTemplate{
		Id: &core.Identifier{
			ResourceType: core.ResourceType_TASK,
			Project:      "project",
			Domain:       "domain",
			Name:         "name",
			Version:      "version",
		},
		Type: "type",
		Metadata: &core.TaskMetadata{
			Runtime: &core.RuntimeMetadata{
				Version: "runtime version 2",
			},
		},
		Interface: &core.TypedInterface{},
		Target: &core.TaskTemplate_Container{
			Container: &core.Container{
				Image: "image",
				Command: []string{
					"command",
				},
			},
		},
	}
	request.Spec.Template = differentTemplate
	_, err = taskManager.CreateTask(context.Background(), request)
	assert.Error(t, err)
	flyteErr, ok = err.(adminErrors.FlyteAdminError)
	assert.True(t, ok, "Error should be of type FlyteAdminError")
	assert.Equal(t, codes.InvalidArgument, flyteErr.Code(), "Error code should be InvalidArgument")
	assert.Contains(t, flyteErr.Error(), "name task with different structure already exists.")
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

	task, err := taskManager.GetTask(context.Background(), &admin.ObjectGetRequest{
		Id: &taskIdentifier,
	})
	assert.NoError(t, err)
	assert.Equal(t, "project", task.GetId().GetProject())
	assert.Equal(t, "domain", task.GetId().GetDomain())
	assert.Equal(t, "name", task.GetId().GetName())
	assert.Equal(t, "version", task.GetId().GetVersion())
	assert.True(t, proto.Equal(testutils.GetTaskClosure(), task.GetClosure()))
}

func TestGetTask_DatabaseError(t *testing.T) {
	repository := getMockTaskRepository()
	expectedErr := errors.New("expected error")
	taskGetFunc := func(input interfaces.Identifier) (models.Task, error) {
		return models.Task{}, expectedErr
	}
	repository.TaskRepo().(*repositoryMocks.MockTaskRepo).SetGetCallback(taskGetFunc)
	taskManager := NewTaskManager(repository, getMockConfigForTaskTest(), getMockTaskCompiler(), mockScope.NewTestScope())
	task, err := taskManager.GetTask(context.Background(), &admin.ObjectGetRequest{
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

	task, err := taskManager.GetTask(context.Background(), &admin.ObjectGetRequest{
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

	taskList, err := taskManager.ListTasks(context.Background(), &admin.ResourceListRequest{
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
	assert.Len(t, taskList.GetTasks(), 2)

	for idx, task := range taskList.GetTasks() {
		assert.Equal(t, projectValue, task.GetId().GetProject())
		assert.Equal(t, domainValue, task.GetId().GetDomain())
		assert.Equal(t, nameValue, task.GetId().GetName())
		assert.Equal(t, fmt.Sprintf("version %v", idx), task.GetId().GetVersion())
		assert.True(t, proto.Equal(&admin.TaskClosure{
			CreatedAt: testutils.MockCreatedAtProto,
		}, task.GetClosure()))
	}
	assert.Equal(t, "2", taskList.GetToken())
}

func TestListTasks_MissingParameters(t *testing.T) {
	repository := getMockTaskRepository()
	taskManager := NewTaskManager(repository, getMockConfigForTaskTest(), getMockTaskCompiler(), mockScope.NewTestScope())
	_, err := taskManager.ListTasks(context.Background(), &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Domain: domainValue,
			Name:   nameValue,
		},
		Limit: limit,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, err.(adminErrors.FlyteAdminError).Code())

	_, err = taskManager.ListTasks(context.Background(), &admin.ResourceListRequest{
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
	_, err := taskManager.ListTasks(context.Background(), &admin.ResourceListRequest{
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

	resp, err := taskManager.ListUniqueTaskIdentifiers(context.Background(), &admin.NamedEntityIdentifierListRequest{
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
	assert.Equal(t, 2, len(resp.GetEntities()))
	assert.Empty(t, resp.GetToken())
}
