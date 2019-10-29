package validators

import (
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
