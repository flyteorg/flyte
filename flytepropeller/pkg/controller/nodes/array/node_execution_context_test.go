package array

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
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
	collectionLiteral = &core.Literal{
		Value: &core.Literal_Collection{
			Collection: &core.LiteralCollection{
				Literals: []*core.Literal{
					literalOne,
				},
			},
		},
	}
	collectionLiteralTwo = &core.Literal{
		Value: &core.Literal_Collection{
			Collection: &core.LiteralCollection{
				Literals: []*core.Literal{
					literalTwo,
				},
			},
		},
	}
	collectionOfCollectionLiteral = &core.Literal{
		Value: &core.Literal_Collection{
			Collection: &core.LiteralCollection{
				Literals: []*core.Literal{
					collectionLiteral,
					collectionLiteralTwo,
				},
			},
		},
	}
)

func TestConstructLiteralMap(t *testing.T) {

	mockArrayNode := mocks.ExecutableArrayNode{}
	mockArrayNode.On("GetBoundInputs").Return([]string{})

	boundInputsArrayNode := mocks.ExecutableArrayNode{}
	boundInputsArrayNode.On("GetBoundInputs").Return([]string{"bar"})

	tests := []struct {
		name                string
		inputLiteralMaps    *core.LiteralMap
		expectedLiteralMaps []*core.LiteralMap
		arrayNode           v1alpha1.ExecutableArrayNode
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
			&mockArrayNode,
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
			&mockArrayNode,
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
			&mockArrayNode,
		},
		{
			"Bound inputs - scalar",
			&core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": collectionLiteralTwo,
					"bar": literalTwo,
				},
			},
			[]*core.LiteralMap{
				&core.LiteralMap{
					Literals: map[string]*core.Literal{
						"foo": literalTwo,
						"bar": literalTwo,
					},
				},
			},
			&boundInputsArrayNode,
		},
		{
			"Bound inputs - collection",
			&core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": collectionOfCollectionLiteral,
					"bar": collectionLiteralTwo,
				},
			},
			[]*core.LiteralMap{
				&core.LiteralMap{
					Literals: map[string]*core.Literal{
						"foo": collectionLiteral,
						"bar": collectionLiteralTwo,
					},
				},
				&core.LiteralMap{
					Literals: map[string]*core.Literal{
						"foo": collectionLiteralTwo,
						"bar": collectionLiteralTwo,
					},
				},
			},
			&boundInputsArrayNode,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for i := 0; i < len(test.expectedLiteralMaps); i++ {
				inputLiteralMap, err := constructLiteralMap(test.inputLiteralMaps, i, test.arrayNode)
				assert.NoError(t, err)
				assert.True(t, proto.Equal(test.expectedLiteralMaps[i], inputLiteralMap))
			}
		})
	}
}
