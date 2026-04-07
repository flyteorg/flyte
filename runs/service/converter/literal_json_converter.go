package converter

import (
	"context"
	"encoding/base64"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
)

var digitsOnlyRegex = regexp.MustCompile(`^\d+$`)

// This converter translates between JSON represntations and literals/variable maps/task specs. Here is an internal link
// to documentation surrounding the RSJF spec we are targeting:
// https://www.notion.so/Translating-Flyte-Literals-to-JSON-Schema-2098cc06513d80638f04e0a5920e6c4d?source=copy_link

// variableMapToMap converts a VariableMap (repeated VariableEntry) to a map for easier access.
func variableMapToMap(variableMap *core.VariableMap) map[string]*core.Variable {
	result := make(map[string]*core.Variable)
	if variableMap == nil {
		return result
	}
	for _, entry := range variableMap.GetVariables() {
		result[entry.GetKey()] = entry.GetValue()
	}
	return result
}

// TaskSpecToLaunchFormJson converts a TaskSpec to an RSJF-compliant JSON schema.
func TaskSpecToLaunchFormJson(ctx context.Context, taskSpec *task.TaskSpec) (*structpb.Struct, error) {
	properties := make(map[string]any)
	var required []any

	variableMap := variableMapToMap(taskSpec.GetTaskTemplate().GetInterface().GetInputs())

	for fieldName, variable := range variableMap {
		// Use helper function to build field schema without default value
		fieldSchema, shouldBeRequired, err := buildFieldSchemaFromVariable(ctx, fieldName, variable, nil)
		if err != nil {
			return nil, err
		}

		properties[fieldName] = fieldSchema

		if shouldBeRequired {
			required = append(required, fieldName)
		}
	}

	// TODO handle default values from task spec and merge in here

	// Build RSJF schema directly
	rootSchema := map[string]any{
		"type":       "object",
		"properties": properties,
		"required":   required,
		"title":      "Launch Form",
	}

	pbStruct, err := structpb.NewStruct(rootSchema)
	if err != nil {
		logger.Errorf(ctx, "error converting root schema to structpb: %v", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("error converting root schema to structpb: %v", err))
	}
	return pbStruct, nil
}

// JSONValuesToLiterals converts a raw JSON object (values) into NamedLiterals
// using the provided type definitions in variableMap. This is a thin wrapper
// around jsonValueToLiteralWithFieldName that understands Flyte's LiteralType.
//
//   - variableMap: describes the expected inputs (types, names)
//   - values: a Struct whose fields are the raw JSON values, e.g.:
//     {
//     "batch_index": 1,
//     "batch_info": { ... },
//     "output_context": {},
//     "queue_offset": 1
//     }
func JSONValuesToLiterals(ctx context.Context, variableMap *core.VariableMap, values *structpb.Struct) ([]*task.NamedLiteral, error) {
	if variableMap == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("variableMap cannot be nil"))
	}
	if values == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("values cannot be nil"))
	}

	valueMap := values.AsMap()
	varMap := variableMapToMap(variableMap)
	literals := make([]*task.NamedLiteral, 0, len(varMap))
	processedFields := make(map[string]bool)

	for name, variable := range varMap {
		rawValue, ok := valueMap[name]
		if !ok {
			// If the field is missing from the JSON payload, but the variable is an
			// optional union (i.e. includes a NONE variant), treat it as an explicit
			// NONE value instead of erroring.
			if unionType, ok := variable.GetType().GetType().(*core.LiteralType_UnionType); ok {
				nullVariant := findNullVariantInUnionType(unionType)
				if nullVariant != nil {
					literal := createNoneTypeUnionLiteralFromVariant(nullVariant)
					literals = append(literals, &task.NamedLiteral{
						Name:  name,
						Value: literal,
					})
					logger.Debugf(ctx, "created NONE type union literal for missing optional field '%s'", name)
					processedFields[name] = true
					continue
				}
			}

			// For non-optional (or misconfigured) variables, still treat a missing
			// value as an error.
			return nil, connect.NewError(
				connect.CodeInvalidArgument,
				fmt.Errorf("missing value for variable %q in JSON payload", name),
			)
		}

		literal, err := jsonValueToLiteralWithFieldName(ctx, rawValue, variable.GetType(), name)
		if err != nil {
			logger.Errorf(ctx, "error converting json value to literal for field '%s'. ValueType: [%T] Type: [%+v] Err: %v", name, rawValue, variable.GetType(), err)
			return nil, connect.NewError(
				connect.CodeInvalidArgument,
				fmt.Errorf("error converting field %q to literal: %w", name, err),
			)
		}

		// Allow null-like values to produce nil literals; skip them in the result.
		if literal == nil {
			logger.Debugf(ctx, "skipping nil literal for field '%s' (null-like value)", name)
			processedFields[name] = true
			continue
		}

		literals = append(literals, &task.NamedLiteral{
			Name:  name,
			Value: literal,
		})
		processedFields[name] = true
	}

	// Check for fields in raw data that don't map to any variable
	var unmappedFields []string
	for fieldName := range valueMap {
		if !processedFields[fieldName] {
			unmappedFields = append(unmappedFields, fieldName)
		}
	}

	if len(unmappedFields) > 0 {
		// Extra fields in values that don't have corresponding variables in variableMap
		// are allowed and will be ignored (with a warning). This is intentional to avoid
		// hard failures when translating literals for a task version that has a different
		// template than the original task these inputs were derived from.
		logger.Warnf(ctx, "found unmapped fields in JSON payload that do not correspond to any variable (ignoring): %v", unmappedFields)
	}

	return literals, nil
}

// isMapType checks if additionalProperties indicates this is a map type rather than a struct.
// Returns true if additionalProperties is a map schema object or boolean true.
func isMapType(additionalProperties any) bool {
	if additionalProperties == nil {
		return false
	}
	// Check if it's a map schema object (e.g., {"type": "string"})
	if _, ok := additionalProperties.(map[string]any); ok {
		return true
	}
	// Check if it's boolean true (allows arbitrary properties of any type)
	if boolVal, ok := additionalProperties.(bool); ok && boolVal {
		return true
	}
	return false
}

// collectDefaultsRecursively recursively collects default values from nested properties in a JSON schema.
// It returns a map of default values and a boolean indicating if any defaults were found.
func collectDefaultsRecursively(properties map[string]any) (map[string]any, bool) {
	defaultObj := make(map[string]any)
	hasAnyDefault := false

	for propName, propSchema := range properties {
		if propMap, ok := propSchema.(map[string]any); ok {
			// Check if this property has a direct default value
			if propDefault, hasPropDefault := propMap["default"]; hasPropDefault {
				defaultObj[propName] = propDefault
				hasAnyDefault = true
			} else {
				// Check if this property is a nested object that might have nested defaults
				schemaType, _ := propMap["type"].(string)
				_, hasFormat := propMap["format"].(string)
				isMap := isMapType(propMap["additionalProperties"])

				// If it's a struct type (object with properties, no additionalProperties map/true, no format)
				if schemaType == "object" && !hasFormat && !isMap {
					if nestedProperties, ok := propMap["properties"].(map[string]any); ok {
						// Recursively collect defaults from nested properties
						nestedDefaults, hasNestedDefaults := collectDefaultsRecursively(nestedProperties)
						if hasNestedDefaults {
							defaultObj[propName] = nestedDefaults
							hasAnyDefault = true
						}
					}
				}
			}
		}
	}

	return defaultObj, hasAnyDefault
}

// LaunchFormJsonToLiterals converts an RSJF JSON schema to a list of NamedLiterals.
func LaunchFormJsonToLiterals(ctx context.Context, jsonSchema *structpb.Struct) ([]*task.NamedLiteral, error) {
	schemaMap := jsonSchema.AsMap()

	// Extract properties from RSJF schema
	properties, ok := schemaMap["properties"].(map[string]any)
	if !ok {
		logger.Errorf(ctx, "missing or invalid properties in JSON schema: [%+v]", jsonSchema)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("missing or invalid properties in JSON schema"))
	}

	var literals []*task.NamedLiteral

	for fieldName, fieldSchemaInterface := range properties {
		fieldSchema, ok := fieldSchemaInterface.(map[string]any)
		if !ok {
			logger.Errorf(ctx, "invalid field schema for field %s: [%+v]", fieldName, fieldSchemaInterface)
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid field schema for field %s", fieldName))
		}

		// Add title back to schema for existing parsing logic
		fieldSchemaWithTitle := make(map[string]any, len(fieldSchema))
		for k, v := range fieldSchema {
			fieldSchemaWithTitle[k] = v
		}
		fieldSchemaWithTitle["title"] = fieldName

		// Extract title as name and default as value
		title := fieldName
		defaultValue, hasDefault := fieldSchemaWithTitle["default"]

		// For blob types, construct default from nested property defaults if top-level default is missing
		if !hasDefault {
			format, hasFormat := fieldSchemaWithTitle["format"].(string)
			if hasFormat && format == "blob" {
				if properties, ok := fieldSchemaWithTitle["properties"].(map[string]any); ok {
					// Construct default object from nested property defaults
					defaultObj := make(map[string]any)
					hasAnyDefault := false

					// Extract uri default
					if uriProp, ok := properties["uri"].(map[string]any); ok {
						if uriDefault, ok := uriProp["default"].(string); ok {
							defaultObj["uri"] = uriDefault
							hasAnyDefault = true
						}
					}

					// Extract format default
					if formatProp, ok := properties["format"].(map[string]any); ok {
						if formatDefault, ok := formatProp["default"].(string); ok {
							defaultObj["format"] = formatDefault
							hasAnyDefault = true
						}
					}

					// Extract dimensionality default (as string name "SINGLE" or "MULTIPART")
					if dimProp, ok := properties["dimensionality"].(map[string]any); ok {
						if dimDefault, ok := dimProp["default"].(string); ok {
							defaultObj["dimensionality"] = dimDefault
							hasAnyDefault = true
						} else if dimDefault, ok := dimProp["default"].(float64); ok {
							// Handle legacy integer format - convert to enum name
							switch dimDefault {
							case 0:
								defaultObj["dimensionality"] = "SINGLE"
							case 1:
								defaultObj["dimensionality"] = "MULTIPART"
							default:
								defaultObj["dimensionality"] = "SINGLE" // Default
							}
							hasAnyDefault = true
						}
					}

					// If we found any defaults, use the constructed default object
					if hasAnyDefault {
						defaultValue = defaultObj
						hasDefault = true
					}
				}
			}
		}

		// For struct/dataclass types, construct default from nested property defaults if top-level default is missing
		if !hasDefault {
			schemaType, _ := fieldSchemaWithTitle["type"].(string)
			_, hasFormat := fieldSchemaWithTitle["format"].(string)
			isMap := isMapType(fieldSchemaWithTitle["additionalProperties"])
			// Check if this is a struct type (object with properties, no additionalProperties map/true, no format)
			if schemaType == "object" && !hasFormat && !isMap {
				if properties, ok := fieldSchemaWithTitle["properties"].(map[string]any); ok {
					// Recursively collect defaults from all nested properties
					defaultObj, hasAnyDefault := collectDefaultsRecursively(properties)

					// If we found any defaults, use the constructed default object
					if hasAnyDefault {
						defaultValue = defaultObj
						hasDefault = true
					}
				}
			}
		}

		// Convert schema back to literal type
		literalType, err := jsonSchemaToLiteralType(fieldSchemaWithTitle)
		if err != nil {
			logger.Errorf(ctx, "error converting json schema to literal type. Schema: [%+v] Err: %v", fieldSchemaWithTitle, err)
			return nil, err
		}

		// Check if this is an optional field (union type with null variant)
		isOptional := isOptionalUnionType(fieldSchemaWithTitle)

		// Convert default value to literal
		var literal *core.Literal
		if hasDefault {
			literal, err = jsonValueToLiteralWithFieldName(ctx, defaultValue, literalType, fieldName)
			if err != nil {
				logger.Errorf(ctx, "error converting default value to literal for field '%s'. Default value: [%v] Type: [%+v] Err: %v", fieldName, defaultValue, literalType, err)
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("error converting field '%s' to literal: %v", fieldName, err))
			}
			// Skip null values - if literal is nil, don't add it to the list
			if literal == nil {
				continue
			}
		} else {
			// For optional fields without a default, create a NoneType literal
			if isOptional {
				var err error
				literal, err = createNoneTypeLiteralForOptionalField(ctx, literalType, fieldName)
				if err != nil {
					return nil, err
				}
			} else {
				// For non-optional fields without defaults, skip them
				logger.Infof(ctx, "missing default value in schema, skipping field: [%+v]", fieldSchemaWithTitle)
				continue
			}
		}

		namedLiteral := &task.NamedLiteral{
			Name:  title,
			Value: literal,
		}

		literals = append(literals, namedLiteral)
	}
	return literals, nil
}

// LiteralsToLaunchFormJson converts a list of NamedLiterals accompanied by a VariableMap to RSJF-compliant JSON schema format.
func LiteralsToLaunchFormJson(ctx context.Context, literals []*task.NamedLiteral, variableMap *core.VariableMap) (*structpb.Struct, error) {
	properties := make(map[string]any)
	var required []any

	varMap := variableMapToMap(variableMap)

	for _, literal := range literals {
		variable, ok := varMap[literal.GetName()]
		if !ok {
			logger.Errorf(ctx, "variable not found in variable map for literal: %s", literal.GetName())
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("variable not found in variable map for literal: %s", literal.GetName()))
		}
		fieldName := literal.GetName()

		literalValue := literal.GetValue()
		value, err := literalToJsonValues(ctx, literalValue)
		if err != nil {
			logger.Errorf(ctx, "error converting literal to value. Literal: [%+v] Err: %v", literal, err)
			return nil, err
		}

		fieldSchema, shouldBeRequired, err := buildFieldSchemaFromVariable(ctx, fieldName, variable, value)
		if err != nil {
			return nil, err
		}

		properties[fieldName] = fieldSchema

		if shouldBeRequired {
			required = append(required, fieldName)
		}
	}

	// Build RSJF schema directly
	rootSchema := map[string]any{
		"type":       "object",
		"properties": properties,
		"required":   required,
		"title":      "Launch Form",
	}

	pbStruct, err := structpb.NewStruct(rootSchema)
	if err != nil {
		logger.Errorf(ctx, "error converting root schema to structpb: %v", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("error converting root schema to structpb: %v", err))
	}
	return pbStruct, nil
}

// buildFieldSchemaFromVariable creates a JSON schema field from a variable definition
func buildFieldSchemaFromVariable(ctx context.Context, fieldName string, variable *core.Variable, defaultValue any) (map[string]any, bool, error) {
	description := ""
	if variable.GetDescription() != "" && variable.GetDescription() != fieldName {
		description = variable.GetDescription()
	}

	literalType := variable.GetType()
	jsonType, err := literalTypeToJsonSchema(ctx, literalType)
	if err != nil {
		logger.Errorf(ctx, "error converting literal type to JSON schema. LiteralType: [%+v] Err: %v", literalType, err)
		return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("error converting literal type to JSON schema: %v", err))
	}

	// Build field schema
	fieldSchema := make(map[string]any)
	for k, v := range jsonType {
		fieldSchema[k] = v
	}

	// For blob types, put defaults in nested properties instead of top-level
	if format, hasFormat := jsonType["format"].(string); hasFormat && format == "blob" {
		if defaultValue != nil {
			if defaultMap, ok := defaultValue.(map[string]any); ok {
				// Extract defaults from the map and put them in nested properties
				if properties, ok := fieldSchema["properties"].(map[string]any); ok {
					// Set uri default if present
					if uri, hasURI := defaultMap["uri"].(string); hasURI {
						if uriProp, ok := properties["uri"].(map[string]any); ok {
							uriProp["default"] = uri
						}
					}
					// Set format default if present
					if format, hasFormat := defaultMap["format"].(string); hasFormat {
						if formatProp, ok := properties["format"].(map[string]any); ok {
							formatProp["default"] = format
						}
					}
					// Set dimensionality default if present (as string name "SINGLE" or "MULTIPART")
					if dimensionality, hasDim := defaultMap["dimensionality"]; hasDim {
						if dimProp, ok := properties["dimensionality"].(map[string]any); ok {
							// Convert to string name
							var dimValue string
							switch v := dimensionality.(type) {
							case string:
								// If already a string, check if it's a valid enum name or number string
								if v == "SINGLE" || v == "MULTIPART" {
									dimValue = v
								} else {
									// Try to parse as number and convert to enum name
									if dimInt, err := strconv.ParseInt(v, 10, 32); err == nil {
										switch dimInt {
										case 0:
											dimValue = "SINGLE"
										case 1:
											dimValue = "MULTIPART"
										default:
											dimValue = "SINGLE" // Default
										}
									} else {
										dimValue = "SINGLE" // Default
									}
								}
							case int32:
								switch v {
								case 0:
									dimValue = "SINGLE"
								case 1:
									dimValue = "MULTIPART"
								default:
									dimValue = "SINGLE" // Default
								}
							case int64:
								switch v {
								case 0:
									dimValue = "SINGLE"
								case 1:
									dimValue = "MULTIPART"
								default:
									dimValue = "SINGLE" // Default
								}
							case float64:
								switch v {
								case 0:
									dimValue = "SINGLE"
								case 1:
									dimValue = "MULTIPART"
								default:
									dimValue = "SINGLE" // Default
								}
							case int:
								switch v {
								case 0:
									dimValue = "SINGLE"
								case 1:
									dimValue = "MULTIPART"
								default:
									dimValue = "SINGLE" // Default
								}
							default:
								dimValue = "SINGLE" // Default
							}
							dimProp["default"] = dimValue
						}
					}
				}
			}
		}
	} else {
		// For non-blob types, add default value at top level if provided
		if defaultValue != nil {
			fieldSchema["default"] = defaultValue
		}
	}

	if description != "" {
		fieldSchema["description"] = description
	}

	// Determine if this field should be required
	shouldBeRequired := true

	// Check if default value indicates this should be optional
	if defaultValue != nil {
		// If default is a map with type: "null", it's optional
		if defaultMap, ok := defaultValue.(map[string]any); ok {
			if defaultMap["type"] == "null" {
				shouldBeRequired = false
			}
		}
	}

	// Check if it's a union type that includes null
	if oneOf, hasOneOf := jsonType["oneOf"]; hasOneOf {
		if oneOfSlice, ok := oneOf.([]any); ok {
			for _, variant := range oneOfSlice {
				if variantMap, ok := variant.(map[string]any); ok {
					if variantMap["type"] == "null" {
						shouldBeRequired = false
						break
					}
				}
			}
		}
	}

	return fieldSchema, shouldBeRequired, nil
}

func literalTypeToJsonSchema(ctx context.Context, literalType *core.LiteralType) (map[string]any, error) {
	if literalType == nil {
		// TODO validate the below is the correct case
		return map[string]any{"type": "null"}, nil
	}

	switch t := literalType.GetType().(type) {
	case *core.LiteralType_Simple:
		switch t.Simple {
		case core.SimpleType_NONE:
			return map[string]any{
				"type": "null",
			}, nil
		case core.SimpleType_STRING:
			return map[string]any{
				"type": "string",
			}, nil
		case core.SimpleType_INTEGER:
			return map[string]any{
				"type": "integer",
			}, nil
		case core.SimpleType_FLOAT:
			return map[string]any{
				"type":   "number",
				"format": "float",
			}, nil
		case core.SimpleType_BOOLEAN:
			return map[string]any{
				"type": "boolean",
			}, nil
		case core.SimpleType_DATETIME:
			return map[string]any{
				"type":   "string",
				"format": "datetime",
			}, nil
		case core.SimpleType_DURATION:
			return map[string]any{
				"type":   "string",
				"format": "duration",
			}, nil
		case core.SimpleType_STRUCT:
			// Pass through the metadata as JSON schema, no additional processing
			if literalType.GetMetadata() != nil {
				metadataMap := literalType.GetMetadata().AsMap()
				// Extract dataclass name from title if present
				dataclassName := metadataMap["title"]
				delete(metadataMap, "title")
				// Remove additionalProperties if present to avoid conflicts for generic structs
				delete(metadataMap, "additionalProperties")
				if dataclassName != nil {
					metadataMap["dataclass"] = dataclassName
				}
				return metadataMap, nil
			}
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("missing metadata for struct type"))
		case core.SimpleType_ERROR:
			return errorTypeToJsonSchema(), nil
		case core.SimpleType_BINARY:
			return map[string]any{
				"type":   "string",
				"format": "binary",
			}, nil
		default:
			logger.Errorf(ctx, "unknown simple type: %s", t.Simple)
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown simple type: %s", t.Simple))
		}
	case *core.LiteralType_Blob:
		// Build properties with defaults from BlobType if available
		properties := make(map[string]any)

		// URI property - always set a default (empty string if not set)
		uriProp := map[string]any{"type": "string"}
		uriProp["default"] = ""
		properties["uri"] = uriProp

		// Format property with default from BlobType if available, otherwise empty string
		formatProp := map[string]any{"type": "string"}
		if t.Blob != nil && t.Blob.GetFormat() != "" {
			formatProp["default"] = t.Blob.GetFormat()
		} else {
			formatProp["default"] = ""
		}
		properties["format"] = formatProp

		// Dimensionality property with default from BlobType if available
		// Return as string with enum values as "SINGLE" and "MULTIPART"
		dimProp := map[string]any{
			"type": "string",
			"enum": []any{"SINGLE", "MULTIPART"}, // Values are "SINGLE" (0) and "MULTIPART" (1)
		}
		if t.Blob != nil {
			// Convert integer enum to string name
			dimValue := t.Blob.GetDimensionality()
			switch dimValue {
			case core.BlobType_SINGLE:
				dimProp["default"] = "SINGLE"
			case core.BlobType_MULTIPART:
				dimProp["default"] = "MULTIPART"
			default:
				dimProp["default"] = "SINGLE" // Default to SINGLE
			}
		} else {
			dimProp["default"] = "SINGLE" // Default to SINGLE
		}
		properties["dimensionality"] = dimProp

		return map[string]any{
			"type":       "object",
			"format":     "blob",
			"properties": properties,
		}, nil
	case *core.LiteralType_CollectionType:
		literalType, err := literalTypeToJsonSchema(ctx, t.CollectionType)
		if err != nil {
			logger.Errorf(ctx, "error converting collection type to JSON schema. CollectionType: [%+v] Err: %v", t.CollectionType, err)
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("error converting collection type to JSON schema: %v", err))
		}
		return map[string]any{
			"type":  "array",
			"items": literalType,
		}, nil
	case *core.LiteralType_MapValueType:
		literalType, err := literalTypeToJsonSchema(ctx, t.MapValueType)
		if err != nil {
			logger.Errorf(ctx, "error converting map value type to JSON schema. MapValueType: [%+v] Err: %v", t.MapValueType, err)
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("error converting union variant to JSON schema: %v", err))
		}
		return map[string]any{
			"type":                 "object",
			"additionalProperties": literalType,
		}, nil
	case *core.LiteralType_EnumType:
		// Convert []string to []any for JSON schema compatibility
		values := t.EnumType.GetValues()
		enumValues := make([]any, len(values))
		for i, v := range values {
			enumValues[i] = v
		}
		return map[string]any{
			"type": "string",
			"enum": enumValues,
		}, nil
	case *core.LiteralType_StructuredDatasetType:
		return map[string]any{
			"type":   "object",
			"format": "structured-dataset",
			"properties": map[string]any{
				"uri":    map[string]any{"type": "string"},
				"format": map[string]any{"type": "string"},
			}}, nil
	case *core.LiteralType_UnionType:
		var oneOf []any
		for _, variant := range t.UnionType.GetVariants() {
			variantSchema, err := literalTypeToJsonSchema(ctx, variant)
			if err != nil {
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("error converting union variant to JSON schema: %v", err))
			}

			// Extract title from variant type
			title := extractTitleFromLiteralType(variant)
			variantSchema["title"] = title

			// Set structure tag, handling cases where Structure is nil
			structureTag := variant.GetStructure().GetTag()
			variantSchema["structure"] = structureTag

			oneOf = append(oneOf, variantSchema)
		}
		return map[string]any{
			"oneOf":  oneOf,
			"format": "union",
		}, nil
	case *core.LiteralType_Schema:
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("deprecated schema type"))
	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown type %s", t))
	}
}

func errorTypeToJsonSchema() map[string]any {
	return map[string]any{
		"type":   "string",
		"format": "error",
	}
}

// extractTitleFromLiteralType extracts a human-readable title for union variant schemas
func extractTitleFromLiteralType(literalType *core.LiteralType) string {
	if literalType == nil {
		return "unknown"
	}

	switch t := literalType.GetType().(type) {
	case *core.LiteralType_Simple:
		switch t.Simple {
		case core.SimpleType_STRING:
			return "string"
		case core.SimpleType_INTEGER:
			return "integer"
		case core.SimpleType_FLOAT:
			return "number"
		case core.SimpleType_BOOLEAN:
			return "boolean"
		case core.SimpleType_DATETIME:
			return "datetime"
		case core.SimpleType_DURATION:
			return "duration"
		case core.SimpleType_STRUCT:
			// For struct types, try to extract the dataclass name from metadata
			if title, ok := literalType.GetMetadata().AsMap()["title"]; ok {
				if titleStr, ok := title.(string); ok {
					return titleStr
				}
			}

			return "object"
		case core.SimpleType_BINARY:
			return "binary"
		case core.SimpleType_NONE:
			return "none"
		default:
			return "unknown"
		}
	case *core.LiteralType_CollectionType:
		return "array"
	case *core.LiteralType_MapValueType:
		return "object"
	case *core.LiteralType_Blob:
		return "blob"
	case *core.LiteralType_EnumType:
		return "enum"
	case *core.LiteralType_StructuredDatasetType:
		return "structured_dataset"
	case *core.LiteralType_UnionType:
		return "union"
	default:
		return "unknown"
	}
}

// isOptionalUnionType checks if a JSON schema represents an optional union type (union with null variant)
func isOptionalUnionType(schema map[string]any) bool {
	if oneOf, hasOneOf := schema["oneOf"]; hasOneOf {
		if oneOfSlice, ok := oneOf.([]any); ok {
			for _, variant := range oneOfSlice {
				if variantMap, ok := variant.(map[string]any); ok {
					if variantMap["type"] == "null" {
						return true
					}
				}
			}
		}
	}
	return false
}

// findNullVariantInUnionType finds the null variant type in a union literal type
func findNullVariantInUnionType(unionType *core.LiteralType_UnionType) *core.LiteralType {
	if unionType == nil || unionType.UnionType == nil {
		return nil
	}
	for _, variant := range unionType.UnionType.GetVariants() {
		if variant.GetType() != nil {
			if simple, ok := variant.GetType().(*core.LiteralType_Simple); ok {
				if simple.Simple == core.SimpleType_NONE {
					return variant
				}
			}
		}
	}
	return nil
}

// createNoneTypeLiteralForOptionalField creates a NoneType literal for an optional field without a default value.
// It returns a union literal containing the NoneType value, or an error if the field is not properly configured.
func createNoneTypeLiteralForOptionalField(ctx context.Context, literalType *core.LiteralType, fieldName string) (*core.Literal, error) {
	// Find the null variant in the union type
	unionType, ok := literalType.GetType().(*core.LiteralType_UnionType)
	if !ok {
		logger.Errorf(ctx, "expected union type for optional field: %s", fieldName)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("expected union type for optional field: %s", fieldName))
	}

	nullVariant := findNullVariantInUnionType(unionType)
	if nullVariant == nil {
		logger.Errorf(ctx, "optional field but could not find null variant in union type: %s", fieldName)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("optional field but could not find null variant in union type: %s", fieldName))
	}

	literal := createNoneTypeUnionLiteralFromVariant(nullVariant)
	logger.Infof(ctx, "optional field with missing default value, creating NoneType literal: %s", fieldName)
	return literal, nil
}

// createNoneTypeUnionLiteralFromVariant creates a union literal containing a NoneType value from a given variant.
// This is used when converting null JSON values to union literals with NoneType.
func createNoneTypeUnionLiteralFromVariant(variant *core.LiteralType) *core.Literal {
	// Create a NoneType literal
	noneLiteral := &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_NoneType{
					NoneType: &core.Void{},
				},
			},
		},
	}

	// Set the structure tag from the variant type
	var tag string
	if structure := variant.GetStructure(); structure != nil {
		tag = structure.GetTag()
	}
	variantWithTag := &core.LiteralType{
		Type:      variant.GetType(),
		Structure: &core.TypeStructure{Tag: tag},
	}

	return &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Union{
					Union: &core.Union{
						Type:  variantWithTag,
						Value: noneLiteral,
					},
				},
			},
		},
	}
}

// parseUnionSchema converts a oneOf schema array to core.UnionType
func parseUnionSchema(oneOf any) (*core.LiteralType, error) {
	oneOfSlice, ok := oneOf.([]any)
	if !ok {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid oneOf schema format"))
	}

	var variants []*core.LiteralType
	for i, variant := range oneOfSlice {
		variantMap, ok := variant.(map[string]any)
		if !ok {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid union variant format at index %d", i))
		}

		// Preserve structure field before removing title (structure is needed for union variant matching)
		structureTag := ""
		if structure, ok := variantMap["structure"].(string); ok {
			structureTag = structure
		}

		// Remove title field before parsing as it's not part of the core schema
		delete(variantMap, "title")

		variantType, err := jsonSchemaToLiteralType(variantMap)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("error parsing union variant at index %d: %v", i, err))
		}

		// Ensure structure tag is set on the variant type
		if structureTag != "" && variantType.GetStructure() == nil {
			variantType.Structure = &core.TypeStructure{
				Tag: structureTag,
			}
		}

		variants = append(variants, variantType)
	}

	return &core.LiteralType{
		Type: &core.LiteralType_UnionType{
			UnionType: &core.UnionType{
				Variants: variants,
			},
		},
	}, nil
}

// findMatchingUnionVariant attempts to convert the value to each union variant and returns the first successful match
func findMatchingUnionVariant(ctx context.Context, value any, variants []*core.LiteralType, fieldName string) (*core.Literal, *core.LiteralType, error) {
	var lastErr error
	var variantErrors []string

	// Try each variant until we find one that works
	for i, variant := range variants {
		// Skip null variants if value is not null
		// Check if this is a NONE type variant by checking the actual type, not just GetSimple()
		// because GetSimple() returns 0 for non-simple types like MapValueType
		isNoneType := false
		switch variant.GetType().(type) {
		case *core.LiteralType_Simple:
			if variant.GetSimple() == core.SimpleType_NONE {
				isNoneType = true
			}
		}

		if isNoneType {
			if value == nil {
				// Try to create a NoneType literal for null values
				literal, err := jsonValueToLiteral(ctx, value, variant)
				if err == nil && literal != nil {
					return literal, variant, nil
				}
				// If NONE type conversion fails, track error and continue to next variant
				if err != nil {
					lastErr = err
					variantErrors = append(variantErrors, fmt.Sprintf("variant %d (null): %v", i, err))
				}
				continue
			} else {
				// Skip null variant for non-null values
				variantErrors = append(variantErrors, fmt.Sprintf("variant %d (null): skipped (value is not null)", i))
				continue
			}
		}

		literal, err := jsonValueToLiteral(ctx, value, variant)
		if err == nil && literal != nil {
			return literal, variant, nil
		}
		if err != nil {
			lastErr = err
			variantType := "unknown"
			// Use type switch to properly detect variant type, consistent with NONE type detection above
			switch variant.GetType().(type) {
			case *core.LiteralType_Simple:
				variantType = variant.GetSimple().String()
			default:
				if variant.GetMapValueType() != nil {
					variantType = "map"
				} else if variant.GetUnionType() != nil {
					variantType = "union"
				} else if variant.GetCollectionType() != nil {
					variantType = "array"
				} else if variant.GetBlob() != nil {
					variantType = "blob"
				} else if variant.GetEnumType() != nil {
					variantType = "enum"
				}
			}
			variantErrors = append(variantErrors, fmt.Sprintf("variant %d (%s): %v", i, variantType, err))
		}
	}

	// If no variant matches, return error with details
	errMsg := fmt.Sprintf("value %v does not match any union variant", value)
	if len(variantErrors) > 0 {
		errMsg += fmt.Sprintf(": %s", strings.Join(variantErrors, "; "))
	}
	if fieldName != "" {
		errMsg = fmt.Sprintf("field '%s': %s", fieldName, errMsg)
	}
	if lastErr != nil {
		return nil, nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("%s (last error: %v)", errMsg, lastErr))
	}
	return nil, nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("%s", errMsg))
}

// literalToJsonValues converts a core.Literal to a JSON-compatible value.
func literalToJsonValues(ctx context.Context, literal *core.Literal) (any, error) {
	if literal == nil {
		logger.Errorf(ctx, "literal cannot be nil")
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("literal cannot be nil"))
	}
	switch v := literal.GetValue().(type) {
	case *core.Literal_Scalar:
		switch s := v.Scalar.GetValue().(type) {
		case *core.Scalar_Primitive:
			switch p := s.Primitive.GetValue().(type) {
			case *core.Primitive_StringValue:
				return p.StringValue, nil
			case *core.Primitive_Integer:
				return p.Integer, nil
			case *core.Primitive_FloatValue:
				return p.FloatValue, nil
			case *core.Primitive_Boolean:
				return p.Boolean, nil
			case *core.Primitive_Datetime:
				if p.Datetime != nil {
					return p.Datetime.AsTime().Format(time.RFC3339), nil
				}
				logger.Errorf(ctx, "datetime cannot be nil")
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("datetime cannot be nil"))
			case *core.Primitive_Duration:
				if p.Duration != nil {
					return p.Duration.AsDuration().String(), nil
				}
				logger.Errorf(ctx, "duration cannot be nil")
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("duration cannot be nil"))
			default:
				logger.Errorf(ctx, "unknown primitive type: %T", p)
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown primitive type"))
			}
		case *core.Scalar_Binary:
			// Try to decode as msgpack, fallback to base64 encoded string
			if s.Binary != nil && len(s.Binary.GetValue()) > 0 {
				var decoded any
				err := msgpack.Unmarshal(s.Binary.GetValue(), &decoded)
				if err == nil {
					return decoded, nil
				}
				// If msgpack decoding fails, return as base64 encoded string
				return base64.StdEncoding.EncodeToString(s.Binary.GetValue()), nil
			}
			if s.Binary != nil {
				return base64.StdEncoding.EncodeToString(s.Binary.GetValue()), nil
			}
			logger.Errorf(ctx, "binary cannot be nil")
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("binary cannot be nil"))
		case *core.Scalar_Generic:
			if s.Generic != nil {
				return s.Generic.AsMap(), nil
			}
			logger.Errorf(ctx, "generic map cannot be nil")
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("generic cannot be nil"))
		case *core.Scalar_Union:
			return unionToValue(ctx, v)
		case *core.Scalar_Blob:
			// Convert dimensionality enum to string name
			dimValue := s.Blob.GetMetadata().GetType().GetDimensionality()
			var dimString string
			switch dimValue {
			case core.BlobType_SINGLE:
				dimString = "SINGLE"
			case core.BlobType_MULTIPART:
				dimString = "MULTIPART"
			default:
				dimString = "SINGLE" // Default to SINGLE
			}
			return map[string]any{
				"uri":            s.Blob.GetUri(),
				"format":         s.Blob.GetMetadata().GetType().GetFormat(),
				"dimensionality": dimString,
			}, nil
		case *core.Scalar_Schema:
			return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("unimplemented schema type"))
		case *core.Scalar_NoneType:
			return map[string]any{
				"type": "null",
			}, nil
		case *core.Scalar_Error:
			return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("unimplemented error type"))
		case *core.Scalar_StructuredDataset:
			return map[string]any{
				"uri":    s.StructuredDataset.GetUri(),
				"format": s.StructuredDataset.GetMetadata().GetStructuredDatasetType().GetFormat(),
			}, nil
		default:
			logger.Errorf(ctx, "unknown scalar type: %T", v)
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown scalar type"))
		}
	case *core.Literal_Collection:
		return collectionToValue(ctx, v)
	case *core.Literal_Map:
		return mapToValue(ctx, v)
	case *core.Literal_OffloadedMetadata:
		// TODO: implement offloaded metadata conversion
		return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("unimplemented offloaded metadata type"))
	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown literal type"))
	}
}

func mapToValue(ctx context.Context, literal *core.Literal_Map) (any, error) {
	if literal == nil || literal.Map == nil {
		logger.Infof(ctx, "map cannot be nil")
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("map cannot be nil"))
	}

	result := make(map[string]any)
	for key, value := range literal.Map.GetLiterals() {
		val, err := literalToJsonValues(ctx, value)
		if err != nil {
			logger.Errorf(ctx, "error converting map value for key '%s': %v", key, err)
			return nil, err
		}
		result[key] = val
	}
	return result, nil
}

func unionToValue(ctx context.Context, literal *core.Literal_Scalar) (any, error) {
	switch s := literal.Scalar.GetValue().(type) {
	case *core.Scalar_Union:
		if s.Union == nil {
			logger.Errorf(ctx, "union cannot be nil")
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("union cannot be nil"))
		}

		// Get the underlying value from the union
		unionValue := s.Union.GetValue()
		if unionValue == nil {
			logger.Errorf(ctx, "union value cannot be nil")
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("union value cannot be nil"))
		}

		// Convert the underlying literal to JSON value
		jsonValue, err := literalToJsonValues(ctx, unionValue)
		if err != nil {
			logger.Errorf(ctx, "error converting union value to json: %v", err)
			return nil, err
		}

		return jsonValue, nil

	default:
		logger.Errorf(ctx, "invalid union literal, expected Scalar_Union but got %T", s)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid union literal type"))
	}
}

func collectionToValue(ctx context.Context, collection *core.Literal_Collection) (any, error) {
	if collection == nil || collection.Collection == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("collection cannot be nil"))
	}

	if len(collection.Collection.GetLiterals()) == 0 {
		return []any{}, nil
	}
	var result []any
	for _, item := range collection.Collection.GetLiterals() {
		value, err := literalToJsonValues(ctx, item)
		if err != nil {
			logger.Info(ctx, "failed when converting collection item to value. Item: [%+v] Err: %v", item, err)
			return nil, err
		}
		result = append(result, value)
	}
	return result, nil
}

// jsonSchemaToLiteralType converts JSON schema back to core.LiteralType
func jsonSchemaToLiteralType(schema map[string]any) (*core.LiteralType, error) {
	// Check for union type first (has oneOf field)
	if oneOf, hasOneOf := schema["oneOf"]; hasOneOf {
		return parseUnionSchema(oneOf)
	}

	var err error
	literalType := &core.LiteralType{}
	// Pass through structure field if present, only relevant for union types
	if schema["structure"] != nil {
		if structure, ok := schema["structure"].(string); ok {
			literalType.Structure = &core.TypeStructure{
				Tag: structure,
			}
		}
	}

	schemaType, ok := schema["type"].(string)
	if !ok {
		logger.Infof(context.Background(), "missing 'type' in json schema: %+v", schema)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("missing 'type' in json"))
	}

	switch schemaType {
	case "string":
		// Check for enum property first
		if enumValues, hasEnum := schema["enum"]; hasEnum {
			if enumSlice, ok := enumValues.([]any); ok {
				// Convert []any to []string for EnumType
				values := make([]string, len(enumSlice))
				for i, v := range enumSlice {
					if str, ok := v.(string); ok {
						values[i] = str
					} else {
						// Convert non-string enum values to strings
						values[i] = fmt.Sprintf("%v", v)
					}
				}
				literalType.Type = &core.LiteralType_EnumType{
					EnumType: &core.EnumType{
						Values: values,
					},
				}
			} else {
				// If enum is not a slice, treat as regular string
				literalType.Type = &core.LiteralType_Simple{
					Simple: core.SimpleType_STRING,
				}
			}
		} else {
			// No enum property, check for format
			format, hasFormat := schema["format"].(string)
			if hasFormat {
				switch format {
				case "datetime":
					literalType.Type = &core.LiteralType_Simple{
						Simple: core.SimpleType_DATETIME,
					}
				case "duration":
					literalType.Type = &core.LiteralType_Simple{
						Simple: core.SimpleType_DURATION,
					}
				default:
					literalType.Type = &core.LiteralType_Simple{
						Simple: core.SimpleType_STRING,
					}
				}
			} else {
				literalType.Type = &core.LiteralType_Simple{
					Simple: core.SimpleType_STRING,
				}
			}
		}

	case "integer":
		// Handle datetime case by looking at format
		format, hasFormat := schema["format"].(string)
		if hasFormat && format == "datetime" {
			literalType.Type = &core.LiteralType_Simple{
				Simple: core.SimpleType_DATETIME,
			}
		} else if hasFormat && format == "duration" {
			literalType.Type = &core.LiteralType_Simple{
				Simple: core.SimpleType_DURATION,
			}
		} else {
			literalType.Type = &core.LiteralType_Simple{
				Simple: core.SimpleType_INTEGER,
			}
		}

	case "number":
		literalType.Type = &core.LiteralType_Simple{
			Simple: core.SimpleType_FLOAT,
		}
	case "boolean":
		literalType.Type = &core.LiteralType_Simple{
			Simple: core.SimpleType_BOOLEAN,
		}
	case "null":
		literalType.Type = &core.LiteralType_Simple{
			Simple: core.SimpleType_NONE,
		}
	case "array":
		items, ok := schema["items"].(map[string]any)
		if ok {
			itemType, err := jsonSchemaToLiteralType(items)
			if err != nil {
				return nil, err
			}
			literalType.Type = &core.LiteralType_CollectionType{
				CollectionType: itemType,
			}
		} else {
			// Default to string array if items not specified
			literalType.Type = &core.LiteralType_CollectionType{
				CollectionType: &core.LiteralType{
					Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
				},
			}
		}
	case "object":
		format, hasFormat := schema["format"].(string)
		if hasFormat {
			switch format {
			case "blob":
				blobType := &core.BlobType{}
				// Extract blob properties from schema if available
				if properties, ok := schema["properties"].(map[string]any); ok {
					// Extract format default value
					if formatProp, ok := properties["format"].(map[string]any); ok {
						if formatDefault, ok := formatProp["default"].(string); ok {
							blobType.Format = formatDefault
						}
					}
					// Extract dimensionality default value (as string name "SINGLE" or "MULTIPART")
					if dimProp, ok := properties["dimensionality"].(map[string]any); ok {
						if dimDefault, ok := dimProp["default"].(string); ok {
							// Parse as enum name or number string
							switch dimDefault {
							case "SINGLE":
								blobType.Dimensionality = core.BlobType_SINGLE
							case "MULTIPART":
								blobType.Dimensionality = core.BlobType_MULTIPART
							default:
								// Try to parse as number string for backward compatibility
								dimInt, err := strconv.ParseInt(dimDefault, 10, 32)
								if err == nil {
									blobType.Dimensionality = core.BlobType_BlobDimensionality(int32(dimInt))
								} else {
									blobType.Dimensionality = core.BlobType_SINGLE // Default
								}
							}
						} else if dimDefault, ok := dimProp["default"].(float64); ok {
							// Handle legacy integer format
							blobType.Dimensionality = core.BlobType_BlobDimensionality(int32(dimDefault))
						}
					}
				}
				literalType.Type = &core.LiteralType_Blob{
					Blob: blobType,
				}
			case "structured-dataset":
				literalType.Type = &core.LiteralType_StructuredDatasetType{
					StructuredDatasetType: &core.StructuredDatasetType{}, // Here we just need the type info to make an informed decision when parsing
				}
			}
		} else {
			// There are two cases where `object` is defined with no format: generic struct and maps
			// 1. Map: additionalProperties defined with a specific type
			// 2. Generic struct: no additionalProperties defined
			if t, ok := schema["additionalProperties"].(map[string]any); ok {
				mapValueType, err := jsonSchemaToLiteralType(t)
				if err != nil {
					return nil, err
				}
				literalType.Type = &core.LiteralType_MapValueType{
					MapValueType: mapValueType,
				}
			} else {
				// If no additionalProperties, treat as generic struct
				literalType.Type = &core.LiteralType_Simple{
					Simple: core.SimpleType_STRUCT,
				}
			}
		}

		// For struct types, pass through the metadata
		if metadata, err := structpb.NewStruct(schema); err == nil {
			literalType.Metadata = metadata
		}
	default:
		literalType.Type = &core.LiteralType_Simple{
			Simple: core.SimpleType_NONE,
		}
	}

	return literalType, err
}

// jsonValueToLiteral converts the json back to core.Literal based on the literal type
func jsonValueToLiteral(ctx context.Context, value any, literalType *core.LiteralType) (*core.Literal, error) {
	return jsonValueToLiteralWithFieldName(ctx, value, literalType, "")
}

// jsonValueToLiteralWithFieldName converts the json back to core.Literal based on the literal type, with field name for error context
func jsonValueToLiteralWithFieldName(ctx context.Context, value any, literalType *core.LiteralType, fieldName string) (*core.Literal, error) {
	// Check if value represents null (Go nil or {"type": "null"})
	isNullValue := false
	if value == nil {
		isNullValue = true
	} else if nullMap, ok := value.(map[string]any); ok {
		if nullType, hasType := nullMap["type"].(string); hasType && nullType == "null" {
			isNullValue = true
		}
	}

	// For union types, handle null values specially by creating a union literal with NoneType
	if isNullValue {
		if unionType, ok := literalType.GetType().(*core.LiteralType_UnionType); ok {
			// Find the NONE variant in the union
			nullVariant := findNullVariantInUnionType(unionType)
			if nullVariant != nil {
				return createNoneTypeUnionLiteralFromVariant(nullVariant), nil
			}
			// If no NONE variant found, fall through to error for non-optional unions
		}
		// For non-union types or unions without NONE variant, null is not allowed
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			fmt.Errorf("null value provided for non-optional variable %q", fieldName),
		)
	}

	switch t := literalType.GetType().(type) {
	case *core.LiteralType_Simple:
		switch t.Simple {
		case core.SimpleType_STRING:
			if str, ok := value.(string); ok {
				return &core.Literal{
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Primitive{
								Primitive: &core.Primitive{
									Value: &core.Primitive_StringValue{
										StringValue: str,
									},
								},
							},
						},
					},
				}, nil
			}
			logger.Errorf(ctx, "invalid string value type: %T, value: %v", value, value)
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid string value type: %T", value))
		case core.SimpleType_INTEGER:
			// Handle both int64 (from structpb) and float64 (from JSON unmarshaling)
			var intNum int64
			if num, ok := value.(int64); ok {
				intNum = num
			} else if num, ok := value.(float64); ok {
				intNum = int64(num)
			} else if num, ok := value.(int); ok {
				intNum = int64(num)
			} else {
				logger.Errorf(ctx, "invalid integer value type: %T, value: %v", value, value)
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid integer value type: %T", value))
			}
			return &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Primitive{
							Primitive: &core.Primitive{
								Value: &core.Primitive_Integer{
									Integer: intNum,
								},
							},
						},
					},
				},
			}, nil
		case core.SimpleType_FLOAT:
			// Handle both int64 (from structpb) and float64 (from JSON unmarshaling)
			var floatNum float64
			if num, ok := value.(float64); ok {
				floatNum = num
			} else if num, ok := value.(int64); ok {
				floatNum = float64(num)
			} else if num, ok := value.(int); ok {
				floatNum = float64(num)
			} else {
				logger.Errorf(ctx, "invalid float value type: %T, value: %v", value, value)
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid float value type: %T", value))
			}
			return &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Primitive{
							Primitive: &core.Primitive{
								Value: &core.Primitive_FloatValue{
									FloatValue: floatNum,
								},
							},
						},
					},
				},
			}, nil
		case core.SimpleType_BOOLEAN:
			if b, ok := value.(bool); ok {
				return &core.Literal{
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Primitive{
								Primitive: &core.Primitive{
									Value: &core.Primitive_Boolean{
										Boolean: b,
									},
								},
							},
						},
					},
				}, nil
			}
			logger.Errorf(ctx, "invalid boolean value type: %T, value: %v", value, value)
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid boolean value type: %T", value))
		case core.SimpleType_NONE:
			// For none type, create a none literal regardless of input value
			return &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_NoneType{
							NoneType: &core.Void{},
						},
					},
				},
			}, nil
		case core.SimpleType_STRUCT:
			// For structs, just passthrough the value without msgpack encoding
			if structValue, ok := value.(map[string]any); ok {
				metadata, err := structpb.NewStruct(structValue)
				if err != nil {
					logger.Errorf(ctx, "error creating structpb.Struct from value: %v, error: %v", structValue, err)
					return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("error creating struct from value: %v", err))
				}
				return &core.Literal{
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Generic{
								Generic: metadata,
							},
						},
					},
				}, nil
			}
			logger.Errorf(ctx, "invalid struct value type: %T, value: %v", value, value)
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid struct value type: %T, expected map[string]any", value))
		case core.SimpleType_DURATION:
			var dtn *durationpb.Duration
			if str, ok := value.(string); ok {
				// Check if the string contains only digits (parse as seconds)
				if digitsOnlyRegex.MatchString(str) {
					seconds, err := strconv.ParseInt(str, 10, 64)
					if err != nil {
						logger.Errorf(ctx, "error parsing duration seconds when converting json to literal. String: [%s] Err: %v", str, err)
						return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("error parsing duration seconds when converting json to literal"))
					}

					dtn = &durationpb.Duration{
						Seconds: seconds,
					}
				} else {
					// Parse the duration string into a time.Duration
					duration, err := time.ParseDuration(str)
					if err != nil {
						logger.Errorf(ctx, "error parsing duration when converting json to literal. String duration: [%s] Err: %v", str, err)
						return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("error parsing duration when converting json to literal"))
					}

					dtn = &durationpb.Duration{
						Seconds: int64(duration.Seconds()),
						Nanos:   int32(duration.Nanoseconds() % 1e9),
					}
				}

			} else if i, ok := value.(float64); ok {
				dtn = &durationpb.Duration{
					Seconds: int64(i),
				}
			} else {
				logger.Errorf(ctx, "error parsing duration type. Only string or float64 are supported.")
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid type for literal type: %v", reflect.TypeOf(value)))
			}

			return &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Primitive{
							Primitive: &core.Primitive{
								Value: &core.Primitive_Duration{
									Duration: dtn,
								},
							},
						},
					},
				},
			}, nil
		case core.SimpleType_DATETIME:
			if str, ok := value.(string); ok {
				// Parse the datetime string into a time.Time
				t, err := time.Parse(time.RFC3339, str)
				if err != nil {
					logger.Errorf(ctx, "error parsing datetime when converting json to literal. String time: [%s] Err: %v", str, err)
					return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("error parsing datetime when converting json to literal"))
				}
				ts := timestamppb.New(t)
				return &core.Literal{
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Primitive{
								Primitive: &core.Primitive{
									Value: &core.Primitive_Datetime{
										Datetime: ts,
									},
								},
							},
						},
					},
				}, nil
			}
		}
	case *core.LiteralType_CollectionType:
		if arr, ok := value.([]any); ok {
			var literals []*core.Literal
			for _, item := range arr {
				itemLiteral, err := jsonValueToLiteral(ctx, item, t.CollectionType)
				if err != nil {
					logger.Errorf(ctx, "error converting array item to literal. Item: [%+v] Err: %v", item, err)
					return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("error converting array item to literal"))
				}
				// Only append non-nil literals
				if itemLiteral != nil {
					literals = append(literals, itemLiteral)
				}
			}
			return &core.Literal{
				Value: &core.Literal_Collection{
					Collection: &core.LiteralCollection{
						Literals: literals,
					},
				},
			}, nil
		}
	case *core.LiteralType_MapValueType:
		if m, ok := value.(map[string]any); ok {
			literals := make(map[string]*core.Literal)
			for key, item := range m {
				itemLiteral, err := jsonValueToLiteral(ctx, item, t.MapValueType)
				if err != nil {
					logger.Errorf(ctx, "error converting map value to literal. Key: %s, Item: [%+v] Err: %v", key, item, err)
					return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("error converting map value to literal for key %s: %v", key, err))
				}
				// Only add non-nil literals to the map
				if itemLiteral != nil {
					literals[key] = itemLiteral
				}
			}
			return &core.Literal{
				Value: &core.Literal_Map{
					Map: &core.LiteralMap{
						Literals: literals,
					},
				},
			}, nil
		}
		logger.Errorf(ctx, "invalid map value type: %T, value: %v", value, value)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid map value type: %T, expected map[string]any", value))
	case *core.LiteralType_StructuredDatasetType:
		if m, ok := value.(map[string]any); ok {
			// Extract URI and format from the map
			uri, hasURI := m["uri"].(string)
			format, hasFormat := m["format"].(string)
			if hasURI && hasFormat {
				return &core.Literal{
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_StructuredDataset{
								StructuredDataset: &core.StructuredDataset{
									Uri: uri,
									Metadata: &core.StructuredDatasetMetadata{
										StructuredDatasetType: &core.StructuredDatasetType{
											Format: format,
										},
									},
								},
							},
						},
					},
				}, nil
			}
			// If URI or format is missing, return an empty literal
			logger.Errorf(ctx, "missing required fields for StructuredDataset type. URI: %v, Format: %v", hasURI, hasFormat)
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("missing required fields for StructuredDataset type"))
		}

	case *core.LiteralType_Blob:
		if m, ok := value.(map[string]any); ok {
			// Extract URI and format from the map
			uri, hasURI := m["uri"].(string)
			format, hasFormat := m["format"].(string)
			dimensionalityValue, hasDimensionality := m["dimensionality"]
			if hasURI && hasFormat && hasDimensionality {
				// Parse dimensionality as string name ("SINGLE" or "MULTIPART") or number (for backward compatibility)
				var dimensionality core.BlobType_BlobDimensionality
				switch v := dimensionalityValue.(type) {
				case string:
					// Try to parse as enum name first
					switch v {
					case "SINGLE":
						dimensionality = core.BlobType_SINGLE
					case "MULTIPART":
						dimensionality = core.BlobType_MULTIPART
					default:
						// Try to parse as number string for backward compatibility
						dimInt, err := strconv.ParseInt(v, 10, 32)
						if err != nil {
							logger.Errorf(ctx, "error parsing dimensionality string: %v", err)
							return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid dimensionality value: %v (expected SINGLE, MULTIPART, 0, or 1)", v))
						}
						dimensionality = core.BlobType_BlobDimensionality(int32(dimInt))
					}
				case float64:
					dimensionality = core.BlobType_BlobDimensionality(int32(v))
				case int32:
					dimensionality = core.BlobType_BlobDimensionality(v)
				case int64:
					dimensionality = core.BlobType_BlobDimensionality(int32(v))
				case int:
					dimensionality = core.BlobType_BlobDimensionality(int32(v))
				default:
					logger.Errorf(ctx, "invalid dimensionality type: %T", v)
					return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid dimensionality type: %T", v))
				}
				return &core.Literal{
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Blob{
								Blob: &core.Blob{
									Uri: uri,
									Metadata: &core.BlobMetadata{
										Type: &core.BlobType{
											Format:         format,
											Dimensionality: dimensionality,
										},
									},
								},
							},
						},
					},
				}, nil
			}
			// If URI, format, or dimensionality is missing, return an empty literal
			logger.Errorf(ctx, "missing required fields for Blob type. URI: %v, Format: %v, Dimensionality: %v", hasURI, hasFormat, hasDimensionality)
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("missing required fields for Blob type"))
		}
	case *core.LiteralType_UnionType:
		// For union types, we need to determine which variant matches the input value
		// and create a Union with the appropriate type and value
		unionLiteral, unionType, err := findMatchingUnionVariant(ctx, value, t.UnionType.GetVariants(), fieldName)
		if err != nil {
			if fieldName != "" {
				logger.Errorf(ctx, "error finding matching union variant for field '%s' with value [%v]: %v", fieldName, value, err)
			} else {
				logger.Errorf(ctx, "error finding matching union variant for value [%v]: %v", value, err)
			}
			return nil, err
		}

		// Set the structure tag from the matched variant type
		var tag string
		if structure := unionType.GetStructure(); structure != nil {
			tag = structure.GetTag()
		}
		unionType.Structure = &core.TypeStructure{
			Tag: tag,
		}

		return &core.Literal{
			Value: &core.Literal_Scalar{
				Scalar: &core.Scalar{
					Value: &core.Scalar_Union{
						Union: &core.Union{
							Type:  unionType,
							Value: unionLiteral,
						},
					},
				},
			},
		}, nil
	case *core.LiteralType_EnumType:
		// Enum types should be converted from string values
		if str, ok := value.(string); ok {
			enumValues := t.EnumType.GetValues()
			// Check if the string value is in the enum values
			for _, enumValue := range enumValues {
				if enumValue == str {
					return &core.Literal{
						Value: &core.Literal_Scalar{
							Scalar: &core.Scalar{
								Value: &core.Scalar_Primitive{
									Primitive: &core.Primitive{
										Value: &core.Primitive_StringValue{
											StringValue: str,
										},
									},
								},
							},
						},
					}, nil
				}
			}
			if fieldName != "" {
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("field '%s': enum value '%s' is not in allowed values: %v", fieldName, str, enumValues))
			}
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("enum value '%s' is not in allowed values: %v", str, enumValues))
		}
		if fieldName != "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("field '%s': invalid enum value type: %T, expected string", fieldName, value))
		}
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid enum value type: %T, expected string", value))
	case *core.LiteralType_Schema: // raise error
		if fieldName != "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("field '%s': schema is deprecated and cannot be converted to a literal", fieldName))
		}
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("schema is deprecated and cannot be converted to a literal"))
	}

	if fieldName != "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("field '%s': unsupported literal type: %v", fieldName, literalType))
	}
	return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unsupported literal type: %v", literalType))
}
