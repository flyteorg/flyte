package gormimpl

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	flyteAdminDbErrors "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

// Implementation of TaskExecutionInterface.
type TaskExecutionRepo struct {
	db               *gorm.DB
	errorTransformer flyteAdminDbErrors.ErrorTransformer
	metrics          gormMetrics
}

func (r *TaskExecutionRepo) Create(ctx context.Context, input models.TaskExecution) error {
	timer := r.metrics.CreateDuration.Start()
	tx := r.db.WithContext(ctx).Omit("id").Create(&input)
	timer.Stop()
	if tx.Error != nil {
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return nil
}

func (r *TaskExecutionRepo) Get(ctx context.Context, input interfaces.GetTaskExecutionInput) (models.TaskExecution, error) {
	var taskExecution models.TaskExecution
	timer := r.metrics.GetDuration.Start()
	tx := r.db.WithContext(ctx).Where(&models.TaskExecution{
		TaskExecutionKey: models.TaskExecutionKey{
			TaskKey: models.TaskKey{
				Project: input.TaskExecutionID.GetTaskId().GetProject(),
				Domain:  input.TaskExecutionID.GetTaskId().GetDomain(),
				Name:    input.TaskExecutionID.GetTaskId().GetName(),
				Version: input.TaskExecutionID.GetTaskId().GetVersion(),
			},
			NodeExecutionKey: models.NodeExecutionKey{
				NodeID: input.TaskExecutionID.GetNodeExecutionId().GetNodeId(),
				ExecutionKey: models.ExecutionKey{
					Project: input.TaskExecutionID.GetNodeExecutionId().GetExecutionId().GetProject(),
					Domain:  input.TaskExecutionID.GetNodeExecutionId().GetExecutionId().GetDomain(),
					Name:    input.TaskExecutionID.GetNodeExecutionId().GetExecutionId().GetName(),
				},
			},
			RetryAttempt: &input.TaskExecutionID.RetryAttempt,
		},
	}).Preload("ChildNodeExecution").Take(&taskExecution)
	timer.Stop()

	if tx.Error != nil && errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return models.TaskExecution{},
			flyteAdminDbErrors.GetMissingEntityError("task execution", &core.TaskExecutionIdentifier{
				TaskId: &core.Identifier{
					Project: input.TaskExecutionID.GetTaskId().GetProject(),
					Domain:  input.TaskExecutionID.GetTaskId().GetDomain(),
					Name:    input.TaskExecutionID.GetTaskId().GetName(),
					Version: input.TaskExecutionID.GetTaskId().GetVersion(),
				},
				NodeExecutionId: &core.NodeExecutionIdentifier{
					NodeId: input.TaskExecutionID.GetNodeExecutionId().GetNodeId(),
					ExecutionId: &core.WorkflowExecutionIdentifier{
						Project: input.TaskExecutionID.GetNodeExecutionId().GetExecutionId().GetProject(),
						Domain:  input.TaskExecutionID.GetNodeExecutionId().GetExecutionId().GetDomain(),
						Name:    input.TaskExecutionID.GetNodeExecutionId().GetExecutionId().GetName(),
					},
				},
			})
	} else if tx.Error != nil {
		return models.TaskExecution{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return taskExecution, nil
}

func (r *TaskExecutionRepo) Update(ctx context.Context, execution models.TaskExecution) error {
	timer := r.metrics.UpdateDuration.Start()
	tx := r.db.WithContext(ctx).Model(&models.TaskExecution{}).Where(getIDFilter(execution.ID)).
		Updates(&execution)
	timer.Stop()

	if err := tx.Error; err != nil {
		return r.errorTransformer.ToFlyteAdminError(err)
	}
	return nil
}

func (r *TaskExecutionRepo) List(ctx context.Context, input interfaces.ListResourceInput) (interfaces.TaskExecutionCollectionOutput, error) {
	if err := ValidateListInput(input); err != nil {
		return interfaces.TaskExecutionCollectionOutput{}, err
	}

	var taskExecutions []models.TaskExecution
	tx := r.db.WithContext(ctx).Limit(input.Limit).Offset(input.Offset).Preload("ChildNodeExecution")

	// And add three join conditions
	// We enable joining on
	// - task x task exec
	// - node exec x task exec
	// - exec x task exec
	if input.JoinTableEntities[common.Task] {
		tx = tx.Joins(leftJoinTaskToTaskExec)
	}
	if input.JoinTableEntities[common.NodeExecution] {
		tx = tx.Joins(innerJoinNodeExecToTaskExec)
	}
	if input.JoinTableEntities[common.Execution] {
		tx = tx.Joins(innerJoinExecToTaskExec)
	}

	// Apply filters
	tx, err := applyScopedFilters(tx, input.InlineFilters, input.MapFilters)
	if err != nil {
		return interfaces.TaskExecutionCollectionOutput{}, err
	}

	// Apply sort ordering.
	if input.SortParameter != nil {
		tx = tx.Order(input.SortParameter.GetGormOrderExpr())
	}

	timer := r.metrics.ListDuration.Start()
	tx = tx.Find(&taskExecutions)
	timer.Stop()
	if tx.Error != nil {
		return interfaces.TaskExecutionCollectionOutput{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}

	return interfaces.TaskExecutionCollectionOutput{
		TaskExecutions: taskExecutions,
	}, nil
}

func (r *TaskExecutionRepo) Count(ctx context.Context, input interfaces.CountResourceInput) (int64, error) {
	var err error
	tx := r.db.WithContext(ctx).Model(&models.TaskExecution{})

	// And add three join conditions
	// We enable joining on
	// - task x task exec
	// - node exec x task exec
	// - exec x task exec
	if input.JoinTableEntities[common.Task] {
		tx = tx.Joins(leftJoinTaskToTaskExec)
	}
	if input.JoinTableEntities[common.NodeExecution] {
		tx = tx.Joins(innerJoinNodeExecToTaskExec)
	}
	if input.JoinTableEntities[common.Execution] {
		tx = tx.Joins(innerJoinExecToTaskExec)
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

// Returns an instance of TaskExecutionRepoInterface
func NewTaskExecutionRepo(
	db *gorm.DB, errorTransformer flyteAdminDbErrors.ErrorTransformer, scope promutils.Scope) interfaces.TaskExecutionRepoInterface {
	metrics := newMetrics(scope)
	return &TaskExecutionRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}
