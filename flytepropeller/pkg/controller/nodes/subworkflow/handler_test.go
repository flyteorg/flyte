package subworkflow

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	mocks4 "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	mocks2 "github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
	mocks3 "github.com/lyft/flytepropeller/pkg/controller/nodes/handler/mocks"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/subworkflow/launchplan"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/subworkflow/launchplan/mocks"
	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/promutils/labeled"
	"github.com/lyft/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type workflowNodeStateHolder struct {
	s handler.WorkflowNodeState
}

func (t *workflowNodeStateHolder) PutTaskNodeState(s handler.TaskNodeState) error {
	panic("not implemented")
}

func (t workflowNodeStateHolder) PutBranchNode(s handler.BranchNodeState) error {
	panic("not implemented")
}

func (t *workflowNodeStateHolder) PutWorkflowNodeState(s handler.WorkflowNodeState) error {
	t.s = s
	return nil
}

func (t workflowNodeStateHolder) PutDynamicNodeState(s handler.DynamicNodeState) error {
	panic("not implemented")
}

func createNodeContext(phase v1alpha1.WorkflowNodePhase, w v1alpha1.ExecutableWorkflow, n v1alpha1.ExecutableNode) *mocks3.NodeExecutionContext {

	wfNodeState := handler.WorkflowNodeState{}
	s := &workflowNodeStateHolder{s: wfNodeState}

	wfExecID := &core.WorkflowExecutionIdentifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
	}

	nm := &mocks3.NodeExecutionMetadata{}
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

	ir := &mocks4.InputReader{}
	inputs := &core.LiteralMap{}
	ir.On("Get", mock.Anything).Return(inputs, nil)

	nCtx := &mocks3.NodeExecutionContext{}
	nCtx.On("Node").Return(n)
	nCtx.On("NodeExecutionMetadata").Return(nm)
	nCtx.On("InputReader").Return(ir)
	nCtx.On("CurrentAttempt").Return(uint32(1))
	nCtx.On("MaxDatasetSizeBytes").Return(int64(1))
	nCtx.On("NodeStatus").Return(ns)
	nCtx.On("NodeID").Return("n1")
	nCtx.On("EnqueueOwner").Return(nil)
	nCtx.On("Workflow").Return(w)

	nr := &mocks3.NodeStateReader{}
	nr.On("GetWorkflowNodeState").Return(handler.WorkflowNodeState{
		Phase: phase,
	})
	nCtx.On("NodeStateReader").Return(nr)
	nCtx.On("NodeStateWriter").Return(s)
	return nCtx
}

func TestWorkflowNodeHandler_StartNode_Launchplan(t *testing.T) {
	ctx := context.TODO()

	nodeID := "n1"
	attempts := uint32(1)

	lpID := &core.Identifier{
		Project:      "p",
		Domain:       "d",
		Name:         "n",
		Version:      "v",
		ResourceType: core.ResourceType_LAUNCH_PLAN,
	}
	mockWfNode := &mocks2.ExecutableWorkflowNode{}
	mockWfNode.On("GetLaunchPlanRefID").Return(&v1alpha1.Identifier{
		Identifier: lpID,
	})
	mockWfNode.On("GetSubWorkflowRef").Return(nil)

	mockNode := &mocks2.ExecutableNode{}
	mockNode.On("GetID").Return("n1")
	mockNode.On("GetWorkflowNode").Return(mockWfNode)

	mockNodeStatus := &mocks2.ExecutableNodeStatus{}
	mockNodeStatus.On("GetAttempts").Return(attempts)
	wfStatus := &mocks2.MutableWorkflowNodeStatus{}
	mockNodeStatus.On("GetOrCreateWorkflowStatus").Return(wfStatus)
	parentID := &core.WorkflowExecutionIdentifier{
		Name:    "x",
		Domain:  "y",
		Project: "z",
	}
	mockWf := &mocks2.ExecutableWorkflow{}
	mockWf.On("GetNodeExecutionStatus", nodeID).Return(mockNodeStatus)
	mockWf.On("GetExecutionID").Return(v1alpha1.WorkflowExecutionIdentifier{
		WorkflowExecutionIdentifier: parentID,
	})

	t.Run("happy", func(t *testing.T) {

		mockLPExec := &mocks.Executor{}
		h := New(nil, mockLPExec, promutils.NewTestScope())
		mockLPExec.On("Launch",
			ctx,
			mock.MatchedBy(func(o launchplan.LaunchContext) bool {
				return o.ParentNodeExecution.NodeId == mockNode.GetID() &&
					o.ParentNodeExecution.ExecutionId == parentID
			}),
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return o.Project == parentID.Project && o.Domain == parentID.Domain
			}),
			mock.MatchedBy(func(o *core.Identifier) bool { return lpID == o }),
			mock.MatchedBy(func(o *core.LiteralMap) bool { return o.Literals == nil }),
		).Return(nil)

		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseUndefined, mockWf, mockNode)
		s, err := h.Handle(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, handler.EPhaseRunning, s.Info().GetPhase())
	})
}

func TestWorkflowNodeHandler_CheckNodeStatus(t *testing.T) {
	ctx := context.TODO()

	nodeID := "n1"
	attempts := uint32(1)
	dataDir := storage.DataReference("data")

	lpID := &core.Identifier{
		Project:      "p",
		Domain:       "d",
		Name:         "n",
		Version:      "v",
		ResourceType: core.ResourceType_LAUNCH_PLAN,
	}
	mockWfNode := &mocks2.ExecutableWorkflowNode{}
	mockWfNode.On("GetLaunchPlanRefID").Return(&v1alpha1.Identifier{
		Identifier: lpID,
	})
	mockWfNode.On("GetSubWorkflowRef").Return(nil)

	mockNode := &mocks2.ExecutableNode{}
	mockNode.On("GetID").Return("n1")
	mockNode.On("GetWorkflowNode").Return(mockWfNode)

	mockNodeStatus := &mocks2.ExecutableNodeStatus{}
	mockNodeStatus.On("GetAttempts").Return(attempts)
	mockNodeStatus.On("GetDataDir").Return(dataDir)

	parentID := &core.WorkflowExecutionIdentifier{
		Name:    "x",
		Domain:  "y",
		Project: "z",
	}
	mockWf := &mocks2.ExecutableWorkflow{}
	mockWf.On("GetNodeExecutionStatus", nodeID).Return(mockNodeStatus)
	mockWf.On("GetExecutionID").Return(v1alpha1.WorkflowExecutionIdentifier{
		WorkflowExecutionIdentifier: parentID,
	})

	t.Run("stillRunning", func(t *testing.T) {

		mockLPExec := &mocks.Executor{}

		h := New(nil, mockLPExec, promutils.NewTestScope())
		mockLPExec.On("GetStatus",
			ctx,
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return o.Project == parentID.Project && o.Domain == parentID.Domain
			}),
		).Return(&admin.ExecutionClosure{
			Phase: core.WorkflowExecution_RUNNING,
		}, nil)

		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockWf, mockNode)
		s, err := h.Handle(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, handler.EPhaseRunning, s.Info().GetPhase())
	})
}

func TestWorkflowNodeHandler_AbortNode(t *testing.T) {
	ctx := context.TODO()

	nodeID := "n1"
	attempts := uint32(1)
	dataDir := storage.DataReference("data")

	lpID := &core.Identifier{
		Project:      "p",
		Domain:       "d",
		Name:         "n",
		Version:      "v",
		ResourceType: core.ResourceType_LAUNCH_PLAN,
	}
	mockWfNode := &mocks2.ExecutableWorkflowNode{}
	mockWfNode.On("GetLaunchPlanRefID").Return(&v1alpha1.Identifier{
		Identifier: lpID,
	})
	mockWfNode.On("GetSubWorkflowRef").Return(nil)

	mockNode := &mocks2.ExecutableNode{}
	mockNode.On("GetID").Return("n1")
	mockNode.On("GetWorkflowNode").Return(mockWfNode)

	mockNodeStatus := &mocks2.ExecutableNodeStatus{}
	mockNodeStatus.On("GetAttempts").Return(attempts)
	mockNodeStatus.On("GetDataDir").Return(dataDir)

	parentID := &core.WorkflowExecutionIdentifier{
		Name:    "x",
		Domain:  "y",
		Project: "z",
	}
	mockWf := &mocks2.ExecutableWorkflow{}
	mockWf.On("GetNodeExecutionStatus", nodeID).Return(mockNodeStatus)
	mockWf.On("GetName").Return("test")
	mockWf.On("GetExecutionID").Return(v1alpha1.WorkflowExecutionIdentifier{
		WorkflowExecutionIdentifier: parentID,
	})

	t.Run("abort", func(t *testing.T) {

		mockLPExec := &mocks.Executor{}
		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockWf, mockNode)

		h := New(nil, mockLPExec, promutils.NewTestScope())
		mockLPExec.On("Kill",
			ctx,
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return o.Project == parentID.Project && o.Domain == parentID.Domain
			}),
			mock.AnythingOfType(reflect.String.String()),
		).Return(nil)

		err := h.Abort(ctx, nCtx, "test")
		assert.NoError(t, err)
	})

	t.Run("abort-fail", func(t *testing.T) {

		mockLPExec := &mocks.Executor{}
		expectedErr := fmt.Errorf("fail")
		h := New(nil, mockLPExec, promutils.NewTestScope())
		mockLPExec.On("Kill",
			ctx,
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return o.Project == parentID.Project && o.Domain == parentID.Domain
			}),
			mock.AnythingOfType(reflect.String.String()),
		).Return(expectedErr)

		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockWf, mockNode)
		err := h.Abort(ctx, nCtx, "test")
		assert.Error(t, err)
		assert.Equal(t, err, expectedErr)
	})
}

func init() {
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)
}
