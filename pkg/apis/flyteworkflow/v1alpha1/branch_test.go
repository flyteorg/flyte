package v1alpha1_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestMarshalUnMarshal_BranchTask(t *testing.T) {
	r, err := ioutil.ReadFile("testdata/branch.json")
	assert.NoError(t, err)
	o := v1alpha1.NodeSpec{}
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
