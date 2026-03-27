package migrations

import (
	"context"
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// AllModels contains all GORM models used in the runs service
var AllModels = []interface{}{
	&models.Action{},
	&models.ActionEvent{},
	&models.Project{},
	&models.Task{},
	&models.TaskSpec{},
}

const MigrationIDInitSchema = "20260327_001_runs_init_schema"

var RunsMigrations = []*gormigrate.Migration{
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

// migrateInitSchema initializes the runs service database schema.
func migrateInitSchema(db *gorm.DB) error {
	ctx := context.Background()

	// Drop stale tables from previous schema versions that are no longer used.
	// "actions" is intentionally excluded since it holds live data.
	oldTables := []string{"runs", "action_attempts", "cluster_events", "phase_transitions"}
	for _, table := range oldTables {
		if db.Migrator().HasTable(table) {
			logger.Infof(ctx, "Dropping old table: %s", table)
			if err := db.Migrator().DropTable(table); err != nil {
				return fmt.Errorf("failed to drop table %s: %w", table, err)
			}
		}
	}

	// AutoMigrate creates missing tables and adds new columns without dropping existing data.
	if err := db.AutoMigrate(AllModels...); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Infof(ctx, "Database migrations completed successfully")
	return nil
}
