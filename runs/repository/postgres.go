package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

// PostgresRepository implements Repository interface using PostgreSQL/SQLite
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository creates a new PostgreSQL/SQLite repository
func NewPostgresRepository(db *gorm.DB) Repository {
	return &PostgresRepository{db: db}
}

// CreateRun creates a new run
func (r *PostgresRepository) CreateRun(ctx context.Context, req *workflow.CreateRunRequest) (*Run, error) {
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

	// Serialize run spec
	runSpecBytes, err := json.Marshal(req.RunSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal run spec: %w", err)
	}

	run := &Run{
		Org:            runID.Org,
		Project:        runID.Project,
		Domain:         runID.Domain,
		Name:           runID.Name,
		RootActionName: "root", // Will be updated after creating root action
		RunSpec:        datatypes.JSON(runSpecBytes),
	}

	if err := r.db.WithContext(ctx).Create(run).Error; err != nil {
		return nil, fmt.Errorf("failed to create run: %w", err)
	}

	logger.Infof(ctx, "Created run: %s/%s/%s/%s (ID: %d)",
		run.Org, run.Project, run.Domain, run.Name, run.ID)

	return run, nil
}

// GetRun retrieves a run by identifier
func (r *PostgresRepository) GetRun(ctx context.Context, runID *common.RunIdentifier) (*Run, error) {
	var run Run
	result := r.db.WithContext(ctx).
		Where("org = ? AND project = ? AND domain = ? AND name = ?",
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
func (r *PostgresRepository) ListRuns(ctx context.Context, req *workflow.ListRunsRequest) ([]*Run, string, error) {
	query := r.db.WithContext(ctx).Model(&Run{})

	// Apply scope filters
	switch scope := req.ScopeBy.(type) {
	case *workflow.ListRunsRequest_Org:
		query = query.Where("org = ?", scope.Org)
	case *workflow.ListRunsRequest_ProjectId:
		query = query.Where("org = ? AND project = ? AND domain = ?",
			scope.ProjectId.Organization, scope.ProjectId.Name, scope.ProjectId.Domain)
	case *workflow.ListRunsRequest_TriggerName:
		query = query.Where("trigger_org = ? AND trigger_project = ? AND trigger_domain = ? AND trigger_name = ?",
			scope.TriggerName.Org, scope.TriggerName.Project, scope.TriggerName.Domain, scope.TriggerName.Name)
	}

	// Apply pagination
	limit := 50
	if req.Request != nil && req.Request.Limit > 0 {
		limit = int(req.Request.Limit)
	}

	var runs []*Run
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
func (r *PostgresRepository) AbortRun(ctx context.Context, runID *common.RunIdentifier, reason string, abortedBy *common.EnrichedIdentity) error {
	// Get the run first
	run, err := r.GetRun(ctx, runID)
	if err != nil {
		return err
	}

	// Update all actions in this run to aborted
	updates := map[string]interface{}{
		"phase":        "PHASE_ABORTED",
		"abort_reason": reason,
		"end_time":     time.Now(),
		"updated_at":   time.Now(),
	}

	if abortedBy != nil {
		// Extract principal from abortedBy
		if abortedBy.GetUser() != nil && abortedBy.GetUser().Id != nil && abortedBy.GetUser().Id.Subject != "" {
			updates["aborted_by_principal"] = abortedBy.GetUser().Id.Subject
		}
	}

	result := r.db.WithContext(ctx).
		Model(&Action{}).
		Where("run_id = ? AND phase NOT IN ?", run.ID, []string{"PHASE_SUCCEEDED", "PHASE_FAILED", "PHASE_ABORTED"}).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to abort run actions: %w", result.Error)
	}

	logger.Infof(ctx, "Aborted %d actions for run: %s/%s/%s/%s",
		result.RowsAffected, runID.Org, runID.Project, runID.Domain, runID.Name)

	return nil
}

// CreateAction creates a new action
func (r *PostgresRepository) CreateAction(ctx context.Context, runID uint, actionSpec *workflow.ActionSpec) (*Action, error) {
	action := &Action{
		Org:     actionSpec.ActionId.Run.Org,
		Project: actionSpec.ActionId.Run.Project,
		Domain:  actionSpec.ActionId.Run.Domain,
		RunName: actionSpec.ActionId.Run.Name,
		Name:    actionSpec.ActionId.Name,
		RunID:   runID,
		Phase:   "PHASE_QUEUED",
	}

	// Set parent action name if provided
	if actionSpec.ParentActionName != nil {
		action.ParentActionName = actionSpec.ParentActionName
	}

	// Set group
	if actionSpec.Group != "" {
		action.ActionGroup = &actionSpec.Group
	}

	// Determine action type and set metadata
	switch spec := actionSpec.Spec.(type) {
	case *workflow.ActionSpec_Task:
		action.ActionType = "task"

		// Serialize task spec
		taskSpecBytes, err := json.Marshal(spec.Task)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal task spec: %w", err)
		}
		action.TaskSpec = datatypes.JSON(taskSpecBytes)

		// Set task metadata if ID is provided
		if spec.Task.Id != nil {
			action.TaskOrg = &spec.Task.Id.Org
			action.TaskProject = &spec.Task.Id.Project
			action.TaskDomain = &spec.Task.Id.Domain
			action.TaskName = &spec.Task.Id.Name
			action.TaskVersion = &spec.Task.Id.Version
		}

	case *workflow.ActionSpec_Trace:
		action.ActionType = "trace"
		action.TraceName = &spec.Trace.Name

		// Serialize trace spec
		traceSpecBytes, err := json.Marshal(spec.Trace)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal trace spec: %w", err)
		}
		action.TraceSpec = datatypes.JSON(traceSpecBytes)

	case *workflow.ActionSpec_Condition:
		action.ActionType = "condition"
		action.ConditionName = &spec.Condition.Name

		// Serialize condition spec
		conditionSpecBytes, err := json.Marshal(spec.Condition)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal condition spec: %w", err)
		}
		action.ConditionSpec = datatypes.JSON(conditionSpecBytes)
	}

	// Set input/output paths
	if actionSpec.InputUri != "" {
		action.InputURI = &actionSpec.InputUri
	}
	if actionSpec.RunOutputBase != "" {
		action.RunOutputBase = &actionSpec.RunOutputBase
	}

	if err := r.db.WithContext(ctx).Create(action).Error; err != nil {
		return nil, fmt.Errorf("failed to create action: %w", err)
	}

	logger.Infof(ctx, "Created action: %s (ID: %d)", action.Name, action.ID)
	return action, nil
}

// GetAction retrieves an action by identifier
func (r *PostgresRepository) GetAction(ctx context.Context, actionID *common.ActionIdentifier) (*Action, error) {
	var action Action
	result := r.db.WithContext(ctx).
		Where("org = ? AND project = ? AND domain = ? AND run_name = ? AND name = ?",
			actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain,
			actionID.Run.Name, actionID.Name).
		First(&action)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("action not found")
		}
		return nil, fmt.Errorf("failed to get action: %w", result.Error)
	}

	return &action, nil
}

// GetActionWithAttempts retrieves an action with all its attempts
func (r *PostgresRepository) GetActionWithAttempts(ctx context.Context, actionID *common.ActionIdentifier) (*Action, error) {
	action, err := r.GetAction(ctx, actionID)
	if err != nil {
		return nil, err
	}

	// Load attempts (GORM will handle this with Preload if we set it up)
	var attempts []*ActionAttempt
	if err := r.db.WithContext(ctx).
		Where("action_id = ?", action.ID).
		Order("attempt_number ASC").
		Find(&attempts).Error; err != nil {
		return nil, fmt.Errorf("failed to load attempts: %w", err)
	}

	// For now, we're not attaching them to the model
	// In a real implementation, you'd add an Attempts field to Action
	logger.Infof(ctx, "Loaded action with %d attempts", len(attempts))

	return action, nil
}

// ListActions lists actions for a run
func (r *PostgresRepository) ListActions(ctx context.Context, runID *common.RunIdentifier, limit int, token string) ([]*Action, string, error) {
	// Get the run first to get its ID
	run, err := r.GetRun(ctx, runID)
	if err != nil {
		return nil, "", err
	}

	var actions []*Action
	query := r.db.WithContext(ctx).
		Where("run_id = ?", run.ID).
		Order("created_at ASC")

	if limit == 0 {
		limit = 50
	}

	result := query.Limit(limit + 1).Find(&actions)
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
func (r *PostgresRepository) UpdateActionPhase(ctx context.Context, actionID *common.ActionIdentifier, phase string, startTime, endTime *string) error {
	updates := map[string]interface{}{
		"phase":      phase,
		"updated_at": time.Now(),
	}

	if startTime != nil {
		if t, err := time.Parse(time.RFC3339, *startTime); err == nil {
			updates["start_time"] = t
		}
	}

	if endTime != nil {
		if t, err := time.Parse(time.RFC3339, *endTime); err == nil {
			updates["end_time"] = t
		}
	}

	result := r.db.WithContext(ctx).
		Model(&Action{}).
		Where("org = ? AND project = ? AND domain = ? AND run_name = ? AND name = ?",
			actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain,
			actionID.Run.Name, actionID.Name).
		Updates(updates)

	return result.Error
}

// AbortAction aborts a specific action
func (r *PostgresRepository) AbortAction(ctx context.Context, actionID *common.ActionIdentifier, reason string, abortedBy *common.EnrichedIdentity) error {
	updates := map[string]interface{}{
		"phase":        "PHASE_ABORTED",
		"abort_reason": reason,
		"end_time":     time.Now(),
		"updated_at":   time.Now(),
	}

	if abortedBy != nil && abortedBy.GetUser() != nil && abortedBy.GetUser().Id != nil {
		updates["aborted_by_principal"] = abortedBy.GetUser().Id.Subject
	}

	result := r.db.WithContext(ctx).
		Model(&Action{}).
		Where("org = ? AND project = ? AND domain = ? AND run_name = ? AND name = ?",
			actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain,
			actionID.Run.Name, actionID.Name).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to abort action: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("action not found")
	}

	logger.Infof(ctx, "Aborted action: %s", actionID.Name)
	return nil
}

// CreateActionAttempt creates a new action attempt
func (r *PostgresRepository) CreateActionAttempt(ctx context.Context, actionID uint, attemptNumber uint) (*ActionAttempt, error) {
	attempt := &ActionAttempt{
		ActionID:      actionID,
		AttemptNumber: attemptNumber,
		Phase:         "PHASE_QUEUED",
		StartTime:     time.Now(),
	}

	if err := r.db.WithContext(ctx).Create(attempt).Error; err != nil {
		return nil, fmt.Errorf("failed to create attempt: %w", err)
	}

	return attempt, nil
}

// GetLatestAttempt gets the latest attempt for an action
func (r *PostgresRepository) GetLatestAttempt(ctx context.Context, actionID uint) (*ActionAttempt, error) {
	var attempt ActionAttempt
	result := r.db.WithContext(ctx).
		Where("action_id = ?", actionID).
		Order("attempt_number DESC").
		First(&attempt)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no attempts found for action")
		}
		return nil, result.Error
	}

	return &attempt, nil
}

// UpdateAttempt updates an attempt
func (r *PostgresRepository) UpdateAttempt(ctx context.Context, attemptID uint, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.WithContext(ctx).
		Model(&ActionAttempt{}).
		Where("id = ?", attemptID).
		Updates(updates).Error
}

// AddClusterEvent adds a cluster event to an attempt
func (r *PostgresRepository) AddClusterEvent(ctx context.Context, attemptID uint, occurredAt string, message string) error {
	t, err := time.Parse(time.RFC3339, occurredAt)
	if err != nil {
		t = time.Now()
	}

	event := &ClusterEvent{
		AttemptID:  attemptID,
		OccurredAt: t,
		Message:    message,
	}

	return r.db.WithContext(ctx).Create(event).Error
}

// GetClusterEvents gets cluster events for an action attempt
func (r *PostgresRepository) GetClusterEvents(ctx context.Context, actionID *common.ActionIdentifier, attemptNumber uint) ([]*ClusterEvent, error) {
	action, err := r.GetAction(ctx, actionID)
	if err != nil {
		return nil, err
	}

	var attempt ActionAttempt
	if err := r.db.WithContext(ctx).
		Where("action_id = ? AND attempt_number = ?", action.ID, attemptNumber).
		First(&attempt).Error; err != nil {
		return nil, err
	}

	var events []*ClusterEvent
	if err := r.db.WithContext(ctx).
		Where("attempt_id = ?", attempt.ID).
		Order("occurred_at ASC").
		Find(&events).Error; err != nil {
		return nil, err
	}

	return events, nil
}

// AddPhaseTransition adds a phase transition to an attempt
func (r *PostgresRepository) AddPhaseTransition(ctx context.Context, attemptID uint, phase string, startTime string, endTime *string) error {
	t, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		t = time.Now()
	}

	transition := &PhaseTransition{
		AttemptID: attemptID,
		Phase:     phase,
		StartTime: t,
	}

	if endTime != nil {
		if et, err := time.Parse(time.RFC3339, *endTime); err == nil {
			transition.EndTime = &et
		}
	}

	return r.db.WithContext(ctx).Create(transition).Error
}

// WatchRunUpdates watches for run updates (simplified polling implementation)
func (r *PostgresRepository) WatchRunUpdates(ctx context.Context, runID *common.RunIdentifier, updates chan<- *Run, errs chan<- error) {
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

// WatchActionUpdates watches for action updates (simplified polling implementation)
func (r *PostgresRepository) WatchActionUpdates(ctx context.Context, runID *common.RunIdentifier, updates chan<- *Action, errs chan<- error) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	lastCheck := time.Now()

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

			var actions []*Action
			if err := r.db.WithContext(ctx).
				Where("run_id = ? AND updated_at > ?", run.ID, lastCheck).
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
