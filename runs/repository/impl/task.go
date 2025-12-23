package impl

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

type tasksRepo struct {
	db *gorm.DB
}

func NewTaskRepo(db *gorm.DB) interfaces.TaskRepo {
	return &tasksRepo{
		db: db,
	}
}

// TODO(nary): add triggers back
func (r *tasksRepo) CreateTask(ctx context.Context, newTask *models.Task) error {
	// Use GORM's Create or Updates based on conflict
	// ON CONFLICT (org, project, domain, name, version) DO UPDATE
	result := r.db.WithContext(ctx).
		Exec(`INSERT INTO tasks (
			org, project, domain, name, version,
			environment, function_name, deployed_by,
			trigger_name, total_triggers, active_triggers,
			trigger_automation_spec, trigger_types,
			task_spec, env_description, short_description
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (org, project, domain, name, version) DO UPDATE SET
			environment = EXCLUDED.environment,
			function_name = EXCLUDED.function_name,
			deployed_by = EXCLUDED.deployed_by,
			trigger_name = EXCLUDED.trigger_name,
			total_triggers = EXCLUDED.total_triggers,
			active_triggers = EXCLUDED.active_triggers,
			trigger_automation_spec = EXCLUDED.trigger_automation_spec,
			trigger_types = EXCLUDED.trigger_types,
			task_spec = EXCLUDED.task_spec,
			env_description = EXCLUDED.env_description,
			short_description = EXCLUDED.short_description`,
			newTask.Org,
			newTask.Project,
			newTask.Domain,
			newTask.Name,
			newTask.Version,
			newTask.Environment,
			newTask.FunctionName,
			newTask.DeployedBy,
			newTask.TriggerName,
			newTask.TotalTriggers,
			newTask.ActiveTriggers,
			newTask.TriggerAutomationSpec,
			newTask.TriggerTypes,
			newTask.TaskSpec,
			newTask.EnvDescription,
			newTask.ShortDescription,
		)

	if result.Error != nil {
		logger.Errorf(ctx, "failed to create task %v: %v", newTask.TaskKey, result.Error)
		return fmt.Errorf("failed to create task %v: %w", newTask.TaskKey, result.Error)
	}

	logger.Infof(ctx, "Created/Updated task: %s/%s/%s/%s version %s",
		newTask.Org, newTask.Project, newTask.Domain, newTask.Name, newTask.Version)

	return nil
}

func (r *tasksRepo) GetTask(ctx context.Context, key models.TaskKey) (*models.Task, error) {
	var task models.Task
	result := r.db.WithContext(ctx).
		Where("org = ? AND project = ? AND domain = ? AND name = ? AND version = ?",
			key.Org, key.Project, key.Domain, key.Name, key.Version).
		First(&task)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("task not found: %v", key)
		}
		logger.Errorf(ctx, "failed to get task %v: %v", key, result.Error)
		return nil, fmt.Errorf("failed to get task %v: %w", key, result.Error)
	}

	return &task, nil
}

func (r *tasksRepo) CreateTaskSpec(ctx context.Context, taskSpec *models.TaskSpec) error {
	// Insert task spec (ignore conflicts since specs are immutable by digest)
	result := r.db.WithContext(ctx).
		Exec(`INSERT INTO task_specs (digest, spec) VALUES (?, ?) ON CONFLICT (digest) DO NOTHING`,
			taskSpec.Digest, taskSpec.Spec)

	if result.Error != nil {
		logger.Errorf(ctx, "failed to create task spec %v: %v", taskSpec.Digest, result.Error)
		return fmt.Errorf("failed to create task spec %v: %w", taskSpec.Digest, result.Error)
	}

	return nil
}

func (r *tasksRepo) GetTaskSpec(ctx context.Context, digest string) (*models.TaskSpec, error) {
	var taskSpec models.TaskSpec
	result := r.db.WithContext(ctx).
		Where("digest = ?", digest).
		First(&taskSpec)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("task spec not found: %s", digest)
		}
		logger.Errorf(ctx, "failed to get task spec %v: %v", digest, result.Error)
		return nil, fmt.Errorf("failed to get task spec %v: %w", digest, result.Error)
	}

	return &taskSpec, nil
}

func (r *tasksRepo) ListTasks(ctx context.Context, input interfaces.ListResourceInput) (*models.TaskListResult, error) {
	// TODO(nary): This is a simplified version. The original had complex CTE queries for:
	// - Selecting latest version per task name using ROW_NUMBER() window function
	// - Filtered and unfiltered counts
	// - Custom filter expressions
	// Need to reimplement with GORM raw queries or refactor the logic

	var tasks []*models.Task
	query := r.db.WithContext(ctx).Model(&models.Task{})

	// Apply filters if provided
	if input.Filter != nil {
		expr, err := input.Filter.GormQueryExpression("")
		if err != nil {
			return nil, fmt.Errorf("failed to build filter: %w", err)
		}
		query = query.Where(expr.Query, expr.Args...)
	}

	if input.ScopeByFilter != nil {
		expr, err := input.ScopeByFilter.GormQueryExpression("")
		if err != nil {
			return nil, fmt.Errorf("failed to build scope filter: %w", err)
		}
		query = query.Where(expr.Query, expr.Args...)
	}

	// Apply sorting
	if len(input.SortParameters) > 0 {
		for _, sp := range input.SortParameters {
			query = query.Order(sp.GetGormOrderExpr())
		}
	} else {
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	query = query.Limit(input.Limit).Offset(input.Offset)

	result := query.Find(&tasks)
	if result.Error != nil {
		logger.Errorf(ctx, "failed to list tasks: %v", result.Error)
		return nil, fmt.Errorf("failed to list tasks: %w", result.Error)
	}

	// Get total counts
	var filteredTotal int64
	var total int64

	countQuery := r.db.WithContext(ctx).Model(&models.Task{})
	if input.Filter != nil {
		expr, _ := input.Filter.GormQueryExpression("")
		countQuery = countQuery.Where(expr.Query, expr.Args...)
	}
	if input.ScopeByFilter != nil {
		expr, _ := input.ScopeByFilter.GormQueryExpression("")
		countQuery = countQuery.Where(expr.Query, expr.Args...)
	}
	countQuery.Count(&filteredTotal)

	totalQuery := r.db.WithContext(ctx).Model(&models.Task{})
	if input.ScopeByFilter != nil {
		expr, _ := input.ScopeByFilter.GormQueryExpression("")
		totalQuery = totalQuery.Where(expr.Query, expr.Args...)
	}
	totalQuery.Count(&total)

	return &models.TaskListResult{
		Tasks:         tasks,
		FilteredTotal: uint32(filteredTotal),
		Total:         uint32(total),
	}, nil
}

func (r *tasksRepo) ListVersions(ctx context.Context, input interfaces.ListResourceInput) ([]*models.TaskVersion, error) {
	var versions []*models.TaskVersion
	query := r.db.WithContext(ctx).
		Model(&models.Task{}).
		Select("version, created_at")

	// Apply filters
	if input.Filter != nil {
		expr, err := input.Filter.GormQueryExpression("")
		if err != nil {
			return nil, fmt.Errorf("failed to build filter: %w", err)
		}
		query = query.Where(expr.Query, expr.Args...)
	}

	// Apply sorting
	if len(input.SortParameters) > 0 {
		for _, sp := range input.SortParameters {
			query = query.Order(sp.GetGormOrderExpr())
		}
	} else {
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	query = query.Limit(input.Limit).Offset(input.Offset)

	result := query.Find(&versions)
	if result.Error != nil {
		logger.Errorf(ctx, "failed to list versions: %v", result.Error)
		return nil, fmt.Errorf("failed to list versions: %w", result.Error)
	}

	return versions, nil
}
