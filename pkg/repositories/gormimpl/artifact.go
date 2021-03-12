package gormimpl

import (
	"context"

	"github.com/flyteorg/datacatalog/pkg/common"
	"github.com/flyteorg/datacatalog/pkg/repositories/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories/interfaces"
	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/jinzhu/gorm"
)

type artifactRepo struct {
	db               *gorm.DB
	errorTransformer errors.ErrorTransformer
	repoMetrics      gormMetrics
}

func NewArtifactRepo(db *gorm.DB, errorTransformer errors.ErrorTransformer, scope promutils.Scope) interfaces.ArtifactRepo {
	return &artifactRepo{
		db:               db,
		errorTransformer: errorTransformer,
		repoMetrics:      newGormMetrics(scope),
	}
}

// Create the artifact in a transaction because ArtifactData will be created and associated along with it
func (h *artifactRepo) Create(ctx context.Context, artifact models.Artifact) error {
	timer := h.repoMetrics.CreateDuration.Start(ctx)
	defer timer.Stop()

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
	timer := h.repoMetrics.GetDuration.Start(ctx)
	defer timer.Stop()

	var artifact models.Artifact
	result := h.db.Preload("ArtifactData").
		Preload("Partitions", func(db *gorm.DB) *gorm.DB {
			return db.Order("partitions.created_at ASC") // preserve the order in which the partitions were created
		}).
		Preload("Tags").
		Order("artifacts.created_at DESC").
		First(
			&artifact,
			&models.Artifact{ArtifactKey: in},
		)

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

func (h *artifactRepo) List(ctx context.Context, datasetKey models.DatasetKey, in models.ListModelsInput) ([]models.Artifact, error) {
	timer := h.repoMetrics.ListDuration.Start(ctx)
	defer timer.Stop()

	artifacts := make([]models.Artifact, 0)
	sourceEntity := common.Artifact

	// add filter for dataset
	datasetUUIDFilter := NewGormValueFilter(common.Equal, "dataset_uuid", datasetKey.UUID)
	datasetFilter := models.ModelFilter{
		Entity:       common.Artifact,
		ValueFilters: []models.ModelValueFilter{datasetUUIDFilter},
	}
	in.ModelFilters = append(in.ModelFilters, datasetFilter)

	// apply filters and joins
	tx, err := applyListModelsInput(h.db, sourceEntity, in)

	if err != nil {
		return nil, err
	} else if tx.Error != nil {
		return []models.Artifact{}, h.errorTransformer.ToDataCatalogError(tx.Error)
	}

	tx = tx.Preload("ArtifactData").
		Preload("Partitions", func(db *gorm.DB) *gorm.DB {
			return db.Order("partitions.created_at ASC") // preserve the order in which the partitions were created
		}).
		Preload("Tags").Find(&artifacts)
	if tx.Error != nil {
		return []models.Artifact{}, h.errorTransformer.ToDataCatalogError(tx.Error)
	}
	return artifacts, nil
}
