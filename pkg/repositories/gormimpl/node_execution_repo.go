package gormimpl

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/jinzhu/gorm"
)

// Implementation of NodeExecutionInterface.
type NodeExecutionRepo struct {
	db               *gorm.DB
	errorTransformer errors.ErrorTransformer
	metrics          gormMetrics
}

// Persist the node execution and the initial event that triggers this execution. If any of the persistence fails
// rollback the transaction all together.
func (r *NodeExecutionRepo) Create(ctx context.Context, event *models.NodeExecutionEvent, execution *models.NodeExecution) error {
	timer := r.metrics.CreateDuration.Start()
	defer timer.Stop()
	// Use a transaction to guarantee no partial updates in
	// creating the execution and event
	tx := r.db.Begin()
	if err := tx.Create(&execution).Error; err != nil {
		tx.Rollback()
		return r.errorTransformer.ToFlyteAdminError(err)
	}

	if err := tx.Create(&event).Error; err != nil {
		tx.Rollback()
		return r.errorTransformer.ToFlyteAdminError(err)
	}
	if err := tx.Commit().Error; err != nil {
		return r.errorTransformer.ToFlyteAdminError(err)
	}
	return nil
}

func (r *NodeExecutionRepo) Get(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
	var nodeExecution models.NodeExecution
	timer := r.metrics.GetDuration.Start()
	tx := r.db.Where(&models.NodeExecution{
		NodeExecutionKey: models.NodeExecutionKey{
			NodeID: input.NodeExecutionIdentifier.NodeId,
			ExecutionKey: models.ExecutionKey{
				Project: input.NodeExecutionIdentifier.ExecutionId.Project,
				Domain:  input.NodeExecutionIdentifier.ExecutionId.Domain,
				Name:    input.NodeExecutionIdentifier.ExecutionId.Name,
			},
		},
	}).Preload("ChildNodeExecutions").Take(&nodeExecution)
	timer.Stop()
	if tx.Error != nil {
		return models.NodeExecution{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	if tx.RecordNotFound() {
		return models.NodeExecution{},
			errors.GetMissingEntityError("node execution", &core.NodeExecutionIdentifier{
				NodeId: input.NodeExecutionIdentifier.NodeId,
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: input.NodeExecutionIdentifier.ExecutionId.Project,
					Domain:  input.NodeExecutionIdentifier.ExecutionId.Domain,
					Name:    input.NodeExecutionIdentifier.ExecutionId.Name,
				},
			})
	}
	return nodeExecution, nil
}

// Persist the event that triggers an update in execution. If any of the persistence fails
// rollback the transaction all together.
func (r *NodeExecutionRepo) Update(ctx context.Context, event *models.NodeExecutionEvent, nodeExecution *models.NodeExecution) error {
	timer := r.metrics.UpdateDuration.Start()
	defer timer.Stop()
	// Use a transaction to guarantee no partial updates.
	tx := r.db.Begin()
	if err := tx.Create(&event).Error; err != nil {
		tx.Rollback()
		return r.errorTransformer.ToFlyteAdminError(err)
	}
	if err := r.db.Model(nodeExecution).Updates(nodeExecution).Error; err != nil {
		tx.Rollback()
		return r.errorTransformer.ToFlyteAdminError(err)
	}
	if err := tx.Commit().Error; err != nil {
		return r.errorTransformer.ToFlyteAdminError(err)
	}
	return nil
}

func (r *NodeExecutionRepo) List(ctx context.Context, input interfaces.ListResourceInput) (
	interfaces.NodeExecutionCollectionOutput, error) {
	// First validate input.
	if err := ValidateListInput(input); err != nil {
		return interfaces.NodeExecutionCollectionOutput{}, err
	}
	var nodeExecutions []models.NodeExecution
	tx := r.db.Limit(input.Limit).Offset(input.Offset).Preload("ChildNodeExecutions")
	// And add join condition (joining multiple tables is fine even we only filter on a subset of table attributes).
	// (this query isn't called for deletes).
	tx = tx.Joins(fmt.Sprintf("INNER JOIN %s ON %s.execution_project = %s.execution_project AND "+
		"%s.execution_domain = %s.execution_domain AND %s.execution_name = %s.execution_name",
		executionTableName, nodeExecutionTableName, executionTableName,
		nodeExecutionTableName, executionTableName, nodeExecutionTableName, executionTableName))

	// Apply filters
	tx, err := applyScopedFilters(tx, input.InlineFilters, input.MapFilters)
	if err != nil {
		return interfaces.NodeExecutionCollectionOutput{}, err
	}
	// Apply sort ordering.
	if input.SortParameter != nil {
		tx = tx.Order(input.SortParameter.GetGormOrderExpr())
	}

	timer := r.metrics.ListDuration.Start()
	tx = tx.Find(&nodeExecutions)
	timer.Stop()
	if tx.Error != nil {
		return interfaces.NodeExecutionCollectionOutput{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return interfaces.NodeExecutionCollectionOutput{
		NodeExecutions: nodeExecutions,
	}, nil
}

func (r *NodeExecutionRepo) ListEvents(
	ctx context.Context, input interfaces.ListResourceInput) (interfaces.NodeExecutionEventCollectionOutput, error) {
	// First validate input.
	if err := ValidateListInput(input); err != nil {
		return interfaces.NodeExecutionEventCollectionOutput{}, err
	}
	var nodeExecutionEvents []models.NodeExecutionEvent
	tx := r.db.Limit(input.Limit).Offset(input.Offset)
	// And add join condition (joining multiple tables is fine even we only filter on a subset of table attributes).
	// (this query isn't called for deletes).
	tx = tx.Joins(innerJoinNodeExecToNodeEvents)
	tx = tx.Joins(innerJoinExecToNodeExec)

	// Apply filters
	tx, err := applyScopedFilters(tx, input.InlineFilters, input.MapFilters)
	if err != nil {
		return interfaces.NodeExecutionEventCollectionOutput{}, err
	}
	// Apply sort ordering.
	if input.SortParameter != nil {
		tx = tx.Order(input.SortParameter.GetGormOrderExpr())
	}

	timer := r.metrics.ListDuration.Start()
	tx = tx.Find(&nodeExecutionEvents)
	timer.Stop()
	if tx.Error != nil {
		return interfaces.NodeExecutionEventCollectionOutput{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return interfaces.NodeExecutionEventCollectionOutput{
		NodeExecutionEvents: nodeExecutionEvents,
	}, nil
}

func (r *NodeExecutionRepo) Exists(ctx context.Context, input interfaces.NodeExecutionResource) (bool, error) {
	var nodeExecution models.NodeExecution
	timer := r.metrics.ExistsDuration.Start()
	tx := r.db.Select(ID).Where(&models.NodeExecution{
		NodeExecutionKey: models.NodeExecutionKey{
			NodeID: input.NodeExecutionIdentifier.NodeId,
			ExecutionKey: models.ExecutionKey{
				Project: input.NodeExecutionIdentifier.ExecutionId.Project,
				Domain:  input.NodeExecutionIdentifier.ExecutionId.Domain,
				Name:    input.NodeExecutionIdentifier.ExecutionId.Name,
			},
		},
	}).Take(&nodeExecution)
	timer.Stop()
	if tx.Error != nil {
		return false, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return !tx.RecordNotFound(), nil
}

// Returns an instance of NodeExecutionRepoInterface
func NewNodeExecutionRepo(
	db *gorm.DB, errorTransformer errors.ErrorTransformer,
	scope promutils.Scope) interfaces.NodeExecutionRepoInterface {
	metrics := newMetrics(scope)
	return &NodeExecutionRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}
