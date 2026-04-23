package impl

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/flyteorg/flyte/v2/flytestdlib/database"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var testDbConfig = database.DbConfig{
	Postgres: database.PostgresConfig{
		Host:         "localhost",
		Port:         15432,
		DbName:       "flyte_runs_test",
		User:         "postgres",
		Password:     "postgres",
		ExtraOptions: "sslmode=disable",
	},
}

func setupActionDB(t *testing.T) *sqlx.DB {
	db := setupDB(t)
	t.Cleanup(func() {
		db.Exec("DELETE FROM action_events")
		db.Exec("DELETE FROM actions")
	})
	return db
}

func TestCreateRun(t *testing.T) {
	db := setupActionDB(t)
	defer func() { db.Exec("DELETE FROM actions") }()
	actionRepo, err := NewActionRepo(db, testDbConfig)
	require.NoError(t, err)
	ctx := context.Background()

	runID := &common.RunIdentifier{
		Org:     "org1",
		Project: "proj1",
		Domain:  "domain1",
		Name:    "run1",
	}
	runModel := &models.Run{
		Project: runID.Project,
		Domain:  runID.Domain,
		RunName: runID.Name,
		Name:    rootActionName,
		Phase:   int32(common.ActionPhase_ACTION_PHASE_QUEUED),
	}

	run, err := actionRepo.CreateAction(ctx, runModel, false)
	require.NoError(t, err)
	require.NotNil(t, run)
	assert.Equal(t, runID.Project, run.Project)
	assert.Equal(t, runID.Domain, run.Domain)
	assert.Equal(t, runID.Name, run.RunName)
	assert.Equal(t, "a0", run.Name)
	assert.Equal(t, int32(common.ActionPhase_ACTION_PHASE_QUEUED), run.Phase)

	// Attempt duplicate run create with same run name should return existing (idempotent)
	run2, err := actionRepo.CreateAction(ctx, runModel, false)
	require.NoError(t, err)
	assert.Equal(t, run.Name, run2.Name)
}

func TestUpdateActionPhasePersistsAttemptsAndCacheStatus(t *testing.T) {
	db := setupActionDB(t)
	defer func() { db.Exec("DELETE FROM actions") }()
	actionRepo, err := NewActionRepo(db, testDbConfig)
	require.NoError(t, err)
	ctx := context.Background()

	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     "org1",
			Project: "proj1",
			Domain:  "domain1",
			Name:    "run1",
		},
		Name: "action1",
	}

	_, err = actionRepo.CreateAction(ctx, models.NewActionModel(actionID), false)
	require.NoError(t, err)

	endTime := time.Now()
	err = actionRepo.UpdateActionPhase(
		ctx,
		actionID,
		common.ActionPhase_ACTION_PHASE_SUCCEEDED,
		3,
		core.CatalogCacheStatus_CACHE_HIT,
		&endTime,
	)
	require.NoError(t, err)

	action, err := actionRepo.GetAction(ctx, actionID)
	require.NoError(t, err)
	assert.Equal(t, int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED), action.Phase)
	assert.Equal(t, uint32(3), action.Attempts)
	assert.Equal(t, core.CatalogCacheStatus_CACHE_HIT, action.CacheStatus)
	assert.True(t, action.EndedAt.Valid)
}

func TestWatchActionUpdates_OnlyStreamsTargetAction(t *testing.T) {
	db := setupActionDB(t)
	defer func() { db.Exec("DELETE FROM actions") }()
	repo, err := NewActionRepo(db, testDbConfig)
	require.NoError(t, err)
	repoImpl := repo.(*actionRepo)

	runID := &common.RunIdentifier{
		Org:     "org1",
		Project: "proj1",
		Domain:  "domain1",
		Name:    "run1",
	}
	targetActionID := &common.ActionIdentifier{Run: runID, Name: "target"}
	otherActionID := &common.ActionIdentifier{Run: runID, Name: "other"}

	ctx := context.Background()

	// Start watcher before creating actions so we can deterministically
	// drain the creation notification and avoid a race where the async
	// NOTIFY arrives after the subscriber registers.
	watchCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	updates := make(chan *models.Action, 2)
	errs := make(chan error, 1)
	go repo.WatchActionUpdates(watchCtx, targetActionID, updates, errs)

	require.Eventually(t, func() bool {
		repoImpl.mu.RLock()
		defer repoImpl.mu.RUnlock()
		return len(repoImpl.actionSubscribers) > 0
	}, 2*time.Second, 10*time.Millisecond, "timed out waiting for watcher registration")

	_, err = repo.CreateAction(ctx, models.NewActionModel(targetActionID), false)
	require.NoError(t, err)
	_, err = repo.CreateAction(ctx, models.NewActionModel(otherActionID), false)
	require.NoError(t, err)

	// Drain the creation notification for the target action.
	select {
	case <-updates:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for creation notification")
	}

	// Update "other" — should NOT produce an update for "target".
	err = repo.UpdateActionPhase(ctx, otherActionID, common.ActionPhase_ACTION_PHASE_RUNNING, 1, core.CatalogCacheStatus_CACHE_DISABLED, nil)
	require.NoError(t, err)

	select {
	case action := <-updates:
		t.Fatalf("unexpected update for action %s", action.Name)
	case err := <-errs:
		require.NoError(t, err)
	case <-time.After(1200 * time.Millisecond):
	}

	// Update "target" — should produce an update.
	err = repo.UpdateActionPhase(ctx, targetActionID, common.ActionPhase_ACTION_PHASE_RUNNING, 1, core.CatalogCacheStatus_CACHE_DISABLED, nil)
	require.NoError(t, err)

	select {
	case action := <-updates:
		require.Equal(t, targetActionID.Name, action.Name)
	case err := <-errs:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for target action update")
	}
}

func TestUpdateActionPhase_AllowsRetryTransition(t *testing.T) {
	db := setupActionDB(t)
	defer func() { db.Exec("DELETE FROM actions") }()
	actionRepo, err := NewActionRepo(db, testDbConfig)
	require.NoError(t, err)
	ctx := context.Background()

	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     "org1",
			Project: "proj1",
			Domain:  "domain1",
			Name:    "run1",
		},
		Name: "action1",
	}

	_, err = actionRepo.CreateAction(ctx, models.NewActionModel(actionID), false)
	require.NoError(t, err)

	// Move to FAILED (terminal state)
	endTime := time.Now()
	err = actionRepo.UpdateActionPhase(ctx, actionID,
		common.ActionPhase_ACTION_PHASE_FAILED, 1,
		core.CatalogCacheStatus_CACHE_DISABLED, &endTime)
	require.NoError(t, err)

	action, err := actionRepo.GetAction(ctx, actionID)
	require.NoError(t, err)
	assert.Equal(t, int32(common.ActionPhase_ACTION_PHASE_FAILED), action.Phase)

	// Retry: transition from FAILED back to QUEUED — should succeed
	err = actionRepo.UpdateActionPhase(ctx, actionID,
		common.ActionPhase_ACTION_PHASE_QUEUED, 2,
		core.CatalogCacheStatus_CACHE_DISABLED, nil)
	require.NoError(t, err)

	action, err = actionRepo.GetAction(ctx, actionID)
	require.NoError(t, err)
	assert.Equal(t, int32(common.ActionPhase_ACTION_PHASE_QUEUED), action.Phase,
		"phase should transition from FAILED to QUEUED on retry")
	assert.Equal(t, uint32(2), action.Attempts)
}

func TestUpdateActionPhase_BlocksBackwardFromNonRetryable(t *testing.T) {
	db := setupActionDB(t)
	defer func() { db.Exec("DELETE FROM actions") }()
	actionRepo, err := NewActionRepo(db, testDbConfig)
	require.NoError(t, err)
	ctx := context.Background()

	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     "org1",
			Project: "proj1",
			Domain:  "domain1",
			Name:    "run1",
		},
		Name: "action-no-backward",
	}

	_, err = actionRepo.CreateAction(ctx, models.NewActionModel(actionID), false)
	require.NoError(t, err)

	// Move to RUNNING
	err = actionRepo.UpdateActionPhase(ctx, actionID,
		common.ActionPhase_ACTION_PHASE_RUNNING, 1,
		core.CatalogCacheStatus_CACHE_DISABLED, nil)
	require.NoError(t, err)

	// Try to downgrade from RUNNING to QUEUED — should be a no-op (phase guard)
	err = actionRepo.UpdateActionPhase(ctx, actionID,
		common.ActionPhase_ACTION_PHASE_QUEUED, 1,
		core.CatalogCacheStatus_CACHE_DISABLED, nil)
	require.NoError(t, err)

	action, err := actionRepo.GetAction(ctx, actionID)
	require.NoError(t, err)
	assert.Equal(t, int32(common.ActionPhase_ACTION_PHASE_RUNNING), action.Phase,
		"phase should not downgrade from RUNNING to QUEUED")
}

func TestUpdateActionPhase_BlocksBackwardFromSucceeded(t *testing.T) {
	db := setupActionDB(t)
	defer func() { db.Exec("DELETE FROM actions") }()
	actionRepo, err := NewActionRepo(db, testDbConfig)
	require.NoError(t, err)
	ctx := context.Background()

	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     "org1",
			Project: "proj1",
			Domain:  "domain1",
			Name:    "run1",
		},
		Name: "action-no-backward-succeeded",
	}

	_, err = actionRepo.CreateAction(ctx, models.NewActionModel(actionID), false)
	require.NoError(t, err)

	// Move to SUCCEEDED (terminal, non-retryable)
	endTime := time.Now()
	err = actionRepo.UpdateActionPhase(ctx, actionID,
		common.ActionPhase_ACTION_PHASE_SUCCEEDED, 1,
		core.CatalogCacheStatus_CACHE_DISABLED, &endTime)
	require.NoError(t, err)

	// Try to downgrade from SUCCEEDED to QUEUED — should be a no-op
	err = actionRepo.UpdateActionPhase(ctx, actionID,
		common.ActionPhase_ACTION_PHASE_QUEUED, 2,
		core.CatalogCacheStatus_CACHE_DISABLED, nil)
	require.NoError(t, err)

	action, err := actionRepo.GetAction(ctx, actionID)
	require.NoError(t, err)
	assert.Equal(t, int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED), action.Phase,
		"phase should not downgrade from SUCCEEDED to QUEUED")
}

func TestListRuns(t *testing.T) {
	db := setupActionDB(t)
	defer func() { db.Exec("DELETE FROM actions") }()
	actionRepo, err := NewActionRepo(db, testDbConfig)
	require.NoError(t, err)
	ctx := context.Background()

	runsToCreate := []string{"run-1", "run-2", "run-3"}
	for _, runName := range runsToCreate {
		_, err := actionRepo.CreateAction(ctx, &models.Run{
			Project: "proj1",
			Domain:  "domain1",
			RunName: runName,
			Name:    rootActionName,
			Phase:   int32(common.ActionPhase_ACTION_PHASE_QUEUED),
		}, false)
		require.NoError(t, err)
	}

	// List all runs (root actions only)
	runs, err := actionRepo.ListActions(ctx, interfaces.ListResourceInput{
		Filter: NewIsRootActionFilter(),
		Limit:  50,
	})
	require.NoError(t, err)
	assert.Len(t, runs, 3)
	runNames := map[string]bool{}
	for _, r := range runs {
		runNames[r.RunName] = true
	}
	assert.True(t, runNames["run-1"])
	assert.True(t, runNames["run-2"])
	assert.True(t, runNames["run-3"])

	// ListActions uses a keyset cursor: it returns up to Limit+1 rows so the
	// caller can detect whether another page exists. Page 1 asks for Limit=2
	// and gets all 3 rows back (limit+1 probe). The caller trims to Limit and
	// uses the last kept row's created_at as the CursorToken for page 2.
	runsPage1, err := actionRepo.ListActions(ctx, interfaces.ListResourceInput{
		Filter: NewIsRootActionFilter(),
		Limit:  2,
	})
	require.NoError(t, err)
	assert.Len(t, runsPage1, 3)

	page1 := runsPage1[:2]
	runsPage2, err := actionRepo.ListActions(ctx, interfaces.ListResourceInput{
		Filter:      NewIsRootActionFilter(),
		Limit:       2,
		CursorToken: page1[len(page1)-1].CreatedAt.UTC().Format(time.RFC3339Nano),
	})
	require.NoError(t, err)
	assert.Len(t, runsPage2, 1)

	// Test project scope filtering doesn't include other project
	_, err = actionRepo.CreateAction(ctx, &models.Run{
		Project: "other-proj",
		Domain:  "domain1",
		RunName: "run-other",
		Name:    rootActionName,
		Phase:   int32(common.ActionPhase_ACTION_PHASE_QUEUED),
	}, false)
	require.NoError(t, err)

	runsFiltered, err := actionRepo.ListActions(ctx, interfaces.ListResourceInput{
		Filter: NewIsRootActionFilter().
			And(NewEqualFilter("project", "proj1")).
			And(NewEqualFilter("domain", "domain1")),
		Limit: 50,
	})
	require.NoError(t, err)
	assert.Len(t, runsFiltered, 3)
	for _, r := range runsFiltered {
		assert.Equal(t, "proj1", r.Project)
		assert.Equal(t, "domain1", r.Domain)
	}
}

func setupActionEventDB(t *testing.T) (*sqlx.DB, *actionRepo) {
	db := setupActionDB(t)
	r, err := NewActionRepo(db, testDbConfig)
	require.NoError(t, err)
	repo := r.(*actionRepo)
	return db, repo
}

var testActionID = &common.ActionIdentifier{
	Run: &common.RunIdentifier{
		Org:     "org1",
		Project: "proj1",
		Domain:  "domain1",
		Name:    "run1",
	},
	Name: "action1",
}

func makeTestEvent(attempt, version uint32, phase common.ActionPhase) *workflow.ActionEvent {
	return &workflow.ActionEvent{
		Id:          testActionID,
		Attempt:     attempt,
		Phase:       phase,
		Version:     version,
		UpdatedTime: timestamppb.Now(),
	}
}

func TestGetLatestEventByAttempt_HappyPath(t *testing.T) {
	_, repo := setupActionEventDB(t)
	ctx := context.Background()

	// Insert two events for the same attempt with different versions
	e1, err := models.NewActionEventModel(makeTestEvent(0, 0, common.ActionPhase_ACTION_PHASE_RUNNING))
	require.NoError(t, err)
	e2, err := models.NewActionEventModel(makeTestEvent(0, 1, common.ActionPhase_ACTION_PHASE_SUCCEEDED))
	require.NoError(t, err)

	require.NoError(t, repo.InsertEvents(ctx, []*models.ActionEvent{e1, e2}))

	// Should return the latest version (version=1)
	event, err := repo.GetLatestEventByAttempt(ctx, testActionID, 0)
	require.NoError(t, err)
	assert.Equal(t, uint32(1), event.Version)
}

func TestGetLatestEventByAttempt_NotFound(t *testing.T) {
	_, repo := setupActionEventDB(t)
	ctx := context.Background()

	_, err := repo.GetLatestEventByAttempt(ctx, testActionID, 99)
	require.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestGetLatestEventByAttempt_DifferentAttempts(t *testing.T) {
	_, repo := setupActionEventDB(t)
	ctx := context.Background()

	// Insert events for attempt 0 and attempt 1
	e0, _ := models.NewActionEventModel(makeTestEvent(0, 0, common.ActionPhase_ACTION_PHASE_RUNNING))
	e1, _ := models.NewActionEventModel(makeTestEvent(1, 0, common.ActionPhase_ACTION_PHASE_RUNNING))
	e1v1, _ := models.NewActionEventModel(makeTestEvent(1, 1, common.ActionPhase_ACTION_PHASE_SUCCEEDED))
	require.NoError(t, repo.InsertEvents(ctx, []*models.ActionEvent{e0, e1, e1v1}))

	// Attempt 0 should return version 0
	event, err := repo.GetLatestEventByAttempt(ctx, testActionID, 0)
	require.NoError(t, err)
	assert.Equal(t, uint32(0), event.Attempt)
	assert.Equal(t, uint32(0), event.Version)

	// Attempt 1 should return version 1 (latest)
	event, err = repo.GetLatestEventByAttempt(ctx, testActionID, 1)
	require.NoError(t, err)
	assert.Equal(t, uint32(1), event.Attempt)
	assert.Equal(t, uint32(1), event.Version)
}

func TestInsertEvents_MultipleEventsForDifferentActions(t *testing.T) {
	_, repo := setupActionEventDB(t)
	ctx := context.Background()

	// Insert events for two different actions in the same batch
	actionID2 := &common.ActionIdentifier{
		Run:  testActionID.Run,
		Name: "action2",
	}
	e1, _ := models.NewActionEventModel(makeTestEvent(0, 0, common.ActionPhase_ACTION_PHASE_RUNNING))
	e2, _ := models.NewActionEventModel(&workflow.ActionEvent{
		Id:          actionID2,
		Attempt:     0,
		Phase:       common.ActionPhase_ACTION_PHASE_RUNNING,
		Version:     0,
		UpdatedTime: timestamppb.Now(),
	})
	require.NoError(t, repo.InsertEvents(ctx, []*models.ActionEvent{e1, e2}))

	// Both events should be retrievable
	got1, err := repo.GetLatestEventByAttempt(ctx, testActionID, 0)
	require.NoError(t, err)
	assert.Equal(t, "action1", got1.Name)

	got2, err := repo.GetLatestEventByAttempt(ctx, actionID2, 0)
	require.NoError(t, err)
	assert.Equal(t, "action2", got2.Name)
}

func TestInsertEvents_Empty(t *testing.T) {
	_, repo := setupActionEventDB(t)
	ctx := context.Background()

	// Insert empty slice should be no-op
	err := repo.InsertEvents(ctx, []*models.ActionEvent{})
	assert.NoError(t, err)
}

func TestNotifyActionUpdate_PayloadWithSpecialChars(t *testing.T) {
	r := &actionRepo{

		actionNotifyCh: make(chan string, 256),
		runNotifyCh:    make(chan string, 256),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Payload with single quotes that would cause SQL injection with string interpolation.
	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     "org",
			Project: "proj",
			Domain:  "domain",
			Name:    "run'; DROP TABLE actions; --",
		},
		Name: "action",
	}

	r.notifyActionUpdate(ctx, actionID)

	select {
	case payload := <-r.actionNotifyCh:
		assert.Equal(t, "proj/domain/run'; DROP TABLE actions; --/action", payload)
	case <-ctx.Done():
		t.Fatal("timed out waiting for payload on actionNotifyCh")
	}
}

func TestNotifyRunUpdate_PayloadWithSpecialChars(t *testing.T) {
	r := &actionRepo{

		actionNotifyCh: make(chan string, 256),
		runNotifyCh:    make(chan string, 256),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	runID := &common.RunIdentifier{
		Org:     "org",
		Project: "proj",
		Domain:  "domain",
		Name:    "run'); SELECT pg_sleep(10); --",
	}

	r.notifyRunUpdate(ctx, runID)

	select {
	case payload := <-r.runNotifyCh:
		assert.Equal(t, "proj/domain/run'); SELECT pg_sleep(10); --", payload)
	case <-ctx.Done():
		t.Fatal("timed out waiting for payload on runNotifyCh")
	}
}

func TestIsConnError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		expect bool
	}{
		{
			name:   "driver.ErrBadConn",
			err:    driver.ErrBadConn,
			expect: true,
		},
		{
			name:   "wrapped driver.ErrBadConn",
			err:    fmt.Errorf("exec failed: %w", driver.ErrBadConn),
			expect: true,
		},
		{
			name:   "net.OpError",
			err:    &net.OpError{Op: "read", Err: errors.New("connection reset")},
			expect: true,
		},
		{
			name:   "pq connection_exception class 08",
			err:    &pq.Error{Code: "08006"},
			expect: true,
		},
		{
			name:   "pq non-connection error",
			err:    &pq.Error{Code: "42P01"},
			expect: false,
		},
		{
			name:   "generic error",
			err:    errors.New("something went wrong"),
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, isConnError(tt.err))
		})
	}
}

func TestRunNotifyLoop_NilConnNoPanic(t *testing.T) {
	// Verify that runNotifyLoop handles a nil connection gracefully
	// (e.g. after a failed reconnect) instead of panicking.
	r := &actionRepo{

		actionNotifyCh: make(chan string, 256),
		runNotifyCh:    make(chan string, 256),
	}

	// Send a notification, then close the channel so the loop exits.
	r.actionNotifyCh <- "proj/domain/run/action"
	close(r.actionNotifyCh)

	// Pass a nil conn — should not panic.
	assert.NotPanics(t, func() {
		r.runNotifyLoop(nil, nil)
	})
}

func TestInsertEvents_WithLogContext(t *testing.T) {
	_, repo := setupActionEventDB(t)
	ctx := context.Background()

	event := &workflow.ActionEvent{
		Id:          testActionID,
		Attempt:     0,
		Phase:       common.ActionPhase_ACTION_PHASE_RUNNING,
		Version:     1,
		UpdatedTime: timestamppb.Now(),
		LogContext: &core.LogContext{
			PrimaryPodName: "my-pod",
			Pods: []*core.PodLogContext{
				{PodName: "my-pod", Namespace: "default"},
			},
		},
	}
	eventModel, err := models.NewActionEventModel(event)
	require.NoError(t, err)

	require.NoError(t, repo.InsertEvents(ctx, []*models.ActionEvent{eventModel}))

	// Fetch it back via GetLatestEventByAttempt and verify log context is preserved
	fetched, err := repo.GetLatestEventByAttempt(ctx, testActionID, 0)
	require.NoError(t, err)
	deserialized, err := fetched.ToActionEvent()
	require.NoError(t, err)
	assert.Equal(t, "my-pod", deserialized.GetLogContext().GetPrimaryPodName())
}

// TestUpdateActionPhase_AbortedDoesNotInsertEvent verifies that transitioning an
// action to ABORTED updates the phase column but does NOT insert a synthetic row
// into action_events. The abort event is now emitted by the controller via
// RecordActionEvents before the TaskAction finalizer is removed.
func TestUpdateActionPhase_AbortedDoesNotInsertEvent(t *testing.T) {
	db := setupActionDB(t)
	actionRepo, err := NewActionRepo(db, testDbConfig)
	require.NoError(t, err)
	ctx := context.Background()

	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{Project: "p", Domain: "d", Name: "run-abort"},
		Name: "abort-action",
	}
	_, err = actionRepo.CreateAction(ctx, models.NewActionModel(actionID), false)
	require.NoError(t, err)

	endTime := time.Now()
	err = actionRepo.UpdateActionPhase(ctx, actionID, common.ActionPhase_ACTION_PHASE_ABORTED, 1, core.CatalogCacheStatus_CACHE_DISABLED, &endTime)
	require.NoError(t, err)

	// Phase column must be updated.
	action, err := actionRepo.GetAction(ctx, actionID)
	require.NoError(t, err)
	assert.Equal(t, int32(common.ActionPhase_ACTION_PHASE_ABORTED), action.Phase)

	// No synthetic event row should have been inserted — the controller now emits the event.
	var count int
	err = db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM action_events WHERE project=$1 AND domain=$2 AND run_name=$3 AND name=$4`,
		actionID.Run.Project, actionID.Run.Domain, actionID.Run.Name, actionID.Name,
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "UpdateActionPhase(ABORTED) must not insert a synthetic action_events row")
}
