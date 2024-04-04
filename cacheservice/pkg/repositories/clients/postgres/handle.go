package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/flyteorg/flyte/cacheservice/pkg/repositories/models"
	"github.com/flyteorg/flyte/flytestdlib/database"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

type DBHandle struct {
	DB *gorm.DB
}

func NewDBHandle(ctx context.Context, dbConfigValues *database.DbConfig) (*DBHandle, error) {
	logConfig := logger.GetConfig()
	db, err := database.GetDB(ctx, dbConfigValues, logConfig)
	if err != nil {
		return nil, err
	}

	return &DBHandle{
		DB: db,
	}, nil
}

func (h *DBHandle) Migrate(ctx context.Context) error {
	if err := h.DB.AutoMigrate(&models.CachedOutput{}); err != nil {
		return err
	}

	if err := h.DB.Debug().AutoMigrate(&models.CacheReservation{}); err != nil {
		return err
	}

	return nil
}
