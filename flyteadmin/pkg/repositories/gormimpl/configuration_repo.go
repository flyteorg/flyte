package gormimpl

import (
	"context"
	"errors"

	"gorm.io/gorm"

	flyteErrs "github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	adminErrors "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

// ConfigurationRepo implements interfaces.ConfigurationRepoInterface
type ConfigurationRepo struct {
	db               *gorm.DB
	errorTransformer adminErrors.ErrorTransformer
	metrics          gormMetrics
}

func (r *ConfigurationRepo) GetActive(ctx context.Context) (models.ConfigurationDocumentMetadata, error) {
	var metadata models.ConfigurationDocumentMetadata
	timer := r.metrics.GetDuration.Start()
	// Get the only one active configuration document
	tx := r.db.WithContext(ctx).Where(&models.ConfigurationDocumentMetadata{
		Active: true,
	}).Take(&metadata)
	timer.Stop()

	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			logger.Error(ctx, "No active configuration document found")
			return models.ConfigurationDocumentMetadata{}, adminErrors.GetSingletonMissingEntityError("configuration")
		}
		logger.Errorf(ctx, "Failed to get active configuration document with err %v", tx.Error)
		return models.ConfigurationDocumentMetadata{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return metadata, nil
}

func (r *ConfigurationRepo) Create(ctx context.Context, input *models.ConfigurationDocumentMetadata) error {
	timer := r.metrics.CreateDuration.Start()
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		activeMetadata := models.ConfigurationDocumentMetadata{}
		err := tx.Model(&models.ConfigurationDocumentMetadata{}).Where(&models.ConfigurationDocumentMetadata{
			Active: true,
		}).Take(&activeMetadata).Error
		if err == nil {
			// Found an active configuration document
			if activeMetadata.Version == input.Version {
				logger.Debugf(ctx, "The document with version %s is already active", input.Version)
				return nil
			}
			logger.Debugf(ctx, "When creating document with version %s, found an active document with version %s. Aborted creation.", input.Version, activeMetadata.Version)
			return flyteErrs.ActiveConfigurationDocumentAlreadyExistsError
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		// Create the new configuration document
		return tx.Create(input).Error
	})
	timer.Stop()
	if err != nil {
		if errors.Is(err, flyteErrs.ActiveConfigurationDocumentAlreadyExistsError) {
			return err
		}
		return r.errorTransformer.ToFlyteAdminError(err)
	}
	return nil
}

func (r *ConfigurationRepo) Update(ctx context.Context, input *interfaces.UpdateConfigurationInput) error {
	timer := r.metrics.CreateDuration.Start()
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Erase the active configuration document
		tx = tx.Model(&models.ConfigurationDocumentMetadata{}).Where(&models.ConfigurationDocumentMetadata{
			Active:  true,
			Version: input.VersionToUpdate,
		}).Update("active", false)
		err := tx.Error
		if err != nil {
			return err
		}
		if tx.RowsAffected == 0 {
			logger.Debugf(ctx, "No active configuration document found with version %s when update document with version %s", input.VersionToUpdate, input.NewConfigurationMetadata.Version)
			return flyteErrs.ConfigurationDocumentStaleError
		}

		// Create the new configuration document
		return tx.Create(input.NewConfigurationMetadata).Error
	})
	timer.Stop()
	if err != nil {
		if errors.Is(err, flyteErrs.ConfigurationDocumentStaleError) {
			return err
		}
		return r.errorTransformer.ToFlyteAdminError(err)
	}
	return nil
}

func NewConfigurationRepo(
	db *gorm.DB, errorTransformer adminErrors.ErrorTransformer, scope promutils.Scope) interfaces.ConfigurationRepoInterface {
	metrics := newMetrics(scope)
	return &ConfigurationRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}
