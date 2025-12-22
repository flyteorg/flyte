package migrations

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// RunMigrations runs all database migrations for the runs service
func RunMigrations(db *gorm.DB) error {
	ctx := context.Background()

	// Drop ALL old tables if they exist (from previous schema)
	// This includes the old actions table which had different columns
	oldTables := []string{"runs", "actions", "action_attempts", "cluster_events", "phase_transitions"}
	for _, table := range oldTables {
		if db.Migrator().HasTable(table) {
			logger.Infof(ctx, "Dropping old table: %s", table)
			if err := db.Migrator().DropTable(table); err != nil {
				return fmt.Errorf("failed to drop table %s: %w", table, err)
			}
		}
	}

	// AutoMigrate will create the new actions table with simplified schema
	if err := db.AutoMigrate(&models.Action{}); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Infof(ctx, "Database migrations completed successfully (recreated actions table)")
	return nil
}
