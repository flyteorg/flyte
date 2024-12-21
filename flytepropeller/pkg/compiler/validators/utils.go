package validators

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func containsBindingByVariableName(bindings []*core.Binding, name string) (found bool) {
	for _, b := range bindings {
		if b.GetVar() == name {
			return true
		}
	}

	return false
}

func findVariableByName(vars *core.VariableMap, name string) (variable *core.Variable, found bool) {
	if vars == nil || vars.Variables == nil {
		return nil, false
	}

	variable, found = vars.GetVariables()[name]
	return
}

// Gets literal type for scalar value. This can be used to compare the underlying type of two scalars for compatibility.
func literalTypeForScalar(scalar *core.Scalar) *core.LiteralType {
	// TODO: Should we just pass the type information with the value?  That way we don't have to guess?
	var literalType *core.LiteralType
	switch v := scalar.GetValue().(type) {
	case *core.Scalar_Primitive:
		literalType = literalTypeForPrimitive(scalar.GetPrimitive())
	case *core.Scalar_Blob:
		if scalar.GetBlob().GetMetadata() == nil {
			return nil
		}

		literalType = &core.LiteralType{Type: &core.LiteralType_Blob{Blob: scalar.GetBlob().GetMetadata().GetType()}}
	case *core.Scalar_Binary:
		// If the binary has a tag, treat it as a structured type (e.g., dict, dataclass, Pydantic BaseModel).
		// Otherwise, treat it as raw binary data.
		// Reference: https://github.com/flyteorg/flyte/blob/master/rfc/system/5741-binary-idl-with-message-pack.md
		if v.Binary.GetTag() == coreutils.MESSAGEPACK {
			literalType = &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRUCT}}
		} else {
			literalType = &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_BINARY}}
		}
	case *core.Scalar_Schema:
		literalType = &core.LiteralType{
			Type: &core.LiteralType_Schema{
				Schema: scalar.GetSchema().GetType(),
			},
		}
	case *core.Scalar_StructuredDataset:
		if v.StructuredDataset == nil || v.StructuredDataset.GetMetadata() == nil {
			return &core.LiteralType{
				Type: &core.LiteralType_StructuredDatasetType{},
			}
		}

		literalType = &core.LiteralType{
			Type: &core.LiteralType_StructuredDatasetType{
				StructuredDatasetType: scalar.GetStructuredDataset().GetMetadata().GetStructuredDatasetType(),
			},
		}
	case *core.Scalar_NoneType:
		literalType = &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_NONE}}
	case *core.Scalar_Error:
		literalType = &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_ERROR}}
	case *core.Scalar_Generic:
		literalType = &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRUCT}}
	case *core.Scalar_Union:
		literalType = &core.LiteralType{
			Type: &core.LiteralType_UnionType{
				UnionType: &core.UnionType{
					Variants: []*core.LiteralType{
						scalar.GetUnion().GetType(),
					},
				},
			},
		}
	default:
		return nil
	}

	return literalType
}

func literalTypeForPrimitive(primitive *core.Primitive) *core.LiteralType {
	simpleType := core.SimpleType_NONE
	switch primitive.GetValue().(type) {
	case *core.Primitive_Integer:
		simpleType = core.SimpleType_INTEGER
	case *core.Primitive_FloatValue:
		simpleType = core.SimpleType_FLOAT
	case *core.Primitive_StringValue:
		simpleType = core.SimpleType_STRING
	case *core.Primitive_Boolean:
		simpleType = core.SimpleType_BOOLEAN
	case *core.Primitive_Datetime:
		simpleType = core.SimpleType_DATETIME
	case *core.Primitive_Duration:
		simpleType = core.SimpleType_DURATION
	}

	return &core.LiteralType{Type: &core.LiteralType_Simple{Simple: simpleType}}
}

func buildVariablesIndex(params *core.VariableMap) (map[string]*core.Variable, sets.String) {
	paramMap := make(map[string]*core.Variable, len(params.GetVariables()))
	paramSet := sets.NewString()
	for paramName, param := range params.GetVariables() {
		paramMap[paramName] = param
		paramSet.Insert(paramName)
	}

	return paramMap, paramSet
}

func filterVariables(vars *core.VariableMap, varNames sets.String) *core.VariableMap {
	res := &core.VariableMap{
		Variables: make(map[string]*core.Variable, len(varNames)),
	}

	for paramName, param := range vars.GetVariables() {
		if varNames.Has(paramName) {
			res.Variables[paramName] = param
		}
	}

	return res
}

func withVariableName(param *core.Variable) (newParam *core.Variable, ok bool) {
	if raw, err := proto.Marshal(param); err == nil {
		newParam = &core.Variable{}
		if err = proto.Unmarshal(raw, newParam); err == nil {
			ok = true
		}
	}

	return
}

func UnionDistinctVariableMaps(m1, m2 map[string]*core.Variable) (map[string]*core.Variable, error) {
	res := make(map[string]*core.Variable, len(m1)+len(m2))
	for k, v := range m1 {
		res[k] = v
	}

	for k, v := range m2 {
		if existingV, exists := res[k]; exists {
			if v.GetType().String() != existingV.GetType().String() {
				return nil, fmt.Errorf("key already exists with a different type. %v has type [%v] on one side "+
					"and type [%v] on the other", k, existingV.GetType().String(), v.GetType().String())
			}
		}

		res[k] = v
	}

	return res, nil
}

// ValidateLiteralType check if the literal type is valid, return error if the literal is invalid.
func ValidateLiteralType(lt *core.LiteralType) error {
	if lt == nil {
		err := fmt.Errorf("got unknown literal type: [%v].\n"+
			"Suggested solution: Please update all your Flyte deployment images to the latest version and try again", lt)
		return err
	}
	if lt.GetCollectionType() != nil {
		return ValidateLiteralType(lt.GetCollectionType())
	}
	if lt.GetMapValueType() != nil {
		return ValidateLiteralType(lt.GetMapValueType())
	}

	return nil
}

type instanceChecker interface {
	isInstance(*core.Literal) bool
}

type trivialInstanceChecker struct {
	literalType *core.LiteralType
}

func (t trivialInstanceChecker) isInstance(lit *core.Literal) bool {
	if _, ok := lit.GetValue().(*core.Literal_Scalar); !ok {
		return false
	}
	targetType := t.literalType
	if targetType.GetEnumType() != nil {
		// If t is an enum, it can be created from a string as Enums as just constrained String aliases
		if _, ok := lit.GetScalar().GetPrimitive().GetValue().(*core.Primitive_StringValue); ok {
			return true
		}
	}

	literalType := literalTypeForScalar(lit.GetScalar())
	err := ValidateLiteralType(literalType)
	if err != nil {
		return false
	}
	return AreTypesCastable(literalType, targetType)
}

type noneInstanceChecker struct{}

func (t noneInstanceChecker) isInstance(lit *core.Literal) bool {
	if lit == nil {
		return true
	}
	_, ok := lit.GetScalar().GetValue().(*core.Scalar_NoneType)
	return ok
}

type collectionInstanceChecker struct {
	literalType *core.LiteralType
}

func (t collectionInstanceChecker) isInstance(lit *core.Literal) bool {
	if _, ok := lit.GetValue().(*core.Literal_Collection); !ok {
		return false
	}
	for _, x := range lit.GetCollection().GetLiterals() {
		if !IsInstance(x, t.literalType.GetCollectionType()) {
			return false
		}
	}
	return true
}

type mapInstanceChecker struct {
	literalType *core.LiteralType
}

func (t mapInstanceChecker) isInstance(lit *core.Literal) bool {
	if _, ok := lit.GetValue().(*core.Literal_Map); !ok {
		return false
	}
	for _, x := range lit.GetMap().GetLiterals() {
		if !IsInstance(x, t.literalType.GetMapValueType()) {
			return false
		}
	}
	return true
}

type blobInstanceChecker struct {
	literalType *core.LiteralType
}

func (t blobInstanceChecker) isInstance(lit *core.Literal) bool {
	if _, ok := lit.GetScalar().GetValue().(*core.Scalar_Blob); !ok {
		return false
	}

	blobType := lit.GetScalar().GetBlob().GetMetadata().GetType()
	if blobType == nil {
		return false
	}

	// Empty blobs should match any blob.
	if blobType.GetFormat() == "" || t.literalType.GetBlob().GetFormat() == "" {
		return true
	}

	return blobType.GetFormat() == t.literalType.GetBlob().GetFormat()
}

type schemaInstanceChecker struct {
	literalType *core.LiteralType
}

func (t schemaInstanceChecker) isInstance(lit *core.Literal) bool {
	if _, ok := lit.GetValue().(*core.Literal_Scalar); !ok {
		return false
	}
	scalar := lit.GetScalar()

	switch v := scalar.GetValue().(type) {
	case *core.Scalar_Schema:
		return schemaCastFromSchema(scalar.GetSchema().GetType(), t.literalType.GetSchema())
	case *core.Scalar_StructuredDataset:
		if v.StructuredDataset == nil || v.StructuredDataset.GetMetadata() == nil {
			return true
		}
		return schemaCastFromStructuredDataset(scalar.GetStructuredDataset().GetMetadata().GetStructuredDatasetType(), t.literalType.GetSchema())
	default:
		return false
	}
}

type structuredDatasetInstanceChecker struct {
	literalType *core.LiteralType
}

func (t structuredDatasetInstanceChecker) isInstance(lit *core.Literal) bool {
	if _, ok := lit.GetValue().(*core.Literal_Scalar); !ok {
		return false
	}
	scalar := lit.GetScalar()

	switch v := scalar.GetValue().(type) {
	case *core.Scalar_NoneType:
		return true
	case *core.Scalar_Schema:
		// Flyte Schema can only be serialized to parquet
		format := t.literalType.GetStructuredDatasetType().GetFormat()
		if len(format) != 0 && !strings.EqualFold(format, "parquet") {
			return false
		}
		return structuredDatasetCastFromSchema(scalar.GetSchema().GetType(), t.literalType.GetStructuredDatasetType())
	case *core.Scalar_StructuredDataset:
		if v.StructuredDataset == nil || v.StructuredDataset.GetMetadata() == nil {
			return true
		}
		return structuredDatasetCastFromStructuredDataset(scalar.GetStructuredDataset().GetMetadata().GetStructuredDatasetType(), t.literalType.GetStructuredDatasetType())
	default:
		return false
	}
}

type unionInstanceChecker struct {
	literalType *core.LiteralType
}

func (t unionInstanceChecker) isInstance(lit *core.Literal) bool {
	unionType := t.literalType.GetUnionType()

	if u := lit.GetScalar().GetUnion().GetType(); u != nil {
		found := false
		for _, d := range unionType.GetVariants() {
			if AreTypesCastable(u, d) {
				found = true
				break
			}
		}
		return found
	}

	// Matches iff we can unambiguously select a variant
	foundOne := false
	for _, x := range unionType.GetVariants() {
		if IsInstance(lit, x) {
			if foundOne {
				return false
			}
			foundOne = true
		}
	}

	return foundOne
}

func getInstanceChecker(t *core.LiteralType) instanceChecker {
	switch t.GetType().(type) {
	case *core.LiteralType_CollectionType:
		return collectionInstanceChecker{
			literalType: t,
		}
	case *core.LiteralType_MapValueType:
		return mapInstanceChecker{
			literalType: t,
		}
	case *core.LiteralType_Blob:
		return blobInstanceChecker{
			literalType: t,
		}
	case *core.LiteralType_Schema:
		return schemaInstanceChecker{
			literalType: t,
		}
	case *core.LiteralType_UnionType:
		return unionInstanceChecker{
			literalType: t,
		}
	case *core.LiteralType_StructuredDatasetType:
		return structuredDatasetInstanceChecker{
			literalType: t,
		}
	default:
		if isNoneType(t) {
			return noneInstanceChecker{}
		}

		return trivialInstanceChecker{
			literalType: t,
		}
	}
}

func IsInstance(lit *core.Literal, t *core.LiteralType) bool {
	instanceChecker := getInstanceChecker(t)

	if lit.GetOffloadedMetadata() != nil {
		return AreTypesCastable(lit.GetOffloadedMetadata().GetInferredType(), t)
	}
	return instanceChecker.isInstance(lit)
}

func GetTagForType(x *core.LiteralType) string {
	if x.GetStructure() == nil {
		return ""
	}
	return x.GetStructure().GetTag()
}
