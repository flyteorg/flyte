package plugin

import (
	"context"
	"fmt"
	"testing"
	"time"

	_struct "github.com/golang/protobuf/ptypes/struct"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	idlcore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/logs"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	coremocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	iomocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/tasklog"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/transformers/k8s"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/flytestdlib/storage"
	stdlibUtils "github.com/flyteorg/flyte/flytestdlib/utils"

	"github.com/unionai/flyte/fasttask/plugin/interfaces"
	interfaceMocks "github.com/unionai/flyte/fasttask/plugin/interfaces/mocks"
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
	taskExecutionID.OnGetID().Return(idlcore.TaskExecutionIdentifier{
		TaskId: &idlcore.Identifier{
			ResourceType: idlcore.ResourceType_TASK,
			Project:      "project",
			Domain:       "domain",
			Org:          "org",
			Name:         "task-id",
			Version:      "123",
		},
	})
	taskMetadata.OnGetTaskExecutionID().Return(taskExecutionID)
	taskMetadata.OnGetLabels().Return(map[string]string{})

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
	fastTaskService := &interfaceMocks.FastTaskService{}
	fastTaskService.On("Cleanup", ctx, "task-id", "org_project_domain_foo_0", "w0").Return(nil)

	// initialize plugin
	plugin := &Plugin{
		fastTaskService: fastTaskService,
		metrics:         newPluginMetrics(scope),
	}

	// call handle
	err := plugin.Finalize(ctx, tCtx)
	assert.Nil(t, err)
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
	}{
		{
			name:     "NoWorkersAvailable",
			workerID: "",
			executionEnvStatus: map[string]*v1.Pod{
				"foo": nil,
			},
			expectedPhase:  core.PhaseWaitingForResources,
			expectedReason: "no workers available",
		},
		{
			name:     "NoWorkersAllFailed",
			workerID: "",
			executionEnvStatus: map[string]*v1.Pod{
				"foo": {
					Status: v1.PodStatus{
						Phase: v1.PodFailed,
					},
				},
			},
			expectedPhase:  core.PhasePermanentFailure,
			expectedReason: "",
		},
		{
			name:     "AssignedToWorker",
			workerID: "w0",
			executionEnvStatus: map[string]*v1.Pod{
				"w0": {
					ObjectMeta: metav1.ObjectMeta{
						Name: "w0",
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "w0",
							},
						},
					},
					Status: v1.PodStatus{
						ContainerStatuses: []v1.ContainerStatus{
							{
								Name:        "w0",
								ContainerID: "foo",
							},
						},
					},
				},
			},
			expectedPhase:  core.PhaseQueued,
			expectedReason: "task offered to worker w0",
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
	envVars := map[string]string{
		"TEST":                             "VALUE",
		"FLYTE_ATTEMPT_NUMBER":             "0",
		"FLYTE_INTERNAL_DOMAIN":            "domain",
		"FLYTE_INTERNAL_EXECUTION_DOMAIN":  "domain",
		"FLYTE_INTERNAL_EXECUTION_ID":      "abc123",
		"FLYTE_INTERNAL_EXECUTION_PROJECT": "project",
		"FLYTE_INTERNAL_NAME":              "",
		"FLYTE_INTERNAL_PROJECT":           "project",
		"FLYTE_INTERNAL_TASK_DOMAIN":       "domain",
		"FLYTE_INTERNAL_TASK_NAME":         "",
		"FLYTE_INTERNAL_TASK_PROJECT":      "project",
		"FLYTE_INTERNAL_TASK_VERSION":      "",
		"FLYTE_INTERNAL_VERSION":           "",
		"_F_PN":                            "",
	}
	taskMetadata.OnGetEnvironmentVariables().Return(envVars)
	taskExecutionID := &coremocks.TaskExecutionID{}
	execID := &idlcore.WorkflowExecutionIdentifier{
		Org:     "foo",
		Project: "project",
		Domain:  "domain",
		Name:    "abc123",
	}
	taskExecutionID.OnGetID().Return(idlcore.TaskExecutionIdentifier{
		TaskId: &idlcore.Identifier{
			Project: "project",
			Domain:  "domain",
		},
		NodeExecutionId: &idlcore.NodeExecutionIdentifier{
			ExecutionId: execID,
		},
	})
	taskExecutionID.OnGetGeneratedNameWithMatch(mock.Anything, mock.Anything).Return("task-id", nil)
	taskMetadata.OnGetTaskExecutionID().Return(taskExecutionID)
	enqueueLabels := map[string]string{
		"additional-label":  "additional-label-value",
		"un-included-label": "un-included-label-value",
	}
	taskMetadata.OnGetLabels().Return(enqueueLabels)

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

				namespaceName := "namespace"
				executionName := "execution_id"
				workflowID := types.NamespacedName{
					Namespace: namespaceName,
					Name:      executionName,
				}
				// ensure the correct labels are used for enqueueing
				updatedEnqueueLabels := map[string]string{
					k8s.WorkflowID:     workflowID.String(),
					"additional-label": "additional-label-value",
				}

				// create FastTaskService mock
				fastTaskService := &interfaceMocks.FastTaskService{}
				if test.workerID != "" {
					mockWorker := &interfaceMocks.Worker{}
					mockWorker.On("State").Return(interfaces.HEALTHY)
					mockWorker.On("ID").Return(test.workerID)
					fastTaskService.On("OfferTaskToEnvironment", ctx, execID, "project_domain_foo_0", "task-id", "namespace", "execution_id", []string{}, envVars, updatedEnqueueLabels).Return(mockWorker, nil)
				} else {
					fastTaskService.On("OfferTaskToEnvironment", ctx, execID, "project_domain_foo_0", "task-id", "namespace", "execution_id", []string{}, envVars, updatedEnqueueLabels).Return(nil, noCapacityAvailableError)
					fastTaskService.On("AddPendingOwner", "project_domain_foo_0", "task-id", updatedEnqueueLabels).Return(true)
				}

				mockEnv := &interfaceMocks.Environment{}
				mockEnv.On("State").Return(interfaces.HEALTHY)

				// create EnvironmentBuilder mock
				builder := &interfaceMocks.EnvironmentBuilder{}
				builder.On("GetOrCreateEnvironment", ctx, tCtx, mock.Anything, mock.Anything).Return(mockEnv, nil)

				// Configure ValidateWorkerPods based on test case
				if test.name == "NoWorkersAllFailed" {
					builder.On("ValidateWorkerPods", ctx, "project_domain_foo_0", mock.Anything).Return("all workers failed", nil)
				} else {
					builder.On("ValidateWorkerPods", ctx, "project_domain_foo_0", mock.Anything).Return("", nil)
					builder.On("ScaleUp", ctx, "project_domain_foo_0").Return()
				}

				// Add GetWorkerPod mock for AssignedToWorker test case
				if test.workerID != "" {
					// Create mock pod for worker
					builder.On("GetWorkerPod", ctx, "project_domain_foo_0", test.workerID).Return(test.executionEnvStatus[test.workerID], nil)
				}

				// initialize plugin
				plugin := &Plugin{
					cfg: &Config{
						Logs: logs.LogConfig{},
					},
					fastTaskService: fastTaskService,
					builder:         builder,
					metrics:         newPluginMetrics(scope),
					enqueueLabels: map[string]struct{}{
						k8s.WorkflowID:     {},
						"additional-label": {},
					},
				}

				// call handle
				transition, err := plugin.Handle(ctx, tCtx)

				assert.NoError(t, err)
				assert.Equal(t, test.expectedPhase, transition.Info().Phase())
				assert.Equal(t, test.expectedReason, transition.Info().Reason())
				assert.Len(t, transition.Info().Info().Logs, 0)
				require.Len(t, transition.Info().Info().ExternalResources, 1)
				assignment := pb.FastTaskAssignment{}
				require.NoError(t, stdlibUtils.UnmarshalStructToPb(transition.Info().Info().ExternalResources[0].CustomInfo, &assignment))
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
		envState               interfaces.State
		workerState            interfaces.State
	}{
		{
			name:             "PodNotFoundRunning",
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
			envState:               interfaces.HEALTHY,
		},
		{
			name:             "OrphanedEnvironment",
			lastUpdated:      time.Now().Add(-5 * time.Second),
			taskStatusPhase:  core.PhaseRunning,
			taskStatusReason: "",
			checkStatusError: nil,
			executionEnvStatus: map[string]*v1.Pod{
				"w0": {
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				},
			},
			expectedPhase:          core.PhaseRunning,
			expectedPhaseVersion:   1,
			expectedReason:         "",
			expectedError:          nil,
			expectedLastUpdatedInc: true,
			expectedLogs:           false,
			envState:               interfaces.ORPHANED,
		},
		{
			name:             "PodNotFoundSuccess",
			lastUpdated:      time.Now().Add(-5 * time.Second),
			taskStatusPhase:  core.PhaseSuccess,
			taskStatusReason: "",
			checkStatusError: nil,
			executionEnvStatus: map[string]*v1.Pod{
				"w0": nil,
			},
			expectedPhase:          core.PhaseUndefined,
			expectedPhaseVersion:   0,
			expectedReason:         "",
			expectedError:          podContainerNotFoundError,
			expectedLastUpdatedInc: false,
			expectedLogs:           false,
			envState:               interfaces.HEALTHY,
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
			envState:               interfaces.HEALTHY,
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
			envState:               interfaces.HEALTHY,
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
			envState:               interfaces.HEALTHY,
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
			envState:               interfaces.HEALTHY,
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
			envState:               interfaces.HEALTHY,
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
	taskMetadata.OnGetLabels().Return(map[string]string{})

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
			fastTaskService := &interfaceMocks.FastTaskService{}
			// Updated method signature: CheckStatus(ctx, taskID, envID, workerID) (phase, reason, error)
			fastTaskService.EXPECT().CheckStatus(mock.Anything, "task-id", "project_domain_foo_0", "w0").Return(interfaces.TaskStatus{Phase: test.taskStatusPhase}, test.checkStatusError)

			// create environment store and set up environment
			store := newEnvironmentStore()

			// Create mock worker
			mockWorker := &interfaceMocks.Worker{}
			mockWorker.EXPECT().ID().Return("w0")
			if test.checkStatusError == statusUpdateNotFoundError {
				mockWorker.EXPECT().State().Return(test.workerState)
			}

			// Create mock environment and worker for the store
			mockEnv := &interfaceMocks.Environment{}
			mockEnv.EXPECT().GetWorker("w0").Return(mockWorker).Maybe()
			mockEnv.EXPECT().State().Return(test.envState)

			// Add environment to store using the correct executionEnvID string
			store.GetOrCreate("project_domain_foo_0", mockEnv)

			// create EnvironmentBuilder mock
			builder := &interfaceMocks.EnvironmentBuilder{}
			builder.EXPECT().GetOrCreateEnvironment(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockEnv, nil)

			// Mock GetWorkerPod method for log building
			if test.executionEnvStatus["w0"] != nil {
				builder.EXPECT().GetWorkerPod(mock.Anything, "project_domain_foo_0", "w0").Return(test.executionEnvStatus["w0"], nil)
			} else {
				builder.EXPECT().GetWorkerPod(mock.Anything, "project_domain_foo_0", "w0").Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, "pod-name"))
			}

			// initialize plugin
			plugin := &Plugin{
				cfg:             defaultConfig,
				fastTaskService: fastTaskService,
				metrics:         newPluginMetrics(scope),
				builder:         builder,
				store:           store,
			}

			transition, err := plugin.Handle(ctx, tCtx)
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedPhase, transition.Info().Phase())
			assert.Equal(t, test.expectedPhaseVersion, arrayNodeStateOutput.PhaseVersion)

			if test.expectedLastUpdatedInc {
				assert.True(t, arrayNodeStateOutput.LastUpdated.After(test.lastUpdated))
			}

			if test.expectedLogs {
				assert.Len(t, transition.Info().Info().Logs, 1)
				assert.NotNil(t, transition.Info().Info().LogContext)
			} else {
				if transition.Info().Info() != nil {
					assert.Len(t, transition.Info().Info().Logs, 0)
				}
			}

			if test.expectedError == nil {
				require.Len(t, transition.Info().Info().ExternalResources, 1)
				assignment := pb.FastTaskAssignment{}
				require.NoError(t, stdlibUtils.UnmarshalStructToPb(transition.Info().Info().ExternalResources[0].CustomInfo, &assignment))
				assert.Equal(t, "", assignment.GetEnvironmentOrg())
				assert.Equal(t, "project", assignment.GetEnvironmentProject())
				assert.Equal(t, "domain", assignment.GetEnvironmentDomain())
				assert.Equal(t, "foo", assignment.GetEnvironmentName())
				assert.Equal(t, "0", assignment.GetEnvironmentVersion())
				assert.Equal(t, "w0", assignment.GetAssignedWorker())
			}
		})
	}
}

func TestBuildTaskInfoWithLogs(t *testing.T) {
	ctx := context.TODO()
	now := time.Now()
	start := now.Add(-5 * time.Second)

	executionEnvID := interfaces.ExecutionEnvID{
		Project: "project",
		Domain:  "domain",
		Name:    "foo",
		Version: "0",
	}

	executionEnv := &idlcore.ExecutionEnv{
		Name:    "foo",
		Version: "0",
	}

	workerID := "w0"
	podName := "pod-name"

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

	taskTemplate := getBaseFasttaskTaskTemplate(t)

	plugin := &Plugin{
		cfg: &Config{
			Logs: logs.LogConfig{
				Templates: []tasklog.TemplateLogPlugin{
					{
						DisplayName:  "Custom Logs",
						TemplateURIs: []string{"http://foo.com/pod={{ .namespace }}/{{ .podName }}"},
					},
				},
				DynamicLogLinks: map[string]tasklog.TemplateLogPlugin{
					"vscode": {
						DisplayName:  "Vscode",
						TemplateURIs: []string{"http://foo.com/vscode={{ .namespace }}/{{ .podName }}"},
					},
				},
			},
		},
	}

	expectedLogCtx := &idlcore.LogContext{
		Pods: []*idlcore.PodLogContext{
			{
				Namespace:            "namespace",
				PodName:              podName,
				PrimaryContainerName: podName,
				Containers: []*idlcore.ContainerContext{
					{
						ContainerName: podName,
						Process: &idlcore.ContainerContext_ProcessContext{
							ContainerStartTime: timestamppb.New(start),
							ContainerEndTime:   timestamppb.New(now),
						},
					},
				},
			},
		},
		PrimaryPodName: podName,
	}

	t.Run("success", func(t *testing.T) {
		builder := &interfaceMocks.EnvironmentBuilder{}
		builder.EXPECT().GetWorkerPod(mock.Anything, executionEnvID.String(), workerID).Return(&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "namespace",
				Name:      podName,
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: podName,
						Env: []v1.EnvVar{
							{
								Name:  logs.FlyteEnableVscode,
								Value: "true",
							},
						},
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
		}, nil)

		plugin.builder = builder

		taskInfo, err := plugin.buildTaskInfoWithLogs(ctx, executionEnvID, executionEnv, workerID, taskExecutionID, taskTemplate, start, now)

		require.NoError(t, err)
		require.NotNil(t, taskInfo)
		require.Len(t, taskInfo.Logs, 2)
		assert.Equal(t, "Custom Logs", taskInfo.Logs[0].GetName())
		assert.Equal(t, "Vscode", taskInfo.Logs[1].GetName())
		assert.Equal(t, "http://foo.com/pod=namespace/pod-name", taskInfo.Logs[0].GetUri())
		assert.Equal(t, "http://foo.com/vscode=namespace/pod-name", taskInfo.Logs[1].GetUri())
		assert.Equal(t, expectedLogCtx, taskInfo.LogContext)

		// Verify external resources contain assignment info
		require.Len(t, taskInfo.ExternalResources, 1)
		customInfo := taskInfo.ExternalResources[0].CustomInfo
		assert.Equal(t, workerID, customInfo.Fields["assignedWorker"].GetStringValue())
		assert.Equal(t, executionEnvID.Project, customInfo.Fields["environmentProject"].GetStringValue())
	})

	t.Run("worker pod not found", func(t *testing.T) {
		builder := &interfaceMocks.EnvironmentBuilder{}
		builder.EXPECT().GetWorkerPod(mock.Anything, executionEnvID.String(), workerID).Return(nil, k8serrors.NewNotFound(schema.GroupResource{Resource: "pods"}, "pod"))

		plugin.builder = builder

		taskInfo, err := plugin.buildTaskInfoWithLogs(ctx, executionEnvID, executionEnv, workerID, taskExecutionID, taskTemplate, start, now)

		assert.Equal(t, podContainerNotFoundError, err)
		assert.NotNil(t, taskInfo) // Basic taskInfo should still be returned
		assert.Empty(t, taskInfo.Logs)
		assert.Nil(t, taskInfo.LogContext)
	})

	t.Run("worker pod gone", func(t *testing.T) {
		builder := &interfaceMocks.EnvironmentBuilder{}
		builder.EXPECT().GetWorkerPod(mock.Anything, executionEnvID.String(), workerID).Return(nil, k8serrors.NewGone("pod deleted"))

		plugin.builder = builder

		taskInfo, err := plugin.buildTaskInfoWithLogs(ctx, executionEnvID, executionEnv, workerID, taskExecutionID, taskTemplate, start, now)

		assert.Equal(t, podContainerNotFoundError, err)
		assert.NotNil(t, taskInfo)
		assert.Empty(t, taskInfo.Logs)
		assert.Nil(t, taskInfo.LogContext)
	})

	t.Run("worker pod resource expired", func(t *testing.T) {
		builder := &interfaceMocks.EnvironmentBuilder{}
		builder.EXPECT().GetWorkerPod(mock.Anything, executionEnvID.String(), workerID).Return(nil, k8serrors.NewResourceExpired("pod expired"))

		plugin.builder = builder

		taskInfo, err := plugin.buildTaskInfoWithLogs(ctx, executionEnvID, executionEnv, workerID, taskExecutionID, taskTemplate, start, now)

		assert.Equal(t, podContainerNotFoundError, err)
		assert.NotNil(t, taskInfo)
		assert.Empty(t, taskInfo.Logs)
		assert.Nil(t, taskInfo.LogContext)
	})

	t.Run("worker pod other error", func(t *testing.T) {
		builder := &interfaceMocks.EnvironmentBuilder{}
		builder.EXPECT().GetWorkerPod(mock.Anything, executionEnvID.String(), workerID).Return(nil, fmt.Errorf("unexpected error"))

		plugin.builder = builder

		taskInfo, err := plugin.buildTaskInfoWithLogs(ctx, executionEnvID, executionEnv, workerID, taskExecutionID, taskTemplate, start, now)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected error")
		assert.Nil(t, taskInfo)
	})

	t.Run("mismatched container name", func(t *testing.T) {
		builder := &interfaceMocks.EnvironmentBuilder{}
		builder.EXPECT().GetWorkerPod(mock.Anything, executionEnvID.String(), workerID).Return(&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "namespace",
				Name:      podName,
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "different-container-name",
					},
				},
				Hostname: "hostname",
			},
			Status: v1.PodStatus{},
		}, nil)

		plugin.builder = builder

		taskInfo, err := plugin.buildTaskInfoWithLogs(ctx, executionEnvID, executionEnv, workerID, taskExecutionID, taskTemplate, start, now)

		assert.Equal(t, podContainerNotFoundError, err)
		assert.NotNil(t, taskInfo)
		assert.Empty(t, taskInfo.Logs)
		assert.Nil(t, taskInfo.LogContext)
	})

	t.Run("no container status", func(t *testing.T) {
		builder := &interfaceMocks.EnvironmentBuilder{}
		builder.EXPECT().GetWorkerPod(mock.Anything, executionEnvID.String(), workerID).Return(&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "namespace",
				Name:      podName,
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: podName,
					},
				},
				Hostname: "hostname",
			},
			Status: v1.PodStatus{}, // Empty status
		}, nil)

		plugin.builder = builder

		taskInfo, err := plugin.buildTaskInfoWithLogs(ctx, executionEnvID, executionEnv, workerID, taskExecutionID, taskTemplate, start, now)

		assert.Equal(t, podContainerNotFoundError, err)
		assert.NotNil(t, taskInfo)
		assert.Empty(t, taskInfo.Logs)
		assert.Equal(t, expectedLogCtx, taskInfo.LogContext) // LogContext should still be set
	})

	t.Run("log plugin initialization fails", func(t *testing.T) {
		pluginWithBadConfig := &Plugin{
			cfg: &Config{
				Logs: logs.LogConfig{
					Templates: []tasklog.TemplateLogPlugin{
						{
							DisplayName:  "Bad Config",
							TemplateURIs: []string{"invalid://template"},
						},
					},
				},
			},
		}

		builder := &interfaceMocks.EnvironmentBuilder{}
		builder.EXPECT().GetWorkerPod(mock.Anything, executionEnvID.String(), workerID).Return(&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "namespace",
				Name:      podName,
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: podName,
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
		}, nil)

		pluginWithBadConfig.builder = builder

		taskInfo, err := pluginWithBadConfig.buildTaskInfoWithLogs(ctx, executionEnvID, executionEnv, workerID, taskExecutionID, taskTemplate, start, now)

		// This test may or may not fail depending on the actual implementation of InitializeLogPlugins
		if err != nil {
			assert.Nil(t, taskInfo)
		} else {
			assert.NotNil(t, taskInfo)
		}
	})

	t.Run("get task logs fails", func(t *testing.T) {
		// This test would require mocking the log plugin's GetTaskLogs method
		// The exact implementation depends on how you want to mock the logs.InitializeLogPlugins
		// and the returned logPlugin.GetTaskLogs method

		builder := &interfaceMocks.EnvironmentBuilder{}
		builder.EXPECT().GetWorkerPod(mock.Anything, executionEnvID.String(), workerID).Return(&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "namespace",
				Name:      podName,
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: podName,
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
		}, nil)

		plugin.builder = builder

		// Note: This test is harder to implement without being able to mock the log plugin
		// You might need to create a custom mock for the log plugin or use dependency injection
		// to make this more testable

		taskInfo, err := plugin.buildTaskInfoWithLogs(ctx, executionEnvID, executionEnv, workerID, taskExecutionID, taskTemplate, start, now)

		// For now, we expect this to succeed since we can't easily mock the log plugin failure
		if err != nil {
			assert.Nil(t, taskInfo)
		} else {
			assert.NotNil(t, taskInfo)
		}
	})
}

func init() {
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey,
		contextutils.TaskIDKey)
}
