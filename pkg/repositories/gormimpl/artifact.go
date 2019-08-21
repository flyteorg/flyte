package gormimpl

import (
	"context"

	"github.com/jinzhu/gorm"
	"github.com/lyft/datacatalog/pkg/repositories/errors"
	"github.com/lyft/datacatalog/pkg/repositories/interfaces"
	"github.com/lyft/datacatalog/pkg/repositories/models"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
)

type artifactRepo struct {
	db               *gorm.DB
	errorTransformer errors.ErrorTransformer
	// TODO: add metrics
}

func NewArtifactRepo(db *gorm.DB, errorTransformer errors.ErrorTransformer) interfaces.ArtifactRepo {
	return &artifactRepo{
		db:               db,
		errorTransformer: errorTransformer,
	}
}

func (h *artifactRepo) Create(ctx context.Context, artifact models.Artifact) error {
	// Create the artifact in a transaction because ArtifactData will be created and associated along with it
	tx := h.db.Begin()

	tx = tx.Create(&artifact)

	if tx.Error != nil {
		tx.Rollback()
		return h.errorTransformer.ToDataCatalogError(tx.Error)
	}

	tx = tx.Commit()
	if tx.Error != nil {
		return h.errorTransformer.ToDataCatalogError(tx.Error)
	}

	return nil
}

func (h *artifactRepo) Get(ctx context.Context, in models.ArtifactKey) (models.Artifact, error) {
	var artifact models.Artifact
	result := h.db.Preload("ArtifactData").Find(&artifact, &models.Artifact{
		ArtifactKey: in,
	})

	if result.Error != nil {
		return models.Artifact{}, h.errorTransformer.ToDataCatalogError(result.Error)
	}
	if result.RecordNotFound() {
		return models.Artifact{}, errors.GetMissingEntityError("Artifact", &datacatalog.Artifact{
			Dataset: &datacatalog.DatasetID{
				Project: in.DatasetProject,
				Domain:  in.DatasetDomain,
				Name:    in.DatasetName,
				Version: in.DatasetVersion,
			},
			Id: in.ArtifactID,
		})
	}

	return artifact, nil
}
