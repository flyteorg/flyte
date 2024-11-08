package transformers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
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
	assert.Equal(t, datasetID.GetProject(), tagKey.DatasetProject)
	assert.Equal(t, datasetID.GetDomain(), tagKey.DatasetDomain)
	assert.Equal(t, datasetID.GetName(), tagKey.DatasetName)
	assert.Equal(t, datasetID.GetVersion(), tagKey.DatasetVersion)
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

	assert.Equal(t, tag.GetName(), tagModel.TagName)
	assert.Equal(t, datasetID.GetProject(), tag.GetDataset().GetProject())
	assert.Equal(t, datasetID.GetDomain(), tag.GetDataset().GetDomain())
	assert.Equal(t, datasetID.GetName(), tag.GetDataset().GetName())
	assert.Equal(t, datasetID.GetVersion(), tag.GetDataset().GetVersion())
	assert.Equal(t, datasetID.GetUUID(), tag.GetDataset().GetUUID())
}
