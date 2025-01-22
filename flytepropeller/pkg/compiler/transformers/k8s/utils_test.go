package k8s

import (
	"testing"

	"github.com/go-test/deep"
	_struct "github.com/golang/protobuf/ptypes/struct"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func TestComputeRetryStrategy(t *testing.T) {

	tests := []struct {
		name            string
		nodeRetries     uint32
		taskRetries     uint32
		expectedRetries uint32
	}{
		{"node-only", 1, 0, 2},
		{"task-only", 0, 1, 2},
		{"node-task", 2, 3, 3},
		{"no-retries", 0, 0, 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			var node *core.Node
			if test.nodeRetries != 0 {
				node = &core.Node{
					Metadata: &core.NodeMetadata{
						Retries: &core.RetryStrategy{
							Retries: test.nodeRetries,
						},
					},
				}
			}

			var tmpl *core.TaskTemplate
			if test.taskRetries != 0 {
				tmpl = &core.TaskTemplate{
					Metadata: &core.TaskMetadata{
						Retries: &core.RetryStrategy{
							Retries: test.taskRetries,
						},
					},
				}
			}

			r := computeRetryStrategy(node, tmpl)
			if test.expectedRetries != 0 {
				assert.NotNil(t, r)
				assert.Equal(t, int(test.expectedRetries), *r.MinAttempts) // #nosec G115
			} else {
				assert.Nil(t, r)
			}
		})
	}

}

func TestStripTypeMetadata(t *testing.T) {

	tests := []struct {
		name string
		args *core.LiteralType
		want *core.LiteralType
	}{
		{
			name: "nil",
			args: nil,
			want: nil,
		},
		{
			name: "simple",
			args: &core.LiteralType{
				Type: &core.LiteralType_Simple{
					Simple: core.SimpleType_INTEGER,
				},
				Metadata: &_struct.Struct{
					Fields: map[string]*_struct.Value{
						"foo": {
							Kind: &_struct.Value_StringValue{
								StringValue: "bar",
							},
						},
					},
				},
			},
			want: &core.LiteralType{
				Type: &core.LiteralType_Simple{
					Simple: core.SimpleType_INTEGER,
				},
			},
		},
		{
			name: "collection",
			args: &core.LiteralType{
				Type: &core.LiteralType_CollectionType{
					CollectionType: &core.LiteralType{
						Type: &core.LiteralType_Simple{
							Simple: core.SimpleType_INTEGER,
						},
					},
				},
				Metadata: &_struct.Struct{
					Fields: map[string]*_struct.Value{
						"foo": {
							Kind: &_struct.Value_StringValue{
								StringValue: "bar",
							},
						},
					},
				},
			},
			want: &core.LiteralType{
				Type: &core.LiteralType_CollectionType{
					CollectionType: &core.LiteralType{
						Type: &core.LiteralType_Simple{
							Simple: core.SimpleType_INTEGER,
						},
					},
				},
			},
		},
		{
			name: "map",
			args: &core.LiteralType{
				Type: &core.LiteralType_MapValueType{
					MapValueType: &core.LiteralType{
						Type: &core.LiteralType_Simple{
							Simple: core.SimpleType_INTEGER,
						},
					},
				},
				Metadata: &_struct.Struct{
					Fields: map[string]*_struct.Value{
						"foo": {
							Kind: &_struct.Value_StringValue{
								StringValue: "bar",
							},
						},
					},
				},
			},
			want: &core.LiteralType{
				Type: &core.LiteralType_MapValueType{
					MapValueType: &core.LiteralType{
						Type: &core.LiteralType_Simple{
							Simple: core.SimpleType_INTEGER,
						},
					},
				},
			},
		},
		{
			name: "union",
			args: &core.LiteralType{
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
								Metadata: &_struct.Struct{
									Fields: map[string]*_struct.Value{
										"foo": {
											Kind: &_struct.Value_StringValue{
												StringValue: "bar",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: &core.LiteralType{
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
		{
			name: "StructuredDataSet",
			args: &core.LiteralType{
				Type: &core.LiteralType_StructuredDatasetType{
					StructuredDatasetType: &core.StructuredDatasetType{
						Columns: []*core.StructuredDatasetType_DatasetColumn{
							{
								Name: "column1",
								LiteralType: &core.LiteralType{
									Type: &core.LiteralType_Simple{
										Simple: core.SimpleType_STRING,
									},
									Metadata: &_struct.Struct{
										Fields: map[string]*_struct.Value{
											"foo": {
												Kind: &_struct.Value_StringValue{
													StringValue: "bar",
												},
											},
										},
									},
									Structure: &core.TypeStructure{
										DataclassType: map[string]*core.LiteralType{
											"foo": {
												Type: &core.LiteralType_Simple{
													Simple: core.SimpleType_STRING,
												},
												Metadata: &_struct.Struct{
													Fields: map[string]*_struct.Value{
														"foo": {
															Kind: &_struct.Value_StringValue{
																StringValue: "bar",
															},
														},
													},
												},
											},
										},
										Tag: "str",
									},
								},
							},
						},
					},
				},
			},
			want: &core.LiteralType{
				Type: &core.LiteralType_StructuredDatasetType{
					StructuredDatasetType: &core.StructuredDatasetType{
						Columns: []*core.StructuredDatasetType_DatasetColumn{
							{
								Name: "column1",
								LiteralType: &core.LiteralType{
									Type: &core.LiteralType_Simple{
										Simple: core.SimpleType_STRING,
									},
									Structure: &core.TypeStructure{
										DataclassType: map[string]*core.LiteralType{
											"foo": {
												Type: &core.LiteralType_Simple{
													Simple: core.SimpleType_STRING,
												},
											},
										},
										Tag: "str",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "cache-key-metadata set",
			args: &core.LiteralType{
				Type: &core.LiteralType_Simple{
					Simple: core.SimpleType_INTEGER,
				},
				Annotation: &core.TypeAnnotation{
					Annotations: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"cache-key-metadata": {
								Kind: &_struct.Value_StructValue{
									StructValue: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"foo": {
												Kind: &_struct.Value_StringValue{
													StringValue: "bar",
												},
											},
										},
									},
								},
							},
						},
					},
				},
				Metadata: &_struct.Struct{
					Fields: map[string]*_struct.Value{
						"foo": {
							Kind: &_struct.Value_StringValue{
								StringValue: "bar",
							},
						},
					},
				},
			},
			want: &core.LiteralType{
				Type: &core.LiteralType_Simple{
					Simple: core.SimpleType_INTEGER,
				},
				Annotation: &core.TypeAnnotation{
					Annotations: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"cache-key-metadata": {
								Kind: &_struct.Value_StructValue{
									StructValue: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"foo": {
												Kind: &_struct.Value_StringValue{
													StringValue: "bar",
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
			name: "cache-key-metadata not present in annotation",
			args: &core.LiteralType{
				Type: &core.LiteralType_Simple{
					Simple: core.SimpleType_INTEGER,
				},
				Annotation: &core.TypeAnnotation{
					Annotations: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"some-key": {
								Kind: &_struct.Value_StringValue{
									StringValue: "some-value",
								},
							},
						},
					},
				},
				Metadata: &_struct.Struct{
					Fields: map[string]*_struct.Value{
						"foo": {
							Kind: &_struct.Value_StringValue{
								StringValue: "bar",
							},
						},
					},
				},
			},
			want: &core.LiteralType{
				Type: &core.LiteralType_Simple{
					Simple: core.SimpleType_INTEGER,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := deep.Equal(StripTypeMetadata(tt.args), tt.want); diff != nil {
				assert.Fail(t, "actual != expected", "Diff: %v", diff)
			}
		})
	}
}

func TestStripInterfaceTypeMetadata(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		assert.Nil(t, StripInterfaceTypeMetadata(nil))
	})

	t.Run("populated", func(t *testing.T) {
		vars := &core.VariableMap{
			Variables: map[string]*core.Variable{
				"a": {
					Type: &core.LiteralType{
						Type: &core.LiteralType_Simple{
							Simple: core.SimpleType_INTEGER,
						},
						Metadata: &_struct.Struct{
							Fields: map[string]*_struct.Value{
								"foo": {
									Kind: &_struct.Value_StringValue{
										StringValue: "bar",
									},
								},
							},
						},
					},
				},
			},
		}

		i := &core.TypedInterface{
			Inputs:  vars,
			Outputs: vars,
		}

		stripped := StripInterfaceTypeMetadata(i)
		assert.Nil(t, stripped.GetInputs().GetVariables()["a"].GetType().GetMetadata())
		assert.Nil(t, stripped.GetOutputs().GetVariables()["a"].GetType().GetMetadata())
	})
}
