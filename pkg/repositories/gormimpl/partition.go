package gormimpl

import (
	"context"

	"github.com/jinzhu/gorm"
	"github.com/lyft/datacatalog/pkg/repositories/errors"
	"github.com/lyft/datacatalog/pkg/repositories/interfaces"
	"github.com/lyft/datacatalog/pkg/repositories/models"
	"github.com/lyft/flytestdlib/promutils"
)

type partitionRepo struct {
	db               *gorm.DB
	errorTransformer errors.ErrorTransformer
	repoMetrics      gormMetrics
}

func NewPartitionRepo(db *gorm.DB, errorTransformer errors.ErrorTransformer, scope promutils.Scope) interfaces.PartitionRepo {
	return &partitionRepo{
		db:               db,
		errorTransformer: errorTransformer,
		repoMetrics:      newGormMetrics(scope),
	}
}

// Create a PartitionEntry model
func (h *partitionRepo) Create(ctx context.Context, in models.Partition) error {
	timer := h.repoMetrics.CreateDuration.Start(ctx)
	defer timer.Stop()

	result := h.db.Create(&in)
	if result.Error != nil {
		return h.errorTransformer.ToDataCatalogError(result.Error)
	}
	return nil
}
