package transformers

import (
	"github.com/lyft/datacatalog/pkg/repositories/models"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
)

func CreateArtifactModel(request datacatalog.CreateArtifactRequest, artifactData []models.ArtifactData, dataset models.Dataset) (models.Artifact, error) {
	datasetID := request.Artifact.Dataset

	serializedMetadata, err := marshalMetadata(request.Artifact.Metadata)
	if err != nil {
		return models.Artifact{}, err
	}

	partitions := make([]models.Partition, len(request.Artifact.Partitions))
	for i, partition := range request.Artifact.GetPartitions() {
		partitions[i] = models.Partition{
			DatasetUUID: dataset.UUID,
			Key:         partition.Key,
			Value:       partition.Value,
		}
	}

	return models.Artifact{
		ArtifactKey: models.ArtifactKey{
			DatasetProject: datasetID.Project,
			DatasetDomain:  datasetID.Domain,
			DatasetName:    datasetID.Name,
			DatasetVersion: datasetID.Version,
			ArtifactID:     request.Artifact.Id,
		},
		DatasetUUID:        dataset.UUID,
		ArtifactData:       artifactData,
		SerializedMetadata: serializedMetadata,
		Partitions:         partitions,
	}, nil
}

func FromArtifactModel(artifact models.Artifact) (datacatalog.Artifact, error) {
	datasetID := datacatalog.DatasetID{
		Project: artifact.DatasetProject,
		Domain:  artifact.DatasetDomain,
		Name:    artifact.DatasetName,
		Version: artifact.DatasetVersion,
		UUID:    artifact.DatasetUUID,
	}

	metadata, err := unmarshalMetadata(artifact.SerializedMetadata)
	if err != nil {
		return datacatalog.Artifact{}, err
	}

	partitions := make([]*datacatalog.Partition, len(artifact.Partitions))
	for i, partition := range artifact.Partitions {
		partitions[i] = &datacatalog.Partition{
			Key:   partition.Key,
			Value: partition.Value,
		}
	}

	tags := make([]*datacatalog.Tag, len(artifact.Tags))
	for i, tag := range artifact.Tags {
		tags[i] = FromTagModel(datasetID, tag)
	}
	return datacatalog.Artifact{
		Id:         artifact.ArtifactID,
		Dataset:    &datasetID,
		Metadata:   metadata,
		Partitions: partitions,
		Tags:       tags,
	}, nil
}

func FromArtifactModels(artifacts []models.Artifact) ([]*datacatalog.Artifact, error) {
	retArtifacts := make([]*datacatalog.Artifact, 0, len(artifacts))
	for _, artifact := range artifacts {
		retArtifact, err := FromArtifactModel(artifact)
		if err != nil {
			return nil, err
		}

		retArtifacts = append(retArtifacts, &retArtifact)
	}

	return retArtifacts, nil
}

func ToArtifactKey(datasetID datacatalog.DatasetID, artifactID string) models.ArtifactKey {
	return models.ArtifactKey{
		DatasetProject: datasetID.Project,
		DatasetDomain:  datasetID.Domain,
		DatasetName:    datasetID.Name,
		DatasetVersion: datasetID.Version,
		ArtifactID:     artifactID,
	}
}
