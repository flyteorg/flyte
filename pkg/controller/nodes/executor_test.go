package nodes

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	mocks4 "github.com/lyft/flytepropeller/pkg/controller/executors/mocks"

	eventsErr "github.com/lyft/flyteidl/clients/go/events/errors"

	"github.com/golang/protobuf/proto"
	"github.com/lyft/flyteidl/clients/go/events"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteplugins/go/tasks/v1/flytek8s"
	pluginV1 "github.com/lyft/flyteplugins/go/tasks/v1/types"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/promutils/labeled"
	"github.com/lyft/flytestdlib/storage"
	goerrors "github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	"github.com/lyft/flytepropeller/pkg/controller/catalog"
	"github.com/lyft/flytepropeller/pkg/controller/executors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
	mocks3 "github.com/lyft/flytepropeller/pkg/controller/nodes/handler/mocks"
	mocks2 "github.com/lyft/flytepropeller/pkg/controller/nodes/mocks"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/subworkflow/launchplan"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/task"
	"github.com/lyft/flytepropeller/pkg/utils"
	flyteassert "github.com/lyft/flytepropeller/pkg/utils/assert"
)

var fakeKubeClient = mocks4.NewFakeKubeClient()

func createSingletonTaskExecutorFactory() task.Factory {
	return &task.FactoryFuncs{
		GetTaskExecutorCb: func(taskType v1alpha1.TaskType) (pluginV1.Executor, error) {
			return nil, nil
		},
		ListAllTaskExecutorsCb: func() []pluginV1.Executor {
			return []pluginV1.Executor{}
		},
	}
}

func init() {
	flytek8s.InitializeFake()
}

func TestSetInputsForStartNode(t *testing.T) {
	ctx := context.Background()
	mockStorage := createInmemoryDataStore(t, testScope.NewSubScope("f"))
	catalogClient, _ := catalog.NewCatalogClient(ctx, mockStorage)
	enQWf := func(workflowID v1alpha1.WorkflowID) {}

	factory := createSingletonTaskExecutorFactory()
	task.SetTestFactory(factory)
	assert.True(t, task.IsTestModeEnabled())

	exec, err := NewExecutor(ctx, mockStorage, enQWf, time.Second, events.NewMockEventSink(), launchplan.NewFailFastLaunchPlanExecutor(), catalogClient, fakeKubeClient, promutils.NewTestScope())
	assert.NoError(t, err)
	inputs := &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"x": utils.MustMakePrimitiveLiteral("hello"),
			"y": utils.MustMakePrimitiveLiteral("blah"),
		},
	}

	t.Run("NoInputs", func(t *testing.T) {
		w := createDummyBaseWorkflow()
		w.DummyStartNode = &v1alpha1.NodeSpec{
			ID: v1alpha1.StartNodeID,
		}
		s, err := exec.SetInputsForStartNode(ctx, w, nil)
		assert.NoError(t, err)
		assert.Equal(t, executors.NodeStatusComplete, s)
	})

	t.Run("WithInputs", func(t *testing.T) {
		w := createDummyBaseWorkflow()
		w.GetNodeExecutionStatus(v1alpha1.StartNodeID).SetDataDir("s3://test-bucket/exec/start-node/data")
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
		w := createDummyBaseWorkflow()
		w.DummyStartNode = &v1alpha1.NodeSpec{
			ID: v1alpha1.StartNodeID,
		}
		s, err := exec.SetInputsForStartNode(ctx, w, inputs)
		assert.Error(t, err)
		assert.Equal(t, executors.NodeStatusUndefined, s)
	})

	failStorage := createFailingDatastore(t, testScope.NewSubScope("failing"))
	execFail, err := NewExecutor(ctx, failStorage, enQWf, time.Second, events.NewMockEventSink(), launchplan.NewFailFastLaunchPlanExecutor(), catalogClient, fakeKubeClient, promutils.NewTestScope())
	assert.NoError(t, err)
	t.Run("StorageFailure", func(t *testing.T) {
		w := createDummyBaseWorkflow()
		w.GetNodeExecutionStatus(v1alpha1.StartNodeID).SetDataDir("s3://test-bucket/exec/start-node/data")
		w.DummyStartNode = &v1alpha1.NodeSpec{
			ID: v1alpha1.StartNodeID,
		}
		s, err := execFail.SetInputsForStartNode(ctx, w, inputs)
		assert.Error(t, err)
		assert.Equal(t, executors.NodeStatusUndefined, s)
	})
}

func TestNodeExecutor_TransitionToPhase(t *testing.T) {
	ctx := context.Background()
	enQWf := func(workflowID v1alpha1.WorkflowID) {
	}
	mockEventSink := events.NewMockEventSink().(*events.MockEventSink)

	factory := createSingletonTaskExecutorFactory()
	task.SetTestFactory(factory)
	assert.True(t, task.IsTestModeEnabled())

	memStore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)

	catalogClient, _ := catalog.NewCatalogClient(ctx, memStore)
	execIface, err := NewExecutor(ctx, memStore, enQWf, time.Second, mockEventSink, launchplan.NewFailFastLaunchPlanExecutor(), catalogClient, fakeKubeClient, promutils.NewTestScope())
	assert.NoError(t, err)
	exec := execIface.(*nodeExecutor)
	execID := &core.WorkflowExecutionIdentifier{}
	nodeID := "n1"

	expectedErr := fmt.Errorf("test err")
	taskErr := fmt.Errorf("task failed")

	// TABLE Tests
	tests := []struct {
		name               string
		nodeStatus         v1alpha1.ExecutableNodeStatus
		toStatus           handler.Status
		expectedErr        bool
		expectedNodeStatus executors.NodeStatus
	}{
		{"notStarted", &v1alpha1.NodeStatus{Phase: v1alpha1.NodePhaseNotYetStarted}, handler.StatusNotStarted, false, executors.NodeStatusPending},
		{"running", &v1alpha1.NodeStatus{Phase: v1alpha1.NodePhaseNotYetStarted}, handler.StatusRunning, false, executors.NodeStatusRunning},
		{"runningRepeated", &v1alpha1.NodeStatus{Phase: v1alpha1.NodePhaseRunning}, handler.StatusRunning, false, executors.NodeStatusRunning},
		{"success", &v1alpha1.NodeStatus{Phase: v1alpha1.NodePhaseRunning}, handler.StatusSuccess, false, executors.NodeStatusSuccess},
		{"succeeding", &v1alpha1.NodeStatus{Phase: v1alpha1.NodePhaseRunning}, handler.StatusSucceeding, false, executors.NodeStatusRunning},
		{"failing", &v1alpha1.NodeStatus{Phase: v1alpha1.NodePhaseRunning}, handler.StatusFailing(nil), false, executors.NodeStatusRunning},
		{"failed", &v1alpha1.NodeStatus{Phase: v1alpha1.NodePhaseFailing}, handler.StatusFailed(taskErr), false, executors.NodeStatusFailed(taskErr)},
		{"undefined", &v1alpha1.NodeStatus{Phase: v1alpha1.NodePhaseNotYetStarted}, handler.StatusUndefined, true, executors.NodeStatusUndefined},
		{"skipped", &v1alpha1.NodeStatus{Phase: v1alpha1.NodePhaseNotYetStarted}, handler.StatusSkipped, false, executors.NodeStatusSuccess},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			node := &mocks.ExecutableNode{}
			node.On("GetID").Return(nodeID)
			n, err := exec.TransitionToPhase(ctx, execID, node, test.nodeStatus, test.toStatus)
			if test.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expectedNodeStatus, n)
		})
	}

	// Testing retries
	t.Run("noRetryAttemptSet", func(t *testing.T) {
		now := v1.Now()
		status := &mocks.ExecutableNodeStatus{}
		status.On("GetPhase").Return(v1alpha1.NodePhaseRunning)
		status.On("GetAttempts").Return(uint32(0))
		status.On("GetDataDir").Return(storage.DataReference("x"))
		status.On("IncrementAttempts").Return(uint32(1))
		status.On("UpdatePhase", v1alpha1.NodePhaseFailed, mock.Anything, mock.AnythingOfType("string"))
		status.On("GetQueuedAt").Return(&now)
		status.On("GetStartedAt").Return(&now)
		status.On("GetStoppedAt").Return(&now)
		status.On("GetWorkflowNodeStatus").Return(nil)

		node := &mocks.ExecutableNode{}
		node.On("GetID").Return(nodeID)
		node.On("GetRetryStrategy").Return(nil)

		n, err := exec.TransitionToPhase(ctx, execID, node, status, handler.StatusRetryableFailure(fmt.Errorf("failed")))
		assert.NoError(t, err)
		assert.Equal(t, executors.NodePhaseFailed, n.NodePhase)
	})

	// Testing retries
	t.Run("maxAttempt0", func(t *testing.T) {
		now := v1.Now()
		status := &mocks.ExecutableNodeStatus{}
		status.On("GetPhase").Return(v1alpha1.NodePhaseRunning)
		status.On("GetAttempts").Return(uint32(0))
		status.On("GetDataDir").Return(storage.DataReference("x"))
		status.On("IncrementAttempts").Return(uint32(1))
		status.On("UpdatePhase", v1alpha1.NodePhaseFailed, mock.Anything, mock.AnythingOfType("string"))
		status.On("GetQueuedAt").Return(&now)
		status.On("GetStartedAt").Return(&now)
		status.On("GetStoppedAt").Return(&now)
		status.On("GetWorkflowNodeStatus").Return(nil)

		maxAttempts := 0
		node := &mocks.ExecutableNode{}
		node.On("GetID").Return(nodeID)
		node.On("GetRetryStrategy").Return(&v1alpha1.RetryStrategy{MinAttempts: &maxAttempts})

		n, err := exec.TransitionToPhase(ctx, execID, node, status, handler.StatusRetryableFailure(fmt.Errorf("failed")))
		assert.NoError(t, err)
		assert.Equal(t, executors.NodePhaseFailed, n.NodePhase)
	})

	// Testing retries
	t.Run("retryAttemptsRemaining", func(t *testing.T) {
		now := v1.Now()
		status := &mocks.ExecutableNodeStatus{}
		status.On("GetPhase").Return(v1alpha1.NodePhaseRunning)
		status.On("GetAttempts").Return(uint32(0))
		status.On("GetDataDir").Return(storage.DataReference("x"))
		status.On("IncrementAttempts").Return(uint32(1))
		status.On("UpdatePhase", v1alpha1.NodePhaseRetryableFailure, mock.Anything, mock.AnythingOfType("string"))
		status.On("GetQueuedAt").Return(&now)
		status.On("GetStartedAt").Return(&now)
		status.On("GetLastUpdatedAt").Return(&now)
		status.On("GetWorkflowNodeStatus").Return(nil)
		var s *v1alpha1.TaskNodeStatus
		status.On("UpdateTaskNodeStatus", s).Times(10)
		status.On("ClearTaskStatus").Return()
		status.On("ClearWorkflowStatus").Return()
		status.On("ClearDynamicNodeStatus").Return()

		maxAttempts := 2
		node := &mocks.ExecutableNode{}
		node.On("GetID").Return(nodeID)
		node.On("GetRetryStrategy").Return(&v1alpha1.RetryStrategy{MinAttempts: &maxAttempts})

		n, err := exec.TransitionToPhase(ctx, execID, node, status, handler.StatusRetryableFailure(fmt.Errorf("failed")))
		assert.NoError(t, err)
		assert.Equal(t, executors.NodePhaseRunning, n.NodePhase, "%+v", n)
	})

	// Testing retries
	t.Run("retriesExhausted", func(t *testing.T) {
		now := v1.Now()
		status := &mocks.ExecutableNodeStatus{}
		status.On("GetPhase").Return(v1alpha1.NodePhaseRunning)
		status.On("GetAttempts").Return(uint32(0))
		status.On("GetDataDir").Return(storage.DataReference("x"))
		// Change to return 3
		status.On("IncrementAttempts").Return(uint32(3))
		status.On("UpdatePhase", v1alpha1.NodePhaseFailed, mock.Anything, mock.AnythingOfType("string"))
		status.On("GetQueuedAt").Return(&now)
		status.On("GetStartedAt").Return(&now)
		status.On("GetStoppedAt").Return(&now)
		status.On("GetWorkflowNodeStatus").Return(nil)

		maxAttempts := 2
		node := &mocks.ExecutableNode{}
		node.On("GetID").Return(nodeID)
		node.On("GetRetryStrategy").Return(&v1alpha1.RetryStrategy{MinAttempts: &maxAttempts})

		n, err := exec.TransitionToPhase(ctx, execID, node, status, handler.StatusRetryableFailure(fmt.Errorf("failed")))
		assert.NoError(t, err)
		assert.Equal(t, executors.NodePhaseFailed, n.NodePhase, "%+v", n.NodePhase)
	})

	t.Run("eventSendFailure", func(t *testing.T) {
		node := &mocks.ExecutableNode{}
		node.On("GetID").Return(nodeID)
		// In case Report event fails
		mockEventSink.SinkCb = func(ctx context.Context, message proto.Message) error {
			return expectedErr
		}
		n, err := exec.TransitionToPhase(ctx, execID, node, &v1alpha1.NodeStatus{Phase: v1alpha1.NodePhaseRunning}, handler.StatusSuccess)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, goerrors.Cause(err))
		assert.Equal(t, executors.NodeStatusUndefined, n)
	})

	t.Run("eventSendMismatch", func(t *testing.T) {
		node := &mocks.ExecutableNode{}
		node.On("GetID").Return(nodeID)
		// In case Report event fails
		mockEventSink.SinkCb = func(ctx context.Context, message proto.Message) error {
			return &eventsErr.EventError{Code: eventsErr.EventAlreadyInTerminalStateError,
				Cause: errors.New("already exists"),
			}
		}
		n, err := exec.TransitionToPhase(ctx, execID, node, &v1alpha1.NodeStatus{Phase: v1alpha1.NodePhaseRunning}, handler.StatusSuccess)
		assert.NoError(t, err)
		assert.Equal(t, executors.NodePhaseFailed, n.NodePhase)
	})

	// Testing that workflow execution name is queried in running
	t.Run("childWorkflows", func(t *testing.T) {
		now := v1.Now()

		wfNodeStatus := &mocks.ExecutableWorkflowNodeStatus{}
		wfNodeStatus.On("GetWorkflowExecutionName").Return("childWfName")

		status := &mocks.ExecutableNodeStatus{}
		status.On("GetPhase").Return(v1alpha1.NodePhaseQueued)
		status.On("GetAttempts").Return(uint32(0))
		status.On("GetDataDir").Return(storage.DataReference("x"))
		status.On("IncrementAttempts").Return(uint32(1))
		status.On("UpdatePhase", v1alpha1.NodePhaseRunning, mock.Anything, mock.AnythingOfType("string"))
		status.On("GetStartedAt").Return(&now)
		status.On("GetQueuedAt").Return(&now)
		status.On("GetStoppedAt").Return(&now)
		status.On("GetOrCreateWorkflowStatus").Return(wfNodeStatus)
		status.On("ClearTaskStatus").Return()
		status.On("ClearWorkflowStatus").Return()
		status.On("GetWorkflowNodeStatus").Return(wfNodeStatus)

		node := &mocks.ExecutableNode{}
		node.On("GetID").Return(nodeID)
		node.On("GetRetryStrategy").Return(nil)

		n, err := exec.TransitionToPhase(ctx, execID, node, status, handler.StatusRunning)
		assert.NoError(t, err)
		assert.Equal(t, executors.NodePhaseRunning, n.NodePhase)
		wfNodeStatus.AssertCalled(t, "GetWorkflowExecutionName")
	})
}

func TestNodeExecutor_Initialize(t *testing.T) {
	ctx := context.Background()
	enQWf := func(workflowID v1alpha1.WorkflowID) {
	}

	mockEventSink := events.NewMockEventSink().(*events.MockEventSink)
	memStore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)

	catalogClient, _ := catalog.NewCatalogClient(ctx, memStore)

	execIface, err := NewExecutor(ctx, memStore, enQWf, time.Second, mockEventSink, launchplan.NewFailFastLaunchPlanExecutor(), catalogClient, fakeKubeClient, promutils.NewTestScope())
	assert.NoError(t, err)
	exec := execIface.(*nodeExecutor)

	assert.NoError(t, exec.Initialize(ctx))
}

func TestNodeExecutor_RecursiveNodeHandler_RecurseStartNodes(t *testing.T) {
	ctx := context.Background()
	enQWf := func(workflowID v1alpha1.WorkflowID) {
	}
	mockEventSink := events.NewMockEventSink().(*events.MockEventSink)
	factory := createSingletonTaskExecutorFactory()
	task.SetTestFactory(factory)
	assert.True(t, task.IsTestModeEnabled())

	store := createInmemoryDataStore(t, promutils.NewTestScope())
	catalogClient, _ := catalog.NewCatalogClient(ctx, store)

	execIface, err := NewExecutor(ctx, store, enQWf, time.Second, mockEventSink,
		launchplan.NewFailFastLaunchPlanExecutor(), catalogClient, fakeKubeClient, promutils.NewTestScope())
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
		}, startNode, startNodeStatus

	}

	// Recurse Child Node Queued previously
	{
		tests := []struct {
			name              string
			currentNodePhase  v1alpha1.NodePhase
			expectedNodePhase v1alpha1.NodePhase
			expectedPhase     executors.NodePhase
			handlerReturn     func() (handler.Status, error)
			expectedError     bool
		}{
			// Starting at Queued
			{"nys->success", v1alpha1.NodePhaseNotYetStarted, v1alpha1.NodePhaseSucceeded, executors.NodePhaseSuccess, func() (handler.Status, error) {
				return handler.StatusSuccess, nil
			}, false},
			{"queued->success", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseSucceeded, executors.NodePhaseSuccess, func() (handler.Status, error) {
				return handler.StatusSuccess, nil
			}, false},
			{"nys->error", v1alpha1.NodePhaseNotYetStarted, v1alpha1.NodePhaseNotYetStarted, executors.NodePhaseUndefined, func() (handler.Status, error) {
				return handler.StatusUndefined, fmt.Errorf("err")
			}, true},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				hf := &mocks2.HandlerFactory{}
				exec.nodeHandlerFactory = hf

				h := &mocks3.IFace{}
				h.On("StartNode",
					mock.MatchedBy(func(ctx context.Context) bool { return true }),
					mock.MatchedBy(func(o v1alpha1.ExecutableWorkflow) bool { return true }),
					mock.MatchedBy(func(o v1alpha1.ExecutableNode) bool { return true }),
					mock.MatchedBy(func(o *handler.Data) bool { return true }),
				).Return(test.handlerReturn())

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

	factory := createSingletonTaskExecutorFactory()
	task.SetTestFactory(factory)
	assert.True(t, task.IsTestModeEnabled())

	store := createInmemoryDataStore(t, promutils.NewTestScope())
	catalogClient, _ := catalog.NewCatalogClient(ctx, store)

	execIface, err := NewExecutor(ctx, store, enQWf, time.Second, mockEventSink, launchplan.NewFailFastLaunchPlanExecutor(), catalogClient, fakeKubeClient, promutils.NewTestScope())
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
					DataDir: "data",
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
				h := &mocks3.IFace{}
				hf.On("GetHandler", v1alpha1.NodeKindEnd).Return(h, nil)
				h.On("StartNode", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(handler.StatusQueued, nil)

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
			}, n, ns

		}
		tests := []struct {
			name              string
			currentNodePhase  v1alpha1.NodePhase
			expectedNodePhase v1alpha1.NodePhase
			expectedPhase     executors.NodePhase
			handlerReturn     func() (handler.Status, error)
			expectedError     bool
		}{
			// Starting at Queued
			{"queued->success", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseSucceeded, executors.NodePhaseSuccess, func() (handler.Status, error) {
				return handler.StatusSuccess, nil
			}, false},

			{"queued->failed", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseFailed, executors.NodePhaseFailed, func() (handler.Status, error) {
				return handler.StatusFailed(fmt.Errorf("err")), nil
			}, false},

			{"queued->failing", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseFailing, executors.NodePhasePending, func() (handler.Status, error) {
				return handler.StatusFailing(fmt.Errorf("err")), nil
			}, false},

			{"queued->error", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseQueued, executors.NodePhaseUndefined, func() (handler.Status, error) {
				return handler.StatusUndefined, fmt.Errorf("err")
			}, true},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				hf := &mocks2.HandlerFactory{}
				exec.nodeHandlerFactory = hf

				h := &mocks3.IFace{}
				h.On("StartNode",
					mock.MatchedBy(func(ctx context.Context) bool { return true }),
					mock.MatchedBy(func(o v1alpha1.ExecutableWorkflow) bool { return true }),
					mock.MatchedBy(func(o v1alpha1.ExecutableNode) bool { return true }),
					mock.MatchedBy(func(o *handler.Data) bool { return true }),
				).Return(test.handlerReturn())

				hf.On("GetHandler", v1alpha1.NodeKindEnd).Return(h, nil)

				mockWf, _, mockNodeStatus := createSingleNodeWf(test.currentNodePhase, 0)
				startNode := mockWf.StartNode()
				startStatus := mockWf.GetNodeExecutionStatus(startNode.GetID())
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

	factory := createSingletonTaskExecutorFactory()
	task.SetTestFactory(factory)
	assert.True(t, task.IsTestModeEnabled())

	store := createInmemoryDataStore(t, promutils.NewTestScope())
	catalogClient, _ := catalog.NewCatalogClient(ctx, store)

	execIface, err := NewExecutor(ctx, store, enQWf, time.Second, mockEventSink, launchplan.NewFailFastLaunchPlanExecutor(), catalogClient, fakeKubeClient, promutils.NewTestScope())
	assert.NoError(t, err)
	exec := execIface.(*nodeExecutor)

	defaultNodeID := "n1"

	createSingleNodeWf := func(p v1alpha1.NodePhase, maxAttempts int) (v1alpha1.ExecutableWorkflow, v1alpha1.ExecutableNode, v1alpha1.ExecutableNodeStatus) {
		n := &v1alpha1.NodeSpec{
			ID:   defaultNodeID,
			Kind: v1alpha1.NodeKindTask,
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
		}, n, ns

	}

	// Recursion test with child Node not yet started
	{
		nodeN0 := "n0"
		nodeN2 := "n2"
		ctx := context.Background()
		connections := &v1alpha1.Connections{
			UpstreamEdges: map[v1alpha1.NodeID][]v1alpha1.NodeID{
				nodeN2: {nodeN0},
			},
		}

		setupNodePhase := func(n0Phase, n2Phase, expectedN2Phase v1alpha1.NodePhase) (*mocks.ExecutableWorkflow, *mocks.ExecutableNodeStatus) {
			// Setup
			mockN2Status := &mocks.ExecutableNodeStatus{}
			// No parent node
			mockN2Status.On("GetParentNodeID").Return(nil)
			mockN2Status.On("GetPhase").Return(n2Phase)
			mockN2Status.On("SetDataDir", mock.AnythingOfType(reflect.TypeOf(storage.DataReference("x")).String()))
			mockN2Status.On("GetDataDir").Return(storage.DataReference("blah"))
			mockN2Status.On("GetWorkflowNodeStatus").Return(nil)
			mockN2Status.On("GetStoppedAt").Return(nil)
			mockN2Status.On("UpdatePhase", expectedN2Phase, mock.Anything, mock.AnythingOfType("string"))
			mockN2Status.On("IsDirty").Return(false)

			mockNode := &mocks.ExecutableNode{}
			mockNode.On("GetID").Return(nodeN2)
			mockNode.On("GetBranchNode").Return(nil)
			mockNode.On("GetKind").Return(v1alpha1.NodeKindTask)
			mockNode.On("IsStartNode").Return(false)
			mockNode.On("IsEndNode").Return(false)

			mockNodeN0 := &mocks.ExecutableNode{}
			mockNodeN0.On("GetID").Return(nodeN0)
			mockNodeN0.On("GetBranchNode").Return(nil)
			mockNodeN0.On("GetKind").Return(v1alpha1.NodeKindTask)
			mockNodeN0.On("IsStartNode").Return(false)
			mockNodeN0.On("IsEndNode").Return(false)
			mockN0Status := &mocks.ExecutableNodeStatus{}
			mockN0Status.On("GetPhase").Return(n0Phase)
			mockN0Status.On("IsDirty").Return(false)

			mockWfStatus := &mocks.ExecutableWorkflowStatus{}
			mockWf := &mocks.ExecutableWorkflow{}
			mockWf.On("StartNode").Return(mockNodeN0)
			mockWf.On("GetNode", nodeN2).Return(mockNode, true)
			mockWf.On("GetNodeExecutionStatus", nodeN0).Return(mockN0Status)
			mockWf.On("GetNodeExecutionStatus", nodeN2).Return(mockN2Status)
			mockWf.On("GetConnections").Return(connections)
			mockWf.On("GetID").Return("w1")
			mockWf.On("FromNode", nodeN0).Return([]string{nodeN2}, nil)
			mockWf.On("FromNode", nodeN2).Return([]string{}, fmt.Errorf("did not expect"))
			mockWf.On("GetExecutionID").Return(v1alpha1.WorkflowExecutionIdentifier{})
			mockWf.On("GetExecutionStatus").Return(mockWfStatus)
			mockWfStatus.On("GetDataDir").Return(storage.DataReference("x"))
			return mockWf, mockN2Status
		}

		tests := []struct {
			name              string
			currentNodePhase  v1alpha1.NodePhase
			parentNodePhase   v1alpha1.NodePhase
			expectedNodePhase v1alpha1.NodePhase
			expectedPhase     executors.NodePhase
			handlerReturn     func() (handler.Status, error)
			expectedError     bool
			updateCalled      bool
		}{
			{"notYetStarted->notYetStarted", v1alpha1.NodePhaseNotYetStarted, v1alpha1.NodePhaseFailed, v1alpha1.NodePhaseNotYetStarted, executors.NodePhaseFailed, func() (handler.Status, error) {
				return handler.StatusNotStarted, nil
			}, false, false},

			{"notYetStarted->skipped", v1alpha1.NodePhaseNotYetStarted, v1alpha1.NodePhaseSkipped, v1alpha1.NodePhaseSkipped, executors.NodePhaseSuccess, func() (handler.Status, error) {
				return handler.StatusSkipped, nil
			}, false, true},

			{"notYetStarted->queued", v1alpha1.NodePhaseNotYetStarted, v1alpha1.NodePhaseSucceeded, v1alpha1.NodePhaseQueued, executors.NodePhasePending, func() (handler.Status, error) {
				return handler.StatusQueued, nil
			}, false, true},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				hf := &mocks2.HandlerFactory{}
				exec.nodeHandlerFactory = hf

				h := &mocks3.IFace{}
				hf.On("GetHandler", v1alpha1.NodeKindTask).Return(h, nil)

				mockWf, _ := setupNodePhase(test.parentNodePhase, test.currentNodePhase, test.expectedNodePhase)
				startNode := mockWf.StartNode()
				s, err := exec.RecursiveNodeHandler(ctx, mockWf, startNode)
				if test.expectedError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, test.expectedPhase, s.NodePhase, "expected: %s, received %s", test.expectedPhase.String(), s.NodePhase.String())
			})
		}
	}

	// Recurse Child Node Queued previously
	{
		tests := []struct {
			name              string
			currentNodePhase  v1alpha1.NodePhase
			expectedNodePhase v1alpha1.NodePhase
			expectedPhase     executors.NodePhase
			handlerReturn     func() (handler.Status, error)
			expectedError     bool
		}{
			// Starting at Queued
			{"queued->running", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseRunning, executors.NodePhasePending, func() (handler.Status, error) {
				return handler.StatusRunning, nil
			}, false},

			{"queued->queued", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseQueued, executors.NodePhasePending, func() (handler.Status, error) {
				return handler.StatusQueued, nil
			}, false},

			{"queued->failed", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseFailed, executors.NodePhaseFailed, func() (handler.Status, error) {
				return handler.StatusFailed(fmt.Errorf("err")), nil
			}, false},

			{"queued->failing", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseFailing, executors.NodePhasePending, func() (handler.Status, error) {
				return handler.StatusFailing(fmt.Errorf("err")), nil
			}, false},

			{"queued->success", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseSucceeded, executors.NodePhaseSuccess, func() (handler.Status, error) {
				return handler.StatusSuccess, nil
			}, false},

			{"queued->error", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseQueued, executors.NodePhaseUndefined, func() (handler.Status, error) {
				return handler.StatusUndefined, fmt.Errorf("err")
			}, true},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				hf := &mocks2.HandlerFactory{}
				exec.nodeHandlerFactory = hf

				h := &mocks3.IFace{}
				h.On("StartNode",
					mock.MatchedBy(func(ctx context.Context) bool { return true }),
					mock.MatchedBy(func(o v1alpha1.ExecutableWorkflow) bool { return true }),
					mock.MatchedBy(func(o v1alpha1.ExecutableNode) bool { return true }),
					mock.MatchedBy(func(o *handler.Data) bool { return true }),
				).Return(test.handlerReturn())

				hf.On("GetHandler", v1alpha1.NodeKindTask).Return(h, nil)

				mockWf, _, mockNodeStatus := createSingleNodeWf(test.currentNodePhase, 0)
				startNode := mockWf.StartNode()
				startStatus := mockWf.GetNodeExecutionStatus(startNode.GetID())
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

	// Recurse Child Node started previously
	{
		tests := []struct {
			name              string
			currentNodePhase  v1alpha1.NodePhase
			expectedNodePhase v1alpha1.NodePhase
			expectedPhase     executors.NodePhase
			handlerReturn     func() (handler.Status, error)
			expectedError     bool
		}{
			// Starting at running
			{"running->running", v1alpha1.NodePhaseRunning, v1alpha1.NodePhaseRunning, executors.NodePhasePending, func() (handler.Status, error) {
				return handler.StatusRunning, nil
			}, false},

			{"running->failing", v1alpha1.NodePhaseRunning, v1alpha1.NodePhaseFailing, executors.NodePhasePending, func() (handler.Status, error) {
				return handler.StatusFailing(fmt.Errorf("err")), nil
			}, false},

			{"running->failed", v1alpha1.NodePhaseRunning, v1alpha1.NodePhaseFailed, executors.NodePhaseFailed, func() (handler.Status, error) {
				return handler.StatusFailed(fmt.Errorf("err")), nil
			}, false},

			{"running->success", v1alpha1.NodePhaseRunning, v1alpha1.NodePhaseSucceeded, executors.NodePhaseSuccess, func() (handler.Status, error) {
				return handler.StatusSuccess, nil
			}, false},

			{"running->error", v1alpha1.NodePhaseRunning, v1alpha1.NodePhaseRunning, executors.NodePhaseUndefined, func() (handler.Status, error) {
				return handler.StatusUndefined, fmt.Errorf("err")
			}, true},

			{"previously-failed", v1alpha1.NodePhaseFailed, v1alpha1.NodePhaseFailed, executors.NodePhaseFailed, func() (handler.Status, error) {
				return handler.StatusQueued, nil
			}, false},

			{"previously-success", v1alpha1.NodePhaseSucceeded, v1alpha1.NodePhaseSucceeded, executors.NodePhaseComplete, func() (handler.Status, error) {
				return handler.StatusQueued, nil
			}, false},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				hf := &mocks2.HandlerFactory{}
				exec.nodeHandlerFactory = hf

				h := &mocks3.IFace{}
				h.On("CheckNodeStatus",
					mock.MatchedBy(func(ctx context.Context) bool { return true }),
					mock.MatchedBy(func(o v1alpha1.ExecutableWorkflow) bool { return true }),
					mock.MatchedBy(func(o v1alpha1.ExecutableNode) bool { return true }),
					mock.MatchedBy(func(o v1alpha1.ExecutableNodeStatus) bool { return true }),
				).Return(test.handlerReturn())

				hf.On("GetHandler", v1alpha1.NodeKindTask).Return(h, nil)

				mockWf, _, mockNodeStatus := createSingleNodeWf(test.currentNodePhase, 0)
				startNode := mockWf.StartNode()
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

func TestNodeExecutor_RecursiveNodeHandler_NoDownstream(t *testing.T) {
	ctx := context.Background()
	enQWf := func(workflowID v1alpha1.WorkflowID) {
	}
	mockEventSink := events.NewMockEventSink().(*events.MockEventSink)

	factory := createSingletonTaskExecutorFactory()
	task.SetTestFactory(factory)
	assert.True(t, task.IsTestModeEnabled())

	store := createInmemoryDataStore(t, promutils.NewTestScope())
	catalogClient, _ := catalog.NewCatalogClient(ctx, store)

	execIface, err := NewExecutor(ctx, store, enQWf, time.Second, mockEventSink, launchplan.NewFailFastLaunchPlanExecutor(), catalogClient, fakeKubeClient, promutils.NewTestScope())
	assert.NoError(t, err)
	exec := execIface.(*nodeExecutor)

	defaultNodeID := "n1"

	createSingleNodeWf := func(p v1alpha1.NodePhase, maxAttempts int) (v1alpha1.ExecutableWorkflow, v1alpha1.ExecutableNode, v1alpha1.ExecutableNodeStatus) {
		n := &v1alpha1.NodeSpec{
			ID:   defaultNodeID,
			Kind: v1alpha1.NodeKindTask,
			RetryStrategy: &v1alpha1.RetryStrategy{
				MinAttempts: &maxAttempts,
			},
		}
		ns := &v1alpha1.NodeStatus{
			Phase: p,
		}

		return &v1alpha1.FlyteWorkflow{
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
					defaultNodeID: n,
				},
				Connections: v1alpha1.Connections{
					UpstreamEdges: map[v1alpha1.NodeID][]v1alpha1.NodeID{
						defaultNodeID: {v1alpha1.StartNodeID},
					},
				},
			},
		}, n, ns

	}

	// Node not yet started
	{
		tests := []struct {
			name              string
			currentNodePhase  v1alpha1.NodePhase
			expectedNodePhase v1alpha1.NodePhase
			expectedPhase     executors.NodePhase
			handlerReturn     func() (handler.Status, error)
			expectedError     bool
		}{
			{"notYetStarted->running", v1alpha1.NodePhaseNotYetStarted, v1alpha1.NodePhaseQueued, executors.NodePhaseQueued, func() (handler.Status, error) {
				return handler.StatusRunning, nil
			}, false},

			{"notYetStarted->queued", v1alpha1.NodePhaseNotYetStarted, v1alpha1.NodePhaseQueued, executors.NodePhaseQueued, func() (handler.Status, error) {
				return handler.StatusQueued, nil
			}, false},

			{"notYetStarted->failed", v1alpha1.NodePhaseNotYetStarted, v1alpha1.NodePhaseQueued, executors.NodePhaseQueued, func() (handler.Status, error) {
				return handler.StatusFailed(fmt.Errorf("err")), nil
			}, false},

			{"notYetStarted->success", v1alpha1.NodePhaseNotYetStarted, v1alpha1.NodePhaseQueued, executors.NodePhaseQueued, func() (handler.Status, error) {
				return handler.StatusSuccess, nil
			}, false},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				hf := &mocks2.HandlerFactory{}
				exec.nodeHandlerFactory = hf

				h := &mocks3.IFace{}
				hf.On("GetHandler", v1alpha1.NodeKindTask).Return(h, nil)

				mockWf, mockNode, mockNodeStatus := createSingleNodeWf(test.currentNodePhase, 0)
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

	// Node queued previously
	{
		tests := []struct {
			name              string
			currentNodePhase  v1alpha1.NodePhase
			expectedNodePhase v1alpha1.NodePhase
			expectedPhase     executors.NodePhase
			handlerReturn     func() (handler.Status, error)
			expectedError     bool
		}{
			// Starting at Queued
			{"queued->running", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseRunning, executors.NodePhaseRunning, func() (handler.Status, error) {
				return handler.StatusRunning, nil
			}, false},

			{"queued->queued", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseQueued, executors.NodePhaseQueued, func() (handler.Status, error) {
				return handler.StatusQueued, nil
			}, false},

			{"queued->failed", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseFailed, executors.NodePhaseFailed, func() (handler.Status, error) {
				return handler.StatusFailed(fmt.Errorf("err")), nil
			}, false},

			{"queued->failing", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseFailing, executors.NodePhaseRunning, func() (handler.Status, error) {
				return handler.StatusFailing(fmt.Errorf("err")), nil
			}, false},

			{"queued->success", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseSucceeded, executors.NodePhaseSuccess, func() (handler.Status, error) {
				return handler.StatusSuccess, nil
			}, false},

			{"queued->error", v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseQueued, executors.NodePhaseUndefined, func() (handler.Status, error) {
				return handler.StatusUndefined, fmt.Errorf("err")
			}, true},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				hf := &mocks2.HandlerFactory{}
				exec.nodeHandlerFactory = hf

				h := &mocks3.IFace{}
				h.On("StartNode",
					mock.MatchedBy(func(ctx context.Context) bool { return true }),
					mock.MatchedBy(func(o v1alpha1.ExecutableWorkflow) bool { return true }),
					mock.MatchedBy(func(o v1alpha1.ExecutableNode) bool { return true }),
					mock.MatchedBy(func(o *handler.Data) bool { return true }),
				).Return(test.handlerReturn())

				hf.On("GetHandler", v1alpha1.NodeKindTask).Return(h, nil)

				mockWf, mockNode, mockNodeStatus := createSingleNodeWf(test.currentNodePhase, 0)
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

	// Node started previously
	{
		tests := []struct {
			name              string
			currentNodePhase  v1alpha1.NodePhase
			expectedNodePhase v1alpha1.NodePhase
			expectedPhase     executors.NodePhase
			handlerReturn     func() (handler.Status, error)
			expectedError     bool
		}{
			// Starting at running
			{"running->running", v1alpha1.NodePhaseRunning, v1alpha1.NodePhaseRunning, executors.NodePhaseRunning, func() (handler.Status, error) {
				return handler.StatusRunning, nil
			}, false},

			{"running->failing", v1alpha1.NodePhaseRunning, v1alpha1.NodePhaseFailing, executors.NodePhaseRunning, func() (handler.Status, error) {
				return handler.StatusFailing(fmt.Errorf("err")), nil
			}, false},

			{"running->failed", v1alpha1.NodePhaseRunning, v1alpha1.NodePhaseFailed, executors.NodePhaseFailed, func() (handler.Status, error) {
				return handler.StatusFailed(fmt.Errorf("err")), nil
			}, false},

			{"running->success", v1alpha1.NodePhaseRunning, v1alpha1.NodePhaseSucceeded, executors.NodePhaseSuccess, func() (handler.Status, error) {
				return handler.StatusSuccess, nil
			}, false},

			{"running->error", v1alpha1.NodePhaseRunning, v1alpha1.NodePhaseRunning, executors.NodePhaseUndefined, func() (handler.Status, error) {
				return handler.StatusUndefined, fmt.Errorf("err")
			}, true},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				hf := &mocks2.HandlerFactory{}
				exec.nodeHandlerFactory = hf

				h := &mocks3.IFace{}
				h.On("CheckNodeStatus",
					mock.MatchedBy(func(ctx context.Context) bool { return true }),
					mock.MatchedBy(func(o v1alpha1.ExecutableWorkflow) bool { return true }),
					mock.MatchedBy(func(o v1alpha1.ExecutableNode) bool { return true }),
					mock.MatchedBy(func(o v1alpha1.ExecutableNodeStatus) bool { return true }),
				).Return(test.handlerReturn())

				hf.On("GetHandler", v1alpha1.NodeKindTask).Return(h, nil)

				mockWf, mockNode, mockNodeStatus := createSingleNodeWf(test.currentNodePhase, 0)
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

	// Node started previously and is failing
	{
		tests := []struct {
			name              string
			currentNodePhase  v1alpha1.NodePhase
			expectedNodePhase v1alpha1.NodePhase
			expectedPhase     executors.NodePhase
			handlerReturn     func() (handler.Status, error)
			expectedError     bool
		}{
			// Starting at Failing
			// TODO this should be illegal
			{"failing->running", v1alpha1.NodePhaseFailing, v1alpha1.NodePhaseRunning, executors.NodePhaseRunning, func() (handler.Status, error) {
				return handler.StatusRunning, nil
			}, false},

			{"failing->failed", v1alpha1.NodePhaseFailing, v1alpha1.NodePhaseFailed, executors.NodePhaseFailed, func() (handler.Status, error) {
				return handler.StatusFailed(fmt.Errorf("err")), nil
			}, false},

			// TODO this should be illegal
			{"failing->success", v1alpha1.NodePhaseFailing, v1alpha1.NodePhaseSucceeded, executors.NodePhaseSuccess, func() (handler.Status, error) {
				return handler.StatusSuccess, nil
			}, false},

			{"failing->error", v1alpha1.NodePhaseFailing, v1alpha1.NodePhaseFailing, executors.NodePhaseUndefined, func() (handler.Status, error) {
				return handler.StatusUndefined, fmt.Errorf("err")
			}, true},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				hf := &mocks2.HandlerFactory{}
				exec.nodeHandlerFactory = hf

				h := &mocks3.IFace{}
				h.On("HandleFailingNode",
					mock.MatchedBy(func(ctx context.Context) bool { return true }),
					mock.MatchedBy(func(o v1alpha1.ExecutableWorkflow) bool { return true }),
					mock.MatchedBy(func(o v1alpha1.ExecutableNode) bool { return true }),
				).Return(test.handlerReturn())

				hf.On("GetHandler", v1alpha1.NodeKindTask).Return(h, nil)

				mockWf, mockNode, mockNodeStatus := createSingleNodeWf(test.currentNodePhase, 0)
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

	// Node started previously and retryable failure
	{
		tests := []struct {
			name              string
			currentNodePhase  v1alpha1.NodePhase
			expectedNodePhase v1alpha1.NodePhase
			expectedPhase     executors.NodePhase
			handlerReturn     func() (handler.Status, error)
			expectedError     bool
		}{
			// Starting at Queued
			{"running->retryable", v1alpha1.NodePhaseRunning, v1alpha1.NodePhaseRetryableFailure, executors.NodePhaseRunning, func() (handler.Status, error) {
				return handler.StatusRetryableFailure(fmt.Errorf("err")), nil
			}, false},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				hf := &mocks2.HandlerFactory{}
				exec.nodeHandlerFactory = hf

				h := &mocks3.IFace{}
				h.On("CheckNodeStatus",
					mock.MatchedBy(func(ctx context.Context) bool { return true }),
					mock.MatchedBy(func(o v1alpha1.ExecutableWorkflow) bool { return true }),
					mock.MatchedBy(func(o v1alpha1.ExecutableNode) bool { return true }),
					mock.MatchedBy(func(o v1alpha1.ExecutableNodeStatus) bool { return true }),
				).Return(test.handlerReturn())

				hf.On("GetHandler", v1alpha1.NodeKindTask).Return(h, nil)

				mockWf, mockNode, mockNodeStatus := createSingleNodeWf(test.currentNodePhase, 2)
				s, err := exec.RecursiveNodeHandler(ctx, mockWf, mockNode)
				if test.expectedError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, test.expectedPhase, s.NodePhase, "expected: %s, received %s", test.expectedPhase.String(), s.NodePhase.String())
				assert.Equal(t, uint32(1), mockNodeStatus.GetAttempts())
				assert.Equal(t, test.expectedNodePhase, mockNodeStatus.GetPhase(), "expected %s, received %s", test.expectedNodePhase.String(), mockNodeStatus.GetPhase().String())
			})
		}
	}

	// Node started previously and retryable failure - but exhausted attempts
	{
		tests := []struct {
			name              string
			currentNodePhase  v1alpha1.NodePhase
			expectedNodePhase v1alpha1.NodePhase
			expectedPhase     executors.NodePhase
			handlerReturn     func() (handler.Status, error)
			expectedError     bool
		}{
			{"running->retryable", v1alpha1.NodePhaseRunning, v1alpha1.NodePhaseFailed, executors.NodePhaseFailed, func() (handler.Status, error) {
				return handler.StatusRetryableFailure(fmt.Errorf("err")), nil
			}, false},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				hf := &mocks2.HandlerFactory{}
				exec.nodeHandlerFactory = hf

				h := &mocks3.IFace{}
				h.On("CheckNodeStatus",
					mock.MatchedBy(func(ctx context.Context) bool { return true }),
					mock.MatchedBy(func(o v1alpha1.ExecutableWorkflow) bool { return true }),
					mock.MatchedBy(func(o v1alpha1.ExecutableNode) bool { return true }),
					mock.MatchedBy(func(o v1alpha1.ExecutableNodeStatus) bool { return true }),
				).Return(test.handlerReturn())

				hf.On("GetHandler", v1alpha1.NodeKindTask).Return(h, nil)

				mockWf, mockNode, mockNodeStatus := createSingleNodeWf(test.currentNodePhase, 1)
				s, err := exec.RecursiveNodeHandler(ctx, mockWf, mockNode)
				if test.expectedError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, test.expectedPhase, s.NodePhase, "expected: %s, received %s", test.expectedPhase.String(), s.NodePhase.String())
				assert.Equal(t, uint32(1), mockNodeStatus.GetAttempts())
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

	factory := createSingletonTaskExecutorFactory()
	task.SetTestFactory(factory)
	assert.True(t, task.IsTestModeEnabled())

	store := createInmemoryDataStore(t, promutils.NewTestScope())
	catalogClient, _ := catalog.NewCatalogClient(ctx, store)

	execIface, err := NewExecutor(ctx, store, enQWf, time.Second, mockEventSink, launchplan.NewFailFastLaunchPlanExecutor(), catalogClient, fakeKubeClient, promutils.NewTestScope())
	assert.NoError(t, err)
	exec := execIface.(*nodeExecutor)

	defaultNodeID := "n1"

	createSingleNodeWf := func(parentPhase v1alpha1.NodePhase, maxAttempts int) (v1alpha1.ExecutableWorkflow, v1alpha1.ExecutableNode, v1alpha1.ExecutableNodeStatus) {
		n := &v1alpha1.NodeSpec{
			ID:   defaultNodeID,
			Kind: v1alpha1.NodeKindTask,
			RetryStrategy: &v1alpha1.RetryStrategy{
				MinAttempts: &maxAttempts,
			},
		}
		ns := &v1alpha1.NodeStatus{}

		return &v1alpha1.FlyteWorkflow{
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
			{"skipped", v1alpha1.NodePhaseSkipped, v1alpha1.NodePhaseSkipped, executors.NodePhaseSuccess, false},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				hf := &mocks2.HandlerFactory{}
				exec.nodeHandlerFactory = hf
				h := &mocks3.IFace{}
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
