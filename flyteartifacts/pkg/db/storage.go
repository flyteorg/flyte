package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/configuration"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/lib"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifact"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/database"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"strconv"
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

		if serviceModel.GetSource().GetWorkflowExecution() != nil {
			var extantWfExec WorkflowExecution
			we := WorkflowExecution{
				ExecutionProject: serviceModel.Artifact.ArtifactId.ArtifactKey.Project,
				ExecutionDomain:  serviceModel.Artifact.ArtifactId.ArtifactKey.Domain,
				ExecutionName:    serviceModel.Source.WorkflowExecution.Name,
			}
			tx.FirstOrCreate(&extantWfExec, we)
			if err := tx.Error; err != nil {
				logger.Errorf(ctx, "Failed to firstorcreate wf exec: %+v", err)
				return err
			}
			gormModel.WorkflowExecutionID = extantWfExec.ID
			gormModel.WorkflowExecution = WorkflowExecution{} // zero out the workflow execution
		}

		tx.Create(&gormModel)
		if tx.Error != nil {
			logger.Errorf(ctx, "Failed to create artifact %+v", tx.Error)
			return tx.Error
		}
		getSaved := tx.Preload("ArtifactKey").Preload("WorkflowExecution").First(&gormModel, "id = ?", gormModel.ID)
		if getSaved.Error != nil {
			logger.Errorf(ctx, "Failed to find artifact that was just saved: %+v", getSaved.Error)
			return getSaved.Error
		}
		return nil
	})
	if err != nil {
		logger.Errorf(ctx, "Failed transaction upsert on key [%v]: %+v", serviceModel.ArtifactId, err)
		return models.Artifact{}, err
	}
	timer.Stop()
	svcModel, err := GormToServiceModel(gormModel)
	if err != nil {
		// metric
		logger.Errorf(ctx, "Failed to convert gorm model to service model: %+v", err)
		return models.Artifact{}, err
	}

	return svcModel, nil
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
	db := r.db.Model(&Artifact{}).InnerJoins("ArtifactKey", r.db.Where(&ak)).Joins("WorkflowExecution")
	if artifactID.Version != "" {
		db = db.Where("version = ?", artifactID.Version)
	}

	// @eduardo - not actually sure if this is doing a strict match.
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
			// todo: return grpc errors at the service layer not here.
			return models.Artifact{}, status.Errorf(codes.NotFound, "artifact [%v] not found", artifactID)
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
	// TODO: Add a check to ensure that the artifact IDs that the trigger is triggering on are unique if more than one.

	// The first transaction just saves artifact keys, in case there are no artifacts
	// yet in the database that the trigger depends on.
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, ak := range dbTrigger.RunsOn {
			var extantKey ArtifactKey
			tx.FirstOrCreate(&extantKey, ak)
			if err := tx.Error; err != nil {
				logger.Errorf(ctx, "Failed to firstorcreate key: %+v", err)
				return err
			}
		}
		return nil
	})

	if err != nil {
		logger.Errorf(ctx, "Failed to pre-save artifact keys [%s]: %+v", trigger.Name, err)
		return models.Trigger{}, err
	}

	// Check to see if the trigger key already exists. Create if not
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

	m, err := GormToServiceTrigger(gotTrigger)
	if err != nil {
		logger.Errorf(ctx, "Failed to convert gorm model to service model: %+v", err)
		return models.Trigger{}, err
	}
	return m, nil
}

// GetTriggersByArtifactKey - Given an artifact key, presumably of a newly created artifact, find active triggers that
// trigger on that key (and potentially others).
func (r *RDSStorage) GetTriggersByArtifactKey(ctx context.Context, key core.ArtifactKey) ([]models.Trigger, error) {
	// First find the trigger keys relevant to the artifact key
	// then find the most recent, active, version of each trigger key.

	var triggerKey []TriggerKey
	db := r.db.Preload("RunsOn").
		Joins("inner join active_trigger_artifact_keys ug on ug.trigger_key_id = trigger_keys.id ").
		Joins("inner join artifact_keys g on g.id= ug.artifact_key_id ").
		Where("g.project = ? and g.domain = ? and g.name = ?", key.Project, key.Domain, key.Name).
		Find(&triggerKey)

	err := db.Error
	if err != nil {
		logger.Errorf(ctx, "Failed to find triggers for artifact key %v: %+v", key, err)
		return nil, err
	}
	logger.Debugf(ctx, "Found trigger keys: %+v for artifact key %v", triggerKey, key)
	if triggerKey == nil || len(triggerKey) == 0 {
		logger.Infof(ctx, "No triggers found for artifact key %v", key)
		return nil, nil
	}

	ts := make([]Trigger, len(triggerKey))

	var triggerCondition = r.db.Where(&triggerKey[0])
	if len(triggerKey) > 1 {
		for _, tk := range triggerKey[1:] {
			triggerCondition = triggerCondition.Or(&tk)
		}
	}

	db = r.db.Preload("RunsOn").Model(&Trigger{}).InnerJoins("TriggerKey", triggerCondition).Where("active = true").Find(&ts)
	if err := db.Error; err != nil {
		logger.Errorf(ctx, "Failed to query for triggers: %+v", err)
		return nil, err
	}
	logger.Debugf(ctx, "Found (%d) triggers %v", len(ts), ts)

	modelTriggers := make([]models.Trigger, len(ts))
	for i, t := range ts {
		st, err := GormToServiceTrigger(t)
		if err != nil {
			logger.Errorf(ctx, "Failed to convert gorm model to service model: %+v", err)
			return nil, err
		}
		modelTriggers[i] = st
	}

	return modelTriggers, nil
}

func (r *RDSStorage) SearchArtifacts(ctx context.Context, request artifact.SearchArtifactsRequest) ([]models.Artifact, string, error) {

	var offset = 0
	var err error
	if request.GetToken() != "" {
		offset, err = strconv.Atoi(request.GetToken())
		if err != nil {
			logger.Errorf(ctx, "Failed to convert offset to int: %+v", err)
			return nil, "", err
		}
	}

	var db *gorm.DB
	if request.GetOptions() != nil && request.GetOptions().LatestByKey {
		db = r.db.Select(`distinct on ("ArtifactKey"."project", "ArtifactKey"."domain", "ArtifactKey"."name") "artifacts".*`).Model(&Artifact{})
	} else {
		db = r.db.Model(&Artifact{})
	}

	if request.GetArtifactKey() != nil {
		gormAk := ArtifactKey{
			Project: request.GetArtifactKey().Project,
			Domain:  request.GetArtifactKey().Domain,
			Name:    request.GetArtifactKey().Name,
		}
		db = db.InnerJoins("ArtifactKey", r.db.Where(&gormAk))
	}

	if request.GetOptions() != nil && request.GetOptions().StrictPartitions {
		// strict means if partitions is not specified, then partitions is set to empty
		if request.GetPartitions() != nil && len(request.GetPartitions().GetValue()) > 0 {
			partitionMap := PartitionsIdlToHstore(request.GetPartitions())
			db = db.Where("partitions = ?", partitionMap)
		} else {
			db = db.Where("partitions is null or partitions = ''")
		}
	} else {
		// make this not-strict, and make the handleArtifactIdGet one strict.
		// this should be what the @> comparison does.
		if request.GetPartitions() != nil && len(request.GetPartitions().GetValue()) > 0 {
			partitionMap := PartitionsIdlToHstore(request.GetPartitions())
			db = db.Where("partitions @> ?", partitionMap)
		}
	}

	if request.Principal != "" {
		db = db.Where("principal = ?", request.Principal)
	}

	if request.Version != "" {
		db = db.Where("version = ?", request.Version)
	}

	var limit = 10
	if request.Limit != 0 {
		limit = int(request.Limit)
	}
	if request.GetOptions() != nil && request.GetOptions().LatestByKey {
		db.Order(`"ArtifactKey"."project", "ArtifactKey"."domain", "ArtifactKey"."name", created_at desc`)
	} else {
		db.Order("created_at desc")
	}

	var results []Artifact
	db.Limit(limit).Offset(offset).Find(&results)

	if len(results) == 0 {
		logger.Debugf(ctx, "No artifacts found for query: %+v", request)
		return nil, "", nil
	}
	var res = make([]models.Artifact, len(results))
	for i, ga := range results {
		a, err := GormToServiceModel(ga)
		if err != nil {
			logger.Errorf(ctx, "Failed to convert %v: %+v", ga, err)
			return nil, "", err
		}
		res[i] = a
	}

	return res, fmt.Sprintf("%d", offset+len(res)), nil
}

func (r *RDSStorage) SetExecutionInputs(ctx context.Context, req *artifact.ExecutionInputsRequest) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		var newInputs = make([]Artifact, len(req.GetInputs()))
		for i, a := range req.GetInputs() {
			if a.Version == "" {
				return fmt.Errorf("version must be specified for artifact %v", a.ArtifactKey)
			}
			ak := ArtifactKey{
				Project: a.ArtifactKey.Project,
				Domain:  a.ArtifactKey.Domain,
				Name:    a.ArtifactKey.Name,
			}
			var dbArtifact Artifact
			r.db.Model(&Artifact{}).InnerJoins("ArtifactKey", r.db.Where(&ak)).Where("version = ?", a.Version).First(&dbArtifact)
			newInputs[i] = dbArtifact
		}
		we := WorkflowExecution{
			ExecutionProject: req.ExecutionId.Project,
			ExecutionDomain:  req.ExecutionId.Domain,
			ExecutionName:    req.ExecutionId.Name,
		}
		var dbWe WorkflowExecution
		// Create the execution if it doesn't exist. When we hit this part, we're keeping track of what artifacts
		// an execution used as inputs. We need to create because the execution itself might otherwise never exist
		// in the artifact db if it doesn't itself produce artifacts.
		tx.FirstOrCreate(&dbWe, we)
		if err := tx.Error; err != nil {
			logger.Errorf(ctx, "Failed to firstorcreate workflow exec %v: %+v", we, err)
			return err
		}

		err := tx.Model(&dbWe).Association("InputArtifacts").Append(&newInputs)

		if err != nil {
			logger.Errorf(ctx, "Failed to update input artifacts: %+v", err)
			return err
		}

		return nil
	})

	return err
}

func (r *RDSStorage) findWorkflowInputs(ctx context.Context, we WorkflowExecution) ([]Artifact, error) {
	var usedAsInputs []Artifact
	var dbWe WorkflowExecution
	err := r.db.Preload("InputArtifacts").Where(we).First(&dbWe).Error
	if err != nil {
		logger.Errorf(ctx, "The requested workflow execution %v does not exist: %+v", we, err)
		return nil, err
	}
	err = r.db.Model(&dbWe).Association("InputArtifacts").Find(&usedAsInputs)
	if err != nil {
		logger.Errorf(ctx, "Failed to find artifacts used as inputs for workflow execution %v: %+v", we, err)
		return nil, err
	}
	return usedAsInputs, nil
}

func (r *RDSStorage) FindByWorkflowExec(ctx context.Context, request *artifact.FindByWorkflowExecRequest) ([]models.Artifact, error) {
	we := WorkflowExecution{
		ExecutionProject: request.ExecId.Project,
		ExecutionDomain:  request.ExecId.Domain,
		ExecutionName:    request.ExecId.Name,
	}
	var artifacts []Artifact
	var err error
	if request.Direction == artifact.FindByWorkflowExecRequest_INPUTS {
		// What are the artifacts that this workflow execution used as inputs?
		artifacts, err = r.findWorkflowInputs(ctx, we)
	} else {
		// What artifacts were produced as outputs by this workflow execution?
		db := r.db.Model(&Artifact{}).InnerJoins("WorkflowExecution", r.db.Where(&we))
		db.Order("created_at desc").Find(&artifacts)
		err = db.Error
	}

	if err != nil {
		logger.Errorf(ctx, "Failed to find artifacts for workflow execution: %+v", err)
		return nil, err
	}

	if len(artifacts) == 0 {
		logger.Debugf(ctx, "No artifacts found for workflow execution: %+v", request)
		return nil, nil
	}
	var res = make([]models.Artifact, len(artifacts))
	for i, ga := range artifacts {
		a, err := GormToServiceModel(ga)
		if err != nil {
			logger.Errorf(ctx, "Failed to convert %v: %+v", ga, err)
			return nil, err
		}
		res[i] = a
	}

	return res, nil
}

func NewStorage(ctx context.Context, scope promutils.Scope) *RDSStorage {
	dbCfg := configuration.ApplicationConfig.GetConfig().(*configuration.ApplicationConfiguration).ArtifactDatabaseConfig
	logConfig := logger.GetConfig()

	logger.Debugf(ctx, "Artifact database config: %v", dbCfg)

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
