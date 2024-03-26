package array

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

var (
	literalOne = &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Primitive{
					Primitive: &core.Primitive{
						Value: &core.Primitive_Integer{
							Integer: 1,
						},
					},
				},
			},
		},
	}
	literalTwo = &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Primitive{
					Primitive: &core.Primitive{
						Value: &core.Primitive_Integer{
							Integer: 2,
						},
					},
				},
			},
		},
	}
)

func TestConstructLiteralMap(t *testing.T) {
	tests := []struct {
		name                string
		inputLiteralMaps    *core.LiteralMap
		expectedLiteralMaps []*core.LiteralMap
	}{
		{
			"SingleList",
			&core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": &core.Literal{
						Value: &core.Literal_Collection{
							Collection: &core.LiteralCollection{
								Literals: []*core.Literal{
									literalOne,
									literalTwo,
								},
							},
						},
					},
				},
			},
			[]*core.LiteralMap{
				&core.LiteralMap{
					Literals: map[string]*core.Literal{
						"foo": literalOne,
					},
				},
				&core.LiteralMap{
					Literals: map[string]*core.Literal{
						"foo": literalTwo,
					},
				},
			},
		},
		{
			"MultiList",
			&core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": &core.Literal{
						Value: &core.Literal_Collection{
							Collection: &core.LiteralCollection{
								Literals: []*core.Literal{
									literalOne,
									literalTwo,
								},
							},
						},
					},
					"bar": &core.Literal{
						Value: &core.Literal_Collection{
							Collection: &core.LiteralCollection{
								Literals: []*core.Literal{
									literalTwo,
									literalOne,
								},
							},
						},
					},
				},
			},
			[]*core.LiteralMap{
				&core.LiteralMap{
					Literals: map[string]*core.Literal{
						"foo": literalOne,
						"bar": literalTwo,
					},
				},
				&core.LiteralMap{
					Literals: map[string]*core.Literal{
						"foo": literalTwo,
						"bar": literalOne,
					},
				},
			},
		},
		{
			"Partial",
			&core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": &core.Literal{
						Value: &core.Literal_Collection{
							Collection: &core.LiteralCollection{
								Literals: []*core.Literal{
									literalOne,
									literalTwo,
								},
							},
						},
					},
					"bar": literalTwo,
				},
			},
			[]*core.LiteralMap{
				&core.LiteralMap{
					Literals: map[string]*core.Literal{
						"foo": literalOne,
						"bar": literalTwo,
					},
				},
				&core.LiteralMap{
					Literals: map[string]*core.Literal{
						"foo": literalTwo,
						"bar": literalTwo,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for i := 0; i < len(test.expectedLiteralMaps); i++ {
				outputLiteralMap, err := constructLiteralMap(test.inputLiteralMaps, i)
				assert.NoError(t, err)
				assert.True(t, proto.Equal(test.expectedLiteralMaps[i], outputLiteralMap))
			}
		})
	}
}
