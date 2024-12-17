package impl

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/golang/protobuf/jsonpb"
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

	"github.com/flyteorg/flyte/flyteadmin/auth"
	"github.com/flyteorg/flyte/flyteadmin/pkg/artifacts"
	artifactMocks "github.com/flyteorg/flyte/flyteadmin/pkg/artifacts/mocks"
	eventWriterMocks "github.com/flyteorg/flyte/flyteadmin/pkg/async/events/mocks"
	notificationMocks "github.com/flyteorg/flyte/flyteadmin/pkg/async/notifications/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/clusterresource/plugin"
	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	commonMocks "github.com/flyteorg/flyte/flyteadmin/pkg/common/mocks"
	commonTestUtils "github.com/flyteorg/flyte/flyteadmin/pkg/common/testutils"
	dataMocks "github.com/flyteorg/flyte/flyteadmin/pkg/data/mocks"
	flyteAdminErrors "github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/executions"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/testutils"
	managerInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"
	managerMocks "github.com/flyteorg/flyte/flyteadmin/pkg/manager/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	repositoryMocks "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/transformers"
	"github.com/flyteorg/flyte/flyteadmin/pkg/runtime"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	runtimeIFaceMocks "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces/mocks"
	runtimeMocks "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/mocks"
	workflowengineInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/workflowengine/interfaces"
	workflowengineMocks "github.com/flyteorg/flyte/flyteadmin/pkg/workflowengine/mocks"
	"github.com/flyteorg/flyte/flyteadmin/plugins"
	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	artifactsIdl "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifacts"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	mockScope "github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

const (
	principal                   = "principal"
	rawOutput                   = "raw_output"
	executionClusterLabel       = "execution_cluster_label"
	executionCreatedAtFilter    = "gte(execution_created_at,2021-01-01T00:00:00Z)"
	executionCreatedAtValue     = "2021-01-01T00:00:00Z"
	executionCreatedAtQueryExpr = "execution_created_at >= ?"
	executionDomainQueryExpr    = "execution_domain = ?"
	executionNameQueryExpr      = "execution_name = ?"
	executionOrgQueryExpr       = "execution_org = ?"
	executionProjectQueryExpr   = "execution_project = ?"
	executionModeNeQueryExpr    = "mode <> ?"
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
	ResolvedSpec: getExpectedSpec(),
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
var outputURI = "s3://bucket/output_uri"
var inputURI = "s3://bucket/inputs"
var resourceDefaults = runtimeInterfaces.TaskResourceSet{
	CPU:    resource.MustParse("200m"),
	Memory: resource.MustParse("200Gi"),
}
var resourceLimits = runtimeInterfaces.TaskResourceSet{
	CPU:    resource.MustParse("300m"),
	Memory: resource.MustParse("500Gi"),
}

var noopSelfServeServicePlugin = plugin.NewNoopClusterResourcePlugin()
var noopPreExecutionValidationPlugin = executions.NewNoopPreExecutionValidationPlugin()

func getDefaultPluginRegistry() *plugins.Registry {
	r := plugins.NewRegistry()
	r.RegisterDefault(plugins.PluginIDUserProperties, shared.DefaultGetUserPropertiesFunc)
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &defaultTestExecutor)
	r.RegisterDefault(plugins.PluginIDClusterResource, noopSelfServeServicePlugin)
	r.RegisterDefault(plugins.PluginIDPreExecutionValidation, noopPreExecutionValidationPlugin)
	return r
}

func getNoopMockResourceManager() managerInterfaces.ResourceInterface {
	mockResourceManager := new(managerMocks.ResourceInterface)
	mockResourceManager.On("GetResource", mock.Anything, mock.Anything).Return(nil, nil)
	return mockResourceManager
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
		ResolvedSpec: getExpectedSpec(),
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
	return r
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

	lpSpecBytes, _ := proto.Marshal(lpSpec)
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
	var mtx sync.RWMutex
	mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).ReadProtobufCb = func(
		ctx context.Context, reference storage.DataReference, msg proto.Message) error {
		mtx.RLock()
		defer mtx.RUnlock()
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
		mtx.Lock()
		defer mtx.Unlock()
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
		return transformers.CreateProjectModel(&admin.Project{Labels: &labels}), nil
	}

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
			assert.Equal(t, len(resources.Requests), len(task.Template.GetContainer().Resources.Requests))
			for i, request := range resources.Requests {
				assert.True(t, proto.Equal(request, task.Template.GetContainer().Resources.Requests[i]))
				assert.True(t, proto.Equal(request, task.Template.GetContainer().Resources.Limits[i]))
			}
		}

		return true
	})).Return(workflowengineInterfaces.ExecutionResponse{
		Cluster: testCluster,
	}, nil)
	mockExecutor.OnID().Return("customMockExecutor")
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)

	qosProvider := &runtimeIFaceMocks.QualityOfServiceConfiguration{}
	qosProvider.OnGetTierExecutionValues().Return(map[core.QualityOfService_Tier]*core.QualityOfServiceSpec{
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
	request := testutils.GetExecutionRequest()
	request.Spec.Metadata = &admin.ExecutionMetadata{
		Principal: "unused - populated from authenticated context",
	}
	request.Spec.RawOutputDataConfig = &admin.RawOutputDataConfig{OutputLocationPrefix: rawOutput}
	request.Spec.ClusterAssignment = &clusterAssignment
	request.Spec.ExecutionClusterLabel = &admin.ExecutionClusterLabel{Value: executionClusterLabel}
	mockResourceManager := new(managerMocks.ResourceInterface)
	defer mockResourceManager.AssertExpectations(t)
	mockResourceManager.
		OnGetResourceMatch(mock.Anything, managerInterfaces.ResourceRequest{
			Org:          request.Org,
			Project:      request.Project,
			Domain:       request.Domain,
			Workflow:     name,
			ResourceType: admin.MatchableResource_TASK_RESOURCE,
		}).
		Return(&managerInterfaces.ResourceResponse{}, nil).
		Once()
	mockResourceManager.
		OnGetResourceMatch(mock.Anything, managerInterfaces.ResourceRequest{
			Org:          request.Org,
			Project:      request.Project,
			Domain:       request.Domain,
			Workflow:     name,
			ResourceType: admin.MatchableResource_EXECUTION_QUEUE,
		}).
		Return(&managerInterfaces.ResourceResponse{}, nil).
		Once()
	mockResourceManager.
		OnGetResourceMatch(mock.Anything, managerInterfaces.ResourceRequest{
			Org:          request.Org,
			Project:      request.Project,
			Domain:       request.Domain,
			Workflow:     name,
			ResourceType: admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG,
		}).
		Return(&managerInterfaces.ResourceResponse{}, nil).
		Once()
	mockResourceManager.
		OnGetResourceMatch(mock.Anything, managerInterfaces.ResourceRequest{
			Org:          request.Org,
			Project:      request.Project,
			ResourceType: admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG,
		}).
		Return(&managerInterfaces.ResourceResponse{}, nil).
		Once()
	mockResourceManager.
		OnGetResourceMatch(mock.Anything, managerInterfaces.ResourceRequest{
			Org:          request.Org,
			Project:      request.Project,
			Domain:       request.Domain,
			ResourceType: admin.MatchableResource_CLUSTER_ASSIGNMENT,
		}).
		Return(&managerInterfaces.ResourceResponse{
			Attributes: &admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_ClusterAssignment{
					ClusterAssignment: &admin.ClusterAssignment{ClusterPoolName: "gpu"},
				},
			},
		}, nil).
		Once()
	mockResourceManager.
		OnGetResourceMatch(mock.Anything, managerInterfaces.ResourceRequest{
			Org:          request.Org,
			Project:      request.Project,
			ResourceType: admin.MatchableResource_EXTERNAL_RESOURCE,
		}).
		Return(&managerInterfaces.ResourceResponse{}, nil).
		Once()
	mockResourceManager.
		OnGetResourceMatch(mock.Anything, managerInterfaces.ResourceRequest{
			Org:          request.Org,
			Project:      request.Project,
			Domain:       request.Domain,
			ResourceType: admin.MatchableResource_EXTERNAL_RESOURCE,
		}).
		Return(&managerInterfaces.ResourceResponse{}, nil).
		Once()
	mockResourceManager.
		OnGetResourceMatch(mock.Anything, managerInterfaces.ResourceRequest{
			Org:          request.Org,
			Project:      request.Project,
			Domain:       request.Domain,
			Workflow:     name,
			LaunchPlan:   name,
			ResourceType: admin.MatchableResource_PLUGIN_OVERRIDE,
		}).
		Return(&managerInterfaces.ResourceResponse{}, nil).
		Once()
	execManager := NewExecutionManager(repository, r, mockConfig, getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, &mockPublisher, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

	identity, err := auth.NewIdentityContext("", principal, "", time.Now(), sets.NewString(), nil, nil)
	assert.NoError(t, err)
	ctx := identity.WithContext(context.Background())
	response, err := execManager.CreateExecution(ctx, request, requestedAt)
	assert.NoError(t, err)

	expectedResponse := &admin.ExecutionCreateResponse{
		Id: &executionIdentifier,
	}
	assert.NoError(t, err)
	assert.True(t, proto.Equal(expectedResponse.Id, response.Id))

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
			assert.EqualValues(t, input.NodeExecutionIdentifier, &parentNodeExecutionID)
			getNodeExecutionCalled = true
			return models.NodeExecution{
				BaseModel: models.BaseModel{
					ID: 1,
				},
			}, nil
		},
	)

	getExecutionCalled := false
	var clusterLabel = &admin.ExecutionClusterLabel{Value: executionClusterLabel}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
			assert.EqualValues(t, input.Project, parentNodeExecutionID.ExecutionId.Project)
			assert.EqualValues(t, input.Domain, parentNodeExecutionID.ExecutionId.Domain)
			assert.EqualValues(t, input.Name, parentNodeExecutionID.ExecutionId.Name)
			spec := &admin.ExecutionSpec{
				Metadata: &admin.ExecutionMetadata{
					Nesting: 1,
				},
				ExecutionClusterLabel: clusterLabel,
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
			assert.Equal(t, executionClusterLabel, spec.ExecutionClusterLabel.Value)
			assert.Equal(t, principal, input.User)
			return nil
		},
	)

	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{
		Cluster: testCluster,
	}, nil)
	mockExecutor.OnID().Return("customMockExecutor")
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)

	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
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
	assert.True(t, proto.Equal(expectedResponse, response))
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
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)

	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
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
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, configProvider, getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

	request := testutils.GetExecutionRequest()
	response, err := execManager.CreateExecution(context.Background(), request, requestedAt)
	assert.Nil(t, err)

	expectedResponse := &admin.ExecutionCreateResponse{
		Id: &executionIdentifier,
	}
	assert.Nil(t, err)
	assert.True(t, proto.Equal(expectedResponse, response))
}

func TestCreateExecutionValidationError(t *testing.T) {
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

	request := testutils.GetExecutionRequest()
	request.Domain = ""
	response, err := execManager.CreateExecution(context.Background(), request, requestedAt)
	assert.EqualError(t, err, "missing domain")
	assert.Nil(t, response)
}

func TestCreateExecution_InvalidLpIdentifier(t *testing.T) {
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

	request := testutils.GetExecutionRequest()
	request.Spec.LaunchPlan = nil
	response, err := execManager.CreateExecution(context.Background(), request, requestedAt)
	assert.EqualError(t, err, "missing id")
	assert.Nil(t, response)
}

func TestCreateExecutionInCompatibleInputs(t *testing.T) {
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

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
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	qosProvider := &runtimeIFaceMocks.QualityOfServiceConfiguration{}
	qosProvider.OnGetTierExecutionValues().Return(map[core.QualityOfService_Tier]*core.QualityOfServiceSpec{
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

	request := testutils.GetExecutionRequest()
	request.Spec.Metadata = &admin.ExecutionMetadata{
		Principal: "unused - populated from authenticated context",
	}
	request.Spec.RawOutputDataConfig = &admin.RawOutputDataConfig{OutputLocationPrefix: rawOutput}

	identity, err := auth.NewIdentityContext("", principal, "", time.Now(), sets.NewString(), nil, nil)
	assert.NoError(t, err)
	ctx := identity.WithContext(context.Background())
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

	expectedResponse := &admin.ExecutionCreateResponse{Id: &executionIdentifier}

	response, err := execManager.CreateExecution(ctx, request, requestedAt)

	assert.NoError(t, err)
	assert.True(t, proto.Equal(expectedResponse, response))
}

func TestCreateExecutionDatabaseFailure(t *testing.T) {
	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{}, nil)
	mockExecutor.OnID().Return("customMockExecutor")
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	expectedErr := flyteAdminErrors.NewFlyteAdminErrorf(codes.Internal, "ABCD")
	exCreateFunc := func(ctx context.Context, input models.Execution) error {
		return expectedErr
	}

	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCreateCallback(exCreateFunc)
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
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
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), storageClient, mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

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
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

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
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

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
	lpSpecBytes, _ := proto.Marshal(lpSpec)
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
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

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
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
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
	assert.True(t, proto.Equal(expectedResponse, response))
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
			r := getDefaultPluginRegistry()
			r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
			mockResourceManager := getNoopMockResourceManager()
			execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

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
			r := getDefaultPluginRegistry()
			r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
			mockResourceManager := getNoopMockResourceManager()
			execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

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
			r := getDefaultPluginRegistry()
			r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
			mockResourceManager := getNoopMockResourceManager()
			execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

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
	r := getDefaultPluginRegistry()
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

	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, mockExecutionsConfigProvider, storageClient, mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

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
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
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
	response, err := execManager.RelaunchExecution(context.Background(), &admin.ExecutionRelaunchRequest{
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

	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

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
	_, err := execManager.RelaunchExecution(context.Background(), &admin.ExecutionRelaunchRequest{
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
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
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
	_, err := execManager.RelaunchExecution(context.Background(), &admin.ExecutionRelaunchRequest{
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
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
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

	_, err := execManager.RelaunchExecution(context.Background(), &admin.ExecutionRelaunchRequest{
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
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
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

		asd, err := execManager.RelaunchExecution(context.Background(), &admin.ExecutionRelaunchRequest{
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

		asd, err := execManager.RelaunchExecution(context.Background(), &admin.ExecutionRelaunchRequest{
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

		asd, err := execManager.RelaunchExecution(context.Background(), &admin.ExecutionRelaunchRequest{
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
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
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

	_, err := execManager.RelaunchExecution(context.Background(), &admin.ExecutionRelaunchRequest{
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
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
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
	response, err := execManager.RecoverExecution(context.Background(), &admin.ExecutionRecoverRequest{
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
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
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
		assert.True(t, proto.Equal(&parentNodeExecution, input.NodeExecutionIdentifier))

		return models.NodeExecution{
			BaseModel: models.BaseModel{
				ID: parentNodeDatabaseID,
			},
		}, nil
	})

	// Issue request.
	response, err := execManager.RecoverExecution(context.Background(), &admin.ExecutionRecoverRequest{
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
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

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
	_, err := execManager.RecoverExecution(context.Background(), &admin.ExecutionRecoverRequest{
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
	mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).WriteProtobufCb = func(
		ctx context.Context, reference storage.DataReference, opts storage.Options, msg proto.Message) error {
		return nil
	}
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), mockStorage, mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
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
	_, err := execManager.RecoverExecution(context.Background(), &admin.ExecutionRecoverRequest{
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
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
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
	response, err := execManager.RecoverExecution(context.Background(), &admin.ExecutionRecoverRequest{
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
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
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
	response, err := execManager.RecoverExecution(context.Background(), &admin.ExecutionRecoverRequest{
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
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
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
	response, err := execManager.RecoverExecution(context.Background(), &admin.ExecutionRecoverRequest{
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
	request := &admin.WorkflowExecutionEventRequest{
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
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, &mockPublisher, &mockPublisher, mockDbEventWriter, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
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
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

	resp, err := execManager.CreateWorkflowEvent(context.Background(), &admin.WorkflowExecutionEventRequest{
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
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

	resp, err := execManager.CreateWorkflowEvent(context.Background(), &admin.WorkflowExecutionEventRequest{
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

	req := &admin.WorkflowExecutionEventRequest{
		RequestId: "1",
		Event: &event.WorkflowExecutionEvent{
			ExecutionId: &executionIdentifier,
			Phase:       core.WorkflowExecution_ABORTED,
			OccurredAt:  timestamppb.New(time.Now()),
		},
	}

	mockDbEventWriter := &eventWriterMocks.WorkflowExecutionEventWriter{}
	mockDbEventWriter.On("Write", req)
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, &mockPublisher, &mockPublisher, mockDbEventWriter, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

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

	closure := &admin.ExecutionClosure{
		Phase:     core.WorkflowExecution_RUNNING,
		StartedAt: occurredAtProto,
		UpdatedAt: occurredAtProto,
		StateChangeDetails: &admin.ExecutionStateChangeDetails{
			State:      admin.ExecutionState_EXECUTION_ACTIVE,
			OccurredAt: testutils.MockCreatedAtProto,
		},
		ResolvedSpec: getExpectedSpec(),
	}
	closureBytes, _ := proto.Marshal(closure)
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
	request := &admin.WorkflowExecutionEventRequest{
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
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, &mockPublisher, &mockPublisher, mockDbEventWriter, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
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

	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
	occurredAtTimestamp, _ := ptypes.TimestampProto(occurredAt)
	resp, err := execManager.CreateWorkflowEvent(context.Background(), &admin.WorkflowExecutionEventRequest{
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

	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
	occurredAtTimestamp, _ := ptypes.TimestampProto(occurredAt)
	resp, err := execManager.CreateWorkflowEvent(context.Background(), &admin.WorkflowExecutionEventRequest{
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
	request := &admin.WorkflowExecutionEventRequest{
		RequestId: "1",
		Event: &event.WorkflowExecutionEvent{
			ExecutionId: &executionIdentifier,
			OccurredAt:  occurredAtTimestamp,
			Phase:       core.WorkflowExecution_QUEUED,
			ProducerId:  newCluster,
		},
	}
	mockDbEventWriter.On("Write", request)
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, &mockPublisher, &mockPublisher, mockDbEventWriter, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

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
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
	resp, err := execManager.CreateWorkflowEvent(context.Background(), &admin.WorkflowExecutionEventRequest{
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

	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
	resp, err := execManager.CreateWorkflowEvent(context.Background(), &admin.WorkflowExecutionEventRequest{
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
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
	resp, err := execManager.CreateWorkflowEvent(context.Background(), &admin.WorkflowExecutionEventRequest{
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
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
	resp, err := execManager.CreateWorkflowEvent(context.Background(), &admin.WorkflowExecutionEventRequest{
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

	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
	occurredAtTimestamp, _ := ptypes.TimestampProto(occurredAt)
	resp, err := execManager.CreateWorkflowEvent(context.Background(), &admin.WorkflowExecutionEventRequest{
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
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
	execution, err := execManager.GetExecution(context.Background(), &admin.WorkflowExecutionGetRequest{
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
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
	execution, err := execManager.GetExecution(context.Background(), &admin.WorkflowExecutionGetRequest{
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
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
	execution, err := execManager.GetExecution(context.Background(), &admin.WorkflowExecutionGetRequest{
		Id: &executionIdentifier,
	})
	assert.Nil(t, execution)
	assert.Equal(t, codes.Internal, err.(flyteAdminErrors.FlyteAdminError).Code())
}

func TestUpdateExecution(t *testing.T) {
	t.Run("invalid execution identifier", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		r := getDefaultPluginRegistry()
		mockResourceManager := getNoopMockResourceManager()
		execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
		_, err := execManager.UpdateExecution(context.Background(), &admin.ExecutionUpdateRequest{
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
		r := getDefaultPluginRegistry()
		mockResourceManager := getNoopMockResourceManager()
		execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
		updateResponse, err := execManager.UpdateExecution(context.Background(), &admin.ExecutionUpdateRequest{
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
		r := getDefaultPluginRegistry()
		mockResourceManager := getNoopMockResourceManager()
		execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
		updateResponse, err := execManager.UpdateExecution(context.Background(), &admin.ExecutionUpdateRequest{
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
		r := getDefaultPluginRegistry()
		mockResourceManager := getNoopMockResourceManager()
		execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
		_, err := execManager.UpdateExecution(context.Background(), &admin.ExecutionUpdateRequest{
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
		r := getDefaultPluginRegistry()
		mockResourceManager := getNoopMockResourceManager()
		execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
		_, err := execManager.UpdateExecution(context.Background(), &admin.ExecutionUpdateRequest{
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
			if queryExpr.Args == projectValue && queryExpr.Query == executionProjectQueryExpr {
				projectFilter = true
			}
			if queryExpr.Args == domainValue && queryExpr.Query == executionDomainQueryExpr {
				domainFilter = true
			}
			if queryExpr.Args == nameValue && queryExpr.Query == executionNameQueryExpr {
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
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

	executionList, err := execManager.ListExecutions(context.Background(), &admin.ResourceListRequest{
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

func TestListExecutions_Filters(t *testing.T) {
	t.Run("with mode filter added", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		sawModeFilter := false
		executionListFunc := func(
			ctx context.Context, input interfaces.ListResourceInput) (interfaces.ExecutionCollectionOutput, error) {
			for _, filter := range input.InlineFilters {
				assert.Equal(t, common.Execution, filter.GetEntity())
				queryExpr, _ := filter.GetGormQueryExpr()
				if queryExpr.Args == admin.ExecutionMetadata_WORKSPACE && queryExpr.Query == "mode <> ?" {
					sawModeFilter = true
				}
			}

			return interfaces.ExecutionCollectionOutput{}, nil

		}
		repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetListCallback(executionListFunc)
		r := getDefaultPluginRegistry()
		mockResourceManager := getNoopMockResourceManager()
		execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

		_, err := execManager.ListExecutions(context.Background(), &admin.ResourceListRequest{
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
		assert.True(t, sawModeFilter)
	})
	t.Run("without mode filter added", func(t *testing.T) {
		repository := repositoryMocks.NewMockRepository()
		sawUserSpecifiedModeFilter := false
		sawWorkspaceExcludedModeFilter := false
		executionListFunc := func(
			ctx context.Context, input interfaces.ListResourceInput) (interfaces.ExecutionCollectionOutput, error) {
			for _, filter := range input.InlineFilters {
				assert.Equal(t, common.Execution, filter.GetEntity())
				queryExpr, _ := filter.GetGormQueryExpr()
				if queryExpr.Query == "mode <> ?" && queryExpr.Args == admin.ExecutionMetadata_WORKSPACE {
					sawWorkspaceExcludedModeFilter = true
				} else if queryExpr.Query == "mode = ?" && queryExpr.Args == "3" {
					sawUserSpecifiedModeFilter = true
				}
			}

			return interfaces.ExecutionCollectionOutput{}, nil

		}
		repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetListCallback(executionListFunc)
		r := getDefaultPluginRegistry()
		mockResourceManager := getNoopMockResourceManager()
		execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

		_, err := execManager.ListExecutions(context.Background(), &admin.ResourceListRequest{
			Id: &admin.NamedEntityIdentifier{
				Project: projectValue,
				Domain:  domainValue,
			},
			Limit: limit,
			SortBy: &admin.Sort{
				Direction: admin.Sort_ASCENDING,
				Key:       "execution_domain",
			},
			Token:   "2",
			Filters: "eq(mode, 3)",
		})
		assert.NoError(t, err)
		assert.True(t, sawUserSpecifiedModeFilter)
		assert.False(t, sawWorkspaceExcludedModeFilter)
	})
}

func TestListExecutions_MissingParameters(t *testing.T) {
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repositoryMocks.NewMockRepository(), r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

	_, err := execManager.ListExecutions(context.Background(), &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Domain: domainValue,
		},
		Limit: limit,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, err.(flyteAdminErrors.FlyteAdminError).Code())

	_, err = execManager.ListExecutions(context.Background(), &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: projectValue,
		},
		Limit: limit,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, err.(flyteAdminErrors.FlyteAdminError).Code())

	_, err = execManager.ListExecutions(context.Background(), &admin.ResourceListRequest{
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
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
	_, err := execManager.ListExecutions(context.Background(), &admin.ResourceListRequest{
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
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

	executionList, err := execManager.ListExecutions(context.Background(), &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: projectValue,
			Domain:  domainValue,
		},
		Limit: limit,
	})
	assert.EqualError(t, err, "failed to unmarshal spec")
	assert.Nil(t, executionList)
}

func TestGetExecutionCounts(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	getExecutionCountsFunc := func(
		ctx context.Context, input interfaces.CountResourceInput) (interfaces.ExecutionCountsByPhaseOutput, error) {
		var orgFilter, projectFilter, domainFilter, updatedAtFilter, nameFilter, modeFilter bool
		for _, filter := range input.InlineFilters {
			assert.Equal(t, common.Execution, filter.GetEntity())
			queryExpr, _ := filter.GetGormQueryExpr()
			if queryExpr.Args == orgValue && queryExpr.Query == executionOrgQueryExpr {
				orgFilter = true
			}
			if queryExpr.Args == projectValue && queryExpr.Query == executionProjectQueryExpr {
				projectFilter = true
			}
			if queryExpr.Args == domainValue && queryExpr.Query == executionDomainQueryExpr {
				domainFilter = true
			}
			if queryExpr.Args == executionCreatedAtValue && queryExpr.Query == executionCreatedAtQueryExpr {
				updatedAtFilter = true
			}
			if queryExpr.Args == nameValue && queryExpr.Query == executionNameQueryExpr {
				nameFilter = true
			}
			if queryExpr.Args == admin.ExecutionMetadata_WORKSPACE && queryExpr.Query == executionModeNeQueryExpr {
				modeFilter = true
			}

		}
		assert.True(t, orgFilter, "Missing org equality filter")
		assert.True(t, projectFilter, "Missing project equality filter")
		assert.True(t, domainFilter, "Missing domain equality filter")
		assert.True(t, updatedAtFilter, "Missing updated at filter")
		assert.False(t, nameFilter, "Included name equality filter")
		assert.True(t, modeFilter, "Missing mode filter")
		return interfaces.ExecutionCountsByPhaseOutput{
			{
				Phase: "FAILED",
				Count: int64(3),
			},
			{
				Phase: "SUCCEEDED",
				Count: int64(4),
			},
		}, nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCountByPhaseCallback(getExecutionCountsFunc)
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

	executionCountsGetResponse, err := execManager.GetExecutionCounts(context.Background(), &admin.ExecutionCountsGetRequest{
		Org:     orgValue,
		Project: projectValue,
		Domain:  domainValue,
		Filters: executionCreatedAtFilter,
	})
	executionCounts := executionCountsGetResponse.ExecutionCounts
	assert.NoError(t, err)
	assert.NotNil(t, executionCounts)
	assert.Len(t, executionCounts, 2)

	assert.Equal(t, core.WorkflowExecution_FAILED, executionCounts[0].Phase)
	assert.Equal(t, int64(3), executionCounts[0].Count)
	assert.Equal(t, core.WorkflowExecution_SUCCEEDED, executionCounts[1].Phase)
	assert.Equal(t, int64(4), executionCounts[1].Count)
}

func TestGetExecutionCounts_MissingParameters(t *testing.T) {
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repositoryMocks.NewMockRepository(), r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

	// Test missing domain
	_, err := execManager.GetExecutionCounts(context.Background(), &admin.ExecutionCountsGetRequest{
		Project: projectValue,
		Filters: executionCreatedAtFilter,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, err.(flyteAdminErrors.FlyteAdminError).Code())

	// Test missing project
	_, err = execManager.GetExecutionCounts(context.Background(), &admin.ExecutionCountsGetRequest{
		Domain:  domainValue,
		Filters: executionCreatedAtFilter,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, err.(flyteAdminErrors.FlyteAdminError).Code())

	// Filter is optional
	_, err = execManager.GetExecutionCounts(context.Background(), &admin.ExecutionCountsGetRequest{
		Project: projectValue,
		Domain:  domainValue,
	})
	assert.NoError(t, err)
}

func TestGetExecutionCounts_DatabaseError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	expectedErr := errors.New("expected error")
	getExecutionCountsFunc := func(
		ctx context.Context, input interfaces.CountResourceInput) (interfaces.ExecutionCountsByPhaseOutput, error) {
		return interfaces.ExecutionCountsByPhaseOutput{}, expectedErr
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCountByPhaseCallback(getExecutionCountsFunc)
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
	_, err := execManager.GetExecutionCounts(context.Background(), &admin.ExecutionCountsGetRequest{
		Project: projectValue,
		Domain:  domainValue,
		Filters: executionCreatedAtFilter,
	})
	assert.EqualError(t, err, expectedErr.Error())
}

func TestGetExecutionCounts_TransformerError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	getExecutionCountsFunc := func(
		ctx context.Context, input interfaces.CountResourceInput) (interfaces.ExecutionCountsByPhaseOutput, error) {
		return interfaces.ExecutionCountsByPhaseOutput{
			{
				Phase: "INVALID_PHASE",
				Count: int64(3),
			},
		}, nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCountByPhaseCallback(getExecutionCountsFunc)
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

	executionCountsGetResponse, err := execManager.GetExecutionCounts(context.Background(), &admin.ExecutionCountsGetRequest{
		Project: projectValue,
		Domain:  domainValue,
		Filters: executionCreatedAtFilter,
	})
	assert.EqualError(t, err, "Failed to transform INVALID_PHASE into an execution phase.")
	assert.Nil(t, executionCountsGetResponse)
}

func TestGetRunningExecutionsCount(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	getRunningExecutionsCountFunc := func(
		ctx context.Context, input interfaces.CountResourceInput) (int64, error) {
		var orgFilter, projectFilter, domainFilter, nameFilter, modeFilter bool
		for _, filter := range input.InlineFilters {
			assert.Equal(t, common.Execution, filter.GetEntity())
			queryExpr, _ := filter.GetGormQueryExpr()
			if queryExpr.Args == orgValue && queryExpr.Query == executionOrgQueryExpr {
				orgFilter = true
			}
			if queryExpr.Args == projectValue && queryExpr.Query == executionProjectQueryExpr {
				projectFilter = true
			}
			if queryExpr.Args == domainValue && queryExpr.Query == executionDomainQueryExpr {
				domainFilter = true
			}
			if queryExpr.Args == nameValue && queryExpr.Query == executionNameQueryExpr {
				nameFilter = true
			}
			if queryExpr.Args == admin.ExecutionMetadata_WORKSPACE && queryExpr.Query == executionModeNeQueryExpr {
				modeFilter = true
			}
		}
		assert.True(t, orgFilter, "Missing org equality filter")
		assert.True(t, projectFilter, "Missing project equality filter")
		assert.True(t, domainFilter, "Missing domain equality filter")
		assert.False(t, nameFilter, "Included name equality filter")
		assert.True(t, modeFilter, "Missing mode filter")
		return 3, nil
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCountCallback(getRunningExecutionsCountFunc)
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

	runningExecutionsCountGetResponse, err := execManager.GetRunningExecutionsCount(context.Background(), &admin.RunningExecutionsCountGetRequest{
		Org:     orgValue,
		Project: projectValue,
		Domain:  domainValue,
	})
	assert.NoError(t, err)
	assert.NotNil(t, runningExecutionsCountGetResponse)
	assert.Equal(t, int64(3), runningExecutionsCountGetResponse.Count)
}

func TestGetRunningExecutionsCount_MissingParameters(t *testing.T) {
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repositoryMocks.NewMockRepository(), r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

	_, err := execManager.GetRunningExecutionsCount(context.Background(), &admin.RunningExecutionsCountGetRequest{
		Project: projectValue,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, err.(flyteAdminErrors.FlyteAdminError).Code())

	_, err = execManager.GetRunningExecutionsCount(context.Background(), &admin.RunningExecutionsCountGetRequest{
		Domain: domainValue,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, err.(flyteAdminErrors.FlyteAdminError).Code())
}

func TestGetRunningExecutionsCount_DatabaseError(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	expectedErr := errors.New("expected error")
	getRunningExecutionsCountFunc := func(
		ctx context.Context, input interfaces.CountResourceInput) (int64, error) {
		return 0, expectedErr
	}
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCountCallback(getRunningExecutionsCountFunc)
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
	_, err := execManager.GetRunningExecutionsCount(context.Background(), &admin.RunningExecutionsCountGetRequest{
		Project: projectValue,
		Domain:  domainValue,
	})
	assert.EqualError(t, err, expectedErr.Error())
}

func TestExecutionManager_PublishNotifications(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	mockResourceManager := getNoopMockResourceManager()
	queue := executions.NewQueueAllocator(getMockExecutionsConfigProvider(), repository, mockResourceManager)

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
	workflowRequest := &admin.WorkflowExecutionEventRequest{
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
	var execClosure = &admin.ExecutionClosure{
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

	execClosureBytes, _ := proto.Marshal(execClosure)
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
	mockResourceManager := getNoopMockResourceManager()
	queue := executions.NewQueueAllocator(getMockExecutionsConfigProvider(), repository, mockResourceManager)

	var execManager = &ExecutionManager{
		db:                 repository,
		config:             getMockExecutionsConfigProvider(),
		storageClient:      getMockStorageForExecTest(context.Background()),
		queueAllocator:     queue,
		_clock:             clock.New(),
		systemMetrics:      newExecutionSystemMetrics(mockScope.NewTestScope()),
		notificationClient: &mockPublisher,
	}

	workflowRequest := &admin.WorkflowExecutionEventRequest{
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
	mockResourceManager := getNoopMockResourceManager()
	queue := executions.NewQueueAllocator(getMockExecutionsConfigProvider(), repository, mockResourceManager)

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
	workflowRequest := &admin.WorkflowExecutionEventRequest{
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
	var execClosure = &admin.ExecutionClosure{
		Notifications: testutils.GetExecutionRequest().Spec.GetNotifications().Notifications,
		WorkflowId: &core.Identifier{
			ResourceType: core.ResourceType_WORKFLOW,
			Project:      "wf_project",
			Domain:       "wf_domain",
			Name:         "wf_name",
			Version:      "wf_version",
		},
	}
	execClosureBytes, _ := proto.Marshal(execClosure)
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
	mockResourceManager := getNoopMockResourceManager()
	queue := executions.NewQueueAllocator(getMockExecutionsConfigProvider(), repository, mockResourceManager)

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
	workflowRequest := &admin.WorkflowExecutionEventRequest{
		Event: &event.WorkflowExecutionEvent{
			Phase: core.WorkflowExecution_SUCCEEDED,
			OutputResult: &event.WorkflowExecutionEvent_OutputUri{
				OutputUri: "somestring",
			},
			ExecutionId: &executionIdentifier,
		},
	}
	var execClosure = &admin.ExecutionClosure{
		Notifications: testutils.GetExecutionRequest().Spec.GetNotifications().Notifications,
	}
	execClosureBytes, _ := proto.Marshal(execClosure)
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
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

	identity, err := auth.NewIdentityContext("", principal, "", time.Now(), sets.NewString(), nil, nil)
	assert.NoError(t, err)
	ctx := identity.WithContext(context.Background())
	resp, err := execManager.TerminateExecution(ctx, &admin.ExecutionTerminateRequest{
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
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)

	updateCalled := false
	repository := repositoryMocks.NewMockRepository()
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetUpdateCallback(func(
		context context.Context, execution models.Execution) error {
		updateCalled = true
		assert.Equal(t, core.WorkflowExecution_ABORTING.String(), execution.Phase)
		return nil
	})
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

	resp, err := execManager.TerminateExecution(context.Background(), &admin.ExecutionTerminateRequest{
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
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
	resp, err := execManager.TerminateExecution(context.Background(), &admin.ExecutionTerminateRequest{
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
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)

	repository := repositoryMocks.NewMockRepository()
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
			return models.Execution{
				Phase: core.WorkflowExecution_SUCCEEDED.String(),
			}, nil
		})
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
	resp, err := execManager.TerminateExecution(context.Background(), &admin.ExecutionTerminateRequest{
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
			InputsURI:    storage.DataReference(inputURI),
		}, nil
	}
	mockExecutionRemoteURL := dataMocks.NewMockRemoteURL()
	mockExecutionRemoteURL.(*dataMocks.MockRemoteURL).GetCallback = func(
		ctx context.Context, uri string) (*admin.UrlBlob, error) {
		switch uri {
		case inputURI, outputURI:
			return &admin.UrlBlob{
				Url:   uri,
				Bytes: 200,
			}, nil
		}

		return &admin.UrlBlob{}, errors.New("unexpected input")
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
		if reference.String() == inputURI {
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
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), mockStorage, mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

	dataResponse, err := execManager.GetExecutionData(context.Background(), &admin.WorkflowExecutionGetDataRequest{
		Id: &executionIdentifier,
	})

	assert.NoError(t, err)
	assert.True(t, proto.Equal(&admin.WorkflowExecutionGetDataResponse{
		Outputs: &admin.UrlBlob{
			Url:   outputURI,
			Bytes: 200,
		},
		Inputs: &admin.UrlBlob{
			Url:   inputURI,
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
	mockResourceManager := new(managerMocks.ResourceInterface)
	mockResourceManager.On("GetResource", mock.Anything, mock.MatchedBy(func(request managerInterfaces.ResourceRequest) bool {
		assert.Equal(t, project, request.Project)
		assert.Equal(t, domain, request.Domain)
		assert.Equal(t, workflowName, request.Workflow)
		assert.Equal(t, launchPlanName, request.LaunchPlan)
		assert.Equal(t, admin.MatchableResource_PLUGIN_OVERRIDE, request.ResourceType)
		return true
	})).Return(&managerInterfaces.ResourceResponse{
		Project: project,
		Domain:  domain,
		Attributes: commonTestUtils.GetPluginOverridesAttributes(map[string][]string{
			"python": {"plugin a"},
			"hive":   {"plugin b"},
		}),
	}, nil)

	r := getDefaultPluginRegistry()

	execManager := NewExecutionManager(db, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

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
	mockResourceManager := new(managerMocks.ResourceInterface)
	mockResourceManager.On("GetResource", mock.Anything, mock.Anything).Return(&managerInterfaces.ResourceResponse{}, errors.New("uh oh"))

	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDUserProperties, shared.DefaultGetUserPropertiesFunc)
	execManager := NewExecutionManager(db, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

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
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()

	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
	execution, err := execManager.GetExecution(context.Background(), &admin.WorkflowExecutionGetRequest{
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
				Org:     "org",
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
		ctx context.Context, uri string) (*admin.UrlBlob, error) {
		if uri == outputURI {
			return &admin.UrlBlob{
				Url:   "outputs",
				Bytes: 200,
			}, nil
		} else if strings.HasSuffix(uri, shared.Inputs) {
			return &admin.UrlBlob{
				Url:   "inputs",
				Bytes: 200,
			}, nil
		}

		return &admin.UrlBlob{}, errors.New("unexpected input")
	}

	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(executionGetFunc)
	storageClient := getMockStorageForExecTest(context.Background())
	r := getDefaultPluginRegistry()

	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), storageClient, mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

	dataResponse, err := execManager.GetExecutionData(context.Background(), &admin.WorkflowExecutionGetDataRequest{
		Id: &executionIdentifier,
	})

	assert.EqualError(t, err, "could not find value in storage [s3://bucket/output_uri]")
	assert.Empty(t, dataResponse)

	var inputs core.LiteralMap

	err = storageClient.ReadProtobuf(context.Background(), "s3://bucket/metadata/project/domain/name/inputs", &inputs)

	assert.NoError(t, err)
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
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
	response, err := execManager.CreateExecution(context.Background(), getLegacyExecutionRequest(), requestedAt)
	assert.Nil(t, err)

	expectedResponse := &admin.ExecutionCreateResponse{
		Id: &executionIdentifier,
	}
	assert.Nil(t, err)
	assert.True(t, proto.Equal(expectedResponse, response))
}

func TestRelaunchExecution_LegacyModel(t *testing.T) {
	// Set up mocks.
	repository := getMockRepositoryForExecTest()
	setDefaultLpCallbackForExecTest(repository)
	storageClient := getMockStorageForExecTest(context.Background())

	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{}, nil)
	mockExecutor.OnID().Return("testMockExecutor")
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), storageClient, mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)
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
	response, err := execManager.RelaunchExecution(context.Background(), &admin.ExecutionRelaunchRequest{
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
			if queryExpr.Args == projectValue && queryExpr.Query == executionProjectQueryExpr {
				projectFilter = true
			}
			if queryExpr.Args == domainValue && queryExpr.Query == executionDomainQueryExpr {
				domainFilter = true
			}
			if queryExpr.Args == nameValue && queryExpr.Query == executionNameQueryExpr {
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
	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

	executionList, err := execManager.ListExecutions(context.Background(), &admin.ResourceListRequest{
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

	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repositoryMocks.NewMockRepository(), r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

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

	r := getDefaultPluginRegistry()
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repositoryMocks.NewMockRepository(), r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

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
		r := getDefaultPluginRegistry()
		mockResourceManager := getNoopMockResourceManager()
		execManager := NewExecutionManager(repositoryMocks.NewMockRepository(), r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

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
		r := getDefaultPluginRegistry()
		mockResourceManager := getNoopMockResourceManager()
		execManager := NewExecutionManager(repositoryMocks.NewMockRepository(), r, getMockExecutionsConfigProvider(), getMockStorageForExecTest(context.Background()), mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

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
		storagePrefix, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil))
	namedEntityManager := NewNamedEntityManager(repository, getMockConfigForNETest(), mockScope.NewTestScope())

	mockExecutor := workflowengineMocks.WorkflowExecutor{}
	mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{}, nil)
	mockExecutor.OnID().Return("testMockExecutor")
	r := getDefaultPluginRegistry()
	r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
	mockResourceManager := getNoopMockResourceManager()
	execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), mockStorage, mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, workflowManager, nil, namedEntityManager, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

	request := &admin.ExecutionCreateRequest{
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
	_, ferr := marshaller.MarshalToString(request)
	assert.NoError(t, ferr)

	// test once to create an initial launchplan
	_, err := execManager.CreateExecution(context.TODO(), request, time.Now())
	assert.NoError(t, err)

	// test again to ensure existing launchplan retrieval works
	_, err = execManager.CreateExecution(context.TODO(), request, time.Now())
	assert.NoError(t, err)
}

func TestCreateSingleTaskFromNodeExecution(t *testing.T) {

	v1 := "v1"
	taskID := &core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      "flytekit",
		Domain:       "production",
		Name:         "simple_task",
		Version:      v1,
	}
	diffVersion := "v2"
	diffVersionTaskID := &core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      "flytekit",
		Domain:       "production",
		Name:         "simple_task",
		Version:      diffVersion,
	}

	arrayNodeId := "array-node-id"
	arrayNode := &core.Node{
		Id: arrayNodeId,
		Target: &core.Node_ArrayNode{
			ArrayNode: &core.ArrayNode{
				Node: &core.Node{
					Target: &core.Node_TaskNode{
						TaskNode: &core.TaskNode{
							Reference: &core.TaskNode_ReferenceId{
								ReferenceId: taskID,
							},
						},
					},
				},
			},
		},
	}
	diffVersionArrayNode := core.Node{
		Id: arrayNodeId,
		Target: &core.Node_ArrayNode{
			ArrayNode: &core.ArrayNode{
				Node: &core.Node{
					Target: &core.Node_TaskNode{
						TaskNode: &core.TaskNode{
							Reference: &core.TaskNode_ReferenceId{
								ReferenceId: diffVersionTaskID,
							},
						},
					},
				},
			},
		},
	}

	taskNodeId := "task-node-id"
	taskNode := &core.Node{
		Id:     taskNodeId,
		Target: &core.Node_TaskNode{},
	}

	workflowClosureV1 := testutils.GetWorkflowClosure()
	workflowClosureV1.CompiledWorkflow.Primary.Template.Nodes = []*core.Node{
		arrayNode,
		taskNode,
	}
	workflowClosureV2 := testutils.GetWorkflowClosure()
	workflowClosureV2.CompiledWorkflow.Primary.Template.Nodes = []*core.Node{
		&diffVersionArrayNode,
	}

	mockStorage := getMockStorageForExecTest(context.Background())

	const wfClosureIdV1 = "s3://flyte/metadata/admin/remote wf_closure v1"
	const wfClosureIdV2 = "s3://flyte/metadata/admin/remote wf_closure v2"
	err := mockStorage.WriteProtobuf(context.TODO(), wfClosureIdV1, defaultStorageOptions, workflowClosureV1)
	assert.NoError(t, err)
	err = mockStorage.WriteProtobuf(context.TODO(), wfClosureIdV2, defaultStorageOptions, workflowClosureV2)
	assert.NoError(t, err)

	testCases := []struct {
		name                           string
		requestTaskId                  *core.Identifier
		requestParentNodeId            string
		node                           *core.Node
		createLaunchPlanFromNodeCalled bool
		diffVersion                    bool
	}{
		{
			name:                           "not ArrayNode",
			requestTaskId:                  taskID,
			requestParentNodeId:            taskNodeId,
			node:                           taskNode,
			createLaunchPlanFromNodeCalled: false,
		},
		{
			name:                           "found array node, no version difference 1",
			requestParentNodeId:            arrayNodeId,
			requestTaskId:                  taskID,
			node:                           arrayNode,
			createLaunchPlanFromNodeCalled: true,
		},
		{
			name:                           "found array node, different version",
			requestParentNodeId:            arrayNodeId,
			requestTaskId:                  diffVersionTaskID,
			node:                           arrayNode,
			createLaunchPlanFromNodeCalled: true,
			diffVersion:                    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRequest := testutils.GetExecutionRequest()
			parentNodeExecutionID := &core.NodeExecutionIdentifier{
				NodeId: tc.requestParentNodeId,
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: mockRequest.Project,
					Domain:  mockRequest.Domain,
					Name:    mockRequest.Name,
				},
			}
			mockRequest.Spec.Metadata = &admin.ExecutionMetadata{
				ParentNodeExecution: parentNodeExecutionID,
			}
			mockRequest.Spec.LaunchPlan = tc.requestTaskId

			repository := repositoryMocks.NewMockRepository()

			mockExecutor := workflowengineMocks.WorkflowExecutor{}
			mockExecutor.OnExecuteMatch(mock.Anything, mock.Anything, mock.Anything).Return(workflowengineInterfaces.ExecutionResponse{}, nil)
			mockExecutor.OnID().Return("testMockExecutor")
			r := getDefaultPluginRegistry()
			r.RegisterDefault(plugins.PluginIDWorkflowExecutor, &mockExecutor)
			r.RegisterDefault(plugins.PluginIDUserProperties, shared.DefaultGetUserPropertiesFunc)

			mockResourceManager := getNoopMockResourceManager()

			generatedLaunchPlanID := &core.Identifier{
				ResourceType: core.ResourceType_LAUNCH_PLAN,
				Project:      "flytekit",
				Domain:       "production",
				Name:         "generated_lp",
				Version:      v1,
			}
			var setCreateLaunchPlanFromNodeCalled bool
			mockLaunchPlanManager := &managerMocks.MockLaunchPlanManager{}
			mockLaunchPlanManager.SetCreateLaunchPlanFromNode(func(ctx context.Context, request *admin.CreateLaunchPlanFromNodeRequest) (*admin.CreateLaunchPlanFromNodeResponse, error) {
				if tc.diffVersion {
					assert.True(t, proto.Equal(&diffVersionArrayNode, request.GetSubNodeSpec()))
				} else {
					assert.True(t, proto.Equal(tc.node, request.GetSubNodeSpec()))
				}

				setCreateLaunchPlanFromNodeCalled = true
				return &admin.CreateLaunchPlanFromNodeResponse{
					LaunchPlan: &admin.LaunchPlan{
						Id: generatedLaunchPlanID,
					},
				}, nil
			})

			executionClosure := &admin.ExecutionClosure{
				WorkflowId: &core.Identifier{
					Project: "flytekit",
					Domain:  "production",
					Name:    "workflow",
					Version: v1,
				},
			}
			executionClosureBytes, err := proto.Marshal(executionClosure)
			assert.NoError(t, err)
			executionSpec := &admin.ExecutionSpec{
				LaunchPlan: &core.Identifier{
					Version: v1,
				},
			}
			executionSpecBytes, err := proto.Marshal(executionSpec)
			assert.NoError(t, err)
			repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
				return models.Execution{
					Closure: executionClosureBytes,
					Spec:    executionSpecBytes,
				}, nil
			})

			nodeExecutionMetadata := admin.NodeExecutionMetaData{
				SpecNodeId: tc.requestParentNodeId,
			}
			nodeExecutionMetadataBytes, err := proto.Marshal(&nodeExecutionMetadata)
			assert.NoError(t, err)
			repository.NodeExecutionRepo().(*repositoryMocks.MockNodeExecutionRepo).SetGetCallback(
				func(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
					assert.EqualValues(t, parentNodeExecutionID.GetNodeId(), input.NodeExecutionIdentifier.GetNodeId())
					assert.EqualValues(t, parentNodeExecutionID.GetExecutionId(), input.NodeExecutionIdentifier.GetExecutionId())
					return models.NodeExecution{
						BaseModel: models.BaseModel{
							ID: 1,
						},
						NodeExecutionMetadata: nodeExecutionMetadataBytes,
					}, nil
				},
			)

			var wfGetCalledCount = 0
			repository.WorkflowRepo().(*repositoryMocks.MockWorkflowRepo).SetGetCallback(
				func(input interfaces.Identifier) (models.Workflow, error) {
					if wfGetCalledCount == 0 {
						assert.Equal(t, executionClosure.WorkflowId.Project, input.Project)
						assert.Equal(t, executionClosure.WorkflowId.Domain, input.Domain)
						assert.Equal(t, executionClosure.WorkflowId.Name, input.Name)
						assert.Equal(t, executionClosure.WorkflowId.Version, input.Version)
					}
					wfGetCalledCount++

					var wfClosureId string
					if input.Version == "v1" {
						wfClosureId = wfClosureIdV1
					} else {
						wfClosureId = wfClosureIdV2
					}
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
						RemoteClosureIdentifier: wfClosureId,
					}, nil
				})

			lpSpec := testutils.GetSampleLpSpecForTest()
			lpSpec.WorkflowId = generatedLaunchPlanID
			lpSpecBytes, err := proto.Marshal(lpSpec)
			assert.NoError(t, err)
			lpClosure := admin.LaunchPlanClosure{
				ExpectedInputs: lpSpec.DefaultInputs,
			}
			lpClosureBytes, err := proto.Marshal(&lpClosure)
			assert.NoError(t, err)
			// since the CreateLaunchPlanFromNode creates or fetches an existing launch plan, we just check if that returned launch plan is used
			repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetGetCallback(func(input interfaces.Identifier) (models.LaunchPlan, error) {
				if tc.createLaunchPlanFromNodeCalled {
					assert.Equal(t, generatedLaunchPlanID.Project, input.Project)
					assert.Equal(t, generatedLaunchPlanID.Domain, input.Domain)
					assert.Equal(t, generatedLaunchPlanID.Name, input.Name)
					assert.Equal(t, generatedLaunchPlanID.Version, input.Version)
				}
				return models.LaunchPlan{
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
				}, nil
			})

			workflowManager := NewWorkflowManager(
				repository,
				getMockWorkflowConfigProvider(), getMockWorkflowCompiler(), mockStorage,
				storagePrefix, mockScope.NewTestScope(), artifacts.NewArtifactRegistry(context.Background(), nil))
			namedEntityManager := NewNamedEntityManager(repository, getMockConfigForNETest(), mockScope.NewTestScope())
			execManager := NewExecutionManager(repository, r, getMockExecutionsConfigProvider(), mockStorage, mockScope.NewTestScope(), mockScope.NewTestScope(), &mockPublisher, mockExecutionRemoteURL, workflowManager, mockLaunchPlanManager, namedEntityManager, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), mockResourceManager)

			marshaller := jsonpb.Marshaler{}
			_, ferr := marshaller.MarshalToString(request)
			assert.NoError(t, ferr)

			_, err = execManager.CreateExecution(context.TODO(), mockRequest, time.Now())
			assert.NoError(t, err)

			if tc.createLaunchPlanFromNodeCalled {
				assert.True(t, setCreateLaunchPlanFromNodeCalled)
			} else {
				assert.False(t, setCreateLaunchPlanFromNodeCalled)
			}
		})
	}
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
	t.Run("value from request matches value from config", func(t *testing.T) {
		reqClusterAssignment := admin.ClusterAssignment{ClusterPoolName: "gpu"}
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
	t.Run("no value in DB nor in config, takes value from request", func(t *testing.T) {
		mockConfig := getMockExecutionsConfigProvider()

		executionManager := ExecutionManager{
			resourceManager: &managerMocks.MockResourceManager{},
			config:          mockConfig,
		}

		reqClusterAssignment := admin.ClusterAssignment{ClusterPoolName: "gpu"}
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
	t.Run("empty value in DB, takes value from request", func(t *testing.T) {
		clusterPoolAsstProvider := &runtimeIFaceMocks.ClusterPoolAssignmentConfiguration{}
		clusterPoolAsstProvider.OnGetClusterPoolAssignments().Return(runtimeInterfaces.ClusterPoolAssignments{
			workflowIdentifier.GetDomain(): runtimeInterfaces.ClusterPoolAssignment{
				Pool: "",
			},
		})
		mockConfig := getMockExecutionsConfigProvider()
		mockConfig.(*runtimeMocks.MockConfigurationProvider).AddClusterPoolAssignmentConfiguration(clusterPoolAsstProvider)

		executionManager := ExecutionManager{
			resourceManager: &managerMocks.MockResourceManager{},
			config:          mockConfig,
		}

		reqClusterAssignment := admin.ClusterAssignment{ClusterPoolName: "gpu"}
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
	t.Run("value from request doesn't match value from config", func(t *testing.T) {
		reqClusterAssignment := admin.ClusterAssignment{ClusterPoolName: "swimming-pool"}
		_, err := executionManager.getClusterAssignment(context.TODO(), &admin.ExecutionCreateRequest{
			Project: workflowIdentifier.Project,
			Domain:  workflowIdentifier.Domain,
			Spec: &admin.ExecutionSpec{
				ClusterAssignment: &reqClusterAssignment,
			},
		})
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Equal(t, `execution with project "project" and domain "domain" cannot run on cluster pool "swimming-pool", because its configured to run on pool "gpu"`, st.Message())
	})
}

func TestGetExternalResourceAttribute(t *testing.T) {
	ExternalResourceAttributes := &admin.ExternalResourceAttributes{
		Connections: map[string]*core.Connection{
			"conn1": {
				TaskType: "openai",
				Secrets:  map[string]string{"key1": "value1"},
				Configs:  map[string]string{"key2": "value2"},
			},
		},
	}
	resourceManager := managerMocks.MockResourceManager{}
	resourceManager.GetResourceFunc = func(ctx context.Context,
		request managerInterfaces.ResourceRequest) (*managerInterfaces.ResourceResponse, error) {
		if len(request.Project) != 0 && len(request.Domain) != 0 {
			assert.EqualValues(t, request, managerInterfaces.ResourceRequest{
				Project:      workflowIdentifier.Project,
				Domain:       workflowIdentifier.Domain,
				ResourceType: admin.MatchableResource_EXTERNAL_RESOURCE,
			})
		}
		if len(request.Project) != 0 && len(request.Domain) == 0 {
			assert.EqualValues(t, request, managerInterfaces.ResourceRequest{
				Project:      workflowIdentifier.Project,
				ResourceType: admin.MatchableResource_EXTERNAL_RESOURCE,
			})
		}
		if len(request.Project) == 0 && len(request.Domain) == 0 {
			assert.EqualValues(t, request, managerInterfaces.ResourceRequest{
				ResourceType: admin.MatchableResource_EXTERNAL_RESOURCE,
			})
		}
		return &managerInterfaces.ResourceResponse{
			Attributes: &admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_ExternalResourceAttributes{
					ExternalResourceAttributes: ExternalResourceAttributes,
				},
			},
		}, nil
	}

	executionManager := ExecutionManager{resourceManager: &resourceManager}

	t.Run("value from db", func(t *testing.T) {
		era, err := executionManager.getExternalResourceAttributes(context.TODO(), &admin.ExecutionCreateRequest{
			Project: workflowIdentifier.Project,
			Domain:  workflowIdentifier.Domain,
			Spec:    &admin.ExecutionSpec{},
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(era.GetConnections()))
		actualConn := era.GetConnections()["conn1"]
		expectedConn := ExternalResourceAttributes.Connections["conn1"]
		assert.Equal(t, actualConn.Secrets, expectedConn.Secrets)
		assert.Equal(t, actualConn.Configs, expectedConn.Configs)
		assert.Equal(t, actualConn.TaskType, expectedConn.TaskType)
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

func TestQueryTemplate(t *testing.T) {
	ctx := context.Background()

	aTime := time.Date(
		2063, 4, 5, 00, 00, 00, 0, time.UTC)

	rawInputs := map[string]interface{}{
		"aStr":   "hello world",
		"anInt":  1,
		"aFloat": 1.3,
		"aTime":  aTime,
	}

	otherInputs, err := coreutils.MakeLiteralMap(rawInputs)
	assert.NoError(t, err)

	m := ExecutionManager{}

	ak := &core.ArtifactKey{
		Project: "project",
		Domain:  "domain",
		Name:    "testname",
	}

	akNamePlusOrg := &core.ArtifactKey{
		Project: "",
		Domain:  "",
		Name:    "testname",
		Org:     "my-gh-handle",
	}

	t.Run("test all present, nothing to fill in", func(t *testing.T) {
		pMap := map[string]*core.LabelValue{
			"partition1": {Value: &core.LabelValue_StaticValue{StaticValue: "my value"}},
			"partition2": {Value: &core.LabelValue_StaticValue{StaticValue: "my value 2"}},
		}
		p := &core.Partitions{Value: pMap}

		q := &core.ArtifactQuery{
			Identifier: &core.ArtifactQuery_ArtifactId{
				ArtifactId: &core.ArtifactID{
					ArtifactKey:   akNamePlusOrg,
					Partitions:    p,
					TimePartition: nil,
				},
			},
		}

		filledQuery, err := m.fillInTemplateArgs(ctx, q, otherInputs.Literals)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(q, filledQuery))
		assert.Equal(t, "my-gh-handle", filledQuery.GetArtifactId().GetArtifactKey().GetOrg())

		q.GetArtifactId().ArtifactKey = ak
		ctx = contextutils.WithProjectDomain(ctx, "project", "domain")
		filledQuery, err = m.fillInTemplateArgs(ctx, q, otherInputs.Literals)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(q, filledQuery))
	})

	t.Run("template date-times, both in explicit tp and not", func(t *testing.T) {
		pMap := map[string]*core.LabelValue{
			"partition1": {Value: &core.LabelValue_InputBinding{InputBinding: &core.InputBindingData{Var: "aTime"}}},
			"partition2": {Value: &core.LabelValue_StaticValue{StaticValue: "my value 2"}},
		}
		p := &core.Partitions{Value: pMap}

		q := &core.ArtifactQuery{
			Identifier: &core.ArtifactQuery_ArtifactId{
				ArtifactId: &core.ArtifactID{
					ArtifactKey:   ak,
					Partitions:    p,
					TimePartition: &core.TimePartition{Value: &core.LabelValue{Value: &core.LabelValue_InputBinding{InputBinding: &core.InputBindingData{Var: "aTime"}}}},
				},
			},
		}

		filledQuery, err := m.fillInTemplateArgs(ctx, q, otherInputs.Literals)
		assert.NoError(t, err)
		staticTime := filledQuery.GetArtifactId().Partitions.Value["partition1"].GetStaticValue()
		assert.Equal(t, "2063-04-05", staticTime)
		assert.Equal(t, int64(2942956800), filledQuery.GetArtifactId().TimePartition.Value.GetTimeValue().Seconds)
	})

	t.Run("something missing", func(t *testing.T) {
		pMap := map[string]*core.LabelValue{
			"partition1": {Value: &core.LabelValue_StaticValue{StaticValue: "my value"}},
			"partition2": {Value: &core.LabelValue_StaticValue{StaticValue: "my value 2"}},
		}
		p := &core.Partitions{Value: pMap}

		q := &core.ArtifactQuery{
			Identifier: &core.ArtifactQuery_ArtifactId{
				ArtifactId: &core.ArtifactID{
					ArtifactKey:   ak,
					Partitions:    p,
					TimePartition: &core.TimePartition{Value: &core.LabelValue{Value: &core.LabelValue_InputBinding{InputBinding: &core.InputBindingData{Var: "wrong var"}}}},
				},
			},
		}

		_, err := m.fillInTemplateArgs(ctx, q, otherInputs.Literals)
		assert.Error(t, err)
	})
}

func TestLiteralParsing(t *testing.T) {
	ctx := context.Background()

	aDate := time.Date(
		2063, 4, 5, 00, 00, 00, 0, time.UTC)
	aTime := time.Date(
		2063, 4, 5, 15, 42, 00, 0, time.UTC)

	rawInputs := map[string]interface{}{
		"aStr":   "hello world",
		"anInt":  1,
		"aFloat": 6.62607015e-34,
		"aDate":  aDate,
		"aTime":  aTime,
		"aBool":  true,
	}

	otherInputs, err := coreutils.MakeLiteralMap(rawInputs)
	assert.NoError(t, err)

	m := ExecutionManager{}
	testCases := []struct {
		varName        string
		expectedString string
	}{
		{"aStr", "hello world"},
		{"anInt", "1"},
		{"aFloat", "0.00"},
		{"aDate", "2063-04-05"},
		{"aTime", "2063-04-05T15:42:00Z"},
		{"aBool", "true"},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Parse %s", tc.varName), func(t *testing.T) {
			binding := &core.InputBindingData{Var: tc.varName}
			strVal, err := m.getStringFromInput(ctx, binding, otherInputs.Literals)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedString, strVal)
		})
	}
}

func TestResolveNotWorking(t *testing.T) {
	mockConfig := getMockExecutionsConfigProvider()

	execManager := NewExecutionManager(nil, nil, mockConfig, nil, mockScope.NewTestScope(), mockScope.NewTestScope(), nil, nil, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), getNoopMockResourceManager()).(*ExecutionManager)

	pm, artifactIDs, err := execManager.ResolveParameterMapArtifacts(context.Background(), nil, nil, "")
	assert.Nil(t, err)
	fmt.Println(pm, artifactIDs)

}

func TestTrackingBitExtract(t *testing.T) {
	mockConfig := getMockExecutionsConfigProvider()

	execManager := NewExecutionManager(nil, nil, mockConfig, nil, mockScope.NewTestScope(), mockScope.NewTestScope(), nil, nil, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil), getNoopMockResourceManager()).(*ExecutionManager)

	lit := core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Primitive{
					Primitive: &core.Primitive{
						Value: &core.Primitive_Integer{
							Integer: 1,
						},
					},
				},
			},
		},
		Metadata: map[string]string{"_ua": "proj/domain/name@version"},
	}
	inputMap := core.LiteralMap{
		Literals: map[string]*core.Literal{
			"a": &lit,
		},
	}
	inputColl := core.LiteralCollection{
		Literals: []*core.Literal{
			&lit,
		},
	}

	var trackers = make(map[string]string)
	execManager.ExtractArtifactTrackers(trackers, &lit)
	assert.Equal(t, 1, len(trackers))

	trackers = make(map[string]string)
	execManager.ExtractArtifactTrackers(trackers, &core.Literal{Value: &core.Literal_Map{Map: &inputMap}})
	assert.Equal(t, 1, len(trackers))

	trackers = make(map[string]string)
	execManager.ExtractArtifactTrackers(trackers, &core.Literal{Value: &core.Literal_Collection{Collection: &inputColl}})
	assert.Equal(t, 1, len(trackers))
	assert.Equal(t, "", trackers["proj/domain/name@version"])
}

func TestResolveParameterMapArtifacts(t *testing.T) {
	ak := &core.ArtifactKey{
		Project: "project",
		Domain:  "domain",
		Name:    "testname",
	}
	returnID := &core.ArtifactID{
		ArtifactKey: ak,
		Version:     "abc",
	}
	one, err := coreutils.MakeLiteral(1)
	assert.NoError(t, err)

	pMap := map[string]*core.LabelValue{
		"partition1": {Value: &core.LabelValue_StaticValue{StaticValue: "my value"}},
		"partition2": {Value: &core.LabelValue_StaticValue{StaticValue: "my value 2"}},
	}
	p := &core.Partitions{Value: pMap}

	q := core.ArtifactQuery{
		Identifier: &core.ArtifactQuery_ArtifactId{
			ArtifactId: &core.ArtifactID{
				ArtifactKey: ak,
				Partitions:  p,
			},
		},
	}

	inputs := core.ParameterMap{
		Parameters: map[string]*core.Parameter{
			"input1": {
				Var:      nil,
				Behavior: &core.Parameter_ArtifactQuery{ArtifactQuery: &q},
			},
		},
	}

	t.Run("context metadata provides github handle", func(t *testing.T) {
		client := artifactMocks.ArtifactRegistryClient{}
		client.On("GetArtifact", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			req := args.Get(1).(*artifactsIdl.GetArtifactRequest)
			assert.Equal(t, "user-handle", req.GetQuery().GetArtifactId().GetArtifactKey().GetOrg())
		}).Return(&artifactsIdl.GetArtifactResponse{Artifact: &artifactsIdl.Artifact{
			ArtifactId: returnID,
			Spec: &artifactsIdl.ArtifactSpec{
				Value: one,
			},
		}}, nil)

		ctx := context.Background()

		m := ExecutionManager{
			artifactRegistry: &artifacts.ArtifactRegistry{Client: &client},
		}

		_, x, err := m.ResolveParameterMapArtifacts(ctx, &inputs, nil, "user-handle")
		assert.NoError(t, err)
		assert.Equal(t, 1, len(x))
	})
}

func TestCompleteResolvedSpec(t *testing.T) {
	existingFields := sets.NewString(
		"LaunchPlan",
		"Inputs",
		"Metadata",
		"NotificationOverrides",
		"Labels",
		"Annotations",
		"SecurityContext",
		"AuthRole",
		"QualityOfService",
		"MaxParallelism",
		"RawOutputDataConfig",
		"ClusterAssignment",
		"Interruptible",
		"OverwriteCache",
		"Envs",
		"Tags",
		"ExecutionClusterLabel",
		"ExecutionEnvAssignments",
		"TaskResourceAttributes",
		"SubNodeIds",
	)
	specType := reflect.ValueOf(admin.ExecutionSpec{}).Type()
	for i := 0; i < specType.NumField(); i++ {
		fieldName := specType.Field(i).Name
		if fieldName == "state" || fieldName == "sizeCache" || fieldName == "unknownFields" {
			continue
		}
		if !existingFields.Has(fieldName) {
			t.Fatalf("This is a warning test. You are adding a new field [%s] to ExecutionSpec.proto, be sure to also modify completeResolvedSpec function in execution_manager.go", fieldName)
		}
	}
}

func TestValidateActiveExecutions(t *testing.T) {
	t.Run("no active executions limit", func(t *testing.T) {
		r := getDefaultPluginRegistry()

		execManager := ExecutionManager{
			pluginRegistry: r,
		}
		err := execManager.validateActiveExecutions(context.TODO())

		assert.NoError(t, err)
	})

	t.Run("active executions exhausted", func(t *testing.T) {
		getUserProperties := func(ctx context.Context) shared.UserProperties {
			return shared.UserProperties{
				Org:              orgValue,
				ActiveExecutions: 5,
			}
		}

		r := getDefaultPluginRegistry()
		r.RegisterDefault(plugins.PluginIDUserProperties, getUserProperties)

		repository := repositoryMocks.NewMockRepository()
		getExecutionCountsFunc := func(
			ctx context.Context, input interfaces.CountResourceInput) (interfaces.ExecutionCountsByPhaseOutput, error) {
			var orgFilter, updatedAtFilter bool
			for _, filter := range input.InlineFilters {
				assert.Equal(t, common.Execution, filter.GetEntity())
				queryExpr, _ := filter.GetGormQueryExpr()
				if queryExpr.Args == orgValue && queryExpr.Query == executionOrgQueryExpr {
					orgFilter = true
				}
				if queryExpr.Query == executionCreatedAtQueryExpr {
					updatedAtFilter = true
				}
			}

			assert.True(t, orgFilter, "Missing org equality filter")
			assert.True(t, updatedAtFilter, "Missing updated at filter")
			return interfaces.ExecutionCountsByPhaseOutput{
				{
					Phase: "FAILED",
					Count: int64(3),
				},
				{
					Phase: "RUNNING",
					Count: int64(7),
				},
			}, nil
		}
		repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCountByPhaseCallback(getExecutionCountsFunc)

		execManager := ExecutionManager{
			pluginRegistry: r,
			db:             repository,
			userMetrics: executionUserMetrics{
				MaxActiveExecutionsReached: mockScope.NewTestScope().MustNewCounter("max_active_executions_reached", "description"),
			},
		}

		err := execManager.validateActiveExecutions(context.TODO())

		flyteAdminErr := err.(flyteAdminErrors.FlyteAdminError)
		assert.Equal(t, codes.ResourceExhausted, flyteAdminErr.Code())
	})

	t.Run("active executions not exhausted", func(t *testing.T) {
		getUserProperties := func(ctx context.Context) shared.UserProperties {
			return shared.UserProperties{
				Org:              orgValue,
				ActiveExecutions: 5,
			}
		}

		r := getDefaultPluginRegistry()
		r.RegisterDefault(plugins.PluginIDUserProperties, getUserProperties)

		repository := repositoryMocks.NewMockRepository()
		getExecutionCountsFunc := func(
			ctx context.Context, input interfaces.CountResourceInput) (interfaces.ExecutionCountsByPhaseOutput, error) {
			var orgFilter, updatedAtFilter bool
			for _, filter := range input.InlineFilters {
				assert.Equal(t, common.Execution, filter.GetEntity())
				queryExpr, _ := filter.GetGormQueryExpr()
				if queryExpr.Args == orgValue && queryExpr.Query == executionOrgQueryExpr {
					orgFilter = true
				}
				if queryExpr.Query == executionCreatedAtQueryExpr {
					updatedAtFilter = true
				}
			}
			assert.True(t, orgFilter, "Missing org equality filter")
			assert.True(t, updatedAtFilter, "Missing updated at filter")
			return interfaces.ExecutionCountsByPhaseOutput{
				{
					Phase: "FAILED",
					Count: int64(6),
				},
				{
					Phase: "RUNNING",
					Count: int64(4),
				},
			}, nil
		}
		repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetCountByPhaseCallback(getExecutionCountsFunc)

		execManager := ExecutionManager{
			pluginRegistry: r,
			db:             repository,
		}

		err := execManager.validateActiveExecutions(context.TODO())

		assert.NoError(t, err)
	})

}
