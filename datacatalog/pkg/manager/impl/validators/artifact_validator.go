package validators

import (
	"fmt"

	"github.com/flyteorg/flyte/datacatalog/pkg/common"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
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

	switch request.GetQueryHandle().(type) {
	case *datacatalog.GetArtifactRequest_ArtifactId:
		if request.GetDataset() != nil {
			err := ValidateDatasetID(request.GetDataset())
			if err != nil {
				return err
			}
		}

		if err := ValidateEmptyStringField(request.GetArtifactId(), artifactID); err != nil {
			return err
		}
	case *datacatalog.GetArtifactRequest_TagName:
		if err := ValidateDatasetID(request.GetDataset()); err != nil {
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

	if err := ValidateDatasetID(artifact.GetDataset()); err != nil {
		return err
	}

	if err := ValidateEmptyStringField(artifact.GetId(), artifactID); err != nil {
		return err
	}

	if err := ValidateEmptyArtifactData(artifact.GetData()); err != nil {
		return err
	}

	return nil
}

// Validate the list request and format the request with proper defaults if not provided
func ValidateListArtifactRequest(request *datacatalog.ListArtifactsRequest) error {
	if err := ValidateDatasetID(request.GetDataset()); err != nil {
		return err
	}

	if err := ValidateArtifactFilterTypes(request.GetFilter().GetFilters()); err != nil {
		return err
	}

	if request.GetPagination() != nil {
		err := ValidatePagination(request.GetPagination())
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

	switch request.GetQueryHandle().(type) {
	case *datacatalog.UpdateArtifactRequest_ArtifactId:
		if request.GetDataset() != nil {
			err := ValidateDatasetID(request.GetDataset())
			if err != nil {
				return err
			}
		}

		if err := ValidateEmptyStringField(request.GetArtifactId(), artifactID); err != nil {
			return err
		}
	case *datacatalog.UpdateArtifactRequest_TagName:
		if err := ValidateDatasetID(request.GetDataset()); err != nil {
			return err
		}

		if err := ValidateEmptyStringField(request.GetTagName(), tagName); err != nil {
			return err
		}
	default:
		return NewInvalidArgumentError("QueryHandle", "invalid type")
	}

	if err := ValidateEmptyArtifactData(request.GetData()); err != nil {
		return err
	}

	return nil
}
