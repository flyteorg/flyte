package impl

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	triggerpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/trigger"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

type triggerRepo struct {
	db *gorm.DB
}

func NewTriggerRepo(db *gorm.DB) interfaces.TriggerRepo {
	return &triggerRepo{db: db}
}

func (r *triggerRepo) SaveTrigger(ctx context.Context, trigger *models.Trigger, expectedRevision uint64) (*models.Trigger, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := upsertTrigger(ctx, tx, trigger, expectedRevision); err != nil {
			return err
		}
		return refreshTaskTriggerMeta(ctx, tx, trigger)
	})
	if err != nil {
		logger.Errorf(ctx, "SaveTrigger failed for %s/%s/%s/%s: %v",
			trigger.Project, trigger.Domain, trigger.TaskName, trigger.Name, err)
		return nil, err
	}
	return trigger, nil
}

// upsertTrigger upserts a trigger row and appends an immutable revision snapshot.
// Pass expectedRevision=0 to skip optimistic locking (e.g. when called from CreateTask).
func upsertTrigger(ctx context.Context, tx *gorm.DB, trigger *models.Trigger, expectedRevision uint64) error {
	now := time.Now()

	// ON CONFLICT: bump revision, update mutable fields.
	// WHERE clause enforces optimistic locking when expectedRevision > 0.
	//
	// The WHERE on the UPDATE half of ON CONFLICT is a PostgreSQL extension:
	// if the condition is false (revision mismatch), the upsert returns 0 rows.
	suffix := `ON CONFLICT (project, domain, task_name, name) DO UPDATE SET
		latest_revision  = triggers.latest_revision + 1,
		deployed_at      = EXCLUDED.deployed_at,
		deployed_by      = EXCLUDED.deployed_by,
		updated_at       = CASE WHEN triggers.active != EXCLUDED.active THEN EXCLUDED.updated_at ELSE triggers.updated_at END,
		updated_by       = CASE WHEN triggers.active != EXCLUDED.active THEN EXCLUDED.updated_by ELSE triggers.updated_by END,
		deleted_at       = NULL,
		active           = EXCLUDED.active,
		task_version     = EXCLUDED.task_version,
		automation_type  = EXCLUDED.automation_type,
		automation_spec  = EXCLUDED.automation_spec,
		spec             = EXCLUDED.spec,
		description      = EXCLUDED.description`

	if expectedRevision > 0 {
		suffix += fmt.Sprintf(" WHERE triggers.latest_revision = %d", expectedRevision)
	}
	suffix += " RETURNING *"

	result := tx.Raw(`
		INSERT INTO triggers (
			project, domain, task_name, name,
			latest_revision,
			spec, automation_spec,
			task_version, active, automation_type,
			deployed_by, updated_by,
			deployed_at, updated_at,
			description
		) VALUES (?, ?, ?, ?, 1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) `+suffix,
		trigger.Project, trigger.Domain, trigger.TaskName, trigger.Name,
		trigger.Spec, trigger.AutomationSpec,
		trigger.TaskVersion, trigger.Active, trigger.AutomationType,
		trigger.DeployedBy, trigger.UpdatedBy,
		now, now,
		trigger.Description,
	).Scan(trigger)

	if result.Error != nil {
		return fmt.Errorf("failed to upsert trigger: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("optimistic lock failure: trigger was modified concurrently, please fetch latest and retry")
	}

	return insertTriggerRevision(ctx, tx, trigger, triggerpb.TriggerRevisionAction_TRIGGER_REVISION_ACTION_DEPLOY)
}

func (r *triggerRepo) GetTrigger(ctx context.Context, key interfaces.TriggerNameKey) (*models.Trigger, error) {
	var t models.Trigger
	q := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if key.TaskName != "" {
		q = q.Where("project = ? AND domain = ? AND task_name = ? AND name = ?",
			key.Project, key.Domain, key.TaskName, key.Name)
	} else {
		q = q.Where("project = ? AND domain = ? AND name = ?",
			key.Project, key.Domain, key.Name)
	}
	result := q.First(&t)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("trigger not found: %s/%s/%s/%s", key.Project, key.Domain, key.TaskName, key.Name)
		}
		return nil, fmt.Errorf("failed to get trigger: %w", result.Error)
	}
	return &t, nil
}

func (r *triggerRepo) GetTriggerRevision(ctx context.Context, project, domain, taskName, name string, revision uint64) (*models.TriggerRevision, error) {
	var rev models.TriggerRevision
	result := r.db.WithContext(ctx).
		Where("project = ? AND domain = ? AND task_name = ? AND name = ? AND revision = ?",
			project, domain, taskName, name, revision).
		First(&rev)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("trigger revision not found: %s/%s/%s/%s@%d", project, domain, taskName, name, revision)
		}
		return nil, fmt.Errorf("failed to get trigger revision: %w", result.Error)
	}
	return &rev, nil
}

func (r *triggerRepo) ListTriggers(ctx context.Context, input interfaces.ListResourceInput) ([]*models.Trigger, error) {
	var triggers []*models.Trigger
	query := r.db.WithContext(ctx).Model(&models.Trigger{}).Where("deleted_at IS NULL")

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
	if len(input.SortParameters) > 0 {
		for _, sp := range input.SortParameters {
			query = query.Order(sp.GetGormOrderExpr())
		}
	} else {
		query = query.Order("active DESC, updated_at DESC")
	}

	if err := query.Limit(input.Limit).Offset(input.Offset).Find(&triggers).Error; err != nil {
		return nil, fmt.Errorf("failed to list triggers: %w", err)
	}
	return triggers, nil
}

func (r *triggerRepo) ListTriggerRevisions(ctx context.Context, project, domain, taskName, name string, input interfaces.ListResourceInput) ([]*models.TriggerRevision, error) {
	var revisions []*models.TriggerRevision
	query := r.db.WithContext(ctx).Model(&models.TriggerRevision{}).
		Where("project = ? AND domain = ? AND task_name = ? AND name = ?",
			project, domain, taskName, name).
		Order("revision DESC").
		Limit(input.Limit).
		Offset(input.Offset)

	if err := query.Find(&revisions).Error; err != nil {
		return nil, fmt.Errorf("failed to list trigger revisions: %w", err)
	}
	return revisions, nil
}

func (r *triggerRepo) UpdateTriggers(ctx context.Context, keys []interfaces.TriggerNameKey, active bool) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, key := range keys {
			var t models.Trigger
			result := tx.Raw(`
				UPDATE triggers SET
					active          = ?,
					updated_at      = NOW(),
					latest_revision = latest_revision + 1
				WHERE project = ? AND domain = ? AND task_name = ? AND name = ?
				  AND deleted_at IS NULL
				  AND active != ?
				RETURNING *`,
				active,
				key.Project, key.Domain, key.TaskName, key.Name,
				active,
			).Scan(&t)

			if result.Error != nil {
				return fmt.Errorf("failed to update trigger %+v: %w", key, result.Error)
			}
			if result.RowsAffected == 0 {
				continue // already in desired state or not found
			}

			action := triggerpb.TriggerRevisionAction_TRIGGER_REVISION_ACTION_DEACTIVATE
			if active {
				action = triggerpb.TriggerRevisionAction_TRIGGER_REVISION_ACTION_ACTIVATE
			}
			if err := insertTriggerRevision(ctx, tx, &t, action); err != nil {
				return err
			}
			if err := refreshTaskTriggerMeta(ctx, tx, &t); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *triggerRepo) DeleteTriggers(ctx context.Context, keys []interfaces.TriggerNameKey) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, key := range keys {
			var t models.Trigger
			result := tx.Raw(`
				UPDATE triggers SET
					deleted_at      = NOW(),
					active          = FALSE,
					latest_revision = latest_revision + 1
				WHERE project = ? AND domain = ? AND task_name = ? AND name = ?
				  AND deleted_at IS NULL
				RETURNING *`,
				key.Project, key.Domain, key.TaskName, key.Name,
			).Scan(&t)

			if result.Error != nil {
				return fmt.Errorf("failed to delete trigger %+v: %w", key, result.Error)
			}
			if result.RowsAffected == 0 {
				continue // already deleted or not found
			}

			if err := insertTriggerRevision(ctx, tx, &t, triggerpb.TriggerRevisionAction_TRIGGER_REVISION_ACTION_DELETE); err != nil {
				return err
			}
			if err := refreshTaskTriggerMeta(ctx, tx, &t); err != nil {
				return err
			}
		}
		return nil
	})
}


// insertTriggerRevision appends an immutable snapshot to trigger_revisions.
func insertTriggerRevision(ctx context.Context, tx *gorm.DB, t *models.Trigger, action triggerpb.TriggerRevisionAction) error {
	rev := &models.TriggerRevision{
		Project:        t.Project,
		Domain:         t.Domain,
		TaskName:       t.TaskName,
		Name:           t.Name,
		Revision:       t.LatestRevision,
		Spec:           t.Spec,
		AutomationSpec: t.AutomationSpec,
		TaskVersion:    t.TaskVersion,
		Active:         t.Active,
		AutomationType: t.AutomationType,
		DeployedBy:     t.DeployedBy,
		UpdatedBy:      t.UpdatedBy,
		DeployedAt:     t.DeployedAt,
		UpdatedAt:      t.UpdatedAt,
		TriggeredAt:    t.TriggeredAt,
		DeletedAt:      t.DeletedAt,
		Action:         action.String(),
	}
	if err := tx.WithContext(ctx).Create(rev).Error; err != nil {
		return fmt.Errorf("failed to insert trigger revision: %w", err)
	}
	return nil
}

// refreshTaskTriggerMeta recomputes and updates denormalized trigger summary fields on the tasks row.
// Package-level so it can be reused by tasksRepo within the same transaction.
func refreshTaskTriggerMeta(ctx context.Context, tx *gorm.DB, t *models.Trigger) error {
	type statRow struct {
		Active         bool           `gorm:"column:active"`
		Count          int            `gorm:"column:count"`
		TriggerName    sql.NullString `gorm:"column:trigger_name"`
		AutomationSpec []byte         `gorm:"column:automation_spec"`
	}

	var rows []statRow
	err := tx.WithContext(ctx).Raw(`
		SELECT active, COUNT(*) AS count,
		       STRING_AGG(name, ',') AS trigger_name,
		       STRING_AGG(encode(automation_spec, 'escape'), ',') AS automation_spec
		FROM triggers
		WHERE project = ? AND domain = ? AND task_name = ? AND task_version = ?
		  AND deleted_at IS NULL
		GROUP BY active`,
		t.Project, t.Domain, t.TaskName, t.TaskVersion,
	).Scan(&rows).Error
	if err != nil {
		return fmt.Errorf("failed to compute task trigger stats: %w", err)
	}

	var totalTriggers, activeTriggers int
	var triggerName sql.NullString
	var automationSpec []byte

	for _, row := range rows {
		totalTriggers += row.Count
		if row.Active {
			activeTriggers += row.Count
		}
		if row.TriggerName.Valid {
			triggerName = row.TriggerName
		}
		if row.AutomationSpec != nil {
			automationSpec = row.AutomationSpec
		}
	}
	// Surface name/spec only when exactly one trigger is attached
	if totalTriggers != 1 {
		automationSpec = nil
		triggerName = sql.NullString{}
	}

	result := tx.WithContext(ctx).Exec(`
		UPDATE tasks SET
			trigger_name            = ?,
			total_triggers          = ?,
			active_triggers         = ?,
			trigger_automation_spec = ?
		WHERE project = ? AND domain = ? AND name = ? AND version = ?`,
		triggerName, totalTriggers, activeTriggers, automationSpec,
		t.Project, t.Domain, t.TaskName, t.TaskVersion,
	)
	if result.Error != nil {
		return fmt.Errorf("failed to update task trigger meta: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		logger.Warnf(ctx, "task %s/%s/%s@%s not found when refreshing trigger meta",
			t.Project, t.Domain, t.TaskName, t.TaskVersion)
	}
	return nil
}
