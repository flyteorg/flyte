package visualize

import (
	"fmt"
	"io/ioutil"
	"testing"

	graphviz "github.com/awalterschulze/gographviz"
	"github.com/flyteorg/flyte/flytectl/pkg/visualize/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRenderWorkflowBranch(t *testing.T) {
	// Sadly we cannot compare the output of svg, as it slightly changes.
	file := []string{"compiled_closure_branch_nested", "compiled_subworkflows"}

	for _, s := range file {
		t.Run(s, func(t *testing.T) {
			r, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s.json", s))
			assert.NoError(t, err)

			c := &core.CompiledWorkflowClosure{}
			err = utils.UnmarshalBytesToPb(r, c)
			assert.NoError(t, err)
			b, err := RenderWorkflow(c)
			fmt.Println(b)
			assert.NoError(t, err)
			assert.NotNil(t, b)
		})
	}
}

func TestAddBranchSubNodeEdge(t *testing.T) {
	attrs := map[string]string{}
	attrs[LHeadAttr] = fmt.Sprintf("\"%s\"", "innerGraph")
	attrs[LabelAttr] = fmt.Sprintf("\"%s\"", "label")
	t.Run("Successful", func(t *testing.T) {
		gb := newGraphBuilder()
		gb.nodeClusters["n"] = "innerGraph"
		parentNode := &graphviz.Node{Name: "parentNode", Attrs: nil}
		n := &graphviz.Node{Name: "n"}

		mockGraph := &mocks.Graphvizer{}
		// Verify the attributes
		mockGraph.OnAddEdgeMatch(mock.Anything, mock.Anything, mock.Anything, attrs).Return(nil)
		mockGraph.OnGetEdgeMatch(mock.Anything, mock.Anything).Return(&graphviz.Edge{})
		err := gb.addBranchSubNodeEdge(mockGraph, parentNode, n, "label")
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		gb := newGraphBuilder()
		gb.nodeClusters["n"] = "innerGraph"
		parentNode := &graphviz.Node{Name: "parentNode", Attrs: nil}
		n := &graphviz.Node{Name: "n"}

		mockGraph := &mocks.Graphvizer{}
		// Verify the attributes
		mockGraph.OnAddEdgeMatch(mock.Anything, mock.Anything, mock.Anything, attrs).Return(fmt.Errorf("error adding edge"))
		err := gb.addBranchSubNodeEdge(mockGraph, parentNode, n, "label")
		assert.NotNil(t, err)
	})
}

func TestConstructBranchNode(t *testing.T) {
	attrs := map[string]string{}
	attrs[LabelAttr] = fmt.Sprintf("\"[%s]\"", "nodeMetadata")
	attrs[ShapeType] = DiamondShape
	t.Run("Successful", func(t *testing.T) {
		gb := newGraphBuilder()
		mockGraph := &mocks.Graphvizer{}
		expectedGraphvizNode := &graphviz.Node{
			Name: "brancheName_id",
			Attrs: map[graphviz.Attr]string{"label": fmt.Sprintf("\"[%s]\"", "nodeMetadata"),
				"shape": "diamond"},
		}
		// Verify the attributes
		mockGraph.OnAddNodeMatch(mock.Anything, mock.Anything, attrs).Return(nil)
		mockGraph.OnGetNodeMatch(mock.Anything).Return(expectedGraphvizNode)
		flyteNode := &core.Node{
			Id: "id",
			Metadata: &core.NodeMetadata{
				Name: "nodeMetadata",
			},
			Target: &core.Node_BranchNode{
				BranchNode: &core.BranchNode{},
			},
		}
		resultBranchNode, err := gb.constructBranchNode("parentGraph", "branchName", mockGraph, flyteNode)
		assert.NoError(t, err)
		assert.NotNil(t, resultBranchNode)
		assert.Equal(t, expectedGraphvizNode, resultBranchNode)
	})

	t.Run("Add Node Error", func(t *testing.T) {
		gb := newGraphBuilder()
		mockGraph := &mocks.Graphvizer{}
		// Verify the attributes
		mockGraph.OnAddNodeMatch(mock.Anything, mock.Anything, attrs).Return(fmt.Errorf("unable to add node"))
		mockGraph.OnGetNodeMatch(mock.Anything).Return(nil)
		flyteNode := &core.Node{
			Id: "id",
			Metadata: &core.NodeMetadata{
				Name: "nodeMetadata",
			},
			Target: &core.Node_BranchNode{
				BranchNode: &core.BranchNode{},
			},
		}
		resultBranchNode, err := gb.constructBranchNode("parentGraph", "branchName", mockGraph, flyteNode)
		assert.NotNil(t, err)
		assert.Nil(t, resultBranchNode)
	})

	t.Run("Add ThenNode Error", func(t *testing.T) {
		gb := newGraphBuilder()
		mockGraph := &mocks.Graphvizer{}
		attrs := map[string]string{}
		attrs[LabelAttr] = fmt.Sprintf("\"[%s]\"", "nodeMetadata")
		attrs[ShapeType] = DiamondShape

		// Verify the attributes
		mockGraph.OnAddNodeMatch(mock.Anything, "branchName_id", attrs).Return(nil)
		mockGraph.OnAddNodeMatch(mock.Anything, "branchName_start_node", mock.Anything).Return(fmt.Errorf("unable to add node"))
		mockGraph.OnGetNodeMatch(mock.Anything).Return(nil)
		flyteNode := &core.Node{
			Id: "id",
			Metadata: &core.NodeMetadata{
				Name: "nodeMetadata",
			},
			Target: &core.Node_BranchNode{
				BranchNode: &core.BranchNode{
					IfElse: &core.IfElseBlock{
						Case: &core.IfBlock{
							Condition: &core.BooleanExpression{},
							ThenNode: &core.Node{
								Id: "start-node",
							},
						},
					},
				},
			},
		}
		resultBranchNode, err := gb.constructBranchNode("parentGraph", "branchName", mockGraph, flyteNode)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("unable to add node"), err)
		assert.Nil(t, resultBranchNode)
	})

	t.Run("Add Condition Node Edge Error", func(t *testing.T) {
		gb := newGraphBuilder()
		mockGraph := &mocks.Graphvizer{}
		attrs := map[string]string{}
		attrs[LabelAttr] = fmt.Sprintf("\"[%s]\"", "nodeMetadata")
		attrs[ShapeType] = DiamondShape

		parentNode := &graphviz.Node{Name: "parentNode", Attrs: nil}
		thenBranchStartNode := &graphviz.Node{Name: "branchName_start_node", Attrs: nil}

		mockGraph.OnAddNodeMatch(mock.Anything, "branchName_id", attrs).Return(nil)
		mockGraph.OnAddNodeMatch(mock.Anything, "branchName_start_node", mock.Anything).Return(nil)
		mockGraph.OnGetNodeMatch("branchName_id").Return(parentNode)
		mockGraph.OnGetNodeMatch("branchName_start_node").Return(thenBranchStartNode)
		mockGraph.OnAddEdgeMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("unable to add edge"))
		flyteNode := &core.Node{
			Id: "id",
			Metadata: &core.NodeMetadata{
				Name: "nodeMetadata",
			},
			Target: &core.Node_BranchNode{
				BranchNode: &core.BranchNode{
					IfElse: &core.IfElseBlock{
						Case: &core.IfBlock{
							Condition: &core.BooleanExpression{
								Expr: &core.BooleanExpression_Comparison{
									Comparison: &core.ComparisonExpression{
										Operator: core.ComparisonExpression_EQ,
										LeftValue: &core.Operand{
											Val: &core.Operand_Primitive{
												Primitive: &core.Primitive{
													Value: &core.Primitive_Integer{
														Integer: 40,
													},
												},
											},
										},
										RightValue: &core.Operand{
											Val: &core.Operand_Primitive{
												Primitive: &core.Primitive{
													Value: &core.Primitive_Integer{
														Integer: 50,
													},
												},
											},
										},
									},
								},
							},
							ThenNode: &core.Node{
								Id: "start-node",
							},
						},
					},
				},
			},
		}
		resultBranchNode, err := gb.constructBranchNode("parentGraph", "branchName", mockGraph, flyteNode)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("unable to add edge"), err)
		assert.Nil(t, resultBranchNode)
	})
}

func TestConstructNode(t *testing.T) {

	t.Run("Start-Node", func(t *testing.T) {
		attrs := map[string]string{}
		attrs[LabelAttr] = "start"
		attrs[ShapeType] = DoubleCircleShape
		attrs[ColorAttr] = Green
		gb := newGraphBuilder()
		mockGraph := &mocks.Graphvizer{}
		expectedGraphvizNode := &graphviz.Node{
			Name:  "start-node",
			Attrs: map[graphviz.Attr]string{"label": "start", "shape": "doublecircle", "color": "green"},
		}
		// Verify the attributes
		mockGraph.OnAddNodeMatch(mock.Anything, mock.Anything, attrs).Return(nil)
		mockGraph.OnGetNodeMatch(mock.Anything).Return(expectedGraphvizNode)
		flyteNode := &core.Node{
			Id: "start-node",
		}
		resultNode, err := gb.constructNode("", "", mockGraph, flyteNode)
		assert.NoError(t, err)
		assert.NotNil(t, resultNode)
		assert.Equal(t, expectedGraphvizNode, resultNode)
	})

	t.Run("End-Node", func(t *testing.T) {
		attrs := map[string]string{}
		attrs[LabelAttr] = "end"
		attrs[ShapeType] = DoubleCircleShape
		attrs[ColorAttr] = Red
		gb := newGraphBuilder()
		mockGraph := &mocks.Graphvizer{}
		expectedGraphvizNode := &graphviz.Node{
			Name:  "end-node",
			Attrs: map[graphviz.Attr]string{"label": "end", "shape": "doublecircle", "color": "red"},
		}
		// Verify the attributes
		mockGraph.OnAddNodeMatch(mock.Anything, mock.Anything, attrs).Return(nil)
		mockGraph.OnGetNodeMatch(mock.Anything).Return(expectedGraphvizNode)
		flyteNode := &core.Node{
			Id: "end-node",
		}
		resultNode, err := gb.constructNode("", "", mockGraph, flyteNode)
		assert.NoError(t, err)
		assert.NotNil(t, resultNode)
		assert.Equal(t, expectedGraphvizNode, resultNode)
	})

	t.Run("Task-Node-Error", func(t *testing.T) {
		gb := newGraphBuilder()
		mockGraph := &mocks.Graphvizer{}
		flyteNode := &core.Node{
			Id: "id",
			Metadata: &core.NodeMetadata{
				Name: "nodeMetadata",
			},
			Target: &core.Node_TaskNode{
				TaskNode: &core.TaskNode{
					Reference: &core.TaskNode_ReferenceId{
						ReferenceId: &core.Identifier{
							Project: "dummyProject",
							Domain:  "dummyDomain",
							Name:    "dummyName",
							Version: "dummyVersion",
						},
					},
				},
			},
		}
		resultNode, err := gb.constructNode("", "", mockGraph, flyteNode)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to find task [%s] in closure",
			flyteNode.GetTaskNode().GetReferenceId().String()), err)
		assert.Nil(t, resultNode)
	})

	t.Run("Branch-Node-SubGraph-Error", func(t *testing.T) {
		gb := newGraphBuilder()
		mockGraph := &mocks.Graphvizer{}
		// Verify the attributes
		mockGraph.OnAddSubGraphMatch("", SubgraphPrefix+"nodeMetadata",
			mock.Anything).Return(fmt.Errorf("unable to create subgraph"))
		flyteNode := &core.Node{
			Id: "id",
			Metadata: &core.NodeMetadata{
				Name: "nodeMetadata",
			},
			Target: &core.Node_BranchNode{
				BranchNode: &core.BranchNode{},
			},
		}
		resultBranchNode, err := gb.constructNode("", "", mockGraph, flyteNode)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("unable to create subgraph"), err)
		assert.Nil(t, resultBranchNode)
	})

	t.Run("Branch-Node-Add-Error", func(t *testing.T) {
		attrs := map[string]string{}
		attrs[LabelAttr] = fmt.Sprintf("\"[%s]\"", "nodeMetadata")
		attrs[ShapeType] = DiamondShape
		gb := newGraphBuilder()
		mockGraph := &mocks.Graphvizer{}

		// Verify the attributes
		mockGraph.OnAddSubGraphMatch(mock.Anything, SubgraphPrefix+"nodeMetadata", mock.Anything).Return(nil)
		mockGraph.OnAddNodeMatch(mock.Anything, mock.Anything, attrs).Return(fmt.Errorf("unable to add node"))
		mockGraph.OnGetNodeMatch(mock.Anything).Return(nil)

		flyteNode := &core.Node{
			Id: "id",
			Metadata: &core.NodeMetadata{
				Name: "nodeMetadata",
			},
			Target: &core.Node_BranchNode{
				BranchNode: &core.BranchNode{},
			},
		}
		resultBranchNode, err := gb.constructNode("", "", mockGraph, flyteNode)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("unable to add node"), err)
		assert.Nil(t, resultBranchNode)
	})

	t.Run("Workflow-Node-Add-Error", func(t *testing.T) {
		attrs := map[string]string{}
		gb := newGraphBuilder()
		mockGraph := &mocks.Graphvizer{}

		// Verify the attributes
		mockGraph.OnAddNodeMatch(mock.Anything, mock.Anything, attrs).Return(fmt.Errorf("unable to add node"))
		mockGraph.OnGetNodeMatch(mock.Anything).Return(nil)

		flyteNode := &core.Node{
			Id: "id",
			Metadata: &core.NodeMetadata{
				Name: "nodeMetadata",
			},
			Target: &core.Node_WorkflowNode{
				WorkflowNode: &core.WorkflowNode{
					Reference: &core.WorkflowNode_LaunchplanRef{
						LaunchplanRef: &core.Identifier{
							Project: "dummyProject",
							Domain:  "dummyDomain",
							Name:    "dummyName",
							Version: "dummyVersion",
						},
					},
				},
			},
		}
		resultWorkflowNode, err := gb.constructNode("", "", mockGraph, flyteNode)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("unable to add node"), err)
		assert.Nil(t, resultWorkflowNode)
	})

	t.Run("Workflow-Node-SubGraph-Error", func(t *testing.T) {
		gb := newGraphBuilder()
		mockGraph := &mocks.Graphvizer{}
		// Verify the attributes
		mockGraph.OnAddSubGraphMatch("", SubgraphPrefix+"id",
			mock.Anything).Return(fmt.Errorf("unable to create subgraph"))
		flyteNode := &core.Node{
			Id: "id",
			Metadata: &core.NodeMetadata{
				Name: "nodeMetadata",
			},
			Target: &core.Node_WorkflowNode{
				WorkflowNode: &core.WorkflowNode{},
			},
		}
		resultWorkflowNode, err := gb.constructNode("", "", mockGraph, flyteNode)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("unable to create subgraph"), err)
		assert.Nil(t, resultWorkflowNode)
	})
	t.Run("Workflow-Node-Subworkflow-NotFound-Error", func(t *testing.T) {
		gb := newGraphBuilder()
		mockGraph := &mocks.Graphvizer{}

		// Verify the attributes
		mockGraph.OnAddSubGraphMatch(mock.Anything, SubgraphPrefix+"id", mock.Anything).Return(nil)

		flyteNode := &core.Node{
			Id: "id",
			Metadata: &core.NodeMetadata{
				Name: "nodeMetadata",
			},
			Target: &core.Node_WorkflowNode{
				WorkflowNode: &core.WorkflowNode{
					Reference: &core.WorkflowNode_SubWorkflowRef{
						SubWorkflowRef: &core.Identifier{
							Project: "dummyProject",
							Domain:  "dummyDomain",
							Name:    "dummyName",
							Version: "dummyVersion",
						},
					},
				},
			},
		}
		resultWorkflowNode, err := gb.constructNode("", "", mockGraph, flyteNode)
		assert.NotNil(t, err)
		utils.AssertEqualWithSanitizedRegex(t, "subworkfow [project:\"dummyProject\" domain:\"dummyDomain\" name:\"dummyName\" version:\"dummyVersion\"] not found", err.Error())
		assert.Nil(t, resultWorkflowNode)
	})

	t.Run("Workflow-Node-Subworkflow-Graph-Create-Error", func(t *testing.T) {
		gb := newGraphBuilder()
		mockGraph := &mocks.Graphvizer{}

		// Verify the attributes
		mockGraph.OnAddSubGraphMatch(mock.Anything, SubgraphPrefix+"id", mock.Anything).Return(nil)
		mockGraph.OnAddNodeMatch(mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("unable to add node"))
		mockGraph.OnGetNodeMatch(mock.Anything).Return(nil)

		gb.subWf = make(map[string]*core.CompiledWorkflow)
		subwfNode := &core.Node{
			Id: "start-node",
		}
		sbwfNodes := []*core.Node{subwfNode}

		flyteNode := &core.Node{
			Id: "id",
			Metadata: &core.NodeMetadata{
				Name: "nodeMetadata",
			},
			Target: &core.Node_WorkflowNode{
				WorkflowNode: &core.WorkflowNode{
					Reference: &core.WorkflowNode_SubWorkflowRef{
						SubWorkflowRef: &core.Identifier{
							Project: "dummyProject",
							Domain:  "dummyDomain",
							Name:    "dummyName",
							Version: "dummyVersion",
						},
					},
				},
			},
		}
		// Since this code depends on the string representation of a protobuf message, we do not
		// use a fix string to compare the error message.
		gb.subWf[flyteNode.GetWorkflowNode().GetSubWorkflowRef().String()] =
			&core.CompiledWorkflow{Template: &core.WorkflowTemplate{Nodes: sbwfNodes}}
		resultWorkflowNode, err := gb.constructNode("", "", mockGraph, flyteNode)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("unable to add node"), err)
		assert.Nil(t, resultWorkflowNode)
	})

}
