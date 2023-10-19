package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/configuration"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/lib"
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

	// Check to see if the artifact key already exists.
	// this is not thread-safe, but whatever, gorm can't do this itself.
	var extantKey ArtifactKey
	db := r.db.Where("project = ? AND domain = ? and name = ?",
		serviceModel.Artifact.ArtifactId.ArtifactKey.Project,
		serviceModel.Artifact.ArtifactId.ArtifactKey.Domain,
		serviceModel.Artifact.ArtifactId.ArtifactKey.Name)
	db.First(&extantKey)
	if err := db.Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Errorf(ctx, "Failed to query for existing key: %+v", err)
			return models.Artifact{}, err
		}
		logger.Debugf(ctx, "Existing key [%+v] not found", extantKey)
	} else {
		gormModel.ArtifactKey = ArtifactKey{}
		gormModel.ArtifactKeyID = extantKey.ID
	}

	tx := r.db.Create(&gormModel)
	timer.Stop()
	if tx.Error != nil {
		logger.Errorf(ctx, "Failed to create artifact %+v", tx.Error)
		return models.Artifact{}, tx.Error
	}

	return models.Artifact{}, nil
}

func (r *RDSStorage) handleUriGet(ctx context.Context, uri string) (models.Artifact, error) {
	artifactID, tag, err := lib.ParseFlyteURL(uri)
	if err != nil {
		logger.Errorf(ctx, "Failed to parse uri [%s]: %+v", uri, err)
		return models.Artifact{}, err
	}
	if tag != "" {
		return models.Artifact{}, fmt.Errorf("tag not implemented yet")
	}
	logger.Debugf(ctx, "Extracted artifact id [%v] from uri [%s], using id handler", artifactID, uri)
	return r.handleArtifactIdGet(ctx, artifactID)
}

func (r *RDSStorage) handleArtifactIdGet(ctx context.Context, artifactID core.ArtifactID) (models.Artifact, error) {

	var gotArtifact models.Artifact
	db := r.db.Where("project = ? AND domain = ? and name = ?",
		artifactID.ArtifactKey.Project,
		artifactID.ArtifactKey.Domain,
		artifactID.ArtifactKey.Name)

	if artifactID.Version != "" {
		db = db.Where("version = ?", artifactID.Version)
	}
	if artifactID.GetPartitions() != nil && len(artifactID.GetPartitions().GetValue()) > 0 {
		partitionMap := PartitionsIdlToHstore(artifactID.GetPartitions())
		db = db.Where("partitions = ?", partitionMap)
	}
	db.Order("created_at desc").Limit(1)
	db = db.First(&gotArtifact)
	if err := db.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Infof(ctx, "Artifact not found: %+v", artifactID)
			return models.Artifact{}, fmt.Errorf("artifact [%v] not found", artifactID)
		}
		logger.Errorf(ctx, "Failed to query for artifact: %+v", err)
		return models.Artifact{}, err
	}
	return gotArtifact, nil
}

func (r *RDSStorage) GetArtifact(ctx context.Context, query core.ArtifactQuery) (models.Artifact, error) {
	timer := r.metrics.GetDuration.Start()

	var resp models.Artifact
	var err error
	if query.GetUri() != "" {
		logger.Debugf(ctx, "found uri in query: %+v", *query.GetArtifactId())
		resp, err = r.handleUriGet(ctx, query.GetUri())
	} else if query.GetArtifactId() != nil {
		logger.Debugf(ctx, "found artifact_id in query: %+v", *query.GetArtifactId())
		resp, err = r.handleArtifactIdGet(ctx, *query.GetArtifactId())
	} else if query.GetArtifactTag() != nil {
		return models.Artifact{}, fmt.Errorf("artifact tag not implemented yet")
	} else {
		return models.Artifact{}, fmt.Errorf("query must contain either uri, artifact_id, or artifact_tag")
	}
	timer.Stop()
	return resp, err
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
