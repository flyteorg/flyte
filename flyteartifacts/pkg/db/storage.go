package db

import (
	"context"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/configuration"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifact"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
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
func (r *RDSStorage) WriteOne(ctx context.Context, gormModel Artifact) (artifact.Artifact, error) {
	timer := r.metrics.CreateDuration.Start()
	logger.Debugf(ctx, "Attempt create artifact %s", gormModel.Version)
	tx := r.db.Create(&gormModel)
	timer.Stop()
	if tx.Error != nil {
		logger.Errorf(ctx, "Failed to create artifact %+v", tx.Error)
		return artifact.Artifact{}, tx.Error
	}
	return artifact.Artifact{}, nil
}

// CreateArtifact helps implement StorageInterface
func (r *RDSStorage) CreateArtifact(ctx context.Context, serviceModel models.Artifact) (models.Artifact, error) {
	timer := r.metrics.CreateDuration.Start()
	logger.Debugf(ctx, "Attempt create artifact [%s:%s]",
		gormModel.Artifact.ArtifactId.ArtifactKey.Name, gormModel.Artifact.ArtifactId.Version)
	tx := r.db.Create(&gormModel)
	timer.Stop()
	if tx.Error != nil {
		logger.Errorf(ctx, "Failed to create artifact %+v", tx.Error)
		return models.Artifact{}, tx.Error
	}

	return models.Artifact{}, nil
}

func (r *RDSStorage) GetArtifact(ctx context.Context, query core.ArtifactQuery, details bool) (models.Artifact, error) {
	return models.Artifact{}, nil
}

func NewStorage(ctx context.Context, scope promutils.Scope) *RDSStorage {
	dbCfg := configuration.ApplicationConfig.GetConfig().(*configuration.ApplicationConfiguration).ArtifactDatabaseConfig
	logConfig := logger.GetConfig()

	db, err := database.GetDB(ctx, &dbCfg, logConfig)
	if err != nil {
		logger.Fatal(ctx, err)
	}
	return &RDSStorage{
		config:  dbCfg,
		db:      db,
		metrics: newMetrics(scope),
	}
}
