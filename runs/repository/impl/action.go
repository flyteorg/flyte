package impl

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/flytestdlib/database"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// actionRepo implements actionRepo interface using PostgreSQL/SQLite
type actionRepo struct {
	db         *gorm.DB
	isPostgres bool
	pgConfig   database.PostgresConfig
	listener   *pq.Listener

	// Subscriber management for LISTEN/NOTIFY
	runSubscribers    map[chan string]bool
	actionSubscribers map[chan string]bool
	mu                sync.RWMutex

	// Dedicated channels for async NOTIFY to avoid pool contention
	actionNotifyCh chan string
	runNotifyCh    chan string
}

const rootActionName = "a0"

// NewActionRepo creates a new PostgreSQL/SQLite repository
func NewActionRepo(db *gorm.DB, dbConfig database.DbConfig) interfaces.ActionRepo {
	// Detect database type
	dbName := db.Name()
	isPostgres := dbName == "postgres"

	repo := &actionRepo{
		db:                db,
		isPostgres:        isPostgres,
		pgConfig:          dbConfig.Postgres,
		runSubscribers:    make(map[chan string]bool),
		actionSubscribers: make(map[chan string]bool),
	}

	// Start LISTEN/NOTIFY for PostgreSQL
	if isPostgres {
		repo.actionNotifyCh = make(chan string, 256)
		repo.runNotifyCh = make(chan string, 256)
		go repo.startPostgresListener()
		go repo.startNotifyLoop()
	}

	return repo
}

// CreateRun creates a new run (root action with parent_action_name = null)
func (r *actionRepo) CreateRun(ctx context.Context, req *workflow.CreateRunRequest, inputUri, runOutputBase string) (*models.Run, error) {
	// Determine run ID
	var runID *common.RunIdentifier
	switch id := req.Id.(type) {
	case *workflow.CreateRunRequest_RunId:
		runID = id.RunId
	default:
		return nil, fmt.Errorf("invalid run ID type")
	}

	// Build ActionSpec from CreateRunRequest
	actionSpec := &workflow.ActionSpec{
		ActionId: &common.ActionIdentifier{
			Run:  runID,
			Name: rootActionName,
		},
		ParentActionName: nil, // NULL for root actions
		RunSpec:          req.RunSpec,
		InputUri:         inputUri + "/inputs.pb",
		RunOutputBase:    runOutputBase,
	}

	// Set the task spec based on the request
	switch taskSpec := req.Task.(type) {
	case *workflow.CreateRunRequest_TaskSpec:
		actionSpec.Spec = &workflow.ActionSpec_Task{
			Task: &workflow.TaskAction{
				Spec: taskSpec.TaskSpec,
			},
		}
	case *workflow.CreateRunRequest_TaskId:
		actionSpec.Spec = &workflow.ActionSpec_Task{
			Task: &workflow.TaskAction{
				Id: taskSpec.TaskId,
			},
		}
	}

	// Serialize the ActionSpec to binary protobuf
	actionSpecBytes, err := proto.Marshal(actionSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal action spec: %w", err)
	}

	// Build RunInfo with storage URIs and task spec digest
	info := &workflow.RunInfo{
		InputsUri: inputUri,
	}

	// Store task spec separately and record its digest
	if taskSpec := actionSpec.GetTask().GetSpec(); taskSpec != nil {
		taskSpecModel, err := models.NewTaskSpecModel(ctx, taskSpec)
		if err != nil {
			return nil, fmt.Errorf("failed to create task spec model: %w", err)
		}
		if taskSpecModel != nil {
			if err := r.db.WithContext(ctx).
				Clauses(clause.OnConflict{DoNothing: true}).
				Create(taskSpecModel).Error; err != nil {
				logger.Warnf(ctx, "CreateRun: failed to store task spec: %v", err)
			} else {
				info.TaskSpecDigest = taskSpecModel.Digest
			}
		}
	}

	detailedInfo, err := proto.Marshal(info)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal run info: %w", err)
	}

	// Marshal RunSpec if present
	var runSpecBytes []byte
	if req.RunSpec != nil {
		runSpecBytes, err = proto.Marshal(req.RunSpec)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal run spec: %w", err)
		}
	}

	// Extract metadata columns from action spec
	meta := extractActionMetadata(actionSpec)

	// Create root action (represents the run)
	run := &models.Run{
		Org:              runID.Org,
		Project:          runID.Project,
		Domain:           runID.Domain,
		RunName:          runID.Name,
		Name:             rootActionName,
		ParentActionName: newNullString(""), // NULL for root actions/runs
		Phase:            int32(common.ActionPhase_ACTION_PHASE_QUEUED),
		ActionType:       meta.ActionType,
		TaskOrg:          meta.TaskOrg,
		TaskProject:      meta.TaskProject,
		TaskDomain:       meta.TaskDomain,
		TaskName:         meta.TaskName,
		TaskVersion:      meta.TaskVersion,
		TaskType:         meta.TaskType,
		TaskShortName:    meta.TaskShortName,
		FunctionName:     meta.FunctionName,
		EnvironmentName:  meta.EnvironmentName,
		ActionSpec:       actionSpecBytes,
		ActionDetails:    []byte("{}"), // Empty details initially
		DetailedInfo:     detailedInfo,
		RunSpec:          runSpecBytes,
		Attempts:         1,
	}

	if err := r.db.WithContext(ctx).Create(run).Error; err != nil {
		return nil, fmt.Errorf("failed to create run: %w", err)
	}

	logger.Infof(ctx, "Created run: %s/%s/%s/%s (ID: %d)",
		run.Org, run.Project, run.Domain, run.Name, run.ID)

	// Notify subscribers of run creation
	r.notifyRunUpdate(ctx, runID)

	return run, nil
}

// GetRun retrieves a run by identifier
func (r *actionRepo) GetRun(ctx context.Context, runID *common.RunIdentifier) (*models.Run, error) {
	var run models.Run
	result := r.db.WithContext(ctx).
		Where("org = ? AND project = ? AND domain = ? AND run_name = ? AND parent_action_name IS NULL",
			runID.Org, runID.Project, runID.Domain, runID.Name).
		First(&run)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("run not found: %s/%s/%s/%s",
				runID.Org, runID.Project, runID.Domain, runID.Name)
		}
		return nil, fmt.Errorf("failed to get run: %w", result.Error)
	}

	return &run, nil
}

// ListRuns lists runs with pagination
func (r *actionRepo) ListRuns(ctx context.Context, req *workflow.ListRunsRequest) ([]*models.Run, string, error) {
	query := r.db.WithContext(ctx).Model(&models.Run{}).
		Where("parent_action_name IS NULL") // Only root actions (runs)

	// Apply scope filters
	switch scope := req.ScopeBy.(type) {
	case *workflow.ListRunsRequest_Org:
		query = query.Where("org = ?", scope.Org)
	case *workflow.ListRunsRequest_ProjectId:
		query = query.Where("org = ? AND project = ? AND domain = ?",
			scope.ProjectId.Organization, scope.ProjectId.Name, scope.ProjectId.Domain)
	}

	// Apply pagination according to token and limit from requests.
	limit := 50
	if req.Request != nil {
		if req.Request.Token != "" {
			tokenID, err := strconv.ParseUint(req.Request.Token, 10, 64)
			if err != nil {
				return nil, "", fmt.Errorf("invalid pagination token: %w", err)
			}
			query = query.Where("id < ?", tokenID)
		}

		if req.Request.Limit > 0 {
			limit = int(req.Request.Limit)
		}
	}

	var runs []*models.Run
	result := query.
		Order("created_at DESC").
		Limit(limit + 1). // Fetch one extra to determine if there are more
		Find(&runs)

	if result.Error != nil {
		return nil, "", fmt.Errorf("failed to list runs: %w", result.Error)
	}

	// Determine next token
	var nextToken string
	if len(runs) > limit {
		runs = runs[:limit]
		nextToken = fmt.Sprintf("%d", runs[len(runs)-1].ID)
	}

	return runs, nextToken, nil
}

// AbortRun aborts a run and all its actions
func (r *actionRepo) AbortRun(ctx context.Context, runID *common.RunIdentifier, reason string, abortedBy *common.EnrichedIdentity) error {
	now := time.Now()
	updates := map[string]interface{}{
		"phase":               int32(common.ActionPhase_ACTION_PHASE_ABORTED),
		"updated_at":          now,
		"abort_requested_at":  now,
		"abort_attempt_count": 0,
		"abort_reason":        reason,
	}

	result := r.db.WithContext(ctx).
		Model(&models.Run{}).
		Where("org = ? AND project = ? AND domain = ? AND run_name = ? AND parent_action_name IS NULL",
			runID.Org, runID.Project, runID.Domain, runID.Name).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to abort run: %w", result.Error)
	}

	// Notify run subscribers.
	r.notifyRunUpdate(ctx, runID)

	logger.Infof(ctx, "Aborted run: %s/%s/%s/%s", runID.Org, runID.Project, runID.Domain, runID.Name)
	return nil
}

// InsertEvents inserts a batch of action events, ignoring duplicates (same PK = idempotent).
func (r *actionRepo) InsertEvents(ctx context.Context, events []*models.ActionEvent) error {
	if len(events) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&events).Error; err != nil {
		return err
	}

	// Notify subscribers so watchers see new events (e.g. log context becoming available).
	notified := make(map[string]bool)
	for _, e := range events {
		actionID := &common.ActionIdentifier{
			Run: &common.RunIdentifier{
				Org:     e.Org,
				Project: e.Project,
				Domain:  e.Domain,
				Name:    e.RunName,
			},
			Name: e.Name,
		}
		key := e.Org + "/" + e.Project + "/" + e.Domain + "/" + e.RunName + "/" + e.Name
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
	result := r.db.WithContext(ctx).
		Where("org = ? AND project = ? AND domain = ? AND run_name = ? AND name = ?",
			actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain, actionID.Run.Name, actionID.Name).
		Order("attempt ASC, phase ASC, version ASC").
		Limit(limit).
		Find(&events)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list action events: %w", result.Error)
	}
	return events, nil
}

// GetLatestEventByAttempt returns the most recent event for a given attempt,
// ordered by version descending, without deserializing all events.
func (r *actionRepo) GetLatestEventByAttempt(ctx context.Context, actionID *common.ActionIdentifier, attempt uint32) (*models.ActionEvent, error) {
	var event models.ActionEvent
	result := r.db.WithContext(ctx).
		Where("org = ? AND project = ? AND domain = ? AND run_name = ? AND name = ? AND attempt = ?",
			actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain, actionID.Run.Name, actionID.Name, attempt).
		Order("phase DESC, version DESC").
		First(&event)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("event not found for attempt %d: %w", attempt, gorm.ErrRecordNotFound)
		}
		return nil, fmt.Errorf("failed to get latest event for attempt %d: %w", attempt, result.Error)
	}
	return &event, nil
}

// CreateAction creates a new action
func (r *actionRepo) CreateAction(ctx context.Context, actionSpec *workflow.ActionSpec, detailedInfo []byte) (*models.Action, error) {
	// Serialize action spec
	actionSpecBytes, err := proto.Marshal(actionSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal action spec: %w", err)
	}

	// Determine parent action name
	var parentActionName string
	if actionSpec.ParentActionName != nil {
		parentActionName = *actionSpec.ParentActionName
	}

	// Extract metadata columns from action spec
	meta := extractActionMetadata(actionSpec)

	action := &models.Action{
		Org:              actionSpec.ActionId.Run.Org,
		Project:          actionSpec.ActionId.Run.Project,
		Domain:           actionSpec.ActionId.Run.Domain,
		RunName:          actionSpec.ActionId.Run.Name,
		Name:             actionSpec.ActionId.Name,
		ParentActionName: newNullString(parentActionName),
		Phase:            int32(common.ActionPhase_ACTION_PHASE_QUEUED),
		ActionType:       meta.ActionType,
		ActionGroup:      newNullString(actionSpec.GetGroup()),
		TaskOrg:          meta.TaskOrg,
		TaskProject:      meta.TaskProject,
		TaskDomain:       meta.TaskDomain,
		TaskName:         meta.TaskName,
		TaskVersion:      meta.TaskVersion,
		TaskType:         meta.TaskType,
		TaskShortName:    meta.TaskShortName,
		FunctionName:     meta.FunctionName,
		EnvironmentName:  meta.EnvironmentName,
		ActionSpec:       actionSpecBytes,
		ActionDetails:    []byte("{}"), // Empty details initially
		DetailedInfo:     detailedInfo,
		Attempts:         1,
	}

	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(action)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create action: %w", result.Error)
	}

	// If no rows were affected, the action already exists — fetch and return it.
	if result.RowsAffected == 0 {
		existing, err := r.GetAction(ctx, actionSpec.ActionId)
		if err != nil {
			return nil, fmt.Errorf("failed to get existing action: %w", err)
		}
		return existing, nil
	}

	logger.Infof(ctx, "Created action: %s (ID: %d)", action.Name, action.ID)

	// Notify subscribers of action creation
	r.notifyActionUpdate(ctx, actionSpec.ActionId)

	return action, nil
}

// GetAction retrieves an action by identifier
func (r *actionRepo) GetAction(ctx context.Context, actionID *common.ActionIdentifier) (*models.Action, error) {
	var action models.Action
	result := r.db.WithContext(ctx).
		Where("org = ? AND project = ? AND domain = ? AND run_name = ? AND name = ?",
			actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain, actionID.Run.Name, actionID.Name).
		First(&action)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("action not found")
		}
		return nil, fmt.Errorf("failed to get action: %w", result.Error)
	}

	return &action, nil
}

// ListActions lists actions for a run
func (r *actionRepo) ListActions(ctx context.Context, runID *common.RunIdentifier, limit int, token string) ([]*models.Action, string, error) {
	if limit == 0 {
		limit = 100
	}

	query := r.db.WithContext(ctx).Model(&models.Action{}).
		Where("org = ? AND project = ? AND domain = ? AND run_name = ?",
			runID.Org, runID.Project, runID.Domain, runID.Name)

	// Apply pagination token
	if token != "" {
		query = query.Where("id > ?", token)
	}

	var actions []*models.Action
	result := query.
		Order("id ASC").
		Limit(limit + 1).
		Find(&actions)

	if result.Error != nil {
		return nil, "", fmt.Errorf("failed to list actions: %w", result.Error)
	}

	// Determine next token
	var nextToken string
	if len(actions) > limit {
		actions = actions[:limit]
		nextToken = fmt.Sprintf("%d", actions[len(actions)-1].ID)
	}

	return actions, nextToken, nil
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
	updates := map[string]interface{}{
		"phase":        phase,
		"attempts":     attempts,
		"cache_status": cacheStatus,
		"updated_at":   time.Now(),
	}

	if endTime != nil {
		if r.isPostgres {
			// Only set ended_at if not already set, clamped to at least created_at.
			updates["ended_at"] = gorm.Expr("COALESCE(ended_at, GREATEST(?, created_at))", *endTime)
			updates["duration_ms"] = gorm.Expr(
				"EXTRACT(EPOCH FROM (COALESCE(ended_at, GREATEST(?, created_at)) - created_at)) * 1000", *endTime)
		} else {
			// SQLite: only set ended_at if not already set.
			updates["ended_at"] = gorm.Expr("COALESCE(ended_at, MAX(?, created_at))", *endTime)
			updates["duration_ms"] = gorm.Expr(
				"CAST((julianday(COALESCE(ended_at, MAX(?, created_at))) - julianday(created_at)) * 86400000 AS INTEGER)", *endTime)
		}
	}

	// Only move the phase forward or re-apply the same phase — never downgrade.
	result := r.db.WithContext(ctx).
		Model(&models.Action{}).
		Where("org = ? AND project = ? AND domain = ? AND run_name = ? AND name = ? AND phase <= ?",
			actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain, actionID.Run.Name, actionID.Name, phase).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	// Notify subscribers of the action update
	r.notifyActionUpdate(ctx, actionID)

	return nil
}

// AbortAction aborts a specific action
func (r *actionRepo) AbortAction(ctx context.Context, actionID *common.ActionIdentifier, reason string, abortedBy *common.EnrichedIdentity) error {
	now := time.Now()
	updates := map[string]interface{}{
		"phase":               int32(common.ActionPhase_ACTION_PHASE_ABORTED),
		"updated_at":          now,
		"abort_requested_at":  now,
		"abort_attempt_count": 0,
		"abort_reason":        reason,
	}

	result := r.db.WithContext(ctx).
		Model(&models.Action{}).
		Where("org = ? AND project = ? AND domain = ? AND run_name = ? AND name = ?",
			actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain, actionID.Run.Name, actionID.Name).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to abort action: %w", result.Error)
	}

	// Notify action subscribers.
	r.notifyActionUpdate(ctx, actionID)

	logger.Infof(ctx, "Aborted action: %s", actionID.Name)
	return nil
}

// ListPendingAborts returns all actions that have abort_requested_at set (i.e. awaiting pod termination).
func (r *actionRepo) ListPendingAborts(ctx context.Context) ([]*models.Action, error) {
	var actions []*models.Action
	result := r.db.WithContext(ctx).
		Where("abort_requested_at IS NOT NULL").
		Find(&actions)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list pending aborts: %w", result.Error)
	}
	return actions, nil
}

// MarkAbortAttempt increments abort_attempt_count and returns the new value.
// Called by the reconciler before each actionsClient.Abort call.
func (r *actionRepo) MarkAbortAttempt(ctx context.Context, actionID *common.ActionIdentifier) (int, error) {
	var action models.Action
	result := r.db.WithContext(ctx).
		Model(&action).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "abort_attempt_count"}}}).
		Where("org = ? AND project = ? AND domain = ? AND run_name = ? AND name = ?",
			actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain, actionID.Run.Name, actionID.Name).
		Updates(map[string]interface{}{
			"abort_attempt_count": gorm.Expr("abort_attempt_count + 1"),
			"updated_at":          time.Now(),
		})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to mark abort attempt: %w", result.Error)
	}
	return action.AbortAttemptCount, nil
}

// ClearAbortRequest clears abort_requested_at (and resets counters) once the pod is confirmed terminated.
func (r *actionRepo) ClearAbortRequest(ctx context.Context, actionID *common.ActionIdentifier) error {
	result := r.db.WithContext(ctx).
		Model(&models.Action{}).
		Where("org = ? AND project = ? AND domain = ? AND run_name = ? AND name = ?",
			actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain, actionID.Run.Name, actionID.Name).
		Updates(map[string]interface{}{
			"abort_requested_at":  nil,
			"abort_attempt_count": 0,
			"abort_reason":        nil,
			"updated_at":          time.Now(),
		})
	if result.Error != nil {
		return fmt.Errorf("failed to clear abort request: %w", result.Error)
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

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	// Extract phase if present
	if phase, ok := stateObj["phase"].(string); ok {
		updates["phase"] = phase
		logger.Infof(ctx, "Updating action %s phase to %s", actionID.Name, phase)
	}

	// Store state in ActionDetails JSON
	// For now, we'll replace the entire ActionDetails with the state
	// In a full implementation, we'd merge it with existing ActionDetails
	updates["action_details"] = []byte(state)

	result := r.db.WithContext(ctx).
		Model(&models.Action{}).
		Where("org = ? AND project = ? AND domain = ? AND run_name = ? AND name = ?",
			actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain, actionID.Run.Name, actionID.Name).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update action state: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("action not found: %s/%s/%s/%s",
			actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain, actionID.Name)
	}

	// Notify subscribers of the update
	r.notifyActionUpdate(ctx, actionID)

	return nil
}

// GetActionState retrieves the state of an action
func (r *actionRepo) GetActionState(ctx context.Context, actionID *common.ActionIdentifier) (string, error) {
	var action models.Action
	result := r.db.WithContext(ctx).
		Select("action_details").
		Where("org = ? AND project = ? AND domain = ? AND run_name = ? AND name = ?",
			actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain, actionID.Run.Name, actionID.Name).
		First(&action)

	if result.Error != nil {
		return "", fmt.Errorf("failed to get action state: %w", result.Error)
	}

	// Extract state from ActionDetails JSON
	// For now, return the whole ActionDetails JSON
	return string(action.ActionDetails), nil
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

// WatchRunUpdates watches for run updates (simplified polling implementation)
func (r *actionRepo) WatchRunUpdates(ctx context.Context, runID *common.RunIdentifier, updates chan<- *models.Run, errs chan<- error) {
	if r.isPostgres {
		// PostgreSQL: Use LISTEN/NOTIFY with dedicated channel for this watcher
		runKey := fmt.Sprintf("%s/%s/%s/%s", runID.Org, runID.Project, runID.Domain, runID.Name)
		notifCh := make(chan string, 100)

		// Register as subscriber
		r.mu.Lock()
		r.runSubscribers[notifCh] = true
		r.mu.Unlock()

		// Unregister on exit
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
				// Check if this notification is for the run we're watching
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
	} else {
		// SQLite: Use polling
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		var lastUpdated time.Time

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				run, err := r.GetRun(ctx, runID)
				if err != nil {
					errs <- err
					return
				}

				if run.UpdatedAt.After(lastUpdated) {
					lastUpdated = run.UpdatedAt
					updates <- run
				}
			}
		}
	}
}

// WatchAllRunUpdates watches for all run updates (not filtered by runID)
func (r *actionRepo) WatchAllRunUpdates(ctx context.Context, updates chan<- *models.Run, errs chan<- error) {
	if r.isPostgres {
		// PostgreSQL: Use LISTEN/NOTIFY with dedicated channel for this watcher
		notifCh := make(chan string, 100)

		// Register as subscriber
		r.mu.Lock()
		r.runSubscribers[notifCh] = true
		r.mu.Unlock()

		// Unregister on exit
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
				// Parse notification payload: org/project/domain/run
				parts := strings.Split(notifPayload, "/")
				if len(parts) != 4 {
					logger.Warnf(ctx, "Invalid run notification payload: %s", notifPayload)
					continue
				}

				runID := &common.RunIdentifier{
					Org:     parts[0],
					Project: parts[1],
					Domain:  parts[2],
					Name:    parts[3],
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
	} else {
		// SQLite: Use polling for all runs
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		lastCheck := time.Now()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Query runs updated since last check
				var runs []*models.Run
				if err := r.db.WithContext(ctx).
					Where("updated_at > ? AND parent_action_name IS NULL", lastCheck).
					Find(&runs).Error; err != nil {
					errs <- err
					return
				}

				for _, run := range runs {
					updates <- run
				}

				lastCheck = time.Now()
			}
		}
	}
}

// WatchActionUpdates watches for action updates
func (r *actionRepo) WatchActionUpdates(ctx context.Context, runID *common.RunIdentifier, updates chan<- *models.Action, errs chan<- error) {
	if r.isPostgres {
		// PostgreSQL: Use LISTEN/NOTIFY with dedicated channel for this watcher
		runPrefix := fmt.Sprintf("%s/%s/%s/%s/", runID.Org, runID.Project, runID.Domain, runID.Name)
		notifCh := make(chan string, 100)

		// Register as subscriber
		r.mu.Lock()
		r.actionSubscribers[notifCh] = true
		r.mu.Unlock()

		// Unregister on exit
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
	} else {
		// SQLite: Use polling
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		lastCheck := time.Now()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Query actions updated since last check
				var actions []*models.Action
				if err := r.db.WithContext(ctx).
					Where("org = ? AND project = ? AND domain = ? AND updated_at > ?",
						runID.Org, runID.Project, runID.Domain, lastCheck).
					Find(&actions).Error; err != nil {
					errs <- err
					return
				}

				for _, action := range actions {
					updates <- action
				}

				lastCheck = time.Now()
			}
		}
	}
}

// startPostgresListener starts the PostgreSQL LISTEN/NOTIFY listener
func (r *actionRepo) startPostgresListener() {
	// Get the underlying SQL DB
	sqlDB, err := r.db.DB()
	if err != nil {
		logger.Errorf(context.Background(), "Failed to get SQL DB: %v", err)
		return
	}

	// Get the connection string
	var connStr string
	row := sqlDB.QueryRow("SELECT current_database()")
	var dbName string
	if err := row.Scan(&dbName); err != nil {
		logger.Errorf(context.Background(), "Failed to get database name: %v", err)
		return
	}

	// Build connection string from the database config
	pgCfg := r.pgConfig
	pgCfg.DbName = dbName
	connStr = database.GetPostgresDsn(context.Background(), pgCfg)

	r.listener = pq.NewListener(connStr, 10*time.Second, time.Minute, func(ev pq.ListenerEventType, err error) {
		if err != nil {
			logger.Errorf(context.Background(), "Listener error: %v", err)
		}
	})

	// Listen to channels
	if err := r.listener.Listen("run_updates"); err != nil {
		logger.Errorf(context.Background(), "Failed to listen to run_updates: %v", err)
		return
	}

	if err := r.listener.Listen("action_updates"); err != nil {
		logger.Errorf(context.Background(), "Failed to listen to action_updates: %v", err)
		return
	}

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
						logger.Warnf(context.Background(), "Action subscriber channel full, dropping notification")
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
// dedicated notify channel, avoiding GORM connection pool contention.
func (r *actionRepo) notifyRunUpdate(ctx context.Context, runID *common.RunIdentifier) {
	if !r.isPostgres {
		return
	}

	payload := fmt.Sprintf("%s/%s/%s/%s", runID.Org, runID.Project, runID.Domain, runID.Name)

	select {
	case r.runNotifyCh <- payload:
	default:
		logger.Warnf(ctx, "Run NOTIFY channel full, dropping notification for %s", payload)
	}
}

// ListRootActions lists root actions (runs) matching scope and date filters.
func (r *actionRepo) ListRootActions(ctx context.Context, org, project, domain string, startDate, endDate *time.Time, limit int) ([]*models.Action, error) {
	query := r.db.WithContext(ctx).Model(&models.Action{}).
		Where("parent_action_name IS NULL")

	if org != "" {
		query = query.Where("org = ?", org)
	}
	if project != "" {
		query = query.Where("project = ?", project)
	}
	if domain != "" {
		query = query.Where("domain = ?", domain)
	}
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}
	if limit <= 0 {
		limit = 1000
	}

	var actions []*models.Action
	result := query.Order("created_at DESC").Limit(limit).Find(&actions)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list root actions: %w", result.Error)
	}
	return actions, nil
}

// startNotifyLoop runs a dedicated goroutine that sends all NOTIFY commands
// over a single persistent connection, avoiding GORM connection pool contention.
func (r *actionRepo) startNotifyLoop() {
	sqlDB, err := r.db.DB()
	if err != nil {
		logger.Errorf(context.Background(), "Failed to get SQL DB for NOTIFY loop: %v", err)
		return
	}

	conn, err := sqlDB.Conn(context.Background())
	if err != nil {
		logger.Errorf(context.Background(), "Failed to acquire dedicated NOTIFY connection: %v", err)
		return
	}
	defer conn.Close()

	reconnect := func() bool {
		conn.Close()
		conn, err = sqlDB.Conn(context.Background())
		if err != nil {
			logger.Errorf(context.Background(), "Failed to re-acquire NOTIFY connection: %v", err)
			return false
		}
		return true
	}

	// drainAndExec collects all immediately available payloads from a channel
	// and sends them along with the initial payload in a single batched SQL call.
	drainAndExec := func(channel, firstPayload string, ch <-chan string) {
		var b strings.Builder
		b.WriteString(fmt.Sprintf("SELECT pg_notify('%s', '%s')", channel, firstPayload))

		// Drain any additional queued payloads to batch them in one round-trip.
	drain:
		for {
			select {
			case payload, ok := <-ch:
				if !ok {
					break drain
				}
				b.WriteString(fmt.Sprintf(", pg_notify('%s', '%s')", channel, payload))
			default:
				break drain
			}
		}

		if _, err := conn.ExecContext(context.Background(), b.String()); err != nil {
			logger.Errorf(context.Background(), "Failed to NOTIFY %s: %v", channel, err)
			reconnect()
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
// dedicated notify channel, avoiding GORM connection pool contention.
func (r *actionRepo) notifyActionUpdate(ctx context.Context, actionID *common.ActionIdentifier) {
	if !r.isPostgres {
		return
	}

	payload := fmt.Sprintf("%s/%s/%s/%s/%s",
		actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain, actionID.Run.Name, actionID.Name)

	select {
	case r.actionNotifyCh <- payload:
	default:
		logger.Warnf(ctx, "Action NOTIFY channel full, dropping notification for %s", payload)
	}
}

// actionMeta holds metadata columns extracted from an ActionSpec.
type actionMeta struct {
	ActionType      int32
	TaskOrg         sql.NullString
	TaskProject     sql.NullString
	TaskDomain      sql.NullString
	TaskName        sql.NullString
	TaskVersion     sql.NullString
	TaskType        string
	TaskShortName   sql.NullString
	FunctionName    string
	EnvironmentName sql.NullString
}

func newNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// extractActionMetadata extracts metadata columns from an ActionSpec proto.
func extractActionMetadata(spec *workflow.ActionSpec) actionMeta {
	var m actionMeta
	switch s := spec.GetSpec().(type) {
	case *workflow.ActionSpec_Task:
		m.ActionType = int32(workflow.ActionType_ACTION_TYPE_TASK)
		// TaskAction.Id takes precedence; fall back to TaskTemplate.Id
		if id := s.Task.GetId(); id != nil {
			m.TaskOrg = newNullString(id.GetOrg())
			m.TaskProject = newNullString(id.GetProject())
			m.TaskDomain = newNullString(id.GetDomain())
			m.TaskName = newNullString(id.GetName())
			m.TaskVersion = newNullString(id.GetVersion())
			m.FunctionName = id.GetName()
			m.TaskShortName = newNullString(id.GetName())
		} else if tmplID := s.Task.GetSpec().GetTaskTemplate().GetId(); tmplID != nil {
			m.TaskOrg = newNullString(tmplID.GetOrg())
			m.TaskProject = newNullString(tmplID.GetProject())
			m.TaskDomain = newNullString(tmplID.GetDomain())
			m.TaskName = newNullString(tmplID.GetName())
			m.TaskVersion = newNullString(tmplID.GetVersion())
			m.FunctionName = tmplID.GetName()
			m.TaskShortName = newNullString(tmplID.GetName())
		}
		if taskSpec := s.Task.GetSpec(); taskSpec != nil {
			m.TaskType = taskSpec.GetTaskTemplate().GetType()
			if taskSpec.GetShortName() != "" {
				m.TaskShortName = newNullString(taskSpec.GetShortName())
			}
			if env := taskSpec.GetEnvironment(); env != nil && env.GetName() != "" {
				m.EnvironmentName = newNullString(env.GetName())
			}
		}
	case *workflow.ActionSpec_Trace:
		m.ActionType = int32(workflow.ActionType_ACTION_TYPE_TRACE)
		m.FunctionName = s.Trace.GetName()
	}
	return m
}
