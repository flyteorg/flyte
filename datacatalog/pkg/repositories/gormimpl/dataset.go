package gormimpl

import (
	"context"

	"gorm.io/gorm"

	"github.com/flyteorg/flyte/datacatalog/pkg/common"
	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/errors"
	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/models"
	idl_datacatalog "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
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
func (h *dataSetRepo) Create(ctx context.Context, id *idl_datacatalog.DatasetID, in models.Dataset) error {
	timer := h.repoMetrics.CreateDuration.Start(ctx)
	defer timer.Stop()

	result := h.db.WithContext(ctx).Create(&in)
	if result.Error != nil {
		return h.errorTransformer.ToDataCatalogError(result.Error)
	}
	return nil
}

// Get Dataset model
func (h *dataSetRepo) Get(ctx context.Context, id *idl_datacatalog.DatasetID) (models.Dataset, error) {
	timer := h.repoMetrics.GetDuration.Start(ctx)
	defer timer.Stop()

	var ds models.Dataset
	result := h.db.WithContext(ctx).Preload("PartitionKeys", func(db *gorm.DB) *gorm.DB {
		return db.WithContext(ctx).Order("partition_keys.created_at ASC") // preserve the order in which the partitions were created
	}).First(&ds, &models.Dataset{DatasetKey: models.DatasetKey{
		Project: id.Project,
		Domain:  id.Domain,
		Name:    id.Name,
		Version: id.Version,
		UUID:    id.UUID,
	}})

	if result.Error != nil {
		logger.Debugf(ctx, "Unable to find Dataset: [%+v], err: %v", id, result.Error)

		if result.Error.Error() == gorm.ErrRecordNotFound.Error() {
			return models.Dataset{}, errors.GetMissingEntityError("Dataset", id)
		}
		return models.Dataset{}, h.errorTransformer.ToDataCatalogError(result.Error)
	}

	return ds, nil
}

func (h *dataSetRepo) List(ctx context.Context, in models.ListModelsInput) ([]models.Dataset, error) {
	timer := h.repoMetrics.ListDuration.Start(ctx)
	defer timer.Stop()

	// apply filters and joins
	tx, err := applyListModelsInput(h.db, common.Dataset, in)
	if err != nil {
		return nil, err
	} else if tx.Error != nil {
		return []models.Dataset{}, h.errorTransformer.ToDataCatalogError(tx.Error)
	}

	datasets := make([]models.Dataset, 0)
	tx = tx.Preload("PartitionKeys").Find(&datasets)
	if tx.Error != nil {
		return []models.Dataset{}, h.errorTransformer.ToDataCatalogError(tx.Error)
	}
	return datasets, nil
}
