package transformers

import (
	"testing"

	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/stretchr/testify/assert"
)

func TestToTagKey(t *testing.T) {
	datasetID := &datacatalog.DatasetID{
		Project: "testProj",
		Domain:  "testDomain",
		Name:    "testName",
		Version: "testVersion",
		UUID:    "test-uuid",
	}

	tagName := "testTag"
	tagKey := ToTagKey(datasetID, tagName)

	assert.Equal(t, tagName, tagKey.TagName)
	assert.Equal(t, datasetID.Project, tagKey.DatasetProject)
	assert.Equal(t, datasetID.Domain, tagKey.DatasetDomain)
	assert.Equal(t, datasetID.Name, tagKey.DatasetName)
	assert.Equal(t, datasetID.Version, tagKey.DatasetVersion)
}

func TestFromTagModel(t *testing.T) {
	datasetID := &datacatalog.DatasetID{
		Project: "testProj",
		Domain:  "testDomain",
		Name:    "testName",
		Version: "testVersion",
		UUID:    "test-uuid",
	}

	tagModel := models.Tag{
		TagKey: models.TagKey{
			TagName: "test-tag",
		},
		DatasetUUID: "dataset-uuid",
	}

	tag := FromTagModel(datasetID, tagModel)

	assert.Equal(t, tag.Name, tagModel.TagName)
	assert.Equal(t, datasetID.Project, tag.Dataset.Project)
	assert.Equal(t, datasetID.Domain, tag.Dataset.Domain)
	assert.Equal(t, datasetID.Name, tag.Dataset.Name)
	assert.Equal(t, datasetID.Version, tag.Dataset.Version)
	assert.Equal(t, datasetID.UUID, tag.Dataset.UUID)
}
