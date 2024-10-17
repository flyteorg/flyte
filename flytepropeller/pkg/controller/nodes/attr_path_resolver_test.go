package nodes

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/shamaton/msgpack/v2"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/errors"
)

// FlyteFile and FlyteDirectory represented as map[interface{}]interface{}
type FlyteFile map[interface{}]interface{}
type FlyteDirectory map[interface{}]interface{}

// InnerDC struct (equivalent to InnerDC dataclass in Python)
type InnerDC struct {
	A int                 `json:"a"`
	B float64             `json:"b"`
	C string              `json:"c"`
	D bool                `json:"d"`
	E []int               `json:"e"`
	F []FlyteFile         `json:"f"`
	G [][]int             `json:"g"`
	H []map[int]bool      `json:"h"`
	I map[int]bool        `json:"i"`
	J map[int]FlyteFile   `json:"j"`
	K map[int][]int       `json:"k"`
	L map[int]map[int]int `json:"l"`
	M map[string]string   `json:"m"`
	N FlyteFile           `json:"n"`
	O FlyteDirectory      `json:"o"`
}

// DC struct (equivalent to DC dataclass in Python)
type DC struct {
	A     int                 `json:"a"`
	B     float64             `json:"b"`
	C     string              `json:"c"`
	D     bool                `json:"d"`
	E     []int               `json:"e"`
	F     []FlyteFile         `json:"f"`
	G     [][]int             `json:"g"`
	H     []map[int]bool      `json:"h"`
	I     map[int]bool        `json:"i"`
	J     map[int]FlyteFile   `json:"j"`
	K     map[int][]int       `json:"k"`
	L     map[int]map[int]int `json:"l"`
	M     map[string]string   `json:"m"`
	N     FlyteFile           `json:"n"`
	O     FlyteDirectory      `json:"o"`
	Inner InnerDC             `json:"inner_dc"`
}

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
		resolved, err := resolveAttrPathInPromise(context.Background(), nil, "", arg.literal, arg.path)
		if arg.hasError {
			assert.Error(t, err, i)
			assert.ErrorContains(t, err, errors.PromiseAttributeResolveError, i)
		} else {
			assert.Equal(t, arg.expected, resolved, i)
		}
	}
}

func createNestedDC() DC {
	flyteFile := FlyteFile{
		"path": "s3://my-s3-bucket/example.txt",
	}

	flyteDirectory := FlyteDirectory{
		"path": "s3://my-s3-bucket/s3_flyte_dir",
	}

	// Example of initializing InnerDC
	innerDC := InnerDC{
		A: -1,
		B: -2.1,
		C: "Hello, Flyte",
		D: false,
		E: []int{0, 1, 2, -1, -2},
		F: []FlyteFile{flyteFile},
		G: [][]int{{0}, {1}, {-1}},
		H: []map[int]bool{{0: false}, {1: true}, {-1: true}},
		I: map[int]bool{0: false, 1: true, -1: false},
		J: map[int]FlyteFile{
			0:  flyteFile,
			1:  flyteFile,
			-1: flyteFile,
		},
		K: map[int][]int{
			0: {0, 1, -1},
		},
		L: map[int]map[int]int{
			1: {-1: 0},
		},
		M: map[string]string{
			"key": "value",
		},
		N: flyteFile,
		O: flyteDirectory,
	}

	// Initializing DC
	dc := DC{
		A: 1,
		B: 2.1,
		C: "Hello, Flyte",
		D: false,
		E: []int{0, 1, 2, -1, -2},
		F: []FlyteFile{flyteFile},
		G: [][]int{{0}, {1}, {-1}},
		H: []map[int]bool{{0: false}, {1: true}, {-1: true}},
		I: map[int]bool{0: false, 1: true, -1: false},
		J: map[int]FlyteFile{
			0:  flyteFile,
			1:  flyteFile,
			-1: flyteFile,
		},
		K: map[int][]int{
			0: {0, 1, -1},
		},
		L: map[int]map[int]int{
			1: {-1: 0},
		},
		M: map[string]string{
			"key": "value",
		},
		N:     flyteFile,
		O:     flyteDirectory,
		Inner: innerDC,
	}
	return dc
}

func TestResolveAttrPathInBinary(t *testing.T) {
	// Helper function to convert a map to msgpack bytes and then to BinaryIDL
	toMsgpackBytes := func(m interface{}) []byte {
		msgpackBytes, err := msgpack.Marshal(m)
		assert.NoError(t, err)
		return msgpackBytes
	}

	flyteFile := FlyteFile{
		"path": "s3://my-s3-bucket/example.txt",
	}

	flyteDirectory := FlyteDirectory{
		"path": "s3://my-s3-bucket/s3_flyte_dir",
	}

	nestedDC := createNestedDC()
	literalNestedDC := &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Binary{
					Binary: &core.Binary{
						Value: toMsgpackBytes(nestedDC),
						Tag:   "msgpack",
					},
				},
			},
		},
	}

	args := []struct {
		literal  *core.Literal
		path     []*core.PromiseAttribute
		expected *core.Literal
		hasError bool
	}{
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "A",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(1),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "B",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(2.1),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "C",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes("Hello, Flyte"),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "D",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(false),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "E",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes([]int{0, 1, 2, -1, -2}),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "F",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes([]FlyteFile{{"path": "s3://my-s3-bucket/example.txt"}}),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "G",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes([][]int{{0}, {1}, {-1}}),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "H",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes([]map[int]bool{{0: false}, {1: true}, {-1: true}}),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "I",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(map[int]bool{0: false, 1: true, -1: false}),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "J",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(
									map[int]FlyteFile{
										0:  flyteFile,
										1:  flyteFile,
										-1: flyteFile,
									}),
								Tag: "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "K",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(
									map[int][]int{
										0: {0, 1, -1},
									}),
								Tag: "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "L",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(
									map[int]map[int]int{
										1: {-1: 0},
									}),
								Tag: "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "M",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(
									map[string]string{
										"key": "value",
									}),
								Tag: "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "N",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(flyteFile),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "O",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(flyteDirectory),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "Inner",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(nestedDC.Inner),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "Inner",
					},
				},
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "A",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(-1),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "Inner",
					},
				},
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "B",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(-2.1),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "Inner",
					},
				},
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "C",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes("Hello, Flyte"),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "Inner",
					},
				},
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "D",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(false),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "Inner",
					},
				},
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "E",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes([]int{0, 1, 2, -1, -2}),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "Inner",
					},
				},
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "F",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes([]FlyteFile{flyteFile}),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "Inner",
					},
				},
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "G",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes([][]int{{0}, {1}, {-1}}),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "Inner",
					},
				},
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "G",
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
								Value: toMsgpackBytes([]int{0}),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "Inner",
					},
				},
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "G",
					},
				},
				{
					Value: &core.PromiseAttribute_IntValue{
						IntValue: 2,
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
								Value: toMsgpackBytes(-1),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "Inner",
					},
				},
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "H",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes([]map[int]bool{{0: false}, {1: true}, {-1: true}}),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "Inner",
					},
				},
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "I",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(map[int]bool{0: false, 1: true, -1: false}),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "Inner",
					},
				},
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "J",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(map[int]FlyteFile{
									0:  flyteFile,
									1:  flyteFile,
									-1: flyteFile,
								}),
								Tag: "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "Inner",
					},
				},
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "K",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(
									map[int][]int{
										0: {0, 1, -1},
									}),
								Tag: "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "Inner",
					},
				},
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "L",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(
									map[int]map[int]int{
										1: {-1: 0},
									}),
								Tag: "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "Inner",
					},
				},
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "M",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(
									map[string]string{
										"key": "value",
									}),
								Tag: "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "Inner",
					},
				},
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "N",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(flyteFile),
								Tag:   "msgpack",
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			literal: literalNestedDC,
			path: []*core.PromiseAttribute{
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "Inner",
					},
				},
				{
					Value: &core.PromiseAttribute_StringValue{
						StringValue: "O",
					},
				},
			},
			expected: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Binary{
							Binary: &core.Binary{
								Value: toMsgpackBytes(flyteDirectory),
								Tag:   "msgpack",
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
	}

	ctx := context.Background()
	for i, arg := range args {
		resolved, err := resolveAttrPathInPromise(ctx, nil, "", arg.literal, arg.path)
		if arg.hasError {
			assert.Error(t, err, i)
			assert.ErrorContains(t, err, errors.PromiseAttributeResolveError, i)
		} else {
			var expectedValue, actualValue interface{}

			// Helper function to unmarshal a Binary Literal into an interface{}
			unmarshalBinaryLiteral := func(literal *core.Literal) (interface{}, error) {
				if scalar, ok := literal.Value.(*core.Literal_Scalar); ok {
					if binary, ok := scalar.Scalar.Value.(*core.Scalar_Binary); ok {
						var value interface{}
						err := msgpack.Unmarshal(binary.Binary.Value, &value)
						return value, err
					}
				}
				return nil, fmt.Errorf("literal is not a Binary Scalar")
			}

			// Unmarshal the expected value
			expectedValue, err := unmarshalBinaryLiteral(arg.expected)
			if err != nil {
				t.Fatalf("Failed to unmarshal expected value in test case %d: %v", i, err)
			}

			// Unmarshal the resolved value
			actualValue, err = unmarshalBinaryLiteral(resolved)
			if err != nil {
				t.Fatalf("Failed to unmarshal resolved value in test case %d: %v", i, err)
			}

			// Deeply compare the expected and actual values, ignoring map ordering
			if !reflect.DeepEqual(expectedValue, actualValue) {
				t.Fatalf("Test case %d: Expected %+v, but got %+v", i, expectedValue, actualValue)
			}
		}
	}
}
