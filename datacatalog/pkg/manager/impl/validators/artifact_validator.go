package validators

import (
	"fmt"

	"github.com/flyteorg/datacatalog/pkg/common"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

const (
	artifactID         = "artifactID"
	artifactDataEntity = "artifactData"
	artifactEntity     = "artifact"
)

func ValidateGetArtifactRequest(request *datacatalog.GetArtifactRequest) error {
	if request.QueryHandle == nil {
		return NewMissingArgumentError(fmt.Sprintf("one of %s/%s", artifactID, tagName))
	}

	switch request.QueryHandle.(type) {
	case *datacatalog.GetArtifactRequest_ArtifactId:
		if request.Dataset != nil {
			err := ValidateDatasetID(request.Dataset)
			if err != nil {
				return err
			}
		}

		if err := ValidateEmptyStringField(request.GetArtifactId(), artifactID); err != nil {
			return err
		}
	case *datacatalog.GetArtifactRequest_TagName:
		if err := ValidateDatasetID(request.Dataset); err != nil {
			return err
		}

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

// Validate the list request and format the request with proper defaults if not provided
func ValidateListArtifactRequest(request *datacatalog.ListArtifactsRequest) error {
	if err := ValidateDatasetID(request.Dataset); err != nil {
		return err
	}

	if err := ValidateArtifactFilterTypes(request.Filter.GetFilters()); err != nil {
		return err
	}

	if request.Pagination != nil {
		err := ValidatePagination(request.Pagination)
		if err != nil {
			return err
		}
	}

	return nil
}

// Artifacts cannot be filtered across Datasets
func ValidateArtifactFilterTypes(filters []*datacatalog.SinglePropertyFilter) error {
	for _, filter := range filters {
		if filter.GetDatasetFilter() != nil {
			return NewInvalidFilterError(common.Artifact, common.Dataset)
		}
	}
	return nil
}

func ValidateUpdateArtifactRequest(request *datacatalog.UpdateArtifactRequest) error {
	if request.QueryHandle == nil {
		return NewMissingArgumentError(fmt.Sprintf("one of %s/%s", artifactID, tagName))
	}

	switch request.QueryHandle.(type) {
	case *datacatalog.UpdateArtifactRequest_ArtifactId:
		if request.Dataset != nil {
			err := ValidateDatasetID(request.Dataset)
			if err != nil {
				return err
			}
		}

		if err := ValidateEmptyStringField(request.GetArtifactId(), artifactID); err != nil {
			return err
		}
	case *datacatalog.UpdateArtifactRequest_TagName:
		if err := ValidateDatasetID(request.Dataset); err != nil {
			return err
		}

		if err := ValidateEmptyStringField(request.GetTagName(), tagName); err != nil {
			return err
		}
	default:
		return NewInvalidArgumentError("QueryHandle", "invalid type")
	}

	if err := ValidateEmptyArtifactData(request.Data); err != nil {
		return err
	}

	return nil
}

func ValidateDeleteArtifactRequest(request *datacatalog.DeleteArtifactRequest) error {
	if request.QueryHandle == nil {
		return NewMissingArgumentError(fmt.Sprintf("one of %s/%s", artifactID, tagName))
	}

	switch request.QueryHandle.(type) {
	case *datacatalog.DeleteArtifactRequest_ArtifactId:
		if request.Dataset != nil {
			err := ValidateDatasetID(request.Dataset)
			if err != nil {
				return err
			}
		}

		if err := ValidateEmptyStringField(request.GetArtifactId(), artifactID); err != nil {
			return err
		}
	case *datacatalog.DeleteArtifactRequest_TagName:
		if err := ValidateDatasetID(request.Dataset); err != nil {
			return err
		}

		if err := ValidateEmptyStringField(request.GetTagName(), tagName); err != nil {
			return err
		}
	default:
		return NewInvalidArgumentError("QueryHandle", "invalid type")
	}

	return nil
}
