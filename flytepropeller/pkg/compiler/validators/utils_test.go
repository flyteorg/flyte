package validators

import (
	"testing"

	"github.com/shamaton/msgpack/v2"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func TestIsInstance(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		assert.True(t, IsInstance(nil, &core.LiteralType{
			Type: &core.LiteralType_Simple{
				Simple: core.SimpleType_NONE,
			},
		}))
	})

	t.Run("binary idl with raw binary data and no tag", func(t *testing.T) {
		// Some arbitrary raw binary data
		rawBinaryData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

		lv := &core.Literal{
			Value: &core.Literal_Scalar{
				Scalar: &core.Scalar{
					Value: &core.Scalar_Binary{
						Binary: &core.Binary{
							Value: rawBinaryData,
							Tag:   "",
						},
					},
				},
			},
		}
		assert.True(t, IsInstance(lv, &core.LiteralType{
			Type: &core.LiteralType_Simple{
				Simple: core.SimpleType_BINARY,
			},
		}))
	})

	t.Run("binary idl with messagepack input map[int]strings", func(t *testing.T) {
		// Create a map[int]string and serialize it using MessagePack.
		data := map[int]string{
			1:  "hello",
			2:  "world",
			-1: "foo",
		}
		// Serializing the map using MessagePack
		serializedBinaryData, err := msgpack.Marshal(data)
		if err != nil {
			t.Fatalf("failed to serialize map: %v", err)
		}
		lv := &core.Literal{
			Value: &core.Literal_Scalar{
				Scalar: &core.Scalar{
					Value: &core.Scalar_Binary{
						Binary: &core.Binary{
							Value: serializedBinaryData,
							Tag:   coreutils.MESSAGEPACK,
						},
					},
				},
			},
		}
		assert.True(t, IsInstance(lv, &core.LiteralType{
			Type: &core.LiteralType_Simple{
				Simple: core.SimpleType_STRUCT,
			},
		}))
	})

	t.Run("binary idl with messagepack input map[float]strings", func(t *testing.T) {
		// Create a map[float]string and serialize it using MessagePack.
		data := map[float64]string{
			1.0:  "hello",
			5.0:  "world",
			-1.0: "foo",
		}
		// Serializing the map using MessagePack
		serializedBinaryData, err := msgpack.Marshal(data)
		if err != nil {
			t.Fatalf("failed to serialize map: %v", err)
		}
		lv := &core.Literal{
			Value: &core.Literal_Scalar{
				Scalar: &core.Scalar{
					Value: &core.Scalar_Binary{
						Binary: &core.Binary{
							Value: serializedBinaryData,
							Tag:   coreutils.MESSAGEPACK,
						},
					},
				},
			},
		}
		assert.True(t, IsInstance(lv, &core.LiteralType{
			Type: &core.LiteralType_Simple{
				Simple: core.SimpleType_STRUCT,
			},
		}))
	})

	t.Run("list with mixed types", func(t *testing.T) {
		literals := &core.Literal{
			Value: &core.Literal_Collection{
				Collection: &core.LiteralCollection{
					Literals: []*core.Literal{
						{
							Value: &core.Literal_Scalar{
								Scalar: &core.Scalar{
									Value: &core.Scalar_Union{
										Union: &core.Union{
											Value: &core.Literal{
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
											},
											Type: &core.LiteralType{
												Type: &core.LiteralType_Simple{
													Simple: core.SimpleType_INTEGER,
												},
											},
										},
									},
								},
							},
						},
						{
							Value: &core.Literal_Scalar{
								Scalar: &core.Scalar{
									Value: &core.Scalar_Union{
										Union: &core.Union{
											Value: &core.Literal{
												Value: &core.Literal_Scalar{
													Scalar: &core.Scalar{
														Value: &core.Scalar_Primitive{
															Primitive: &core.Primitive{
																Value: &core.Primitive_StringValue{
																	StringValue: "foo",
																},
															},
														},
													},
												},
											},
											Type: &core.LiteralType{
												Type: &core.LiteralType_Simple{
													Simple: core.SimpleType_STRING,
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
		}

		expectedLt := &core.LiteralType{
			Type: &core.LiteralType_CollectionType{
				CollectionType: &core.LiteralType{
					Type: &core.LiteralType_UnionType{
						UnionType: &core.UnionType{
							Variants: []*core.LiteralType{
								{
									Type: &core.LiteralType_UnionType{
										UnionType: &core.UnionType{
											Variants: []*core.LiteralType{
												{
													Type: &core.LiteralType_Simple{
														Simple: core.SimpleType_INTEGER,
													},
												},
											},
										},
									},
								},
								{
									Type: &core.LiteralType_UnionType{
										UnionType: &core.UnionType{
											Variants: []*core.LiteralType{
												{
													Type: &core.LiteralType_Simple{
														Simple: core.SimpleType_STRING,
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
		}

		assert.True(t, IsInstance(literals, expectedLt))
	})

	t.Run("nested lists with empty list", func(t *testing.T) {
		literals := &core.Literal{
			Value: &core.Literal_Collection{
				Collection: &core.LiteralCollection{
					Literals: []*core.Literal{
						{
							Value: &core.Literal_Collection{
								Collection: &core.LiteralCollection{
									Literals: []*core.Literal{
										{
											Value: &core.Literal_Scalar{
												Scalar: &core.Scalar{
													Value: &core.Scalar_Primitive{
														Primitive: &core.Primitive{
															Value: &core.Primitive_StringValue{
																StringValue: "foo",
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
							Value: &core.Literal_Collection{
								Collection: &core.LiteralCollection{},
							},
						},
					},
				},
			},
		}

		expectedLt := &core.LiteralType{
			Type: &core.LiteralType_CollectionType{
				CollectionType: &core.LiteralType{
					Type: &core.LiteralType_CollectionType{
						CollectionType: &core.LiteralType{
							Type: &core.LiteralType_Simple{
								Simple: core.SimpleType_STRING,
							},
						},
					},
				},
			},
		}

		assert.True(t, IsInstance(literals, expectedLt))
	})

	t.Run("nested Lists with different types", func(t *testing.T) {
		literals := &core.Literal{
			Value: &core.Literal_Collection{
				Collection: &core.LiteralCollection{
					Literals: []*core.Literal{
						{
							Value: &core.Literal_Collection{
								Collection: &core.LiteralCollection{
									Literals: []*core.Literal{
										{
											Value: &core.Literal_Scalar{
												Scalar: &core.Scalar{
													Value: &core.Scalar_Union{
														Union: &core.Union{
															Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}},
															Value: &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{
																Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_Integer{Integer: 1}}}}}},
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
							Value: &core.Literal_Collection{
								Collection: &core.LiteralCollection{
									Literals: []*core.Literal{
										{
											Value: &core.Literal_Scalar{
												Scalar: &core.Scalar{
													Value: &core.Scalar_Union{
														Union: &core.Union{
															Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
															Value: &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{
																Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_StringValue{StringValue: "foo"}}}}}},
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
		}

		expectedLt := &core.LiteralType{
			Type: &core.LiteralType_CollectionType{
				CollectionType: &core.LiteralType{
					Type: &core.LiteralType_CollectionType{
						CollectionType: &core.LiteralType{
							Type: &core.LiteralType_UnionType{
								UnionType: &core.UnionType{
									Variants: []*core.LiteralType{
										{
											Type: &core.LiteralType_Simple{
												Simple: core.SimpleType_INTEGER,
											},
										},
										{
											Type: &core.LiteralType_Simple{
												Simple: core.SimpleType_STRING,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		assert.True(t, IsInstance(literals, expectedLt))
	})

	t.Run("empty nested listed", func(t *testing.T) {
		literals := &core.Literal{
			Value: &core.Literal_Collection{
				Collection: &core.LiteralCollection{
					Literals: []*core.Literal{
						{
							Value: &core.Literal_Collection{
								Collection: &core.LiteralCollection{},
							},
						},
					},
				},
			},
		}

		expectedLt := &core.LiteralType{
			Type: &core.LiteralType_CollectionType{
				CollectionType: &core.LiteralType{
					Type: &core.LiteralType_CollectionType{
						CollectionType: &core.LiteralType{
							Type: &core.LiteralType_Simple{
								Simple: core.SimpleType_NONE,
							},
						},
					},
				},
			},
		}

		assert.True(t, IsInstance(literals, expectedLt))
	})

	t.Run("nested Lists with different types", func(t *testing.T) {
		inferredType := &core.LiteralType{
			Type: &core.LiteralType_CollectionType{
				CollectionType: &core.LiteralType{
					Type: &core.LiteralType_CollectionType{
						CollectionType: &core.LiteralType{
							Type: &core.LiteralType_UnionType{
								UnionType: &core.UnionType{
									Variants: []*core.LiteralType{
										{
											Type: &core.LiteralType_Simple{
												Simple: core.SimpleType_INTEGER,
											},
										},
										{
											Type: &core.LiteralType_Simple{
												Simple: core.SimpleType_STRING,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		literals := &core.Literal{
			Value: &core.Literal_OffloadedMetadata{
				OffloadedMetadata: &core.LiteralOffloadedMetadata{
					Uri:          "dummy/uri",
					SizeBytes:    1000,
					InferredType: inferredType,
				},
			},
		}
		expectedLt := inferredType
		assert.True(t, IsInstance(literals, expectedLt))
	})

}

func TestJoinVariableMapsUniqueKeys(t *testing.T) {
	intType := &core.LiteralType{
		Type: &core.LiteralType_Simple{
			Simple: core.SimpleType_INTEGER,
		},
	}

	strType := &core.LiteralType{
		Type: &core.LiteralType_Simple{
			Simple: core.SimpleType_STRING,
		},
	}

	t.Run("Simple", func(t *testing.T) {
		m1 := map[string]*core.Variable{
			"x": {
				Type: intType,
			},
		}

		m2 := map[string]*core.Variable{
			"y": {
				Type: intType,
			},
		}

		res, err := UnionDistinctVariableMaps(m1, m2)
		assert.NoError(t, err)
		assert.Len(t, res, 2)
	})

	t.Run("No type collision", func(t *testing.T) {
		m1 := map[string]*core.Variable{
			"x": {
				Type: intType,
			},
		}

		m2 := map[string]*core.Variable{
			"x": {
				Type: intType,
			},
		}

		res, err := UnionDistinctVariableMaps(m1, m2)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("Type collision", func(t *testing.T) {
		m1 := map[string]*core.Variable{
			"x": {
				Type: intType,
			},
		}

		m2 := map[string]*core.Variable{
			"x": {
				Type: strType,
			},
		}

		_, err := UnionDistinctVariableMaps(m1, m2)
		assert.Error(t, err)
	})
}
