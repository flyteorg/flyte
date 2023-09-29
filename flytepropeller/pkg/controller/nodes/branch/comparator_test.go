package branch

import (
	"fmt"
	"testing"
	"time"

	"github.com/flyteorg/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

func TestEvaluate_int(t *testing.T) {
	p1 := coreutils.MustMakePrimitive(1)
	p2 := coreutils.MustMakePrimitive(2)
	{
		// p1 > p2 = false
		b, err := Evaluate(p1, p2, core.ComparisonExpression_GT)
		assert.NoError(t, err)
		assert.False(t, b)
		b, err = Evaluate(p2, p1, core.ComparisonExpression_GT)
		assert.NoError(t, err)
		assert.True(t, b)
	}
	{
		// p1 >= p2 = false
		b, err := Evaluate(p1, p2, core.ComparisonExpression_GTE)
		assert.NoError(t, err)
		assert.False(t, b)
		b, err = Evaluate(p2, p1, core.ComparisonExpression_GTE)
		assert.NoError(t, err)
		assert.True(t, b)
	}
	{
		// p1 < p2 = true
		b, err := Evaluate(p1, p2, core.ComparisonExpression_LT)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p2, p1, core.ComparisonExpression_LT)
		assert.NoError(t, err)
		assert.False(t, b)
	}
	{
		// p1 <= p2 = true
		b, err := Evaluate(p1, p2, core.ComparisonExpression_LTE)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p2, p1, core.ComparisonExpression_LTE)
		assert.NoError(t, err)
		assert.False(t, b)
	}
	{
		b, err := Evaluate(p1, p2, core.ComparisonExpression_NEQ)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p2, p1, core.ComparisonExpression_EQ)
		assert.NoError(t, err)
		assert.False(t, b)
		b, err = Evaluate(p1, p1, core.ComparisonExpression_EQ)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p1, p1, core.ComparisonExpression_NEQ)
		assert.NoError(t, err)
		assert.False(t, b)
	}
	{
		b, err := Evaluate(p1, p1, core.ComparisonExpression_LTE)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p1, p1, core.ComparisonExpression_LT)
		assert.NoError(t, err)
		assert.False(t, b)
	}
	{
		b, err := Evaluate(p1, p1, core.ComparisonExpression_GTE)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p1, p1, core.ComparisonExpression_GT)
		assert.NoError(t, err)
		assert.False(t, b)
	}
}

func TestEvaluate_float(t *testing.T) {
	p1 := coreutils.MustMakePrimitive(1.0)
	p2 := coreutils.MustMakePrimitive(2.0)
	{
		// p1 > p2 = false
		b, err := Evaluate(p1, p2, core.ComparisonExpression_GT)
		assert.NoError(t, err)
		assert.False(t, b)
		b, err = Evaluate(p2, p1, core.ComparisonExpression_GT)
		assert.NoError(t, err)
		assert.True(t, b)
	}
	{
		// p1 >= p2 = false
		b, err := Evaluate(p1, p2, core.ComparisonExpression_GTE)
		assert.NoError(t, err)
		assert.False(t, b)
		b, err = Evaluate(p2, p1, core.ComparisonExpression_GTE)
		assert.NoError(t, err)
		assert.True(t, b)
	}
	{
		// p1 < p2 = true
		b, err := Evaluate(p1, p2, core.ComparisonExpression_LT)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p2, p1, core.ComparisonExpression_LT)
		assert.NoError(t, err)
		assert.False(t, b)
	}
	{
		// p1 <= p2 = true
		b, err := Evaluate(p1, p2, core.ComparisonExpression_LTE)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p2, p1, core.ComparisonExpression_LTE)
		assert.NoError(t, err)
		assert.False(t, b)
	}
	{
		b, err := Evaluate(p1, p2, core.ComparisonExpression_NEQ)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p2, p1, core.ComparisonExpression_EQ)
		assert.NoError(t, err)
		assert.False(t, b)
		b, err = Evaluate(p1, p1, core.ComparisonExpression_EQ)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p1, p1, core.ComparisonExpression_NEQ)
		assert.NoError(t, err)
		assert.False(t, b)
	}
	{
		b, err := Evaluate(p1, p1, core.ComparisonExpression_LTE)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p1, p1, core.ComparisonExpression_LT)
		assert.NoError(t, err)
		assert.False(t, b)
	}
	{
		b, err := Evaluate(p1, p1, core.ComparisonExpression_GTE)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p1, p1, core.ComparisonExpression_GT)
		assert.NoError(t, err)
		assert.False(t, b)
	}
}

func TestEvaluate_string(t *testing.T) {
	p1 := coreutils.MustMakePrimitive("a")
	p2 := coreutils.MustMakePrimitive("b")
	{
		// p1 > p2 = false
		b, err := Evaluate(p1, p2, core.ComparisonExpression_GT)
		assert.NoError(t, err)
		assert.False(t, b)
		b, err = Evaluate(p2, p1, core.ComparisonExpression_GT)
		assert.NoError(t, err)
		assert.True(t, b)
	}
	{
		// p1 >= p2 = false
		b, err := Evaluate(p1, p2, core.ComparisonExpression_GTE)
		assert.NoError(t, err)
		assert.False(t, b)
		b, err = Evaluate(p2, p1, core.ComparisonExpression_GTE)
		assert.NoError(t, err)
		assert.True(t, b)
	}
	{
		// p1 < p2 = true
		b, err := Evaluate(p1, p2, core.ComparisonExpression_LT)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p2, p1, core.ComparisonExpression_LT)
		assert.NoError(t, err)
		assert.False(t, b)
	}
	{
		// p1 <= p2 = true
		b, err := Evaluate(p1, p2, core.ComparisonExpression_LTE)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p2, p1, core.ComparisonExpression_LTE)
		assert.NoError(t, err)
		assert.False(t, b)
	}
	{
		b, err := Evaluate(p1, p2, core.ComparisonExpression_NEQ)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p2, p1, core.ComparisonExpression_EQ)
		assert.NoError(t, err)
		assert.False(t, b)
		b, err = Evaluate(p1, p1, core.ComparisonExpression_EQ)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p1, p1, core.ComparisonExpression_NEQ)
		assert.NoError(t, err)
		assert.False(t, b)
	}
	{
		b, err := Evaluate(p1, p1, core.ComparisonExpression_LTE)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p1, p1, core.ComparisonExpression_LT)
		assert.NoError(t, err)
		assert.False(t, b)
	}
	{
		b, err := Evaluate(p1, p1, core.ComparisonExpression_GTE)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p1, p1, core.ComparisonExpression_GT)
		assert.NoError(t, err)
		assert.False(t, b)
	}
}

func TestEvaluate_datetime(t *testing.T) {
	p1 := coreutils.MustMakePrimitive(time.Date(2018, 7, 4, 12, 00, 00, 00, time.UTC))
	p2 := coreutils.MustMakePrimitive(time.Date(2018, 7, 4, 12, 00, 01, 00, time.UTC))
	{
		// p1 > p2 = false
		b, err := Evaluate(p1, p2, core.ComparisonExpression_GT)
		assert.NoError(t, err)
		assert.False(t, b)
		b, err = Evaluate(p2, p1, core.ComparisonExpression_GT)
		assert.NoError(t, err)
		assert.True(t, b)
	}
	{
		// p1 >= p2 = false
		b, err := Evaluate(p1, p2, core.ComparisonExpression_GTE)
		assert.NoError(t, err)
		assert.False(t, b)
		b, err = Evaluate(p2, p1, core.ComparisonExpression_GTE)
		assert.NoError(t, err)
		assert.True(t, b)
	}
	{
		// p1 < p2 = true
		b, err := Evaluate(p1, p2, core.ComparisonExpression_LT)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p2, p1, core.ComparisonExpression_LT)
		assert.NoError(t, err)
		assert.False(t, b)
	}
	{
		// p1 <= p2 = true
		b, err := Evaluate(p1, p2, core.ComparisonExpression_LTE)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p2, p1, core.ComparisonExpression_LTE)
		assert.NoError(t, err)
		assert.False(t, b)
	}
	{
		b, err := Evaluate(p1, p2, core.ComparisonExpression_NEQ)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p2, p1, core.ComparisonExpression_EQ)
		assert.NoError(t, err)
		assert.False(t, b)
		b, err = Evaluate(p1, p1, core.ComparisonExpression_EQ)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p1, p1, core.ComparisonExpression_NEQ)
		assert.NoError(t, err)
		assert.False(t, b)
	}
	{
		b, err := Evaluate(p1, p1, core.ComparisonExpression_LTE)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p1, p1, core.ComparisonExpression_LT)
		assert.NoError(t, err)
		assert.False(t, b)
	}
	{
		b, err := Evaluate(p1, p1, core.ComparisonExpression_GTE)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p1, p1, core.ComparisonExpression_GT)
		assert.NoError(t, err)
		assert.False(t, b)
	}
}

func TestEvaluate_duration(t *testing.T) {
	p1 := coreutils.MustMakePrimitive(10 * time.Second)
	p2 := coreutils.MustMakePrimitive(11 * time.Second)
	{
		// p1 > p2 = false
		b, err := Evaluate(p1, p2, core.ComparisonExpression_GT)
		assert.NoError(t, err)
		assert.False(t, b)
		b, err = Evaluate(p2, p1, core.ComparisonExpression_GT)
		assert.NoError(t, err)
		assert.True(t, b)
	}
	{
		// p1 >= p2 = false
		b, err := Evaluate(p1, p2, core.ComparisonExpression_GTE)
		assert.NoError(t, err)
		assert.False(t, b)
		b, err = Evaluate(p2, p1, core.ComparisonExpression_GTE)
		assert.NoError(t, err)
		assert.True(t, b)
	}
	{
		// p1 < p2 = true
		b, err := Evaluate(p1, p2, core.ComparisonExpression_LT)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p2, p1, core.ComparisonExpression_LT)
		assert.NoError(t, err)
		assert.False(t, b)
	}
	{
		// p1 <= p2 = true
		b, err := Evaluate(p1, p2, core.ComparisonExpression_LTE)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p2, p1, core.ComparisonExpression_LTE)
		assert.NoError(t, err)
		assert.False(t, b)
	}
	{
		b, err := Evaluate(p1, p2, core.ComparisonExpression_NEQ)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p2, p1, core.ComparisonExpression_EQ)
		assert.NoError(t, err)
		assert.False(t, b)
		b, err = Evaluate(p1, p1, core.ComparisonExpression_EQ)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p1, p1, core.ComparisonExpression_NEQ)
		assert.NoError(t, err)
		assert.False(t, b)
	}
	{
		b, err := Evaluate(p1, p1, core.ComparisonExpression_LTE)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p1, p1, core.ComparisonExpression_LT)
		assert.NoError(t, err)
		assert.False(t, b)
	}
	{
		b, err := Evaluate(p1, p1, core.ComparisonExpression_GTE)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p1, p1, core.ComparisonExpression_GT)
		assert.NoError(t, err)
		assert.False(t, b)
	}
}

func TestEvaluate_boolean(t *testing.T) {
	p1 := coreutils.MustMakePrimitive(true)
	p2 := coreutils.MustMakePrimitive(false)
	f := func(op core.ComparisonExpression_Operator) {
		// GT/LT = false
		msg := fmt.Sprintf("Evaluating: [%s]", op.String())
		b, err := Evaluate(p1, p2, op)
		assert.Error(t, err, msg)
		assert.False(t, b, msg)
		b, err = Evaluate(p2, p1, op)
		assert.Error(t, err, msg)
		assert.False(t, b, msg)
		b, err = Evaluate(p1, p1, op)
		assert.Error(t, err, msg)
		assert.False(t, b, msg)
	}
	f(core.ComparisonExpression_GT)
	f(core.ComparisonExpression_LT)
	f(core.ComparisonExpression_GTE)
	f(core.ComparisonExpression_LTE)

	{
		b, err := Evaluate(p1, p2, core.ComparisonExpression_NEQ)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p2, p1, core.ComparisonExpression_EQ)
		assert.NoError(t, err)
		assert.False(t, b)
		b, err = Evaluate(p1, p1, core.ComparisonExpression_EQ)
		assert.NoError(t, err)
		assert.True(t, b)
		b, err = Evaluate(p1, p1, core.ComparisonExpression_NEQ)
		assert.NoError(t, err)
		assert.False(t, b)
	}
}
