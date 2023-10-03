package nodes

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

func NewScalarLiteral(value string) *core.Literal {
	return &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Primitive{
					Primitive: &core.Primitive{
						Value: &core.Primitive_StringValue{
							StringValue: value,
						},
					},
				},
			},
		},
	}
}

func NewStructFromMap(m map[string]interface{}) *structpb.Struct {
	st, _ := structpb.NewStruct(m)
	return st
}

func TestResolveAttrPathIn(t *testing.T) {
	// - map {"foo": "bar"}
	// - collection ["foo", "bar"]
	// - struct1 {"foo": "bar"}
	// - struct2 {"foo": ["bar1", "bar2"]}
	// - nested list struct {"foo": [["bar1", "bar2"]]}
	// - map+collection+struct {"foo": [{"bar": "car"}]}
	// - exception key error with map
	// - exception out of range with collection
	// - exception key error with struct
	// - exception out of range with struct

	args := []struct {
		literal  *core.Literal
		path     []*core.PromiseAttribute
		expected *core.Literal
		hasError bool
	}{
		{
			literal: &core.Literal{
				Value: &core.Literal_Map{
					Map: &core.LiteralMap{
						Literals: map[string]*core.Literal{
							"foo": NewScalarLiteral("bar"),
						},
					},
				},
			},
			path: []*core.PromiseAttribute{
				&core.PromiseAttribute{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "foo",
					},
				},
			},
			expected: NewScalarLiteral("bar"),
			hasError: false,
		},
		{
			literal: &core.Literal{
				Value: &core.Literal_Collection{
					Collection: &core.LiteralCollection{
						Literals: []*core.Literal{
							NewScalarLiteral("foo"),
							NewScalarLiteral("bar"),
						},
					},
				},
			},
			path: []*core.PromiseAttribute{
				&core.PromiseAttribute{
					Value: &core.PromiseAttribute_IntValue{
						IntValue: 1,
					},
				},
			},
			expected: NewScalarLiteral("bar"),
			hasError: false,
		},
		{
			literal: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Generic{
							Generic: NewStructFromMap(map[string]interface{}{"foo": "bar"}),
						},
					},
				},
			},
			path: []*core.PromiseAttribute{
				&core.PromiseAttribute{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "foo",
					},
				},
			},
			expected: NewScalarLiteral("bar"),
			hasError: false,
		},
		{
			literal: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Generic{
							Generic: NewStructFromMap(
								map[string]interface{}{
									"foo": []interface{}{"bar1", "bar2"},
								},
							),
						},
					},
				},
			},
			path: []*core.PromiseAttribute{
				&core.PromiseAttribute{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "foo",
					},
				},
				&core.PromiseAttribute{
					Value: &core.PromiseAttribute_IntValue{
						IntValue: 1,
					},
				},
			},
			expected: NewScalarLiteral("bar2"),
			hasError: false,
		},
		{
			literal: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Generic{
							Generic: NewStructFromMap(
								map[string]interface{}{
									"foo": []interface{}{[]interface{}{"bar1", "bar2"}},
								},
							),
						},
					},
				},
			},
			path: []*core.PromiseAttribute{
				&core.PromiseAttribute{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "foo",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Collection{
					Collection: &core.LiteralCollection{
						Literals: []*core.Literal{
							&core.Literal{
								Value: &core.Literal_Collection{
									Collection: &core.LiteralCollection{
										Literals: []*core.Literal{
											NewScalarLiteral("bar1"),
											NewScalarLiteral("bar2"),
										},
									},
								},
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: &core.Literal{
				Value: &core.Literal_Map{
					Map: &core.LiteralMap{
						Literals: map[string]*core.Literal{
							"foo": &core.Literal{
								Value: &core.Literal_Collection{
									Collection: &core.LiteralCollection{
										Literals: []*core.Literal{
											&core.Literal{
												Value: &core.Literal_Scalar{
													Scalar: &core.Scalar{
														Value: &core.Scalar_Generic{
															Generic: NewStructFromMap(map[string]interface{}{"bar": "car"}),
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
				},
			},
			path: []*core.PromiseAttribute{
				&core.PromiseAttribute{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "foo",
					},
				},
				&core.PromiseAttribute{
					Value: &core.PromiseAttribute_IntValue{
						IntValue: 0,
					},
				},
				&core.PromiseAttribute{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "bar",
					},
				},
			},
			expected: NewScalarLiteral("car"),
			hasError: false,
		},
		{
			literal: &core.Literal{
				Value: &core.Literal_Map{
					Map: &core.LiteralMap{
						Literals: map[string]*core.Literal{
							"foo": NewScalarLiteral("bar"),
						},
					},
				},
			},
			path: []*core.PromiseAttribute{
				&core.PromiseAttribute{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "random",
					},
				},
			},
			expected: &core.Literal{},
			hasError: true,
		},
		{
			literal: &core.Literal{
				Value: &core.Literal_Collection{
					Collection: &core.LiteralCollection{
						Literals: []*core.Literal{
							NewScalarLiteral("foo"),
							NewScalarLiteral("bar"),
						},
					},
				},
			},
			path: []*core.PromiseAttribute{
				&core.PromiseAttribute{
					Value: &core.PromiseAttribute_IntValue{
						IntValue: 2,
					},
				},
			},
			expected: &core.Literal{},
			hasError: true,
		},
		{
			literal: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Generic{
							Generic: NewStructFromMap(map[string]interface{}{"foo": "bar"}),
						},
					},
				},
			},
			path: []*core.PromiseAttribute{
				&core.PromiseAttribute{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "random",
					},
				},
			},
			expected: &core.Literal{},
			hasError: true,
		},
		{
			literal: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Generic{
							Generic: NewStructFromMap(
								map[string]interface{}{
									"foo": []interface{}{"bar1", "bar2"},
								},
							),
						},
					},
				},
			},
			path: []*core.PromiseAttribute{
				&core.PromiseAttribute{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "foo",
					},
				},
				&core.PromiseAttribute{
					Value: &core.PromiseAttribute_IntValue{
						IntValue: 100,
					},
				},
			},
			expected: &core.Literal{},
			hasError: true,
		},
	}

	for i, arg := range args {
		resolved, err := resolveAttrPathInPromise(nil, "", arg.literal, arg.path)
		if arg.hasError {
			assert.Error(t, err, i)
			assert.ErrorContains(t, err, errors.PromiseAttributeResolveError, i)
		} else {
			assert.Equal(t, arg.expected, resolved, i)
		}
	}
}
