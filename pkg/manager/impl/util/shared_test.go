package util

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/golang/protobuf/ptypes/wrappers"

	"github.com/flyteorg/flyteadmin/pkg/common"
	commonMocks "github.com/flyteorg/flyteadmin/pkg/common/mocks"
	flyteAdminErrors "github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"
	managerInterfaces "github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	managerMocks "github.com/flyteorg/flyteadmin/pkg/manager/mocks"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	repositoryMocks "github.com/flyteorg/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/storage"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

const project = "project"
const domain = "domain"
const name = "name"
const description = "description"
const version = "version"
const resourceType = core.ResourceType_WORKFLOW
const remoteClosureIdentifier = "remote closure id"

var errExpected = errors.New("expected error")

func TestPopulateExecutionID(t *testing.T) {
	name := GetExecutionName(admin.ExecutionCreateRequest{
		Project: "project",
		Domain:  "domain",
	})
	assert.NotEmpty(t, name)
	assert.Len(t, name, common.ExecutionIDLength)
}

func TestPopulateExecutionID_ExistingName(t *testing.T) {
	name := GetExecutionName(admin.ExecutionCreateRequest{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
	})
	assert.Equal(t, "name", name)
}

func TestGetTask(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	taskGetFunc := func(input interfaces.Identifier) (models.Task, error) {
		assert.Equal(t, project, input.Project)
		assert.Equal(t, domain, input.Domain)
		assert.Equal(t, name, input.Name)
		assert.Equal(t, version, input.Version)
		return models.Task{
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
	task, err := GetTask(context.Background(), repository, core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      "project",
		Domain:       "domain",
		Name:         "name",
		Version:      "version",
	})
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, project, task.Id.Project)
	assert.Equal(t, domain, task.Id.Domain)
	assert.Equal(t, name, task.Id.Name)
	assert.Equal(t, version, task.Id.Version)
}

func TestGetTask_DatabaseError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	taskGetFunc := func(input interfaces.Identifier) (models.Task, error) {
		return models.Task{}, errExpected
	}
	repository.TaskRepo().(*repositoryMocks.MockTaskRepo).SetGetCallback(taskGetFunc)
	task, err := GetTask(context.Background(), repository, core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      "project",
		Domain:       "domain",
		Name:         "name",
		Version:      "version",
	})
	assert.EqualError(t, err, errExpected.Error())
	assert.Nil(t, task)
}

func TestGetTask_TransformerError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	taskGetFunc := func(input interfaces.Identifier) (models.Task, error) {
		assert.Equal(t, project, input.Project)
		assert.Equal(t, domain, input.Domain)
		assert.Equal(t, name, input.Name)
		assert.Equal(t, version, input.Version)
		return models.Task{
			TaskKey: models.TaskKey{
				Project: input.Project,
				Domain:  input.Domain,
				Name:    input.Name,
				Version: input.Version,
			},
			Closure: []byte("i'm invalid"),
		}, nil
	}
	repository.TaskRepo().(*repositoryMocks.MockTaskRepo).SetGetCallback(taskGetFunc)
	task, err := GetTask(context.Background(), repository, core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      "project",
		Domain:       "domain",
		Name:         "name",
		Version:      "version",
	})
	assert.Equal(t, codes.Internal, err.(flyteAdminErrors.FlyteAdminError).Code())
	assert.Nil(t, task)
}

func TestGetWorkflowModel(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	workflowGetFunc := func(input interfaces.Identifier) (models.Workflow, error) {
		assert.Equal(t, project, input.Project)
		assert.Equal(t, domain, input.Domain)
		assert.Equal(t, name, input.Name)
		assert.Equal(t, version, input.Version)
		return models.Workflow{
			WorkflowKey: models.WorkflowKey{
				Project: input.Project,
				Domain:  input.Domain,
				Name:    input.Name,
				Version: input.Version,
			},
			TypedInterface:          testutils.GetWorkflowRequestInterfaceBytes(),
			RemoteClosureIdentifier: remoteClosureIdentifier,
		}, nil
	}
	repository.WorkflowRepo().(*repositoryMocks.MockWorkflowRepo).SetGetCallback(workflowGetFunc)
	workflow, err := GetWorkflowModel(context.Background(), repository, core.Identifier{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      "project",
		Domain:       "domain",
		Name:         "name",
		Version:      "version",
	})
	assert.NoError(t, err)
	assert.NotNil(t, workflow)
	assert.Equal(t, project, workflow.Project)
	assert.Equal(t, domain, workflow.Domain)
	assert.Equal(t, name, workflow.Name)
	assert.Equal(t, version, workflow.Version)
}

func TestGetWorkflowModel_DatabaseError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	workflowGetFunc := func(input interfaces.Identifier) (models.Workflow, error) {
		return models.Workflow{}, errExpected
	}
	repository.WorkflowRepo().(*repositoryMocks.MockWorkflowRepo).SetGetCallback(workflowGetFunc)
	workflow, err := GetWorkflowModel(context.Background(), repository, core.Identifier{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      "project",
		Domain:       "domain",
		Name:         "name",
		Version:      "version",
	})
	assert.EqualError(t, err, errExpected.Error())
	assert.Empty(t, workflow)
}

func TestFetchAndGetWorkflowClosure(t *testing.T) {
	mockStorageClient := commonMocks.GetMockStorageClient()
	mockStorageClient.ComposedProtobufStore.(*commonMocks.TestDataStore).ReadProtobufCb =
		func(ctx context.Context, reference storage.DataReference, msg proto.Message) error {
			assert.Equal(t, remoteClosureIdentifier, reference.String())
			compiledWorkflowClosure := testutils.GetWorkflowClosure()
			workflowBytes, _ := proto.Marshal(compiledWorkflowClosure)
			_ = proto.Unmarshal(workflowBytes, msg)
			return nil
		}
	closure, err := FetchAndGetWorkflowClosure(context.Background(), mockStorageClient, remoteClosureIdentifier)
	assert.Nil(t, err)
	assert.NotNil(t, closure)
}

func TestFetchAndGetWorkflowClosure_RemoteReadError(t *testing.T) {
	mockStorageClient := commonMocks.GetMockStorageClient()
	mockStorageClient.ComposedProtobufStore.(*commonMocks.TestDataStore).ReadProtobufCb =
		func(ctx context.Context, reference storage.DataReference, msg proto.Message) error {
			return errExpected
		}
	closure, err := FetchAndGetWorkflowClosure(context.Background(), mockStorageClient, remoteClosureIdentifier)
	assert.Equal(t, codes.Internal, err.(flyteAdminErrors.FlyteAdminError).Code())
	assert.Nil(t, closure)
}

func TestGetWorkflow(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	workflowGetFunc := func(input interfaces.Identifier) (models.Workflow, error) {
		assert.Equal(t, project, input.Project)
		assert.Equal(t, domain, input.Domain)
		assert.Equal(t, name, input.Name)
		assert.Equal(t, version, input.Version)
		return models.Workflow{
			WorkflowKey: models.WorkflowKey{
				Project: input.Project,
				Domain:  input.Domain,
				Name:    input.Name,
				Version: input.Version,
			},
			TypedInterface:          testutils.GetWorkflowRequestInterfaceBytes(),
			RemoteClosureIdentifier: remoteClosureIdentifier,
		}, nil
	}
	repository.WorkflowRepo().(*repositoryMocks.MockWorkflowRepo).SetGetCallback(workflowGetFunc)

	mockStorageClient := commonMocks.GetMockStorageClient()
	mockStorageClient.ComposedProtobufStore.(*commonMocks.TestDataStore).ReadProtobufCb =
		func(ctx context.Context, reference storage.DataReference, msg proto.Message) error {
			assert.Equal(t, remoteClosureIdentifier, reference.String())
			compiledWorkflowClosure := testutils.GetWorkflowClosure()
			workflowBytes, _ := proto.Marshal(compiledWorkflowClosure)
			_ = proto.Unmarshal(workflowBytes, msg)
			return nil
		}
	workflow, err := GetWorkflow(
		context.Background(), repository, mockStorageClient, core.Identifier{
			ResourceType: core.ResourceType_WORKFLOW,
			Project:      "project",
			Domain:       "domain",
			Name:         "name",
			Version:      "version",
		})
	assert.Nil(t, err)
	assert.NotNil(t, workflow)
}

func TestGetLaunchPlanModel(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	getLaunchPlanFunc := func(input interfaces.Identifier) (models.LaunchPlan, error) {
		assert.Equal(t, project, input.Project)
		assert.Equal(t, domain, input.Domain)
		assert.Equal(t, name, input.Name)
		assert.Equal(t, version, input.Version)
		return models.LaunchPlan{
			LaunchPlanKey: models.LaunchPlanKey{
				Project: input.Project,
				Domain:  input.Domain,
				Name:    input.Name,
				Version: input.Version,
			},
		}, nil
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetGetCallback(getLaunchPlanFunc)
	launchPlan, err := GetLaunchPlanModel(context.Background(), repository, core.Identifier{
		ResourceType: core.ResourceType_LAUNCH_PLAN,
		Project:      "project",
		Domain:       "domain",
		Name:         "name",
		Version:      "version",
	})
	assert.Nil(t, err)
	assert.NotNil(t, launchPlan)
	assert.Equal(t, project, launchPlan.Project)
	assert.Equal(t, domain, launchPlan.Domain)
	assert.Equal(t, name, launchPlan.Name)
	assert.Equal(t, version, launchPlan.Version)
}

func TestGetLaunchPlanModel_DatabaseError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	getLaunchPlanFunc := func(input interfaces.Identifier) (models.LaunchPlan, error) {
		return models.LaunchPlan{}, errExpected
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetGetCallback(getLaunchPlanFunc)
	launchPlan, err := GetLaunchPlanModel(context.Background(), repository, core.Identifier{
		ResourceType: core.ResourceType_LAUNCH_PLAN,
		Project:      "project",
		Domain:       "domain",
		Name:         "name",
		Version:      "version",
	})
	assert.EqualError(t, err, errExpected.Error())
	assert.Empty(t, launchPlan)
}

func TestGetLaunchPlan(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	getLaunchPlanFunc := func(input interfaces.Identifier) (models.LaunchPlan, error) {
		assert.Equal(t, project, input.Project)
		assert.Equal(t, domain, input.Domain)
		assert.Equal(t, name, input.Name)
		assert.Equal(t, version, input.Version)
		return models.LaunchPlan{
			LaunchPlanKey: models.LaunchPlanKey{
				Project: input.Project,
				Domain:  input.Domain,
				Name:    input.Name,
				Version: input.Version,
			},
		}, nil
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetGetCallback(getLaunchPlanFunc)
	launchPlan, err := GetLaunchPlan(context.Background(), repository, core.Identifier{
		ResourceType: core.ResourceType_LAUNCH_PLAN,
		Project:      "project",
		Domain:       "domain",
		Name:         "name",
		Version:      "version",
	})
	assert.Nil(t, err)
	assert.NotNil(t, launchPlan)
	assert.Equal(t, project, launchPlan.Id.Project)
	assert.Equal(t, domain, launchPlan.Id.Domain)
	assert.Equal(t, name, launchPlan.Id.Name)
	assert.Equal(t, version, launchPlan.Id.Version)
}

func TestGetLaunchPlan_TransformerError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	getLaunchPlanFunc := func(input interfaces.Identifier) (models.LaunchPlan, error) {
		assert.Equal(t, project, input.Project)
		assert.Equal(t, domain, input.Domain)
		assert.Equal(t, name, input.Name)
		assert.Equal(t, version, input.Version)
		return models.LaunchPlan{
			LaunchPlanKey: models.LaunchPlanKey{
				Project: input.Project,
				Domain:  input.Domain,
				Name:    input.Name,
				Version: input.Version,
			},
			Spec: []byte("I'm invalid"),
		}, nil
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetGetCallback(getLaunchPlanFunc)
	launchPlan, err := GetLaunchPlan(context.Background(), repository, core.Identifier{
		ResourceType: core.ResourceType_LAUNCH_PLAN,
		Project:      "project",
		Domain:       "domain",
		Name:         "name",
		Version:      "version",
	})
	assert.Equal(t, codes.Internal, err.(flyteAdminErrors.FlyteAdminError).Code())
	assert.Empty(t, launchPlan)
}

func TestGetNamedEntityModel(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	getNamedEntityFunc := func(input interfaces.GetNamedEntityInput) (models.NamedEntity, error) {
		assert.Equal(t, project, input.Project)
		assert.Equal(t, domain, input.Domain)
		assert.Equal(t, name, input.Name)
		assert.Equal(t, resourceType, input.ResourceType)
		return models.NamedEntity{
			NamedEntityKey: models.NamedEntityKey{
				Project:      input.Project,
				Domain:       input.Domain,
				Name:         input.Name,
				ResourceType: input.ResourceType,
			},
			NamedEntityMetadataFields: models.NamedEntityMetadataFields{
				Description: description,
			},
		}, nil
	}
	repository.NamedEntityRepo().(*repositoryMocks.MockNamedEntityRepo).SetGetCallback(getNamedEntityFunc)
	entity, err := GetNamedEntityModel(context.Background(), repository,
		core.ResourceType_WORKFLOW,
		admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		})
	assert.Nil(t, err)
	assert.NotNil(t, entity)
	assert.Equal(t, project, entity.Project)
	assert.Equal(t, domain, entity.Domain)
	assert.Equal(t, name, entity.Name)
	assert.Equal(t, description, entity.Description)
	assert.Equal(t, resourceType, entity.ResourceType)
}

func TestGetNamedEntityModel_DatabaseError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	getNamedEntityFunc := func(input interfaces.GetNamedEntityInput) (models.NamedEntity, error) {
		return models.NamedEntity{}, errExpected
	}
	repository.NamedEntityRepo().(*repositoryMocks.MockNamedEntityRepo).SetGetCallback(getNamedEntityFunc)
	launchPlan, err := GetNamedEntityModel(context.Background(), repository,
		core.ResourceType_WORKFLOW,
		admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		})
	assert.EqualError(t, err, errExpected.Error())
	assert.Empty(t, launchPlan)
}

func TestGetNamedEntity(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	getNamedEntityFunc := func(input interfaces.GetNamedEntityInput) (models.NamedEntity, error) {
		assert.Equal(t, project, input.Project)
		assert.Equal(t, domain, input.Domain)
		assert.Equal(t, name, input.Name)
		assert.Equal(t, resourceType, input.ResourceType)
		return models.NamedEntity{
			NamedEntityKey: models.NamedEntityKey{
				Project:      input.Project,
				Domain:       input.Domain,
				Name:         input.Name,
				ResourceType: core.ResourceType_WORKFLOW,
			},
			NamedEntityMetadataFields: models.NamedEntityMetadataFields{
				Description: description,
			},
		}, nil
	}
	repository.NamedEntityRepo().(*repositoryMocks.MockNamedEntityRepo).SetGetCallback(getNamedEntityFunc)
	entity, err := GetNamedEntity(context.Background(), repository,
		core.ResourceType_WORKFLOW,
		admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		})
	assert.Nil(t, err)
	assert.NotNil(t, entity)
	assert.Equal(t, project, entity.Id.Project)
	assert.Equal(t, domain, entity.Id.Domain)
	assert.Equal(t, name, entity.Id.Name)
	assert.Equal(t, description, entity.Metadata.Description)
	assert.Equal(t, resourceType, entity.ResourceType)
}

func TestGetActiveLaunchPlanVersionFilters(t *testing.T) {
	filters, err := GetActiveLaunchPlanVersionFilters(project, domain, name)
	assert.Nil(t, err)
	assert.NotNil(t, filters)
	assert.Len(t, filters, 4)
	for _, filter := range filters {
		filterExpr, err := filter.GetGormQueryExpr()
		assert.Nil(t, err)
		assert.True(t, strings.Contains(filterExpr.Query, "="))
	}
}

func TestListActiveLaunchPlanVersionsFilters(t *testing.T) {
	filters, err := ListActiveLaunchPlanVersionsFilters(project, domain)
	assert.Nil(t, err)
	assert.Len(t, filters, 3)

	projectExpr, _ := filters[0].GetGormQueryExpr()
	domainExpr, _ := filters[1].GetGormQueryExpr()
	activeExpr, _ := filters[2].GetGormQueryExpr()

	assert.Equal(t, projectExpr.Args, project)
	assert.Equal(t, projectExpr.Query, testutils.ProjectQueryPattern)
	assert.Equal(t, domainExpr.Args, domain)
	assert.Equal(t, domainExpr.Query, testutils.DomainQueryPattern)
	assert.Equal(t, activeExpr.Args, int32(admin.LaunchPlanState_ACTIVE))
	assert.Equal(t, activeExpr.Query, testutils.StateQueryPattern)
}

func TestGetMatchableResource(t *testing.T) {
	resourceType := admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG
	project := "dummyProject"
	domain := "dummyDomain"
	workflow := "dummyWorkflow"
	t.Run("successful fetch", func(t *testing.T) {
		resourceManager := &managerMocks.MockResourceManager{}
		resourceManager.GetResourceFunc = func(ctx context.Context,
			request managerInterfaces.ResourceRequest) (*managerInterfaces.ResourceResponse, error) {
			assert.EqualValues(t, request, managerInterfaces.ResourceRequest{
				Project:      project,
				Domain:       domain,
				ResourceType: resourceType,
			})
			return &managerInterfaces.ResourceResponse{
				Attributes: &admin.MatchingAttributes{
					Target: &admin.MatchingAttributes_WorkflowExecutionConfig{
						WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{
							MaxParallelism: 12,
						},
					},
				},
			}, nil
		}

		mr, err := GetMatchableResource(context.Background(), resourceManager, resourceType, project, domain, "")
		assert.Equal(t, int32(12), mr.Attributes.GetWorkflowExecutionConfig().MaxParallelism)
		assert.Nil(t, err)
	})
	t.Run("successful fetch workflow matchable", func(t *testing.T) {
		resourceManager := &managerMocks.MockResourceManager{}
		resourceManager.GetResourceFunc = func(ctx context.Context,
			request managerInterfaces.ResourceRequest) (*managerInterfaces.ResourceResponse, error) {
			assert.EqualValues(t, request, managerInterfaces.ResourceRequest{
				Project:      project,
				Domain:       domain,
				Workflow:     workflow,
				ResourceType: resourceType,
			})
			return &managerInterfaces.ResourceResponse{
				Attributes: &admin.MatchingAttributes{
					Target: &admin.MatchingAttributes_WorkflowExecutionConfig{
						WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{
							MaxParallelism: 12,
						},
					},
				},
			}, nil
		}

		mr, err := GetMatchableResource(context.Background(), resourceManager, resourceType, project, domain, workflow)
		assert.Equal(t, int32(12), mr.Attributes.GetWorkflowExecutionConfig().MaxParallelism)
		assert.Nil(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		resourceManager := &managerMocks.MockResourceManager{}
		resourceManager.GetResourceFunc = func(ctx context.Context,
			request managerInterfaces.ResourceRequest) (*managerInterfaces.ResourceResponse, error) {
			assert.EqualValues(t, request, managerInterfaces.ResourceRequest{
				Project:      project,
				Domain:       domain,
				ResourceType: resourceType,
			})
			return nil, flyteAdminErrors.NewFlyteAdminError(codes.NotFound, "resource not found")
		}

		_, err := GetMatchableResource(context.Background(), resourceManager, resourceType, project, domain, "")
		assert.Nil(t, err)
	})

	t.Run("internal error", func(t *testing.T) {
		resourceManager := &managerMocks.MockResourceManager{}
		resourceManager.GetResourceFunc = func(ctx context.Context,
			request managerInterfaces.ResourceRequest) (*managerInterfaces.ResourceResponse, error) {
			assert.EqualValues(t, request, managerInterfaces.ResourceRequest{
				Project:      project,
				Domain:       domain,
				ResourceType: resourceType,
			})
			return nil, flyteAdminErrors.NewFlyteAdminError(codes.Internal, "internal error")
		}

		_, err := GetMatchableResource(context.Background(), resourceManager, resourceType, project, domain, "")
		assert.NotNil(t, err)
	})
}

func TestGetDescriptionEntityModel(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	t.Run("Get Description Entity model", func(t *testing.T) {
		entity, err := GetDescriptionEntityModel(context.Background(), repository,
			core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      project,
				Domain:       domain,
				Name:         name,
				Version:      version,
			})
		assert.Nil(t, err)
		assert.NotNil(t, entity)
		assert.Equal(t, "hello world", entity.ShortDescription)
	})

	t.Run("Failed to get DescriptionEntity model", func(t *testing.T) {
		getFunction := func(input interfaces.GetDescriptionEntityInput) (models.DescriptionEntity, error) {
			return models.DescriptionEntity{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound, "NotFound")
		}
		repository.DescriptionEntityRepo().(*repositoryMocks.MockDescriptionEntityRepo).SetGetCallback(getFunction)
		entity, err := GetDescriptionEntityModel(context.Background(), repository,
			core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      project,
				Domain:       domain,
				Name:         name,
				Version:      version,
			})
		assert.Error(t, err)
		assert.Equal(t, "", entity.Name)
	})
}

func TestGetDescriptionEntity(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	t.Run("Get Description Entity", func(t *testing.T) {
		entity, err := GetDescriptionEntity(context.Background(), repository,
			core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      project,
				Domain:       domain,
				Name:         name,
				Version:      version,
			})
		assert.Nil(t, err)
		assert.NotNil(t, entity)
		assert.Equal(t, "hello world", entity.ShortDescription)
	})

	t.Run("Failed to get DescriptionEntity", func(t *testing.T) {
		getFunction := func(input interfaces.GetDescriptionEntityInput) (models.DescriptionEntity, error) {
			return models.DescriptionEntity{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound, "NotFound")
		}
		repository.DescriptionEntityRepo().(*repositoryMocks.MockDescriptionEntityRepo).SetGetCallback(getFunction)
		entity, err := GetDescriptionEntity(context.Background(), repository,
			core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      project,
				Domain:       domain,
				Name:         name,
				Version:      version,
			})
		assert.Error(t, err)
		assert.Nil(t, entity)

		getFunction = func(input interfaces.GetDescriptionEntityInput) (models.DescriptionEntity, error) {
			return models.DescriptionEntity{LongDescription: []byte("???")}, nil
		}
		repository.DescriptionEntityRepo().(*repositoryMocks.MockDescriptionEntityRepo).SetGetCallback(getFunction)
		entity, err = GetDescriptionEntity(context.Background(), repository,
			core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      project,
				Domain:       domain,
				Name:         name,
				Version:      version,
			})
		assert.Error(t, err)
		assert.Nil(t, entity)
	})
}

func TestMergeIntoExecConfig(t *testing.T) {
	var res admin.WorkflowExecutionConfig
	parameters := []struct {
		higher, lower, expected admin.WorkflowExecutionConfig
	}{
		// Max Parallelism taken from higher
		{
			admin.WorkflowExecutionConfig{
				MaxParallelism: 5,
				RawOutputDataConfig: &admin.RawOutputDataConfig{
					OutputLocationPrefix: "s3://test-bucket",
				},
				Labels: &admin.Labels{
					Values: map[string]string{"lab1": "val1"},
				},
				Annotations: &admin.Annotations{
					Values: map[string]string{"ann1": "annval"},
				},
			},
			admin.WorkflowExecutionConfig{
				MaxParallelism: 0,
				RawOutputDataConfig: &admin.RawOutputDataConfig{
					OutputLocationPrefix: "s3://asdf",
				},
				Labels: &admin.Labels{
					Values: map[string]string{"lab1": "oldvalue"},
				},
			},
			admin.WorkflowExecutionConfig{
				MaxParallelism: 5,
				RawOutputDataConfig: &admin.RawOutputDataConfig{
					OutputLocationPrefix: "s3://test-bucket",
				},
				Labels: &admin.Labels{
					Values: map[string]string{"lab1": "val1"},
				},
				Annotations: &admin.Annotations{
					Values: map[string]string{"ann1": "annval"},
				},
			},
		},

		// Values that are set to empty in higher priority get overwritten
		{
			admin.WorkflowExecutionConfig{
				RawOutputDataConfig: &admin.RawOutputDataConfig{
					OutputLocationPrefix: "",
				},
				Labels: &admin.Labels{
					Values: map[string]string{},
				},
				Annotations: &admin.Annotations{
					Values: map[string]string{},
				},
			},
			admin.WorkflowExecutionConfig{
				RawOutputDataConfig: &admin.RawOutputDataConfig{
					OutputLocationPrefix: "s3://asdf",
				},
				Labels: &admin.Labels{
					Values: map[string]string{"lab1": "oldvalue"},
				},
				Annotations: &admin.Annotations{
					Values: map[string]string{"ann1": "annval"},
				},
			},
			admin.WorkflowExecutionConfig{
				RawOutputDataConfig: &admin.RawOutputDataConfig{
					OutputLocationPrefix: "s3://asdf",
				},
				Labels: &admin.Labels{
					Values: map[string]string{"lab1": "oldvalue"},
				},
				Annotations: &admin.Annotations{
					Values: map[string]string{"ann1": "annval"},
				},
			},
		},

		// Values that are not set at all get merged in
		{
			admin.WorkflowExecutionConfig{},
			admin.WorkflowExecutionConfig{
				RawOutputDataConfig: &admin.RawOutputDataConfig{
					OutputLocationPrefix: "s3://asdf",
				},
				Labels: &admin.Labels{
					Values: map[string]string{"lab1": "oldvalue"},
				},
				Annotations: &admin.Annotations{
					Values: map[string]string{"ann1": "annval"},
				},
			},
			admin.WorkflowExecutionConfig{
				RawOutputDataConfig: &admin.RawOutputDataConfig{
					OutputLocationPrefix: "s3://asdf",
				},
				Labels: &admin.Labels{
					Values: map[string]string{"lab1": "oldvalue"},
				},
				Annotations: &admin.Annotations{
					Values: map[string]string{"ann1": "annval"},
				},
			},
		},

		// Interruptible
		{
			admin.WorkflowExecutionConfig{
				Interruptible: &wrappers.BoolValue{
					Value: false,
				},
			},
			admin.WorkflowExecutionConfig{
				Interruptible: &wrappers.BoolValue{
					Value: true,
				},
			},
			admin.WorkflowExecutionConfig{
				Interruptible: &wrappers.BoolValue{
					Value: false,
				},
			},
		},
	}

	for i := range parameters {
		t.Run(fmt.Sprintf("Testing [%v]", i), func(t *testing.T) {
			res = MergeIntoExecConfig(parameters[i].higher, &parameters[i].lower)
			assert.True(t, proto.Equal(&parameters[i].expected, &res))
		})
	}
}
