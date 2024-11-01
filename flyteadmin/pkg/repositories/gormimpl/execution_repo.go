package gormimpl

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	adminErrors "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

const phase = "phase"

var terminalPhases = []any{
	core.WorkflowExecution_SUCCEEDED.String(),
	core.WorkflowExecution_FAILED.String(),
	core.WorkflowExecution_TIMED_OUT.String(),
	core.WorkflowExecution_ABORTED.String(),
}

var nonTerminalExecutions = []any{
	core.WorkflowExecution_UNDEFINED.String(),
	core.WorkflowExecution_QUEUED.String(),
	core.WorkflowExecution_RUNNING.String(),
	core.WorkflowExecution_SUCCEEDING.String(),
	core.WorkflowExecution_FAILING.String(),
	core.WorkflowExecution_ABORTING.String(),
}

// Implementation of ExecutionInterface.
type ExecutionRepo struct {
	db               *gorm.DB
	errorTransformer adminErrors.ErrorTransformer
	metrics          gormMetrics
}

func applyJoinTableEntitiesOnExecution(tx *gorm.DB, joinTableEntities map[common.Entity]bool) *gorm.DB {
	if ok := joinTableEntities[common.LaunchPlan]; ok {
		tx = tx.Joins(fmt.Sprintf("INNER JOIN %s ON %s.launch_plan_id = %s.id",
			launchPlanTableName, executionTableName, launchPlanTableName))
	}
	if ok := joinTableEntities[common.Workflow]; ok {
		tx = tx.Joins(fmt.Sprintf("INNER JOIN %s ON %s.workflow_id = %s.id",
			workflowTableName, executionTableName, workflowTableName))
	}
	if ok := joinTableEntities[common.Task]; ok {
		tx = tx.Joins(fmt.Sprintf("INNER JOIN %s ON %s.task_id = %s.id",
			taskTableName, executionTableName, taskTableName))
	}

	if ok := joinTableEntities[common.ExecutionTag]; ok {
		tx = tx.Joins(fmt.Sprintf("INNER JOIN %s ON %s.execution_name = %s.execution_name AND %s.execution_org = %s.execution_org",
			executionTagsTableName, executionTableName, executionTagsTableName, executionTableName, executionTagsTableName))
	}
	return tx
}

func (r *ExecutionRepo) Create(ctx context.Context, input models.Execution, executionTagModel []*models.ExecutionTag) error {
	timer := r.metrics.CreateDuration.Start()
	err := r.db.WithContext(ctx).Transaction(func(_ *gorm.DB) error {
		if len(executionTagModel) > 0 {
			tx := r.db.WithContext(ctx).Create(executionTagModel)
			if tx.Error != nil {
				return r.errorTransformer.ToFlyteAdminError(tx.Error)
			}
		}

		tx := r.db.WithContext(ctx).Create(&input)
		if tx.Error != nil {
			return r.errorTransformer.ToFlyteAdminError(tx.Error)
		}
		return nil
	})
	timer.Stop()
	return err
}

func (r *ExecutionRepo) Get(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
	var execution models.Execution
	timer := r.metrics.GetDuration.Start()
	tx := r.db.WithContext(ctx).Where(&models.Execution{
		ExecutionKey: models.ExecutionKey{
			Project: input.Project,
			Domain:  input.Domain,
			Name:    input.Name,
		},
	}).Where(getExecutionOrgFilter(input.Org)).Take(&execution)
	timer.Stop()

	if tx.Error != nil && errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return models.Execution{}, adminErrors.GetMissingEntityError("execution", &core.Identifier{
			Project: input.Project,
			Domain:  input.Domain,
			Name:    input.Name,
			Org:     input.Org,
		})
	} else if tx.Error != nil {
		return models.Execution{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return execution, nil
}

func (r *ExecutionRepo) Update(ctx context.Context, execution models.Execution) error {
	timer := r.metrics.UpdateDuration.Start()
	tx := r.db.WithContext(ctx).Model(&models.Execution{}).Where(getIdFilter(execution.ID)).Updates(execution)
	timer.Stop()
	if err := tx.Error; err != nil {
		return r.errorTransformer.ToFlyteAdminError(err)
	}
	return nil
}

func (r *ExecutionRepo) List(ctx context.Context, input interfaces.ListResourceInput) (
	interfaces.ExecutionCollectionOutput, error) {
	var err error
	// First validate input.
	if err = ValidateListInput(input); err != nil {
		return interfaces.ExecutionCollectionOutput{}, err
	}
	var executions []models.Execution
	tx := r.db.WithContext(ctx).Limit(input.Limit).Offset(input.Offset)

	// Add join condition as required by user-specified filters (which can potentially include join table attrs).
	tx = applyJoinTableEntitiesOnExecution(tx, input.JoinTableEntities)

	// Apply filters
	tx, err = applyScopedFilters(tx, input.InlineFilters, input.MapFilters)
	if err != nil {
		return interfaces.ExecutionCollectionOutput{}, err
	}
	// Apply sort ordering.
	if input.SortParameter != nil {
		tx = tx.Order(input.SortParameter.GetGormOrderExpr())
	}

	timer := r.metrics.ListDuration.Start()
	tx = tx.Find(&executions)
	timer.Stop()
	if tx.Error != nil {
		return interfaces.ExecutionCollectionOutput{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return interfaces.ExecutionCollectionOutput{
		Executions: executions,
	}, nil
}

func (r *ExecutionRepo) Count(ctx context.Context, input interfaces.CountResourceInput) (int64, error) {
	var err error
	tx := r.db.WithContext(ctx).Model(&models.Execution{})

	// Add join condition as required by user-specified filters (which can potentially include join table attrs).
	tx = applyJoinTableEntitiesOnExecution(tx, input.JoinTableEntities)

	// Apply filters
	tx, err = applyScopedFilters(tx, input.InlineFilters, input.MapFilters)
	if err != nil {
		return 0, err
	}

	// Run the query
	timer := r.metrics.CountDuration.Start()
	var count int64
	tx = tx.Count(&count)
	timer.Stop()
	if tx.Error != nil {
		return 0, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return count, nil
}

func (r *ExecutionRepo) CountByPhase(ctx context.Context, input interfaces.CountResourceInput) (interfaces.ExecutionCountsByPhaseOutput, error) {
	var err error
	tx := r.db.WithContext(ctx).Model(&models.Execution{})

	// Add join condition as required by user-specified filters (which can potentially include join table attrs).
	tx = applyJoinTableEntitiesOnExecution(tx, input.JoinTableEntities)

	// Apply filters
	tx, err = applyScopedFilters(tx, input.InlineFilters, input.MapFilters)
	if err != nil {
		return interfaces.ExecutionCountsByPhaseOutput{}, err
	}

	// Prepare query and output
	query := fmt.Sprintf("%s as phase, COUNT(%s) as count", phase, phase)
	var counts interfaces.ExecutionCountsByPhaseOutput

	// Run the query
	timer := r.metrics.CountDuration.Start()
	tx = tx.Select(query).Group(phase).Scan(&counts)
	timer.Stop()
	if tx.Error != nil {
		return interfaces.ExecutionCountsByPhaseOutput{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return counts, nil
}

func (r *ExecutionRepo) FindStatusUpdates(ctx context.Context, cluster string, id uint, limit, offset int) ([]interfaces.ExecutionStatus, error) {
	defer r.metrics.FindStatusUpdates.Start().Stop()

	var statuses []interfaces.ExecutionStatus
	err := r.db.Model(&models.Execution{}).
		WithContext(ctx).
		Where("parent_cluster = ? AND id > ?", cluster, id).
		Order("id ASC").
		Limit(limit).
		Offset(offset).
		Find(&statuses).
		Error
	if err != nil {
		logger.Errorf(ctx, "failed to find execution status updates: %v", err)
		return nil, r.errorTransformer.ToFlyteAdminError(err)
	}

	return statuses, nil
}

// FindNextStatusUpdatesCheckpoint for a given parent cluster tries to find latest terminal execution
// before earliest non-terminal execution after given id.
// If there are no such, it tries to find just any latest terminal execution for given parent cluster after given id.
// If table is empty, returns default value 0. DB auto-increment always starts from 1.
func (r *ExecutionRepo) FindNextStatusUpdatesCheckpoint(ctx context.Context, cluster string, id uint) (uint, error) {
	defer r.metrics.FindNextCheckpoint.Start().Stop()

	var execIDs []uint
	err := r.db.
		WithContext(ctx).
		Raw(`
SELECT id 
FROM executions
WHERE parent_cluster = @parent_cluster AND phase IN @terminal_phases AND id > @id AND 
    COALESCE(id < (
		SELECT id FROM executions
		WHERE parent_cluster = @parent_cluster AND phase IN @non_terminal_phases AND id > @id
		ORDER BY id ASC
		LIMIT 1
	), TRUE)
ORDER BY id DESC
LIMIT 1`,
			sql.Named("parent_cluster", cluster),
			sql.Named("terminal_phases", terminalPhases),
			sql.Named("non_terminal_phases", nonTerminalExecutions),
			sql.Named("id", id)).
		Scan(&execIDs).
		Error
	if err != nil {
		logger.Errorf(ctx, "failed to find next status updates checkpoint: %v", err)
		return 0, r.errorTransformer.ToFlyteAdminError(err)
	}

	if len(execIDs) == 0 {
		return id, nil
	}
	return execIDs[0], nil
}

// NewExecutionRepo returns an instance of ExecutionRepoInterface
func NewExecutionRepo(
	db *gorm.DB, errorTransformer adminErrors.ErrorTransformer, scope promutils.Scope) interfaces.ExecutionRepoInterface {
	metrics := newMetrics(scope)
	return &ExecutionRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}
