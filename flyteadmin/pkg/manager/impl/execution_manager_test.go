package impl

import (
	"context"
	"errors"
	"testing"

	"github.com/lyft/flyteadmin/pkg/common"
	commonMocks "github.com/lyft/flyteadmin/pkg/common/mocks"

	"github.com/benbjohnson/clock"

	"github.com/lyft/flytestdlib/storage"

	"time"

	"github.com/golang/protobuf/ptypes"

	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	notificationMocks "github.com/lyft/flyteadmin/pkg/async/notifications/mocks"
	dataMocks "github.com/lyft/flyteadmin/pkg/data/mocks"
	flyteAdminErrors "github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/manager/impl/executions"
	"github.com/lyft/flyteadmin/pkg/manager/impl/shared"
	"github.com/lyft/flyteadmin/pkg/manager/impl/testutils"
	"github.com/lyft/flyteadmin/pkg/repositories"
	"github.com/lyft/flyteadmin/pkg/repositories/interfaces"
	repositoryMocks "github.com/lyft/flyteadmin/pkg/repositories/mocks"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	runtimeInterfaces "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	runtimeMocks "github.com/lyft/flyteadmin/pkg/runtime/mocks"
	workflowengineInterfaces "github.com/lyft/flyteadmin/pkg/workflowengine/interfaces"
	workflowengineMocks "github.com/lyft/flyteadmin/pkg/workflowengine/mocks"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/lyft/flytepropeller/pkg/utils"
	mockScope "github.com/lyft/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

var spec = testutils.GetExecutionRequest().Spec
var specBytes, _ = proto.Marshal(spec)
var phase = core.WorkflowExecution_RUNNING.String()
var closure = admin.ExecutionClosure{
	Phase: core.WorkflowExecution_RUNNING,
}
var closureBytes, _ = proto.Marshal(&closure)

var executionIdentifier = core.WorkflowExecutionIdentifier{
	Project: "project",
	Domain:  "domain",
	Name:    "name",
}
var mockPublisher notificationMocks.MockPublisher
var mockExecutionRemoteURL = dataMocks.NewMockRemoteURL()
var requestedAt = time.Now()
var testCluster = "C1"
var outputURI = "output uri"

func getLegacySpec() *admin.ExecutionSpec {
	executionRequest := testutils.GetExecutionRequest()
	legacySpec := executionRequest.Spec
	legacySpec.Inputs = executionRequest.Inputs
	return legacySpec
}

func getLegacySpecBytes() []byte {
	b, _ := proto.Marshal(getLegacySpec())
	return b
}

func getLegacyClosure() *admin.ExecutionClosure {
	return &admin.ExecutionClosure{
		Phase:          core.WorkflowExecution_RUNNING,
		ComputedInputs: getLegacySpec().Inputs,
	}
}

func getLegacyClosureBytes() []byte {
	b, _ := proto.Marshal(getLegacyClosure())
	return b
}

func getLegacyExecutionRequest() *admin.ExecutionCreateRequest {
	r := testutils.GetExecutionRequest()
	r.Spec.Inputs = r.Inputs
	r.Inputs = nil
	return &r
}

func getMockExecutionsConfigProvider() runtimeInterfaces.Configuration {
	mockExecutionsConfigProvider := runtimeMocks.NewMockConfigurationProvider(
		testutils.GetApplicationConfigWithDefaultProjects(),
		runtimeMocks.NewMockQueueConfigurationProvider(
			[]runtimeInterfaces.ExecutionQueue{}, []runtimeInterfaces.WorkflowConfig{}),
		nil, nil, nil, nil)
	mockExecutionsConfigProvider.(*runtimeMocks.MockConfigurationProvider).AddRegistrationValidationConfiguration(
		runtimeMocks.NewMockRegistrationValidationProvider())
	return mockExecutionsConfigProvider
}

func setDefaultLpCallbackForExecTest(repository repositories.RepositoryInterface) {
	lpSpec := testutils.GetSampleLpSpecForTest()
	lpSpec.Labels = &admin.Labels{
		Values: map[string]string{
			"label1": "1",
			"label2": "2",
		},
	}
	lpSpec.Annotations = &admin.Annotations{
		Values: map[string]string{
			"annotation3": "3",
			"annotation4": "4",
		},
	}
	lpSpecBytes, _ := proto.Marshal(&lpSpec)
	lpClosure := admin.LaunchPlanClosure{
		ExpectedInputs: lpSpec.DefaultInputs,
	}
	lpClosureBytes, _ := proto.Marshal(&lpClosure)

	lpGetFunc := func(input interfaces.GetResourceInput) (models.LaunchPlan, error) {
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
			Spec:    lpSpecBytes,
			Closure: lpClosureBytes,
		}
		return lpModel, nil
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetGetCallback(lpGetFunc)
}

func getMockStorageForExecTest(ctx context.Context) *storage.DataStore {
	mockStorage := commonMocks.GetMockStorageClient()
	mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).ReadProtobufCb = func(
		ctx context.Context, reference storage.DataReference, msg proto.Message) error {
		if val, ok := mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).Store[reference]; ok {
			_ = proto.Unmarshal(val, msg)
			return nil
		}
		return fmt.Errorf("could not find value in storage [%v]", reference.String())
	}
	mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).WriteProtobufCb = func(
		ctx context.Context, reference storage.DataReference, opts storage.Options, msg proto.Message) error {
		bytes, err := proto.Marshal(msg)
		if err != nil {
			return err
		}
		mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).Store[reference] = bytes
		return nil
	}
	workflowClosure := testutils.GetWorkflowClosure()
	if err := mockStorage.WriteProtobuf(ctx, storage.DataReference(remoteClosureIdentifier), defaultStorageOptions, workflowClosure); err != nil {
		return nil
	}
	return mockStorage
}

func getMockRepositoryForExecTest() repositories.RepositoryInterface {
	repository := repositoryMocks.NewMockRepository()
	repository.WorkflowRepo().(*repositoryMocks.MockWorkflowRepo).SetGetCallback(
		func(input interfaces.GetResourceInput) (models.Workflow, error) {
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
		})
	return repository
}

func TestCreateExecution(t *testing.T) {
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	mockExecutor := workflowengineMocks.NewMockExecutor()
	mockExecutor.(*workflowengineMocks.MockExecutor).SetExecuteWorkflowCallback(
		func(inputs workflowengineInterfaces.ExecuteWorkflowInput) (*workflowengineInterfaces.ExecutionInfo, error) {
			assert.EqualValues(t, map[string]string{
				"label1": "1",
				"label2": "2",
			}, inputs.Labels)
			assert.EqualValues(t, map[string]string{
				"annotation3": "3",
				"annotation4": "4",
			}, inputs.Annotations)
			return &workflowengineInterfaces.ExecutionInfo{
				Cluster: testCluster,
			}, nil
		})
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockExecutor,
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	request := testutils.GetExecutionRequest()
	response, err := execManager.CreateExecution(context.Background(), request, requestedAt)
	assert.Nil(t, err)

	expectedResponse := &admin.ExecutionCreateResponse{
		Id: &executionIdentifier,
	}
	assert.Nil(t, err)
	assert.Equal(t, expectedResponse, response)

	// TODO: Check for offloaded inputs
}

func TestCreateExecutionFromWorkflowNode(t *testing.T) {
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)

	parentNodeExecutionID := core.NodeExecutionIdentifier{
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "parent-name",
		},
		NodeId: "node-name",
	}

	getNodeExecutionCalled := false
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.GetNodeExecutionInput) (models.NodeExecution, error) {
			assert.EqualValues(t, input.NodeExecutionIdentifier, parentNodeExecutionID)
			getNodeExecutionCalled = true
			return models.NodeExecution{
				BaseModel: models.BaseModel{
					ID: 1,
				},
			}, nil
		},
	)
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(
		func(ctx context.Context, input models.Execution) error {
			assert.Equal(t, input.ParentNodeExecutionID, uint(1))
			return nil
		},
	)

	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	request := testutils.GetExecutionRequest()
	request.Spec.Metadata = &admin.ExecutionMetadata{
		Mode:                admin.ExecutionMetadata_SYSTEM,
		ParentNodeExecution: &parentNodeExecutionID,
	}
	response, err := execManager.CreateExecution(context.Background(), request, requestedAt)
	assert.Nil(t, err)
	assert.True(t, getNodeExecutionCalled)
	expectedResponse := &admin.ExecutionCreateResponse{
		Id: &executionIdentifier,
	}
	assert.Nil(t, err)
	assert.Equal(t, expectedResponse, response)
}

func TestCreateExecution_NoAssignedName(t *testing.T) {
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(
		func(ctx context.Context, input models.Execution) error {
			assert.Equal(t, executionIdentifier.Project, input.Project)
			assert.Equal(t, executionIdentifier.Domain, input.Domain)
			assert.NotEmpty(t, input.Name)
			return nil
		})
	mockExecutor := workflowengineMocks.NewMockExecutor()
	mockExecutor.(*workflowengineMocks.MockExecutor).SetExecuteWorkflowCallback(
		func(inputs workflowengineInterfaces.ExecuteWorkflowInput) (*workflowengineInterfaces.ExecutionInfo, error) {
			assert.NotEmpty(t, inputs.ExecutionID.Name)
			assert.Equal(t, requestedAt, inputs.AcceptedAt)
			return &workflowengineInterfaces.ExecutionInfo{
				Cluster: testCluster,
			}, nil
		})
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockExecutor, mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	request := testutils.GetExecutionRequest()
	request.Name = ""
	response, err := execManager.CreateExecution(context.Background(), request, requestedAt)
	assert.Nil(t, err)

	expectedResponse := &admin.ExecutionCreateResponse{
		Id: &executionIdentifier,
	}
	assert.Nil(t, err)
	assert.Equal(t, expectedResponse.Id.Project, response.Id.Project)
	assert.Equal(t, expectedResponse.Id.Domain, response.Id.Domain)
	assert.NotEmpty(t, response.Id.Name)
}

func TestCreateExecution_TaggedQueue(t *testing.T) {
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	configProvider := runtimeMocks.NewMockConfigurationProvider(
		testutils.GetApplicationConfigWithDefaultProjects(),
		runtimeMocks.NewMockQueueConfigurationProvider([]runtimeInterfaces.ExecutionQueue{
			{
				Primary:    "primary Q",
				Dynamic:    "dynamic Q",
				Attributes: []string{"tag"},
			},
		}, []runtimeInterfaces.WorkflowConfig{
			{
				Project:      "project",
				Domain:       "domain",
				WorkflowName: "name",
				Tags:         []string{"tag"},
			},
		}),
		nil, nil, nil, nil)
	configProvider.(*runtimeMocks.MockConfigurationProvider).AddRegistrationValidationConfiguration(
		runtimeMocks.NewMockRegistrationValidationProvider())
	mockExecutor := workflowengineMocks.NewMockExecutor()
	mockExecutor.(*workflowengineMocks.MockExecutor).SetExecuteWorkflowCallback(
		func(inputs workflowengineInterfaces.ExecuteWorkflowInput) (*workflowengineInterfaces.ExecutionInfo, error) {
			assert.NotEmpty(t, inputs.WfClosure.Tasks)
			for _, task := range inputs.WfClosure.Tasks {
				assert.Len(t, task.Template.GetContainer().Config, 2)
				assert.Contains(t, parentContainerQueueKey, task.Template.GetContainer().Config[0].Key)
				assert.Contains(t, "primary Q", task.Template.GetContainer().Config[0].Value)
				assert.Contains(t, childContainerQueueKey, task.Template.GetContainer().Config[1].Key)
				assert.Contains(t, "dynamic Q", task.Template.GetContainer().Config[1].Value)
			}
			assert.Equal(t, requestedAt, inputs.AcceptedAt)
			return &workflowengineInterfaces.ExecutionInfo{
				Cluster: testCluster,
			}, nil
		})
	execManager := NewExecutionManager(
		repository, configProvider, getMockStorageForExecTest(context.Background()), mockExecutor, mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)

	request := testutils.GetExecutionRequest()
	response, err := execManager.CreateExecution(context.Background(), request, requestedAt)
	assert.Nil(t, err)

	expectedResponse := &admin.ExecutionCreateResponse{
		Id: &executionIdentifier,
	}
	assert.Nil(t, err)
	assert.Equal(t, expectedResponse, response)
}

func TestCreateExecutionValidationError(t *testing.T) {
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)

	request := testutils.GetExecutionRequest()
	request.Domain = ""
	response, err := execManager.CreateExecution(context.Background(), request, requestedAt)
	assert.EqualError(t, err, "missing domain")
	assert.Nil(t, response)
}

func TestCreateExecution_InvalidLpIdentifier(t *testing.T) {
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)

	request := testutils.GetExecutionRequest()
	request.Spec.LaunchPlan = nil
	response, err := execManager.CreateExecution(context.Background(), request, requestedAt)
	assert.EqualError(t, err, "missing id")
	assert.Nil(t, response)
}

func TestCreateExecutionInCompatibleInputs(t *testing.T) {
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)

	request := testutils.GetExecutionRequest()
	request.Inputs = &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"foo-1": utils.MustMakeLiteral("foo-value-1"),
		},
	}
	response, err := execManager.CreateExecution(context.Background(), request, requestedAt)
	assert.EqualError(t, err, "invalid input foo-1")
	assert.Nil(t, response)
}

func TestCreateExecutionPropellerFailure(t *testing.T) {
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	mockExecutor := workflowengineMocks.NewMockExecutor()
	expectedErr := flyteAdminErrors.NewFlyteAdminErrorf(codes.Internal, "ABC")

	createFunc := func(inputs workflowengineInterfaces.ExecuteWorkflowInput) (*workflowengineInterfaces.ExecutionInfo, error) {
		assert.Equal(t, requestedAt, inputs.AcceptedAt)
		return nil, expectedErr
	}
	mockExecutor.(*workflowengineMocks.MockExecutor).SetExecuteWorkflowCallback(createFunc)
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockExecutor, mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)

	request := testutils.GetExecutionRequest()

	response, err := execManager.CreateExecution(context.Background(), request, requestedAt)
	assert.EqualError(t, err, expectedErr.Error())
	assert.Nil(t, response)
}

func TestCreateExecutionDatabaseFailure(t *testing.T) {

	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	expectedErr := flyteAdminErrors.NewFlyteAdminErrorf(codes.Internal, "ABCD")
	exCreateFunc := func(ctx context.Context, input models.Execution) error {
		return expectedErr
	}

	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(exCreateFunc)
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	request := testutils.GetExecutionRequest()

	response, err := execManager.CreateExecution(context.Background(), request, requestedAt)
	assert.EqualError(t, err, expectedErr.Error())
	assert.Nil(t, response)
}

func TestCreateExecutionVerifyDbModel(t *testing.T) {
	request := testutils.GetExecutionRequest()
	repository := getMockRepositoryForExecTest()
	storageClient := getMockStorageForExecTest(context.Background())
	setDefaultLpCallbackForExecTest(repository)
	mockClock := clock.NewMock()
	createdAt := time.Now()
	mockClock.Set(createdAt)
	exCreateFunc := func(ctx context.Context, input models.Execution) error {
		assert.Equal(t, "name", input.Name)
		assert.Equal(t, "domain", input.Domain)
		assert.Equal(t, "project", input.Project)
		assert.Equal(t, uint(100), input.LaunchPlanID)
		assert.Equal(t, core.WorkflowExecution_UNDEFINED.String(), input.Phase)

		var specValue admin.ExecutionSpec
		err := proto.Unmarshal(input.Spec, &specValue)
		if err != nil {
			return err
		}
		assert.Nil(t, specValue.Inputs)

		var closureValue admin.ExecutionClosure
		err = proto.Unmarshal(input.Closure, &closureValue)
		if err != nil {
			return err
		}
		assert.Nil(t, closureValue.ComputedInputs)

		var userInputs, inputs core.LiteralMap
		if err := storageClient.ReadProtobuf(ctx, input.UserInputsURI, &userInputs); err != nil {
			return err
		}
		if err := storageClient.ReadProtobuf(ctx, input.InputsURI, &inputs); err != nil {
			return err
		}
		fooValue := utils.MustMakeLiteral("foo-value-1")
		assert.Equal(t, 1, len(userInputs.Literals))
		assert.EqualValues(t, userInputs.Literals["foo"], fooValue)
		barValue := utils.MustMakeLiteral("bar-value")
		assert.Equal(t, len(inputs.Literals), 2)
		assert.EqualValues(t, inputs.Literals["foo"], fooValue)
		assert.EqualValues(t, inputs.Literals["bar"], barValue)
		assert.Equal(t, core.WorkflowExecution_UNDEFINED, closureValue.Phase)
		assert.Equal(t, createdAt, *input.ExecutionCreatedAt)
		assert.Equal(t, 1, len(closureValue.Notifications))
		assert.Equal(t, 1, len(closureValue.Notifications[0].Phases))
		assert.Equal(t, request.Spec.GetNotifications().Notifications[0].Phases[0], closureValue.Notifications[0].Phases[0])
		assert.IsType(t, &admin.Notification_Slack{}, closureValue.Notifications[0].GetType())
		assert.Equal(t, request.Spec.GetNotifications().Notifications[0].GetSlack().RecipientsEmail, closureValue.Notifications[0].GetSlack().RecipientsEmail)

		return nil
	}

	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(exCreateFunc)
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), storageClient, workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)

	execManager.(*ExecutionManager)._clock = mockClock

	response, err := execManager.CreateExecution(context.Background(), request, requestedAt)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&executionIdentifier, response.Id))
}

func TestCreateExecutionDefaultNotifications(t *testing.T) {
	// Remove notifications settings for the CreateExecutionRequest.
	request := testutils.GetExecutionRequest()
	request.Spec.NotificationOverrides = &admin.ExecutionSpec_Notifications{
		Notifications: &admin.NotificationList{
			Notifications: []*admin.Notification{},
		},
	}

	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)

	// Create a callback method to ensure the default notification settings from the LaunchPlan is
	// stored in the resulting models.Execution.
	exCreateFunc := func(ctx context.Context, input models.Execution) error {
		var closureValue admin.ExecutionClosure
		err := proto.Unmarshal(input.Closure, &closureValue)
		if err != nil {
			return err
		}

		assert.Equal(t, 1, len(closureValue.Notifications))
		assert.Equal(t, 1, len(closureValue.Notifications[0].Phases))
		assert.Equal(t, core.WorkflowExecution_SUCCEEDED, closureValue.Notifications[0].Phases[0])
		assert.IsType(t, &admin.Notification_Email{}, closureValue.Notifications[0].GetType())

		return nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(exCreateFunc)
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(), mockScope.NewTestScope(),
		mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)

	response, err := execManager.CreateExecution(context.Background(), request, requestedAt)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&core.WorkflowExecutionIdentifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
	}, response.Id))
}

func TestCreateExecutionDisableNotifications(t *testing.T) {
	// Disable notifications for the CreateExecutionRequest.
	request := testutils.GetExecutionRequest()
	request.Spec.NotificationOverrides = &admin.ExecutionSpec_DisableAll{
		DisableAll: true,
	}

	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)

	// Create a callback method to ensure the default notification settings from the LaunchPlan is
	// stored in the resulting models.Execution.
	exCreateFunc := func(ctx context.Context, input models.Execution) error {
		var closureValue admin.ExecutionClosure
		err := proto.Unmarshal(input.Closure, &closureValue)
		if err != nil {
			return err
		}

		assert.Empty(t, closureValue.Notifications)
		return nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(exCreateFunc)
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(), mockScope.NewTestScope(),
		mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)

	response, err := execManager.CreateExecution(context.Background(), request, requestedAt)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&core.WorkflowExecutionIdentifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
	}, response.Id))
}

func TestCreateExecutionNoNotifications(t *testing.T) {
	// Remove notifications settings for the CreateExecutionRequest.
	request := testutils.GetExecutionRequest()
	request.Spec.NotificationOverrides = &admin.ExecutionSpec_Notifications{
		Notifications: &admin.NotificationList{
			Notifications: []*admin.Notification{},
		},
	}

	// Remove notifications settings for the LaunchPlan associated with the
	// CreateExecutionRequest.
	lpSpec := testutils.GetSampleLpSpecForTest()
	lpSpec.EntityMetadata.Notifications = nil
	lpSpecBytes, _ := proto.Marshal(&lpSpec)
	lpClosure := admin.LaunchPlanClosure{
		ExpectedInputs: lpSpec.DefaultInputs,
	}
	lpClosureBytes, _ := proto.Marshal(&lpClosure)

	// The LaunchPlan is retrieved within the CreateExecution call to ExecutionManager.
	// Create a callback method used by the mock to retrieve a LaunchPlan.
	lpGetFunc := func(input interfaces.GetResourceInput) (models.LaunchPlan, error) {
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
			Spec:    lpSpecBytes,
			Closure: lpClosureBytes,
		}
		return lpModel, nil
	}

	repository := getMockRepositoryForExecTest()
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetGetCallback(lpGetFunc)

	// Create a callback method to validate no notifications are set when storing the
	// resulting models.Execution by CreateExecution.
	exCreateFunc := func(ctx context.Context, input models.Execution) error {

		var closureValue admin.ExecutionClosure
		err := proto.Unmarshal(input.Closure, &closureValue)
		if err != nil {
			return err
		}
		assert.Nil(t, closureValue.GetNotifications())
		return nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(exCreateFunc)
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)

	response, err := execManager.CreateExecution(context.Background(), request, requestedAt)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&core.WorkflowExecutionIdentifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
	}, response.Id))
}

func TestCreateExecutionDynamicLabelsAndAnnotations(t *testing.T) {
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	mockExecutor := workflowengineMocks.NewMockExecutor()
	mockExecutor.(*workflowengineMocks.MockExecutor).SetExecuteWorkflowCallback(
		func(inputs workflowengineInterfaces.ExecuteWorkflowInput) (*workflowengineInterfaces.ExecutionInfo, error) {
			assert.EqualValues(t, map[string]string{
				"dynamiclabel1": "dynamic1",
				"dynamiclabel2": "dynamic2",
			}, inputs.Labels)
			assert.EqualValues(t, map[string]string{
				"dynamicannotation3": "dynamic3",
				"dynamicannotation4": "dynamic4",
			}, inputs.Annotations)
			return &workflowengineInterfaces.ExecutionInfo{
				Cluster: testCluster,
			}, nil
		})
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockExecutor,
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	request := testutils.GetExecutionRequest()
	request.Spec.Labels = &admin.Labels{
		Values: map[string]string{
			"dynamiclabel1": "dynamic1",
			"dynamiclabel2": "dynamic2",
		},
	}
	request.Spec.Annotations = &admin.Annotations{
		Values: map[string]string{
			"dynamicannotation3": "dynamic3",
			"dynamicannotation4": "dynamic4",
		},
	}
	response, err := execManager.CreateExecution(context.Background(), request, requestedAt)
	assert.Nil(t, err)

	expectedResponse := &admin.ExecutionCreateResponse{
		Id: &executionIdentifier,
	}
	assert.Nil(t, err)
	assert.Equal(t, expectedResponse, response)
}

func makeExecutionGetFunc(
	t *testing.T, closureBytes []byte, startTime *time.Time) repositoryMocks.GetExecutionFunc {
	return func(ctx context.Context, input interfaces.GetResourceInput) (models.Execution, error) {
		assert.Equal(t, "project", input.Project)
		assert.Equal(t, "domain", input.Domain)
		assert.Equal(t, "name", input.Name)
		return models.Execution{
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
			BaseModel: models.BaseModel{
				ID: uint(8),
			},
			Spec:         specBytes,
			Phase:        core.WorkflowExecution_QUEUED.String(),
			Closure:      closureBytes,
			LaunchPlanID: uint(1),
			WorkflowID:   uint(2),
			StartedAt:    startTime,
			Cluster:      testCluster,
		}, nil
	}
}

func makeLegacyExecutionGetFunc(
	t *testing.T, closureBytes []byte, startTime *time.Time) repositoryMocks.GetExecutionFunc {
	return func(ctx context.Context, input interfaces.GetResourceInput) (models.Execution, error) {
		assert.Equal(t, "project", input.Project)
		assert.Equal(t, "domain", input.Domain)
		assert.Equal(t, "name", input.Name)
		return models.Execution{
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
			BaseModel: models.BaseModel{
				ID: uint(8),
			},
			Spec:         getLegacySpecBytes(),
			Phase:        core.WorkflowExecution_QUEUED.String(),
			Closure:      closureBytes,
			LaunchPlanID: uint(1),
			WorkflowID:   uint(2),
			StartedAt:    startTime,
			Cluster:      testCluster,
		}, nil
	}
}

func TestRelaunchExecution(t *testing.T) {
	// Set up mocks.
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	startTime := time.Now()
	startTimeProto, _ := ptypes.TimestampProto(startTime)
	existingClosure := admin.ExecutionClosure{
		Phase:     core.WorkflowExecution_RUNNING,
		StartedAt: startTimeProto,
	}
	existingClosureBytes, _ := proto.Marshal(&existingClosure)
	executionGetFunc := makeExecutionGetFunc(t, existingClosureBytes, &startTime)
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)

	var createCalled bool
	exCreateFunc := func(ctx context.Context, input models.Execution) error {
		createCalled = true
		assert.Equal(t, "relaunchy", input.Name)
		assert.Equal(t, "domain", input.Domain)
		assert.Equal(t, "project", input.Project)
		assert.Equal(t, uint(8), input.SourceExecutionID)
		var spec admin.ExecutionSpec
		err := proto.Unmarshal(input.Spec, &spec)
		assert.Nil(t, err)
		assert.Equal(t, admin.ExecutionMetadata_RELAUNCH, spec.Metadata.Mode)
		assert.Equal(t, int32(admin.ExecutionMetadata_RELAUNCH), input.Mode)
		return nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(exCreateFunc)

	// Issue request.
	response, err := execManager.RelaunchExecution(context.Background(), admin.ExecutionRelaunchRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Name: "relaunchy",
	}, requestedAt)

	// And verify response.
	assert.Nil(t, err)

	expectedResponse := &admin.ExecutionCreateResponse{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "relaunchy",
		},
	}
	assert.True(t, createCalled)
	assert.True(t, proto.Equal(expectedResponse, response))

	// TODO: Test with inputs
}

func TestRelaunchExecution_GetExistingFailure(t *testing.T) {
	// Set up mocks.
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)

	expectedErr := errors.New("expected error")
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.GetResourceInput) (models.Execution, error) {
			return models.Execution{}, expectedErr
		})

	var createCalled bool
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(
		func(ctx context.Context, input models.Execution) error {
			createCalled = true
			return nil
		})

	// Issue request.
	_, err := execManager.RelaunchExecution(context.Background(), admin.ExecutionRelaunchRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Name: "relaunchy",
	}, requestedAt)

	// And verify response.
	assert.EqualError(t, err, expectedErr.Error())
	assert.False(t, createCalled)
}

func TestRelaunchExecution_CreateFailure(t *testing.T) {
	// Set up mocks.
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	startTime := time.Now()
	startTimeProto, _ := ptypes.TimestampProto(startTime)
	existingClosure := admin.ExecutionClosure{
		Phase:     core.WorkflowExecution_RUNNING,
		StartedAt: startTimeProto,
	}
	existingClosureBytes, _ := proto.Marshal(&existingClosure)
	executionGetFunc := makeExecutionGetFunc(t, existingClosureBytes, &startTime)
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)

	expectedErr := errors.New("expected error")
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(
		func(ctx context.Context, input models.Execution) error {
			return expectedErr
		})

	// Issue request.
	_, err := execManager.RelaunchExecution(context.Background(), admin.ExecutionRelaunchRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Name: "relaunchy",
	}, requestedAt)

	// And verify response.
	assert.EqualError(t, err, expectedErr.Error())
}

func TestCreateWorkflowEvent(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	startTime := time.Now()
	startTimeProto, _ := ptypes.TimestampProto(startTime)
	duration := time.Second
	durationProto := ptypes.DurationProto(duration)
	existingClosure := admin.ExecutionClosure{
		Phase:     core.WorkflowExecution_RUNNING,
		StartedAt: startTimeProto,
	}
	existingClosureBytes, _ := proto.Marshal(&existingClosure)
	executionGetFunc := makeExecutionGetFunc(t, existingClosureBytes, &startTime)
	executionError := core.ExecutionError{
		Code:    "foo",
		Message: "bar baz",
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)
	endTime := startTime.Add(duration)
	occurredAt, _ := ptypes.TimestampProto(endTime)
	closure := admin.ExecutionClosure{
		Phase:     core.WorkflowExecution_FAILED,
		StartedAt: startTimeProto,
		UpdatedAt: occurredAt,
		Duration:  durationProto,
		OutputResult: &admin.ExecutionClosure_Error{
			Error: &executionError,
		},
	}
	closureBytes, _ := proto.Marshal(&closure)
	updateExecutionFunc := func(
		context context.Context, event models.ExecutionEvent, execution models.Execution) error {
		assert.Equal(t, models.ExecutionKey{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		}, event.ExecutionKey)
		assert.Equal(t, "1", event.RequestID)
		assert.Equal(t, endTime.Second(), event.OccurredAt.Second())
		assert.Equal(t, endTime.Nanosecond(), event.OccurredAt.Nanosecond())
		assert.Equal(t, core.WorkflowExecution_FAILED.String(), event.Phase)

		assert.Equal(t, "project", execution.Project)
		assert.Equal(t, "domain", execution.Domain)
		assert.Equal(t, "name", execution.Name)
		assert.Equal(t, uint(1), execution.LaunchPlanID)
		assert.Equal(t, uint(2), execution.WorkflowID)
		assert.Equal(t, core.WorkflowExecution_FAILED.String(), execution.Phase)
		assert.Equal(t, closureBytes, execution.Closure)
		assert.Equal(t, specBytes, execution.Spec)
		assert.Equal(t, startTime, *execution.StartedAt)
		assert.Equal(t, duration, execution.Duration)
		return nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetUpdateCallback(updateExecutionFunc)
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	resp, err := execManager.CreateWorkflowEvent(context.Background(), admin.WorkflowExecutionEventRequest{
		RequestId: "1",
		Event: &event.WorkflowExecutionEvent{
			ExecutionId: &executionIdentifier,
			OccurredAt:  occurredAt,
			Phase:       core.WorkflowExecution_FAILED,
			OutputResult: &event.WorkflowExecutionEvent_Error{
				Error: &executionError,
			},
		},
	})
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestCreateWorkflowEvent_TerminalState(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	executionGetFunc := func(ctx context.Context, input interfaces.GetResourceInput) (models.Execution, error) {
		return models.Execution{
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
			BaseModel: models.BaseModel{
				ID: uint(8),
			},
			Spec:  specBytes,
			Phase: core.WorkflowExecution_FAILED.String(),
		}, nil
	}

	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)
	updateExecutionFunc := func(context context.Context, event models.ExecutionEvent, execution models.Execution) error {
		return nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetUpdateCallback(updateExecutionFunc)
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)

	resp, err := execManager.CreateWorkflowEvent(context.Background(), admin.WorkflowExecutionEventRequest{
		RequestId: "1",
		Event: &event.WorkflowExecutionEvent{
			ExecutionId: &executionIdentifier,
			Phase:       core.WorkflowExecution_SUCCEEDED,
		},
	})
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	adminError := err.(flyteAdminErrors.FlyteAdminError)
	assert.Equal(t, adminError.Code(), codes.FailedPrecondition)
	details, ok := adminError.GRPCStatus().Details()[0].(*admin.EventFailureReason)
	assert.True(t, ok)
	_, ok = details.GetReason().(*admin.EventFailureReason_AlreadyInTerminalState)
	assert.True(t, ok)
}

func TestCreateWorkflowEvent_StartedRunning(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	occurredAt := time.Now().UTC()
	occurredAtProto, _ := ptypes.TimestampProto(occurredAt)
	executionGetFunc := makeExecutionGetFunc(t, closureBytes, nil)
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)

	closure := admin.ExecutionClosure{
		Phase:     core.WorkflowExecution_RUNNING,
		StartedAt: occurredAtProto,
		UpdatedAt: occurredAtProto,
	}
	closureBytes, _ := proto.Marshal(&closure)
	updateExecutionFunc := func(
		context context.Context, event models.ExecutionEvent, execution models.Execution) error {
		assert.Equal(t, "1", event.RequestID)
		assert.Equal(t, models.ExecutionKey{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		}, event.ExecutionKey)
		assert.Equal(t, occurredAt, event.OccurredAt)
		assert.Equal(t, "project", execution.Project)
		assert.Equal(t, "domain", execution.Domain)
		assert.Equal(t, "name", execution.Name)
		assert.Equal(t, uint(1), execution.LaunchPlanID)
		assert.Equal(t, uint(2), execution.WorkflowID)
		assert.Equal(t, core.WorkflowExecution_RUNNING.String(), execution.Phase)
		assert.Equal(t, closureBytes, execution.Closure)
		assert.Equal(t, specBytes, execution.Spec)
		assert.Equal(t, occurredAt, *execution.StartedAt)
		assert.Equal(t, time.Duration(0), execution.Duration)
		assert.Equal(t, occurredAt, *execution.ExecutionUpdatedAt)
		return nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetUpdateCallback(updateExecutionFunc)
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	occurredAtTimestamp, _ := ptypes.TimestampProto(occurredAt)
	resp, err := execManager.CreateWorkflowEvent(context.Background(), admin.WorkflowExecutionEventRequest{
		RequestId: "1",
		Event: &event.WorkflowExecutionEvent{
			ExecutionId: &executionIdentifier,
			OccurredAt:  occurredAtTimestamp,
			Phase:       core.WorkflowExecution_RUNNING,
		},
	})
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestCreateWorkflowEvent_DuplicateRunning(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	occurredAt := time.Now().UTC()

	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.GetResourceInput) (models.Execution, error) {
			return models.Execution{
				ExecutionKey: models.ExecutionKey{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
				BaseModel: models.BaseModel{
					ID: uint(8),
				},
				Spec:         specBytes,
				Phase:        core.WorkflowExecution_RUNNING.String(),
				Closure:      closureBytes,
				LaunchPlanID: uint(1),
				WorkflowID:   uint(2),
				StartedAt:    &occurredAt,
			}, nil
		},
	)

	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	occurredAtTimestamp, _ := ptypes.TimestampProto(occurredAt)
	resp, err := execManager.CreateWorkflowEvent(context.Background(), admin.WorkflowExecutionEventRequest{
		RequestId: "1",
		Event: &event.WorkflowExecutionEvent{
			ExecutionId: &executionIdentifier,
			OccurredAt:  occurredAtTimestamp,
			Phase:       core.WorkflowExecution_RUNNING,
		},
	})
	assert.NotNil(t, err)
	adminError := err.(flyteAdminErrors.FlyteAdminError)
	assert.Equal(t, adminError.Code(), codes.AlreadyExists)
	assert.Nil(t, resp)
}

func TestCreateWorkflowEvent_InvalidPhaseChange(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	occurredAt := time.Now().UTC()

	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.GetResourceInput) (models.Execution, error) {
			return models.Execution{
				ExecutionKey: models.ExecutionKey{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
				BaseModel: models.BaseModel{
					ID: uint(8),
				},
				Spec:         specBytes,
				Phase:        core.WorkflowExecution_SUCCEEDED.String(),
				Closure:      closureBytes,
				LaunchPlanID: uint(1),
				WorkflowID:   uint(2),
				StartedAt:    &occurredAt,
			}, nil
		},
	)

	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	occurredAtTimestamp, _ := ptypes.TimestampProto(occurredAt)
	resp, err := execManager.CreateWorkflowEvent(context.Background(), admin.WorkflowExecutionEventRequest{
		RequestId: "1",
		Event: &event.WorkflowExecutionEvent{
			ExecutionId: &executionIdentifier,
			OccurredAt:  occurredAtTimestamp,
			Phase:       core.WorkflowExecution_RUNNING,
		},
	})
	assert.NotNil(t, err)
	adminError := err.(flyteAdminErrors.FlyteAdminError)
	assert.Equal(t, adminError.Code(), codes.FailedPrecondition)
	assert.Nil(t, resp)
	details, ok := adminError.GRPCStatus().Details()[0].(*admin.EventFailureReason)
	assert.True(t, ok)
	_, ok = details.GetReason().(*admin.EventFailureReason_AlreadyInTerminalState)
	assert.True(t, ok)
}

func TestCreateWorkflowEvent_InvalidEvent(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	startTime := time.Now()
	executionGetFunc := makeExecutionGetFunc(t, closureBytes, &startTime)
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)
	executionError := core.ExecutionError{
		Code:    "foo",
		Message: "bar baz",
	}
	updateExecutionFunc := func(
		context context.Context, event models.ExecutionEvent, execution models.Execution) error {
		return nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetUpdateCallback(updateExecutionFunc)
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	resp, err := execManager.CreateWorkflowEvent(context.Background(), admin.WorkflowExecutionEventRequest{
		RequestId: "1",
		Event: &event.WorkflowExecutionEvent{
			ExecutionId: &executionIdentifier,
			Phase:       core.WorkflowExecution_FAILED,
			OutputResult: &event.WorkflowExecutionEvent_Error{
				Error: &executionError,
			},
		},
	})
	assert.NotNil(t, err)
	assert.Nil(t, resp)
}

func TestCreateWorkflowEvent_UpdateModelError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	startTime := time.Now()
	executionGetFunc := makeExecutionGetFunc(t, []byte("invalid serialized closure"), &startTime)
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)
	duration := time.Second
	endTime := startTime.Add(duration)
	occurredAt, _ := ptypes.TimestampProto(endTime)
	executionError := core.ExecutionError{
		Code:    "foo",
		Message: "bar baz",
	}

	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	resp, err := execManager.CreateWorkflowEvent(context.Background(), admin.WorkflowExecutionEventRequest{
		RequestId: "1",
		Event: &event.WorkflowExecutionEvent{
			ExecutionId: &executionIdentifier,
			OccurredAt:  occurredAt,
			Phase:       core.WorkflowExecution_FAILED,
			OutputResult: &event.WorkflowExecutionEvent_Error{
				Error: &executionError,
			},
		},
	})
	assert.Nil(t, resp)
	assert.Equal(t, codes.Internal, err.(flyteAdminErrors.FlyteAdminError).Code())
}

func TestCreateWorkflowEvent_DatabaseGetError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	startTime := time.Now()

	expectedErr := errors.New("expected error")
	executionGetFunc := func(ctx context.Context, input interfaces.GetResourceInput) (models.Execution, error) {
		return models.Execution{}, expectedErr
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)
	duration := time.Second
	endTime := startTime.Add(duration)
	occurredAt, _ := ptypes.TimestampProto(endTime)
	executionError := core.ExecutionError{
		Code:    "foo",
		Message: "bar baz",
	}
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	resp, err := execManager.CreateWorkflowEvent(context.Background(), admin.WorkflowExecutionEventRequest{
		RequestId: "1",
		Event: &event.WorkflowExecutionEvent{
			ExecutionId: &executionIdentifier,
			OccurredAt:  occurredAt,
			Phase:       core.WorkflowExecution_FAILED,
			OutputResult: &event.WorkflowExecutionEvent_Error{
				Error: &executionError,
			},
		},
	})
	assert.Nil(t, resp)
	assert.EqualError(t, expectedErr, err.Error())
}

func TestCreateWorkflowEvent_DatabaseUpdateError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	startTime := time.Now()
	executionGetFunc := makeExecutionGetFunc(t, closureBytes, &startTime)
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)
	duration := time.Second
	endTime := startTime.Add(duration)
	occurredAt, _ := ptypes.TimestampProto(endTime)
	executionError := core.ExecutionError{
		Code:    "foo",
		Message: "bar baz",
	}
	expectedErr := errors.New("expected error")
	updateExecutionFunc := func(
		context context.Context, event models.ExecutionEvent, execution models.Execution) error {
		return expectedErr
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetUpdateCallback(updateExecutionFunc)
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	resp, err := execManager.CreateWorkflowEvent(context.Background(), admin.WorkflowExecutionEventRequest{
		RequestId: "1",
		Event: &event.WorkflowExecutionEvent{
			ExecutionId: &executionIdentifier,
			OccurredAt:  occurredAt,
			Phase:       core.WorkflowExecution_FAILED,
			OutputResult: &event.WorkflowExecutionEvent_Error{
				Error: &executionError,
			},
		},
	})
	assert.Nil(t, resp)
	assert.EqualError(t, expectedErr, err.Error())
}

func TestGetExecution(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	startedAt := time.Date(2018, 8, 30, 0, 0, 0, 0, time.UTC)
	executionGetFunc := func(ctx context.Context, input interfaces.GetResourceInput) (models.Execution, error) {
		assert.Equal(t, "project", input.Project)
		assert.Equal(t, "domain", input.Domain)
		assert.Equal(t, "name", input.Name)
		return models.Execution{
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
			Spec:         specBytes,
			Phase:        phase,
			Closure:      closureBytes,
			LaunchPlanID: uint(1),
			WorkflowID:   uint(2),
			StartedAt:    &startedAt,
			// TODO: Input uri
		}, nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	execution, err := execManager.GetExecution(context.Background(), admin.WorkflowExecutionGetRequest{
		Id: &executionIdentifier,
	})
	assert.NoError(t, err)
	assert.True(t, proto.Equal(&executionIdentifier, execution.Id))
	assert.True(t, proto.Equal(spec, execution.Spec))
	assert.True(t, proto.Equal(&closure, execution.Closure))
}

func TestGetExecution_DatabaseError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	expectedErr := errors.New("expected error")

	executionGetFunc := func(ctx context.Context, input interfaces.GetResourceInput) (models.Execution, error) {
		assert.Equal(t, "project", input.Project)
		assert.Equal(t, "domain", input.Domain)
		assert.Equal(t, "name", input.Name)
		return models.Execution{}, expectedErr
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	execution, err := execManager.GetExecution(context.Background(), admin.WorkflowExecutionGetRequest{
		Id: &executionIdentifier,
	})
	assert.Nil(t, execution)
	assert.Equal(t, expectedErr, err)
}

func TestGetExecution_TransformerError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	startedAt := time.Date(2018, 8, 30, 0, 0, 0, 0, time.UTC)
	executionGetFunc := func(ctx context.Context, input interfaces.GetResourceInput) (models.Execution, error) {
		assert.Equal(t, "project", input.Project)
		assert.Equal(t, "domain", input.Domain)
		assert.Equal(t, "name", input.Name)
		return models.Execution{
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
			Spec:         []byte("invalid spec"),
			Phase:        phase,
			Closure:      closureBytes,
			LaunchPlanID: uint(1),
			WorkflowID:   uint(2),
			StartedAt:    &startedAt,
		}, nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	execution, err := execManager.GetExecution(context.Background(), admin.WorkflowExecutionGetRequest{
		Id: &executionIdentifier,
	})
	assert.Nil(t, execution)
	assert.Equal(t, codes.Internal, err.(flyteAdminErrors.FlyteAdminError).Code())
}

func TestListExecutions(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	executionListFunc := func(
		ctx context.Context, input interfaces.ListResourceInput) (interfaces.ExecutionCollectionOutput, error) {
		var projectFilter, domainFilter, nameFilter bool
		for _, filter := range input.InlineFilters {
			assert.Equal(t, common.Execution, filter.GetEntity())
			queryExpr, _ := filter.GetGormQueryExpr()
			if queryExpr.Args == projectValue && queryExpr.Query == "execution_project = ?" {
				projectFilter = true
			}
			if queryExpr.Args == domainValue && queryExpr.Query == "execution_domain = ?" {
				domainFilter = true
			}
			if queryExpr.Args == nameValue && queryExpr.Query == "execution_name = ?" {
				nameFilter = true
			}
		}
		assert.True(t, projectFilter, "Missing project equality filter")
		assert.True(t, domainFilter, "Missing domain equality filter")
		assert.False(t, nameFilter, "Included name equality filter")
		assert.Equal(t, limit, input.Limit)
		assert.Equal(t, "domain asc", input.SortParameter.GetGormOrderExpr())
		assert.Equal(t, 2, input.Offset)
		return interfaces.ExecutionCollectionOutput{
			Executions: []models.Execution{
				{
					ExecutionKey: models.ExecutionKey{
						Project: projectValue,
						Domain:  domainValue,
						Name:    "my awesome execution",
					},
					Spec:    specBytes,
					Closure: closureBytes,
				},
				{
					ExecutionKey: models.ExecutionKey{
						Project: projectValue,
						Domain:  domainValue,
						Name:    "my other execution",
					},
					Phase:   core.WorkflowExecution_SUCCEEDED.String(),
					Spec:    specBytes,
					Closure: closureBytes,
				},
			},
		}, nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetListCallback(executionListFunc)
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)

	executionList, err := execManager.ListExecutions(context.Background(), admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: projectValue,
			Domain:  domainValue,
		},
		Limit: limit,
		SortBy: &admin.Sort{
			Direction: admin.Sort_ASCENDING,
			Key:       "domain",
		},
		Token: "2",
	})
	assert.NoError(t, err)
	assert.NotNil(t, executionList)
	assert.Len(t, executionList.Executions, 2)

	for idx, execution := range executionList.Executions {
		assert.Equal(t, projectValue, execution.Id.Project)
		assert.Equal(t, domainValue, execution.Id.Domain)
		if idx == 0 {
			assert.Equal(t, "my awesome execution", execution.Id.Name)
		}
		assert.True(t, proto.Equal(spec, execution.Spec))
		assert.True(t, proto.Equal(&closure, execution.Closure))
	}
	assert.Empty(t, executionList.Token)
}

func TestListExecutions_MissingParameters(t *testing.T) {
	execManager := NewExecutionManager(
		repositoryMocks.NewMockRepository(),
		getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	_, err := execManager.ListExecutions(context.Background(), admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Domain: domainValue,
		},
		Limit: limit,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, err.(flyteAdminErrors.FlyteAdminError).Code())

	_, err = execManager.ListExecutions(context.Background(), admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: projectValue,
		},
		Limit: limit,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, err.(flyteAdminErrors.FlyteAdminError).Code())

	_, err = execManager.ListExecutions(context.Background(), admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: projectValue,
			Domain:  domainValue,
		},
	})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, err.(flyteAdminErrors.FlyteAdminError).Code())
}

func TestListExecutions_DatabaseError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	expectedErr := errors.New("expected error")
	executionListFunc := func(
		ctx context.Context, input interfaces.ListResourceInput) (interfaces.ExecutionCollectionOutput, error) {
		return interfaces.ExecutionCollectionOutput{}, expectedErr
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetListCallback(executionListFunc)
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	_, err := execManager.ListExecutions(context.Background(), admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: projectValue,
			Domain:  domainValue,
			Name:    nameValue,
		},
		Limit: limit,
	})
	assert.EqualError(t, err, expectedErr.Error())
}

func TestListExecutions_TransformerError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	executionListFunc := func(
		ctx context.Context, input interfaces.ListResourceInput) (interfaces.ExecutionCollectionOutput, error) {
		return interfaces.ExecutionCollectionOutput{
			Executions: []models.Execution{
				{
					ExecutionKey: models.ExecutionKey{
						Project: projectValue,
						Domain:  domainValue,
						Name:    "my awesome execution",
					},
					Spec:    []byte("I am invalid"),
					Closure: closureBytes,
				},
			},
		}, nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetListCallback(executionListFunc)
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)

	executionList, err := execManager.ListExecutions(context.Background(), admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: projectValue,
			Domain:  domainValue,
		},
		Limit: limit,
	})
	assert.EqualError(t, err, "failed to unmarshal spec")
	assert.Nil(t, executionList)
}

func TestExecutionManager_PublishNotifications(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	queue := executions.NewQueueAllocator(getMockExecutionsConfigProvider())

	mockApplicationConfig := runtimeMocks.MockApplicationProvider{}
	mockApplicationConfig.SetNotificationsConfig(runtimeInterfaces.NotificationsConfig{
		NotificationsEmailerConfig: runtimeInterfaces.NotificationsEmailerConfig{
			Body: "http://example.com/console/projects/%s/domains/%s/executions/%s",
		},
	})
	mockRuntime := runtimeMocks.NewMockConfigurationProvider(
		&mockApplicationConfig,
		runtimeMocks.NewMockQueueConfigurationProvider(
			[]runtimeInterfaces.ExecutionQueue{}, []runtimeInterfaces.WorkflowConfig{}),
		nil, nil, nil, nil)

	var myExecManager = &ExecutionManager{
		db:                 repository,
		config:             mockRuntime,
		storageClient:      getMockStorageForExecTest(context.Background()),
		queueAllocator:     queue,
		_clock:             clock.New(),
		systemMetrics:      newExecutionSystemMetrics(mockScope.NewTestScope()),
		notificationClient: &mockPublisher,
	}
	// Currently this doesn't do anything special as the code to invoke pushing to SNS isn't enabled yet.
	// This sets up the skeleton for it and appeases the go lint overlords.
	workflowRequest := admin.WorkflowExecutionEventRequest{
		Event: &event.WorkflowExecutionEvent{
			Phase: core.WorkflowExecution_FAILED,
			//ExecutionId: "1234",
			OutputResult: &event.WorkflowExecutionEvent_Error{
				Error: &core.ExecutionError{
					Code:    "CodeBad",
					Message: "oopsie my bad",
				},
			},
			ExecutionId: &executionIdentifier,
		},
	}
	var execClosure = admin.ExecutionClosure{
		Notifications: testutils.GetExecutionRequest().Spec.GetNotifications().Notifications,
		WorkflowId: &core.Identifier{
			ResourceType: core.ResourceType_WORKFLOW,
			Project:      "wf_project",
			Domain:       "wf_domain",
			Name:         "wf_name",
			Version:      "wf_version",
		},
	}
	var extraNotifications = []*admin.Notification{
		{
			Phases: []core.WorkflowExecution_Phase{
				core.WorkflowExecution_FAILED,
			},
			Type: &admin.Notification_PagerDuty{
				PagerDuty: &admin.PagerDutyNotification{
					RecipientsEmail: []string{
						"pagerduty@example.com",
					},
				},
			},
		},
		{
			Phases: []core.WorkflowExecution_Phase{
				core.WorkflowExecution_SUCCEEDED,
				core.WorkflowExecution_FAILED,
			},
			Type: &admin.Notification_Email{
				Email: &admin.EmailNotification{
					RecipientsEmail: []string{
						"email@example.com",
					},
				},
			},
		},
	}
	execClosure.Notifications = append(execClosure.Notifications, extraNotifications[0])
	execClosure.Notifications = append(execClosure.Notifications, extraNotifications[1])

	execClosureBytes, _ := proto.Marshal(&execClosure)
	executionModel := models.Execution{
		ExecutionKey: models.ExecutionKey{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Phase:        core.WorkflowExecution_FAILED.String(),
		LaunchPlanID: uint(1),
		WorkflowID:   uint(2),
		Closure:      execClosureBytes,
		Spec:         specBytes,
	}
	assert.Nil(t, myExecManager.publishNotifications(context.Background(), workflowRequest, executionModel))
}

func TestExecutionManager_PublishNotificationsTransformError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	queue := executions.NewQueueAllocator(getMockExecutionsConfigProvider())
	var execManager = &ExecutionManager{
		db:                 repository,
		config:             getMockExecutionsConfigProvider(),
		storageClient:      getMockStorageForExecTest(context.Background()),
		queueAllocator:     queue,
		_clock:             clock.New(),
		systemMetrics:      newExecutionSystemMetrics(mockScope.NewTestScope()),
		notificationClient: &mockPublisher,
	}

	workflowRequest := admin.WorkflowExecutionEventRequest{
		Event: &event.WorkflowExecutionEvent{
			Phase: core.WorkflowExecution_FAILED,
			//ExecutionId: "1234",
			OutputResult: &event.WorkflowExecutionEvent_Error{
				Error: &core.ExecutionError{
					Code:    "CodeBad",
					Message: "oopsie my bad",
				},
			},
			ExecutionId: &executionIdentifier,
		},
	}
	// Ensure that an error is thrown when transforming an incorrect models.Execution
	executionModel := models.Execution{
		ExecutionKey: models.ExecutionKey{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Phase:        core.WorkflowExecution_FAILED.String(),
		LaunchPlanID: uint(1),
		WorkflowID:   uint(2),
		Spec:         []byte("I am invalid"),
	}
	assert.Error(t, execManager.publishNotifications(context.Background(), workflowRequest, executionModel))
}

func TestExecutionManager_TestExecutionManager_PublishNotificationsTransformError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	queue := executions.NewQueueAllocator(getMockExecutionsConfigProvider())
	publishFunc := func(ctx context.Context, key string, msg proto.Message) error {
		return errors.New("error publishing message")
	}

	mockPublisher.SetPublishCallback(publishFunc)
	mockApplicationConfig := runtimeMocks.MockApplicationProvider{}
	mockApplicationConfig.SetNotificationsConfig(runtimeInterfaces.NotificationsConfig{
		NotificationsEmailerConfig: runtimeInterfaces.NotificationsEmailerConfig{
			Body: "http://example.com/console/projects/%s/domains/%s/executions/%s",
		},
	})
	mockRuntime := runtimeMocks.NewMockConfigurationProvider(
		&mockApplicationConfig,
		runtimeMocks.NewMockQueueConfigurationProvider(
			[]runtimeInterfaces.ExecutionQueue{}, []runtimeInterfaces.WorkflowConfig{}),
		nil, nil, nil, nil)

	var myExecManager = &ExecutionManager{
		db:                 repository,
		config:             mockRuntime,
		storageClient:      getMockStorageForExecTest(context.Background()),
		queueAllocator:     queue,
		_clock:             clock.New(),
		systemMetrics:      newExecutionSystemMetrics(mockScope.NewTestScope()),
		notificationClient: &mockPublisher,
	}
	// Currently this doesn't do anything special as the code to invoke pushing to SNS isn't enabled yet.
	// This sets up the skeleton for it and appeases the go lint overlords.
	workflowRequest := admin.WorkflowExecutionEventRequest{
		Event: &event.WorkflowExecutionEvent{
			Phase: core.WorkflowExecution_FAILED,
			//ExecutionId: "1234",
			OutputResult: &event.WorkflowExecutionEvent_Error{
				Error: &core.ExecutionError{
					Code:    "CodeBad",
					Message: "oopsie my bad",
				},
			},
			ExecutionId: &executionIdentifier,
		},
	}
	var execClosure = admin.ExecutionClosure{
		Notifications: testutils.GetExecutionRequest().Spec.GetNotifications().Notifications,
		WorkflowId: &core.Identifier{
			ResourceType: core.ResourceType_WORKFLOW,
			Project:      "wf_project",
			Domain:       "wf_domain",
			Name:         "wf_name",
			Version:      "wf_version",
		},
	}
	execClosureBytes, _ := proto.Marshal(&execClosure)
	executionModel := models.Execution{
		ExecutionKey: models.ExecutionKey{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Phase:        core.WorkflowExecution_FAILED.String(),
		LaunchPlanID: uint(1),
		WorkflowID:   uint(2),
		Closure:      execClosureBytes,
		Spec:         specBytes,
	}
	assert.Nil(t, myExecManager.publishNotifications(context.Background(), workflowRequest, executionModel))

}

func TestExecutionManager_PublishNotificationsNoPhaseMatch(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	queue := executions.NewQueueAllocator(getMockExecutionsConfigProvider())

	var myExecManager = &ExecutionManager{
		db:                 repository,
		config:             getMockExecutionsConfigProvider(),
		storageClient:      getMockStorageForExecTest(context.Background()),
		queueAllocator:     queue,
		_clock:             clock.New(),
		systemMetrics:      newExecutionSystemMetrics(mockScope.NewTestScope()),
		notificationClient: &mockPublisher,
	}
	// Currently this doesn't do anything special as the code to invoke pushing to SNS isn't enabled yet.
	// This sets up the skeleton for it and appeases the go lint overlords.
	workflowRequest := admin.WorkflowExecutionEventRequest{
		Event: &event.WorkflowExecutionEvent{
			Phase: core.WorkflowExecution_SUCCEEDED,
			OutputResult: &event.WorkflowExecutionEvent_OutputUri{
				OutputUri: "somestring",
			},
			ExecutionId: &executionIdentifier,
		},
	}
	var execClosure = admin.ExecutionClosure{
		Notifications: testutils.GetExecutionRequest().Spec.GetNotifications().Notifications,
	}
	execClosureBytes, _ := proto.Marshal(&execClosure)
	executionModel := models.Execution{
		ExecutionKey: models.ExecutionKey{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Phase:        core.WorkflowExecution_FAILED.String(),
		LaunchPlanID: uint(1),
		WorkflowID:   uint(2),
		Closure:      execClosureBytes,
	}
	assert.Nil(t, myExecManager.publishNotifications(context.Background(), workflowRequest, executionModel))
}

func TestTerminateExecution(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	startTime := time.Now()
	executionGetFunc := makeExecutionGetFunc(t, []byte{}, &startTime)
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)

	abortCause := "abort cause"
	updateExecutionFunc := func(
		context context.Context, execution models.Execution) error {
		assert.Equal(t, "project", execution.Project)
		assert.Equal(t, "domain", execution.Domain)
		assert.Equal(t, "name", execution.Name)
		assert.Equal(t, uint(1), execution.LaunchPlanID)
		assert.Equal(t, uint(2), execution.WorkflowID)
		assert.Equal(t, core.WorkflowExecution_QUEUED.String(), execution.Phase,
			"an abort call should not update the execution status until a corresponding execution event "+
				"is received")
		assert.Equal(t, execution.ExecutionCreatedAt, execution.ExecutionUpdatedAt,
			"an abort call should not change ExecutionUpdatedAt until a corresponding execution event is received")
		assert.Equal(t, abortCause, execution.AbortCause)
		assert.Equal(t, testCluster, execution.Cluster)
		return nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetUpdateExecutionCallback(updateExecutionFunc)

	mockExecutor := workflowengineMocks.NewMockExecutor()
	mockExecutor.(*workflowengineMocks.MockExecutor).SetTerminateExecutionCallback(
		func(ctx context.Context, input workflowengineInterfaces.TerminateWorkflowInput) error {
			assert.True(t, proto.Equal(&core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			}, input.ExecutionID))
			return nil
		})
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockExecutor,
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)

	resp, err := execManager.TerminateExecution(context.Background(), admin.ExecutionTerminateRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Cause: abortCause,
	})

	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestTerminateExecution_PropellerError(t *testing.T) {
	var expectedError = errors.New("expected error")

	mockExecutor := workflowengineMocks.NewMockExecutor()
	mockExecutor.(*workflowengineMocks.MockExecutor).SetTerminateExecutionCallback(
		func(ctx context.Context, input workflowengineInterfaces.TerminateWorkflowInput) error {
			return expectedError
		})
	repository := repositoryMocks.NewMockRepository()
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetUpdateExecutionCallback(func(
		context context.Context, execution models.Execution) error {
		t.Fatal("update should not be called when propeller fails to terminate an execution")
		return nil
	})
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockExecutor,
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)

	resp, err := execManager.TerminateExecution(context.Background(), admin.ExecutionTerminateRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Cause: "abort cause",
	})
	assert.Nil(t, resp)
	assert.EqualError(t, err, expectedError.Error())
}

func TestTerminateExecution_DatabaseError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	startTime := time.Now()
	executionGetFunc := makeExecutionGetFunc(t, []byte{}, &startTime)
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)

	var expectedError = errors.New("expected error")
	updateExecutionFunc := func(
		context context.Context, execution models.Execution) error {
		return expectedError
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetUpdateExecutionCallback(updateExecutionFunc)

	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	resp, err := execManager.TerminateExecution(context.Background(), admin.ExecutionTerminateRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Cause: "abort cause",
	})

	assert.Nil(t, resp)
	assert.EqualError(t, err, expectedError.Error())
}

func TestGetExecutionData(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	startedAt := time.Date(2018, 8, 30, 0, 0, 0, 0, time.UTC)
	var closure = admin.ExecutionClosure{
		Phase: core.WorkflowExecution_RUNNING,
		OutputResult: &admin.ExecutionClosure_Outputs{
			Outputs: &admin.LiteralMapBlob{
				Data: &admin.LiteralMapBlob_Uri{
					Uri: outputURI,
				},
			},
		},
	}
	var closureBytes, _ = proto.Marshal(&closure)

	executionGetFunc := func(ctx context.Context, input interfaces.GetResourceInput) (models.Execution, error) {
		return models.Execution{
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
			Spec:         specBytes,
			Phase:        phase,
			Closure:      closureBytes,
			LaunchPlanID: uint(1),
			WorkflowID:   uint(2),
			StartedAt:    &startedAt,
			InputsURI:    shared.Inputs,
		}, nil
	}
	mockExecutionRemoteURL := dataMocks.NewMockRemoteURL()
	mockExecutionRemoteURL.(*dataMocks.MockRemoteURL).GetCallback = func(
		ctx context.Context, uri string) (admin.UrlBlob, error) {
		if uri == outputURI {
			return admin.UrlBlob{
				Url:   "outputs",
				Bytes: 200,
			}, nil
		} else if strings.HasSuffix(uri, shared.Inputs) {
			return admin.UrlBlob{
				Url:   "inputs",
				Bytes: 200,
			}, nil
		}

		return admin.UrlBlob{}, errors.New("unexpected input")
	}

	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	dataResponse, err := execManager.GetExecutionData(context.Background(), admin.WorkflowExecutionGetDataRequest{
		Id: &executionIdentifier,
	})
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&admin.WorkflowExecutionGetDataResponse{
		Outputs: &admin.UrlBlob{
			Url:   "outputs",
			Bytes: 200,
		},
		Inputs: &admin.UrlBlob{
			Url:   "inputs",
			Bytes: 200,
		},
	}, dataResponse))
}

func TestAddLabelsAndAnnotationsRuntimeLimitsObserved(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	setDefaultLpCallbackForExecTest(repository)

	mockRegistrationValidationConfig := runtimeMocks.NewMockRegistrationValidationProvider()
	mockRegistrationValidationConfig.(*runtimeMocks.MockRegistrationValidationProvider).MaxAnnotationEntries = 1

	configProvider := getMockExecutionsConfigProvider()
	configProvider.(*runtimeMocks.MockConfigurationProvider).AddRegistrationValidationConfiguration(
		mockRegistrationValidationConfig)
	execManager := NewExecutionManager(
		repository, configProvider, getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	request := testutils.GetExecutionRequest()
	request.Spec.Labels = &admin.Labels{
		Values: map[string]string{
			"dynamiclabel1": "dynamic1",
			"dynamiclabel2": "dynamic2",
		},
	}
	request.Spec.Annotations = &admin.Annotations{
		Values: map[string]string{
			"dynamicannotation3": "dynamic3",
			"dynamicannotation4": "dynamic4",
		},
	}
	launchPlanSpec := testutils.GetSampleLpSpecForTest()
	err := execManager.(*ExecutionManager).addLabelsAndAnnotations(request.Spec, &workflowengineInterfaces.ExecuteWorkflowInput{
		Reference: admin.LaunchPlan{
			Spec: &launchPlanSpec,
		},
	})
	assert.EqualError(t, err, "Annotations has too many entries [2 > 1]")

	mockRegistrationValidationConfig.(*runtimeMocks.MockRegistrationValidationProvider).MaxAnnotationEntries = 0
	mockRegistrationValidationConfig.(*runtimeMocks.MockRegistrationValidationProvider).MaxLabelEntries = 1
	err = execManager.(*ExecutionManager).addLabelsAndAnnotations(request.Spec, &workflowengineInterfaces.ExecuteWorkflowInput{
		Reference: admin.LaunchPlan{
			Spec: &launchPlanSpec,
		},
	})
	assert.EqualError(t, err, "Labels has too many entries [2 > 1]")
}

func TestGetExecution_Legacy(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	startedAt := time.Date(2018, 8, 30, 0, 0, 0, 0, time.UTC)
	executionGetFunc := func(ctx context.Context, input interfaces.GetResourceInput) (models.Execution, error) {
		assert.Equal(t, "project", input.Project)
		assert.Equal(t, "domain", input.Domain)
		assert.Equal(t, "name", input.Name)
		return models.Execution{
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
			Spec:         getLegacySpecBytes(),
			Phase:        phase,
			Closure:      getLegacyClosureBytes(),
			LaunchPlanID: uint(1),
			WorkflowID:   uint(2),
			StartedAt:    &startedAt,
		}, nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	execution, err := execManager.GetExecution(context.Background(), admin.WorkflowExecutionGetRequest{
		Id: &executionIdentifier,
	})
	assert.NoError(t, err)
	assert.True(t, proto.Equal(&executionIdentifier, execution.Id))
	assert.True(t, proto.Equal(getLegacySpec(), execution.Spec))
	assert.True(t, proto.Equal(getLegacyClosure(), execution.Closure))
}

func TestGetExecution_LegacyClient_OffloadedData(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	startedAt := time.Date(2018, 8, 30, 0, 0, 0, 0, time.UTC)
	executionGetFunc := func(ctx context.Context, input interfaces.GetResourceInput) (models.Execution, error) {
		assert.Equal(t, "project", input.Project)
		assert.Equal(t, "domain", input.Domain)
		assert.Equal(t, "name", input.Name)
		return models.Execution{
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
			Spec:          specBytes,
			Phase:         phase,
			Closure:       closureBytes,
			LaunchPlanID:  uint(1),
			WorkflowID:    uint(2),
			StartedAt:     &startedAt,
			UserInputsURI: shared.UserInputs,
			InputsURI:     shared.Inputs,
		}, nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)
	storageClient := getMockStorageForExecTest(context.Background())
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), storageClient, workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	_ = storageClient.WriteProtobuf(context.Background(), storage.DataReference(shared.UserInputs), storage.Options{}, getLegacySpec().Inputs)
	_ = storageClient.WriteProtobuf(context.Background(), storage.DataReference(shared.Inputs), storage.Options{}, getLegacyClosure().ComputedInputs)
	execution, err := execManager.GetExecution(context.Background(), admin.WorkflowExecutionGetRequest{
		Id: &executionIdentifier,
	})
	assert.NoError(t, err)
	assert.True(t, proto.Equal(&executionIdentifier, execution.Id))
	assert.True(t, proto.Equal(getLegacySpec(), execution.Spec))
	assert.True(t, proto.Equal(getLegacyClosure(), execution.Closure))
}

func TestGetExecutionData_LegacyModel(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	startedAt := time.Date(2018, 8, 30, 0, 0, 0, 0, time.UTC)
	closure := getLegacyClosure()
	closure.OutputResult = &admin.ExecutionClosure_Outputs{
		Outputs: &admin.LiteralMapBlob{
			Data: &admin.LiteralMapBlob_Uri{
				Uri: outputURI,
			},
		},
	}
	var closureBytes, _ = proto.Marshal(closure)

	executionGetFunc := func(ctx context.Context, input interfaces.GetResourceInput) (models.Execution, error) {
		return models.Execution{
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
			Spec:         getLegacySpecBytes(),
			Phase:        phase,
			Closure:      closureBytes,
			LaunchPlanID: uint(1),
			WorkflowID:   uint(2),
			StartedAt:    &startedAt,
		}, nil
	}
	mockExecutionRemoteURL := dataMocks.NewMockRemoteURL()
	mockExecutionRemoteURL.(*dataMocks.MockRemoteURL).GetCallback = func(
		ctx context.Context, uri string) (admin.UrlBlob, error) {
		if uri == outputURI {
			return admin.UrlBlob{
				Url:   "outputs",
				Bytes: 200,
			}, nil
		} else if strings.HasSuffix(uri, shared.Inputs) {
			return admin.UrlBlob{
				Url:   "inputs",
				Bytes: 200,
			}, nil
		}

		return admin.UrlBlob{}, errors.New("unexpected input")
	}

	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)
	storageClient := getMockStorageForExecTest(context.Background())
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), storageClient, workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	dataResponse, err := execManager.GetExecutionData(context.Background(), admin.WorkflowExecutionGetDataRequest{
		Id: &executionIdentifier,
	})
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&admin.WorkflowExecutionGetDataResponse{
		Outputs: &admin.UrlBlob{
			Url:   "outputs",
			Bytes: 200,
		},
		Inputs: &admin.UrlBlob{
			Url:   "inputs",
			Bytes: 200,
		},
	}, dataResponse))
	var inputs core.LiteralMap
	err = storageClient.ReadProtobuf(context.Background(), storage.DataReference("s3://bucket/metadata/project/domain/name/inputs"), &inputs)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&inputs, closure.ComputedInputs))
}

func TestCreateExecution_LegacyClient(t *testing.T) {
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	mockExecutor := workflowengineMocks.NewMockExecutor()
	mockExecutor.(*workflowengineMocks.MockExecutor).SetExecuteWorkflowCallback(
		func(inputs workflowengineInterfaces.ExecuteWorkflowInput) (*workflowengineInterfaces.ExecutionInfo, error) {
			assert.EqualValues(t, map[string]string{
				"label1": "1",
				"label2": "2",
			}, inputs.Labels)
			assert.EqualValues(t, map[string]string{
				"annotation3": "3",
				"annotation4": "4",
			}, inputs.Annotations)
			return &workflowengineInterfaces.ExecutionInfo{
				Cluster: testCluster,
			}, nil
		})
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockExecutor,
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	response, err := execManager.CreateExecution(context.Background(), *getLegacyExecutionRequest(), requestedAt)
	assert.Nil(t, err)

	expectedResponse := &admin.ExecutionCreateResponse{
		Id: &executionIdentifier,
	}
	assert.Nil(t, err)
	assert.Equal(t, expectedResponse, response)
}

func TestRelaunchExecution_LegacyModel(t *testing.T) {
	// Set up mocks.
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	storageClient := getMockStorageForExecTest(context.Background())
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), storageClient, workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)
	startTime := time.Now()
	startTimeProto, _ := ptypes.TimestampProto(startTime)
	existingClosure := getLegacyClosure()
	existingClosure.Phase = core.WorkflowExecution_RUNNING
	existingClosure.StartedAt = startTimeProto
	existingClosure.ComputedInputs.Literals["bar"] = utils.MustMakeLiteral("bar-value")
	existingClosureBytes, _ := proto.Marshal(existingClosure)
	executionGetFunc := makeLegacyExecutionGetFunc(t, existingClosureBytes, &startTime)
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)

	var createCalled bool
	exCreateFunc := func(ctx context.Context, input models.Execution) error {
		createCalled = true
		assert.Equal(t, "relaunchy", input.Name)
		assert.Equal(t, "domain", input.Domain)
		assert.Equal(t, "project", input.Project)
		assert.Equal(t, uint(8), input.SourceExecutionID)
		var spec admin.ExecutionSpec
		err := proto.Unmarshal(input.Spec, &spec)
		assert.Nil(t, err)
		assert.Equal(t, admin.ExecutionMetadata_RELAUNCH, spec.Metadata.Mode)
		assert.Equal(t, int32(admin.ExecutionMetadata_RELAUNCH), input.Mode)
		assert.True(t, proto.Equal(spec.Inputs, getLegacySpec().Inputs))
		return nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(exCreateFunc)

	// Issue request.
	response, err := execManager.RelaunchExecution(context.Background(), admin.ExecutionRelaunchRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Name: "relaunchy",
	}, requestedAt)

	// And verify response.
	assert.Nil(t, err)

	expectedResponse := &admin.ExecutionCreateResponse{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "relaunchy",
		},
	}
	assert.True(t, createCalled)
	assert.True(t, proto.Equal(expectedResponse, response))

	var userInputs core.LiteralMap
	err = storageClient.ReadProtobuf(context.Background(), "s3://bucket/metadata/project/domain/relaunchy/user_inputs", &userInputs)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&userInputs, getLegacySpec().Inputs))

	var inputs core.LiteralMap
	err = storageClient.ReadProtobuf(context.Background(), "s3://bucket/metadata/project/domain/relaunchy/inputs", &inputs)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&inputs, existingClosure.ComputedInputs))
}

func TestListExecutions_LegacyModel(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	executionListFunc := func(
		ctx context.Context, input interfaces.ListResourceInput) (interfaces.ExecutionCollectionOutput, error) {
		var projectFilter, domainFilter, nameFilter bool
		for _, filter := range input.InlineFilters {
			assert.Equal(t, common.Execution, filter.GetEntity())
			queryExpr, _ := filter.GetGormQueryExpr()
			if queryExpr.Args == projectValue && queryExpr.Query == "execution_project = ?" {
				projectFilter = true
			}
			if queryExpr.Args == domainValue && queryExpr.Query == "execution_domain = ?" {
				domainFilter = true
			}
			if queryExpr.Args == nameValue && queryExpr.Query == "execution_name = ?" {
				nameFilter = true
			}
		}
		assert.True(t, projectFilter, "Missing project equality filter")
		assert.True(t, domainFilter, "Missing domain equality filter")
		assert.False(t, nameFilter, "Included name equality filter")
		assert.Equal(t, limit, input.Limit)
		assert.Equal(t, "domain asc", input.SortParameter.GetGormOrderExpr())
		assert.Equal(t, 2, input.Offset)
		return interfaces.ExecutionCollectionOutput{
			Executions: []models.Execution{
				{
					ExecutionKey: models.ExecutionKey{
						Project: projectValue,
						Domain:  domainValue,
						Name:    "my awesome execution",
					},
					Spec:    getLegacySpecBytes(),
					Closure: getLegacyClosureBytes(),
				},
				{
					ExecutionKey: models.ExecutionKey{
						Project: projectValue,
						Domain:  domainValue,
						Name:    "my other execution",
					},
					Phase:   core.WorkflowExecution_SUCCEEDED.String(),
					Spec:    specBytes,
					Closure: closureBytes,
				},
			},
		}, nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetListCallback(executionListFunc)
	execManager := NewExecutionManager(
		repository, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), workflowengineMocks.NewMockExecutor(),
		mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL)

	executionList, err := execManager.ListExecutions(context.Background(), admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: projectValue,
			Domain:  domainValue,
		},
		Limit: limit,
		SortBy: &admin.Sort{
			Direction: admin.Sort_ASCENDING,
			Key:       "domain",
		},
		Token: "2",
	})
	assert.NoError(t, err)
	assert.NotNil(t, executionList)
	assert.Len(t, executionList.Executions, 2)

	for idx, execution := range executionList.Executions {
		assert.Equal(t, projectValue, execution.Id.Project)
		assert.Equal(t, domainValue, execution.Id.Domain)
		if idx == 0 {
			assert.Equal(t, "my awesome execution", execution.Id.Name)
		}
		assert.True(t, proto.Equal(spec, execution.Spec))
		assert.True(t, proto.Equal(&closure, execution.Closure))
	}
	assert.Empty(t, executionList.Token)
}
