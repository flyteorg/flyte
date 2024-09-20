package nodes

import (
	"testing"

	"github.com/shamaton/msgpack/v2"
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

func TestResolveAttrPathInStruct(t *testing.T) {

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
				{
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
				{
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
				{
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
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "foo",
					},
				},
				{
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
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "foo",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Collection{
					Collection: &core.LiteralCollection{
						Literals: []*core.Literal{
							{
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
							"foo": {
								Value: &core.Literal_Collection{
									Collection: &core.LiteralCollection{
										Literals: []*core.Literal{
											{
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
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "foo",
					},
				},
				{
					Value: &core.PromiseAttribute_IntValue{
						IntValue: 0,
					},
				},
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "bar",
					},
				},
			},
			expected: NewScalarLiteral("car"),
			hasError: false,
		},
		// - nested map {"foo": {"bar": {"baz": 42}}}
		{
			literal: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Generic{
							Generic: NewStructFromMap(
								map[string]interface{}{
									"foo": map[string]interface{}{
										"bar": map[string]interface{}{
											"baz": 42,
										},
									},
								},
							),
						},
					},
				},
			},
			// Test accessing the entire nested map at foo.bar
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "foo",
					},
				},
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "bar",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Generic{
							Generic: NewStructFromMap(
								map[string]interface{}{
									"baz": 42,
								},
							),
						},
					},
				},
			},
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
				{
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
				{
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
				{
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
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "foo",
					},
				},
				{
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

func TestResolveAttrPathInBinary(t *testing.T) {
	// Helper function to convert a map to msgpack bytes and then to BinaryIDL
	toMsgpackBytes := func(m interface{}) []byte {
		msgpackBytes, _ := msgpack.Marshal(m)
		return msgpackBytes
	}

	args := []struct {
		literal  *core.Literal
		path     []*core.PromiseAttribute
		expected *core.Literal
		hasError bool
	}{
		// - nested map {"foo": {"bar": 42, "baz": {"qux": 3.14, "quux": "str"}}}
		{
			literal: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(map[string]interface{}{
									"foo": map[string]interface{}{
										"bar": int64(42),
										"baz": map[string]interface{}{
											"qux":  3.14,
											"quux": "str",
										},
									},
								}),
								Tag: "msgpack",
							},
						},
					},
				},
			},
			// Test accessing the int value at foo.bar
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "foo",
					},
				},
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "bar",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(int64(42)),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		// - nested map {"foo": {"bar": 42, "baz": {"qux": 3.14, "quux": "str"}}}
		{
			literal: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(map[string]interface{}{
									"foo": map[string]interface{}{
										"bar": int64(42),
										"baz": map[string]interface{}{
											"qux":  3.14,
											"quux": "str",
										},
									},
								}),
								Tag: "msgpack",
							},
						},
					},
				},
			},
			// Test accessing the float value at foo.baz.qux
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "foo",
					},
				},
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "baz",
					},
				},
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "qux",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(3.14),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		// - nested map {"foo": {"bar": 42, "baz": {"qux": 3.14, "quux": "str"}}}
		{
			literal: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(map[string]interface{}{
									"foo": map[string]interface{}{
										"bar": int64(42),
										"baz": map[string]interface{}{
											"qux":  3.14,
											"quux": "str",
										},
									},
								}),
								Tag: "msgpack",
							},
						},
					},
				},
			},
			// Test accessing the string value at foo.baz.quux
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "foo",
					},
				},
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "baz",
					},
				},
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "quux",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes("str"),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		// - nested list {"foo": [42, 3.14, "str"]}
		{
			literal: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(map[string]interface{}{
									"foo": []interface{}{int64(-42), 3.14, "str"},
								}),
								Tag: "msgpack",
							},
						},
					},
				},
			},
			// Test accessing the int value at foo[0]
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "foo",
					},
				},
				{
					Value: &core.PromiseAttribute_IntValue{
						IntValue: 0,
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(int64(-42)),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		// - nested list {"foo": [42, 3.14, "str"]}
		{
			literal: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(map[string]interface{}{
									"foo": []interface{}{int64(42), 3.14, "str"},
								}),
								Tag: "msgpack",
							},
						},
					},
				},
			},
			// Test accessing the float value at foo[1]
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "foo",
					},
				},
				{
					Value: &core.PromiseAttribute_IntValue{
						IntValue: 1,
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(3.14),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		// - nested list {"foo": [42, 3.14, "str"]}
		{
			literal: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(map[string]interface{}{
									"foo": []interface{}{int64(42), 3.14, "str"},
								}),
								Tag: "msgpack",
							},
						},
					},
				},
			},
			// Test accessing the string value at foo[2]
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "foo",
					},
				},
				{
					Value: &core.PromiseAttribute_IntValue{
						IntValue: 2,
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes("str"),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		// - test extracting a nested map as a Binary object {"foo": {"bar": {"baz": 42}}}
		{
			literal: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(map[string]interface{}{
									"foo": map[string]interface{}{
										"bar": map[string]interface{}{
											"baz": int64(42),
										},
									},
								}),
								Tag: "msgpack",
							},
						},
					},
				},
			},
			// Test accessing the entire nested map at foo.bar
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "foo",
					},
				},
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "bar",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(map[string]interface{}{
									"baz": int64(42),
								}),
								Tag: "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		// - exception case with non-existing key in nested map
		{
			literal: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(map[string]interface{}{
									"foo": map[string]interface{}{
										"bar": int64(42),
										"baz": map[string]interface{}{
											"qux":  3.14,
											"quux": "str",
										},
									},
								}),
								Tag: "msgpack",
							},
						},
					},
				},
			},
			// Test accessing a non-existing key in the nested map
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "foo",
					},
				},
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "baz",
					},
				},
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "unknown",
					},
				},
			},
			expected: &core.Literal{},
			hasError: true,
		},
		// - exception case with out-of-range index in list
		{
			literal: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(map[string]interface{}{
									"foo": []interface{}{int64(42), 3.14, "str"},
								}),
								Tag: "msgpack",
							},
						},
					},
				},
			},
			// Test accessing an out-of-range index in the list
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "foo",
					},
				},
				{
					Value: &core.PromiseAttribute_IntValue{
						IntValue: 10,
					},
				},
			},
			expected: &core.Literal{},
			hasError: true,
		},
		// - nested list struct {"foo": [["bar1", "bar2"]]}
		{
			literal: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(map[string]interface{}{
									"foo": []interface{}{[]interface{}{"bar1", "bar2"}},
								}),
								Tag: "msgpack",
							},
						},
					},
				},
			},
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "foo",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes([]interface{}{[]interface{}{"bar1", "bar2"}}),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
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
