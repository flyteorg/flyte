package gormimpl

import (
	"context"

	"github.com/jinzhu/gorm"
	"github.com/lyft/datacatalog/pkg/repositories/errors"
	"github.com/lyft/datacatalog/pkg/repositories/interfaces"
	"github.com/lyft/datacatalog/pkg/repositories/models"
	idl_datacatalog "github.com/lyft/datacatalog/protos/gen"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/promutils"
)

type dataSetRepo struct {
	db               *gorm.DB
	errorTransformer errors.ErrorTransformer
	repoMetrics      gormMetrics
}

func NewDatasetRepo(db *gorm.DB, errorTransformer errors.ErrorTransformer, scope promutils.Scope) interfaces.DatasetRepo {
	return &dataSetRepo{
		db:               db,
		errorTransformer: errorTransformer,
		repoMetrics:      newGormMetrics(scope),
	}
}

// Create a Dataset model
func (h *dataSetRepo) Create(ctx context.Context, in models.Dataset) error {
	timer := h.repoMetrics.CreateDuration.Start(ctx)
	defer timer.Stop()

	result := h.db.Create(&in)
	if result.Error != nil {
		return h.errorTransformer.ToDataCatalogError(result.Error)
	}
	return nil
}

// Get Dataset model
func (h *dataSetRepo) Get(ctx context.Context, in models.DatasetKey) (models.Dataset, error) {
	timer := h.repoMetrics.GetDuration.Start(ctx)
	defer timer.Stop()

	var ds models.Dataset
	result := h.db.Where(&models.Dataset{DatasetKey: in}).First(&ds)

	if result.Error != nil {
		logger.Debugf(ctx, "Unable to find Dataset: [%+v], err: %v", in, result.Error)
		return models.Dataset{}, h.errorTransformer.ToDataCatalogError(result.Error)
	}
	if result.RecordNotFound() {
		return models.Dataset{}, errors.GetMissingEntityError("Dataset", &idl_datacatalog.DatasetID{
			Project: in.Project,
			Domain:  in.Domain,
			Name:    in.Name,
			Version: in.Version,
		})
	}

	return ds, nil
}
