package validators

import (
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
func (t schemaTypeChecker) CastsFrom(upstreamType *flyte.LiteralType) bool {
	// Schemas are nullable
	if isVoid(upstreamType) {
		return true
	}

	schemaType := upstreamType.GetSchema()
	if schemaType == nil {
		return false
	}

	// If no columns are specified, this is a generic schema and it can accept any schema type.
	if len(t.literalType.GetSchema().Columns) == 0 {
		return true
	}

	nameToTypeMap := make(map[string]flyte.SchemaType_SchemaColumn_SchemaColumnType)
	for _, column := range schemaType.Columns {
		nameToTypeMap[column.Name] = column.Type
	}

	// Check that the downstream schema is a strict sub-set of the upstream schema.
	for _, column := range t.literalType.GetSchema().Columns {
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
