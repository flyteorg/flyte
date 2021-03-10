package validators

import (
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flytepropeller/pkg/utils"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/compiler/common/mocks"
	compilerErrors "github.com/flyteorg/flytepropeller/pkg/compiler/errors"
)

func TestValidateBindings(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		wf := &mocks.WorkflowBuilder{}
		n := &mocks.NodeBuilder{}
		bindings := []*core.Binding{}
		vars := &core.VariableMap{}
		compileErrors := compilerErrors.NewCompileErrors()
		resolved, ok := ValidateBindings(wf, n, bindings, vars, true, compileErrors)
		assert.True(t, ok)
		assert.Empty(t, resolved.Variables)
	})

	t.Run("Variable not in inputs", func(t *testing.T) {
		wf := &mocks.WorkflowBuilder{}
		n := &mocks.NodeBuilder{}
		n.OnGetId().Return("node1")
		bindings := []*core.Binding{
			{
				Var: "x",
			},
		}
		vars := &core.VariableMap{}
		compileErrors := compilerErrors.NewCompileErrors()
		_, ok := ValidateBindings(wf, n, bindings, vars, true, compileErrors)
		assert.False(t, ok)
		if !compileErrors.HasErrors() {
			assert.Error(t, compileErrors)
		}
	})

	t.Run("Bind the same variable twice", func(t *testing.T) {
		wf := &mocks.WorkflowBuilder{}
		n := &mocks.NodeBuilder{}
		n.OnGetId().Return("node1")

		bindings := []*core.Binding{
			{
				Var:     "x",
				Binding: LiteralToBinding(utils.MustMakeLiteral(5)),
			},
			{
				Var:     "x",
				Binding: LiteralToBinding(utils.MustMakeLiteral(5)),
			},
		}

		vars := &core.VariableMap{
			Variables: map[string]*core.Variable{
				"x": {
					Type: LiteralTypeForLiteral(utils.MustMakeLiteral(5)),
				},
			},
		}

		compileErrors := compilerErrors.NewCompileErrors()
		_, ok := ValidateBindings(wf, n, bindings, vars, true, compileErrors)
		assert.False(t, ok)
		if !compileErrors.HasErrors() {
			assert.Error(t, compileErrors)
		}
		assert.Equal(t, "ParameterBoundMoreThanOnce", string(compileErrors.Errors().List()[0].Code()))
	})

	t.Run("Happy Path", func(t *testing.T) {
		wf := &mocks.WorkflowBuilder{}
		n := &mocks.NodeBuilder{}
		n.OnGetId().Return("node1")

		bindings := []*core.Binding{
			{
				Var:     "x",
				Binding: LiteralToBinding(utils.MustMakeLiteral([]interface{}{5})),
			},
		}

		vars := &core.VariableMap{
			Variables: map[string]*core.Variable{
				"x": {
					Type: LiteralTypeForLiteral(utils.MustMakeLiteral([]interface{}{5})),
				},
			},
		}

		compileErrors := compilerErrors.NewCompileErrors()
		_, ok := ValidateBindings(wf, n, bindings, vars, true, compileErrors)
		assert.True(t, ok)
		if compileErrors.HasErrors() {
			assert.NoError(t, compileErrors)
		}
	})

	t.Run("Maps", func(t *testing.T) {
		wf := &mocks.WorkflowBuilder{}
		n := &mocks.NodeBuilder{}
		n.OnGetId().Return("node1")

		bindings := []*core.Binding{
			{
				Var: "x",
				Binding: LiteralToBinding(utils.MustMakeLiteral(
					map[string]interface{}{
						"xy": 5,
					})),
			},
		}

		vars := &core.VariableMap{
			Variables: map[string]*core.Variable{
				"x": {
					Type: LiteralTypeForLiteral(utils.MustMakeLiteral(
						map[string]interface{}{
							"xy": 5,
						})),
				},
			},
		}

		compileErrors := compilerErrors.NewCompileErrors()
		_, ok := ValidateBindings(wf, n, bindings, vars, true, compileErrors)
		assert.True(t, ok)
		if compileErrors.HasErrors() {
			assert.NoError(t, compileErrors)
		}
	})

	t.Run("Promises", func(t *testing.T) {
		n := &mocks.NodeBuilder{}
		n.OnGetId().Return("node1")
		n.OnGetInterface().Return(&core.TypedInterface{
			Inputs: &core.VariableMap{
				Variables: map[string]*core.Variable{},
			},
			Outputs: &core.VariableMap{
				Variables: map[string]*core.Variable{},
			},
		})

		n2 := &mocks.NodeBuilder{}
		n2.OnGetId().Return("node2")
		n2.OnGetOutputAliases().Return(nil)
		n2.OnGetInterface().Return(&core.TypedInterface{
			Inputs: &core.VariableMap{
				Variables: map[string]*core.Variable{},
			},
			Outputs: &core.VariableMap{
				Variables: map[string]*core.Variable{
					"n2_out": {
						Type: LiteralTypeForLiteral(utils.MustMakeLiteral(2)),
					},
				},
			},
		})

		wf := &mocks.WorkflowBuilder{}
		wf.OnGetNode("n2").Return(n2, true)
		wf.On("AddExecutionEdge", mock.Anything, mock.Anything).Return(nil)

		bindings := []*core.Binding{
			{
				Var: "x",
				Binding: &core.BindingData{
					Value: &core.BindingData_Promise{
						Promise: &core.OutputReference{
							Var:    "n2_out",
							NodeId: "n2",
						},
					},
				},
			},
		}

		vars := &core.VariableMap{
			Variables: map[string]*core.Variable{
				"x": {
					Type: LiteralTypeForLiteral(utils.MustMakeLiteral(5)),
				},
			},
		}

		compileErrors := compilerErrors.NewCompileErrors()
		_, ok := ValidateBindings(wf, n, bindings, vars, true, compileErrors)
		assert.True(t, ok)
		if compileErrors.HasErrors() {
			assert.NoError(t, compileErrors)
		}
	})
}
