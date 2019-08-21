package gormimpl

import (
	"context"

	"github.com/jinzhu/gorm"
	"github.com/lyft/datacatalog/pkg/repositories/errors"
	"github.com/lyft/datacatalog/pkg/repositories/interfaces"
	"github.com/lyft/datacatalog/pkg/repositories/models"
	idl_datacatalog "github.com/lyft/datacatalog/protos/gen"
)

type tagRepo struct {
	db               *gorm.DB
	errorTransformer errors.ErrorTransformer
	// TODO: add metrics
}

func NewTagRepo(db *gorm.DB, errorTransformer errors.ErrorTransformer) interfaces.TagRepo {
	return &tagRepo{
		db:               db,
		errorTransformer: errorTransformer,
	}
}

func (h *tagRepo) Create(ctx context.Context, tag models.Tag) error {
	db := h.db.Create(&tag)

	if db.Error != nil {
		return h.errorTransformer.ToDataCatalogError(db.Error)
	}
	return nil
}

func (h *tagRepo) Get(ctx context.Context, in models.TagKey) (models.Tag, error) {
	var tag models.Tag
	result := h.db.Preload("Artifact").Preload("Artifact.ArtifactData").Find(&tag, &models.Tag{
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
