package gormimpl

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	adminErrors "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

// Implementation of NodeExecutionInterface.
type NodeExecutionRepo struct {
	db               *gorm.DB
	errorTransformer adminErrors.ErrorTransformer
	metrics          gormMetrics
}

func (r *NodeExecutionRepo) Create(ctx context.Context, execution *models.NodeExecution) error {
	timer := r.metrics.CreateDuration.Start()
	tx := r.db.WithContext(ctx).Omit("id").Create(&execution)
	timer.Stop()
	if tx.Error != nil {
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return nil
}

func (r *NodeExecutionRepo) Get(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
	var nodeExecution models.NodeExecution
	timer := r.metrics.GetDuration.Start()
	tx := r.db.WithContext(ctx).Where(&models.NodeExecution{
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

	if tx.Error != nil && errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return models.NodeExecution{},
			adminErrors.GetMissingEntityError("node execution", &core.NodeExecutionIdentifier{
				NodeId: input.NodeExecutionIdentifier.NodeId,
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: input.NodeExecutionIdentifier.ExecutionId.Project,
					Domain:  input.NodeExecutionIdentifier.ExecutionId.Domain,
					Name:    input.NodeExecutionIdentifier.ExecutionId.Name,
				},
			})
	} else if tx.Error != nil {
		return models.NodeExecution{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}

	return nodeExecution, nil
}

func (r *NodeExecutionRepo) GetWithChildren(ctx context.Context, input interfaces.NodeExecutionResource) (models.NodeExecution, error) {
	var nodeExecution models.NodeExecution
	timer := r.metrics.GetDuration.Start()
	tx := r.db.WithContext(ctx).Where(&models.NodeExecution{
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

	if tx.Error != nil && errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return models.NodeExecution{},
			adminErrors.GetMissingEntityError("node execution", &core.NodeExecutionIdentifier{
				NodeId: input.NodeExecutionIdentifier.NodeId,
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: input.NodeExecutionIdentifier.ExecutionId.Project,
					Domain:  input.NodeExecutionIdentifier.ExecutionId.Domain,
					Name:    input.NodeExecutionIdentifier.ExecutionId.Name,
				},
			})
	} else if tx.Error != nil {
		return models.NodeExecution{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}

	return nodeExecution, nil
}

func (r *NodeExecutionRepo) Update(ctx context.Context, nodeExecution *models.NodeExecution) error {
	timer := r.metrics.UpdateDuration.Start()
	tx := r.db.WithContext(ctx).Model(&nodeExecution).Updates(nodeExecution)
	timer.Stop()
	if err := tx.Error; err != nil {
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
	tx := r.db.WithContext(ctx).Limit(input.Limit).Offset(input.Offset).Preload("ChildNodeExecutions")
	// And add join condition, if any
	if input.JoinTableEntities[common.Execution] {
		tx = tx.Joins(innerJoinExecToNodeExec)
	}

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

func (r *NodeExecutionRepo) Exists(ctx context.Context, input interfaces.NodeExecutionResource) (bool, error) {
	var nodeExecution models.NodeExecution
	timer := r.metrics.ExistsDuration.Start()
	tx := r.db.WithContext(ctx).Select(ID).Where(&models.NodeExecution{
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
	return true, nil
}

func (r *NodeExecutionRepo) Count(ctx context.Context, input interfaces.CountResourceInput) (int64, error) {
	var err error
	tx := r.db.WithContext(ctx).Model(&models.NodeExecution{}).Preload("ChildNodeExecutions")

	// And add join condition, if any
	if input.JoinTableEntities[common.Execution] {
		tx = tx.Joins(innerJoinExecToNodeExec)
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

// Returns an instance of NodeExecutionRepoInterface
func NewNodeExecutionRepo(
	db *gorm.DB, errorTransformer adminErrors.ErrorTransformer,
	scope promutils.Scope) interfaces.NodeExecutionRepoInterface {
	metrics := newMetrics(scope)
	return &NodeExecutionRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}
