package validators

import (
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

const (
	tagName   = "tagName"
	tagEntity = "tag"
)

func ValidateTag(tag *datacatalog.Tag) error {
	if tag == nil {
		return NewMissingArgumentError(tagEntity)
	}
	if err := ValidateDatasetID(tag.GetDataset()); err != nil {
		return err
	}

	if err := ValidateEmptyStringField(tag.GetName(), tagName); err != nil {
		return err
	}

	if err := ValidateEmptyStringField(tag.GetArtifactId(), artifactID); err != nil {
		return err
	}
	return nil
}
