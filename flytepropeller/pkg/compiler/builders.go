package compiler

import (
	"fmt"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	c "github.com/flyteorg/flytepropeller/pkg/compiler/common"
)

type flyteTask = core.TaskTemplate //nolint:unused
type flyteWorkflow = core.CompiledWorkflow
type flyteNode = core.Node

// A builder object for the Graph struct. This contains information the compiler uses while building the final Graph
// struct.
type workflowBuilder struct {
	CoreWorkflow     *flyteWorkflow
	LaunchPlans      map[c.WorkflowIDKey]c.InterfaceProvider
	Tasks            c.TaskIndex
	downstreamNodes  c.StringAdjacencyList
	upstreamNodes    c.StringAdjacencyList
	Nodes            c.NodeIndex
	NodeBuilderIndex c.NodeIndex

	// These are references to all subgraphs and tasks passed to CompileWorkflow. They will be passed around but will
	// not show in their entirety in the final Graph. The required subset of these will be added to each subgraph as
	// the compile traverses them.
	allLaunchPlans          map[string]c.InterfaceProvider
	allTasks                c.TaskIndex
	allSubWorkflows         c.WorkflowIndex
	allCompiledSubWorkflows c.WorkflowIndex
}

func (w workflowBuilder) GetFailureNode() c.Node {
	if w.GetCoreWorkflow() != nil && w.GetCoreWorkflow().GetTemplate() != nil && w.GetCoreWorkflow().GetTemplate().FailureNode != nil {
		return w.GetOrCreateNodeBuilder(w.GetCoreWorkflow().GetTemplate().FailureNode)
	}

	return nil
}

func (w workflowBuilder) GetNodes() c.NodeIndex {
	return w.Nodes
}

func (w workflowBuilder) GetTasks() c.TaskIndex {
	return w.Tasks
}

func (w workflowBuilder) GetDownstreamNodes() c.StringAdjacencyList {
	return w.downstreamNodes
}

func (w workflowBuilder) GetUpstreamNodes() c.StringAdjacencyList {
	return w.upstreamNodes
}

func (w workflowBuilder) GetOrCreateNodeBuilder(n *flyteNode) c.NodeBuilder {
	address := fmt.Sprintf("%p", n)
	if existingBuilder, found := w.NodeBuilderIndex[address]; found {
		return existingBuilder
	}

	newObj := &nodeBuilder{flyteNode: n}
	w.NodeBuilderIndex[address] = newObj
	return newObj
}

func (w workflowBuilder) GetNode(id c.NodeID) (node c.NodeBuilder, found bool) {
	node, found = w.Nodes[id]
	return
}

func (w workflowBuilder) GetTask(id c.TaskID) (task c.Task, found bool) {
	task, found = w.Tasks[id.String()]
	return
}

func (w workflowBuilder) GetLaunchPlan(id c.LaunchPlanID) (wf c.InterfaceProvider, found bool) {
	wf, found = w.LaunchPlans[id.String()]
	return
}

func (w workflowBuilder) StoreCompiledSubWorkflow(id c.WorkflowID, compiledWorkflow *core.CompiledWorkflow) {
	w.allCompiledSubWorkflows[id.String()] = compiledWorkflow
}

func (w workflowBuilder) GetCompiledSubWorkflow(id c.WorkflowID) (wf *core.CompiledWorkflow, found bool) {
	wf, found = w.allCompiledSubWorkflows[id.String()]
	return
}

func (w workflowBuilder) GetSubWorkflow(id c.WorkflowID) (wf *core.CompiledWorkflow, found bool) {
	wf, found = w.allSubWorkflows[id.String()]
	return
}

func (w workflowBuilder) GetCoreWorkflow() *flyteWorkflow {
	return w.CoreWorkflow
}

// A wrapper around core.nodeBuilder to augment with computed fields during compilation
type nodeBuilder struct {
	*flyteNode
	subWorkflow c.Workflow
	Task        c.Task
	Iface       *core.TypedInterface
}

func (n nodeBuilder) GetTask() c.Task {
	return n.Task
}

func (n *nodeBuilder) SetTask(task c.Task) {
	n.Task = task
}

func (n nodeBuilder) GetSubWorkflow() c.Workflow {
	return n.subWorkflow
}

func (n nodeBuilder) GetCoreNode() *core.Node {
	return n.flyteNode
}

func (n nodeBuilder) GetInterface() *core.TypedInterface {
	return n.Iface
}

func (n *nodeBuilder) SetInterface(iface *core.TypedInterface) {
	n.Iface = iface
}

func (n *nodeBuilder) SetSubWorkflow(wf c.Workflow) {
	n.subWorkflow = wf
}

func (n *nodeBuilder) SetInputs(inputs []*core.Binding) {
	n.Inputs = inputs
}

func (n *nodeBuilder) SetID(id string) {
	n.Id = id
}

type taskBuilder struct {
	*flyteTask
}

func (t taskBuilder) GetCoreTask() *core.TaskTemplate {
	return t.flyteTask
}

func (t taskBuilder) GetID() c.Identifier {
	if t.flyteTask.Id != nil {
		return *t.flyteTask.Id
	}

	return c.Identifier{}
}
