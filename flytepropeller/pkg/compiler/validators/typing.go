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
	if len(structuredDatasetType.Format) != 0 && !strings.EqualFold(structuredDatasetType.Format, "parquet") {
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
		format := t.literalType.GetStructuredDatasetType().Format
		if len(format) != 0 && !strings.EqualFold(format, "parquet") {
			return false
		}
		return structuredDatasetCastFromSchema(schemaType, t.literalType.GetStructuredDatasetType())
	}
	return structuredDatasetCastFromStructuredDataset(structuredDatasetType, t.literalType.GetStructuredDatasetType())
}

// Upstream (schema) -> downstream (schema)
func schemaCastFromSchema(upstream *flyte.SchemaType, downstream *flyte.SchemaType) bool {
	if len(upstream.Columns) == 0 || len(downstream.Columns) == 0 {
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
	if len(upstream.Format) != 0 && len(downstream.Format) != 0 && !strings.EqualFold(upstream.Format, downstream.Format) {
		return false
	}

	if len(upstream.Columns) == 0 || len(downstream.Columns) == 0 {
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
	if len(upstream.Columns) == 0 || len(downstream.Columns) == 0 {
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
	if len(upstream.Columns) == 0 || len(downstream.Columns) == 0 {
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
