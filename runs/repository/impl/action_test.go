package impl

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
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
		Project:          runID.Project,
		Domain:           runID.Domain,
		RunName:          runID.Name,
		Name:             rootActionName,
		Phase:            int32(common.ActionPhase_ACTION_PHASE_QUEUED),
		CreatedBySubject: sql.NullString{String: "00uABC", Valid: true},
	}

	run, err := actionRepo.CreateAction(ctx, runModel, false)
	require.NoError(t, err)
	require.NotNil(t, run)
	assert.Equal(t, runID.Project, run.Project)
	assert.Equal(t, runID.Domain, run.Domain)
	assert.Equal(t, runID.Name, run.RunName)
	assert.Equal(t, "a0", run.Name)
	assert.Equal(t, int32(common.ActionPhase_ACTION_PHASE_QUEUED), run.Phase)
	// created_by_subject (indexed owner subject) round-trips for filtering/listing by owner.
	assert.Equal(t, "00uABC", run.CreatedBySubject.String)

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
		nil,
	)
	require.NoError(t, err)

	action, err := actionRepo.GetAction(ctx, actionID)
	require.NoError(t, err)
	assert.Equal(t, int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED), action.Phase)
	assert.Equal(t, uint32(3), action.Attempts)
	assert.Equal(t, core.CatalogCacheStatus_CACHE_HIT, action.CacheStatus)
	assert.True(t, action.EndedAt.Valid)
}

// A supplied start time (the action's CRD creationTimestamp) must drive duration, so a
// long-held action recorded late — e.g. when coalesced/backlogged events collapse
// created_at toward the terminal time — still reports its real wall-clock duration.
func TestUpdateActionPhase_StartTimeCorrectsDuration(t *testing.T) {
	db := setupActionDB(t)
	defer func() { db.Exec("DELETE FROM actions") }()
	actionRepo, err := NewActionRepo(db, testDbConfig)
	require.NoError(t, err)
	ctx := context.Background()

	actionID := &common.ActionIdentifier{
		Run:  &common.RunIdentifier{Org: "org1", Project: "proj1", Domain: "domain1", Name: "run1"},
		Name: "held-action",
	}
	// Row created "late" (created_at defaults to now), as when a coalesced event records
	// a long-running action only at terminal time.
	_, err = actionRepo.CreateAction(ctx, models.NewActionModel(actionID), false)
	require.NoError(t, err)

	// The action actually started 120s ago (its CRD creationTimestamp) and just ended.
	endTime := time.Now()
	startTime := endTime.Add(-120 * time.Second)
	err = actionRepo.UpdateActionPhase(ctx, actionID, common.ActionPhase_ACTION_PHASE_SUCCEEDED,
		1, core.CatalogCacheStatus_CACHE_DISABLED, &endTime, &startTime)
	require.NoError(t, err)

	action, err := actionRepo.GetAction(ctx, actionID)
	require.NoError(t, err)
	require.True(t, action.DurationMs.Valid)
	// Duration ~= ended - start (120s), not ended - created_at (~0).
	assert.InDelta(t, 120000, action.DurationMs.Int64, 5000,
		"duration must come from the supplied start time, not the late created_at")
}

// Without a start time, duration falls back to ended - created_at (unchanged behaviour).
func TestUpdateActionPhase_NilStartTimeUsesCreatedAt(t *testing.T) {
	db := setupActionDB(t)
	defer func() { db.Exec("DELETE FROM actions") }()
	actionRepo, err := NewActionRepo(db, testDbConfig)
	require.NoError(t, err)
	ctx := context.Background()

	actionID := &common.ActionIdentifier{
		Run:  &common.RunIdentifier{Org: "org1", Project: "proj1", Domain: "domain1", Name: "run1"},
		Name: "quick-action",
	}
	_, err = actionRepo.CreateAction(ctx, models.NewActionModel(actionID), false)
	require.NoError(t, err)

	endTime := time.Now()
	err = actionRepo.UpdateActionPhase(ctx, actionID, common.ActionPhase_ACTION_PHASE_SUCCEEDED,
		1, core.CatalogCacheStatus_CACHE_DISABLED, &endTime, nil)
	require.NoError(t, err)

	action, err := actionRepo.GetAction(ctx, actionID)
	require.NoError(t, err)
	require.True(t, action.DurationMs.Valid)
	assert.Less(t, action.DurationMs.Int64, int64(5000), "without a start time, duration is created_at-based")
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
	err = repo.UpdateActionPhase(ctx, otherActionID, common.ActionPhase_ACTION_PHASE_RUNNING, 1, core.CatalogCacheStatus_CACHE_DISABLED, nil, nil)
	require.NoError(t, err)

	select {
	case action := <-updates:
		t.Fatalf("unexpected update for action %s", action.Name)
	case err := <-errs:
		require.NoError(t, err)
	case <-time.After(1200 * time.Millisecond):
	}

	// Update "target" — should produce an update.
	err = repo.UpdateActionPhase(ctx, targetActionID, common.ActionPhase_ACTION_PHASE_RUNNING, 1, core.CatalogCacheStatus_CACHE_DISABLED, nil, nil)
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
		core.CatalogCacheStatus_CACHE_DISABLED, &endTime, nil)
	require.NoError(t, err)

	action, err := actionRepo.GetAction(ctx, actionID)
	require.NoError(t, err)
	assert.Equal(t, int32(common.ActionPhase_ACTION_PHASE_FAILED), action.Phase)

	// Retry: transition from FAILED back to QUEUED — should succeed
	err = actionRepo.UpdateActionPhase(ctx, actionID,
		common.ActionPhase_ACTION_PHASE_QUEUED, 2,
		core.CatalogCacheStatus_CACHE_DISABLED, nil, nil)
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
		core.CatalogCacheStatus_CACHE_DISABLED, nil, nil)
	require.NoError(t, err)

	// Try to downgrade from RUNNING to QUEUED — should be a no-op (phase guard)
	err = actionRepo.UpdateActionPhase(ctx, actionID,
		common.ActionPhase_ACTION_PHASE_QUEUED, 1,
		core.CatalogCacheStatus_CACHE_DISABLED, nil, nil)
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
		core.CatalogCacheStatus_CACHE_DISABLED, &endTime, nil)
	require.NoError(t, err)

	// Try to downgrade from SUCCEEDED to QUEUED — should be a no-op
	err = actionRepo.UpdateActionPhase(ctx, actionID,
		common.ActionPhase_ACTION_PHASE_QUEUED, 2,
		core.CatalogCacheStatus_CACHE_DISABLED, nil, nil)
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

	// ListActions offset-paginates: it returns up to Limit+1 rows so the caller can detect
	// whether another page exists. Page 1 asks for Limit=2 and gets all 3 rows back
	// (limit+1 probe); page 2 continues at Offset=2 and returns the last row.
	runsPage1, err := actionRepo.ListActions(ctx, interfaces.ListResourceInput{
		Filter: NewIsRootActionFilter(),
		Limit:  2,
	})
	require.NoError(t, err)
	assert.Len(t, runsPage1, 3)

	runsPage2, err := actionRepo.ListActions(ctx, interfaces.ListResourceInput{
		Filter: NewIsRootActionFilter(),
		Limit:  2,
		Offset: 2,
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

func TestListRuns_HasPausedActionFilter(t *testing.T) {
	db := setupActionDB(t)
	defer func() { db.Exec("DELETE FROM actions") }()
	actionRepo, err := NewActionRepo(db, testDbConfig)
	require.NoError(t, err)
	ctx := context.Background()

	// run-paused: RUNNING root with a PAUSED child (HITL gate awaiting input).
	// run-plain: RUNNING root whose child is also RUNNING.
	for _, a := range []*models.Action{
		{Project: "proj1", Domain: "domain1", RunName: "run-paused", Name: rootActionName,
			Phase: int32(common.ActionPhase_ACTION_PHASE_RUNNING)},
		{Project: "proj1", Domain: "domain1", RunName: "run-paused", Name: "gate-node",
			ParentActionName: sql.NullString{String: rootActionName, Valid: true},
			Phase:            int32(common.ActionPhase_ACTION_PHASE_PAUSED)},
		{Project: "proj1", Domain: "domain1", RunName: "run-plain", Name: rootActionName,
			Phase: int32(common.ActionPhase_ACTION_PHASE_RUNNING)},
		{Project: "proj1", Domain: "domain1", RunName: "run-plain", Name: "worker-node",
			ParentActionName: sql.NullString{String: rootActionName, Valid: true},
			Phase:            int32(common.ActionPhase_ACTION_PHASE_RUNNING)},
	} {
		_, err := actionRepo.CreateAction(ctx, a, false)
		require.NoError(t, err)
	}

	runs, err := actionRepo.ListActions(ctx, interfaces.ListResourceInput{
		Filter: NewIsRootActionFilter().And(NewHasPausedActionFilter()),
		Limit:  50,
	})
	require.NoError(t, err)
	require.Len(t, runs, 1)
	assert.Equal(t, "run-paused", runs[0].RunName)
	assert.False(t, runs[0].ParentActionName.Valid, "only the root action should be returned")
}

// TestListActions_KeysetPagination covers the O(n) keyset paging used by the
// WatchActions snapshot: pages continue after the previous page's (created_at, name)
// instead of by OFFSET. It forces tied created_at (the bulk-created map-task case) so
// paging must rely on the unique "name" tiebreaker to stay a total order and cover
// every action exactly once.
func TestListActions_KeysetPagination(t *testing.T) {
	db := setupActionDB(t)
	defer func() { db.Exec("DELETE FROM actions") }()
	actionRepo, err := NewActionRepo(db, testDbConfig)
	require.NoError(t, err)
	ctx := context.Background()

	runID := &common.RunIdentifier{Project: "proj1", Domain: "domain1", Name: "run1"}
	const total = 250
	for i := 0; i < total; i++ {
		aid := &common.ActionIdentifier{Run: runID, Name: fmt.Sprintf("n%04d", i)}
		_, err := actionRepo.CreateAction(ctx, models.NewActionModel(aid), false)
		require.NoError(t, err)
	}
	// Force one shared created_at (map-task tie case) so keyset leans on the name tiebreaker.
	_, err = db.Exec("UPDATE actions SET created_at = '2024-01-01T00:00:00Z'")
	require.NoError(t, err)

	const pageSize = 50
	sortAsc := []interfaces.SortParameter{
		NewSortParameter("created_at", interfaces.SortOrderAscending),
		NewSortParameter("name", interfaces.SortOrderAscending),
	}
	seen := map[string]struct{}{}
	var ordered []string
	var afterCreatedAt *time.Time
	var afterName string
	for {
		batch, err := actionRepo.ListActions(ctx, interfaces.ListResourceInput{
			Filter:               NewRunActionsFilter(runID),
			Limit:                pageSize,
			KeysetAfterCreatedAt: afterCreatedAt,
			KeysetAfterName:      afterName,
			SortParameters:       sortAsc,
		})
		require.NoError(t, err)
		hasMore := len(batch) > pageSize
		if hasMore {
			batch = batch[:pageSize]
		}
		for _, a := range batch {
			seen[a.Name] = struct{}{}
			ordered = append(ordered, a.Name)
		}
		if !hasMore || len(batch) == 0 {
			break
		}
		last := batch[len(batch)-1]
		afterCreatedAt = &last.CreatedAt
		afterName = last.Name
	}

	assert.Len(t, seen, total, "keyset paging over tied created_at must cover every action exactly once")
	assert.True(t, sort.IsSorted(sort.StringSlice(ordered)),
		"keyset order must be a total order (no skips/overlaps)")

	// Keyset is mutually exclusive with Offset.
	now := time.Now()
	_, err = actionRepo.ListActions(ctx, interfaces.ListResourceInput{
		Filter:               NewRunActionsFilter(runID),
		Limit:                10,
		KeysetAfterCreatedAt: &now,
		Offset:               5,
	})
	require.Error(t, err)

	// Keyset requires the (created_at ASC, name ASC) sort; any other sort is rejected
	// because the keyset WHERE would otherwise skip/duplicate rows across pages.
	_, err = actionRepo.ListActions(ctx, interfaces.ListResourceInput{
		Filter:               NewRunActionsFilter(runID),
		Limit:                10,
		KeysetAfterCreatedAt: &now,
		KeysetAfterName:      "n0001",
		SortParameters:       []interfaces.SortParameter{NewSortParameter("name", interfaces.SortOrderAscending)},
	})
	require.Error(t, err)
}

// TestListActions_OffsetPagination covers the offset paging used by the ListRuns/ListActions
// RPC: the page token is the running offset. It forces tied created_at (the bulk-created
// map-task case) to confirm the default sort's (run_name, name) tiebreaker keeps paging
// stable — every action is returned exactly once with no skips or repeats.
func TestListActions_OffsetPagination(t *testing.T) {
	db := setupActionDB(t)
	defer func() { db.Exec("DELETE FROM actions") }()
	actionRepo, err := NewActionRepo(db, testDbConfig)
	require.NoError(t, err)
	ctx := context.Background()

	runID := &common.RunIdentifier{Project: "proj1", Domain: "domain1", Name: "run1"}
	const total = 250
	for i := 0; i < total; i++ {
		aid := &common.ActionIdentifier{Run: runID, Name: fmt.Sprintf("n%04d", i)}
		_, err := actionRepo.CreateAction(ctx, models.NewActionModel(aid), false)
		require.NoError(t, err)
	}
	_, err = db.Exec("UPDATE actions SET created_at = '2024-01-01T00:00:00Z'")
	require.NoError(t, err)

	const pageSize = 50
	seen := map[string]struct{}{}
	for offset := 0; ; offset += pageSize {
		batch, err := actionRepo.ListActions(ctx, interfaces.ListResourceInput{
			Filter: NewRunActionsFilter(runID),
			Limit:  pageSize,
			Offset: offset,
		})
		require.NoError(t, err)
		hasMore := len(batch) > pageSize
		if hasMore {
			batch = batch[:pageSize]
		}
		for _, a := range batch {
			_, dup := seen[a.Name]
			require.False(t, dup, "action %s returned on more than one page", a.Name)
			seen[a.Name] = struct{}{}
		}
		if !hasMore || len(batch) == 0 {
			break
		}
	}
	assert.Len(t, seen, total, "offset paging must cover every action exactly once")

	// Negative offset is rejected.
	_, err = actionRepo.ListActions(ctx, interfaces.ListResourceInput{
		Filter: NewRunActionsFilter(runID), Limit: 10, Offset: -1,
	})
	require.Error(t, err)

	// Offset is mutually exclusive with keyset.
	now := time.Now()
	_, err = actionRepo.ListActions(ctx, interfaces.ListResourceInput{
		Filter: NewRunActionsFilter(runID), Limit: 10, Offset: 5, KeysetAfterCreatedAt: &now,
	})
	require.Error(t, err)
}

// TestListActions_OffsetPaginationClientSortTiedCreatedAt guards against skipping/duplicating
// rows when a client-supplied sort is not a total order. `created_at DESC` over bulk-created
// map-task children (all tied on created_at) leaves ties in an arbitrary order per query, so
// OFFSET paging would be unstable without the appended (run_name, name) tiebreakers.
func TestListActions_OffsetPaginationClientSortTiedCreatedAt(t *testing.T) {
	db := setupActionDB(t)
	defer func() { db.Exec("DELETE FROM actions") }()
	actionRepo, err := NewActionRepo(db, testDbConfig)
	require.NoError(t, err)
	ctx := context.Background()

	runID := &common.RunIdentifier{Project: "proj1", Domain: "domain1", Name: "run1"}
	const total = 250
	for i := 0; i < total; i++ {
		aid := &common.ActionIdentifier{Run: runID, Name: fmt.Sprintf("n%04d", i)}
		_, err := actionRepo.CreateAction(ctx, models.NewActionModel(aid), false)
		require.NoError(t, err)
	}
	_, err = db.Exec("UPDATE actions SET created_at = '2024-01-01T00:00:00Z'")
	require.NoError(t, err)

	sortDesc := []interfaces.SortParameter{NewSortParameter("created_at", interfaces.SortOrderDescending)}
	const pageSize = 50
	seen := map[string]struct{}{}
	for offset := 0; ; offset += pageSize {
		batch, err := actionRepo.ListActions(ctx, interfaces.ListResourceInput{
			Filter:         NewRunActionsFilter(runID),
			Limit:          pageSize,
			Offset:         offset,
			SortParameters: sortDesc,
		})
		require.NoError(t, err)
		hasMore := len(batch) > pageSize
		if hasMore {
			batch = batch[:pageSize]
		}
		for _, a := range batch {
			_, dup := seen[a.Name]
			require.False(t, dup, "action %s returned on more than one page", a.Name)
			seen[a.Name] = struct{}{}
		}
		if !hasMore || len(batch) == 0 {
			break
		}
	}
	assert.Len(t, seen, total, "client-sorted offset paging over tied created_at must cover every action exactly once")
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

// A batch larger than Postgres' 65,535-bind-parameter budget (9 params/row ->
// 7,281 rows) must be chunked into multiple INSERTs rather than failing whole.
// RecordActionEvents accepts arbitrary event counts, so this is a real API path.
func TestInsertEvents_ChunksOversizedBatch(t *testing.T) {
	_, repo := setupActionEventDB(t)
	ctx := context.Background()

	const n = 7300 // > 7,281, forcing at least two chunks
	events := make([]*models.ActionEvent, 0, n)
	for i := 0; i < n; i++ {
		e, err := models.NewActionEventModel(&workflow.ActionEvent{
			Id:          testActionID,
			Attempt:     0,
			Phase:       common.ActionPhase_ACTION_PHASE_RUNNING,
			Version:     uint32(i), // distinct versions so rows are not deduped
			UpdatedTime: timestamppb.Now(),
		})
		require.NoError(t, err)
		events = append(events, e)
	}
	require.NoError(t, repo.InsertEvents(ctx, events))

	got, err := repo.ListEvents(ctx, testActionID, n+1)
	require.NoError(t, err)
	assert.Len(t, got, n, "all chunked rows must be inserted")
}

func TestInsertEvents_Empty(t *testing.T) {
	_, repo := setupActionEventDB(t)
	ctx := context.Background()

	// Insert empty slice should be no-op
	err := repo.InsertEvents(ctx, []*models.ActionEvent{})
	assert.NoError(t, err)
}

func TestNotifyActionUpdate_PayloadWithSpecialChars(t *testing.T) {
	r := newNotifyTestRepo()

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

	actions, _ := r.takePendingNotifications()
	assert.Contains(t, actions, "proj/domain/run'; DROP TABLE actions; --/action")
}

func TestNotifyRunUpdate_PayloadWithSpecialChars(t *testing.T) {
	r := newNotifyTestRepo()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	runID := &common.RunIdentifier{
		Org:     "org",
		Project: "proj",
		Domain:  "domain",
		Name:    "run'); SELECT pg_sleep(10); --",
	}

	r.notifyRunUpdate(ctx, runID)

	_, runs := r.takePendingNotifications()
	assert.Contains(t, runs, "proj/domain/run'); SELECT pg_sleep(10); --")
}

// TestNotifyActionUpdate_CoalescesRepeats checks the property that makes a set
// safe here: repeated updates to one action between drains collapse into a
// single wakeup, because the listener re-reads the latest state anyway.
func TestNotifyActionUpdate_CoalescesRepeats(t *testing.T) {
	r := newNotifyTestRepo()
	ctx := context.Background()

	for i := 0; i < 500; i++ {
		r.notifyActionUpdate(ctx, notifyTestActionID("same-action"))
	}
	r.notifyActionUpdate(ctx, notifyTestActionID("other-action"))

	actions, _ := r.pendingCounts()
	assert.Equal(t, 2, actions, "repeated updates to one action must collapse into one pending entry")
}

// TestNotifyActionUpdate_KeepsWakeupAfterContextCancel covers the loss path
// that existed before: a client disconnecting mid-request used to discard a
// wakeup that other watchers still needed.
func TestNotifyActionUpdate_KeepsWakeupAfterContextCancel(t *testing.T) {
	r := newNotifyTestRepo()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	r.notifyActionUpdate(ctx, notifyTestActionID("cancelled-caller"))
	r.notifyRunUpdate(ctx, &common.RunIdentifier{Project: "proj", Domain: "domain", Name: "run"})

	actions, runs := r.takePendingNotifications()
	assert.Contains(t, actions, "proj/domain/run/cancelled-caller")
	assert.Contains(t, runs, "proj/domain/run")
}

// TestRunNotifyLoop_RetriesUndeliveredPayloads verifies that a failed
// pg_notify keeps the payload pending instead of dropping it. A nil connection
// with no database to reconnect to is the simplest permanent failure. The
// whole batch must survive, not just the payload that failed first: a dead
// connection aborts the drain rather than retrying the connection per payload.
func TestRunNotifyLoop_RetriesUndeliveredPayloads(t *testing.T) {
	r := newNotifyTestRepo()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < 3; i++ {
		r.notifyActionUpdate(ctx, notifyTestActionID(fmt.Sprintf("undeliverable-%d", i)))
	}
	for i := 0; i < 2; i++ {
		r.notifyRunUpdate(ctx, &common.RunIdentifier{
			Project: "proj", Domain: "domain", Name: fmt.Sprintf("run-%d", i),
		})
	}

	loopDone := make(chan struct{})
	go func() {
		defer close(loopDone)
		r.runNotifyLoop(ctx, nil, nil)
	}()

	// Let the pump take the payloads, fail to deliver them, and requeue.
	assert.Eventually(t, func() bool {
		actions, runs := r.pendingCounts()
		return actions == 3 && runs == 2
	}, 5*time.Second, 20*time.Millisecond, "undelivered payloads must stay pending for retry")

	cancel()
	select {
	case <-loopDone:
	case <-time.After(5 * time.Second):
		t.Fatal("runNotifyLoop did not return after its context was cancelled")
	}
}

// newNotifyTestRepo builds a repo with the notify plumbing initialized but no
// pump running, so the pending work is observable and nothing drains it.
func newNotifyTestRepo() *actionRepo {
	return &actionRepo{
		pendingActions: make(map[string]int),
		pendingRuns:    make(map[string]int),
		pendingCh:      make(chan struct{}, 1),
	}
}

// pendingCounts reports how many payloads are queued for the pump.
func (r *actionRepo) pendingCounts() (actions, runs int) {
	r.notifyMu.Lock()
	defer r.notifyMu.Unlock()
	return len(r.pendingActions), len(r.pendingRuns)
}

// pendingDropped reports when the given payload has been absent from the
// pending set on enough consecutive checks to rule out the short window where
// the pump is holding a taken batch and has not requeued it yet. A single
// absent reading is not evidence of a drop.
func (r *actionRepo) pendingDropped(payload string) func() bool {
	const consecutive = 25
	absent := 0
	return func() bool {
		r.notifyMu.Lock()
		_, queued := r.pendingActions[payload]
		r.notifyMu.Unlock()

		if queued {
			absent = 0
			return false
		}
		absent++
		return absent >= consecutive
	}
}

// newNotifyRepoWithDB builds a repo wired to a real database and listener but
// with no pump running, so a test can drive runNotifyLoop itself and watch
// what actually reaches the wire.
func newNotifyRepoWithDB(t *testing.T) (*actionRepo, *sql.DB, *sql.Conn) {
	t.Helper()
	db := setupActionDB(t)

	r := &actionRepo{
		db:                db,
		dsn:               database.GetPostgresDsn(context.Background(), testDbConfig.Postgres),
		runSubscribers:    make(map[chan string]bool),
		actionSubscribers: make(map[chan string]bool),
		pendingActions:    make(map[string]int),
		pendingRuns:       make(map[string]int),
		pendingCh:         make(chan struct{}, 1),
	}
	require.NoError(t, r.startPostgresListener())

	conn, err := db.DB.Conn(context.Background())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() }) //nolint:errcheck

	return r, db.DB, conn
}

// subscribeActions registers a raw subscriber the way WatchActionUpdates does.
func subscribeActions(t *testing.T, r *actionRepo, size int) chan string {
	t.Helper()
	ch := make(chan string, size)
	r.mu.Lock()
	r.actionSubscribers[ch] = true
	r.mu.Unlock()
	t.Cleanup(func() {
		r.mu.Lock()
		delete(r.actionSubscribers, ch)
		r.mu.Unlock()
	})
	return ch
}

func notifyTestActionID(name string) *common.ActionIdentifier {
	return &common.ActionIdentifier{
		Run:  &common.RunIdentifier{Org: "org", Project: "proj", Domain: "domain", Name: "run"},
		Name: name,
	}
}

// TestNotifyActionUpdate_DoesNotBlockOnStalledPump pins the property the write
// path depends on: the row is already committed by the time we notify, so a
// stalled pump must never turn into RPC latency.
func TestNotifyActionUpdate_DoesNotBlockOnStalledPump(t *testing.T) {
	r := newNotifyTestRepo()

	// No pump is running, so nothing consumes what the writer produces. Push
	// far more updates than any fixed-size buffer would hold.
	const updates = 5000

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < updates; i++ {
			r.notifyActionUpdate(ctx, notifyTestActionID(fmt.Sprintf("action-%d", i)))
		}
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("notifyActionUpdate blocked while the notify pump was stalled")
	}
}

// TestNotifyRunUpdate_DoesNotBlockOnStalledPump is the run-side twin of the
// test above; both notify paths share the same pump.
func TestNotifyRunUpdate_DoesNotBlockOnStalledPump(t *testing.T) {
	r := newNotifyTestRepo()

	const updates = 5000

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < updates; i++ {
			r.notifyRunUpdate(ctx, &common.RunIdentifier{
				Org: "org", Project: "proj", Domain: "domain", Name: fmt.Sprintf("run-%d", i),
			})
		}
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("notifyRunUpdate blocked while the notify pump was stalled")
	}
}

// TestWatchActionUpdates_DeliversPhaseChange exercises the full product path
// against a real database: UpdateActionPhase writes the row, the notify pump
// issues pg_notify, the listener fans out to subscribers, and the watcher
// re-reads the action. It guards the other half of the contract, that making
// the writer non-blocking must not lose a wakeup.
func TestWatchActionUpdates_DeliversPhaseChange(t *testing.T) {
	db := setupActionDB(t)
	repo, err := NewActionRepo(db, testDbConfig)
	require.NoError(t, err)
	repoImpl, ok := repo.(*actionRepo)
	require.True(t, ok)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	actionID := &common.ActionIdentifier{
		Run:  &common.RunIdentifier{Project: "p", Domain: "d", Name: "run-watch"},
		Name: "watched-action",
	}
	_, err = repo.CreateAction(ctx, models.NewActionModel(actionID), false)
	require.NoError(t, err)

	updates := make(chan *models.Action, 8)
	errs := make(chan error, 8)
	go repo.WatchActionUpdates(ctx, actionID, updates, errs)

	require.Eventually(t, func() bool {
		repoImpl.mu.RLock()
		defer repoImpl.mu.RUnlock()
		return len(repoImpl.actionSubscribers) > 0
	}, 2*time.Second, 10*time.Millisecond, "timed out waiting for watcher registration")

	require.NoError(t, repo.UpdateActionPhase(ctx, actionID,
		common.ActionPhase_ACTION_PHASE_RUNNING, 0, core.CatalogCacheStatus_CACHE_DISABLED, nil, nil))

	// CreateAction notifies as well, so the watcher can legitimately see the
	// initial phase first. Wait for the transition we triggered.
	deadline := time.After(15 * time.Second)
	for {
		select {
		case got := <-updates:
			if got.Phase == int32(common.ActionPhase_ACTION_PHASE_RUNNING) {
				return
			}
		case err := <-errs:
			t.Fatalf("watch reported an error: %v", err)
		case <-deadline:
			t.Fatal("phase change never reached the watcher")
		}
	}
}

// TestNotifyPump_ConcurrentWritersDeliverEveryAction runs many writers against
// a live pump and a real database, and checks the delivery half of the
// contract: every action that was notified reaches the wire. The stalled-pump
// tests above own the "writers never block" half.
func TestNotifyPump_ConcurrentWritersDeliverEveryAction(t *testing.T) {
	db := setupActionDB(t)
	repoIface, err := NewActionRepo(db, testDbConfig)
	require.NoError(t, err)
	repo, ok := repoIface.(*actionRepo)
	require.True(t, ok)

	const writers = 40
	const perWriter = 25
	const total = writers * perWriter

	// Subscribe the way WatchActionUpdates does, so we observe what actually
	// came back through pg_notify and the listener.
	delivered := make(chan string, 4*total)
	repo.mu.Lock()
	repo.actionSubscribers[delivered] = true
	repo.mu.Unlock()
	defer func() {
		repo.mu.Lock()
		delete(repo.actionSubscribers, delivered)
		repo.mu.Unlock()
	}()

	ctx := context.Background()
	var wg sync.WaitGroup
	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < perWriter; i++ {
				repo.notifyActionUpdate(ctx, notifyTestActionID(fmt.Sprintf("a-%d-%d", w, i)))
			}
		}(w)
	}
	wg.Wait()

	seen := make(map[string]struct{}, total)
	deadline := time.After(60 * time.Second)
	for len(seen) < total {
		select {
		case payload := <-delivered:
			seen[payload] = struct{}{}
		case <-deadline:
			t.Fatalf("only %d of %d notifications were delivered", len(seen), total)
		}
	}

	for w := 0; w < writers; w++ {
		for i := 0; i < perWriter; i++ {
			assert.Contains(t, seen, fmt.Sprintf("proj/domain/run/a-%d-%d", w, i))
		}
	}
}

// TestUpdateActionPhase_CompletesWithStalledPump is the acceptance criterion
// the issue states directly: with the pump stalled the writer returns
// immediately and the row is still written.
func TestUpdateActionPhase_CompletesWithStalledPump(t *testing.T) {
	db := setupActionDB(t)
	r := &actionRepo{
		db:             db,
		pendingActions: make(map[string]int),
		pendingRuns:    make(map[string]int),
		pendingCh:      make(chan struct{}, 1),
	}
	// No pump is started, so nothing drains what the write path queues.

	ctx := context.Background()
	actionID := &common.ActionIdentifier{
		Run:  &common.RunIdentifier{Project: "p", Domain: "d", Name: "run-stalled"},
		Name: "stalled-action",
	}
	_, err := r.CreateAction(ctx, models.NewActionModel(actionID), false)
	require.NoError(t, err)

	// Back the queue up past what the old 256-slot buffer could hold.
	for i := 0; i < 300; i++ {
		r.notifyActionUpdate(ctx, notifyTestActionID(fmt.Sprintf("backlog-%d", i)))
	}

	start := time.Now()
	require.NoError(t, r.UpdateActionPhase(ctx, actionID,
		common.ActionPhase_ACTION_PHASE_RUNNING, 0, core.CatalogCacheStatus_CACHE_DISABLED, nil, nil))
	assert.Less(t, time.Since(start), 2*time.Second, "the write must not wait on a stalled pump")

	action, err := r.GetAction(ctx, actionID)
	require.NoError(t, err)
	assert.Equal(t, int32(common.ActionPhase_ACTION_PHASE_RUNNING), action.Phase,
		"the row must still be written while the pump is stalled")

	actions, _ := r.pendingCounts()
	assert.Equal(t, 301, actions, "the notification must be queued rather than dropped")
}

// TestNotifyPump_CoalescesRepeatsOnTheWire checks the collapse where it counts,
// on the wire rather than in the map. Every update is queued before the pump
// starts, so the whole burst is one drain.
func TestNotifyPump_CoalescesRepeatsOnTheWire(t *testing.T) {
	r, sqlDB, conn := newNotifyRepoWithDB(t)
	delivered := subscribeActions(t, r, 64)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < 500; i++ {
		r.notifyActionUpdate(ctx, notifyTestActionID("busy-action"))
	}

	go r.runNotifyLoop(ctx, sqlDB, conn)

	select {
	case payload := <-delivered:
		assert.Equal(t, "proj/domain/run/busy-action", payload)
	case <-time.After(15 * time.Second):
		t.Fatal("the coalesced notification was never delivered")
	}

	select {
	case extra := <-delivered:
		t.Fatalf("500 updates to one action must produce one notification, also got %q", extra)
	case <-time.After(2 * time.Second):
	}
}

// TestRunNotifyLoop_DropsPayloadPostgresWillNeverAccept covers the failure the
// retry design would otherwise turn into a permanent stall. pg_notify rejects
// a payload of 8000 bytes or more, and that is not a connection error, so it
// fails identically on every attempt. Such a payload must be given up on, and
// must not hold up notifications that are fine.
func TestRunNotifyLoop_DropsPayloadPostgresWillNeverAccept(t *testing.T) {
	r, sqlDB, conn := newNotifyRepoWithDB(t)
	delivered := subscribeActions(t, r, 16)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	poison := strings.Repeat("x", 8000)
	r.notifyMu.Lock()
	markPending(r.pendingActions, poison)
	r.notifyMu.Unlock()

	go r.runNotifyLoop(ctx, sqlDB, conn)

	// Keep ordinary traffic flowing alongside the bad payload.
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			case <-time.After(20 * time.Millisecond):
				r.notifyActionUpdate(ctx, notifyTestActionID(fmt.Sprintf("healthy-%d", i)))
			}
		}
	}()

	select {
	case payload := <-delivered:
		assert.Contains(t, payload, "proj/domain/run/healthy-")
	case <-time.After(15 * time.Second):
		t.Fatal("a deliverable notification was held up behind an undeliverable one")
	}

	assert.Eventually(t, r.pendingDropped(poison), 15*time.Second, 20*time.Millisecond,
		"a payload that can never be delivered must be dropped, not retried forever")
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
	r := newNotifyTestRepo()

	// Queue a notification, then cancel so the loop exits after one attempt.
	r.notifyActionUpdate(context.Background(), notifyTestActionID("action"))

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Pass a nil conn: should not panic.
	assert.NotPanics(t, func() {
		r.runNotifyLoop(ctx, nil, nil)
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
		Run:  &common.RunIdentifier{Project: "p", Domain: "d", Name: "run-abort"},
		Name: "abort-action",
	}
	_, err = actionRepo.CreateAction(ctx, models.NewActionModel(actionID), false)
	require.NoError(t, err)

	endTime := time.Now()
	err = actionRepo.UpdateActionPhase(ctx, actionID, common.ActionPhase_ACTION_PHASE_ABORTED, 1, core.CatalogCacheStatus_CACHE_DISABLED, &endTime, nil)
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
