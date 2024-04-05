package gormimpl

import (
	"context"
	"errors"
	adminErrors "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"gorm.io/gorm"
)

type ConfigurationRepo struct {
	db               *gorm.DB
	errorTransformer adminErrors.ErrorTransformer
	metrics          gormMetrics
}

func (r *ConfigurationRepo) GetActive(ctx context.Context) (models.Configuration, error) {
	var configuration models.Configuration
	timer := r.metrics.GetDuration.Start()
	tx := r.db.Where(&models.Configuration{
		Active: true,
	}).First(&configuration)
	timer.Stop()
	// TODO: what if the table is empty?
	if tx.Error != nil && errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return models.Configuration{}, adminErrors.GetSingletonMissingEntityError("configuration")
	} else if tx.Error != nil {
		return models.Configuration{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return configuration, nil
}

func (r *ConfigurationRepo) EraseActiveAndCreate(ctx context.Context, versionToUpdate string, newConfiguration models.Configuration) error {
	timer := r.metrics.CreateDuration.Start()
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check if the active override attributes are outdated
		var configuration models.Configuration
		err := tx.Where(&models.Configuration{
			Active:  true,
			Version: versionToUpdate,
		}).First(&configuration).Error
		if err != nil {
			return err
		}

		// Erase the active override attributes
		err = tx.Model(&models.Configuration{}).Where(&models.Configuration{
			Active:  true,
			Version: versionToUpdate,
		}).Update("active", false).Error
		if err != nil {
			return err
		}

		// Create the new override attributes
		err = tx.Create(&newConfiguration).Error
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

// Returns an instance of ConfigurationRepoInterface
func NewConfigurationRepo(
	db *gorm.DB, errorTransformer adminErrors.ErrorTransformer, scope promutils.Scope) interfaces.ConfigurationInterface {
	metrics := newMetrics(scope)
	return &ConfigurationRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}
