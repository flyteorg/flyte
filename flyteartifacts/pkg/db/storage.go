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
	// do the create in a transaction
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var extantKey ArtifactKey
		ak := ArtifactKey{
			Project: serviceModel.Artifact.ArtifactId.ArtifactKey.Project,
			Domain:  serviceModel.Artifact.ArtifactId.ArtifactKey.Domain,
			Name:    serviceModel.Artifact.ArtifactId.ArtifactKey.Name,
		}
		tx.FirstOrCreate(&extantKey, ak)
		if err := tx.Error; err != nil {
			logger.Errorf(ctx, "Failed to firstorcreate key: %+v", err)
			return err
		}
		gormModel.ArtifactKeyID = extantKey.ID
		gormModel.ArtifactKey = ArtifactKey{} // zero out the artifact key
		tx.Create(&gormModel)
		if tx.Error != nil {
			logger.Errorf(ctx, "Failed to create artifact %+v", tx.Error)
			return tx.Error
		}
		return nil
	})
	if err != nil {
		logger.Errorf(ctx, "Failed transaction upsert on key [%v]: %+v", serviceModel.ArtifactId, err)
		return models.Artifact{}, err
	}
	timer.Stop()

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
	var gotArtifact Artifact
	ak := ArtifactKey{
		Project: artifactID.ArtifactKey.Project,
		Domain:  artifactID.ArtifactKey.Domain,
		Name:    artifactID.ArtifactKey.Name,
	}
	db := r.db.Model(&Artifact{}).InnerJoins("ArtifactKey", r.db.Where(&ak))
	if artifactID.Version != "" {
		db = db.Where("version = ?", artifactID.Version)
	}

	if artifactID.GetPartitions() != nil && len(artifactID.GetPartitions().GetValue()) > 0 {
		partitionMap := PartitionsIdlToHstore(artifactID.GetPartitions())
		db = db.Where("partitions = ?", partitionMap)
	} else {
		// Be strict about partitions. If not specified, that means, we're looking
		// for null
		db = db.Where("partitions is null")
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
	logger.Debugf(ctx, "Found and returning artifact key %v", gotArtifact)
	m, err := GormToServiceModel(gotArtifact)
	if err != nil {
		logger.Errorf(ctx, "Failed to convert gorm model to service model: %+v", err)
		return models.Artifact{}, err
	}
	return m, nil
}

func (r *RDSStorage) GetArtifact(ctx context.Context, query core.ArtifactQuery) (models.Artifact, error) {
	timer := r.metrics.GetDuration.Start()

	var resp models.Artifact
	var err error
	if query.GetUri() != "" {
		logger.Debugf(ctx, "found uri in query: %+v", query.GetUri())
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

func (r *RDSStorage) CreateTrigger(ctx context.Context, trigger models.Trigger) (models.Trigger, error) {

	timer := r.metrics.CreateTriggerDuration.Start()
	logger.Debugf(ctx, "Attempt create trigger [%s:%s]", trigger.Name, trigger.Version)
	dbTrigger := ServiceToGormTrigger(trigger)

	// Check to see if the trigger key already exists. Create if not
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var extantKey TriggerKey

		tx.FirstOrCreate(&extantKey, dbTrigger.TriggerKey)
		if err := tx.Error; err != nil {
			logger.Errorf(ctx, "Failed to firstorcreate key: %+v", err)
			return err
		}

		// Look for all earlier versions of the trigger and mark them inactive
		setFalse := tx.Model(&Trigger{}).Where("trigger_key_id = ?", extantKey.ID).Update("active", false)
		if tx.Error != nil {
			logger.Errorf(ctx, "transaction error marking earlier versions inactive for %s: %+v", dbTrigger.TriggerKey.Name, tx.Error)
			return tx.Error
		}
		if setFalse.Error != nil {
			logger.Errorf(ctx, "Failed to mark earlier versions inactive for %s: %+v", dbTrigger.TriggerKey.Name, setFalse.Error)
			return setFalse.Error
		}

		var artifactKeys []ArtifactKey
		// should we use tx here?
		db := r.db.Model(&ArtifactKey{}).Where(&dbTrigger.RunsOn).Find(&artifactKeys)
		if db.Error != nil {
			logger.Errorf(ctx, "Error %v", db.Error)
			return db.Error
		}
		if len(artifactKeys) != len(dbTrigger.RunsOn) {
			logger.Errorf(ctx, "Could not find all artifact keys for trigger: %+v, only found %v", dbTrigger.RunsOn, artifactKeys)
			return fmt.Errorf("could not find all artifact keys for trigger")
		}
		dbTrigger.RunsOn = artifactKeys

		dbTrigger.TriggerKeyID = extantKey.ID
		dbTrigger.TriggerKey = TriggerKey{} // zero out the artifact key
		// This create should update the join table between individual triggers and artifact keys
		tt := tx.Save(&dbTrigger)
		if tx.Error != nil || tt.Error != nil {
			if tx.Error != nil {
				logger.Errorf(ctx, "Transaction error: %v", tx.Error)
				return tx.Error
			}
			logger.Errorf(ctx, "Save query failed with: %v", tt.Error)
			return tt.Error
		}
		var savedTrigger Trigger
		tt = tx.Preload("TriggerKey").Preload("RunsOn").First(&savedTrigger, "id = ?", dbTrigger.ID)
		if tx.Error != nil || tt.Error != nil {
			if tx.Error != nil {
				logger.Errorf(ctx, "Transaction error: %v", tx.Error)
				return tx.Error
			}
			logger.Errorf(ctx, "Failed to find trigger that was just saved: %+v", tx.Error)
			return tt.Error
		}

		// Next update the active_trigger_artifact_keys join table that keeps track of active key relationships
		// That is, if you have a trigger_on=[artifactA, artifactB], this table links the trigger's name to those
		// artifact names. If you register a new version of the trigger that is just trigger_on=[artifactC], then
		// this table should just hold the reference to artifactC.
		err := tx.Model(&savedTrigger.TriggerKey).Association("RunsOn").Replace(savedTrigger.RunsOn)
		if err != nil {
			logger.Errorf(ctx, "Failed to update active_trigger_artifact_keys: %+v", err)
			return err
		}

		return nil
	})
	if err != nil {
		if database.IsPgErrorWithCode(err, database.PgDuplicatedKey) {
			logger.Infof(ctx, "Duplicate key detected, the current transaction will be cancelled: %s %s", trigger.Name, trigger.Version)
			// TODO: Replace with the retrieved Trigger object maybe
			// TODO: Add an error handling layer that translates from pg errors to a general service error.
			return models.Trigger{}, err
		} else {
			logger.Errorf(ctx, "Failed transaction upsert on key [%s]: %+v", trigger.Name, err)
			return models.Trigger{}, err
		}
	}
	timer.Stop()

	return models.Trigger{}, nil
}

func (r *RDSStorage) GetLatestTrigger(ctx context.Context, project, domain, name string) (models.Trigger, error) {
	var gotTrigger Trigger
	tk := TriggerKey{
		Project: project,
		Domain:  domain,
		Name:    name,
	}
	db := r.db.Model(&Trigger{}).InnerJoins("TriggerKey", r.db.Where(&tk))
	db = db.Where("active = true").Order("created_at desc").Limit(1).First(&gotTrigger)
	if err := db.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Infof(ctx, "Trigger not found: %+v", tk)
			return models.Trigger{}, fmt.Errorf("could not find a latest trigger")
		}
		logger.Errorf(ctx, "Failed to query for triggers: %+v", err)
		return models.Trigger{}, err
	}
	logger.Debugf(ctx, "Found and returning trigger obj %v", gotTrigger)

	m := GormToServiceTrigger(gotTrigger)
	return m, nil
}

func (r *RDSStorage) GetTriggersByArtifactKey(ctx context.Context, key core.ArtifactKey) ([]models.Trigger, error) {
	return nil, nil
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
