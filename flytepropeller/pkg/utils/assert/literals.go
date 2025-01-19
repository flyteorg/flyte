package assert

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func EqualLiteralType(t *testing.T, lt1 *core.LiteralType, lt2 *core.LiteralType) {
	if !assert.Equal(t, lt1 == nil, lt2 == nil) {
		assert.FailNow(t, "One of the values is nil")
	}
	assert.Equal(t, reflect.TypeOf(lt1.GetType()), reflect.TypeOf(lt2.GetType()))
	switch lt1.GetType().(type) {
	case *core.LiteralType_Simple:
		assert.Equal(t, lt1.GetType().(*core.LiteralType_Simple).Simple, lt2.GetType().(*core.LiteralType_Simple).Simple)
	default:
		assert.FailNow(t, "Not yet implemented for types %v", reflect.TypeOf(lt1.GetType()))
	}
	structure1 := lt1.GetStructure()
	structure2 := lt2.GetStructure()
	if structure1 != nil && structure2 != nil {
		assert.Equal(t, structure1.GetTag(), structure2.GetTag())
	}
}

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

func EqualError(t *testing.T, e1 *core.Error, e2 *core.Error) {
	if !assert.Equal(t, e1 == nil, e2 == nil) {
		assert.FailNow(t, "One of the values is nil")
	}
	assert.Equal(t, e1.GetMessage(), e2.GetMessage())
	assert.Equal(t, e1.GetFailedNodeId(), e2.GetFailedNodeId())
}

func EqualUnion(t *testing.T, u1 *core.Union, u2 *core.Union) {
	if !assert.Equal(t, u1 == nil, u2 == nil) {
		assert.FailNow(t, "One of the values is nil")
	}
	assert.Equal(t, reflect.TypeOf(u1.GetValue()), reflect.TypeOf(u2.GetValue()))
	EqualLiterals(t, u1.GetValue(), u2.GetValue())
	EqualLiteralType(t, u1.GetType(), u2.GetType())
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
	case *core.Scalar_Error:
		EqualError(t, p1.GetError(), p2.GetError())
	case *core.Scalar_Union:
		EqualUnion(t, p1.GetUnion(), p2.GetUnion())
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
