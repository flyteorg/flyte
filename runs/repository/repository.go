package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/flyteorg/flyte/v2/flytestdlib/database"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/runs/repository/impl"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
)

// repository implements the Repository interface
type repository struct {
	actionRepo  interfaces.ActionRepo
	taskRepo    interfaces.TaskRepo
	triggerRepo interfaces.TriggerRepo
}

// NewRepository creates a new Repository instance. Each underlying repository
// gets its own metrics sub-scope (action/task/trigger) so the per-operation DB
// metrics they register under "db" do not collide. Pass a nil scope to disable
// metrics.
func NewRepository(db *sqlx.DB, dbConfig database.DbConfig, scope promutils.Scope) (interfaces.Repository, error) {
	actionRepo, err := impl.NewActionRepo(db, dbConfig, subScope(scope, "action"))
	if err != nil {
		return nil, fmt.Errorf("failed to create action repo: %w", err)
	}
	return &repository{
		actionRepo:  actionRepo,
		taskRepo:    impl.NewTaskRepo(db, subScope(scope, "task")),
		triggerRepo: impl.NewTriggerRepo(db, subScope(scope, "trigger")),
	}, nil
}

// subScope returns a named sub-scope, or nil if the parent scope is nil so that
// metrics can be disabled entirely (e.g. in tests).
func subScope(scope promutils.Scope, name string) promutils.Scope {
	if scope == nil {
		return nil
	}
	return scope.NewSubScope(name)
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
