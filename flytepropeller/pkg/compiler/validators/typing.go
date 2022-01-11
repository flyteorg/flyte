package validators

import (
	"strings"

	flyte "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	structpb "github.com/golang/protobuf/ptypes/struct"
)

type typeChecker interface {
	CastsFrom(*flyte.LiteralType) bool
}

type trivialChecker struct {
	literalType *flyte.LiteralType
}

type voidChecker struct{}

type mapTypeChecker struct {
	literalType *flyte.LiteralType
}

type collectionTypeChecker struct {
	literalType *flyte.LiteralType
}

type schemaTypeChecker struct {
	literalType *flyte.LiteralType
}

type structuredDatasetChecker struct {
	literalType *flyte.LiteralType
}

// The trivial type checker merely checks if types match exactly.
func (t trivialChecker) CastsFrom(upstreamType *flyte.LiteralType) bool {
	// Everything is nullable currently
	if isVoid(upstreamType) {
		return true
	}

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

	// Ignore metadata when comparing types.
	upstreamTypeCopy := *upstreamType
	downstreamTypeCopy := *t.literalType
	upstreamTypeCopy.Metadata = &structpb.Struct{}
	downstreamTypeCopy.Metadata = &structpb.Struct{}
	return upstreamTypeCopy.String() == downstreamTypeCopy.String()
}

// The void type matches everything
func (t voidChecker) CastsFrom(upstreamType *flyte.LiteralType) bool {
	return true
}

// For a map type checker, we need to ensure both the key types and value types match.
func (t mapTypeChecker) CastsFrom(upstreamType *flyte.LiteralType) bool {
	// Maps are nullable
	if isVoid(upstreamType) {
		return true
	}

	mapLiteralType := upstreamType.GetMapValueType()
	if mapLiteralType != nil {
		return getTypeChecker(t.literalType.GetMapValueType()).CastsFrom(mapLiteralType)
	}
	return false
}

// For a collection type, we need to ensure that the nesting is correct and the final sub-types match.
func (t collectionTypeChecker) CastsFrom(upstreamType *flyte.LiteralType) bool {
	// Collections are nullable
	if isVoid(upstreamType) {
		return true
	}

	collectionType := upstreamType.GetCollectionType()
	if collectionType != nil {
		return getTypeChecker(t.literalType.GetCollectionType()).CastsFrom(collectionType)
	}
	return false
}

// Schemas are more complex types in the Flyte ecosystem. A schema is considered castable in the following
// cases.
//
//    1. The downstream schema has no column types specified.  In such a case, it accepts all schema input since it is
//       generic.
//
//    2. The downstream schema has a subset of the upstream columns and they match perfectly.
//
//    3. The upstream type can be Schema type or structured dataset type
//
func (t schemaTypeChecker) CastsFrom(upstreamType *flyte.LiteralType) bool {
	// Schemas are nullable
	if isVoid(upstreamType) {
		return true
	}

	schemaType := upstreamType.GetSchema()
	structuredDatasetType := upstreamType.GetStructuredDatasetType()
	if structuredDatasetType == nil && schemaType == nil {
		return false
	}

	if schemaType != nil {
		return schemaCastFromSchema(schemaType, t.literalType.GetSchema())
	}

	// Flyte Schema can only be serialized to parquet
	if !strings.EqualFold(structuredDatasetType.Format, "parquet") {
		return false
	}

	return schemaCastFromStructuredDataset(structuredDatasetType, t.literalType.GetSchema())
}

// Structured dataset are more complex types in the Flyte ecosystem. A structured dataset is considered castable in the following
// cases.
//
//    1. The downstream structured dataset has no column types specified.  In such a case, it accepts all structured dataset input since it is
//       generic.
//
//    2. The downstream structured dataset has a subset of the upstream structured dataset columns and they match perfectly.
//
//    3. The upstream type can be Schema type or structured dataset type
//
func (t structuredDatasetChecker) CastsFrom(upstreamType *flyte.LiteralType) bool {
	// structured datasets are nullable
	if isVoid(upstreamType) {
		return true
	}
	structuredDatasetType := upstreamType.GetStructuredDatasetType()
	schemaType := upstreamType.GetSchema()
	if structuredDatasetType == nil && schemaType == nil {
		return false
	}
	if schemaType != nil {
		// Flyte Schema can only be serialized to parquet
		if !strings.EqualFold(t.literalType.GetStructuredDatasetType().Format, "parquet") {
			return false
		}
		return structuredDatasetCastFromSchema(schemaType, t.literalType.GetStructuredDatasetType())
	}
	if !strings.EqualFold(structuredDatasetType.Format, t.literalType.GetStructuredDatasetType().Format) {
		return false
	}
	return structuredDatasetCastFromStructuredDataset(structuredDatasetType, t.literalType.GetStructuredDatasetType())
}

// Upstream (schema) -> downstream (schema)
func schemaCastFromSchema(upstream *flyte.SchemaType, downstream *flyte.SchemaType) bool {
	if len(downstream.Columns) == 0 {
		return true
	}

	nameToTypeMap := make(map[string]flyte.SchemaType_SchemaColumn_SchemaColumnType)
	for _, column := range upstream.Columns {
		nameToTypeMap[column.Name] = column.Type
	}

	// Check that the downstream schema is a strict sub-set of the upstream schema.
	for _, column := range downstream.Columns {
		upstreamType, ok := nameToTypeMap[column.Name]
		if !ok {
			return false
		}
		if upstreamType != column.Type {
			return false
		}
	}
	return true
}

// Upstream (structuredDatasetType) -> downstream (structuredDatasetType)
func structuredDatasetCastFromStructuredDataset(upstream *flyte.StructuredDatasetType, downstream *flyte.StructuredDatasetType) bool {
	if len(downstream.Columns) == 0 {
		return true
	}

	nameToTypeMap := make(map[string]*flyte.LiteralType)
	for _, column := range upstream.Columns {
		nameToTypeMap[column.Name] = column.LiteralType
	}

	// Check that the downstream structured dataset is a strict sub-set of the upstream structured dataset.
	for _, column := range downstream.Columns {
		upstreamType, ok := nameToTypeMap[column.Name]
		if !ok {
			return false
		}
		if !getTypeChecker(column.LiteralType).CastsFrom(upstreamType) {
			return false
		}
	}
	return true
}

// Upstream (schemaType) -> downstream (structuredDatasetType)
func structuredDatasetCastFromSchema(upstream *flyte.SchemaType, downstream *flyte.StructuredDatasetType) bool {
	if len(downstream.Columns) == 0 {
		return true
	}
	nameToTypeMap := make(map[string]flyte.SchemaType_SchemaColumn_SchemaColumnType)
	for _, column := range upstream.Columns {
		nameToTypeMap[column.Name] = column.GetType()
	}

	// Check that the downstream structuredDataset is a strict sub-set of the upstream schema.
	for _, column := range downstream.Columns {
		upstreamType, ok := nameToTypeMap[column.Name]
		if !ok {
			return false
		}
		if !schemaTypeIsMatchStructuredDatasetType(upstreamType, column.LiteralType.GetSimple()) {
			return false
		}
	}
	return true
}

// Upstream (structuredDatasetType) -> downstream (schemaType)
func schemaCastFromStructuredDataset(upstream *flyte.StructuredDatasetType, downstream *flyte.SchemaType) bool {
	if len(downstream.Columns) == 0 {
		return true
	}
	nameToTypeMap := make(map[string]flyte.SimpleType)
	for _, column := range upstream.Columns {
		nameToTypeMap[column.Name] = column.LiteralType.GetSimple()
	}

	// Check that the downstream schema is a strict sub-set of the upstream structuredDataset.
	for _, column := range downstream.Columns {
		upstreamType, ok := nameToTypeMap[column.Name]
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

func isVoid(t *flyte.LiteralType) bool {
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
	case *flyte.LiteralType_Schema:
		return schemaTypeChecker{
			literalType: t,
		}
	case *flyte.LiteralType_StructuredDatasetType:
		return structuredDatasetChecker{
			literalType: t,
		}
	default:
		if isVoid(t) {
			return voidChecker{}
		}
		return trivialChecker{
			literalType: t,
		}
	}
}

func AreTypesCastable(upstreamType, downstreamType *flyte.LiteralType) bool {
	return getTypeChecker(downstreamType).CastsFrom(upstreamType)
}
