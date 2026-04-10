package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/flyteorg/flyte/v2/flytestdlib/database"
	"github.com/flyteorg/flyte/v2/runs/repository/impl"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
)

// repository implements the Repository interface
type repository struct {
	actionRepo  interfaces.ActionRepo
	taskRepo    interfaces.TaskRepo
	triggerRepo interfaces.TriggerRepo
}

// NewRepository creates a new Repository instance
func NewRepository(db *sqlx.DB, dbConfig database.DbConfig) (interfaces.Repository, error) {
	actionRepo, err := impl.NewActionRepo(db, dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create action repo: %w", err)
	}
	return &repository{
		actionRepo:  actionRepo,
		taskRepo:    impl.NewTaskRepo(db),
		triggerRepo: impl.NewTriggerRepo(db),
	}, nil
}

// ActionRepo returns the action repository
func (r *repository) ActionRepo() interfaces.ActionRepo {
	return r.actionRepo
}

// TaskRepo returns the task repository
func (r *repository) TaskRepo() interfaces.TaskRepo {
	return r.taskRepo
}

// TriggerRepo returns the trigger repository
func (r *repository) TriggerRepo() interfaces.TriggerRepo {
	return r.triggerRepo
}
