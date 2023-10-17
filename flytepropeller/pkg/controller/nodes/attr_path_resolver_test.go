package nodes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/errors"
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

	args := []struct {
		literal  *core.Literal
		path     []*core.PromiseAttribute
		expected *core.Literal
		hasError bool
	}{
		// - map {"foo": "bar"}
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
		// - collection ["foo", "bar"]
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
		// - struct1 {"foo": "bar"}
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
		// - struct2 {"foo": ["bar1", "bar2"]}
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
		// - nested list struct {"foo": [["bar1", "bar2"]]}
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
		// - map+collection+struct {"foo": [{"bar": "car"}]}
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
		// - exception key error with map
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
		// - exception out of range with collection
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
		// - exception key error with struct
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
		// - exception out of range with struct
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
		resolved, err := resolveAttrPathInPromise("", arg.literal, arg.path)
		if arg.hasError {
			assert.Error(t, err, i)
			assert.ErrorContains(t, err, errors.PromiseAttributeResolveError, i)
		} else {
			assert.Equal(t, arg.expected, resolved, i)
		}
	}
}
