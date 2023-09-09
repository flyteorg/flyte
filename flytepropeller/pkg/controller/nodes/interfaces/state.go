package interfaces

import (
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/handler"
)

type NodeStateWriter interface {
	PutTaskNodeState(s handler.TaskNodeState) error
	PutBranchNode(s handler.BranchNodeState) error
	PutDynamicNodeState(s handler.DynamicNodeState) error
	PutWorkflowNodeState(s handler.WorkflowNodeState) error
	PutGateNodeState(s handler.GateNodeState) error
	PutArrayNodeState(s handler.ArrayNodeState) error
	ClearNodeStatus()
}

type NodeStateReader interface {
	HasTaskNodeState() bool
	GetTaskNodeState() handler.TaskNodeState
	HasBranchNodeState() bool
	GetBranchNodeState() handler.BranchNodeState
	HasDynamicNodeState() bool
	GetDynamicNodeState() handler.DynamicNodeState
	HasWorkflowNodeState() bool
	GetWorkflowNodeState() handler.WorkflowNodeState
	HasGateNodeState() bool
	GetGateNodeState() handler.GateNodeState
	HasArrayNodeState() bool
	GetArrayNodeState() handler.ArrayNodeState
}
