package validators

import (
	"testing"

	c "github.com/flyteorg/flytepropeller/pkg/compiler/common"

	"github.com/flyteorg/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/compiler/common/mocks"
	compilerErrors "github.com/flyteorg/flytepropeller/pkg/compiler/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// LiteralToBinding converts a literal to a non-promise binding data.
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

func TestValidateBindings(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		wf := &mocks.WorkflowBuilder{}
		n := &mocks.NodeBuilder{}
		bindings := []*core.Binding{}
		vars := &core.VariableMap{}
		compileErrors := compilerErrors.NewCompileErrors()
		resolved, ok := ValidateBindings(wf, n, bindings, vars, true, c.EdgeDirectionBidirectional, compileErrors)
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
		_, ok := ValidateBindings(wf, n, bindings, vars, true, c.EdgeDirectionBidirectional, compileErrors)
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
				Binding: LiteralToBinding(coreutils.MustMakeLiteral(5)),
			},
			{
				Var:     "x",
				Binding: LiteralToBinding(coreutils.MustMakeLiteral(5)),
			},
		}

		vars := &core.VariableMap{
			Variables: map[string]*core.Variable{
				"x": {
					Type: LiteralTypeForLiteral(coreutils.MustMakeLiteral(5)),
				},
			},
		}

		compileErrors := compilerErrors.NewCompileErrors()
		_, ok := ValidateBindings(wf, n, bindings, vars, true, c.EdgeDirectionBidirectional, compileErrors)
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
				Binding: LiteralToBinding(coreutils.MustMakeLiteral([]interface{}{5})),
			},
		}

		vars := &core.VariableMap{
			Variables: map[string]*core.Variable{
				"x": {
					Type: LiteralTypeForLiteral(coreutils.MustMakeLiteral([]interface{}{5})),
				},
			},
		}

		compileErrors := compilerErrors.NewCompileErrors()
		_, ok := ValidateBindings(wf, n, bindings, vars, true, c.EdgeDirectionBidirectional, compileErrors)
		assert.True(t, ok)
		if compileErrors.HasErrors() {
			assert.NoError(t, compileErrors)
		}
	})

	t.Run("Enum legal string", func(t *testing.T) {
		wf := &mocks.WorkflowBuilder{}
		n := &mocks.NodeBuilder{}
		n.OnGetId().Return("node1")

		bindings := []*core.Binding{
			{
				Var:     "x",
				Binding: LiteralToBinding(coreutils.MustMakeLiteral("x")),
			},
		}

		vars := &core.VariableMap{
			Variables: map[string]*core.Variable{
				"x": {
					Type: &core.LiteralType{Type: &core.LiteralType_EnumType{
						EnumType: &core.EnumType{
							Values: []string{"x", "y", "z"},
						},
					}},
				},
			},
		}

		compileErrors := compilerErrors.NewCompileErrors()
		_, ok := ValidateBindings(wf, n, bindings, vars, true, c.EdgeDirectionBidirectional, compileErrors)
		assert.True(t, ok)
		if compileErrors.HasErrors() {
			assert.NoError(t, compileErrors)
		}
	})

	t.Run("Enum illegal string", func(t *testing.T) {
		wf := &mocks.WorkflowBuilder{}
		n := &mocks.NodeBuilder{}
		n.OnGetId().Return("node1")

		bindings := []*core.Binding{
			{
				Var:     "m",
				Binding: LiteralToBinding(coreutils.MustMakeLiteral("x")),
			},
		}

		vars := &core.VariableMap{
			Variables: map[string]*core.Variable{
				"x": {
					Type: &core.LiteralType{Type: &core.LiteralType_EnumType{
						EnumType: &core.EnumType{
							Values: []string{"x", "y", "z"},
						},
					}},
				},
			},
		}

		compileErrors := compilerErrors.NewCompileErrors()
		_, ok := ValidateBindings(wf, n, bindings, vars, true, c.EdgeDirectionBidirectional, compileErrors)
		assert.False(t, ok)
		if !compileErrors.HasErrors() {
			assert.Error(t, compileErrors)
		}
	})

	t.Run("Maps", func(t *testing.T) {
		wf := &mocks.WorkflowBuilder{}
		n := &mocks.NodeBuilder{}
		n.OnGetId().Return("node1")

		bindings := []*core.Binding{
			{
				Var: "x",
				Binding: LiteralToBinding(coreutils.MustMakeLiteral(
					map[string]interface{}{
						"xy": 5,
					})),
			},
		}

		vars := &core.VariableMap{
			Variables: map[string]*core.Variable{
				"x": {
					Type: LiteralTypeForLiteral(coreutils.MustMakeLiteral(
						map[string]interface{}{
							"xy": 5,
						})),
				},
			},
		}

		compileErrors := compilerErrors.NewCompileErrors()
		_, ok := ValidateBindings(wf, n, bindings, vars, true, c.EdgeDirectionBidirectional, compileErrors)
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
						Type: LiteralTypeForLiteral(coreutils.MustMakeLiteral(2)),
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
					Type: LiteralTypeForLiteral(coreutils.MustMakeLiteral(5)),
				},
			},
		}

		compileErrors := compilerErrors.NewCompileErrors()
		_, ok := ValidateBindings(wf, n, bindings, vars, true, c.EdgeDirectionBidirectional, compileErrors)
		assert.True(t, ok)
		if compileErrors.HasErrors() {
			assert.NoError(t, compileErrors)
		}
	})

	t.Run("Nil Binding Value", func(t *testing.T) {
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
						Type: LiteralTypeForLiteral(coreutils.MustMakeLiteral(2)),
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
					Value: nil, // Error case, the SDK should not produce an empty Value
				},
			},
		}

		vars := &core.VariableMap{
			Variables: map[string]*core.Variable{
				"x": {
					Type: LiteralTypeForLiteral(coreutils.MustMakeLiteral(5)),
				},
			},
		}

		compileErrors := compilerErrors.NewCompileErrors()
		_, ok := ValidateBindings(wf, n, bindings, vars, true, c.EdgeDirectionBidirectional, compileErrors)
		assert.False(t, ok)
		assert.True(t, compileErrors.HasErrors())
		assert.Equal(t, 1, len(compileErrors.Errors().List()))
		t.Log(compileErrors.Error())
	})

	t.Run("Int to Unambiguous Union Binding", func(t *testing.T) {
		wf := &mocks.WorkflowBuilder{}
		n := &mocks.NodeBuilder{}
		n.OnGetId().Return("node1")

		bindings := []*core.Binding{
			{
				Var:     "x",
				Binding: LiteralToBinding(coreutils.MustMakeLiteral(5)),
			},
		}

		vars := &core.VariableMap{
			Variables: map[string]*core.Variable{
				"x": {
					Type: &core.LiteralType{
						Type: &core.LiteralType_UnionType{
							UnionType: &core.UnionType{
								Variants: []*core.LiteralType{
									{
										Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
										Structure: &core.TypeStructure{
											Tag: "str",
										},
									},
									{
										Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
										Structure: &core.TypeStructure{
											Tag: "int",
										},
									},
								},
							},
						},
					},
				},
			},
		}

		compileErrors := compilerErrors.NewCompileErrors()
		_, ok := ValidateBindings(wf, n, bindings, vars, true, c.EdgeDirectionBidirectional, compileErrors)
		assert.True(t, ok)
		if compileErrors.HasErrors() {
			assert.NoError(t, compileErrors)
		}
	})

	t.Run("Int to Ambiguous Union Binding", func(t *testing.T) {
		wf := &mocks.WorkflowBuilder{}
		n := &mocks.NodeBuilder{}
		n.OnGetId().Return("node1")

		bindings := []*core.Binding{
			{
				Var:     "x",
				Binding: LiteralToBinding(coreutils.MustMakeLiteral(5)),
			},
		}

		vars := &core.VariableMap{
			Variables: map[string]*core.Variable{
				"x": {
					Type: &core.LiteralType{
						Type: &core.LiteralType_UnionType{
							UnionType: &core.UnionType{
								Variants: []*core.LiteralType{
									{
										Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
										Structure: &core.TypeStructure{
											Tag: "int1",
										},
									},
									{
										Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
										Structure: &core.TypeStructure{
											Tag: "int2",
										},
									},
								},
							},
						},
					},
				},
			},
		}

		compileErrors := compilerErrors.NewCompileErrors()
		_, ok := ValidateBindings(wf, n, bindings, vars, true, c.EdgeDirectionBidirectional, compileErrors)
		assert.False(t, ok)
	})

	t.Run("Int to Incompatible Union Binding", func(t *testing.T) {
		// Should still succeed because ambiguity checking is deferred to the SDK which knows about type transformers
		wf := &mocks.WorkflowBuilder{}
		n := &mocks.NodeBuilder{}
		n.OnGetId().Return("node1")

		bindings := []*core.Binding{
			{
				Var:     "x",
				Binding: LiteralToBinding(coreutils.MustMakeLiteral(5)),
			},
		}

		vars := &core.VariableMap{
			Variables: map[string]*core.Variable{
				"x": {
					Type: &core.LiteralType{
						Type: &core.LiteralType_UnionType{
							UnionType: &core.UnionType{
								Variants: []*core.LiteralType{
									{
										Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
										Structure: &core.TypeStructure{
											Tag: "str",
										},
									},
								},
							},
						},
					},
				},
			},
		}

		compileErrors := compilerErrors.NewCompileErrors()
		_, ok := ValidateBindings(wf, n, bindings, vars, true, c.EdgeDirectionBidirectional, compileErrors)
		assert.False(t, ok)
		assert.Equal(t, "MismatchingTypes", string(compileErrors.Errors().List()[0].Code()))
	})

	t.Run("Union Literal to Union Binding", func(t *testing.T) {
		// Should still succeed because ambiguity checking is deferred to the SDK which knows about type transformers
		wf := &mocks.WorkflowBuilder{}
		n := &mocks.NodeBuilder{}
		n.OnGetId().Return("node1")

		bindings := []*core.Binding{
			{
				Var: "x",
				Binding: LiteralToBinding(&core.Literal{
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Union{
								Union: &core.Union{
									Value: coreutils.MustMakeLiteral(5),
									Type: &core.LiteralType{
										Type: &core.LiteralType_Simple{
											Simple: core.SimpleType_INTEGER,
										},
										Structure: &core.TypeStructure{
											Tag: "int",
										},
									},
								},
							},
						},
					},
				}),
			},
		}

		vars := &core.VariableMap{
			Variables: map[string]*core.Variable{
				"x": {
					Type: &core.LiteralType{
						Type: &core.LiteralType_UnionType{
							UnionType: &core.UnionType{
								Variants: []*core.LiteralType{
									{
										Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
										Structure: &core.TypeStructure{
											Tag: "int1",
										},
									},
									{
										Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
										Structure: &core.TypeStructure{
											Tag: "str",
										},
									},
									{
										Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
										Structure: &core.TypeStructure{
											Tag: "int",
										},
									},
								},
							},
						},
					},
				},
			},
		}

		compileErrors := compilerErrors.NewCompileErrors()
		_, ok := ValidateBindings(wf, n, bindings, vars, true, c.EdgeDirectionBidirectional, compileErrors)
		assert.True(t, ok)
		if compileErrors.HasErrors() {
			assert.NoError(t, compileErrors)
		}
	})

	t.Run("Union Literal to Incompatible Union Binding", func(t *testing.T) {
		// Should still succeed because ambiguity checking is deferred to the SDK which knows about type transformers
		wf := &mocks.WorkflowBuilder{}
		n := &mocks.NodeBuilder{}
		n.OnGetId().Return("node1")

		bindings := []*core.Binding{
			{
				Var: "x",
				Binding: LiteralToBinding(&core.Literal{
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Union{
								Union: &core.Union{
									Value: coreutils.MustMakeLiteral(5),
									Type: &core.LiteralType{
										Type: &core.LiteralType_Simple{
											Simple: core.SimpleType_INTEGER,
										},
										Structure: &core.TypeStructure{
											Tag: "int",
										},
									},
								},
							},
						},
					},
				}),
			},
		}

		vars := &core.VariableMap{
			Variables: map[string]*core.Variable{
				"x": {
					Type: &core.LiteralType{
						Type: &core.LiteralType_UnionType{
							UnionType: &core.UnionType{
								Variants: []*core.LiteralType{
									{
										Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
										Structure: &core.TypeStructure{
											Tag: "int_other",
										},
									},
								},
							},
						},
					},
				},
			},
		}

		compileErrors := compilerErrors.NewCompileErrors()
		_, ok := ValidateBindings(wf, n, bindings, vars, true, c.EdgeDirectionBidirectional, compileErrors)
		assert.False(t, ok)
		assert.Equal(t, "MismatchingTypes", string(compileErrors.Errors().List()[0].Code()))
	})

	t.Run("List of Int to List of Unions Binding", func(t *testing.T) {
		wf := &mocks.WorkflowBuilder{}
		n := &mocks.NodeBuilder{}
		n.OnGetId().Return("node1")

		bindings := []*core.Binding{
			{
				Var:     "x",
				Binding: LiteralToBinding(coreutils.MustMakeLiteral([]interface{}{5})),
			},
		}

		vars := &core.VariableMap{
			Variables: map[string]*core.Variable{
				"x": {
					Type: &core.LiteralType{
						Type: &core.LiteralType_CollectionType{
							CollectionType: &core.LiteralType{
								Type: &core.LiteralType_UnionType{
									UnionType: &core.UnionType{
										Variants: []*core.LiteralType{
											{
												Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
												Structure: &core.TypeStructure{
													Tag: "str",
												},
											},
											{
												Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
												Structure: &core.TypeStructure{
													Tag: "int",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		compileErrors := compilerErrors.NewCompileErrors()
		_, ok := ValidateBindings(wf, n, bindings, vars, true, c.EdgeDirectionBidirectional, compileErrors)
		assert.True(t, ok)
		if compileErrors.HasErrors() {
			assert.NoError(t, compileErrors)
		}
	})

	t.Run("List of Int to List of Incompatible Unions Binding", func(t *testing.T) {
		wf := &mocks.WorkflowBuilder{}
		n := &mocks.NodeBuilder{}
		n.OnGetId().Return("node1")

		bindings := []*core.Binding{
			{
				Var:     "x",
				Binding: LiteralToBinding(coreutils.MustMakeLiteral([]interface{}{5})),
			},
		}

		vars := &core.VariableMap{
			Variables: map[string]*core.Variable{
				"x": {
					Type: &core.LiteralType{
						Type: &core.LiteralType_CollectionType{
							CollectionType: &core.LiteralType{
								Type: &core.LiteralType_UnionType{
									UnionType: &core.UnionType{
										Variants: []*core.LiteralType{
											{
												Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
												Structure: &core.TypeStructure{
													Tag: "str",
												},
											},
											{
												Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
												Structure: &core.TypeStructure{
													Tag: "str1",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		compileErrors := compilerErrors.NewCompileErrors()
		_, ok := ValidateBindings(wf, n, bindings, vars, true, c.EdgeDirectionBidirectional, compileErrors)
		assert.False(t, ok)
		assert.Equal(t, "MismatchingTypes", string(compileErrors.Errors().List()[0].Code()))
	})

	t.Run("List of Int to Union of Lists Binding", func(t *testing.T) {
		wf := &mocks.WorkflowBuilder{}
		n := &mocks.NodeBuilder{}
		n.OnGetId().Return("node1")

		bindings := []*core.Binding{
			{
				Var:     "x",
				Binding: LiteralToBinding(coreutils.MustMakeLiteral([]interface{}{5})),
			},
		}

		vars := &core.VariableMap{
			Variables: map[string]*core.Variable{
				"x": {
					Type: &core.LiteralType{
						Type: &core.LiteralType_UnionType{
							UnionType: &core.UnionType{
								Variants: []*core.LiteralType{
									{
										Type: &core.LiteralType_CollectionType{
											CollectionType: &core.LiteralType{
												Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
											},
										},
										Structure: &core.TypeStructure{
											Tag: "list",
										},
									},
									{
										Type: &core.LiteralType_CollectionType{
											CollectionType: &core.LiteralType{
												Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
											},
										},
										Structure: &core.TypeStructure{
											Tag: "list",
										},
									},
								},
							},
						},
					},
				},
			},
		}

		compileErrors := compilerErrors.NewCompileErrors()
		_, ok := ValidateBindings(wf, n, bindings, vars, true, c.EdgeDirectionBidirectional, compileErrors)
		assert.True(t, ok)
		if compileErrors.HasErrors() {
			assert.NoError(t, compileErrors)
		}
	})

	t.Run("List of Int to Incompatible Union of Lists Binding", func(t *testing.T) {
		wf := &mocks.WorkflowBuilder{}
		n := &mocks.NodeBuilder{}
		n.OnGetId().Return("node1")

		bindings := []*core.Binding{
			{
				Var:     "x",
				Binding: LiteralToBinding(coreutils.MustMakeLiteral([]interface{}{5})),
			},
		}

		vars := &core.VariableMap{
			Variables: map[string]*core.Variable{
				"x": {
					Type: &core.LiteralType{
						Type: &core.LiteralType_UnionType{
							UnionType: &core.UnionType{
								Variants: []*core.LiteralType{
									{
										Type: &core.LiteralType_CollectionType{
											CollectionType: &core.LiteralType{
												Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
											},
										},
										Structure: &core.TypeStructure{
											Tag: "list",
										},
									},
									{
										Type: &core.LiteralType_CollectionType{
											CollectionType: &core.LiteralType{
												Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
											},
										},
										Structure: &core.TypeStructure{
											Tag: "list",
										},
									},
								},
							},
						},
					},
				},
			},
		}

		compileErrors := compilerErrors.NewCompileErrors()
		_, ok := ValidateBindings(wf, n, bindings, vars, true, c.EdgeDirectionBidirectional, compileErrors)
		assert.False(t, ok)
		assert.Equal(t, "IncompatibleBindingUnionValue", string(compileErrors.Errors().List()[0].Code()))
	})

	t.Run("List of Int to Ambiguous Union of Lists Binding", func(t *testing.T) {
		wf := &mocks.WorkflowBuilder{}
		n := &mocks.NodeBuilder{}
		n.OnGetId().Return("node1")

		bindings := []*core.Binding{
			{
				Var:     "x",
				Binding: LiteralToBinding(coreutils.MustMakeLiteral([]interface{}{5})),
			},
		}

		vars := &core.VariableMap{
			Variables: map[string]*core.Variable{
				"x": {
					Type: &core.LiteralType{
						Type: &core.LiteralType_UnionType{
							UnionType: &core.UnionType{
								Variants: []*core.LiteralType{
									{
										Type: &core.LiteralType_CollectionType{
											CollectionType: &core.LiteralType{
												Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
											},
										},
										Structure: &core.TypeStructure{
											Tag: "list1",
										},
									},
									{
										Type: &core.LiteralType_CollectionType{
											CollectionType: &core.LiteralType{
												Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
											},
										},
										Structure: &core.TypeStructure{
											Tag: "list2",
										},
									},
								},
							},
						},
					},
				},
			},
		}

		compileErrors := compilerErrors.NewCompileErrors()
		_, ok := ValidateBindings(wf, n, bindings, vars, true, c.EdgeDirectionBidirectional, compileErrors)
		assert.False(t, ok)
		assert.Equal(t, "AmbiguousBindingUnionValue", string(compileErrors.Errors().List()[0].Code()))
	})

	t.Run("Union Promise Unambiguous", func(t *testing.T) {
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
						Type: LiteralTypeForLiteral(coreutils.MustMakeLiteral(2)),
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
					Type: &core.LiteralType{
						Type: &core.LiteralType_UnionType{
							UnionType: &core.UnionType{
								Variants: []*core.LiteralType{
									{
										Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
										Structure: &core.TypeStructure{
											Tag: "str",
										},
									},
									{
										Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
										Structure: &core.TypeStructure{
											Tag: "int",
										},
									},
								},
							},
						},
					},
				},
			},
		}

		compileErrors := compilerErrors.NewCompileErrors()
		_, ok := ValidateBindings(wf, n, bindings, vars, true, c.EdgeDirectionBidirectional, compileErrors)
		assert.True(t, ok)
		if compileErrors.HasErrors() {
			assert.NoError(t, compileErrors)
		}
	})

	t.Run("Union Promise Ambiguous", func(t *testing.T) {
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
						Type: LiteralTypeForLiteral(coreutils.MustMakeLiteral(2)),
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
					Type: &core.LiteralType{
						Type: &core.LiteralType_UnionType{
							UnionType: &core.UnionType{
								Variants: []*core.LiteralType{
									{
										Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
										Structure: &core.TypeStructure{
											Tag: "str",
										},
									},
									{
										Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
										Structure: &core.TypeStructure{
											Tag: "int1",
										},
									},
									{
										Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
										Structure: &core.TypeStructure{
											Tag: "int2",
										},
									},
								},
							},
						},
					},
				},
			},
		}

		compileErrors := compilerErrors.NewCompileErrors()
		_, ok := ValidateBindings(wf, n, bindings, vars, true, c.EdgeDirectionBidirectional, compileErrors)
		assert.False(t, ok)
		assert.Equal(t, "MismatchingTypes", string(compileErrors.Errors().List()[0].Code()))
	})

	t.Run("Union Promise Union Literal", func(t *testing.T) {
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
						Type: LiteralTypeForLiteral(&core.Literal{
							Value: &core.Literal_Scalar{
								Scalar: &core.Scalar{
									Value: &core.Scalar_Union{
										Union: &core.Union{
											Value: coreutils.MustMakeLiteral(5),
											Type: &core.LiteralType{
												Type: &core.LiteralType_Simple{
													Simple: core.SimpleType_INTEGER,
												},
												Structure: &core.TypeStructure{
													Tag: "int1",
												},
											},
										},
									},
								},
							},
						}),
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
					Type: &core.LiteralType{
						Type: &core.LiteralType_UnionType{
							UnionType: &core.UnionType{
								Variants: []*core.LiteralType{
									{
										Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
										Structure: &core.TypeStructure{
											Tag: "str",
										},
									},
									{
										Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
										Structure: &core.TypeStructure{
											Tag: "int1",
										},
									},
									{
										Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
										Structure: &core.TypeStructure{
											Tag: "int2",
										},
									},
								},
							},
						},
					},
				},
				"y": {
					Type: &core.LiteralType{
						Type: &core.LiteralType_UnionType{
							UnionType: &core.UnionType{
								Variants: []*core.LiteralType{
									{
										Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
										Structure: &core.TypeStructure{
											Tag: "str",
										},
									},
									{
										Type: &core.LiteralType_Simple{Simple: core.SimpleType_NONE},
										Structure: &core.TypeStructure{
											Tag: "none",
										},
									},
								},
							},
						},
					},
				},
			},
		}

		compileErrors := compilerErrors.NewCompileErrors()
		_, ok := ValidateBindings(wf, n, bindings, vars, true, c.EdgeDirectionBidirectional, compileErrors)
		assert.True(t, ok)
		if compileErrors.HasErrors() {
			assert.NoError(t, compileErrors)
		}
	})
}
