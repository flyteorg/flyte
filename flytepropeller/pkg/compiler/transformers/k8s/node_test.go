package k8s

import (
	"testing"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/compiler/common"
	"github.com/flyteorg/flytepropeller/pkg/compiler/errors"
	"github.com/stretchr/testify/assert"
)

func createNodeWithTask() *core.Node {
	return &core.Node{
		Id: "n_1",
		Target: &core.Node_TaskNode{
			TaskNode: &core.TaskNode{
				Reference: &core.TaskNode_ReferenceId{
					ReferenceId: &core.Identifier{Name: "ref_1"},
				},
			},
		},
		Metadata: &core.NodeMetadata{InterruptibleValue: &core.NodeMetadata_Interruptible{Interruptible: true}},
	}
}

func TestBuildNodeSpec(t *testing.T) {
	n := mockNode{
		id:   "n_1",
		Node: &core.Node{},
	}

	tasks := []*core.CompiledTask{
		{
			Template: &core.TaskTemplate{
				Id: &core.Identifier{Name: "ref_1"},
			},
		},
		{
			Template: &core.TaskTemplate{
				Id: &core.Identifier{Name: "ref_2"},
				Target: &core.TaskTemplate_Container{
					Container: &core.Container{
						Resources: &core.Resources{
							Requests: []*core.Resources_ResourceEntry{
								{
									Name:  core.Resources_CPU,
									Value: "10Mi",
								},
							},
						},
					},
				},
			},
		},
	}

	errors.SetConfig(errors.Config{IncludeSource: true})
	errs := errors.NewCompileErrors()

	mustBuild := func(t testing.TB, n common.Node, expectedInnerNodesCount int, errs errors.CompileErrors) *v1alpha1.NodeSpec {
		specs, ok := buildNodeSpec(n.GetCoreNode(), tasks, errs)
		assert.Len(t, specs, expectedInnerNodesCount)
		spec := specs[0]
		assert.Nil(t, spec.Interruptibe)
		assert.False(t, errs.HasErrors())
		assert.True(t, ok)
		assert.NotNil(t, spec)

		if errs.HasErrors() {
			assert.Fail(t, errs.Error())
		}

		return spec
	}

	t.Run("Task", func(t *testing.T) {
		n.Node.Target = &core.Node_TaskNode{
			TaskNode: &core.TaskNode{
				Reference: &core.TaskNode_ReferenceId{
					ReferenceId: &core.Identifier{Name: "ref_1"},
				},
			},
		}

		mustBuild(t, n, 1, errs.NewScope())
	})

	t.Run("Task with resources", func(t *testing.T) {
		expectedCPU := resource.MustParse("10Mi")
		n.Node.Target = &core.Node_TaskNode{
			TaskNode: &core.TaskNode{
				Reference: &core.TaskNode_ReferenceId{
					ReferenceId: &core.Identifier{Name: "ref_2"},
				},
			},
		}

		spec := mustBuild(t, n, 1, errs.NewScope())
		assert.NotNil(t, spec.Resources)
		assert.NotNil(t, spec.Resources.Requests.Cpu())
		assert.Equal(t, expectedCPU.Value(), spec.Resources.Requests.Cpu().Value())
	})

	t.Run("node with resource overrides", func(t *testing.T) {
		expectedCPU := resource.MustParse("20Mi")
		n.Node.Target = &core.Node_TaskNode{
			TaskNode: &core.TaskNode{
				Reference: &core.TaskNode_ReferenceId{
					ReferenceId: &core.Identifier{Name: "ref_2"},
				},
				Overrides: &core.TaskNodeOverrides{
					Resources: &core.Resources{
						Requests: []*core.Resources_ResourceEntry{
							{
								Name:  core.Resources_CPU,
								Value: "20Mi",
							},
						},
					},
				},
			},
		}

		spec := mustBuild(t, n, 1, errs.NewScope())
		assert.NotNil(t, spec.Resources)
		assert.NotNil(t, spec.Resources.Requests.Cpu())
		assert.Equal(t, expectedCPU.Value(), spec.Resources.Requests.Cpu().Value())
	})

	t.Run("LaunchPlanRef", func(t *testing.T) {
		n.Node.Target = &core.Node_WorkflowNode{
			WorkflowNode: &core.WorkflowNode{
				Reference: &core.WorkflowNode_LaunchplanRef{
					LaunchplanRef: &core.Identifier{Name: "ref_1"},
				},
			},
		}

		mustBuild(t, n, 1, errs.NewScope())
	})

	t.Run("Workflow", func(t *testing.T) {
		n.subWF = createSampleMockWorkflow()
		n.Node.Target = &core.Node_WorkflowNode{
			WorkflowNode: &core.WorkflowNode{
				Reference: &core.WorkflowNode_SubWorkflowRef{
					SubWorkflowRef: n.subWF.GetCoreWorkflow().Template.Id,
				},
			},
		}

		mustBuild(t, n, 1, errs.NewScope())
	})

	t.Run("Branch", func(t *testing.T) {
		n.Node.Target = &core.Node_BranchNode{
			BranchNode: &core.BranchNode{
				IfElse: &core.IfElseBlock{
					Other: []*core.IfBlock{},
					Default: &core.IfElseBlock_Error{
						Error: &core.Error{
							Message: "failed",
						},
					},
					Case: &core.IfBlock{
						ThenNode: &core.Node{
							Target: &core.Node_TaskNode{
								TaskNode: &core.TaskNode{
									Reference: &core.TaskNode_ReferenceId{
										ReferenceId: &core.Identifier{Name: "ref_1"},
									},
								},
							},
						},
						Condition: &core.BooleanExpression{
							Expr: &core.BooleanExpression_Comparison{
								Comparison: &core.ComparisonExpression{
									Operator: core.ComparisonExpression_EQ,
									LeftValue: &core.Operand{
										Val: &core.Operand_Primitive{
											Primitive: &core.Primitive{
												Value: &core.Primitive_Integer{
													Integer: 123,
												},
											},
										},
									},
									RightValue: &core.Operand{
										Val: &core.Operand_Primitive{
											Primitive: &core.Primitive{
												Value: &core.Primitive_Integer{
													Integer: 123,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		mustBuild(t, n, 2, errs.NewScope())
	})

}
