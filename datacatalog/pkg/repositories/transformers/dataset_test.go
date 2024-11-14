package transformers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

var metadata = &datacatalog.Metadata{
	KeyMap: map[string]string{
		"testKey1": "testValue1",
		"testKey2": "testValue2",
	},
}

var datasetID = &datacatalog.DatasetID{
	Project: "test-project",
	Domain:  "test-domain",
	Name:    "test-name",
	Version: "test-version",
	UUID:    "test-uuid",
}

func assertDatasetIDEqualsModel(t *testing.T, idlDataset *datacatalog.DatasetID, model *models.DatasetKey) {
	assert.Equal(t, idlDataset.GetProject(), model.Project)
	assert.Equal(t, idlDataset.GetDomain(), model.Domain)
	assert.Equal(t, idlDataset.GetName(), model.Name)
	assert.Equal(t, idlDataset.GetVersion(), model.Version)
	assert.Equal(t, idlDataset.GetUUID(), model.UUID)
}

func TestCreateDatasetModelNoParitions(t *testing.T) {
	dataset := &datacatalog.Dataset{
		Id:       datasetID,
		Metadata: metadata,
	}

	datasetModel, err := CreateDatasetModel(dataset)
	assert.NoError(t, err)
	assertDatasetIDEqualsModel(t, dataset.GetId(), &datasetModel.DatasetKey)

	unmarshaledMetadata, err := unmarshalMetadata(datasetModel.SerializedMetadata)
	assert.NoError(t, err)
	assert.EqualValues(t, unmarshaledMetadata.GetKeyMap(), metadata.GetKeyMap())

	assert.Len(t, datasetModel.PartitionKeys, 0)
}

func TestCreateDatasetModel(t *testing.T) {
	dataset := &datacatalog.Dataset{
		Id:            datasetID,
		Metadata:      metadata,
		PartitionKeys: []string{"key1", "key2"},
	}

	datasetModel, err := CreateDatasetModel(dataset)
	assert.NoError(t, err)
	assertDatasetIDEqualsModel(t, dataset.GetId(), &datasetModel.DatasetKey)

	unmarshaledMetadata, err := unmarshalMetadata(datasetModel.SerializedMetadata)
	assert.NoError(t, err)
	assert.EqualValues(t, unmarshaledMetadata.GetKeyMap(), metadata.GetKeyMap())

	assert.Len(t, datasetModel.PartitionKeys, 2)
	assert.Equal(t, datasetModel.PartitionKeys[0], models.PartitionKey{Name: dataset.GetPartitionKeys()[0]})
	assert.Equal(t, datasetModel.PartitionKeys[1], models.PartitionKey{Name: dataset.GetPartitionKeys()[1]})
}

func TestFromDatasetID(t *testing.T) {
	datasetKey := FromDatasetID(datasetID)
	assertDatasetIDEqualsModel(t, datasetID, &datasetKey)
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
	assertDatasetIDEqualsModel(t, dataset.GetId(), &datasetModel.DatasetKey)
	assert.Len(t, dataset.GetMetadata().GetKeyMap(), 0)
	assert.Len(t, dataset.GetPartitionKeys(), 0)
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
	assertDatasetIDEqualsModel(t, dataset.GetId(), &datasetModel.DatasetKey)
	assert.Len(t, dataset.GetMetadata().GetKeyMap(), 2)
	assert.EqualValues(t, dataset.GetMetadata().GetKeyMap(), metadata.GetKeyMap())
	assert.Len(t, dataset.GetPartitionKeys(), 2)
}
