package migrations

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/flyteorg/flyte/v2/runs/repository"
)

// RunMigrations runs all database migrations for the runs service
func RunMigrations(db *gorm.DB) error {
	// AutoMigrate will create tables, missing columns, missing indexes
	// Order matters due to foreign key constraints
	if err := db.AutoMigrate(
		&repository.Run{},
		&repository.Action{},
		&repository.ActionAttempt{},
		&repository.ClusterEvent{},
		&repository.PhaseTransition{},
	); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
