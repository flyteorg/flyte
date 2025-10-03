package migrations

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/flyteorg/flyte/v2/queue/repository"
)

// RunMigrations runs all database migrations for the queue service
func RunMigrations(db *gorm.DB) error {
	// AutoMigrate will create tables, missing columns, missing indexes
	// It will NOT delete columns or change column types
	if err := db.AutoMigrate(
		&repository.QueuedAction{},
	); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create additional indexes that GORM doesn't handle automatically
	// Note: GORM will create indexes defined in struct tags

	return nil
}
