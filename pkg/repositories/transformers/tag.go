package transformers

import (
	"github.com/lyft/datacatalog/pkg/repositories/models"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
)

func ToTagKey(datasetID datacatalog.DatasetID, tagName string) models.TagKey {
	return models.TagKey{
		DatasetProject: datasetID.Project,
		DatasetDomain:  datasetID.Domain,
		DatasetName:    datasetID.Name,
		DatasetVersion: datasetID.Version,
		TagName:        tagName,
	}
}

func FromTagModel(datasetID datacatalog.DatasetID, tag models.Tag) *datacatalog.Tag {
	return &datacatalog.Tag{
		Name:       tag.TagName,
		ArtifactId: tag.ArtifactID,
		Dataset:    &datasetID,
	}
}
