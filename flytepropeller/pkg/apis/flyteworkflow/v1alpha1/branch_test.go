package v1alpha1

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func TestMarshalUnMarshal_BranchTask(t *testing.T) {
	r, err := ioutil.ReadFile("testdata/branch.json")
	assert.NoError(t, err)
	o := NodeSpec{}
	err = json.Unmarshal(r, &o)
	assert.NoError(t, err)
	assert.NotNil(t, o.BranchNode.If)
	assert.Equal(t, core.ComparisonExpression_GT, o.BranchNode.If.Condition.BooleanExpression.GetComparison().Operator)
	assert.Equal(t, 1, len(o.InputBindings))
	raw, err := json.Marshal(o)
	if assert.NoError(t, err) {
		assert.NotEmpty(t, raw)
	}
}

func TestBranchNodeSpecMethods(t *testing.T) {
	// Creating a core.BooleanExpression instance for testing
	boolExpr := &core.BooleanExpression{}

	// Creating an Error instance for testing
	errorMessage := &Error{
		Error: &core.Error{
			Message: "Test error",
		},
	}

	ifNode := NodeID("ifNode")
	elifNode := NodeID("elifNode")
	elseNode := NodeID("elseNode")

	// Creating a BranchNodeSpec instance for testing
	branchNodeSpec := BranchNodeSpec{
		If: IfBlock{
			Condition: BooleanExpression{
				BooleanExpression: boolExpr,
			},
			ThenNode: &ifNode,
		},
		ElseIf: []*IfBlock{
			{
				Condition: BooleanExpression{
					BooleanExpression: boolExpr,
				},
				ThenNode: &elifNode,
			},
		},
		Else:     &elseNode,
		ElseFail: errorMessage,
	}

	assert.Equal(t, boolExpr, branchNodeSpec.If.GetCondition())

	assert.Equal(t, &ifNode, branchNodeSpec.If.GetThenNode())

	assert.Equal(t, &branchNodeSpec.If, branchNodeSpec.GetIf())

	assert.Equal(t, &elseNode, branchNodeSpec.GetElse())

	elifs := branchNodeSpec.GetElseIf()
	assert.Equal(t, 1, len(elifs))
	assert.Equal(t, boolExpr, elifs[0].GetCondition())
	assert.Equal(t, &elifNode, elifs[0].GetThenNode())

	assert.Equal(t, errorMessage.Error, branchNodeSpec.GetElseFail())

	branchNodeSpec.ElseFail = nil
	assert.Nil(t, branchNodeSpec.GetElseFail())
}

func TestWrappedBooleanExpressionDeepCopy(t *testing.T) {
	// 1. Set up proto
	protoBoolExpr := &core.BooleanExpression{
		Expr: &core.BooleanExpression_Comparison{
			Comparison: &core.ComparisonExpression{
				Operator: core.ComparisonExpression_GT,
				LeftValue: &core.Operand{
					Val: &core.Operand_Primitive{
						Primitive: &core.Primitive{
							Value: &core.Primitive_Integer{
								Integer: 10,
							},
						},
					},
				},
				RightValue: &core.Operand{
					Val: &core.Operand_Primitive{
						Primitive: &core.Primitive{
							Value: &core.Primitive_Integer{
								Integer: 11,
							},
						},
					},
				},
			},
		},
	}

	// 2. Define the wrapper object
	boolExpr := BooleanExpression{
		BooleanExpression: protoBoolExpr,
	}

	// 3. Deep copy the wrapper object
	copyBoolExpr := boolExpr.DeepCopy()

	// 4. Compare the pointers and the actual values
	// Assert that the pointers are different
	assert.True(t, boolExpr.BooleanExpression != copyBoolExpr.BooleanExpression)

	// Assert that the values stored in the proto messages are equal
	assert.True(t, proto.Equal(boolExpr.BooleanExpression, copyBoolExpr.BooleanExpression))
}

func TestWrappedErrorDeepCopy(t *testing.T) {
	// 1. Set up proto
	protoError := &core.Error{
		Message: "an error",
	}

	// 2. Define the wrapper object
	error := Error{
		Error: protoError,
	}

	// 3. Deep copy the wrapper object
	errorCopy := error.DeepCopy()

	// 4. Assert that the pointers are different
	assert.True(t, error.Error != errorCopy.Error)

	// 5. Assert that the values stored in the proto messages are equal
	assert.True(t, proto.Equal(error.Error, errorCopy.Error))
}
