package nodes

import (
	"fmt"
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
	// 1. map {"foo": "bar"}
	// 2. collection ["foo", "bar"]
	// 3. struct1 {"foo": "bar"}
	// 4. struct2 {"foo": ["bar1", "bar2"]}
	// 5. map+collection+struct {"foo": [{"bar": "car"}]}
	// 6. exception key error with map
	// 7. exception out of range with collection
	// 8. exception key error with struct
	// 9. exception out of range with struct

	args := []struct {
		literal  *core.Literal
		path     []*core.PromiseAtrribute
		expected string
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
			path: []*core.PromiseAtrribute{
				&core.PromiseAtrribute{
					Value: &core.PromiseAtrribute_StringValue{
						StringValue: "foo",
					},
				},
			},
			expected: "bar",
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
			path: []*core.PromiseAtrribute{
				&core.PromiseAtrribute{
					Value: &core.PromiseAtrribute_IntValue{
						IntValue: 1,
					},
				},
			},
			expected: "bar",
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
			path: []*core.PromiseAtrribute{
				&core.PromiseAtrribute{
					Value: &core.PromiseAtrribute_StringValue{
						StringValue: "foo",
					},
				},
			},
			expected: "bar",
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
			path: []*core.PromiseAtrribute{
				&core.PromiseAtrribute{
					Value: &core.PromiseAtrribute_StringValue{
						StringValue: "foo",
					},
				},
				&core.PromiseAtrribute{
					Value: &core.PromiseAtrribute_IntValue{
						IntValue: 1,
					},
				},
			},
			expected: "bar2",
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
			path: []*core.PromiseAtrribute{
				&core.PromiseAtrribute{
					Value: &core.PromiseAtrribute_StringValue{
						StringValue: "foo",
					},
				},
				&core.PromiseAtrribute{
					Value: &core.PromiseAtrribute_IntValue{
						IntValue: 0,
					},
				},
				&core.PromiseAtrribute{
					Value: &core.PromiseAtrribute_StringValue{
						StringValue: "bar",
					},
				},
			},
			expected: "car",
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
			path: []*core.PromiseAtrribute{
				&core.PromiseAtrribute{
					Value: &core.PromiseAtrribute_StringValue{
						StringValue: "random",
					},
				},
			},
			expected: "",
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
			path: []*core.PromiseAtrribute{
				&core.PromiseAtrribute{
					Value: &core.PromiseAtrribute_IntValue{
						IntValue: 2,
					},
				},
			},
			expected: "",
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
			path: []*core.PromiseAtrribute{
				&core.PromiseAtrribute{
					Value: &core.PromiseAtrribute_StringValue{
						StringValue: "random",
					},
				},
			},
			expected: "",
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
			path: []*core.PromiseAtrribute{
				&core.PromiseAtrribute{
					Value: &core.PromiseAtrribute_StringValue{
						StringValue: "foo",
					},
				},
				&core.PromiseAtrribute{
					Value: &core.PromiseAtrribute_IntValue{
						IntValue: 100,
					},
				},
			},
			expected: "",
			hasError: true,
		},
	}

	for _, arg := range args {
		resolved, err := resolveAttrPathInPromise(nil, "", arg.literal, arg.path)
		if arg.hasError {
			assert.Error(t, err)
			assert.ErrorContains(t, err, errors.PromiseAttributeResolveError)
			fmt.Println(err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, arg.expected, resolved.GetScalar().GetPrimitive().GetStringValue())
		}
	}
}
