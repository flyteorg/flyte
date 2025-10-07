package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

// PostgresRepository implements Repository interface using PostgreSQL
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *gorm.DB) Repository {
	return &PostgresRepository{db: db}
}

// EnqueueAction persists a new action to the queue
func (r *PostgresRepository) EnqueueAction(ctx context.Context, req *workflow.EnqueueActionRequest) error {
	// Serialize the entire request as JSON
	specBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal action spec: %w", err)
	}

	var actionGroup *string
	if req.Group != "" {
		actionGroup = &req.Group
	}

	var subject *string
	if req.Subject != "" {
		subject = &req.Subject
	}

	action := &QueuedAction{
		Org:              req.ActionId.Run.Org,
		Project:          req.ActionId.Run.Project,
		Domain:           req.ActionId.Run.Domain,
		RunName:          req.ActionId.Run.Name,
		ActionName:       req.ActionId.Name,
		ParentActionName: req.ParentActionName,
		ActionGroup:      actionGroup,
		Subject:          subject,
		ActionSpec:       datatypes.JSON(specBytes),
		InputURI:         req.InputUri,
		RunOutputBase:    req.RunOutputBase,
		Status:           "queued",
		EnqueuedAt:       time.Now(),
	}

	result := r.db.WithContext(ctx).Create(action)
	if result.Error != nil {
		return fmt.Errorf("failed to enqueue action: %w", result.Error)
	}

	logger.Infof(ctx, "Enqueued action: %s/%s/%s/%s/%s",
		action.Org, action.Project, action.Domain, action.RunName, action.ActionName)

	return nil
}

// AbortQueuedRun marks all actions in a run as aborted
func (r *PostgresRepository) AbortQueuedRun(ctx context.Context, runID *common.RunIdentifier, reason string) error {
	result := r.db.WithContext(ctx).
		Model(&QueuedAction{}).
		Where("org = ? AND project = ? AND domain = ? AND run_name = ? AND status IN ?",
			runID.Org, runID.Project, runID.Domain, runID.Name, []string{"queued", "processing"}).
		Updates(map[string]interface{}{
			"status":       "aborted",
			"abort_reason": reason,
			"updated_at":   time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to abort queued run: %w", result.Error)
	}

	logger.Infof(ctx, "Aborted %d actions for run: %s/%s/%s/%s",
		result.RowsAffected, runID.Org, runID.Project, runID.Domain, runID.Name)

	return nil
}

// AbortQueuedAction marks a specific action as aborted
func (r *PostgresRepository) AbortQueuedAction(ctx context.Context, actionID *common.ActionIdentifier, reason string) error {
	result := r.db.WithContext(ctx).
		Model(&QueuedAction{}).
		Where("org = ? AND project = ? AND domain = ? AND run_name = ? AND action_name = ? AND status IN ?",
			actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain,
			actionID.Run.Name, actionID.Name, []string{"queued", "processing"}).
		Updates(map[string]interface{}{
			"status":       "aborted",
			"abort_reason": reason,
			"updated_at":   time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to abort queued action: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("action not found or already processed: %s/%s/%s/%s/%s",
			actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain,
			actionID.Run.Name, actionID.Name)
	}

	logger.Infof(ctx, "Aborted action: %s/%s/%s/%s/%s",
		actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain,
		actionID.Run.Name, actionID.Name)

	return nil
}

// GetQueuedActions retrieves queued actions ready for processing
func (r *PostgresRepository) GetQueuedActions(ctx context.Context, limit int) ([]*QueuedAction, error) {
	var actions []*QueuedAction

	result := r.db.WithContext(ctx).
		Where("status = ?", "queued").
		Order("enqueued_at ASC").
		Limit(limit).
		Find(&actions)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get queued actions: %w", result.Error)
	}

	return actions, nil
}

// MarkAsProcessing marks an action as being processed
func (r *PostgresRepository) MarkAsProcessing(ctx context.Context, actionID *common.ActionIdentifier) error {
	result := r.db.WithContext(ctx).
		Model(&QueuedAction{}).
		Where("org = ? AND project = ? AND domain = ? AND run_name = ? AND action_name = ? AND status = ?",
			actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain,
			actionID.Run.Name, actionID.Name, "queued").
		Updates(map[string]interface{}{
			"status":     "processing",
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to mark action as processing: %w", result.Error)
	}

	return nil
}

// MarkAsCompleted marks an action as completed
func (r *PostgresRepository) MarkAsCompleted(ctx context.Context, actionID *common.ActionIdentifier) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&QueuedAction{}).
		Where("org = ? AND project = ? AND domain = ? AND run_name = ? AND action_name = ?",
			actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain,
			actionID.Run.Name, actionID.Name).
		Updates(map[string]interface{}{
			"status":       "completed",
			"processed_at": now,
			"updated_at":   now,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to mark action as completed: %w", result.Error)
	}

	return nil
}

// MarkAsFailed marks an action as failed with an error message
func (r *PostgresRepository) MarkAsFailed(ctx context.Context, actionID *common.ActionIdentifier, errorMsg string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&QueuedAction{}).
		Where("org = ? AND project = ? AND domain = ? AND run_name = ? AND action_name = ?",
			actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain,
			actionID.Run.Name, actionID.Name).
		Updates(map[string]interface{}{
			"status":        "failed",
			"error_message": errorMsg,
			"processed_at":  now,
			"updated_at":    now,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to mark action as failed: %w", result.Error)
	}

	return nil
}
