package gormimpl

import (
	"context"

	"github.com/jinzhu/gorm"
	"github.com/lyft/datacatalog/pkg/repositories/errors"
	"github.com/lyft/datacatalog/pkg/repositories/interfaces"
	"github.com/lyft/datacatalog/pkg/repositories/models"
	idl_datacatalog "github.com/lyft/datacatalog/protos/gen"
	"github.com/lyft/flytestdlib/promutils"
)

type tagRepo struct {
	db               *gorm.DB
	errorTransformer errors.ErrorTransformer
	repoMetrics      gormMetrics
}

func NewTagRepo(db *gorm.DB, errorTransformer errors.ErrorTransformer, scope promutils.Scope) interfaces.TagRepo {
	return &tagRepo{
		db:               db,
		errorTransformer: errorTransformer,
		repoMetrics:      newGormMetrics(scope),
	}
}

func (h *tagRepo) Create(ctx context.Context, tag models.Tag) error {
	timer := h.repoMetrics.CreateDuration.Start(ctx)
	defer timer.Stop()

	db := h.db.Create(&tag)

	if db.Error != nil {
		return h.errorTransformer.ToDataCatalogError(db.Error)
	}
	return nil
}

func (h *tagRepo) Get(ctx context.Context, in models.TagKey) (models.Tag, error) {
	timer := h.repoMetrics.GetDuration.Start(ctx)
	defer timer.Stop()

	var tag models.Tag
	result := h.db.Preload("Artifact").Preload("Artifact.ArtifactData").
		Preload("Artifact.Partitions").Preload("Artifact.Tags").
		Order("tags.created_at DESC").
		First(&tag, &models.Tag{
			TagKey: in,
		})

	if result.Error != nil {
		return models.Tag{}, h.errorTransformer.ToDataCatalogError(result.Error)
	}
	if result.RecordNotFound() {
		return models.Tag{}, errors.GetMissingEntityError("Tag", &idl_datacatalog.Tag{
			Name: tag.TagName,
		})
	}

	return tag, nil
}
