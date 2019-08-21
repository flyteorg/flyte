package transformers

import (
	"testing"

	"github.com/lyft/datacatalog/pkg/repositories/models"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
	"github.com/stretchr/testify/assert"
)

var metadata = datacatalog.Metadata{
	KeyMap: map[string]string{
		"testKey1": "testValue1",
		"testKey2": "testValue2",
	},
}

var datasetID = datacatalog.DatasetID{
	Project: "test-project",
	Domain:  "test-domain",
	Name:    "test-name",
	Version: "test-version",
}

func assertDatasetIDEqualsModel(t *testing.T, idlDataset *datacatalog.DatasetID, model *models.DatasetKey) {
	assert.Equal(t, idlDataset.Project, model.Project)
	assert.Equal(t, idlDataset.Domain, model.Domain)
	assert.Equal(t, idlDataset.Name, model.Name)
	assert.Equal(t, idlDataset.Version, model.Version)
}

func TestCreateDatasetModel(t *testing.T) {
	dataset := &datacatalog.Dataset{
		Id:       &datasetID,
		Metadata: &metadata,
	}

	datasetModel, err := CreateDatasetModel(dataset)
	assert.NoError(t, err)
	assertDatasetIDEqualsModel(t, dataset.Id, &datasetModel.DatasetKey)

	unmarshaledMetadata, err := unmarshalMetadata(datasetModel.SerializedMetadata)
	assert.NoError(t, err)
	assert.EqualValues(t, unmarshaledMetadata.KeyMap, metadata.KeyMap)
}

func TestFromDatasetID(t *testing.T) {
	datasetKey := FromDatasetID(datasetID)
	assertDatasetIDEqualsModel(t, &datasetID, &datasetKey)
}

func TestFromDatasetModel(t *testing.T) {
	datasetModel := &models.Dataset{
		DatasetKey: models.DatasetKey{
			Project: "test-project",
			Domain:  "test-domain",
			Name:    "test-name",
			Version: "test-version",
		},
		SerializedMetadata: []byte{},
	}
	dataset, err := FromDatasetModel(*datasetModel)
	assert.NoError(t, err)
	assertDatasetIDEqualsModel(t, dataset.Id, &datasetModel.DatasetKey)
	assert.Len(t, dataset.Metadata.KeyMap, 0)
}
