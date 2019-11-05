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
	UUID:    "test-uuid",
}

func assertDatasetIDEqualsModel(t *testing.T, idlDataset *datacatalog.DatasetID, model *models.DatasetKey) {
	assert.Equal(t, idlDataset.Project, model.Project)
	assert.Equal(t, idlDataset.Domain, model.Domain)
	assert.Equal(t, idlDataset.Name, model.Name)
	assert.Equal(t, idlDataset.Version, model.Version)
	assert.Equal(t, idlDataset.UUID, model.UUID)
}

func TestCreateDatasetModelNoParitions(t *testing.T) {
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

	assert.Len(t, datasetModel.PartitionKeys, 0)
}

func TestCreateDatasetModel(t *testing.T) {
	dataset := &datacatalog.Dataset{
		Id:            &datasetID,
		Metadata:      &metadata,
		PartitionKeys: []string{"key1", "key2"},
	}

	datasetModel, err := CreateDatasetModel(dataset)
	assert.NoError(t, err)
	assertDatasetIDEqualsModel(t, dataset.Id, &datasetModel.DatasetKey)

	unmarshaledMetadata, err := unmarshalMetadata(datasetModel.SerializedMetadata)
	assert.NoError(t, err)
	assert.EqualValues(t, unmarshaledMetadata.KeyMap, metadata.KeyMap)

	assert.Len(t, datasetModel.PartitionKeys, 2)
	assert.Equal(t, datasetModel.PartitionKeys[0], models.PartitionKey{Name: dataset.PartitionKeys[0]})
	assert.Equal(t, datasetModel.PartitionKeys[1], models.PartitionKey{Name: dataset.PartitionKeys[1]})
}

func TestFromDatasetID(t *testing.T) {
	datasetKey := FromDatasetID(datasetID)
	assertDatasetIDEqualsModel(t, &datasetID, &datasetKey)
}

func TestFromDatasetModelNoPartitionsOrMetadata(t *testing.T) {
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
	assert.Len(t, dataset.PartitionKeys, 0)
}

func TestFromDatasetModelWithPartitions(t *testing.T) {
	datasetModel := &models.Dataset{
		DatasetKey: models.DatasetKey{
			Project: "test-project",
			Domain:  "test-domain",
			Name:    "test-name",
			Version: "test-version",
			UUID:    "test-uuid",
		},
		SerializedMetadata: []byte{10, 22, 10, 8, 116, 101, 115, 116, 75, 101, 121, 49, 18, 10, 116, 101, 115, 116, 86, 97, 108, 117, 101, 49, 10, 22, 10, 8, 116, 101, 115, 116, 75, 101, 121, 50, 18, 10, 116, 101, 115, 116, 86, 97, 108, 117, 101, 50},
		PartitionKeys: []models.PartitionKey{
			{DatasetUUID: "test-uuid", Name: "key1"},
			{DatasetUUID: "test-uuid", Name: "key2"},
		},
	}
	dataset, err := FromDatasetModel(*datasetModel)
	assert.NoError(t, err)
	assertDatasetIDEqualsModel(t, dataset.Id, &datasetModel.DatasetKey)
	assert.Len(t, dataset.Metadata.KeyMap, 2)
	assert.EqualValues(t, dataset.Metadata.KeyMap, metadata.KeyMap)
	assert.Len(t, dataset.PartitionKeys, 2)
}
