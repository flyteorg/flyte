package transformers

import (
	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

// Create a dataset model from the Dataset api object. This will serialize the metadata in the dataset as part of the transform
func CreateDatasetModel(dataset *datacatalog.Dataset) (*models.Dataset, error) {
	serializedMetadata, err := marshalMetadata(dataset.GetMetadata())
	if err != nil {
		return nil, err
	}

	partitionKeys := make([]models.PartitionKey, len(dataset.GetPartitionKeys()))

	for i, partitionKey := range dataset.GetPartitionKeys() {
		partitionKeys[i] = models.PartitionKey{
			Name: partitionKey,
		}
	}

	return &models.Dataset{
		DatasetKey: models.DatasetKey{
			Project: dataset.GetId().GetProject(),
			Domain:  dataset.GetId().GetDomain(),
			Name:    dataset.GetId().GetName(),
			Version: dataset.GetId().GetVersion(),
			UUID:    dataset.GetId().GetUUID(),
		},
		SerializedMetadata: serializedMetadata,
		PartitionKeys:      partitionKeys,
	}, nil
}

// Create a dataset ID from the dataset key model
func FromDatasetID(datasetID *datacatalog.DatasetID) models.DatasetKey {
	return models.DatasetKey{
		Project: datasetID.GetProject(),
		Domain:  datasetID.GetDomain(),
		Name:    datasetID.GetName(),
		Version: datasetID.GetVersion(),
		UUID:    datasetID.GetUUID(),
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
