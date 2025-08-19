package datacatalog

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog"
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
	assert.NotEmpty(t, datasetID.GetVersion())
	assert.Equal(t, "1.0.0-GKw-c0Pw-GKw-c0Pw", datasetID.GetVersion())
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
	assert.NotEmpty(t, datasetID.GetVersion())
	assert.Equal(t, "1.0.0-GKw-c0Pw-GKw-c0Pw", datasetID.GetVersion())

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
	assert.NotEmpty(t, datasetID.GetVersion())
	assert.Equal(t, "1.0.0-UxVtPm0k-GKw-c0Pw", datasetID.GetVersion())

	key.TypedInterface.Inputs = &core.VariableMap{
		Variables: map[string]*core.Variable{
			"2": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
			"1": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
		},
	}
	datasetIDDupe, err := GenerateDatasetIDForTask(context.TODO(), key)
	assert.NoError(t, err)

	assert.Equal(t, "1.0.0-UxVtPm0k-GKw-c0Pw", datasetIDDupe.GetVersion())
	assert.Equal(t, datasetID.String(), datasetIDDupe.String())
}

// Ensure correct format of ArtifactTagName
func TestGenerateArtifactTagName(t *testing.T) {
	literalMap, err := coreutils.MakeLiteralMap(map[string]interface{}{"1": 1, "2": 2})
	assert.NoError(t, err)

	tag, err := GenerateArtifactTagName(context.TODO(), literalMap, nil)
	assert.NoError(t, err)
	assert.Equal(t, "flyte_cached-GQid5LjHbakcW68DS3P2jp80QLbiF0olFHF2hTh5bg8", tag)
}

func TestGenerateArtifactTagNameWithIgnore(t *testing.T) {
	literalMap, err := coreutils.MakeLiteralMap(map[string]interface{}{"1": 1, "2": 2, "3": 3})
	assert.NoError(t, err)
	cacheIgnoreInputVars := []string{"3"}
	tag, err := GenerateArtifactTagName(context.TODO(), literalMap, cacheIgnoreInputVars)
	assert.NoError(t, err)
	assert.Equal(t, "flyte_cached-GQid5LjHbakcW68DS3P2jp80QLbiF0olFHF2hTh5bg8", tag)
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
			execTaskAttemptKey: strconv.Itoa(int(tID.GetRetryAttempt())),
			execProjectKey:     tID.GetNodeExecutionId().GetExecutionId().GetProject(),
			execDomainKey:      tID.GetNodeExecutionId().GetExecutionId().GetDomain(),
			execNodeIDKey:      tID.GetNodeExecutionId().GetNodeId(),
			execNameKey:        tID.GetNodeExecutionId().GetExecutionId().GetName(),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetArtifactMetadataForSource(tt.args.taskExecutionID); !reflect.DeepEqual(got.GetKeyMap(), tt.want) {
				t.Errorf("GetMetadataForSource() = %v, want %v", got.GetKeyMap(), tt.want)
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
		{"legacy", args{datasetMd: GetDatasetMetadataForSource(&tID).GetKeyMap(), currentID: currentTaskID}, &core.TaskExecutionIdentifier{
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Name:         "x",
				Project:      "project",
				Domain:       "development",
				Version:      tID.GetTaskId().GetVersion(),
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
		{"latest", args{datasetMd: GetDatasetMetadataForSource(&tID).GetKeyMap(), artifactMd: GetArtifactMetadataForSource(&tID).GetKeyMap(), currentID: currentTaskID}, &tID},
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
	assert.Equal(t, core.ResourceType_DATASET, id.GetResourceType())
	assert.Equal(t, "n", id.GetName())
	assert.Equal(t, "p", id.GetProject())
	assert.Equal(t, "d", id.GetDomain())
	assert.Equal(t, "v", id.GetVersion())
}

func TestGenerateTaskOutputsFromArtifact_IDLNotFound(t *testing.T) {
	taskID := core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      "project",
		Domain:       "domain",
		Name:         "name",
		Version:      "version",
	}

	taskInterface := core.TypedInterface{
		Outputs: &core.VariableMap{
			Variables: map[string]*core.Variable{
				"output1": {
					Type: &core.LiteralType{
						Type: &core.LiteralType_Simple{
							Simple: 1000,
						},
					},
				},
			},
		},
	}

	artifact := &datacatalog.Artifact{
		Id: "artifact_id",
		Data: []*datacatalog.ArtifactData{
			{
				Name:  "output1",
				Value: &core.Literal{},
			},
		},
	}

	_, err := GenerateTaskOutputsFromArtifact(taskID, taskInterface, artifact)

	expectedContainedErrorMsg := "unexpected artifactData: [output1] val: [] does not match any task output type"
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedContainedErrorMsg)
}
