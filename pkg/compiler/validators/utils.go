package validators

import (
	"fmt"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/golang/protobuf/proto"
	"k8s.io/apimachinery/pkg/util/sets"
)

func containsBindingByVariableName(bindings []*core.Binding, name string) (found bool) {
	for _, b := range bindings {
		if b.Var == name {
			return true
		}
	}

	return false
}

func findVariableByName(vars *core.VariableMap, name string) (variable *core.Variable, found bool) {
	if vars == nil || vars.Variables == nil {
		return nil, false
	}

	variable, found = vars.Variables[name]
	return
}

// Gets literal type for scalar value. This can be used to compare the underlying type of two scalars for compatibility.
func literalTypeForScalar(scalar *core.Scalar) *core.LiteralType {
	// TODO: Should we just pass the type information with the value?  That way we don't have to guess?
	var literalType *core.LiteralType
	switch scalar.GetValue().(type) {
	case *core.Scalar_Primitive:
		literalType = literalTypeForPrimitive(scalar.GetPrimitive())
	case *core.Scalar_Blob:
		if scalar.GetBlob().GetMetadata() == nil {
			return nil
		}

		literalType = &core.LiteralType{Type: &core.LiteralType_Blob{Blob: scalar.GetBlob().GetMetadata().GetType()}}
	case *core.Scalar_Binary:
		literalType = &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_BINARY}}
	case *core.Scalar_Schema:
		literalType = &core.LiteralType{
			Type: &core.LiteralType_Schema{
				Schema: scalar.GetSchema().Type,
			},
		}
	case *core.Scalar_NoneType:
		literalType = &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_NONE}}
	case *core.Scalar_Error:
		literalType = &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_ERROR}}
	case *core.Scalar_Generic:
		literalType = &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRUCT}}
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
	paramMap := make(map[string]*core.Variable, len(params.Variables))
	paramSet := sets.NewString()
	for paramName, param := range params.Variables {
		paramMap[paramName] = param
		paramSet.Insert(paramName)
	}

	return paramMap, paramSet
}

func filterVariables(vars *core.VariableMap, varNames sets.String) *core.VariableMap {
	res := &core.VariableMap{
		Variables: make(map[string]*core.Variable, len(varNames)),
	}

	for paramName, param := range vars.Variables {
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
			if v.Type.String() != existingV.Type.String() {
				return nil, fmt.Errorf("key already exists with a different type. %v has type [%v] on one side "+
					"and type [%v] on the other", k, existingV.Type.String(), v.Type.String())
			}
		}

		res[k] = v
	}

	return res, nil
}

// Gets LiteralType for literal, nil if the value of literal is unknown, or type None if the literal is a non-homogeneous
// type.
func LiteralTypeForLiteral(l *core.Literal) *core.LiteralType {
	switch l.GetValue().(type) {
	case *core.Literal_Scalar:
		return literalTypeForScalar(l.GetScalar())
	case *core.Literal_Collection:
		if len(l.GetCollection().Literals) == 0 {
			return &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_NONE}}
		}

		// Ensure literal collection types are homogeneous.
		var innerType *core.LiteralType
		for _, x := range l.GetCollection().Literals {
			otherType := LiteralTypeForLiteral(x)
			if innerType != nil && !AreTypesCastable(otherType, innerType) {
				return &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_NONE}}
			}

			innerType = otherType
		}

		return &core.LiteralType{Type: &core.LiteralType_CollectionType{CollectionType: innerType}}
	case *core.Literal_Map:
		if len(l.GetMap().Literals) == 0 {
			return &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_NONE}}
		}

		// Ensure literal map types are homogeneous.
		var innerType *core.LiteralType
		for _, x := range l.GetMap().Literals {
			otherType := LiteralTypeForLiteral(x)
			if innerType != nil && !AreTypesCastable(otherType, innerType) {
				return &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_NONE}}
			}

			innerType = otherType
		}

		return &core.LiteralType{Type: &core.LiteralType_MapValueType{MapValueType: innerType}}
	}

	return nil
}

// Converts a literal to a non-promise binding data.
func LiteralToBinding(l *core.Literal) *core.BindingData {
	switch l.GetValue().(type) {
	case *core.Literal_Scalar:
		return &core.BindingData{
			Value: &core.BindingData_Scalar{
				Scalar: l.GetScalar(),
			},
		}
	case *core.Literal_Collection:
		x := make([]*core.BindingData, 0, len(l.GetCollection().Literals))
		for _, sub := range l.GetCollection().Literals {
			x = append(x, LiteralToBinding(sub))
		}

		return &core.BindingData{
			Value: &core.BindingData_Collection{
				Collection: &core.BindingDataCollection{
					Bindings: x,
				},
			},
		}
	case *core.Literal_Map:
		x := make(map[string]*core.BindingData, len(l.GetMap().Literals))
		for key, val := range l.GetMap().Literals {
			x[key] = LiteralToBinding(val)
		}

		return &core.BindingData{
			Value: &core.BindingData_Map{
				Map: &core.BindingDataMap{
					Bindings: x,
				},
			},
		}
	}

	return nil
}
