package v1alpha1

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func TestMarshalUnMarshal_BranchTask(t *testing.T) {
	r, err := ioutil.ReadFile("testdata/branch.json")
	assert.NoError(t, err)
	o := NodeSpec{}
	err = json.Unmarshal(r, &o)
	assert.NoError(t, err)
	assert.NotNil(t, o.BranchNode.If)
	assert.Equal(t, core.ComparisonExpression_GT, o.GetBranchNode().GetIf().GetCondition().GetComparison().GetOperator())
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
	errorMessage := &core.Error{
		Message: "Test error",
	}

	ifNode := "ifNode"
	elifNode := "elifNode"
	elseNode := "elseNode"

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

	assert.Equal(t, errorMessage, branchNodeSpec.GetElseFail())

	branchNodeSpec.ElseFail = nil
	assert.Nil(t, branchNodeSpec.GetElseFail())
}
