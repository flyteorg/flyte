package utils

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

const primitiveString = "hello"

func TestMakePrimitiveBinding(t *testing.T) {
	{
		v := 1.0
		xb, err := MakePrimitiveBindingData(v)
		x := MakeBinding("x", xb)
		assert.NoError(t, err)
		assert.Equal(t, "x", x.GetVar())
		p := x.GetBinding()
		assert.NotNil(t, p.GetScalar())
		assert.Equal(t, "*core.Primitive_FloatValue", reflect.TypeOf(p.GetScalar().GetPrimitive().Value).String())
		assert.Equal(t, v, p.GetScalar().GetPrimitive().GetFloatValue())
	}
	{
		v := struct {
		}{}
		_, err := MakePrimitiveBindingData(v)
		assert.Error(t, err)
	}
}

func TestMustMakePrimitiveBinding(t *testing.T) {
	{
		v := 1.0
		x := MakeBinding("x", MustMakePrimitiveBindingData(v))
		assert.Equal(t, "x", x.GetVar())
		p := x.GetBinding()
		assert.NotNil(t, p.GetScalar())
		assert.Equal(t, "*core.Primitive_FloatValue", reflect.TypeOf(p.GetScalar().GetPrimitive().Value).String())
		assert.Equal(t, v, p.GetScalar().GetPrimitive().GetFloatValue())
	}
	{
		v := struct {
		}{}
		assert.Panics(t, func() {
			MustMakePrimitiveBindingData(v)
		})
	}
}

func TestMakeBindingDataCollection(t *testing.T) {
	v1 := int64(1)
	v2 := primitiveString
	c := MakeBindingDataCollection(
		MustMakePrimitiveBindingData(v1),
		MustMakePrimitiveBindingData(v2),
	)

	c2 := MakeBindingDataCollection(
		MustMakePrimitiveBindingData(v1),
		c,
	)

	assert.NotNil(t, c.GetCollection())
	assert.Equal(t, 2, len(c.GetCollection().Bindings))
	{
		p := c.GetCollection().GetBindings()[0]
		assert.NotNil(t, p.GetScalar())
		assert.Equal(t, "*core.Primitive_Integer", reflect.TypeOf(p.GetScalar().GetPrimitive().Value).String())
		assert.Equal(t, v1, p.GetScalar().GetPrimitive().GetInteger())
	}
	{
		p := c.GetCollection().GetBindings()[1]
		assert.NotNil(t, p.GetScalar())
		assert.Equal(t, "*core.Primitive_StringValue", reflect.TypeOf(p.GetScalar().GetPrimitive().Value).String())
		assert.Equal(t, v2, p.GetScalar().GetPrimitive().GetStringValue())
	}

	assert.NotNil(t, c2.GetCollection())
	assert.Equal(t, 2, len(c2.GetCollection().Bindings))
	{
		p := c2.GetCollection().GetBindings()[0]
		assert.NotNil(t, p.GetScalar())
		assert.Equal(t, "*core.Primitive_Integer", reflect.TypeOf(p.GetScalar().GetPrimitive().Value).String())
		assert.Equal(t, v1, p.GetScalar().GetPrimitive().GetInteger())
	}
	{
		p := c2.GetCollection().GetBindings()[1]
		assert.NotNil(t, p.GetCollection())
		assert.Equal(t, c.GetCollection(), p.GetCollection())
	}
}

func TestMakeBindingDataMap(t *testing.T) {
	v1 := int64(1)
	v2 := primitiveString
	c := MakeBindingDataCollection(
		MustMakePrimitiveBindingData(v1),
		MustMakePrimitiveBindingData(v2),
	)

	m := MakeBindingDataMap(
		NewPair("x", MustMakePrimitiveBindingData(v1)),
		NewPair("y", c),
	)

	m2 := MakeBindingDataMap(
		NewPair("x", MustMakePrimitiveBindingData(v1)),
		NewPair("y", m),
	)
	assert.NotNil(t, m.GetMap())
	assert.Equal(t, 2, len(m.GetMap().GetBindings()))
	{
		p := m.GetMap().GetBindings()["x"]
		assert.NotNil(t, p.GetScalar())
		assert.Equal(t, "*core.Primitive_Integer", reflect.TypeOf(p.GetScalar().GetPrimitive().Value).String())
		assert.Equal(t, v1, p.GetScalar().GetPrimitive().GetInteger())
	}
	{
		p := m.GetMap().GetBindings()["y"]
		assert.NotNil(t, p.GetCollection())
		assert.Equal(t, c.GetCollection(), p.GetCollection())
	}

	assert.NotNil(t, m2.GetMap())
	assert.Equal(t, 2, len(m2.GetMap().GetBindings()))
	{
		p := m2.GetMap().GetBindings()["x"]
		assert.NotNil(t, p.GetScalar())
		assert.Equal(t, "*core.Primitive_Integer", reflect.TypeOf(p.GetScalar().GetPrimitive().Value).String())
		assert.Equal(t, v1, p.GetScalar().GetPrimitive().GetInteger())
	}
	{
		p := m2.GetMap().GetBindings()["y"]
		assert.NotNil(t, p.GetMap())
		assert.Equal(t, m.GetMap(), p.GetMap())
	}

}

func TestMakeBindingPromise(t *testing.T) {
	p := MakeBindingPromise("n1", "x", "y")
	assert.NotNil(t, p)
	assert.Equal(t, "y", p.GetVar())
	assert.NotNil(t, p.GetBinding().GetPromise())
	assert.Equal(t, "n1", p.GetBinding().GetPromise().GetNodeId())
	assert.Equal(t, "x", p.GetBinding().GetPromise().GetVar())
}

func TestMakeBindingDataPromise(t *testing.T) {
	p := MakeBindingDataPromise("n1", "x")
	assert.NotNil(t, p)
	assert.NotNil(t, p.GetPromise())
	assert.Equal(t, "n1", p.GetPromise().GetNodeId())
	assert.Equal(t, "x", p.GetPromise().GetVar())
}
