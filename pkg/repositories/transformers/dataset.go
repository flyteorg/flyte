package transformers

import (
	"github.com/lyft/datacatalog/pkg/repositories/models"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
)

// Create a dataset model from the Dataset api object. This will serialize the metadata in the dataset as part of the transform
func CreateDatasetModel(dataset *datacatalog.Dataset) (*models.Dataset, error) {
	serializedMetadata, err := marshalMetadata(dataset.Metadata)
	if err != nil {
		return nil, err
	}

	return &models.Dataset{
		DatasetKey: models.DatasetKey{
			Project: dataset.Id.Project,
			Domain:  dataset.Id.Domain,
			Name:    dataset.Id.Name,
			Version: dataset.Id.Version,
		},
		SerializedMetadata: serializedMetadata,
	}, nil
}

// Create a dataset ID from the dataset key model
func FromDatasetID(datasetID datacatalog.DatasetID) models.DatasetKey {
	return models.DatasetKey{
		Project: datasetID.Project,
		Domain:  datasetID.Domain,
		Name:    datasetID.Name,
		Version: datasetID.Version,
	}
}

// Create a Dataset api object given a model, this will unmarshal the metadata into the object as part of the transform
func FromDatasetModel(dataset models.Dataset) (*datacatalog.Dataset, error) {
	metadata, err := unmarshalMetadata(dataset.SerializedMetadata)
	if err != nil {
		return nil, err
	}

	return &datacatalog.Dataset{
		Id: &datacatalog.DatasetID{
			Project: dataset.Project,
			Domain:  dataset.Domain,
			Name:    dataset.Name,
			Version: dataset.Version,
		},
		Metadata: metadata,
	}, nil
}
