package nodes

import (
	"context"

	pluginCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/handler"
)

type nodeStateManager struct {
	nodeStatus v1alpha1.ExecutableNodeStatus
	t          *handler.TaskNodeState
	b          *handler.BranchNodeState
	d          *handler.DynamicNodeState
	w          *handler.WorkflowNodeState
}

func (n *nodeStateManager) PutTaskNodeState(s handler.TaskNodeState) error {
	n.t = &s
	return nil
}

func (n *nodeStateManager) PutBranchNode(s handler.BranchNodeState) error {
	n.b = &s
	return nil
}

func (n *nodeStateManager) PutDynamicNodeState(s handler.DynamicNodeState) error {
	n.d = &s
	return nil
}

func (n *nodeStateManager) PutWorkflowNodeState(s handler.WorkflowNodeState) error {
	n.w = &s
	return nil
}

func (n nodeStateManager) GetTaskNodeState() handler.TaskNodeState {
	tn := n.nodeStatus.GetTaskNodeStatus()
	if tn != nil {
		return handler.TaskNodeState{
			PluginPhase:        pluginCore.Phase(tn.GetPhase()),
			PluginPhaseVersion: tn.GetPhaseVersion(),
			PluginStateVersion: tn.GetPluginStateVersion(),
			PluginState:        tn.GetPluginState(),
			BarrierClockTick:   tn.GetBarrierClockTick(),
			LastPhaseUpdatedAt: tn.GetLastPhaseUpdatedAt(),
		}
	}
	return handler.TaskNodeState{}
}

func (n nodeStateManager) GetBranchNode() handler.BranchNodeState {
	bn := n.nodeStatus.GetBranchStatus()
	bs := handler.BranchNodeState{}
	if bn != nil {
		bs.Phase = bn.GetPhase()
		bs.FinalizedNodeID = bn.GetFinalizedNode()
	}
	return bs
}

func (n nodeStateManager) GetDynamicNodeState() handler.DynamicNodeState {
	dn := n.nodeStatus.GetDynamicNodeStatus()
	ds := handler.DynamicNodeState{}
	if dn != nil {
		ds.Phase = dn.GetDynamicNodePhase()
		ds.Reason = dn.GetDynamicNodeReason()
		ds.Error = dn.GetExecutionError()
	}

	return ds
}

func (n nodeStateManager) GetWorkflowNodeState() handler.WorkflowNodeState {
	wn := n.nodeStatus.GetWorkflowNodeStatus()
	ws := handler.WorkflowNodeState{}
	if wn != nil {
		ws.Phase = wn.GetWorkflowNodePhase()
	}
	return ws
}

func (n *nodeStateManager) clearNodeStatus() {
	n.t = nil
	n.b = nil
	n.d = nil
	n.w = nil
	n.nodeStatus.ClearLastAttemptStartedAt()
}

func newNodeStateManager(_ context.Context, status v1alpha1.ExecutableNodeStatus) *nodeStateManager {
	return &nodeStateManager{nodeStatus: status}
}
