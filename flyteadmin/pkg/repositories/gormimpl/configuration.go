package gormimpl

import (
	"context"
	"errors"

	"gorm.io/gorm"

	adminErrors "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

type ConfigurationDocumentRepo struct {
	db               *gorm.DB
	errorTransformer adminErrors.ErrorTransformer
	metrics          gormMetrics
}

func (r *ConfigurationDocumentRepo) GetActive(ctx context.Context) (models.ConfigurationDocument, error) {
	var configuration models.ConfigurationDocument
	timer := r.metrics.GetDuration.Start()
	tx := r.db.Where(&models.ConfigurationDocument{
		Active: true,
	}).First(&configuration)
	timer.Stop()
	// TODO: what if the table is empty?
	if tx.Error != nil && errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return models.ConfigurationDocument{}, adminErrors.GetSingletonMissingEntityError("configuration")
	} else if tx.Error != nil {
		return models.ConfigurationDocument{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return configuration, nil
}

func (r *ConfigurationDocumentRepo) Create(ctx context.Context, input *models.ConfigurationDocument) error {
	timer := r.metrics.CreateDuration.Start()
	err := r.db.Create(input).Error
	timer.Stop()
	if err != nil {
		return r.errorTransformer.ToFlyteAdminError(err)
	}
	return nil
}

func (r *ConfigurationDocumentRepo) Update(ctx context.Context, input *interfaces.UpdateConfigurationInput) error {
	timer := r.metrics.CreateDuration.Start()
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check if the active configuration document is outdated
		var configuration models.ConfigurationDocument
		err := tx.Where(&models.ConfigurationDocument{
			Active:  true,
			Version: input.VersionToUpdate,
		}).First(&configuration).Error

		if err != nil {
			return err
		}

		// Erase the active configuration document
		err = tx.Model(&models.ConfigurationDocument{}).Where(&models.ConfigurationDocument{
			Active:  true,
			Version: input.VersionToUpdate,
		}).Update("active", false).Error
		if err != nil {
			return err
		}

		// Create the new configuration document
		err = tx.Create(input.NewConfigurationDocument).Error
		if err != nil {
			return err
		}
		return nil
	})
	timer.Stop()
	if err != nil {
		return r.errorTransformer.ToFlyteAdminError(err)
	}
	return nil
}

// Returns an instance of ConfigurationDocumentRepoInterface
func NewConfigurationRepo(
	db *gorm.DB, errorTransformer adminErrors.ErrorTransformer, scope promutils.Scope) interfaces.ConfigurationDocumentRepoInterface {
	metrics := newMetrics(scope)
	return &ConfigurationDocumentRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}
