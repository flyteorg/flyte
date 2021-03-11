package gormimpl

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/jinzhu/gorm"
)

// Implementation of TaskExecutionInterface.
type TaskExecutionRepo struct {
	db               *gorm.DB
	errorTransformer errors.ErrorTransformer
	metrics          gormMetrics
}

func (r *TaskExecutionRepo) Create(ctx context.Context, input models.TaskExecution) error {
	timer := r.metrics.CreateDuration.Start()
	tx := r.db.Create(&input)
	timer.Stop()
	if tx.Error != nil {
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return nil
}

func (r *TaskExecutionRepo) Get(ctx context.Context, input interfaces.GetTaskExecutionInput) (models.TaskExecution, error) {
	var taskExecution models.TaskExecution
	timer := r.metrics.GetDuration.Start()
	tx := r.db.Where(&models.TaskExecution{
		TaskExecutionKey: models.TaskExecutionKey{
			TaskKey: models.TaskKey{
				Project: input.TaskExecutionID.TaskId.Project,
				Domain:  input.TaskExecutionID.TaskId.Domain,
				Name:    input.TaskExecutionID.TaskId.Name,
				Version: input.TaskExecutionID.TaskId.Version,
			},
			NodeExecutionKey: models.NodeExecutionKey{
				NodeID: input.TaskExecutionID.NodeExecutionId.NodeId,
				ExecutionKey: models.ExecutionKey{
					Project: input.TaskExecutionID.NodeExecutionId.ExecutionId.Project,
					Domain:  input.TaskExecutionID.NodeExecutionId.ExecutionId.Domain,
					Name:    input.TaskExecutionID.NodeExecutionId.ExecutionId.Name,
				},
			},
			RetryAttempt: &input.TaskExecutionID.RetryAttempt,
		},
	}).Preload("ChildNodeExecution").First(&taskExecution)
	timer.Stop()
	if tx.Error != nil {
		return models.TaskExecution{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	if tx.RecordNotFound() {
		return models.TaskExecution{},
			errors.GetMissingEntityError("task execution", &core.TaskExecutionIdentifier{
				TaskId: &core.Identifier{
					Project: input.TaskExecutionID.TaskId.Project,
					Domain:  input.TaskExecutionID.TaskId.Domain,
					Name:    input.TaskExecutionID.TaskId.Name,
					Version: input.TaskExecutionID.TaskId.Version,
				},
				NodeExecutionId: &core.NodeExecutionIdentifier{
					NodeId: input.TaskExecutionID.NodeExecutionId.NodeId,
					ExecutionId: &core.WorkflowExecutionIdentifier{
						Project: input.TaskExecutionID.NodeExecutionId.ExecutionId.Project,
						Domain:  input.TaskExecutionID.NodeExecutionId.ExecutionId.Domain,
						Name:    input.TaskExecutionID.NodeExecutionId.ExecutionId.Name,
					},
				},
			})
	}
	return taskExecution, nil
}

func (r *TaskExecutionRepo) Update(ctx context.Context, execution models.TaskExecution) error {
	timer := r.metrics.UpdateDuration.Start()
	tx := r.db.Save(&execution)
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
	tx := r.db.Limit(input.Limit).Offset(input.Offset).Preload("ChildNodeExecution")

	// And add three join conditions (joining multiple tables is fine even we only filter on a subset of table attributes).
	// We are joining on task -> taskExec->NodeExec -> Exec.
	// NOTE: the order in which the joins are called below are important because postgres will only know about certain
	// tables as they are joined. So we should do it in the order specified above.
	tx = tx.Joins(leftJoinTaskToTaskExec)
	tx = tx.Joins(innerJoinNodeExecToTaskExec)
	tx = tx.Joins(innerJoinExecToNodeExec)

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

// Returns an instance of TaskExecutionRepoInterface
func NewTaskExecutionRepo(
	db *gorm.DB, errorTransformer errors.ErrorTransformer, scope promutils.Scope) interfaces.TaskExecutionRepoInterface {
	metrics := newMetrics(scope)
	return &TaskExecutionRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}
