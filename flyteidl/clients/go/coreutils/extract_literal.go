// extract_literal.go
// Utility methods to extract a native golang value from a given Literal.
// Usage:
// 1] string literal extraction
//    lit, _ := MakeLiteral("test_string")
//    val, _ := ExtractFromLiteral(lit)
// 2] integer literal extraction. integer would be extracted in type int64.
//    lit, _ := MakeLiteral([]interface{}{1, 2, 3})
//    val, _ := ExtractFromLiteral(lit)
// 3] float literal extraction. float would be extracted in type float64.
//    lit, _ := MakeLiteral([]interface{}{1.0, 2.0, 3.0})
//    val, _ := ExtractFromLiteral(lit)
// 4] map of boolean literal extraction.
//    mapInstance := map[string]interface{}{
//			"key1": []interface{}{1, 2, 3},
//			"key2": []interface{}{5},
//		}
//	  lit, _ := MakeLiteral(mapInstance)
//    val, _ := ExtractFromLiteral(lit)
// For further examples check the test TestFetchLiteral in extract_literal_test.go

package coreutils

import (
	"fmt"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

func ExtractFromLiteral(literal *core.Literal) (interface{}, error) {
	switch literalValue := literal.Value.(type) {
	case *core.Literal_Scalar:
		switch scalarValue := literalValue.Scalar.Value.(type) {
		case *core.Scalar_Primitive:
			switch scalarPrimitive := scalarValue.Primitive.Value.(type) {
			case *core.Primitive_Integer:
				scalarPrimitiveInt := scalarPrimitive.Integer
				return scalarPrimitiveInt, nil
			case *core.Primitive_FloatValue:
				scalarPrimitiveFloat := scalarPrimitive.FloatValue
				return scalarPrimitiveFloat, nil
			case *core.Primitive_StringValue:
				scalarPrimitiveString := scalarPrimitive.StringValue
				return scalarPrimitiveString, nil
			case *core.Primitive_Boolean:
				scalarPrimitiveBoolean := scalarPrimitive.Boolean
				return scalarPrimitiveBoolean, nil
			case *core.Primitive_Datetime:
				scalarPrimitiveDateTime := scalarPrimitive.Datetime.AsTime()
				return scalarPrimitiveDateTime, nil
			case *core.Primitive_Duration:
				scalarPrimitiveDuration := scalarPrimitive.Duration.AsDuration()
				return scalarPrimitiveDuration, nil
			default:
				return nil, fmt.Errorf("unsupported literal scalar primitive type %T", scalarValue)
			}
		case *core.Scalar_Blob:
			return scalarValue.Blob.Uri, nil
		case *core.Scalar_Schema:
			return scalarValue.Schema.Uri, nil
		case *core.Scalar_Generic:
			return scalarValue.Generic, nil
		case *core.Scalar_StructuredDataset:
			return scalarValue.StructuredDataset.Uri, nil
		case *core.Scalar_Union:
			// extract the value of the union but not the actual union object
			extractedVal, err := ExtractFromLiteral(scalarValue.Union.Value)
			if err != nil {
				return nil, err
			}
			return extractedVal, nil
		case *core.Scalar_NoneType:
			return nil, nil
		default:
			return nil, fmt.Errorf("unsupported literal scalar type %T", scalarValue)
		}
	case *core.Literal_Collection:
		collectionValue := literalValue.Collection.Literals
		collection := make([]interface{}, len(collectionValue))
		for index, val := range collectionValue {
			if collectionElem, err := ExtractFromLiteral(val); err == nil {
				collection[index] = collectionElem
			} else {
				return nil, err
			}
		}
		return collection, nil
	case *core.Literal_Map:
		mapLiteralValue := literalValue.Map.Literals
		mapResult := make(map[string]interface{}, len(mapLiteralValue))
		for key, val := range mapLiteralValue {
			if val, err := ExtractFromLiteral(val); err == nil {
				mapResult[key] = val
			} else {
				return nil, err
			}
		}
		return mapResult, nil
	}
	return nil, fmt.Errorf("unsupported literal type %T", literal)
}
