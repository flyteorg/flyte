package branch

import (
	"context"
	"testing"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/errors"
	"github.com/lyft/flytepropeller/pkg/utils"
)

// Creates a ComparisonExpression, comparing 2 literals
func getComparisonExpression(lV interface{}, op core.ComparisonExpression_Operator, rV interface{}) (*core.ComparisonExpression, *core.LiteralMap) {
	exp := &core.ComparisonExpression{
		LeftValue: &core.Operand{
			Val: &core.Operand_Var{
				Var: "x",
			},
		},
		Operator: op,
		RightValue: &core.Operand{
			Val: &core.Operand_Var{
				Var: "y",
			},
		},
	}
	inputs := &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"x": utils.MustMakePrimitiveLiteral(lV),
			"y": utils.MustMakePrimitiveLiteral(rV),
		},
	}
	return exp, inputs
}

func createUnaryConjunction(l *core.ComparisonExpression, op core.ConjunctionExpression_LogicalOperator, r *core.ComparisonExpression) *core.ConjunctionExpression {
	return &core.ConjunctionExpression{
		LeftExpression: &core.BooleanExpression{
			Expr: &core.BooleanExpression_Comparison{
				Comparison: l,
			},
		},
		Operator: op,
		RightExpression: &core.BooleanExpression{
			Expr: &core.BooleanExpression_Comparison{
				Comparison: r,
			},
		},
	}
}

func TestEvaluateComparison(t *testing.T) {
	t.Run("ComparePrimitives", func(t *testing.T) {
		// Compare primitives
		exp := &core.ComparisonExpression{
			LeftValue: &core.Operand{
				Val: &core.Operand_Primitive{
					Primitive: utils.MustMakePrimitive(1),
				},
			},
			Operator: core.ComparisonExpression_GT,
			RightValue: &core.Operand{
				Val: &core.Operand_Primitive{
					Primitive: utils.MustMakePrimitive(2),
				},
			},
		}
		v, err := EvaluateComparison(exp, nil)
		assert.NoError(t, err)
		assert.False(t, v)
	})
	t.Run("ComparePrimitiveAndLiteral", func(t *testing.T) {
		// Compare lVal -> primitive and rVal -> literal
		exp := &core.ComparisonExpression{
			LeftValue: &core.Operand{
				Val: &core.Operand_Primitive{
					Primitive: utils.MustMakePrimitive(1),
				},
			},
			Operator: core.ComparisonExpression_GT,
			RightValue: &core.Operand{
				Val: &core.Operand_Var{
					Var: "y",
				},
			},
		}
		inputs := &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"y": utils.MustMakePrimitiveLiteral(2),
			},
		}
		v, err := EvaluateComparison(exp, inputs)
		assert.NoError(t, err)
		assert.False(t, v)
	})
	t.Run("CompareLiteralAndPrimitive", func(t *testing.T) {

		// Compare lVal -> literal and rVal -> primitive
		exp := &core.ComparisonExpression{
			LeftValue: &core.Operand{
				Val: &core.Operand_Var{
					Var: "x",
				},
			},
			Operator: core.ComparisonExpression_GT,
			RightValue: &core.Operand{
				Val: &core.Operand_Primitive{
					Primitive: utils.MustMakePrimitive(2),
				},
			},
		}
		inputs := &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"x": utils.MustMakePrimitiveLiteral(1),
				"y": utils.MustMakePrimitiveLiteral(3),
			},
		}
		v, err := EvaluateComparison(exp, inputs)
		assert.NoError(t, err)
		assert.False(t, v)
	})

	t.Run("CompareLiterals", func(t *testing.T) {
		// Compare lVal -> literal and rVal -> literal
		exp, inputs := getComparisonExpression(1, core.ComparisonExpression_EQ, 1)
		v, err := EvaluateComparison(exp, inputs)
		assert.NoError(t, err)
		assert.True(t, v)
	})

	t.Run("CompareLiterals2", func(t *testing.T) {
		// Compare lVal -> literal and rVal -> literal
		exp, inputs := getComparisonExpression(1, core.ComparisonExpression_NEQ, 1)
		v, err := EvaluateComparison(exp, inputs)
		assert.NoError(t, err)
		assert.False(t, v)
	})
	t.Run("ComparePrimitiveAndLiteralNotFound", func(t *testing.T) {
		// Compare lVal -> primitive and rVal -> literal
		exp := &core.ComparisonExpression{
			LeftValue: &core.Operand{
				Val: &core.Operand_Primitive{
					Primitive: utils.MustMakePrimitive(1),
				},
			},
			Operator: core.ComparisonExpression_GT,
			RightValue: &core.Operand{
				Val: &core.Operand_Var{
					Var: "y",
				},
			},
		}
		inputs := &core.LiteralMap{
			Literals: map[string]*core.Literal{},
		}
		_, err := EvaluateComparison(exp, inputs)
		assert.Error(t, err)

		_, err = EvaluateComparison(exp, nil)
		assert.Error(t, err)
	})

	t.Run("CompareLiteralNotFoundAndPrimitive", func(t *testing.T) {
		// Compare lVal -> primitive and rVal -> literal
		exp := &core.ComparisonExpression{
			LeftValue: &core.Operand{
				Val: &core.Operand_Var{
					Var: "y",
				},
			},
			Operator: core.ComparisonExpression_GT,
			RightValue: &core.Operand{
				Val: &core.Operand_Primitive{
					Primitive: utils.MustMakePrimitive(1),
				},
			},
		}
		inputs := &core.LiteralMap{
			Literals: map[string]*core.Literal{},
		}
		_, err := EvaluateComparison(exp, inputs)
		assert.Error(t, err)

		_, err = EvaluateComparison(exp, nil)
		assert.Error(t, err)
	})

}

func TestEvaluateBooleanExpression(t *testing.T) {
	{
		// Simple comparison only
		ce, inputs := getComparisonExpression(1, core.ComparisonExpression_EQ, 1)
		exp := &core.BooleanExpression{
			Expr: &core.BooleanExpression_Comparison{
				Comparison: ce,
			},
		}
		v, err := EvaluateBooleanExpression(exp, inputs)
		assert.NoError(t, err)
		assert.True(t, v)
	}
	{
		// AND of 2 comparisons. Inputs are the same for both.
		l, lInputs := getComparisonExpression(1, core.ComparisonExpression_EQ, 1)
		r, _ := getComparisonExpression(1, core.ComparisonExpression_NEQ, 1)

		exp := &core.BooleanExpression{
			Expr: &core.BooleanExpression_Conjunction{
				Conjunction: createUnaryConjunction(l, core.ConjunctionExpression_AND, r),
			},
		}
		v, err := EvaluateBooleanExpression(exp, lInputs)
		assert.NoError(t, err)
		assert.False(t, v)
	}
	{
		// OR of 2 comparisons
		l, _ := getComparisonExpression(1, core.ComparisonExpression_EQ, 1)
		r, inputs := getComparisonExpression(1, core.ComparisonExpression_NEQ, 1)

		exp := &core.BooleanExpression{
			Expr: &core.BooleanExpression_Conjunction{
				Conjunction: createUnaryConjunction(l, core.ConjunctionExpression_OR, r),
			},
		}
		v, err := EvaluateBooleanExpression(exp, inputs)
		assert.NoError(t, err)
		assert.True(t, v)
	}
	{
		// Conjunction of comparison and a conjunction, AND
		l, _ := getComparisonExpression(1, core.ComparisonExpression_EQ, 1)
		r, inputs := getComparisonExpression(1, core.ComparisonExpression_NEQ, 1)

		innerExp := &core.BooleanExpression{
			Expr: &core.BooleanExpression_Conjunction{
				Conjunction: createUnaryConjunction(l, core.ConjunctionExpression_OR, r),
			},
		}

		outerComparison := &core.ComparisonExpression{
			LeftValue: &core.Operand{
				Val: &core.Operand_Var{
					Var: "a",
				},
			},
			Operator: core.ComparisonExpression_GT,
			RightValue: &core.Operand{
				Val: &core.Operand_Var{
					Var: "b",
				},
			},
		}
		outerInputs := &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"a": utils.MustMakePrimitiveLiteral(5),
				"b": utils.MustMakePrimitiveLiteral(4),
			},
		}

		outerExp := &core.BooleanExpression{
			Expr: &core.BooleanExpression_Conjunction{
				Conjunction: &core.ConjunctionExpression{
					LeftExpression: &core.BooleanExpression{
						Expr: &core.BooleanExpression_Comparison{
							Comparison: outerComparison,
						},
					},
					Operator:        core.ConjunctionExpression_AND,
					RightExpression: innerExp,
				},
			},
		}

		for k, v := range inputs.Literals {
			outerInputs.Literals[k] = v
		}

		v, err := EvaluateBooleanExpression(outerExp, outerInputs)
		assert.NoError(t, err)
		assert.True(t, v)
	}
}

func TestEvaluateIfBlock(t *testing.T) {
	{
		// AND of 2 comparisons
		l, _ := getComparisonExpression(1, core.ComparisonExpression_EQ, 1)
		r, inputs := getComparisonExpression(1, core.ComparisonExpression_NEQ, 1)

		thenNode := "test"
		block := &v1alpha1.IfBlock{
			Condition: v1alpha1.BooleanExpression{
				BooleanExpression: &core.BooleanExpression{
					Expr: &core.BooleanExpression_Conjunction{
						Conjunction: createUnaryConjunction(l, core.ConjunctionExpression_AND, r),
					},
				},
			},
			ThenNode: &thenNode,
		}

		skippedNodeIds := make([]*v1alpha1.NodeID, 0)
		accp, skippedNodeIds, err := EvaluateIfBlock(block, inputs, skippedNodeIds)
		assert.NoError(t, err)
		assert.Nil(t, accp)
		assert.Equal(t, 1, len(skippedNodeIds))
		assert.Equal(t, "test", *skippedNodeIds[0])
	}
	{
		// OR of 2 comparisons
		l, _ := getComparisonExpression(1, core.ComparisonExpression_EQ, 1)
		r, inputs := getComparisonExpression(1, core.ComparisonExpression_NEQ, 1)

		thenNode := "test"
		block := &v1alpha1.IfBlock{
			Condition: v1alpha1.BooleanExpression{
				BooleanExpression: &core.BooleanExpression{
					Expr: &core.BooleanExpression_Conjunction{
						Conjunction: createUnaryConjunction(l, core.ConjunctionExpression_OR, r),
					},
				},
			},
			ThenNode: &thenNode,
		}

		skippedNodeIds := make([]*v1alpha1.NodeID, 0)
		accp, skippedNodeIds, err := EvaluateIfBlock(block, inputs, skippedNodeIds)
		assert.NoError(t, err)
		assert.NotNil(t, accp)
		assert.Equal(t, "test", *accp)
		assert.Equal(t, 0, len(skippedNodeIds))
	}
}

func TestDecideBranch(t *testing.T) {
	ctx := context.Background()

	t.Run("EmptyIfBlock", func(t *testing.T) {
		w := &v1alpha1.FlyteWorkflow{
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID:    "w1",
				Nodes: map[v1alpha1.NodeID]*v1alpha1.NodeSpec{},
			},
		}
		branchNode := &v1alpha1.BranchNodeSpec{}
		b, err := DecideBranch(ctx, w, "n1", branchNode, nil)
		assert.Error(t, err)
		assert.Nil(t, b)
	})

	t.Run("MissingThenNode", func(t *testing.T) {
		w := &v1alpha1.FlyteWorkflow{
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID:    "w1",
				Nodes: map[v1alpha1.NodeID]*v1alpha1.NodeSpec{},
			},
		}
		exp, inputs := getComparisonExpression(1.0, core.ComparisonExpression_EQ, 1.0)
		branchNode := &v1alpha1.BranchNodeSpec{
			If: v1alpha1.IfBlock{
				Condition: v1alpha1.BooleanExpression{
					BooleanExpression: &core.BooleanExpression{
						Expr: &core.BooleanExpression_Comparison{
							Comparison: exp,
						},
					},
				},
				ThenNode: nil,
			},
		}
		b, err := DecideBranch(ctx, w, "n1", branchNode, inputs)
		assert.Error(t, err)
		assert.Nil(t, b)
		assert.Equal(t, errors.NoBranchTakenError, err.(*errors.NodeError).Code)
	})

	t.Run("WithThenNode", func(t *testing.T) {
		n1 := "n1"
		w := &v1alpha1.FlyteWorkflow{
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
				Nodes: map[v1alpha1.NodeID]*v1alpha1.NodeSpec{
					n1: {
						ID: n1,
					},
				},
			},
		}
		exp, inputs := getComparisonExpression(1.0, core.ComparisonExpression_EQ, 1.0)
		branchNode := &v1alpha1.BranchNodeSpec{
			If: v1alpha1.IfBlock{
				Condition: v1alpha1.BooleanExpression{
					BooleanExpression: &core.BooleanExpression{
						Expr: &core.BooleanExpression_Comparison{
							Comparison: exp,
						},
					},
				},
				ThenNode: &n1,
			},
		}
		b, err := DecideBranch(ctx, w, "n1", branchNode, inputs)
		assert.NoError(t, err)
		assert.NotNil(t, b)
		assert.Equal(t, n1, *b)
	})

	t.Run("RepeatedCondition", func(t *testing.T) {
		n1 := "n1"
		n2 := "n2"
		w := &v1alpha1.FlyteWorkflow{
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
				Nodes: map[v1alpha1.NodeID]*v1alpha1.NodeSpec{
					n1: {
						ID: n1,
					},
					n2: {
						ID: n2,
					},
				},
			},
		}
		exp, inputs := getComparisonExpression(1.0, core.ComparisonExpression_EQ, 1.0)
		branchNode := &v1alpha1.BranchNodeSpec{
			If: v1alpha1.IfBlock{
				Condition: v1alpha1.BooleanExpression{
					BooleanExpression: &core.BooleanExpression{
						Expr: &core.BooleanExpression_Comparison{
							Comparison: exp,
						},
					},
				},
				ThenNode: &n1,
			},
			ElseIf: []*v1alpha1.IfBlock{
				{
					Condition: v1alpha1.BooleanExpression{
						BooleanExpression: &core.BooleanExpression{
							Expr: &core.BooleanExpression_Comparison{
								Comparison: exp,
							},
						},
					},
					ThenNode: &n2,
				},
			},
		}
		b, err := DecideBranch(ctx, w, "n", branchNode, inputs)
		assert.NoError(t, err)
		assert.NotNil(t, b)
		assert.Equal(t, n1, *b)
		assert.Equal(t, v1alpha1.NodePhaseSkipped, w.Status.NodeStatus[n2].GetPhase())
		assert.Nil(t, w.Status.NodeStatus[n1])
	})

	t.Run("SecondCondition", func(t *testing.T) {
		n1 := "n1"
		n2 := "n2"
		w := &v1alpha1.FlyteWorkflow{
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
				Nodes: map[v1alpha1.NodeID]*v1alpha1.NodeSpec{
					n1: {
						ID: n1,
					},
					n2: {
						ID: n2,
					},
				},
			},
		}
		exp1, inputs := getComparisonExpression(1, core.ComparisonExpression_NEQ, 1)
		exp2, _ := getComparisonExpression(1, core.ComparisonExpression_EQ, 1)
		branchNode := &v1alpha1.BranchNodeSpec{
			If: v1alpha1.IfBlock{
				Condition: v1alpha1.BooleanExpression{
					BooleanExpression: &core.BooleanExpression{
						Expr: &core.BooleanExpression_Comparison{
							Comparison: exp1,
						},
					},
				},
				ThenNode: &n1,
			},
			ElseIf: []*v1alpha1.IfBlock{
				{
					Condition: v1alpha1.BooleanExpression{
						BooleanExpression: &core.BooleanExpression{
							Expr: &core.BooleanExpression_Comparison{
								Comparison: exp2,
							},
						},
					},
					ThenNode: &n2,
				},
			},
		}
		b, err := DecideBranch(ctx, w, "n", branchNode, inputs)
		assert.NoError(t, err)
		assert.NotNil(t, b)
		assert.Equal(t, n2, *b)
		assert.Nil(t, w.Status.NodeStatus[n2])
		assert.Equal(t, v1alpha1.NodePhaseSkipped, w.Status.NodeStatus[n1].GetPhase())
	})

	t.Run("ElseCase", func(t *testing.T) {
		n1 := "n1"
		n2 := "n2"
		n3 := "n3"
		w := &v1alpha1.FlyteWorkflow{
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
				Nodes: map[v1alpha1.NodeID]*v1alpha1.NodeSpec{
					n1: {
						ID: n1,
					},
					n2: {
						ID: n2,
					},
				},
			},
		}
		exp1, inputs := getComparisonExpression(1, core.ComparisonExpression_NEQ, 1)
		exp2, _ := getComparisonExpression(1, core.ComparisonExpression_NEQ, 1)
		branchNode := &v1alpha1.BranchNodeSpec{
			If: v1alpha1.IfBlock{
				Condition: v1alpha1.BooleanExpression{
					BooleanExpression: &core.BooleanExpression{
						Expr: &core.BooleanExpression_Comparison{
							Comparison: exp1,
						},
					},
				},
				ThenNode: &n1,
			},
			ElseIf: []*v1alpha1.IfBlock{
				{
					Condition: v1alpha1.BooleanExpression{
						BooleanExpression: &core.BooleanExpression{
							Expr: &core.BooleanExpression_Comparison{
								Comparison: exp2,
							},
						},
					},
					ThenNode: &n2,
				},
			},
			Else: &n3,
		}
		b, err := DecideBranch(ctx, w, "n", branchNode, inputs)
		assert.NoError(t, err)
		assert.NotNil(t, b)
		assert.Equal(t, n3, *b)
		assert.Equal(t, v1alpha1.NodePhaseSkipped, w.Status.NodeStatus[n1].GetPhase())
		assert.Equal(t, v1alpha1.NodePhaseSkipped, w.Status.NodeStatus[n2].GetPhase())
	})

	t.Run("MissingNode", func(t *testing.T) {
		n1 := "n1"
		n2 := "n2"
		n3 := "n3"
		w := &v1alpha1.FlyteWorkflow{
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
				Nodes: map[v1alpha1.NodeID]*v1alpha1.NodeSpec{
					n1: {
						ID: n1,
					},
				},
			},
		}
		exp1, inputs := getComparisonExpression(1, core.ComparisonExpression_NEQ, 1)
		exp2, _ := getComparisonExpression(1, core.ComparisonExpression_NEQ, 1)
		branchNode := &v1alpha1.BranchNodeSpec{
			If: v1alpha1.IfBlock{
				Condition: v1alpha1.BooleanExpression{
					BooleanExpression: &core.BooleanExpression{
						Expr: &core.BooleanExpression_Comparison{
							Comparison: exp1,
						},
					},
				},
				ThenNode: &n1,
			},
			ElseIf: []*v1alpha1.IfBlock{
				{
					Condition: v1alpha1.BooleanExpression{
						BooleanExpression: &core.BooleanExpression{
							Expr: &core.BooleanExpression_Comparison{
								Comparison: exp2,
							},
						},
					},
					ThenNode: &n2,
				},
			},
			Else: &n3,
		}
		b, err := DecideBranch(ctx, w, "n", branchNode, inputs)
		assert.Error(t, err)
		assert.Nil(t, b)
		assert.Equal(t, errors.DownstreamNodeNotFoundError, err.(*errors.NodeError).Code)
	})

	t.Run("ElseFailCase", func(t *testing.T) {
		n1 := "n1"
		n2 := "n2"
		userError := "User error"
		w := &v1alpha1.FlyteWorkflow{
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
				Nodes: map[v1alpha1.NodeID]*v1alpha1.NodeSpec{
					n1: {
						ID: n1,
					},
					n2: {
						ID: n2,
					},
				},
			},
		}
		exp1, inputs := getComparisonExpression(1, core.ComparisonExpression_NEQ, 1)
		exp2, _ := getComparisonExpression(1, core.ComparisonExpression_NEQ, 1)
		branchNode := &v1alpha1.BranchNodeSpec{
			If: v1alpha1.IfBlock{
				Condition: v1alpha1.BooleanExpression{
					BooleanExpression: &core.BooleanExpression{
						Expr: &core.BooleanExpression_Comparison{
							Comparison: exp1,
						},
					},
				},
				ThenNode: &n1,
			},
			ElseIf: []*v1alpha1.IfBlock{
				{
					Condition: v1alpha1.BooleanExpression{
						BooleanExpression: &core.BooleanExpression{
							Expr: &core.BooleanExpression_Comparison{
								Comparison: exp2,
							},
						},
					},
					ThenNode: &n2,
				},
			},
			ElseFail: &v1alpha1.Error{
				Error: &core.Error{
					Message: userError,
				},
			},
		}
		b, err := DecideBranch(ctx, w, "n", branchNode, inputs)
		assert.Error(t, err)
		assert.Nil(t, b)
		assert.Equal(t, errors.UserProvidedError, err.(*errors.NodeError).Code)
		assert.Equal(t, userError, err.(*errors.NodeError).Message)
		assert.Equal(t, v1alpha1.NodePhaseSkipped, w.Status.NodeStatus[n1].GetPhase())
		assert.Equal(t, v1alpha1.NodePhaseSkipped, w.Status.NodeStatus[n2].GetPhase())
	})
}
