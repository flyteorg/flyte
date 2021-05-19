package gormimpl

import (
	"context"

	idl_datacatalog "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"

	"github.com/flyteorg/datacatalog/pkg/common"
	"github.com/flyteorg/datacatalog/pkg/repositories/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories/interfaces"
	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"gorm.io/gorm"
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
	result := h.db.Preload("PartitionKeys", func(db *gorm.DB) *gorm.DB {
		return db.Order("partition_keys.created_at ASC") // preserve the order in which the partitions were created
	}).First(&ds, &models.Dataset{DatasetKey: in})

	if result.Error != nil {
		logger.Debugf(ctx, "Unable to find Dataset: [%+v], err: %v", in, result.Error)

		if result.Error.Error() == gorm.ErrRecordNotFound.Error() {
			return models.Dataset{}, errors.GetMissingEntityError("Dataset", &idl_datacatalog.DatasetID{
				Project: in.Project,
				Domain:  in.Domain,
				Name:    in.Name,
				Version: in.Version,
			})
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
