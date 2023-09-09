package gormimpl

import (
	"context"

	idl_datacatalog "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"

	"github.com/flyteorg/datacatalog/pkg/repositories/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories/interfaces"
	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flytestdlib/promutils"
	"gorm.io/gorm"
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
	result := h.db.Preload("Artifact").
		Preload("Artifact.ArtifactData").
		Preload("Artifact.Partitions", func(db *gorm.DB) *gorm.DB {
			return db.Order("partitions.created_at ASC") // preserve the order in which the partitions were created
		}).
		Preload("Artifact.Tags").
		Order("tags.created_at DESC").
		First(&tag, &models.Tag{
			TagKey: in,
		})

	if result.Error != nil {
		if result.Error.Error() == gorm.ErrRecordNotFound.Error() {
			return models.Tag{}, errors.GetMissingEntityError("Tag", &idl_datacatalog.Tag{
				Name: tag.TagName,
			})
		}
		return models.Tag{}, h.errorTransformer.ToDataCatalogError(result.Error)
	}

	return tag, nil
}
