package assert

import (
	"reflect"
	"testing"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

func EqualPrimitive(t *testing.T, p1 *core.Primitive, p2 *core.Primitive) {
	if p1 != nil {
		assert.NotNil(t, p2)
	}
	assert.Equal(t, reflect.TypeOf(p1.Value), reflect.TypeOf(p2.Value))
	switch p1.Value.(type) {
	case *core.Primitive_Integer:
		assert.Equal(t, p1.GetInteger(), p2.GetInteger())
	case *core.Primitive_StringValue:
		assert.Equal(t, p1.GetStringValue(), p2.GetStringValue())
	default:
		assert.FailNow(t, "Not yet implemented for types %v", reflect.TypeOf(p1.Value))
	}
}

func EqualScalar(t *testing.T, p1 *core.Scalar, p2 *core.Scalar) {
	if p1 != nil {
		assert.NotNil(t, p2)
	}
	assert.Equal(t, reflect.TypeOf(p1.Value), reflect.TypeOf(p2.Value))
	switch p1.Value.(type) {
	case *core.Scalar_Primitive:
		EqualPrimitive(t, p1.GetPrimitive(), p2.GetPrimitive())
	default:
		assert.FailNow(t, "Not yet implemented for types %v", reflect.TypeOf(p1.Value))
	}
}

func EqualLiterals(t *testing.T, l1 *core.Literal, l2 *core.Literal) {
	if l1 != nil {
		assert.NotNil(t, l2)
	} else {
		assert.FailNow(t, "expected value is nil")
	}
	assert.Equal(t, reflect.TypeOf(l1.Value), reflect.TypeOf(l2.Value))
	switch l1.Value.(type) {
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
		assert.Equal(t, len(l1.Literals), len(l2.Literals))
		for k, v := range l1.Literals {
			actual, ok := l2.Literals[k]
			assert.True(t, ok)
			EqualLiterals(t, v, actual)
		}
	}
}

func EqualLiteralCollection(t *testing.T, l1 *core.LiteralCollection, l2 *core.LiteralCollection) {
	if assert.NotNil(t, l2) {
		assert.Equal(t, len(l1.Literals), len(l2.Literals))
		for i, v := range l1.Literals {
			EqualLiterals(t, v, l2.Literals[i])
		}
	}
}
