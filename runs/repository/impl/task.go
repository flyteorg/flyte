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

	// Prune triggers that are no longer attached to this task.
	if len(triggers) > 0 {
		newTriggerNames := make(map[string]struct{}, len(triggers))
		for _, t := range triggers {
			newTriggerNames[t.Name] = struct{}{}
		}
		oldTriggers, err := listTaskTriggers(ctx, tx, triggers[0].Project, triggers[0].Domain, triggers[0].TaskName)
		if err != nil {
			return fmt.Errorf("failed to list existing task triggers: %w", err)
		}
		for _, oldName := range oldTriggers {
			if _, keep := newTriggerNames[oldName]; !keep {
				if err := softDeleteTrigger(ctx, tx, triggers[0].Project, triggers[0].Domain, triggers[0].TaskName, oldName); err != nil {
					return fmt.Errorf("failed to delete stale trigger %q: %w", oldName, err)
				}
			}
		}
	}

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

// ListTasks returns the latest version of each unique (project, domain, name) task.
// A CTE with ROW_NUMBER() partitioned by (project, domain, name) and ordered by
// created_at DESC is used so that only rn=1 rows (the latest version) are returned.
func (r *tasksRepo) ListTasks(ctx context.Context, input interfaces.ListResourceInput) (*models.TaskListResult, error) {
	// allArgs accumulates all bind parameters in the order they appear in the query.
	var allArgs []interface{}
	argIdx := 1

	// filtered_tasks WHERE: filter + scope (used for paginated result and filtered_total).
	var filteredWhereClauses []string
	if input.Filter != nil {
		expr, err := input.Filter.QueryExpression("")
		if err != nil {
			return nil, fmt.Errorf("failed to build filter: %w", err)
		}
		rewritten, rewrittenArgs := rewritePlaceholders(expr.Query, expr.Args, argIdx)
		filteredWhereClauses = append(filteredWhereClauses, rewritten)
		allArgs = append(allArgs, rewrittenArgs...)
		argIdx += len(rewrittenArgs)
	}
	if input.ScopeByFilter != nil {
		expr, err := input.ScopeByFilter.QueryExpression("")
		if err != nil {
			return nil, fmt.Errorf("failed to build scope filter: %w", err)
		}
		rewritten, rewrittenArgs := rewritePlaceholders(expr.Query, expr.Args, argIdx)
		filteredWhereClauses = append(filteredWhereClauses, rewritten)
		allArgs = append(allArgs, rewrittenArgs...)
		argIdx += len(rewrittenArgs)
	}
	filteredWhere := ""
	if len(filteredWhereClauses) > 0 {
		filteredWhere = "WHERE " + strings.Join(filteredWhereClauses, " AND ")
	}

	// unfiltered_tasks WHERE: scope only (for the total count).
	unfilteredWhere := ""
	if input.ScopeByFilter != nil {
		expr, err := input.ScopeByFilter.QueryExpression("")
		if err != nil {
			return nil, fmt.Errorf("failed to build scope filter: %w", err)
		}
		rewritten, rewrittenArgs := rewritePlaceholders(expr.Query, expr.Args, argIdx)
		unfilteredWhere = "WHERE " + rewritten
		allArgs = append(allArgs, rewrittenArgs...)
		argIdx += len(rewrittenArgs)
	}

	// Build ORDER BY for the outer query.
	orderBy := "created_at DESC"
	if len(input.SortParameters) > 0 {
		parts := make([]string, 0, len(input.SortParameters))
		for _, sp := range input.SortParameters {
			parts = append(parts, sp.GetOrderExpr())
		}
		orderBy = strings.Join(parts, ", ")
	}

	allArgs = append(allArgs, input.Limit, input.Offset)

	// CTE: rank all rows within each (project, domain, name) by created_at DESC,
	// then select only rn=1 (latest version per task name).
	cte := fmt.Sprintf(`
WITH filtered_tasks AS (
	SELECT *, ROW_NUMBER() OVER (PARTITION BY project, domain, name ORDER BY created_at DESC) AS rn
	FROM tasks %s
),
unfiltered_tasks AS (
	SELECT *, ROW_NUMBER() OVER (PARTITION BY project, domain, name ORDER BY created_at DESC) AS rn
	FROM tasks %s
),
total_counts AS (
	SELECT
		(SELECT COUNT(*) FROM filtered_tasks   WHERE rn = 1) AS filtered_total,
		(SELECT COUNT(*) FROM unfiltered_tasks WHERE rn = 1) AS total
)
SELECT filtered_tasks.*, total_counts.filtered_total, total_counts.total
FROM filtered_tasks, total_counts
WHERE filtered_tasks.rn = 1
ORDER BY %s
LIMIT $%d OFFSET $%d`, filteredWhere, unfilteredWhere, orderBy, argIdx, argIdx+1)

	type taskWithCounts struct {
		models.Task
		Rn            uint32 `db:"rn"`
		FilteredTotal uint32 `db:"filtered_total"`
		Total         uint32 `db:"total"`
	}
	var rows []taskWithCounts
	if err := sqlx.SelectContext(ctx, r.db, &rows, cte, allArgs...); err != nil {
		logger.Errorf(ctx, "failed to list tasks: %v", err)
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	tasks := make([]*models.Task, 0, len(rows))
	var filteredTotal, total uint32
	for i, row := range rows {
		t := row.Task
		tasks = append(tasks, &t)
		if i == 0 {
			filteredTotal = row.FilteredTotal
			total = row.Total
		}
	}

	return &models.TaskListResult{
		Tasks:         tasks,
		FilteredTotal: filteredTotal,
		Total:         total,
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
