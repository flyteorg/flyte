package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

type TaskRepo interface {
	// TODO: add triggers back
	CreateTask(ctx context.Context, task *models.Task) error
	GetTask(ctx context.Context, key models.TaskKey) (*models.Task, error)
	ListTasks(ctx context.Context, input ListResourceInput) (*models.TaskListResult, error)
	ListVersions(ctx context.Context, input ListResourceInput) ([]*models.TaskVersion, error)

	CreateTaskSpec(ctx context.Context, taskSpec *models.TaskSpec) error
	GetTaskSpec(ctx context.Context, digest string) (*models.TaskSpec, error)
}
