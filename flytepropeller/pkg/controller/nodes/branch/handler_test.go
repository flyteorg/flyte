package branch

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	mocks3 "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	mocks2 "github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors"
	execMocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/common"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces/mocks"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

var eventConfig = &config.EventConfig{
	RawOutputPolicy: config.RawOutputPolicyReference,
}

type branchNodeStateHolder struct {
	s handler.BranchNodeState
}

func (t *branchNodeStateHolder) ClearNodeStatus() {
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

func (t branchNodeStateHolder) PutGateNodeState(s handler.GateNodeState) error {
	panic("not implemented")
}

func (t branchNodeStateHolder) PutArrayNodeState(s handler.ArrayNodeState) error {
	panic("not implemented")
}

type parentInfo struct {
}

func (parentInfo) GetUniqueID() v1alpha1.NodeID {
	return "u1"
}

func (parentInfo) CurrentAttempt() uint32 {
	return uint32(2)
}

func (parentInfo) IsInDynamicChain() bool {
	return false
}

func createNodeContext(phase v1alpha1.BranchNodePhase, childNodeID *v1alpha1.NodeID, n v1alpha1.ExecutableNode,
	inputs *core.LiteralMap, nl *execMocks.NodeLookup, eCtx executors.ExecutionContext) (*mocks.NodeExecutionContext, *branchNodeStateHolder) {
	branchNodeState := handler.BranchNodeState{
		FinalizedNodeID: childNodeID,
		Phase:           phase,
	}
	s := &branchNodeStateHolder{s: branchNodeState}

	wfExecID := &core.WorkflowExecutionIdentifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
	}

	nm := &mocks.NodeExecutionMetadata{}
	nm.EXPECT().GetAnnotations().Return(map[string]string{})
	nm.EXPECT().GetNodeExecutionID().Return(&core.NodeExecutionIdentifier{
		ExecutionId: wfExecID,
		NodeId:      n.GetID(),
	})
	nm.EXPECT().GetK8sServiceAccount().Return("service-account")
	nm.EXPECT().GetLabels().Return(map[string]string{})
	nm.EXPECT().GetNamespace().Return("namespace")
	nm.EXPECT().GetOwnerID().Return(types.NamespacedName{Namespace: "namespace", Name: "name"})
	nm.EXPECT().GetOwnerReference().Return(v1.OwnerReference{
		Kind: "sample",
		Name: "name",
	})

	ns := &mocks2.ExecutableNodeStatus{}
	ns.EXPECT().GetDataDir().Return(storage.DataReference("data-dir"))
	ns.EXPECT().GetOutputDir().Return(storage.DataReference("output-dir"))
	ns.EXPECT().GetPhase().Return(v1alpha1.NodePhaseNotYetStarted)

	ir := &mocks3.InputReader{}
	ir.EXPECT().Get(mock.Anything).Return(inputs, nil)

	nCtx := &mocks.NodeExecutionContext{}
	nCtx.EXPECT().NodeExecutionMetadata().Return(nm)
	nCtx.EXPECT().Node().Return(n)
	nCtx.EXPECT().InputReader().Return(ir)
	tmpDataStore, _ := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	nCtx.EXPECT().DataStore().Return(tmpDataStore)
	nCtx.EXPECT().CurrentAttempt().Return(uint32(1))
	nCtx.EXPECT().NodeStatus().Return(ns)

	nCtx.EXPECT().NodeID().Return("n1")
	nCtx.EXPECT().EnqueueOwnerFunc().Return(nil)

	nr := &mocks.NodeStateReader{}
	nr.EXPECT().GetBranchNodeState().Return(handler.BranchNodeState{
		FinalizedNodeID: childNodeID,
		Phase:           phase,
	})
	nCtx.EXPECT().NodeStateReader().Return(nr)
	nCtx.EXPECT().NodeStateWriter().Return(s)
	nCtx.EXPECT().ExecutionContext().Return(eCtx)
	if nl == nil {
		nCtx.EXPECT().ContextualNodeLookup().Return(nil)
	} else {
		nCtx.EXPECT().ContextualNodeLookup().Return(nl)
	}
	return nCtx, s
}

func TestBranchHandler_RecurseDownstream(t *testing.T) {
	ctx := context.TODO()

	childNodeID := "child"
	nodeID := "n1"

	res := &v12.ResourceRequirements{}
	n := &mocks2.ExecutableNode{}
	n.EXPECT().GetResources().Return(res)
	n.EXPECT().GetID().Return(nodeID)

	expectedError := &core.ExecutionError{}
	bn := &mocks2.ExecutableNode{}
	bn.EXPECT().GetID().Return(childNodeID)

	tests := []struct {
		name            string
		ns              interfaces.NodeStatus
		err             error
		branchTakenNode v1alpha1.ExecutableNode
		isErr           bool
		expectedPhase   handler.EPhase
		childPhase      v1alpha1.NodePhase
		hasOutputs      bool
		upstreamNodeID  string
	}{
		{"upstreamNodeExists", interfaces.NodeStatusPending, nil,
			bn, false, handler.EPhaseRunning, v1alpha1.NodePhaseQueued, true, "n2"},
		{"childNodeError", interfaces.NodeStatusUndefined, fmt.Errorf("err"),
			bn, true, handler.EPhaseUndefined, v1alpha1.NodePhaseFailed, true, ""},
		{"childPending", interfaces.NodeStatusPending, nil,
			bn, false, handler.EPhaseRunning, v1alpha1.NodePhaseQueued, true, ""},
		{"childStillRunning", interfaces.NodeStatusRunning, nil,
			bn, false, handler.EPhaseRunning, v1alpha1.NodePhaseRunning, true, ""},
		{"childFailure", interfaces.NodeStatusFailed(expectedError), nil,
			bn, false, handler.EPhaseFailed, v1alpha1.NodePhaseFailed, true, ""},
		{"childTimedOut", interfaces.NodeStatusTimedOut, nil,
			bn, false, handler.EPhaseFailed, v1alpha1.NodePhaseTimedOut, true, ""},
		{"childComplete", interfaces.NodeStatusComplete, nil,
			bn, false, handler.EPhaseSuccess, v1alpha1.NodePhaseSucceeded, true, ""},
		{"childCompleteNoOutputs", interfaces.NodeStatusComplete, nil,
			bn, false, handler.EPhaseSuccess, v1alpha1.NodePhaseSucceeded, false, ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			eCtx := &execMocks.ExecutionContext{}
			eCtx.EXPECT().GetParentInfo().Return(parentInfo{})

			mockNodeLookup := &execMocks.NodeLookup{}
			if len(test.upstreamNodeID) > 0 {
				mockNodeLookup.EXPECT().ToNode(childNodeID).Return([]string{test.upstreamNodeID}, nil)
			} else {
				mockNodeLookup.EXPECT().ToNode(childNodeID).Return(nil, nil)
			}

			nCtx, _ := createNodeContext(v1alpha1.BranchNodeNotYetEvaluated, &childNodeID, n, nil, mockNodeLookup, eCtx)
			newParentInfo, _ := common.CreateParentInfo(parentInfo{}, nCtx.NodeID(), nCtx.CurrentAttempt(), false)
			expectedExecContext := executors.NewExecutionContextWithParentInfo(nCtx.ExecutionContext(), newParentInfo)
			mockNodeExecutor := &mocks.Node{}
			mockNodeExecutor.EXPECT().RecursiveNodeHandler(
				mock.Anything, // ctx
				mock.MatchedBy(func(e executors.ExecutionContext) bool { return assert.Equal(t, e, expectedExecContext) }),
				mock.MatchedBy(func(d executors.DAGStructure) bool {
					if assert.NotNil(t, d) {
						fList, err1 := d.FromNode("x")
						dList, err2 := d.ToNode(childNodeID)
						b := assert.NoError(t, err1)
						b = b && assert.Equal(t, []v1alpha1.NodeID{}, fList)
						b = b && assert.NoError(t, err2)
						dListExpected := []v1alpha1.NodeID{nodeID}
						if len(test.upstreamNodeID) > 0 {
							dListExpected = append([]string{test.upstreamNodeID}, dListExpected...)
						}
						b = b && assert.Equal(t, dListExpected, dList)
						return b
					}
					return false
				}),
				mock.MatchedBy(func(lookup executors.NodeLookup) bool { return assert.Equal(t, lookup, mockNodeLookup) }),
				mock.MatchedBy(func(n v1alpha1.ExecutableNode) bool { return assert.Equal(t, n.GetID(), childNodeID) }),
			).Return(test.ns, test.err)

			childNodeStatus := &mocks2.ExecutableNodeStatus{}
			childNodeStatus.EXPECT().GetAttempts().Return(0)
			childNodeStatus.On("SetDataDir", storage.DataReference("/output-dir/child")).Once()
			childNodeStatus.On("SetOutputDir", storage.DataReference("/output-dir/child/0")).Once()
			mockNodeLookup.EXPECT().GetNodeExecutionStatus(ctx, childNodeID).Return(childNodeStatus)
			if test.childPhase == v1alpha1.NodePhaseSucceeded && test.hasOutputs {
				_ = nCtx.DataStore().WriteRaw(ctx, storage.DataReference("/output-dir/child/0/outputs.pb"), 0, storage.Options{}, bytes.NewReader([]byte{}))
			}

			branch := New(mockNodeExecutor, eventConfig, promutils.NewTestScope()).(*branchHandler)
			h, err := branch.recurseDownstream(ctx, nCtx, test.branchTakenNode)
			if test.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expectedPhase, h.Info().GetPhase())
		})
	}
}

func TestBranchHandler_AbortNode(t *testing.T) {
	ctx := context.TODO()
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
	assert.NotNil(t, w)

	t.Run("NoBranchNode", func(t *testing.T) {
		mockNodeExecutor := &mocks.Node{}
		mockNodeExecutor.EXPECT().AbortHandler(mock.Anything,
			mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("err"))
		eCtx := &execMocks.ExecutionContext{}
		eCtx.EXPECT().GetParentInfo().Return(nil)
		nCtx, _ := createNodeContext(v1alpha1.BranchNodeError, nil, n, nil, nil, eCtx)
		branch := New(mockNodeExecutor, eventConfig, promutils.NewTestScope())
		err := branch.Abort(ctx, nCtx, "")
		assert.NoError(t, err)
	})

	t.Run("BranchNodeSuccess", func(t *testing.T) {
		mockNodeExecutor := &mocks.Node{}
		mockNodeLookup := &execMocks.NodeLookup{}
		mockNodeLookup.EXPECT().ToNode(mock.Anything).Return(nil, nil)
		eCtx := &execMocks.ExecutionContext{}
		eCtx.EXPECT().GetParentInfo().Return(parentInfo{})
		nCtx, s := createNodeContext(v1alpha1.BranchNodeSuccess, &n1, n, nil, mockNodeLookup, eCtx)
		newParentInfo, _ := common.CreateParentInfo(parentInfo{}, nCtx.NodeID(), nCtx.CurrentAttempt(), false)
		expectedExecContext := executors.NewExecutionContextWithParentInfo(nCtx.ExecutionContext(), newParentInfo)
		mockNodeExecutor.EXPECT().AbortHandler(mock.Anything,
			mock.MatchedBy(func(e executors.ExecutionContext) bool { return assert.Equal(t, e, expectedExecContext) }),
			mock.Anything,
			mock.Anything, mock.Anything, mock.Anything).Return(nil)
		mockNodeLookup.EXPECT().GetNode(*s.s.FinalizedNodeID).Return(n, true)
		branch := New(mockNodeExecutor, eventConfig, promutils.NewTestScope())
		err := branch.Abort(ctx, nCtx, "")
		assert.NoError(t, err)
	})
}

func TestBranchHandler_Initialize(t *testing.T) {
	ctx := context.TODO()
	mockNodeExecutor := &mocks.Node{}
	branch := New(mockNodeExecutor, eventConfig, promutils.NewTestScope())
	assert.NoError(t, branch.Setup(ctx, nil))
}

// TODO incomplete test suite, add more
func TestBranchHandler_HandleNode(t *testing.T) {
	ctx := context.TODO()
	mockNodeExecutor := &mocks.Node{}
	branch := New(mockNodeExecutor, eventConfig, promutils.NewTestScope())
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
	assert.NotNil(t, w)

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
			n.EXPECT().GetResources().Return(res)
			n.EXPECT().GetBranchNode().Return(nil)
			n.EXPECT().GetID().Return("n1")
			eCtx := &execMocks.ExecutionContext{}
			eCtx.EXPECT().GetParentInfo().Return(nil)
			nCtx, _ := createNodeContext(v1alpha1.BranchNodeSuccess, &childNodeID, n, inputs, nil, eCtx)

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
