package validators

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/shamaton/msgpack/v2"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func TestLiteralTypeForLiterals(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		lt := literalTypeForLiterals(nil)
		assert.Equal(t, core.SimpleType_NONE.String(), lt.GetSimple().String())
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
		lt := LiteralTypeForLiteral(lv)
		assert.Equal(t, core.SimpleType_BINARY.String(), lt.GetSimple().String())
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
		lt := LiteralTypeForLiteral(lv)
		assert.Equal(t, core.SimpleType_STRUCT.String(), lt.GetSimple().String())
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
		lt := LiteralTypeForLiteral(lv)
		assert.Equal(t, core.SimpleType_STRUCT.String(), lt.GetSimple().String())
	})

	t.Run("homogeneous", func(t *testing.T) {
		lt := literalTypeForLiterals([]*core.Literal{
			coreutils.MustMakeLiteral(5),
			coreutils.MustMakeLiteral(0),
			coreutils.MustMakeLiteral(5),
		})

		assert.Equal(t, core.SimpleType_INTEGER.String(), lt.GetSimple().String())
	})

	t.Run("non-homogenous", func(t *testing.T) {
		lt := literalTypeForLiterals([]*core.Literal{
			coreutils.MustMakeLiteral("hello"),
			coreutils.MustMakeLiteral(5),
			coreutils.MustMakeLiteral("world"),
			coreutils.MustMakeLiteral(0),
			coreutils.MustMakeLiteral(2),
		})

		assert.Len(t, lt.GetUnionType().Variants, 2)
		assert.Equal(t, core.SimpleType_INTEGER.String(), lt.GetUnionType().Variants[0].GetSimple().String())
		assert.Equal(t, core.SimpleType_STRING.String(), lt.GetUnionType().Variants[1].GetSimple().String())
	})

	t.Run("non-homogenous ensure ordering", func(t *testing.T) {
		lt := literalTypeForLiterals([]*core.Literal{
			coreutils.MustMakeLiteral(5),
			coreutils.MustMakeLiteral("world"),
			coreutils.MustMakeLiteral(0),
			coreutils.MustMakeLiteral(2),
		})

		assert.Len(t, lt.GetUnionType().Variants, 2)
		assert.Equal(t, core.SimpleType_INTEGER.String(), lt.GetUnionType().Variants[0].GetSimple().String())
		assert.Equal(t, core.SimpleType_STRING.String(), lt.GetUnionType().Variants[1].GetSimple().String())
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

		lt := LiteralTypeForLiteral(literals)

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

		assert.True(t, proto.Equal(expectedLt, lt))
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

		lt := LiteralTypeForLiteral(literals)

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

		assert.True(t, proto.Equal(expectedLt, lt))
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

		lt := LiteralTypeForLiteral(literals)

		assert.True(t, proto.Equal(expectedLt, lt))
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

		lt := LiteralTypeForLiteral(literals)

		assert.True(t, proto.Equal(expectedLt, lt))
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
		lt := LiteralTypeForLiteral(literals)
		assert.True(t, proto.Equal(expectedLt, lt))
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
