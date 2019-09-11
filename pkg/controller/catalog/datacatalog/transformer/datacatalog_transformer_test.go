package transformer

import (
	"context"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytepropeller/pkg/utils"
	"github.com/stretchr/testify/assert"
)

// add test for raarranged Literal maps for input values

func TestNilParamTask(t *testing.T) {
	task := &core.TaskTemplate{
		Id: &core.Identifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
			Version: "1.0.0",
		},
		Metadata: &core.TaskMetadata{
			DiscoveryVersion: "1.0.0",
		},
		Interface: &core.TypedInterface{
			Inputs:  nil,
			Outputs: nil,
		},
	}
	datasetID, err := GenerateDatasetIDForTask(context.TODO(), task)
	assert.NoError(t, err)
	assert.NotEmpty(t, datasetID.Version)
	assert.Equal(t, "1.0.0-V-K42BDF-V-K42BDF", datasetID.Version)
}

// Ensure that empty parameters generate the same dataset as nil parameters
func TestEmptyParamTask(t *testing.T) {
	task := &core.TaskTemplate{
		Id: &core.Identifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
			Version: "1.0.0",
		},
		Metadata: &core.TaskMetadata{
			DiscoveryVersion: "1.0.0",
		},
		Interface: &core.TypedInterface{
			Inputs:  &core.VariableMap{},
			Outputs: &core.VariableMap{},
		},
	}
	datasetID, err := GenerateDatasetIDForTask(context.TODO(), task)
	assert.NoError(t, err)
	assert.NotEmpty(t, datasetID.Version)
	assert.Equal(t, "1.0.0-V-K42BDF-V-K42BDF", datasetID.Version)

	task.Interface.Inputs = nil
	task.Interface.Outputs = nil
	datasetIDDupe, err := GenerateDatasetIDForTask(context.TODO(), task)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(datasetIDDupe, datasetID))
}

// Ensure the key order on the map generates the same dataset
func TestVariableMapOrder(t *testing.T) {
	task := &core.TaskTemplate{
		Id: &core.Identifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
			Version: "1.0.0",
		},
		Metadata: &core.TaskMetadata{
			DiscoveryVersion: "1.0.0",
		},
		Interface: &core.TypedInterface{
			Inputs: &core.VariableMap{
				Variables: map[string]*core.Variable{
					"1": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
					"2": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
				},
			},
		},
	}
	datasetID, err := GenerateDatasetIDForTask(context.TODO(), task)
	assert.NoError(t, err)
	assert.NotEmpty(t, datasetID.Version)
	assert.Equal(t, "1.0.0-UxVtPm0k-V-K42BDF", datasetID.Version)

	task.Interface.Inputs = &core.VariableMap{
		Variables: map[string]*core.Variable{
			"2": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
			"1": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
		},
	}
	datasetIDDupe, err := GenerateDatasetIDForTask(context.TODO(), task)
	assert.NoError(t, err)

	assert.Equal(t, "1.0.0-UxVtPm0k-V-K42BDF", datasetIDDupe.Version)
	assert.True(t, proto.Equal(datasetID, datasetIDDupe))
}

// Ensure the key order on the inputs generates the same tag
func TestInputValueSorted(t *testing.T) {
	literalMap, err := utils.MakeLiteralMap(map[string]interface{}{"1": 1, "2": 2})
	assert.NoError(t, err)

	tag, err := GenerateArtifactTagName(context.TODO(), literalMap)
	assert.NoError(t, err)
	assert.Equal(t, "flyte_cached-GQid5LjHbakcW68DS3P2jp80QLbiF0olFHF2hTh5bg8", tag)

	literalMap, err = utils.MakeLiteralMap(map[string]interface{}{"2": 2, "1": 1})
	assert.NoError(t, err)

	tagDupe, err := GenerateArtifactTagName(context.TODO(), literalMap)
	assert.NoError(t, err)
	assert.Equal(t, tagDupe, tag)
}

// Ensure that empty inputs are hashed the same way
func TestNoInputValues(t *testing.T) {
	tag, err := GenerateArtifactTagName(context.TODO(), nil)
	assert.NoError(t, err)
	assert.Equal(t, "flyte_cached-m4vFNUOHOFEFIiZSyOyid92TkWFFBDha4UOkkBb47XU", tag)

	tagDupe, err := GenerateArtifactTagName(context.TODO(), &core.LiteralMap{Literals: nil})
	assert.NoError(t, err)
	assert.Equal(t, "flyte_cached-m4vFNUOHOFEFIiZSyOyid92TkWFFBDha4UOkkBb47XU", tagDupe)
	assert.Equal(t, tagDupe, tag)
}
