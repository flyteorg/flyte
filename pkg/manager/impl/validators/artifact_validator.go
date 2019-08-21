package validators

import (
	"fmt"

	datacatalog "github.com/lyft/datacatalog/protos/gen"
)

const (
	artifactID         = "artifactID"
	artifactDataEntity = "artifactData"
	artifactEntity     = "artifact"
)

func ValidateGetArtifactRequest(request datacatalog.GetArtifactRequest) error {
	if err := ValidateDatasetID(request.Dataset); err != nil {
		return err
	}

	if request.QueryHandle == nil {
		return NewMissingArgumentError(fmt.Sprintf("one of %s/%s", artifactID, tagName))
	}

	switch request.QueryHandle.(type) {
	case *datacatalog.GetArtifactRequest_ArtifactId:
		if err := ValidateEmptyStringField(request.GetArtifactId(), artifactID); err != nil {
			return err
		}
	case *datacatalog.GetArtifactRequest_TagName:
		if err := ValidateEmptyStringField(request.GetTagName(), tagName); err != nil {
			return err
		}
	default:
		return NewInvalidArgumentError("QueryHandle", "invalid type")
	}

	return nil
}

func ValidateEmptyArtifactData(artifactData []*datacatalog.ArtifactData) error {
	if len(artifactData) == 0 {
		return NewMissingArgumentError(artifactDataEntity)
	}

	return nil
}

func ValidateArtifact(artifact *datacatalog.Artifact) error {
	if artifact == nil {
		return NewMissingArgumentError(artifactEntity)
	}

	if err := ValidateDatasetID(artifact.Dataset); err != nil {
		return err
	}

	if err := ValidateEmptyStringField(artifact.Id, artifactID); err != nil {
		return err
	}

	if err := ValidateEmptyArtifactData(artifact.Data); err != nil {
		return err
	}

	return nil
}
