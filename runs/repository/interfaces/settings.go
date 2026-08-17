package interfaces

import (
	"context"
	"errors"

	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

var (
	ErrSettingsNotFound        = errors.New("settings not found")
	ErrSettingsVersionConflict = errors.New("settings version conflict")
	ErrSettingsAlreadyExists   = errors.New("settings already exists")
)

type SettingsRepo interface {
	CreateSettings(ctx context.Context, settings *models.Settings) error
	GetSettings(ctx context.Context, key string) (*models.Settings, error)
	// GetSettingsByKeys returns the settings rows for the given keys. Keys with
	// no stored row are simply absent from the result; missing rows are not an error.
	GetSettingsByKeys(ctx context.Context, keys []string) ([]*models.Settings, error)
	// UpdateSettings persists settings for its key only if the row still has
	// settings.Version; returns ErrSettingsVersionConflict when the stored version differs,
	// and ErrSettingsNotFound when no row exists for the key.
	UpdateSettings(ctx context.Context, settings *models.Settings) error
}
