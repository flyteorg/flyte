package catalog

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func TestHashLiteralMap_LiteralsWithHashSet(t *testing.T) {
	tests := []struct {
		name            string
		literal         *core.Literal
		expectedLiteral *core.Literal
	}{
		{
			name:            "single literal where hash is not set",
			literal:         coreutils.MustMakeLiteral(42),
			expectedLiteral: coreutils.MustMakeLiteral(42),
		},
		{
			name: "single literal containing hash",
			literal: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_StructuredDataset{
							StructuredDataset: &core.StructuredDataset{
								Uri: "my-blob-stora://some-address",
								Metadata: &core.StructuredDatasetMetadata{
									StructuredDatasetType: &core.StructuredDatasetType{
										Format: "my-columnar-data-format",
									},
								},
							},
						},
					},
				},
				Hash: "abcde",
			},
			expectedLiteral: &core.Literal{
				Value: nil,
				Hash:  "abcde",
			},
		},
		{
			name: "list of literals containing a single item where literal sets its hash",
			literal: &core.Literal{
				Value: &core.Literal_Collection{
					Collection: &core.LiteralCollection{
						Literals: []*core.Literal{
							{
								Value: &core.Literal_Scalar{
									Scalar: &core.Scalar{
										Value: &core.Scalar_StructuredDataset{
											StructuredDataset: &core.StructuredDataset{
												Uri: "my-blob-stora://some-address",
												Metadata: &core.StructuredDatasetMetadata{
													StructuredDatasetType: &core.StructuredDatasetType{
														Format: "my-columnar-data-format",
													},
												},
											},
										},
									},
								},
								Hash: "hash1",
							},
						},
					},
				},
			},
			expectedLiteral: &core.Literal{
				Value: &core.Literal_Collection{
					Collection: &core.LiteralCollection{
						Literals: []*core.Literal{
							{
								Value: nil,
								Hash:  "hash1",
							},
						},
					},
				},
			},
		},
		{
			name: "list of literals containing two items where each literal sets its hash",
			literal: &core.Literal{
				Value: &core.Literal_Collection{
					Collection: &core.LiteralCollection{
						Literals: []*core.Literal{
							{
								Value: &core.Literal_Scalar{
									Scalar: &core.Scalar{
										Value: &core.Scalar_StructuredDataset{
											StructuredDataset: &core.StructuredDataset{
												Uri: "my-blob-stora://some-address",
												Metadata: &core.StructuredDatasetMetadata{
													StructuredDatasetType: &core.StructuredDatasetType{
														Format: "my-columnar-data-format",
													},
												},
											},
										},
									},
								},
								Hash: "hash1",
							},
							{
								Value: &core.Literal_Scalar{
									Scalar: &core.Scalar{
										Value: &core.Scalar_StructuredDataset{
											StructuredDataset: &core.StructuredDataset{
												Uri: "my-blob-stora://another-address",
												Metadata: &core.StructuredDatasetMetadata{
													StructuredDatasetType: &core.StructuredDatasetType{
														Format: "my-columnar-data-format",
													},
												},
											},
										},
									},
								},
								Hash: "hash2",
							},
						},
					},
				},
			},
			expectedLiteral: &core.Literal{
				Value: &core.Literal_Collection{
					Collection: &core.LiteralCollection{
						Literals: []*core.Literal{
							{
								Value: nil,
								Hash:  "hash1",
							},
							{
								Value: nil,
								Hash:  "hash2",
							},
						},
					},
				},
			},
		},
		{
			name: "list of literals containing two items where only one literal has its hash set",
			literal: &core.Literal{
				Value: &core.Literal_Collection{
					Collection: &core.LiteralCollection{
						Literals: []*core.Literal{
							{
								Value: &core.Literal_Scalar{
									Scalar: &core.Scalar{
										Value: &core.Scalar_StructuredDataset{
											StructuredDataset: &core.StructuredDataset{
												Uri: "my-blob-stora://some-address",
												Metadata: &core.StructuredDatasetMetadata{
													StructuredDatasetType: &core.StructuredDatasetType{
														Format: "my-columnar-data-format",
													},
												},
											},
										},
									},
								},
								Hash: "hash1",
							},
							{
								Value: &core.Literal_Scalar{
									Scalar: &core.Scalar{
										Value: &core.Scalar_StructuredDataset{
											StructuredDataset: &core.StructuredDataset{
												Uri: "my-blob-stora://another-address",
												Metadata: &core.StructuredDatasetMetadata{
													StructuredDatasetType: &core.StructuredDatasetType{
														Format: "my-columnar-data-format",
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
			expectedLiteral: &core.Literal{
				Value: &core.Literal_Collection{
					Collection: &core.LiteralCollection{
						Literals: []*core.Literal{
							{
								Value: nil,
								Hash:  "hash1",
							},
							{
								Value: &core.Literal_Scalar{
									Scalar: &core.Scalar{
										Value: &core.Scalar_StructuredDataset{
											StructuredDataset: &core.StructuredDataset{
												Uri: "my-blob-stora://another-address",
												Metadata: &core.StructuredDatasetMetadata{
													StructuredDatasetType: &core.StructuredDatasetType{
														Format: "my-columnar-data-format",
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
		{
			name: "map of literals containing a single item where literal sets its hash",
			literal: &core.Literal{
				Value: &core.Literal_Map{
					Map: &core.LiteralMap{
						Literals: map[string]*core.Literal{
							"literal1": {
								Value: &core.Literal_Scalar{
									Scalar: &core.Scalar{
										Value: &core.Scalar_StructuredDataset{
											StructuredDataset: &core.StructuredDataset{
												Uri: "my-blob-stora://some-address",
												Metadata: &core.StructuredDatasetMetadata{
													StructuredDatasetType: &core.StructuredDatasetType{
														Format: "my-columnar-data-format",
													},
												},
											},
										},
									},
								},
								Hash: "hash-42",
							},
						},
					},
				},
			},
			expectedLiteral: &core.Literal{
				Value: &core.Literal_Map{
					Map: &core.LiteralMap{
						Literals: map[string]*core.Literal{
							"literal1": {
								Value: nil,
								Hash:  "hash-42",
							},
						},
					},
				},
			},
		},
		{
			name: "map of literals containing a three items where only one literal sets its hash",
			literal: &core.Literal{
				Value: &core.Literal_Map{
					Map: &core.LiteralMap{
						Literals: map[string]*core.Literal{
							"literal1": {
								Value: &core.Literal_Scalar{
									Scalar: &core.Scalar{
										Value: &core.Scalar_StructuredDataset{
											StructuredDataset: &core.StructuredDataset{
												Uri: "my-blob-stora://some-address",
												Metadata: &core.StructuredDatasetMetadata{
													StructuredDatasetType: &core.StructuredDatasetType{
														Format: "my-columnar-data-format",
													},
												},
											},
										},
									},
								},
							},
							"literal2-set-its-hash": {
								Value: &core.Literal_Scalar{
									Scalar: &core.Scalar{
										Value: &core.Scalar_StructuredDataset{
											StructuredDataset: &core.StructuredDataset{
												Uri: "my-blob-stora://some-address-for-literal-2",
												Metadata: &core.StructuredDatasetMetadata{
													StructuredDatasetType: &core.StructuredDatasetType{
														Format: "my-columnar-data-format",
													},
												},
											},
										},
									},
								},
								Hash: "literal-2-hash",
							},
							"literal3": {
								Value: &core.Literal_Scalar{
									Scalar: &core.Scalar{
										Value: &core.Scalar_StructuredDataset{
											StructuredDataset: &core.StructuredDataset{
												Uri: "my-blob-stora://some-address-for-literal-3",
												Metadata: &core.StructuredDatasetMetadata{
													StructuredDatasetType: &core.StructuredDatasetType{
														Format: "my-columnar-data-format",
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
			expectedLiteral: &core.Literal{
				Value: &core.Literal_Map{
					Map: &core.LiteralMap{
						Literals: map[string]*core.Literal{
							"literal1": {
								Value: &core.Literal_Scalar{
									Scalar: &core.Scalar{
										Value: &core.Scalar_StructuredDataset{
											StructuredDataset: &core.StructuredDataset{
												Uri: "my-blob-stora://some-address",
												Metadata: &core.StructuredDatasetMetadata{
													StructuredDatasetType: &core.StructuredDatasetType{
														Format: "my-columnar-data-format",
													},
												},
											},
										},
									},
								},
							},
							"literal2-set-its-hash": {
								Value: nil,
								Hash:  "literal-2-hash",
							},
							"literal3": {
								Value: &core.Literal_Scalar{
									Scalar: &core.Scalar{
										Value: &core.Scalar_StructuredDataset{
											StructuredDataset: &core.StructuredDataset{
												Uri: "my-blob-stora://some-address-for-literal-3",
												Metadata: &core.StructuredDatasetMetadata{
													StructuredDatasetType: &core.StructuredDatasetType{
														Format: "my-columnar-data-format",
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
		{
			name: "list of map of literals containing a mixture of literals have their hashes set or not set",
			literal: &core.Literal{
				Value: &core.Literal_Collection{
					Collection: &core.LiteralCollection{
						Literals: []*core.Literal{
							{
								Value: &core.Literal_Map{
									Map: &core.LiteralMap{
										Literals: map[string]*core.Literal{
											"literal1": {
												Value: &core.Literal_Scalar{
													Scalar: &core.Scalar{
														Value: &core.Scalar_StructuredDataset{
															StructuredDataset: &core.StructuredDataset{
																Uri: "my-blob-stora://some-address",
																Metadata: &core.StructuredDatasetMetadata{
																	StructuredDatasetType: &core.StructuredDatasetType{
																		Format: "my-columnar-data-format",
																	},
																},
															},
														},
													},
												},
											},
											"literal2-set-its-hash": {
												Value: &core.Literal_Scalar{
													Scalar: &core.Scalar{
														Value: &core.Scalar_StructuredDataset{
															StructuredDataset: &core.StructuredDataset{
																Uri: "my-blob-stora://some-address-for-literal-2",
																Metadata: &core.StructuredDatasetMetadata{
																	StructuredDatasetType: &core.StructuredDatasetType{
																		Format: "my-columnar-data-format",
																	},
																},
															},
														},
													},
												},
												Hash: "literal-2-hash",
											},
											"literal3": {
												Value: &core.Literal_Scalar{
													Scalar: &core.Scalar{
														Value: &core.Scalar_StructuredDataset{
															StructuredDataset: &core.StructuredDataset{
																Uri: "my-blob-stora://some-address-for-literal-3",
																Metadata: &core.StructuredDatasetMetadata{
																	StructuredDatasetType: &core.StructuredDatasetType{
																		Format: "my-columnar-data-format",
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
							{
								Value: &core.Literal_Map{
									Map: &core.LiteralMap{
										Literals: map[string]*core.Literal{
											"another-literal-1": {
												Value: &core.Literal_Scalar{
													Scalar: &core.Scalar{
														Value: &core.Scalar_StructuredDataset{
															StructuredDataset: &core.StructuredDataset{
																Uri: "my-blob-stora://some-address-for-another-literal-1",
																Metadata: &core.StructuredDatasetMetadata{
																	StructuredDatasetType: &core.StructuredDatasetType{
																		Format: "my-columnar-data-format",
																	},
																},
															},
														},
													},
												},
												Hash: "another-literal-1-hash",
											},
											"another-literal2-set-its-hash": {
												Value: &core.Literal_Scalar{
													Scalar: &core.Scalar{
														Value: &core.Scalar_StructuredDataset{
															StructuredDataset: &core.StructuredDataset{
																Uri: "my-blob-stora://some-address-for-literal-2",
																Metadata: &core.StructuredDatasetMetadata{
																	StructuredDatasetType: &core.StructuredDatasetType{
																		Format: "my-columnar-data-format",
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
					},
				},
			},
			expectedLiteral: &core.Literal{
				Value: &core.Literal_Collection{
					Collection: &core.LiteralCollection{
						Literals: []*core.Literal{
							{
								Value: &core.Literal_Map{
									Map: &core.LiteralMap{
										Literals: map[string]*core.Literal{
											"literal1": {
												Value: &core.Literal_Scalar{
													Scalar: &core.Scalar{
														Value: &core.Scalar_StructuredDataset{
															StructuredDataset: &core.StructuredDataset{
																Uri: "my-blob-stora://some-address",
																Metadata: &core.StructuredDatasetMetadata{
																	StructuredDatasetType: &core.StructuredDatasetType{
																		Format: "my-columnar-data-format",
																	},
																},
															},
														},
													},
												},
											},
											"literal2-set-its-hash": {
												Value: nil,
												Hash:  "literal-2-hash",
											},
											"literal3": {
												Value: &core.Literal_Scalar{
													Scalar: &core.Scalar{
														Value: &core.Scalar_StructuredDataset{
															StructuredDataset: &core.StructuredDataset{
																Uri: "my-blob-stora://some-address-for-literal-3",
																Metadata: &core.StructuredDatasetMetadata{
																	StructuredDatasetType: &core.StructuredDatasetType{
																		Format: "my-columnar-data-format",
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
							{
								Value: &core.Literal_Map{
									Map: &core.LiteralMap{
										Literals: map[string]*core.Literal{
											"another-literal-1": {
												Value: nil,
												Hash:  "another-literal-1-hash",
											},
											"another-literal2-set-its-hash": {
												Value: &core.Literal_Scalar{
													Scalar: &core.Scalar{
														Value: &core.Scalar_StructuredDataset{
															StructuredDataset: &core.StructuredDataset{
																Uri: "my-blob-stora://some-address-for-literal-2",
																Metadata: &core.StructuredDatasetMetadata{
																	StructuredDatasetType: &core.StructuredDatasetType{
																		Format: "my-columnar-data-format",
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
					},
				},
			},
		},
		{
			name: "literal map containing hash",
			literal: &core.Literal{
				Value: &core.Literal_Map{
					Map: &core.LiteralMap{
						Literals: map[string]*core.Literal{
							"hello": {
								Value: &core.Literal_Scalar{
									Scalar: &core.Scalar{
										Value: &core.Scalar_Primitive{
											Primitive: &core.Primitive{
												Value: &core.Primitive_StringValue{
													StringValue: "world",
												},
											},
										},
									},
								},
							},
						},
					},
				},
				Hash: "0xffff",
			},
			expectedLiteral: &core.Literal{
				Value: nil,
				Hash:  "0xffff",
			},
		},
		{
			name: "literal collection containing hash",
			literal: &core.Literal{
				Value: &core.Literal_Collection{
					Collection: &core.LiteralCollection{
						Literals: []*core.Literal{
							{
								Value: &core.Literal_Scalar{
									Scalar: &core.Scalar{
										Value: &core.Scalar_Primitive{
											Primitive: &core.Primitive{
												Value: &core.Primitive_Integer{
													Integer: 42,
												},
											},
										},
									},
								},
							},
						},
					},
				},
				Hash: "0xabcdef",
			},
			expectedLiteral: &core.Literal{
				Value: nil,
				Hash:  "0xabcdef",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedLiteral, hashify(tt.literal))

			// Double-check that generating a tag is successful
			literalMap := &core.LiteralMap{Literals: map[string]*core.Literal{"o0": tt.literal}}
			hash, err := HashLiteralMap(context.TODO(), literalMap, nil)
			assert.NoError(t, err)
			assert.NotEmpty(t, hash)
		})
	}
}

// Ensure the key order on the inputs generates the same hash
func TestInputValueSorted(t *testing.T) {
	literalMap, err := coreutils.MakeLiteralMap(map[string]interface{}{"1": 1, "2": 2})
	assert.NoError(t, err)

	hash, err := HashLiteralMap(context.TODO(), literalMap, nil)
	assert.NoError(t, err)
	assert.Equal(t, "GQid5LjHbakcW68DS3P2jp80QLbiF0olFHF2hTh5bg8", hash)

	literalMap, err = coreutils.MakeLiteralMap(map[string]interface{}{"2": 2, "1": 1})
	assert.NoError(t, err)

	hashDupe, err := HashLiteralMap(context.TODO(), literalMap, nil)
	assert.NoError(t, err)
	assert.Equal(t, hashDupe, hash)
}

// Ensure that empty inputs are hashed the same way
func TestNoInputValues(t *testing.T) {
	hash, err := HashLiteralMap(context.TODO(), nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "GKw-c0PwFokMUQ6T-TUmEWnZ4_VlQ2Qpgw-vCTT0-OQ", hash)

	hashDupe, err := HashLiteralMap(context.TODO(), &core.LiteralMap{Literals: nil}, nil)
	assert.NoError(t, err)
	assert.Equal(t, "GKw-c0PwFokMUQ6T-TUmEWnZ4_VlQ2Qpgw-vCTT0-OQ", hashDupe)
	assert.Equal(t, hashDupe, hash)
}

// Ensure that msgpack Binary scalars with different key orderings produce the same hash
func TestHashLiteralMap_MsgpackBinaryDeterministic(t *testing.T) {
	// Manually constructed msgpack bytes for {a:1, b:2, c:3} with key order a, b, c
	msgpackABC := []byte{
		0x83,                   // fixmap, 3 entries
		0xa1, 0x61, 0x01,      // "a": 1
		0xa1, 0x62, 0x02,      // "b": 2
		0xa1, 0x63, 0x03,      // "c": 3
	}
	// Same map {a:1, b:2, c:3} with key order c, a, b
	msgpackCAB := []byte{
		0x83,                   // fixmap, 3 entries
		0xa1, 0x63, 0x03,      // "c": 3
		0xa1, 0x61, 0x01,      // "a": 1
		0xa1, 0x62, 0x02,      // "b": 2
	}

	litABC := &core.Literal{
		Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Binary{Binary: &core.Binary{Value: msgpackABC, Tag: "msgpack"}},
		}},
	}
	litCAB := &core.Literal{
		Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Binary{Binary: &core.Binary{Value: msgpackCAB, Tag: "msgpack"}},
		}},
	}

	hashABC, err := HashLiteralMap(context.TODO(), &core.LiteralMap{Literals: map[string]*core.Literal{"o0": litABC}}, nil)
	assert.NoError(t, err)
	hashCAB, err := HashLiteralMap(context.TODO(), &core.LiteralMap{Literals: map[string]*core.Literal{"o0": litCAB}}, nil)
	assert.NoError(t, err)
	assert.Equal(t, hashABC, hashCAB, "identical dicts with different msgpack key orderings must hash equally")

	// Nested: {a: {x:1, y:2}, b:3} with order a(x,y), b
	msgpackNested1 := []byte{
		0x82,                         // fixmap, 2 entries
		0xa1, 0x61,                   // "a"
		0x82, 0xa1, 0x78, 0x01, 0xa1, 0x79, 0x02, // {x:1, y:2}
		0xa1, 0x62, 0x03,             // "b": 3
	}
	// Same nested map with order b, a(y,x)
	msgpackNested2 := []byte{
		0x82,                         // fixmap, 2 entries
		0xa1, 0x62, 0x03,             // "b": 3
		0xa1, 0x61,                   // "a"
		0x82, 0xa1, 0x79, 0x02, 0xa1, 0x78, 0x01, // {y:2, x:1}
	}

	litNested1 := &core.Literal{
		Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Binary{Binary: &core.Binary{Value: msgpackNested1, Tag: "msgpack"}},
		}},
	}
	litNested2 := &core.Literal{
		Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Binary{Binary: &core.Binary{Value: msgpackNested2, Tag: "msgpack"}},
		}},
	}

	hashNested1, err := HashLiteralMap(context.TODO(), &core.LiteralMap{Literals: map[string]*core.Literal{"o0": litNested1}}, nil)
	assert.NoError(t, err)
	hashNested2, err := HashLiteralMap(context.TODO(), &core.LiteralMap{Literals: map[string]*core.Literal{"o0": litNested2}}, nil)
	assert.NoError(t, err)
	assert.Equal(t, hashNested1, hashNested2, "identical nested dicts with different msgpack key orderings must hash equally")
}

// Ensure that empty inputs are hashed the same way
func TestCacheIgnoreInputVars(t *testing.T) {
	literalMap, err := coreutils.MakeLiteralMap(map[string]interface{}{"1": 1, "2": 2})
	assert.NoError(t, err)

	hash, err := HashLiteralMap(context.TODO(), literalMap, nil)
	assert.NoError(t, err)
	assert.Equal(t, "GQid5LjHbakcW68DS3P2jp80QLbiF0olFHF2hTh5bg8", hash)

	literalMap, err = coreutils.MakeLiteralMap(map[string]interface{}{"2": 2, "1": 1, "3": 3})
	assert.NoError(t, err)

	hashDupe, err := HashLiteralMap(context.TODO(), literalMap, []string{"3"})
	assert.NoError(t, err)
	assert.Equal(t, hashDupe, hash)
}
