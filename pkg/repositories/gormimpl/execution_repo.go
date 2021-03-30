package gormimpl

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyteadmin/pkg/common"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/jinzhu/gorm"
)

// Implementation of ExecutionInterface.
type ExecutionRepo struct {
	db               *gorm.DB
	errorTransformer errors.ErrorTransformer
	metrics          gormMetrics
}

func (r *ExecutionRepo) Create(ctx context.Context, input models.Execution) error {
	timer := r.metrics.CreateDuration.Start()
	tx := r.db.Create(&input)
	timer.Stop()
	if tx.Error != nil {
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return nil
}

func (r *ExecutionRepo) Get(ctx context.Context, input interfaces.GetResourceInput) (models.Execution, error) {
	var execution models.Execution
	timer := r.metrics.GetDuration.Start()
	tx := r.db.Where(&models.Execution{
		ExecutionKey: models.ExecutionKey{
			Project: input.Project,
			Domain:  input.Domain,
			Name:    input.Name,
		},
	}).Take(&execution)
	timer.Stop()
	if tx.Error != nil {
		return models.Execution{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	if tx.RecordNotFound() {
		return models.Execution{}, errors.GetMissingEntityError("execution", &core.Identifier{
			Project: input.Project,
			Domain:  input.Domain,
			Name:    input.Name,
		})
	}
	return execution, nil
}

func (r *ExecutionRepo) Update(ctx context.Context, event models.ExecutionEvent, execution models.Execution) error {
	timer := r.metrics.UpdateDuration.Start()
	defer timer.Stop()
	// Use a transaction to guarantee no partial updates.
	tx := r.db.Begin()
	if err := tx.Create(&event).Error; err != nil {
		tx.Rollback()
		return r.errorTransformer.ToFlyteAdminError(err)
	}
	if err := r.db.Model(&execution).Updates(execution).Error; err != nil {
		tx.Rollback()
		return r.errorTransformer.ToFlyteAdminError(err)
	}
	if err := tx.Commit().Error; err != nil {
		return r.errorTransformer.ToFlyteAdminError(err)
	}
	return nil
}

func (r *ExecutionRepo) UpdateExecution(ctx context.Context, execution models.Execution) error {
	timer := r.metrics.UpdateDuration.Start()
	tx := r.db.Model(&execution).Updates(execution)
	timer.Stop()
	if err := tx.Error; err != nil {
		return r.errorTransformer.ToFlyteAdminError(err)
	}
	return nil
}

func (r *ExecutionRepo) List(ctx context.Context, input interfaces.ListResourceInput) (
	interfaces.ExecutionCollectionOutput, error) {
	// First validate input.
	if err := ValidateListInput(input); err != nil {
		return interfaces.ExecutionCollectionOutput{}, err
	}
	var executions []models.Execution
	tx := r.db.Limit(input.Limit).Offset(input.Offset)
	// And add join condition as required by user-specified filters (which can potentially include join table attrs).
	if ok := input.JoinTableEntities[common.LaunchPlan]; ok {
		tx = tx.Joins(fmt.Sprintf("INNER JOIN %s ON %s.launch_plan_id = %s.id",
			launchPlanTableName, executionTableName, launchPlanTableName))
	}
	if ok := input.JoinTableEntities[common.Workflow]; ok {
		tx = tx.Joins(fmt.Sprintf("INNER JOIN %s ON %s.workflow_id = %s.id",
			workflowTableName, executionTableName, workflowTableName))
	}
	if ok := input.JoinTableEntities[common.Task]; ok {
		tx = tx.Joins(fmt.Sprintf("INNER JOIN %s ON %s.task_id = %s.id",
			taskTableName, executionTableName, taskTableName))
	}

	// Apply filters
	tx, err := applyScopedFilters(tx, input.InlineFilters, input.MapFilters)
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

// Returns an instance of ExecutionRepoInterface
func NewExecutionRepo(
	db *gorm.DB, errorTransformer errors.ErrorTransformer, scope promutils.Scope) interfaces.ExecutionRepoInterface {
	metrics := newMetrics(scope)
	return &ExecutionRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}
