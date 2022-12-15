package datacatalog

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	"github.com/flyteorg/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/stretchr/testify/assert"
)

// add test for raarranged Literal maps for input values

func TestNilParamTask(t *testing.T) {
	key := catalog.Key{
		Identifier: core.Identifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
			Version: "1.0.0",
		},
		CacheVersion: "1.0.0",
		TypedInterface: core.TypedInterface{
			Inputs:  nil,
			Outputs: nil,
		},
	}
	datasetID, err := GenerateDatasetIDForTask(context.TODO(), key)
	assert.NoError(t, err)
	assert.NotEmpty(t, datasetID.Version)
	assert.Equal(t, "1.0.0-GKw-c0Pw-GKw-c0Pw", datasetID.Version)
}

// Ensure that empty parameters generate the same dataset as nil parameters
func TestEmptyParamTask(t *testing.T) {
	key := catalog.Key{
		Identifier: core.Identifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
			Version: "1.0.0",
		},
		CacheVersion: "1.0.0",
		TypedInterface: core.TypedInterface{
			Inputs:  &core.VariableMap{},
			Outputs: &core.VariableMap{},
		},
	}
	datasetID, err := GenerateDatasetIDForTask(context.TODO(), key)
	assert.NoError(t, err)
	assert.NotEmpty(t, datasetID.Version)
	assert.Equal(t, "1.0.0-GKw-c0Pw-GKw-c0Pw", datasetID.Version)

	key.TypedInterface.Inputs = nil
	key.TypedInterface.Outputs = nil
	datasetIDDupe, err := GenerateDatasetIDForTask(context.TODO(), key)
	assert.NoError(t, err)
	assert.Equal(t, datasetIDDupe.String(), datasetID.String())
}

// Ensure the key order on the map generates the same dataset
func TestVariableMapOrder(t *testing.T) {
	key := catalog.Key{
		Identifier: core.Identifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
			Version: "1.0.0",
		},
		CacheVersion: "1.0.0",
		TypedInterface: core.TypedInterface{
			Inputs: &core.VariableMap{
				Variables: map[string]*core.Variable{
					"1": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
					"2": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
				},
			},
		},
	}
	datasetID, err := GenerateDatasetIDForTask(context.TODO(), key)
	assert.NoError(t, err)
	assert.NotEmpty(t, datasetID.Version)
	assert.Equal(t, "1.0.0-UxVtPm0k-GKw-c0Pw", datasetID.Version)

	key.TypedInterface.Inputs = &core.VariableMap{
		Variables: map[string]*core.Variable{
			"2": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
			"1": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
		},
	}
	datasetIDDupe, err := GenerateDatasetIDForTask(context.TODO(), key)
	assert.NoError(t, err)

	assert.Equal(t, "1.0.0-UxVtPm0k-GKw-c0Pw", datasetIDDupe.Version)
	assert.Equal(t, datasetID.String(), datasetIDDupe.String())
}

// Ensure the key order on the inputs generates the same tag
func TestInputValueSorted(t *testing.T) {
	literalMap, err := coreutils.MakeLiteralMap(map[string]interface{}{"1": 1, "2": 2})
	assert.NoError(t, err)

	tag, err := GenerateArtifactTagName(context.TODO(), literalMap)
	assert.NoError(t, err)
	assert.Equal(t, "flyte_cached-GQid5LjHbakcW68DS3P2jp80QLbiF0olFHF2hTh5bg8", tag)

	literalMap, err = coreutils.MakeLiteralMap(map[string]interface{}{"2": 2, "1": 1})
	assert.NoError(t, err)

	tagDupe, err := GenerateArtifactTagName(context.TODO(), literalMap)
	assert.NoError(t, err)
	assert.Equal(t, tagDupe, tag)
}

// Ensure that empty inputs are hashed the same way
func TestNoInputValues(t *testing.T) {
	tag, err := GenerateArtifactTagName(context.TODO(), nil)
	assert.NoError(t, err)
	assert.Equal(t, "flyte_cached-GKw-c0PwFokMUQ6T-TUmEWnZ4_VlQ2Qpgw-vCTT0-OQ", tag)

	tagDupe, err := GenerateArtifactTagName(context.TODO(), &core.LiteralMap{Literals: nil})
	assert.NoError(t, err)
	assert.Equal(t, "flyte_cached-GKw-c0PwFokMUQ6T-TUmEWnZ4_VlQ2Qpgw-vCTT0-OQ", tagDupe)
	assert.Equal(t, tagDupe, tag)
}

func TestGetOrDefault(t *testing.T) {
	type args struct {
		m            map[string]string
		key          string
		defaultValue string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"default", args{m: map[string]string{"x": "val"}, key: "y", defaultValue: "def"}, "def"},
		{"original", args{m: map[string]string{"y": "val"}, key: "y", defaultValue: "def"}, "val"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetOrDefault(tt.args.m, tt.args.key, tt.args.defaultValue); got != tt.want {
				t.Errorf("GetOrDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetArtifactMetadataForSource(t *testing.T) {
	type args struct {
		taskExecutionID *core.TaskExecutionIdentifier
	}

	tID := &core.TaskExecutionIdentifier{
		TaskId: &core.Identifier{
			ResourceType: core.ResourceType_TASK,
			Name:         "x",
			Project:      "project",
			Domain:       "development",
			Version:      "ver",
		},
		NodeExecutionId: &core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Name:    "wf",
				Project: "p1",
				Domain:  "d1",
			},
			NodeId: "n",
		},
		RetryAttempt: 1,
	}

	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{"nil TaskExec", args{}, nil},
		{"TaskExec", args{tID}, map[string]string{
			execTaskAttemptKey: strconv.Itoa(int(tID.RetryAttempt)),
			execProjectKey:     tID.NodeExecutionId.ExecutionId.Project,
			execDomainKey:      tID.NodeExecutionId.ExecutionId.Domain,
			execNodeIDKey:      tID.NodeExecutionId.NodeId,
			execNameKey:        tID.NodeExecutionId.ExecutionId.Name,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetArtifactMetadataForSource(tt.args.taskExecutionID); !reflect.DeepEqual(got.KeyMap, tt.want) {
				t.Errorf("GetMetadataForSource() = %v, want %v", got.KeyMap, tt.want)
			}
		})
	}
}

func TestGetSourceFromMetadata(t *testing.T) {
	tID := core.TaskExecutionIdentifier{
		TaskId: &core.Identifier{
			ResourceType: core.ResourceType_TASK,
			Name:         "x",
			Project:      "project",
			Domain:       "development",
			Version:      "ver",
		},
		NodeExecutionId: &core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Name:    "wf",
				Project: "p1",
				Domain:  "d1",
			},
			NodeId: "n",
		},
		RetryAttempt: 1,
	}

	currentTaskID := core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Name:         "x",
		Project:      "project",
		Domain:       "development",
		Version:      "ver2",
	}

	type args struct {
		datasetMd  map[string]string
		artifactMd map[string]string
		currentID  core.Identifier
	}
	tests := []struct {
		name string
		args args
		want *core.TaskExecutionIdentifier
	}{
		// EVerything is missing
		{"missing", args{currentID: currentTaskID}, &core.TaskExecutionIdentifier{
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Name:         "x",
				Project:      "project",
				Domain:       "development",
				Version:      "unknown",
			},
			NodeExecutionId: &core.NodeExecutionIdentifier{
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Name:    "unknown",
					Project: "project",
					Domain:  "development",
				},
				NodeId: "unknown",
			},
			RetryAttempt: 0,
		}},
		// In legacy only taskVersionKey is available
		{"legacy", args{datasetMd: GetDatasetMetadataForSource(&tID).KeyMap, currentID: currentTaskID}, &core.TaskExecutionIdentifier{
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Name:         "x",
				Project:      "project",
				Domain:       "development",
				Version:      tID.TaskId.Version,
			},
			NodeExecutionId: &core.NodeExecutionIdentifier{
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Name:    "unknown",
					Project: "project",
					Domain:  "development",
				},
				NodeId: "unknown",
			},
			RetryAttempt: 0,
		}},
		// Completely available
		{"latest", args{datasetMd: GetDatasetMetadataForSource(&tID).KeyMap, artifactMd: GetArtifactMetadataForSource(&tID).KeyMap, currentID: currentTaskID}, &tID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := GetSourceFromMetadata(&datacatalog.Metadata{KeyMap: tt.args.datasetMd}, &datacatalog.Metadata{KeyMap: tt.args.artifactMd}, tt.args.currentID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSourceFromMetadata() = %v, want %v", got, tt.want)
				assert.NoError(t, err)
			}
		})
	}
}

func TestEventCatalogMetadata(t *testing.T) {
	tID := core.TaskExecutionIdentifier{
		TaskId: &core.Identifier{
			ResourceType: core.ResourceType_TASK,
			Name:         "x",
			Project:      "project",
			Domain:       "development",
			Version:      "ver",
		},
		NodeExecutionId: &core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Name:    "wf",
				Project: "p1",
				Domain:  "d1",
			},
			NodeId: "n",
		},
		RetryAttempt: 1,
	}
	datasetID := &datacatalog.DatasetID{Project: "p", Domain: "d", Name: "n", Version: "v"}
	type args struct {
		datasetID *datacatalog.DatasetID
		tag       *datacatalog.Tag
		sourceID  *core.TaskExecutionIdentifier
	}
	tests := []struct {
		name string
		args args
		want *core.CatalogMetadata
	}{
		{"only datasetID", args{datasetID: datasetID}, &core.CatalogMetadata{DatasetId: DatasetIDToIdentifier(datasetID)}},
		{"tag", args{datasetID: datasetID, tag: &datacatalog.Tag{Name: "n", ArtifactId: "a"}}, &core.CatalogMetadata{DatasetId: DatasetIDToIdentifier(datasetID), ArtifactTag: &core.CatalogArtifactTag{Name: "n", ArtifactId: "a"}}},
		{"source", args{datasetID: datasetID, tag: &datacatalog.Tag{Name: "n", ArtifactId: "a"}, sourceID: &tID}, &core.CatalogMetadata{DatasetId: DatasetIDToIdentifier(datasetID), ArtifactTag: &core.CatalogArtifactTag{Name: "n", ArtifactId: "a"}, SourceExecution: &core.CatalogMetadata_SourceTaskExecution{
			SourceTaskExecution: &tID,
		}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EventCatalogMetadata(tt.args.datasetID, tt.args.tag, tt.args.sourceID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EventCatalogMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDatasetIDToIdentifier(t *testing.T) {
	id := DatasetIDToIdentifier(&datacatalog.DatasetID{Project: "p", Domain: "d", Name: "n", Version: "v"})
	assert.Equal(t, core.ResourceType_DATASET, id.ResourceType)
	assert.Equal(t, "n", id.Name)
	assert.Equal(t, "p", id.Project)
	assert.Equal(t, "d", id.Domain)
	assert.Equal(t, "v", id.Version)
}

func TestGenerateArtifactTagName_LiteralsWithHashSet(t *testing.T) {
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
							&core.Literal{
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
							&core.Literal{
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
							&core.Literal{
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
							&core.Literal{
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
							&core.Literal{
								Value: nil,
								Hash:  "hash1",
							},
							&core.Literal{
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
							&core.Literal{
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
							&core.Literal{
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
							&core.Literal{
								Value: nil,
								Hash:  "hash1",
							},
							&core.Literal{
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
							"literal1": &core.Literal{
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
							"literal1": &core.Literal{
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
							"literal1": &core.Literal{
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
							"literal2-set-its-hash": &core.Literal{
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
							"literal3": &core.Literal{
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
							"literal1": &core.Literal{
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
							"literal2-set-its-hash": &core.Literal{
								Value: nil,
								Hash:  "literal-2-hash",
							},
							"literal3": &core.Literal{
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
							&core.Literal{
								Value: &core.Literal_Map{
									Map: &core.LiteralMap{
										Literals: map[string]*core.Literal{
											"literal1": &core.Literal{
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
											"literal2-set-its-hash": &core.Literal{
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
											"literal3": &core.Literal{
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
							&core.Literal{
								Value: &core.Literal_Map{
									Map: &core.LiteralMap{
										Literals: map[string]*core.Literal{
											"another-literal-1": &core.Literal{
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
											"another-literal2-set-its-hash": &core.Literal{
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
							&core.Literal{
								Value: &core.Literal_Map{
									Map: &core.LiteralMap{
										Literals: map[string]*core.Literal{
											"literal1": &core.Literal{
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
											"literal2-set-its-hash": &core.Literal{
												Value: nil,
												Hash:  "literal-2-hash",
											},
											"literal3": &core.Literal{
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
							&core.Literal{
								Value: &core.Literal_Map{
									Map: &core.LiteralMap{
										Literals: map[string]*core.Literal{
											"another-literal-1": &core.Literal{
												Value: nil,
												Hash:  "another-literal-1-hash",
											},
											"another-literal2-set-its-hash": &core.Literal{
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedLiteral, hashify(tt.literal))

			// Double-check that generating a tag is successful
			literalMap := &core.LiteralMap{Literals: map[string]*core.Literal{"o0": tt.literal}}
			tag, err := GenerateArtifactTagName(context.TODO(), literalMap)
			assert.NoError(t, err)
			assert.NotEmpty(t, tag)
		})
	}
}
