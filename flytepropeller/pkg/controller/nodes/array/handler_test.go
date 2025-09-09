package array

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	idlcore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	pluginiomocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	eventmocks "github.com/flyteorg/flyte/flytepropeller/events/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	execmocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/catalog"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/errors"
	gatemocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/gate/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces/mocks"
	recoverymocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/recovery/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/subworkflow/launchplan"
	"github.com/flyteorg/flyte/flytestdlib/bitarray"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

var (
	taskRef       = "taskRef"
	arrayNodeSpec = v1alpha1.NodeSpec{
		ID: "foo",
		ArrayNode: &v1alpha1.ArrayNodeSpec{
			SubNodeSpec: &v1alpha1.NodeSpec{
				Kind:    v1alpha1.NodeKindTask,
				TaskRef: &taskRef,
			},
		},
	}
	workflowMaxParallelism = uint32(10)
	testError              = &idlcore.ExecutionError{Message: "test error"}
)

func createArrayNodeHandler(ctx context.Context, t *testing.T, nodeHandler interfaces.NodeHandler, dataStore *storage.DataStore, scope promutils.Scope) (interfaces.NodeHandler, error) {
	// mock components
	adminClient := launchplan.NewFailFastLaunchPlanExecutor()
	enqueueWorkflowFunc := func(workflowID v1alpha1.WorkflowID) {}
	eventConfig := &config.EventConfig{}
	offloadingConfig := config.LiteralOffloadingConfig{Enabled: false}
	literalOffloadingConfig := config.LiteralOffloadingConfig{Enabled: true, MinSizeInMBForOffloading: 1024, MaxSizeInMBForOffloading: 1024 * 1024}
	mockEventSink := eventmocks.NewMockEventSink()
	mockHandlerFactory := &mocks.HandlerFactory{}
	mockHandlerFactory.EXPECT().GetHandler(mock.Anything).Return(nodeHandler, nil)
	mockKubeClient := execmocks.NewFakeKubeClient()
	mockRecoveryClient := &recoverymocks.Client{}
	mockSignalClient := &gatemocks.SignalServiceClient{}
	noopCatalogClient := catalog.NOOPCatalog{}

	// create node executor
	nodeExecutor, err := nodes.NewExecutor(ctx, config.GetConfig().NodeConfig, dataStore, enqueueWorkflowFunc, mockEventSink, adminClient,
		adminClient, "s3://bucket/", mockKubeClient, noopCatalogClient, mockRecoveryClient, offloadingConfig, eventConfig, "clusterID", mockSignalClient, mockHandlerFactory, scope)
	assert.NoError(t, err)

	// return ArrayNodeHandler
	arrayNodeHandler, err := New(nodeExecutor, eventConfig, literalOffloadingConfig, scope)
	if err != nil {
		return nil, err
	}

	err = arrayNodeHandler.Setup(ctx, nil)
	return arrayNodeHandler, err
}

func createNodeExecutionContext(dataStore *storage.DataStore, eventRecorder interfaces.EventRecorder, outputVariables []string,
	inputLiteralMap *idlcore.LiteralMap, arrayNodeSpec *v1alpha1.NodeSpec, arrayNodeState *handler.ArrayNodeState,
	currentParallelism uint32, maxParallelism uint32) interfaces.NodeExecutionContext {

	nCtx := &mocks.NodeExecutionContext{}
	nCtx.EXPECT().CurrentAttempt().Return(uint32(0))

	// ContextualNodeLookup
	nodeLookup := &execmocks.NodeLookup{}
	nodeLookup.EXPECT().FromNode(mock.Anything).Return(nil, nil)
	nCtx.EXPECT().ContextualNodeLookup().Return(nodeLookup)

	// DataStore
	nCtx.EXPECT().DataStore().Return(dataStore)

	// ExecutionContext
	executionContext := &execmocks.ExecutionContext{}
	executionContext.EXPECT().GetEventVersion().Return(1)
	executionContext.EXPECT().GetExecutionConfig().Return(v1alpha1.ExecutionConfig{
		MaxParallelism: maxParallelism,
	})
	executionContext.EXPECT().GetExecutionID().Return(
		v1alpha1.ExecutionID{
			WorkflowExecutionIdentifier: &idlcore.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
		})
	executionContext.EXPECT().GetLabels().Return(nil)
	executionContext.EXPECT().GetRawOutputDataConfig().Return(v1alpha1.RawOutputDataConfig{})
	executionContext.EXPECT().IsInterruptible().Return(false)
	executionContext.EXPECT().GetParentInfo().Return(nil)
	outputVariableMap := make(map[string]*idlcore.Variable)
	for _, outputVariable := range outputVariables {
		outputVariableMap[outputVariable] = &idlcore.Variable{}
	}
	executionContext.EXPECT().GetTask(taskRef).Return(
		&v1alpha1.TaskSpec{
			TaskTemplate: &idlcore.TaskTemplate{
				Interface: &idlcore.TypedInterface{
					Outputs: &idlcore.VariableMap{
						Variables: outputVariableMap,
					},
				},
			},
		},
		nil,
	)
	executionContext.EXPECT().CurrentParallelism().Return(currentParallelism)
	executionContext.On("IncrementParallelism").Run(func(args mock.Arguments) {}).Return(currentParallelism)
	executionContext.EXPECT().IncrementNodeExecutionCount().Return(1)
	executionContext.EXPECT().IncrementTaskExecutionCount().Return(1)
	executionContext.EXPECT().CurrentNodeExecutionCount().Return(1)
	executionContext.EXPECT().CurrentTaskExecutionCount().Return(1)
	nCtx.EXPECT().ExecutionContext().Return(executionContext)

	// EventsRecorder
	nCtx.EXPECT().EventsRecorder().Return(eventRecorder)

	// InputReader
	inputFilePaths := &pluginiomocks.InputFilePaths{}
	inputFilePaths.EXPECT().GetInputPath().Return(storage.DataReference("s3://bucket/input"))
	nCtx.EXPECT().InputReader().Return(
		newStaticInputReader(
			inputFilePaths,
			inputLiteralMap,
		))

	// Node
	nCtx.EXPECT().Node().Return(arrayNodeSpec)

	// NodeExecutionMetadata
	nodeExecutionMetadata := &mocks.NodeExecutionMetadata{}
	nodeExecutionMetadata.EXPECT().GetNodeExecutionID().Return(&idlcore.NodeExecutionIdentifier{
		NodeId: "foo",
		ExecutionId: &idlcore.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
	})
	nodeExecutionMetadata.EXPECT().GetOwnerID().Return(types.NamespacedName{
		Namespace: "wf-namespace",
		Name:      "wf-name",
	})
	nCtx.EXPECT().NodeExecutionMetadata().Return(nodeExecutionMetadata)

	// NodeID
	nCtx.EXPECT().NodeID().Return("foo")

	// NodeStateReader
	nodeStateReader := &mocks.NodeStateReader{}
	nodeStateReader.EXPECT().GetArrayNodeState().Return(*arrayNodeState)
	nCtx.EXPECT().NodeStateReader().Return(nodeStateReader)

	// NodeStateWriter
	nodeStateWriter := &mocks.NodeStateWriter{}
	nodeStateWriter.EXPECT().PutArrayNodeState(mock.Anything).Run(
		func(s handler.ArrayNodeState) {
			*arrayNodeState = s
		},
	).Return(nil)
	nCtx.EXPECT().NodeStateWriter().Return(nodeStateWriter)

	// NodeStatus
	nowMinus := time.Now().Add(time.Duration(-5) * time.Second)
	metav1NowMinus := metav1.Time{
		Time: nowMinus,
	}
	nCtx.EXPECT().NodeStatus().Return(&v1alpha1.NodeStatus{
		DataDir:              storage.DataReference("s3://bucket/data"),
		OutputDir:            storage.DataReference("s3://bucket/output"),
		LastAttemptStartedAt: &metav1NowMinus,
		StartedAt:            &metav1NowMinus,
	})

	return nCtx
}

func TestAbort(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name                           string
		inputMap                       map[string][]int64
		subNodePhases                  []v1alpha1.NodePhase
		subNodeTaskPhases              []core.Phase
		expectedAbortCalls             int
		expectedExternalResourcePhases []idlcore.TaskExecution_Phase
		arrayNodeStatePhase            v1alpha1.ArrayNodePhase
		arrayNodeStateError            *idlcore.ExecutionError
		expectedTaskExecutionPhase     idlcore.TaskExecution_Phase
		expectTaskExecutionError       bool
	}{
		{
			name: "Aborted after failed",
			inputMap: map[string][]int64{
				"foo": []int64{0, 1, 2},
			},
			subNodePhases:                  []v1alpha1.NodePhase{v1alpha1.NodePhaseSucceeded, v1alpha1.NodePhaseRunning, v1alpha1.NodePhaseNotYetStarted},
			subNodeTaskPhases:              []core.Phase{core.PhaseSuccess, core.PhaseRunning, core.PhaseUndefined},
			expectedExternalResourcePhases: []idlcore.TaskExecution_Phase{idlcore.TaskExecution_SUCCEEDED, idlcore.TaskExecution_ABORTED, idlcore.TaskExecution_UNDEFINED},
			expectedAbortCalls:             1,
			arrayNodeStatePhase:            v1alpha1.ArrayNodePhaseFailing,
			arrayNodeStateError:            testError,
			expectedTaskExecutionPhase:     idlcore.TaskExecution_FAILED,
			expectTaskExecutionError:       true,
		},
		{
			name: "Aborted while running",
			inputMap: map[string][]int64{
				"foo": []int64{0, 1, 2},
			},
			subNodePhases:                  []v1alpha1.NodePhase{v1alpha1.NodePhaseSucceeded, v1alpha1.NodePhaseRunning, v1alpha1.NodePhaseNotYetStarted},
			subNodeTaskPhases:              []core.Phase{core.PhaseSuccess, core.PhaseRunning, core.PhaseUndefined},
			expectedExternalResourcePhases: []idlcore.TaskExecution_Phase{idlcore.TaskExecution_SUCCEEDED, idlcore.TaskExecution_ABORTED, idlcore.TaskExecution_UNDEFINED},
			expectedAbortCalls:             1,
			arrayNodeStatePhase:            v1alpha1.ArrayNodePhaseExecuting,
			arrayNodeStateError:            testError,
			expectedTaskExecutionPhase:     idlcore.TaskExecution_ABORTED,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scope := promutils.NewTestScope()
			dataStore, err := storage.NewDataStore(&storage.Config{
				Type: storage.TypeMemory,
			}, scope)
			assert.NoError(t, err)

			nodeHandler := &mocks.NodeHandler{}
			nodeHandler.EXPECT().Abort(mock.Anything, mock.Anything, mock.Anything).Return(nil)
			nodeHandler.EXPECT().Finalize(mock.Anything, mock.Anything).Return(nil)

			// initialize ArrayNodeHandler
			arrayNodeHandler, err := createArrayNodeHandler(ctx, t, nodeHandler, dataStore, scope)
			assert.NoError(t, err)

			// initialize universal variables
			literalMap := convertMapToArrayLiterals(test.inputMap)

			size := -1
			for _, v := range test.inputMap {
				if size == -1 {
					size = len(v)
				} else if len(v) > size { // calculating size as largest input list
					size = len(v)
				}
			}

			// initialize ArrayNodeState
			arrayNodeState := &handler.ArrayNodeState{
				Phase: test.arrayNodeStatePhase,
				Error: test.arrayNodeStateError,
			}
			for _, item := range []struct {
				arrayReference *bitarray.CompactArray
				maxValue       int
			}{
				{arrayReference: &arrayNodeState.SubNodePhases, maxValue: int(v1alpha1.NodePhaseRecovered)},
				{arrayReference: &arrayNodeState.SubNodeTaskPhases, maxValue: len(core.Phases) - 1},
				{arrayReference: &arrayNodeState.SubNodeRetryAttempts, maxValue: 1},
				{arrayReference: &arrayNodeState.SubNodeSystemFailures, maxValue: 1},
				{arrayReference: &arrayNodeState.SubNodeDeltaTimestamps, maxValue: 1024},
			} {

				*item.arrayReference, err = bitarray.NewCompactArray(uint(size), bitarray.Item(item.maxValue)) // #nosec G115
				assert.NoError(t, err)
			}

			for i, nodePhase := range test.subNodePhases {
				arrayNodeState.SubNodePhases.SetItem(i, bitarray.Item(nodePhase)) // #nosec G115
			}
			for i, taskPhase := range test.subNodeTaskPhases {
				arrayNodeState.SubNodeTaskPhases.SetItem(i, bitarray.Item(taskPhase)) // #nosec G115
			}

			// create NodeExecutionContext
			eventRecorder := newBufferedEventRecorder()
			nCtx := createNodeExecutionContext(dataStore, eventRecorder, nil, literalMap, &arrayNodeSpec, arrayNodeState, 0, workflowMaxParallelism)

			// evaluate node
			err = arrayNodeHandler.Abort(ctx, nCtx, "foo")
			assert.NoError(t, err)

			nodeHandler.AssertNumberOfCalls(t, "Abort", test.expectedAbortCalls)
			if len(test.expectedExternalResourcePhases) > 0 {
				assert.Equal(t, 1, len(eventRecorder.taskExecutionEvents))
				assert.Equal(t, test.expectedTaskExecutionPhase, eventRecorder.taskExecutionEvents[0].GetPhase())

				if test.expectTaskExecutionError {
					assert.Equal(t, testError.GetMessage(), eventRecorder.taskExecutionEvents[0].GetError().GetMessage())
				} else {
					assert.Nil(t, eventRecorder.taskExecutionEvents[0].GetError())
				}

				externalResources := eventRecorder.taskExecutionEvents[0].GetMetadata().GetExternalResources()
				assert.Equal(t, len(test.expectedExternalResourcePhases), len(externalResources))
				for i, expectedPhase := range test.expectedExternalResourcePhases {
					assert.Equal(t, expectedPhase, externalResources[i].GetPhase())
				}
			} else {
				assert.Equal(t, 0, len(eventRecorder.taskExecutionEvents))
			}
		})
	}
}

func TestFinalize(t *testing.T) {
	ctx := context.Background()
	scope := promutils.NewTestScope()
	dataStore, err := storage.NewDataStore(&storage.Config{
		Type: storage.TypeMemory,
	}, scope)
	assert.NoError(t, err)

	nodeHandler := &mocks.NodeHandler{}
	nodeHandler.EXPECT().Finalize(mock.Anything, mock.Anything).Return(nil)

	// initialize ArrayNodeHandler
	arrayNodeHandler, err := createArrayNodeHandler(ctx, t, nodeHandler, dataStore, scope)
	assert.NoError(t, err)

	tests := []struct {
		name                  string
		inputMap              map[string][]int64
		subNodePhases         []v1alpha1.NodePhase
		subNodeTaskPhases     []core.Phase
		expectedFinalizeCalls int
	}{
		{
			name: "Success",
			inputMap: map[string][]int64{
				"foo": []int64{0, 1, 2},
			},
			subNodePhases:         []v1alpha1.NodePhase{v1alpha1.NodePhaseSucceeded, v1alpha1.NodePhaseRunning, v1alpha1.NodePhaseNotYetStarted},
			subNodeTaskPhases:     []core.Phase{core.PhaseSuccess, core.PhaseRunning, core.PhaseUndefined},
			expectedFinalizeCalls: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// initialize universal variables
			literalMap := convertMapToArrayLiterals(test.inputMap)

			size := -1
			for _, v := range test.inputMap {
				if size == -1 {
					size = len(v)
				} else if len(v) > size { // calculating size as largest input list
					size = len(v)
				}
			}

			// initialize ArrayNodeState
			arrayNodeState := &handler.ArrayNodeState{
				Phase: v1alpha1.ArrayNodePhaseFailing,
			}
			for _, item := range []struct {
				arrayReference *bitarray.CompactArray
				maxValue       int
			}{
				{arrayReference: &arrayNodeState.SubNodePhases, maxValue: int(v1alpha1.NodePhaseRecovered)},
				{arrayReference: &arrayNodeState.SubNodeTaskPhases, maxValue: len(core.Phases) - 1},
				{arrayReference: &arrayNodeState.SubNodeRetryAttempts, maxValue: 1},
				{arrayReference: &arrayNodeState.SubNodeSystemFailures, maxValue: 1},
				{arrayReference: &arrayNodeState.SubNodeDeltaTimestamps, maxValue: 1024},
			} {
				*item.arrayReference, err = bitarray.NewCompactArray(uint(size), bitarray.Item(item.maxValue)) // #nosec G115
				assert.NoError(t, err)
			}

			for i, nodePhase := range test.subNodePhases {
				arrayNodeState.SubNodePhases.SetItem(i, bitarray.Item(nodePhase)) // #nosec G115
			}
			for i, taskPhase := range test.subNodeTaskPhases {
				arrayNodeState.SubNodeTaskPhases.SetItem(i, bitarray.Item(taskPhase)) // #nosec G115
			}

			// create NodeExecutionContext
			eventRecorder := newBufferedEventRecorder()
			nCtx := createNodeExecutionContext(dataStore, eventRecorder, nil, literalMap, &arrayNodeSpec, arrayNodeState, 0, workflowMaxParallelism)

			// evaluate node
			err := arrayNodeHandler.Finalize(ctx, nCtx)
			assert.NoError(t, err)

			// validate
			nodeHandler.AssertNumberOfCalls(t, "Finalize", test.expectedFinalizeCalls)
		})
	}
}

func TestHandleArrayNodePhaseNone(t *testing.T) {
	ctx := context.Background()
	scope := promutils.NewTestScope()
	dataStore, err := storage.NewDataStore(&storage.Config{
		Type: storage.TypeMemory,
	}, scope)
	assert.NoError(t, err)
	nodeHandler := &mocks.NodeHandler{}

	// initialize ArrayNodeHandler
	arrayNodeHandler, err := createArrayNodeHandler(ctx, t, nodeHandler, dataStore, scope)
	assert.NoError(t, err)

	tests := []struct {
		name                           string
		inputMap                       *idlcore.LiteralMap
		expectedArrayNodePhase         v1alpha1.ArrayNodePhase
		expectedTransitionPhase        handler.EPhase
		expectedExternalResourcePhases []idlcore.TaskExecution_Phase
		boundInputs                    []string
	}{
		{
			name: "Success",
			inputMap: &idlcore.LiteralMap{
				Literals: map[string]*idlcore.Literal{
					"foo": convertListToLiterals([]int64{1, 2}),
				},
			},
			expectedArrayNodePhase:         v1alpha1.ArrayNodePhaseExecuting,
			expectedTransitionPhase:        handler.EPhaseRunning,
			expectedExternalResourcePhases: []idlcore.TaskExecution_Phase{idlcore.TaskExecution_UNDEFINED, idlcore.TaskExecution_UNDEFINED},
		},
		{
			name: "SuccessMultipleInputs",
			inputMap: &idlcore.LiteralMap{
				Literals: map[string]*idlcore.Literal{
					"foo": convertListToLiterals([]int64{1, 2, 3}),
					"bar": convertListToLiterals([]int64{4, 5, 6}),
				},
			},
			expectedArrayNodePhase:         v1alpha1.ArrayNodePhaseExecuting,
			expectedTransitionPhase:        handler.EPhaseRunning,
			expectedExternalResourcePhases: []idlcore.TaskExecution_Phase{idlcore.TaskExecution_UNDEFINED, idlcore.TaskExecution_UNDEFINED, idlcore.TaskExecution_UNDEFINED},
		},
		{
			name: "FailureDifferentInputListLengths",
			inputMap: &idlcore.LiteralMap{
				Literals: map[string]*idlcore.Literal{
					"foo": convertListToLiterals([]int64{1, 2}),
					"bar": convertListToLiterals([]int64{3}),
				},
			},
			expectedArrayNodePhase:         v1alpha1.ArrayNodePhaseNone,
			expectedTransitionPhase:        handler.EPhaseFailed,
			expectedExternalResourcePhases: nil,
		},
		{
			name: "boundInputs",
			inputMap: &idlcore.LiteralMap{
				Literals: map[string]*idlcore.Literal{
					"foo": convertListToLiterals([]int64{1, 2, 3}),
					"bar": {
						Value: &idlcore.Literal_Scalar{
							Scalar: &idlcore.Scalar{
								Value: &idlcore.Scalar_Primitive{
									Primitive: &idlcore.Primitive{
										Value: &idlcore.Primitive_Integer{
											Integer: 1,
										},
									},
								},
							},
						},
					},
				},
			},
			boundInputs:                    []string{"bar"},
			expectedArrayNodePhase:         v1alpha1.ArrayNodePhaseExecuting,
			expectedTransitionPhase:        handler.EPhaseRunning,
			expectedExternalResourcePhases: []idlcore.TaskExecution_Phase{idlcore.TaskExecution_UNDEFINED, idlcore.TaskExecution_UNDEFINED, idlcore.TaskExecution_UNDEFINED},
		},
		{
			name: "All boundInputs",
			inputMap: &idlcore.LiteralMap{
				Literals: map[string]*idlcore.Literal{
					"foo": convertListToLiterals([]int64{1, 2, 3}),
					"bar": convertListToLiterals([]int64{1, 2, 3}),
				},
			},
			boundInputs:                    []string{"foo", "bar"},
			expectedArrayNodePhase:         v1alpha1.ArrayNodePhaseExecuting,
			expectedTransitionPhase:        handler.EPhaseRunning,
			expectedExternalResourcePhases: []idlcore.TaskExecution_Phase{idlcore.TaskExecution_UNDEFINED},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create NodeExecutionContext
			eventRecorder := newBufferedEventRecorder()
			arrayNodeState := &handler.ArrayNodeState{
				Phase: v1alpha1.ArrayNodePhaseNone,
			}
			arrayNodeSpec.ArrayNode.BoundInputs = test.boundInputs
			nCtx := createNodeExecutionContext(dataStore, eventRecorder, nil, test.inputMap, &arrayNodeSpec, arrayNodeState, 0, workflowMaxParallelism)

			// evaluate node
			transition, err := arrayNodeHandler.Handle(ctx, nCtx)
			assert.NoError(t, err)

			// validate results
			assert.Equal(t, test.expectedArrayNodePhase, arrayNodeState.Phase)
			assert.Equal(t, test.expectedTransitionPhase, transition.Info().GetPhase())

			if len(test.expectedExternalResourcePhases) > 0 {
				assert.Equal(t, 1, len(eventRecorder.taskExecutionEvents))

				externalResources := eventRecorder.taskExecutionEvents[0].GetMetadata().GetExternalResources()
				assert.Equal(t, len(test.expectedExternalResourcePhases), len(externalResources))
				for i, expectedPhase := range test.expectedExternalResourcePhases {
					assert.Equal(t, expectedPhase, externalResources[i].GetPhase())
				}
			} else {
				assert.Equal(t, 0, len(eventRecorder.taskExecutionEvents))
			}
		})
	}
}

func uint32Ptr(v uint32) *uint32 {
	return &v
}

type fakeEventRecorder struct {
	taskErr                  error
	phaseVersionFailures     uint32
	recordTaskEventCallCount int
}

func (f *fakeEventRecorder) RecordNodeEvent(ctx context.Context, event *event.NodeExecutionEvent, eventConfig *config.EventConfig) error {
	return nil
}

func (f *fakeEventRecorder) RecordTaskEvent(ctx context.Context, event *event.TaskExecutionEvent, eventConfig *config.EventConfig) error {
	f.recordTaskEventCallCount++
	if f.phaseVersionFailures == 0 || event.GetPhaseVersion() < f.phaseVersionFailures {
		return f.taskErr
	}
	return nil
}

func TestHandleArrayNodePhaseExecuting(t *testing.T) {
	ctx := context.Background()

	// setting default parallelism behavior on ArrayNode to "hybrid" to test the largest scope of functionality
	flyteConfig := config.GetConfig()
	flyteConfig.ArrayNode.DefaultParallelismBehavior = config.ParallelismBehaviorHybrid

	minSuccessRatio := float32(0.5)

	// initialize universal variables
	inputMap := map[string][]int64{
		"foo": []int64{0, 1},
		"bar": []int64{2, 3},
	}
	literalMap := convertMapToArrayLiterals(inputMap)

	size := -1
	for _, v := range inputMap {
		if size == -1 {
			size = len(v)
		} else if len(v) > size { // calculating size as largest input list
			size = len(v)
		}
	}

	tests := []struct {
		name                                    string
		parallelism                             *uint32
		minSuccessRatio                         *float32
		subNodePhases                           []v1alpha1.NodePhase
		subNodeTaskPhases                       []core.Phase
		subNodeDeltaTimestamps                  []uint64
		subNodeTransitions                      []handler.Transition
		expectedArrayNodePhase                  v1alpha1.ArrayNodePhase
		expectedArrayNodeSubPhases              []v1alpha1.NodePhase
		expectedDiffArrayNodeSubDeltaTimestamps []bool
		expectedTransitionPhase                 handler.EPhase
		expectedExternalResourcePhases          []idlcore.TaskExecution_Phase
		currentWfParallelism                    uint32
		maxWfParallelism                        uint32
		incrementParallelismCount               uint32
		useFakeEventRecorder                    bool
		eventRecorderFailures                   uint32
		eventRecorderError                      error
		expectedTaskPhaseVersion                uint32
		expectHandleError                       bool
		expectedEventingCalls                   int
	}{
		{
			name:        "StartAllSubNodes",
			parallelism: uint32Ptr(0),
			subNodePhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseQueued,
				v1alpha1.NodePhaseQueued,
			},
			subNodeTaskPhases: []core.Phase{
				core.PhaseUndefined,
				core.PhaseUndefined,
			},
			subNodeTransitions: []handler.Transition{
				handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRunning(&handler.ExecutionInfo{})),
				handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRunning(&handler.ExecutionInfo{})),
			},
			expectedArrayNodePhase: v1alpha1.ArrayNodePhaseExecuting,
			expectedArrayNodeSubPhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseRunning,
				v1alpha1.NodePhaseRunning,
			},
			expectedTaskPhaseVersion:       1,
			expectedTransitionPhase:        handler.EPhaseRunning,
			expectedExternalResourcePhases: []idlcore.TaskExecution_Phase{idlcore.TaskExecution_RUNNING, idlcore.TaskExecution_RUNNING},
			incrementParallelismCount:      1,
		},
		{
			name:        "StartOneSubNodeParallelism",
			parallelism: uint32Ptr(1),
			subNodePhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseQueued,
				v1alpha1.NodePhaseQueued,
			},
			subNodeTaskPhases: []core.Phase{
				core.PhaseUndefined,
				core.PhaseUndefined,
			},
			subNodeTransitions: []handler.Transition{
				handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRunning(&handler.ExecutionInfo{})),
			},
			expectedArrayNodePhase: v1alpha1.ArrayNodePhaseExecuting,
			expectedArrayNodeSubPhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseRunning,
				v1alpha1.NodePhaseQueued,
			},
			expectedTaskPhaseVersion:       1,
			expectedTransitionPhase:        handler.EPhaseRunning,
			expectedExternalResourcePhases: []idlcore.TaskExecution_Phase{idlcore.TaskExecution_RUNNING},
			incrementParallelismCount:      1,
		},
		{
			name:        "UtilizeWfParallelismAllSubNodes",
			parallelism: nil,
			subNodePhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseQueued,
				v1alpha1.NodePhaseQueued,
			},
			subNodeTaskPhases: []core.Phase{
				core.PhaseUndefined,
				core.PhaseUndefined,
			},
			subNodeTransitions: []handler.Transition{
				handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRunning(&handler.ExecutionInfo{})),
				handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRunning(&handler.ExecutionInfo{})),
			},
			expectedArrayNodePhase: v1alpha1.ArrayNodePhaseExecuting,
			expectedArrayNodeSubPhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseRunning,
				v1alpha1.NodePhaseRunning,
			},
			expectedTaskPhaseVersion:       1,
			expectedTransitionPhase:        handler.EPhaseRunning,
			expectedExternalResourcePhases: []idlcore.TaskExecution_Phase{idlcore.TaskExecution_RUNNING, idlcore.TaskExecution_RUNNING},
			currentWfParallelism:           0,
			incrementParallelismCount:      2,
		},
		{
			name:        "UtilizeWfParallelismSomeSubNodes",
			parallelism: nil,
			subNodePhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseQueued,
				v1alpha1.NodePhaseQueued,
			},
			subNodeTaskPhases: []core.Phase{
				core.PhaseUndefined,
				core.PhaseUndefined,
			},
			subNodeTransitions: []handler.Transition{
				handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRunning(&handler.ExecutionInfo{})),
			},
			expectedArrayNodePhase: v1alpha1.ArrayNodePhaseExecuting,
			expectedArrayNodeSubPhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseRunning,
				v1alpha1.NodePhaseQueued,
			},
			expectedTaskPhaseVersion:       1,
			expectedTransitionPhase:        handler.EPhaseRunning,
			expectedExternalResourcePhases: []idlcore.TaskExecution_Phase{idlcore.TaskExecution_RUNNING},
			currentWfParallelism:           workflowMaxParallelism - 1,
			incrementParallelismCount:      1,
		},
		{
			name:        "UtilizeWfParallelismNoSubNodes",
			parallelism: nil,
			subNodePhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseQueued,
				v1alpha1.NodePhaseQueued,
			},
			subNodeTaskPhases: []core.Phase{
				core.PhaseUndefined,
				core.PhaseUndefined,
			},
			subNodeTransitions:     []handler.Transition{},
			expectedArrayNodePhase: v1alpha1.ArrayNodePhaseExecuting,
			expectedArrayNodeSubPhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseQueued,
				v1alpha1.NodePhaseQueued,
			},
			expectedTaskPhaseVersion:       0,
			expectedTransitionPhase:        handler.EPhaseRunning,
			expectedExternalResourcePhases: []idlcore.TaskExecution_Phase{},
			currentWfParallelism:           workflowMaxParallelism,
			incrementParallelismCount:      0,
		},
		{
			name:        "StartSubNodesNewAttempts",
			parallelism: uint32Ptr(0),
			subNodePhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseQueued,
				v1alpha1.NodePhaseQueued,
			},
			subNodeTaskPhases: []core.Phase{
				core.PhaseRetryableFailure,
				core.PhaseWaitingForResources,
			},
			subNodeTransitions: []handler.Transition{
				handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRunning(&handler.ExecutionInfo{})),
				handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRunning(&handler.ExecutionInfo{})),
			},
			expectedArrayNodePhase: v1alpha1.ArrayNodePhaseExecuting,
			expectedArrayNodeSubPhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseRunning,
				v1alpha1.NodePhaseRunning,
			},
			expectedTaskPhaseVersion:       1,
			expectedTransitionPhase:        handler.EPhaseRunning,
			expectedExternalResourcePhases: []idlcore.TaskExecution_Phase{idlcore.TaskExecution_RUNNING, idlcore.TaskExecution_RUNNING},
			incrementParallelismCount:      1,
		},
		{
			name:        "AllSubNodesSuccedeed",
			parallelism: uint32Ptr(0),
			subNodePhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseRunning,
				v1alpha1.NodePhaseRunning,
			},
			subNodeTaskPhases: []core.Phase{
				core.PhaseRunning,
				core.PhaseRunning,
			},
			subNodeTransitions: []handler.Transition{
				handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoSuccess(&handler.ExecutionInfo{})),
				handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoSuccess(&handler.ExecutionInfo{})),
			},
			expectedArrayNodePhase: v1alpha1.ArrayNodePhaseSucceeding,
			expectedArrayNodeSubPhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseSucceeded,
				v1alpha1.NodePhaseSucceeded,
			},
			expectedTaskPhaseVersion:       0,
			expectedTransitionPhase:        handler.EPhaseRunning,
			expectedExternalResourcePhases: []idlcore.TaskExecution_Phase{idlcore.TaskExecution_SUCCEEDED, idlcore.TaskExecution_SUCCEEDED},
		},
		{
			name:            "OneSubNodeSuccedeedMinSuccessRatio",
			parallelism:     uint32Ptr(0),
			minSuccessRatio: &minSuccessRatio,
			subNodePhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseRunning,
				v1alpha1.NodePhaseRunning,
			},
			subNodeTaskPhases: []core.Phase{
				core.PhaseRunning,
				core.PhaseRunning,
			},
			subNodeTransitions: []handler.Transition{
				handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoSuccess(&handler.ExecutionInfo{})),
				handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(0, "", "", &handler.ExecutionInfo{})),
			},
			expectedArrayNodePhase: v1alpha1.ArrayNodePhaseSucceeding,
			expectedArrayNodeSubPhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseSucceeded,
				v1alpha1.NodePhaseFailed,
			},
			expectedTaskPhaseVersion:       0,
			expectedTransitionPhase:        handler.EPhaseRunning,
			expectedExternalResourcePhases: []idlcore.TaskExecution_Phase{idlcore.TaskExecution_SUCCEEDED, idlcore.TaskExecution_FAILED},
		},
		{
			name:        "OneSubNodeFailed",
			parallelism: uint32Ptr(0),
			subNodePhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseRunning,
				v1alpha1.NodePhaseRunning,
			},
			subNodeTaskPhases: []core.Phase{
				core.PhaseRunning,
				core.PhaseRunning,
			},
			subNodeTransitions: []handler.Transition{
				handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(0, "", "", &handler.ExecutionInfo{})),
				handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoSuccess(&handler.ExecutionInfo{})),
			},
			expectedArrayNodePhase: v1alpha1.ArrayNodePhaseFailing,
			expectedArrayNodeSubPhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseFailed,
				v1alpha1.NodePhaseSucceeded,
			},
			expectedTaskPhaseVersion:       0,
			expectedTransitionPhase:        handler.EPhaseRunning,
			expectedExternalResourcePhases: []idlcore.TaskExecution_Phase{idlcore.TaskExecution_FAILED, idlcore.TaskExecution_SUCCEEDED},
		},
		{
			name:        "EventingFails",
			parallelism: uint32Ptr(0),
			subNodePhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseQueued,
				v1alpha1.NodePhaseQueued,
			},
			subNodeTaskPhases: []core.Phase{
				core.PhaseRunning,
				core.PhaseRunning,
			},
			subNodeTransitions: []handler.Transition{
				handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRunning(&handler.ExecutionInfo{})),
				handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRunning(&handler.ExecutionInfo{})),
			},
			expectedArrayNodeSubPhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseQueued,
				v1alpha1.NodePhaseQueued,
			},
			expectedExternalResourcePhases: []idlcore.TaskExecution_Phase{idlcore.TaskExecution_RUNNING, idlcore.TaskExecution_RUNNING},
			useFakeEventRecorder:           true,
			eventRecorderError:             fmt.Errorf("err"),
			expectHandleError:              true,
			expectedEventingCalls:          1,
		},
		{
			name:        "DeltaTimestampUpdates",
			parallelism: uint32Ptr(0),
			subNodePhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseQueued,
				v1alpha1.NodePhaseRunning,
			},
			subNodeTaskPhases: []core.Phase{
				core.PhaseUndefined,
				core.PhaseUndefined,
			},
			subNodeTransitions: []handler.Transition{
				handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRunning(&handler.ExecutionInfo{})),
				handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRetryableFailure(idlcore.ExecutionError_SYSTEM, "", "", &handler.ExecutionInfo{})),
			},
			expectedArrayNodePhase: v1alpha1.ArrayNodePhaseExecuting,
			expectedArrayNodeSubPhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseRunning,
				v1alpha1.NodePhaseRetryableFailure,
			},
			expectedTaskPhaseVersion:       1,
			expectedTransitionPhase:        handler.EPhaseRunning,
			expectedExternalResourcePhases: []idlcore.TaskExecution_Phase{idlcore.TaskExecution_RUNNING, idlcore.TaskExecution_FAILED},
			incrementParallelismCount:      1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scope := promutils.NewTestScope()
			dataStore, err := storage.NewDataStore(&storage.Config{
				Type: storage.TypeMemory,
			}, scope)
			assert.NoError(t, err)

			var eventRecorder interfaces.EventRecorder
			if test.useFakeEventRecorder {
				eventRecorder = &fakeEventRecorder{
					phaseVersionFailures: test.eventRecorderFailures,
					taskErr:              test.eventRecorderError,
				}
			} else {
				eventRecorder = newBufferedEventRecorder()
			}
			// initialize ArrayNodeState
			arrayNodeState := &handler.ArrayNodeState{
				Phase: v1alpha1.ArrayNodePhaseExecuting,
			}
			for _, item := range []struct {
				arrayReference *bitarray.CompactArray
				maxValue       int
			}{
				{arrayReference: &arrayNodeState.SubNodePhases, maxValue: int(v1alpha1.NodePhaseRecovered)},
				{arrayReference: &arrayNodeState.SubNodeTaskPhases, maxValue: len(core.Phases) - 1},
				{arrayReference: &arrayNodeState.SubNodeRetryAttempts, maxValue: 1},
				{arrayReference: &arrayNodeState.SubNodeSystemFailures, maxValue: 1},
				{arrayReference: &arrayNodeState.SubNodeDeltaTimestamps, maxValue: 1024},
			} {
				*item.arrayReference, err = bitarray.NewCompactArray(uint(size), bitarray.Item(item.maxValue)) // #nosec G115
				assert.NoError(t, err)
			}

			for i, nodePhase := range test.subNodePhases {
				arrayNodeState.SubNodePhases.SetItem(i, bitarray.Item(nodePhase)) // #nosec G115
			}

			for i, deltaTimestmap := range test.subNodeDeltaTimestamps {
				arrayNodeState.SubNodeDeltaTimestamps.SetItem(i, deltaTimestmap) // #nosec G115
			}

			nodeSpec := arrayNodeSpec
			nodeSpec.ArrayNode.Parallelism = test.parallelism
			nodeSpec.ArrayNode.MinSuccessRatio = test.minSuccessRatio

			nCtx := createNodeExecutionContext(dataStore, eventRecorder, nil, literalMap, &nodeSpec, arrayNodeState, test.currentWfParallelism, workflowMaxParallelism)

			// initialize ArrayNodeHandler
			nodeHandler := &mocks.NodeHandler{}
			nodeHandler.EXPECT().FinalizeRequired().Return(false)
			for i, transition := range test.subNodeTransitions {
				nodeID := fmt.Sprintf("n%d", i)
				transitionPhase := test.expectedExternalResourcePhases[i]

				nodeHandler.EXPECT().Handle(mock.Anything, mock.MatchedBy(func(arrayNCtx interfaces.NodeExecutionContext) bool {
					return arrayNCtx.NodeID() == nodeID // match on NodeID using index to ensure each subNode is handled independently
				})).Run(
					func(ctx context.Context, executionContext interfaces.NodeExecutionContext) {
						// mock sending TaskExecutionEvent from handler to show task state transition
						taskExecutionEvent := &event.TaskExecutionEvent{
							Phase: transitionPhase,
						}

						err := executionContext.EventsRecorder().RecordTaskEvent(ctx, taskExecutionEvent, &config.EventConfig{})
						assert.NoError(t, err)
					},
				).Return(transition, nil)
			}

			arrayNodeHandler, err := createArrayNodeHandler(ctx, t, nodeHandler, dataStore, scope)
			assert.NoError(t, err)

			// evaluate node
			transition, err := arrayNodeHandler.Handle(ctx, nCtx)

			fakeEventRecorder, ok := eventRecorder.(*fakeEventRecorder)
			if ok {
				assert.Equal(t, test.expectedEventingCalls, fakeEventRecorder.recordTaskEventCallCount)
			}

			if !test.expectHandleError {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				return
			}

			// validate results
			assert.Equal(t, test.expectedArrayNodePhase, arrayNodeState.Phase)
			assert.Equal(t, test.expectedTransitionPhase, transition.Info().GetPhase())
			assert.Equal(t, test.expectedTaskPhaseVersion, arrayNodeState.TaskPhaseVersion)

			for i, expectedPhase := range test.expectedArrayNodeSubPhases {
				assert.Equal(t, expectedPhase, v1alpha1.NodePhase(arrayNodeState.SubNodePhases.GetItem(i))) // #nosec G115
			}

			for i, expectedDiffDeltaTimestamps := range test.expectedDiffArrayNodeSubDeltaTimestamps {
				if expectedDiffDeltaTimestamps {
					assert.NotEqual(t, arrayNodeState.SubNodeDeltaTimestamps.GetItem(i), test.subNodeDeltaTimestamps[i])
				} else {
					assert.Equal(t, arrayNodeState.SubNodeDeltaTimestamps.GetItem(i), test.subNodeDeltaTimestamps[i])
				}
			}

			bufferedEventRecorder, ok := eventRecorder.(*bufferedEventRecorder)
			if ok {
				if len(test.expectedExternalResourcePhases) > 0 {
					assert.Equal(t, 1, len(bufferedEventRecorder.taskExecutionEvents))

					externalResources := bufferedEventRecorder.taskExecutionEvents[0].GetMetadata().GetExternalResources()
					assert.Equal(t, len(test.expectedExternalResourcePhases), len(externalResources))
					for i, expectedPhase := range test.expectedExternalResourcePhases {
						assert.Equal(t, expectedPhase, externalResources[i].GetPhase())
					}
				} else {
					assert.Equal(t, 0, len(bufferedEventRecorder.taskExecutionEvents))
				}
			}

			nCtx.ExecutionContext().(*execmocks.ExecutionContext).AssertNumberOfCalls(t, "IncrementParallelism", int(test.incrementParallelismCount))
		})
	}
}

func TestHandle_InvalidLiteralType(t *testing.T) {
	ctx := context.Background()
	scope := promutils.NewTestScope()
	dataStore, err := storage.NewDataStore(&storage.Config{
		Type: storage.TypeMemory,
	}, scope)
	assert.NoError(t, err)
	nodeHandler := &mocks.NodeHandler{}

	// Initialize ArrayNodeHandler
	arrayNodeHandler, err := createArrayNodeHandler(ctx, t, nodeHandler, dataStore, scope)
	assert.NoError(t, err)

	// Test cases
	tests := []struct {
		name                      string
		inputLiteral              *idlcore.Literal
		expectedTransitionType    handler.TransitionType
		expectedPhase             handler.EPhase
		expectedErrorCode         string
		expectedContainedErrorMsg string
	}{
		{
			name: "InvalidLiteralType",
			inputLiteral: &idlcore.Literal{
				Value: &idlcore.Literal_Scalar{
					Scalar: &idlcore.Scalar{},
				},
			},
			expectedTransitionType:    handler.TransitionTypeEphemeral,
			expectedPhase:             handler.EPhaseFailed,
			expectedErrorCode:         errors.IDLNotFoundErr,
			expectedContainedErrorMsg: "Failed to validate literal type",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create NodeExecutionContext
			literalMap := &idlcore.LiteralMap{
				Literals: map[string]*idlcore.Literal{
					"invalidInput": test.inputLiteral,
				},
			}
			arrayNodeState := &handler.ArrayNodeState{
				Phase: v1alpha1.ArrayNodePhaseNone,
			}
			nCtx := createNodeExecutionContext(dataStore, newBufferedEventRecorder(), nil, literalMap, &arrayNodeSpec, arrayNodeState, 0, workflowMaxParallelism)

			// Evaluate node
			transition, err := arrayNodeHandler.Handle(ctx, nCtx)
			assert.NoError(t, err)

			// Validate results
			assert.Equal(t, test.expectedTransitionType, transition.Type())
			assert.Equal(t, test.expectedPhase, transition.Info().GetPhase())
			assert.Equal(t, test.expectedErrorCode, transition.Info().GetErr().GetCode())
			assert.Contains(t, transition.Info().GetErr().GetMessage(), test.expectedContainedErrorMsg)
		})
	}
}

func TestHandleArrayNodePhaseExecutingSubNodeFailures(t *testing.T) {
	ctx := context.Background()

	inputValues := map[string][]int64{
		"foo": []int64{1},
		"bar": []int64{2},
	}
	literalMap := convertMapToArrayLiterals(inputValues)

	tests := []struct {
		name               string
		defaultMaxAttempts int32
		maxSystemFailures  int64
		ignoreRetryCause   bool
		transition         handler.Transition
		expectedAttempts   int
	}{
		{
			name:               "UserFailure",
			defaultMaxAttempts: 3,
			maxSystemFailures:  10,
			ignoreRetryCause:   false,
			transition: handler.DoTransition(handler.TransitionTypeEphemeral,
				handler.PhaseInfoRetryableFailure(idlcore.ExecutionError_USER, "", "", &handler.ExecutionInfo{})),
			expectedAttempts: 3,
		},
		{
			name:               "SystemFailure",
			defaultMaxAttempts: 3,
			maxSystemFailures:  10,
			ignoreRetryCause:   false,
			transition: handler.DoTransition(handler.TransitionTypeEphemeral,
				handler.PhaseInfoRetryableFailure(idlcore.ExecutionError_SYSTEM, "", "", &handler.ExecutionInfo{})),
			expectedAttempts: 11,
		},
		{
			name:               "UserFailureIgnoreRetryCause",
			defaultMaxAttempts: 3,
			maxSystemFailures:  10,
			ignoreRetryCause:   true,
			transition: handler.DoTransition(handler.TransitionTypeEphemeral,
				handler.PhaseInfoRetryableFailure(idlcore.ExecutionError_USER, "", "", &handler.ExecutionInfo{})),
			expectedAttempts: 3,
		},
		{
			name:               "SystemFailureIgnoreRetryCause",
			defaultMaxAttempts: 3,
			maxSystemFailures:  10,
			ignoreRetryCause:   true,
			transition: handler.DoTransition(handler.TransitionTypeEphemeral,
				handler.PhaseInfoRetryableFailure(idlcore.ExecutionError_SYSTEM, "", "", &handler.ExecutionInfo{})),
			expectedAttempts: 3,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config.GetConfig().NodeConfig.DefaultMaxAttempts = test.defaultMaxAttempts
			config.GetConfig().NodeConfig.MaxNodeRetriesOnSystemFailures = test.maxSystemFailures
			config.GetConfig().NodeConfig.IgnoreRetryCause = test.ignoreRetryCause

			// create NodeExecutionContext
			scope := promutils.NewTestScope()
			dataStore, err := storage.NewDataStore(&storage.Config{
				Type: storage.TypeMemory,
			}, scope)
			assert.NoError(t, err)
			eventRecorder := newBufferedEventRecorder()
			arrayNodeState := &handler.ArrayNodeState{
				Phase: v1alpha1.ArrayNodePhaseNone,
			}
			nCtx := createNodeExecutionContext(dataStore, eventRecorder, nil, literalMap, &arrayNodeSpec, arrayNodeState, 0, workflowMaxParallelism)

			// initialize ArrayNodeHandler
			nodeHandler := &mocks.NodeHandler{}
			nodeHandler.EXPECT().Abort(mock.Anything, mock.Anything, mock.Anything).Return(nil)
			nodeHandler.EXPECT().Finalize(mock.Anything, mock.Anything).Return(nil)
			nodeHandler.EXPECT().FinalizeRequired().Return(false)
			nodeHandler.EXPECT().Handle(mock.Anything, mock.Anything).Return(test.transition, nil)

			arrayNodeHandler, err := createArrayNodeHandler(ctx, t, nodeHandler, dataStore, scope)
			assert.NoError(t, err)

			// evaluate node to transition to Executing
			_, err = arrayNodeHandler.Handle(ctx, nCtx)
			assert.NoError(t, err)
			assert.Equal(t, v1alpha1.ArrayNodePhaseExecuting, arrayNodeState.Phase)

			for i := 0; i < len(arrayNodeState.SubNodePhases.GetItems()); i++ {
				arrayNodeState.SubNodePhases.SetItem(i, bitarray.Item(v1alpha1.NodePhaseRunning))
			}

			for i := 0; i < len(arrayNodeState.SubNodeTaskPhases.GetItems()); i++ {
				arrayNodeState.SubNodeTaskPhases.SetItem(i, bitarray.Item(core.PhaseRunning))
			}

			// evaluate node until failure
			attempts := 1
			for {
				nCtx := createNodeExecutionContext(dataStore, eventRecorder, nil, literalMap, &arrayNodeSpec, arrayNodeState, 0, workflowMaxParallelism)
				_, err = arrayNodeHandler.Handle(ctx, nCtx)
				assert.NoError(t, err)

				if arrayNodeState.Phase == v1alpha1.ArrayNodePhaseFailing {
					break
				}

				// failing a task requires two calls to Handle, the first to return a
				// RetryableFailure and the second to abort. therefore, we only increment the
				// number of attempts once in this loop.
				if arrayNodeState.SubNodePhases.GetItem(0) == bitarray.Item(v1alpha1.NodePhaseRetryableFailure) {
					attempts++
				}
			}

			assert.Equal(t, test.expectedAttempts, attempts)
		})
	}
}

func TestHandleArrayNodePhaseSucceeding(t *testing.T) {
	ctx := context.Background()
	scope := promutils.NewTestScope()
	dataStore, err := storage.NewDataStore(&storage.Config{
		Type: storage.TypeMemory,
	}, scope)
	assert.NoError(t, err)
	nodeHandler := &mocks.NodeHandler{}
	valueOne := 1

	// initialize ArrayNodeHandler
	arrayNodeHandler, err := createArrayNodeHandler(ctx, t, nodeHandler, dataStore, scope)
	assert.NoError(t, err)

	tests := []struct {
		name                    string
		outputVariable          string
		outputValues            []*int
		subNodePhases           []v1alpha1.NodePhase
		expectedArrayNodePhase  v1alpha1.ArrayNodePhase
		expectedTransitionPhase handler.EPhase
	}{
		{
			name:           "Success",
			outputValues:   []*int{&valueOne, nil},
			outputVariable: "foo",
			subNodePhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseSucceeded,
				v1alpha1.NodePhaseFailed,
			},
			expectedArrayNodePhase:  v1alpha1.ArrayNodePhaseSucceeding,
			expectedTransitionPhase: handler.EPhaseSuccess,
		},
		{
			name:                    "SuccessEmptyInput",
			outputValues:            []*int{},
			outputVariable:          "foo",
			subNodePhases:           []v1alpha1.NodePhase{},
			expectedArrayNodePhase:  v1alpha1.ArrayNodePhaseSucceeding,
			expectedTransitionPhase: handler.EPhaseSuccess,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// initialize ArrayNodeState
			subNodePhases, err := bitarray.NewCompactArray(uint(len(test.subNodePhases)), bitarray.Item(v1alpha1.NodePhaseRecovered))
			assert.NoError(t, err)
			for i, nodePhase := range test.subNodePhases {
				subNodePhases.SetItem(i, bitarray.Item(nodePhase)) // #nosec G115
			}

			retryAttempts, err := bitarray.NewCompactArray(uint(len(test.subNodePhases)), bitarray.Item(1))
			assert.NoError(t, err)

			arrayNodeState := &handler.ArrayNodeState{
				Phase:                v1alpha1.ArrayNodePhaseSucceeding,
				SubNodePhases:        subNodePhases,
				SubNodeRetryAttempts: retryAttempts,
				Error:                testError,
			}

			// create NodeExecutionContext
			eventRecorder := newBufferedEventRecorder()
			literalMap := &idlcore.LiteralMap{}
			nCtx := createNodeExecutionContext(dataStore, eventRecorder, []string{test.outputVariable}, literalMap, &arrayNodeSpec, arrayNodeState, 0, workflowMaxParallelism)

			// write mocked output files
			for i, outputValue := range test.outputValues {
				if outputValue == nil {
					continue
				}

				outputFile := storage.DataReference(fmt.Sprintf("s3://bucket/output/%d/0/outputs.pb", i))
				outputLiteralMap := &idlcore.LiteralMap{
					Literals: map[string]*idlcore.Literal{
						test.outputVariable: &idlcore.Literal{
							Value: &idlcore.Literal_Scalar{
								Scalar: &idlcore.Scalar{
									Value: &idlcore.Scalar_Primitive{
										Primitive: &idlcore.Primitive{
											Value: &idlcore.Primitive_Integer{
												Integer: int64(*outputValue),
											},
										},
									},
								},
							},
						},
					},
				}

				err := nCtx.DataStore().WriteProtobuf(ctx, outputFile, storage.Options{}, outputLiteralMap)
				assert.NoError(t, err)
			}

			// evaluate node
			transition, err := arrayNodeHandler.Handle(ctx, nCtx)
			assert.NoError(t, err)

			// validate results
			assert.Equal(t, test.expectedArrayNodePhase, arrayNodeState.Phase)
			assert.Equal(t, test.expectedTransitionPhase, transition.Info().GetPhase())

			// validate output file
			var outputs idlcore.LiteralMap
			outputFile := v1alpha1.GetOutputsFile(nCtx.NodeStatus().GetOutputDir())
			err = nCtx.DataStore().ReadProtobuf(ctx, outputFile, &outputs)
			assert.NoError(t, err)

			assert.Len(t, outputs.GetLiterals(), 1)

			collection := outputs.GetLiterals()[test.outputVariable].GetCollection()
			assert.NotNil(t, collection)

			assert.Len(t, collection.GetLiterals(), len(test.outputValues))
			for i, outputValue := range test.outputValues {
				if outputValue == nil {
					assert.NotNil(t, collection.GetLiterals()[i].GetScalar())
				} else {
					assert.Equal(t, int64(*outputValue), collection.GetLiterals()[i].GetScalar().GetPrimitive().GetInteger())
				}
			}

			assert.Equal(t, 1, len(eventRecorder.taskExecutionEvents))
			assert.Equal(t, idlcore.TaskExecution_SUCCEEDED, eventRecorder.taskExecutionEvents[0].GetPhase())
			assert.Nil(t, eventRecorder.taskExecutionEvents[0].GetError())
		})
	}
}

func TestHandleArrayNodePhaseFailing(t *testing.T) {
	ctx := context.Background()
	scope := promutils.NewTestScope()
	dataStore, err := storage.NewDataStore(&storage.Config{
		Type: storage.TypeMemory,
	}, scope)
	assert.NoError(t, err)

	nodeHandler := &mocks.NodeHandler{}
	nodeHandler.EXPECT().Abort(mock.Anything, mock.Anything, mock.Anything).Return(nil)
	nodeHandler.EXPECT().Finalize(mock.Anything, mock.Anything).Return(nil)

	// initialize ArrayNodeHandler
	arrayNodeHandler, err := createArrayNodeHandler(ctx, t, nodeHandler, dataStore, scope)
	assert.NoError(t, err)

	tests := []struct {
		name                    string
		subNodePhases           []v1alpha1.NodePhase
		expectedArrayNodePhase  v1alpha1.ArrayNodePhase
		expectedTransitionPhase handler.EPhase
		expectedAbortCalls      int
	}{
		{
			name: "Success",
			subNodePhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseRunning,
				v1alpha1.NodePhaseSucceeded,
				v1alpha1.NodePhaseFailed,
			},
			expectedArrayNodePhase:  v1alpha1.ArrayNodePhaseFailing,
			expectedTransitionPhase: handler.EPhaseFailed,
			expectedAbortCalls:      1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// initialize ArrayNodeState
			arrayNodeState := &handler.ArrayNodeState{
				Phase: v1alpha1.ArrayNodePhaseFailing,
				Error: testError,
			}

			for _, item := range []struct {
				arrayReference *bitarray.CompactArray
				maxValue       int
			}{
				{arrayReference: &arrayNodeState.SubNodePhases, maxValue: int(v1alpha1.NodePhaseRecovered)},
				{arrayReference: &arrayNodeState.SubNodeTaskPhases, maxValue: len(core.Phases) - 1},
				{arrayReference: &arrayNodeState.SubNodeRetryAttempts, maxValue: 1},
				{arrayReference: &arrayNodeState.SubNodeSystemFailures, maxValue: 1},
				{arrayReference: &arrayNodeState.SubNodeDeltaTimestamps, maxValue: 1024},
			} {
				*item.arrayReference, err = bitarray.NewCompactArray(uint(len(test.subNodePhases)), bitarray.Item(item.maxValue)) // #nosec G115
				assert.NoError(t, err)
			}

			for i, nodePhase := range test.subNodePhases {
				arrayNodeState.SubNodePhases.SetItem(i, bitarray.Item(nodePhase)) // #nosec G115
			}

			// create NodeExecutionContext
			eventRecorder := newBufferedEventRecorder()
			literalMap := &idlcore.LiteralMap{}
			nCtx := createNodeExecutionContext(dataStore, eventRecorder, nil, literalMap, &arrayNodeSpec, arrayNodeState, 0, workflowMaxParallelism)

			// evaluate node
			transition, err := arrayNodeHandler.Handle(ctx, nCtx)
			assert.NoError(t, err)

			// validate results
			assert.Equal(t, test.expectedArrayNodePhase, arrayNodeState.Phase)
			assert.Equal(t, test.expectedTransitionPhase, transition.Info().GetPhase())
			nodeHandler.AssertNumberOfCalls(t, "Abort", test.expectedAbortCalls)

			assert.Equal(t, 1, len(eventRecorder.taskExecutionEvents))
			assert.Equal(t, idlcore.TaskExecution_FAILED, eventRecorder.taskExecutionEvents[0].GetPhase())
			assert.Equal(t, testError.GetMessage(), eventRecorder.taskExecutionEvents[0].GetError().GetMessage())
		})
	}
}

func init() {
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)
}

func convertMapToArrayLiterals(values map[string][]int64) *idlcore.LiteralMap {
	literalMap := make(map[string]*idlcore.Literal)
	for k, v := range values {
		// create LiteralCollection
		literalList := make([]*idlcore.Literal, 0, len(v))
		for _, x := range v {
			literalList = append(literalList, &idlcore.Literal{
				Value: &idlcore.Literal_Scalar{
					Scalar: &idlcore.Scalar{
						Value: &idlcore.Scalar_Primitive{
							Primitive: &idlcore.Primitive{
								Value: &idlcore.Primitive_Integer{
									Integer: x,
								},
							},
						},
					},
				},
			})
		}

		// add LiteralCollection to map
		literalMap[k] = &idlcore.Literal{
			Value: &idlcore.Literal_Collection{
				Collection: &idlcore.LiteralCollection{
					Literals: literalList,
				},
			},
		}
	}

	return &idlcore.LiteralMap{
		Literals: literalMap,
	}
}

func convertListToLiterals(values []int64) *idlcore.Literal {
	literalList := make([]*idlcore.Literal, 0, len(values))
	for _, x := range values {
		literalList = append(literalList, &idlcore.Literal{
			Value: &idlcore.Literal_Scalar{
				Scalar: &idlcore.Scalar{
					Value: &idlcore.Scalar_Primitive{
						Primitive: &idlcore.Primitive{
							Value: &idlcore.Primitive_Integer{
								Integer: x,
							},
						},
					},
				},
			},
		})
	}
	return &idlcore.Literal{
		Value: &idlcore.Literal_Collection{
			Collection: &idlcore.LiteralCollection{
				Literals: literalList,
			},
		},
	}
}
