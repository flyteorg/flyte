package create

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

func TestMakeLiteralForTypes(t *testing.T) {
	inputTypes := map[string]*core.LiteralType{
		"a": {
			Type: &core.LiteralType_Simple{
				Simple: core.SimpleType_INTEGER,
			},
		},
		"x": {
			Type: &core.LiteralType_Simple{
				Simple: core.SimpleType_INTEGER,
			},
		},
		"b": {
			Type: &core.LiteralType_Simple{
				Simple: core.SimpleType_STRING,
			},
		},
		"y": {
			Type: &core.LiteralType_Simple{
				Simple: core.SimpleType_STRING,
			},
		},
	}

	t.Run("Happy path", func(t *testing.T) {
		inputValues := map[string]interface{}{
			"a": 5,
			"b": "hello",
		}

		m, err := MakeLiteralForTypes(inputValues, inputTypes)
		assert.NoError(t, err)
		assert.Len(t, m, len(inputValues))
	})

	t.Run("Type not found", func(t *testing.T) {
		inputValues := map[string]interface{}{
			"notfound": 5,
		}

		_, err := MakeLiteralForTypes(inputValues, inputTypes)
		assert.Error(t, err)
	})

	t.Run("Invalid value", func(t *testing.T) {
		inputValues := map[string]interface{}{
			"a": "hello",
		}

		_, err := MakeLiteralForTypes(inputValues, inputTypes)
		assert.Error(t, err)
	})
}

func TestMakeLiteralForParams(t *testing.T) {
	inputValues := map[string]interface{}{
		"a": "hello",
	}

	t.Run("Happy path", func(t *testing.T) {
		inputParams := map[string]*core.Parameter{
			"a": {
				Var: &core.Variable{
					Type: &core.LiteralType{
						Type: &core.LiteralType_Simple{
							Simple: core.SimpleType_STRING,
						},
					},
				},
			},
		}

		m, err := MakeLiteralForParams(inputValues, inputParams)
		assert.NoError(t, err)
		assert.Len(t, m, len(inputValues))
	})

	t.Run("Invalid Param", func(t *testing.T) {
		inputParams := map[string]*core.Parameter{
			"a": nil,
		}

		_, err := MakeLiteralForParams(inputValues, inputParams)
		assert.Error(t, err)
	})

	t.Run("Invalid Type", func(t *testing.T) {
		inputParams := map[string]*core.Parameter{
			"a": {
				Var: &core.Variable{},
			},
		}

		_, err := MakeLiteralForParams(inputValues, inputParams)
		assert.Error(t, err)
	})
}

func TestMakeLiteralForVariables(t *testing.T) {
	inputValues := map[string]interface{}{
		"a": "hello",
	}

	t.Run("Happy path", func(t *testing.T) {
		inputVariables := map[string]*core.Variable{
			"a": {
				Type: &core.LiteralType{
					Type: &core.LiteralType_Simple{
						Simple: core.SimpleType_STRING,
					},
				},
			},
		}

		m, err := MakeLiteralForVariables(inputValues, inputVariables)
		assert.NoError(t, err)
		assert.Len(t, m, len(inputValues))
	})

	t.Run("Invalid Variable", func(t *testing.T) {
		inputVariables := map[string]*core.Variable{
			"a": nil,
		}

		_, err := MakeLiteralForVariables(inputValues, inputVariables)
		assert.Error(t, err)
	})

	t.Run("Invalid Type", func(t *testing.T) {
		inputVariables := map[string]*core.Variable{
			"a": {
				Type: nil,
			},
		}

		_, err := MakeLiteralForVariables(inputValues, inputVariables)
		assert.Error(t, err)
	})
}
