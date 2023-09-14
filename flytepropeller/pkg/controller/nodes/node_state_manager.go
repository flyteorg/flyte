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
	g          *handler.GateNodeState
	a          *handler.ArrayNodeState
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

func (n *nodeStateManager) PutGateNodeState(s handler.GateNodeState) error {
	n.g = &s
	return nil
}

func (n *nodeStateManager) PutArrayNodeState(s handler.ArrayNodeState) error {
	n.a = &s
	return nil
}

func (n *nodeStateManager) HasTaskNodeState() bool {
	return n.t != nil
}

func (n *nodeStateManager) HasBranchNodeState() bool {
	return n.b != nil
}

func (n *nodeStateManager) HasDynamicNodeState() bool {
	return n.d != nil
}

func (n *nodeStateManager) HasWorkflowNodeState() bool {
	return n.w != nil
}

func (n *nodeStateManager) HasGateNodeState() bool {
	return n.g != nil
}

func (n *nodeStateManager) HasArrayNodeState() bool {
	return n.a != nil
}

func (n nodeStateManager) GetTaskNodeState() handler.TaskNodeState {
	if n.t != nil {
		return *n.t
	}

	tn := n.nodeStatus.GetTaskNodeStatus()
	if tn != nil {
		return handler.TaskNodeState{
			PluginPhase:                        pluginCore.Phase(tn.GetPhase()),
			PluginPhaseVersion:                 tn.GetPhaseVersion(),
			PluginStateVersion:                 tn.GetPluginStateVersion(),
			PluginState:                        tn.GetPluginState(),
			LastPhaseUpdatedAt:                 tn.GetLastPhaseUpdatedAt(),
			PreviousNodeExecutionCheckpointURI: tn.GetPreviousNodeExecutionCheckpointPath(),
			CleanupOnFailure:                   tn.GetCleanupOnFailure(),
		}
	}
	return handler.TaskNodeState{}
}

func (n nodeStateManager) GetBranchNodeState() handler.BranchNodeState {
	if n.b != nil {
		return *n.b
	}

	bn := n.nodeStatus.GetBranchStatus()
	bs := handler.BranchNodeState{}
	if bn != nil {
		bs.Phase = bn.GetPhase()
		bs.FinalizedNodeID = bn.GetFinalizedNode()
	}
	return bs
}

func (n nodeStateManager) GetDynamicNodeState() handler.DynamicNodeState {
	if n.d != nil {
		return *n.d
	}

	dn := n.nodeStatus.GetDynamicNodeStatus()
	ds := handler.DynamicNodeState{}
	if dn != nil {
		ds.Phase = dn.GetDynamicNodePhase()
		ds.Reason = dn.GetDynamicNodeReason()
		ds.Error = dn.GetExecutionError()
		ds.IsFailurePermanent = dn.GetIsFailurePermanent()
	}

	return ds
}

func (n nodeStateManager) GetWorkflowNodeState() handler.WorkflowNodeState {
	if n.w != nil {
		return *n.w
	}

	wn := n.nodeStatus.GetWorkflowNodeStatus()
	ws := handler.WorkflowNodeState{}
	if wn != nil {
		ws.Phase = wn.GetWorkflowNodePhase()
		ws.Error = wn.GetExecutionError()
	}
	return ws
}

func (n nodeStateManager) GetGateNodeState() handler.GateNodeState {
	if n.g != nil {
		return *n.g
	}

	gn := n.nodeStatus.GetGateNodeStatus()
	gs := handler.GateNodeState{}
	if gn != nil {
		gs.Phase = gn.GetGateNodePhase()
	}
	return gs
}

func (n nodeStateManager) GetArrayNodeState() handler.ArrayNodeState {
	if n.a != nil {
		return *n.a
	}

	an := n.nodeStatus.GetArrayNodeStatus()
	as := handler.ArrayNodeState{}
	if an != nil {
		as.Phase = an.GetArrayNodePhase()
		as.Error = an.GetExecutionError()
		as.SubNodePhases = an.GetSubNodePhases()
		as.SubNodeTaskPhases = an.GetSubNodeTaskPhases()
		as.SubNodeRetryAttempts = an.GetSubNodeRetryAttempts()
		as.SubNodeSystemFailures = an.GetSubNodeSystemFailures()
		as.TaskPhaseVersion = an.GetTaskPhaseVersion()
	}
	return as
}

func (n *nodeStateManager) ClearNodeStatus() {
	n.t = nil
	n.b = nil
	n.d = nil
	n.w = nil
	n.g = nil
	n.a = nil
	n.nodeStatus.ClearLastAttemptStartedAt()
}

func newNodeStateManager(_ context.Context, status v1alpha1.ExecutableNodeStatus) *nodeStateManager {
	return &nodeStateManager{nodeStatus: status}
}
