package gormimpl

import (
	"context"

	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"k8s.io/utils/clock"

	"github.com/flyteorg/flyte/datacatalog/pkg/common"
	catalogErrors "github.com/flyteorg/flyte/datacatalog/pkg/errors"
	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/errors"
	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

type artifactRepo struct {
	db               *gorm.DB
	errorTransformer errors.ErrorTransformer
	repoMetrics      gormMetrics
	clock            clock.Clock
}

func NewArtifactRepo(
	db *gorm.DB,
	errorTransformer errors.ErrorTransformer,
	scope promutils.Scope,
	clock clock.Clock,
) interfaces.ArtifactRepo {
	return &artifactRepo{
		db:               db,
		errorTransformer: errorTransformer,
		repoMetrics:      newGormMetrics(scope),
		clock:            clock,
	}
}

// Create the artifact in a transaction because ArtifactData will be created and associated along with it
func (h *artifactRepo) Create(ctx context.Context, artifact models.Artifact) error {
	timer := h.repoMetrics.CreateDuration.Start(ctx)
	defer timer.Stop()

	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tx = tx.
			Where("artifacts.expires_at is null or artifacts.expires_at < ?", h.clock.Now().UTC()).
			Order("artifacts.created_at DESC"). // Always pick the most recent
			Find(
				&models.Artifact{ArtifactKey: artifact.ArtifactKey},
			)

		if tx.Error != nil {
			return tx.Error
		}

		if tx.RowsAffected > 0 {
			return catalogErrors.NewDataCatalogErrorf(codes.AlreadyExists, "artifact already exists")
		}

		tx = tx.Create(&artifact)
		return tx.Error
	})

	if err != nil {
		return h.errorTransformer.ToDataCatalogError(err)
	}

	return nil
}

func (h *artifactRepo) GetAndFilterExpired(ctx context.Context, in models.ArtifactKey) (models.Artifact, error) {
	timer := h.repoMetrics.GetDuration.Start(ctx)
	defer timer.Stop()

	var artifact models.Artifact
	result := h.db.WithContext(ctx).
		Where("artifacts.expires_at is null or artifacts.expires_at < ?", h.clock.Now().UTC()).
		Preload("ArtifactData").
		Preload("Partitions", func(db *gorm.DB) *gorm.DB {
			return db.WithContext(ctx).Order("partitions.created_at ASC") // preserve the order in which the partitions were created
		}).
		Preload("Tags").
		Order("artifacts.created_at DESC"). // Always pick the most recent
		First(
			&artifact,
			&models.Artifact{ArtifactKey: in},
		)

	if result.Error != nil {
		if result.Error.Error() == gorm.ErrRecordNotFound.Error() {
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

		return models.Artifact{}, h.errorTransformer.ToDataCatalogError(result.Error)
	}

	return artifact, nil
}

func (h *artifactRepo) ListAndFilterExpired(ctx context.Context, datasetKey models.DatasetKey, in models.ListModelsInput) ([]models.Artifact, error) {
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
			return db.WithContext(ctx).Order("partitions.created_at ASC") // preserve the order in which the partitions were created
		}).
		Preload("Tags").Find(&artifacts)
	if tx.Error != nil {
		return []models.Artifact{}, h.errorTransformer.ToDataCatalogError(tx.Error)
	}
	return artifacts, nil
}

// Update updates the given artifact and its associated ArtifactData in database. The ArtifactData entries are upserted
// (ignoring conflicts, as no updates to the database model are to be expected) and any longer existing data is deleted.
func (h *artifactRepo) Update(ctx context.Context, artifact models.Artifact) error {
	timer := h.repoMetrics.UpdateDuration.Start(ctx)
	defer timer.Stop()

	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// ensure all artifact fields in DB are up-to-date
		if res := tx.Model(&models.Artifact{
			ArtifactKey: artifact.ArtifactKey,
		}).
			Where("artifacts.expires_at is null or artifacts.expires_at < ?", h.clock.Now().UTC()).
			Updates(artifact); res.Error != nil {
			return h.errorTransformer.ToDataCatalogError(res.Error)
		} else if res.RowsAffected == 0 {
			// no rows affected --> artifact not found
			return errors.GetMissingEntityError(string(common.Artifact), &datacatalog.Artifact{
				Dataset: &datacatalog.DatasetID{
					Project: artifact.DatasetProject,
					Domain:  artifact.DatasetDomain,
					Name:    artifact.DatasetName,
					Version: artifact.DatasetVersion,
				},
				Id: artifact.ArtifactID,
			})
		}

		artifactDataNames := make([]string, len(artifact.ArtifactData))
		for i := range artifact.ArtifactData {
			artifactDataNames[i] = artifact.ArtifactData[i].Name
			// ensure artifact data is fully associated with correct artifact
			artifact.ArtifactData[i].ArtifactKey = artifact.ArtifactKey
		}

		// delete all removed artifact data entries from the DB
		if err := tx.Where(&models.ArtifactData{ArtifactKey: artifact.ArtifactKey}).Where("name NOT IN ?", artifactDataNames).Delete(&models.ArtifactData{}).Error; err != nil {
			return h.errorTransformer.ToDataCatalogError(err)
		}

		// upsert artifact data, adding new entries and ignoring conflicts (no actual data changed)
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(artifact.ArtifactData).Error; err != nil {
			return h.errorTransformer.ToDataCatalogError(err)
		}

		return nil
	})

	return err
}
