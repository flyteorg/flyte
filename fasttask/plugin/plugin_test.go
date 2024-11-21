package plugin

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	_struct "github.com/golang/protobuf/ptypes/struct"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	idlcore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/logs"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	coremocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	iomocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/tasklog"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils/secrets"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/flytestdlib/storage"

	"github.com/unionai/flyte/fasttask/plugin/mocks"
	"github.com/unionai/flyte/fasttask/plugin/pb"
)

func buildFasttaskEnvironment(t *testing.T, fastTaskExtant *pb.FastTaskEnvironment, fastTaskSpec *pb.FastTaskEnvironmentSpec) *_struct.Struct {
	executionEnv := &idlcore.ExecutionEnv{
		Name:    "foo",
		Type:    "fast-task",
		Version: "0",
	}

	if fastTaskExtant != nil {
		extant := &_struct.Struct{}
		err := utils.MarshalStruct(fastTaskExtant, extant)
		assert.Nil(t, err)
		executionEnv.Environment = &idlcore.ExecutionEnv_Extant{
			Extant: extant,
		}
	} else if fastTaskSpec != nil {
		spec := &_struct.Struct{}
		err := utils.MarshalStruct(fastTaskSpec, spec)
		assert.Nil(t, err)
		executionEnv.Environment = &idlcore.ExecutionEnv_Spec{
			Spec: spec,
		}
	}

	executionEnvStruct := &_struct.Struct{}
	err := utils.MarshalStruct(executionEnv, executionEnvStruct)
	assert.Nil(t, err)

	return executionEnvStruct
}

func getBaseFasttaskTaskTemplate(t *testing.T) *idlcore.TaskTemplate {
	executionEnv := buildFasttaskEnvironment(t, &pb.FastTaskEnvironment{
		QueueId: "foo",
	}, nil)

	executionEnvStruct := &_struct.Struct{}
	err := utils.MarshalStruct(executionEnv, executionEnvStruct)
	assert.Nil(t, err)

	return &idlcore.TaskTemplate{
		Custom: executionEnvStruct,
		Target: &idlcore.TaskTemplate_Container{
			Container: &idlcore.Container{
				Command: []string{""},
				Args:    []string{},
			},
		},
	}
}

func getBaseFasttaskTaskTemplateK8SPod(t *testing.T) *idlcore.TaskTemplate {
	executionEnv := buildFasttaskEnvironment(t, &pb.FastTaskEnvironment{
		QueueId: "foo",
	}, nil)

	executionEnvStruct := &_struct.Struct{}
	err := utils.MarshalStruct(executionEnv, executionEnvStruct)
	assert.Nil(t, err)

	metadata := &idlcore.K8SObjectMetadata{
		Labels: map[string]string{
			"l": "a",
		},
		Annotations: map[string]string{
			"a": "b",
		},
	}

	podSpec := v1.PodSpec{
		Containers: []v1.Container{
			{
				Name:    "primary",
				Command: []string{""},
				Args:    []string{},
			},
		},
	}
	podSpecStruct, err := utils.MarshalObjToStruct(podSpec)
	if err != nil {
		t.Fatal(err)
	}

	return &idlcore.TaskTemplate{
		Custom: executionEnvStruct,
		Target: &idlcore.TaskTemplate_K8SPod{
			K8SPod: &idlcore.K8SPod{
				Metadata: metadata,
				PodSpec:  podSpecStruct,
			},
		},
		Config: map[string]string{
			flytek8s.PrimaryContainerKey: "primary",
		},
	}
}

func TestFinalize(t *testing.T) {
	ctx := context.TODO()
	scope := promutils.NewTestScope()

	// initialize fasttask TaskTemplate
	taskTemplate := getBaseFasttaskTaskTemplate(t)
	taskReader := &coremocks.TaskReader{}
	taskReader.OnReadMatch(mock.Anything).Return(taskTemplate, nil)

	// initialize static execution context attributes
	taskMetadata := &coremocks.TaskExecutionMetadata{}
	taskExecutionID := &coremocks.TaskExecutionID{}
	taskExecutionID.OnGetGeneratedNameWithMatch(mock.Anything, mock.Anything).Return("task-id", nil)
	taskMetadata.OnGetTaskExecutionID().Return(taskExecutionID)

	// create TaskExecutionContext
	tCtx := &coremocks.TaskExecutionContext{}
	tCtx.OnTaskExecutionMetadata().Return(taskMetadata)
	tCtx.OnTaskReader().Return(taskReader)

	arrayNodeStateInput := &State{
		SubmissionPhase: Submitted,
		WorkerID:        "w0",
	}
	pluginStateReader := &coremocks.PluginStateReader{}
	pluginStateReader.On("Get", mock.Anything).Return(
		func(v interface{}) uint8 {
			*v.(*State) = *arrayNodeStateInput
			return 0
		},
		nil,
	)
	tCtx.OnPluginStateReader().Return(pluginStateReader)

	// create FastTaskService mock
	fastTaskService := &mocks.FastTaskService{}
	fastTaskService.OnCleanup(ctx, "task-id", "foo", "w0").Return(nil)

	// initialize plugin
	plugin := &Plugin{
		fastTaskService: fastTaskService,
		metrics:         newPluginMetrics(scope),
	}

	// call handle
	err := plugin.Finalize(ctx, tCtx)
	assert.Nil(t, err)
}

func TestGetExecutionEnv(t *testing.T) {
	ctx := context.TODO()
	tCtx := &coremocks.TaskExecutionContext{}
	tCtx.OnTaskReader().Return(&coremocks.TaskReader{})

	executionEnvID := core.ExecutionEnvID{
		Project: "project",
		Domain:  "domain",
		Name:    "foo",
		Version: "0",
	}

	expectedExtant := &pb.FastTaskEnvironment{
		QueueId: executionEnvID.String(),
	}
	expectedExtantStruct := &_struct.Struct{}
	err := utils.MarshalStruct(expectedExtant, expectedExtantStruct)
	assert.Nil(t, err)

	toFastTaskSpec := func(spec *pb.FastTaskEnvironmentSpec) *structpb.Struct {
		specStruct := &_struct.Struct{}
		err := utils.MarshalStruct(spec, specStruct)
		assert.Nil(t, err)
		return specStruct
	}

	podTemplateSpec := &v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
			},
			Labels: map[string]string{
				"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
			},
			Namespace: "test-namespace",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Command: []string{"bar"},
				},
			},
		},
	}
	podTemplateSpecBytes, err := json.Marshal(podTemplateSpec)
	assert.Nil(t, err)

	tests := []struct {
		name                     string
		fastTaskExtant           *pb.FastTaskEnvironment
		fastTaskSpec             *pb.FastTaskEnvironmentSpec
		clientGetExists          bool
		createExectionEnvMatcher interface{} // func (environmentSpec *structpb.Struct) bool
	}{
		{
			name: "ExecutionExtant",
			fastTaskExtant: &pb.FastTaskEnvironment{
				QueueId: executionEnvID.String(),
			},
		},
		{
			name:                     "ExecutionSpecExists",
			fastTaskSpec:             &pb.FastTaskEnvironmentSpec{},
			clientGetExists:          true,
			createExectionEnvMatcher: expectedExtantStruct,
		},
		{
			name: "ExecutionSpecCreate",
			fastTaskSpec: &pb.FastTaskEnvironmentSpec{
				PodTemplateSpec: podTemplateSpecBytes,
			},
			clientGetExists: false,
			createExectionEnvMatcher: toFastTaskSpec(
				&pb.FastTaskEnvironmentSpec{
					PodTemplateSpec: podTemplateSpecBytes,
				},
			),
		},
		{
			name:            "ExecutionSpecInjectPodTemplateAndCreate",
			fastTaskSpec:    &pb.FastTaskEnvironmentSpec{},
			clientGetExists: false,
			createExectionEnvMatcher: mock.MatchedBy(func(environmentSpec *structpb.Struct) bool {
				spec := &pb.FastTaskEnvironmentSpec{}
				err := utils.UnmarshalStruct(environmentSpec, spec)
				assert.Nil(t, err)
				var podTemplateSpec v1.PodTemplateSpec
				err = json.Unmarshal(spec.GetPodTemplateSpec(), &podTemplateSpec)
				assert.Nil(t, err)
				return podTemplateSpec.Namespace == "test-namespace" && spec.GetPrimaryContainerName() == "task-id"
			}),
		},
	}

	// initialize static execution context attributes
	inputReader := &iomocks.InputReader{}
	inputReader.OnGetInputPrefixPath().Return("test-data-prefix")
	inputReader.OnGetInputPath().Return("test-data-reference")
	inputReader.OnGetMatch(mock.Anything).Return(&idlcore.LiteralMap{}, nil)

	outputReader := &iomocks.OutputWriter{}
	outputReader.OnGetOutputPath().Return("/data/outputs.pb")
	outputReader.OnGetOutputPrefixPath().Return("/data/")
	outputReader.OnGetRawOutputPrefix().Return("")
	outputReader.OnGetCheckpointPrefix().Return("/checkpoint")
	outputReader.OnGetPreviousCheckpointsPrefix().Return("/prev")

	taskMetadata := &coremocks.TaskExecutionMetadata{}
	taskMetadata.OnGetAnnotations().Return(map[string]string{})
	taskMetadata.OnGetEnvironmentVariables().Return(nil)
	taskMetadata.OnGetLabels().Return(map[string]string{})
	taskMetadata.OnGetK8sServiceAccount().Return("service-account")
	taskMetadata.OnGetNamespace().Return("test-namespace")
	taskMetadata.OnGetPlatformResources().Return(&v1.ResourceRequirements{})
	taskMetadata.OnGetSecurityContext().Return(idlcore.SecurityContext{})
	taskMetadata.OnIsInterruptible().Return(true)

	taskExecutionID := &coremocks.TaskExecutionID{}
	taskExecutionID.OnGetIDMatch().Return(idlcore.TaskExecutionIdentifier{
		NodeExecutionId: &idlcore.NodeExecutionIdentifier{
			ExecutionId: &idlcore.WorkflowExecutionIdentifier{
				Name:    "my_name",
				Project: "my_project",
				Domain:  "my_domain",
			},
		},
		TaskId: &idlcore.Identifier{
			Project: "project",
			Domain:  "domain",
		},
	})
	taskExecutionID.OnGetGeneratedNameMatch().Return("task-id")
	taskMetadata.OnGetTaskExecutionID().Return(taskExecutionID)
	taskExecutionID.OnGetGeneratedNameWithMatch(mock.Anything, mock.Anything).Return("task-id", nil)
	taskMetadata.OnGetTaskExecutionID().Return(taskExecutionID)

	taskOverrides := &coremocks.TaskOverrides{}
	taskOverrides.OnGetResourcesMatch().Return(&v1.ResourceRequirements{})
	taskOverrides.OnGetExtendedResourcesMatch().Return(nil)
	taskOverrides.OnGetContainerImageMatch().Return("")
	taskMetadata.OnGetOverridesMatch().Return(taskOverrides)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scope := promutils.NewTestScope()

			// initialize fasttask TaskTemplate
			executionEnvStruct := buildFasttaskEnvironment(t, test.fastTaskExtant, test.fastTaskSpec)
			taskTemplate := &idlcore.TaskTemplate{
				Custom: executionEnvStruct,
				Target: &idlcore.TaskTemplate_Container{
					Container: &idlcore.Container{
						Command: []string{""},
						Args:    []string{},
					},
				},
				Config: map[string]string{
					flytek8s.PrimaryContainerKey: "primary",
				},
			}

			// create ExecutionEnvClient mock
			executionEnvClient := &coremocks.ExecutionEnvClient{}
			if test.clientGetExists {
				executionEnvClient.OnGetMatch(ctx, mock.Anything).Return(expectedExtantStruct)
			} else {
				executionEnvClient.OnGetMatch(ctx, mock.Anything).Return(nil)
			}
			executionEnvClient.OnCreateMatch(ctx, executionEnvID, test.createExectionEnvMatcher).Return(expectedExtantStruct, nil)

			// create TaskExecutionContext
			tCtx := &coremocks.TaskExecutionContext{}
			tCtx.OnInputReader().Return(inputReader)
			tCtx.OnOutputWriter().Return(outputReader)
			tCtx.OnTaskExecutionMetadata().Return(taskMetadata)

			taskReader := &coremocks.TaskReader{}
			taskReader.OnReadMatch(mock.Anything).Return(taskTemplate, nil)
			tCtx.OnTaskReader().Return(taskReader)

			tCtx.OnGetExecutionEnvClient().Return(executionEnvClient)

			// initialize plugin
			plugin := &Plugin{
				metrics: newPluginMetrics(scope),
			}

			// call handle
			_, fastTaskEnvironment, err := plugin.getExecutionEnv(ctx, tCtx)
			assert.Nil(t, err)
			assert.True(t, proto.Equal(expectedExtant, fastTaskEnvironment))
		})
	}
}

func TestAddObjectMetadata(t *testing.T) {
	ctx := context.Background()
	scope := promutils.NewTestScope()
	plugin := &Plugin{
		metrics: newPluginMetrics(scope),
	}

	taskMetadata := &coremocks.TaskExecutionMetadata{}
	taskMetadata.OnGetNamespace().Return("test-namespace")
	taskMetadata.OnGetAnnotations().Return(map[string]string{
		"metadataAnnotation": "metadataAnnotation",
	})
	taskMetadata.OnGetLabels().Return(map[string]string{
		"metadataLabel": "metadataLabel",
	})

	taskReader := &coremocks.TaskReader{}
	taskReader.OnRead(ctx).Return(&idlcore.TaskTemplate{
		SecurityContext: &idlcore.SecurityContext{
			Secrets: []*idlcore.Secret{
				{
					Group:            "my_group",
					Key:              "my_key",
					MountRequirement: idlcore.Secret_ENV_VAR,
				},
			},
		},
	}, nil)

	tCtx := &coremocks.TaskExecutionContext{}
	tCtx.OnTaskExecutionMetadata().Return(taskMetadata)
	tCtx.OnTaskReader().Return(taskReader)

	cfg := &config.K8sPluginConfig{
		DefaultAnnotations: map[string]string{
			"defaultAnnotation": "defaultAnnotation",
		},
		DefaultLabels: map[string]string{
			"defaultLabel": "defaultLabel",
		},
	}

	spec := &v1.PodTemplateSpec{}
	err := plugin.addObjectMetadata(ctx, tCtx, spec, cfg)

	assert.Nil(t, err)
	assert.Equal(t, map[string]string{
		"defaultAnnotation":  "defaultAnnotation",
		"metadataAnnotation": "metadataAnnotation",
		"flyte.secrets/s0":   "m4zg54lqhiqce2lzl4txe22voarau12fpe4caitnpfpwwzlzeifg122vnz1f53tfof1ws3tfnvsw34b1ebcu3vs6kzavecq",
	}, spec.GetAnnotations())
	assert.Equal(t, map[string]string{
		secrets.PodLabel: secrets.PodLabelValue,
		"defaultLabel":   "defaultLabel",
		"metadataLabel":  "metadataLabel",
	}, spec.GetLabels())
	assert.Equal(t, "test-namespace", spec.GetNamespace())
	assert.Len(t, spec.GetOwnerReferences(), 0)
	assert.Len(t, spec.GetFinalizers(), 0)
}

func TestHandleNotYetStarted(t *testing.T) {
	ctx := context.TODO()
	tests := []struct {
		name               string
		workerID           string
		lastUpdated        time.Time
		executionEnvStatus map[string]*v1.Pod
		expectedPhase      core.Phase
		expectedReason     string
		expectedError      error
	}{
		{
			name:     "NoWorkersAvailable",
			workerID: "",
			executionEnvStatus: map[string]*v1.Pod{
				"foo": nil,
			},
			expectedPhase:  core.PhaseWaitingForResources,
			expectedReason: "no workers available",
			expectedError:  nil,
		},
		{
			name:     "NoWorkersAllFailed",
			workerID: "",
			executionEnvStatus: map[string]*v1.Pod{
				"foo": &v1.Pod{
					Status: v1.PodStatus{
						Phase: v1.PodFailed,
					},
				},
			},
			expectedPhase:  core.PhasePermanentFailure,
			expectedReason: "",
			expectedError:  nil,
		},
		{
			name:           "AssignedToWorker",
			workerID:       "w0",
			expectedPhase:  core.PhaseQueued,
			expectedReason: "task offered to worker w0",
			expectedError:  nil,
		},
	}

	// initialize fasttask TaskTemplate
	taskTemplateTests := []*idlcore.TaskTemplate{getBaseFasttaskTaskTemplate(t), getBaseFasttaskTaskTemplateK8SPod(t)}

	// initialize static execution context attributes
	inputReader := &iomocks.InputReader{}
	inputReader.OnGetInputPrefixPath().Return("test-data-prefix")
	inputReader.OnGetInputPath().Return("test-data-reference")
	inputReader.OnGetMatch(mock.Anything).Return(&idlcore.LiteralMap{}, nil)

	outputReader := &iomocks.OutputWriter{}
	outputReader.OnGetOutputPath().Return("/data/outputs.pb")
	outputReader.OnGetOutputPrefixPath().Return("/data/")
	outputReader.OnGetRawOutputPrefix().Return("")
	outputReader.OnGetCheckpointPrefix().Return("/checkpoint")
	outputReader.OnGetPreviousCheckpointsPrefix().Return("/prev")

	taskMetadata := &coremocks.TaskExecutionMetadata{}
	taskMetadata.OnGetOwnerIDMatch().Return(types.NamespacedName{
		Namespace: "namespace",
		Name:      "execution_id",
	})
	envVars := map[string]string{"TEST": "VALUE"}
	taskMetadata.OnGetEnvironmentVariables().Return(envVars)
	taskExecutionID := &coremocks.TaskExecutionID{}
	taskExecutionID.OnGetID().Return(idlcore.TaskExecutionIdentifier{
		TaskId: &idlcore.Identifier{
			Project: "project",
			Domain:  "domain",
		},
	})
	taskExecutionID.OnGetGeneratedNameWithMatch(mock.Anything, mock.Anything).Return("task-id", nil)
	taskMetadata.OnGetTaskExecutionID().Return(taskExecutionID)

	for _, test := range tests {
		for _, taskTemplate := range taskTemplateTests {
			t.Run(test.name, func(t *testing.T) {
				scope := promutils.NewTestScope()

				// initialize fasttask TaskTemplate
				taskReader := &coremocks.TaskReader{}
				taskReader.OnReadMatch(mock.Anything).Return(taskTemplate, nil)

				// create TaskExecutionContext
				tCtx := &coremocks.TaskExecutionContext{}
				tCtx.OnInputReader().Return(inputReader)
				tCtx.OnOutputWriter().Return(outputReader)
				tCtx.OnTaskExecutionMetadata().Return(taskMetadata)
				tCtx.OnTaskReader().Return(taskReader)

				executionEnvClient := &coremocks.ExecutionEnvClient{}
				executionEnvClient.OnStatusMatch(ctx, mock.Anything).Return(test.executionEnvStatus, nil)
				tCtx.OnGetExecutionEnvClient().Return(executionEnvClient)

				arrayNodeStateInput := &State{
					SubmissionPhase: NotSubmitted,
					LastUpdated:     test.lastUpdated,
				}
				pluginStateReader := &coremocks.PluginStateReader{}
				pluginStateReader.On("Get", mock.Anything).Return(
					func(v interface{}) uint8 {
						*v.(*State) = *arrayNodeStateInput
						return 0
					},
					nil,
				)
				tCtx.OnPluginStateReader().Return(pluginStateReader)

				arrayNodeStateOutput := &State{}
				pluginStateWriter := &coremocks.PluginStateWriter{}
				pluginStateWriter.On("Put", mock.Anything, mock.Anything).Return(
					func(stateVersion uint8, v interface{}) error {
						*arrayNodeStateOutput = *v.(*State)
						return nil
					},
				)
				tCtx.OnPluginStateWriter().Return(pluginStateWriter)

				// create FastTaskService mock
				fastTaskService := &mocks.FastTaskService{}
				fastTaskService.OnOfferOnQueue(ctx, "foo", "task-id", "namespace", "execution_id", []string{}, envVars).Return(test.workerID, nil)

				// initialize plugin
				plugin := &Plugin{
					fastTaskService: fastTaskService,
					metrics:         newPluginMetrics(scope),
				}

				// call handle
				transition, err := plugin.Handle(ctx, tCtx)
				assert.Equal(t, test.expectedError, err)
				assert.Equal(t, test.expectedPhase, transition.Info().Phase())
				assert.Equal(t, test.expectedReason, transition.Info().Reason())
				assert.Len(t, transition.Info().Info().Logs, 0)

				require.Len(t, transition.Info().Info().ExternalResources, 1)
				assignment := pb.FastTaskAssignment{}
				require.NoError(t, utils.UnmarshalStruct(transition.Info().Info().ExternalResources[0].CustomInfo, &assignment))
				assert.Equal(t, "", assignment.GetEnvironmentOrg())
				assert.Equal(t, "project", assignment.GetEnvironmentProject())
				assert.Equal(t, "domain", assignment.GetEnvironmentDomain())
				assert.Equal(t, "foo", assignment.GetEnvironmentName())
				assert.Equal(t, "0", assignment.GetEnvironmentVersion())
				assert.Equal(t, test.workerID, assignment.GetAssignedWorker())

				if len(test.workerID) > 0 {
					assert.Equal(t, test.workerID, arrayNodeStateOutput.WorkerID)
				}
			})
		}
	}
}

func TestHandleRunning(t *testing.T) {
	ctx := context.TODO()
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "namespace",
			Name:      "pod-name",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "pod-name",
				},
			},
			Hostname: "hostname",
		},
		Status: v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					ContainerID: "container-id",
				},
			},
		},
	}

	tests := []struct {
		name                   string
		lastUpdated            time.Time
		taskStatusPhase        core.Phase
		taskStatusReason       string
		checkStatusError       error
		executionEnvStatus     map[string]*v1.Pod
		expectedPhase          core.Phase
		expectedPhaseVersion   uint32
		expectedReason         string
		expectedError          error
		expectedLastUpdatedInc bool
		expectedLogs           bool
	}{
		{
			name:             "Running",
			lastUpdated:      time.Now().Add(-5 * time.Second),
			taskStatusPhase:  core.PhaseRunning,
			taskStatusReason: "",
			checkStatusError: nil,
			executionEnvStatus: map[string]*v1.Pod{
				"w0": nil,
			},
			expectedPhase:          core.PhaseRunning,
			expectedPhaseVersion:   1,
			expectedReason:         "",
			expectedError:          nil,
			expectedLastUpdatedInc: true,
			expectedLogs:           false,
		},
		{
			name:             "RunningWithLogs",
			lastUpdated:      time.Now().Add(-5 * time.Second),
			taskStatusPhase:  core.PhaseRunning,
			taskStatusReason: "",
			checkStatusError: nil,
			executionEnvStatus: map[string]*v1.Pod{
				"w0": pod,
			},
			expectedPhase:          core.PhaseRunning,
			expectedPhaseVersion:   1,
			expectedReason:         "",
			expectedError:          nil,
			expectedLastUpdatedInc: true,
			expectedLogs:           true,
		},
		{
			name:             "RunningStatusNotFound",
			lastUpdated:      time.Now().Add(-5 * time.Second),
			taskStatusPhase:  core.PhaseRunning,
			taskStatusReason: "",
			checkStatusError: statusUpdateNotFoundError,
			executionEnvStatus: map[string]*v1.Pod{
				"w0": pod,
			},
			expectedPhase:          core.PhaseRunning,
			expectedReason:         "",
			expectedError:          nil,
			expectedLastUpdatedInc: false,
			expectedLogs:           true,
		},
		{
			name:             "RetryableFailure",
			lastUpdated:      time.Now().Add(-5 * time.Second),
			taskStatusPhase:  core.PhaseRetryableFailure,
			checkStatusError: nil,
			executionEnvStatus: map[string]*v1.Pod{
				"w0": pod,
			},
			expectedPhase:          core.PhaseRetryableFailure,
			expectedError:          nil,
			expectedLastUpdatedInc: false,
			expectedLogs:           true,
		},
		{
			name:             "StatusNotFoundTimeout",
			lastUpdated:      time.Now().Add(-600 * time.Second),
			taskStatusPhase:  core.PhaseUndefined,
			checkStatusError: statusUpdateNotFoundError,
			executionEnvStatus: map[string]*v1.Pod{
				"w0": pod,
			},
			expectedPhase:          core.PhaseRetryableFailure,
			expectedError:          nil,
			expectedLastUpdatedInc: false,
			expectedLogs:           true,
		},
		{
			name:             "Success",
			lastUpdated:      time.Now().Add(-5 * time.Second),
			taskStatusPhase:  core.PhaseSuccess,
			checkStatusError: nil,
			executionEnvStatus: map[string]*v1.Pod{
				"w0": pod,
			},
			expectedPhase:          core.PhaseSuccess,
			expectedError:          nil,
			expectedLastUpdatedInc: false,
			expectedLogs:           true,
		},
	}

	// initialize fasttask TaskTemplate
	taskTemplate := getBaseFasttaskTaskTemplate(t)
	taskReader := &coremocks.TaskReader{}
	taskReader.OnReadMatch(mock.Anything).Return(taskTemplate, nil)

	// initialize static execution context attributes
	taskMetadata := &coremocks.TaskExecutionMetadata{}
	taskMetadata.OnGetOwnerIDMatch().Return(types.NamespacedName{
		Namespace: "namespace",
		Name:      "execution_id",
	})
	taskExecutionID := &coremocks.TaskExecutionID{}
	taskExecutionID.OnGetGeneratedNameWithMatch(mock.Anything, mock.Anything).Return("task-id", nil)
	taskExecutionID.OnGetIDMatch().Return(idlcore.TaskExecutionIdentifier{
		NodeExecutionId: &idlcore.NodeExecutionIdentifier{
			ExecutionId: &idlcore.WorkflowExecutionIdentifier{
				Name:    "my_name",
				Project: "my_project",
				Domain:  "my_domain",
			},
		},
		TaskId: &idlcore.Identifier{
			Project: "project",
			Domain:  "domain",
		},
	})
	taskExecutionID.OnGetUniqueNodeID().Return("task-id")
	taskExecutionID.OnGetGeneratedName().Return("task-name")
	taskMetadata.OnGetTaskExecutionID().Return(taskExecutionID)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scope := promutils.NewTestScope()

			// create TaskExecutionContext
			tCtx := &coremocks.TaskExecutionContext{}
			tCtx.OnTaskExecutionMetadata().Return(taskMetadata)
			tCtx.OnTaskReader().Return(taskReader)

			dataStore, err := storage.NewDataStore(
				&storage.Config{
					Type: storage.TypeMemory,
				},
				scope,
			)
			assert.NoError(t, err)
			tCtx.OnDataStoreMatch().Return(dataStore)

			executionEnvClient := &coremocks.ExecutionEnvClient{}
			executionEnvClient.OnStatusMatch(ctx, mock.Anything).Return(test.executionEnvStatus, nil)
			tCtx.OnGetExecutionEnvClient().Return(executionEnvClient)

			outputWriter := &iomocks.OutputWriter{}
			outputWriter.OnPutMatch(ctx, mock.Anything).Return(nil)
			tCtx.OnOutputWriterMatch().Return(outputWriter)

			arrayNodeStateInput := &State{
				SubmissionPhase: Submitted,
				WorkerID:        "w0",
				LastUpdated:     test.lastUpdated,
			}
			pluginStateReader := &coremocks.PluginStateReader{}
			pluginStateReader.On("Get", mock.Anything).Return(
				func(v interface{}) uint8 {
					*v.(*State) = *arrayNodeStateInput
					return 0
				},
				nil,
			)
			tCtx.OnPluginStateReader().Return(pluginStateReader)

			arrayNodeStateOutput := &State{}
			pluginStateWriter := &coremocks.PluginStateWriter{}
			pluginStateWriter.On("Put", mock.Anything, mock.Anything).Return(
				func(stateVersion uint8, v interface{}) error {
					*arrayNodeStateOutput = *v.(*State)
					return nil
				},
			)
			tCtx.OnPluginStateWriter().Return(pluginStateWriter)

			// create FastTaskService mock
			fastTaskService := &mocks.FastTaskService{}
			fastTaskService.OnCheckStatusMatch(ctx, "task-id", "foo", "w0").Return(test.taskStatusPhase, "", test.checkStatusError)

			// initialize plugin
			plugin := &Plugin{
				cfg:             defaultConfig,
				fastTaskService: fastTaskService,
				metrics:         newPluginMetrics(scope),
			}

			// call handle
			transition, err := plugin.Handle(ctx, tCtx)
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedPhase, transition.Info().Phase())
			assert.Equal(t, test.expectedPhaseVersion, arrayNodeStateOutput.PhaseVersion)

			if test.expectedLastUpdatedInc {
				assert.True(t, arrayNodeStateOutput.LastUpdated.After(test.lastUpdated))
			}

			if test.expectedLogs {
				assert.Len(t, transition.Info().Info().Logs, 1)
			} else {
				assert.Len(t, transition.Info().Info().Logs, 0)
			}

			require.Len(t, transition.Info().Info().ExternalResources, 1)
			assignment := pb.FastTaskAssignment{}
			require.NoError(t, utils.UnmarshalStruct(transition.Info().Info().ExternalResources[0].CustomInfo, &assignment))
			assert.Equal(t, "", assignment.GetEnvironmentOrg())
			assert.Equal(t, "project", assignment.GetEnvironmentProject())
			assert.Equal(t, "domain", assignment.GetEnvironmentDomain())
			assert.Equal(t, "foo", assignment.GetEnvironmentName())
			assert.Equal(t, "0", assignment.GetEnvironmentVersion())
			assert.Equal(t, "w0", assignment.GetAssignedWorker())
		})
	}
}

func TestGetTaskInfo(t *testing.T) {
	ctx := context.TODO()

	now := time.Now()
	start := now.Add(-5 * time.Second)
	executionEnv := &idlcore.ExecutionEnv{
		Name:    "foo",
		Version: "0",
	}
	queueID := "foo"
	workerID := "w0"

	plugin := &Plugin{
		cfg: &Config{
			Logs: logs.LogConfig{
				Templates: []tasklog.TemplateLogPlugin{
					{
						DisplayName:  "Custom Logs",
						TemplateURIs: []string{"http://foo.com/pod={{ .namespace }}/{{ .podName }}"},
					},
				},
			},
		},
	}

	expectedLogCtx := &idlcore.LogContext{
		Pods: []*idlcore.PodLogContext{
			{
				Namespace:            "namespace",
				PodName:              "pod-name",
				PrimaryContainerName: "pod-name",
				Containers: []*idlcore.ContainerContext{
					{
						ContainerName: "pod-name",
						Process: &idlcore.ContainerContext_ProcessContext{
							ContainerStartTime: timestamppb.New(start),
							ContainerEndTime:   timestamppb.New(now),
						},
					},
				},
			},
		},
		PrimaryPodName: "pod-name",
	}

	t.Run("available", func(t *testing.T) {
		executionEnvClient := &coremocks.ExecutionEnvClient{}
		executionEnvClient.OnStatusMatch(ctx, mock.Anything).Return(map[string]*v1.Pod{
			"w0": {
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "namespace",
					Name:      "pod-name",
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "pod-name",
						},
					},
					Hostname: "hostname",
				},
				Status: v1.PodStatus{
					ContainerStatuses: []v1.ContainerStatus{
						{
							ContainerID: "container-id",
						},
					},
				},
			},
		}, nil)
		tCtx := &coremocks.TaskExecutionContext{}
		tCtx.OnGetExecutionEnvClient().Return(executionEnvClient)

		taskMetadata := &coremocks.TaskExecutionMetadata{}
		taskMetadata.OnGetOwnerIDMatch().Return(types.NamespacedName{
			Namespace: "namespace",
			Name:      "execution_id",
		})
		taskExecutionID := &coremocks.TaskExecutionID{}
		taskExecutionID.OnGetID().Return(idlcore.TaskExecutionIdentifier{
			TaskId: &idlcore.Identifier{
				Project: "project",
				Domain:  "domain",
			},
		})
		taskExecutionID.OnGetGeneratedNameWithMatch(mock.Anything, mock.Anything).Return("task-id", nil)
		taskExecutionID.OnGetUniqueNodeID().Return("task-id")
		taskExecutionID.OnGetGeneratedName().Return("task-name")
		taskMetadata.OnGetTaskExecutionID().Return(taskExecutionID)
		tCtx.OnTaskExecutionMetadata().Return(taskMetadata)

		taskTemplate := getBaseFasttaskTaskTemplate(t)
		taskReader := &coremocks.TaskReader{}
		taskReader.OnReadMatch(mock.Anything).Return(taskTemplate, nil)
		tCtx.OnTaskReader().Return(taskReader)

		taskInfo, err := plugin.getTaskInfo(ctx, tCtx, start, now, executionEnv, queueID, workerID)

		require.Nil(t, err)
		require.Len(t, taskInfo.Logs, 1)
		assert.Equal(t, "Custom Logs", taskInfo.Logs[0].GetName())
		assert.Equal(t, "http://foo.com/pod=namespace/pod-name", taskInfo.Logs[0].GetUri())
		assert.Equal(t, expectedLogCtx, taskInfo.LogContext)
	})

	t.Run("mismatched container name", func(t *testing.T) {
		executionEnvClient := &coremocks.ExecutionEnvClient{}
		executionEnvClient.OnStatusMatch(ctx, mock.Anything).Return(map[string]*v1.Pod{
			"w0": {
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "namespace",
					Name:      "pod-name",
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "container-name",
						},
					},
					Hostname: "hostname",
				},
				Status: v1.PodStatus{},
			},
		}, nil)
		tCtx := &coremocks.TaskExecutionContext{}
		tCtx.OnGetExecutionEnvClient().Return(executionEnvClient)

		taskMetadata := &coremocks.TaskExecutionMetadata{}
		taskMetadata.OnGetOwnerIDMatch().Return(types.NamespacedName{
			Namespace: "namespace",
			Name:      "execution_id",
		})
		taskExecutionID := &coremocks.TaskExecutionID{}
		taskExecutionID.OnGetID().Return(idlcore.TaskExecutionIdentifier{
			TaskId: &idlcore.Identifier{
				Project: "project",
				Domain:  "domain",
			},
		})
		taskExecutionID.OnGetGeneratedNameWithMatch(mock.Anything, mock.Anything).Return("task-id", nil)
		taskExecutionID.OnGetUniqueNodeID().Return("task-id")
		taskExecutionID.OnGetGeneratedName().Return("task-name")
		taskMetadata.OnGetTaskExecutionID().Return(taskExecutionID)
		tCtx.OnTaskExecutionMetadata().Return(taskMetadata)

		taskInfo, err := plugin.getTaskInfo(ctx, tCtx, start, now, executionEnv, queueID, workerID)

		require.Nil(t, err)
		require.Empty(t, taskInfo.Logs)
		assert.Nil(t, taskInfo.LogContext)
	})

	t.Run("no container id", func(t *testing.T) {
		executionEnvClient := &coremocks.ExecutionEnvClient{}
		executionEnvClient.OnStatusMatch(ctx, mock.Anything).Return(map[string]*v1.Pod{
			"w0": {
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "namespace",
					Name:      "pod-name",
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "pod-name",
						},
					},
					Hostname: "hostname",
				},
				Status: v1.PodStatus{},
			},
		}, nil)
		tCtx := &coremocks.TaskExecutionContext{}
		tCtx.OnGetExecutionEnvClient().Return(executionEnvClient)

		taskMetadata := &coremocks.TaskExecutionMetadata{}
		taskMetadata.OnGetOwnerIDMatch().Return(types.NamespacedName{
			Namespace: "namespace",
			Name:      "execution_id",
		})
		taskExecutionID := &coremocks.TaskExecutionID{}
		taskExecutionID.OnGetID().Return(idlcore.TaskExecutionIdentifier{
			TaskId: &idlcore.Identifier{
				Project: "project",
				Domain:  "domain",
			},
		})
		taskExecutionID.OnGetGeneratedNameWithMatch(mock.Anything, mock.Anything).Return("task-id", nil)
		taskExecutionID.OnGetUniqueNodeID().Return("task-id")
		taskExecutionID.OnGetGeneratedName().Return("task-name")
		taskMetadata.OnGetTaskExecutionID().Return(taskExecutionID)
		tCtx.OnTaskExecutionMetadata().Return(taskMetadata)

		taskInfo, err := plugin.getTaskInfo(ctx, tCtx, start, now, executionEnv, queueID, workerID)

		require.Nil(t, err)
		require.Empty(t, taskInfo.Logs)
		assert.Nil(t, taskInfo.LogContext)
	})
}

func init() {
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey,
		contextutils.TaskIDKey)
}
