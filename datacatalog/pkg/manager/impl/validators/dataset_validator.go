package validators

import (
	"github.com/flyteorg/flyte/datacatalog/pkg/common"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

const (
	datasetEntity  = "dataset"
	datasetProject = "project"
	datasetDomain  = "domain"
	datasetName    = "name"
	datasetVersion = "version"
)

// Validate that the DatasetID has all the fields filled
func ValidateDatasetID(ds *datacatalog.DatasetID) error {
	if ds == nil {
		return NewMissingArgumentError(datasetEntity)
	}
	if err := ValidateEmptyStringField(ds.GetProject(), datasetProject); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(ds.GetDomain(), datasetDomain); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(ds.GetName(), datasetName); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(ds.GetVersion(), datasetVersion); err != nil {
		return err
	}
	return nil
}

// Ensure list Datasets request is properly constructed
func ValidateListDatasetsRequest(request *datacatalog.ListDatasetsRequest) error {
	if request.GetPagination() != nil {
		err := ValidatePagination(request.GetPagination())
		if err != nil {
			return err
		}
	}

	// Datasets cannot be filtered by tag, partitions or artifacts
	for _, filter := range request.GetFilter().GetFilters() {
		if filter.GetTagFilter() != nil {
			return NewInvalidFilterError(common.Dataset, common.Tag)
		} else if filter.GetPartitionFilter() != nil {
			return NewInvalidFilterError(common.Dataset, common.Partition)
		} else if filter.GetArtifactFilter() != nil {
			return NewInvalidFilterError(common.Dataset, common.Artifact)
		}
	}
	return nil
}
