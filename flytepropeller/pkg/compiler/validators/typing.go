package validators

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/santhosh-tekuri/jsonschema"
	"github.com/wI2L/jsondiff"
	jscmp "gitlab.com/yvesf/json-schema-compare"

	flyte "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

type typeChecker interface {
	CastsFrom(*flyte.LiteralType) bool
}

type trivialChecker struct {
	literalType *flyte.LiteralType
}

func isSuperTypeInJSON(sourceMetaData, targetMetaData *structpb.Struct) bool {
	// Check if the source schema is a supertype of the target schema, beyond simple inheritance.
	// For custom types, we expect the JSON schemas in the metadata to come from the same JSON schema package,
	// specifically draft 2020-12 from Mashumaro.

	srcSchemaBytes, _ := json.Marshal(sourceMetaData.GetFields())
	tgtSchemaBytes, _ := json.Marshal(targetMetaData.GetFields())

	compiler := jsonschema.NewCompiler()

	err := compiler.AddResource("src", bytes.NewReader(srcSchemaBytes))
	if err != nil {
		return false
	}
	err = compiler.AddResource("tgt", bytes.NewReader(tgtSchemaBytes))
	if err != nil {
		return false
	}

	srcSchema, _ := compiler.Compile("src")
	tgtSchema, _ := compiler.Compile("tgt")

	// Compare the two schemas
	errs := jscmp.Compare(tgtSchema, srcSchema)

	// Ignore not implemented errors in json-schema-compare (additionalProperties)
	// TODO: Explain why we use for loop here to check error msg
	for _, err := range errs {
		if !strings.Contains(err.Error(), "not implemented") {
			return false
		}
	}

	return true
}

func isSameTypeInJSON(sourceMetaData, targetMetaData *structpb.Struct) bool {
	srcSchemaBytes, err := json.Marshal(sourceMetaData.GetFields())
	if err != nil {
		logger.Infof(context.Background(), "Failed to marshal source metadata: %v", err)
		return false
	}

	tgtSchemaBytes, err := json.Marshal(targetMetaData.GetFields())
	if err != nil {
		logger.Infof(context.Background(), "Failed to marshal target metadata: %v", err)
		return false
	}

	// Use jsondiff to compare the two schemas
	patch, err := jsondiff.CompareJSON(srcSchemaBytes, tgtSchemaBytes)
	if err != nil {
		logger.Infof(context.Background(), "Failed to compare JSON schemas: %v", err)
		return false
	}

	// If the length of the patch is zero, the two JSON structures are identical
	return len(patch) == 0
}

// CastsFrom is a trivial type checker merely checks if types match exactly.
func (t trivialChecker) CastsFrom(upstreamType *flyte.LiteralType) bool {
	// If upstream is an enum, it can be consumed as a string downstream
	if upstreamType.GetEnumType() != nil {
		if t.literalType.GetSimple() == flyte.SimpleType_STRING {
			return true
		}
	}
	// If t is an enum, it can be created from a string as Enums as just constrained String aliases
	if t.literalType.GetEnumType() != nil {
		if upstreamType.GetSimple() == flyte.SimpleType_STRING {
			return true
		}
	}

	if GetTagForType(upstreamType) != "" && GetTagForType(t.literalType) != GetTagForType(upstreamType) {
		return false
	}

	// Related Issue: https://github.com/flyteorg/flyte/issues/5489
	// RFC: https://github.com/flyteorg/flyte/blob/master/rfc/system/5741-binary-idl-with-message-pack.md#flytepropeller
	if upstreamType.GetSimple() == flyte.SimpleType_STRUCT && t.literalType.GetSimple() == flyte.SimpleType_STRUCT {
		// Json Schema is stored in Metadata
		upstreamMetaData := upstreamType.GetMetadata()
		downstreamMetaData := t.literalType.GetMetadata()

		// There's bug in flytekit's dataclass Transformer to generate JSON Scheam before,
		// in some case, we the JSON Schema will be nil, so we can only pass it to support
		// backward compatible. (reference task should be supported.)
		if upstreamMetaData == nil || downstreamMetaData == nil {
			return true
		}

		return isSameTypeInJSON(upstreamMetaData, downstreamMetaData) || isSuperTypeInJSON(upstreamMetaData, downstreamMetaData)
	}

	// Ignore metadata when comparing types.
	upstreamTypeCopy := *upstreamType
	downstreamTypeCopy := *t.literalType
	upstreamTypeCopy.Structure = &flyte.TypeStructure{}
	downstreamTypeCopy.Structure = &flyte.TypeStructure{}
	upstreamTypeCopy.Metadata = &structpb.Struct{}
	downstreamTypeCopy.Metadata = &structpb.Struct{}
	upstreamTypeCopy.Annotation = &flyte.TypeAnnotation{}
	downstreamTypeCopy.Annotation = &flyte.TypeAnnotation{}
	return upstreamTypeCopy.String() == downstreamTypeCopy.String()
}

type noneTypeChecker struct{}

// CastsFrom matches only void
func (t noneTypeChecker) CastsFrom(upstreamType *flyte.LiteralType) bool {
	return isNoneType(upstreamType)
}

type mapTypeChecker struct {
	literalType *flyte.LiteralType
}

// CastsFrom checks that the target map type can be cast to the current map type. We need to ensure both the key types
// and value types match.
func (t mapTypeChecker) CastsFrom(upstreamType *flyte.LiteralType) bool {
	// Empty maps should match any collection.
	mapLiteralType := upstreamType.GetMapValueType()
	if isNoneType(mapLiteralType) {
		return true
	} else if mapLiteralType != nil {
		return getTypeChecker(t.literalType.GetMapValueType()).CastsFrom(mapLiteralType)
	}

	return false
}

type blobTypeChecker struct {
	literalType *flyte.LiteralType
}

// CastsFrom checks that the target blob type can be cast to the current blob type. When the blob has no format
// specified, it accepts all blob inputs since it is generic.
func (t blobTypeChecker) CastsFrom(upstreamType *flyte.LiteralType) bool {
	blobType := upstreamType.GetBlob()
	if blobType == nil {
		return false
	}

	// Empty blobs should match any blob.
	if blobType.GetFormat() == "" || t.literalType.GetBlob().GetFormat() == "" {
		return true
	}

	return blobType.GetFormat() == t.literalType.GetBlob().GetFormat()
}

type collectionTypeChecker struct {
	literalType *flyte.LiteralType
}

// CastsFrom checks whether two collection types match. We need to ensure that the nesting is correct and the final
// subtypes match.
func (t collectionTypeChecker) CastsFrom(upstreamType *flyte.LiteralType) bool {
	// Empty collections should match any collection.
	collectionType := upstreamType.GetCollectionType()
	if isNoneType(upstreamType.GetCollectionType()) {
		return true
	} else if collectionType != nil {
		return getTypeChecker(t.literalType.GetCollectionType()).CastsFrom(collectionType)
	}

	return false
}

type schemaTypeChecker struct {
	literalType *flyte.LiteralType
}

// CastsFrom handles type casting to the underlying schema type.
// Schemas are more complex types in the Flyte ecosystem. A schema is considered castable in the following
// cases.
//
//  1. The downstream schema has no column types specified.  In such a case, it accepts all schema input since it is
//     generic.
//
//  2. The downstream schema has a subset of the upstream columns and they match perfectly.
//
//  3. The upstream type can be Schema type or structured dataset type
func (t schemaTypeChecker) CastsFrom(upstreamType *flyte.LiteralType) bool {
	schemaType := upstreamType.GetSchema()
	structuredDatasetType := upstreamType.GetStructuredDatasetType()
	if structuredDatasetType == nil && schemaType == nil {
		return false
	}

	if schemaType != nil {
		return schemaCastFromSchema(schemaType, t.literalType.GetSchema())
	}

	// Flyte Schema can only be serialized to parquet
	if len(structuredDatasetType.GetFormat()) != 0 && !strings.EqualFold(structuredDatasetType.GetFormat(), "parquet") {
		return false
	}

	return schemaCastFromStructuredDataset(structuredDatasetType, t.literalType.GetSchema())
}

type structuredDatasetChecker struct {
	literalType *flyte.LiteralType
}

// CastsFrom for Structured dataset are more complex types in the Flyte ecosystem. A structured dataset is considered
// castable in the following cases:
//
//  1. The downstream structured dataset has no column types specified.  In such a case, it accepts all structured dataset input since it is
//     generic.
//
//  2. The downstream structured dataset has a subset of the upstream structured dataset columns and they match perfectly.
//
//  3. The upstream type can be Schema type or structured dataset type
func (t structuredDatasetChecker) CastsFrom(upstreamType *flyte.LiteralType) bool {
	// structured datasets are nullable
	if isNoneType(upstreamType) {
		return true
	}
	structuredDatasetType := upstreamType.GetStructuredDatasetType()
	schemaType := upstreamType.GetSchema()
	if structuredDatasetType == nil && schemaType == nil {
		return false
	}
	if schemaType != nil {
		// Flyte Schema can only be serialized to parquet
		format := t.literalType.GetStructuredDatasetType().GetFormat()
		if len(format) != 0 && !strings.EqualFold(format, "parquet") {
			return false
		}
		return structuredDatasetCastFromSchema(schemaType, t.literalType.GetStructuredDatasetType())
	}
	return structuredDatasetCastFromStructuredDataset(structuredDatasetType, t.literalType.GetStructuredDatasetType())
}

// Upstream (schema) -> downstream (schema)
func schemaCastFromSchema(upstream *flyte.SchemaType, downstream *flyte.SchemaType) bool {
	if len(upstream.GetColumns()) == 0 || len(downstream.GetColumns()) == 0 {
		return true
	}

	nameToTypeMap := make(map[string]flyte.SchemaType_SchemaColumn_SchemaColumnType)
	for _, column := range upstream.GetColumns() {
		nameToTypeMap[column.GetName()] = column.GetType()
	}

	// Check that the downstream schema is a strict sub-set of the upstream schema.
	for _, column := range downstream.GetColumns() {
		upstreamType, ok := nameToTypeMap[column.GetName()]
		if !ok {
			return false
		}
		if upstreamType != column.GetType() {
			return false
		}
	}
	return true
}

type unionTypeChecker struct {
	literalType *flyte.LiteralType
}

func (t unionTypeChecker) CastsFrom(upstreamType *flyte.LiteralType) bool {
	unionType := t.literalType.GetUnionType()

	upstreamUnionType := upstreamType.GetUnionType()
	if upstreamUnionType != nil {
		// For each upstream variant we must find a compatible downstream variant
		for _, u := range upstreamUnionType.GetVariants() {
			found := false
			for _, d := range unionType.GetVariants() {
				if AreTypesCastable(u, d) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}

		return true
	}

	// Matches iff we can unambiguously select a variant
	foundOne := false
	for _, x := range unionType.GetVariants() {
		if AreTypesCastable(upstreamType, x) {
			if foundOne {
				return false
			}
			foundOne = true
		}
	}

	return foundOne
}

// Upstream (structuredDatasetType) -> downstream (structuredDatasetType)
func structuredDatasetCastFromStructuredDataset(upstream *flyte.StructuredDatasetType, downstream *flyte.StructuredDatasetType) bool {
	// Skip the format check here when format is empty. https://github.com/flyteorg/flyte/issues/2864
	if len(upstream.GetFormat()) != 0 && len(downstream.GetFormat()) != 0 && !strings.EqualFold(upstream.GetFormat(), downstream.GetFormat()) {
		return false
	}

	if len(upstream.GetColumns()) == 0 || len(downstream.GetColumns()) == 0 {
		return true
	}

	nameToTypeMap := make(map[string]*flyte.LiteralType)
	for _, column := range upstream.GetColumns() {
		nameToTypeMap[column.GetName()] = column.GetLiteralType()
	}

	// Check that the downstream structured dataset is a strict sub-set of the upstream structured dataset.
	for _, column := range downstream.GetColumns() {
		upstreamType, ok := nameToTypeMap[column.GetName()]
		if !ok {
			return false
		}
		if !getTypeChecker(column.GetLiteralType()).CastsFrom(upstreamType) {
			return false
		}
	}
	return true
}

// Upstream (schemaType) -> downstream (structuredDatasetType)
func structuredDatasetCastFromSchema(upstream *flyte.SchemaType, downstream *flyte.StructuredDatasetType) bool {
	if len(upstream.GetColumns()) == 0 || len(downstream.GetColumns()) == 0 {
		return true
	}
	nameToTypeMap := make(map[string]flyte.SchemaType_SchemaColumn_SchemaColumnType)
	for _, column := range upstream.GetColumns() {
		nameToTypeMap[column.GetName()] = column.GetType()
	}

	// Check that the downstream structuredDataset is a strict sub-set of the upstream schema.
	for _, column := range downstream.GetColumns() {
		upstreamType, ok := nameToTypeMap[column.GetName()]
		if !ok {
			return false
		}
		if !schemaTypeIsMatchStructuredDatasetType(upstreamType, column.GetLiteralType().GetSimple()) {
			return false
		}
	}
	return true
}

// Upstream (structuredDatasetType) -> downstream (schemaType)
func schemaCastFromStructuredDataset(upstream *flyte.StructuredDatasetType, downstream *flyte.SchemaType) bool {
	if len(upstream.GetColumns()) == 0 || len(downstream.GetColumns()) == 0 {
		return true
	}
	nameToTypeMap := make(map[string]flyte.SimpleType)
	for _, column := range upstream.GetColumns() {
		nameToTypeMap[column.GetName()] = column.GetLiteralType().GetSimple()
	}

	// Check that the downstream schema is a strict sub-set of the upstream structuredDataset.
	for _, column := range downstream.GetColumns() {
		upstreamType, ok := nameToTypeMap[column.GetName()]
		if !ok {
			return false
		}
		if !schemaTypeIsMatchStructuredDatasetType(column.GetType(), upstreamType) {
			return false
		}
	}
	return true
}

func schemaTypeIsMatchStructuredDatasetType(schemaType flyte.SchemaType_SchemaColumn_SchemaColumnType, structuredDatasetType flyte.SimpleType) bool {
	switch schemaType {
	case flyte.SchemaType_SchemaColumn_INTEGER:
		return structuredDatasetType == flyte.SimpleType_INTEGER
	case flyte.SchemaType_SchemaColumn_FLOAT:
		return structuredDatasetType == flyte.SimpleType_FLOAT
	case flyte.SchemaType_SchemaColumn_STRING:
		return structuredDatasetType == flyte.SimpleType_STRING
	case flyte.SchemaType_SchemaColumn_BOOLEAN:
		return structuredDatasetType == flyte.SimpleType_BOOLEAN
	case flyte.SchemaType_SchemaColumn_DATETIME:
		return structuredDatasetType == flyte.SimpleType_DATETIME
	case flyte.SchemaType_SchemaColumn_DURATION:
		return structuredDatasetType == flyte.SimpleType_DURATION
	}
	return false
}

func isNoneType(t *flyte.LiteralType) bool {
	switch t.GetType().(type) {
	case *flyte.LiteralType_Simple:
		return t.GetSimple() == flyte.SimpleType_NONE
	default:
		return false
	}
}

func getTypeChecker(t *flyte.LiteralType) typeChecker {
	switch t.GetType().(type) {
	case *flyte.LiteralType_CollectionType:
		return collectionTypeChecker{
			literalType: t,
		}
	case *flyte.LiteralType_MapValueType:
		return mapTypeChecker{
			literalType: t,
		}
	case *flyte.LiteralType_Blob:
		return blobTypeChecker{
			literalType: t,
		}
	case *flyte.LiteralType_Schema:
		return schemaTypeChecker{
			literalType: t,
		}
	case *flyte.LiteralType_UnionType:
		return unionTypeChecker{
			literalType: t,
		}
	case *flyte.LiteralType_StructuredDatasetType:
		return structuredDatasetChecker{
			literalType: t,
		}
	default:
		if isNoneType(t) {
			return noneTypeChecker{}
		}

		return trivialChecker{
			literalType: t,
		}
	}
}

func AreTypesCastable(upstreamType, downstreamType *flyte.LiteralType) bool {
	typeChecker := getTypeChecker(downstreamType)

	// if upstream is a singular union we check if the downstream type is castable from the union variant
	if upstreamType.GetUnionType() != nil && len(upstreamType.GetUnionType().GetVariants()) == 1 {
		variants := upstreamType.GetUnionType().GetVariants()
		if len(variants) == 1 && typeChecker.CastsFrom(variants[0]) {
			return true
		}
	}

	return typeChecker.CastsFrom(upstreamType)
}
