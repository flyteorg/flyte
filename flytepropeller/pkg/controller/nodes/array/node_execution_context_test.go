package array

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	"github.com/flyteorg/flyte/flytestdlib/pbhash"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
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
)

func TestConstructLiteralMap(t *testing.T) {
	ctx := context.TODO()
	scope := promutils.NewTestScope()
	dataStore, err := storage.NewDataStore(&storage.Config{
		Type: storage.TypeMemory,
	}, scope)
	assert.NoError(t, err)

	dataReference := storage.DataReference("s3://bucket/key")
	err = dataStore.WriteProtobuf(ctx, dataReference, storage.Options{}, collectionLiteral)
	assert.Nil(t, err)

	individualInputFilesArrayNode := mocks.ExecutableArrayNode{}
	individualInputFilesArrayNode.OnGetDataMode().Return(core.ArrayNode_INDIVIDUAL_INPUT_FILES)

	singleInputFileArrayNode := mocks.ExecutableArrayNode{}
	singleInputFileArrayNode.OnGetDataMode().Return(core.ArrayNode_SINGLE_INPUT_FILE)

	literalOneHash, err := pbhash.ComputeHash(ctx, literalOne)
	assert.NoError(t, err)
	offloadedLiteral := &core.Literal{
		Value: &core.Literal_OffloadedMetadata{
			OffloadedMetadata: &core.LiteralOffloadedMetadata{
				Uri:       "s3://bucket/key",
				SizeBytes: 100,
				InferredType: &core.LiteralType{
					Type: &core.LiteralType_CollectionType{
						CollectionType: &core.LiteralType{
							Type: &core.LiteralType_Simple{
								Simple: core.SimpleType_INTEGER,
							},
						},
					},
				},
			},
		},
		Hash: base64.RawURLEncoding.EncodeToString(literalOneHash),
	}

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
			&mocks.ExecutableArrayNode{},
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
			&mocks.ExecutableArrayNode{},
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
			&mocks.ExecutableArrayNode{},
		},
		{
			"Offloaded Literal Collection - Individual Input Files",
			&core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": {
						Value: &core.Literal_OffloadedMetadata{
							OffloadedMetadata: &core.LiteralOffloadedMetadata{
								Uri:       "s3://bucket/key",
								SizeBytes: 100,
								InferredType: &core.LiteralType{
									Type: &core.LiteralType_CollectionType{
										CollectionType: &core.LiteralType{
											Type: &core.LiteralType_Simple{
												Simple: core.SimpleType_INTEGER,
											},
										},
									},
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
			},
			&individualInputFilesArrayNode,
		},
		{
			"Offloaded Literal Collection - Single Input File",
			&core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": {
						Value: &core.Literal_OffloadedMetadata{
							OffloadedMetadata: &core.LiteralOffloadedMetadata{
								Uri:       "s3://bucket/key",
								SizeBytes: 100,
								InferredType: &core.LiteralType{
									Type: &core.LiteralType_CollectionType{
										CollectionType: &core.LiteralType{
											Type: &core.LiteralType_Simple{
												Simple: core.SimpleType_INTEGER,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			[]*core.LiteralMap{
				&core.LiteralMap{
					Literals: map[string]*core.Literal{
						"foo": offloadedLiteral,
					},
				},
			},
			&singleInputFileArrayNode,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for i := 0; i < len(test.expectedLiteralMaps); i++ {
				outputLiteralMap, err := constructLiteralMap(ctx, dataStore, test.arrayNode, test.inputLiteralMaps, i)
				assert.NoError(t, err)
				assert.True(t, proto.Equal(test.expectedLiteralMaps[i], outputLiteralMap))
			}
		})
	}
}
