package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// actionRepo implements actionRepo interface using PostgreSQL/SQLite
type actionRepo struct {
	db         *gorm.DB
	isPostgres bool
	listener   *pq.Listener

	// Subscriber management for LISTEN/NOTIFY
	runSubscribers    map[chan string]bool
	actionSubscribers map[chan string]bool
	mu                sync.RWMutex
}

// NewActionRepo creates a new PostgreSQL/SQLite repository
func NewActionRepo(db *gorm.DB) interfaces.ActionRepo {
	// Detect database type
	dbName := db.Name()
	isPostgres := dbName == "postgres"

	repo := &actionRepo{
		db:                db,
		isPostgres:        isPostgres,
		runSubscribers:    make(map[chan string]bool),
		actionSubscribers: make(map[chan string]bool),
	}

	// Start LISTEN/NOTIFY for PostgreSQL
	if isPostgres {
		go repo.startPostgresListener()
	}

	return repo
}

// CreateRun creates a new run (root action with parent_action_name = null)
func (r *actionRepo) CreateRun(ctx context.Context, req *workflow.CreateRunRequest) (*models.Run, error) {
	// Determine run ID
	var runID *common.RunIdentifier
	switch id := req.Id.(type) {
	case *workflow.CreateRunRequest_RunId:
		runID = id.RunId
	case *workflow.CreateRunRequest_ProjectId:
		// Generate a run name (simplified - in production, use a better generator)
		runID = &common.RunIdentifier{
			Org:     id.ProjectId.Organization,
			Project: id.ProjectId.Name,
			Domain:  id.ProjectId.Domain,
			Name:    fmt.Sprintf("run-%d", time.Now().Unix()),
		}
	default:
		return nil, fmt.Errorf("invalid run ID type")
	}

	// Build ActionSpec from CreateRunRequest
	actionSpec := &workflow.ActionSpec{
		ActionId: &common.ActionIdentifier{
			Run:  runID,
			Name: runID.Name, // For root actions, action name = run name
		},
		ParentActionName: nil, // NULL for root actions
		RunSpec:          req.RunSpec,
		InputUri:         "", // TODO: build from inputs
		RunOutputBase:    "", // TODO: build output path
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

	// Serialize the ActionSpec to JSON
	actionSpecBytes, err := json.Marshal(actionSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal action spec: %w", err)
	}

	// Create root action (represents the run)
	run := &models.Run{
		Org:              runID.Org,
		Project:          runID.Project,
		Domain:           runID.Domain,
		Name:             runID.Name,
		ParentActionName: nil, // NULL for root actions/runs
		Phase:            "PHASE_QUEUED",
		ActionSpec:       datatypes.JSON(actionSpecBytes),
		ActionDetails:    datatypes.JSON([]byte("{}")), // Empty details initially
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
		Where("org = ? AND project = ? AND domain = ? AND name = ? AND parent_action_name IS NULL",
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

	// Apply pagination
	limit := 50
	if req.Request != nil && req.Request.Limit > 0 {
		limit = int(req.Request.Limit)
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
	// Update the run action to aborted
	updates := map[string]interface{}{
		"phase":      "PHASE_ABORTED",
		"updated_at": time.Now(),
	}

	result := r.db.WithContext(ctx).
		Model(&models.Run{}).
		Where("org = ? AND project = ? AND domain = ? AND name = ? AND parent_action_name IS NULL",
			runID.Org, runID.Project, runID.Domain, runID.Name).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to abort run: %w", result.Error)
	}

	// Notify subscribers
	r.notifyRunUpdate(ctx, runID)

	logger.Infof(ctx, "Aborted run: %s/%s/%s/%s", runID.Org, runID.Project, runID.Domain, runID.Name)
	return nil
}

// CreateAction creates a new action
func (r *actionRepo) CreateAction(ctx context.Context, runID uint, actionSpec *workflow.ActionSpec) (*models.Action, error) {
	// Serialize action spec
	actionSpecBytes, err := json.Marshal(actionSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal action spec: %w", err)
	}

	// Determine parent action name
	var parentActionName *string
	if actionSpec.ParentActionName != nil {
		parentActionName = actionSpec.ParentActionName
	}

	action := &models.Action{
		Org:              actionSpec.ActionId.Run.Org,
		Project:          actionSpec.ActionId.Run.Project,
		Domain:           actionSpec.ActionId.Run.Domain,
		Name:             actionSpec.ActionId.Name,
		ParentActionName: parentActionName,
		Phase:            "PHASE_QUEUED",
		ActionSpec:       datatypes.JSON(actionSpecBytes),
		ActionDetails:    datatypes.JSON([]byte("{}")), // Empty details initially
	}

	if err := r.db.WithContext(ctx).Create(action).Error; err != nil {
		return nil, fmt.Errorf("failed to create action: %w", err)
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
		Where("org = ? AND project = ? AND domain = ? AND name = ?",
			actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain, actionID.Name).
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
		Where("org = ? AND project = ? AND domain = ?",
							runID.Org, runID.Project, runID.Domain).
		Where("parent_action_name IS NOT NULL") // Exclude the root action/run itself

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

// UpdateActionPhase updates the phase of an action
func (r *actionRepo) UpdateActionPhase(ctx context.Context, actionID *common.ActionIdentifier, phase string, startTime, endTime *string) error {
	updates := map[string]interface{}{
		"phase":      phase,
		"updated_at": time.Now(),
	}

	result := r.db.WithContext(ctx).
		Model(&models.Action{}).
		Where("org = ? AND project = ? AND domain = ? AND name = ?",
			actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain, actionID.Name).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	// Notify subscribers of action update
	r.notifyActionUpdate(ctx, actionID)

	return nil
}

// AbortAction aborts a specific action
func (r *actionRepo) AbortAction(ctx context.Context, actionID *common.ActionIdentifier, reason string, abortedBy *common.EnrichedIdentity) error {
	updates := map[string]interface{}{
		"phase":      "PHASE_ABORTED",
		"updated_at": time.Now(),
	}

	result := r.db.WithContext(ctx).
		Model(&models.Action{}).
		Where("org = ? AND project = ? AND domain = ? AND name = ?",
			actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain, actionID.Name).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to abort action: %w", result.Error)
	}

	// Notify subscribers
	r.notifyActionUpdate(ctx, actionID)

	logger.Infof(ctx, "Aborted action: %s", actionID.Name)
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
	updates["action_details"] = datatypes.JSON([]byte(state))

	result := r.db.WithContext(ctx).
		Model(&models.Action{}).
		Where("org = ? AND project = ? AND domain = ? AND name = ?",
			actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain, actionID.Name).
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
		Where("org = ? AND project = ? AND domain = ? AND name = ?",
			actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain, actionID.Name).
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
		notifCh := make(chan string, 10)

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
						errs <- err
						return
					}
					updates <- run
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
		notifCh := make(chan string, 10)

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

				run, err := r.GetRun(ctx, runID)
				if err != nil {
					logger.Errorf(ctx, "Failed to get run from notification: %v", err)
					continue
				}
				updates <- run
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
		notifCh := make(chan string, 10)

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
				// Check if this notification is for an action in the run we're watching
				// Payload format: org/project/domain/run/action
				if len(notifPayload) > len(runPrefix) && notifPayload[:len(runPrefix)] == runPrefix {
					// Extract action name from payload
					actionName := notifPayload[len(runPrefix):]
					actionID := &common.ActionIdentifier{
						Run:  runID,
						Name: actionName,
					}

					action, err := r.GetAction(ctx, actionID)
					if err != nil {
						logger.Errorf(ctx, "Failed to get action from notification: %v", err)
						continue
					}
					updates <- action
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
					Where("org = ? AND project = ? AND domain = ? AND updated_at > ? AND parent_action_name IS NOT NULL",
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

	// Build connection string from the existing connection
	// This is a simplified approach - in production, get from config
	row = sqlDB.QueryRow("SHOW server_version")
	var version string
	_ = row.Scan(&version) // Ignore error, just need a connection

	// Create listener with a simple connection string
	// In production, get this from the database config
	connStr = "user=postgres password=mysecretpassword host=localhost port=5432 dbname=" + dbName + " sslmode=disable"

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
						// Channel full, skip this subscriber
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

// notifyRunUpdate sends a notification about a run update
func (r *actionRepo) notifyRunUpdate(ctx context.Context, runID *common.RunIdentifier) {
	if !r.isPostgres {
		return
	}

	payload := fmt.Sprintf("%s/%s/%s/%s", runID.Org, runID.Project, runID.Domain, runID.Name)

	// Execute NOTIFY
	sql := fmt.Sprintf("NOTIFY run_updates, '%s'", payload)
	if err := r.db.WithContext(ctx).Exec(sql).Error; err != nil {
		logger.Errorf(ctx, "Failed to NOTIFY run_updates: %v", err)
	}
}

// notifyActionUpdate sends a notification about an action update
func (r *actionRepo) notifyActionUpdate(ctx context.Context, actionID *common.ActionIdentifier) {
	if !r.isPostgres {
		return
	}

	payload := fmt.Sprintf("%s/%s/%s/%s/%s",
		actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain, actionID.Run.Name, actionID.Name)

	// Execute NOTIFY
	sql := fmt.Sprintf("NOTIFY action_updates, '%s'", payload)
	if err := r.db.WithContext(ctx).Exec(sql).Error; err != nil {
		logger.Errorf(ctx, "Failed to NOTIFY action_updates: %v", err)
	}
}
