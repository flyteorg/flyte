package impl

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/flyteorg/flyteidl/clients/go/coreutils"

	"github.com/flyteorg/flyteadmin/pkg/async/schedule/mocks"

	scheduleInterfaces "github.com/flyteorg/flyteadmin/pkg/async/schedule/interfaces"
	"github.com/golang/protobuf/ptypes"

	"github.com/flyteorg/flyteadmin/pkg/common"
	flyteAdminErrors "github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	repositoryMocks "github.com/flyteorg/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"
	"github.com/golang/protobuf/proto"

	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	runtimeMocks "github.com/flyteorg/flyteadmin/pkg/runtime/mocks"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	mockScope "github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

var active = int32(admin.LaunchPlanState_ACTIVE)
var inactive = int32(admin.LaunchPlanState_INACTIVE)
var mockScheduler = mocks.NewMockEventScheduler()
var launchPlanIdentifier = core.Identifier{
	ResourceType: core.ResourceType_LAUNCH_PLAN,
	Project:      project,
	Domain:       domain,
	Name:         name,
	Version:      version,
}

var launchPlanNamedIdentifier = core.Identifier{
	Project: project,
	Domain:  domain,
	Name:    name,
	Version: "version",
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
			assert.Equal(t, []byte{0xc9, 0xa9, 0x1b, 0xf3, 0x0, 0x65, 0xe5, 0xce, 0xdb, 0xde, 0xbe, 0x14, 0x1b, 0x9b,
				0x60, 0x8d, 0xeb, 0x69, 0x47, 0x69, 0xed, 0x82, 0xae, 0x2c, 0xde, 0x11, 0x70, 0xba, 0xdc, 0x11, 0xe8, 0xdb}, input.Digest)
			createCalled = true
			return nil
		})
	setDefaultWorkflowCallbackForLpTest(repository)
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
	request := testutils.GetLaunchPlanRequest()
	response, err := lpManager.CreateLaunchPlan(context.Background(), request)
	assert.Nil(t, err)

	expectedResponse := &admin.LaunchPlanCreateResponse{}
	assert.Equal(t, expectedResponse, response)
	assert.True(t, createCalled)
}

func TestLaunchPlanManager_GetLaunchPlan(t *testing.T) {
	repository := getMockRepositoryForLpTest()
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
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
	response, err := lpManager.GetLaunchPlan(context.Background(), admin.ObjectGetRequest{
		Id: &launchPlanIdentifier,
	})
	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestLaunchPlanManager_GetActiveLaunchPlan(t *testing.T) {
	repository := getMockRepositoryForLpTest()
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
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
		assert.Len(t, input.InlineFilters, 4)

		projectExpr, _ := input.InlineFilters[0].GetGormQueryExpr()
		domainExpr, _ := input.InlineFilters[1].GetGormQueryExpr()
		nameExpr, _ := input.InlineFilters[2].GetGormQueryExpr()
		activeExpr, _ := input.InlineFilters[3].GetGormQueryExpr()

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
	response, err := lpManager.GetActiveLaunchPlan(context.Background(), admin.ActiveLaunchPlanRequest{
		Id: &admin.NamedEntityIdentifier{
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
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
	lpRequest := testutils.GetLaunchPlanRequest()

	launchPlanListFunc := func(input interfaces.ListResourceInput) (interfaces.LaunchPlanCollectionOutput, error) {
		return interfaces.LaunchPlanCollectionOutput{}, nil
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetListCallback(launchPlanListFunc)
	response, err := lpManager.GetActiveLaunchPlan(context.Background(), admin.ActiveLaunchPlanRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: lpRequest.Id.Project,
			Domain:  lpRequest.Id.Domain,
			Name:    lpRequest.Id.Name,
		},
	})
	assert.EqualError(t, err, "No active launch plan could be found: project:domain:name")
	assert.Nil(t, response)
}

func TestLaunchPlanManager_GetActiveLaunchPlan_InvalidRequest(t *testing.T) {
	repository := getMockRepositoryForLpTest()
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
	response, err := lpManager.GetActiveLaunchPlan(context.Background(), admin.ActiveLaunchPlanRequest{
		Id: &admin.NamedEntityIdentifier{
			Domain: domain,
			Name:   name,
		},
	})
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestLaunchPlan_ValidationError(t *testing.T) {
	lpManager := NewLaunchPlanManager(repositoryMocks.NewMockRepository(), getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
	request := testutils.GetLaunchPlanRequest()
	request.Id = nil
	response, err := lpManager.CreateLaunchPlan(context.Background(), request)
	assert.EqualError(t, err, "missing id")
	assert.Nil(t, response)
}

func TestLaunchPlanManager_CreateLaunchPlanErrorDueToBadLabels(t *testing.T) {
	repository := getMockRepositoryForLpTest()
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
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
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
	request := testutils.GetLaunchPlanRequest()
	response, err := lpManager.CreateLaunchPlan(context.Background(), request)
	assert.EqualError(t, err, expectedErr.Error())
	assert.Nil(t, response)
}

func TestCreateLaunchPlanInCompatibleInputs(t *testing.T) {
	repository := getMockRepositoryForLpTest()
	setDefaultWorkflowCallbackForLpTest(repository)
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
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
		assert.EqualValues(t, testutils.GetLaunchPlanRequest().Spec, launchPlan.Spec)
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
		assert.EqualValues(t, expectedInputs, launchPlan.Closure.ExpectedInputs)
		assert.EqualValues(
			t, testutils.GetSampleWorkflowSpecForTest().Template.Interface.Outputs,
			launchPlan.Closure.ExpectedOutputs)
		return nil
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetCreateCallback(lpCreateFunc)

	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
	request := testutils.GetLaunchPlanRequest()
	response, err := lpManager.CreateLaunchPlan(context.Background(), request)
	assert.Nil(t, err)

	expectedResponse := &admin.LaunchPlanCreateResponse{}
	assert.Equal(t, expectedResponse, response)
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

	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
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

func TestEnableSchedule(t *testing.T) {
	repository := getMockRepositoryForLpTest()
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
			assert.True(t, proto.Equal(&launchPlanNamedIdentifier, &input.Identifier))
			assert.True(t, proto.Equal(&scheduleExpression, &input.ScheduleExpression))
			assert.Equal(t, "{\"time\":<time>,\"kickoff_time_arg\":\"\",\"payload\":"+
				"\"Cgdwcm9qZWN0EgZkb21haW4aBG5hbWU=\"}",
				*input.Payload)
			return nil
		})
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
	err := lpManager.(*LaunchPlanManager).enableSchedule(
		context.Background(),
		launchPlanNamedIdentifier,
		admin.LaunchPlanSpec{
			EntityMetadata: &admin.LaunchPlanMetadata{
				Schedule: &scheduleExpression,
			},
		})
	assert.Nil(t, err)
}

func TestEnableSchedule_Error(t *testing.T) {
	expectedErr := errors.New("expected error")

	repository := getMockRepositoryForLpTest()
	mockScheduler := mocks.NewMockEventScheduler()
	mockScheduler.(*mocks.MockEventScheduler).SetAddScheduleFunc(
		func(ctx context.Context, input scheduleInterfaces.AddScheduleInput) error {
			return expectedErr
		})
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
	err := lpManager.(*LaunchPlanManager).enableSchedule(
		context.Background(),
		launchPlanNamedIdentifier, admin.LaunchPlanSpec{
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
			assert.True(t, proto.Equal(&launchPlanNamedIdentifier, &input.Identifier))
			return nil
		})
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
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
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
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
			assert.True(t, proto.Equal(&launchPlanNamedIdentifier, &input.Identifier))
			removeCalled = true
			return nil
		})
	var addCalled bool
	mockScheduler.(*mocks.MockEventScheduler).SetAddScheduleFunc(
		func(ctx context.Context, input scheduleInterfaces.AddScheduleInput) error {
			assert.True(t, proto.Equal(&launchPlanNamedIdentifier, &input.Identifier))
			assert.True(t, proto.Equal(&newScheduleExpression, &input.ScheduleExpression))
			addCalled = true
			return nil
		})
	repository := getMockRepositoryForLpTest()
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
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
			}, &input.Identifier))
			assert.True(t, proto.Equal(&newScheduleExpression, &input.ScheduleExpression))
			addCalled = true
			return nil
		})
	repository := getMockRepositoryForLpTest()
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
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
			}, &input.Identifier)
			assert.True(t, areEqual)
			removeCalled = true
			return nil
		})

	repository := getMockRepositoryForLpTest()
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
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
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
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
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
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
			assert.True(t, proto.Equal(&launchPlanNamedIdentifier, &input.Identifier))
			removeScheduleFuncCalled = true
			return nil
		})

	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetUpdateCallback(disableFunc)

	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
	_, err := lpManager.UpdateLaunchPlan(context.Background(), admin.LaunchPlanUpdateRequest{
		Id:    &launchPlanIdentifier,
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
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
	_, err := lpManager.UpdateLaunchPlan(context.Background(), admin.LaunchPlanUpdateRequest{
		Id:    &launchPlanIdentifier,
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
	lpManager = NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
	_, err = lpManager.UpdateLaunchPlan(context.Background(), admin.LaunchPlanUpdateRequest{
		Id:    &launchPlanIdentifier,
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

	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
	_, err := lpManager.UpdateLaunchPlan(context.Background(), admin.LaunchPlanUpdateRequest{
		Id:    &launchPlanIdentifier,
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

	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
	_, err := lpManager.UpdateLaunchPlan(context.Background(), admin.LaunchPlanUpdateRequest{
		Id:    &launchPlanIdentifier,
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
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
	_, err := lpManager.UpdateLaunchPlan(context.Background(), admin.LaunchPlanUpdateRequest{
		Id:    &launchPlanIdentifier,
		State: admin.LaunchPlanState_ACTIVE,
	})
	assert.EqualError(t, err, expectedError.Error(), "Failures on getting the existing launch plan should propagate")

	lpGetFunc = makeLaunchPlanRepoGetCallback(t)
	lpManager = NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
	listFunc := func(input interfaces.ListResourceInput) (interfaces.LaunchPlanCollectionOutput, error) {
		return interfaces.LaunchPlanCollectionOutput{}, expectedError
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetListCallback(listFunc)
	_, err = lpManager.UpdateLaunchPlan(context.Background(), admin.LaunchPlanUpdateRequest{
		Id:    &launchPlanIdentifier,
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
	lpManager = NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
	_, err = lpManager.UpdateLaunchPlan(context.Background(), admin.LaunchPlanUpdateRequest{
		Id:    &launchPlanIdentifier,
		State: admin.LaunchPlanState_ACTIVE,
	})
	assert.EqualError(t, err, expectedError.Error(), "Errors on setting the desired launch plan to active should propagate")
}

func TestLaunchPlanManager_ListLaunchPlans(t *testing.T) {
	repository := getMockRepositoryForLpTest()
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
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
	lpList, err := lpManager.ListLaunchPlans(context.Background(), admin.ResourceListRequest{
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
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
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
	lpList, err := lpManager.ListLaunchPlanIds(context.Background(), admin.NamedEntityIdentifierListRequest{
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
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
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
	lpList, err := lpManager.ListActiveLaunchPlans(context.Background(), admin.ActiveLaunchPlanListRequest{
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
	lpManager := NewLaunchPlanManager(repository, getMockConfigForLpTest(), mockScheduler, mockScope.NewTestScope())
	lpList, err := lpManager.ListActiveLaunchPlans(context.Background(), admin.ActiveLaunchPlanListRequest{
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
		sp := admin.LaunchPlanSpec{
			EntityMetadata: &admin.LaunchPlanMetadata{
				Schedule: &admin.Schedule{
					ScheduleExpression: &admin.Schedule_CronExpression{},
				},
			},
		}
		assert.True(t, isScheduleEmpty(sp))
	})
	t.Run("deprecated cron expression used", func(t *testing.T) {
		sp := admin.LaunchPlanSpec{
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
		sp := admin.LaunchPlanSpec{
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
		sp := admin.LaunchPlanSpec{
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
		sp := admin.LaunchPlanSpec{
			EntityMetadata: &admin.LaunchPlanMetadata{
				Schedule: &admin.Schedule{
					KickoffTimeInputArg: "name",
				},
			},
		}
		assert.True(t, isScheduleEmpty(sp))
	})
}
