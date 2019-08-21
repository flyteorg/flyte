package transformers

import (
	"github.com/lyft/datacatalog/pkg/repositories/models"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
)

func CreateArtifactModel(request datacatalog.CreateArtifactRequest, artifactData []models.ArtifactData) (models.Artifact, error) {
	datasetID := request.Artifact.Dataset

	serializedMetadata, err := marshalMetadata(request.Artifact.Metadata)
	if err != nil {
		return models.Artifact{}, err
	}

	return models.Artifact{
		ArtifactKey: models.ArtifactKey{
			DatasetProject: datasetID.Project,
			DatasetDomain:  datasetID.Domain,
			DatasetName:    datasetID.Name,
			DatasetVersion: datasetID.Version,
			ArtifactID:     request.Artifact.Id,
		},
		ArtifactData:       artifactData,
		SerializedMetadata: serializedMetadata,
	}, nil
}

func FromArtifactModel(artifact models.Artifact) (datacatalog.Artifact, error) {
	metadata, err := unmarshalMetadata(artifact.SerializedMetadata)
	if err != nil {
		return datacatalog.Artifact{}, err
	}

	return datacatalog.Artifact{
		Id: artifact.ArtifactID,
		Dataset: &datacatalog.DatasetID{
			Project: artifact.DatasetProject,
			Domain:  artifact.DatasetDomain,
			Name:    artifact.DatasetName,
			Version: artifact.DatasetVersion,
		},
		Metadata: metadata,
	}, nil
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
