package impl

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/flyteorg/flyte/flyteadmin/pkg/artifacts"
	artifactMocks "github.com/flyteorg/flyte/flyteadmin/pkg/artifacts/mocks"
	scheduleInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/async/schedule/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/async/schedule/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	commonMocks "github.com/flyteorg/flyte/flyteadmin/pkg/common/mocks"
	flyteAdminErrors "github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/testutils"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	repositoryMocks "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/transformers"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	runtimeMocks "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/mocks"
	"github.com/flyteorg/flyte/flyteadmin/plugins"
	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	artifactsIdl "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifacts"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	mockScope "github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

var active = int32(admin.LaunchPlanState_ACTIVE)
var inactive = int32(admin.LaunchPlanState_INACTIVE)
var mockScheduler = mocks.NewMockEventScheduler()
var launchPlanIdentifier = &core.Identifier{
	ResourceType: core.ResourceType_LAUNCH_PLAN,
	Project:      project,
	Domain:       domain,
	Name:         name,
	Version:      version,
}

var launchPlanNamedIdentifier = &core.Identifier{
	Project: project,
	Domain:  domain,
	Name:    name,
	Version: "version",
}

func getMockPluginRegistry() *plugins.Registry {
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDUserProperties, shared.DefaultGetUserPropertiesFunc)
	return r
}

func getMockRepositoryForLpTest() interfaces.Repository {
	return repositoryMocks.NewMockRepository()
}

func getMockConfigForLpTest() runtimeInterfaces.Configuration {
	mockConfig := runtimeMocks.NewMockConfigurationProvider(
		testutils.GetApplicationConfigWithDefaultDomains(), nil, nil, nil, nil, nil)
	return mockConfig
}

func setDefaultWorkflowCallbackForLpTest(repository interfaces.Repository) {
	workflowSpec := testutils.GetSampleWorkflowSpecForTest()
	typedInterface, _ := proto.Marshal(workflowSpec.Template.Interface)
	workflowGetFunc := func(input interfaces.Identifier) (models.Workflow, error) {
		return models.Workflow{
			WorkflowKey: models.WorkflowKey{
				Project: input.Project,
				Domain:  input.Domain,
				Name:    input.Name,
				Version: input.Version,
			},
			TypedInterface: typedInterface,
		}, nil
	}
	repository.WorkflowRepo().(*repositoryMocks.MockWorkflowRepo).SetGetCallback(workflowGetFunc)
}

func TestCreateLaunchPlan(t *testing.T) {
	repository := getMockRepositoryForLpTest()
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetGetCallback(
		func(input interfaces.Identifier) (models.LaunchPlan, error) {
			return models.LaunchPlan{}, errors.New("foo")
		})
	var createCalled bool
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetCreateCallback(
		func(input models.LaunchPlan) error {
			assert.Equal(t, []byte{0x67, 0xc9, 0xe8, 0xc2, 0xa3, 0x6a, 0x1b, 0xb4, 0x1d, 0xb, 0x44, 0xc7, 0xca, 0xa5, 0x2a, 0xc1, 0x91, 0x32, 0xf, 0x31, 0x33, 0x94, 0xe, 0x2a, 0x8d, 0x89, 0x6b, 0x2d, 0xd7, 0x34, 0x49, 0xe2}, input.Digest)
			createCalled = true
			return nil
		})
	setDefaultWorkflowCallbackForLpTest(repository)
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	request := testutils.GetLaunchPlanRequest()
	response, err := lpManager.CreateLaunchPlan(context.Background(), request)
	assert.Nil(t, err)

	expectedResponse := &admin.LaunchPlanCreateResponse{}
	assert.Equal(t, expectedResponse, response)
	assert.True(t, createCalled)
}

func TestLaunchPlanManager_GetLaunchPlan(t *testing.T) {
	repository := getMockRepositoryForLpTest()
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	state := int32(0)
	lpRequest := testutils.GetLaunchPlanRequest()
	workflowRequest := testutils.GetWorkflowRequest()

	closure := admin.LaunchPlanClosure{
		ExpectedInputs:  lpRequest.Spec.DefaultInputs,
		ExpectedOutputs: workflowRequest.Spec.Template.Interface.Outputs,
	}
	specBytes, _ := proto.Marshal(lpRequest.Spec)
	closureBytes, _ := proto.Marshal(&closure)

	launchPlanGetFunc := func(input interfaces.Identifier) (models.LaunchPlan, error) {
		return models.LaunchPlan{
			LaunchPlanKey: models.LaunchPlanKey{
				Project: input.Project,
				Domain:  input.Domain,
				Name:    input.Name,
				Version: input.Version,
			},
			Spec:       specBytes,
			Closure:    closureBytes,
			WorkflowID: 1,
			State:      &state,
		}, nil
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetGetCallback(launchPlanGetFunc)
	response, err := lpManager.GetLaunchPlan(context.Background(), &admin.ObjectGetRequest{
		Id: launchPlanIdentifier,
	})
	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestLaunchPlanManager_GetActiveLaunchPlan(t *testing.T) {
	repository := getMockRepositoryForLpTest()
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	state := int32(1)
	lpRequest := testutils.GetLaunchPlanRequest()
	workflowRequest := testutils.GetWorkflowRequest()

	closure := admin.LaunchPlanClosure{
		ExpectedInputs:  lpRequest.Spec.DefaultInputs,
		ExpectedOutputs: workflowRequest.Spec.Template.Interface.Outputs,
	}
	specBytes, _ := proto.Marshal(lpRequest.Spec)
	closureBytes, _ := proto.Marshal(&closure)

	launchPlanListFunc := func(input interfaces.ListResourceInput) (interfaces.LaunchPlanCollectionOutput, error) {
		assert.Len(t, input.InlineFilters, 5)

		orgExpr, _ := input.InlineFilters[0].GetGormQueryExpr()
		projectExpr, _ := input.InlineFilters[1].GetGormQueryExpr()
		domainExpr, _ := input.InlineFilters[2].GetGormQueryExpr()
		nameExpr, _ := input.InlineFilters[3].GetGormQueryExpr()
		activeExpr, _ := input.InlineFilters[4].GetGormQueryExpr()

		assert.Equal(t, orgExpr.Args, org)
		assert.Equal(t, orgExpr.Query, testutils.OrgQueryPattern)
		assert.Equal(t, projectExpr.Args, project)
		assert.Equal(t, projectExpr.Query, testutils.ProjectQueryPattern)
		assert.Equal(t, domainExpr.Args, domain)
		assert.Equal(t, domainExpr.Query, testutils.DomainQueryPattern)
		assert.Equal(t, nameExpr.Args, name)
		assert.Equal(t, nameExpr.Query, testutils.NameQueryPattern)
		assert.Equal(t, activeExpr.Args, state)
		assert.Equal(t, activeExpr.Query, testutils.StateQueryPattern)
		return interfaces.LaunchPlanCollectionOutput{
			LaunchPlans: []models.LaunchPlan{
				{
					LaunchPlanKey: models.LaunchPlanKey{
						Org:     lpRequest.Id.Org,
						Project: lpRequest.Id.Project,
						Domain:  lpRequest.Id.Domain,
						Name:    lpRequest.Id.Name,
						Version: lpRequest.Id.Version,
					},
					Spec:       specBytes,
					Closure:    closureBytes,
					WorkflowID: 1,
					State:      &state,
				},
			},
		}, nil
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetListCallback(launchPlanListFunc)
	response, err := lpManager.GetActiveLaunchPlan(context.Background(), &admin.ActiveLaunchPlanRequest{
		Id: &admin.NamedEntityIdentifier{
			Org:     lpRequest.Id.Org,
			Project: lpRequest.Id.Project,
			Domain:  lpRequest.Id.Domain,
			Name:    lpRequest.Id.Name,
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestLaunchPlanManager_GetActiveLaunchPlan_NoneActive(t *testing.T) {
	repository := getMockRepositoryForLpTest()
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	lpRequest := testutils.GetLaunchPlanRequest()

	launchPlanListFunc := func(input interfaces.ListResourceInput) (interfaces.LaunchPlanCollectionOutput, error) {
		return interfaces.LaunchPlanCollectionOutput{}, nil
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetListCallback(launchPlanListFunc)
	response, err := lpManager.GetActiveLaunchPlan(context.Background(), &admin.ActiveLaunchPlanRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: lpRequest.Id.Project,
			Domain:  lpRequest.Id.Domain,
			Name:    lpRequest.Id.Name,
		},
	})
	assert.EqualError(t, err, "No active launch plan could be found: :project:domain:name")
	assert.Nil(t, response)
}

func TestLaunchPlanManager_GetActiveLaunchPlan_InvalidRequest(t *testing.T) {
	repository := getMockRepositoryForLpTest()
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	response, err := lpManager.GetActiveLaunchPlan(context.Background(), &admin.ActiveLaunchPlanRequest{
		Id: &admin.NamedEntityIdentifier{
			Domain: domain,
			Name:   name,
		},
	})
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestLaunchPlan_ValidationError(t *testing.T) {
	lpManager := NewLaunchPlanManager(repositoryMocks.NewMockRepository(), getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	request := testutils.GetLaunchPlanRequest()
	request.Id = nil
	response, err := lpManager.CreateLaunchPlan(context.Background(), request)
	assert.EqualError(t, err, "missing id")
	assert.Nil(t, response)
}

func TestLaunchPlanManager_CreateLaunchPlanErrorDueToBadLabels(t *testing.T) {
	repository := getMockRepositoryForLpTest()
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	request := testutils.GetLaunchPlanRequest()
	request.Spec.Labels = &admin.Labels{
		Values: map[string]string{
			"foo": "#badlabel",
			"bar": "baz",
		}}
	response, err := lpManager.CreateLaunchPlan(context.Background(), request)
	assert.EqualError(t, err, "invalid label value [#badlabel]: [a valid label must be an empty string or consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyValue',  or 'my_value',  or '12345', regex used for validation is '(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?')]")
	assert.Nil(t, response)
}

func TestLaunchPlan_DatabaseError(t *testing.T) {
	repository := getMockRepositoryForLpTest()
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetGetCallback(
		func(input interfaces.Identifier) (models.LaunchPlan, error) {
			return models.LaunchPlan{}, errors.New("foo")
		})
	setDefaultWorkflowCallbackForLpTest(repository)
	expectedErr := errors.New("expected error")
	lpCreateFunc := func(input models.LaunchPlan) error {
		return expectedErr
	}

	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetCreateCallback(lpCreateFunc)
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	request := testutils.GetLaunchPlanRequest()
	response, err := lpManager.CreateLaunchPlan(context.Background(), request)
	assert.EqualError(t, err, expectedErr.Error())
	assert.Nil(t, response)
}

func TestCreateLaunchPlanInCompatibleInputs(t *testing.T) {
	repository := getMockRepositoryForLpTest()
	setDefaultWorkflowCallbackForLpTest(repository)
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	request := testutils.GetLaunchPlanRequest()
	request.Spec.DefaultInputs = &core.ParameterMap{
		Parameters: map[string]*core.Parameter{
			"boo": {
				Var: &core.Variable{
					Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
				},
			},
		},
	}
	response, err := lpManager.CreateLaunchPlan(context.Background(), request)
	assert.Nil(t, response)
	assert.EqualError(t, err, "Invalid variable boo in default_inputs - variable has neither default, nor is required. One must be specified")
}

func TestCreateLaunchPlanValidateCreate(t *testing.T) {
	repository := getMockRepositoryForLpTest()
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetGetCallback(
		func(input interfaces.Identifier) (models.LaunchPlan, error) {
			return models.LaunchPlan{}, errors.New("foo")
		})
	setDefaultWorkflowCallbackForLpTest(repository)
	lpCreateFunc := func(input models.LaunchPlan) error {
		launchPlan, _ := transformers.FromLaunchPlanModel(input)
		assert.Equal(t, project, launchPlan.Id.Project)
		assert.Equal(t, domain, launchPlan.Id.Domain)
		assert.Equal(t, name, launchPlan.Id.Name)
		assert.Equal(t, version, launchPlan.Id.Version)
		assert.True(t, proto.Equal(testutils.GetLaunchPlanRequest().Spec, launchPlan.Spec))
		expectedInputs := &core.ParameterMap{
			Parameters: map[string]*core.Parameter{
				"foo": {
					Var: &core.Variable{
						Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
					},
					Behavior: &core.Parameter_Default{
						Default: coreutils.MustMakeLiteral("foo-value"),
					},
				},
			},
		}
		assert.True(t, proto.Equal(expectedInputs, launchPlan.Closure.ExpectedInputs))
		assert.True(t, proto.Equal(testutils.GetSampleWorkflowSpecForTest().Template.Interface.Outputs,
			launchPlan.Closure.ExpectedOutputs))
		return nil
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetCreateCallback(lpCreateFunc)

	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	request := testutils.GetLaunchPlanRequest()
	response, err := lpManager.CreateLaunchPlan(context.Background(), request)
	assert.Nil(t, err)

	expectedResponse := &admin.LaunchPlanCreateResponse{}
	assert.True(t, proto.Equal(expectedResponse, response))
}

func TestCreateLaunchPlan_ArtifactBehavior(t *testing.T) {
	// Test that enabling artifacts feature flag will not call RegisterArtifactConsumer if no artifact queries present.
	repository := getMockRepositoryForLpTest()
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetGetCallback(
		func(input interfaces.Identifier) (models.LaunchPlan, error) {
			return models.LaunchPlan{}, errors.New("foo")
		})
	client := artifactMocks.ArtifactRegistryClient{}
	registry := artifacts.ArtifactRegistry{
		Client: &client,
	}

	client.On("RegisterConsumer", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		req := args.Get(1).(*artifactsIdl.RegisterConsumerRequest)
		id := req.Consumers[0].EntityId
		assert.Equal(t, "project", id.Project)
		assert.Equal(t, "domain", id.Domain)
		assert.Equal(t, "name", id.Name)
		assert.Equal(t, "version", id.Version)
	}).Return(&artifactsIdl.RegisterResponse{}, nil)

	setDefaultWorkflowCallbackForLpTest(repository)
	mockConfig := getMockConfigForLpTest()
	mockConfig.(*runtimeMocks.MockConfigurationProvider).ApplicationConfiguration().GetTopLevelConfig().FeatureGates.EnableArtifacts = true
	lpManager := NewLaunchPlanManager(repository, mockConfig, mockScheduler, mockScope.NewTestScope(), &registry, getMockPluginRegistry(), nil, nil, nil)
	request := testutils.GetLaunchPlanRequest()
	response, err := lpManager.CreateLaunchPlan(context.Background(), request)
	assert.Nil(t, err)

	expectedResponse := &admin.LaunchPlanCreateResponse{}
	assert.True(t, proto.Equal(expectedResponse, response))
	client.AssertNotCalled(t, "RegisterConsumer", mock.Anything, mock.Anything, mock.Anything)

	// If the launch plan interface has a query however, then the service should be called.
	aq := &core.Parameter_ArtifactQuery{
		ArtifactQuery: &core.ArtifactQuery{
			Identifier: &core.ArtifactQuery_ArtifactId{
				ArtifactId: testutils.GetArtifactID(),
			},
		},
	}
	request.GetSpec().GetDefaultInputs().GetParameters()["foo"] = &core.Parameter{
		Var: &core.Variable{
			Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
		},
		Behavior: aq,
	}
	response, err = lpManager.CreateLaunchPlan(context.Background(), request)
	assert.Nil(t, err)
	expectedResponse = &admin.LaunchPlanCreateResponse{}
	assert.True(t, proto.Equal(expectedResponse, response))
	client.AssertCalled(t, "RegisterConsumer", mock.Anything, mock.Anything, mock.Anything)
}

func TestCreateLaunchPlanNoWorkflowInterface(t *testing.T) {
	repository := getMockRepositoryForLpTest()
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetGetCallback(
		func(input interfaces.Identifier) (models.LaunchPlan, error) {
			return models.LaunchPlan{}, errors.New("foo")
		})
	workflowGetFunc := func(input interfaces.Identifier) (models.Workflow, error) {
		return models.Workflow{
			WorkflowKey: models.WorkflowKey{
				Project: input.Project,
				Domain:  input.Domain,
				Name:    input.Name,
				Version: input.Version,
			},
		}, nil
	}
	repository.WorkflowRepo().(*repositoryMocks.MockWorkflowRepo).SetGetCallback(workflowGetFunc)
	lpCreateFunc := func(input models.LaunchPlan) error {
		launchPlan, _ := transformers.FromLaunchPlanModel(input)
		assert.Equal(t, project, launchPlan.Id.Project)
		assert.Equal(t, domain, launchPlan.Id.Domain)
		assert.Equal(t, name, launchPlan.Id.Name)
		assert.Equal(t, version, launchPlan.Id.Version)
		expectedLaunchPlanSpec := testutils.GetLaunchPlanRequest().Spec
		expectedLaunchPlanSpec.FixedInputs = nil
		expectedLaunchPlanSpec.DefaultInputs.Parameters = map[string]*core.Parameter{}
		assert.EqualValues(t, expectedLaunchPlanSpec.String(), launchPlan.Spec.String())
		assert.Empty(t, launchPlan.Closure.ExpectedInputs)
		return nil
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetCreateCallback(lpCreateFunc)

	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	request := testutils.GetLaunchPlanRequest()
	request.Spec.FixedInputs = nil
	request.Spec.DefaultInputs = nil
	response, err := lpManager.CreateLaunchPlan(context.Background(), request)
	assert.Nil(t, err)

	expectedResponse := &admin.LaunchPlanCreateResponse{}
	assert.Equal(t, expectedResponse, response)
}

func makeLaunchPlanRepoGetCallback(t *testing.T) repositoryMocks.GetLaunchPlanFunc {
	return func(input interfaces.Identifier) (models.LaunchPlan, error) {
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
}

func makeActiveLaunchPlanModelGetCallback(t *testing.T) repositoryMocks.GetLaunchPlanFunc {
	return func(input interfaces.Identifier) (models.LaunchPlan, error) {

		closureBytes, _ := proto.Marshal(&admin.LaunchPlanClosure{
			State: admin.LaunchPlanState_ACTIVE,
		})

		return models.LaunchPlan{
			BaseModel: models.BaseModel{
				CreatedAt: testutils.MockCreatedAtValue,
			},
			LaunchPlanKey: models.LaunchPlanKey{
				Project: input.Project,
				Domain:  input.Domain,
				Name:    input.Name,
				Version: input.Version,
			},
			State:   &active,
			Closure: closureBytes,
		}, nil
	}
}

func makeSchedLaunchPlanRepoGetCallback(t *testing.T) repositoryMocks.GetLaunchPlanFunc {
	return func(input interfaces.Identifier) (models.LaunchPlan, error) {
		specBytes, _ := proto.Marshal(&admin.LaunchPlanSpec{
			EntityMetadata: &admin.LaunchPlanMetadata{
				Schedule: &admin.Schedule{
					ScheduleExpression: &admin.Schedule_Rate{
						Rate: &admin.FixedRate{
							Value: 2,
							Unit:  admin.FixedRateUnit_HOUR,
						},
					},
				},
			},
		})
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
			Spec: specBytes,
		}, nil
	}
}

func makeArtifactTriggerLaunchPlanRepoGetCallback(t *testing.T) repositoryMocks.GetLaunchPlanFunc {
	return func(input interfaces.Identifier) (models.LaunchPlan, error) {
		trigger := &artifactsIdl.Trigger{
			Trigger: &core.ArtifactID{
				ArtifactKey:   nil,
				Version:       "",
				Partitions:    nil,
				TimePartition: nil,
			},
		}
		a, err := anypb.New(trigger)
		assert.NoError(t, err)
		specBytes, _ := proto.Marshal(&admin.LaunchPlanSpec{
			EntityMetadata: &admin.LaunchPlanMetadata{
				LaunchConditions: a,
			},
		})
		x := models.LaunchConditionTypeARTIFACT
		return models.LaunchPlan{
			LaunchPlanKey: models.LaunchPlanKey{
				Project: input.Project,
				Domain:  input.Domain,
				Name:    input.Name,
				Version: input.Version,
			},
			Spec:                specBytes,
			LaunchConditionType: &x,
		}, nil
	}
}

func TestUpdateTrigger(t *testing.T) {
	client := artifactMocks.ArtifactRegistryClient{}
	registry := artifacts.ArtifactRegistry{
		Client: &client,
	}
	repository := getMockRepositoryForLpTest()
	mockScheduler := mocks.NewMockEventScheduler()

	client.On("ActivateTrigger", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		req := args.Get(1).(*artifactsIdl.ActivateTriggerRequest)
		assert.Equal(t, "2", req.TriggerId.Version)
	}).Return(&artifactsIdl.ActivateTriggerResponse{}, nil)

	client.On("DeactivateTrigger", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		req := args.Get(1).(*artifactsIdl.DeactivateTriggerRequest)
		assert.Equal(t, "1", req.TriggerId.Version)
	}).Return(&artifactsIdl.DeactivateTriggerResponse{}, nil)

	x := models.LaunchConditionTypeARTIFACT
	oldLP := models.LaunchPlan{
		LaunchPlanKey: models.LaunchPlanKey{
			Project: "proj",
			Domain:  "dev",
			Name:    "triggered-launch-plan",
			Version: "1",
			Org:     "sample-tenant",
		},
		Spec:                nil,
		WorkflowID:          0,
		Closure:             nil,
		State:               nil,
		Digest:              nil,
		ScheduleType:        "",
		LaunchConditionType: &x,
	}
	newLP := models.LaunchPlan{
		LaunchPlanKey: models.LaunchPlanKey{
			Project: "proj",
			Domain:  "dev",
			Name:    "triggered-launch-plan",
			Version: "2",
			Org:     "sample-tenant",
		},
		Spec:                nil,
		WorkflowID:          0,
		Closure:             nil,
		State:               nil,
		Digest:              nil,
		ScheduleType:        "",
		LaunchConditionType: &x,
	}
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), &registry, getMockPluginRegistry(), nil, nil, nil)
	err := lpManager.(*LaunchPlanManager).updateTriggers(
		context.Background(),
		newLP, &oldLP)
	assert.Nil(t, err)
}

func TestEnableLaunchPlanSched(t *testing.T) {
	repository := getMockRepositoryForLpTest()
	client := artifactMocks.ArtifactRegistryClient{}
	registry := artifacts.ArtifactRegistry{
		Client: &client,
	}
	mockScheduler := mocks.NewMockEventScheduler()
	scheduleExpression := admin.Schedule{
		ScheduleExpression: &admin.Schedule_Rate{
			Rate: &admin.FixedRate{
				Value: 2,
				Unit:  admin.FixedRateUnit_HOUR,
			},
		},
	}
	mockScheduler.(*mocks.MockEventScheduler).SetAddScheduleFunc(
		func(ctx context.Context, input scheduleInterfaces.AddScheduleInput) error {
			assert.True(t, proto.Equal(launchPlanNamedIdentifier, input.Identifier))
			assert.True(t, proto.Equal(&scheduleExpression, input.ScheduleExpression))
			assert.Equal(t, "{\"time\":<time>,\"kickoff_time_arg\":\"\",\"payload\":"+
				"\"Cgdwcm9qZWN0EgZkb21haW4aBG5hbWU=\"}",
				*input.Payload)
			return nil
		})
	lpGetFunc := makeSchedLaunchPlanRepoGetCallback(t)
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetGetCallback(lpGetFunc)
	closureBytes, _ := proto.Marshal(&admin.LaunchPlanClosure{
		State: admin.LaunchPlanState_ACTIVE,
	})
	listFunc := func(input interfaces.ListResourceInput) (interfaces.LaunchPlanCollectionOutput, error) {
		return interfaces.LaunchPlanCollectionOutput{
			LaunchPlans: []models.LaunchPlan{
				{
					BaseModel: models.BaseModel{
						CreatedAt: testutils.MockCreatedAtValue,
					},
					LaunchPlanKey: models.LaunchPlanKey{
						Project: project,
						Domain:  domain,
						Name:    name,
						Version: "old version",
					},
					State:   &active,
					Closure: closureBytes,
				},
			},
		}, nil
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetListCallback(listFunc)

	enableFunc := func(toEnable models.LaunchPlan, toDisable *models.LaunchPlan) error {
		assert.Equal(t, project, toEnable.Project)
		assert.Equal(t, domain, toEnable.Domain)
		assert.Equal(t, name, toEnable.Name)
		assert.Equal(t, version, toEnable.Version)
		assert.Equal(t, active, *toEnable.State)

		assert.Equal(t, project, toDisable.Project)
		assert.Equal(t, domain, toDisable.Domain)
		assert.Equal(t, name, toDisable.Name)
		assert.Equal(t, "old version", toDisable.Version)
		assert.Equal(t, inactive, *toDisable.State)
		var closure admin.LaunchPlanClosure
		err := proto.Unmarshal(toDisable.Closure, &closure)
		assert.NoError(t, err)
		return nil
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetSetActiveCallback(enableFunc)
	mockConfig := getMockConfigForLpTest()
	mockConfig.(*runtimeMocks.MockConfigurationProvider).ApplicationConfiguration().GetTopLevelConfig().FeatureGates.EnableArtifacts = true

	lpManager := NewLaunchPlanManager(repository, mockConfig, mockScheduler, mockScope.NewTestScope(), &registry, getMockPluginRegistry(), nil, nil, nil)
	_, err := lpManager.UpdateLaunchPlan(context.Background(), &admin.LaunchPlanUpdateRequest{
		Id:    launchPlanIdentifier,
		State: admin.LaunchPlanState_ACTIVE,
	})
	assert.NoError(t, err)
	client.AssertNotCalled(t, "ActivateTrigger", mock.Anything, mock.Anything)
	client.AssertNotCalled(t, "DeactivateTrigger", mock.Anything, mock.Anything)
}

func getFakeLaunchPlanManagerWithRegistryClient(t *testing.T, oldModel, newModel models.LaunchPlan, addScheduleFunc mocks.AddScheduleFunc, client *artifactMocks.ArtifactRegistryClient) *LaunchPlanManager {

	repository := getMockRepositoryForLpTest()
	registry := artifacts.ArtifactRegistry{
		Client: client,
	}
	mockScheduler := mocks.NewMockEventScheduler()
	if addScheduleFunc != nil {
		mockScheduler.(*mocks.MockEventScheduler).SetAddScheduleFunc(addScheduleFunc)
	}

	// Get needs to return the new model
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetGetCallback(func(_ interfaces.Identifier) (models.LaunchPlan, error) {
		return newModel, nil
	})
	// List returns the old model
	listFunc := func(_ interfaces.ListResourceInput) (interfaces.LaunchPlanCollectionOutput, error) {
		return interfaces.LaunchPlanCollectionOutput{
			LaunchPlans: []models.LaunchPlan{
				oldModel,
			},
		}, nil
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetListCallback(listFunc)

	enableFunc := func(toEnable models.LaunchPlan, toDisable *models.LaunchPlan) error {
		assert.Equal(t, newModel.Project, toEnable.Project)
		assert.Equal(t, newModel.Domain, toEnable.Domain)
		assert.Equal(t, newModel.Name, toEnable.Name)
		assert.Equal(t, newModel.Version, toEnable.Version)
		assert.Equal(t, active, *toEnable.State)

		assert.Equal(t, oldModel.Project, toDisable.Project)
		assert.Equal(t, oldModel.Domain, toDisable.Domain)
		assert.Equal(t, oldModel.Name, toDisable.Name)
		assert.Equal(t, oldModel.Version, toDisable.Version)
		assert.Equal(t, inactive, *toDisable.State)
		var closure admin.LaunchPlanClosure
		err := proto.Unmarshal(toDisable.Closure, &closure)
		assert.NoError(t, err)
		return nil
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetSetActiveCallback(enableFunc)
	mockConfig := getMockConfigForLpTest()
	mockConfig.(*runtimeMocks.MockConfigurationProvider).ApplicationConfiguration().GetTopLevelConfig().FeatureGates.EnableArtifacts = true

	lpManager := NewLaunchPlanManager(repository, mockConfig, mockScheduler, mockScope.NewTestScope(), &registry, getMockPluginRegistry(), nil, nil, nil).(*LaunchPlanManager)

	return lpManager
}

func TestDisableArtfActivateNothing(t *testing.T) {
	makerWithArtifact := makeArtifactTriggerLaunchPlanRepoGetCallback(t)
	oldID := interfaces.Identifier{
		Project: "p",
		Domain:  "dd",
		Name:    "name1",
		Version: "oldversion",
		Org:     "",
	}
	oldModel, err := makerWithArtifact(oldID)
	assert.NoError(t, err)

	plainMaker := makeActiveLaunchPlanModelGetCallback(t)
	newID := interfaces.Identifier{
		Project: "p",
		Domain:  "dd",
		Name:    "name1",
		Version: "newversion",
		Org:     "",
	}
	newModel, err := plainMaker(newID)
	assert.NoError(t, err)

	client := artifactMocks.ArtifactRegistryClient{}

	lpManager := getFakeLaunchPlanManagerWithRegistryClient(t, oldModel, newModel, nil, &client)
	// Set the mock to return
	client.OnDeactivateTriggerMatch(mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		lpID := args.Get(1).(*artifactsIdl.DeactivateTriggerRequest).TriggerId
		assert.Equal(t, "oldversion", lpID.Version)
	}).Return(&artifactsIdl.DeactivateTriggerResponse{}, nil)

	_, err = lpManager.UpdateLaunchPlan(context.Background(), &admin.LaunchPlanUpdateRequest{
		Id:    launchPlanIdentifier,
		State: admin.LaunchPlanState_ACTIVE,
	})
	assert.NoError(t, err)
	client.AssertNotCalled(t, "ActivateTrigger", mock.Anything, mock.Anything)
	client.AssertCalled(t, "DeactivateTrigger", mock.Anything, mock.Anything)
}

func TestEnableLaunchPlanArtifact(t *testing.T) {
	makerWithArtifact := makeArtifactTriggerLaunchPlanRepoGetCallback(t)
	newID := interfaces.Identifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
		Version: "version",
		Org:     "",
	}
	newModel, err := makerWithArtifact(newID)
	assert.NoError(t, err)

	plainMaker := makeActiveLaunchPlanModelGetCallback(t)
	oldID := interfaces.Identifier{
		Project: "p",
		Domain:  "dd",
		Name:    "name1",
		Version: "oldversion",
		Org:     "",
	}
	oldModel, err := plainMaker(oldID)
	assert.NoError(t, err)

	client := artifactMocks.ArtifactRegistryClient{}
	lpManager := getFakeLaunchPlanManagerWithRegistryClient(t, oldModel, newModel, nil, &client)

	// Set the mock to return
	client.OnActivateTriggerMatch(mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		lpID := args.Get(1).(*artifactsIdl.ActivateTriggerRequest).TriggerId
		assert.Equal(t, "version", lpID.Version)
	}).Return(&artifactsIdl.ActivateTriggerResponse{}, nil)
	// Make the call to trigger everything
	_, err = lpManager.UpdateLaunchPlan(context.Background(), &admin.LaunchPlanUpdateRequest{
		Id:    launchPlanIdentifier,
		State: admin.LaunchPlanState_ACTIVE,
	})
	assert.NoError(t, err)
	client.AssertCalled(t, "ActivateTrigger", mock.Anything, mock.Anything)
	client.AssertNotCalled(t, "DeactivateTrigger", mock.Anything, mock.Anything)
}

func TestEnableSchedule(t *testing.T) {
	repository := getMockRepositoryForLpTest()
	mockScheduler := mocks.NewMockEventScheduler()
	scheduleExpression := &admin.Schedule{
		ScheduleExpression: &admin.Schedule_Rate{
			Rate: &admin.FixedRate{
				Value: 2,
				Unit:  admin.FixedRateUnit_HOUR,
			},
		},
	}
	mockScheduler.(*mocks.MockEventScheduler).SetAddScheduleFunc(
		func(ctx context.Context, input scheduleInterfaces.AddScheduleInput) error {
			assert.True(t, proto.Equal(launchPlanNamedIdentifier, input.Identifier))
			assert.True(t, proto.Equal(scheduleExpression, input.ScheduleExpression))
			assert.Equal(t, "{\"time\":<time>,\"kickoff_time_arg\":\"\",\"payload\":"+
				"\"Cgdwcm9qZWN0EgZkb21haW4aBG5hbWU=\"}",
				*input.Payload)
			return nil
		})
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	err := lpManager.(*LaunchPlanManager).enableSchedule(
		context.Background(),
		launchPlanNamedIdentifier,
		&admin.LaunchPlanSpec{
			EntityMetadata: &admin.LaunchPlanMetadata{
				Schedule: scheduleExpression,
			},
		})
	assert.Nil(t, err)
}

func TestEnableSchedule_ActiveLaunchPlanLimit(t *testing.T) {
	repository := getMockRepositoryForLpTest()
	launchPlanCountFunc := func(ctx context.Context, input interfaces.CountResourceInput) (int64, error) {
		assert.Len(t, input.InlineFilters, 2)
		orgExpr, _ := input.InlineFilters[0].GetGormQueryExpr()
		activeExpr, _ := input.InlineFilters[1].GetGormQueryExpr()

		assert.Equal(t, orgExpr.Args, org)
		assert.Equal(t, orgExpr.Query, testutils.OrgQueryPattern)

		assert.Equal(t, activeExpr.Args, int32(admin.LaunchPlanState_ACTIVE))
		assert.Equal(t, activeExpr.Query, testutils.StateQueryPattern)

		return 2, nil
	}

	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetCountCallback(launchPlanCountFunc)

	getUserProperties := func(ctx context.Context) shared.UserProperties {
		return shared.UserProperties{
			ActiveLaunchPlans: 1,
			Org:               org,
		}
	}
	registry := plugins.NewRegistry()
	registry.RegisterDefault(plugins.PluginIDUserProperties, getUserProperties)

	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), registry, nil, nil, nil)
	_, err := lpManager.(*LaunchPlanManager).enableLaunchPlan(
		context.Background(),
		&admin.LaunchPlanUpdateRequest{
			Id: &core.Identifier{
				ResourceType: core.ResourceType_LAUNCH_PLAN,
				Project:      project,
				Domain:       domain,
				Name:         name,
				Version:      version,
				Org:          org,
			},
		})

	flyteAdminErr := err.(flyteAdminErrors.FlyteAdminError)
	assert.Equal(t, codes.ResourceExhausted, flyteAdminErr.Code())

}

func TestEnableSchedule_Error(t *testing.T) {
	expectedErr := errors.New("expected error")

	repository := getMockRepositoryForLpTest()
	mockScheduler := mocks.NewMockEventScheduler()
	mockScheduler.(*mocks.MockEventScheduler).SetAddScheduleFunc(
		func(ctx context.Context, input scheduleInterfaces.AddScheduleInput) error {
			return expectedErr
		})
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	err := lpManager.(*LaunchPlanManager).enableSchedule(
		context.Background(),
		launchPlanNamedIdentifier, &admin.LaunchPlanSpec{
			EntityMetadata: &admin.LaunchPlanMetadata{
				Schedule: &admin.Schedule{},
			},
		})
	assert.EqualError(t, err, expectedErr.Error())
}

func TestDisableSchedule(t *testing.T) {
	repository := getMockRepositoryForLpTest()
	mockScheduler := mocks.NewMockEventScheduler()
	mockScheduler.(*mocks.MockEventScheduler).SetRemoveScheduleFunc(
		func(ctx context.Context, input scheduleInterfaces.RemoveScheduleInput) error {
			assert.True(t, proto.Equal(launchPlanNamedIdentifier, input.Identifier))
			return nil
		})
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	err := lpManager.(*LaunchPlanManager).disableSchedule(context.Background(), launchPlanNamedIdentifier)
	assert.Nil(t, err)
}

func TestDisableSchedule_Error(t *testing.T) {
	expectedErr := errors.New("expected error")

	repository := getMockRepositoryForLpTest()
	mockScheduler := mocks.NewMockEventScheduler()
	mockScheduler.(*mocks.MockEventScheduler).SetRemoveScheduleFunc(
		func(ctx context.Context, input scheduleInterfaces.RemoveScheduleInput) error {
			return expectedErr
		})
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	err := lpManager.(*LaunchPlanManager).disableSchedule(context.Background(), launchPlanNamedIdentifier)
	assert.EqualError(t, err, expectedErr.Error())
}

func TestUpdateSchedules(t *testing.T) {
	oldScheduleExpression := admin.Schedule{
		ScheduleExpression: &admin.Schedule_Rate{
			Rate: &admin.FixedRate{
				Value: 2,
				Unit:  admin.FixedRateUnit_HOUR,
			},
		},
	}
	oldLaunchPlanSpec := admin.LaunchPlanSpec{
		EntityMetadata: &admin.LaunchPlanMetadata{
			Schedule: &oldScheduleExpression,
		},
	}
	oldLaunchPlanSpecBytes, _ := proto.Marshal(&oldLaunchPlanSpec)
	newScheduleExpression := admin.Schedule{
		ScheduleExpression: &admin.Schedule_CronExpression{
			CronExpression: "cron",
		},
	}
	newLaunchPlanSpec := admin.LaunchPlanSpec{
		EntityMetadata: &admin.LaunchPlanMetadata{
			Schedule: &newScheduleExpression,
		},
	}
	newLaunchPlanSpecBytes, _ := proto.Marshal(&newLaunchPlanSpec)
	mockScheduler := mocks.NewMockEventScheduler()
	var removeCalled bool
	mockScheduler.(*mocks.MockEventScheduler).SetRemoveScheduleFunc(
		func(ctx context.Context, input scheduleInterfaces.RemoveScheduleInput) error {
			assert.True(t, proto.Equal(launchPlanNamedIdentifier, input.Identifier))
			removeCalled = true
			return nil
		})
	var addCalled bool
	mockScheduler.(*mocks.MockEventScheduler).SetAddScheduleFunc(
		func(ctx context.Context, input scheduleInterfaces.AddScheduleInput) error {
			assert.True(t, proto.Equal(launchPlanNamedIdentifier, input.Identifier))
			assert.True(t, proto.Equal(&newScheduleExpression, input.ScheduleExpression))
			addCalled = true
			return nil
		})
	repository := getMockRepositoryForLpTest()
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	err := lpManager.(*LaunchPlanManager).updateSchedules(
		context.Background(),
		models.LaunchPlan{
			LaunchPlanKey: models.LaunchPlanKey{
				Project: project,
				Domain:  domain,
				Name:    name,
				Version: "version",
			},
			Spec: newLaunchPlanSpecBytes,
		},
		&models.LaunchPlan{
			LaunchPlanKey: models.LaunchPlanKey{
				Project: project,
				Domain:  domain,
				Name:    name,
				Version: "version",
			},
			Spec: oldLaunchPlanSpecBytes,
		})
	assert.Nil(t, err)
	assert.True(t, removeCalled)
	assert.True(t, addCalled)
}

func TestUpdateSchedules_NothingToDisableButRedo(t *testing.T) {
	newScheduleExpression := admin.Schedule{
		ScheduleExpression: &admin.Schedule_CronExpression{
			CronExpression: "cron",
		},
	}
	newLaunchPlanSpec := admin.LaunchPlanSpec{
		EntityMetadata: &admin.LaunchPlanMetadata{
			Schedule: &newScheduleExpression,
		},
	}
	newLaunchPlanSpecBytes, _ := proto.Marshal(&newLaunchPlanSpec)
	mockScheduler := mocks.NewMockEventScheduler()
	var addCalled bool
	mockScheduler.(*mocks.MockEventScheduler).SetAddScheduleFunc(
		func(ctx context.Context, input scheduleInterfaces.AddScheduleInput) error {
			assert.True(t, proto.Equal(&core.Identifier{
				Project: project,
				Domain:  domain,
				Name:    name,
				Version: "v1",
			}, input.Identifier))
			assert.True(t, proto.Equal(&newScheduleExpression, input.ScheduleExpression))
			addCalled = true
			return nil
		})
	repository := getMockRepositoryForLpTest()
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	err := lpManager.(*LaunchPlanManager).updateSchedules(context.Background(), models.LaunchPlan{
		LaunchPlanKey: models.LaunchPlanKey{
			Project: project,
			Domain:  domain,
			Name:    name,
			Version: "v1",
		},
		Spec: newLaunchPlanSpecBytes,
	}, nil)
	assert.Nil(t, err)
	assert.True(t, addCalled)

	addCalled = false
	mockScheduler.(*mocks.MockEventScheduler).SetAddScheduleFunc(
		func(ctx context.Context, input scheduleInterfaces.AddScheduleInput) error {
			addCalled = true
			return nil
		})
	oldLaunchPlanSpec := admin.LaunchPlanSpec{
		EntityMetadata: &admin.LaunchPlanMetadata{},
	}
	oldLaunchPlanSpecBytes, _ := proto.Marshal(&oldLaunchPlanSpec)
	err = lpManager.(*LaunchPlanManager).updateSchedules(context.Background(), models.LaunchPlan{
		LaunchPlanKey: models.LaunchPlanKey{
			Project: project,
			Domain:  domain,
			Name:    name,
			Version: "v1",
		},
		Spec: newLaunchPlanSpecBytes,
	}, &models.LaunchPlan{
		Spec: oldLaunchPlanSpecBytes,
	})
	assert.Nil(t, err)
	assert.True(t, addCalled)
}

func TestUpdateSchedules_NothingToEnableButRedo(t *testing.T) {
	oldScheduleExpression := admin.Schedule{
		ScheduleExpression: &admin.Schedule_Rate{
			Rate: &admin.FixedRate{
				Value: 2,
				Unit:  admin.FixedRateUnit_HOUR,
			},
		},
	}
	oldLaunchPlanSpec := admin.LaunchPlanSpec{
		EntityMetadata: &admin.LaunchPlanMetadata{
			Schedule: &oldScheduleExpression,
		},
	}
	oldLaunchPlanSpecBytes, _ := proto.Marshal(&oldLaunchPlanSpec)
	mockScheduler := mocks.NewMockEventScheduler()
	var removeCalled bool
	mockScheduler.(*mocks.MockEventScheduler).SetRemoveScheduleFunc(
		func(ctx context.Context, input scheduleInterfaces.RemoveScheduleInput) error {
			areEqual := proto.Equal(&core.Identifier{
				Project: project,
				Domain:  domain,
				Name:    name,
				Version: "v1",
			}, input.Identifier)
			assert.True(t, areEqual)
			removeCalled = true
			return nil
		})

	repository := getMockRepositoryForLpTest()
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	err := lpManager.(*LaunchPlanManager).updateSchedules(context.Background(), models.LaunchPlan{
		LaunchPlanKey: models.LaunchPlanKey{
			Project: project,
			Domain:  domain,
			Name:    name,
			Version: "v1",
		},
	}, &models.LaunchPlan{
		LaunchPlanKey: models.LaunchPlanKey{
			Project: project,
			Domain:  domain,
			Name:    name,
			Version: "v1",
		},
		Spec: oldLaunchPlanSpecBytes,
	})
	assert.Nil(t, err)
	assert.True(t, removeCalled)
}

func TestUpdateSchedules_NothingToDoButRedo(t *testing.T) {
	scheduleExpression := admin.Schedule{
		ScheduleExpression: &admin.Schedule_CronExpression{
			CronExpression: "cron",
		},
	}
	launchPlanSpec := admin.LaunchPlanSpec{
		EntityMetadata: &admin.LaunchPlanMetadata{
			Schedule: &scheduleExpression,
		},
	}
	launchPlanSpecBytes, _ := proto.Marshal(&launchPlanSpec)

	mockScheduler := mocks.NewMockEventScheduler()
	var removeCalled bool
	mockScheduler.(*mocks.MockEventScheduler).SetRemoveScheduleFunc(
		func(ctx context.Context, input scheduleInterfaces.RemoveScheduleInput) error {
			removeCalled = true
			return nil
		})
	var addCalled bool
	mockScheduler.(*mocks.MockEventScheduler).SetAddScheduleFunc(
		func(ctx context.Context, input scheduleInterfaces.AddScheduleInput) error {
			addCalled = true
			return nil
		})

	repository := getMockRepositoryForLpTest()
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	err := lpManager.(*LaunchPlanManager).updateSchedules(context.Background(), models.LaunchPlan{
		LaunchPlanKey: models.LaunchPlanKey{
			Project: project,
			Domain:  domain,
			Name:    name,
			Version: "v1",
		},
		Spec: launchPlanSpecBytes,
	}, &models.LaunchPlan{
		LaunchPlanKey: models.LaunchPlanKey{
			Project: project,
			Domain:  domain,
			Name:    name,
			Version: "v1",
		},
		Spec: launchPlanSpecBytes,
	})
	assert.Nil(t, err)
	assert.True(t, removeCalled)
	assert.True(t, addCalled)

	err = lpManager.(*LaunchPlanManager).updateSchedules(context.Background(), models.LaunchPlan{
		LaunchPlanKey: models.LaunchPlanKey{
			Project: project,
			Domain:  domain,
			Name:    name,
			Version: "v1",
		},
	}, &models.LaunchPlan{
		LaunchPlanKey: models.LaunchPlanKey{
			Project: project,
			Domain:  domain,
			Name:    name,
			Version: "v1",
		},
	})
	assert.Nil(t, err)
	assert.True(t, removeCalled)
	assert.True(t, addCalled)
}

func TestUpdateSchedules_EnableNoSchedule(t *testing.T) {
	launchPlanSpec := admin.LaunchPlanSpec{}
	launchPlanSpecBytes, _ := proto.Marshal(&launchPlanSpec)

	mockScheduler := mocks.NewMockEventScheduler()
	var removeCalled bool
	mockScheduler.(*mocks.MockEventScheduler).SetRemoveScheduleFunc(
		func(ctx context.Context, input scheduleInterfaces.RemoveScheduleInput) error {
			removeCalled = true
			return nil
		})
	var addCalled bool
	mockScheduler.(*mocks.MockEventScheduler).SetAddScheduleFunc(
		func(ctx context.Context, input scheduleInterfaces.AddScheduleInput) error {
			addCalled = true
			return nil
		})

	repository := getMockRepositoryForLpTest()
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	err := lpManager.(*LaunchPlanManager).updateSchedules(context.Background(), models.LaunchPlan{
		LaunchPlanKey: models.LaunchPlanKey{
			Project: project,
			Domain:  domain,
			Name:    name,
		},
		Spec: launchPlanSpecBytes,
	}, nil)
	assert.Nil(t, err)
	assert.False(t, removeCalled)
	assert.False(t, addCalled)

	err = lpManager.(*LaunchPlanManager).updateSchedules(context.Background(), models.LaunchPlan{
		LaunchPlanKey: models.LaunchPlanKey{
			Project: project,
			Domain:  domain,
			Name:    name,
		},
	}, &models.LaunchPlan{
		LaunchPlanKey: models.LaunchPlanKey{
			Project: project,
			Domain:  domain,
			Name:    name,
		},
	})
	assert.Nil(t, err)
	assert.False(t, removeCalled)
	assert.False(t, addCalled)
}

func TestDisableLaunchPlan(t *testing.T) {
	repository := getMockRepositoryForLpTest()

	lpGetFunc := func(input interfaces.Identifier) (models.LaunchPlan, error) {
		assert.Equal(t, project, input.Project)
		assert.Equal(t, domain, input.Domain)
		assert.Equal(t, name, input.Name)
		assert.Equal(t, version, input.Version)
		specWithSchedule := admin.LaunchPlanSpec{
			EntityMetadata: &admin.LaunchPlanMetadata{
				Schedule: &admin.Schedule{
					ScheduleExpression: &admin.Schedule_CronExpression{
						CronExpression: "foo",
					},
				},
			},
		}
		specWithScheduleBytes, _ := proto.Marshal(&specWithSchedule)
		return models.LaunchPlan{
			LaunchPlanKey: models.LaunchPlanKey{
				Project: input.Project,
				Domain:  input.Domain,
				Name:    input.Name,
				Version: input.Version,
			},
			State: &active,
			Spec:  specWithScheduleBytes,
		}, nil
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetGetCallback(lpGetFunc)
	disableFunc := func(toDisable models.LaunchPlan) error {
		assert.Equal(t, project, toDisable.Project)
		assert.Equal(t, domain, toDisable.Domain)
		assert.Equal(t, name, toDisable.Name)
		assert.Equal(t, version, toDisable.Version)
		assert.Equal(t, inactive, *toDisable.State)
		return nil
	}

	var removeScheduleFuncCalled bool
	mockScheduler := mocks.NewMockEventScheduler()
	mockScheduler.(*mocks.MockEventScheduler).SetRemoveScheduleFunc(
		func(ctx context.Context, input scheduleInterfaces.RemoveScheduleInput) error {
			assert.True(t, proto.Equal(launchPlanNamedIdentifier, input.Identifier))
			removeScheduleFuncCalled = true
			return nil
		})

	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetUpdateCallback(disableFunc)

	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	_, err := lpManager.UpdateLaunchPlan(context.Background(), &admin.LaunchPlanUpdateRequest{
		Id:    launchPlanIdentifier,
		State: admin.LaunchPlanState_INACTIVE,
	})
	assert.NoError(t, err)
	assert.True(t, removeScheduleFuncCalled)
}

func TestDisableLaunchPlan_DatabaseError(t *testing.T) {
	repository := getMockRepositoryForLpTest()
	expectedError := errors.New("expected error")

	lpGetFunc := func(input interfaces.Identifier) (models.LaunchPlan, error) {
		assert.Equal(t, project, input.Project)
		assert.Equal(t, domain, input.Domain)
		assert.Equal(t, name, input.Name)
		assert.Equal(t, version, input.Version)
		return models.LaunchPlan{}, expectedError
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetGetCallback(lpGetFunc)
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	_, err := lpManager.UpdateLaunchPlan(context.Background(), &admin.LaunchPlanUpdateRequest{
		Id:    launchPlanIdentifier,
		State: admin.LaunchPlanState_INACTIVE,
	})
	assert.EqualError(t, err, expectedError.Error(),
		"Failures on getting the existing launch plan should propagate")

	lpGetFunc = func(input interfaces.Identifier) (models.LaunchPlan, error) {
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
			State: &active,
		}, nil
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetGetCallback(lpGetFunc)
	disableFunc := func(toDisable models.LaunchPlan) error {
		assert.Equal(t, project, toDisable.Project)
		assert.Equal(t, domain, toDisable.Domain)
		assert.Equal(t, name, toDisable.Name)
		assert.Equal(t, version, toDisable.Version)
		assert.Equal(t, inactive, *toDisable.State)
		return expectedError
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetUpdateCallback(disableFunc)
	lpManager = NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	_, err = lpManager.UpdateLaunchPlan(context.Background(), &admin.LaunchPlanUpdateRequest{
		Id:    launchPlanIdentifier,
		State: admin.LaunchPlanState_INACTIVE,
	})
	assert.EqualError(t, err, expectedError.Error(),
		"Errors on setting the desired launch plan to inactive should propagate")
}

func TestEnableLaunchPlan(t *testing.T) {
	repository := getMockRepositoryForLpTest()

	lpGetFunc := makeLaunchPlanRepoGetCallback(t)
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetGetCallback(lpGetFunc)
	closureBytes, _ := proto.Marshal(&admin.LaunchPlanClosure{
		State: admin.LaunchPlanState_ACTIVE,
	})
	listFunc := func(input interfaces.ListResourceInput) (interfaces.LaunchPlanCollectionOutput, error) {
		return interfaces.LaunchPlanCollectionOutput{
			LaunchPlans: []models.LaunchPlan{
				{
					BaseModel: models.BaseModel{
						CreatedAt: testutils.MockCreatedAtValue,
					},
					LaunchPlanKey: models.LaunchPlanKey{
						Project: project,
						Domain:  domain,
						Name:    name,
						Version: "old version",
					},
					State:   &active,
					Closure: closureBytes,
				},
			},
		}, nil
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetListCallback(listFunc)

	enableFunc := func(toEnable models.LaunchPlan, toDisable *models.LaunchPlan) error {
		assert.Equal(t, project, toEnable.Project)
		assert.Equal(t, domain, toEnable.Domain)
		assert.Equal(t, name, toEnable.Name)
		assert.Equal(t, version, toEnable.Version)
		assert.Equal(t, active, *toEnable.State)

		assert.Equal(t, project, toDisable.Project)
		assert.Equal(t, domain, toDisable.Domain)
		assert.Equal(t, name, toDisable.Name)
		assert.Equal(t, "old version", toDisable.Version)
		assert.Equal(t, inactive, *toDisable.State)
		var closure admin.LaunchPlanClosure
		err := proto.Unmarshal(toDisable.Closure, &closure)
		assert.NoError(t, err)
		return nil
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetSetActiveCallback(enableFunc)

	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	_, err := lpManager.UpdateLaunchPlan(context.Background(), &admin.LaunchPlanUpdateRequest{
		Id:    launchPlanIdentifier,
		State: admin.LaunchPlanState_ACTIVE,
	})
	assert.NoError(t, err)
}

func TestEnableLaunchPlan_NoCurrentlyActiveVersion(t *testing.T) {
	repository := getMockRepositoryForLpTest()

	lpGetFunc := makeLaunchPlanRepoGetCallback(t)
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetGetCallback(lpGetFunc)
	listFunc := func(input interfaces.ListResourceInput) (interfaces.LaunchPlanCollectionOutput, error) {
		return interfaces.LaunchPlanCollectionOutput{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound, "foo")
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetListCallback(listFunc)

	enableFunc := func(toEnable models.LaunchPlan, toDisable *models.LaunchPlan) error {
		assert.Equal(t, project, toEnable.Project)
		assert.Equal(t, domain, toEnable.Domain)
		assert.Equal(t, name, toEnable.Name)
		assert.Equal(t, version, toEnable.Version)
		assert.Equal(t, active, *toEnable.State)
		assert.Nil(t, toDisable)
		return nil
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetSetActiveCallback(enableFunc)

	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	_, err := lpManager.UpdateLaunchPlan(context.Background(), &admin.LaunchPlanUpdateRequest{
		Id:    launchPlanIdentifier,
		State: admin.LaunchPlanState_ACTIVE,
	})
	assert.NoError(t, err)
}

func TestEnableLaunchPlan_DatabaseError(t *testing.T) {
	repository := getMockRepositoryForLpTest()
	expectedError := errors.New("expected error")

	lpGetFunc := func(input interfaces.Identifier) (models.LaunchPlan, error) {
		assert.Equal(t, project, input.Project)
		assert.Equal(t, domain, input.Domain)
		assert.Equal(t, name, input.Name)
		assert.Equal(t, version, input.Version)
		return models.LaunchPlan{}, expectedError
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetGetCallback(lpGetFunc)
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	_, err := lpManager.UpdateLaunchPlan(context.Background(), &admin.LaunchPlanUpdateRequest{
		Id:    launchPlanIdentifier,
		State: admin.LaunchPlanState_ACTIVE,
	})
	assert.EqualError(t, err, expectedError.Error(), "Failures on getting the existing launch plan should propagate")

	lpGetFunc = makeLaunchPlanRepoGetCallback(t)
	lpManager = NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	listFunc := func(input interfaces.ListResourceInput) (interfaces.LaunchPlanCollectionOutput, error) {
		return interfaces.LaunchPlanCollectionOutput{}, expectedError
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetListCallback(listFunc)
	_, err = lpManager.UpdateLaunchPlan(context.Background(), &admin.LaunchPlanUpdateRequest{
		Id:    launchPlanIdentifier,
		State: admin.LaunchPlanState_ACTIVE,
	})
	assert.EqualError(t, err, expectedError.Error(),
		"Failures on listing the existing active launch plan version should propagate")

	closureBytes, _ := proto.Marshal(&admin.LaunchPlanClosure{
		State: admin.LaunchPlanState_ACTIVE,
	})
	listFunc = func(input interfaces.ListResourceInput) (interfaces.LaunchPlanCollectionOutput, error) {
		return interfaces.LaunchPlanCollectionOutput{
			LaunchPlans: []models.LaunchPlan{
				{
					LaunchPlanKey: models.LaunchPlanKey{
						Project: project,
						Domain:  domain,
						Name:    name,
						Version: "old version",
					},
					State:   &active,
					Closure: closureBytes,
				},
			},
		}, nil
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetListCallback(listFunc)
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetGetCallback(lpGetFunc)
	enableFunc := func(toEnable models.LaunchPlan, toDisable *models.LaunchPlan) error {
		assert.Equal(t, project, toEnable.Project)
		assert.Equal(t, domain, toEnable.Domain)
		assert.Equal(t, name, toEnable.Name)
		assert.Equal(t, version, toEnable.Version)
		assert.Equal(t, active, *toEnable.State)

		assert.Equal(t, project, toDisable.Project)
		assert.Equal(t, domain, toDisable.Domain)
		assert.Equal(t, name, toDisable.Name)
		assert.Equal(t, "old version", toDisable.Version)
		assert.Equal(t, inactive, *toDisable.State)
		return expectedError
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetSetActiveCallback(enableFunc)
	lpManager = NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	_, err = lpManager.UpdateLaunchPlan(context.Background(), &admin.LaunchPlanUpdateRequest{
		Id:    launchPlanIdentifier,
		State: admin.LaunchPlanState_ACTIVE,
	})
	assert.EqualError(t, err, expectedError.Error(), "Errors on setting the desired launch plan to active should propagate")
}

func TestLaunchPlanManager_ListLaunchPlans(t *testing.T) {
	repository := getMockRepositoryForLpTest()
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	state := int32(0)
	lpRequest := testutils.GetLaunchPlanRequest()
	workflowRequest := testutils.GetWorkflowRequest()

	closure := admin.LaunchPlanClosure{
		ExpectedInputs:  lpRequest.Spec.DefaultInputs,
		ExpectedOutputs: workflowRequest.Spec.Template.Interface.Outputs,
	}
	specBytes, _ := proto.Marshal(lpRequest.Spec)
	closureBytes, _ := proto.Marshal(&closure)

	createdAt := time.Now()
	createdAtProto, _ := ptypes.TimestampProto(createdAt)
	updatedAt := createdAt.Add(time.Second)
	updatedAtProto, _ := ptypes.TimestampProto(updatedAt)

	launchPlanListFunc := func(input interfaces.ListResourceInput) (
		interfaces.LaunchPlanCollectionOutput, error) {
		var projectFilter, domainFilter, nameFilter bool

		for _, filter := range input.InlineFilters {
			assert.Equal(t, common.LaunchPlan, filter.GetEntity())
			queryExpr, _ := filter.GetGormQueryExpr()
			if queryExpr.Args == project && queryExpr.Query == testutils.ProjectQueryPattern {
				projectFilter = true
			}
			if queryExpr.Args == domain && queryExpr.Query == testutils.DomainQueryPattern {
				domainFilter = true
			}
			if queryExpr.Args == name && queryExpr.Query == testutils.NameQueryPattern {
				nameFilter = true
			}
		}
		assert.True(t, projectFilter, "Missing project equality filter")
		assert.True(t, domainFilter, "Missing domain equality filter")
		assert.True(t, nameFilter, "Missing name equality filter")
		assert.Equal(t, 10, input.Limit)
		assert.Equal(t, 2, input.Offset)
		assert.Equal(t, "domain asc", input.SortParameter.GetGormOrderExpr())

		return interfaces.LaunchPlanCollectionOutput{
			LaunchPlans: []models.LaunchPlan{
				{
					BaseModel: models.BaseModel{
						CreatedAt: createdAt,
						UpdatedAt: updatedAt,
					},
					LaunchPlanKey: models.LaunchPlanKey{
						Project: project,
						Domain:  domain,
						Name:    name,
						Version: "1",
					},
					Spec:       specBytes,
					Closure:    closureBytes,
					WorkflowID: 1,
					State:      &state,
				},
				{
					BaseModel: models.BaseModel{
						CreatedAt: createdAt,
						UpdatedAt: updatedAt,
					},
					LaunchPlanKey: models.LaunchPlanKey{
						Project: project,
						Domain:  domain,
						Name:    name,
						Version: "2",
					},
					Spec:       specBytes,
					Closure:    closureBytes,
					WorkflowID: 1,
					State:      &state,
				},
			},
		}, nil
	}

	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetListCallback(launchPlanListFunc)
	lpList, err := lpManager.ListLaunchPlans(context.Background(), &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: project,
			Domain:  domain,
			Name:    name,
		},
		Limit: 10,
		Token: "2",
		SortBy: &admin.Sort{
			Direction: admin.Sort_ASCENDING,
			Key:       domain,
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(lpList.LaunchPlans))
	for idx, lp := range lpList.LaunchPlans {
		assert.Equal(t, project, lp.Id.Project)
		assert.Equal(t, domain, lp.Id.Domain)
		assert.Equal(t, name, lp.Id.Name)
		assert.Equal(t, fmt.Sprintf("%v", idx+1), lp.Id.Version)
		assert.True(t, proto.Equal(createdAtProto, lp.Closure.CreatedAt))
		assert.True(t, proto.Equal(updatedAtProto, lp.Closure.UpdatedAt))
	}
}

func TestLaunchPlanManager_ListLaunchPlanIds(t *testing.T) {
	repository := getMockRepositoryForLpTest()
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	state := int32(0)
	lpRequest := testutils.GetLaunchPlanRequest()
	workflowRequest := testutils.GetWorkflowRequest()

	closure := admin.LaunchPlanClosure{
		ExpectedInputs:  lpRequest.Spec.DefaultInputs,
		ExpectedOutputs: workflowRequest.Spec.Template.Interface.Outputs,
	}
	specBytes, _ := proto.Marshal(lpRequest.Spec)
	closureBytes, _ := proto.Marshal(&closure)

	launchPlanListFunc := func(input interfaces.ListResourceInput) (
		interfaces.LaunchPlanCollectionOutput, error) {
		var projectFilter, domainFilter bool

		for _, filter := range input.InlineFilters {
			assert.Equal(t, common.LaunchPlan, filter.GetEntity())
			queryExpr, _ := filter.GetGormQueryExpr()
			if queryExpr.Args == project && queryExpr.Query == testutils.ProjectQueryPattern {
				projectFilter = true
			}
			if queryExpr.Args == domain && queryExpr.Query == testutils.DomainQueryPattern {
				domainFilter = true
			}
		}
		assert.True(t, projectFilter, "Missing project equality filter")
		assert.True(t, domainFilter, "Missing domain equality filter")
		assert.Equal(t, 10, input.Limit)
		assert.Equal(t, "domain asc", input.SortParameter.GetGormOrderExpr())

		return interfaces.LaunchPlanCollectionOutput{
			LaunchPlans: []models.LaunchPlan{
				{
					LaunchPlanKey: models.LaunchPlanKey{
						Project: project,
						Domain:  domain,
						Name:    name,
						Version: "1",
					},
					Spec:       specBytes,
					Closure:    closureBytes,
					WorkflowID: 1,
					State:      &state,
				},
				{
					LaunchPlanKey: models.LaunchPlanKey{
						Project: project,
						Domain:  domain,
						Name:    name,
						Version: "2",
					},
					Spec:       specBytes,
					Closure:    closureBytes,
					WorkflowID: 1,
					State:      &state,
				},
			},
		}, nil
	}

	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetListLaunchPlanIdentifiersCallback(
		launchPlanListFunc)
	lpList, err := lpManager.ListLaunchPlanIds(context.Background(), &admin.NamedEntityIdentifierListRequest{
		Project: project,
		Domain:  domain,
		Limit:   10,
		SortBy: &admin.Sort{
			Direction: admin.Sort_ASCENDING,
			Key:       domain,
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(lpList.Entities))
	for _, id := range lpList.Entities {
		assert.Equal(t, project, id.Project)
		assert.Equal(t, domain, id.Domain)
		assert.Equal(t, name, id.Name)
	}
}

func TestLaunchPlanManager_ListActiveLaunchPlans(t *testing.T) {
	repository := getMockRepositoryForLpTest()
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	state := int32(admin.LaunchPlanState_ACTIVE)
	lpRequest := testutils.GetLaunchPlanRequest()
	workflowRequest := testutils.GetWorkflowRequest()

	closure := admin.LaunchPlanClosure{
		ExpectedInputs:  lpRequest.Spec.DefaultInputs,
		ExpectedOutputs: workflowRequest.Spec.Template.Interface.Outputs,
	}
	specBytes, _ := proto.Marshal(lpRequest.Spec)
	closureBytes, _ := proto.Marshal(&closure)

	launchPlanListFunc := func(input interfaces.ListResourceInput) (
		interfaces.LaunchPlanCollectionOutput, error) {
		var projectFilter, domainFilter, activeFilter bool

		for _, filter := range input.InlineFilters {
			assert.Equal(t, common.LaunchPlan, filter.GetEntity())
			queryExpr, _ := filter.GetGormQueryExpr()
			if queryExpr.Args == project && queryExpr.Query == testutils.ProjectQueryPattern {
				projectFilter = true
			}
			if queryExpr.Args == domain && queryExpr.Query == testutils.DomainQueryPattern {
				domainFilter = true
			}
			if queryExpr.Args == state && queryExpr.Query == testutils.StateQueryPattern {
				activeFilter = true
			}
		}
		assert.True(t, projectFilter, "Missing project equality filter")
		assert.True(t, domainFilter, "Missing domain equality filter")
		assert.True(t, activeFilter, "Missing active filter")
		assert.Equal(t, 10, input.Limit)
		assert.Equal(t, "domain asc", input.SortParameter.GetGormOrderExpr())

		return interfaces.LaunchPlanCollectionOutput{
			LaunchPlans: []models.LaunchPlan{
				{
					LaunchPlanKey: models.LaunchPlanKey{
						Project: project,
						Domain:  domain,
						Name:    name,
						Version: "1",
					},
					Spec:       specBytes,
					Closure:    closureBytes,
					WorkflowID: 1,
					State:      &state,
				},
				{
					LaunchPlanKey: models.LaunchPlanKey{
						Project: project,
						Domain:  domain,
						Name:    name,
						Version: "2",
					},
					Spec:       specBytes,
					Closure:    closureBytes,
					WorkflowID: 1,
					State:      &state,
				},
			},
		}, nil
	}

	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetListCallback(
		launchPlanListFunc)
	lpList, err := lpManager.ListActiveLaunchPlans(context.Background(), &admin.ActiveLaunchPlanListRequest{
		Project: project,
		Domain:  domain,
		Limit:   10,
		SortBy: &admin.Sort{
			Direction: admin.Sort_ASCENDING,
			Key:       domain,
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(lpList.LaunchPlans))
	for _, id := range lpList.LaunchPlans {
		assert.Equal(t, project, id.Id.Project)
		assert.Equal(t, domain, id.Id.Domain)
		assert.Equal(t, name, id.Id.Name)
	}
}

func TestLaunchPlanManager_ListActiveLaunchPlans_BadRequest(t *testing.T) {
	repository := getMockRepositoryForLpTest()
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), nil, nil, nil)
	lpList, err := lpManager.ListActiveLaunchPlans(context.Background(), &admin.ActiveLaunchPlanListRequest{
		Domain: domain,
		Limit:  10,
		SortBy: &admin.Sort{
			Direction: admin.Sort_ASCENDING,
			Key:       domain,
		},
	})
	assert.Error(t, err)
	assert.Nil(t, lpList)
}

func TestIsScheduleEmpty(t *testing.T) {
	t.Run("deprecated cron expression used", func(t *testing.T) {
		sp := &admin.LaunchPlanSpec{
			EntityMetadata: &admin.LaunchPlanMetadata{
				Schedule: &admin.Schedule{
					ScheduleExpression: &admin.Schedule_CronExpression{},
				},
			},
		}
		assert.True(t, isScheduleEmpty(sp))
	})
	t.Run("deprecated cron expression used", func(t *testing.T) {
		sp := &admin.LaunchPlanSpec{
			EntityMetadata: &admin.LaunchPlanMetadata{
				Schedule: &admin.Schedule{
					ScheduleExpression: &admin.Schedule_CronExpression{
						CronExpression: "* * * * *",
					},
				},
			},
		}
		assert.False(t, isScheduleEmpty(sp))
	})
	t.Run("fixed rate used", func(t *testing.T) {
		sp := &admin.LaunchPlanSpec{
			EntityMetadata: &admin.LaunchPlanMetadata{
				Schedule: &admin.Schedule{
					ScheduleExpression: &admin.Schedule_Rate{
						Rate: &admin.FixedRate{
							Value: 10,
							Unit:  admin.FixedRateUnit_HOUR,
						},
					},
				},
			},
		}
		assert.False(t, isScheduleEmpty(sp))
	})
	t.Run("cron schedule used", func(t *testing.T) {
		sp := &admin.LaunchPlanSpec{
			EntityMetadata: &admin.LaunchPlanMetadata{
				Schedule: &admin.Schedule{
					ScheduleExpression: &admin.Schedule_CronSchedule{
						CronSchedule: &admin.CronSchedule{
							Schedule: "* * * * *",
						},
					},
				},
			},
		}
		assert.False(t, isScheduleEmpty(sp))
	})
	t.Run("kick off time used", func(t *testing.T) {
		sp := &admin.LaunchPlanSpec{
			EntityMetadata: &admin.LaunchPlanMetadata{
				Schedule: &admin.Schedule{
					KickoffTimeInputArg: "name",
				},
			},
		}
		assert.True(t, isScheduleEmpty(sp))
	})
}

func TestCreateLaunchPlanFromNode(t *testing.T) {
	subNodeID := "node 1"
	sudNodeID2 := "node 2"
	subWorkflowNodeID := "sub_wf"
	subWorkflowNodeID1 := "wub_wf_1"
	simpleIntInput := &core.VariableMap{
		Variables: map[string]*core.Variable{
			"a": {
				Type: &core.LiteralType{
					Type: &core.LiteralType_Simple{
						Simple: core.SimpleType_INTEGER,
					},
				},
			},
		},
	}
	simpleIntOutput := &core.VariableMap{
		Variables: map[string]*core.Variable{
			"b": {
				Type: &core.LiteralType{
					Type: &core.LiteralType_Simple{
						Simple: core.SimpleType_INTEGER,
					},
				},
			},
		},
	}
	tests := []struct {
		name              string
		targetType        *core.Node
		subNodeID         []string
		expectErr         bool
		expectedErrCode   codes.Code
		inputVariableMap  *core.VariableMap
		outputVariableMap *core.VariableMap
	}{
		{
			name: "TaskNode",
			targetType: &core.Node{
				Target: &core.Node_TaskNode{},
			},
			subNodeID:         []string{subNodeID},
			inputVariableMap:  simpleIntInput,
			outputVariableMap: simpleIntOutput,
		},
		{
			name: "ArrayNode TaskNode",
			targetType: &core.Node{
				Target: &core.Node_ArrayNode{},
			},
			subNodeID: []string{subNodeID},
			inputVariableMap: &core.VariableMap{
				Variables: map[string]*core.Variable{
					"a": {
						Type: &core.LiteralType{
							Type: &core.LiteralType_CollectionType{
								CollectionType: &core.LiteralType{
									Type: &core.LiteralType_Simple{
										Simple: core.SimpleType_INTEGER,
									},
								},
							},
						},
					},
				},
			},
			outputVariableMap: &core.VariableMap{
				Variables: map[string]*core.Variable{
					"b": {
						Type: &core.LiteralType{
							Type: &core.LiteralType_CollectionType{
								CollectionType: &core.LiteralType{
									Type: &core.LiteralType_Simple{
										Simple: core.SimpleType_INTEGER,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "SubNode Does Not Exist",
			targetType: &core.Node{
				Target: &core.Node_WorkflowNode{},
			},
			subNodeID:       []string{"does not exist"},
			expectErr:       true,
			expectedErrCode: codes.NotFound,
		},
		{
			name: "SubWorkflow TaskNode",
			targetType: &core.Node{
				Target: &core.Node_WorkflowNode{},
			},
			subNodeID:         []string{sudNodeID2, subWorkflowNodeID1},
			inputVariableMap:  simpleIntInput,
			outputVariableMap: simpleIntOutput,
		},
		{
			name: "Nested SubWorkflow TaskNode",
			targetType: &core.Node{
				Target: &core.Node_WorkflowNode{},
			},
			subNodeID:         []string{subNodeID, subWorkflowNodeID, subWorkflowNodeID1},
			inputVariableMap:  simpleIntInput,
			outputVariableMap: simpleIntOutput,
		},
		{
			name:      "SubNode Does Not Exist in SubWorkflow",
			subNodeID: []string{subNodeID, "does not exist"},
			targetType: &core.Node{
				Target: &core.Node_WorkflowNode{},
			},
			expectErr:       true,
			expectedErrCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := testutils.GetLaunchPlanRequest()
			createLaunchPlanFromNodeRequest := &admin.CreateLaunchPlanFromNodeRequest{
				LaunchPlanId: request.Id,
				SubNodes: &admin.CreateLaunchPlanFromNodeRequest_SubNodeIds{
					SubNodeIds: &admin.SubNodeList{
						SubNodeIds: []*admin.SubNodeIdAsList{
							{
								SubNodeId: tt.subNodeID,
							},
						},
					},
				},
			}

			mockStorage := getMockStorageForExecTest(context.Background())

			repository := getMockRepositoryForExecTest()
			var newlyCreatedWorkflow models.Workflow
			workflowCreateFunc := func(input models.Workflow, descriptionEntity *models.DescriptionEntity) error {
				newlyCreatedWorkflow = input
				return nil
			}
			repository.WorkflowRepo().(*repositoryMocks.MockWorkflowRepo).SetCreateCallback(workflowCreateFunc)

			lpSpecBytes, err := proto.Marshal(request.Spec)
			assert.Nil(t, err)

			var launchplan *models.LaunchPlan
			repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetCreateCallback(func(input models.LaunchPlan) error {
				launchplan = &input
				return nil
			})
			var getLaunchPlanCalledCount = 0
			repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetGetCallback(func(input interfaces.Identifier) (models.LaunchPlan, error) {
				if getLaunchPlanCalledCount == 0 {
					lpModel := models.LaunchPlan{
						LaunchPlanKey: models.LaunchPlanKey{
							Project: input.Project,
							Domain:  input.Domain,
							Name:    input.Name,
							Version: input.Version,
						},
						BaseModel: models.BaseModel{
							ID: uint(100),
						},
						Spec: lpSpecBytes,
					}
					getLaunchPlanCalledCount++
					return lpModel, nil
				}
				if launchplan == nil {
					return models.LaunchPlan{}, flyteAdminErrors.NewFlyteAdminError(codes.NotFound, "launchplan not found")
				}
				return *launchplan, nil
			})

			var originWfGetCalledCount = 0
			var newWfGetCalledCount = 0
			workflowGetFunc := func(input interfaces.Identifier) (models.Workflow, error) {
				if strings.Contains(input.Name, "flytegen") {
					if newWfGetCalledCount <= 1 {
						newWfGetCalledCount++
						return models.Workflow{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound, "not found")
					}
					return newlyCreatedWorkflow, nil
				}
				if originWfGetCalledCount == 1 {
					originWfGetCalledCount++
					return models.Workflow{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound, "not found")
				}
				originWfGetCalledCount++
				return newlyCreatedWorkflow, nil
			}
			repository.WorkflowRepo().(*repositoryMocks.MockWorkflowRepo).SetGetCallback(workflowGetFunc)
			taskIdentifier := &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "flytekit",
				Domain:       "production",
				Name:         "simple_task",
				Version:      "12345",
			}
			repository.TaskRepo().(*repositoryMocks.MockTaskRepo).SetGetCallback(
				func(input interfaces.Identifier) (models.Task, error) {
					createdAt := time.Now()
					createdAtProto, _ := ptypes.TimestampProto(createdAt)
					taskClosure := &admin.TaskClosure{
						CompiledTask: &core.CompiledTask{
							Template: &core.TaskTemplate{
								Id:   taskIdentifier,
								Type: "python-task",
								Metadata: &core.TaskMetadata{
									Runtime: &core.RuntimeMetadata{
										Type:    core.RuntimeMetadata_FLYTE_SDK,
										Version: "0.6.2",
										Flavor:  "python",
									},
									Timeout: ptypes.DurationProto(time.Second),
								},
								Interface: &core.TypedInterface{
									Inputs:  tt.inputVariableMap,
									Outputs: tt.outputVariableMap,
								},
								Custom: nil,
								Target: &core.TaskTemplate_Container{
									Container: &core.Container{
										Image: "docker.io/my:image",
										Args: []string{
											"pyflyte-execute",
											"--task-module",
											"workflows.simple",
											"--task-name",
											"simple_task",
											"--inputs",
											"{{.input}}",
											"--output-prefix",
											"{{.outputPrefix}}",
										},
										Env: []*core.KeyValuePair{
											{
												Key:   "FLYTE_INTERNAL_PROJECT",
												Value: "flytekit",
											},
											{
												Key:   "FLYTE_INTERNAL_DOMAIN",
												Value: "production",
											},
											{
												Key:   "FLYTE_INTERNAL_NAME",
												Value: "simple_task",
											},
											{
												Key:   "FLYTE_INTERNAL_VERSION",
												Value: "12345",
											},
										},
									},
								},
							},
						},
						CreatedAt: createdAtProto,
					}
					serializedTaskClosure, err := proto.Marshal(taskClosure)
					assert.NoError(t, err)
					return models.Task{
						TaskKey: models.TaskKey{
							Project: "flytekit",
							Domain:  "production",
							Name:    "simple_task",
							Version: "12345",
						},
						Closure: serializedTaskClosure,
						Digest:  []byte("simple_task"),
						Type:    "python",
					}, nil
				})

			workflowClosure := testutils.GetWorkflowClosure()
			node := workflowClosure.CompiledWorkflow.Primary.Template.Nodes[0]
			node2 := workflowClosure.CompiledWorkflow.Primary.Template.Nodes[1]
			switch tt.targetType.Target.(type) {
			case *core.Node_TaskNode:
				node.Target = &core.Node_TaskNode{
					TaskNode: &core.TaskNode{
						Reference: &core.TaskNode_ReferenceId{
							ReferenceId: &core.Identifier{Name: "ref_1"},
						},
					},
				}
			case *core.Node_ArrayNode:
				node.Target = &core.Node_ArrayNode{
					ArrayNode: &core.ArrayNode{
						Node: &core.Node{
							Id: "subnode",
							Target: &core.Node_TaskNode{
								TaskNode: &core.TaskNode{
									Reference: &core.TaskNode_ReferenceId{
										ReferenceId: &core.Identifier{Name: "ref_1"},
									},
								},
							},
						},
					},
				}
			case *core.Node_WorkflowNode:
				node.Target = &core.Node_WorkflowNode{
					WorkflowNode: &core.WorkflowNode{
						Reference: &core.WorkflowNode_SubWorkflowRef{
							SubWorkflowRef: &core.Identifier{
								Project: "flytekit",
								Domain:  "production",
								Name:    "nested_sub_workflow",
								Version: "12345",
							},
						},
					},
				}
				node2.Target = &core.Node_WorkflowNode{
					WorkflowNode: &core.WorkflowNode{
						Reference: &core.WorkflowNode_SubWorkflowRef{
							SubWorkflowRef: &core.Identifier{
								Project: "flytekit",
								Domain:  "production",
								Name:    "sub_workflow",
								Version: "12345",
							},
						},
					},
				}
				workflowClosure.CompiledWorkflow.SubWorkflows = []*core.CompiledWorkflow{
					{
						Template: &core.WorkflowTemplate{
							Id: &core.Identifier{
								Project: "flytekit",
								Domain:  "production",
								Name:    "sub_workflow",
								Version: "12345",
							},
							Nodes: []*core.Node{
								{
									Id: subWorkflowNodeID1,
									Target: &core.Node_TaskNode{
										TaskNode: &core.TaskNode{
											Reference: &core.TaskNode_ReferenceId{
												ReferenceId: &core.Identifier{Name: "ref_1"},
											},
										},
									},
								},
							},
						},
					},
					{
						Template: &core.WorkflowTemplate{
							Id: &core.Identifier{
								Project: "flytekit",
								Domain:  "production",
								Name:    "nested_sub_workflow",
								Version: "12345",
							},
							Nodes: []*core.Node{
								{
									Id: subWorkflowNodeID,
									Target: &core.Node_WorkflowNode{
										WorkflowNode: &core.WorkflowNode{
											Reference: &core.WorkflowNode_SubWorkflowRef{
												SubWorkflowRef: &core.Identifier{
													Project: "flytekit",
													Domain:  "production",
													Name:    "sub_workflow",
													Version: "12345",
												},
											},
										},
									},
								},
							},
						},
					},
				}
			}

			// override the created workflow
			mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).ReadProtobufCb =
				func(ctx context.Context, reference storage.DataReference, msg proto.Message) error {
					workflowBytes, _ := proto.Marshal(workflowClosure)
					_ = proto.Unmarshal(workflowBytes, msg)
					return nil
				}
			err = mockStorage.WriteProtobuf(context.Background(), remoteClosureIdentifier, defaultStorageOptions, workflowClosure)
			assert.Nil(t, err)

			workflowManager := NewWorkflowManager(
				repository,
				getMockWorkflowConfigProvider(), getMockWorkflowCompiler(), mockStorage,
				storagePrefix, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil))
			namedEntityManager := NewNamedEntityManager(repository, getMockConfigForNETest(), mockScope.NewTestScope())

			lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil), getMockPluginRegistry(), mockStorage, workflowManager, namedEntityManager)

			response, err := lpManager.CreateLaunchPlanFromNode(context.Background(), createLaunchPlanFromNodeRequest)
			if tt.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, response)
				inputVars := map[string]*core.Variable{}
				for k, v := range response.LaunchPlan.Closure.ExpectedInputs.Parameters {
					inputVars[k] = v.Var
				}
				inputMap := &core.VariableMap{
					Variables: inputVars,
				}
				assert.True(t, proto.Equal(tt.inputVariableMap, inputMap))
				assert.True(t, proto.Equal(tt.outputVariableMap, response.LaunchPlan.Closure.ExpectedOutputs))
			}
		})
	}
}
