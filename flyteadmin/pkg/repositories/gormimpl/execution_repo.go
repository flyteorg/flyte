package gormimpl

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	adminErrors "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

// Implementation of ExecutionInterface.
type ExecutionRepo struct {
	db               *gorm.DB
	errorTransformer adminErrors.ErrorTransformer
	metrics          gormMetrics
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
	}).Take(&execution)
	timer.Stop()

	if tx.Error != nil && errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return models.Execution{}, adminErrors.GetMissingEntityError("execution", &core.Identifier{
			Project: input.Project,
			Domain:  input.Domain,
			Name:    input.Name,
		})
	} else if tx.Error != nil {
		return models.Execution{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return execution, nil
}

func (r *ExecutionRepo) Update(ctx context.Context, execution models.Execution) error {
	timer := r.metrics.UpdateDuration.Start()
	tx := r.db.WithContext(ctx).Model(&execution).Updates(execution)
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

	if ok := input.JoinTableEntities[common.ExecutionTag]; ok {
		tx = tx.Joins(fmt.Sprintf("INNER JOIN %s ON %s.execution_name = %s.execution_name",
			executionTagsTableName, executionTableName, executionTagsTableName))
	}

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

// Returns an instance of ExecutionRepoInterface
func NewExecutionRepo(
	db *gorm.DB, errorTransformer adminErrors.ErrorTransformer, scope promutils.Scope) interfaces.ExecutionRepoInterface {
	metrics := newMetrics(scope)
	return &ExecutionRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}
