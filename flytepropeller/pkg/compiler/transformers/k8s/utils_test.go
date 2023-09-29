package k8s

import (
	"testing"

	"github.com/go-test/deep"
	_struct "github.com/golang/protobuf/ptypes/struct"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

func TestComputeRetryStrategy(t *testing.T) {

	tests := []struct {
		name            string
		nodeRetries     int
		taskRetries     int
		expectedRetries int
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
							Retries: uint32(test.nodeRetries),
						},
					},
				}
			}

			var tmpl *core.TaskTemplate
			if test.taskRetries != 0 {
				tmpl = &core.TaskTemplate{
					Metadata: &core.TaskMetadata{
						Retries: &core.RetryStrategy{
							Retries: uint32(test.taskRetries),
						},
					},
				}
			}

			r := computeRetryStrategy(node, tmpl)
			if test.expectedRetries != 0 {
				assert.NotNil(t, r)
				assert.Equal(t, test.expectedRetries, *r.MinAttempts)
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
									Structure: &core.TypeStructure{Tag: "str"},
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
									Structure: &core.TypeStructure{Tag: "str"},
								},
							},
						},
					},
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
		assert.Nil(t, stripped.Inputs.Variables["a"].Type.Metadata)
		assert.Nil(t, stripped.Outputs.Variables["a"].Type.Metadata)
	})
}
