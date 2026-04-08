package impl

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

type tasksRepo struct {
	db *sqlx.DB
}

func NewTaskRepo(db *sqlx.DB) interfaces.TaskRepo {
	return &tasksRepo{
		db: db,
	}
}

// CreateTask upserts a task and its associated triggers in one transaction.
// Trigger summary fields (total_triggers, active_triggers, trigger_name,
// trigger_automation_spec) on the task row are recomputed from the current
// set of triggers via refreshTaskTriggerMeta, so they are intentionally NOT
// set by the task upsert itself.
func (r *tasksRepo) CreateTask(ctx context.Context, newTask *models.Task, triggers []*models.Trigger) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now()
	_, err = tx.ExecContext(ctx,
		`INSERT INTO tasks (
			project, domain, name, version,
			environment, function_name, deployed_by,
			trigger_types,
			task_spec, env_description, short_description,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (project, domain, name, version) DO UPDATE SET
			environment       = EXCLUDED.environment,
			function_name     = EXCLUDED.function_name,
			deployed_by       = EXCLUDED.deployed_by,
			task_spec         = EXCLUDED.task_spec,
			env_description   = EXCLUDED.env_description,
			short_description = EXCLUDED.short_description,
			updated_at        = EXCLUDED.updated_at`,
		newTask.Project,
		newTask.Domain,
		newTask.Name,
		newTask.Version,
		newTask.Environment,
		newTask.FunctionName,
		newTask.DeployedBy,
		newTask.TriggerTypes,
		newTask.TaskSpec,
		newTask.EnvDescription,
		newTask.ShortDescription,
		now,
		now,
	)
	if err != nil {
		logger.Errorf(ctx, "failed to upsert task %v: %v", newTask.TaskKey, err)
		return fmt.Errorf("failed to upsert task %v: %w", newTask.TaskKey, err)
	}
	logger.Infof(ctx, "Upserted task: %s/%s/%s version %s",
		newTask.Project, newTask.Domain, newTask.Name, newTask.Version)

	// Upsert each trigger and append a revision snapshot within the same transaction.
	for _, t := range triggers {
		if err := upsertTrigger(ctx, tx, t, 0); err != nil {
			return fmt.Errorf("failed to upsert trigger %q: %w", t.Name, err)
		}
	}

	// Refresh task trigger summary once after all triggers are written.
	if len(triggers) > 0 {
		if err := refreshTaskTriggerMeta(ctx, tx, triggers[0]); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}
	return nil
}

func (r *tasksRepo) GetTask(ctx context.Context, key models.TaskKey) (*models.Task, error) {
	var task models.Task
	err := sqlx.GetContext(ctx, r.db, &task,
		"SELECT * FROM tasks WHERE project = $1 AND domain = $2 AND name = $3 AND version = $4",
		key.Project, key.Domain, key.Name, key.Version)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("task not found: %v", key)
		}
		logger.Errorf(ctx, "failed to get task %v: %v", key, err)
		return nil, fmt.Errorf("failed to get task %v: %w", key, err)
	}

	return &task, nil
}

func (r *tasksRepo) CreateTaskSpec(ctx context.Context, taskSpec *models.TaskSpec) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO task_specs (digest, spec) VALUES ($1, $2) ON CONFLICT (digest) DO NOTHING`,
		taskSpec.Digest, taskSpec.Spec)

	if err != nil {
		logger.Errorf(ctx, "failed to create task spec %v: %v", taskSpec.Digest, err)
		return fmt.Errorf("failed to create task spec %v: %w", taskSpec.Digest, err)
	}

	return nil
}

func (r *tasksRepo) GetTaskSpec(ctx context.Context, digest string) (*models.TaskSpec, error) {
	var taskSpec models.TaskSpec
	err := sqlx.GetContext(ctx, r.db, &taskSpec,
		"SELECT * FROM task_specs WHERE digest = $1", digest)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("task spec not found: %s", digest)
		}
		logger.Errorf(ctx, "failed to get task spec %v: %v", digest, err)
		return nil, fmt.Errorf("failed to get task spec %v: %w", digest, err)
	}

	return &taskSpec, nil
}

func (r *tasksRepo) ListTasks(ctx context.Context, input interfaces.ListResourceInput) (*models.TaskListResult, error) {
	var queryBuilder strings.Builder
	var args []interface{}
	argIdx := 1

	queryBuilder.WriteString("SELECT * FROM tasks")

	// Build WHERE clause from filters
	var whereClauses []string

	if input.Filter != nil {
		expr, err := input.Filter.QueryExpression("")
		if err != nil {
			return nil, fmt.Errorf("failed to build filter: %w", err)
		}
		rewritten, rewrittenArgs := rewritePlaceholders(expr.Query, expr.Args, argIdx)
		whereClauses = append(whereClauses, rewritten)
		args = append(args, rewrittenArgs...)
		argIdx += len(rewrittenArgs)
	}

	if input.ScopeByFilter != nil {
		expr, err := input.ScopeByFilter.QueryExpression("")
		if err != nil {
			return nil, fmt.Errorf("failed to build scope filter: %w", err)
		}
		rewritten, rewrittenArgs := rewritePlaceholders(expr.Query, expr.Args, argIdx)
		whereClauses = append(whereClauses, rewritten)
		args = append(args, rewrittenArgs...)
		argIdx += len(rewrittenArgs)
	}

	if len(whereClauses) > 0 {
		queryBuilder.WriteString(" WHERE ")
		queryBuilder.WriteString(strings.Join(whereClauses, " AND "))
	}

	// Apply sorting
	if len(input.SortParameters) > 0 {
		queryBuilder.WriteString(" ORDER BY ")
		for i, sp := range input.SortParameters {
			if i > 0 {
				queryBuilder.WriteString(", ")
			}
			queryBuilder.WriteString(sp.GetOrderExpr())
		}
	} else {
		queryBuilder.WriteString(" ORDER BY created_at DESC")
	}

	// Apply pagination
	queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1))
	args = append(args, input.Limit, input.Offset)
	argIdx += 2

	var tasks []*models.Task
	if err := sqlx.SelectContext(ctx, r.db, &tasks, queryBuilder.String(), args...); err != nil {
		logger.Errorf(ctx, "failed to list tasks: %v", err)
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	// Get total counts
	var filteredTotal int64
	var total int64

	// Filtered count query
	{
		var countBuilder strings.Builder
		var countArgs []interface{}
		countIdx := 1

		countBuilder.WriteString("SELECT COUNT(*) FROM tasks")

		var countWhere []string
		if input.Filter != nil {
			expr, err := input.Filter.QueryExpression("")
			if err != nil {
				return nil, fmt.Errorf("failed to build filter expression: %w", err)
			}
			rewritten, rewrittenArgs := rewritePlaceholders(expr.Query, expr.Args, countIdx)
			countWhere = append(countWhere, rewritten)
			countArgs = append(countArgs, rewrittenArgs...)
			countIdx += len(rewrittenArgs)
		}
		if input.ScopeByFilter != nil {
			expr, err := input.ScopeByFilter.QueryExpression("")
			if err != nil {
				return nil, fmt.Errorf("failed to build scope filter expression: %w", err)
			}
			rewritten, rewrittenArgs := rewritePlaceholders(expr.Query, expr.Args, countIdx)
			countWhere = append(countWhere, rewritten)
			countArgs = append(countArgs, rewrittenArgs...)
			countIdx += len(rewrittenArgs)
		}
		if len(countWhere) > 0 {
			countBuilder.WriteString(" WHERE ")
			countBuilder.WriteString(strings.Join(countWhere, " AND "))
		}

		if err := r.db.QueryRowContext(ctx, countBuilder.String(), countArgs...).Scan(&filteredTotal); err != nil {
			logger.Errorf(ctx, "failed to count filtered tasks: %v", err)
		}
	}

	// Total count query (only scoped, no filter)
	{
		var totalBuilder strings.Builder
		var totalArgs []interface{}
		totalIdx := 1

		totalBuilder.WriteString("SELECT COUNT(*) FROM tasks")

		if input.ScopeByFilter != nil {
			expr, err := input.ScopeByFilter.QueryExpression("")
			if err != nil {
				return nil, fmt.Errorf("failed to build scope filter expression: %w", err)
			}
			rewritten, rewrittenArgs := rewritePlaceholders(expr.Query, expr.Args, totalIdx)
			totalBuilder.WriteString(" WHERE ")
			totalBuilder.WriteString(rewritten)
			totalArgs = append(totalArgs, rewrittenArgs...)
		}

		if err := r.db.QueryRowContext(ctx, totalBuilder.String(), totalArgs...).Scan(&total); err != nil {
			logger.Errorf(ctx, "failed to count total tasks: %v", err)
		}
	}

	return &models.TaskListResult{
		Tasks:         tasks,
		FilteredTotal: uint32(filteredTotal),
		Total:         uint32(total),
	}, nil
}

func (r *tasksRepo) ListVersions(ctx context.Context, input interfaces.ListResourceInput) ([]*models.TaskVersion, error) {
	var queryBuilder strings.Builder
	var args []interface{}
	argIdx := 1

	queryBuilder.WriteString("SELECT version, created_at FROM tasks")

	// Apply filters
	if input.Filter != nil {
		expr, err := input.Filter.QueryExpression("")
		if err != nil {
			return nil, fmt.Errorf("failed to build filter: %w", err)
		}
		rewritten, rewrittenArgs := rewritePlaceholders(expr.Query, expr.Args, argIdx)
		queryBuilder.WriteString(" WHERE ")
		queryBuilder.WriteString(rewritten)
		args = append(args, rewrittenArgs...)
		argIdx += len(rewrittenArgs)
	}

	// Apply sorting
	if len(input.SortParameters) > 0 {
		queryBuilder.WriteString(" ORDER BY ")
		for i, sp := range input.SortParameters {
			if i > 0 {
				queryBuilder.WriteString(", ")
			}
			queryBuilder.WriteString(sp.GetOrderExpr())
		}
	} else {
		queryBuilder.WriteString(" ORDER BY created_at DESC")
	}

	// Apply pagination
	queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1))
	args = append(args, input.Limit, input.Offset)

	var versions []*models.TaskVersion
	if err := sqlx.SelectContext(ctx, r.db, &versions, queryBuilder.String(), args...); err != nil {
		logger.Errorf(ctx, "failed to list versions: %v", err)
		return nil, fmt.Errorf("failed to list versions: %w", err)
	}

	return versions, nil
}
