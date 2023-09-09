// Package compiler provides compiler services for flyte workflows. It performs static analysis on the Workflow and produces
// CompilerErrors for any detected issue. A flyte workflow should only be considered valid for execution if it passed through
// the compiler first. The intended usage for the compiler is as follows:
// 1) Call GetRequirements(...) and load/retrieve all tasks/workflows referenced in the response.
// 2) Call CompileWorkflow(...) and make sure it reports no errors.
// 3) Use one of the transformer packages (e.g. transformer/k8s) to build the final executable workflow.
//
//	               +-------------------+
//	               | start(StartNode)  |
//	               +-------------------+
//	                 |
//	                 | wf_input
//	                 v
//	+--------+     +-------------------+
//	| static | --> | node_1(TaskNode)  |
//	+--------+     +-------------------+
//	  |              |
//	  |              | x
//	  |              v
//	  |            +-------------------+
//	  +----------> | node_2(TaskNode)  |
//	               +-------------------+
//	                 |
//	                 | n2_output
//	                 v
//	               +-------------------+
//	               |   end(EndNode)    |
//	               +-------------------+
//	               +-------------------+
//	               | Workflow Id: repo |
//	               +-------------------+
package compiler

import (
	"strings"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	c "github.com/flyteorg/flytepropeller/pkg/compiler/common"
	"github.com/flyteorg/flytepropeller/pkg/compiler/errors"
	v "github.com/flyteorg/flytepropeller/pkg/compiler/validators"

	// #noSA1019
	"github.com/golang/protobuf/proto"
	"k8s.io/apimachinery/pkg/util/sets"
)

// Updates workflows and tasks references to reflect the needed ones for this workflow (ignoring subworkflows)
func (w *workflowBuilder) updateRequiredReferences() {
	reqs := getRequirements(w.CoreWorkflow.Template, w.allSubWorkflows, false, errors.NewCompileErrors())
	workflows := map[c.WorkflowIDKey]c.InterfaceProvider{}
	tasks := c.TaskIndex{}
	for _, workflowID := range reqs.launchPlanIds {
		if wf, ok := w.allLaunchPlans[workflowID.String()]; ok {
			workflows[workflowID.String()] = wf
		}
	}

	for _, taskID := range reqs.taskIds {
		if task, ok := w.allTasks[taskID.String()]; ok {
			tasks[taskID.String()] = task
		}
	}

	w.Tasks = tasks
	w.LaunchPlans = workflows
}

// Validates the coreWorkflow contains no cycles and that all nodes are reachable.
func (w workflowBuilder) validateReachable(errs errors.CompileErrors) (ok bool) {
	neighbors := func(nodeId string) sets.String {
		downNodes := w.downstreamNodes[nodeId]
		if downNodes == nil {
			return sets.String{}
		}

		return downNodes
	}

	// TODO: If a branch node can exist in a cycle and not actually be a cycle since it can branch off...
	if cycle, visited, detected := detectCycle(c.StartNodeID, neighbors); detected {
		errs.Collect(errors.NewCycleDetectedInWorkflowErr(c.StartNodeID, strings.Join(cycle, ">")))
	} else {
		// If no cycles are detected, we expect all nodes to have been visited. Otherwise there are unreachable
		// node(s)..
		if visited.Len() != len(w.Nodes) {
			// report unreachable nodes
			allNodes := toNodeIdsSet(w.Nodes)
			unreachableNodes := allNodes.Difference(visited).Difference(sets.NewString(c.EndNodeID))
			if len(unreachableNodes) > 0 {
				errs.Collect(errors.NewUnreachableNodesErr(c.StartNodeID, strings.Join(toSlice(unreachableNodes), ",")))
			}
		}
	}

	return !errs.HasErrors()
}

// Adds unique nodes to the workflow.
func (w workflowBuilder) AddNode(n c.NodeBuilder, errs errors.CompileErrors) (node c.NodeBuilder, ok bool) {
	if _, ok := w.Nodes[n.GetId()]; ok {
		return n, !errs.HasErrors()
	}

	node = n
	w.Nodes[n.GetId()] = node
	ok = !errs.HasErrors()
	w.CoreWorkflow.Template.Nodes = append(w.CoreWorkflow.Template.Nodes, node.GetCoreNode())
	return
}

func (w workflowBuilder) AddUpstreamEdge(nodeProvider, nodeDependent c.NodeID) {
	if nodeProvider == "" {
		nodeProvider = c.StartNodeID
	}

	if _, found := w.upstreamNodes[nodeDependent]; !found {
		w.upstreamNodes[nodeDependent] = sets.String{}
		w.CoreWorkflow.Connections.Upstream[nodeDependent] = &core.ConnectionSet_IdList{
			Ids: make([]string, 1),
		}
	}

	w.upstreamNodes[nodeDependent].Insert(nodeProvider)
	w.CoreWorkflow.Connections.Upstream[nodeDependent].Ids = w.upstreamNodes[nodeDependent].List()
}

func (w workflowBuilder) AddDownstreamEdge(nodeProvider, nodeDependent c.NodeID) {
	if nodeProvider == "" {
		nodeProvider = c.StartNodeID
	}

	if _, found := w.downstreamNodes[nodeProvider]; !found {
		w.downstreamNodes[nodeProvider] = sets.String{}
		w.CoreWorkflow.Connections.Downstream[nodeProvider] = &core.ConnectionSet_IdList{}
	}

	w.downstreamNodes[nodeProvider].Insert(nodeDependent)
	w.CoreWorkflow.Connections.Downstream[nodeProvider].Ids = w.downstreamNodes[nodeProvider].List()
}

func (w workflowBuilder) AddExecutionEdge(nodeFrom, nodeTo c.NodeID) {
	w.AddDownstreamEdge(nodeFrom, nodeTo)
	w.AddUpstreamEdge(nodeFrom, nodeTo)
}

func (w workflowBuilder) AddEdges(n c.NodeBuilder, edgeDirection c.EdgeDirection, errs errors.CompileErrors) (ok bool) {
	if n.GetInterface() == nil {
		// If there were errors computing node's interface, don't add any edges and just bail.
		return
	}

	// Add explicitly declared edges
	switch edgeDirection {
	case c.EdgeDirectionDownstream:
		fallthrough
	case c.EdgeDirectionBidirectional:
		for _, upNode := range n.GetUpstreamNodeIds() {
			w.AddExecutionEdge(upNode, n.GetId())
		}
	}

	// Add implicit Edges
	_, ok = v.ValidateBindings(&w, n, n.GetInputs(), n.GetInterface().GetInputs(),
		true /* validateParamTypes */, edgeDirection, errs.NewScope())
	return
}

// Contains the main validation logic for the coreWorkflow. If successful, it'll build an executable Workflow.
func (w workflowBuilder) ValidateWorkflow(fg *flyteWorkflow, errs errors.CompileErrors) (c.Workflow, bool) {
	if len(fg.Template.Nodes) == 0 {
		errs.Collect(errors.NewNoNodesFoundErr(fg.Template.Id.String()))
		return nil, !errs.HasErrors()
	}

	// Initialize workflow
	wf := w.newWorkflowBuilder(fg)
	wf.updateRequiredReferences()

	// Start building out the workflow
	// Create global sentinel nodeBuilder with the workflow as its interface.
	startNode := &core.Node{
		Id: c.StartNodeID,
	}

	var ok bool
	if wf.CoreWorkflow.Template.Interface, ok = v.ValidateInterface(c.StartNodeID, wf.CoreWorkflow.Template.Interface, errs.NewScope()); !ok {
		return nil, !errs.HasErrors()
	}

	checkpoint := make([]*core.Node, 0, len(fg.Template.Nodes))
	checkpoint = append(checkpoint, fg.Template.Nodes...)
	fg.Template.Nodes = make([]*core.Node, 0, len(fg.Template.Nodes))
	wf.GetCoreWorkflow().Connections = &core.ConnectionSet{
		Downstream: make(map[string]*core.ConnectionSet_IdList),
		Upstream:   make(map[string]*core.ConnectionSet_IdList),
	}

	globalInputNode, _ := wf.AddNode(wf.GetOrCreateNodeBuilder(startNode), errs)
	globalInputNode.SetInterface(&core.TypedInterface{Outputs: wf.CoreWorkflow.Template.Interface.Inputs})

	endNode := &core.Node{Id: c.EndNodeID}
	globalOutputNode, _ := wf.AddNode(wf.GetOrCreateNodeBuilder(endNode), errs)
	globalOutputNode.SetInterface(&core.TypedInterface{Inputs: wf.CoreWorkflow.Template.Interface.Outputs})
	globalOutputNode.SetInputs(wf.CoreWorkflow.Template.Outputs)

	// Track top level nodes (a branch in a branch node is NOT a top level node). The final graph should ensure that all
	// top level nodes are executed before the end node. We do that by adding execution edges from leaf nodes that do not
	// contribute to the final outputs to the end node.
	topLevelNodes := sets.NewString()

	// Add and validate all other nodes
	for _, n := range checkpoint {
		topLevelNodes.Insert(n.Id)
		if node, addOk := wf.AddNode(wf.GetOrCreateNodeBuilder(n), errs.NewScope()); addOk {
			v.ValidateNode(&wf, node, false /* validateConditionTypes */, errs.NewScope())
		}
	}

	// Add explicitly and implicitly declared edges
	for nodeID, n := range wf.Nodes {
		if nodeID == c.StartNodeID {
			continue
		}

		wf.AddEdges(n, c.EdgeDirectionBidirectional, errs.NewScope())
	}

	// Add execution edges for orphan nodes that don't have any inward/outward edges.
	for nodeID := range wf.Nodes {
		if nodeID == c.StartNodeID || nodeID == c.EndNodeID {
			continue
		}

		// Nodes that do not have a upstream dependencies means they do not rely on the workflow inputs to execute.
		// This is a rare but possible occurrence, by explicitly adding an execution edge from the start node to these
		// nodes, we ensure that propeller starts executing the workflow by running all such nodes and then their
		// downstream dependencies.
		if topLevelNodes.Has(nodeID) {
			if _, foundUpStream := wf.upstreamNodes[nodeID]; !foundUpStream {
				wf.AddExecutionEdge(c.StartNodeID, nodeID)
			}
		}

		// When propeller executes nodes it'll ensure that any node does not start executing until all of its upstream
		// dependencies have finished successfully. By explicitly adding execution edges from such nodes to end-node, we
		// ensure that execution continues until all nodes successfully finish.
		if topLevelNodes.Has(nodeID) {
			if _, foundDownStream := wf.downstreamNodes[nodeID]; !foundDownStream {
				wf.AddExecutionEdge(nodeID, c.EndNodeID)
			}
		}
	}

	// Validate workflow outputs are bound
	if _, wfIfaceOk := v.ValidateInterface(globalOutputNode.GetId(), globalOutputNode.GetInterface(), errs.NewScope()); wfIfaceOk {
		v.ValidateBindings(&wf, globalOutputNode, globalOutputNode.GetInputs(),
			globalOutputNode.GetInterface().GetInputs(), true, /* validateParamTypes */
			c.EdgeDirectionBidirectional, errs.NewScope())
	}

	// Validate no cycles are detected.
	wf.validateReachable(errs.NewScope())

	return wf, !errs.HasErrors()
}

// Validates that all requirements for the coreWorkflow and its subworkflows are present.
func (w workflowBuilder) validateAllRequirements(errs errors.CompileErrors) bool {
	reqs := getRequirements(w.CoreWorkflow.Template, w.allSubWorkflows, true, errs)

	for _, lp := range reqs.launchPlanIds {
		if _, ok := w.allLaunchPlans[lp.String()]; !ok {
			errs.Collect(errors.NewWorkflowReferenceNotFoundErr(c.StartNodeID, lp.String()))
		}
	}

	for _, taskID := range reqs.taskIds {
		if _, ok := w.allTasks[taskID.String()]; !ok {
			errs.Collect(errors.NewTaskReferenceNotFoundErr(c.StartNodeID, taskID.String()))
		}
	}

	return !errs.HasErrors()
}

// CompileWorkflow compiles a flyte workflow a and all of its dependencies into a single executable Workflow. Refer to
// GetRequirements() to obtain a list of launchplan and Task ids to load/compile first.
// Returns an executable Workflow (if no errors are found) or a list of errors that must be addressed before the Workflow
// can be executed. Cast the error to errors.CompileErrors to inspect individual errors.
func CompileWorkflow(primaryWf *core.WorkflowTemplate, subworkflows []*core.WorkflowTemplate, tasks []*core.CompiledTask,
	launchPlans []c.InterfaceProvider) (*core.CompiledWorkflowClosure, error) {

	errs := errors.NewCompileErrors()

	if primaryWf == nil {
		errs.Collect(errors.NewValueRequiredErr("root", "wf"))
		return nil, errs
	}

	wf := proto.Clone(primaryWf).(*core.WorkflowTemplate)

	if tasks == nil {
		errs.Collect(errors.NewValueRequiredErr("root", "tasks"))
		return nil, errs
	}

	// Validate all tasks are valid... invalid tasks won't be passed on to the workflow validator
	uniqueTasks := sets.NewString()
	taskBuilders := make([]c.Task, 0, len(tasks))
	for _, task := range tasks {
		if task.Template == nil || task.Template.Id == nil {
			errs.Collect(errors.NewValueRequiredErr("task", "Template.Id"))
			return nil, errs
		}

		if uniqueTasks.Has(task.Template.Id.String()) {
			continue
		}

		taskBuilders = append(taskBuilders, &taskBuilder{flyteTask: task.Template})
		uniqueTasks.Insert(task.Template.Id.String())
	}

	// Validate overall requirements of the coreWorkflow.
	compiledSubWorkflows := toCompiledWorkflows(subworkflows...)
	wfIndex, ok := c.NewWorkflowIndex(compiledSubWorkflows, errs.NewScope())
	if !ok {
		return nil, errs
	}

	compiledWf := &core.CompiledWorkflow{Template: wf}

	gb := newWorkflowBuilder(compiledWf, wfIndex, c.NewTaskIndex(taskBuilders...), toInterfaceProviderMap(launchPlans))
	// Terminate early if there are some required component not present.
	if !gb.validateAllRequirements(errs.NewScope()) {
		return nil, errs
	}

	validatedWf, ok := gb.ValidateWorkflow(compiledWf, errs.NewScope())
	if ok {
		compiledTasks := make([]*core.CompiledTask, 0, len(taskBuilders))
		for _, t := range taskBuilders {
			compiledTasks = append(compiledTasks, &core.CompiledTask{Template: t.GetCoreTask()})
		}

		coreWf := validatedWf.GetCoreWorkflow()

		return &core.CompiledWorkflowClosure{
			Primary:      coreWf,
			Tasks:        compiledTasks,
			SubWorkflows: compiledSubWorkflows,
		}, nil
	}

	return nil, errs
}

func (w workflowBuilder) newWorkflowBuilder(fg *flyteWorkflow) workflowBuilder {
	return newWorkflowBuilder(fg, w.allSubWorkflows, w.allTasks, w.allLaunchPlans)
}

func newWorkflowBuilder(fg *flyteWorkflow, wfIndex c.WorkflowIndex, tasks c.TaskIndex,
	workflows map[string]c.InterfaceProvider) workflowBuilder {

	return workflowBuilder{
		CoreWorkflow:            fg,
		LaunchPlans:             map[string]c.InterfaceProvider{},
		Nodes:                   c.NewNodeIndex(),
		NodeBuilderIndex:        c.NewNodeIndex(),
		Tasks:                   c.NewTaskIndex(),
		downstreamNodes:         c.StringAdjacencyList{},
		upstreamNodes:           c.StringAdjacencyList{},
		allSubWorkflows:         wfIndex,
		allCompiledSubWorkflows: c.WorkflowIndex{},
		allLaunchPlans:          workflows,
		allTasks:                tasks,
	}
}
