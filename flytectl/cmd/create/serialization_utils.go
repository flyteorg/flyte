package create

import (
	"github.com/flyteorg/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

// TODO: Move all functions to flyteidl
func MakeLiteralForVariables(serialize map[string]interface{}, variables map[string]*core.Variable) (map[string]*core.Literal, error) {
	result := make(map[string]*core.Literal)
	var err error
	for k, v := range variables {
		if result[k], err = coreutils.MakeLiteralForType(v.Type, serialize[k]); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func MakeLiteralForParams(serialize map[string]interface{}, parameters map[string]*core.Parameter) (map[string]*core.Literal, error) {
	result := make(map[string]*core.Literal)
	var err error
	for k, v := range parameters {
		if result[k], err = coreutils.MakeLiteralForType(v.GetVar().Type, serialize[k]); err != nil {
			return nil, err
		}
	}
	return result, nil
}
