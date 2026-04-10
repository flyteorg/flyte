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
	triggerpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/trigger"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

type triggerRepo struct {
	db *sqlx.DB
}

func NewTriggerRepo(db *sqlx.DB) interfaces.TriggerRepo {
	return &triggerRepo{db: db}
}

func (r *triggerRepo) SaveTrigger(ctx context.Context, trigger *models.Trigger, expectedRevision uint64) (*models.Trigger, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := upsertTrigger(ctx, tx, trigger, expectedRevision); err != nil {
		logger.Errorf(ctx, "SaveTrigger failed for %s/%s/%s/%s: %v",
			trigger.Project, trigger.Domain, trigger.TaskName, trigger.Name, err)
		return nil, err
	}
	if err := refreshTaskTriggerMeta(ctx, tx, trigger); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit tx: %w", err)
	}
	return trigger, nil
}

// upsertTrigger upserts a trigger row and appends an immutable revision snapshot.
// Pass expectedRevision=0 to skip optimistic locking (e.g. when called from CreateTask).
func upsertTrigger(ctx context.Context, tx *sqlx.Tx, trigger *models.Trigger, expectedRevision uint64) error {
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

	query := `
		INSERT INTO triggers (
			project, domain, task_name, name,
			latest_revision,
			spec, automation_spec,
			task_version, active, automation_type,
			deployed_by, updated_by,
			deployed_at, updated_at,
			description
		) VALUES ($1, $2, $3, $4, 1, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) ` + suffix

	err := tx.QueryRowxContext(ctx, query,
		trigger.Project, trigger.Domain, trigger.TaskName, trigger.Name,
		trigger.Spec, trigger.AutomationSpec,
		trigger.TaskVersion, trigger.Active, trigger.AutomationType,
		trigger.DeployedBy, trigger.UpdatedBy,
		now, now,
		trigger.Description,
	).StructScan(trigger)

	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("optimistic lock failure: trigger was modified concurrently, please fetch latest and retry")
	}
	if err != nil {
		return fmt.Errorf("failed to upsert trigger: %w", err)
	}

	return insertTriggerRevision(ctx, tx, trigger, triggerpb.TriggerRevisionAction_TRIGGER_REVISION_ACTION_DEPLOY)
}

func (r *triggerRepo) GetTrigger(ctx context.Context, key interfaces.TriggerNameKey) (*models.Trigger, error) {
	var t models.Trigger
	var err error
	if key.TaskName != "" {
		err = sqlx.GetContext(ctx, r.db, &t,
			`SELECT * FROM triggers WHERE deleted_at IS NULL
			   AND project = $1 AND domain = $2 AND task_name = $3 AND name = $4`,
			key.Project, key.Domain, key.TaskName, key.Name)
	} else {
		err = sqlx.GetContext(ctx, r.db, &t,
			`SELECT * FROM triggers WHERE deleted_at IS NULL
			   AND project = $1 AND domain = $2 AND name = $3`,
			key.Project, key.Domain, key.Name)
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("trigger not found: %s/%s/%s/%s: %w", key.Project, key.Domain, key.TaskName, key.Name, err)
		}
		return nil, fmt.Errorf("failed to get trigger: %w", err)
	}
	return &t, nil
}

func (r *triggerRepo) GetTriggerRevision(ctx context.Context, project, domain, taskName, name string, revision uint64) (*models.TriggerRevision, error) {
	var rev models.TriggerRevision
	err := sqlx.GetContext(ctx, r.db, &rev,
		`SELECT * FROM trigger_revisions
		 WHERE project = $1 AND domain = $2 AND task_name = $3 AND name = $4 AND revision = $5`,
		project, domain, taskName, name, revision)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("trigger revision not found: %s/%s/%s/%s@%d: %w", project, domain, taskName, name, revision, err)
		}
		return nil, fmt.Errorf("failed to get trigger revision: %w", err)
	}
	return &rev, nil
}

func (r *triggerRepo) ListTriggers(ctx context.Context, input interfaces.ListResourceInput) ([]*models.Trigger, error) {
	var queryBuilder strings.Builder
	var args []interface{}
	argIdx := 1

	queryBuilder.WriteString("SELECT * FROM triggers WHERE deleted_at IS NULL")

	if input.Filter != nil {
		expr, err := input.Filter.QueryExpression("")
		if err != nil {
			return nil, fmt.Errorf("failed to build filter: %w", err)
		}
		rewritten, rewrittenArgs := rewritePlaceholders(expr.Query, expr.Args, argIdx)
		queryBuilder.WriteString(" AND ")
		queryBuilder.WriteString(rewritten)
		args = append(args, rewrittenArgs...)
		argIdx += len(rewrittenArgs)
	}
	if input.ScopeByFilter != nil {
		expr, err := input.ScopeByFilter.QueryExpression("")
		if err != nil {
			return nil, fmt.Errorf("failed to build scope filter: %w", err)
		}
		rewritten, rewrittenArgs := rewritePlaceholders(expr.Query, expr.Args, argIdx)
		queryBuilder.WriteString(" AND ")
		queryBuilder.WriteString(rewritten)
		args = append(args, rewrittenArgs...)
		argIdx += len(rewrittenArgs)
	}

	if len(input.SortParameters) > 0 {
		queryBuilder.WriteString(" ORDER BY ")
		for i, sp := range input.SortParameters {
			if i > 0 {
				queryBuilder.WriteString(", ")
			}
			queryBuilder.WriteString(sp.GetOrderExpr())
		}
	} else {
		queryBuilder.WriteString(" ORDER BY active DESC, updated_at DESC")
	}

	queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1))
	args = append(args, input.Limit, input.Offset)

	var triggers []*models.Trigger
	if err := sqlx.SelectContext(ctx, r.db, &triggers, queryBuilder.String(), args...); err != nil {
		return nil, fmt.Errorf("failed to list triggers: %w", err)
	}
	return triggers, nil
}

func (r *triggerRepo) ListTriggerRevisions(ctx context.Context, project, domain, taskName, name string, input interfaces.ListResourceInput) ([]*models.TriggerRevision, error) {
	var revisions []*models.TriggerRevision
	err := sqlx.SelectContext(ctx, r.db, &revisions,
		`SELECT * FROM trigger_revisions
		 WHERE project = $1 AND domain = $2 AND task_name = $3 AND name = $4
		 ORDER BY revision DESC
		 LIMIT $5 OFFSET $6`,
		project, domain, taskName, name, input.Limit, input.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list trigger revisions: %w", err)
	}
	return revisions, nil
}

func (r *triggerRepo) UpdateTriggers(ctx context.Context, keys []interfaces.TriggerNameKey, active bool) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, key := range keys {
		var t models.Trigger
		err := tx.QueryRowxContext(ctx, `
			UPDATE triggers SET
				active          = $1,
				updated_at      = NOW(),
				latest_revision = latest_revision + 1
			WHERE project = $2 AND domain = $3 AND task_name = $4 AND name = $5
			  AND deleted_at IS NULL
			  AND active != $1
			RETURNING *`,
			active, key.Project, key.Domain, key.TaskName, key.Name,
		).StructScan(&t)

		if errors.Is(err, sql.ErrNoRows) {
			continue // already in desired state or not found
		}
		if err != nil {
			return fmt.Errorf("failed to update trigger %+v: %w", key, err)
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

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}
	return nil
}

func (r *triggerRepo) DeleteTriggers(ctx context.Context, keys []interfaces.TriggerNameKey) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, key := range keys {
		var t models.Trigger
		err := tx.QueryRowxContext(ctx, `
			UPDATE triggers SET
				deleted_at      = NOW(),
				active          = FALSE,
				latest_revision = latest_revision + 1
			WHERE project = $1 AND domain = $2 AND task_name = $3 AND name = $4
			  AND deleted_at IS NULL
			RETURNING *`,
			key.Project, key.Domain, key.TaskName, key.Name,
		).StructScan(&t)

		if errors.Is(err, sql.ErrNoRows) {
			continue // already deleted or not found
		}
		if err != nil {
			return fmt.Errorf("failed to delete trigger %+v: %w", key, err)
		}

		if err := insertTriggerRevision(ctx, tx, &t, triggerpb.TriggerRevisionAction_TRIGGER_REVISION_ACTION_DELETE); err != nil {
			return err
		}
		if err := refreshTaskTriggerMeta(ctx, tx, &t); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}
	return nil
}

// insertTriggerRevision appends an immutable snapshot to trigger_revisions.
func insertTriggerRevision(ctx context.Context, tx *sqlx.Tx, t *models.Trigger, action triggerpb.TriggerRevisionAction) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO trigger_revisions (
			project, domain, task_name, name, revision,
			spec, automation_spec,
			task_version, active, automation_type,
			deployed_by, updated_by,
			deployed_at, updated_at, triggered_at, deleted_at,
			action, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, NOW())`,
		t.Project, t.Domain, t.TaskName, t.Name, t.LatestRevision,
		t.Spec, t.AutomationSpec,
		t.TaskVersion, t.Active, t.AutomationType,
		t.DeployedBy, t.UpdatedBy,
		t.DeployedAt, t.UpdatedAt, t.TriggeredAt, t.DeletedAt,
		action.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to insert trigger revision: %w", err)
	}
	return nil
}

// listTaskTriggers returns the names of all non-deleted triggers for a given task (across all task versions).
func listTaskTriggers(ctx context.Context, tx *sqlx.Tx, project, domain, taskName string) ([]string, error) {
	var names []string
	err := sqlx.SelectContext(ctx, tx, &names,
		`SELECT name FROM triggers
		 WHERE project = $1 AND domain = $2 AND task_name = $3
		   AND deleted_at IS NULL`,
		project, domain, taskName)
	if err != nil {
		return nil, fmt.Errorf("failed to list task triggers: %w", err)
	}
	return names, nil
}

// softDeleteTrigger soft-deletes a trigger by name and appends a DELETE revision snapshot.
func softDeleteTrigger(ctx context.Context, tx *sqlx.Tx, project, domain, taskName, name string) error {
	var t models.Trigger
	err := tx.QueryRowxContext(ctx, `
		UPDATE triggers SET
			deleted_at      = NOW(),
			active          = FALSE,
			latest_revision = latest_revision + 1
		WHERE project = $1 AND domain = $2 AND task_name = $3 AND name = $4
		  AND deleted_at IS NULL
		RETURNING *`,
		project, domain, taskName, name,
	).StructScan(&t)
	if errors.Is(err, sql.ErrNoRows) {
		return nil // already deleted
	}
	if err != nil {
		return fmt.Errorf("failed to soft-delete trigger %q: %w", name, err)
	}
	return insertTriggerRevision(ctx, tx, &t, triggerpb.TriggerRevisionAction_TRIGGER_REVISION_ACTION_DELETE)
}

// refreshTaskTriggerMeta is the single source of truth for the denormalized
// trigger summary columns on the tasks row. It recomputes them from the current
// set of (non-deleted) rows in the triggers table and writes:
//
//	tasks.trigger_name            — name of the sole attached trigger, or NULL
//	tasks.total_triggers          — count of attached triggers
//	tasks.active_triggers         — count of attached triggers with active = true
//	tasks.trigger_automation_spec — automation_spec of the sole attached trigger, or NULL
//
// Callers that modify the triggers table (upsert/update/delete) MUST invoke this
// within the same transaction so the task summary stays consistent. Do not write
// these columns from any other code path — they will be overwritten here.
func refreshTaskTriggerMeta(ctx context.Context, tx *sqlx.Tx, t *models.Trigger) error {
	type statRow struct {
		Active         bool           `db:"active"`
		Count          int            `db:"count"`
		TriggerName    sql.NullString `db:"trigger_name"`
		AutomationSpec []byte         `db:"automation_spec"`
	}

	var rows []statRow
	err := sqlx.SelectContext(ctx, tx, &rows, `
		SELECT active, COUNT(*) AS count,
		       STRING_AGG(name, ',') AS trigger_name,
		       STRING_AGG(encode(automation_spec, 'escape'), ',') AS automation_spec
		FROM triggers
		WHERE project = $1 AND domain = $2 AND task_name = $3 AND task_version = $4
		  AND deleted_at IS NULL
		GROUP BY active`,
		t.Project, t.Domain, t.TaskName, t.TaskVersion,
	)
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

	result, err := tx.ExecContext(ctx, `
		UPDATE tasks SET
			trigger_name            = $1,
			total_triggers          = $2,
			active_triggers         = $3,
			trigger_automation_spec = $4
		WHERE project = $5 AND domain = $6 AND name = $7 AND version = $8`,
		triggerName, totalTriggers, activeTriggers, automationSpec,
		t.Project, t.Domain, t.TaskName, t.TaskVersion,
	)
	if err != nil {
		return fmt.Errorf("failed to update task trigger meta: %w", err)
	}
	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		logger.Warnf(ctx, "task %s/%s/%s@%s not found when refreshing trigger meta",
			t.Project, t.Domain, t.TaskName, t.TaskVersion)
	}
	return nil
}
