package impl

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/flyteorg/flyte/v2/flytestdlib/database"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

const rootActionName = "a0"

// actionRepo implements actionRepo interface using PostgreSQL
type actionRepo struct {
	db       *sqlx.DB
	dsn      string // stored for pq.Listener
	listener *pq.Listener

	// Subscriber management for LISTEN/NOTIFY
	runSubscribers    map[chan string]bool
	actionSubscribers map[chan string]bool
	mu                sync.RWMutex

	// Dedicated channels for async NOTIFY to avoid pool contention
	actionNotifyCh chan string
	runNotifyCh    chan string
}

// NewActionRepo creates a new PostgreSQL repository
func NewActionRepo(db *sqlx.DB, dbConfig database.DbConfig) (interfaces.ActionRepo, error) {
	dsn := database.GetPostgresDsn(context.Background(), dbConfig.Postgres)
	repo := &actionRepo{
		db:                db,
		dsn:               dsn,
		runSubscribers:    make(map[chan string]bool),
		actionSubscribers: make(map[chan string]bool),
	}

	repo.actionNotifyCh = make(chan string, 256)
	repo.runNotifyCh = make(chan string, 256)

	if err := repo.startPostgresListener(); err != nil {
		return nil, fmt.Errorf("failed to start postgres listener: %w", err)
	}
	if err := repo.startNotifyLoop(); err != nil {
		return nil, fmt.Errorf("failed to start notify loop: %w", err)
	}

	return repo, nil
}

// GetRun retrieves a run by identifier
func (r *actionRepo) GetRun(ctx context.Context, runID *common.RunIdentifier) (*models.Run, error) {
	var run models.Run
	err := sqlx.GetContext(ctx, r.db, &run,
		"SELECT * FROM actions WHERE project = $1 AND domain = $2 AND run_name = $3 AND parent_action_name IS NULL",
		runID.Project, runID.Domain, runID.Name)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("run not found: %s/%s/%s",
				runID.Project, runID.Domain, runID.Name)
		}
		return nil, fmt.Errorf("failed to get run: %w", err)
	}

	return &run, nil
}

// AbortRun marks only the root action as ABORTED and sets abort_requested_at on it.
// K8s cascades CRD deletion to child actions via OwnerReferences; the action service
// informer handles marking them ABORTED in DB when their CRDs are deleted.
func (r *actionRepo) AbortRun(ctx context.Context, runID *common.RunIdentifier, reason string, abortedBy *common.EnrichedIdentity) error {
	now := time.Now()

	var rootName string
	err := r.db.QueryRowxContext(ctx,
		`UPDATE actions SET phase = $1, updated_at = $2, abort_requested_at = $3, abort_attempt_count = $4, abort_reason = $5,
		                    ended_at = COALESCE(ended_at, GREATEST($2, created_at)),
		                    duration_ms = EXTRACT(EPOCH FROM (COALESCE(ended_at, GREATEST($2, created_at)) - created_at)) * 1000
		 WHERE project = $6 AND domain = $7 AND run_name = $8 AND parent_action_name IS NULL
		 RETURNING name`,
		int32(common.ActionPhase_ACTION_PHASE_ABORTED), now, now, 0, reason,
		runID.Project, runID.Domain, runID.Name,
	).Scan(&rootName)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("run not found: %w", sql.ErrNoRows)
	}
	if err != nil {
		return fmt.Errorf("failed to abort run: %w", err)
	}

	rootID := &common.ActionIdentifier{Run: runID, Name: rootName}
	r.notifyRunUpdate(ctx, runID)
	r.notifyActionUpdate(ctx, rootID)

	logger.Infof(ctx, "Aborted run: %s/%s/%s", runID.Project, runID.Domain, runID.Name)
	return nil
}

// InsertEvents inserts a batch of action events, ignoring duplicates (same PK = idempotent).
func (r *actionRepo) InsertEvents(ctx context.Context, events []*models.ActionEvent) error {
	if len(events) == 0 {
		return nil
	}

	for _, e := range events {
		_, err := r.db.ExecContext(ctx,
			`INSERT INTO action_events (project, domain, run_name, name, attempt, phase, version, info, error_kind, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			 ON CONFLICT DO NOTHING`,
			e.Project, e.Domain, e.RunName, e.Name, e.Attempt, e.Phase, e.Version, e.Info, e.ErrorKind)
		if err != nil {
			return err
		}
	}

	// Notify subscribers so watchers see new events (e.g. log context becoming available).
	notified := make(map[string]bool)
	for _, e := range events {
		actionID := &common.ActionIdentifier{
			Run: &common.RunIdentifier{
				Project: e.Project,
				Domain:  e.Domain,
				Name:    e.RunName,
			},
			Name: e.Name,
		}
		key := e.Project + "/" + e.Domain + "/" + e.RunName + "/" + e.Name
		if !notified[key] {
			r.notifyActionUpdate(ctx, actionID)
			notified[key] = true
		}
	}
	return nil
}

// ListEvents lists action events for a given action identifier.
func (r *actionRepo) ListEvents(ctx context.Context, actionID *common.ActionIdentifier, limit int) ([]*models.ActionEvent, error) {
	var events []*models.ActionEvent
	err := sqlx.SelectContext(ctx, r.db, &events,
		`SELECT * FROM action_events
		 WHERE project = $1 AND domain = $2 AND run_name = $3 AND name = $4
		 ORDER BY attempt ASC, phase ASC, version ASC
		 LIMIT $5`,
		actionID.Run.Project, actionID.Run.Domain, actionID.Run.Name, actionID.Name, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list action events: %w", err)
	}
	return events, nil
}

// ListEventsSince lists action events for a given action identifier updated at or after the provided checkpoint.
func (r *actionRepo) ListEventsSince(
	ctx context.Context,
	actionID *common.ActionIdentifier,
	attempt uint32,
	since time.Time,
	offset, limit int,
) ([]*models.ActionEvent, error) {
	var events []*models.ActionEvent
	err := sqlx.SelectContext(ctx, r.db, &events,
		`SELECT * FROM action_events
		 WHERE project = $1 AND domain = $2 AND run_name = $3 AND name = $4 AND attempt = $5 AND updated_at > $6
		 ORDER BY updated_at ASC, attempt ASC, phase ASC, version ASC
		 OFFSET $7 LIMIT $8`,
		actionID.Run.Project, actionID.Run.Domain, actionID.Run.Name, actionID.Name, attempt, since,
		offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list action events since %s: %w", since.Format(time.RFC3339Nano), err)
	}
	return events, nil
}

// GetLatestEventByAttempt returns the most recent event for a given attempt,
// ordered by version descending, without deserializing all events.
func (r *actionRepo) GetLatestEventByAttempt(ctx context.Context, actionID *common.ActionIdentifier, attempt uint32) (*models.ActionEvent, error) {
	var event models.ActionEvent
	err := sqlx.GetContext(ctx, r.db, &event,
		`SELECT * FROM action_events
		 WHERE project = $1 AND domain = $2 AND run_name = $3 AND name = $4 AND attempt = $5
		 ORDER BY phase DESC, version DESC
		 LIMIT 1`,
		actionID.Run.Project, actionID.Run.Domain, actionID.Run.Name, actionID.Name, attempt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("event not found for attempt %d: %w", attempt, sql.ErrNoRows)
		}
		return nil, fmt.Errorf("failed to get latest event for attempt %d: %w", attempt, err)
	}
	return &event, nil
}

// CreateAction inserts an Action model into the database.
// When updateTriggeredAt is true and the action has a TriggerName set, triggered_at on the
// corresponding trigger row is updated to now() in the same transaction.
func (r *actionRepo) CreateAction(ctx context.Context, action *models.Action, updateTriggeredAt bool) (*models.Action, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Use the model's CreatedAt if set (e.g. trace actions use start_time so parents
	// sort before children), otherwise fall back to the current time.
	createdAt := action.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	result, err := tx.ExecContext(ctx,
		`INSERT INTO actions (project, domain, run_name, name, parent_action_name, phase, run_source, action_type, action_group, task_project, task_domain, task_name, task_version, task_type, task_short_name, function_name, environment_name, action_spec, action_details, detailed_info, run_spec, attempts, cache_status, trigger_name, trigger_task_name, trigger_revision, created_at, ended_at, duration_ms)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, CASE WHEN $28::timestamptz IS NOT NULL THEN EXTRACT(EPOCH FROM (GREATEST($28::timestamptz, $27) - $27)) * 1000 ELSE NULL END)
		 ON CONFLICT DO NOTHING`,
		action.Project, action.Domain, action.RunName, action.Name, action.ParentActionName, action.Phase, action.RunSource, action.ActionType, action.ActionGroup,
		action.TaskProject, action.TaskDomain, action.TaskName, action.TaskVersion, action.TaskType, action.TaskShortName, action.FunctionName, action.EnvironmentName,
		action.ActionSpec, action.ActionDetails, action.DetailedInfo, action.RunSpec, action.Attempts, action.CacheStatus,
		action.TriggerName, action.TriggerTaskName, action.TriggerRevision, createdAt, action.EndedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create action: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 && updateTriggeredAt && action.TriggerName.Valid {
		if _, err := tx.ExecContext(ctx,
			`UPDATE triggers SET triggered_at = NOW() WHERE project = $1 AND domain = $2 AND task_name = $3 AND name = $4`,
			action.Project, action.Domain, action.TriggerTaskName.String, action.TriggerName.String,
		); err != nil {
			return nil, fmt.Errorf("failed to update triggered_at: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit tx: %w", err)
	}

	actionID := &common.ActionIdentifier{
		Name: action.Name,
		Run: &common.RunIdentifier{
			Project: action.Project,
			Domain:  action.Domain,
			Name:    action.RunName,
		},
	}

	// If no rows were affected, the action already exists — fetch and return it.
	if rowsAffected == 0 {
		existing, err := r.GetAction(ctx, actionID)
		if err != nil {
			return nil, fmt.Errorf("failed to get existing action: %w", err)
		}
		return existing, nil
	}

	// Fetch the created action to get DB-generated fields (created_at, updated_at)
	created, err := r.GetAction(ctx, actionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created action: %w", err)
	}

	logger.Infof(ctx, "Created action: %s", created.Name)

	// Notify subscribers of action creation
	r.notifyActionUpdate(ctx, actionID)

	return created, nil
}

// GetAction retrieves an action by identifier
func (r *actionRepo) GetAction(ctx context.Context, actionID *common.ActionIdentifier) (*models.Action, error) {
	var action models.Action
	err := sqlx.GetContext(ctx, r.db, &action,
		"SELECT * FROM actions WHERE project = $1 AND domain = $2 AND run_name = $3 AND name = $4",
		actionID.Run.Project, actionID.Run.Domain, actionID.Run.Name, actionID.Name)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("action not found")
		}
		return nil, fmt.Errorf("failed to get action: %w", err)
	}

	return &action, nil
}

// ListActions lists actions matching the given input filter, sort, and pagination.
func (r *actionRepo) ListActions(ctx context.Context, input interfaces.ListResourceInput) ([]*models.Action, error) {
	var queryBuilder strings.Builder
	var args []interface{}

	queryBuilder.WriteString("SELECT * FROM actions")

	if input.Filter != nil {
		expr, err := input.Filter.QueryExpression("")
		if err != nil {
			return nil, fmt.Errorf("failed to build filter expression: %w", err)
		}
		queryBuilder.WriteString(" WHERE ")
		queryBuilder.WriteString(expr.Query)
		args = append(args, expr.Args...)
	}

	if input.CursorToken != "" {
		if t, err := time.Parse(time.RFC3339Nano, input.CursorToken); err == nil {
			// If a filter was already applied above, the WHERE clause is already open
			// and we extend it with AND. Otherwise we open a new WHERE clause.
			// Use < because the default sort is DESC (newest first): each page
			// continues from rows older than the last row of the previous page.
			if input.Filter != nil {
				queryBuilder.WriteString(" AND created_at < ?")
			} else {
				queryBuilder.WriteString(" WHERE created_at < ?")
			}
			args = append(args, t)
		}
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
		// Default sorting non-terminal runs at the top, then by most recent start time.
		queryBuilder.WriteString(" ORDER BY phase ASC, created_at DESC")
	}

	queryBuilder.WriteString(" LIMIT ?")
	args = append(args, input.Limit+1)

	query := sqlx.Rebind(sqlx.DOLLAR, queryBuilder.String())

	var actions []*models.Action
	if err := sqlx.SelectContext(ctx, r.db, &actions, query, args...); err != nil {
		return nil, fmt.Errorf("failed to list actions: %w", err)
	}

	return actions, nil
}

// UpdateActionPhase updates the phase of an action.
// endTime should be set when the action reaches a terminal phase.
func (r *actionRepo) UpdateActionPhase(
	ctx context.Context,
	actionID *common.ActionIdentifier,
	phase common.ActionPhase,
	attempts uint32,
	cacheStatus core.CatalogCacheStatus,
	endTime *time.Time,
) error {
	now := time.Now()
	retryablePhases := []int32{
		int32(common.ActionPhase_ACTION_PHASE_FAILED),
		int32(common.ActionPhase_ACTION_PHASE_TIMED_OUT),
	}

	var queryBuilder strings.Builder
	var args []interface{}
	argIdx := 1

	queryBuilder.WriteString("UPDATE actions SET ")
	queryBuilder.WriteString(fmt.Sprintf("phase = $%d, attempts = $%d, cache_status = $%d, updated_at = $%d", argIdx, argIdx+1, argIdx+2, argIdx+3))
	args = append(args, phase, attempts, cacheStatus, now)
	argIdx += 4

	if endTime != nil {
		queryBuilder.WriteString(fmt.Sprintf(", ended_at = COALESCE(ended_at, GREATEST($%d, created_at))", argIdx))
		args = append(args, *endTime)
		argIdx++
		queryBuilder.WriteString(fmt.Sprintf(", duration_ms = EXTRACT(EPOCH FROM (COALESCE(ended_at, GREATEST($%d, created_at)) - created_at)) * 1000", argIdx))
		args = append(args, *endTime)
		argIdx++
	}

	queryBuilder.WriteString(fmt.Sprintf(" WHERE project = $%d AND domain = $%d AND run_name = $%d AND name = $%d AND (phase <= $%d OR phase = ANY($%d))",
		argIdx, argIdx+1, argIdx+2, argIdx+3, argIdx+4, argIdx+5))
	args = append(args, actionID.Run.Project, actionID.Run.Domain, actionID.Run.Name, actionID.Name, phase, pq.Array(retryablePhases))

	result, err := r.db.ExecContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected > 0 {
		r.notifyActionUpdate(ctx, actionID)
	}

	// If this is the root action (the run itself), also notify run subscribers
	// so that WatchRuns streams reflect phase transitions (e.g. RUNNING → SUCCEEDED).
	if actionID.Name == rootActionName {
		r.notifyRunUpdate(ctx, actionID.Run)
	}

	return nil
}

// AbortAction marks only the targeted action as ABORTED and sets abort_requested_at.
// K8s cascades CRD deletion to descendants via OwnerReferences; the action service
// informer handles marking them ABORTED in DB when their CRDs are deleted.
func (r *actionRepo) AbortAction(ctx context.Context, actionID *common.ActionIdentifier, reason string, abortedBy *common.EnrichedIdentity) error {
	now := time.Now()

	result, err := r.db.ExecContext(ctx,
		`UPDATE actions SET phase = $1, updated_at = $2, abort_requested_at = $3, abort_attempt_count = $4, abort_reason = $5,
		                    ended_at = COALESCE(ended_at, GREATEST($2, created_at)),
		                    duration_ms = EXTRACT(EPOCH FROM (COALESCE(ended_at, GREATEST($2, created_at)) - created_at)) * 1000
		 WHERE project = $6 AND domain = $7 AND run_name = $8 AND name = $9`,
		int32(common.ActionPhase_ACTION_PHASE_ABORTED), now, now, 0, reason,
		actionID.Run.Project, actionID.Run.Domain, actionID.Run.Name, actionID.Name)
	if err != nil {
		return fmt.Errorf("failed to abort action: %w", err)
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return fmt.Errorf("action not found: %w", sql.ErrNoRows)
	}

	r.notifyActionUpdate(ctx, actionID)

	logger.Infof(ctx, "AbortAction: aborted %s", actionID.Name)
	return nil
}

// ListPendingAborts returns all actions that have abort_requested_at set (i.e. awaiting pod termination).
func (r *actionRepo) ListPendingAborts(ctx context.Context) ([]*models.Action, error) {
	var actions []*models.Action
	err := sqlx.SelectContext(ctx, r.db, &actions,
		"SELECT * FROM actions WHERE abort_requested_at IS NOT NULL")
	if err != nil {
		return nil, fmt.Errorf("failed to list pending aborts: %w", err)
	}
	return actions, nil
}

// MarkAbortAttempt increments abort_attempt_count and returns the new value.
// Called by the reconciler before each actionsClient.Abort call.
func (r *actionRepo) MarkAbortAttempt(ctx context.Context, actionID *common.ActionIdentifier) (int, error) {
	var abortAttemptCount int
	err := r.db.QueryRowxContext(ctx,
		`UPDATE actions SET abort_attempt_count = abort_attempt_count + 1, updated_at = $1
		 WHERE project = $2 AND domain = $3 AND run_name = $4 AND name = $5
		 RETURNING abort_attempt_count`,
		time.Now(), actionID.Run.Project, actionID.Run.Domain, actionID.Run.Name, actionID.Name,
	).Scan(&abortAttemptCount)
	if err != nil {
		return 0, fmt.Errorf("failed to mark abort attempt: %w", err)
	}
	return abortAttemptCount, nil
}

// ClearAbortRequest clears abort_requested_at (and resets counters) once the pod is confirmed terminated.
func (r *actionRepo) ClearAbortRequest(ctx context.Context, actionID *common.ActionIdentifier) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE actions SET abort_requested_at = NULL, abort_attempt_count = 0, updated_at = $1
		 WHERE project = $2 AND domain = $3 AND run_name = $4 AND name = $5`,
		time.Now(), actionID.Run.Project, actionID.Run.Domain, actionID.Run.Name, actionID.Name)
	if err != nil {
		return fmt.Errorf("failed to clear abort request: %w", err)
	}
	return nil
}

// UpdateActionState updates the state of an action
func (r *actionRepo) UpdateActionState(ctx context.Context, actionID *common.ActionIdentifier, state string) error {
	// Parse the state JSON to extract the phase
	var stateObj map[string]interface{}
	if err := json.Unmarshal([]byte(state), &stateObj); err != nil {
		return fmt.Errorf("failed to unmarshal state JSON: %w", err)
	}

	now := time.Now()

	// Extract phase if present
	var phase interface{}
	if p, ok := stateObj["phase"].(string); ok {
		phase = p
		logger.Infof(ctx, "Updating action %s phase to %s", actionID.Name, p)
	}

	var result sql.Result
	var err error
	if phase != nil {
		result, err = r.db.ExecContext(ctx,
			`UPDATE actions SET phase = $1, action_details = $2, updated_at = $3
			 WHERE project = $4 AND domain = $5 AND run_name = $6 AND name = $7`,
			phase, []byte(state), now,
			actionID.Run.Project, actionID.Run.Domain, actionID.Run.Name, actionID.Name)
	} else {
		result, err = r.db.ExecContext(ctx,
			`UPDATE actions SET action_details = $1, updated_at = $2
			 WHERE project = $3 AND domain = $4 AND run_name = $5 AND name = $6`,
			[]byte(state), now,
			actionID.Run.Project, actionID.Run.Domain, actionID.Run.Name, actionID.Name)
	}

	if err != nil {
		return fmt.Errorf("failed to update action state: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("action not found: %s/%s/%s/%s",
			actionID.Run.Project, actionID.Run.Domain, actionID.Run.Name, actionID.Name)
	}

	// Notify subscribers of the update
	r.notifyActionUpdate(ctx, actionID)

	return nil
}

// GetActionState retrieves the state of an action
func (r *actionRepo) GetActionState(ctx context.Context, actionID *common.ActionIdentifier) (string, error) {
	var actionDetails []byte
	err := r.db.QueryRowContext(ctx,
		"SELECT action_details FROM actions WHERE project = $1 AND domain = $2 AND run_name = $3 AND name = $4",
		actionID.Run.Project, actionID.Run.Domain, actionID.Run.Name, actionID.Name,
	).Scan(&actionDetails)
	if err != nil {
		return "", fmt.Errorf("failed to get action state: %w", err)
	}

	return string(actionDetails), nil
}

// NotifyStateUpdate sends a notification about a state update
func (r *actionRepo) NotifyStateUpdate(ctx context.Context, actionID *common.ActionIdentifier) error {
	// This is already handled by notifyActionUpdate
	r.notifyActionUpdate(ctx, actionID)
	return nil
}

// WatchStateUpdates watches for state updates (simplified implementation)
func (r *actionRepo) WatchStateUpdates(ctx context.Context, updates chan<- *common.ActionIdentifier, errs chan<- error) {
	// For now, just block until context is cancelled
	// In a full implementation, this would listen for state notifications
	<-ctx.Done()
}

// WatchRunUpdates watches for run updates via LISTEN/NOTIFY
func (r *actionRepo) WatchRunUpdates(ctx context.Context, runID *common.RunIdentifier, updates chan<- *models.Run, errs chan<- error) {
	runKey := fmt.Sprintf("%s/%s/%s", runID.Project, runID.Domain, runID.Name)
	notifCh := make(chan string, 100)

	r.mu.Lock()
	r.runSubscribers[notifCh] = true
	r.mu.Unlock()

	defer func() {
		r.mu.Lock()
		delete(r.runSubscribers, notifCh)
		close(notifCh)
		r.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case notifPayload := <-notifCh:
			if notifPayload == runKey {
				run, err := r.GetRun(ctx, runID)
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					errs <- err
					return
				}
				select {
				case updates <- run:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

// WatchAllRunUpdates watches for all run updates via LISTEN/NOTIFY
func (r *actionRepo) WatchAllRunUpdates(ctx context.Context, updates chan<- *models.Run, errs chan<- error) {
	notifCh := make(chan string, 100)

	r.mu.Lock()
	r.runSubscribers[notifCh] = true
	r.mu.Unlock()

	defer func() {
		r.mu.Lock()
		delete(r.runSubscribers, notifCh)
		close(notifCh)
		r.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case notifPayload := <-notifCh:
			parts := strings.Split(notifPayload, "/")
			if len(parts) != 3 {
				logger.Warnf(ctx, "Invalid run notification payload: %s", notifPayload)
				continue
			}

			runID := &common.RunIdentifier{
				Project: parts[0],
				Domain:  parts[1],
				Name:    parts[2],
			}

			if ctx.Err() != nil {
				return
			}
			run, err := r.GetRun(ctx, runID)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				logger.Errorf(ctx, "Failed to get run from notification: %v", err)
				continue
			}
			select {
			case updates <- run:
			case <-ctx.Done():
				return
			}
		}
	}
}

// WatchAllActionUpdates watches for all action updates for a run via LISTEN/NOTIFY
func (r *actionRepo) WatchAllActionUpdates(ctx context.Context, runID *common.RunIdentifier, updates chan<- *models.Action, errs chan<- error) {
	runPrefix := fmt.Sprintf("%s/%s/%s/", runID.Project, runID.Domain, runID.Name)
	notifCh := make(chan string, 100)

	r.mu.Lock()
	r.actionSubscribers[notifCh] = true
	r.mu.Unlock()

	defer func() {
		r.mu.Lock()
		delete(r.actionSubscribers, notifCh)
		close(notifCh)
		r.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case notifPayload := <-notifCh:
			// Collect this notification plus any others already queued,
			// deduplicating by action name so we only query the DB once
			// per action during a burst of updates.
			pending := make(map[string]bool)
			if len(notifPayload) > len(runPrefix) && notifPayload[:len(runPrefix)] == runPrefix {
				pending[notifPayload[len(runPrefix):]] = true
			}
			// Drain buffered notifications without blocking.
		drain:
			for {
				select {
				case extra := <-notifCh:
					if len(extra) > len(runPrefix) && extra[:len(runPrefix)] == runPrefix {
						pending[extra[len(runPrefix):]] = true
					}
				default:
					break drain
				}
			}

			for actionName := range pending {
				actionID := &common.ActionIdentifier{
					Run:  runID,
					Name: actionName,
				}
				if ctx.Err() != nil {
					return
				}
				action, err := r.GetAction(ctx, actionID)
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					logger.Errorf(ctx, "Failed to get action from notification: %v", err)
					continue
				}
				select {
				case updates <- action:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

// WatchActionUpdates watches the current action update via LISTEN/NOTIFY
func (r *actionRepo) WatchActionUpdates(ctx context.Context, actionID *common.ActionIdentifier, updates chan<- *models.Action, errs chan<- error) {
	targetPayload := fmt.Sprintf("%s/%s/%s/%s",
		actionID.Run.Project, actionID.Run.Domain, actionID.Run.Name, actionID.Name)
	notifCh := make(chan string, 100)

	r.mu.Lock()
	r.actionSubscribers[notifCh] = true
	r.mu.Unlock()

	defer func() {
		r.mu.Lock()
		delete(r.actionSubscribers, notifCh)
		close(notifCh)
		r.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case notifPayload := <-notifCh:
			if notifPayload != targetPayload {
				continue
			}

		drain:
			for {
				// Prevent continuous update from the action
				select {
				case extra := <-notifCh:
					if extra != targetPayload {
						continue
					}
				default:
					break drain
				}
			}

			action, err := r.GetAction(ctx, actionID)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				logger.Errorf(ctx, "Failed to get action from notification: %v", err)
				continue
			}

			select {
			case updates <- action:
			case <-ctx.Done():
				return
			}
		}
	}
}

// startPostgresListener starts the PostgreSQL LISTEN/NOTIFY listener.
// Returns an error if the connection or LISTEN setup fails.
func (r *actionRepo) startPostgresListener() error {
	connStr := r.dsn

	r.listener = pq.NewListener(connStr, 10*time.Second, time.Minute, func(ev pq.ListenerEventType, err error) {
		if err != nil {
			logger.Errorf(context.Background(), "Listener error: %v", err)
		}
	})

	// Listen to channels
	if err := r.listener.Listen("run_updates"); err != nil {
		return fmt.Errorf("listen to run_updates: %w", err)
	}

	if err := r.listener.Listen("action_updates"); err != nil {
		return fmt.Errorf("listen to action_updates: %w", err)
	}

	// Process notifications in background
	go r.processNotifications()

	return nil
}

// processNotifications handles incoming LISTEN/NOTIFY messages.
func (r *actionRepo) processNotifications() {
	logger.Infof(context.Background(), "PostgreSQL LISTEN/NOTIFY started")

	// Process notifications
	for {
		select {
		case notif := <-r.listener.Notify:
			if notif == nil {
				continue
			}

			switch notif.Channel {
			case "run_updates":
				// Broadcast to all run subscribers
				r.mu.RLock()
				for ch := range r.runSubscribers {
					select {
					case ch <- notif.Extra:
					default:
						// Channel full, skip this subscriber
						logger.Warnf(context.Background(), "Run subscriber channel full, dropping notification")
					}
				}
				r.mu.RUnlock()

			case "action_updates":
				// Broadcast to all action subscribers
				r.mu.RLock()
				for ch := range r.actionSubscribers {
					select {
					case ch <- notif.Extra:
					default:
						logger.Warnf(context.Background(), "Action subscriber channel full, dropping notification payload=%s", notif.Extra)
					}
				}
				r.mu.RUnlock()

			}

		case <-time.After(90 * time.Second):
			// Ping to keep connection alive
			if err := r.listener.Ping(); err != nil {
				logger.Errorf(context.Background(), "Listener ping failed: %v", err)
				return
			}
		}
	}
}

// notifyRunUpdate sends a notification about a run update via the
// dedicated notify channel, avoiding connection pool contention.
func (r *actionRepo) notifyRunUpdate(ctx context.Context, runID *common.RunIdentifier) {
	payload := fmt.Sprintf("%s/%s/%s", runID.Project, runID.Domain, runID.Name)

	select {
	case r.runNotifyCh <- payload:
	case <-ctx.Done():
		logger.Warnf(ctx, "Run NOTIFY send cancelled for %s: %v", payload, ctx.Err())
	}
}

// ListRootActions lists root actions (runs) matching scope and date filters.
func (r *actionRepo) ListRootActions(ctx context.Context, project, domain string, startDate, endDate *time.Time, limit int) ([]*models.Action, error) {
	var queryBuilder strings.Builder
	var args []interface{}
	argIdx := 1

	queryBuilder.WriteString("SELECT * FROM actions WHERE parent_action_name IS NULL")

	if project != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND project = $%d", argIdx))
		args = append(args, project)
		argIdx++
	}
	if domain != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND domain = $%d", argIdx))
		args = append(args, domain)
		argIdx++
	}
	if startDate != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND created_at >= $%d", argIdx))
		args = append(args, *startDate)
		argIdx++
	}
	if endDate != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND created_at <= $%d", argIdx))
		args = append(args, *endDate)
		argIdx++
	}
	if limit <= 0 {
		limit = 1000
	}

	queryBuilder.WriteString(fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", argIdx))
	args = append(args, limit)

	var actions []*models.Action
	if err := sqlx.SelectContext(ctx, r.db, &actions, queryBuilder.String(), args...); err != nil {
		return nil, fmt.Errorf("failed to list root actions: %w", err)
	}
	return actions, nil
}

// startNotifyLoop acquires a dedicated DB connection for NOTIFY commands and
// starts processing in the background. Returns an error if the initial
// connection cannot be established.
func (r *actionRepo) startNotifyLoop() error {
	sqlDB := r.db.DB

	conn, err := sqlDB.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("acquire dedicated NOTIFY connection: %w", err)
	}

	go r.runNotifyLoop(sqlDB, conn)
	return nil
}

// isConnError reports whether err indicates a broken or lost database connection.
func isConnError(err error) bool {
	if errors.Is(err, driver.ErrBadConn) {
		return true
	}
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		return true
	}
	// lib/pq wraps connection-class errors (Class 08) as *pq.Error.
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code.Class() == "08" {
		return true
	}
	return false
}

// runNotifyLoop processes notify channels using the given connection.
// On connection errors it attempts to reconnect; if reconnection fails
// the error is logged and the notification is skipped.
func (r *actionRepo) runNotifyLoop(sqlDB *sql.DB, conn *sql.Conn) {
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()

	reconnect := func() {
		if conn != nil {
			conn.Close()
			conn = nil
		}
		if sqlDB == nil {
			return
		}
		var err error
		conn, err = sqlDB.Conn(context.Background())
		if err != nil {
			logger.Errorf(context.Background(), "Failed to re-acquire NOTIFY connection: %v", err)
			conn = nil
		}
	}

	execNotify := func(channel, payload string) {
		if conn == nil {
			reconnect()
		}
		if conn == nil {
			logger.Errorf(context.Background(), "No NOTIFY connection available, dropping %s notification", channel)
			return
		}
		if _, err := conn.ExecContext(context.Background(), "SELECT pg_notify($1, $2)", channel, payload); err != nil {
			logger.Errorf(context.Background(), "Failed to NOTIFY %s: %v", channel, err)
			if isConnError(err) {
				reconnect()
			}
		}
	}

	drainAndExec := func(channel, firstPayload string, ch <-chan string) {
		execNotify(channel, firstPayload)
		for {
			select {
			case payload, ok := <-ch:
				if !ok {
					return
				}
				execNotify(channel, payload)
			default:
				return
			}
		}
	}

	for {
		select {
		case payload, ok := <-r.actionNotifyCh:
			if !ok {
				return
			}
			drainAndExec("action_updates", payload, r.actionNotifyCh)
		case payload, ok := <-r.runNotifyCh:
			if !ok {
				return
			}
			drainAndExec("run_updates", payload, r.runNotifyCh)
		}
	}
}

// notifyActionUpdate sends a notification about an action update via the
// dedicated notify channel, avoiding connection pool contention.
func (r *actionRepo) notifyActionUpdate(ctx context.Context, actionID *common.ActionIdentifier) {
	payload := fmt.Sprintf("%s/%s/%s/%s",
		actionID.Run.Project, actionID.Run.Domain, actionID.Run.Name, actionID.Name)

	select {
	case r.actionNotifyCh <- payload:
	case <-ctx.Done():
		logger.Errorf(ctx, "Action NOTIFY send cancelled for %s: %v", payload, ctx.Err())
	}
}
