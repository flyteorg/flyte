package impl

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/flyteorg/flyte/v2/flytestdlib/database"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

type settingsRepo struct {
	db *sqlx.DB
}

// NewSettingsRepo returns a SettingsRepo backed by Postgres.
func NewSettingsRepo(db *sqlx.DB) interfaces.SettingsRepo {
	return &settingsRepo{db: db}
}

func (r *settingsRepo) CreateSettings(ctx context.Context, settings *models.Settings) error {
	if len(settings.Data) == 0 {
		return fmt.Errorf("settings data is required: %s", settings.Key)
	}

	now := time.Now().UTC()
	settings.CreatedAt = now
	settings.UpdatedAt = now
	settings.Version = 1

	result, err := r.db.ExecContext(ctx,
		`INSERT INTO settings (key, data, version, created_at, updated_at)
                VALUES ($1, $2, $3, $4, $5)
                ON CONFLICT (key) DO NOTHING`,
		settings.Key, settings.Data, settings.Version, settings.CreatedAt, settings.UpdatedAt)
	if err != nil {
		if database.IsPgErrorWithCode(err, database.PgDuplicatedKey) {
			return fmt.Errorf("%w: %s", interfaces.ErrSettingsAlreadyExists, settings.Key)
		}
		return fmt.Errorf("failed to create settings %s: %w", settings.Key, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w: %s", interfaces.ErrSettingsAlreadyExists, settings.Key)
	}
	return nil
}

func (r *settingsRepo) GetSettings(ctx context.Context, key string) (*models.Settings, error) {
	var settings models.Settings
	err := sqlx.GetContext(ctx, r.db, &settings, "SELECT * FROM settings WHERE key = $1", key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %s", interfaces.ErrSettingsNotFound, key)
		}
		return nil, fmt.Errorf("failed to get settings %s: %w", key, err)
	}
	return &settings, nil
}

func (r *settingsRepo) GetSettingsByKeys(ctx context.Context, keys []string) ([]*models.Settings, error) {
	var settings []*models.Settings
	if err := sqlx.SelectContext(ctx, r.db, &settings, "SELECT * FROM settings WHERE key = ANY($1)", pq.Array(keys)); err != nil {
		return nil, fmt.Errorf("failed to get settings by keys: %w", err)
	}
	return settings, nil
}

func (r *settingsRepo) UpdateSettings(ctx context.Context, settings *models.Settings) error {
	now := time.Now().UTC()

	result, err := r.db.ExecContext(ctx,
		`UPDATE settings SET data = $1, version = version + 1, updated_at = $2 WHERE key = $3 AND version = $4`,
		settings.Data, now, settings.Key, settings.Version)
	if err != nil {
		return fmt.Errorf("failed to update settings %s: %w", settings.Key, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	// Zero rows means the version was stale or the row does not exist.
	// We distinguish it, so callers know whether to retry or to create.
	if rowsAffected == 0 {
		if _, getErr := r.GetSettings(ctx, settings.Key); getErr != nil {
			return getErr
		}
		return fmt.Errorf("%w: %s", interfaces.ErrSettingsVersionConflict, settings.Key)
	}

	// Keep the caller's struct matching the row the DB now holds.
	settings.Version++
	settings.UpdatedAt = now
	return nil
}
