package array

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	pluginiomocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	nodesmocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces/mocks"
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

func TestSubNodeInputResolver_Initialize(t *testing.T) {
	ctx := context.TODO()
	scope := promutils.NewTestScope()
	dataStore, err := storage.NewDataStore(&storage.Config{
		Type: storage.TypeMemory,
	}, scope)
	assert.NoError(t, err)

	dataReference := storage.DataReference("s3://bucket/offloaded")
	offloadedCollection := &core.Literal{
		Value: &core.Literal_Collection{
			Collection: &core.LiteralCollection{
				Literals: []*core.Literal{literalOne, literalTwo},
			},
		},
	}
	err = dataStore.WriteProtobuf(ctx, dataReference, storage.Options{}, offloadedCollection)
	assert.NoError(t, err)

	tests := []struct {
		name                     string
		inputLiteralMap          *core.LiteralMap
		boundInputs              []string
		expectOffloadedDownloads int
	}{
		{
			name: "NoOffloadedLiterals",
			inputLiteralMap: &core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": {
						Value: &core.Literal_Collection{
							Collection: &core.LiteralCollection{
								Literals: []*core.Literal{literalOne, literalTwo},
							},
						},
					},
				},
			},
			boundInputs:              []string{},
			expectOffloadedDownloads: 0,
		},
		{
			name: "WithOffloadedLiteral",
			inputLiteralMap: &core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": {
						Value: &core.Literal_OffloadedMetadata{
							OffloadedMetadata: &core.LiteralOffloadedMetadata{
								Uri:       "s3://bucket/offloaded",
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
			boundInputs:              []string{},
			expectOffloadedDownloads: 1,
		},
		{
			name: "BoundInputsSkipped",
			inputLiteralMap: &core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": {
						Value: &core.Literal_OffloadedMetadata{
							OffloadedMetadata: &core.LiteralOffloadedMetadata{
								Uri:       "s3://bucket/offloaded",
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
			boundInputs:              []string{"foo"},
			expectOffloadedDownloads: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			inputFilePaths := &pluginiomocks.InputFilePaths{}
			inputFilePaths.OnGetInputPath().Return(storage.DataReference("s3://bucket/input"))
			inputReader := newStaticInputReader(inputFilePaths, test.inputLiteralMap)

			nCtx := &nodesmocks.NodeExecutionContext{}
			nCtx.OnInputReader().Return(inputReader)
			nCtx.OnDataStore().Return(dataStore)

			arrayNode := &mocks.ExecutableArrayNode{}
			arrayNode.OnGetBoundInputs().Return(test.boundInputs)

			resolver := newSubNodeInputResolver(nCtx, arrayNode)
			err := resolver.Initialize(ctx)
			assert.NoError(t, err)

			// Verify parent inputs were cached
			assert.NotNil(t, resolver.parentInputs)
			assert.Equal(t, len(test.inputLiteralMap.Literals), len(resolver.parentInputs.Literals))

			// Verify offloaded literals were downloaded
			assert.Equal(t, test.expectOffloadedDownloads, len(resolver.offloadedCollections))
		})
	}
}

func TestSubNodeInputResolver_GetSubNodeInputs(t *testing.T) {
	ctx := context.TODO()
	scope := promutils.NewTestScope()
	dataStore, err := storage.NewDataStore(&storage.Config{
		Type: storage.TypeMemory,
	}, scope)
	assert.NoError(t, err)

	dataReference := storage.DataReference("s3://bucket/offloaded")
	offloadedCollection := &core.Literal{
		Value: &core.Literal_Collection{
			Collection: &core.LiteralCollection{
				Literals: []*core.Literal{literalOne, literalTwo},
			},
		},
	}
	err = dataStore.WriteProtobuf(ctx, dataReference, storage.Options{}, offloadedCollection)
	assert.NoError(t, err)

	// Pre-compute hash for SINGLE_INPUT_FILE offloaded literal test
	literalOneHash, err := pbhash.ComputeHash(ctx, literalOne)
	assert.NoError(t, err)

	tests := []struct {
		name                string
		inputLiteralMap     *core.LiteralMap
		boundInputs         []string
		dataMode            core.ArrayNode_DataMode
		subNodeIndex        int
		expectedLiteralMap  *core.LiteralMap
		expectInputBindings bool
	}{
		{
			name: "CollectionLiteral_Index0",
			inputLiteralMap: &core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": {
						Value: &core.Literal_Collection{
							Collection: &core.LiteralCollection{
								Literals: []*core.Literal{literalOne, literalTwo},
							},
						},
					},
				},
			},
			boundInputs:  []string{},
			dataMode:     core.ArrayNode_INDIVIDUAL_INPUT_FILES,
			subNodeIndex: 0,
			expectedLiteralMap: &core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": literalOne,
				},
			},
			expectInputBindings: true,
		},
		{
			name: "CollectionLiteral_Index1",
			inputLiteralMap: &core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": {
						Value: &core.Literal_Collection{
							Collection: &core.LiteralCollection{
								Literals: []*core.Literal{literalOne, literalTwo},
							},
						},
					},
				},
			},
			boundInputs:  []string{},
			dataMode:     core.ArrayNode_INDIVIDUAL_INPUT_FILES,
			subNodeIndex: 1,
			expectedLiteralMap: &core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": literalTwo,
				},
			},
			expectInputBindings: true,
		},
		{
			name: "OffloadedLiteral_Index0",
			inputLiteralMap: &core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": {
						Value: &core.Literal_OffloadedMetadata{
							OffloadedMetadata: &core.LiteralOffloadedMetadata{
								Uri:       "s3://bucket/offloaded",
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
			boundInputs:  []string{},
			dataMode:     core.ArrayNode_INDIVIDUAL_INPUT_FILES,
			subNodeIndex: 0,
			expectedLiteralMap: &core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": literalOne,
				},
			},
			expectInputBindings: true,
		},
		{
			name: "BoundInput_PassedThrough",
			inputLiteralMap: &core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": {
						Value: &core.Literal_Collection{
							Collection: &core.LiteralCollection{
								Literals: []*core.Literal{literalOne, literalTwo},
							},
						},
					},
					"bar": literalTwo,
				},
			},
			boundInputs:  []string{"bar"},
			dataMode:     core.ArrayNode_INDIVIDUAL_INPUT_FILES,
			subNodeIndex: 0,
			expectedLiteralMap: &core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": literalOne,
					"bar": literalTwo,
				},
			},
			expectInputBindings: true,
		},
		{
			name: "MultipleCollections",
			inputLiteralMap: &core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": {
						Value: &core.Literal_Collection{
							Collection: &core.LiteralCollection{
								Literals: []*core.Literal{literalOne, literalTwo},
							},
						},
					},
					"bar": {
						Value: &core.Literal_Collection{
							Collection: &core.LiteralCollection{
								Literals: []*core.Literal{literalTwo, literalOne},
							},
						},
					},
				},
			},
			boundInputs:  []string{},
			dataMode:     core.ArrayNode_INDIVIDUAL_INPUT_FILES,
			subNodeIndex: 0,
			expectedLiteralMap: &core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": literalOne,
					"bar": literalTwo,
				},
			},
			expectInputBindings: true,
		},
		{
			name: "PartialCollectionAndScalar",
			inputLiteralMap: &core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": {
						Value: &core.Literal_Collection{
							Collection: &core.LiteralCollection{
								Literals: []*core.Literal{literalOne, literalTwo},
							},
						},
					},
					"bar": literalTwo,
				},
			},
			boundInputs:  []string{},
			dataMode:     core.ArrayNode_INDIVIDUAL_INPUT_FILES,
			subNodeIndex: 1,
			expectedLiteralMap: &core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": literalTwo,
					"bar": literalTwo,
				},
			},
			expectInputBindings: true,
		},
		{
			name: "BoundInput_Collection",
			inputLiteralMap: &core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": collectionOfCollectionLiteral,
					"bar": collectionLiteralTwo,
				},
			},
			boundInputs:  []string{"bar"},
			dataMode:     core.ArrayNode_INDIVIDUAL_INPUT_FILES,
			subNodeIndex: 0,
			expectedLiteralMap: &core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": collectionLiteral,
					"bar": collectionLiteralTwo,
				},
			},
			expectInputBindings: true,
		},
		{
			name: "SingleInputFile_NilBindings",
			inputLiteralMap: &core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": {
						Value: &core.Literal_Collection{
							Collection: &core.LiteralCollection{
								Literals: []*core.Literal{literalOne, literalTwo},
							},
						},
					},
				},
			},
			boundInputs:  []string{},
			dataMode:     core.ArrayNode_SINGLE_INPUT_FILE,
			subNodeIndex: 0,
			expectedLiteralMap: &core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": literalOne,
				},
			},
			expectInputBindings: false,
		},
		{
			name: "OffloadedLiteral_SingleInputFile",
			inputLiteralMap: &core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": {
						Value: &core.Literal_OffloadedMetadata{
							OffloadedMetadata: &core.LiteralOffloadedMetadata{
								Uri:       "s3://bucket/offloaded",
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
			boundInputs:  []string{},
			dataMode:     core.ArrayNode_SINGLE_INPUT_FILE,
			subNodeIndex: 0,
			expectedLiteralMap: &core.LiteralMap{
				Literals: map[string]*core.Literal{
					"foo": {
						Value: &core.Literal_OffloadedMetadata{
							OffloadedMetadata: &core.LiteralOffloadedMetadata{
								Uri:       "s3://bucket/offloaded",
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
					},
				},
			},
			expectInputBindings: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Setup mocks
			inputFilePaths := &pluginiomocks.InputFilePaths{}
			inputFilePaths.OnGetInputPath().Return(storage.DataReference("s3://bucket/input"))
			inputReader := newStaticInputReader(inputFilePaths, test.inputLiteralMap)

			nCtx := &nodesmocks.NodeExecutionContext{}
			nCtx.OnInputReader().Return(inputReader)
			nCtx.OnDataStore().Return(dataStore)

			// Create a mock subNodeSpec for constructInputBindings
			subNodeSpec := &v1alpha1.NodeSpec{}

			arrayNode := &mocks.ExecutableArrayNode{}
			arrayNode.OnGetBoundInputs().Return(test.boundInputs)
			arrayNode.OnGetDataMode().Return(test.dataMode)
			arrayNode.OnGetSubNodeSpec().Return(subNodeSpec)

			resolver := newSubNodeInputResolver(nCtx, arrayNode)
			err := resolver.Initialize(ctx)
			assert.NoError(t, err)

			subDataDir := storage.DataReference("s3://bucket/subnode")
			inputReaderResult, inputBindings, err := resolver.GetSubNodeInputs(ctx, test.subNodeIndex, subDataDir)
			assert.NoError(t, err)

			resultLiteralMap, err := inputReaderResult.Get(ctx)
			assert.NoError(t, err)
			assert.True(t, proto.Equal(test.expectedLiteralMap, resultLiteralMap))

			if !test.expectInputBindings {
				assert.Nil(t, inputBindings)
			}
		})
	}
}
