package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

type TaskRepo interface {
	// CreateTask upserts a task and its associated triggers in one transaction.
	// Trigger models are optional; pass nil or an empty slice if there are no triggers.
	CreateTask(ctx context.Context, task *models.Task, triggers []*models.Trigger) error
	GetTask(ctx context.Context, key models.TaskKey) (*models.Task, error)
	ListTasks(ctx context.Context, input ListResourceInput) (*models.TaskListResult, error)
	ListVersions(ctx context.Context, input ListResourceInput) ([]*models.TaskVersion, error)

	CreateTaskSpec(ctx context.Context, taskSpec *models.TaskSpec) error
	GetTaskSpec(ctx context.Context, digest string) (*models.TaskSpec, error)
}
