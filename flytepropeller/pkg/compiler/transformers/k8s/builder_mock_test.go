package k8s

import (
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/compiler/common"
)

type mockWorkflow struct {
	*core.CompiledWorkflow
	nodes       common.NodeIndex
	tasks       common.TaskIndex
	wfs         map[common.WorkflowIDKey]common.InterfaceProvider
	failureNode *mockNode
	downstream  common.StringAdjacencyList
	upstream    common.StringAdjacencyList
}

func (m mockWorkflow) GetCompiledSubWorkflow(id common.WorkflowID) (wf *core.CompiledWorkflow, found bool) {
	panic("method invocation not expected")
}

func (m mockWorkflow) GetSubWorkflow(id common.WorkflowID) (wf *core.CompiledWorkflow, found bool) {
	panic("method invocation not expected")
}

func (m mockWorkflow) GetNode(id common.NodeID) (node common.NodeBuilder, found bool) {
	node, found = m.nodes[id]
	return
}

func (m mockWorkflow) GetTask(id common.TaskID) (task common.Task, found bool) {
	task, found = m.tasks[id.String()]
	return
}

func (m mockWorkflow) GetLaunchPlan(id common.LaunchPlanID) (wf common.InterfaceProvider, found bool) {
	wf, found = m.wfs[id.String()]
	return
}

func (m mockWorkflow) GetCoreWorkflow() *core.CompiledWorkflow {
	return m.CompiledWorkflow
}

func (m mockWorkflow) GetFailureNode() common.Node {
	if m.failureNode == nil {
		return nil
	}

	return m.failureNode
}

func (m mockWorkflow) GetNodes() common.NodeIndex {
	return m.nodes
}

func (m mockWorkflow) GetTasks() common.TaskIndex {
	return m.tasks
}

func (m mockWorkflow) GetDownstreamNodes() common.StringAdjacencyList {
	return m.downstream
}

func (m mockWorkflow) GetUpstreamNodes() common.StringAdjacencyList {
	return m.upstream
}

type mockNode struct {
	*core.Node
	id       common.NodeID
	iface    *core.TypedInterface
	inputs   []*core.Binding
	aliases  []*core.Alias
	upstream []string
	task     common.Task
	subWF    common.Workflow
}

func (n *mockNode) SetID(id common.NodeID) {
	n.id = id
}

func (n mockNode) GetID() common.NodeID {
	return n.id
}

func (n mockNode) GetInterface() *core.TypedInterface {
	return n.iface
}

func (n mockNode) GetInputs() []*core.Binding {
	return n.inputs
}

func (n mockNode) GetOutputAliases() []*core.Alias {
	return n.aliases
}

func (n mockNode) GetUpstreamNodeIds() []string {
	return n.upstream
}

func (n mockNode) GetCoreNode() *core.Node {
	return n.Node
}

func (n mockNode) GetTask() common.Task {
	return n.task
}

func (n *mockNode) SetTask(task common.Task) {
	n.task = task
}

func (n mockNode) GetSubWorkflow() common.Workflow {
	return n.subWF
}

func (n *mockNode) SetInterface(iface *core.TypedInterface) {
	n.iface = iface
}

func (n *mockNode) SetInputs(inputs []*core.Binding) {
	n.inputs = inputs
}

func (n *mockNode) SetSubWorkflow(wf common.Workflow) {
	n.subWF = wf
}

type mockTask struct {
	id    common.TaskID
	task  *core.TaskTemplate
	iface *core.TypedInterface
}

func (m mockTask) GetID() common.TaskID {
	return m.id
}

func (m mockTask) GetCoreTask() *core.TaskTemplate {
	return m.task
}

func (m mockTask) GetInterface() *core.TypedInterface {
	return m.iface
}
