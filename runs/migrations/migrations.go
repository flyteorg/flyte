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

const MigrationIDInitSchema = "20260327_runs_init_schema"
const MigrationIDFixActionIndexes = "20260330_fix_action_indexes"

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
	{
		ID: MigrationIDFixActionIndexes,
		Migrate: func(tx *gorm.DB) error {
			return migrateFixActionIndexes(tx)
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},
}

// migrateFixActionIndexes drops high-contention indexes and creates replacements.
//
// idx_actions_updated and idx_actions_ended caused B-tree right-hand page contention:
// UpdateActionPhase writes updated_at/ended_at ≈ NOW() on every call, so all concurrent
// writers compete for the same rightmost leaf page, producing 200ms+ lock-wait latency.
//
// idx_actions_org/project/domain/run_name are replaced by the composite
// idx_actions_run_lookup(org, project, domain, run_name) which serves ListActions,
// AbortAction, MarkAbortAttempt, and ClearAbortRequest with a single selective index.
func migrateFixActionIndexes(db *gorm.DB) error {
	staleIndexes := []string{
		"idx_actions_updated",
		"idx_actions_ended",
		"idx_actions_org",
		"idx_actions_project",
		"idx_actions_domain",
		"idx_actions_run_name",
	}
	for _, idx := range staleIndexes {
		if err := db.Exec("DROP INDEX IF EXISTS " + idx).Error; err != nil {
			return fmt.Errorf("failed to drop stale index %s: %w", idx, err)
		}
	}
	// AutoMigrate picks up idx_actions_run_lookup from the updated model struct tags.
	return db.AutoMigrate(&models.Action{})
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
