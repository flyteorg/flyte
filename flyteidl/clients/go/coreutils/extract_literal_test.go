// extract_literal_test.go
// Test class for the utility methods which extract a native golang value from a flyte Literal.

package coreutils

import (
	"testing"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/stretchr/testify/assert"
)

func TestFetchLiteral(t *testing.T) {
	t.Run("Primitive", func(t *testing.T) {
		lit, err := MakeLiteral("test_string")
		assert.NoError(t, err)
		val, err := ExtractFromLiteral(lit)
		assert.NoError(t, err)
		assert.Equal(t, "test_string", val)
	})

	t.Run("Timestamp", func(t *testing.T) {
		now := time.Now().UTC()
		lit, err := MakeLiteral(now)
		assert.NoError(t, err)
		val, err := ExtractFromLiteral(lit)
		assert.NoError(t, err)
		assert.Equal(t, now, val)
	})

	t.Run("Duration", func(t *testing.T) {
		duration := time.Second * 10
		lit, err := MakeLiteral(duration)
		assert.NoError(t, err)
		val, err := ExtractFromLiteral(lit)
		assert.NoError(t, err)
		assert.Equal(t, duration, val)
	})

	t.Run("Array", func(t *testing.T) {
		lit, err := MakeLiteral([]interface{}{1, 2, 3})
		assert.NoError(t, err)
		val, err := ExtractFromLiteral(lit)
		assert.NoError(t, err)
		arr := []interface{}{int64(1), int64(2), int64(3)}
		assert.Equal(t, arr, val)
	})

	t.Run("Map", func(t *testing.T) {
		mapInstance := map[string]interface{}{
			"key1": []interface{}{1, 2, 3},
			"key2": []interface{}{5},
		}
		lit, err := MakeLiteral(mapInstance)
		assert.NoError(t, err)
		val, err := ExtractFromLiteral(lit)
		assert.NoError(t, err)
		expectedMapInstance := map[string]interface{}{
			"key1": []interface{}{int64(1), int64(2), int64(3)},
			"key2": []interface{}{int64(5)},
		}
		assert.Equal(t, expectedMapInstance, val)
	})

	t.Run("Map_Booleans", func(t *testing.T) {
		mapInstance := map[string]interface{}{
			"key1": []interface{}{true, false, true},
			"key2": []interface{}{false},
		}
		lit, err := MakeLiteral(mapInstance)
		assert.NoError(t, err)
		val, err := ExtractFromLiteral(lit)
		assert.NoError(t, err)
		assert.Equal(t, mapInstance, val)
	})

	t.Run("Map_Floats", func(t *testing.T) {
		mapInstance := map[string]interface{}{
			"key1": []interface{}{1.0, 2.0, 3.0},
			"key2": []interface{}{1.0},
		}
		lit, err := MakeLiteral(mapInstance)
		assert.NoError(t, err)
		val, err := ExtractFromLiteral(lit)
		assert.NoError(t, err)
		expectedMapInstance := map[string]interface{}{
			"key1": []interface{}{float64(1.0), float64(2.0), float64(3.0)},
			"key2": []interface{}{float64(1.0)},
		}
		assert.Equal(t, expectedMapInstance, val)
	})

	t.Run("NestedMap", func(t *testing.T) {
		mapInstance := map[string]interface{}{
			"key1": map[string]interface{}{"key11": 1.0, "key12": 2.0, "key13": 3.0},
			"key2": map[string]interface{}{"key21": 1.0},
		}
		lit, err := MakeLiteral(mapInstance)
		assert.NoError(t, err)
		val, err := ExtractFromLiteral(lit)
		assert.NoError(t, err)
		expectedMapInstance := map[string]interface{}{
			"key1": map[string]interface{}{"key11": float64(1.0), "key12": float64(2.0), "key13": float64(3.0)},
			"key2": map[string]interface{}{"key21": float64(1.0)},
		}
		assert.Equal(t, expectedMapInstance, val)
	})

	t.Run("Binary", func(t *testing.T) {
		s := MakeBinaryLiteral([]byte{'h'})
		assert.Equal(t, []byte{'h'}, s.GetScalar().GetBinary().GetValue())
		_, err := ExtractFromLiteral(s)
		assert.NotNil(t, err)
	})

	t.Run("NoneType", func(t *testing.T) {
		p, err := MakeLiteral(nil)
		assert.NoError(t, err)
		assert.NotNil(t, p.GetScalar())
		_, err = ExtractFromLiteral(p)
		assert.Nil(t, err)
	})

	t.Run("Generic", func(t *testing.T) {
		literalVal := map[string]interface{}{
			"x": 1,
			"y": "ystringvalue",
		}
		var literalType = &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRUCT}}
		lit, err := MakeLiteralForType(literalType, literalVal)
		assert.NoError(t, err)
		extractedLiteralVal, err := ExtractFromLiteral(lit)
		assert.NoError(t, err)
		fieldsMap := map[string]*structpb.Value{
			"x": {
				Kind: &structpb.Value_NumberValue{NumberValue: 1},
			},
			"y": {
				Kind: &structpb.Value_StringValue{StringValue: "ystringvalue"},
			},
		}
		expectedStructVal := &structpb.Struct{
			Fields: fieldsMap,
		}
		extractedStructValue := extractedLiteralVal.(*structpb.Struct)
		assert.Equal(t, len(expectedStructVal.Fields), len(extractedStructValue.Fields))
		for key, val := range expectedStructVal.Fields {
			assert.Equal(t, val.Kind, extractedStructValue.Fields[key].Kind)
		}
	})

	t.Run("Generic Passed As String", func(t *testing.T) {
		literalVal := "{\"x\": 1,\"y\": \"ystringvalue\"}"
		var literalType = &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRUCT}}
		lit, err := MakeLiteralForType(literalType, literalVal)
		assert.NoError(t, err)
		extractedLiteralVal, err := ExtractFromLiteral(lit)
		assert.NoError(t, err)
		fieldsMap := map[string]*structpb.Value{
			"x": {
				Kind: &structpb.Value_NumberValue{NumberValue: 1},
			},
			"y": {
				Kind: &structpb.Value_StringValue{StringValue: "ystringvalue"},
			},
		}
		expectedStructVal := &structpb.Struct{
			Fields: fieldsMap,
		}
		extractedStructValue := extractedLiteralVal.(*structpb.Struct)
		assert.Equal(t, len(expectedStructVal.Fields), len(extractedStructValue.Fields))
		for key, val := range expectedStructVal.Fields {
			assert.Equal(t, val.Kind, extractedStructValue.Fields[key].Kind)
		}
	})

	t.Run("Structured dataset", func(t *testing.T) {
		literalVal := "s3://blah/blah/blah"
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

		lit, err := MakeLiteralForType(literalType, literalVal)
		assert.NoError(t, err)
		extractedLiteralVal, err := ExtractFromLiteral(lit)
		assert.NoError(t, err)
		assert.Equal(t, literalVal, extractedLiteralVal)
	})

	t.Run("Union", func(t *testing.T) {
		literalVal := int64(1)
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
		lit, err := MakeLiteralForType(literalType, literalVal)
		assert.NoError(t, err)
		extractedLiteralVal, err := ExtractFromLiteral(lit)
		assert.NoError(t, err)
		assert.Equal(t, literalVal, extractedLiteralVal)
	})

	t.Run("Union with None", func(t *testing.T) {
		var literalType = &core.LiteralType{
			Type: &core.LiteralType_UnionType{
				UnionType: &core.UnionType{
					Variants: []*core.LiteralType{
						{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}},
						{Type: &core.LiteralType_Simple{Simple: core.SimpleType_NONE}},
					},
				},
			},
		}
		lit, err := MakeLiteralForType(literalType, nil)

		assert.NoError(t, err)
		extractedLiteralVal, err := ExtractFromLiteral(lit)
		assert.NoError(t, err)
		assert.Nil(t, extractedLiteralVal)
	})
}
