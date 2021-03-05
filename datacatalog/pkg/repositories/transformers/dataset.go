package transformers

import (
	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	datacatalog "github.com/flyteorg/datacatalog/protos/gen"
)

// Create a dataset model from the Dataset api object. This will serialize the metadata in the dataset as part of the transform
func CreateDatasetModel(dataset *datacatalog.Dataset) (*models.Dataset, error) {
	serializedMetadata, err := marshalMetadata(dataset.Metadata)
	if err != nil {
		return nil, err
	}

	partitionKeys := make([]models.PartitionKey, len(dataset.PartitionKeys))

	for i, partitionKey := range dataset.GetPartitionKeys() {
		partitionKeys[i] = models.PartitionKey{
			Name: partitionKey,
		}
	}

	return &models.Dataset{
		DatasetKey: models.DatasetKey{
			Project: dataset.Id.Project,
			Domain:  dataset.Id.Domain,
			Name:    dataset.Id.Name,
			Version: dataset.Id.Version,
			UUID:    dataset.Id.UUID,
		},
		SerializedMetadata: serializedMetadata,
		PartitionKeys:      partitionKeys,
	}, nil
}

// Create a dataset ID from the dataset key model
func FromDatasetID(datasetID *datacatalog.DatasetID) models.DatasetKey {
	return models.DatasetKey{
		Project: datasetID.Project,
		Domain:  datasetID.Domain,
		Name:    datasetID.Name,
		Version: datasetID.Version,
		UUID:    datasetID.UUID,
	}
}

// Create a Dataset api object given a model, this will unmarshal the metadata into the object as part of the transform
func FromDatasetModel(dataset models.Dataset) (*datacatalog.Dataset, error) {
	metadata, err := unmarshalMetadata(dataset.SerializedMetadata)
	if err != nil {
		return nil, err
	}

	partitionKeyStrings := FromPartitionKeyModel(dataset.PartitionKeys)
	return &datacatalog.Dataset{
		Id: &datacatalog.DatasetID{
			UUID:    dataset.UUID,
			Project: dataset.Project,
			Domain:  dataset.Domain,
			Name:    dataset.Name,
			Version: dataset.Version,
		},
		Metadata:      metadata,
		PartitionKeys: partitionKeyStrings,
	}, nil
}
