package repository

import (
	"gorm.io/gorm"

	"github.com/flyteorg/flyte/v2/runs/repository/impl"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
)

// repository implements the Repository interface
type repository struct {
	actionRepo interfaces.ActionRepo
	taskRepo   interfaces.TaskRepo
}

// NewRepository creates a new Repository instance
func NewRepository(db *gorm.DB) interfaces.Repository {
	return &repository{
		actionRepo: impl.NewActionRepo(db),
		taskRepo:   impl.NewTaskRepo(db),
	}
}

// ActionRepo returns the action repository
func (r *repository) ActionRepo() interfaces.ActionRepo {
	return r.actionRepo
}

// TaskRepo returns the task repository
func (r *repository) TaskRepo() interfaces.TaskRepo {
	return r.taskRepo
}
