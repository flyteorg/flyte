package converter

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
)

// makeVariableMap is a test helper that creates a VariableMap from a map of key->Variable
func makeVariableMap(variables map[string]*core.Variable) *core.VariableMap {
	entries := make([]*core.VariableEntry, 0, len(variables))
	for key, value := range variables {
		entries = append(entries, &core.VariableEntry{
			Key:   key,
			Value: value,
		})
	}
	return &core.VariableMap{Variables: entries}
}

func TestLiteralsToJsonSchema(t *testing.T) {
	t.Run("mismatched lengths, should pass and return json schema", func(t *testing.T) {
		literals := []*task.NamedLiteral{
			{Name: "test", Value: &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_StringValue{StringValue: "test_string value"}}}}}}},
		}
		variableMap := makeVariableMap(map[string]*core.Variable{
			"test":  {Description: "description", Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}}},
			"test2": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
		})

		resp, err := LiteralsToLaunchFormJson(context.Background(), literals, variableMap)
		require.NoError(t, err)

		// Check RSJF structure
		schema := resp.AsMap()
		assert.Equal(t, "object", schema["type"])
		assert.Contains(t, schema, "properties")

		// Check field in properties
		properties := schema["properties"].(map[string]any)
		testField := properties["test"].(map[string]any)
		assert.Equal(t, "string", testField["type"])
		assert.Equal(t, "test_string value", testField["default"])
		assert.Equal(t, "description", testField["description"])
	})

	t.Run("string literal", func(t *testing.T) {
		literals := []*task.NamedLiteral{
			{
				Name: "test_string",
				Value: &core.Literal{
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Primitive{
								Primitive: &core.Primitive{
									Value: &core.Primitive_StringValue{
										StringValue: "hello world",
									},
								},
							},
						},
					},
				},
			},
		}
		variableMap := makeVariableMap(map[string]*core.Variable{
			"test_string": {
				Type: &core.LiteralType{
					Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
				},
				Description: "A test string",
			},
		})

		result, err := LiteralsToLaunchFormJson(context.Background(), literals, variableMap)
		require.NoError(t, err)

		schema := result.AsMap()
		assert.Equal(t, "object", schema["type"])

		properties := schema["properties"].(map[string]any)
		testStringField := properties["test_string"].(map[string]any)
		assert.Equal(t, "string", testStringField["type"])
		assert.Equal(t, "hello world", testStringField["default"])
		assert.Equal(t, "A test string", testStringField["description"])
	})

	t.Run("integer literal", func(t *testing.T) {
		literals := []*task.NamedLiteral{
			{
				Name: "test_int",
				Value: &core.Literal{
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Primitive{
								Primitive: &core.Primitive{
									Value: &core.Primitive_Integer{
										Integer: 42,
									},
								},
							},
						},
					},
				},
			},
		}
		variableMap := makeVariableMap(map[string]*core.Variable{
			"test_int": {
				Type: &core.LiteralType{
					Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
				},
				Description: "A test integer",
			},
		})

		literalsToJsonResult, err := LiteralsToLaunchFormJson(context.Background(), literals, variableMap)
		require.NoError(t, err)

		schema := literalsToJsonResult.AsMap()
		assert.Equal(t, "object", schema["type"])

		properties := schema["properties"].(map[string]any)
		testIntField := properties["test_int"].(map[string]any)
		assert.Equal(t, "integer", testIntField["type"])
		assert.InDelta(t, float64(42), testIntField["default"], 0.001)
		assert.Equal(t, "A test integer", testIntField["description"])

		jsonToLiteralsResult, err := LaunchFormJsonToLiterals(context.Background(), literalsToJsonResult)
		if err != nil {
			t.Fatalf("failed to transform JSON to literals: %v", err)
		}

		// Validate the output
		assert.Len(t, jsonToLiteralsResult, 1)
		assert.Equal(t, "test_int", jsonToLiteralsResult[0].GetName())
		assert.NotNil(t, jsonToLiteralsResult[0].GetValue())
		assert.Equal(t, int64(42), jsonToLiteralsResult[0].GetValue().GetScalar().GetPrimitive().GetInteger())
	})

	t.Run("float literal", func(t *testing.T) {
		literals := []*task.NamedLiteral{
			{
				Name: "test_float",
				Value: &core.Literal{
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Primitive{
								Primitive: &core.Primitive{
									Value: &core.Primitive_FloatValue{
										FloatValue: 3.14,
									},
								},
							},
						},
					},
				},
			},
		}
		variableMap := makeVariableMap(map[string]*core.Variable{
			"test_float": {
				Type: &core.LiteralType{
					Type: &core.LiteralType_Simple{Simple: core.SimpleType_FLOAT},
				},
				Description: "A test float",
			},
		})

		literalsToJsonResult, err := LiteralsToLaunchFormJson(context.Background(), literals, variableMap)
		require.NoError(t, err)

		schema := literalsToJsonResult.AsMap()
		assert.Equal(t, "object", schema["type"])

		properties := schema["properties"].(map[string]any)
		testFloatField := properties["test_float"].(map[string]any)
		assert.Equal(t, "number", testFloatField["type"])
		assert.Equal(t, "float", testFloatField["format"])
		assert.InDelta(t, 3.14, testFloatField["default"], 0.001)
		assert.Equal(t, "A test float", testFloatField["description"])

		jsonToLiteralsResult, err := LaunchFormJsonToLiterals(context.Background(), literalsToJsonResult)
		if err != nil {
			t.Fatalf("failed to transform JSON to literals: %v", err)
		}

		// Validate the output
		assert.Len(t, jsonToLiteralsResult, 1)
		assert.Equal(t, "test_float", jsonToLiteralsResult[0].GetName())
		assert.NotNil(t, jsonToLiteralsResult[0].GetValue())
		assert.InDelta(t, 3.14, jsonToLiteralsResult[0].GetValue().GetScalar().GetPrimitive().GetFloatValue(), 0.001)
	})

	t.Run("boolean literal", func(t *testing.T) {
		literals := []*task.NamedLiteral{
			{
				Name: "test_bool",
				Value: &core.Literal{
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Primitive{
								Primitive: &core.Primitive{
									Value: &core.Primitive_Boolean{
										Boolean: true,
									},
								},
							},
						},
					},
				},
			},
		}
		variableMap := makeVariableMap(map[string]*core.Variable{
			"test_bool": {
				Type: &core.LiteralType{
					Type: &core.LiteralType_Simple{Simple: core.SimpleType_BOOLEAN},
				},
				Description: "A test boolean",
			},
		})

		literalsToJsonResult, err := LiteralsToLaunchFormJson(context.Background(), literals, variableMap)
		require.NoError(t, err)

		schema := literalsToJsonResult.AsMap()
		assert.Equal(t, "object", schema["type"])

		properties := schema["properties"].(map[string]any)
		testBoolField := properties["test_bool"].(map[string]any)
		assert.Equal(t, "boolean", testBoolField["type"])
		assert.Equal(t, true, testBoolField["default"])
		assert.Equal(t, "A test boolean", testBoolField["description"])

		jsonToLiteralsResult, err := LaunchFormJsonToLiterals(context.Background(), literalsToJsonResult)
		if err != nil {
			t.Fatalf("failed to transform JSON to literals: %v", err)
		}

		// Validate the output
		assert.Len(t, jsonToLiteralsResult, 1)
		assert.Equal(t, "test_bool", jsonToLiteralsResult[0].GetName())
		assert.NotNil(t, jsonToLiteralsResult[0].GetValue())
		assert.True(t, jsonToLiteralsResult[0].GetValue().GetScalar().GetPrimitive().GetBoolean())
	})

	t.Run("datetime literal", func(t *testing.T) {
		now := time.Now()
		literals := []*task.NamedLiteral{
			{
				Name: "test_datetime",
				Value: &core.Literal{
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Primitive{
								Primitive: &core.Primitive{
									Value: &core.Primitive_Datetime{
										Datetime: timestamppb.New(now),
									},
								},
							},
						},
					},
				},
			},
		}
		variableMap := makeVariableMap(map[string]*core.Variable{
			"test_datetime": {
				Type: &core.LiteralType{
					Type: &core.LiteralType_Simple{Simple: core.SimpleType_DATETIME},
				},
				Description: "A test datetime",
			},
		})

		literalsToJsonResult, err := LiteralsToLaunchFormJson(context.Background(), literals, variableMap)
		require.NoError(t, err)

		schema := literalsToJsonResult.AsMap()
		assert.Equal(t, "object", schema["type"])

		properties := schema["properties"].(map[string]any)
		testDatetimeField := properties["test_datetime"].(map[string]any)
		assert.Equal(t, "string", testDatetimeField["type"])
		assert.Equal(t, "datetime", testDatetimeField["format"])
		assert.Equal(t, now.UTC().Format(time.RFC3339), testDatetimeField["default"])
		assert.Equal(t, "A test datetime", testDatetimeField["description"])

		jsonToLiteralsResult, err := LaunchFormJsonToLiterals(context.Background(), literalsToJsonResult)
		if err != nil {
			t.Fatalf("failed to transform JSON to literals: %v", err)
		}

		// Validate the output
		assert.Len(t, jsonToLiteralsResult, 1)
		assert.Equal(t, "test_datetime", jsonToLiteralsResult[0].GetName())
		assert.NotNil(t, jsonToLiteralsResult[0].GetValue())
		resultTime := jsonToLiteralsResult[0].GetValue().GetScalar().GetPrimitive().GetDatetime().AsTime()
		assert.Equal(t, now.UTC().Format(time.RFC3339), resultTime.UTC().Format(time.RFC3339))
	})

	t.Run("duration literal", func(t *testing.T) {
		duration := time.Hour + 30*time.Minute
		literals := []*task.NamedLiteral{
			{
				Name: "test_duration",
				Value: &core.Literal{
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Primitive{
								Primitive: &core.Primitive{
									Value: &core.Primitive_Duration{
										Duration: durationpb.New(duration),
									},
								},
							},
						},
					},
				},
			},
		}
		variableMap := makeVariableMap(map[string]*core.Variable{
			"test_duration": {
				Type: &core.LiteralType{
					Type: &core.LiteralType_Simple{Simple: core.SimpleType_DURATION},
				},
				Description: "A test duration",
			},
		})

		literalsToJsonResult, err := LiteralsToLaunchFormJson(context.Background(), literals, variableMap)
		require.NoError(t, err)

		schema := literalsToJsonResult.AsMap()
		assert.Equal(t, "object", schema["type"])

		properties := schema["properties"].(map[string]any)
		testDurationField := properties["test_duration"].(map[string]any)
		assert.Equal(t, "string", testDurationField["type"])
		assert.Equal(t, "duration", testDurationField["format"])
		assert.Equal(t, "PT1H30M", testDurationField["default"])
		assert.Equal(t, "A test duration", testDurationField["description"])

		jsonToLiteralsResult, err := LaunchFormJsonToLiterals(context.Background(), literalsToJsonResult)
		if err != nil {
			t.Fatalf("failed to transform JSON to literals: %v", err)
		}

		// Validate the output
		assert.Len(t, jsonToLiteralsResult, 1)
		assert.Equal(t, "test_duration", jsonToLiteralsResult[0].GetName())
		assert.NotNil(t, jsonToLiteralsResult[0].GetValue())
		resultDuration := jsonToLiteralsResult[0].GetValue().GetScalar().GetPrimitive().GetDuration().AsDuration()
		assert.Equal(t, duration, resultDuration)
	})

	t.Run("generic struct", func(t *testing.T) {
		literals, variableMap := createSimpleDataclassLiteralSingleType()

		literalsToJsonResult, err := LiteralsToLaunchFormJson(context.Background(), literals, variableMap)
		require.NoError(t, err)

		schema := literalsToJsonResult.AsMap()
		assert.Equal(t, "object", schema["type"])

		properties := schema["properties"].(map[string]any)
		testStructField := properties["test_struct"].(map[string]any)
		assert.Equal(t, "object", testStructField["type"])
		assert.Equal(t, "A test struct", testStructField["description"])
		assert.Equal(t, map[string]interface{}{
			"field1": map[string]interface{}{
				"type": "string",
			},
		}, testStructField["properties"])
		assert.Equal(t, []interface{}{"field1"}, testStructField["required"])
		assert.Equal(t, map[string]interface{}{
			"field1": "value1",
		}, testStructField["default"])
	})
}

func TestLiteralTypeToJsonSchema(t *testing.T) {
	t.Run("nil literal type", func(t *testing.T) {
		result, err := literalTypeToJsonSchema(context.Background(), nil)
		require.NoError(t, err)
		assert.Equal(t, map[string]interface{}{"type": "null"}, result)
	})

	t.Run("simple types", func(t *testing.T) {
		testCases := []struct {
			simpleType core.SimpleType
			expected   map[string]interface{}
		}{
			{core.SimpleType_STRING, map[string]interface{}{"type": "string"}},
			{core.SimpleType_INTEGER, map[string]interface{}{"type": "integer"}},
			{core.SimpleType_FLOAT, map[string]interface{}{"type": "number", "format": "float"}},
			{core.SimpleType_BOOLEAN, map[string]interface{}{"type": "boolean"}},
			{core.SimpleType_DATETIME, map[string]interface{}{"type": "string", "format": "datetime"}},
			{core.SimpleType_DURATION, map[string]interface{}{"type": "string", "format": "duration"}},
			{core.SimpleType_BINARY, map[string]interface{}{"format": "binary", "type": "string"}},
			{core.SimpleType_ERROR, map[string]interface{}{"format": "error", "type": "string"}},
			{core.SimpleType_NONE, map[string]interface{}{"type": "null"}},
		}

		for _, tc := range testCases {
			t.Run(tc.simpleType.String(), func(t *testing.T) {
				literalType := &core.LiteralType{
					Type: &core.LiteralType_Simple{Simple: tc.simpleType},
				}
				result, err := literalTypeToJsonSchema(context.Background(), literalType)
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("blob type", func(t *testing.T) {
		literalType := &core.LiteralType{
			Type: &core.LiteralType_Blob{
				Blob: &core.BlobType{
					Dimensionality: core.BlobType_SINGLE,
					Format:         "csv",
				},
			},
		}
		result, err := literalTypeToJsonSchema(context.Background(), literalType)
		require.NoError(t, err)
		assert.Equal(t, "object", result["type"])
		assert.Equal(t, "blob", result["format"])

		properties := result["properties"].(map[string]any)
		assert.Contains(t, properties, "uri")
		assert.Contains(t, properties, "format")
		assert.Contains(t, properties, "dimensionality")

		formatProp := properties["format"].(map[string]any)
		assert.Equal(t, "csv", formatProp["default"])

		dimProp := properties["dimensionality"].(map[string]any)
		assert.Equal(t, "SINGLE", dimProp["default"])
	})

	t.Run("collection type", func(t *testing.T) {
		literalType := &core.LiteralType{
			Type: &core.LiteralType_CollectionType{
				CollectionType: &core.LiteralType{
					Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
				},
			},
		}
		result, err := literalTypeToJsonSchema(context.Background(), literalType)
		require.NoError(t, err)
		assert.Equal(t, "array", result["type"])
		items := result["items"].(map[string]any)
		assert.Equal(t, "string", items["type"])
	})

	t.Run("map type", func(t *testing.T) {
		literalType := &core.LiteralType{
			Type: &core.LiteralType_MapValueType{
				MapValueType: &core.LiteralType{
					Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
				},
			},
		}
		result, err := literalTypeToJsonSchema(context.Background(), literalType)
		require.NoError(t, err)
		assert.Equal(t, "object", result["type"])
		additionalProps := result["additionalProperties"].(map[string]any)
		assert.Equal(t, "integer", additionalProps["type"])
	})

	t.Run("enum type", func(t *testing.T) {
		literalType := &core.LiteralType{
			Type: &core.LiteralType_EnumType{
				EnumType: &core.EnumType{
					Values: []string{"RED", "GREEN", "BLUE"},
				},
			},
		}
		result, err := literalTypeToJsonSchema(context.Background(), literalType)
		require.NoError(t, err)
		assert.Equal(t, "string", result["type"])
		assert.Equal(t, []any{"RED", "GREEN", "BLUE"}, result["enum"])
	})
}

func TestJSONValuesToLiterals(t *testing.T) {
	t.Run("basic types", func(t *testing.T) {
		variableMap := makeVariableMap(map[string]*core.Variable{
			"str":  {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}}},
			"num":  {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
			"bool": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_BOOLEAN}}},
		})

		values, err := structpb.NewStruct(map[string]any{
			"str":  "hello",
			"num":  42,
			"bool": true,
		})
		require.NoError(t, err)

		literals, err := JSONValuesToLiterals(context.Background(), variableMap, values)
		require.NoError(t, err)
		assert.Len(t, literals, 3)

		// Find each literal by name and verify
		for _, lit := range literals {
			switch lit.Name {
			case "str":
				assert.Equal(t, "hello", lit.Value.GetScalar().GetPrimitive().GetStringValue())
			case "num":
				assert.Equal(t, int64(42), lit.Value.GetScalar().GetPrimitive().GetInteger())
			case "bool":
				assert.Equal(t, true, lit.Value.GetScalar().GetPrimitive().GetBoolean())
			}
		}
	})

	t.Run("missing required field", func(t *testing.T) {
		variableMap := makeVariableMap(map[string]*core.Variable{
			"required_field": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}}},
		})

		values, err := structpb.NewStruct(map[string]any{})
		require.NoError(t, err)

		_, err = JSONValuesToLiterals(context.Background(), variableMap, values)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing value for variable")
	})

	t.Run("nil variableMap", func(t *testing.T) {
		values, err := structpb.NewStruct(map[string]any{"foo": "bar"})
		require.NoError(t, err)

		_, err = JSONValuesToLiterals(context.Background(), nil, values)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "variableMap cannot be nil")
	})

	t.Run("nil values", func(t *testing.T) {
		variableMap := makeVariableMap(map[string]*core.Variable{
			"test": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}}},
		})

		_, err := JSONValuesToLiterals(context.Background(), variableMap, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "values cannot be nil")
	})
}

func createSimpleDataclassLiteralSingleType() ([]*task.NamedLiteral, *core.VariableMap) {
	structValue, _ := structpb.NewStruct(map[string]any{"field1": "value1"})
	literals := []*task.NamedLiteral{
		{
			Name: "test_struct",
			Value: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Generic{
							Generic: structValue,
						},
					},
				},
			},
		},
	}

	metadata, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"field1": map[string]any{
				"type": "string",
			},
		},
		"required": []any{"field1"},
	})
	variableMap := makeVariableMap(map[string]*core.Variable{
		"test_struct": {
			Type: &core.LiteralType{
				Type:     &core.LiteralType_Simple{Simple: core.SimpleType_STRUCT},
				Metadata: metadata,
			},
			Description: "A test struct",
		},
	})

	return literals, variableMap
}

func TestFormatISO8601Duration(t *testing.T) {
	cases := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"zero", 0, "P0D"},
		{"seconds", 5 * time.Second, "PT5S"},
		{"minutes and seconds", 90 * time.Second, "PT1M30S"},
		{"hours and minutes", time.Hour + 30*time.Minute, "PT1H30M"},
		{"days", 48 * time.Hour, "P2D"},
		{"days hours minutes seconds", 49*time.Hour + 4*time.Minute + 5*time.Second, "P2DT1H4M5S"},
		{"fractional seconds", 1500 * time.Millisecond, "PT1.5S"},
		{"negative", -(time.Hour + 30*time.Minute), "-PT1H30M"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, formatISO8601Duration(tc.duration))
		})
	}
}

func TestParseISO8601Duration(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		cases := []struct {
			input    string
			expected time.Duration
		}{
			{"P0D", 0},
			{"PT5S", 5 * time.Second},
			{"PT1M30S", 90 * time.Second},
			{"PT1H30M", time.Hour + 30*time.Minute},
			{"P2D", 48 * time.Hour},
			{"P2DT1H4M5S", 49*time.Hour + 4*time.Minute + 5*time.Second},
			{"P1W", 7 * 24 * time.Hour},
			{"PT1.5S", 1500 * time.Millisecond},
			{"-PT1H30M", -(time.Hour + 30*time.Minute)},
			{"P1M", 30 * 24 * time.Hour},
			{"P1Y", 365 * 24 * time.Hour},
		}
		for _, tc := range cases {
			t.Run(tc.input, func(t *testing.T) {
				parsed, err := parseISO8601Duration(tc.input)
				require.NoError(t, err)
				assert.Equal(t, tc.expected, parsed.AsDuration())
			})
		}
	})

	t.Run("invalid", func(t *testing.T) {
		for _, input := range []string{"P", "PT", "", "2h30m", "not-a-duration", "P1H"} {
			_, err := parseISO8601Duration(input)
			assert.Error(t, err, "expected error for %q", input)
		}
	})

	t.Run("round trips with formatISO8601Duration", func(t *testing.T) {
		for _, d := range []time.Duration{0, 5 * time.Second, time.Hour + 30*time.Minute, 49*time.Hour + 4*time.Minute + 5*time.Second, 1500 * time.Millisecond} {
			parsed, err := parseISO8601Duration(formatISO8601Duration(d))
			require.NoError(t, err)
			assert.Equal(t, d, parsed.AsDuration())
		}
	})
}

func TestJSONValuesToLiteralsDuration(t *testing.T) {
	variableMap := makeVariableMap(map[string]*core.Variable{
		"duration": {
			Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_DURATION}},
		},
	})

	cases := []struct {
		name     string
		value    string
		expected time.Duration
	}{
		{"iso-8601", "P2DT3H", 51 * time.Hour},
		{"iso-8601 minutes", "PT30M", 30 * time.Minute},
		{"bare seconds (backward compat)", "3600", time.Hour},
		{"go-style readable (backward compat)", "2h30m", 2*time.Hour + 30*time.Minute},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			values, err := structpb.NewStruct(map[string]any{"duration": tc.value})
			require.NoError(t, err)

			literals, err := JSONValuesToLiterals(context.Background(), variableMap, values)
			require.NoError(t, err)
			require.Len(t, literals, 1)
			assert.Equal(t, tc.expected, literals[0].GetValue().GetScalar().GetPrimitive().GetDuration().AsDuration())
		})
	}

	t.Run("invalid iso-8601 errors", func(t *testing.T) {
		values, err := structpb.NewStruct(map[string]any{"duration": "P1H"})
		require.NoError(t, err)

		_, err = JSONValuesToLiterals(context.Background(), variableMap, values)
		assert.Error(t, err)
	})
}
