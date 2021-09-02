package create

import (
	"fmt"

	"github.com/flyteorg/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

// TODO: Move all functions to flyteidl
// MakeLiteralForVariables builds a map of literals for the provided serialized values. If a provided value does not have
// a corresponding variable or if that variable is invalid (e.g. doesn't have Type property populated), it returns an
// error.
func MakeLiteralForVariables(serialize map[string]interface{}, variables []*core.VariableMapEntry) (map[string]*core.Literal, error) {
	types := make(map[string]*core.LiteralType)
	for _, e := range variables {
		t := e.GetVar().GetType()
		if t == nil {
			return nil, fmt.Errorf("variable [%v] has nil type", e.GetName())
		}

		types[e.GetName()] = t
	}

	return MakeLiteralForTypes(serialize, types)
}

// MakeLiteralForParams builds a map of literals for the provided serialized values. If a provided value does not have
// a corresponding parameter or if that parameter is invalid (e.g. doesn't have Type property populated), it returns an
// error.
func MakeLiteralForParams(serialize map[string]interface{}, parameters []*core.ParameterMapEntry) (map[string]*core.Literal, error) {
	types := make(map[string]*core.LiteralType)
	for _, e := range parameters {
		if variable := e.GetParameter().GetVar(); variable == nil {
			return nil, fmt.Errorf("parameter [%v] has nil Variable", e.GetName())
		} else if t := variable.GetType(); t == nil {
			return nil, fmt.Errorf("parameter [%v] has nil variable type", e.GetName())
		} else {
			types[e.GetName()] = t
		}
	}

	return MakeLiteralForTypes(serialize, types)
}

// MakeLiteralForTypes builds a map of literals for the provided serialized values. If a provided value does not have
// a corresponding type or if it fails to create a literal for the given type and value, it returns an error.
func MakeLiteralForTypes(serialize map[string]interface{}, types map[string]*core.LiteralType) (map[string]*core.Literal, error) {
	result := make(map[string]*core.Literal)
	var err error
	for k, v := range serialize {
		if t, typeFound := types[k]; typeFound {
			if result[k], err = coreutils.MakeLiteralForType(t, v); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("no matching type for [%v]", k)
		}
	}

	return result, nil
}
