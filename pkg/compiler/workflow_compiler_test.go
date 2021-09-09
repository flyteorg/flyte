package compiler

import (
	"fmt"
	"strings"
	"testing"

	"github.com/flyteorg/flytepropeller/pkg/compiler/common/mocks"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/compiler/common"
	"github.com/flyteorg/flytepropeller/pkg/compiler/errors"
	v "github.com/flyteorg/flytepropeller/pkg/compiler/validators"
	"github.com/flyteorg/flytepropeller/pkg/visualize"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/sets"
)

func createEmptyVariableMap() *core.VariableMap {
	res := &core.VariableMap{
		Variables: map[string]*core.Variable{},
	}
	return res
}

func createVariableMap(variableMap map[string]*core.Variable) *core.VariableMap {
	res := &core.VariableMap{
		Variables: variableMap,
	}
	return res
}

func dumpIdentifierNames(ids []common.Identifier) []string {
	res := make([]string, 0, len(ids))

	for _, id := range ids {
		res = append(res, id.Name)
	}

	return res
}

func ExampleCompileWorkflow_basic() {
	inputWorkflow := &core.WorkflowTemplate{
		Id: &core.Identifier{Name: "repo"},
		Interface: &core.TypedInterface{
			Inputs:  createEmptyVariableMap(),
			Outputs: createEmptyVariableMap(),
		},
		Nodes: []*core.Node{
			{
				Id: "FirstNode",
				Target: &core.Node_TaskNode{
					TaskNode: &core.TaskNode{
						Reference: &core.TaskNode_ReferenceId{
							ReferenceId: &core.Identifier{Name: "task_123"},
						},
					},
				},
			},
		},
	}

	// Detect what other workflows/tasks does this coreWorkflow reference
	subWorkflows := make([]*core.WorkflowTemplate, 0)
	reqs, err := GetRequirements(inputWorkflow, subWorkflows)
	if err != nil {
		fmt.Printf("failed to get requirements. Error: %v", err)
		return
	}

	fmt.Printf("Needed Tasks: [%v], Needed Workflows [%v]\n",
		strings.Join(dumpIdentifierNames(reqs.GetRequiredTaskIds()), ","),
		strings.Join(dumpIdentifierNames(reqs.GetRequiredLaunchPlanIds()), ","))

	// Replace with logic to satisfy the requirements
	workflows := make([]common.InterfaceProvider, 0)
	tasks := []*core.TaskTemplate{
		{
			Id: &core.Identifier{Name: "task_123"},
			Interface: &core.TypedInterface{
				Inputs:  createEmptyVariableMap(),
				Outputs: createEmptyVariableMap(),
			},
			Target: &core.TaskTemplate_Container{
				Container: &core.Container{
					Image:   "image://",
					Command: []string{"cmd"},
					Args:    []string{"args"},
				},
			},
		},
	}

	compiledTasks := make([]*core.CompiledTask, 0, len(tasks))
	for _, task := range tasks {
		compiledTask, err := CompileTask(task)
		if err != nil {
			fmt.Printf("failed to compile task [%v]. Error: %v", task.Id, err)
			return
		}

		compiledTasks = append(compiledTasks, compiledTask)
	}

	output, errs := CompileWorkflow(inputWorkflow, subWorkflows, compiledTasks, workflows)
	fmt.Printf("Compiled Workflow in GraphViz: %v\n", visualize.ToGraphViz(output.Primary))
	fmt.Printf("Compile Errors: %v\n", errs)

	// Output:
	// Needed Tasks: [task_123], Needed Workflows []
	// Compiled Workflow in GraphViz: digraph G {rankdir=TB;workflow[label="Workflow Id: name:"repo" "];node[style=filled];"start-node(start)" [shape=Msquare];"start-node(start)" -> "FirstNode()" [label="execution",style="dashed"];"FirstNode()" -> "end-node(end)" [label="execution",style="dashed"];}
	// Compile Errors: <nil>
}

func ExampleCompileWorkflow_inputsOutputsBinding() {
	inputWorkflow := &core.WorkflowTemplate{
		Id: &core.Identifier{Name: "repo"},
		Interface: &core.TypedInterface{
			Inputs: createVariableMap(map[string]*core.Variable{
				"wf_input": {
					Type: getIntegerLiteralType(),
				},
			}),
			Outputs: createVariableMap(map[string]*core.Variable{
				"wf_output": {
					Type: getIntegerLiteralType(),
				},
			}),
		},
		Nodes: []*core.Node{
			{
				Id: "node_1",
				Target: &core.Node_TaskNode{
					TaskNode: &core.TaskNode{Reference: &core.TaskNode_ReferenceId{ReferenceId: &core.Identifier{Name: "task_123"}}},
				},
				Inputs: []*core.Binding{
					newVarBinding("", "wf_input", "x"), newIntegerBinding(124, "y"),
				},
			},
			{
				Id: "node_2",
				Target: &core.Node_TaskNode{
					TaskNode: &core.TaskNode{Reference: &core.TaskNode_ReferenceId{ReferenceId: &core.Identifier{Name: "task_123"}}},
				},
				Inputs: []*core.Binding{
					newIntegerBinding(124, "y"), newVarBinding("node_1", "x", "x"),
				},
				OutputAliases: []*core.Alias{{Var: "x", Alias: "n2_output"}},
			},
		},
		Outputs: []*core.Binding{newVarBinding("node_2", "n2_output", "wf_output")},
	}

	// Detect what other graphs/tasks does this coreWorkflow reference
	subWorkflows := make([]*core.WorkflowTemplate, 0)
	reqs, err := GetRequirements(inputWorkflow, subWorkflows)
	if err != nil {
		fmt.Printf("Failed to get requirements. Error: %v", err)
		return
	}

	fmt.Printf("Needed Tasks: [%v], Needed Graphs [%v]\n",
		strings.Join(dumpIdentifierNames(reqs.GetRequiredTaskIds()), ","),
		strings.Join(dumpIdentifierNames(reqs.GetRequiredLaunchPlanIds()), ","))

	// Replace with logic to satisfy the requirements
	graphs := make([]common.InterfaceProvider, 0)
	inputTasks := []*core.TaskTemplate{
		{
			Id:       &core.Identifier{Name: "task_123"},
			Metadata: &core.TaskMetadata{},
			Interface: &core.TypedInterface{
				Inputs: createVariableMap(map[string]*core.Variable{
					"x": {
						Type: getIntegerLiteralType(),
					},
					"y": {
						Type: getIntegerLiteralType(),
					},
				}),
				Outputs: createVariableMap(map[string]*core.Variable{
					"x": {
						Type: getIntegerLiteralType(),
					},
				}),
			},
			Target: &core.TaskTemplate_Container{
				Container: &core.Container{
					Image:   "image://",
					Command: []string{"cmd"},
					Args:    []string{"args"},
				},
			},
		},
	}

	// Compile all tasks before proceeding with Workflow
	compiledTasks := make([]*core.CompiledTask, 0, len(inputTasks))
	for _, task := range inputTasks {
		compiledTask, err := CompileTask(task)
		if err != nil {
			fmt.Printf("Failed to compile task [%v]. Error: %v", task.Id, err)
			return
		}

		compiledTasks = append(compiledTasks, compiledTask)
	}

	output, errs := CompileWorkflow(inputWorkflow, subWorkflows, compiledTasks, graphs)
	if errs != nil {
		fmt.Printf("Compile Errors: %v\n", errs)
	} else {
		fmt.Printf("Compiled Workflow in GraphViz: %v\n", visualize.ToGraphViz(output.Primary))
	}

	// Output:
	// Needed Tasks: [task_123], Needed Graphs []
	// Compiled Workflow in GraphViz: digraph G {rankdir=TB;workflow[label="Workflow Id: name:"repo" "];node[style=filled];"start-node(start)" [shape=Msquare];"start-node(start)" -> "node_1()" [label="wf_input",style="solid"];"node_1()" -> "node_2()" [label="x",style="solid"];"static" -> "node_1()" [label=""];"node_2()" -> "end-node(end)" [label="n2_output",style="solid"];"static" -> "node_2()" [label=""];}
}

func ExampleCompileWorkflow_compileErrors() {
	inputWorkflow := &core.WorkflowTemplate{
		Id: &core.Identifier{Name: "repo"},
		Interface: &core.TypedInterface{
			Inputs:  createEmptyVariableMap(),
			Outputs: createEmptyVariableMap(),
		},
		Nodes: []*core.Node{
			{
				Target: &core.Node_TaskNode{
					TaskNode: &core.TaskNode{
						Reference: &core.TaskNode_ReferenceId{
							ReferenceId: &core.Identifier{Name: "task_123"},
						},
					},
				},
			},
		},
	}

	// Detect what other workflows/tasks does this coreWorkflow reference
	subWorkflows := make([]*core.WorkflowTemplate, 0)
	reqs, err := GetRequirements(inputWorkflow, subWorkflows)
	if err != nil {
		fmt.Printf("Failed to get requirements. Error: %v", err)
		return
	}

	fmt.Printf("Needed Tasks: [%v], Needed Workflows [%v]\n",
		strings.Join(dumpIdentifierNames(reqs.GetRequiredTaskIds()), ","),
		strings.Join(dumpIdentifierNames(reqs.GetRequiredLaunchPlanIds()), ","))

	// Replace with logic to satisfy the requirements
	workflows := make([]common.InterfaceProvider, 0)
	_, errs := CompileWorkflow(inputWorkflow, subWorkflows, []*core.CompiledTask{}, workflows)
	fmt.Printf("Compile Errors: %v\n", errs)

	// Output:
	// Needed Tasks: [task_123], Needed Workflows []
	// Compile Errors: Collected Errors: 1
	// 	Error 0: Code: TaskReferenceNotFound, Node Id: start-node, Description: Referenced Task [name:"task_123" ] not found.
}

func newIntegerPrimitive(value int64) *core.Primitive {
	return &core.Primitive{Value: &core.Primitive_Integer{Integer: value}}
}

func newStringPrimitive(value string) *core.Primitive {
	return &core.Primitive{Value: &core.Primitive_StringValue{StringValue: value}}
}

func newScalarInteger(value int64) *core.Scalar {
	return &core.Scalar{
		Value: &core.Scalar_Primitive{
			Primitive: newIntegerPrimitive(value),
		},
	}
}

func newIntegerLiteral(value int64) *core.Literal {
	return &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: newScalarInteger(value),
		},
	}
}

func getIntegerLiteralType() *core.LiteralType {
	return getSimpleLiteralType(core.SimpleType_INTEGER)
}

func getSimpleLiteralType(simpleType core.SimpleType) *core.LiteralType {
	return &core.LiteralType{
		Type: &core.LiteralType_Simple{
			Simple: simpleType,
		},
	}
}

func newIntegerBinding(value int64, toVar string) *core.Binding {
	return &core.Binding{
		Binding: &core.BindingData{
			Value: &core.BindingData_Scalar{Scalar: newIntegerLiteral(value).GetScalar()},
		},
		Var: toVar,
	}
}

func newVarBinding(fromNodeID, fromVar, toVar string) *core.Binding {
	return &core.Binding{
		Binding: &core.BindingData{
			Value: &core.BindingData_Promise{
				Promise: &core.OutputReference{
					NodeId: fromNodeID,
					Var:    fromVar,
				},
			},
		},
		Var: toVar,
	}
}

func TestComparisonExpression_MissingLeftRight(t *testing.T) {
	bExpr := &core.BooleanExpression{
		Expr: &core.BooleanExpression_Comparison{
			Comparison: &core.ComparisonExpression{
				Operator: core.ComparisonExpression_GT,
			},
		},
	}

	w := &mocks.WorkflowBuilder{}

	errs := errors.NewCompileErrors()
	v.ValidateBooleanExpression(w, &nodeBuilder{flyteNode: &flyteNode{}}, bExpr, true, errs)
	assert.Error(t, errs)
	assert.Equal(t, 2, errs.ErrorCount())
}

func TestComparisonExpression(t *testing.T) {
	bExpr := &core.BooleanExpression{
		Expr: &core.BooleanExpression_Comparison{
			Comparison: &core.ComparisonExpression{
				Operator:   core.ComparisonExpression_GT,
				LeftValue:  &core.Operand{Val: &core.Operand_Primitive{Primitive: newIntegerPrimitive(123)}},
				RightValue: &core.Operand{Val: &core.Operand_Primitive{Primitive: newStringPrimitive("hello")}},
			},
		},
	}

	w := &mocks.WorkflowBuilder{}
	errs := errors.NewCompileErrors()
	v.ValidateBooleanExpression(w, &nodeBuilder{flyteNode: &flyteNode{}}, bExpr, true, errs)
	assert.True(t, errs.HasErrors())
	assert.Equal(t, 1, errs.ErrorCount())
}

func TestBooleanExpression_BranchNodeHasNoCondition(t *testing.T) {
	bExpr := &core.BooleanExpression{
		Expr: &core.BooleanExpression_Conjunction{
			Conjunction: &core.ConjunctionExpression{
				Operator: core.ConjunctionExpression_AND,
				RightExpression: &core.BooleanExpression{
					Expr: &core.BooleanExpression_Comparison{
						Comparison: &core.ComparisonExpression{
							Operator:   core.ComparisonExpression_GT,
							LeftValue:  &core.Operand{Val: &core.Operand_Primitive{Primitive: newIntegerPrimitive(123)}},
							RightValue: &core.Operand{Val: &core.Operand_Primitive{Primitive: newIntegerPrimitive(345)}},
						},
					},
				},
			},
		},
	}

	w := &mocks.WorkflowBuilder{}
	errs := errors.NewCompileErrors()
	v.ValidateBooleanExpression(w, &nodeBuilder{flyteNode: &flyteNode{}}, bExpr, true, errs)
	assert.True(t, errs.HasErrors())
	assert.Equal(t, 1, errs.ErrorCount())
	for e := range *errs.Errors() {
		assert.Equal(t, errors.BranchNodeHasNoCondition, e.Code())
	}
}

func newNodeIDSet(nodeIDs ...common.NodeID) sets.String {
	return sets.NewString(nodeIDs...)
}

func TestValidateReachable(t *testing.T) {
	graph := &workflowBuilder{
		NodeBuilderIndex: common.NewNodeIndex(),
	}

	graph.downstreamNodes = map[string]sets.String{
		v1alpha1.StartNodeID: newNodeIDSet("1"),
		"1":                  newNodeIDSet("5", "2"),
		"2":                  newNodeIDSet("3"),
		"3":                  newNodeIDSet("4"),
		"4":                  newNodeIDSet(v1alpha1.EndNodeID),
	}

	for range graph.downstreamNodes {
		graph.Nodes = common.NewNodeIndex(graph.GetOrCreateNodeBuilder(nil))
	}

	errs := errors.NewCompileErrors()
	assert.False(t, graph.validateReachable(errs))
	assert.True(t, errs.HasErrors())
}

func TestValidateUnderlyingInterface(parentT *testing.T) {
	graphIface := &core.TypedInterface{
		Inputs: createVariableMap(map[string]*core.Variable{
			"x": {
				Type: getIntegerLiteralType(),
			},
		}),
		Outputs: createVariableMap(map[string]*core.Variable{
			"x": {
				Type: getIntegerLiteralType(),
			},
		}),
	}

	inputWorkflow := &core.WorkflowTemplate{
		Id:        &core.Identifier{Name: "repo"},
		Interface: graphIface,
		Nodes: []*core.Node{
			{
				Id: "node_123",
				Target: &core.Node_TaskNode{
					TaskNode: &core.TaskNode{Reference: &core.TaskNode_ReferenceId{ReferenceId: &core.Identifier{Name: "task_123"}}},
				},
			},
		},
	}

	taskIface := &core.TypedInterface{
		Inputs: createVariableMap(map[string]*core.Variable{
			"x": {
				Type: getIntegerLiteralType(),
			},
			"y": {
				Type: getIntegerLiteralType(),
			},
		}),
		Outputs: createVariableMap(map[string]*core.Variable{
			"x": {
				Type: getIntegerLiteralType(),
			},
		}),
	}

	inputTasks := []*core.TaskTemplate{
		{
			Id:        &core.Identifier{Name: "task_123"},
			Metadata:  &core.TaskMetadata{},
			Interface: taskIface,
			Target: &core.TaskTemplate_Container{
				Container: &core.Container{
					Image:   "Image://",
					Command: []string{"blah"},
					Args:    []string{"bloh"},
				},
			},
		},
	}

	errs := errors.NewCompileErrors()
	compiledTasks := make([]common.Task, 0, len(inputTasks))
	for _, inputTask := range inputTasks {
		t := compileTaskInternal(inputTask, errs)
		compiledTasks = append(compiledTasks, t)
		assert.False(parentT, errs.HasErrors())
		if errs.HasErrors() {
			assert.FailNow(parentT, errs.Error())
		}
	}

	g := newWorkflowBuilder(
		&core.CompiledWorkflow{Template: inputWorkflow},
		mustBuildWorkflowIndex(inputWorkflow),
		common.NewTaskIndex(compiledTasks...),
		map[string]common.InterfaceProvider{})
	(&g).Tasks = common.NewTaskIndex(compiledTasks...)

	parentT.Run("TaskNode", func(t *testing.T) {
		errs := errors.NewCompileErrors()
		iface, ifaceOk := v.ValidateUnderlyingInterface(&g, &nodeBuilder{flyteNode: inputWorkflow.Nodes[0]}, errs)
		assert.True(t, ifaceOk)
		assert.False(t, errs.HasErrors())
		assert.Equal(t, taskIface, iface)
	})

	parentT.Run("GraphNode", func(t *testing.T) {
		errs := errors.NewCompileErrors()
		iface, ifaceOk := v.ValidateUnderlyingInterface(&g, &nodeBuilder{flyteNode: &core.Node{
			Target: &core.Node_WorkflowNode{
				WorkflowNode: &core.WorkflowNode{
					Reference: &core.WorkflowNode_SubWorkflowRef{
						SubWorkflowRef: inputWorkflow.Id,
					},
				},
			},
		}}, errs)
		assert.True(t, ifaceOk)
		assert.False(t, errs.HasErrors())
		assert.Equal(t, graphIface, iface)
	})

	parentT.Run("BranchNode", func(branchT *testing.T) {
		branchT.Run("OneCase", func(t *testing.T) {
			errs := errors.NewCompileErrors()
			iface, ifaceOk := v.ValidateUnderlyingInterface(&g, &nodeBuilder{flyteNode: &core.Node{
				Target: &core.Node_BranchNode{
					BranchNode: &core.BranchNode{
						IfElse: &core.IfElseBlock{
							Case: &core.IfBlock{
								ThenNode: inputWorkflow.Nodes[0],
							},
						},
					},
				},
			}}, errs)
			assert.True(t, ifaceOk)
			assert.False(t, errs.HasErrors())
			assert.Equal(t, taskIface.Outputs, iface.Outputs)
		})

		branchT.Run("TwoCases", func(t *testing.T) {
			errs := errors.NewCompileErrors()
			_, ifaceOk := v.ValidateUnderlyingInterface(&g, &nodeBuilder{flyteNode: &core.Node{
				Target: &core.Node_BranchNode{
					BranchNode: &core.BranchNode{
						IfElse: &core.IfElseBlock{
							Case: &core.IfBlock{
								ThenNode: inputWorkflow.Nodes[0],
							},
							Other: []*core.IfBlock{
								{
									ThenNode: &core.Node{
										Target: &core.Node_WorkflowNode{
											WorkflowNode: &core.WorkflowNode{
												Reference: &core.WorkflowNode_SubWorkflowRef{
													SubWorkflowRef: inputWorkflow.Id,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}}, errs)
			assert.True(t, ifaceOk)
			assert.False(t, errs.HasErrors())
		})
	})
}

func TestCompileWorkflow(t *testing.T) {
	inputWorkflow := &core.WorkflowTemplate{
		Id: &core.Identifier{Name: "repo"},
		Interface: &core.TypedInterface{
			Inputs: createVariableMap(map[string]*core.Variable{
				"x": {
					Type: getIntegerLiteralType(),
				},
			}),
			Outputs: createVariableMap(map[string]*core.Variable{
				"x": {
					Type: getIntegerLiteralType(),
				},
			}),
		},
		Nodes: []*core.Node{
			{
				Id: "node_123",
				Target: &core.Node_TaskNode{
					TaskNode: &core.TaskNode{Reference: &core.TaskNode_ReferenceId{ReferenceId: &core.Identifier{Name: "task_123"}}},
				},
				Inputs: []*core.Binding{
					newIntegerBinding(123, "x"), newIntegerBinding(123, "y"),
				},
			},
			{
				Id: "node_456",
				Target: &core.Node_TaskNode{
					TaskNode: &core.TaskNode{Reference: &core.TaskNode_ReferenceId{ReferenceId: &core.Identifier{Name: "task_123"}}},
				},
				Inputs: []*core.Binding{
					newIntegerBinding(123, "y"), newVarBinding("node_123", "x", "x"),
				},
				UpstreamNodeIds: []string{"node_123"},
			},
		},
		Outputs: []*core.Binding{newVarBinding("node_456", "x", "x")},
	}

	inputTasks := []*core.TaskTemplate{
		{
			Id: &core.Identifier{Name: "task_123"}, Metadata: &core.TaskMetadata{},
			Interface: &core.TypedInterface{
				Inputs: createVariableMap(map[string]*core.Variable{
					"x": {
						Type: getIntegerLiteralType(),
					},
					"y": {
						Type: getIntegerLiteralType(),
					},
				}),
				Outputs: createVariableMap(map[string]*core.Variable{
					"x": {
						Type: getIntegerLiteralType(),
					},
				}),
			},
			Target: &core.TaskTemplate_Container{
				Container: &core.Container{
					Command: []string{},
					Image:   "image://123",
				},
			},
		},
	}

	errors.SetConfig(errors.Config{PanicOnError: true})
	defer errors.SetConfig(errors.Config{})
	output, errs := CompileWorkflow(inputWorkflow, []*core.WorkflowTemplate{}, mustCompileTasks(inputTasks), []common.InterfaceProvider{})
	assert.NoError(t, errs)
	assert.NotNil(t, output)
	if output != nil {
		t.Logf("Graph Repr: %v", visualize.ToGraphViz(output.Primary))

		assert.Equal(t, []string{"node_123"}, output.Primary.Connections.Upstream["node_456"].Ids)
	}
}

func TestNoNodesFound(t *testing.T) {
	inputWorkflow := &core.WorkflowTemplate{
		Id: &core.Identifier{Name: "repo"},
		Interface: &core.TypedInterface{
			Inputs: createVariableMap(map[string]*core.Variable{
				"x": {
					Type: getIntegerLiteralType(),
				},
			}),
			Outputs: createVariableMap(map[string]*core.Variable{
				"x": {
					Type: getIntegerLiteralType(),
				},
			}),
		},
		Nodes:   []*core.Node{},
		Outputs: []*core.Binding{newVarBinding("node_456", "x", "x")},
	}

	_, errs := CompileWorkflow(inputWorkflow, []*core.WorkflowTemplate{},
		mustCompileTasks(make([]*core.TaskTemplate, 0)), []common.InterfaceProvider{})
	assert.Contains(t, errs.Error(), errors.NoNodesFound)
}

func mustCompileTasks(tasks []*core.TaskTemplate) []*core.CompiledTask {
	res := make([]*core.CompiledTask, 0, len(tasks))
	for _, t := range tasks {
		compiledT, err := CompileTask(t)
		if err != nil {
			panic(err)
		}

		res = append(res, compiledT)
	}
	return res
}

func mustBuildWorkflowIndex(wfs ...*core.WorkflowTemplate) common.WorkflowIndex {
	compiledWfs := make([]*core.CompiledWorkflow, 0, len(wfs))
	for _, wf := range wfs {
		compiledWfs = append(compiledWfs, &core.CompiledWorkflow{Template: wf})
	}

	err := errors.NewCompileErrors()
	if index, ok := common.NewWorkflowIndex(compiledWfs, err); !ok {
		panic(err)
	} else {
		return index
	}
}
