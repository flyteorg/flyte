package array

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/types"

	idlcore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	pluginmocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	eventmocks "github.com/flyteorg/flyte/flytepropeller/events/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	execmocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/catalog"
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
)

func createArrayNodeHandler(ctx context.Context, t *testing.T, nodeHandler interfaces.NodeHandler, dataStore *storage.DataStore, scope promutils.Scope) (interfaces.NodeHandler, error) {
	// mock components
	adminClient := launchplan.NewFailFastLaunchPlanExecutor()
	enqueueWorkflowFunc := func(workflowID v1alpha1.WorkflowID) {}
	eventConfig := &config.EventConfig{}
	mockEventSink := eventmocks.NewMockEventSink()
	mockHandlerFactory := &mocks.HandlerFactory{}
	mockHandlerFactory.OnGetHandlerMatch(mock.Anything).Return(nodeHandler, nil)
	mockKubeClient := execmocks.NewFakeKubeClient()
	mockRecoveryClient := &recoverymocks.Client{}
	mockSignalClient := &gatemocks.SignalServiceClient{}
	noopCatalogClient := catalog.NOOPCatalog{}

	// create node executor
	nodeExecutor, err := nodes.NewExecutor(ctx, config.GetConfig().NodeConfig, dataStore, enqueueWorkflowFunc, mockEventSink, adminClient,
		adminClient, "s3://bucket/", mockKubeClient, noopCatalogClient, mockRecoveryClient, eventConfig, "clusterID", mockSignalClient, mockHandlerFactory, scope)
	assert.NoError(t, err)

	// return ArrayNodeHandler
	arrayNodeHandler, err := New(nodeExecutor, eventConfig, scope)
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
	nCtx.OnCurrentAttempt().Return(uint32(0))

	// ContextualNodeLookup
	nodeLookup := &execmocks.NodeLookup{}
	nodeLookup.OnFromNodeMatch(mock.Anything).Return(nil, nil)
	nCtx.OnContextualNodeLookup().Return(nodeLookup)

	// DataStore
	nCtx.OnDataStore().Return(dataStore)

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
	nCtx.OnExecutionContext().Return(executionContext)

	// EventsRecorder
	nCtx.OnEventsRecorder().Return(eventRecorder)

	// InputReader
	inputFilePaths := &pluginmocks.InputFilePaths{}
	inputFilePaths.OnGetInputPath().Return(storage.DataReference("s3://bucket/input"))
	nCtx.OnInputReader().Return(
		newStaticInputReader(
			inputFilePaths,
			inputLiteralMap,
		))

	// Node
	nCtx.OnNode().Return(arrayNodeSpec)

	// NodeExecutionMetadata
	nodeExecutionMetadata := &mocks.NodeExecutionMetadata{}
	nodeExecutionMetadata.OnGetNodeExecutionID().Return(&idlcore.NodeExecutionIdentifier{
		NodeId: "foo",
		ExecutionId: &idlcore.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
	})
	nodeExecutionMetadata.OnGetOwnerID().Return(types.NamespacedName{
		Namespace: "wf-namespace",
		Name:      "wf-name",
	})
	nCtx.OnNodeExecutionMetadata().Return(nodeExecutionMetadata)

	// NodeID
	nCtx.OnNodeID().Return("foo")

	// NodeStateReader
	nodeStateReader := &mocks.NodeStateReader{}
	nodeStateReader.OnGetArrayNodeState().Return(*arrayNodeState)
	nCtx.OnNodeStateReader().Return(nodeStateReader)

	// NodeStateWriter
	nodeStateWriter := &mocks.NodeStateWriter{}
	nodeStateWriter.OnPutArrayNodeStateMatch(mock.Anything, mock.Anything).Run(
		func(args mock.Arguments) {
			*arrayNodeState = args.Get(0).(handler.ArrayNodeState)
		},
	).Return(nil)
	nCtx.OnNodeStateWriter().Return(nodeStateWriter)

	// NodeStatus
	nCtx.OnNodeStatus().Return(&v1alpha1.NodeStatus{
		DataDir:   storage.DataReference("s3://bucket/data"),
		OutputDir: storage.DataReference("s3://bucket/output"),
	})

	return nCtx
}

func TestAbort(t *testing.T) {
	ctx := context.Background()
	scope := promutils.NewTestScope()
	dataStore, err := storage.NewDataStore(&storage.Config{
		Type: storage.TypeMemory,
	}, scope)
	assert.NoError(t, err)

	nodeHandler := &mocks.NodeHandler{}
	nodeHandler.OnAbortMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil)
	nodeHandler.OnFinalizeMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// initialize ArrayNodeHandler
	arrayNodeHandler, err := createArrayNodeHandler(ctx, t, nodeHandler, dataStore, scope)
	assert.NoError(t, err)

	tests := []struct {
		name                           string
		inputMap                       map[string][]int64
		subNodePhases                  []v1alpha1.NodePhase
		subNodeTaskPhases              []core.Phase
		expectedExternalResourcePhases []idlcore.TaskExecution_Phase
	}{
		{
			name: "Success",
			inputMap: map[string][]int64{
				"foo": []int64{0, 1, 2},
			},
			subNodePhases:                  []v1alpha1.NodePhase{v1alpha1.NodePhaseSucceeded, v1alpha1.NodePhaseRunning, v1alpha1.NodePhaseNotYetStarted},
			subNodeTaskPhases:              []core.Phase{core.PhaseSuccess, core.PhaseRunning, core.PhaseUndefined},
			expectedExternalResourcePhases: []idlcore.TaskExecution_Phase{idlcore.TaskExecution_ABORTED},
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
			} {

				*item.arrayReference, err = bitarray.NewCompactArray(uint(size), bitarray.Item(item.maxValue))
				assert.NoError(t, err)
			}

			for i, nodePhase := range test.subNodePhases {
				arrayNodeState.SubNodePhases.SetItem(i, bitarray.Item(nodePhase))
			}
			for i, taskPhase := range test.subNodeTaskPhases {
				arrayNodeState.SubNodeTaskPhases.SetItem(i, bitarray.Item(taskPhase))
			}

			// create NodeExecutionContext
			eventRecorder := newBufferedEventRecorder()
			nCtx := createNodeExecutionContext(dataStore, eventRecorder, nil, literalMap, &arrayNodeSpec, arrayNodeState, 0, workflowMaxParallelism)

			// evaluate node
			err := arrayNodeHandler.Abort(ctx, nCtx, "foo")
			assert.NoError(t, err)

			nodeHandler.AssertNumberOfCalls(t, "Abort", len(test.expectedExternalResourcePhases))
			if len(test.expectedExternalResourcePhases) > 0 {
				assert.Equal(t, 1, len(eventRecorder.taskExecutionEvents))

				externalResources := eventRecorder.taskExecutionEvents[0].Metadata.GetExternalResources()
				assert.Equal(t, len(test.expectedExternalResourcePhases), len(externalResources))
				for i, expectedPhase := range test.expectedExternalResourcePhases {
					assert.Equal(t, expectedPhase, externalResources[i].Phase)
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
	nodeHandler.OnFinalizeMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil)

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
			} {

				*item.arrayReference, err = bitarray.NewCompactArray(uint(size), bitarray.Item(item.maxValue))
				assert.NoError(t, err)
			}

			for i, nodePhase := range test.subNodePhases {
				arrayNodeState.SubNodePhases.SetItem(i, bitarray.Item(nodePhase))
			}
			for i, taskPhase := range test.subNodeTaskPhases {
				arrayNodeState.SubNodeTaskPhases.SetItem(i, bitarray.Item(taskPhase))
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
		inputValues                    map[string][]int64
		expectedArrayNodePhase         v1alpha1.ArrayNodePhase
		expectedTransitionPhase        handler.EPhase
		expectedExternalResourcePhases []idlcore.TaskExecution_Phase
	}{
		{
			name: "Success",
			inputValues: map[string][]int64{
				"foo": []int64{1, 2},
			},
			expectedArrayNodePhase:         v1alpha1.ArrayNodePhaseExecuting,
			expectedTransitionPhase:        handler.EPhaseRunning,
			expectedExternalResourcePhases: []idlcore.TaskExecution_Phase{idlcore.TaskExecution_UNDEFINED, idlcore.TaskExecution_UNDEFINED},
		},
		{
			name: "SuccessMultipleInputs",
			inputValues: map[string][]int64{
				"foo": []int64{1, 2, 3},
				"bar": []int64{4, 5, 6},
			},
			expectedArrayNodePhase:         v1alpha1.ArrayNodePhaseExecuting,
			expectedTransitionPhase:        handler.EPhaseRunning,
			expectedExternalResourcePhases: []idlcore.TaskExecution_Phase{idlcore.TaskExecution_UNDEFINED, idlcore.TaskExecution_UNDEFINED, idlcore.TaskExecution_UNDEFINED},
		},
		{
			name: "FailureDifferentInputListLengths",
			inputValues: map[string][]int64{
				"foo": []int64{1, 2},
				"bar": []int64{3},
			},
			expectedArrayNodePhase:         v1alpha1.ArrayNodePhaseNone,
			expectedTransitionPhase:        handler.EPhaseFailed,
			expectedExternalResourcePhases: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create NodeExecutionContext
			eventRecorder := newBufferedEventRecorder()
			literalMap := convertMapToArrayLiterals(test.inputValues)
			arrayNodeState := &handler.ArrayNodeState{
				Phase: v1alpha1.ArrayNodePhaseNone,
			}
			nCtx := createNodeExecutionContext(dataStore, eventRecorder, nil, literalMap, &arrayNodeSpec, arrayNodeState, 0, workflowMaxParallelism)

			// evaluate node
			transition, err := arrayNodeHandler.Handle(ctx, nCtx)
			assert.NoError(t, err)

			// validate results
			assert.Equal(t, test.expectedArrayNodePhase, arrayNodeState.Phase)
			assert.Equal(t, test.expectedTransitionPhase, transition.Info().GetPhase())

			if len(test.expectedExternalResourcePhases) > 0 {
				assert.Equal(t, 1, len(eventRecorder.taskExecutionEvents))

				externalResources := eventRecorder.taskExecutionEvents[0].Metadata.GetExternalResources()
				assert.Equal(t, len(test.expectedExternalResourcePhases), len(externalResources))
				for i, expectedPhase := range test.expectedExternalResourcePhases {
					assert.Equal(t, expectedPhase, externalResources[i].Phase)
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
		name                           string
		parallelism                    *uint32
		minSuccessRatio                *float32
		subNodePhases                  []v1alpha1.NodePhase
		subNodeTaskPhases              []core.Phase
		subNodeTransitions             []handler.Transition
		expectedArrayNodePhase         v1alpha1.ArrayNodePhase
		expectedTransitionPhase        handler.EPhase
		expectedExternalResourcePhases []idlcore.TaskExecution_Phase
		currentWfParallelism           uint32
		maxWfParallelism               uint32
		incrementParallelismCount      uint32
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
			expectedArrayNodePhase:         v1alpha1.ArrayNodePhaseExecuting,
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
			expectedArrayNodePhase:         v1alpha1.ArrayNodePhaseExecuting,
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
			expectedArrayNodePhase:         v1alpha1.ArrayNodePhaseExecuting,
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
			expectedArrayNodePhase:         v1alpha1.ArrayNodePhaseExecuting,
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
			subNodeTransitions:             []handler.Transition{},
			expectedArrayNodePhase:         v1alpha1.ArrayNodePhaseExecuting,
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
			expectedArrayNodePhase:         v1alpha1.ArrayNodePhaseExecuting,
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
			expectedArrayNodePhase:         v1alpha1.ArrayNodePhaseSucceeding,
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
			expectedArrayNodePhase:         v1alpha1.ArrayNodePhaseSucceeding,
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
			expectedArrayNodePhase:         v1alpha1.ArrayNodePhaseFailing,
			expectedTransitionPhase:        handler.EPhaseRunning,
			expectedExternalResourcePhases: []idlcore.TaskExecution_Phase{idlcore.TaskExecution_FAILED, idlcore.TaskExecution_SUCCEEDED},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scope := promutils.NewTestScope()
			dataStore, err := storage.NewDataStore(&storage.Config{
				Type: storage.TypeMemory,
			}, scope)
			assert.NoError(t, err)

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
			} {

				*item.arrayReference, err = bitarray.NewCompactArray(uint(size), bitarray.Item(item.maxValue))
				assert.NoError(t, err)
			}

			for i, nodePhase := range test.subNodePhases {
				arrayNodeState.SubNodePhases.SetItem(i, bitarray.Item(nodePhase))
			}
			for i, taskPhase := range test.subNodeTaskPhases {
				arrayNodeState.SubNodeTaskPhases.SetItem(i, bitarray.Item(taskPhase))
			}

			// create NodeExecutionContext
			eventRecorder := newBufferedEventRecorder()

			nodeSpec := arrayNodeSpec
			nodeSpec.ArrayNode.Parallelism = test.parallelism
			nodeSpec.ArrayNode.MinSuccessRatio = test.minSuccessRatio

			nCtx := createNodeExecutionContext(dataStore, eventRecorder, nil, literalMap, &arrayNodeSpec, arrayNodeState, test.currentWfParallelism, workflowMaxParallelism)

			// initialize ArrayNodeHandler
			nodeHandler := &mocks.NodeHandler{}
			nodeHandler.OnFinalizeRequired().Return(false)
			for i, transition := range test.subNodeTransitions {
				nodeID := fmt.Sprintf("n%d", i)
				transitionPhase := test.expectedExternalResourcePhases[i]

				nodeHandler.OnHandleMatch(mock.Anything, mock.MatchedBy(func(arrayNCtx interfaces.NodeExecutionContext) bool {
					return arrayNCtx.NodeID() == nodeID // match on NodeID using index to ensure each subNode is handled independently
				})).Run(
					func(args mock.Arguments) {
						// mock sending TaskExecutionEvent from handler to show task state transition
						taskExecutionEvent := &event.TaskExecutionEvent{
							Phase: transitionPhase,
						}

						err := args.Get(1).(interfaces.NodeExecutionContext).EventsRecorder().RecordTaskEvent(ctx, taskExecutionEvent, &config.EventConfig{})
						assert.NoError(t, err)
					},
				).Return(transition, nil)
			}

			arrayNodeHandler, err := createArrayNodeHandler(ctx, t, nodeHandler, dataStore, scope)
			assert.NoError(t, err)

			// evaluate node
			transition, err := arrayNodeHandler.Handle(ctx, nCtx)
			assert.NoError(t, err)

			// validate results
			assert.Equal(t, test.expectedArrayNodePhase, arrayNodeState.Phase)
			assert.Equal(t, test.expectedTransitionPhase, transition.Info().GetPhase())

			if len(test.expectedExternalResourcePhases) > 0 {
				assert.Equal(t, 1, len(eventRecorder.taskExecutionEvents))

				externalResources := eventRecorder.taskExecutionEvents[0].Metadata.GetExternalResources()
				assert.Equal(t, len(test.expectedExternalResourcePhases), len(externalResources))
				for i, expectedPhase := range test.expectedExternalResourcePhases {
					assert.Equal(t, expectedPhase, externalResources[i].Phase)
				}
			} else {
				assert.Equal(t, 0, len(eventRecorder.taskExecutionEvents))
			}

			nCtx.ExecutionContext().(*execmocks.ExecutionContext).AssertNumberOfCalls(t, "IncrementParallelism", int(test.incrementParallelismCount))
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
			nodeHandler.OnAbortMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil)
			nodeHandler.OnFinalizeMatch(mock.Anything, mock.Anything).Return(nil)
			nodeHandler.OnFinalizeRequired().Return(false)
			nodeHandler.OnHandleMatch(mock.Anything, mock.Anything).Return(test.transition, nil)

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
				subNodePhases.SetItem(i, bitarray.Item(nodePhase))
			}

			retryAttempts, err := bitarray.NewCompactArray(uint(len(test.subNodePhases)), bitarray.Item(1))
			assert.NoError(t, err)

			arrayNodeState := &handler.ArrayNodeState{
				Phase:                v1alpha1.ArrayNodePhaseSucceeding,
				SubNodePhases:        subNodePhases,
				SubNodeRetryAttempts: retryAttempts,
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
	nodeHandler.OnAbortMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil)
	nodeHandler.OnFinalizeMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil)

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
			}

			for _, item := range []struct {
				arrayReference *bitarray.CompactArray
				maxValue       int
			}{
				{arrayReference: &arrayNodeState.SubNodePhases, maxValue: int(v1alpha1.NodePhaseRecovered)},
				{arrayReference: &arrayNodeState.SubNodeTaskPhases, maxValue: len(core.Phases) - 1},
				{arrayReference: &arrayNodeState.SubNodeRetryAttempts, maxValue: 1},
				{arrayReference: &arrayNodeState.SubNodeSystemFailures, maxValue: 1},
			} {

				*item.arrayReference, err = bitarray.NewCompactArray(uint(len(test.subNodePhases)), bitarray.Item(item.maxValue))
				assert.NoError(t, err)
			}

			for i, nodePhase := range test.subNodePhases {
				arrayNodeState.SubNodePhases.SetItem(i, bitarray.Item(nodePhase))
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
