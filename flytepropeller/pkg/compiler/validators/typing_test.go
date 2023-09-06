package validators

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/stretchr/testify/assert"
)

func TestSimpleLiteralCasting(t *testing.T) {
	t.Run("BaseCase_Integer", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
			},
			&core.LiteralType{
				Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
			},
		)
		assert.True(t, castable, "Integers should be castable to other integers")
	})

	t.Run("IntegerToFloat", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
			},
			&core.LiteralType{
				Type: &core.LiteralType_Simple{Simple: core.SimpleType_FLOAT},
			},
		)
		assert.False(t, castable, "Integers should not be castable to floats")
	})

	t.Run("FloatToInteger", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_Simple{Simple: core.SimpleType_FLOAT},
			},
			&core.LiteralType{
				Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
			},
		)
		assert.False(t, castable, "Floats should not be castable to integers")
	})

	t.Run("VoidToInteger", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_Simple{Simple: core.SimpleType_NONE},
			},
			&core.LiteralType{
				Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
			},
		)
		assert.False(t, castable, "Non-optional types are non-nullable")
	})

	t.Run("IgnoreMetadata", func(t *testing.T) {
		s := structpb.Struct{
			Fields: map[string]*structpb.Value{
				"a": {},
			},
		}
		castable := AreTypesCastable(
			&core.LiteralType{
				Type:     &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
				Metadata: &s,
			},
			&core.LiteralType{
				Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
			},
		)
		assert.True(t, castable, "Metadata should be ignored")
	})

	t.Run("EnumToString", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_EnumType{EnumType: &core.EnumType{
					Values: []string{"x", "y"},
				}},
			},
			&core.LiteralType{
				Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
			},
		)
		assert.True(t, castable, "Enum should be castable to string")
	})

	t.Run("EnumToEnum", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_EnumType{EnumType: &core.EnumType{
					Values: []string{"x", "y"},
				}},
			},
			&core.LiteralType{
				Type: &core.LiteralType_EnumType{EnumType: &core.EnumType{
					Values: []string{"x", "y"},
				}},
			},
		)
		assert.True(t, castable, "Enum should be castable to Enums if they are identical")
	})

	t.Run("EnumToEnum", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_EnumType{EnumType: &core.EnumType{
					Values: []string{"x", "y"},
				}},
			},
			&core.LiteralType{
				Type: &core.LiteralType_EnumType{EnumType: &core.EnumType{
					Values: []string{"m", "n"},
				}},
			},
		)
		assert.False(t, castable, "Enum should not be castable to non matching enums")
	})

	t.Run("StringToEnum", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
			},
			&core.LiteralType{
				Type: &core.LiteralType_EnumType{EnumType: &core.EnumType{
					Values: []string{"x", "y"},
				}},
			},
		)
		assert.True(t, castable, "Strings should be castable to enums - may result in runtime failure")
	})
}

func TestUnionCasting(t *testing.T) {
	t.Run("StringToUnionUnambiguously", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
			},
			&core.LiteralType{
				Type: &core.LiteralType_UnionType{
					UnionType: &core.UnionType{
						Variants: []*core.LiteralType{
							{
								Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
								Structure: &core.TypeStructure{
									Tag: "int",
								},
							},
							{
								Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
								Structure: &core.TypeStructure{
									Tag: "str",
								},
							},
						},
					},
				},
			},
		)
		assert.True(t, castable, "Strings should be castable to (str | int)")
	})

	t.Run("StringToUnionAmbiguously", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
			},
			&core.LiteralType{
				Type: &core.LiteralType_UnionType{
					UnionType: &core.UnionType{
						Variants: []*core.LiteralType{
							{
								Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
								Structure: &core.TypeStructure{
									Tag: "str1",
								},
							},
							{
								Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
								Structure: &core.TypeStructure{
									Tag: "str2",
								},
							},
						},
					},
				},
			},
		)
		assert.False(t, castable, "Raw string literals should not be ambiguously castable to (str | str)")
	})

	t.Run("UnionToUnionSuperset", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_UnionType{
					UnionType: &core.UnionType{
						Variants: []*core.LiteralType{
							{
								Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
								Structure: &core.TypeStructure{
									Tag: "str1",
								},
							},
							{
								Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
								Structure: &core.TypeStructure{
									Tag: "str2",
								},
							},
						},
					},
				},
			},
			&core.LiteralType{
				Type: &core.LiteralType_UnionType{
					UnionType: &core.UnionType{
						Variants: []*core.LiteralType{
							{
								Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
								Structure: &core.TypeStructure{
									Tag: "str1",
								},
							},
							{
								Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
								Structure: &core.TypeStructure{
									Tag: "int1",
								},
							},
							{
								Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
								Structure: &core.TypeStructure{
									Tag: "str2",
								},
							},
						},
					},
				},
			},
		)
		assert.True(t, castable, "Union types can be cast to a union that contains a superset of variants")
	})

	t.Run("UnionToUnionTagMismatch", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_UnionType{
					UnionType: &core.UnionType{
						Variants: []*core.LiteralType{
							{
								Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
								Structure: &core.TypeStructure{
									Tag: "str1",
								},
							},
							{
								Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
								Structure: &core.TypeStructure{
									Tag: "str2",
								},
							},
						},
					},
				},
			},
			&core.LiteralType{
				Type: &core.LiteralType_UnionType{
					UnionType: &core.UnionType{
						Variants: []*core.LiteralType{
							{
								Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
								Structure: &core.TypeStructure{
									Tag: "str2",
								},
							},
							{
								Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
								Structure: &core.TypeStructure{
									Tag: "str3",
								},
							},
						},
					},
				},
			},
		)
		assert.False(t, castable, "Union types can only be cast to a union that contains a superset of variants")
	})

	t.Run("UnionToUnionTypeMismatch", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_UnionType{
					UnionType: &core.UnionType{
						Variants: []*core.LiteralType{
							{
								Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
								Structure: &core.TypeStructure{
									Tag: "test",
								},
							},
						},
					},
				},
			},
			&core.LiteralType{
				Type: &core.LiteralType_UnionType{
					UnionType: &core.UnionType{
						Variants: []*core.LiteralType{
							{
								Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
								Structure: &core.TypeStructure{
									Tag: "test",
								},
							},
						},
					},
				},
			},
		)
		assert.False(t, castable, "Union types can only be cast to a union that contains a superset of variants")
	})

	t.Run("SingularUnionToUnderlyingType", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_UnionType{
					UnionType: &core.UnionType{
						Variants: []*core.LiteralType{
							{
								Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
								Structure: &core.TypeStructure{
									Tag: "string",
								},
							},
						},
					},
				},
			},
			&core.LiteralType{
				Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
				Structure: &core.TypeStructure{
					Tag: "string",
				},
			},
		)
		assert.True(t, castable, "Singular unions should be castable to their underlying type")
	})
}

func TestCollectionCasting(t *testing.T) {
	t.Run("BaseCase_SingleIntegerCollection", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_CollectionType{
					CollectionType: &core.LiteralType{
						Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
					},
				},
			},
			&core.LiteralType{
				Type: &core.LiteralType_CollectionType{
					CollectionType: &core.LiteralType{
						Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
					},
				},
			},
		)
		assert.True(t, castable, "[Integer] should be castable to [Integer].")
	})

	t.Run("Empty collection", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_CollectionType{
					CollectionType: &core.LiteralType{
						Type: &core.LiteralType_Simple{Simple: core.SimpleType_NONE},
					},
				},
			},
			&core.LiteralType{
				Type: &core.LiteralType_CollectionType{
					CollectionType: &core.LiteralType{
						Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
					},
				},
			},
		)
		assert.True(t, castable, "[] should be castable to [Integer].")
	})

	t.Run("SingleIntegerCollectionToSingleFloatCollection", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_CollectionType{
					CollectionType: &core.LiteralType{
						Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
					},
				},
			},
			&core.LiteralType{
				Type: &core.LiteralType_CollectionType{
					CollectionType: &core.LiteralType{
						Type: &core.LiteralType_Simple{Simple: core.SimpleType_FLOAT},
					},
				},
			},
		)
		assert.False(t, castable, "[Integer] should not be castable to [Float]")
	})

	t.Run("MismatchedNestLevels_Scalar", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_CollectionType{
					CollectionType: &core.LiteralType{
						Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
					},
				},
			},
			&core.LiteralType{
				Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
			},
		)
		assert.False(t, castable, "[Integer] should not be castable to Integer")
	})

	t.Run("MismatchedNestLevels_Collections", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_CollectionType{
					CollectionType: &core.LiteralType{
						Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
					},
				},
			},
			&core.LiteralType{
				Type: &core.LiteralType_CollectionType{
					CollectionType: &core.LiteralType{
						Type: &core.LiteralType_CollectionType{
							CollectionType: &core.LiteralType{
								Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
							},
						},
					},
				},
			},
		)
		assert.False(t, castable, "[Integer] should not be castable to [[Integer]]")
	})

	t.Run("Nullable_Collections", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_Simple{
					Simple: core.SimpleType_NONE,
				},
			},
			&core.LiteralType{
				Type: &core.LiteralType_CollectionType{
					CollectionType: &core.LiteralType{
						Type: &core.LiteralType_CollectionType{
							CollectionType: &core.LiteralType{
								Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
							},
						},
					},
				},
			},
		)
		assert.False(t, castable, "Non-optional collections are not nullable")
	})
}

func TestMapCasting(t *testing.T) {
	t.Run("BaseCase_SingleIntegerMap", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_MapValueType{
					MapValueType: &core.LiteralType{
						Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
					},
				},
			},
			&core.LiteralType{
				Type: &core.LiteralType_MapValueType{
					MapValueType: &core.LiteralType{
						Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
					},
				},
			},
		)
		assert.True(t, castable, "{k: Integer} should be castable to {k: Integer}.")
	})

	t.Run("ScalarIntegerMapToScalarFloatMap", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_MapValueType{
					MapValueType: &core.LiteralType{
						Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
					},
				},
			},
			&core.LiteralType{
				Type: &core.LiteralType_MapValueType{
					MapValueType: &core.LiteralType{
						Type: &core.LiteralType_Simple{Simple: core.SimpleType_FLOAT},
					},
				},
			},
		)
		assert.False(t, castable, "{k: Integer} should not be castable to {k: Float}")
	})

	t.Run("ScalarIntegerMapToScalarFloatMap", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_MapValueType{
					MapValueType: &core.LiteralType{
						Type: &core.LiteralType_Simple{Simple: core.SimpleType_NONE},
					},
				},
			},
			&core.LiteralType{
				Type: &core.LiteralType_MapValueType{
					MapValueType: &core.LiteralType{
						Type: &core.LiteralType_Simple{Simple: core.SimpleType_FLOAT},
					},
				},
			},
		)

		assert.True(t, castable, "{k: None} should be castable to {k: Float}")
	})

	t.Run("ScalarStructToStruct", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_Simple{
					Simple: core.SimpleType_STRUCT,
				},
			},
			&core.LiteralType{
				Type: &core.LiteralType_Simple{
					Simple: core.SimpleType_STRUCT,
				},
			},
		)
		assert.True(t, castable, "castable from Struct to struct")
	})

	t.Run("MismatchedMapNestLevels_Scalar", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_MapValueType{
					MapValueType: &core.LiteralType{
						Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
					},
				},
			},
			&core.LiteralType{
				Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
			},
		)
		assert.False(t, castable, "{k: Integer} should not be castable to Integer")
	})

	t.Run("MismatchedMapNestLevels_Maps", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_MapValueType{
					MapValueType: &core.LiteralType{
						Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
					},
				},
			},
			&core.LiteralType{
				Type: &core.LiteralType_MapValueType{
					MapValueType: &core.LiteralType{
						Type: &core.LiteralType_MapValueType{
							MapValueType: &core.LiteralType{
								Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
							},
						},
					},
				},
			},
		)
		assert.False(t, castable, "{k: Integer} should not be castable to {k: {k: Integer}}")
	})
}

func TestSchemaCasting(t *testing.T) {
	genericSchema := &core.LiteralType{
		Type: &core.LiteralType_Schema{
			Schema: &core.SchemaType{
				Columns: []*core.SchemaType_SchemaColumn{},
			},
		},
	}
	genericStructuredDataset := &core.LiteralType{
		Type: &core.LiteralType_StructuredDatasetType{
			StructuredDatasetType: &core.StructuredDatasetType{
				Columns: []*core.StructuredDatasetType_DatasetColumn{},
				Format:  "",
			},
		},
	}
	subsetIntegerSchema := &core.LiteralType{
		Type: &core.LiteralType_Schema{
			Schema: &core.SchemaType{
				Columns: []*core.SchemaType_SchemaColumn{
					{
						Name: "a",
						Type: core.SchemaType_SchemaColumn_INTEGER,
					},
				},
			},
		},
	}
	supersetIntegerAndFloatSchema := &core.LiteralType{
		Type: &core.LiteralType_Schema{
			Schema: &core.SchemaType{
				Columns: []*core.SchemaType_SchemaColumn{
					{
						Name: "a",
						Type: core.SchemaType_SchemaColumn_INTEGER,
					},
					{
						Name: "b",
						Type: core.SchemaType_SchemaColumn_FLOAT,
					},
				},
			},
		},
	}
	supersetStructuredDataset := &core.LiteralType{
		Type: &core.LiteralType_StructuredDatasetType{
			StructuredDatasetType: &core.StructuredDatasetType{
				Columns: []*core.StructuredDatasetType_DatasetColumn{
					{
						Name:        "a",
						LiteralType: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}},
					},
					{
						Name:        "b",
						LiteralType: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_FLOAT}},
					},
				},
				Format: "parquet",
			},
		},
	}
	mismatchedSubsetSchema := &core.LiteralType{
		Type: &core.LiteralType_Schema{
			Schema: &core.SchemaType{
				Columns: []*core.SchemaType_SchemaColumn{
					{
						Name: "a",
						Type: core.SchemaType_SchemaColumn_FLOAT,
					},
				},
			},
		},
	}

	t.Run("BaseCase_GenericSchema", func(t *testing.T) {
		castable := AreTypesCastable(genericSchema, genericSchema)
		assert.True(t, castable, "Schema() should be castable to Schema()")
	})

	t.Run("GenericSchemaToNonGeneric", func(t *testing.T) {
		castable := AreTypesCastable(genericSchema, subsetIntegerSchema)
		assert.True(t, castable, "Schema() should be castable to Schema(a=Integer)")
	})

	t.Run("NonGenericSchemaToGeneric", func(t *testing.T) {
		castable := AreTypesCastable(subsetIntegerSchema, genericSchema)
		assert.True(t, castable, "Schema(a=Integer) should be castable to Schema()")
	})

	t.Run("SupersetToSubsetTypedSchema", func(t *testing.T) {
		castable := AreTypesCastable(supersetIntegerAndFloatSchema, subsetIntegerSchema)
		assert.True(t, castable, "Schema(a=Integer, b=Float) should be castable to Schema(a=Integer)")
	})

	t.Run("GenericToSubsetTypedSchema", func(t *testing.T) {
		castable := AreTypesCastable(genericStructuredDataset, subsetIntegerSchema)
		assert.True(t, castable, "StructuredDataset() with generic format should be castable to Schema(a=Integer)")
	})

	t.Run("SubsetTypedSchemaToGeneric", func(t *testing.T) {
		castable := AreTypesCastable(subsetIntegerSchema, genericStructuredDataset)
		assert.True(t, castable, "Schema(a=Integer) should be castable to StructuredDataset() with generic format")
	})

	t.Run("SupersetStructuredToSubsetTypedSchema", func(t *testing.T) {
		castable := AreTypesCastable(supersetStructuredDataset, subsetIntegerSchema)
		assert.True(t, castable, "StructuredDataset(a=Integer, b=Float) should be castable to Schema(a=Integer)")
	})

	t.Run("SubsetToSupersetSchema", func(t *testing.T) {
		castable := AreTypesCastable(subsetIntegerSchema, supersetIntegerAndFloatSchema)
		assert.False(t, castable, "Schema(a=Integer) should not be castable to Schema(a=Integer, b=Float)")
	})

	t.Run("MismatchedColumns", func(t *testing.T) {
		castable := AreTypesCastable(subsetIntegerSchema, mismatchedSubsetSchema)
		assert.False(t, castable, "Schema(a=Integer) should not be castable to Schema(a=Float)")
	})

	t.Run("MismatchedColumnsFlipped", func(t *testing.T) {
		castable := AreTypesCastable(mismatchedSubsetSchema, subsetIntegerSchema)
		assert.False(t, castable, "Schema(a=Float) should not be castable to Schema(a=Integer)")
	})

	t.Run("SchemasAreNullable", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_Simple{
					Simple: core.SimpleType_NONE,
				},
			},
			subsetIntegerSchema)
		assert.False(t, castable, "Non-optional schemas are not nullable")
	})
}

func TestStructuredDatasetCasting(t *testing.T) {
	emptyStructuredDataset := &core.LiteralType{
		Type: &core.LiteralType_StructuredDatasetType{
			StructuredDatasetType: &core.StructuredDatasetType{
				Columns: []*core.StructuredDatasetType_DatasetColumn{},
				Format:  "",
			},
		},
	}
	genericStructuredDataset := &core.LiteralType{
		Type: &core.LiteralType_StructuredDatasetType{
			StructuredDatasetType: &core.StructuredDatasetType{
				Columns: []*core.StructuredDatasetType_DatasetColumn{},
				Format:  "parquet",
			},
		},
	}
	subsetStructuredDataset := &core.LiteralType{
		Type: &core.LiteralType_StructuredDatasetType{
			StructuredDatasetType: &core.StructuredDatasetType{
				Columns: []*core.StructuredDatasetType_DatasetColumn{
					{
						Name:        "a",
						LiteralType: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}},
					},
					{
						Name:        "b",
						LiteralType: &core.LiteralType{Type: &core.LiteralType_CollectionType{CollectionType: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
					},
				},
				Format: "parquet",
			},
		},
	}
	supersetStructuredDataset := &core.LiteralType{
		Type: &core.LiteralType_StructuredDatasetType{
			StructuredDatasetType: &core.StructuredDatasetType{
				Columns: []*core.StructuredDatasetType_DatasetColumn{
					{
						Name:        "a",
						LiteralType: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}},
					},
					{
						Name:        "b",
						LiteralType: &core.LiteralType{Type: &core.LiteralType_CollectionType{CollectionType: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
					},
					{
						Name:        "c",
						LiteralType: &core.LiteralType{Type: &core.LiteralType_MapValueType{MapValueType: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
					},
				},
				Format: "parquet",
			},
		},
	}
	integerSchema := &core.LiteralType{
		Type: &core.LiteralType_Schema{
			Schema: &core.SchemaType{
				Columns: []*core.SchemaType_SchemaColumn{
					{
						Name: "a",
						Type: core.SchemaType_SchemaColumn_INTEGER,
					},
				},
			},
		},
	}
	integerStructuredDataset := &core.LiteralType{
		Type: &core.LiteralType_StructuredDatasetType{
			StructuredDatasetType: &core.StructuredDatasetType{
				Columns: []*core.StructuredDatasetType_DatasetColumn{
					{
						Name:        "a",
						LiteralType: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}},
					},
				},
				Format: "parquet",
			},
		},
	}
	mismatchedSubsetStructuredDataset := &core.LiteralType{
		Type: &core.LiteralType_StructuredDatasetType{
			StructuredDatasetType: &core.StructuredDatasetType{
				Columns: []*core.StructuredDatasetType_DatasetColumn{
					{
						Name:        "a",
						LiteralType: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_FLOAT}},
					},
				},
			},
		},
	}

	t.Run("BaseCase_GenericStructuredDataset", func(t *testing.T) {
		castable := AreTypesCastable(genericStructuredDataset, genericStructuredDataset)
		assert.True(t, castable, "StructuredDataset() should be castable to StructuredDataset()")
	})

	t.Run("GenericStructuredDatasetToNonGeneric", func(t *testing.T) {
		castable := AreTypesCastable(genericStructuredDataset, subsetStructuredDataset)
		assert.True(t, castable, "StructuredDataset() should be castable to StructuredDataset(a=Integer, b=Collection)")
	})

	t.Run("NonGenericStructuredDatasetToGeneric", func(t *testing.T) {
		castable := AreTypesCastable(subsetStructuredDataset, genericStructuredDataset)
		assert.True(t, castable, "StructuredDataset(a=Integer, b=Collection) should be castable to StructuredDataset()")
	})

	t.Run("SupersetToSubsetTypedStructuredDataset", func(t *testing.T) {
		castable := AreTypesCastable(supersetStructuredDataset, subsetStructuredDataset)
		assert.True(t, castable, "StructuredDataset(a=Integer, b=Collection, c=Map) should be castable to StructuredDataset(a=Integer, b=Collection)")
	})

	t.Run("SubsetToSupersetStructuredDataset", func(t *testing.T) {
		castable := AreTypesCastable(subsetStructuredDataset, supersetStructuredDataset)
		assert.False(t, castable, "StructuredDataset(a=Integer, b=Collection) should not be castable to StructuredDataset(a=Integer, b=Collection, c=Map)")
	})

	t.Run("SchemaToStructuredDataset", func(t *testing.T) {
		castable := AreTypesCastable(integerSchema, integerStructuredDataset)
		assert.True(t, castable, "Schema(a=Integer) should be castable to StructuredDataset(a=Integer)")
	})

	t.Run("MismatchedSchemaColumns", func(t *testing.T) {
		castable := AreTypesCastable(integerSchema, mismatchedSubsetStructuredDataset)
		assert.False(t, castable, "Schema(a=Integer) should not be castable to StructuredDataset(a=Float)")
	})

	t.Run("MismatchedColumns", func(t *testing.T) {
		castable := AreTypesCastable(subsetStructuredDataset, mismatchedSubsetStructuredDataset)
		assert.False(t, castable, "StructuredDataset(a=Integer, b=Collection) should not be castable to StructuredDataset(a=Float)")
	})

	t.Run("MismatchedColumnsFlipped", func(t *testing.T) {
		castable := AreTypesCastable(mismatchedSubsetStructuredDataset, subsetStructuredDataset)
		assert.False(t, castable, "StructuredDataset(a=Float) should not be castable to StructuredDataset(a=Integer, b=Collection)")
	})

	t.Run("GenericToEmptyFormat", func(t *testing.T) {
		castable := AreTypesCastable(genericStructuredDataset, emptyStructuredDataset)
		assert.True(t, castable, "StructuredDataset(format='Parquet') should be castable to StructuredDataset()")
	})

	t.Run("EmptyFormatToGeneric", func(t *testing.T) {
		castable := AreTypesCastable(genericStructuredDataset, emptyStructuredDataset)
		assert.True(t, castable, "StructuredDataset() should be castable to StructuredDataset(format='Parquet')")
	})

	t.Run("StructuredDatasetsAreNullable", func(t *testing.T) {
		castable := AreTypesCastable(
			&core.LiteralType{
				Type: &core.LiteralType_Simple{
					Simple: core.SimpleType_NONE,
				},
			},
			subsetStructuredDataset)
		assert.True(t, castable, "StructuredDataset are nullable")
	})
}
