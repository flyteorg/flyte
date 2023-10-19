package db

import (
	"context"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/configuration"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/models"
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

// CreateArtifact helps implement StorageInterface
func (r *RDSStorage) CreateArtifact(ctx context.Context, serviceModel models.Artifact) (models.Artifact, error) {
	timer := r.metrics.CreateDuration.Start()
	logger.Debugf(ctx, "Attempt create artifact [%s:%s]",
		serviceModel.Artifact.ArtifactId.ArtifactKey.Name, serviceModel.Artifact.ArtifactId.Version)
	gormModel, err := ServiceToGormModel(serviceModel)
	if err != nil {
		logger.Errorf(ctx, "Failed to convert service model to gorm model: %+v", err)
		return models.Artifact{}, err
	}
	tx := r.db.Create(&gormModel)
	timer.Stop()
	if tx.Error != nil {
		logger.Errorf(ctx, "Failed to create artifact %+v", tx.Error)
		return models.Artifact{}, tx.Error
	}

	return models.Artifact{}, nil
}

func (r *RDSStorage) GetArtifact(ctx context.Context, query core.ArtifactQuery, details bool) (models.Artifact, error) {
	timer := r.metrics.GetDuration.Start()

	//var artifacts []models.Artifact

	timer.Stop()
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
