package k8s

import (
	"testing"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/compiler/common"
	"github.com/flyteorg/flytepropeller/pkg/compiler/errors"

	"github.com/stretchr/testify/assert"

	"google.golang.org/protobuf/types/known/durationpb"

	"k8s.io/apimachinery/pkg/api/resource"
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
		assert.Nil(t, spec.Interruptible)
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

	t.Run("GateNodeApprove", func(t *testing.T) {
		n.Node.Target = &core.Node_GateNode{
			GateNode: &core.GateNode{
				Condition: &core.GateNode_Approve{
					Approve: &core.ApproveCondition{
						SignalId: "foo",
					},
				},
			},
		}

		mustBuild(t, n, 1, errs.NewScope())
	})

	t.Run("GateNodeSignal", func(t *testing.T) {
		n.Node.Target = &core.Node_GateNode{
			GateNode: &core.GateNode{
				Condition: &core.GateNode_Signal{
					Signal: &core.SignalCondition{
						SignalId: "foo",
						Type: &core.LiteralType{
							Type: &core.LiteralType_Simple{
								Simple: core.SimpleType_BOOLEAN,
							},
						},
					},
				},
			},
		}

		mustBuild(t, n, 1, errs.NewScope())
	})

	t.Run("GateNodeSleep", func(t *testing.T) {
		n.Node.Target = &core.Node_GateNode{
			GateNode: &core.GateNode{
				Condition: &core.GateNode_Sleep{
					Sleep: &core.SleepCondition{
						Duration: durationpb.New(time.Minute),
					},
				},
			},
		}

		mustBuild(t, n, 1, errs.NewScope())
	})

	t.Run("ArrayNode", func(t *testing.T) {
		n.Node.Target = &core.Node_ArrayNode{
			ArrayNode: &core.ArrayNode{
				Node: &core.Node{
					Id: "foo",
					Target: &core.Node_TaskNode{
						TaskNode: &core.TaskNode{
							Reference: &core.TaskNode_ReferenceId{
								ReferenceId: &core.Identifier{Name: "ref_1"},
							},
						},
					},
				},
				Parallelism: 10,
				SuccessCriteria: &core.ArrayNode_MinSuccessRatio{
					MinSuccessRatio: 0.5,
				},
			},
		}

		mustBuild(t, n, 1, errs.NewScope())
	})
}

func TestBuildTasks(t *testing.T) {

	withoutAnnotations := make(map[string]*core.Variable)
	withoutAnnotations["a"] = &core.Variable{
		Type: &core.LiteralType{},
	}

	tasks := []*core.CompiledTask{
		{
			Template: &core.TaskTemplate{
				Id: &core.Identifier{Name: "annotatedInput"},
				Interface: &core.TypedInterface{
					Inputs: &core.VariableMap{
						Variables: withoutAnnotations,
					},
				},
			},
		},
		{
			Template: &core.TaskTemplate{
				Id: &core.Identifier{Name: "unannotatedInput"},
				Interface: &core.TypedInterface{
					Inputs: &core.VariableMap{
						Variables: withoutAnnotations,
					},
				},
			},
		},
		{
			Template: &core.TaskTemplate{
				Id: &core.Identifier{Name: "annotatedOutput"},
				Interface: &core.TypedInterface{
					Outputs: &core.VariableMap{
						Variables: withoutAnnotations,
					},
				},
			},
		},
		{
			Template: &core.TaskTemplate{
				Id: &core.Identifier{Name: "unannotatedOutput"},
				Interface: &core.TypedInterface{
					Outputs: &core.VariableMap{
						Variables: withoutAnnotations,
					},
				},
			},
		},
	}

	errs := errors.NewCompileErrors()

	t.Run("Tasks with annotations", func(t *testing.T) {
		taskMap := buildTasks(tasks, errs)

		annInputTask := taskMap[(&core.Identifier{Name: "annotatedInput"}).String()]
		assert.Nil(t, annInputTask.Interface.Inputs.Variables["a"].Type.Annotation)

		unAnnInputTask := taskMap[(&core.Identifier{Name: "unannotatedInput"}).String()]
		assert.Nil(t, unAnnInputTask.Interface.Inputs.Variables["a"].Type.Annotation)

		annOutputTask := taskMap[(&core.Identifier{Name: "annotatedOutput"}).String()]
		assert.Nil(t, annOutputTask.Interface.Outputs.Variables["a"].Type.Annotation)

		unAnnOutputTask := taskMap[(&core.Identifier{Name: "unannotatedOutput"}).String()]
		assert.Nil(t, unAnnOutputTask.Interface.Outputs.Variables["a"].Type.Annotation)
	})
}
