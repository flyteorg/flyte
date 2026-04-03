package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"

	"github.com/flyteorg/flyte/v2/cache_service/repository/models"
)

var allModels = []interface{}{
	&models.CachedOutput{},
	&models.Reservation{},
}

const MigrationIDInitSchema = "20260327_cache_service_init_schema"

var CacheServiceMigrations = []*gormigrate.Migration{
	{
		ID: MigrationIDInitSchema,
		Migrate: func(tx *gorm.DB) error {
			return migrateInitSchema(tx)
		},
		Rollback: func(tx *gorm.DB) error {
			// Intentionally no-op for now; this migration can contain destructive steps.
			return nil
		},
	},
}

// migrateInitSchema initializes the cache service database schema.
func migrateInitSchema(db *gorm.DB) error {
	if err := db.AutoMigrate(allModels...); err != nil {
		return fmt.Errorf("failed to initialize cache service schema: %w", err)
	}
	return nil
}
