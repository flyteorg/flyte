package db

import (
	"context"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/models"
	"github.com/flyteorg/flyte/flytestdlib/database"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"gorm.io/gorm"
)

// RDSStorage should implement StorageInterface
type RDSStorage struct {
	config  database.DbConfig
	db      *gorm.DB
	metrics gormMetrics
}

// WriteOne is a test function
func (r *RDSStorage) WriteOne(ctx context.Context, gormModel Artifact) (models.Artifact, error) {
	timer := r.metrics.CreateDuration.Start()
	logger.Debugf(ctx, "Attempt create artifact %s", gormModel.Version)
	tx := r.db.Omit("id").Create(&gormModel)
	timer.Stop()
	if tx.Error != nil {
		return models.Artifact{}, tx.Error
	}
	return models.Artifact{}, nil
}

func NewStorage(ctx context.Context, scope promutils.Scope) *RDSStorage {
	dbCfg := database.GetConfig()
	logConfig := logger.GetConfig()

	db, err := database.GetDB(ctx, dbCfg, logConfig)
	if err != nil {
		logger.Fatal(ctx, err)
	}
	return &RDSStorage{
		config:  *dbCfg,
		db:      db,
		metrics: newMetrics(scope.NewSubScope("rds")),
	}
}
