package branch

import (
	"context"
	"fmt"
	"testing"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	mocks3 "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/promutils/labeled"
	"github.com/lyft/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	mocks2 "github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	"github.com/lyft/flytepropeller/pkg/controller/executors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/errors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler/mocks"
)

type recursiveNodeHandlerFn func(ctx context.Context, w v1alpha1.ExecutableWorkflow, currentNode v1alpha1.ExecutableNode) (executors.NodeStatus, error)
type abortNodeHandlerCbFn func(ctx context.Context, w v1alpha1.ExecutableWorkflow, currentNode v1alpha1.ExecutableNode) error

type mockNodeExecutor struct {
	executors.Node
	RecursiveNodeHandlerCB recursiveNodeHandlerFn
	AbortNodeHandlerCB     abortNodeHandlerCbFn
}

type branchNodeStateHolder struct {
	s handler.BranchNodeState
}

func (t *branchNodeStateHolder) PutTaskNodeState(s handler.TaskNodeState) error {
	panic("not implemented")
}

func (t *branchNodeStateHolder) PutBranchNode(s handler.BranchNodeState) error {
	t.s = s
	return nil
}

func (t branchNodeStateHolder) PutWorkflowNodeState(s handler.WorkflowNodeState) error {
	panic("not implemented")
}

func (t branchNodeStateHolder) PutDynamicNodeState(s handler.DynamicNodeState) error {
	panic("not implemented")
}

func (m *mockNodeExecutor) RecursiveNodeHandler(ctx context.Context, w v1alpha1.ExecutableWorkflow, currentNode v1alpha1.ExecutableNode) (executors.NodeStatus, error) {
	return m.RecursiveNodeHandlerCB(ctx, w, currentNode)
}

func (m *mockNodeExecutor) AbortHandler(ctx context.Context, w v1alpha1.ExecutableWorkflow, currentNode v1alpha1.ExecutableNode, reason string) error {
	return m.AbortNodeHandlerCB(ctx, w, currentNode)
}

func createNodeContext(phase v1alpha1.BranchNodePhase, childNodeID *v1alpha1.NodeID, w v1alpha1.ExecutableWorkflow, n v1alpha1.ExecutableNode, inputs *core.LiteralMap) *mocks.NodeExecutionContext {
	nodeID := "nodeID"
	branchNodeState := handler.BranchNodeState{
		FinalizedNodeID: &nodeID,
		Phase:           phase,
	}
	s := &branchNodeStateHolder{s: branchNodeState}

	wfExecID := &core.WorkflowExecutionIdentifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
	}

	nm := &mocks.NodeExecutionMetadata{}
	nm.On("GetAnnotations").Return(map[string]string{})
	nm.On("GetExecutionID").Return(v1alpha1.WorkflowExecutionIdentifier{
		WorkflowExecutionIdentifier: wfExecID,
	})
	nm.On("GetK8sServiceAccount").Return("service-account")
	nm.On("GetLabels").Return(map[string]string{})
	nm.On("GetNamespace").Return("namespace")
	nm.On("GetOwnerID").Return(types.NamespacedName{Namespace: "namespace", Name: "name"})
	nm.On("GetOwnerReference").Return(v1.OwnerReference{
		Kind: "sample",
		Name: "name",
	})

	ns := &mocks2.ExecutableNodeStatus{}
	ns.On("GetDataDir").Return(storage.DataReference("data-dir"))
	ns.On("GetPhase").Return(v1alpha1.NodePhaseNotYetStarted)

	ir := &mocks3.InputReader{}
	ir.On("Get", mock.Anything).Return(inputs, nil)

	nCtx := &mocks.NodeExecutionContext{}
	nCtx.On("NodeExecutionMetadata").Return(nm)
	nCtx.On("Node").Return(n)
	nCtx.On("InputReader").Return(ir)
	nCtx.On("DataStore").Return(storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope()))
	nCtx.On("CurrentAttempt").Return(uint32(1))
	nCtx.On("MaxDatasetSizeBytes").Return(int64(1))
	nCtx.On("NodeStatus").Return(ns)

	nCtx.On("NodeID").Return("n1")
	nCtx.On("EnqueueOwner").Return(nil)
	nCtx.On("Workflow").Return(w)

	nr := &mocks.NodeStateReader{}
	nr.On("GetBranchNode").Return(handler.BranchNodeState{
		FinalizedNodeID: childNodeID,
		Phase:           phase,
	})
	nCtx.On("NodeStateReader").Return(nr)
	nCtx.On("NodeStateWriter").Return(s)
	return nCtx
}

func TestBranchHandler_RecurseDownstream(t *testing.T) {
	ctx := context.TODO()
	m := &mockNodeExecutor{}

	branch := New(m, promutils.NewTestScope()).(*branchHandler)
	childNodeID := "child"
	childDatadir := v1alpha1.DataReference("test")
	w := &v1alpha1.FlyteWorkflow{
		Status: v1alpha1.WorkflowStatus{
			NodeStatus: map[v1alpha1.NodeID]*v1alpha1.NodeStatus{
				childNodeID: {
					DataDir: childDatadir,
				},
			},
		},
	}

	res := &v12.ResourceRequirements{}
	n := &mocks2.ExecutableNode{}
	n.On("GetResources").Return(res)

	expectedError := fmt.Errorf("error")
	recursiveNodeHandlerFnArchetype := func(status executors.NodeStatus, err error) recursiveNodeHandlerFn {
		return func(ctx context.Context, w v1alpha1.ExecutableWorkflow, currentNode v1alpha1.ExecutableNode) (executors.NodeStatus, error) {
			return status, err
		}
	}

	tests := []struct {
		name                   string
		recursiveNodeHandlerFn recursiveNodeHandlerFn
		nodeStatus             v1alpha1.ExecutableNodeStatus
		branchTakenNode        v1alpha1.ExecutableNode
		isErr                  bool
		expectedPhase          handler.EPhase
		childPhase             v1alpha1.NodePhase
	}{
		{"childNodeError", recursiveNodeHandlerFnArchetype(executors.NodeStatusUndefined, expectedError),
			nil, &v1alpha1.NodeSpec{}, true, handler.EPhaseUndefined, v1alpha1.NodePhaseFailed},
		{"childPending", recursiveNodeHandlerFnArchetype(executors.NodeStatusPending, nil),
			nil, &v1alpha1.NodeSpec{}, false, handler.EPhaseRunning, v1alpha1.NodePhaseQueued},
		{"childStillRunning", recursiveNodeHandlerFnArchetype(executors.NodeStatusRunning, nil),
			nil, &v1alpha1.NodeSpec{}, false, handler.EPhaseRunning, v1alpha1.NodePhaseRunning},
		{"childFailure", recursiveNodeHandlerFnArchetype(executors.NodeStatusFailed(expectedError), nil),
			nil, &v1alpha1.NodeSpec{}, false, handler.EPhaseFailed, v1alpha1.NodePhaseFailed},
		{"childComplete", recursiveNodeHandlerFnArchetype(executors.NodeStatusComplete, nil),
			&v1alpha1.NodeStatus{}, &v1alpha1.NodeSpec{ID: childNodeID}, false, handler.EPhaseSuccess, v1alpha1.NodePhaseSucceeded},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m.RecursiveNodeHandlerCB = test.recursiveNodeHandlerFn

			nCtx := createNodeContext(v1alpha1.BranchNodeNotYetEvaluated, &childNodeID, w, n, nil)
			h, err := branch.recurseDownstream(ctx, nCtx, test.nodeStatus, test.branchTakenNode)
			if test.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expectedPhase, h.Info().GetPhase())
			if test.nodeStatus != nil {
				assert.Equal(t, w.GetNodeExecutionStatus(ctx, test.branchTakenNode.GetID()).GetDataDir(), test.nodeStatus.GetDataDir())
			}
		})
	}
}

func TestBranchHandler_AbortNode(t *testing.T) {
	ctx := context.TODO()
	m := &mockNodeExecutor{}
	branch := New(m, promutils.NewTestScope())
	b1 := "b1"
	n1 := "n1"
	n2 := "n2"

	exp, _ := getComparisonExpression(1.0, core.ComparisonExpression_EQ, 1.0)
	branchNode := &v1alpha1.BranchNodeSpec{

		If: v1alpha1.IfBlock{
			Condition: v1alpha1.BooleanExpression{
				BooleanExpression: &core.BooleanExpression{
					Expr: &core.BooleanExpression_Comparison{
						Comparison: exp,
					},
				},
			},
			ThenNode: &n1,
		},
		ElseIf: []*v1alpha1.IfBlock{
			{
				Condition: v1alpha1.BooleanExpression{
					BooleanExpression: &core.BooleanExpression{
						Expr: &core.BooleanExpression_Comparison{
							Comparison: exp,
						},
					},
				},
				ThenNode: &n2,
			},
		},
	}

	n := &v1alpha1.NodeSpec{
		ID:         n2,
		BranchNode: branchNode,
	}

	w := &v1alpha1.FlyteWorkflow{
		WorkflowSpec: &v1alpha1.WorkflowSpec{
			ID: "test",
			Nodes: map[v1alpha1.NodeID]*v1alpha1.NodeSpec{
				n1: {
					ID: n1,
				},
				n2: n,
			},
		},
		Status: v1alpha1.WorkflowStatus{
			NodeStatus: map[v1alpha1.NodeID]*v1alpha1.NodeStatus{
				b1: {
					Phase: v1alpha1.NodePhaseRunning,
					BranchStatus: &v1alpha1.BranchNodeStatus{
						FinalizedNodeID: &n1,
					},
				},
			},
		},
	}

	t.Run("NoBranchNode", func(t *testing.T) {
		nCtx := createNodeContext(v1alpha1.BranchNodeError, nil, w, n, nil)
		err := branch.Abort(ctx, nCtx, "")
		assert.Error(t, err)
		assert.True(t, errors.Matches(err, errors.UserProvidedError))
	})

	t.Run("BranchNodeSuccess", func(t *testing.T) {

		nCtx := createNodeContext(v1alpha1.BranchNodeSuccess, &n1, w, n, nil)
		m.AbortNodeHandlerCB = func(ctx context.Context, w v1alpha1.ExecutableWorkflow, currentNode v1alpha1.ExecutableNode) error {
			assert.Equal(t, n1, currentNode.GetID())
			return nil
		}
		err := branch.Abort(ctx, nCtx, "")
		assert.NoError(t, err)
	})
}

func TestBranchHandler_Initialize(t *testing.T) {
	ctx := context.TODO()
	m := &mockNodeExecutor{}
	branch := New(m, promutils.NewTestScope())
	assert.NoError(t, branch.Setup(ctx, nil))
}

// TODO incomplete test suite, add more
func TestBranchHandler_HandleNode(t *testing.T) {
	ctx := context.TODO()
	m := &mockNodeExecutor{}
	branch := New(m, promutils.NewTestScope())
	childNodeID := "child"
	childDatadir := v1alpha1.DataReference("test")
	w := &v1alpha1.FlyteWorkflow{
		WorkflowSpec: &v1alpha1.WorkflowSpec{
			ID: "test",
		},
		Status: v1alpha1.WorkflowStatus{
			NodeStatus: map[v1alpha1.NodeID]*v1alpha1.NodeStatus{
				childNodeID: {
					DataDir: childDatadir,
				},
			},
		},
	}
	_, inputs := getComparisonExpression(1, core.ComparisonExpression_NEQ, 1)

	tests := []struct {
		name          string
		node          v1alpha1.ExecutableNode
		isErr         bool
		expectedPhase handler.EPhase
	}{
		{"NoBranchNode", &v1alpha1.NodeSpec{}, false, handler.EPhaseFailed},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := &v12.ResourceRequirements{}
			n := &mocks2.ExecutableNode{}
			n.On("GetResources").Return(res)
			n.On("GetBranchNode").Return(nil)
			nCtx := createNodeContext(v1alpha1.BranchNodeSuccess, &childNodeID, w, n, inputs)

			s, err := branch.Handle(ctx, nCtx)
			if test.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expectedPhase, s.Info().GetPhase())
		})
	}
}

func init() {
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)
}
