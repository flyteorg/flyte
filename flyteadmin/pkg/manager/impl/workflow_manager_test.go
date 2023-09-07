package impl

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/common"
	commonMocks "github.com/flyteorg/flyteadmin/pkg/common/mocks"
	adminErrors "github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	repositoryMocks "github.com/flyteorg/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/golang/protobuf/proto"

	flyteErrors "github.com/flyteorg/flyteadmin/pkg/errors"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	runtimeMocks "github.com/flyteorg/flyteadmin/pkg/runtime/mocks"
	workflowengineInterfaces "github.com/flyteorg/flyteadmin/pkg/workflowengine/interfaces"
	workflowengineMocks "github.com/flyteorg/flyteadmin/pkg/workflowengine/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/compiler"
	engine "github.com/flyteorg/flytepropeller/pkg/compiler/common"
	mockScope "github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

const remoteClosureIdentifier = "s3://flyte/metadata/admin/remote closure id"
const returnWorkflowOnGet = true

var workflowIdentifier = core.Identifier{
	ResourceType: core.ResourceType_WORKFLOW,
	Project:      "project",
	Domain:       "domain",
	Name:         "name",
	Version:      "version",
}

var storagePrefix = []string{"metadata", "admin"}

var workflowClosure = admin.WorkflowClosure{
	CompiledWorkflow: &core.CompiledWorkflowClosure{
		Primary: &core.CompiledWorkflow{
			Template: &core.WorkflowTemplate{
				Nodes: []*core.Node{
					{
						Id: engine.StartNodeID,
					},
					{
						Id: engine.EndNodeID,
					},
					{
						Id: "node 1",
					},
					{
						Id: "node 2",
					},
				},
			},
		},
	},
}
var workflowClosureBytes, _ = proto.Marshal(&workflowClosure)

func getMockWorkflowConfigProvider() runtimeInterfaces.Configuration {
	mockWorkflowConfigProvider := runtimeMocks.NewMockConfigurationProvider(
		testutils.GetApplicationConfigWithDefaultDomains(), nil, nil, nil, nil, nil)
	mockWorkflowConfigProvider.(*runtimeMocks.MockConfigurationProvider).AddRegistrationValidationConfiguration(
		runtimeMocks.NewMockRegistrationValidationProvider())
	return mockWorkflowConfigProvider
}

func getMockRepository(workflowOnGet bool) interfaces.Repository {
	mockRepo := repositoryMocks.NewMockRepository()
	if !workflowOnGet {
		mockRepo.(*repositoryMocks.MockRepository).WorkflowRepo().(*repositoryMocks.MockWorkflowRepo).SetGetCallback(
			func(input interfaces.Identifier) (models.Workflow, error) {
				return models.Workflow{}, adminErrors.NewFlyteAdminError(codes.NotFound, "not found")
			})
	}
	return mockRepo
}

func getMockWorkflowCompiler() workflowengineInterfaces.Compiler {
	mockCompiler := workflowengineMocks.NewMockCompiler()
	mockCompiler.(*workflowengineMocks.MockCompiler).AddGetRequirementCallback(
		func(fg *core.WorkflowTemplate, subWfs []*core.WorkflowTemplate) (
			reqs compiler.WorkflowExecutionRequirements, err error) {
			return compiler.WorkflowExecutionRequirements{}, nil
		})
	mockCompiler.(*workflowengineMocks.MockCompiler).AddCompileWorkflowCallback(func(
		primaryWf *core.WorkflowTemplate, subworkflows []*core.WorkflowTemplate, tasks []*core.CompiledTask,
		launchPlans []engine.InterfaceProvider) (*core.CompiledWorkflowClosure, error) {
		return &core.CompiledWorkflowClosure{
			Primary: &core.CompiledWorkflow{
				Template: primaryWf,
			},
		}, nil
	})
	return mockCompiler
}

func getMockStorage() *storage.DataStore {
	mockStorage := commonMocks.GetMockStorageClient()
	mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).WriteProtobufCb =
		func(ctx context.Context, reference storage.DataReference, opts storage.Options, msg proto.Message) error {
			return nil
		}
	mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).ReadProtobufCb =
		func(ctx context.Context, reference storage.DataReference, msg proto.Message) error {
			return nil
		}
	return mockStorage
}

func TestSetWorkflowDefaults(t *testing.T) {
	workflowManager := NewWorkflowManager(
		getMockRepository(returnWorkflowOnGet),
		getMockWorkflowConfigProvider(), getMockWorkflowCompiler(), commonMocks.GetMockStorageClient(), storagePrefix,
		mockScope.NewTestScope())
	request := testutils.GetWorkflowRequest()
	finalizedRequest, err := workflowManager.(*WorkflowManager).setDefaults(request)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(&workflowIdentifier, finalizedRequest.Spec.Template.Id))
}

func TestCreateWorkflow(t *testing.T) {
	repository := getMockRepository(!returnWorkflowOnGet)
	var createCalled bool
	repository.WorkflowRepo().(*repositoryMocks.MockWorkflowRepo).SetCreateCallback(func(input models.Workflow, descriptionEntity *models.DescriptionEntity) error {
		assert.Equal(t, []byte{
			0x2c, 0x69, 0x58, 0x2f, 0xd5, 0x3e, 0x68, 0x7d, 0x5, 0x8e, 0xd9, 0xc8, 0x7d, 0xbd, 0xd1, 0xc7, 0xa7, 0x69,
			0xeb, 0x2e, 0x54, 0x6, 0x3e, 0x67, 0x82, 0xcd, 0x54, 0x7a, 0x91, 0xb3, 0x35, 0x81}, input.Digest)
		createCalled = true
		return nil
	})

	workflowManager := NewWorkflowManager(
		repository,
		getMockWorkflowConfigProvider(), getMockWorkflowCompiler(), getMockStorage(), storagePrefix, mockScope.NewTestScope())
	request := testutils.GetWorkflowRequest()
	response, err := workflowManager.CreateWorkflow(context.Background(), request)
	assert.NoError(t, err)

	expectedResponse := &admin.WorkflowCreateResponse{}
	assert.Equal(t, expectedResponse, response)
	assert.True(t, createCalled)

	repository.WorkflowRepo().(*repositoryMocks.MockWorkflowRepo).SetCreateCallback(func(input models.Workflow, descriptionEntity *models.DescriptionEntity) error {
		return errors.New("failed to insert record into workflow table")
	})
	response, err = workflowManager.CreateWorkflow(context.Background(), request)
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestCreateWorkflow_ValidationError(t *testing.T) {
	workflowManager := NewWorkflowManager(
		repositoryMocks.NewMockRepository(),
		getMockWorkflowConfigProvider(), getMockWorkflowCompiler(), commonMocks.GetMockStorageClient(), storagePrefix,
		mockScope.NewTestScope())
	request := testutils.GetWorkflowRequest()
	request.Id = nil
	response, err := workflowManager.CreateWorkflow(context.Background(), request)
	assert.EqualError(t, err, "missing id")
	assert.Nil(t, response)
}

func TestCreateWorkflow_ExistingWorkflow(t *testing.T) {
	mockStorageClient := commonMocks.GetMockStorageClient()
	mockStorageClient.ComposedProtobufStore.(*commonMocks.TestDataStore).ReadProtobufCb =
		func(ctx context.Context, reference storage.DataReference, msg proto.Message) error {
			workflowClosure := testutils.GetWorkflowClosure()
			bytes, _ := proto.Marshal(workflowClosure)
			_ = proto.Unmarshal(bytes, msg)
			return nil
		}
	workflowManager := NewWorkflowManager(
		getMockRepository(returnWorkflowOnGet),
		getMockWorkflowConfigProvider(), getMockWorkflowCompiler(), mockStorageClient, storagePrefix, mockScope.NewTestScope())
	request := testutils.GetWorkflowRequest()
	response, err := workflowManager.CreateWorkflow(context.Background(), request)
	assert.EqualError(t, err, "workflow with different structure already exists")
	assert.Nil(t, response)
}

func TestCreateWorkflow_ExistingWorkflow_Different(t *testing.T) {
	mockStorageClient := commonMocks.GetMockStorageClient()

	mockStorageClient.ComposedProtobufStore.(*commonMocks.TestDataStore).ReadProtobufCb =
		func(ctx context.Context, reference storage.DataReference, msg proto.Message) error {
			_ = proto.Unmarshal(workflowClosureBytes, msg)
			return nil
		}
	workflowManager := NewWorkflowManager(
		getMockRepository(returnWorkflowOnGet),
		getMockWorkflowConfigProvider(), getMockWorkflowCompiler(), mockStorageClient, storagePrefix, mockScope.NewTestScope())

	request := testutils.GetWorkflowRequest()
	response, err := workflowManager.CreateWorkflow(context.Background(), request)
	assert.EqualError(t, err, "workflow with different structure already exists")
	flyteErr := err.(flyteErrors.FlyteAdminError)
	assert.Equal(t, codes.InvalidArgument, flyteErr.Code())
	assert.Nil(t, response)
}

func TestCreateWorkflow_CompilerGetRequirementsError(t *testing.T) {
	expectedErr := errors.New("expected error")
	mockCompiler := getMockWorkflowCompiler()
	mockCompiler.(*workflowengineMocks.MockCompiler).AddGetRequirementCallback(
		func(fg *core.WorkflowTemplate, subWfs []*core.WorkflowTemplate) (
			reqs compiler.WorkflowExecutionRequirements, err error) {
			return compiler.WorkflowExecutionRequirements{}, expectedErr
		})

	workflowManager := NewWorkflowManager(
		getMockRepository(!returnWorkflowOnGet),
		getMockWorkflowConfigProvider(), mockCompiler, getMockStorage(), storagePrefix, mockScope.NewTestScope())
	request := testutils.GetWorkflowRequest()
	response, err := workflowManager.CreateWorkflow(context.Background(), request)
	assert.EqualError(t, err, fmt.Sprintf(
		"failed to compile workflow for [resource_type:WORKFLOW project:\"project\" domain:\"domain\" "+
			"name:\"name\" version:\"version\" ] with err %v", expectedErr.Error()))
	assert.Nil(t, response)
}

func TestCreateWorkflow_CompileWorkflowError(t *testing.T) {
	expectedErr := errors.New("expected error")
	mockCompiler := getMockWorkflowCompiler()
	mockCompiler.(*workflowengineMocks.MockCompiler).AddCompileWorkflowCallback(func(
		primaryWf *core.WorkflowTemplate, subworkflows []*core.WorkflowTemplate, tasks []*core.CompiledTask,
		launchPlans []engine.InterfaceProvider) (*core.CompiledWorkflowClosure, error) {
		return &core.CompiledWorkflowClosure{}, expectedErr
	})

	workflowManager := NewWorkflowManager(
		getMockRepository(!returnWorkflowOnGet),
		getMockWorkflowConfigProvider(), mockCompiler, getMockStorage(), storagePrefix, mockScope.NewTestScope())
	request := testutils.GetWorkflowRequest()
	response, err := workflowManager.CreateWorkflow(context.Background(), request)
	assert.EqualError(t, err, fmt.Sprintf(
		"failed to compile workflow for [resource_type:WORKFLOW project:\"project\" domain:\"domain\" "+
			"name:\"name\" version:\"version\" ] with err %v", expectedErr.Error()))
	assert.Nil(t, response)
}

func TestCreateWorkflow_DatabaseError(t *testing.T) {
	repository := getMockRepository(!returnWorkflowOnGet)
	expectedErr := errors.New("expected error")
	workflowCreateFunc := func(input models.Workflow, descriptionEntity *models.DescriptionEntity) error {
		return expectedErr
	}

	repository.WorkflowRepo().(*repositoryMocks.MockWorkflowRepo).SetCreateCallback(workflowCreateFunc)
	workflowManager := NewWorkflowManager(
		repository, getMockWorkflowConfigProvider(), getMockWorkflowCompiler(), getMockStorage(), storagePrefix,
		mockScope.NewTestScope())
	request := testutils.GetWorkflowRequest()
	response, err := workflowManager.CreateWorkflow(context.Background(), request)
	assert.EqualError(t, err, expectedErr.Error())
	assert.Nil(t, response)
}

func TestGetWorkflow(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	workflowGetFunc := func(input interfaces.Identifier) (models.Workflow, error) {
		assert.Equal(t, "project", input.Project)
		assert.Equal(t, "domain", input.Domain)
		assert.Equal(t, "name", input.Name)
		assert.Equal(t, "version", input.Version)
		return models.Workflow{
			BaseModel: models.BaseModel{
				CreatedAt: testutils.MockCreatedAtValue,
			},
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
			workflowClosure := testutils.GetWorkflowClosure()
			bytes, _ := proto.Marshal(workflowClosure)
			_ = proto.Unmarshal(bytes, msg)
			return nil
		}
	workflowManager := NewWorkflowManager(
		repository, getMockWorkflowConfigProvider(), getMockWorkflowCompiler(), mockStorageClient, storagePrefix,
		mockScope.NewTestScope())
	workflow, err := workflowManager.GetWorkflow(context.Background(), admin.ObjectGetRequest{
		Id: &workflowIdentifier,
	})
	assert.NoError(t, err)
	assert.Equal(t, "project", workflow.Id.Project)
	assert.Equal(t, "domain", workflow.Id.Domain)
	assert.Equal(t, "name", workflow.Id.Name)
	assert.Equal(t, "version", workflow.Id.Version)
	assert.True(t, proto.Equal(testutils.GetWorkflowClosure(), workflow.Closure),
		"%+v !=\n %+v", testutils.GetWorkflowClosure(), workflow.Closure)
}

func TestGetWorkflow_DatabaseError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	expectedErr := errors.New("expected error")
	workflowGetFunc := func(input interfaces.Identifier) (models.Workflow, error) {
		return models.Workflow{}, expectedErr
	}
	repository.WorkflowRepo().(*repositoryMocks.MockWorkflowRepo).SetGetCallback(workflowGetFunc)
	workflowManager := NewWorkflowManager(
		repository, getMockWorkflowConfigProvider(), getMockWorkflowCompiler(), commonMocks.GetMockStorageClient(),
		storagePrefix, mockScope.NewTestScope())
	workflow, err := workflowManager.GetWorkflow(context.Background(), admin.ObjectGetRequest{
		Id: &workflowIdentifier,
	})
	assert.Nil(t, workflow)
	assert.EqualError(t, err, expectedErr.Error())
}

func TestGetWorkflow_TransformerError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	workflowGetFunc := func(input interfaces.Identifier) (models.Workflow, error) {
		assert.Equal(t, "project", input.Project)
		assert.Equal(t, "domain", input.Domain)
		assert.Equal(t, "name", input.Name)
		assert.Equal(t, "version", input.Version)
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
			return errors.New("womp womp")
		}

	workflowManager := NewWorkflowManager(
		repository, getMockWorkflowConfigProvider(), getMockWorkflowCompiler(), mockStorageClient, storagePrefix,
		mockScope.NewTestScope())
	workflow, err := workflowManager.GetWorkflow(context.Background(), admin.ObjectGetRequest{
		Id: &workflowIdentifier,
	})
	assert.Nil(t, workflow)
	assert.Equal(t, codes.Internal, err.(adminErrors.FlyteAdminError).Code())
}

func TestListWorkflows(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	workflowListFunc := func(input interfaces.ListResourceInput) (interfaces.WorkflowCollectionOutput, error) {
		var projectFilter, domainFilter, nameFilter bool
		assert.Len(t, input.InlineFilters, 3)
		for _, filter := range input.InlineFilters {
			assert.Equal(t, common.Workflow, filter.GetEntity())
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
		assert.Equal(t, limit, input.Limit)
		assert.Equal(t, "domain asc", input.SortParameter.GetGormOrderExpr())
		assert.Equal(t, 10, input.Offset)
		return interfaces.WorkflowCollectionOutput{
			Workflows: []models.Workflow{
				{
					BaseModel: models.BaseModel{
						CreatedAt: testutils.MockCreatedAtValue,
					},
					WorkflowKey: models.WorkflowKey{
						Project: projectValue,
						Domain:  domainValue,
						Name:    nameValue,
						Version: "version 0",
					},
					RemoteClosureIdentifier: remoteClosureIdentifier,
				},
				{
					BaseModel: models.BaseModel{
						CreatedAt: testutils.MockCreatedAtValue,
					},
					WorkflowKey: models.WorkflowKey{
						Project: projectValue,
						Domain:  domainValue,
						Name:    nameValue,
						Version: "version 1",
					},
					RemoteClosureIdentifier: remoteClosureIdentifier,
				},
			},
		}, nil
	}
	repository.WorkflowRepo().(*repositoryMocks.MockWorkflowRepo).SetListCallback(workflowListFunc)
	mockStorageClient := commonMocks.GetMockStorageClient()
	mockStorageClient.ComposedProtobufStore.(*commonMocks.TestDataStore).ReadProtobufCb =
		func(ctx context.Context, reference storage.DataReference, msg proto.Message) error {
			assert.Equal(t, remoteClosureIdentifier, reference.String())
			assert.True(t, proto.Equal(testutils.GetWorkflowClosure(), msg))
			return nil
		}
	workflowManager := NewWorkflowManager(
		repository, getMockWorkflowConfigProvider(), getMockWorkflowCompiler(), mockStorageClient, storagePrefix,
		mockScope.NewTestScope())

	workflowList, err := workflowManager.ListWorkflows(context.Background(), admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: projectValue,
			Domain:  domainValue,
			Name:    nameValue,
		},
		Limit: limit,
		SortBy: &admin.Sort{
			Direction: admin.Sort_ASCENDING,
			Key:       "domain",
		},
		Token: "10",
	})
	assert.NoError(t, err)
	assert.NotNil(t, workflowList)
	assert.Len(t, workflowList.Workflows, 2)

	for idx, workflow := range workflowList.Workflows {
		assert.Equal(t, projectValue, workflow.Id.Project)
		assert.Equal(t, domainValue, workflow.Id.Domain)
		assert.Equal(t, nameValue, workflow.Id.Name)
		assert.Equal(t, fmt.Sprintf("version %v", idx), workflow.Id.Version)
		assert.True(t, proto.Equal(&admin.WorkflowClosure{
			CreatedAt: testutils.MockCreatedAtProto,
		}, workflow.Closure))
	}
	assert.Empty(t, workflowList.Token)
}

func TestListWorkflows_MissingParameters(t *testing.T) {
	workflowManager := NewWorkflowManager(
		repositoryMocks.NewMockRepository(),
		getMockWorkflowConfigProvider(), getMockWorkflowCompiler(), commonMocks.GetMockStorageClient(), storagePrefix,
		mockScope.NewTestScope())
	_, err := workflowManager.ListWorkflows(context.Background(), admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Domain: domainValue,
			Name:   nameValue,
		},
		Limit: limit,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, err.(adminErrors.FlyteAdminError).Code())

	_, err = workflowManager.ListWorkflows(context.Background(), admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: projectValue,
			Name:    nameValue,
		},
		Limit: limit,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, err.(adminErrors.FlyteAdminError).Code())
}

func TestListWorkflows_DatabaseError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	expectedErr := errors.New("expected error")
	workflowListFunc := func(input interfaces.ListResourceInput) (interfaces.WorkflowCollectionOutput, error) {
		return interfaces.WorkflowCollectionOutput{}, expectedErr
	}

	repository.WorkflowRepo().(*repositoryMocks.MockWorkflowRepo).SetListCallback(workflowListFunc)
	workflowManager := NewWorkflowManager(repository,
		getMockWorkflowConfigProvider(), getMockWorkflowCompiler(), commonMocks.GetMockStorageClient(), storagePrefix,
		mockScope.NewTestScope())
	_, err := workflowManager.ListWorkflows(context.Background(), admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: projectValue,
			Domain:  domainValue,
			Name:    nameValue,
		},
		Limit: limit,
	})
	assert.EqualError(t, err, expectedErr.Error())
}

func TestWorkflowManager_ListWorkflowIdentifiers(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	workflowListIdsFunc := func(input interfaces.ListResourceInput) (interfaces.WorkflowCollectionOutput, error) {
		var projectFilter, domainFilter bool
		for _, filter := range input.InlineFilters {
			assert.Equal(t, common.Workflow, filter.GetEntity())
			queryExpr, _ := filter.GetGormQueryExpr()
			if queryExpr.Args == projectValue && queryExpr.Query == testutils.ProjectQueryPattern {
				projectFilter = true
			}
			if queryExpr.Args == domainValue && queryExpr.Query == testutils.DomainQueryPattern {
				domainFilter = true
			}
		}
		assert.True(t, projectFilter, "Missing project equality filter")
		assert.True(t, domainFilter, "Missing domain equality filter")
		assert.Equal(t, limit, input.Limit)
		assert.Equal(t, "domain asc", input.SortParameter.GetGormOrderExpr())
		return interfaces.WorkflowCollectionOutput{
			Workflows: []models.Workflow{
				{
					WorkflowKey: models.WorkflowKey{
						Project: projectValue,
						Domain:  domainValue,
						Name:    nameValue,
						Version: "version 0",
					},
					RemoteClosureIdentifier: remoteClosureIdentifier,
				},
				{
					WorkflowKey: models.WorkflowKey{
						Project: projectValue,
						Domain:  domainValue,
						Name:    nameValue,
						Version: "version 1",
					},
					RemoteClosureIdentifier: remoteClosureIdentifier,
				},
			},
		}, nil
	}
	repository.WorkflowRepo().(*repositoryMocks.MockWorkflowRepo).SetListIdentifiersFunc(workflowListIdsFunc)
	mockStorageClient := commonMocks.GetMockStorageClient()
	mockStorageClient.ComposedProtobufStore.(*commonMocks.TestDataStore).ReadProtobufCb =
		func(ctx context.Context, reference storage.DataReference, msg proto.Message) error {
			assert.Equal(t, remoteClosureIdentifier, reference.String())
			assert.True(t, proto.Equal(testutils.GetWorkflowClosure(), msg))
			return nil
		}
	workflowManager := NewWorkflowManager(
		repository, getMockWorkflowConfigProvider(), getMockWorkflowCompiler(), mockStorageClient, storagePrefix,
		mockScope.NewTestScope())

	workflowList, err := workflowManager.ListWorkflowIdentifiers(context.Background(),
		admin.NamedEntityIdentifierListRequest{
			Project: projectValue,
			Domain:  domainValue,
			Limit:   100,
			SortBy: &admin.Sort{
				Direction: admin.Sort_ASCENDING,
				Key:       "domain",
			},
		})
	assert.NoError(t, err)
	assert.NotNil(t, workflowList)
	assert.Len(t, workflowList.Entities, 2)

	for _, entity := range workflowList.Entities {
		assert.Equal(t, projectValue, entity.Project)
		assert.Equal(t, domainValue, entity.Domain)
		assert.Equal(t, nameValue, entity.Name)
	}
}
