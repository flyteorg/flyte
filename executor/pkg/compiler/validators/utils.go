package validators

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/v2/flyteidl2/clients/go/coreutils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
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

func buildMultipleTypeUnion(innerType []*core.LiteralType) *core.LiteralType {
	var variants []*core.LiteralType
	isNested := false

	for _, x := range innerType {
		unionType := x.GetCollectionType().GetUnionType()
		if unionType != nil {
			isNested = true
			variants = append(variants, unionType.GetVariants()...)
		} else {
			variants = append(variants, x)
		}
	}
	unionLiteralType := &core.LiteralType{
		Type: &core.LiteralType_UnionType{
			UnionType: &core.UnionType{
				Variants: variants,
			},
		},
	}

	if isNested {
		return &core.LiteralType{
			Type: &core.LiteralType_CollectionType{
				CollectionType: unionLiteralType,
			},
		}
	}

	return unionLiteralType
}

func literalTypeForLiterals(literals []*core.Literal) *core.LiteralType {
	innerType := make([]*core.LiteralType, 0, 1)
	innerTypeSet := sets.NewString()
	var noneType *core.LiteralType
	for _, x := range literals {
		otherType := LiteralTypeForLiteral(x)
		otherTypeKey := otherType.String()
		if _, ok := x.GetValue().(*core.Literal_Collection); ok {
			if x.GetCollection().GetLiterals() == nil {
				noneType = otherType
				continue
			}
		}

		if !innerTypeSet.Has(otherTypeKey) {
			innerType = append(innerType, otherType)
			innerTypeSet.Insert(otherTypeKey)
		}
	}

	// only add none type if there aren't other types
	if len(innerType) == 0 && noneType != nil {
		innerType = append(innerType, noneType)
	}

	if len(innerType) == 0 {
		return &core.LiteralType{
			Type: &core.LiteralType_Simple{Simple: core.SimpleType_NONE},
		}
	} else if len(innerType) == 1 {
		return innerType[0]
	}

	// sort inner types to ensure consistent union types are generated
	slices.SortFunc(innerType, func(a, b *core.LiteralType) int {
		aStr := a.String()
		bStr := b.String()
		if aStr < bStr {
			return -1
		} else if aStr > bStr {
			return 1
		}

		return 0
	})
	return buildMultipleTypeUnion(innerType)
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

// LiteralTypeForLiteral gets LiteralType for literal, nil if the value of literal is unknown, or type collection/map of
// type None if the literal is a non-homogeneous type.
func LiteralTypeForLiteral(l *core.Literal) *core.LiteralType {
	switch l.GetValue().(type) {
	case *core.Literal_Scalar:
		return literalTypeForScalar(l.GetScalar())
	case *core.Literal_Collection:
		return &core.LiteralType{
			Type: &core.LiteralType_CollectionType{
				CollectionType: literalTypeForLiterals(l.GetCollection().GetLiterals()),
			},
		}
	case *core.Literal_Map:
		return &core.LiteralType{
			Type: &core.LiteralType_MapValueType{
				MapValueType: literalTypeForLiterals(maps.Values(l.GetMap().GetLiterals())),
			},
		}
	case *core.Literal_OffloadedMetadata:
		return l.GetOffloadedMetadata().GetInferredType()
	}
	return nil
}

func GetTagForType(x *core.LiteralType) string {
	if x.GetStructure() == nil {
		return ""
	}
	return x.GetStructure().GetTag()
}
