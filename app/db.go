package app

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/flyteorg/flyte/v2/flytestdlib/database"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

// InitDB creates a gorm.DB from the given config.
func InitDB(ctx context.Context, cfg *database.DbConfig) (*gorm.DB, error) {
	db, err := database.GetDB(ctx, cfg, logger.GetConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	logger.Infof(ctx, "Database connection established")
	return db, nil
}
