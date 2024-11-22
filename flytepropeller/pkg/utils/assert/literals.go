package assert

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func EqualPrimitive(t *testing.T, p1 *core.Primitive, p2 *core.Primitive) {
	if !assert.Equal(t, p1 == nil, p2 == nil) {
		assert.FailNow(t, "One of the values is nil")
	}
	if p1 == nil {
		return
	}
	assert.Equal(t, reflect.TypeOf(p1.GetValue()), reflect.TypeOf(p2.GetValue()))
	switch p1.GetValue().(type) {
	case *core.Primitive_Integer:
		assert.Equal(t, p1.GetInteger(), p2.GetInteger())
	case *core.Primitive_StringValue:
		assert.Equal(t, p1.GetStringValue(), p2.GetStringValue())
	default:
		assert.FailNow(t, "Not yet implemented for types %v", reflect.TypeOf(p1.GetValue()))
	}
}

func EqualScalar(t *testing.T, p1 *core.Scalar, p2 *core.Scalar) {
	if !assert.Equal(t, p1 == nil, p2 == nil) {
		assert.FailNow(t, "One of the values is nil")
	}
	if p1 == nil {
		return
	}
	assert.Equal(t, reflect.TypeOf(p1.GetValue()), reflect.TypeOf(p2.GetValue()))
	switch p1.GetValue().(type) {
	case *core.Scalar_Primitive:
		EqualPrimitive(t, p1.GetPrimitive(), p2.GetPrimitive())
	default:
		assert.FailNow(t, "Not yet implemented for types %v", reflect.TypeOf(p1.GetValue()))
	}
}

func EqualLiterals(t *testing.T, l1 *core.Literal, l2 *core.Literal) {
	if !assert.Equal(t, l1 == nil, l2 == nil) {
		assert.FailNow(t, "One of the values is nil")
	}
	if l1 == nil {
		return
	}
	assert.Equal(t, reflect.TypeOf(l1.GetValue()), reflect.TypeOf(l2.GetValue()))
	switch l1.GetValue().(type) {
	case *core.Literal_Scalar:
		EqualScalar(t, l1.GetScalar(), l2.GetScalar())
	case *core.Literal_Map:
		EqualLiteralMap(t, l1.GetMap(), l2.GetMap())
	default:
		assert.FailNow(t, "Not supported test type")
	}
}

func EqualLiteralMap(t *testing.T, l1 *core.LiteralMap, l2 *core.LiteralMap) {
	if assert.NotNil(t, l1, "l1 is nil") && assert.NotNil(t, l2, "l2 is nil") {
		assert.Equal(t, len(l1.GetLiterals()), len(l2.GetLiterals()))
		for k, v := range l1.GetLiterals() {
			actual, ok := l2.GetLiterals()[k]
			assert.True(t, ok)
			EqualLiterals(t, v, actual)
		}
	}
}

func EqualLiteralCollection(t *testing.T, l1 *core.LiteralCollection, l2 *core.LiteralCollection) {
	if assert.NotNil(t, l2) {
		assert.Equal(t, len(l1.GetLiterals()), len(l2.GetLiterals()))
		for i, v := range l1.GetLiterals() {
			EqualLiterals(t, v, l2.GetLiterals()[i])
		}
	}
}
