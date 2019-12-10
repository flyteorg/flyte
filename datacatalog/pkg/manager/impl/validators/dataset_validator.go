package validators

import (
	"github.com/lyft/datacatalog/pkg/common"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
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
	if err := ValidateEmptyStringField(ds.Project, datasetProject); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(ds.Domain, datasetDomain); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(ds.Name, datasetName); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(ds.Version, datasetVersion); err != nil {
		return err
	}
	return nil
}

// Ensure list Datasets request is properly constructed
func ValidateListDatasetsRequest(request *datacatalog.ListDatasetsRequest) error {
	if request.Pagination != nil {
		err := ValidatePagination(*request.Pagination)
		if err != nil {
			return err
		}
	}

	// Datasets cannot be filtered by tag, partitions or artifacts
	for _, filter := range request.Filter.GetFilters() {
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
