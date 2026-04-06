package app

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/flyteorg/flyte/v2/flytestdlib/database"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

// InitDB creates a sqlx.DB from the given config.
func InitDB(ctx context.Context, cfg *database.DbConfig) (*sqlx.DB, error) {
	db, err := database.GetDB(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	logger.Infof(ctx, "Database connection established")
	return db, nil
}
