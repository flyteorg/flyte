package common

import (
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

type NodeID = string
type TaskID = Identifier
type WorkflowID = Identifier
type LaunchPlanID = Identifier
type TaskIDKey = string
type WorkflowIDKey = string

// An immutable workflow that represents the final output of the compiler.
type Workflow interface {
	GetNode(id NodeID) (node NodeBuilder, found bool)
	GetTask(id TaskID) (task Task, found bool)
	GetLaunchPlan(id LaunchPlanID) (wf InterfaceProvider, found bool)
	GetSubWorkflow(id WorkflowID) (wf *core.CompiledWorkflow, found bool)
	GetCompiledSubWorkflow(id WorkflowID) (wf *core.CompiledWorkflow, found bool)
	GetCoreWorkflow() *core.CompiledWorkflow
	GetFailureNode() Node
	GetNodes() NodeIndex
	GetTasks() TaskIndex
	GetDownstreamNodes() StringAdjacencyList
	GetUpstreamNodes() StringAdjacencyList
}

// An immutable Node that represents the final output of the compiler.
type Node interface {
	GetId() NodeID
	GetInterface() *core.TypedInterface
	GetInputs() []*core.Binding
	GetWorkflowNode() *core.WorkflowNode
	GetOutputAliases() []*core.Alias
	GetUpstreamNodeIds() []string
	GetCoreNode() *core.Node
	GetBranchNode() *core.BranchNode
	GetTaskNode() *core.TaskNode
	GetMetadata() *core.NodeMetadata
	GetTask() Task
	GetSubWorkflow() Workflow
	GetGateNode() *core.GateNode
	GetArrayNode() *core.ArrayNode
}

// An immutable task that represents the final output of the compiler.
type Task interface {
	GetID() TaskID
	GetCoreTask() *core.TaskTemplate
	GetInterface() *core.TypedInterface
}

type InterfaceProvider interface {
	GetID() *core.Identifier
	GetExpectedInputs() *core.ParameterMap
	GetExpectedOutputs() *core.VariableMap
}
