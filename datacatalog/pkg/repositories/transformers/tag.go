package transformers

import (
	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

func ToTagKey(datasetID *datacatalog.DatasetID, tagName string) models.TagKey {
	return models.TagKey{
		DatasetProject: datasetID.GetProject(),
		DatasetDomain:  datasetID.GetDomain(),
		DatasetName:    datasetID.GetName(),
		DatasetVersion: datasetID.GetVersion(),
		TagName:        tagName,
	}
}

func FromTagModel(datasetID *datacatalog.DatasetID, tag models.Tag) *datacatalog.Tag {
	return &datacatalog.Tag{
		Name:       tag.TagName,
		ArtifactId: tag.ArtifactID,
		Dataset:    datasetID,
	}
}
