package impl

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/flyteorg/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	mockScope "github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyteadmin/auth"
	eventWriterMocks "github.com/flyteorg/flyteadmin/pkg/async/events/mocks"
	notificationMocks "github.com/flyteorg/flyteadmin/pkg/async/notifications/mocks"
	"github.com/flyteorg/flyteadmin/pkg/common"
	commonMocks "github.com/flyteorg/flyteadmin/pkg/common/mocks"
	commonTestUtils "github.com/flyteorg/flyteadmin/pkg/common/testutils"
	dataMocks "github.com/flyteorg/flyteadmin/pkg/data/mocks"
	flyteAdminErrors "github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/executions"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"
	managerInterfaces "github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	managerMocks "github.com/flyteorg/flyteadmin/pkg/manager/mocks"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	repositoryMocks "github.com/flyteorg/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"
	"github.com/flyteorg/flyteadmin/pkg/runtime"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	runtimeIFaceMocks "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces/mocks"
	runtimeMocks "github.com/flyteorg/flyteadmin/pkg/runtime/mocks"
	workflowengineInterfaces "github.com/flyteorg/flyteadmin/pkg/workflowengine/interfaces"
	workflowengineMocks "github.com/flyteorg/flyteadmin/pkg/workflowengine/mocks"
	"github.com/flyteorg/flyteadmin/plugins"
)

var spec = testutils.GetExecutionRequest().Spec
var specBytes, _ = proto.Marshal(spec)
var phase = core.WorkflowExecution_RUNNING.String()
var closure = admin.ExecutionClosure{
	Phase: core.WorkflowExecution_RUNNING,
	StateChangeDetails: &admin.ExecutionStateChangeDetails{
		State:      admin.ExecutionState_EXECUTION_ACTIVE,
		OccurredAt: testutils.MockCreatedAtProto,
	},
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

var resourceDefaults = runtimeInterfaces.TaskResourceSet{
	CPU:    resource.MustParse("200m"),
	Memory: resource.MustParse("200Gi"),
}
var resourceLimits = runtimeInterfaces.TaskResourceSet{
	CPU:    resource.MustParse("300m"),
	Memory: resource.MustParse("500Gi"),
}

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

func getExpectedLegacySpec() *admin.ExecutionSpec {
	expectedLegacySpec := getLegacySpec()
	expectedLegacySpec.Metadata = &admin.ExecutionMetadata{
		SystemMetadata: &admin.SystemMetadata{
			Namespace: "project-domain",
		},
	}
	return expectedLegacySpec
}

func getExpectedLegacySpecBytes() []byte {
	expectedLegacySpec := getExpectedLegacySpec()
	b, _ := proto.Marshal(expectedLegacySpec)
	return b
}

func getExpectedSpec() *admin.ExecutionSpec {
	expectedSpec := testutils.GetExecutionRequest().Spec
	expectedSpec.Metadata = &admin.ExecutionMetadata{
		SystemMetadata: &admin.SystemMetadata{
			Namespace: "project-domain",
		},
	}
	return expectedSpec
}

func getExpectedSpecBytes() []byte {
	specBytes, _ := proto.Marshal(getExpectedSpec())
	return specBytes
}

func getLegacyClosure() *admin.ExecutionClosure {
	return &admin.ExecutionClosure{
		Phase:          core.WorkflowExecution_RUNNING,
		ComputedInputs: getLegacySpec().Inputs,
		StateChangeDetails: &admin.ExecutionStateChangeDetails{
			State:      admin.ExecutionState_EXECUTION_ACTIVE,
			OccurredAt: testutils.MockCreatedAtProto,
		},
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

func getMockNamespaceMappingConfig() runtimeInterfaces.NamespaceMappingConfiguration {
	mockNs := runtimeMocks.NamespaceMappingConfiguration{}
	mockNs.OnGetNamespaceTemplate().Return("{{ project }}-{{ domain }}")
	return &mockNs
}

func getMockExecutionsConfigProvider() runtimeInterfaces.Configuration {
	mockExecutionsConfigProvider := runtimeMocks.NewMockConfigurationProvider(
		testutils.GetApplicationConfigWithDefaultDomains(),
		runtimeMocks.NewMockQueueConfigurationProvider(
			[]runtimeInterfaces.ExecutionQueue{}, []runtimeInterfaces.WorkflowConfig{}),
		nil,
		runtimeMocks.NewMockTaskResourceConfiguration(resourceDefaults, resourceLimits), nil, getMockNamespaceMappingConfig())
	mockExecutionsConfigProvider.(*runtimeMocks.MockConfigurationProvider).AddRegistrationValidationConfiguration(
		runtimeMocks.NewMockRegistrationValidationProvider())
	return mockExecutionsConfigProvider
}

func setDefaultLpCallbackForExecTest(repository interfaces.Repository) {
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

	lpGetFunc := func(input interfaces.Identifier) (models.LaunchPlan, error) {
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

func setDefaultTaskCallbackForExecTest(repository interfaces.Repository) {
	taskGetFunc := func(input interfaces.Identifier) (models.Task, error) {
		return models.Task{
			TaskKey: models.TaskKey{
				Project: input.Project,
				Domain:  input.Domain,
				Name:    input.Name,
				Version: input.Version,
			},
			BaseModel: models.BaseModel{
				ID:        uint(123),
				CreatedAt: testutils.MockCreatedAtValue,
			},
			Closure: testutils.GetTaskClosureBytes(),
			Digest:  []byte(input.Name),
			Type:    "python",
		}, nil
	}
	repository.TaskRepo().(*repositoryMocks.MockTaskRepo).SetGetCallback(taskGetFunc)
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
	if err := mockStorage.WriteProtobuf(ctx, remoteClosureIdentifier, defaultStorageOptions, workflowClosure); err != nil {
		return nil
	}
	return mockStorage
}

func getMockRepositoryForExecTest() interfaces.Repository {
	repository := repositoryMocks.NewMockRepository()
	repository.WorkflowRepo().(*repositoryMocks.MockWorkflowRepo).SetGetCallback(
		func(input interfaces.Identifier) (models.Workflow, error) {
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

var defaultTestExecutor = workflowengineMocks.WorkflowExecutor{}

func init() {
	defaultTestExecutor.OnID().Return("testDefault")
}

func TestCreateExecution(t *testing.T) {
	repository := getMockRepositoryForExecTest()
	labels := admin.Labels{
		Values: map[string]string{
			"label3": "3",
			"label2": "1", // common label, will be dropped
		}}
	repository.ProjectRepo().(*repositoryMocks.MockProjectRepo).GetFunction = func(
		ctx context.Context, projectID string) (models.Project, error) {
		return transformers.CreateProjectModel(&admin.Project{
			Labels: &labels}), nil
	}

	principal := "principal"
	rawOutput := "raw_output"
	clusterAssignment := admin.ClusterAssignment{ClusterPoolName: "gpu"}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(
		func(ctx context.Context, input models.Execution) error {
			var spec admin.ExecutionSpec
			err := proto.Unmarshal(input.Spec, &spec)
			assert.NoError(t, err)
			assert.Equal(t, principal, spec.Metadata.Principal)
			assert.Equal(t, rawOutput, spec.RawOutputDataConfig.OutputLocationPrefix)
			assert.True(t, proto.Equal(spec.ClusterAssignment, &clusterAssignment))
			assert.Equal(t, "launch_plan", input.LaunchEntity)
			assert.Equal(t, spec.GetMetadata().GetSystemMetadata().Namespace, "project-domain")
			return nil
		})
	setDefaultLpCallbackForExecTest(repository)
	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	resources := &core.Resources{
		Requests: []*core.Resources_ResourceEntry{
			{
				Name:  core.Resources_CPU,
				Value: "200m",
			},
			{
				Name:  core.Resources_MEMORY,
				Value: "200Gi",
			},
		},
		Limits: []*core.Resources_ResourceEntry{
			{
				Name:  core.Resources_CPU,
				Value: "300m",
			},
			{
				Name:  core.Resources_MEMORY,
				Value: "500Gi",
			},
		},
	}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.MatchedBy(func(data workflowengineInterfaces.ExecutionData) bool {
		tasks := data.WorkflowClosure.GetTasks()
		for _, task := range tasks {
			assert.EqualValues(t, resources.Requests,
				task.Template.GetContainer().Resources.Requests)
			assert.EqualValues(t, resources.Requests,
				task.Template.GetContainer().Resources.Limits)
		}

		return true
	})).Return(workflowengineInterfaces.ExecutionResponse{
		Cluster: testCluster,
	}, nil)
	mockExecutor.OnID().Return("customMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)

	qosProvider := &runtimeIFaceMocks.QualityOfServiceConfiguration{}
	qosProvider.OnGetTierExecutionValues().Return(map[core.QualityOfService_Tier]core.QualityOfServiceSpec{
		core.QualityOfService_HIGH: {
			QueueingBudget: ptypes.DurationProto(10 * time.Minute),
		},
		core.QualityOfService_MEDIUM: {
			QueueingBudget: ptypes.DurationProto(20 * time.Minute),
		},
		core.QualityOfService_LOW: {
			QueueingBudget: ptypes.DurationProto(30 * time.Minute),
		},
	})

	qosProvider.OnGetDefaultTiers().Return(map[string]core.QualityOfService_Tier{
		"domain": core.QualityOfService_HIGH,
	})

	mockConfig := getMockExecutionsConfigProvider()
	mockConfig.(*runtimeMocks.MockConfigurationProvider).AddQualityOfServiceConfiguration(qosProvider)

	execManager := NewExecutionManager(repository, r, mockConfig, getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, &mockPublisher, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
	request := testutils.GetExecutionRequest()
	request.Spec.Metadata = &admin.ExecutionMetadata{
		Principal: "unused - populated from authenticated context",
	}
	request.Spec.RawOutputDataConfig = &admin.RawOutputDataConfig{OutputLocationPrefix: rawOutput}
	request.Spec.ClusterAssignment = &clusterAssignment

	identity, err := auth.NewIdentityContext("", principal, "", time.Now(), sets.NewString(), nil, nil)
	assert.NoError(t, err)
	ctx := identity.WithContext(context.Background())
	response, err := execManager.CreateExecution(ctx, request, requestedAt)
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
		func(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
			assert.EqualValues(t, input.NodeExecutionIdentifier, parentNodeExecutionID)
			getNodeExecutionCalled = true
			return models.NodeExecution{
				BaseModel: models.BaseModel{
					ID: 1,
				},
			}, nil
		},
	)

	principal := "feeny"
	getExecutionCalled := false
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
			assert.EqualValues(t, input.Project, parentNodeExecutionID.ExecutionId.Project)
			assert.EqualValues(t, input.Domain, parentNodeExecutionID.ExecutionId.Domain)
			assert.EqualValues(t, input.Name, parentNodeExecutionID.ExecutionId.Name)
			spec := &admin.ExecutionSpec{
				Metadata: &admin.ExecutionMetadata{
					Nesting: 1,
				},
			}
			specBytes, _ := proto.Marshal(spec)
			getExecutionCalled = true
			return models.Execution{
				BaseModel: models.BaseModel{
					ID: 2,
				},
				Spec: specBytes,
				User: principal,
			}, nil
		},
	)

	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(
		func(ctx context.Context, input models.Execution) error {
			assert.Equal(t, input.ParentNodeExecutionID, uint(1))
			var spec admin.ExecutionSpec
			err := proto.Unmarshal(input.Spec, &spec)
			assert.NoError(t, err)
			assert.Equal(t, admin.ExecutionMetadata_CHILD_WORKFLOW, spec.Metadata.Mode)
			assert.True(t, proto.Equal(&parentNodeExecutionID, spec.Metadata.ParentNodeExecution))
			assert.EqualValues(t, input.ParentNodeExecutionID, 1)
			assert.EqualValues(t, input.SourceExecutionID, 2)
			assert.Equal(t, 2, int(spec.Metadata.Nesting))
			assert.Equal(t, principal, spec.Metadata.Principal)
			assert.Equal(t, principal, input.User)
			return nil
		},
	)

	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{
		Cluster: testCluster,
	}, nil)
	mockExecutor.OnID().Return("customMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)

	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
	request := testutils.GetExecutionRequest()
	request.Spec.Metadata = &admin.ExecutionMetadata{
		Mode:                admin.ExecutionMetadata_CHILD_WORKFLOW,
		ParentNodeExecution: &parentNodeExecutionID,
	}
	response, err := execManager.CreateExecution(context.Background(), request, requestedAt)
	assert.Nil(t, err)
	assert.True(t, getNodeExecutionCalled)
	assert.True(t, getExecutionCalled)
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
	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.MatchedBy(func(data workflowengineInterfaces.ExecutionData) bool {
		return len(data.ExecutionID.Name) > 0
	})).Return(workflowengineInterfaces.ExecutionResponse{
		Cluster: testCluster,
	}, nil)
	mockExecutor.OnID().Return("customMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)

	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
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
		testutils.GetApplicationConfigWithDefaultDomains(),
		runtimeMocks.NewMockQueueConfigurationProvider([]runtimeInterfaces.ExecutionQueue{
			{
				Dynamic:    "dynamic Q",
				Attributes: []string{"tag"},
			},
		}, []runtimeInterfaces.WorkflowConfig{
			{
				Domain: "domain",
				Tags:   []string{"tag"},
			},
		}),
		nil,
		runtimeMocks.NewMockTaskResourceConfiguration(resourceDefaults, resourceLimits), nil, getMockNamespaceMappingConfig())
	configProvider.(*runtimeMocks.MockConfigurationProvider).AddRegistrationValidationConfiguration(
		runtimeMocks.NewMockRegistrationValidationProvider())

	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.MatchedBy(func(data workflowengineInterfaces.ExecutionData) bool {
		assert.NotEmpty(t, data.WorkflowClosure.Tasks)
		for _, task := range data.WorkflowClosure.Tasks {
			assert.Len(t, task.Template.GetContainer().Config, 1)
			assert.Contains(t, childContainerQueueKey, task.Template.GetContainer().Config[0].Key)
			assert.Contains(t, "dynamic Q", task.Template.GetContainer().Config[0].Value)
		}
		return true
	})).Return(workflowengineInterfaces.ExecutionResponse{
		Cluster: testCluster,
	}, nil)
	mockExecutor.OnID().Return("customMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	execManager := NewExecutionManager(repository, r, configProvider, getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})

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
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})

	request := testutils.GetExecutionRequest()
	request.Domain = ""
	response, err := execManager.CreateExecution(context.Background(), request, requestedAt)
	assert.EqualError(t, err, "missing domain")
	assert.Nil(t, response)
}

func TestCreateExecution_InvalidLpIdentifier(t *testing.T) {
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})

	request := testutils.GetExecutionRequest()
	request.Spec.LaunchPlan = nil
	response, err := execManager.CreateExecution(context.Background(), request, requestedAt)
	assert.EqualError(t, err, "missing id")
	assert.Nil(t, response)
}

func TestCreateExecutionInCompatibleInputs(t *testing.T) {
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})

	request := testutils.GetExecutionRequest()
	request.Inputs = &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"foo-1": coreutils.MustMakeLiteral("foo-value-1"),
		},
	}
	response, err := execManager.CreateExecution(context.Background(), request, requestedAt)
	assert.EqualError(t, err, "invalid input foo-1")
	assert.Nil(t, response)
}

func TestCreateExecutionPropellerFailure(t *testing.T) {
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	expectedErr := flyteAdminErrors.NewFlyteAdminErrorf(codes.Internal, "ABC")
	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{}, expectedErr)
	mockExecutor.OnID().Return("customMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})

	request := testutils.GetExecutionRequest()

	response, err := execManager.CreateExecution(context.Background(), request, requestedAt)
	assert.EqualError(t, err, expectedErr.Error())
	assert.Nil(t, response)
}

func TestCreateExecutionDatabaseFailure(t *testing.T) {
	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{}, nil)
	mockExecutor.OnID().Return("customMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	expectedErr := flyteAdminErrors.NewFlyteAdminErrorf(codes.Internal, "ABCD")
	exCreateFunc := func(ctx context.Context, input models.Execution) error {
		return expectedErr
	}

	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(exCreateFunc)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
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
		fooValue := coreutils.MustMakeLiteral("foo-value-1")
		assert.Equal(t, 1, len(userInputs.Literals))
		assert.EqualValues(t, userInputs.Literals["foo"], fooValue)
		barValue := coreutils.MustMakeLiteral("bar-value")
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
	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{}, nil)
	mockExecutor.OnID().Return("testMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), storageClient, mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})

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
	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{}, nil)
	mockExecutor.OnID().Return("testMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})

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
	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{}, nil)
	mockExecutor.OnID().Return("testMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})

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
	lpGetFunc := func(input interfaces.Identifier) (models.LaunchPlan, error) {
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
	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{}, nil)
	mockExecutor.OnID().Return("testMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})

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
	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.MatchedBy(func(executionData workflowengineInterfaces.ExecutionData) bool {
		assert.EqualValues(t, map[string]string{
			"dynamiclabel1": "dynamic1",
			"dynamiclabel2": "dynamic2",
		}, executionData.ExecutionParameters.Labels)
		assert.EqualValues(t, map[string]string{
			"dynamicannotation3": "dynamic3",
			"dynamicannotation4": "dynamic4",
		}, executionData.ExecutionParameters.Annotations)
		return true
	})).Return(workflowengineInterfaces.ExecutionResponse{}, nil)
	mockExecutor.OnID().Return("customMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
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

func TestCreateExecutionInterruptible(t *testing.T) {
	enable := true
	disable := false
	tests := []struct {
		name          string
		task          bool
		interruptible *bool
		want          bool
	}{
		{
			name:          "LaunchPlanDefault",
			task:          false,
			interruptible: nil,
			want:          false,
		},
		{
			name:          "LaunchPlanDisable",
			task:          false,
			interruptible: &disable,
			want:          false,
		},
		{
			name:          "LaunchPlanEnable",
			task:          false,
			interruptible: &enable,
			want:          true,
		},
		{
			name:          "TaskDefault",
			task:          true,
			interruptible: nil,
			want:          false,
		},
		{
			name:          "TaskDisable",
			task:          true,
			interruptible: &disable,
			want:          false,
		},
		{
			name:          "TaskEnable",
			task:          true,
			interruptible: &enable,
			want:          true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request := testutils.GetExecutionRequest()
			if tt.task {
				request.Spec.LaunchPlan.ResourceType = core.ResourceType_TASK
			}
			if tt.interruptible == nil {
				request.Spec.Interruptible = nil
			} else {
				request.Spec.Interruptible = &wrappers.BoolValue{Value: *tt.interruptible}
			}

			repository := getMockRepositoryForExecTest()
			setDefaultLpCallbackForExecTest(repository)
			setDefaultTaskCallbackForExecTest(repository)

			exCreateFunc := func(ctx context.Context, input models.Execution) error {
				var spec admin.ExecutionSpec
				err := proto.Unmarshal(input.Spec, &spec)
				assert.Nil(t, err)

				if tt.task {
					assert.Equal(t, uint(0), input.LaunchPlanID)
					assert.NotEqual(t, uint(0), input.TaskID)
				} else {
					assert.NotEqual(t, uint(0), input.LaunchPlanID)
					assert.Equal(t, uint(0), input.TaskID)
				}

				if tt.interruptible == nil {
					assert.Nil(t, spec.GetInterruptible())
				} else {
					assert.NotNil(t, spec.GetInterruptible())
					assert.Equal(t, *tt.interruptible, spec.GetInterruptible().GetValue())
				}

				return nil
			}

			repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(exCreateFunc)
			mockExecutor := workflowengineMocks.WorkflowExecutor{}
			mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{}, nil)
			mockExecutor.OnID().Return("testMockExecutor")
			r := plugins.NewRegistry()
			r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
			execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})

			_, err := execManager.CreateExecution(context.Background(), request, requestedAt)
			assert.Nil(t, err)
		})
	}
}

func TestCreateExecutionOverwriteCache(t *testing.T) {
	tests := []struct {
		name           string
		task           bool
		overwriteCache bool
		want           bool
	}{
		{
			name:           "LaunchPlanDefault",
			task:           false,
			overwriteCache: false,
			want:           false,
		},
		{
			name:           "LaunchPlanEnable",
			task:           false,
			overwriteCache: true,
			want:           true,
		},
		{
			name:           "TaskDefault",
			task:           false,
			overwriteCache: false,
			want:           false,
		},
		{
			name:           "TaskEnable",
			task:           true,
			overwriteCache: true,
			want:           true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request := testutils.GetExecutionRequest()
			if tt.task {
				request.Spec.LaunchPlan.ResourceType = core.ResourceType_TASK
			}
			request.Spec.OverwriteCache = tt.overwriteCache

			repository := getMockRepositoryForExecTest()
			setDefaultLpCallbackForExecTest(repository)
			setDefaultTaskCallbackForExecTest(repository)

			exCreateFunc := func(ctx context.Context, input models.Execution) error {
				var spec admin.ExecutionSpec
				err := proto.Unmarshal(input.Spec, &spec)
				assert.Nil(t, err)

				if tt.task {
					assert.Equal(t, uint(0), input.LaunchPlanID)
					assert.NotEqual(t, uint(0), input.TaskID)
				} else {
					assert.NotEqual(t, uint(0), input.LaunchPlanID)
					assert.Equal(t, uint(0), input.TaskID)
				}

				assert.Equal(t, tt.overwriteCache, spec.GetOverwriteCache())

				return nil
			}

			repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(exCreateFunc)
			mockExecutor := workflowengineMocks.WorkflowExecutor{}
			mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{}, nil)
			mockExecutor.OnID().Return("testMockExecutor")
			r := plugins.NewRegistry()
			r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
			execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})

			_, err := execManager.CreateExecution(context.Background(), request, requestedAt)
			assert.Nil(t, err)
		})
	}
}

func TestCreateExecutionWithEnvs(t *testing.T) {
	tests := []struct {
		name string
		task bool
		envs []*core.KeyValuePair
		want []*core.KeyValuePair
	}{
		{
			name: "LaunchPlanDefault",
			task: false,
			envs: nil,
			want: nil,
		},
		{
			name: "LaunchPlanEnable",
			task: false,
			envs: []*core.KeyValuePair{{Key: "foo", Value: "bar"}},
			want: []*core.KeyValuePair{{Key: "foo", Value: "bar"}},
		},
		{
			name: "TaskDefault",
			task: false,
			envs: nil,
			want: nil,
		},
		{
			name: "TaskEnable",
			task: true,
			envs: []*core.KeyValuePair{{Key: "foo", Value: "bar"}},
			want: []*core.KeyValuePair{{Key: "foo", Value: "bar"}},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request := testutils.GetExecutionRequest()
			if tt.task {
				request.Spec.LaunchPlan.ResourceType = core.ResourceType_TASK
			}
			request.Spec.Envs.Values = tt.envs

			repository := getMockRepositoryForExecTest()
			setDefaultLpCallbackForExecTest(repository)
			setDefaultTaskCallbackForExecTest(repository)

			exCreateFunc := func(ctx context.Context, input models.Execution) error {
				var spec admin.ExecutionSpec
				err := proto.Unmarshal(input.Spec, &spec)
				assert.Nil(t, err)

				if tt.task {
					assert.Equal(t, uint(0), input.LaunchPlanID)
					assert.NotEqual(t, uint(0), input.TaskID)
				} else {
					assert.NotEqual(t, uint(0), input.LaunchPlanID)
					assert.Equal(t, uint(0), input.TaskID)
				}
				if len(tt.envs) != 0 {
					assert.Equal(t, tt.envs[0].Key, spec.GetEnvs().Values[0].Key)
					assert.Equal(t, tt.envs[0].Value, spec.GetEnvs().Values[0].Value)
				} else {
					assert.Nil(t, spec.GetEnvs().GetValues())
				}

				return nil
			}

			repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(exCreateFunc)
			mockExecutor := workflowengineMocks.WorkflowExecutor{}
			mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{}, nil)
			mockExecutor.OnID().Return("testMockExecutor")
			r := plugins.NewRegistry()
			r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
			execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})

			_, err := execManager.CreateExecution(context.Background(), request, requestedAt)
			assert.Nil(t, err)
		})
	}
}

func TestCreateExecution_CustomNamespaceMappingConfig(t *testing.T) {
	request := testutils.GetExecutionRequest()
	repository := getMockRepositoryForExecTest()
	storageClient := getMockStorageForExecTest(context.Background())
	setDefaultLpCallbackForExecTest(repository)
	mockClock := clock.NewMock()
	createdAt := time.Now()
	mockClock.Set(createdAt)
	exCreateFunc := func(ctx context.Context, input models.Execution) error {
		var spec admin.ExecutionSpec
		err := proto.Unmarshal(input.Spec, &spec)
		assert.NoError(t, err)
		assert.Equal(t, spec.GetMetadata().GetSystemMetadata().Namespace, "project")
		return nil
	}

	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(exCreateFunc)
	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{}, nil)
	mockExecutor.OnID().Return("testMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)

	mockNs := runtimeMocks.NamespaceMappingConfiguration{}
	mockNs.OnGetNamespaceTemplate().Return("{{ project }}")
	mockExecutionsConfigProvider := runtimeMocks.NewMockConfigurationProvider(
		testutils.GetApplicationConfigWithDefaultDomains(),
		runtimeMocks.NewMockQueueConfigurationProvider(
			[]runtimeInterfaces.ExecutionQueue{}, []runtimeInterfaces.WorkflowConfig{}),
		nil,
		runtimeMocks.NewMockTaskResourceConfiguration(resourceDefaults, resourceLimits), nil, &mockNs)
	mockExecutionsConfigProvider.(*runtimeMocks.MockConfigurationProvider).AddRegistrationValidationConfiguration(
		runtimeMocks.NewMockRegistrationValidationProvider())

	execManager := NewExecutionManager(repository, r, mockExecutionsConfigProvider, storageClient, mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})

	execManager.(*ExecutionManager)._clock = mockClock

	response, err := execManager.CreateExecution(context.Background(), request, requestedAt)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&executionIdentifier, response.Id))
}

func makeExecutionGetFunc(
	t *testing.T, closureBytes []byte, startTime *time.Time) repositoryMocks.GetExecutionFunc {
	return func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
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
			Spec:         getExpectedSpecBytes(),
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
	return func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
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

func makeExecutionInterruptibleGetFunc(
	t *testing.T, closureBytes []byte, startTime *time.Time, interruptible *bool) repositoryMocks.GetExecutionFunc {
	return func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
		assert.Equal(t, "project", input.Project)
		assert.Equal(t, "domain", input.Domain)
		assert.Equal(t, "name", input.Name)

		request := testutils.GetExecutionRequest()
		if interruptible == nil {
			request.Spec.Interruptible = nil
		} else {
			request.Spec.Interruptible = &wrappers.BoolValue{Value: *interruptible}
		}

		specBytes, err := proto.Marshal(request.Spec)
		assert.Nil(t, err)

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

func makeExecutionOverwriteCacheGetFunc(
	t *testing.T, closureBytes []byte, startTime *time.Time, overwriteCache bool) repositoryMocks.GetExecutionFunc {
	return func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
		assert.Equal(t, "project", input.Project)
		assert.Equal(t, "domain", input.Domain)
		assert.Equal(t, "name", input.Name)

		request := testutils.GetExecutionRequest()
		request.Spec.OverwriteCache = overwriteCache

		specBytes, err := proto.Marshal(request.Spec)
		assert.Nil(t, err)

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

func makeExecutionWithEnvs(
	t *testing.T, closureBytes []byte, startTime *time.Time, envs []*core.KeyValuePair) repositoryMocks.GetExecutionFunc {
	return func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
		assert.Equal(t, "project", input.Project)
		assert.Equal(t, "domain", input.Domain)
		assert.Equal(t, "name", input.Name)

		request := testutils.GetExecutionRequest()
		request.Spec.Envs.Values = envs

		specBytes, err := proto.Marshal(request.Spec)
		assert.Nil(t, err)

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

func TestRelaunchExecution(t *testing.T) {
	// Set up mocks.
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{}, nil)
	mockExecutor.OnID().Return("testMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
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

	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})

	expectedErr := errors.New("expected error")
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
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
	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{}, nil)
	mockExecutor.OnID().Return("testMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
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

func TestRelaunchExecutionInterruptibleOverride(t *testing.T) {
	// Set up mocks.
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{}, nil)
	mockExecutor.OnID().Return("testMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
	startTime := time.Now()
	startTimeProto, _ := ptypes.TimestampProto(startTime)
	existingClosure := admin.ExecutionClosure{
		Phase:     core.WorkflowExecution_RUNNING,
		StartedAt: startTimeProto,
	}
	existingClosureBytes, _ := proto.Marshal(&existingClosure)
	interruptible := true
	executionGetFunc := makeExecutionInterruptibleGetFunc(t, existingClosureBytes, &startTime, &interruptible)
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
		assert.NotNil(t, spec.GetInterruptible())
		assert.True(t, spec.GetInterruptible().GetValue())
		return nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(exCreateFunc)

	_, err := execManager.RelaunchExecution(context.Background(), admin.ExecutionRelaunchRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Name: "relaunchy",
	}, requestedAt)
	assert.Nil(t, err)
	assert.True(t, createCalled)
}

func TestRelaunchExecutionOverwriteCacheOverride(t *testing.T) {
	// Set up mocks.
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{}, nil)
	mockExecutor.OnID().Return("testMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
	startTime := time.Now()
	startTimeProto, _ := ptypes.TimestampProto(startTime)
	existingClosure := admin.ExecutionClosure{
		Phase:     core.WorkflowExecution_RUNNING,
		StartedAt: startTimeProto,
	}
	existingClosureBytes, _ := proto.Marshal(&existingClosure)

	t.Run("override enable", func(t *testing.T) {
		executionGetFunc := makeExecutionOverwriteCacheGetFunc(t, existingClosureBytes, &startTime, false)
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
			assert.True(t, spec.GetOverwriteCache())
			return nil
		}
		repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(exCreateFunc)

		asd, err := execManager.RelaunchExecution(context.Background(), admin.ExecutionRelaunchRequest{
			Id: &core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
			Name:           "relaunchy",
			OverwriteCache: true,
		}, requestedAt)
		assert.Nil(t, err)
		assert.NotNil(t, asd)
		assert.True(t, createCalled)
	})

	t.Run("override disable", func(t *testing.T) {
		executionGetFunc := makeExecutionOverwriteCacheGetFunc(t, existingClosureBytes, &startTime, true)
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
			assert.False(t, spec.GetOverwriteCache())
			return nil
		}
		repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(exCreateFunc)

		asd, err := execManager.RelaunchExecution(context.Background(), admin.ExecutionRelaunchRequest{
			Id: &core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
			Name:           "relaunchy",
			OverwriteCache: false,
		}, requestedAt)
		assert.Nil(t, err)
		assert.NotNil(t, asd)
		assert.True(t, createCalled)
	})

	t.Run("override omitted", func(t *testing.T) {
		executionGetFunc := makeExecutionOverwriteCacheGetFunc(t, existingClosureBytes, &startTime, true)
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
			assert.False(t, spec.GetOverwriteCache())
			return nil
		}
		repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(exCreateFunc)

		asd, err := execManager.RelaunchExecution(context.Background(), admin.ExecutionRelaunchRequest{
			Id: &core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
			Name: "relaunchy",
		}, requestedAt)
		assert.Nil(t, err)
		assert.NotNil(t, asd)
		assert.True(t, createCalled)
	})
}

func TestRelaunchExecutionEnvsOverride(t *testing.T) {
	// Set up mocks.
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{}, nil)
	mockExecutor.OnID().Return("testMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
	startTime := time.Now()
	startTimeProto, _ := ptypes.TimestampProto(startTime)
	existingClosure := admin.ExecutionClosure{
		Phase:     core.WorkflowExecution_RUNNING,
		StartedAt: startTimeProto,
	}
	existingClosureBytes, _ := proto.Marshal(&existingClosure)
	env := []*core.KeyValuePair{{Key: "foo", Value: "bar"}}
	executionGetFunc := makeExecutionWithEnvs(t, existingClosureBytes, &startTime, env)
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
		assert.NotNil(t, spec.GetEnvs())
		assert.Equal(t, spec.GetEnvs().Values[0].Key, env[0].Key)
		assert.Equal(t, spec.GetEnvs().Values[0].Value, env[0].Value)
		return nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(exCreateFunc)

	_, err := execManager.RelaunchExecution(context.Background(), admin.ExecutionRelaunchRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Name: "relaunchy",
	}, requestedAt)
	assert.Nil(t, err)
	assert.True(t, createCalled)
}

func TestRecoverExecution(t *testing.T) {
	// Set up mocks.
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{}, nil)
	mockExecutor.OnID().Return("testMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
	startTime := time.Now()
	startTimeProto, _ := ptypes.TimestampProto(startTime)
	existingClosure := admin.ExecutionClosure{
		Phase:     core.WorkflowExecution_SUCCEEDED,
		StartedAt: startTimeProto,
	}
	existingClosureBytes, _ := proto.Marshal(&existingClosure)
	executionGetFunc := makeExecutionGetFunc(t, existingClosureBytes, &startTime)
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)

	var createCalled bool
	exCreateFunc := func(ctx context.Context, input models.Execution) error {
		createCalled = true
		assert.Equal(t, "recovered", input.Name)
		assert.Equal(t, "domain", input.Domain)
		assert.Equal(t, "project", input.Project)
		assert.Equal(t, uint(8), input.SourceExecutionID)
		var spec admin.ExecutionSpec
		err := proto.Unmarshal(input.Spec, &spec)
		assert.Nil(t, err)
		assert.Equal(t, admin.ExecutionMetadata_RECOVERED, spec.Metadata.Mode)
		assert.Equal(t, int32(admin.ExecutionMetadata_RECOVERED), input.Mode)
		return nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(exCreateFunc)

	// Issue request.
	response, err := execManager.RecoverExecution(context.Background(), admin.ExecutionRecoverRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Name: "recovered",
	}, requestedAt)

	// And verify response.
	assert.Nil(t, err)

	expectedResponse := &admin.ExecutionCreateResponse{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "recovered",
		},
	}
	assert.True(t, createCalled)
	assert.True(t, proto.Equal(expectedResponse, response))
}

func TestRecoverExecution_RecoveredChildNode(t *testing.T) {
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{}, nil)
	mockExecutor.OnID().Return("testMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
	startTime := time.Now()
	startTimeProto, _ := ptypes.TimestampProto(startTime)
	existingClosure := admin.ExecutionClosure{
		Phase:     core.WorkflowExecution_SUCCEEDED,
		StartedAt: startTimeProto,
	}
	existingClosureBytes, _ := proto.Marshal(&existingClosure)
	referencedExecutionID := uint(123)
	ignoredExecutionID := uint(456)
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
		switch input.Name {
		case "name":
			return models.Execution{
				Spec:    getExpectedSpecBytes(),
				Closure: existingClosureBytes,
				BaseModel: models.BaseModel{
					ID: referencedExecutionID,
				},
			}, nil
		case "orig":
			return models.Execution{
				BaseModel: models.BaseModel{
					ID: ignoredExecutionID,
				},
			}, nil
		default:
			return models.Execution{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.InvalidArgument, "unexpected get for execution %s", input.Name)
		}
	})

	parentNodeDatabaseID := uint(12345)
	var createCalled bool
	exCreateFunc := func(ctx context.Context, input models.Execution) error {
		createCalled = true
		assert.Equal(t, "recovered", input.Name)
		assert.Equal(t, "domain", input.Domain)
		assert.Equal(t, "project", input.Project)
		var spec admin.ExecutionSpec
		err := proto.Unmarshal(input.Spec, &spec)
		assert.Nil(t, err)
		assert.Equal(t, admin.ExecutionMetadata_RECOVERED, spec.Metadata.Mode)
		assert.Equal(t, int32(admin.ExecutionMetadata_RECOVERED), input.Mode)
		assert.Equal(t, parentNodeDatabaseID, input.ParentNodeExecutionID)
		assert.Equal(t, referencedExecutionID, input.SourceExecutionID)

		return nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(exCreateFunc)

	parentNodeExecution := core.NodeExecutionIdentifier{
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "p",
			Domain:  "d",
			Name:    "orig",
		},
		NodeId: "parent",
	}
	repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetGetCallback(func(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
		assert.True(t, proto.Equal(&parentNodeExecution, &input.NodeExecutionIdentifier))

		return models.NodeExecution{
			BaseModel: models.BaseModel{
				ID: parentNodeDatabaseID,
			},
		}, nil
	})

	// Issue request.
	response, err := execManager.RecoverExecution(context.Background(), admin.ExecutionRecoverRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Name: "recovered",
		Metadata: &admin.ExecutionMetadata{
			ParentNodeExecution: &parentNodeExecution,
		},
	}, requestedAt)

	// And verify response.
	assert.Nil(t, err)

	expectedResponse := &admin.ExecutionCreateResponse{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "recovered",
		},
	}
	assert.True(t, createCalled)
	assert.True(t, proto.Equal(expectedResponse, response))
}

func TestRecoverExecution_GetExistingFailure(t *testing.T) {
	// Set up mocks.
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})

	expectedErr := errors.New("expected error")
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
			return models.Execution{}, expectedErr
		})

	var createCalled bool
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(
		func(ctx context.Context, input models.Execution) error {
			createCalled = true
			return nil
		})

	// Issue request.
	_, err := execManager.RecoverExecution(context.Background(), admin.ExecutionRecoverRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Name: "recovered",
	}, requestedAt)

	// And verify response.
	assert.EqualError(t, err, expectedErr.Error())
	assert.False(t, createCalled)
}

func TestRecoverExecution_GetExistingInputsFailure(t *testing.T) {
	// Set up mocks.
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)

	expectedErr := errors.New("foo")
	mockStorage := commonMocks.GetMockStorageClient()
	mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).ReadProtobufCb = func(
		ctx context.Context, reference storage.DataReference, msg proto.Message) error {
		return expectedErr
	}
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), mockStorage, mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
	startTime := time.Now()
	startTimeProto, _ := ptypes.TimestampProto(startTime)
	existingClosure := admin.ExecutionClosure{
		Phase:     core.WorkflowExecution_SUCCEEDED,
		StartedAt: startTimeProto,
	}
	existingClosureBytes, _ := proto.Marshal(&existingClosure)
	executionGetFunc := makeExecutionGetFunc(t, existingClosureBytes, &startTime)
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)

	// Issue request.
	_, err := execManager.RecoverExecution(context.Background(), admin.ExecutionRecoverRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Name: "recovered",
	}, requestedAt)

	// And verify response.
	assert.EqualError(t, err, "Unable to read WorkflowClosure from location s3://flyte/metadata/admin/remote closure id : foo")
}

func TestRecoverExecutionInterruptibleOverride(t *testing.T) {
	// Set up mocks.
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{}, nil)
	mockExecutor.OnID().Return("testMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
	startTime := time.Now()
	startTimeProto, _ := ptypes.TimestampProto(startTime)
	existingClosure := admin.ExecutionClosure{
		Phase:     core.WorkflowExecution_SUCCEEDED,
		StartedAt: startTimeProto,
	}
	existingClosureBytes, _ := proto.Marshal(&existingClosure)
	interruptible := true
	executionGetFunc := makeExecutionInterruptibleGetFunc(t, existingClosureBytes, &startTime, &interruptible)
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)

	var createCalled bool
	exCreateFunc := func(ctx context.Context, input models.Execution) error {
		createCalled = true
		assert.Equal(t, "recovered", input.Name)
		assert.Equal(t, "domain", input.Domain)
		assert.Equal(t, "project", input.Project)
		assert.Equal(t, uint(8), input.SourceExecutionID)
		var spec admin.ExecutionSpec
		err := proto.Unmarshal(input.Spec, &spec)
		assert.Nil(t, err)
		assert.Equal(t, admin.ExecutionMetadata_RECOVERED, spec.Metadata.Mode)
		assert.Equal(t, int32(admin.ExecutionMetadata_RECOVERED), input.Mode)
		assert.NotNil(t, spec.GetInterruptible())
		assert.True(t, spec.GetInterruptible().GetValue())
		return nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(exCreateFunc)

	// Issue request.
	response, err := execManager.RecoverExecution(context.Background(), admin.ExecutionRecoverRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Name: "recovered",
	}, requestedAt)

	// And verify response.
	assert.Nil(t, err)

	expectedResponse := &admin.ExecutionCreateResponse{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "recovered",
		},
	}
	assert.True(t, createCalled)
	assert.True(t, proto.Equal(expectedResponse, response))
}

func TestRecoverExecutionOverwriteCacheOverride(t *testing.T) {
	// Set up mocks.
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{}, nil)
	mockExecutor.OnID().Return("testMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
	startTime := time.Now()
	startTimeProto, _ := ptypes.TimestampProto(startTime)
	existingClosure := admin.ExecutionClosure{
		Phase:     core.WorkflowExecution_SUCCEEDED,
		StartedAt: startTimeProto,
	}
	existingClosureBytes, _ := proto.Marshal(&existingClosure)
	executionGetFunc := makeExecutionOverwriteCacheGetFunc(t, existingClosureBytes, &startTime, true)
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)

	var createCalled bool
	exCreateFunc := func(ctx context.Context, input models.Execution) error {
		createCalled = true
		assert.Equal(t, "recovered", input.Name)
		assert.Equal(t, "domain", input.Domain)
		assert.Equal(t, "project", input.Project)
		assert.Equal(t, uint(8), input.SourceExecutionID)
		var spec admin.ExecutionSpec
		err := proto.Unmarshal(input.Spec, &spec)
		assert.Nil(t, err)
		assert.Equal(t, admin.ExecutionMetadata_RECOVERED, spec.Metadata.Mode)
		assert.Equal(t, int32(admin.ExecutionMetadata_RECOVERED), input.Mode)
		assert.True(t, spec.GetOverwriteCache())
		return nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(exCreateFunc)

	// Issue request.
	response, err := execManager.RecoverExecution(context.Background(), admin.ExecutionRecoverRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Name: "recovered",
	}, requestedAt)

	// And verify response.
	assert.Nil(t, err)

	expectedResponse := &admin.ExecutionCreateResponse{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "recovered",
		},
	}
	assert.True(t, createCalled)
	assert.True(t, proto.Equal(expectedResponse, response))
}

func TestRecoverExecutionEnvsOverride(t *testing.T) {
	// Set up mocks.
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{}, nil)
	mockExecutor.OnID().Return("testMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
	startTime := time.Now()
	startTimeProto, _ := ptypes.TimestampProto(startTime)
	existingClosure := admin.ExecutionClosure{
		Phase:     core.WorkflowExecution_SUCCEEDED,
		StartedAt: startTimeProto,
	}
	existingClosureBytes, _ := proto.Marshal(&existingClosure)
	env := []*core.KeyValuePair{{Key: "foo", Value: "bar"}}
	executionGetFunc := makeExecutionWithEnvs(t, existingClosureBytes, &startTime, env)
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)

	exCreateFunc := func(ctx context.Context, input models.Execution) error {
		assert.Equal(t, "recovered", input.Name)
		assert.Equal(t, "domain", input.Domain)
		assert.Equal(t, "project", input.Project)
		assert.Equal(t, uint(8), input.SourceExecutionID)
		var spec admin.ExecutionSpec
		err := proto.Unmarshal(input.Spec, &spec)
		assert.Nil(t, err)
		assert.Equal(t, admin.ExecutionMetadata_RECOVERED, spec.Metadata.Mode)
		assert.Equal(t, int32(admin.ExecutionMetadata_RECOVERED), input.Mode)
		assert.NotNil(t, spec.GetEnvs())
		assert.Equal(t, spec.GetEnvs().GetValues()[0].Key, env[0].Key)
		assert.Equal(t, spec.GetEnvs().GetValues()[0].Value, env[0].Value)
		return nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(exCreateFunc)

	// Issue request.
	response, err := execManager.RecoverExecution(context.Background(), admin.ExecutionRecoverRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Name: "recovered",
	}, requestedAt)

	// And verify response.
	assert.Nil(t, err)

	expectedResponse := &admin.ExecutionCreateResponse{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "recovered",
		},
	}
	assert.True(t, proto.Equal(expectedResponse, response))
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
		context context.Context, execution models.Execution) error {

		assert.Equal(t, "project", execution.Project)
		assert.Equal(t, "domain", execution.Domain)
		assert.Equal(t, "name", execution.Name)
		assert.Equal(t, uint(1), execution.LaunchPlanID)
		assert.Equal(t, uint(2), execution.WorkflowID)
		assert.Equal(t, core.WorkflowExecution_FAILED.String(), execution.Phase)
		assert.Equal(t, closureBytes, execution.Closure)
		assert.Equal(t, getExpectedSpecBytes(), execution.Spec)
		assert.Equal(t, startTime, *execution.StartedAt)
		assert.Equal(t, duration, execution.Duration)
		return nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetUpdateCallback(updateExecutionFunc)
	request := admin.WorkflowExecutionEventRequest{
		RequestId: "1",
		Event: &event.WorkflowExecutionEvent{
			ExecutionId: &executionIdentifier,
			OccurredAt:  occurredAt,
			Phase:       core.WorkflowExecution_FAILED,
			OutputResult: &event.WorkflowExecutionEvent_Error{
				Error: &executionError,
			},
			ProducerId: testCluster,
		},
	}
	mockDbEventWriter := &eventWriterMocks.WorkflowExecutionEventWriter{}
	mockDbEventWriter.On("Write", request)
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, &mockPublisher, &mockPublisher, mockDbEventWriter)
	resp, err := execManager.CreateWorkflowEvent(context.Background(), request)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestCreateWorkflowEvent_TerminalState(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	executionGetFunc := func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
		return models.Execution{
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
			BaseModel: models.BaseModel{
				ID: uint(8),
			},
			Spec:  getExpectedSpecBytes(),
			Phase: core.WorkflowExecution_FAILED.String(),
		}, nil
	}

	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)
	updateExecutionFunc := func(context context.Context, execution models.Execution) error {
		return nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetUpdateCallback(updateExecutionFunc)
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})

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

func TestCreateWorkflowEvent_NoRunningToQueued(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	executionGetFunc := func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
		return models.Execution{
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
			Spec:  getExpectedSpecBytes(),
			Phase: core.WorkflowExecution_RUNNING.String(),
		}, nil
	}

	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)
	updateExecutionFunc := func(context context.Context, execution models.Execution) error {
		return nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetUpdateCallback(updateExecutionFunc)
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})

	resp, err := execManager.CreateWorkflowEvent(context.Background(), admin.WorkflowExecutionEventRequest{
		RequestId: "1",
		Event: &event.WorkflowExecutionEvent{
			ExecutionId: &executionIdentifier,
			Phase:       core.WorkflowExecution_QUEUED,
		},
	})
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	adminError := err.(flyteAdminErrors.FlyteAdminError)
	assert.Equal(t, adminError.Code(), codes.FailedPrecondition)
}

func TestCreateWorkflowEvent_CurrentlyAborting(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	executionGetFunc := func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
		return models.Execution{
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
			Spec:  getExpectedSpecBytes(),
			Phase: core.WorkflowExecution_ABORTING.String(),
		}, nil
	}

	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)
	updateExecutionFunc := func(context context.Context, execution models.Execution) error {
		return nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetUpdateCallback(updateExecutionFunc)

	req := admin.WorkflowExecutionEventRequest{
		RequestId: "1",
		Event: &event.WorkflowExecutionEvent{
			ExecutionId: &executionIdentifier,
			Phase:       core.WorkflowExecution_ABORTED,
			OccurredAt:  timestamppb.New(time.Now()),
		},
	}

	mockDbEventWriter := &eventWriterMocks.WorkflowExecutionEventWriter{}
	mockDbEventWriter.On("Write", req)
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, &mockPublisher, &mockPublisher, mockDbEventWriter)

	resp, err := execManager.CreateWorkflowEvent(context.Background(), req)
	assert.NotNil(t, resp)
	assert.NoError(t, err)

	req.Event.Phase = core.WorkflowExecution_QUEUED
	resp, err = execManager.CreateWorkflowEvent(context.Background(), req)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	adminError := err.(flyteAdminErrors.FlyteAdminError)
	assert.Equal(t, adminError.Code(), codes.FailedPrecondition)

	req.Event.Phase = core.WorkflowExecution_RUNNING
	resp, err = execManager.CreateWorkflowEvent(context.Background(), req)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	adminError = err.(flyteAdminErrors.FlyteAdminError)
	assert.Equal(t, adminError.Code(), codes.FailedPrecondition)
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
		StateChangeDetails: &admin.ExecutionStateChangeDetails{
			State:      admin.ExecutionState_EXECUTION_ACTIVE,
			OccurredAt: testutils.MockCreatedAtProto,
		},
	}
	closureBytes, _ := proto.Marshal(&closure)
	updateExecutionFunc := func(
		context context.Context, execution models.Execution) error {
		assert.Equal(t, "project", execution.Project)
		assert.Equal(t, "domain", execution.Domain)
		assert.Equal(t, "name", execution.Name)
		assert.Equal(t, uint(1), execution.LaunchPlanID)
		assert.Equal(t, uint(2), execution.WorkflowID)
		assert.Equal(t, core.WorkflowExecution_RUNNING.String(), execution.Phase)
		assert.Equal(t, closureBytes, execution.Closure)
		assert.Equal(t, getExpectedSpecBytes(), execution.Spec)
		assert.Equal(t, occurredAt, *execution.StartedAt)
		assert.Equal(t, time.Duration(0), execution.Duration)
		assert.Equal(t, occurredAt, *execution.ExecutionUpdatedAt)
		return nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetUpdateCallback(updateExecutionFunc)
	occurredAtTimestamp, _ := ptypes.TimestampProto(occurredAt)
	request := admin.WorkflowExecutionEventRequest{
		RequestId: "1",
		Event: &event.WorkflowExecutionEvent{
			ExecutionId: &executionIdentifier,
			OccurredAt:  occurredAtTimestamp,
			Phase:       core.WorkflowExecution_RUNNING,
			ProducerId:  testCluster,
		},
	}
	mockDbEventWriter := &eventWriterMocks.WorkflowExecutionEventWriter{}
	mockDbEventWriter.On("Write", request)
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, &mockPublisher, &mockPublisher, mockDbEventWriter)
	resp, err := execManager.CreateWorkflowEvent(context.Background(), request)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestCreateWorkflowEvent_DuplicateRunning(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	occurredAt := time.Now().UTC()

	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
			return models.Execution{
				ExecutionKey: models.ExecutionKey{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
				BaseModel: models.BaseModel{
					ID: uint(8),
				},
				Spec:         getExpectedSpecBytes(),
				Phase:        core.WorkflowExecution_RUNNING.String(),
				Closure:      closureBytes,
				LaunchPlanID: uint(1),
				WorkflowID:   uint(2),
				StartedAt:    &occurredAt,
			}, nil
		},
	)

	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
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
		func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
			return models.Execution{
				ExecutionKey: models.ExecutionKey{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
				BaseModel: models.BaseModel{
					ID: uint(8),
				},
				Spec:         getExpectedSpecBytes(),
				Phase:        core.WorkflowExecution_SUCCEEDED.String(),
				Closure:      closureBytes,
				LaunchPlanID: uint(1),
				WorkflowID:   uint(2),
				StartedAt:    &occurredAt,
			}, nil
		},
	)

	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
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

func TestCreateWorkflowEvent_ClusterReassignmentOnQueued(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	occurredAt := time.Now().UTC()

	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
			return models.Execution{
				ExecutionKey: models.ExecutionKey{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
				BaseModel: models.BaseModel{
					ID: uint(8),
				},
				Spec:         getExpectedSpecBytes(),
				Phase:        core.WorkflowExecution_UNDEFINED.String(),
				Closure:      closureBytes,
				LaunchPlanID: uint(1),
				WorkflowID:   uint(2),
				StartedAt:    &occurredAt,
			}, nil
		},
	)
	newCluster := "C2"
	updateExecutionFunc := func(
		context context.Context, execution models.Execution) error {
		assert.Equal(t, core.WorkflowExecution_QUEUED.String(), execution.Phase)
		assert.Equal(t, newCluster, execution.Cluster)
		return nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetUpdateCallback(updateExecutionFunc)

	occurredAtTimestamp, _ := ptypes.TimestampProto(occurredAt)
	mockDbEventWriter := &eventWriterMocks.WorkflowExecutionEventWriter{}
	request := admin.WorkflowExecutionEventRequest{
		RequestId: "1",
		Event: &event.WorkflowExecutionEvent{
			ExecutionId: &executionIdentifier,
			OccurredAt:  occurredAtTimestamp,
			Phase:       core.WorkflowExecution_QUEUED,
			ProducerId:  newCluster,
		},
	}
	mockDbEventWriter.On("Write", request)
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, &mockPublisher, &mockPublisher, mockDbEventWriter)

	resp, err := execManager.CreateWorkflowEvent(context.Background(), request)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
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
		context context.Context, execution models.Execution) error {
		return nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetUpdateCallback(updateExecutionFunc)
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
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

	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
	resp, err := execManager.CreateWorkflowEvent(context.Background(), admin.WorkflowExecutionEventRequest{
		RequestId: "1",
		Event: &event.WorkflowExecutionEvent{
			ExecutionId: &executionIdentifier,
			OccurredAt:  occurredAt,
			Phase:       core.WorkflowExecution_FAILED,
			OutputResult: &event.WorkflowExecutionEvent_Error{
				Error: &executionError,
			},
			ProducerId: testCluster,
		},
	})
	assert.Nil(t, resp)
	assert.Equal(t, codes.Internal, err.(flyteAdminErrors.FlyteAdminError).Code())
}

func TestCreateWorkflowEvent_DatabaseGetError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	startTime := time.Now()

	expectedErr := errors.New("expected error")
	executionGetFunc := func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
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
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
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
		context context.Context, execution models.Execution) error {
		return expectedErr
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetUpdateCallback(updateExecutionFunc)
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
	resp, err := execManager.CreateWorkflowEvent(context.Background(), admin.WorkflowExecutionEventRequest{
		RequestId: "1",
		Event: &event.WorkflowExecutionEvent{
			ExecutionId: &executionIdentifier,
			OccurredAt:  occurredAt,
			Phase:       core.WorkflowExecution_FAILED,
			OutputResult: &event.WorkflowExecutionEvent_Error{
				Error: &executionError,
			},
			ProducerId: testCluster,
		},
	})
	assert.Nil(t, resp)
	assert.EqualError(t, expectedErr, err.Error())
}

func TestCreateWorkflowEvent_IncompatibleCluster(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	occurredAt := time.Now().UTC()

	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
			return models.Execution{
				ExecutionKey: models.ExecutionKey{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
				BaseModel: models.BaseModel{
					ID: uint(8),
				},
				Spec:         getExpectedSpecBytes(),
				Phase:        core.WorkflowExecution_RUNNING.String(),
				Closure:      closureBytes,
				LaunchPlanID: uint(1),
				WorkflowID:   uint(2),
				StartedAt:    &occurredAt,
				Cluster:      testCluster,
			}, nil
		},
	)

	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
	occurredAtTimestamp, _ := ptypes.TimestampProto(occurredAt)
	resp, err := execManager.CreateWorkflowEvent(context.Background(), admin.WorkflowExecutionEventRequest{
		RequestId: "1",
		Event: &event.WorkflowExecutionEvent{
			ExecutionId: &executionIdentifier,
			OccurredAt:  occurredAtTimestamp,
			Phase:       core.WorkflowExecution_ABORTING,
			ProducerId:  "C2",
		},
	})
	assert.NotNil(t, err)
	adminError := err.(flyteAdminErrors.FlyteAdminError)
	assert.Equal(t, adminError.Code(), codes.FailedPrecondition)
	s, ok := status.FromError(err)
	assert.True(t, ok)
	var seenIncompatibleClusterErr bool
	for _, detail := range s.Details() {
		failureReason, ok := detail.(*admin.EventFailureReason)
		if ok {
			if failureReason.GetIncompatibleCluster() != nil {
				seenIncompatibleClusterErr = true
				break
			}
		}
	}
	assert.True(t, seenIncompatibleClusterErr)
	assert.Nil(t, resp)
}

func TestGetExecution(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	startedAt := time.Date(2018, 8, 30, 0, 0, 0, 0, time.UTC)
	executionGetFunc := func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
		assert.Equal(t, "project", input.Project)
		assert.Equal(t, "domain", input.Domain)
		assert.Equal(t, "name", input.Name)
		return models.Execution{
			BaseModel: models.BaseModel{
				CreatedAt: testutils.MockCreatedAtValue,
			},
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
			Spec:         getExpectedSpecBytes(),
			Phase:        phase,
			Closure:      closureBytes,
			LaunchPlanID: uint(1),
			WorkflowID:   uint(2),
			StartedAt:    &startedAt,
			// TODO: Input uri
		}, nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
	execution, err := execManager.GetExecution(context.Background(), admin.WorkflowExecutionGetRequest{
		Id: &executionIdentifier,
	})
	assert.NoError(t, err)
	assert.True(t, proto.Equal(&executionIdentifier, execution.Id))
	assert.True(t, proto.Equal(getExpectedSpec(), execution.Spec))
	assert.True(t, proto.Equal(&closure, execution.Closure))
}

func TestGetExecution_DatabaseError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	expectedErr := errors.New("expected error")

	executionGetFunc := func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
		assert.Equal(t, "project", input.Project)
		assert.Equal(t, "domain", input.Domain)
		assert.Equal(t, "name", input.Name)
		return models.Execution{}, expectedErr
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
	execution, err := execManager.GetExecution(context.Background(), admin.WorkflowExecutionGetRequest{
		Id: &executionIdentifier,
	})
	assert.Nil(t, execution)
	assert.Equal(t, expectedErr, err)
}

func TestGetExecution_TransformerError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	startedAt := time.Date(2018, 8, 30, 0, 0, 0, 0, time.UTC)
	executionGetFunc := func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
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
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
	execution, err := execManager.GetExecution(context.Background(), admin.WorkflowExecutionGetRequest{
		Id: &executionIdentifier,
	})
	assert.Nil(t, execution)
	assert.Equal(t, codes.Internal, err.(flyteAdminErrors.FlyteAdminError).Code())
}

func TestUpdateExecution(t *testing.T) {
	t.Run("invalid execution identifier", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		r := plugins.NewRegistry()
		r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
		execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
		_, err := execManager.UpdateExecution(context.Background(), admin.ExecutionUpdateRequest{
			Id: &core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
			},
		}, time.Now())
		assert.Error(t, err)
	})

	t.Run("empty status passed", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		updateExecFuncCalled := false
		updateExecFunc := func(ctx context.Context, execModel models.Execution) error {
			stateInt := int32(admin.ExecutionState_EXECUTION_ACTIVE)
			assert.Equal(t, stateInt, *execModel.State)
			updateExecFuncCalled = true
			return nil
		}
		repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetUpdateCallback(updateExecFunc)
		r := plugins.NewRegistry()
		r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
		execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
		updateResponse, err := execManager.UpdateExecution(context.Background(), admin.ExecutionUpdateRequest{
			Id: &executionIdentifier,
		}, time.Now())
		assert.NoError(t, err)
		assert.NotNil(t, updateResponse)
		assert.True(t, updateExecFuncCalled)
	})

	t.Run("archive status passed", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		updateExecFuncCalled := false
		updateExecFunc := func(ctx context.Context, execModel models.Execution) error {
			stateInt := int32(admin.ExecutionState_EXECUTION_ARCHIVED)
			assert.Equal(t, stateInt, *execModel.State)
			updateExecFuncCalled = true
			return nil
		}
		repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetUpdateCallback(updateExecFunc)
		r := plugins.NewRegistry()
		r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
		execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
		updateResponse, err := execManager.UpdateExecution(context.Background(), admin.ExecutionUpdateRequest{
			Id:    &executionIdentifier,
			State: admin.ExecutionState_EXECUTION_ARCHIVED,
		}, time.Now())
		assert.NoError(t, err)
		assert.NotNil(t, updateResponse)
		assert.True(t, updateExecFuncCalled)
	})

	t.Run("update error", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		updateExecFunc := func(ctx context.Context, execModel models.Execution) error {
			return fmt.Errorf("some db error")
		}
		repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetUpdateCallback(updateExecFunc)
		r := plugins.NewRegistry()
		r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
		execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
		_, err := execManager.UpdateExecution(context.Background(), admin.ExecutionUpdateRequest{
			Id:    &executionIdentifier,
			State: admin.ExecutionState_EXECUTION_ARCHIVED,
		}, time.Now())
		assert.Error(t, err)
		assert.Equal(t, "some db error", err.Error())
	})

	t.Run("get execution error", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		getExecFunc := func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
			return models.Execution{}, fmt.Errorf("some db error")
		}
		repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(getExecFunc)
		r := plugins.NewRegistry()
		r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
		execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
		_, err := execManager.UpdateExecution(context.Background(), admin.ExecutionUpdateRequest{
			Id:    &executionIdentifier,
			State: admin.ExecutionState_EXECUTION_ARCHIVED,
		}, time.Now())
		assert.Error(t, err)
		assert.Equal(t, "some db error", err.Error())
	})
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
		assert.Equal(t, "execution_domain asc", input.SortParameter.GetGormOrderExpr())
		assert.Equal(t, 2, input.Offset)
		assert.EqualValues(t, map[common.Entity]bool{
			common.Execution: true,
		}, input.JoinTableEntities)
		return interfaces.ExecutionCollectionOutput{
			Executions: []models.Execution{
				{
					BaseModel: models.BaseModel{
						CreatedAt: testutils.MockCreatedAtValue,
					},
					ExecutionKey: models.ExecutionKey{
						Project: projectValue,
						Domain:  domainValue,
						Name:    "my awesome execution",
					},
					Spec:    getExpectedSpecBytes(),
					Closure: closureBytes,
				},
				{
					BaseModel: models.BaseModel{
						CreatedAt: testutils.MockCreatedAtValue,
					},
					ExecutionKey: models.ExecutionKey{
						Project: projectValue,
						Domain:  domainValue,
						Name:    "my other execution",
					},
					Phase:   core.WorkflowExecution_SUCCEEDED.String(),
					Spec:    getExpectedSpecBytes(),
					Closure: closureBytes,
				},
			},
		}, nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetListCallback(executionListFunc)
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})

	executionList, err := execManager.ListExecutions(context.Background(), admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: projectValue,
			Domain:  domainValue,
		},
		Limit: limit,
		SortBy: &admin.Sort{
			Direction: admin.Sort_ASCENDING,
			Key:       "execution_domain",
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
		assert.True(t, proto.Equal(getExpectedSpec(), execution.Spec))
		assert.True(t, proto.Equal(&closure, execution.Closure))
	}
	assert.Empty(t, executionList.Token)
}

func TestListExecutions_MissingParameters(t *testing.T) {
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repositoryMocks.NewMockRepository(), r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
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
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
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
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})

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
	queue := executions.NewQueueAllocator(getMockExecutionsConfigProvider(), repository)

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
		Spec:         getExpectedSpecBytes(),
	}
	assert.Nil(t, myExecManager.publishNotifications(context.Background(), workflowRequest, executionModel))
}

func TestExecutionManager_PublishNotificationsTransformError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	queue := executions.NewQueueAllocator(getMockExecutionsConfigProvider(), repository)
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
	queue := executions.NewQueueAllocator(getMockExecutionsConfigProvider(), repository)
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
		Spec:         getExpectedSpecBytes(),
	}
	assert.Nil(t, myExecManager.publishNotifications(context.Background(), workflowRequest, executionModel))

}

func TestExecutionManager_PublishNotificationsNoPhaseMatch(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	queue := executions.NewQueueAllocator(getMockExecutionsConfigProvider(), repository)

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
	principal := "principal"
	updateExecutionFunc := func(
		context context.Context, execution models.Execution) error {
		assert.Equal(t, "project", execution.Project)
		assert.Equal(t, "domain", execution.Domain)
		assert.Equal(t, "name", execution.Name)
		assert.Equal(t, uint(1), execution.LaunchPlanID)
		assert.Equal(t, uint(2), execution.WorkflowID)
		assert.Equal(t, core.WorkflowExecution_ABORTING.String(), execution.Phase)
		assert.Equal(t, execution.ExecutionCreatedAt, execution.ExecutionUpdatedAt,
			"an abort call should not change ExecutionUpdatedAt until a corresponding execution event is received")
		assert.Equal(t, abortCause, execution.AbortCause)
		assert.Equal(t, testCluster, execution.Cluster)

		var unmarshaledClosure admin.ExecutionClosure
		err := proto.Unmarshal(execution.Closure, &unmarshaledClosure)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(&admin.AbortMetadata{
			Cause:     abortCause,
			Principal: principal,
		}, unmarshaledClosure.GetAbortMetadata()))
		return nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetUpdateCallback(updateExecutionFunc)

	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnAbortMatch(mock.Anything, mock.MatchedBy(func(data workflowengineInterfaces.AbortData) bool {
		assert.True(t, proto.Equal(&core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		}, data.ExecutionID))
		return true
	})).Return(nil)
	mockExecutor.OnID().Return("customMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})

	identity, err := auth.NewIdentityContext("", principal, "", time.Now(), sets.NewString(), nil, nil)
	assert.NoError(t, err)
	ctx := identity.WithContext(context.Background())
	resp, err := execManager.TerminateExecution(ctx, admin.ExecutionTerminateRequest{
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

	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnAbortMatch(mock.Anything, mock.Anything).Return(expectedError)
	mockExecutor.OnID().Return("customMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)

	updateCalled := false
	repository := repositoryMocks.NewMockRepository()
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetUpdateCallback(func(
		context context.Context, execution models.Execution) error {
		updateCalled = true
		assert.Equal(t, core.WorkflowExecution_ABORTING.String(), execution.Phase)
		return nil
	})
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})

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
	assert.True(t, updateCalled)
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
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetUpdateCallback(updateExecutionFunc)
	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnAbortMatch(mock.Anything, mock.Anything).Return(nil)
	mockExecutor.OnID().Return("testMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
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

func TestTerminateExecution_AlreadyTerminated(t *testing.T) {
	var expectedError = errors.New("expected error")

	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnAbortMatch(mock.Anything, mock.Anything).Return(expectedError)
	mockExecutor.OnID().Return("customMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)

	repository := repositoryMocks.NewMockRepository()
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
			return models.Execution{
				Phase: core.WorkflowExecution_SUCCEEDED.String(),
			}, nil
		})
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
	resp, err := execManager.TerminateExecution(context.Background(), admin.ExecutionTerminateRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Cause: "abort cause",
	})

	assert.Nil(t, resp)
	s, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.FailedPrecondition, s.Code())
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

	executionGetFunc := func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
		return models.Execution{
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
			Spec:         getExpectedSpecBytes(),
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
	mockStorage := commonMocks.GetMockStorageClient()
	fullInputs := &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"foo": testutils.MakeStringLiteral("foo-value-1"),
		},
	}
	fullOutputs := &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"bar": testutils.MakeStringLiteral("bar-value-1"),
		},
	}
	mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).ReadProtobufCb = func(
		ctx context.Context, reference storage.DataReference, msg proto.Message) error {
		if reference.String() == "inputs" {
			marshalled, _ := proto.Marshal(fullInputs)
			_ = proto.Unmarshal(marshalled, msg)
			return nil
		} else if reference.String() == outputURI {
			marshalled, _ := proto.Marshal(fullOutputs)
			_ = proto.Unmarshal(marshalled, msg)
			return nil
		}
		return fmt.Errorf("unexpected call to find value in storage [%v]", reference.String())
	}

	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), mockStorage, mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
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
		FullInputs:  fullInputs,
		FullOutputs: fullOutputs,
	}, dataResponse))
}

func TestResolveStringMap_RuntimeLimitsObserved(t *testing.T) {
	_, err := resolveStringMap(&admin.Labels{
		Values: map[string]string{
			"dynamiclabel1": "dynamic1",
			"dynamiclabel2": "dynamic2",
		},
	}, &admin.Labels{
		Values: map[string]string{
			"existing1": "value1",
		},
	}, "labels", 1)
	assert.EqualError(t, err, "labels has too many entries [2 > 1]")
}

func TestAddPluginOverrides(t *testing.T) {
	executionID := &core.WorkflowExecutionIdentifier{
		Project: project,
		Domain:  domain,
		Name:    "unused",
	}
	workflowName := "workflow_name"
	launchPlanName := "launch_plan_name"

	db := repositoryMocks.NewMockRepository()
	db.ResourceRepo().(*repositoryMocks.MockResourceRepo).GetFunction = func(ctx context.Context, ID interfaces.ResourceID) (
		models.Resource, error) {
		assert.Equal(t, project, ID.Project)
		assert.Equal(t, domain, ID.Domain)
		assert.Equal(t, workflowName, ID.Workflow)
		assert.Equal(t, launchPlanName, ID.LaunchPlan)
		existingAttributes := commonTestUtils.GetPluginOverridesAttributes(map[string][]string{
			"python": {"plugin a"},
			"hive":   {"plugin b"},
		})
		bytes, err := proto.Marshal(existingAttributes)
		if err != nil {
			t.Fatal(err)
		}
		return models.Resource{
			Project:    project,
			Domain:     domain,
			Attributes: bytes,
		}, nil
	}
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(db, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})

	taskPluginOverrides, err := execManager.(*ExecutionManager).addPluginOverrides(
		context.Background(), executionID, workflowName, launchPlanName)
	assert.NoError(t, err)
	assert.Len(t, taskPluginOverrides, 2)
	for _, override := range taskPluginOverrides {
		if override.TaskType == "python" {
			assert.EqualValues(t, []string{"plugin a"}, override.PluginId)
		} else if override.TaskType == "hive" {
			assert.EqualValues(t, []string{"plugin b"}, override.PluginId)
		} else {
			t.Errorf("Unexpected task type [%s] plugin override committed to db", override.TaskType)
		}
	}
}

func TestPluginOverrides_ResourceGetFailure(t *testing.T) {
	executionID := &core.WorkflowExecutionIdentifier{
		Project: project,
		Domain:  domain,
		Name:    "unused",
	}
	workflowName := "workflow_name"
	launchPlanName := "launch_plan_name"

	db := repositoryMocks.NewMockRepository()
	db.ResourceRepo().(*repositoryMocks.MockResourceRepo).GetFunction = func(ctx context.Context, ID interfaces.ResourceID) (
		models.Resource, error) {
		return models.Resource{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.Aborted, "uh oh")
	}
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(db, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})

	_, err := execManager.(*ExecutionManager).addPluginOverrides(
		context.Background(), executionID, workflowName, launchPlanName)
	assert.Error(t, err, "uh oh")
}

func TestGetExecution_Legacy(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	startedAt := time.Date(2018, 8, 30, 0, 0, 0, 0, time.UTC)
	executionGetFunc := func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
		assert.Equal(t, "project", input.Project)
		assert.Equal(t, "domain", input.Domain)
		assert.Equal(t, "name", input.Name)
		return models.Execution{
			BaseModel: models.BaseModel{
				CreatedAt: testutils.MockCreatedAtValue,
			},
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
			Spec:         getExpectedLegacySpecBytes(),
			Phase:        phase,
			Closure:      getLegacyClosureBytes(),
			LaunchPlanID: uint(1),
			WorkflowID:   uint(2),
			StartedAt:    &startedAt,
		}, nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
	execution, err := execManager.GetExecution(context.Background(), admin.WorkflowExecutionGetRequest{
		Id: &executionIdentifier,
	})
	assert.NoError(t, err)
	assert.True(t, proto.Equal(&executionIdentifier, execution.Id))
	assert.True(t, proto.Equal(getExpectedLegacySpec(), execution.Spec))
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

	executionGetFunc := func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
		return models.Execution{
			BaseModel: models.BaseModel{
				CreatedAt: testutils.MockCreatedAtValue,
			},
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
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), storageClient, mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
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
		FullInputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"foo": testutils.MakeStringLiteral("foo-value-1"),
			},
		},
		FullOutputs: &core.LiteralMap{},
	}, dataResponse))
	var inputs core.LiteralMap
	err = storageClient.ReadProtobuf(context.Background(), storage.DataReference("s3://bucket/metadata/project/domain/name/inputs"), &inputs)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&inputs, closure.ComputedInputs))
}

func TestCreateExecution_LegacyClient(t *testing.T) {
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)

	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.MatchedBy(func(execData workflowengineInterfaces.ExecutionData) bool {
		assert.EqualValues(t, map[string]string{
			"label1": "1",
			"label2": "2",
		}, execData.ExecutionParameters.Labels)
		assert.EqualValues(t, map[string]string{
			"annotation3": "3",
			"annotation4": "4",
		}, execData.ExecutionParameters.Annotations)
		return true
	})).Return(workflowengineInterfaces.ExecutionResponse{
		Cluster: testCluster,
	}, nil)
	mockExecutor.OnID().Return("customMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
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

	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{}, nil)
	mockExecutor.OnID().Return("testMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), storageClient, mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
	startTime := time.Now()
	startTimeProto, _ := ptypes.TimestampProto(startTime)
	existingClosure := getLegacyClosure()
	existingClosure.Phase = core.WorkflowExecution_RUNNING
	existingClosure.StartedAt = startTimeProto
	existingClosure.ComputedInputs.Literals["bar"] = coreutils.MustMakeLiteral("bar-value")
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
		assert.Equal(t, "default_raw_output", spec.RawOutputDataConfig.OutputLocationPrefix)
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
		assert.Equal(t, "execution_domain asc", input.SortParameter.GetGormOrderExpr())
		assert.Equal(t, 2, input.Offset)
		return interfaces.ExecutionCollectionOutput{
			Executions: []models.Execution{
				{
					BaseModel: models.BaseModel{
						CreatedAt: testutils.MockCreatedAtValue,
					},
					ExecutionKey: models.ExecutionKey{
						Project: projectValue,
						Domain:  domainValue,
						Name:    "my awesome execution",
					},
					Spec:    getLegacySpecBytes(),
					Closure: getLegacyClosureBytes(),
				},
				{
					BaseModel: models.BaseModel{
						CreatedAt: testutils.MockCreatedAtValue,
					},
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
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})

	executionList, err := execManager.ListExecutions(context.Background(), admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: projectValue,
			Domain:  domainValue,
		},
		Limit: limit,
		SortBy: &admin.Sort{
			Direction: admin.Sort_ASCENDING,
			Key:       "execution_domain",
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

func TestSetDefaults(t *testing.T) {
	task := &core.CompiledTask{
		Template: &core.TaskTemplate{
			Target: &core.TaskTemplate_Container{
				Container: &core.Container{
					Resources: &core.Resources{
						Requests: []*core.Resources_ResourceEntry{
							{
								Name:  core.Resources_CPU,
								Value: "250m",
							},
						},
					},
				},
			},
			Id: &core.Identifier{
				Project: "project",
				Domain:  "domain",
				Name:    "task_name",
				Version: "version",
			},
		},
	}

	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repositoryMocks.NewMockRepository(), r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
	execManager.(*ExecutionManager).setCompiledTaskDefaults(context.Background(), task, workflowengineInterfaces.TaskResources{
		Defaults: runtimeInterfaces.TaskResourceSet{
			CPU:              resource.MustParse("200m"),
			GPU:              resource.MustParse("4"),
			Memory:           resource.MustParse("200Gi"),
			EphemeralStorage: resource.MustParse("500Mi"),
		},
		Limits: runtimeInterfaces.TaskResourceSet{
			CPU:              resource.MustParse("300m"),
			GPU:              resource.MustParse("8"),
			Memory:           resource.MustParse("500Gi"),
			EphemeralStorage: resource.MustParse("501Mi"),
		},
	})
	assert.True(t, proto.Equal(
		&core.Container{
			Resources: &core.Resources{
				Requests: []*core.Resources_ResourceEntry{
					{
						Name:  core.Resources_CPU,
						Value: "250m",
					},
					{
						Name:  core.Resources_MEMORY,
						Value: "200Gi",
					},
					{
						Name:  core.Resources_EPHEMERAL_STORAGE,
						Value: "500Mi",
					},
					{
						Name:  core.Resources_GPU,
						Value: "4",
					},
				},
				Limits: []*core.Resources_ResourceEntry{
					{
						Name:  core.Resources_CPU,
						Value: "250m",
					},
					{
						Name:  core.Resources_MEMORY,
						Value: "200Gi",
					},
					{
						Name:  core.Resources_EPHEMERAL_STORAGE,
						Value: "500Mi",
					},
					{
						Name:  core.Resources_GPU,
						Value: "4",
					},
				},
			},
		},
		task.Template.GetContainer()), fmt.Sprintf("%+v", task.Template.GetContainer()))
}

func TestSetDefaults_MissingRequests_ExistingRequestsPreserved(t *testing.T) {
	task := &core.CompiledTask{
		Template: &core.TaskTemplate{
			Target: &core.TaskTemplate_Container{
				Container: &core.Container{
					Resources: &core.Resources{
						Requests: []*core.Resources_ResourceEntry{
							{
								Name:  core.Resources_CPU,
								Value: "250m",
							},
						},
					},
				},
			},
			Id: &core.Identifier{
				Project: "project",
				Domain:  "domain",
				Name:    "task_name",
				Version: "version",
			},
		},
	}

	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	execManager := NewExecutionManager(repositoryMocks.NewMockRepository(), r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
	execManager.(*ExecutionManager).setCompiledTaskDefaults(context.Background(), task, workflowengineInterfaces.TaskResources{
		Defaults: runtimeInterfaces.TaskResourceSet{
			CPU:    resource.MustParse("200m"),
			GPU:    resource.MustParse("4"),
			Memory: resource.MustParse("200Gi"),
		},
		Limits: runtimeInterfaces.TaskResourceSet{
			CPU: resource.MustParse("300m"),
			GPU: resource.MustParse("8"),
			// Because only the limit is set, this resource should not be injected.
			EphemeralStorage: resource.MustParse("100"),
		},
	})
	assert.True(t, proto.Equal(
		&core.Container{
			Resources: &core.Resources{
				Requests: []*core.Resources_ResourceEntry{
					{
						Name:  core.Resources_CPU,
						Value: "250m",
					},
					{
						Name:  core.Resources_MEMORY,
						Value: "200Gi",
					},
					{
						Name:  core.Resources_GPU,
						Value: "4",
					},
				},
				Limits: []*core.Resources_ResourceEntry{
					{
						Name:  core.Resources_CPU,
						Value: "250m",
					},
					{
						Name:  core.Resources_MEMORY,
						Value: "200Gi",
					},
					{
						Name:  core.Resources_GPU,
						Value: "4",
					},
				},
			},
		},
		task.Template.GetContainer()), fmt.Sprintf("%+v", task.Template.GetContainer()))
}

func TestSetDefaults_OptionalRequiredResources(t *testing.T) {
	taskConfigLimits := runtimeInterfaces.TaskResourceSet{
		CPU:              resource.MustParse("300m"),
		GPU:              resource.MustParse("1"),
		Memory:           resource.MustParse("500Gi"),
		EphemeralStorage: resource.MustParse("501Mi"),
	}

	task := &core.CompiledTask{
		Template: &core.TaskTemplate{
			Target: &core.TaskTemplate_Container{
				Container: &core.Container{
					Resources: &core.Resources{
						Requests: []*core.Resources_ResourceEntry{
							{
								Name:  core.Resources_CPU,
								Value: "200m",
							},
						},
					},
				},
			},
			Id: &taskIdentifier,
		},
	}
	t.Run("don't inject ephemeral storage or gpu when only the limit is set in config", func(t *testing.T) {
		r := plugins.NewRegistry()
		r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
		execManager := NewExecutionManager(repositoryMocks.NewMockRepository(), r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
		execManager.(*ExecutionManager).setCompiledTaskDefaults(context.Background(), task, workflowengineInterfaces.TaskResources{
			Defaults: runtimeInterfaces.TaskResourceSet{
				CPU:    resource.MustParse("200m"),
				Memory: resource.MustParse("200Gi"),
			},
			Limits: taskConfigLimits,
		})
		assert.True(t, proto.Equal(
			&core.Container{
				Resources: &core.Resources{
					Requests: []*core.Resources_ResourceEntry{
						{
							Name:  core.Resources_CPU,
							Value: "200m",
						},
						{
							Name:  core.Resources_MEMORY,
							Value: "200Gi",
						},
					},
					Limits: []*core.Resources_ResourceEntry{
						{
							Name:  core.Resources_CPU,
							Value: "200m",
						},
						{
							Name:  core.Resources_MEMORY,
							Value: "200Gi",
						},
					},
				},
			},
			task.Template.GetContainer()), fmt.Sprintf("%+v", task.Template.GetContainer()))
	})

	t.Run("respect non-required resources when defaults exist in config", func(t *testing.T) {
		r := plugins.NewRegistry()
		r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
		execManager := NewExecutionManager(repositoryMocks.NewMockRepository(), r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
		execManager.(*ExecutionManager).setCompiledTaskDefaults(context.Background(), task, workflowengineInterfaces.TaskResources{
			Limits: taskConfigLimits,
			Defaults: runtimeInterfaces.TaskResourceSet{
				CPU:              resource.MustParse("200m"),
				Memory:           resource.MustParse("200Gi"),
				EphemeralStorage: resource.MustParse("1"),
			},
		})
		assert.True(t, proto.Equal(
			&core.Container{
				Resources: &core.Resources{
					Requests: []*core.Resources_ResourceEntry{
						{
							Name:  core.Resources_CPU,
							Value: "200m",
						},
						{
							Name:  core.Resources_MEMORY,
							Value: "200Gi",
						},
						{
							Name:  core.Resources_EPHEMERAL_STORAGE,
							Value: "1",
						},
					},
					Limits: []*core.Resources_ResourceEntry{
						{
							Name:  core.Resources_CPU,
							Value: "200m",
						},
						{
							Name:  core.Resources_MEMORY,
							Value: "200Gi",
						},
						{
							Name:  core.Resources_EPHEMERAL_STORAGE,
							Value: "1",
						},
					},
				},
			},
			task.Template.GetContainer()), fmt.Sprintf("%+v", task.Template.GetContainer()))
	})

}
func TestCreateSingleTaskExecution(t *testing.T) {
	repository := getMockRepositoryForExecTest()
	var getCalledCount = 0
	var newlyCreatedWorkflow models.Workflow
	workflowCreateFunc := func(input models.Workflow, descriptionEntity *models.DescriptionEntity) error {
		newlyCreatedWorkflow = input
		return nil
	}
	repository.WorkflowRepo().(*repositoryMocks.MockWorkflowRepo).SetCreateCallback(workflowCreateFunc)

	workflowGetFunc := func(input interfaces.Identifier) (models.Workflow, error) {
		if getCalledCount <= 1 {
			getCalledCount++
			return models.Workflow{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound, "not found")
		}
		getCalledCount++
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
							Inputs: &core.VariableMap{
								Variables: map[string]*core.Variable{
									"a": {
										Type: &core.LiteralType{
											Type: &core.LiteralType_Simple{
												Simple: core.SimpleType_INTEGER,
											},
										},
									},
								},
							},
							Outputs: &core.VariableMap{
								Variables: map[string]*core.Variable{
									"b": {
										Type: &core.LiteralType{
											Type: &core.LiteralType_Simple{
												Simple: core.SimpleType_INTEGER,
											},
										},
									},
								},
							},
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
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(
		func(ctx context.Context, input models.Execution) error {
			var spec admin.ExecutionSpec
			err := proto.Unmarshal(input.Spec, &spec)
			assert.NoError(t, err)
			assert.Equal(t, models.ExecutionKey{
				Project: "flytekit",
				Domain:  "production",
				Name:    "singletaskexec",
			}, input.ExecutionKey)
			assert.Equal(t, "task", input.LaunchEntity)
			assert.Equal(t, "UNDEFINED", input.Phase)
			assert.True(t, proto.Equal(taskIdentifier, spec.LaunchPlan))
			return nil
		})

	var launchplan *models.LaunchPlan
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetCreateCallback(func(input models.LaunchPlan) error {
		launchplan = &input
		return nil
	})
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetGetCallback(func(input interfaces.Identifier) (models.LaunchPlan, error) {
		if launchplan == nil {
			return models.LaunchPlan{}, flyteAdminErrors.NewFlyteAdminError(codes.NotFound, "launchplan not found")
		}
		return *launchplan, nil
	})

	mockStorage := getMockStorageForExecTest(context.Background())
	workflowManager := NewWorkflowManager(
		repository,
		getMockWorkflowConfigProvider(), getMockWorkflowCompiler(), mockStorage,
		storagePrefix, mockScope.NewTestScope())
	namedEntityManager := NewNamedEntityManager(repository, getMockConfigForNETest(), mockScope.NewTestScope())

	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{}, nil)
	mockExecutor.OnID().Return("testMockExecutor")
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), mockStorage, mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, workflowManager, namedEntityManager, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{})
	request := admin.ExecutionCreateRequest{
		Project: "flytekit",
		Domain:  "production",
		Name:    "singletaskexec",
		Spec: &admin.ExecutionSpec{
			LaunchPlan: taskIdentifier,
		},
		Inputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"a": {
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Primitive{
								Primitive: &core.Primitive{
									Value: &core.Primitive_Integer{
										Integer: 999,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	marshaller := jsonpb.Marshaler{}
	_, ferr := marshaller.MarshalToString(&request)
	assert.NoError(t, ferr)

	// test once to create an initial launchplan
	_, err := execManager.CreateExecution(context.TODO(), request, time.Now())
	assert.NoError(t, err)

	// test again to ensure existing launchplan retrieval works
	_, err = execManager.CreateExecution(context.TODO(), request, time.Now())
	assert.NoError(t, err)
}

func TestGetExecutionConfigOverrides(t *testing.T) {
	requestLabels := map[string]string{"requestLabelKey": "requestLabelValue"}
	requestAnnotations := map[string]string{"requestAnnotationKey": "requestAnnotationValue"}
	requestOutputLocationPrefix := "requestOutputLocationPrefix"
	requestK8sServiceAccount := "requestK8sServiceAccount"
	requestMaxParallelism := int32(10)
	requestInterruptible := false
	requestOverwriteCache := false
	requestEnvironmentVariables := []*core.KeyValuePair{{Key: "hello", Value: "world"}}

	launchPlanLabels := map[string]string{"launchPlanLabelKey": "launchPlanLabelValue"}
	launchPlanAnnotations := map[string]string{"launchPlanAnnotationKey": "launchPlanAnnotationValue"}
	launchPlanOutputLocationPrefix := "launchPlanOutputLocationPrefix"
	launchPlanK8sServiceAccount := "launchPlanK8sServiceAccount"
	launchPlanAssumableIamRole := "launchPlanAssumableIamRole"
	launchPlanMaxParallelism := int32(50)
	launchPlanInterruptible := true
	launchPlanOverwriteCache := true
	launchPlanEnvironmentVariables := []*core.KeyValuePair{{Key: "foo", Value: "bar"}}

	applicationConfig := runtime.NewConfigurationProvider()

	defaultK8sServiceAccount := applicationConfig.ApplicationConfiguration().GetTopLevelConfig().K8SServiceAccount
	defaultMaxParallelism := applicationConfig.ApplicationConfiguration().GetTopLevelConfig().MaxParallelism

	deprecatedLaunchPlanK8sServiceAccount := "deprecatedLaunchPlanK8sServiceAccount"
	rmLabels := map[string]string{"rmLabelKey": "rmLabelValue"}
	rmAnnotations := map[string]string{"rmAnnotationKey": "rmAnnotationValue"}
	rmOutputLocationPrefix := "rmOutputLocationPrefix"
	rmK8sServiceAccount := "rmK8sServiceAccount"
	rmMaxParallelism := int32(80)
	rmInterruptible := false
	rmOverwriteCache := false

	resourceManager := managerMocks.MockResourceManager{}
	executionManager := ExecutionManager{
		resourceManager: &resourceManager,
		config:          applicationConfig,
	}
	resourceManager.GetResourceFunc = func(ctx context.Context,
		request managerInterfaces.ResourceRequest) (*managerInterfaces.ResourceResponse, error) {
		// two requests will be made, one with empty domain and one with filled in domain
		assert.Contains(t, []managerInterfaces.ResourceRequest{{
			Project:      workflowIdentifier.Project,
			Domain:       workflowIdentifier.Domain,
			ResourceType: admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG,
		}, {Project: workflowIdentifier.Project,
			Domain:       "",
			ResourceType: admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG},
		}, request)
		projectDomainResponse := &managerInterfaces.ResourceResponse{
			Attributes: &admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_WorkflowExecutionConfig{
					WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{
						MaxParallelism: rmMaxParallelism,
						Interruptible:  &wrappers.BoolValue{Value: rmInterruptible},
						OverwriteCache: rmOverwriteCache,
						Annotations:    &admin.Annotations{Values: rmAnnotations},
						RawOutputDataConfig: &admin.RawOutputDataConfig{
							OutputLocationPrefix: rmOutputLocationPrefix,
						},
						SecurityContext: &core.SecurityContext{
							RunAs: &core.Identity{
								K8SServiceAccount: rmK8sServiceAccount,
							},
						},
					},
				},
			},
		}

		projectResponse := &managerInterfaces.ResourceResponse{
			Attributes: &admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_WorkflowExecutionConfig{
					WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{
						Labels: &admin.Labels{Values: rmLabels},
						RawOutputDataConfig: &admin.RawOutputDataConfig{
							OutputLocationPrefix: "shouldnotbeused",
						},
					},
				},
			},
		}
		if request.Domain == "" {
			return projectResponse, nil
		}
		return projectDomainResponse, nil
	}

	t.Run("request with full config", func(t *testing.T) {
		request := &admin.ExecutionCreateRequest{
			Project: workflowIdentifier.Project,
			Domain:  workflowIdentifier.Domain,
			Spec: &admin.ExecutionSpec{
				Labels:      &admin.Labels{Values: requestLabels},
				Annotations: &admin.Annotations{Values: requestAnnotations},
				RawOutputDataConfig: &admin.RawOutputDataConfig{
					OutputLocationPrefix: requestOutputLocationPrefix,
				},
				SecurityContext: &core.SecurityContext{
					RunAs: &core.Identity{
						K8SServiceAccount: requestK8sServiceAccount,
					},
				},
				MaxParallelism: requestMaxParallelism,
				Interruptible:  &wrappers.BoolValue{Value: requestInterruptible},
				OverwriteCache: requestOverwriteCache,
				Envs:           &admin.Envs{Values: requestEnvironmentVariables},
			},
		}
		identityContext, err := auth.NewIdentityContext("", "", "", time.Now(), sets.String{}, nil, nil)
		assert.NoError(t, err)
		identityContext = identityContext.WithExecutionUserIdentifier("yeee")
		ctx := identityContext.WithContext(context.Background())
		execConfig, err := executionManager.getExecutionConfig(ctx, request, nil)
		assert.NoError(t, err)
		assert.Equal(t, requestMaxParallelism, execConfig.MaxParallelism)
		assert.Equal(t, requestK8sServiceAccount, execConfig.SecurityContext.RunAs.K8SServiceAccount)
		assert.Equal(t, requestInterruptible, execConfig.Interruptible.Value)
		assert.Equal(t, requestOverwriteCache, execConfig.OverwriteCache)
		assert.Equal(t, requestOutputLocationPrefix, execConfig.RawOutputDataConfig.OutputLocationPrefix)
		assert.Equal(t, requestLabels, execConfig.GetLabels().Values)
		assert.Equal(t, requestAnnotations, execConfig.GetAnnotations().Values)
		assert.Equal(t, "yeee", execConfig.GetSecurityContext().GetRunAs().GetExecutionIdentity())
		assert.Equal(t, requestEnvironmentVariables, execConfig.GetEnvs().Values)
	})
	t.Run("request with partial config", func(t *testing.T) {
		request := &admin.ExecutionCreateRequest{
			Project: workflowIdentifier.Project,
			Domain:  workflowIdentifier.Domain,
			Spec: &admin.ExecutionSpec{
				Labels: &admin.Labels{Values: requestLabels},
				RawOutputDataConfig: &admin.RawOutputDataConfig{
					OutputLocationPrefix: requestOutputLocationPrefix,
				},
				MaxParallelism: requestMaxParallelism,
			},
		}
		launchPlan := &admin.LaunchPlan{
			Spec: &admin.LaunchPlanSpec{
				Annotations:         &admin.Annotations{Values: launchPlanAnnotations},
				Labels:              &admin.Labels{Values: launchPlanLabels},
				RawOutputDataConfig: &admin.RawOutputDataConfig{OutputLocationPrefix: launchPlanOutputLocationPrefix},
				SecurityContext: &core.SecurityContext{
					RunAs: &core.Identity{
						K8SServiceAccount: launchPlanK8sServiceAccount,
						IamRole:           launchPlanAssumableIamRole,
					},
				},
				MaxParallelism: launchPlanMaxParallelism,
				Interruptible:  &wrappers.BoolValue{Value: launchPlanInterruptible},
				OverwriteCache: launchPlanOverwriteCache,
				Envs:           &admin.Envs{Values: launchPlanEnvironmentVariables},
			},
		}
		execConfig, err := executionManager.getExecutionConfig(context.TODO(), request, launchPlan)
		assert.NoError(t, err)
		assert.Equal(t, requestMaxParallelism, execConfig.MaxParallelism)
		assert.Equal(t, launchPlanInterruptible, execConfig.Interruptible.Value)
		assert.Equal(t, launchPlanOverwriteCache, execConfig.OverwriteCache)
		assert.True(t, proto.Equal(launchPlan.Spec.SecurityContext, execConfig.SecurityContext))
		assert.True(t, proto.Equal(launchPlan.Spec.Annotations, execConfig.Annotations))
		assert.Equal(t, requestOutputLocationPrefix, execConfig.RawOutputDataConfig.OutputLocationPrefix)
		assert.Equal(t, requestLabels, execConfig.GetLabels().Values)
		assert.Equal(t, launchPlanEnvironmentVariables, execConfig.GetEnvs().Values)
	})
	t.Run("request with empty security context", func(t *testing.T) {
		request := &admin.ExecutionCreateRequest{
			Project: workflowIdentifier.Project,
			Domain:  workflowIdentifier.Domain,
			Spec: &admin.ExecutionSpec{
				SecurityContext: &core.SecurityContext{
					RunAs: &core.Identity{
						K8SServiceAccount: "",
						IamRole:           "",
					},
				},
			},
		}
		launchPlan := &admin.LaunchPlan{
			Spec: &admin.LaunchPlanSpec{
				Annotations:         &admin.Annotations{Values: launchPlanAnnotations},
				Labels:              &admin.Labels{Values: launchPlanLabels},
				RawOutputDataConfig: &admin.RawOutputDataConfig{OutputLocationPrefix: launchPlanOutputLocationPrefix},
				SecurityContext: &core.SecurityContext{
					RunAs: &core.Identity{
						K8SServiceAccount: launchPlanK8sServiceAccount,
					},
				},
				MaxParallelism: launchPlanMaxParallelism,
				Interruptible:  &wrappers.BoolValue{Value: launchPlanInterruptible},
				OverwriteCache: launchPlanOverwriteCache,
				Envs:           &admin.Envs{Values: launchPlanEnvironmentVariables},
			},
		}
		execConfig, err := executionManager.getExecutionConfig(context.TODO(), request, launchPlan)
		assert.NoError(t, err)
		assert.Equal(t, launchPlanMaxParallelism, execConfig.MaxParallelism)
		assert.Equal(t, launchPlanInterruptible, execConfig.Interruptible.Value)
		assert.Equal(t, launchPlanOverwriteCache, execConfig.OverwriteCache)
		assert.Equal(t, launchPlanK8sServiceAccount, execConfig.SecurityContext.RunAs.K8SServiceAccount)
		assert.Equal(t, launchPlanOutputLocationPrefix, execConfig.RawOutputDataConfig.OutputLocationPrefix)
		assert.Equal(t, launchPlanLabels, execConfig.GetLabels().Values)
		assert.Equal(t, launchPlanEnvironmentVariables, execConfig.GetEnvs().Values)
	})
	t.Run("request with no config", func(t *testing.T) {
		request := &admin.ExecutionCreateRequest{
			Project: workflowIdentifier.Project,
			Domain:  workflowIdentifier.Domain,
			Spec:    &admin.ExecutionSpec{},
		}
		launchPlan := &admin.LaunchPlan{
			Spec: &admin.LaunchPlanSpec{
				Labels:      &admin.Labels{Values: launchPlanLabels},
				Annotations: &admin.Annotations{Values: launchPlanAnnotations},
				RawOutputDataConfig: &admin.RawOutputDataConfig{
					OutputLocationPrefix: launchPlanOutputLocationPrefix,
				},
				SecurityContext: &core.SecurityContext{
					RunAs: &core.Identity{
						K8SServiceAccount: launchPlanK8sServiceAccount,
					},
				},
				MaxParallelism: launchPlanMaxParallelism,
				Interruptible:  &wrappers.BoolValue{Value: launchPlanInterruptible},
				OverwriteCache: launchPlanOverwriteCache,
				Envs:           &admin.Envs{Values: launchPlanEnvironmentVariables},
			},
		}
		execConfig, err := executionManager.getExecutionConfig(context.TODO(), request, launchPlan)
		assert.NoError(t, err)
		assert.Equal(t, launchPlanMaxParallelism, execConfig.MaxParallelism)
		assert.Equal(t, launchPlanInterruptible, execConfig.Interruptible.Value)
		assert.Equal(t, launchPlanOverwriteCache, execConfig.OverwriteCache)
		assert.Equal(t, launchPlanK8sServiceAccount, execConfig.SecurityContext.RunAs.K8SServiceAccount)
		assert.Equal(t, launchPlanOutputLocationPrefix, execConfig.RawOutputDataConfig.OutputLocationPrefix)
		assert.Equal(t, launchPlanLabels, execConfig.GetLabels().Values)
		assert.Equal(t, launchPlanAnnotations, execConfig.GetAnnotations().Values)
		assert.Equal(t, launchPlanEnvironmentVariables, execConfig.GetEnvs().Values)
	})
	t.Run("launchplan with partial config", func(t *testing.T) {
		request := &admin.ExecutionCreateRequest{
			Project: workflowIdentifier.Project,
			Domain:  workflowIdentifier.Domain,
			Spec:    &admin.ExecutionSpec{},
		}
		launchPlan := &admin.LaunchPlan{
			Spec: &admin.LaunchPlanSpec{
				Labels:      &admin.Labels{Values: launchPlanLabels},
				Annotations: &admin.Annotations{Values: launchPlanAnnotations},
				RawOutputDataConfig: &admin.RawOutputDataConfig{
					OutputLocationPrefix: launchPlanOutputLocationPrefix,
				},
				SecurityContext: &core.SecurityContext{
					RunAs: &core.Identity{
						K8SServiceAccount: launchPlanK8sServiceAccount,
					},
				},
				MaxParallelism: launchPlanMaxParallelism,
			},
		}
		execConfig, err := executionManager.getExecutionConfig(context.TODO(), request, launchPlan)
		assert.NoError(t, err)
		assert.Equal(t, launchPlanMaxParallelism, execConfig.MaxParallelism)
		assert.Equal(t, rmInterruptible, execConfig.Interruptible.Value)
		assert.Equal(t, rmOverwriteCache, execConfig.OverwriteCache)
		assert.Equal(t, launchPlanK8sServiceAccount, execConfig.SecurityContext.RunAs.K8SServiceAccount)
		assert.Equal(t, launchPlanOutputLocationPrefix, execConfig.RawOutputDataConfig.OutputLocationPrefix)
		assert.Equal(t, launchPlanLabels, execConfig.GetLabels().Values)
		assert.Equal(t, launchPlanAnnotations, execConfig.GetAnnotations().Values)
	})
	t.Run("launchplan with no config", func(t *testing.T) {
		request := &admin.ExecutionCreateRequest{
			Project: workflowIdentifier.Project,
			Domain:  workflowIdentifier.Domain,
			Spec:    &admin.ExecutionSpec{},
		}
		launchPlan := &admin.LaunchPlan{
			Spec: &admin.LaunchPlanSpec{},
		}
		execConfig, err := executionManager.getExecutionConfig(context.TODO(), request, launchPlan)
		assert.NoError(t, err)
		assert.Equal(t, rmMaxParallelism, execConfig.MaxParallelism)
		assert.Equal(t, rmInterruptible, execConfig.Interruptible.Value)
		assert.Equal(t, rmOverwriteCache, execConfig.OverwriteCache)
		assert.Equal(t, rmK8sServiceAccount, execConfig.SecurityContext.RunAs.K8SServiceAccount)
		assert.Equal(t, rmOutputLocationPrefix, execConfig.RawOutputDataConfig.OutputLocationPrefix)
		assert.Equal(t, rmLabels, execConfig.GetLabels().Values)
		assert.Equal(t, rmAnnotations, execConfig.GetAnnotations().Values)
		assert.Nil(t, execConfig.GetEnvs())
	})
	t.Run("matchable resource partial config", func(t *testing.T) {
		resourceManager.GetResourceFunc = func(ctx context.Context,
			request managerInterfaces.ResourceRequest) (*managerInterfaces.ResourceResponse, error) {
			assert.Contains(t, []managerInterfaces.ResourceRequest{{
				Project:      workflowIdentifier.Project,
				Domain:       workflowIdentifier.Domain,
				ResourceType: admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG,
			}, {Project: workflowIdentifier.Project,
				Domain:       "",
				ResourceType: admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG},
			}, request)

			return &managerInterfaces.ResourceResponse{
				Attributes: &admin.MatchingAttributes{
					Target: &admin.MatchingAttributes_WorkflowExecutionConfig{
						WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{
							MaxParallelism: rmMaxParallelism,
							Annotations:    &admin.Annotations{Values: rmAnnotations},
							SecurityContext: &core.SecurityContext{
								RunAs: &core.Identity{
									K8SServiceAccount: rmK8sServiceAccount,
								},
							},
						},
					},
				},
			}, nil
		}
		request := &admin.ExecutionCreateRequest{
			Project: workflowIdentifier.Project,
			Domain:  workflowIdentifier.Domain,
			Spec:    &admin.ExecutionSpec{},
		}
		launchPlan := &admin.LaunchPlan{
			Spec: &admin.LaunchPlanSpec{},
		}
		execConfig, err := executionManager.getExecutionConfig(context.TODO(), request, launchPlan)
		assert.NoError(t, err)
		assert.Equal(t, rmMaxParallelism, execConfig.MaxParallelism)
		assert.Nil(t, execConfig.GetInterruptible())
		assert.False(t, execConfig.OverwriteCache)
		assert.Equal(t, rmK8sServiceAccount, execConfig.SecurityContext.RunAs.K8SServiceAccount)
		assert.Nil(t, execConfig.GetRawOutputDataConfig())
		assert.Nil(t, execConfig.GetLabels())
		assert.Equal(t, rmAnnotations, execConfig.GetAnnotations().Values)
		assert.Nil(t, execConfig.GetEnvs())
	})
	t.Run("matchable resource with no config", func(t *testing.T) {
		resourceManager.GetResourceFunc = func(ctx context.Context,
			request managerInterfaces.ResourceRequest) (*managerInterfaces.ResourceResponse, error) {
			assert.Contains(t, []managerInterfaces.ResourceRequest{{
				Project:      workflowIdentifier.Project,
				Domain:       workflowIdentifier.Domain,
				ResourceType: admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG,
			}, {Project: workflowIdentifier.Project,
				Domain:       "",
				ResourceType: admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG},
			}, request)
			return &managerInterfaces.ResourceResponse{
				Attributes: &admin.MatchingAttributes{
					Target: &admin.MatchingAttributes_WorkflowExecutionConfig{
						WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{},
					},
				},
			}, nil
		}
		request := &admin.ExecutionCreateRequest{
			Project: workflowIdentifier.Project,
			Domain:  workflowIdentifier.Domain,
			Spec:    &admin.ExecutionSpec{},
		}
		launchPlan := &admin.LaunchPlan{
			Spec: &admin.LaunchPlanSpec{},
		}
		execConfig, err := executionManager.getExecutionConfig(context.TODO(), request, launchPlan)
		assert.NoError(t, err)
		assert.Equal(t, defaultMaxParallelism, execConfig.MaxParallelism)
		assert.Nil(t, execConfig.GetInterruptible())
		assert.False(t, execConfig.OverwriteCache)
		assert.Equal(t, defaultK8sServiceAccount, execConfig.SecurityContext.RunAs.K8SServiceAccount)
		assert.Nil(t, execConfig.GetRawOutputDataConfig())
		assert.Nil(t, execConfig.GetLabels())
		assert.Nil(t, execConfig.GetAnnotations())
		assert.Nil(t, execConfig.GetEnvs())
	})
	t.Run("fetch security context from deprecated config", func(t *testing.T) {
		resourceManager.GetResourceFunc = func(ctx context.Context,
			request managerInterfaces.ResourceRequest) (*managerInterfaces.ResourceResponse, error) {
			assert.Contains(t, []managerInterfaces.ResourceRequest{{
				Project:      workflowIdentifier.Project,
				Domain:       workflowIdentifier.Domain,
				ResourceType: admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG,
			}, {Project: workflowIdentifier.Project,
				Domain:       "",
				ResourceType: admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG},
			}, request)

			return &managerInterfaces.ResourceResponse{
				Attributes: &admin.MatchingAttributes{
					Target: &admin.MatchingAttributes_WorkflowExecutionConfig{
						WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{},
					},
				},
			}, nil
		}
		request := &admin.ExecutionCreateRequest{
			Project: workflowIdentifier.Project,
			Domain:  workflowIdentifier.Domain,
			Spec:    &admin.ExecutionSpec{},
		}
		launchPlan := &admin.LaunchPlan{
			Spec: &admin.LaunchPlanSpec{
				AuthRole: &admin.AuthRole{
					KubernetesServiceAccount: deprecatedLaunchPlanK8sServiceAccount,
				},
			},
		}
		execConfig, err := executionManager.getExecutionConfig(context.TODO(), request, launchPlan)
		assert.NoError(t, err)
		assert.Equal(t, defaultMaxParallelism, execConfig.MaxParallelism)
		assert.Nil(t, execConfig.GetInterruptible())
		assert.False(t, execConfig.OverwriteCache)
		assert.Equal(t, deprecatedLaunchPlanK8sServiceAccount, execConfig.SecurityContext.RunAs.K8SServiceAccount)
		assert.Nil(t, execConfig.GetRawOutputDataConfig())
		assert.Nil(t, execConfig.GetLabels())
		assert.Nil(t, execConfig.GetAnnotations())
		assert.Nil(t, execConfig.GetEnvs())
	})
	t.Run("matchable resource workflow resource", func(t *testing.T) {
		resourceManager.GetResourceFunc = func(ctx context.Context,
			request managerInterfaces.ResourceRequest) (*managerInterfaces.ResourceResponse, error) {
			assert.Contains(t, []managerInterfaces.ResourceRequest{{
				Project:      workflowIdentifier.Project,
				Domain:       workflowIdentifier.Domain,
				ResourceType: admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG,
				Workflow:     workflowIdentifier.Name,
			}, {Project: workflowIdentifier.Project,
				Domain:       "",
				Workflow:     "",
				ResourceType: admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG},
			}, request)

			return &managerInterfaces.ResourceResponse{
				Attributes: &admin.MatchingAttributes{
					Target: &admin.MatchingAttributes_WorkflowExecutionConfig{
						WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{
							MaxParallelism: 300,
							Interruptible:  &wrappers.BoolValue{Value: true},
							OverwriteCache: true,
							SecurityContext: &core.SecurityContext{
								RunAs: &core.Identity{
									K8SServiceAccount: "workflowDefault",
								},
							},
						},
					},
				},
			}, nil
		}
		request := &admin.ExecutionCreateRequest{
			Project: workflowIdentifier.Project,
			Domain:  workflowIdentifier.Domain,
			Spec:    &admin.ExecutionSpec{},
		}
		launchPlan := &admin.LaunchPlan{
			Spec: &admin.LaunchPlanSpec{
				WorkflowId: &core.Identifier{
					Name: workflowIdentifier.Name,
				},
			},
		}
		execConfig, err := executionManager.getExecutionConfig(context.TODO(), request, launchPlan)
		assert.NoError(t, err)
		assert.Equal(t, int32(300), execConfig.MaxParallelism)
		assert.True(t, execConfig.Interruptible.Value)
		assert.True(t, execConfig.OverwriteCache)
		assert.Equal(t, "workflowDefault", execConfig.SecurityContext.RunAs.K8SServiceAccount)
		assert.Nil(t, execConfig.GetRawOutputDataConfig())
		assert.Nil(t, execConfig.GetLabels())
		assert.Nil(t, execConfig.GetAnnotations())
		assert.Nil(t, execConfig.GetEnvs())
	})
	t.Run("matchable resource failure", func(t *testing.T) {
		resourceManager.GetResourceFunc = func(ctx context.Context,
			request managerInterfaces.ResourceRequest) (*managerInterfaces.ResourceResponse, error) {
			assert.Contains(t, []managerInterfaces.ResourceRequest{{
				Project:      workflowIdentifier.Project,
				Domain:       workflowIdentifier.Domain,
				ResourceType: admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG,
			}, {Project: workflowIdentifier.Project,
				Domain:       "",
				ResourceType: admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG},
			}, request)
			return nil, fmt.Errorf("failed to fetch the resources")
		}
		request := &admin.ExecutionCreateRequest{
			Project: workflowIdentifier.Project,
			Domain:  workflowIdentifier.Domain,
			Spec:    &admin.ExecutionSpec{},
		}
		launchPlan := &admin.LaunchPlan{
			Spec: &admin.LaunchPlanSpec{},
		}
		execConfig, err := executionManager.getExecutionConfig(context.TODO(), request, launchPlan)
		assert.Equal(t, fmt.Errorf("failed to fetch the resources"), err)
		assert.Nil(t, execConfig.GetInterruptible())
		assert.False(t, execConfig.GetOverwriteCache())
		assert.Nil(t, execConfig.GetSecurityContext())
		assert.Nil(t, execConfig.GetRawOutputDataConfig())
		assert.Nil(t, execConfig.GetLabels())
		assert.Nil(t, execConfig.GetAnnotations())
		assert.Nil(t, execConfig.GetEnvs())
	})

	t.Run("application configuration", func(t *testing.T) {
		resourceManager.GetResourceFunc = func(ctx context.Context,
			request managerInterfaces.ResourceRequest) (*managerInterfaces.ResourceResponse, error) {
			assert.Contains(t, []managerInterfaces.ResourceRequest{{
				Project:      workflowIdentifier.Project,
				Domain:       workflowIdentifier.Domain,
				ResourceType: admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG,
			}, {Project: workflowIdentifier.Project,
				Domain:       "",
				ResourceType: admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG},
			}, request)
			return &managerInterfaces.ResourceResponse{
				Attributes: &admin.MatchingAttributes{
					Target: &admin.MatchingAttributes_WorkflowExecutionConfig{
						WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{},
					},
				},
			}, nil
		}

		executionManager.config.ApplicationConfiguration().GetTopLevelConfig().Interruptible = true
		executionManager.config.ApplicationConfiguration().GetTopLevelConfig().OverwriteCache = true

		t.Run("request with interruptible override disabled", func(t *testing.T) {
			request := &admin.ExecutionCreateRequest{
				Project: workflowIdentifier.Project,
				Domain:  workflowIdentifier.Domain,
				Spec: &admin.ExecutionSpec{
					Interruptible: &wrappers.BoolValue{Value: false},
				},
			}

			execConfig, err := executionManager.getExecutionConfig(context.TODO(), request, nil)
			assert.NoError(t, err)
			assert.Equal(t, defaultMaxParallelism, execConfig.MaxParallelism)
			assert.False(t, execConfig.Interruptible.Value)
			assert.Equal(t, defaultK8sServiceAccount, execConfig.SecurityContext.RunAs.K8SServiceAccount)
			assert.Nil(t, execConfig.GetRawOutputDataConfig())
			assert.Nil(t, execConfig.GetLabels())
			assert.Nil(t, execConfig.GetAnnotations())
		})
		t.Run("request with interruptible override enabled", func(t *testing.T) {
			request := &admin.ExecutionCreateRequest{
				Project: workflowIdentifier.Project,
				Domain:  workflowIdentifier.Domain,
				Spec: &admin.ExecutionSpec{
					Interruptible: &wrappers.BoolValue{Value: true},
				},
			}

			execConfig, err := executionManager.getExecutionConfig(context.TODO(), request, nil)
			assert.NoError(t, err)
			assert.Equal(t, defaultMaxParallelism, execConfig.MaxParallelism)
			assert.True(t, execConfig.Interruptible.Value)
			assert.Equal(t, defaultK8sServiceAccount, execConfig.SecurityContext.RunAs.K8SServiceAccount)
			assert.Nil(t, execConfig.GetRawOutputDataConfig())
			assert.Nil(t, execConfig.GetLabels())
			assert.Nil(t, execConfig.GetAnnotations())
		})
		t.Run("request with no interruptible override specified", func(t *testing.T) {
			request := &admin.ExecutionCreateRequest{
				Project: workflowIdentifier.Project,
				Domain:  workflowIdentifier.Domain,
				Spec:    &admin.ExecutionSpec{},
			}

			execConfig, err := executionManager.getExecutionConfig(context.TODO(), request, nil)
			assert.NoError(t, err)
			assert.Equal(t, defaultMaxParallelism, execConfig.MaxParallelism)
			assert.True(t, execConfig.Interruptible.Value)
			assert.Equal(t, defaultK8sServiceAccount, execConfig.SecurityContext.RunAs.K8SServiceAccount)
			assert.Nil(t, execConfig.GetRawOutputDataConfig())
			assert.Nil(t, execConfig.GetLabels())
			assert.Nil(t, execConfig.GetAnnotations())
		})
		t.Run("launch plan with interruptible override disabled", func(t *testing.T) {
			request := &admin.ExecutionCreateRequest{
				Project: workflowIdentifier.Project,
				Domain:  workflowIdentifier.Domain,
				Spec:    &admin.ExecutionSpec{},
			}

			launchPlan := &admin.LaunchPlan{
				Spec: &admin.LaunchPlanSpec{
					Interruptible: &wrappers.BoolValue{Value: false},
				},
			}

			execConfig, err := executionManager.getExecutionConfig(context.TODO(), request, launchPlan)
			assert.NoError(t, err)
			assert.Equal(t, defaultMaxParallelism, execConfig.MaxParallelism)
			assert.False(t, execConfig.Interruptible.Value)
			assert.Equal(t, defaultK8sServiceAccount, execConfig.SecurityContext.RunAs.K8SServiceAccount)
			assert.Nil(t, execConfig.GetRawOutputDataConfig())
			assert.Nil(t, execConfig.GetLabels())
			assert.Nil(t, execConfig.GetAnnotations())
		})
		t.Run("launch plan with interruptible override enabled", func(t *testing.T) {
			request := &admin.ExecutionCreateRequest{
				Project: workflowIdentifier.Project,
				Domain:  workflowIdentifier.Domain,
				Spec:    &admin.ExecutionSpec{},
			}

			launchPlan := &admin.LaunchPlan{
				Spec: &admin.LaunchPlanSpec{
					Interruptible: &wrappers.BoolValue{Value: true},
					Envs:          &admin.Envs{Values: []*core.KeyValuePair{{Key: "foo", Value: "bar"}}},
				},
			}

			execConfig, err := executionManager.getExecutionConfig(context.TODO(), request, launchPlan)
			assert.NoError(t, err)
			assert.Equal(t, defaultMaxParallelism, execConfig.MaxParallelism)
			assert.True(t, execConfig.Interruptible.Value)
			assert.Equal(t, defaultK8sServiceAccount, execConfig.SecurityContext.RunAs.K8SServiceAccount)
			assert.Nil(t, execConfig.GetRawOutputDataConfig())
			assert.Nil(t, execConfig.GetLabels())
			assert.Nil(t, execConfig.GetAnnotations())
			assert.Equal(t, 1, len(execConfig.Envs.Values))
			assert.Equal(t, "foo", execConfig.Envs.Values[0].Key)
			assert.Equal(t, "bar", execConfig.Envs.Values[0].Value)
		})
		t.Run("launch plan with no interruptible override specified", func(t *testing.T) {
			request := &admin.ExecutionCreateRequest{
				Project: workflowIdentifier.Project,
				Domain:  workflowIdentifier.Domain,
				Spec:    &admin.ExecutionSpec{},
			}

			launchPlan := &admin.LaunchPlan{
				Spec: &admin.LaunchPlanSpec{},
			}

			execConfig, err := executionManager.getExecutionConfig(context.TODO(), request, launchPlan)
			assert.NoError(t, err)
			assert.Equal(t, defaultMaxParallelism, execConfig.MaxParallelism)
			assert.True(t, execConfig.Interruptible.Value)
			assert.Equal(t, defaultK8sServiceAccount, execConfig.SecurityContext.RunAs.K8SServiceAccount)
			assert.Nil(t, execConfig.GetRawOutputDataConfig())
			assert.Nil(t, execConfig.GetLabels())
			assert.Nil(t, execConfig.GetAnnotations())
		})
		t.Run("request and launch plan with different interruptible overrides", func(t *testing.T) {
			request := &admin.ExecutionCreateRequest{
				Project: workflowIdentifier.Project,
				Domain:  workflowIdentifier.Domain,
				Spec: &admin.ExecutionSpec{
					Interruptible: &wrappers.BoolValue{Value: true},
				},
			}

			launchPlan := &admin.LaunchPlan{
				Spec: &admin.LaunchPlanSpec{
					Interruptible: &wrappers.BoolValue{Value: false},
				},
			}

			execConfig, err := executionManager.getExecutionConfig(context.TODO(), request, launchPlan)
			assert.NoError(t, err)
			assert.Equal(t, defaultMaxParallelism, execConfig.MaxParallelism)
			assert.True(t, execConfig.Interruptible.Value)
			assert.Equal(t, defaultK8sServiceAccount, execConfig.SecurityContext.RunAs.K8SServiceAccount)
			assert.Nil(t, execConfig.GetRawOutputDataConfig())
			assert.Nil(t, execConfig.GetLabels())
			assert.Nil(t, execConfig.GetAnnotations())
		})
		t.Run("request with skip cache override enabled", func(t *testing.T) {
			request := &admin.ExecutionCreateRequest{
				Project: workflowIdentifier.Project,
				Domain:  workflowIdentifier.Domain,
				Spec: &admin.ExecutionSpec{
					OverwriteCache: true,
				},
			}

			execConfig, err := executionManager.getExecutionConfig(context.TODO(), request, nil)
			assert.NoError(t, err)
			assert.Equal(t, defaultMaxParallelism, execConfig.MaxParallelism)
			assert.True(t, execConfig.OverwriteCache)
			assert.Equal(t, defaultK8sServiceAccount, execConfig.SecurityContext.RunAs.K8SServiceAccount)
			assert.Nil(t, execConfig.GetRawOutputDataConfig())
			assert.Nil(t, execConfig.GetLabels())
			assert.Nil(t, execConfig.GetAnnotations())
		})
		t.Run("request with no skip cache override specified", func(t *testing.T) {
			request := &admin.ExecutionCreateRequest{
				Project: workflowIdentifier.Project,
				Domain:  workflowIdentifier.Domain,
				Spec:    &admin.ExecutionSpec{},
			}

			execConfig, err := executionManager.getExecutionConfig(context.TODO(), request, nil)
			assert.NoError(t, err)
			assert.Equal(t, defaultMaxParallelism, execConfig.MaxParallelism)
			assert.True(t, execConfig.OverwriteCache)
			assert.Equal(t, defaultK8sServiceAccount, execConfig.SecurityContext.RunAs.K8SServiceAccount)
			assert.Nil(t, execConfig.GetRawOutputDataConfig())
			assert.Nil(t, execConfig.GetLabels())
			assert.Nil(t, execConfig.GetAnnotations())
		})
		t.Run("launch plan with skip cache override enabled", func(t *testing.T) {
			request := &admin.ExecutionCreateRequest{
				Project: workflowIdentifier.Project,
				Domain:  workflowIdentifier.Domain,
				Spec:    &admin.ExecutionSpec{},
			}

			launchPlan := &admin.LaunchPlan{
				Spec: &admin.LaunchPlanSpec{
					OverwriteCache: true,
				},
			}

			execConfig, err := executionManager.getExecutionConfig(context.TODO(), request, launchPlan)
			assert.NoError(t, err)
			assert.Equal(t, defaultMaxParallelism, execConfig.MaxParallelism)
			assert.True(t, execConfig.OverwriteCache)
			assert.Equal(t, defaultK8sServiceAccount, execConfig.SecurityContext.RunAs.K8SServiceAccount)
			assert.Nil(t, execConfig.GetRawOutputDataConfig())
			assert.Nil(t, execConfig.GetLabels())
			assert.Nil(t, execConfig.GetAnnotations())
		})
		t.Run("launch plan with no skip cache override specified", func(t *testing.T) {
			request := &admin.ExecutionCreateRequest{
				Project: workflowIdentifier.Project,
				Domain:  workflowIdentifier.Domain,
				Spec:    &admin.ExecutionSpec{},
			}

			launchPlan := &admin.LaunchPlan{
				Spec: &admin.LaunchPlanSpec{},
			}

			execConfig, err := executionManager.getExecutionConfig(context.TODO(), request, launchPlan)
			assert.NoError(t, err)
			assert.Equal(t, defaultMaxParallelism, execConfig.MaxParallelism)
			assert.True(t, execConfig.OverwriteCache)
			assert.Equal(t, defaultK8sServiceAccount, execConfig.SecurityContext.RunAs.K8SServiceAccount)
			assert.Nil(t, execConfig.GetRawOutputDataConfig())
			assert.Nil(t, execConfig.GetLabels())
			assert.Nil(t, execConfig.GetAnnotations())
		})
		t.Run("request and launch plan with different skip cache overrides", func(t *testing.T) {
			request := &admin.ExecutionCreateRequest{
				Project: workflowIdentifier.Project,
				Domain:  workflowIdentifier.Domain,
				Spec: &admin.ExecutionSpec{
					OverwriteCache: true,
				},
			}

			launchPlan := &admin.LaunchPlan{
				Spec: &admin.LaunchPlanSpec{
					OverwriteCache: false,
				},
			}

			execConfig, err := executionManager.getExecutionConfig(context.TODO(), request, launchPlan)
			assert.NoError(t, err)
			assert.Equal(t, defaultMaxParallelism, execConfig.MaxParallelism)
			assert.True(t, execConfig.OverwriteCache)
			assert.Equal(t, defaultK8sServiceAccount, execConfig.SecurityContext.RunAs.K8SServiceAccount)
			assert.Nil(t, execConfig.GetRawOutputDataConfig())
			assert.Nil(t, execConfig.GetLabels())
			assert.Nil(t, execConfig.GetAnnotations())
		})

		t.Run("test pick up security context from admin system config", func(t *testing.T) {
			executionManager.config.ApplicationConfiguration().GetTopLevelConfig().K8SServiceAccount = "flyte-test"
			request := &admin.ExecutionCreateRequest{
				Project: workflowIdentifier.Project,
				Domain:  workflowIdentifier.Domain,
				Spec:    &admin.ExecutionSpec{},
			}
			execConfig, err := executionManager.getExecutionConfig(context.TODO(), request, nil)
			assert.NoError(t, err)
			assert.Equal(t, "flyte-test", execConfig.SecurityContext.RunAs.K8SServiceAccount)
			executionManager.config.ApplicationConfiguration().GetTopLevelConfig().K8SServiceAccount = defaultK8sServiceAccount
		})
	})
}

func TestGetExecutionConfig(t *testing.T) {
	resourceManager := managerMocks.MockResourceManager{}
	resourceManager.GetResourceFunc = func(ctx context.Context,
		request managerInterfaces.ResourceRequest) (*managerInterfaces.ResourceResponse, error) {
		assert.Contains(t, []managerInterfaces.ResourceRequest{{
			Project:      workflowIdentifier.Project,
			Domain:       workflowIdentifier.Domain,
			ResourceType: admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG,
		}, {Project: workflowIdentifier.Project,
			Domain:       "",
			ResourceType: admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG},
		}, request)
		return &managerInterfaces.ResourceResponse{
			Attributes: &admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_WorkflowExecutionConfig{
					WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{
						MaxParallelism: 100,
						OverwriteCache: true,
					},
				},
			},
		}, nil
	}

	applicationConfig := runtime.NewConfigurationProvider()
	executionManager := ExecutionManager{
		resourceManager: &resourceManager,
		config:          applicationConfig,
	}
	execConfig, err := executionManager.getExecutionConfig(context.TODO(), &admin.ExecutionCreateRequest{
		Project: workflowIdentifier.Project,
		Domain:  workflowIdentifier.Domain,
		Spec:    &admin.ExecutionSpec{},
	}, nil)
	assert.NoError(t, err)
	assert.Equal(t, execConfig.MaxParallelism, int32(100))
	assert.True(t, execConfig.OverwriteCache)
}

func TestGetExecutionConfig_Spec(t *testing.T) {
	resourceManager := managerMocks.MockResourceManager{}
	resourceManager.GetResourceFunc = func(ctx context.Context,
		request managerInterfaces.ResourceRequest) (*managerInterfaces.ResourceResponse, error) {
		return nil, nil
	}
	applicationConfig := runtime.NewConfigurationProvider()
	executionManager := ExecutionManager{
		resourceManager: &resourceManager,
		config:          applicationConfig,
	}
	execConfig, err := executionManager.getExecutionConfig(context.TODO(), &admin.ExecutionCreateRequest{
		Project: workflowIdentifier.Project,
		Domain:  workflowIdentifier.Domain,
		Spec: &admin.ExecutionSpec{
			MaxParallelism: 100,
			OverwriteCache: true,
		},
	}, &admin.LaunchPlan{
		Spec: &admin.LaunchPlanSpec{
			MaxParallelism: 50,
			OverwriteCache: false, // explicitly set to false for clarity
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, int32(100), execConfig.MaxParallelism)
	assert.True(t, execConfig.OverwriteCache)

	execConfig, err = executionManager.getExecutionConfig(context.TODO(), &admin.ExecutionCreateRequest{
		Project: workflowIdentifier.Project,
		Domain:  workflowIdentifier.Domain,
		Spec:    &admin.ExecutionSpec{},
	}, &admin.LaunchPlan{
		Spec: &admin.LaunchPlanSpec{
			MaxParallelism: 50,
			OverwriteCache: true,
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, int32(50), execConfig.MaxParallelism)
	assert.True(t, execConfig.OverwriteCache)

	resourceManager = managerMocks.MockResourceManager{}
	resourceManager.GetResourceFunc = func(ctx context.Context,
		request managerInterfaces.ResourceRequest) (*managerInterfaces.ResourceResponse, error) {
		return nil, nil
	}
	executionManager = ExecutionManager{
		resourceManager: &resourceManager,
		config:          applicationConfig,
	}

	executionManager.config.ApplicationConfiguration().GetTopLevelConfig().OverwriteCache = true

	execConfig, err = executionManager.getExecutionConfig(context.TODO(), &admin.ExecutionCreateRequest{
		Project: workflowIdentifier.Project,
		Domain:  workflowIdentifier.Domain,
		Spec:    &admin.ExecutionSpec{},
	}, &admin.LaunchPlan{
		Spec: &admin.LaunchPlanSpec{},
	})
	assert.NoError(t, err)
	assert.Equal(t, execConfig.MaxParallelism, int32(25))
	assert.True(t, execConfig.OverwriteCache)
}

func TestGetClusterAssignment(t *testing.T) {
	clusterAssignment := admin.ClusterAssignment{ClusterPoolName: "gpu"}
	resourceManager := managerMocks.MockResourceManager{}
	resourceManager.GetResourceFunc = func(ctx context.Context,
		request managerInterfaces.ResourceRequest) (*managerInterfaces.ResourceResponse, error) {
		assert.EqualValues(t, request, managerInterfaces.ResourceRequest{
			Project:      workflowIdentifier.Project,
			Domain:       workflowIdentifier.Domain,
			ResourceType: admin.MatchableResource_CLUSTER_ASSIGNMENT,
		})
		return &managerInterfaces.ResourceResponse{
			Attributes: &admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_ClusterAssignment{
					ClusterAssignment: &clusterAssignment,
				},
			},
		}, nil
	}

	executionManager := ExecutionManager{
		resourceManager: &resourceManager,
	}
	t.Run("value from db", func(t *testing.T) {
		ca, err := executionManager.getClusterAssignment(context.TODO(), &admin.ExecutionCreateRequest{
			Project: workflowIdentifier.Project,
			Domain:  workflowIdentifier.Domain,
			Spec:    &admin.ExecutionSpec{},
		})
		assert.NoError(t, err)
		assert.True(t, proto.Equal(ca, &clusterAssignment))
	})
	t.Run("value from request", func(t *testing.T) {
		reqClusterAssignment := admin.ClusterAssignment{ClusterPoolName: "swimming-pool"}
		ca, err := executionManager.getClusterAssignment(context.TODO(), &admin.ExecutionCreateRequest{
			Project: workflowIdentifier.Project,
			Domain:  workflowIdentifier.Domain,
			Spec: &admin.ExecutionSpec{
				ClusterAssignment: &reqClusterAssignment,
			},
		})
		assert.NoError(t, err)
		assert.True(t, proto.Equal(ca, &reqClusterAssignment))
	})
	t.Run("value from config", func(t *testing.T) {
		customCP := "my_cp"
		clusterPoolAsstProvider := &runtimeIFaceMocks.ClusterPoolAssignmentConfiguration{}
		clusterPoolAsstProvider.OnGetClusterPoolAssignments().Return(runtimeInterfaces.ClusterPoolAssignments{
			workflowIdentifier.GetDomain(): runtimeInterfaces.ClusterPoolAssignment{
				Pool: customCP,
			},
		})
		mockConfig := getMockExecutionsConfigProvider()
		mockConfig.(*runtimeMocks.MockConfigurationProvider).AddClusterPoolAssignmentConfiguration(clusterPoolAsstProvider)

		executionManager := ExecutionManager{
			resourceManager: &managerMocks.MockResourceManager{},
			config:          mockConfig,
		}

		ca, err := executionManager.getClusterAssignment(context.TODO(), &admin.ExecutionCreateRequest{
			Project: workflowIdentifier.Project,
			Domain:  workflowIdentifier.Domain,
			Spec:    &admin.ExecutionSpec{},
		})
		assert.NoError(t, err)
		assert.Equal(t, customCP, ca.GetClusterPoolName())
	})
}

func TestResolvePermissions(t *testing.T) {
	assumableIamRole := "role"
	k8sServiceAccount := "sa"

	assumableIamRoleLp := "roleLp"
	k8sServiceAccountLp := "saLp"

	assumableIamRoleSc := "roleSc"
	k8sServiceAccountSc := "saSc"

	t.Run("backward compat use request values from auth", func(t *testing.T) {
		execRequest := &admin.ExecutionCreateRequest{
			Spec: &admin.ExecutionSpec{
				AuthRole: &admin.AuthRole{
					AssumableIamRole:         assumableIamRole,
					KubernetesServiceAccount: k8sServiceAccount,
				},
			},
		}
		lp := &admin.LaunchPlan{
			Spec: &admin.LaunchPlanSpec{
				AuthRole: &admin.AuthRole{
					AssumableIamRole:         "lp role",
					KubernetesServiceAccount: "k8s sa",
				},
			},
		}
		execConfigSecCtx := &core.SecurityContext{
			RunAs: &core.Identity{
				IamRole:           assumableIamRole,
				K8SServiceAccount: k8sServiceAccount,
			},
		}
		authRole := resolveAuthRole(execRequest, lp)
		sc := resolveSecurityCtx(context.TODO(), execConfigSecCtx, authRole)
		assert.Equal(t, assumableIamRole, authRole.AssumableIamRole)
		assert.Equal(t, k8sServiceAccount, authRole.KubernetesServiceAccount)
		assert.Equal(t, &core.SecurityContext{
			RunAs: &core.Identity{
				IamRole:           assumableIamRole,
				K8SServiceAccount: k8sServiceAccount,
			}}, sc)
	})
	t.Run("use request values security context", func(t *testing.T) {
		execRequest := &admin.ExecutionCreateRequest{
			Spec: &admin.ExecutionSpec{
				SecurityContext: &core.SecurityContext{
					RunAs: &core.Identity{
						IamRole:           assumableIamRoleSc,
						K8SServiceAccount: k8sServiceAccountSc,
					},
				},
			},
		}
		lp := &admin.LaunchPlan{
			Spec: &admin.LaunchPlanSpec{
				SecurityContext: &core.SecurityContext{
					RunAs: &core.Identity{
						IamRole:           assumableIamRoleSc,
						K8SServiceAccount: k8sServiceAccountSc,
					},
				},
			},
		}
		authRole := resolveAuthRole(execRequest, lp)
		execConfigSecCtx := &core.SecurityContext{
			RunAs: &core.Identity{
				IamRole:           assumableIamRoleSc,
				K8SServiceAccount: k8sServiceAccountSc,
			},
		}
		sc := resolveSecurityCtx(context.TODO(), execConfigSecCtx, authRole)
		assert.Equal(t, "", authRole.AssumableIamRole)
		assert.Equal(t, "", authRole.KubernetesServiceAccount)
		assert.Equal(t, assumableIamRoleSc, sc.RunAs.IamRole)
		assert.Equal(t, k8sServiceAccountSc, sc.RunAs.K8SServiceAccount)
	})
	t.Run("prefer lp auth role over auth", func(t *testing.T) {
		execRequest := &admin.ExecutionCreateRequest{
			Spec: &admin.ExecutionSpec{},
		}
		lp := &admin.LaunchPlan{
			Spec: &admin.LaunchPlanSpec{
				AuthRole: &admin.AuthRole{
					AssumableIamRole:         assumableIamRole,
					KubernetesServiceAccount: k8sServiceAccount,
				},
				Auth: &admin.Auth{
					AssumableIamRole:         "lp role",
					KubernetesServiceAccount: "k8s sa",
				},
			},
		}
		authRole := resolveAuthRole(execRequest, lp)
		execConfigSecCtx := &core.SecurityContext{
			RunAs: &core.Identity{},
		}
		sc := resolveSecurityCtx(context.TODO(), execConfigSecCtx, authRole)
		assert.Equal(t, assumableIamRole, authRole.AssumableIamRole)
		assert.Equal(t, k8sServiceAccount, authRole.KubernetesServiceAccount)
		assert.Equal(t, &core.SecurityContext{
			RunAs: &core.Identity{
				IamRole:           assumableIamRole,
				K8SServiceAccount: k8sServiceAccount,
			},
		}, sc)
	})
	t.Run("prefer security context over auth context", func(t *testing.T) {
		execRequest := &admin.ExecutionCreateRequest{
			Spec: &admin.ExecutionSpec{
				AuthRole: &admin.AuthRole{
					AssumableIamRole:         assumableIamRole,
					KubernetesServiceAccount: k8sServiceAccount,
				},
				SecurityContext: &core.SecurityContext{
					RunAs: &core.Identity{
						IamRole:           assumableIamRoleSc,
						K8SServiceAccount: k8sServiceAccountSc,
					},
				},
			},
		}
		lp := &admin.LaunchPlan{
			Spec: &admin.LaunchPlanSpec{
				AuthRole: &admin.AuthRole{
					AssumableIamRole:         assumableIamRole,
					KubernetesServiceAccount: k8sServiceAccount,
				},
				SecurityContext: &core.SecurityContext{
					RunAs: &core.Identity{
						IamRole:           assumableIamRoleSc,
						K8SServiceAccount: k8sServiceAccountSc,
					},
				},
			},
		}
		authRole := resolveAuthRole(execRequest, lp)
		execConfigSecCtx := &core.SecurityContext{
			RunAs: &core.Identity{
				IamRole:           assumableIamRoleSc,
				K8SServiceAccount: k8sServiceAccountSc,
			},
		}
		sc := resolveSecurityCtx(context.TODO(), execConfigSecCtx, authRole)
		assert.Equal(t, assumableIamRole, authRole.AssumableIamRole)
		assert.Equal(t, k8sServiceAccount, authRole.KubernetesServiceAccount)
		assert.Equal(t, assumableIamRoleSc, sc.RunAs.IamRole)
		assert.Equal(t, k8sServiceAccountSc, sc.RunAs.K8SServiceAccount)
	})
	t.Run("prefer lp auth over role", func(t *testing.T) {
		execRequest := &admin.ExecutionCreateRequest{
			Spec: &admin.ExecutionSpec{},
		}
		lp := &admin.LaunchPlan{
			Spec: &admin.LaunchPlanSpec{
				Auth: &admin.Auth{
					AssumableIamRole:         assumableIamRole,
					KubernetesServiceAccount: k8sServiceAccount,
				},
				Role: "old role",
			},
		}
		authRole := resolveAuthRole(execRequest, lp)
		execConfigSecCtx := &core.SecurityContext{
			RunAs: &core.Identity{
				IamRole:           assumableIamRole,
				K8SServiceAccount: k8sServiceAccount,
			},
		}
		sc := resolveSecurityCtx(context.TODO(), execConfigSecCtx, authRole)
		assert.Equal(t, assumableIamRole, authRole.AssumableIamRole)
		assert.Equal(t, k8sServiceAccount, authRole.KubernetesServiceAccount)
		assert.Equal(t, &core.SecurityContext{
			RunAs: &core.Identity{
				IamRole:           assumableIamRole,
				K8SServiceAccount: k8sServiceAccount,
			},
		}, sc)
	})
	t.Run("prefer lp auth over role", func(t *testing.T) {
		authRole := resolveAuthRole(&admin.ExecutionCreateRequest{
			Spec: &admin.ExecutionSpec{},
		}, &admin.LaunchPlan{
			Spec: &admin.LaunchPlanSpec{
				Auth: &admin.Auth{
					AssumableIamRole:         assumableIamRoleLp,
					KubernetesServiceAccount: k8sServiceAccountLp,
				},
				Role: "old role",
			},
		})
		assert.Equal(t, assumableIamRoleLp, authRole.AssumableIamRole)
		assert.Equal(t, k8sServiceAccountLp, authRole.KubernetesServiceAccount)
	})
}

func TestAddStateFilter(t *testing.T) {
	t.Run("empty filters", func(t *testing.T) {
		var filters []common.InlineFilter
		updatedFilters, err := addStateFilter(filters)
		assert.Nil(t, err)
		assert.NotNil(t, updatedFilters)
		assert.Equal(t, 1, len(updatedFilters))

		assert.Equal(t, shared.State, updatedFilters[0].GetField())
		assert.Equal(t, common.Execution, updatedFilters[0].GetEntity())

		expression, err := updatedFilters[0].GetGormQueryExpr()
		assert.NoError(t, err)
		assert.Equal(t, "state = ?", expression.Query)
	})

	t.Run("passed state filter", func(t *testing.T) {
		filter, err := common.NewSingleValueFilter(common.Execution, common.NotEqual, "state", "0")
		assert.NoError(t, err)
		filters := []common.InlineFilter{filter}

		updatedFilters, err := addStateFilter(filters)
		assert.Nil(t, err)
		assert.NotNil(t, updatedFilters)
		assert.Equal(t, 1, len(updatedFilters))

		assert.Equal(t, shared.State, updatedFilters[0].GetField())
		assert.Equal(t, common.Execution, updatedFilters[0].GetEntity())

		expression, err := updatedFilters[0].GetGormQueryExpr()
		assert.NoError(t, err)
		assert.Equal(t, "state <> ?", expression.Query)
	})

}
