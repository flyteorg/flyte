package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/common/mocks"
	compilerErrors "github.com/flyteorg/flyte/flytepropeller/pkg/compiler/errors"
)

func createBooleanOperand(val bool) *core.Operand {
	return &core.Operand{
		Val: &core.Operand_Primitive{
			Primitive: &core.Primitive{
				Value: &core.Primitive_Boolean{
					Boolean: val,
				},
			},
		},
	}
}

func Test_validateBranchInterface(t *testing.T) {
	identifier := core.Identifier{
		Name:    "taskName",
		Project: "project",
		Domain:  "domain",
	}

	taskNode := &core.TaskNode{
		Reference: &core.TaskNode_ReferenceId{
			ReferenceId: &identifier,
		},
	}

	coreN2 := &core.Node{
		Id: "n2",
		Target: &core.Node_TaskNode{
			TaskNode: taskNode,
		},
	}

	n2 := &mocks.NodeBuilder{}
	n2.EXPECT().GetId().Return("n2")
	n2.EXPECT().GetCoreNode().Return(coreN2)
	n2.EXPECT().GetTaskNode().Return(taskNode)
	n2.On("SetInterface", mock.Anything)
	n2.EXPECT().GetInputs().Return([]*core.Binding{})
	n2.On("SetID", mock.Anything).Return()
	n2.EXPECT().GetInterface().Return(nil)

	task := &mocks.Task{}
	task.EXPECT().GetInterface().Return(&core.TypedInterface{})

	wf := &mocks.WorkflowBuilder{}
	wf.On("GetTask", mock.Anything).Return(task, true)

	errs := compilerErrors.NewCompileErrors()
	wf.EXPECT().GetOrCreateNodeBuilder(coreN2).Return(n2)

	t.Run("single branch", func(t *testing.T) {
		n := &mocks.NodeBuilder{}
		n.EXPECT().GetInterface().Return(nil)
		n.On("SetID", mock.Anything).Return()
		n.EXPECT().GetId().Return("n1")
		n.EXPECT().GetBranchNode().Return(&core.BranchNode{
			IfElse: &core.IfElseBlock{
				Case: &core.IfBlock{
					Condition: &core.BooleanExpression{
						Expr: &core.BooleanExpression_Comparison{
							Comparison: &core.ComparisonExpression{
								LeftValue:  createBooleanOperand(true),
								RightValue: createBooleanOperand(false),
							},
						},
					},
					ThenNode: coreN2,
				},
			},
		})

		n.EXPECT().GetInputs().Return([]*core.Binding{})

		_, ok := validateBranchInterface(wf, n, errs)
		assert.True(t, ok)
		if errs.HasErrors() {
			assert.NoError(t, errs)
		}
	})

	t.Run("two conditions", func(t *testing.T) {
		n := &mocks.NodeBuilder{}
		n.EXPECT().GetId().Return("n1")
		n.EXPECT().GetInterface().Return(nil)
		n.EXPECT().GetInputs().Return([]*core.Binding{})
		n.EXPECT().GetBranchNode().Return(&core.BranchNode{
			IfElse: &core.IfElseBlock{
				Case: &core.IfBlock{
					Condition: &core.BooleanExpression{
						Expr: &core.BooleanExpression_Comparison{
							Comparison: &core.ComparisonExpression{
								LeftValue:  createBooleanOperand(true),
								RightValue: createBooleanOperand(false),
							},
						},
					},
					ThenNode: coreN2,
				},
				Other: []*core.IfBlock{
					{
						Condition: &core.BooleanExpression{
							Expr: &core.BooleanExpression_Comparison{
								Comparison: &core.ComparisonExpression{
									LeftValue:  createBooleanOperand(true),
									RightValue: createBooleanOperand(false),
								},
							},
						},
						ThenNode: coreN2,
					},
				},
			},
		})

		_, ok := validateBranchInterface(wf, n, errs)
		assert.True(t, ok)
		if errs.HasErrors() {
			assert.NoError(t, errs)
		}
	})
}
