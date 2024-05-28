package plugin

import (
	"context"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	_struct "github.com/golang/protobuf/ptypes/struct"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	idlcore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	coremocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	iomocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"

	"github.com/unionai/flyte/fasttask/plugin/mocks"
	"github.com/unionai/flyte/fasttask/plugin/pb"
)

func buildFasttaskEnvironment(t *testing.T, fastTaskExtant *pb.FastTaskEnvironment, fastTaskSpec *pb.FastTaskEnvironmentSpec) *_struct.Struct {
	executionEnv := &idlcore.ExecutionEnv{
		Id:   "foo",
		Type: "fast-task",
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

func TestFinalize(t *testing.T) {
	ctx := context.TODO()

	// initialize fasttask TaskTemplate
	taskTemplate := getBaseFasttaskTaskTemplate(t)
	taskReader := &coremocks.TaskReader{}
	taskReader.OnReadMatch(mock.Anything).Return(taskTemplate, nil)

	// initialize static execution context attributes
	taskMetadata := &coremocks.TaskExecutionMetadata{}
	taskExecutionID := &coremocks.TaskExecutionID{}
	taskExecutionID.OnGetGeneratedNameWithMatch(mock.Anything, mock.Anything).Return("task_id", nil)
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
	fastTaskService.OnCleanup(ctx, "task_id", "foo", "w0").Return(nil)

	// initialize plugin
	plugin := &Plugin{
		fastTaskService: fastTaskService,
	}

	// call handle
	err := plugin.Finalize(ctx, tCtx)
	assert.Nil(t, err)
}

func TestGetExecutionEnv(t *testing.T) {
	ctx := context.TODO()

	expectedExtant := &pb.FastTaskEnvironment{
		QueueId: "foo",
	}
	expectedExtantStruct := &_struct.Struct{}
	err := utils.MarshalStruct(expectedExtant, expectedExtantStruct)
	assert.Nil(t, err)

	tests := []struct {
		name            string
		fastTaskExtant  *pb.FastTaskEnvironment
		fastTaskSpec    *pb.FastTaskEnvironmentSpec
		clientGetExists bool
	}{
		{
			name: "ExecutionExtant",
			fastTaskExtant: &pb.FastTaskEnvironment{
				QueueId: "foo",
			},
		},
		{
			name:            "ExecutionSpecExists",
			fastTaskSpec:    &pb.FastTaskEnvironmentSpec{},
			clientGetExists: true,
		},
		{
			name: "ExecutionSpecCreate",
			fastTaskSpec: &pb.FastTaskEnvironmentSpec{
				PodTemplateSpec: []byte("bar"),
			},
			clientGetExists: false,
		},
		{
			name:            "ExecutionSpecInjectPodTemplateAndCreate",
			fastTaskSpec:    &pb.FastTaskEnvironmentSpec{},
			clientGetExists: false,
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
	taskMetadata.OnGetEnvironmentVariables().Return(nil)
	taskMetadata.OnGetK8sServiceAccount().Return("service-account")
	taskMetadata.OnGetNamespace().Return("test-namespace")
	taskMetadata.OnGetPlatformResources().Return(&v1.ResourceRequirements{})
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
	})
	taskExecutionID.OnGetGeneratedNameMatch().Return("task_id")
	taskMetadata.OnGetTaskExecutionID().Return(taskExecutionID)
	taskExecutionID.OnGetGeneratedNameWithMatch(mock.Anything, mock.Anything).Return("task_id", nil)
	taskMetadata.OnGetTaskExecutionID().Return(taskExecutionID)

	taskOverrides := &coremocks.TaskOverrides{}
	taskOverrides.OnGetResourcesMatch().Return(&v1.ResourceRequirements{})
	taskOverrides.OnGetExtendedResourcesMatch().Return(nil)
	taskOverrides.OnGetContainerImageMatch().Return("")
	taskMetadata.OnGetOverridesMatch().Return(taskOverrides)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
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
			}

			// create ExecutionEnvClient mock
			executionEnvClient := &coremocks.ExecutionEnvClient{}
			if test.clientGetExists {
				executionEnvClient.OnGetMatch(ctx, mock.Anything).Return(expectedExtantStruct)
			} else {
				executionEnvClient.OnGetMatch(ctx, mock.Anything).Return(nil)
			}
			executionEnvClient.OnCreateMatch(ctx, "foo", mock.Anything).Return(expectedExtantStruct, nil)

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
			plugin := &Plugin{}

			// call handle
			fastTaskEnvironment, err := plugin.getExecutionEnv(ctx, tCtx)
			assert.Nil(t, err)
			assert.True(t, proto.Equal(expectedExtant, fastTaskEnvironment))
		})
	}
}

func TestHandleNotYetStarted(t *testing.T) {
	ctx := context.TODO()
	tests := []struct {
		name           string
		workerID       string
		lastUpdated    time.Time
		expectedPhase  core.Phase
		expectedReason string
		expectedError  error
	}{
		{
			name:           "NoWorkersAvailable",
			workerID:       "",
			expectedPhase:  core.PhaseWaitingForResources,
			expectedReason: "no workers available",
			expectedError:  nil,
		},
		{
			name:           "NoWorkersAvailableGracePeriodFailure",
			workerID:       "",
			lastUpdated:    time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
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
	taskTemplate := getBaseFasttaskTaskTemplate(t)
	taskReader := &coremocks.TaskReader{}
	taskReader.OnReadMatch(mock.Anything).Return(taskTemplate, nil)

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
	taskExecutionID := &coremocks.TaskExecutionID{}
	taskExecutionID.OnGetGeneratedNameWithMatch(mock.Anything, mock.Anything).Return("task_id", nil)
	taskMetadata.OnGetTaskExecutionID().Return(taskExecutionID)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
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

			// create FastTaskService mock
			fastTaskService := &mocks.FastTaskService{}
			fastTaskService.OnOfferOnQueue(ctx, "foo", "task_id", "namespace", "execution_id", []string{}).Return(test.workerID, nil)

			// initialize plugin
			plugin := &Plugin{
				fastTaskService: fastTaskService,
			}

			// call handle
			transition, err := plugin.Handle(ctx, tCtx)
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedPhase, transition.Info().Phase())
			assert.Equal(t, test.expectedReason, transition.Info().Reason())

			if len(test.workerID) > 0 {
				assert.Equal(t, test.workerID, arrayNodeStateOutput.WorkerID)
			}
		})
	}
}

func TestHandleRunning(t *testing.T) {
	ctx := context.TODO()
	tests := []struct {
		name                   string
		lastUpdated            time.Time
		taskStatusPhase        core.Phase
		taskStatusReason       string
		checkStatusError       error
		expectedPhase          core.Phase
		expectedReason         string
		expectedError          error
		expectedLastUpdatedInc bool
	}{
		{
			name:                   "Running",
			lastUpdated:            time.Now().Add(-5 * time.Second),
			taskStatusPhase:        core.PhaseRunning,
			taskStatusReason:       "",
			checkStatusError:       nil,
			expectedPhase:          core.PhaseRunning,
			expectedReason:         "",
			expectedError:          nil,
			expectedLastUpdatedInc: true,
		},
		{
			name:                   "RetryableFailure",
			lastUpdated:            time.Now().Add(-5 * time.Second),
			taskStatusPhase:        core.PhaseRetryableFailure,
			checkStatusError:       nil,
			expectedPhase:          core.PhaseRetryableFailure,
			expectedError:          nil,
			expectedLastUpdatedInc: false,
		},
		{
			name:                   "StatusNotFoundTimeout",
			lastUpdated:            time.Now().Add(-600 * time.Second),
			taskStatusPhase:        core.PhaseUndefined,
			checkStatusError:       statusUpdateNotFoundError,
			expectedPhase:          core.PhaseRetryableFailure,
			expectedError:          nil,
			expectedLastUpdatedInc: false,
		},
		{
			name:                   "Success",
			lastUpdated:            time.Now().Add(-5 * time.Second),
			taskStatusPhase:        core.PhaseSuccess,
			checkStatusError:       nil,
			expectedPhase:          core.PhaseSuccess,
			expectedError:          nil,
			expectedLastUpdatedInc: false,
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
	taskExecutionID.OnGetGeneratedNameWithMatch(mock.Anything, mock.Anything).Return("task_id", nil)
	taskMetadata.OnGetTaskExecutionID().Return(taskExecutionID)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create TaskExecutionContext
			tCtx := &coremocks.TaskExecutionContext{}
			tCtx.OnTaskExecutionMetadata().Return(taskMetadata)
			tCtx.OnTaskReader().Return(taskReader)

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
			fastTaskService.OnCheckStatusMatch(ctx, "task_id", "foo", "w0").Return(test.taskStatusPhase, "", test.checkStatusError)

			// initialize plugin
			plugin := &Plugin{
				fastTaskService: fastTaskService,
			}

			// call handle
			transition, err := plugin.Handle(ctx, tCtx)
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedPhase, transition.Info().Phase())

			if test.expectedLastUpdatedInc {
				assert.True(t, arrayNodeStateOutput.LastUpdated.After(test.lastUpdated))
			}
		})
	}
}
