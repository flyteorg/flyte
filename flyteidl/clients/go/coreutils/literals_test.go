// extract_literal_test.go
// Test class for the utility methods which construct flyte literals.

package coreutils

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/go-test/deep"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/storage"

	"github.com/golang/protobuf/ptypes"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/pkg/errors"
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
		_, err = MakePrimitive(time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC))
		assert.Error(t, err)
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
	type args struct {
		name        string
		ty          core.SimpleType
		tyName      string
		isPrimitive bool
	}
	tests := []args{
		{"None", core.SimpleType_NONE, "*core.Scalar_NoneType", false},
		{"Binary", core.SimpleType_BINARY, "*core.Scalar_Binary", false},
		{"Integer", core.SimpleType_INTEGER, "*core.Primitive_Integer", true},
		{"Float", core.SimpleType_FLOAT, "*core.Primitive_FloatValue", true},
		{"String", core.SimpleType_STRING, "*core.Primitive_StringValue", true},
		{"Boolean", core.SimpleType_BOOLEAN, "*core.Primitive_Boolean", true},
		{"Duration", core.SimpleType_DURATION, "*core.Primitive_Duration", true},
		{"Datetime", core.SimpleType_DATETIME, "*core.Primitive_Datetime", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l, err := MakeDefaultLiteralForType(&core.LiteralType{Type: &core.LiteralType_Simple{Simple: test.ty}})
			assert.NoError(t, err)
			if test.isPrimitive {
				assert.Equal(t, test.tyName, reflect.TypeOf(l.GetScalar().GetPrimitive().Value).String())
			} else {
				assert.Equal(t, test.tyName, reflect.TypeOf(l.GetScalar().Value).String())
			}
		})
	}

	t.Run("Binary", func(t *testing.T) {
		s, err := MakeLiteral([]byte{'h'})
		assert.NoError(t, err)
		assert.Equal(t, []byte{'h'}, s.GetScalar().GetBinary().GetValue())
	})

	t.Run("Blob", func(t *testing.T) {
		l, err := MakeDefaultLiteralForType(&core.LiteralType{Type: &core.LiteralType_Blob{}})
		assert.NoError(t, err)
		assert.Equal(t, "*core.Scalar_Blob", reflect.TypeOf(l.GetScalar().Value).String())
	})

	t.Run("Collection", func(t *testing.T) {
		l, err := MakeDefaultLiteralForType(&core.LiteralType{Type: &core.LiteralType_CollectionType{CollectionType: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}})
		assert.NoError(t, err)
		assert.Equal(t, "*core.LiteralCollection", reflect.TypeOf(l.GetCollection()).String())
	})

	t.Run("Map", func(t *testing.T) {
		l, err := MakeDefaultLiteralForType(&core.LiteralType{Type: &core.LiteralType_MapValueType{MapValueType: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}})
		assert.NoError(t, err)
		assert.Equal(t, "*core.LiteralMap", reflect.TypeOf(l.GetMap()).String())
	})

	t.Run("error", func(t *testing.T) {
		l, err := MakeDefaultLiteralForType(&core.LiteralType{Type: &core.LiteralType_Simple{
			Simple: core.SimpleType_ERROR,
		}})
		assert.NoError(t, err)
		assert.NotNil(t, l.GetScalar().GetError())
	})

	t.Run("struct", func(t *testing.T) {
		l, err := MakeDefaultLiteralForType(&core.LiteralType{Type: &core.LiteralType_Simple{
			Simple: core.SimpleType_STRUCT,
		}})
		assert.NoError(t, err)
		assert.NotNil(t, l.GetScalar().GetGeneric())
	})

	t.Run("enum", func(t *testing.T) {
		l, err := MakeDefaultLiteralForType(&core.LiteralType{Type: &core.LiteralType_EnumType{
			EnumType: &core.EnumType{Values: []string{"x", "y", "z"}},
		}})
		assert.NoError(t, err)
		assert.NotNil(t, l.GetScalar().GetPrimitive().GetStringValue())
		expected := &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_StringValue{StringValue: "x"}}}}}}
		assert.Equal(t, expected, l)
	})

	t.Run("union", func(t *testing.T) {
		l, err := MakeDefaultLiteralForType(
			&core.LiteralType{
				Type: &core.LiteralType_UnionType{
					UnionType: &core.UnionType{
						Variants: []*core.LiteralType{
							{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}},
							{Type: &core.LiteralType_Simple{Simple: core.SimpleType_FLOAT}},
						},
					},
				},
			},
		)
		assert.NoError(t, err)
		assert.Equal(t, "*core.Union", reflect.TypeOf(l.GetScalar().GetUnion()).String())
	})
}

func TestMustMakeDefaultLiteralForType(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		assert.Panics(t, func() {
			MustMakeDefaultLiteralForType(nil)
		})
	})

	t.Run("Blob", func(t *testing.T) {
		l := MustMakeDefaultLiteralForType(&core.LiteralType{Type: &core.LiteralType_Blob{}})
		assert.Equal(t, "*core.Scalar_Blob", reflect.TypeOf(l.GetScalar().Value).String())
	})
}

func TestMakePrimitiveForType(t *testing.T) {
	n := time.Now()
	type args struct {
		t core.SimpleType
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    *core.Primitive
		wantErr bool
	}{
		{"error-type", args{core.SimpleType_NONE, "x"}, nil, true},

		{"error-int", args{core.SimpleType_INTEGER, "x"}, nil, true},
		{"int", args{core.SimpleType_INTEGER, "1"}, MustMakePrimitive(1), false},

		{"error-bool", args{core.SimpleType_BOOLEAN, "x"}, nil, true},
		{"bool", args{core.SimpleType_BOOLEAN, "true"}, MustMakePrimitive(true), false},

		{"error-float", args{core.SimpleType_FLOAT, "x"}, nil, true},
		{"float", args{core.SimpleType_FLOAT, "3.1416"}, MustMakePrimitive(3.1416), false},

		{"string", args{core.SimpleType_STRING, "string"}, MustMakePrimitive("string"), false},

		{"error-dt", args{core.SimpleType_DATETIME, "x"}, nil, true},
		{"dt", args{core.SimpleType_DATETIME, n.Format(time.RFC3339Nano)}, MustMakePrimitive(n), false},

		{"error-dur", args{core.SimpleType_DURATION, "x"}, nil, true},
		{"dur", args{core.SimpleType_DURATION, time.Hour.String()}, MustMakePrimitive(time.Hour), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MakePrimitiveForType(tt.args.t, tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakePrimitiveForType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakePrimitiveForType() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeLiteralForSimpleType(t *testing.T) {
	type args struct {
		t core.SimpleType
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    *core.Literal
		wantErr bool
	}{
		{"error-int", args{core.SimpleType_INTEGER, "x"}, nil, true},
		{"int", args{core.SimpleType_INTEGER, "1"}, MustMakeLiteral(1), false},

		{"error-struct", args{core.SimpleType_STRUCT, "x"}, nil, true},
		{"struct", args{core.SimpleType_STRUCT, `{"x": 1}`}, MustMakeLiteral(&structpb.Struct{Fields: map[string]*structpb.Value{"x": {Kind: &structpb.Value_NumberValue{NumberValue: 1}}}}), false},

		{"bin", args{core.SimpleType_BINARY, "x"}, MustMakeLiteral([]byte("x")), false},

		{"error", args{core.SimpleType_ERROR, "err"}, MustMakeLiteral(&core.Error{Message: "err"}), false},

		{"none", args{core.SimpleType_NONE, "null"}, MustMakeLiteral(nil), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MakeLiteralForSimpleType(tt.args.t, tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeLiteralForSimpleType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := deep.Equal(tt.want, got); diff != nil {
				t.Errorf("MakeLiteralForSimpleType() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeLiteralForBlob(t *testing.T) {
	type args struct {
		path   storage.DataReference
		isDir  bool
		format string
	}
	tests := []struct {
		name string
		args args
		want *core.Blob
	}{
		{"simple-key", args{path: "/key", isDir: false, format: "xyz"}, &core.Blob{Uri: "/key", Metadata: &core.BlobMetadata{Type: &core.BlobType{Format: "xyz", Dimensionality: core.BlobType_SINGLE}}}},
		{"simple-dir", args{path: "/key", isDir: true, format: "xyz"}, &core.Blob{Uri: "/key", Metadata: &core.BlobMetadata{Type: &core.BlobType{Format: "xyz", Dimensionality: core.BlobType_MULTIPART}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MakeLiteralForBlob(tt.args.path, tt.args.isDir, tt.args.format); !reflect.DeepEqual(got.GetScalar().GetBlob(), tt.want) {
				t.Errorf("MakeLiteralForBlob() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeLiteralForType(t *testing.T) {
	t.Run("SimpleInteger", func(t *testing.T) {
		var literalType = &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}
		val, err := MakeLiteralForType(literalType, 1)
		assert.NoError(t, err)
		literalVal := &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_Integer{Integer: 1}}}}}}
		expectedVal, _ := ExtractFromLiteral(literalVal)
		actualVal, _ := ExtractFromLiteral(val)
		assert.Equal(t, expectedVal, actualVal)
	})

	t.Run("IntegerComingInAsFloatOverFlow", func(t *testing.T) {
		var literalType = &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}
		_, err := MakeLiteralForType(literalType, 8.888888e+19)
		assert.NotNil(t, err)
		numError := &strconv.NumError{
			Func: "ParseInt",
			Num:  "88888880000000000000",
			Err:  fmt.Errorf("value out of range"),
		}
		parseIntError := errors.WithMessage(numError, "failed to parse integer value")
		assert.Equal(t, errors.WithStack(parseIntError).Error(), err.Error())
	})

	t.Run("IntegerComingInAsFloat", func(t *testing.T) {
		var literalType = &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}
		val, err := MakeLiteralForType(literalType, 8.888888e+18)
		assert.NoError(t, err)
		literalVal := &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_Integer{Integer: 8.888888e+18}}}}}}
		expectedVal, _ := ExtractFromLiteral(literalVal)
		actualVal, _ := ExtractFromLiteral(val)
		assert.Equal(t, expectedVal, actualVal)
	})

	t.Run("SimpleFloat", func(t *testing.T) {
		var literalType = &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_FLOAT}}
		val, err := MakeLiteralForType(literalType, 1)
		assert.NoError(t, err)
		literalVal := &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_FloatValue{FloatValue: 1.0}}}}}}
		expectedVal, _ := ExtractFromLiteral(literalVal)
		actualVal, _ := ExtractFromLiteral(val)
		assert.Equal(t, expectedVal, actualVal)
	})

	t.Run("ArrayStrings", func(t *testing.T) {
		var literalType = &core.LiteralType{Type: &core.LiteralType_CollectionType{
			CollectionType: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}}}}
		strArray := []interface{}{"hello", "world"}
		val, err := MakeLiteralForType(literalType, strArray)
		assert.NoError(t, err)
		literalVal1 := &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_StringValue{StringValue: "hello"}}}}}}
		literalVal2 := &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_StringValue{StringValue: "world"}}}}}}
		literalCollection := []*core.Literal{literalVal1, literalVal2}
		literalVal := &core.Literal{Value: &core.Literal_Collection{Collection: &core.LiteralCollection{Literals: literalCollection}}}
		expectedVal, _ := ExtractFromLiteral(literalVal)
		actualVal, _ := ExtractFromLiteral(val)
		assert.Equal(t, expectedVal, actualVal)
	})

	t.Run("ArrayOfArrayStringsNotSupported", func(t *testing.T) {
		var literalType = &core.LiteralType{Type: &core.LiteralType_CollectionType{
			CollectionType: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}}}}
		strArrayOfArray := [][]interface{}{{"hello1", "world1"}, {"hello2", "world2"}}
		_, err := MakeLiteralForType(literalType, strArrayOfArray)
		expectedErrorf := fmt.Errorf("collection type expected but found [][]interface {}")
		assert.Equal(t, expectedErrorf, err)
	})

	t.Run("ArrayOfArrayStringsTypeErasure", func(t *testing.T) {
		var collectionType = &core.LiteralType{Type: &core.LiteralType_CollectionType{
			CollectionType: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}}}}
		var literalType = &core.LiteralType{Type: &core.LiteralType_CollectionType{
			CollectionType: collectionType}}

		createList1 := func() interface{} {
			return []interface{}{"hello1", "world1"}
		}
		createList2 := func() interface{} {
			return []interface{}{"hello2", "world2"}
		}
		createNestedList := func() interface{} {
			return []interface{}{createList1(), createList2()}
		}
		var strArrayOfArray = createNestedList()
		val, err := MakeLiteralForType(literalType, strArrayOfArray)
		assert.NoError(t, err)
		literalVal11 := &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_StringValue{StringValue: "hello1"}}}}}}
		literalVal12 := &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_StringValue{StringValue: "world1"}}}}}}
		literalCollection1Val := []*core.Literal{literalVal11, literalVal12}

		literalCollection1 := &core.Literal{Value: &core.Literal_Collection{Collection: &core.LiteralCollection{Literals: literalCollection1Val}}}

		literalVal21 := &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_StringValue{StringValue: "hello2"}}}}}}
		literalVal22 := &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_StringValue{StringValue: "world2"}}}}}}
		literalCollection2Val := []*core.Literal{literalVal21, literalVal22}
		literalCollection2 := &core.Literal{Value: &core.Literal_Collection{Collection: &core.LiteralCollection{Literals: literalCollection2Val}}}
		literalCollection := []*core.Literal{literalCollection1, literalCollection2}

		literalVal := &core.Literal{Value: &core.Literal_Collection{Collection: &core.LiteralCollection{Literals: literalCollection}}}
		expectedVal, _ := ExtractFromLiteral(literalVal)
		actualVal, _ := ExtractFromLiteral(val)
		assert.Equal(t, expectedVal, actualVal)
	})

	t.Run("MapStrings", func(t *testing.T) {
		var literalType = &core.LiteralType{Type: &core.LiteralType_MapValueType{
			MapValueType: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}}}}
		mapVal := map[string]interface{}{"hello1": "world1", "hello2": "world2"}
		val, err := MakeLiteralForType(literalType, mapVal)
		assert.NoError(t, err)
		literalVal1 := &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_StringValue{StringValue: "world1"}}}}}}
		literalVal2 := &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_StringValue{StringValue: "world2"}}}}}}
		literalMapVal := map[string]*core.Literal{"hello1": literalVal1, "hello2": literalVal2}
		literalVal := &core.Literal{Value: &core.Literal_Map{Map: &core.LiteralMap{Literals: literalMapVal}}}
		expectedVal, _ := ExtractFromLiteral(literalVal)
		actualVal, _ := ExtractFromLiteral(val)
		assert.Equal(t, expectedVal, actualVal)
	})

	t.Run("MapArrayOfStringsFail", func(t *testing.T) {
		var literalType = &core.LiteralType{Type: &core.LiteralType_MapValueType{
			MapValueType: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}}}}
		strArray := map[string][]interface{}{"hello1": {"world11", "world12"}, "hello2": {"world21", "world22"}}
		_, err := MakeLiteralForType(literalType, strArray)
		expectedErrorf := fmt.Errorf("map value types can only be of type map[string]interface{}, but found map[string][]interface {}")
		assert.Equal(t, expectedErrorf, err)
	})

	t.Run("MapArrayOfStringsTypeErasure", func(t *testing.T) {
		var collectionType = &core.LiteralType{Type: &core.LiteralType_CollectionType{
			CollectionType: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}}}}
		var literalType = &core.LiteralType{Type: &core.LiteralType_MapValueType{
			MapValueType: collectionType}}
		createList1 := func() interface{} {
			return []interface{}{"world11", "world12"}
		}
		createList2 := func() interface{} {
			return []interface{}{"world21", "world22"}
		}
		strArray := map[string]interface{}{"hello1": createList1(), "hello2": createList2()}
		val, err := MakeLiteralForType(literalType, strArray)
		assert.NoError(t, err)
		literalVal11 := &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_StringValue{StringValue: "world11"}}}}}}
		literalVal12 := &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_StringValue{StringValue: "world12"}}}}}}
		literalCollection1 := []*core.Literal{literalVal11, literalVal12}
		literalVal1 := &core.Literal{Value: &core.Literal_Collection{Collection: &core.LiteralCollection{Literals: literalCollection1}}}
		literalVal21 := &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_StringValue{StringValue: "world21"}}}}}}
		literalVal22 := &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_StringValue{StringValue: "world22"}}}}}}
		literalCollection2 := []*core.Literal{literalVal21, literalVal22}
		literalVal2 := &core.Literal{Value: &core.Literal_Collection{Collection: &core.LiteralCollection{Literals: literalCollection2}}}
		literalMapVal := map[string]*core.Literal{"hello1": literalVal1, "hello2": literalVal2}
		literalVal := &core.Literal{Value: &core.Literal_Map{Map: &core.LiteralMap{Literals: literalMapVal}}}
		expectedVal, _ := ExtractFromLiteral(literalVal)
		actualVal, _ := ExtractFromLiteral(val)
		assert.Equal(t, expectedVal, actualVal)
	})

	t.Run("Schema", func(t *testing.T) {
		var schemaColumns []*core.SchemaType_SchemaColumn
		schemaColumns = append(schemaColumns, &core.SchemaType_SchemaColumn{
			Name: "Price",
			Type: core.SchemaType_SchemaColumn_FLOAT,
		})
		var literalType = &core.LiteralType{Type: &core.LiteralType_Schema{Schema: &core.SchemaType{
			Columns: schemaColumns,
		}}}

		expectedLV := &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Schema{
				Schema: &core.Schema{
					Uri: "s3://blah/blah/blah",
					Type: &core.SchemaType{
						Columns: schemaColumns,
					},
				},
			},
		}}}
		lv, err := MakeLiteralForType(literalType, "s3://blah/blah/blah")
		assert.NoError(t, err)

		assert.Equal(t, expectedLV, lv)

		expectedVal, err := ExtractFromLiteral(expectedLV)
		assert.NoError(t, err)
		actualVal, err := ExtractFromLiteral(lv)
		assert.NoError(t, err)
		assert.Equal(t, expectedVal, actualVal)
	})

	t.Run("Blob", func(t *testing.T) {
		var literalType = &core.LiteralType{Type: &core.LiteralType_Blob{Blob: &core.BlobType{
			Dimensionality: core.BlobType_SINGLE,
		}}}
		expectedLV := &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Blob{
				Blob: &core.Blob{
					Uri: "s3://blah/blah/blah",
					Metadata: &core.BlobMetadata{
						Type: &core.BlobType{
							Dimensionality: core.BlobType_SINGLE,
						},
					},
				},
			},
		}}}
		lv, err := MakeLiteralForType(literalType, "s3://blah/blah/blah")
		assert.NoError(t, err)

		assert.Equal(t, expectedLV, lv)

		expectedVal, err := ExtractFromLiteral(expectedLV)
		assert.NoError(t, err)
		actualVal, err := ExtractFromLiteral(lv)
		assert.NoError(t, err)
		assert.Equal(t, expectedVal, actualVal)
	})

	t.Run("MultipartBlob", func(t *testing.T) {
		var literalType = &core.LiteralType{Type: &core.LiteralType_Blob{Blob: &core.BlobType{
			Dimensionality: core.BlobType_MULTIPART,
		}}}
		expectedLV := &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Blob{
				Blob: &core.Blob{
					Uri: "s3://blah/blah/blah",
					Metadata: &core.BlobMetadata{
						Type: &core.BlobType{
							Dimensionality: core.BlobType_MULTIPART,
						},
					},
				},
			},
		}}}
		lv, err := MakeLiteralForType(literalType, "s3://blah/blah/blah")
		assert.NoError(t, err)

		assert.Equal(t, expectedLV, lv)

		expectedVal, err := ExtractFromLiteral(expectedLV)
		assert.NoError(t, err)
		actualVal, err := ExtractFromLiteral(lv)
		assert.NoError(t, err)
		assert.Equal(t, expectedVal, actualVal)
	})

	t.Run("enumtype-nil", func(t *testing.T) {
		var literalType = &core.LiteralType{Type: &core.LiteralType_EnumType{EnumType: &core.EnumType{}}}
		_, err := MakeLiteralForType(literalType, nil)
		assert.Error(t, err)
		_, err = MakeLiteralForType(literalType, "")
		assert.Error(t, err)
	})

	t.Run("enumtype-happy", func(t *testing.T) {
		var literalType = &core.LiteralType{Type: &core.LiteralType_EnumType{EnumType: &core.EnumType{Values: []string{"x", "y", "z"}}}}
		expected := &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_StringValue{StringValue: "x"}}}}}}
		v, err := MakeLiteralForType(literalType, "x")
		assert.NoError(t, err)
		assert.Equal(t, expected, v)
		_, err = MakeLiteralForType(literalType, "")
		assert.Error(t, err)
	})

	t.Run("enumtype-illegal-val", func(t *testing.T) {
		var literalType = &core.LiteralType{Type: &core.LiteralType_EnumType{EnumType: &core.EnumType{Values: []string{"x", "y", "z"}}}}
		_, err := MakeLiteralForType(literalType, "m")
		assert.Error(t, err)
	})

	t.Run("Nil string", func(t *testing.T) {
		var literalType = &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}}
		l, err := MakeLiteralForType(literalType, nil)
		assert.NoError(t, err)
		assert.Equal(t, "", l.GetScalar().GetPrimitive().GetStringValue())
		l, err = MakeLiteralForType(literalType, "")
		assert.NoError(t, err)
		assert.Equal(t, "", l.GetScalar().GetPrimitive().GetStringValue())
	})

	t.Run("Structured Data Set", func(t *testing.T) {
		var dataSetColumns []*core.StructuredDatasetType_DatasetColumn
		dataSetColumns = append(dataSetColumns, &core.StructuredDatasetType_DatasetColumn{
			Name: "Price",
			LiteralType: &core.LiteralType{
				Type: &core.LiteralType_Simple{
					Simple: core.SimpleType_FLOAT,
				},
			},
		})
		var literalType = &core.LiteralType{Type: &core.LiteralType_StructuredDatasetType{StructuredDatasetType: &core.StructuredDatasetType{
			Columns: dataSetColumns,
			Format:  "testFormat",
		}}}

		expectedLV := &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_StructuredDataset{
				StructuredDataset: &core.StructuredDataset{
					Uri: "s3://blah/blah/blah",
					Metadata: &core.StructuredDatasetMetadata{
						StructuredDatasetType: &core.StructuredDatasetType{
							Columns: dataSetColumns,
							Format:  "testFormat",
						},
					},
				},
			},
		}}}
		lv, err := MakeLiteralForType(literalType, "s3://blah/blah/blah")
		assert.NoError(t, err)

		assert.Equal(t, expectedLV, lv)

		expectedVal, err := ExtractFromLiteral(expectedLV)
		assert.NoError(t, err)
		actualVal, err := ExtractFromLiteral(lv)
		assert.NoError(t, err)
		assert.Equal(t, expectedVal, actualVal)
	})

	t.Run("Union", func(t *testing.T) {
		var literalType = &core.LiteralType{
			Type: &core.LiteralType_UnionType{
				UnionType: &core.UnionType{
					Variants: []*core.LiteralType{
						{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}},
						{Type: &core.LiteralType_Simple{Simple: core.SimpleType_FLOAT}},
					},
				},
			},
		}
		expectedLV := &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Union{
				Union: &core.Union{
					Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_FLOAT}},
					Value: &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{
						Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_FloatValue{FloatValue: 0.1}}}}}},
				},
			},
		}}}
		lv, err := MakeLiteralForType(literalType, float64(0.1))
		assert.NoError(t, err)
		assert.Equal(t, expectedLV, lv)
		expectedVal, err := ExtractFromLiteral(expectedLV)
		assert.NoError(t, err)
		actualVal, err := ExtractFromLiteral(lv)
		assert.NoError(t, err)
		assert.Equal(t, expectedVal, actualVal)
	})
}
