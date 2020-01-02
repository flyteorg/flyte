package nodes

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/lyft/flytestdlib/promutils/labeled"
	"github.com/lyft/flytestdlib/storage"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	mocks4 "github.com/lyft/flytepropeller/pkg/controller/executors/mocks"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
	nodeHandlerMocks "github.com/lyft/flytepropeller/pkg/controller/nodes/handler/mocks"
	mocks2 "github.com/lyft/flytepropeller/pkg/controller/nodes/mocks"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/task/catalog"

	"github.com/lyft/flyteidl/clients/go/events"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"

	"errors"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/controller/config"
	"github.com/lyft/flytepropeller/pkg/controller/executors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/subworkflow/launchplan"
	"github.com/lyft/flytepropeller/pkg/utils"
	flyteassert "github.com/lyft/flytepropeller/pkg/utils/assert"
)

var fakeKubeClient = mocks4.NewFakeKubeClient()
var catalogClient = catalog.NOOPCatalog{}

const taskID = "tID"

func TestSetInputsForStartNode(t *testing.T) {
	ctx := context.Background()
	mockStorage := createInmemoryDataStore(t, testScope.NewSubScope("f"))
	enQWf := func(workflowID v1alpha1.WorkflowID) {}

	exec, err := NewExecutor(ctx, config.GetConfig().DefaultDeadlines, mockStorage, enQWf, events.NewMockEventSink(), launchplan.NewFailFastLaunchPlanExecutor(), 10, fakeKubeClient, catalogClient, promutils.NewTestScope())
	assert.NoError(t, err)
	inputs := &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"x": utils.MustMakePrimitiveLiteral("hello"),
			"y": utils.MustMakePrimitiveLiteral("blah"),
		},
	}

	t.Run("NoInputs", func(t *testing.T) {
		w := createDummyBaseWorkflow(mockStorage)
		w.DummyStartNode = &v1alpha1.NodeSpec{
			ID: v1alpha1.StartNodeID,
		}
		s, err := exec.SetInputsForStartNode(ctx, w, nil)
		assert.NoError(t, err)
		assert.Equal(t, executors.NodeStatusComplete, s)
	})

	t.Run("WithInputs", func(t *testing.T) {
		w := createDummyBaseWorkflow(mockStorage)
		w.GetNodeExecutionStatus(ctx, v1alpha1.StartNodeID).SetDataDir("s3://test-bucket/exec/start-node/data")
		w.DummyStartNode = &v1alpha1.NodeSpec{
			ID: v1alpha1.StartNodeID,
		}
		s, err := exec.SetInputsForStartNode(ctx, w, inputs)
		assert.NoError(t, err)
		assert.Equal(t, executors.NodeStatusComplete, s)
		actual := &core.LiteralMap{}
		if assert.NoError(t, mockStorage.ReadProtobuf(ctx, "s3://test-bucket/exec/start-node/data/outputs.pb", actual)) {
			flyteassert.EqualLiteralMap(t, inputs, actual)
		}
	})

	t.Run("DataDirNotSet", func(t *testing.T) {
		w := createDummyBaseWorkflow(mockStorage)
		w.DummyStartNode = &v1alpha1.NodeSpec{
			ID: v1alpha1.StartNodeID,
		}
		s, err := exec.SetInputsForStartNode(ctx, w, inputs)
		assert.Error(t, err)
		assert.Equal(t, executors.NodeStatusUndefined, s)
	})

	failStorage := createFailingDatastore(t, testScope.NewSubScope("failing"))
	execFail, err := NewExecutor(ctx, config.GetConfig().DefaultDeadlines, failStorage, enQWf, events.NewMockEventSink(), launchplan.NewFailFastLaunchPlanExecutor(), 10, fakeKubeClient, catalogClient, promutils.NewTestScope())
	assert.NoError(t, err)
	t.Run("StorageFailure", func(t *testing.T) {
		w := createDummyBaseWorkflow(mockStorage)
		w.GetNodeExecutionStatus(ctx, v1alpha1.StartNodeID).SetDataDir("s3://test-bucket/exec/start-node/data")
		w.DummyStartNode = &v1alpha1.NodeSpec{
			ID: v1alpha1.StartNodeID,
		}
		s, err := execFail.SetInputsForStartNode(ctx, w, inputs)
		assert.Error(t, err)
		assert.Equal(t, executors.NodeStatusUndefined, s)
	})
}

func TestNodeExecutor_Initialize(t *testing.T) {
	ctx := context.Background()
	enQWf := func(workflowID v1alpha1.WorkflowID) {
	}

	mockEventSink := events.NewMockEventSink().(*events.MockEventSink)
	memStore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)

	t.Run("happy", func(t *testing.T) {
		execIface, err := NewExecutor(ctx, config.GetConfig().DefaultDeadlines, memStore, enQWf, mockEventSink, launchplan.NewFailFastLaunchPlanExecutor(), 10, fakeKubeClient, catalogClient, promutils.NewTestScope())
		assert.NoError(t, err)
		exec := execIface.(*nodeExecutor)

		hf := &mocks2.HandlerFactory{}
		exec.nodeHandlerFactory = hf

		hf.On("Setup", mock.Anything, mock.Anything).Return(nil)

		assert.NoError(t, exec.Initialize(ctx))
	})

	t.Run("error", func(t *testing.T) {
		execIface, err := NewExecutor(ctx, config.GetConfig().DefaultDeadlines, memStore, enQWf, mockEventSink, launchplan.NewFailFastLaunchPlanExecutor(), 10, fakeKubeClient, catalogClient, promutils.NewTestScope())
		assert.NoError(t, err)
		exec := execIface.(*nodeExecutor)

		hf := &mocks2.HandlerFactory{}
		exec.nodeHandlerFactory = hf

		hf.On("Setup", mock.Anything, mock.Anything).Return(fmt.Errorf("error"))

		assert.Error(t, exec.Initialize(ctx))
	})
}

func TestNodeExecutor_RecursiveNodeHandler_RecurseStartNodes(t *testing.T) {
	ctx := context.Background()
	enQWf := func(workflowID v1alpha1.WorkflowID) {
	}
	mockEventSink := events.NewMockEventSink().(*events.MockEventSink)

	store := createInmemoryDataStore(t, promutils.NewTestScope())

	execIface, err := NewExecutor(ctx, config.GetConfig().DefaultDeadlines, store, enQWf, mockEventSink, launchplan.NewFailFastLaunchPlanExecutor(), 10, fakeKubeClient, catalogClient, promutils.NewTestScope())
	assert.NoError(t, err)
	exec := execIface.(*nodeExecutor)

	defaultNodeID := "n1"

	createStartNodeWf := func(p v1alpha1.NodePhase, _ int) (v1alpha1.ExecutableWorkflow, v1alpha1.ExecutableNode, v1alpha1.ExecutableNodeStatus) {
		startNode := &v1alpha1.NodeSpec{
			Kind: v1alpha1.NodeKindStart,
			ID:   v1alpha1.StartNodeID,
		}
		startNodeStatus := &v1alpha1.NodeStatus{
			Phase: p,
		}
		return &v1alpha1.FlyteWorkflow{
			Status: v1alpha1.WorkflowStatus{
				NodeStatus: map[v1alpha1.NodeID]*v1alpha1.NodeStatus{
					v1alpha1.StartNodeID: startNodeStatus,
				},
				DataDir: "data",
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "wf",
				Nodes: map[v1alpha1.NodeID]*v1alpha1.NodeSpec{
					v1alpha1.StartNodeID: startNode,
				},
				Connections: v1alpha1.Connections{
					UpstreamEdges: map[v1alpha1.NodeID][]v1alpha1.NodeID{
						defaultNodeID: {v1alpha1.StartNodeID},
					},
					DownstreamEdges: map[v1alpha1.NodeID][]v1alpha1.NodeID{
						v1alpha1.StartNodeID: {defaultNodeID},
					},
				},
			},
			DataReferenceConstructor: store,
		}, startNode, startNodeStatus

	}

	// Recurse Child Node Queued previously
	{
		tests := []struct {
			name              string
			currentNodePhase  v1alpha1.NodePhase
			expectedNodePhase v1alpha1.NodePhase
			expectedPhase     executors.NodePhase
			handlerReturn     func() (handler.Transition, error)
			expectedError     bool
		}{
			// Starting at Queued
			{"nys->success", v1alpha1.NodePhaseNotYetStarted, v1alpha1.NodePhaseSucceeded, executors.NodePhaseSuccess, func() (handler.Transition, error) {
				return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoSuccess(nil)), nil
			}, false},
			{"queued->success", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseSucceeded, executors.NodePhaseSuccess, func() (handler.Transition, error) {
				return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoSuccess(nil)), nil
			}, false},
			{"nys->error", v1alpha1.NodePhaseNotYetStarted, v1alpha1.NodePhaseNotYetStarted, executors.NodePhaseUndefined, func() (handler.Transition, error) {
				return handler.UnknownTransition, fmt.Errorf("err")
			}, true},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				hf := &mocks2.HandlerFactory{}
				exec.nodeHandlerFactory = hf

				h := &nodeHandlerMocks.Node{}
				h.On("Handle",
					mock.MatchedBy(func(ctx context.Context) bool { return true }),
					mock.MatchedBy(func(o handler.NodeExecutionContext) bool { return true }),
				).Return(test.handlerReturn())
				h.On("FinalizeRequired").Return(false)

				hf.On("GetHandler", v1alpha1.NodeKindStart).Return(h, nil)

				mockWf, startNode, startNodeStatus := createStartNodeWf(test.currentNodePhase, 0)
				s, err := exec.RecursiveNodeHandler(ctx, mockWf, startNode)
				if test.expectedError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, test.expectedPhase, s.NodePhase, "expected: %s, received %s", test.expectedPhase.String(), s.NodePhase.String())
				assert.Equal(t, uint32(0), startNodeStatus.GetAttempts())
				assert.Equal(t, test.expectedNodePhase, startNodeStatus.GetPhase(), "expected %s, received %s", test.expectedNodePhase.String(), startNodeStatus.GetPhase().String())
			})
		}
	}
}

func TestNodeExecutor_RecursiveNodeHandler_RecurseEndNode(t *testing.T) {
	ctx := context.Background()
	enQWf := func(workflowID v1alpha1.WorkflowID) {
	}
	mockEventSink := events.NewMockEventSink().(*events.MockEventSink)

	store := createInmemoryDataStore(t, promutils.NewTestScope())

	execIface, err := NewExecutor(ctx, config.GetConfig().DefaultDeadlines, store, enQWf, mockEventSink, launchplan.NewFailFastLaunchPlanExecutor(), 10, fakeKubeClient, catalogClient, promutils.NewTestScope())
	assert.NoError(t, err)
	exec := execIface.(*nodeExecutor)

	// Node not yet started
	{
		createSingleNodeWf := func(parentPhase v1alpha1.NodePhase, _ int) (v1alpha1.ExecutableWorkflow, v1alpha1.ExecutableNode, v1alpha1.ExecutableNodeStatus) {
			n := &v1alpha1.NodeSpec{
				ID:   v1alpha1.EndNodeID,
				Kind: v1alpha1.NodeKindEnd,
			}
			ns := &v1alpha1.NodeStatus{}

			return &v1alpha1.FlyteWorkflow{
				Status: v1alpha1.WorkflowStatus{
					NodeStatus: map[v1alpha1.NodeID]*v1alpha1.NodeStatus{
						v1alpha1.EndNodeID: ns,
						v1alpha1.StartNodeID: {
							Phase: parentPhase,
						},
					},
					DataDir: "wf-data",
				},
				WorkflowSpec: &v1alpha1.WorkflowSpec{
					ID: "wf",
					Nodes: map[v1alpha1.NodeID]*v1alpha1.NodeSpec{
						v1alpha1.EndNodeID: n,
					},
					Connections: v1alpha1.Connections{
						UpstreamEdges: map[v1alpha1.NodeID][]v1alpha1.NodeID{
							v1alpha1.EndNodeID: {v1alpha1.StartNodeID},
						},
					},
				},
				DataReferenceConstructor: store,
			}, n, ns

		}
		tests := []struct {
			name              string
			parentNodePhase   v1alpha1.NodePhase
			expectedNodePhase v1alpha1.NodePhase
			expectedPhase     executors.NodePhase
			expectedError     bool
		}{
			{"notYetStarted", v1alpha1.NodePhaseNotYetStarted, v1alpha1.NodePhaseNotYetStarted, executors.NodePhasePending, false},
			{"running", v1alpha1.NodePhaseRunning, v1alpha1.NodePhaseNotYetStarted, executors.NodePhasePending, false},
			{"queued", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseNotYetStarted, executors.NodePhasePending, false},
			{"retryable", v1alpha1.NodePhaseRetryableFailure, v1alpha1.NodePhaseNotYetStarted, executors.NodePhasePending, false},
			{"failing", v1alpha1.NodePhaseFailing, v1alpha1.NodePhaseNotYetStarted, executors.NodePhasePending, false},
			{"skipped", v1alpha1.NodePhaseSkipped, v1alpha1.NodePhaseSkipped, executors.NodePhaseSuccess, false},
			{"success", v1alpha1.NodePhaseSucceeded, v1alpha1.NodePhaseQueued, executors.NodePhaseQueued, false},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				hf := &mocks2.HandlerFactory{}
				exec.nodeHandlerFactory = hf
				h := &nodeHandlerMocks.Node{}
				hf.On("GetHandler", v1alpha1.NodeKindEnd).Return(h, nil)

				mockWf, mockNode, mockNodeStatus := createSingleNodeWf(test.parentNodePhase, 0)
				s, err := exec.RecursiveNodeHandler(ctx, mockWf, mockNode)
				if test.expectedError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, test.expectedPhase, s.NodePhase, "expected: %s, received %s", test.expectedPhase.String(), s.NodePhase.String())
				assert.Equal(t, uint32(0), mockNodeStatus.GetAttempts())
				assert.Equal(t, test.expectedNodePhase, mockNodeStatus.GetPhase(), "expected %s, received %s", test.expectedNodePhase.String(), mockNodeStatus.GetPhase().String())

				if test.expectedNodePhase == v1alpha1.NodePhaseQueued {
					assert.Equal(t, mockNodeStatus.GetDataDir(), storage.DataReference("/wf-data/end-node/data"))
				}
			})
		}
	}

	// Recurse End Node Queued previously
	{
		createSingleNodeWf := func(endNodePhase v1alpha1.NodePhase, _ int) (v1alpha1.ExecutableWorkflow, v1alpha1.ExecutableNode, v1alpha1.ExecutableNodeStatus) {
			n := &v1alpha1.NodeSpec{
				ID:   v1alpha1.EndNodeID,
				Kind: v1alpha1.NodeKindEnd,
			}
			ns := &v1alpha1.NodeStatus{
				Phase: endNodePhase,
			}

			return &v1alpha1.FlyteWorkflow{
				Status: v1alpha1.WorkflowStatus{
					NodeStatus: map[v1alpha1.NodeID]*v1alpha1.NodeStatus{
						v1alpha1.EndNodeID: ns,
						v1alpha1.StartNodeID: {
							Phase: v1alpha1.NodePhaseSucceeded,
						},
					},
					DataDir: "data",
				},
				WorkflowSpec: &v1alpha1.WorkflowSpec{
					ID: "wf",
					Nodes: map[v1alpha1.NodeID]*v1alpha1.NodeSpec{
						v1alpha1.StartNodeID: {
							ID:   v1alpha1.StartNodeID,
							Kind: v1alpha1.NodeKindStart,
						},
						v1alpha1.EndNodeID: n,
					},
					Connections: v1alpha1.Connections{
						UpstreamEdges: map[v1alpha1.NodeID][]v1alpha1.NodeID{
							v1alpha1.EndNodeID: {v1alpha1.StartNodeID},
						},
						DownstreamEdges: map[v1alpha1.NodeID][]v1alpha1.NodeID{
							v1alpha1.StartNodeID: {v1alpha1.EndNodeID},
						},
					},
				},
				DataReferenceConstructor: store,
			}, n, ns

		}
		tests := []struct {
			name              string
			currentNodePhase  v1alpha1.NodePhase
			expectedNodePhase v1alpha1.NodePhase
			expectedPhase     executors.NodePhase
			handlerReturn     func() (handler.Transition, error)
			expectedError     bool
		}{
			// Starting at Queued
			{"queued->success", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseSucceeded, executors.NodePhaseSuccess, func() (handler.Transition, error) {
				return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoSuccess(nil)), nil
			}, false},

			{"queued->failed", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseFailed, executors.NodePhaseFailed, func() (handler.Transition, error) {
				return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure("code", "mesage", nil)), nil
			}, false},

			{"queued->running", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseRunning, executors.NodePhasePending, func() (handler.Transition, error) {
				return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRunning(nil)), nil
			}, false},

			{"queued->error", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseQueued, executors.NodePhaseUndefined, func() (handler.Transition, error) {
				return handler.UnknownTransition, fmt.Errorf("err")
			}, true},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				hf := &mocks2.HandlerFactory{}
				exec.nodeHandlerFactory = hf

				h := &nodeHandlerMocks.Node{}
				h.On("Handle",
					mock.MatchedBy(func(ctx context.Context) bool { return true }),
					mock.MatchedBy(func(o handler.NodeExecutionContext) bool { return true }),
				).Return(test.handlerReturn())
				h.On("FinalizeRequired").Return(false)

				hf.On("GetHandler", v1alpha1.NodeKindEnd).Return(h, nil)

				mockWf, _, mockNodeStatus := createSingleNodeWf(test.currentNodePhase, 0)
				startNode := mockWf.StartNode()
				startStatus := mockWf.GetNodeExecutionStatus(ctx, startNode.GetID())
				assert.Equal(t, v1alpha1.NodePhaseSucceeded, startStatus.GetPhase())
				s, err := exec.RecursiveNodeHandler(ctx, mockWf, startNode)
				if test.expectedError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, test.expectedPhase, s.NodePhase, "expected: %s, received %s", test.expectedPhase.String(), s.NodePhase.String())
				assert.Equal(t, uint32(0), mockNodeStatus.GetAttempts())
				assert.Equal(t, test.expectedNodePhase, mockNodeStatus.GetPhase(), "expected %s, received %s", test.expectedNodePhase.String(), mockNodeStatus.GetPhase().String())
			})
		}
	}
}

func TestNodeExecutor_RecursiveNodeHandler_Recurse(t *testing.T) {
	ctx := context.Background()
	enQWf := func(workflowID v1alpha1.WorkflowID) {
	}
	mockEventSink := events.NewMockEventSink().(*events.MockEventSink)

	defaultNodeID := "n1"
	taskID := taskID

	store := createInmemoryDataStore(t, promutils.NewTestScope())
	createSingleNodeWf := func(p v1alpha1.NodePhase, maxAttempts int) (v1alpha1.ExecutableWorkflow, v1alpha1.ExecutableNode, v1alpha1.ExecutableNodeStatus) {
		n := &v1alpha1.NodeSpec{
			ID:      defaultNodeID,
			TaskRef: &taskID,
			Kind:    v1alpha1.NodeKindTask,
			RetryStrategy: &v1alpha1.RetryStrategy{
				MinAttempts: &maxAttempts,
			},
		}
		ns := &v1alpha1.NodeStatus{
			Phase: p,
		}

		startNode := &v1alpha1.NodeSpec{
			Kind: v1alpha1.NodeKindStart,
			ID:   v1alpha1.StartNodeID,
		}
		return &v1alpha1.FlyteWorkflow{
			Tasks: map[v1alpha1.TaskID]*v1alpha1.TaskSpec{
				taskID: {
					TaskTemplate: &core.TaskTemplate{},
				},
			},
			Status: v1alpha1.WorkflowStatus{
				NodeStatus: map[v1alpha1.NodeID]*v1alpha1.NodeStatus{
					defaultNodeID: ns,
					v1alpha1.StartNodeID: {
						Phase: v1alpha1.NodePhaseSucceeded,
					},
				},
				DataDir: "data",
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "wf",
				Nodes: map[v1alpha1.NodeID]*v1alpha1.NodeSpec{
					defaultNodeID:        n,
					v1alpha1.StartNodeID: startNode,
				},
				Connections: v1alpha1.Connections{
					UpstreamEdges: map[v1alpha1.NodeID][]v1alpha1.NodeID{
						defaultNodeID: {v1alpha1.StartNodeID},
					},
					DownstreamEdges: map[v1alpha1.NodeID][]v1alpha1.NodeID{
						v1alpha1.StartNodeID: {defaultNodeID},
					},
				},
			},
			DataReferenceConstructor: store,
		}, n, ns

	}

	// Recursion test with child Node not yet started
	t.Run("ChildNodeNotYetStarted", func(t *testing.T) {
		nodeN0 := "n0"
		nodeN2 := "n2"
		ctx := context.Background()
		connections := &v1alpha1.Connections{
			UpstreamEdges: map[v1alpha1.NodeID][]v1alpha1.NodeID{
				nodeN2: {nodeN0},
			},
		}

		setupNodePhase := func(n0Phase, n2Phase, expectedN2Phase v1alpha1.NodePhase) (*mocks.ExecutableWorkflow, *mocks.ExecutableNodeStatus) {
			taskID := "id"
			taskID0 := "id1"
			// Setup
			mockN2Status := &mocks.ExecutableNodeStatus{}
			// No parent node
			mockN2Status.On("GetParentNodeID").Return(nil)
			mockN2Status.On("GetParentTaskID").Return(nil)
			mockN2Status.On("GetPhase").Return(n2Phase)
			mockN2Status.On("SetDataDir", mock.AnythingOfType(reflect.TypeOf(storage.DataReference("x")).String()))
			mockN2Status.On("GetDataDir").Return(storage.DataReference("blah"))
			mockN2Status.On("GetWorkflowNodeStatus").Return(nil)
			mockN2Status.On("GetStoppedAt").Return(nil)
			mockN2Status.On("UpdatePhase", expectedN2Phase, mock.Anything, mock.AnythingOfType("string"))
			mockN2Status.On("IsDirty").Return(false)
			mockN2Status.On("GetTaskNodeStatus").Return(nil)
			mockN2Status.On("ClearDynamicNodeStatus").Return(nil)

			mockNode := &mocks.ExecutableNode{}
			mockNode.On("GetID").Return(nodeN2)
			mockNode.On("GetBranchNode").Return(nil)
			mockNode.On("GetKind").Return(v1alpha1.NodeKindTask)
			mockNode.On("IsStartNode").Return(false)
			mockNode.On("IsEndNode").Return(false)
			mockNode.On("GetTaskID").Return(&taskID)
			mockNode.On("GetInputBindings").Return([]*v1alpha1.Binding{})

			mockNodeN0 := &mocks.ExecutableNode{}
			mockNodeN0.On("GetID").Return(nodeN0)
			mockNodeN0.On("GetBranchNode").Return(nil)
			mockNodeN0.On("GetKind").Return(v1alpha1.NodeKindTask)
			mockNodeN0.On("IsStartNode").Return(false)
			mockNodeN0.On("IsEndNode").Return(false)
			mockNodeN0.On("GetTaskID").Return(&taskID0)
			mockN0Status := &mocks.ExecutableNodeStatus{}
			mockN0Status.On("GetPhase").Return(n0Phase)
			mockN0Status.On("IsDirty").Return(false)
			mockN0Status.On("GetParentTaskID").Return(nil)
			n := v1.Now()
			mockN0Status.On("GetStoppedAt").Return(&n)

			tk := &mocks.ExecutableTask{}
			tk.On("CoreTask").Return(&core.TaskTemplate{})
			mockWfStatus := &mocks.ExecutableWorkflowStatus{}
			mockWf := &mocks.ExecutableWorkflow{}
			mockWf.On("StartNode").Return(mockNodeN0)
			mockWf.On("GetNode", nodeN2).Return(mockNode, true)
			mockWf.OnGetNodeExecutionStatusMatch(mock.Anything, nodeN0).Return(mockN0Status)
			mockWf.OnGetNodeExecutionStatusMatch(mock.Anything, nodeN2).Return(mockN2Status)
			mockWf.On("GetConnections").Return(connections)
			mockWf.On("GetID").Return("w1")
			mockWf.On("FromNode", nodeN0).Return([]string{nodeN2}, nil)
			mockWf.On("FromNode", nodeN2).Return([]string{}, fmt.Errorf("did not expect"))
			mockWf.On("GetExecutionID").Return(v1alpha1.WorkflowExecutionIdentifier{})
			mockWf.On("GetExecutionStatus").Return(mockWfStatus)
			mockWf.On("GetTask", taskID0).Return(tk, nil)
			mockWf.On("GetTask", taskID).Return(tk, nil)
			mockWf.On("GetLabels").Return(make(map[string]string))
			mockWfStatus.On("GetDataDir").Return(storage.DataReference("x"))
			return mockWf, mockN2Status
		}

		tests := []struct {
			name              string
			currentNodePhase  v1alpha1.NodePhase
			parentNodePhase   v1alpha1.NodePhase
			expectedNodePhase v1alpha1.NodePhase
			expectedPhase     executors.NodePhase
			expectedError     bool
			updateCalled      bool
		}{
			{"notYetStarted->notYetStarted", v1alpha1.NodePhaseNotYetStarted, v1alpha1.NodePhaseFailed, v1alpha1.NodePhaseNotYetStarted, executors.NodePhaseFailed, false, false},
			{"notYetStarted->skipped", v1alpha1.NodePhaseNotYetStarted, v1alpha1.NodePhaseSkipped, v1alpha1.NodePhaseSkipped, executors.NodePhaseSuccess, false, true},
			{"notYetStarted->queued", v1alpha1.NodePhaseNotYetStarted, v1alpha1.NodePhaseSucceeded, v1alpha1.NodePhaseQueued, executors.NodePhasePending, false, true},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				hf := &mocks2.HandlerFactory{}

				h := &nodeHandlerMocks.Node{}
				h.On("Handle",
					mock.MatchedBy(func(ctx context.Context) bool { return true }),
					mock.MatchedBy(func(o handler.NodeExecutionContext) bool { return true }),
				).Return(handler.UnknownTransition, fmt.Errorf("should not be called"))
				h.On("FinalizeRequired").Return(false)
				hf.On("GetHandler", v1alpha1.NodeKindTask).Return(h, nil)

				mockWf, _ := setupNodePhase(test.parentNodePhase, test.currentNodePhase, test.expectedNodePhase)
				startNode := mockWf.StartNode()
				store := createInmemoryDataStore(t, promutils.NewTestScope())

				execIface, err := NewExecutor(ctx, config.GetConfig().DefaultDeadlines, store, enQWf, mockEventSink, launchplan.NewFailFastLaunchPlanExecutor(), 10, fakeKubeClient, catalogClient, promutils.NewTestScope())
				assert.NoError(t, err)
				exec := execIface.(*nodeExecutor)
				exec.nodeHandlerFactory = hf

				s, err := exec.RecursiveNodeHandler(ctx, mockWf, startNode)
				if test.expectedError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, test.expectedPhase, s.NodePhase, "expected: %s, received %s", test.expectedPhase.String(), s.NodePhase.String())
			})
		}
	})

	// Recurse Child Node Queued previously
	t.Run("ChildNodeQueuedPreviously", func(t *testing.T) {
		tests := []struct {
			name              string
			currentNodePhase  v1alpha1.NodePhase
			expectedNodePhase v1alpha1.NodePhase
			expectedPhase     executors.NodePhase
			handlerReturn     func() (handler.Transition, error)
			finalizeReturnErr bool
			expectedError     bool
			eventRecorded     bool
			eventPhase        core.NodeExecution_Phase
		}{
			// Starting at Queued
			{"queued->running", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseRunning, executors.NodePhasePending, func() (handler.Transition, error) {
				return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRunning(nil)), nil
			}, true, false, true, core.NodeExecution_RUNNING},

			{"queued->queued", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseQueued, executors.NodePhasePending, func() (handler.Transition, error) {
				return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoQueued("reason")), nil
			}, true, false, false, core.NodeExecution_QUEUED},

			{"queued->failing", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseFailing, executors.NodePhasePending, func() (handler.Transition, error) {
				return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure("code", "reason", nil)), nil
			}, true, false, true, core.NodeExecution_FAILED},

			{"failing->failed", v1alpha1.NodePhaseFailing, v1alpha1.NodePhaseFailed, executors.NodePhaseFailed, func() (handler.Transition, error) {
				return handler.UnknownTransition, fmt.Errorf("error")
			}, false, false, false, core.NodeExecution_FAILED},

			{"failing->failed(error)", v1alpha1.NodePhaseFailing, v1alpha1.NodePhaseFailing, executors.NodePhaseUndefined, func() (handler.Transition, error) {
				return handler.UnknownTransition, fmt.Errorf("error")
			}, true, true, false, core.NodeExecution_FAILING},

			{"queued->succeeding", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseSucceeding, executors.NodePhasePending, func() (handler.Transition, error) {
				return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoSuccess(nil)), nil
			}, true, false, true, core.NodeExecution_SUCCEEDED},

			{"succeeding->success", v1alpha1.NodePhaseSucceeding, v1alpha1.NodePhaseSucceeded, executors.NodePhaseSuccess, func() (handler.Transition, error) {
				return handler.UnknownTransition, fmt.Errorf("error")
			}, false, false, false, core.NodeExecution_SUCCEEDED},

			{"succeeding->success(error)", v1alpha1.NodePhaseSucceeding, v1alpha1.NodePhaseSucceeding, executors.NodePhaseUndefined, func() (handler.Transition, error) {

				return handler.UnknownTransition, fmt.Errorf("error")
			}, true, true, false, core.NodeExecution_SUCCEEDED},

			{"queued->error", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseQueued, executors.NodePhaseUndefined, func() (handler.Transition, error) {
				return handler.UnknownTransition, fmt.Errorf("error")
			}, true, true, false, core.NodeExecution_RUNNING},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				hf := &mocks2.HandlerFactory{}

				store := createInmemoryDataStore(t, promutils.NewTestScope())
				execIface, err := NewExecutor(ctx, config.GetConfig().DefaultDeadlines, store, enQWf, mockEventSink, launchplan.NewFailFastLaunchPlanExecutor(), 10, fakeKubeClient, catalogClient, promutils.NewTestScope())
				assert.NoError(t, err)
				exec := execIface.(*nodeExecutor)
				exec.nodeHandlerFactory = hf

				called := false
				exec.nodeRecorder = &events.MockRecorder{
					RecordNodeEventCb: func(ctx context.Context, ev *event.NodeExecutionEvent) error {
						assert.NotNil(t, ev)
						assert.Equal(t, test.eventPhase, ev.Phase)
						called = true
						return nil
					},
				}

				h := &nodeHandlerMocks.Node{}
				h.On("Handle",
					mock.MatchedBy(func(ctx context.Context) bool { return true }),
					mock.MatchedBy(func(o handler.NodeExecutionContext) bool { return true }),
				).Return(test.handlerReturn())
				h.On("FinalizeRequired").Return(true)

				if test.finalizeReturnErr {
					h.On("Finalize", mock.Anything, mock.Anything).Return(fmt.Errorf("error"))
				} else {
					h.On("Finalize", mock.Anything, mock.Anything).Return(nil)
				}
				hf.On("GetHandler", v1alpha1.NodeKindTask).Return(h, nil)

				mockWf, _, mockNodeStatus := createSingleNodeWf(test.currentNodePhase, 0)
				startNode := mockWf.StartNode()
				startStatus := mockWf.GetNodeExecutionStatus(ctx, startNode.GetID())
				assert.Equal(t, v1alpha1.NodePhaseSucceeded, startStatus.GetPhase())
				s, err := exec.RecursiveNodeHandler(ctx, mockWf, startNode)
				if test.expectedError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, test.expectedPhase, s.NodePhase, "expected: %s, received %s", test.expectedPhase.String(), s.NodePhase.String())
				assert.Equal(t, uint32(0), mockNodeStatus.GetAttempts())
				assert.Equal(t, test.expectedNodePhase, mockNodeStatus.GetPhase(), "expected %s, received %s", test.expectedNodePhase.String(), mockNodeStatus.GetPhase().String())
				assert.Equal(t, test.eventRecorded, called, "event recording expected: %v, but got %v", test.eventRecorded, called)
			})
		}
	})

	// Recurse Child Node started previously
	t.Run("ChildNodeStartedPreviously", func(t *testing.T) {
		tests := []struct {
			name              string
			currentNodePhase  v1alpha1.NodePhase
			expectedNodePhase v1alpha1.NodePhase
			expectedPhase     executors.NodePhase
			handlerReturn     func() (handler.Transition, error)
			expectedError     bool
			eventRecorded     bool
			eventPhase        core.NodeExecution_Phase
			attempts          int
		}{
			{"running->running", v1alpha1.NodePhaseRunning, v1alpha1.NodePhaseRunning, executors.NodePhasePending, func() (handler.Transition, error) {
				return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRunning(nil)), nil
			}, false, false, core.NodeExecution_RUNNING, 0},

			{"running->failing", v1alpha1.NodePhaseRunning, v1alpha1.NodePhaseFailing, executors.NodePhasePending,
				func() (handler.Transition, error) {
					return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRetryableFailure("x", "y", nil)), nil
				},
				false, true, core.NodeExecution_FAILED, 0},

			{"(retryablefailure->running", v1alpha1.NodePhaseRetryableFailure, v1alpha1.NodePhaseRunning, executors.NodePhasePending, func() (handler.Transition, error) {
				return handler.UnknownTransition, fmt.Errorf("should not be invoked")
			}, false, false, core.NodeExecution_RUNNING, 0},

			{"running->failing", v1alpha1.NodePhaseRunning, v1alpha1.NodePhaseFailing, executors.NodePhasePending, func() (handler.Transition, error) {
				return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure("code", "reason", nil)), nil
			}, false, true, core.NodeExecution_FAILED, 0},

			{"running->succeeding", v1alpha1.NodePhaseRunning, v1alpha1.NodePhaseSucceeding, executors.NodePhasePending, func() (handler.Transition, error) {
				return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoSuccess(nil)), nil
			}, false, true, core.NodeExecution_SUCCEEDED, 0},

			{"running->error", v1alpha1.NodePhaseRunning, v1alpha1.NodePhaseRunning, executors.NodePhaseUndefined, func() (handler.Transition, error) {
				return handler.UnknownTransition, fmt.Errorf("error")
			}, true, false, core.NodeExecution_RUNNING, 0},

			{"previously-failed", v1alpha1.NodePhaseFailed, v1alpha1.NodePhaseFailed, executors.NodePhaseFailed, func() (handler.Transition, error) {
				return handler.UnknownTransition, fmt.Errorf("error")
			}, false, false, core.NodeExecution_RUNNING, 0},

			{"previously-success", v1alpha1.NodePhaseSucceeded, v1alpha1.NodePhaseSucceeded, executors.NodePhaseComplete, func() (handler.Transition, error) {
				return handler.UnknownTransition, fmt.Errorf("error")
			}, false, false, core.NodeExecution_RUNNING, 0},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				hf := &mocks2.HandlerFactory{}
				store := createInmemoryDataStore(t, promutils.NewTestScope())
				execIface, err := NewExecutor(ctx, config.GetConfig().DefaultDeadlines, store, enQWf, mockEventSink, launchplan.NewFailFastLaunchPlanExecutor(), 10, fakeKubeClient, catalogClient, promutils.NewTestScope())
				assert.NoError(t, err)
				exec := execIface.(*nodeExecutor)
				exec.nodeHandlerFactory = hf

				called := false
				exec.nodeRecorder = &events.MockRecorder{
					RecordNodeEventCb: func(ctx context.Context, ev *event.NodeExecutionEvent) error {
						assert.NotNil(t, ev)
						assert.Equal(t, test.eventPhase.String(), ev.Phase.String())
						called = true
						return nil
					},
				}

				h := &nodeHandlerMocks.Node{}
				h.On("Handle",
					mock.MatchedBy(func(ctx context.Context) bool { return true }),
					mock.MatchedBy(func(o handler.NodeExecutionContext) bool { return true }),
				).Return(test.handlerReturn())
				h.On("FinalizeRequired").Return(true)
				if test.currentNodePhase == v1alpha1.NodePhaseRetryableFailure {
					h.On("Finalize", mock.Anything, mock.Anything).Return(nil)
				} else {
					h.On("Finalize", mock.Anything, mock.Anything).Return(fmt.Errorf("error"))
				}
				hf.On("GetHandler", v1alpha1.NodeKindTask).Return(h, nil)

				mockWf, _, mockNodeStatus := createSingleNodeWf(test.currentNodePhase, 1)
				startNode := mockWf.StartNode()
				s, err := exec.RecursiveNodeHandler(ctx, mockWf, startNode)
				if test.expectedError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, test.expectedPhase, s.NodePhase, "expected: %s, received %s", test.expectedPhase.String(), s.NodePhase.String())
				assert.Equal(t, uint32(test.attempts), mockNodeStatus.GetAttempts())
				assert.Equal(t, test.expectedNodePhase.String(), mockNodeStatus.GetPhase().String(), "expected %s, received %s", test.expectedNodePhase.String(), mockNodeStatus.GetPhase().String())
				assert.Equal(t, test.eventRecorded, called, "event recording expected: %v, but got %v", test.eventRecorded, called)
			})
		}
	})

	// Extinguished retries
	t.Run("retries-exhausted", func(t *testing.T) {
		hf := &mocks2.HandlerFactory{}
		store := createInmemoryDataStore(t, promutils.NewTestScope())
		execIface, err := NewExecutor(ctx, config.GetConfig().DefaultDeadlines, store, enQWf, mockEventSink, launchplan.NewFailFastLaunchPlanExecutor(), 10, fakeKubeClient, catalogClient, promutils.NewTestScope())
		assert.NoError(t, err)
		exec := execIface.(*nodeExecutor)
		exec.nodeHandlerFactory = hf

		h := &nodeHandlerMocks.Node{}
		h.On("Handle",
			mock.MatchedBy(func(ctx context.Context) bool { return true }),
			mock.MatchedBy(func(o handler.NodeExecutionContext) bool { return true }),
		).Return(handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRetryableFailure("x", "y", nil)), nil)
		h.On("FinalizeRequired").Return(true)
		h.On("Finalize", mock.Anything, mock.Anything).Return(fmt.Errorf("error"))
		hf.On("GetHandler", v1alpha1.NodeKindTask).Return(h, nil)

		mockWf, _, mockNodeStatus := createSingleNodeWf(v1alpha1.NodePhaseRunning, 0)
		startNode := mockWf.StartNode()
		s, err := exec.RecursiveNodeHandler(ctx, mockWf, startNode)
		assert.NoError(t, err)
		assert.Equal(t, executors.NodePhasePending.String(), s.NodePhase.String())
		assert.Equal(t, uint32(0), mockNodeStatus.GetAttempts())
		assert.Equal(t, v1alpha1.NodePhaseFailing.String(), mockNodeStatus.GetPhase().String())
	})

	// Remaining retries
	t.Run("retries-exhausted", func(t *testing.T) {
		hf := &mocks2.HandlerFactory{}
		store := createInmemoryDataStore(t, promutils.NewTestScope())
		execIface, err := NewExecutor(ctx, config.GetConfig().DefaultDeadlines, store, enQWf, mockEventSink, launchplan.NewFailFastLaunchPlanExecutor(), 10, fakeKubeClient, catalogClient, promutils.NewTestScope())
		assert.NoError(t, err)
		exec := execIface.(*nodeExecutor)
		exec.nodeHandlerFactory = hf

		h := &nodeHandlerMocks.Node{}
		h.On("Handle",
			mock.MatchedBy(func(ctx context.Context) bool { return true }),
			mock.MatchedBy(func(o handler.NodeExecutionContext) bool { return true }),
		).Return(handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRetryableFailure("x", "y", nil)), nil)
		h.On("FinalizeRequired").Return(true)
		h.On("Finalize", mock.Anything, mock.Anything).Return(fmt.Errorf("error"))
		hf.On("GetHandler", v1alpha1.NodeKindTask).Return(h, nil)

		mockWf, _, mockNodeStatus := createSingleNodeWf(v1alpha1.NodePhaseRunning, 1)
		startNode := mockWf.StartNode()
		s, err := exec.RecursiveNodeHandler(ctx, mockWf, startNode)
		assert.NoError(t, err)
		assert.Equal(t, executors.NodePhasePending.String(), s.NodePhase.String())
		assert.Equal(t, uint32(0), mockNodeStatus.GetAttempts())
		assert.Equal(t, v1alpha1.NodePhaseFailing.String(), mockNodeStatus.GetPhase().String())
	})
}

func TestNodeExecutor_RecursiveNodeHandler_NoDownstream(t *testing.T) {
	ctx := context.Background()
	enQWf := func(workflowID v1alpha1.WorkflowID) {
	}
	mockEventSink := events.NewMockEventSink().(*events.MockEventSink)

	store := createInmemoryDataStore(t, promutils.NewTestScope())

	execIface, err := NewExecutor(ctx, config.GetConfig().DefaultDeadlines, store, enQWf, mockEventSink, launchplan.NewFailFastLaunchPlanExecutor(), 10, fakeKubeClient, catalogClient, promutils.NewTestScope())
	assert.NoError(t, err)
	exec := execIface.(*nodeExecutor)

	defaultNodeID := "n1"
	taskID := "tID"

	createSingleNodeWf := func(p v1alpha1.NodePhase, maxAttempts int) (v1alpha1.ExecutableWorkflow, v1alpha1.ExecutableNode, v1alpha1.ExecutableNodeStatus) {
		n := &v1alpha1.NodeSpec{
			ID:      defaultNodeID,
			TaskRef: &taskID,
			Kind:    v1alpha1.NodeKindTask,
			RetryStrategy: &v1alpha1.RetryStrategy{
				MinAttempts: &maxAttempts,
			},
		}
		ns := &v1alpha1.NodeStatus{
			Phase: p,
		}

		startNode := &v1alpha1.NodeSpec{
			Kind: v1alpha1.NodeKindStart,
			ID:   v1alpha1.StartNodeID,
		}
		return &v1alpha1.FlyteWorkflow{
			Tasks: map[v1alpha1.TaskID]*v1alpha1.TaskSpec{
				taskID: {
					TaskTemplate: &core.TaskTemplate{},
				},
			},
			Status: v1alpha1.WorkflowStatus{
				NodeStatus: map[v1alpha1.NodeID]*v1alpha1.NodeStatus{
					defaultNodeID: ns,
					v1alpha1.StartNodeID: {
						Phase: v1alpha1.NodePhaseSucceeded,
					},
				},
				DataDir: "data",
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "wf",
				Nodes: map[v1alpha1.NodeID]*v1alpha1.NodeSpec{
					defaultNodeID:        n,
					v1alpha1.StartNodeID: startNode,
				},
				Connections: v1alpha1.Connections{
					UpstreamEdges: map[v1alpha1.NodeID][]v1alpha1.NodeID{
						defaultNodeID: {v1alpha1.StartNodeID},
					},
					DownstreamEdges: map[v1alpha1.NodeID][]v1alpha1.NodeID{
						v1alpha1.StartNodeID: {defaultNodeID},
					},
				},
			},
			DataReferenceConstructor: store,
		}, n, ns

	}

	// Node failed or succeeded
	{
		tests := []struct {
			name              string
			currentNodePhase  v1alpha1.NodePhase
			expectedNodePhase v1alpha1.NodePhase
			expectedPhase     executors.NodePhase
			expectedError     bool
		}{
			{"succeeded", v1alpha1.NodePhaseSucceeded, v1alpha1.NodePhaseSucceeded, executors.NodePhaseComplete, false},
			{"failed", v1alpha1.NodePhaseFailed, v1alpha1.NodePhaseFailed, executors.NodePhaseFailed, false},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				hf := &mocks2.HandlerFactory{}
				exec.nodeHandlerFactory = hf

				h := &nodeHandlerMocks.Node{}
				h.On("Handle",
					mock.MatchedBy(func(ctx context.Context) bool { return true }),
					mock.MatchedBy(func(o handler.NodeExecutionContext) bool { return true }),
				).Return(handler.UnknownTransition, fmt.Errorf("should not be called"))
				h.On("FinalizeRequired").Return(true)
				h.On("Finalize", mock.Anything, mock.Anything).Return(fmt.Errorf("error"))

				hf.On("GetHandler", v1alpha1.NodeKindTask).Return(h, nil)

				mockWf, mockNode, mockNodeStatus := createSingleNodeWf(test.currentNodePhase, 1)
				s, err := exec.RecursiveNodeHandler(ctx, mockWf, mockNode)
				if test.expectedError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, test.expectedPhase, s.NodePhase, "expected: %s, received %s", test.expectedPhase.String(), s.NodePhase.String())
				assert.Equal(t, test.expectedNodePhase, mockNodeStatus.GetPhase(), "expected %s, received %s", test.expectedNodePhase.String(), mockNodeStatus.GetPhase().String())
			})
		}
	}
}

func TestNodeExecutor_RecursiveNodeHandler_UpstreamNotReady(t *testing.T) {
	ctx := context.Background()
	enQWf := func(workflowID v1alpha1.WorkflowID) {
	}
	mockEventSink := events.NewMockEventSink().(*events.MockEventSink)

	store := createInmemoryDataStore(t, promutils.NewTestScope())

	execIface, err := NewExecutor(ctx, config.GetConfig().DefaultDeadlines, store, enQWf, mockEventSink, launchplan.NewFailFastLaunchPlanExecutor(), 10, fakeKubeClient, catalogClient, promutils.NewTestScope())
	assert.NoError(t, err)
	exec := execIface.(*nodeExecutor)

	defaultNodeID := "n1"
	taskID := taskID

	createSingleNodeWf := func(parentPhase v1alpha1.NodePhase, maxAttempts int) (v1alpha1.ExecutableWorkflow, v1alpha1.ExecutableNode, v1alpha1.ExecutableNodeStatus) {
		n := &v1alpha1.NodeSpec{
			ID:      defaultNodeID,
			TaskRef: &taskID,
			Kind:    v1alpha1.NodeKindTask,
			RetryStrategy: &v1alpha1.RetryStrategy{
				MinAttempts: &maxAttempts,
			},
		}
		ns := &v1alpha1.NodeStatus{}

		return &v1alpha1.FlyteWorkflow{
			Tasks: map[v1alpha1.TaskID]*v1alpha1.TaskSpec{
				taskID: {
					TaskTemplate: &core.TaskTemplate{},
				},
			},
			Status: v1alpha1.WorkflowStatus{
				NodeStatus: map[v1alpha1.NodeID]*v1alpha1.NodeStatus{
					defaultNodeID: ns,
					v1alpha1.StartNodeID: {
						Phase: parentPhase,
					},
				},
				DataDir: "data",
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "wf",
				Nodes: map[v1alpha1.NodeID]*v1alpha1.NodeSpec{
					defaultNodeID: n,
				},
				Connections: v1alpha1.Connections{
					UpstreamEdges: map[v1alpha1.NodeID][]v1alpha1.NodeID{
						defaultNodeID: {v1alpha1.StartNodeID},
					},
				},
			},
			DataReferenceConstructor: store,
		}, n, ns

	}

	// Node not yet started
	{
		tests := []struct {
			name              string
			parentNodePhase   v1alpha1.NodePhase
			expectedNodePhase v1alpha1.NodePhase
			expectedPhase     executors.NodePhase
			expectedError     bool
		}{
			{"notYetStarted", v1alpha1.NodePhaseNotYetStarted, v1alpha1.NodePhaseNotYetStarted, executors.NodePhasePending, false},
			{"running", v1alpha1.NodePhaseRunning, v1alpha1.NodePhaseNotYetStarted, executors.NodePhasePending, false},
			{"queued", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseNotYetStarted, executors.NodePhasePending, false},
			{"retryable", v1alpha1.NodePhaseRetryableFailure, v1alpha1.NodePhaseNotYetStarted, executors.NodePhasePending, false},
			{"failing", v1alpha1.NodePhaseFailing, v1alpha1.NodePhaseNotYetStarted, executors.NodePhasePending, false},
			{"failing", v1alpha1.NodePhaseSucceeding, v1alpha1.NodePhaseNotYetStarted, executors.NodePhasePending, false},
			{"skipped", v1alpha1.NodePhaseSkipped, v1alpha1.NodePhaseSkipped, executors.NodePhaseSuccess, false},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				hf := &mocks2.HandlerFactory{}
				exec.nodeHandlerFactory = hf
				h := &nodeHandlerMocks.Node{}
				h.On("Handle",
					mock.MatchedBy(func(ctx context.Context) bool { return true }),
					mock.MatchedBy(func(o handler.NodeExecutionContext) bool { return true }),
				).Return(handler.UnknownTransition, fmt.Errorf("should not be called"))
				h.On("FinalizeRequired").Return(true)
				h.On("Finalize", mock.Anything, mock.Anything).Return(fmt.Errorf("error"))

				hf.On("GetHandler", v1alpha1.NodeKindTask).Return(h, nil)

				mockWf, mockNode, mockNodeStatus := createSingleNodeWf(test.parentNodePhase, 0)
				s, err := exec.RecursiveNodeHandler(ctx, mockWf, mockNode)
				if test.expectedError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, test.expectedPhase, s.NodePhase, "expected: %s, received %s", test.expectedPhase.String(), s.NodePhase.String())
				assert.Equal(t, uint32(0), mockNodeStatus.GetAttempts())
				assert.Equal(t, test.expectedNodePhase, mockNodeStatus.GetPhase(), "expected %s, received %s", test.expectedNodePhase.String(), mockNodeStatus.GetPhase().String())
			})
		}
	}
}

func Test_nodeExecutor_RecordTransitionLatency(t *testing.T) {
	testScope := promutils.NewTestScope()
	type fields struct {
		nodeHandlerFactory HandlerFactory
		enqueueWorkflow    v1alpha1.EnqueueWorkflow
		store              *storage.DataStore
		nodeRecorder       events.NodeEventRecorder
		metrics            *nodeMetrics
	}
	type args struct {
		w          v1alpha1.ExecutableWorkflow
		node       v1alpha1.ExecutableNode
		nodeStatus v1alpha1.ExecutableNodeStatus
	}

	nsf := func(phase v1alpha1.NodePhase, lastUpdated *time.Time) *mocks.ExecutableNodeStatus {
		ns := &mocks.ExecutableNodeStatus{}
		ns.On("GetPhase").Return(phase)
		var t *v1.Time
		if lastUpdated != nil {
			t = &v1.Time{Time: *lastUpdated}
		}
		ns.On("GetLastUpdatedAt").Return(t)
		return ns
	}
	testTime := time.Now()
	tests := []struct {
		name              string
		fields            fields
		args              args
		recordingExpected bool
	}{
		{
			"retryable-failure",
			fields{metrics: &nodeMetrics{TransitionLatency: labeled.NewStopWatch("test", "xyz", time.Millisecond, testScope)}},
			args{nodeStatus: nsf(v1alpha1.NodePhaseRetryableFailure, &testTime)},
			true,
		},
		{
			"retryable-failure-notime",
			fields{metrics: &nodeMetrics{TransitionLatency: labeled.NewStopWatch("test2", "xyz", time.Millisecond, testScope)}},
			args{nodeStatus: nsf(v1alpha1.NodePhaseRetryableFailure, nil)},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &nodeExecutor{
				nodeHandlerFactory: tt.fields.nodeHandlerFactory,
				enqueueWorkflow:    tt.fields.enqueueWorkflow,
				store:              tt.fields.store,
				nodeRecorder:       tt.fields.nodeRecorder,
				metrics:            tt.fields.metrics,
			}
			c.RecordTransitionLatency(context.TODO(), tt.args.w, tt.args.node, tt.args.nodeStatus)

			ch := make(chan prometheus.Metric, 2)
			tt.fields.metrics.TransitionLatency.Collect(ch)
			assert.Equal(t, len(ch) == 1, tt.recordingExpected)
		})
	}
}

func Test_nodeExecutor_timeout(t *testing.T) {
	tests := []struct {
		name              string
		phaseInfo         handler.PhaseInfo
		expectedPhase     handler.EPhase
		activeDeadline    time.Duration
		executionDeadline time.Duration
		err               error
	}{
		{
			name:              "timeout",
			phaseInfo:         handler.PhaseInfoRunning(nil),
			expectedPhase:     handler.EPhaseTimedout,
			activeDeadline:    time.Second * 5,
			executionDeadline: time.Second * 5,
			err:               nil,
		},
		{
			name:              "retryable-failure",
			phaseInfo:         handler.PhaseInfoRunning(nil),
			expectedPhase:     handler.EPhaseRetryableFailure,
			activeDeadline:    time.Second * 15,
			executionDeadline: time.Second * 5,
			err:               nil,
		},
		{
			name:              "expired-but-terminal-phase",
			phaseInfo:         handler.PhaseInfoSuccess(nil),
			expectedPhase:     handler.EPhaseSuccess,
			activeDeadline:    time.Second * 10,
			executionDeadline: time.Second * 5,
			err:               nil,
		},
		{
			name:              "not-expired",
			phaseInfo:         handler.PhaseInfoRunning(nil),
			expectedPhase:     handler.EPhaseRunning,
			activeDeadline:    time.Second * 15,
			executionDeadline: time.Second * 15,
			err:               nil,
		},
		{
			name:              "handler-failure",
			phaseInfo:         handler.PhaseInfoRunning(nil),
			expectedPhase:     handler.EPhaseUndefined,
			activeDeadline:    time.Second * 15,
			executionDeadline: time.Second * 15,
			err:               errors.New("test-error"),
		},
	}
	// mocking status
	queuedAt := time.Now().Add(-1 * time.Second * 10)
	ns := &mocks.ExecutableNodeStatus{}
	queuedAtTime := &v1.Time{Time: queuedAt}
	ns.On("GetQueuedAt").Return(queuedAtTime)
	ns.On("GetLastAttemptStartedAt").Return(queuedAtTime)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &nodeExecutor{}
			handlerReturn := func() (handler.Transition, error) {
				return handler.DoTransition(handler.TransitionTypeEphemeral, tt.phaseInfo), tt.err
			}
			h := &nodeHandlerMocks.Node{}
			h.On("Handle",
				mock.MatchedBy(func(ctx context.Context) bool { return true }),
				mock.MatchedBy(func(o handler.NodeExecutionContext) bool { return true }),
			).Return(handlerReturn())
			h.On("FinalizeRequired").Return(true)
			h.On("Finalize", mock.Anything, mock.Anything).Return(nil)

			hf := &mocks2.HandlerFactory{}
			hf.On("GetHandler", v1alpha1.NodeKindStart).Return(h, nil)
			c.nodeHandlerFactory = hf

			mockNode := &mocks.ExecutableNode{}
			mockNode.On("GetID").Return("node")
			mockNode.On("GetBranchNode").Return(nil)
			mockNode.On("GetKind").Return(v1alpha1.NodeKindTask)
			mockNode.On("IsStartNode").Return(false)
			mockNode.On("IsEndNode").Return(false)
			mockNode.On("GetInputBindings").Return([]*v1alpha1.Binding{})
			mockNode.On("GetActiveDeadline").Return(&tt.activeDeadline)
			mockNode.On("GetExecutionDeadline").Return(&tt.executionDeadline)

			nCtx := &execContext{node: mockNode, nsm: &nodeStateManager{}}
			phaseInfo, err := c.execute(context.TODO(), h, nCtx, ns)

			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedPhase, phaseInfo.GetPhase())
		})
	}
}

func Test_nodeExecutor_abort(t *testing.T) {
	ctx := context.Background()
	exec := nodeExecutor{}
	nCtx := &execContext{}

	t.Run("abort error calls finalize", func(t *testing.T) {
		h := &nodeHandlerMocks.Node{}
		h.OnAbortMatch(mock.Anything, mock.Anything, mock.Anything).Return(errors.New("test error"))
		h.OnFinalizeRequired().Return(true)
		var called bool
		h.OnFinalizeMatch(mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			called = true
		}).Return(nil)

		err := exec.abort(ctx, h, nCtx, "testing")
		assert.Equal(t, "test error", err.Error())
		assert.True(t, called)
	})

	t.Run("abort error calls finalize with error", func(t *testing.T) {
		h := &nodeHandlerMocks.Node{}
		h.OnAbortMatch(mock.Anything, mock.Anything, mock.Anything).Return(errors.New("test error"))
		h.OnFinalizeRequired().Return(true)
		var called bool
		h.OnFinalizeMatch(mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			called = true
		}).Return(errors.New("finalize error"))

		err := exec.abort(ctx, h, nCtx, "testing")
		assert.Equal(t, "0: test error\r\n1: finalize error\r\n", err.Error())
		assert.True(t, called)
	})

	t.Run("abort calls finalize when no errors", func(t *testing.T) {
		h := &nodeHandlerMocks.Node{}
		h.OnAbortMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil)
		h.OnFinalizeRequired().Return(true)
		var called bool
		h.OnFinalizeMatch(mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			called = true
		}).Return(nil)

		err := exec.abort(ctx, h, nCtx, "testing")
		assert.NoError(t, err)
		assert.True(t, called)
	})
}
