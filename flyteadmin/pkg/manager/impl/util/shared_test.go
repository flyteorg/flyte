package util

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteadmin/pkg/common"
	commonMocks "github.com/flyteorg/flyteadmin/pkg/common/mocks"
	flyteAdminErrors "github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	repositoryMocks "github.com/flyteorg/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
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
	taskGetFunc := func(input interfaces.GetResourceInput) (models.Task, error) {
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
	taskGetFunc := func(input interfaces.GetResourceInput) (models.Task, error) {
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
	taskGetFunc := func(input interfaces.GetResourceInput) (models.Task, error) {
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
	workflowGetFunc := func(input interfaces.GetResourceInput) (models.Workflow, error) {
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
	workflowGetFunc := func(input interfaces.GetResourceInput) (models.Workflow, error) {
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
	workflowGetFunc := func(input interfaces.GetResourceInput) (models.Workflow, error) {
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
	getLaunchPlanFunc := func(input interfaces.GetResourceInput) (models.LaunchPlan, error) {
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
	getLaunchPlanFunc := func(input interfaces.GetResourceInput) (models.LaunchPlan, error) {
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
	getLaunchPlanFunc := func(input interfaces.GetResourceInput) (models.LaunchPlan, error) {
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
	getLaunchPlanFunc := func(input interfaces.GetResourceInput) (models.LaunchPlan, error) {
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
