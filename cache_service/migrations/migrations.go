package migrations

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/flyteorg/flyte/v2/cache_service/repository/models"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

var allModels = []interface{}{
	&models.CachedOutput{},
	&models.Reservation{},
}

func RunMigrations(db *gorm.DB) error {
	if err := db.AutoMigrate(allModels...); err != nil {
		return fmt.Errorf("failed to run cache service migrations: %w", err)
	}

	logger.Infof(context.Background(), "Cache service migrations completed successfully")
	return nil
}
