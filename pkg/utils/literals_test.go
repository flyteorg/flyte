package utils

import (
	"reflect"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

func TestMakePrimitive(t *testing.T) {
	{
		v := 1
		p, err := MakePrimitive(v)
		assert.NoError(t, err)
		assert.Equal(t, "*core.Primitive_Integer", reflect.TypeOf(p.Value).String())
		assert.Equal(t, int64(v), p.GetInteger())
	}
	{
		v := int64(1)
		p, err := MakePrimitive(v)
		assert.NoError(t, err)
		assert.Equal(t, "*core.Primitive_Integer", reflect.TypeOf(p.Value).String())
		assert.Equal(t, v, p.GetInteger())
	}
	{
		v := 1.0
		p, err := MakePrimitive(v)
		assert.NoError(t, err)
		assert.Equal(t, "*core.Primitive_FloatValue", reflect.TypeOf(p.Value).String())
		assert.Equal(t, v, p.GetFloatValue())
	}
	{
		v := "blah"
		p, err := MakePrimitive(v)
		assert.NoError(t, err)
		assert.Equal(t, "*core.Primitive_StringValue", reflect.TypeOf(p.Value).String())
		assert.Equal(t, v, p.GetStringValue())
	}
	{
		v := true
		p, err := MakePrimitive(v)
		assert.NoError(t, err)
		assert.Equal(t, "*core.Primitive_Boolean", reflect.TypeOf(p.Value).String())
		assert.Equal(t, v, p.GetBoolean())
	}
	{
		v := time.Now()
		p, err := MakePrimitive(v)
		assert.NoError(t, err)
		assert.Equal(t, "*core.Primitive_Datetime", reflect.TypeOf(p.Value).String())
		j, err := ptypes.TimestampProto(v)
		assert.NoError(t, err)
		assert.Equal(t, j, p.GetDatetime())
	}
	{
		v := time.Second * 10
		p, err := MakePrimitive(v)
		assert.NoError(t, err)
		assert.Equal(t, "*core.Primitive_Duration", reflect.TypeOf(p.Value).String())
		assert.Equal(t, ptypes.DurationProto(v), p.GetDuration())
	}
	{
		v := struct {
		}{}
		_, err := MakePrimitive(v)
		assert.Error(t, err)
	}
}

func TestMustMakePrimitive(t *testing.T) {
	{
		v := struct {
		}{}
		assert.Panics(t, func() {
			MustMakePrimitive(v)
		})
	}
	{
		v := time.Second * 10
		p := MustMakePrimitive(v)
		assert.Equal(t, "*core.Primitive_Duration", reflect.TypeOf(p.Value).String())
		assert.Equal(t, ptypes.DurationProto(v), p.GetDuration())
	}
}

func TestMakePrimitiveLiteral(t *testing.T) {
	{
		v := 1.0
		p, err := MakePrimitiveLiteral(v)
		assert.NoError(t, err)
		assert.NotNil(t, p.GetScalar())
		assert.Equal(t, "*core.Primitive_FloatValue", reflect.TypeOf(p.GetScalar().GetPrimitive().Value).String())
		assert.Equal(t, v, p.GetScalar().GetPrimitive().GetFloatValue())
	}
	{
		v := struct {
		}{}
		_, err := MakePrimitiveLiteral(v)
		assert.Error(t, err)
	}
}

func TestMustMakePrimitiveLiteral(t *testing.T) {
	t.Run("Panic", func(t *testing.T) {
		v := struct {
		}{}
		assert.Panics(t, func() {
			MustMakePrimitiveLiteral(v)
		})
	})
	t.Run("FloatValue", func(t *testing.T) {
		v := 1.0
		p := MustMakePrimitiveLiteral(v)
		assert.NotNil(t, p.GetScalar())
		assert.Equal(t, "*core.Primitive_FloatValue", reflect.TypeOf(p.GetScalar().GetPrimitive().Value).String())
		assert.Equal(t, v, p.GetScalar().GetPrimitive().GetFloatValue())
	})
}

func TestMakeLiteral(t *testing.T) {
	t.Run("Primitive", func(t *testing.T) {
		lit, err := MakeLiteral("test_string")
		assert.NoError(t, err)
		assert.Equal(t, "*core.Primitive_StringValue", reflect.TypeOf(lit.GetScalar().GetPrimitive().Value).String())
	})

	t.Run("Array", func(t *testing.T) {
		lit, err := MakeLiteral([]interface{}{1, 2, 3})
		assert.NoError(t, err)
		assert.Equal(t, "*core.Literal_Collection", reflect.TypeOf(lit.GetValue()).String())
		assert.Equal(t, "*core.Primitive_Integer", reflect.TypeOf(lit.GetCollection().Literals[0].GetScalar().GetPrimitive().Value).String())
	})

	t.Run("Map", func(t *testing.T) {
		lit, err := MakeLiteral(map[string]interface{}{
			"key1": []interface{}{1, 2, 3},
			"key2": []interface{}{5},
		})
		assert.NoError(t, err)
		assert.Equal(t, "*core.Literal_Map", reflect.TypeOf(lit.GetValue()).String())
		assert.Equal(t, "*core.Literal_Collection", reflect.TypeOf(lit.GetMap().Literals["key1"].GetValue()).String())
	})

	t.Run("Binary", func(t *testing.T) {
		s := MakeBinaryLiteral([]byte{'h'})
		assert.Equal(t, []byte{'h'}, s.GetScalar().GetBinary().GetValue())
	})

	t.Run("NoneType", func(t *testing.T) {
		p, err := MakeLiteral(nil)
		assert.NoError(t, err)
		assert.NotNil(t, p.GetScalar())
		assert.Equal(t, "*core.Scalar_NoneType", reflect.TypeOf(p.GetScalar().Value).String())
	})
}

func TestMustMakeLiteral(t *testing.T) {
	v := "hello"
	l := MustMakeLiteral(v)
	assert.NotNil(t, l.GetScalar())
	assert.Equal(t, v, l.GetScalar().GetPrimitive().GetStringValue())
}

func TestMakeBinaryLiteral(t *testing.T) {
	s := MakeBinaryLiteral([]byte{'h'})
	assert.Equal(t, []byte{'h'}, s.GetScalar().GetBinary().GetValue())
}

func TestMakeDefaultLiteralForType(t *testing.T) {

	tests := [][]interface{}{
		{"Integer", core.SimpleType_INTEGER, "*core.Primitive_Integer"},
		{"Float", core.SimpleType_FLOAT, "*core.Primitive_FloatValue"},
		{"String", core.SimpleType_STRING, "*core.Primitive_StringValue"},
		{"Boolean", core.SimpleType_BOOLEAN, "*core.Primitive_Boolean"},
		{"Duration", core.SimpleType_DURATION, "*core.Primitive_Duration"},
		{"Datetime", core.SimpleType_DATETIME, "*core.Primitive_Datetime"},
	}

	for i := range tests {
		name := tests[i][0].(string)
		ty := tests[i][1].(core.SimpleType)
		tyName := tests[i][2].(string)

		t.Run(name, func(t *testing.T) {
			l, err := MakeDefaultLiteralForType(&core.LiteralType{Type: &core.LiteralType_Simple{Simple: ty}})
			assert.NoError(t, err)
			assert.Equal(t, tyName, reflect.TypeOf(l.GetScalar().GetPrimitive().Value).String())
		})
	}

	t.Run("Binary", func(t *testing.T) {
		s, err := MakeLiteral([]byte{'h'})
		assert.NoError(t, err)
		assert.Equal(t, []byte{'h'}, s.GetScalar().GetBinary().GetValue())
	})
}
