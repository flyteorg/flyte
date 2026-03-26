package impl

import (
	"context"
	"testing"
	"time"

	"github.com/flyteorg/flyte/v2/flytestdlib/database"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

func setupActionDB(t *testing.T) *gorm.DB {
	db := setupDB(t)

	var err error
	err = db.Exec(`CREATE TABLE actions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		org TEXT NOT NULL,
		project TEXT NOT NULL,
		domain TEXT NOT NULL,
		run_name TEXT NOT NULL DEFAULT '',
		name TEXT NOT NULL,
		parent_action_name TEXT,
		phase INTEGER NOT NULL DEFAULT 1,
		run_source TEXT NOT NULL DEFAULT '',
		action_type INTEGER NOT NULL DEFAULT 0,
		action_group TEXT,
		task_org TEXT,
		task_project TEXT,
		task_domain TEXT,
		task_name TEXT,
		task_version TEXT,
		task_type TEXT NOT NULL DEFAULT '',
		task_short_name TEXT,
		function_name TEXT NOT NULL DEFAULT '',
		environment_name TEXT,
		action_spec BLOB,
		action_details BLOB,
		detailed_info BLOB,
		run_spec BLOB,
		abort_requested_at DATETIME,
		abort_attempt_count INTEGER NOT NULL DEFAULT 0,
		abort_reason TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		started_at DATETIME,
		ended_at DATETIME,
		duration_ms INTEGER,
		attempts INTEGER NOT NULL DEFAULT 1,
		cache_status INTEGER NOT NULL DEFAULT 0,
		CONSTRAINT uq_actions_identifier UNIQUE (org, project, domain, name)
	)`).Error
	require.NoError(t, err)
	err = db.Exec(`CREATE INDEX idx_actions_org ON actions(org)`).Error
	require.NoError(t, err)
	err = db.Exec(`CREATE INDEX idx_actions_project ON actions(project)`).Error
	require.NoError(t, err)
	err = db.Exec(`CREATE INDEX idx_actions_domain ON actions(domain)`).Error
	require.NoError(t, err)
	err = db.Exec(`CREATE INDEX idx_actions_run_name ON actions(run_name)`).Error
	require.NoError(t, err)
	err = db.Exec(`CREATE INDEX idx_actions_parent ON actions(parent_action_name)`).Error
	require.NoError(t, err)
	err = db.Exec(`CREATE INDEX idx_actions_phase ON actions(phase)`).Error
	require.NoError(t, err)
	err = db.Exec(`CREATE INDEX idx_actions_created ON actions(created_at)`).Error
	require.NoError(t, err)
	err = db.Exec(`CREATE INDEX idx_actions_updated ON actions(updated_at)`).Error
	require.NoError(t, err)
	err = db.Exec(`CREATE INDEX idx_actions_ended ON actions(ended_at)`).Error
	require.NoError(t, err)

	return db
}

func TestCreateRun(t *testing.T) {
	db := setupActionDB(t)
	defer func() { _ = db.Exec("DELETE FROM actions") }()
	actionRepo := NewActionRepo(db, database.DbConfig{})
	ctx := context.Background()

	runID := &common.RunIdentifier{
		Org:     "org1",
		Project: "proj1",
		Domain:  "domain1",
		Name:    "run1",
	}
	req := &workflow.CreateRunRequest{
		Id:      &workflow.CreateRunRequest_RunId{RunId: runID},
		RunSpec: nil,
		Task:    &workflow.CreateRunRequest_TaskId{TaskId: &task.TaskIdentifier{Name: "task1"}},
	}

	run, err := actionRepo.CreateRun(ctx, req, "s3://input", "s3://output")
	require.NoError(t, err)
	require.NotNil(t, run)
	assert.Equal(t, runID.Org, run.Org)
	assert.Equal(t, runID.Project, run.Project)
	assert.Equal(t, runID.Domain, run.Domain)
	assert.Equal(t, runID.Name, run.RunName)
	assert.Equal(t, runID.Name, run.Name)
	assert.Equal(t, int32(common.ActionPhase_ACTION_PHASE_QUEUED), run.Phase)
	require.NotZero(t, run.ID)
	require.NotEmpty(t, run.ActionSpec)

	// Attempt duplicate run create with same run name should fail unique constraint
	_, err = actionRepo.CreateRun(ctx, req, "s3://input2", "s3://output2")
	require.Error(t, err)
}

func TestUpdateActionPhasePersistsAttemptsAndCacheStatus(t *testing.T) {
	db := setupActionDB(t)
	defer func() { _ = db.Exec("DELETE FROM actions") }()
	actionRepo := NewActionRepo(db, database.DbConfig{})
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

	_, err := actionRepo.CreateAction(ctx, &workflow.ActionSpec{
		ActionId: actionID,
		InputUri: "s3://bucket/input",
	}, nil)
	require.NoError(t, err)

	endTime := time.Now()
	err = actionRepo.UpdateActionPhase(
		ctx,
		actionID,
		common.ActionPhase_ACTION_PHASE_SUCCEEDED,
		3,
		core.CatalogCacheStatus_CACHE_HIT,
		nil,
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

func TestUpdateActionPhase_SetsStartedAtOnExecutionPhase(t *testing.T) {
	// Tasks may skip RUNNING and go QUEUED → INITIALIZING → terminal.
	// started_at should be set on any execution phase (INITIALIZING, RUNNING, etc.)
	// when a non-nil startTime is provided.
	for _, phase := range []common.ActionPhase{
		common.ActionPhase_ACTION_PHASE_INITIALIZING,
		common.ActionPhase_ACTION_PHASE_RUNNING,
	} {
		t.Run(phase.String(), func(t *testing.T) {
			db := setupActionDB(t)
			defer func() { _ = db.Exec("DELETE FROM actions") }()
			actionRepo := NewActionRepo(db, database.DbConfig{})
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

			_, err := actionRepo.CreateAction(ctx, &workflow.ActionSpec{
				ActionId: actionID,
				InputUri: "s3://bucket/input",
			}, nil)
			require.NoError(t, err)

			// Transition to execution phase with an explicit startTime — should set started_at
			startTime := time.Now().Add(-10 * time.Second)
			err = actionRepo.UpdateActionPhase(ctx, actionID,
				phase, 1,
				core.CatalogCacheStatus_CACHE_DISABLED, &startTime, nil)
			require.NoError(t, err)

			action, err := actionRepo.GetAction(ctx, actionID)
			require.NoError(t, err)
			assert.True(t, action.StartedAt.Valid, "started_at should be set after %s transition with non-nil startTime", phase)
			startedAt := action.StartedAt.Time

			// Transition to SUCCEEDED with an endTime — duration should be based on started_at
			endTime := startedAt.Add(5 * time.Second)
			err = actionRepo.UpdateActionPhase(ctx, actionID,
				common.ActionPhase_ACTION_PHASE_SUCCEEDED, 1,
				core.CatalogCacheStatus_CACHE_DISABLED, nil, &endTime)
			require.NoError(t, err)

			action, err = actionRepo.GetAction(ctx, actionID)
			require.NoError(t, err)
			assert.True(t, action.DurationMs.Valid)
			// Duration should be ~5000ms (endTime - startedAt), not endTime - createdAt
			assert.InDelta(t, 5000, action.DurationMs.Int64, 1000,
				"duration should reflect execution time (started_at to ended_at), not queue time")
		})
	}
}

func TestUpdateActionPhase_NilStartTimeDoesNotSetStartedAt(t *testing.T) {
	// When startTime is nil (e.g. from the fast recordEvents path), started_at
	// should NOT be set — we avoid time.Now() fallbacks that can produce
	// started_at values after the actual pod end time.
	for _, phase := range []common.ActionPhase{
		common.ActionPhase_ACTION_PHASE_INITIALIZING,
		common.ActionPhase_ACTION_PHASE_RUNNING,
		common.ActionPhase_ACTION_PHASE_SUCCEEDED,
	} {
		t.Run(phase.String(), func(t *testing.T) {
			db := setupActionDB(t)
			defer func() { _ = db.Exec("DELETE FROM actions") }()
			actionRepo := NewActionRepo(db, database.DbConfig{})
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

			_, err := actionRepo.CreateAction(ctx, &workflow.ActionSpec{
				ActionId: actionID,
				InputUri: "s3://bucket/input",
			}, nil)
			require.NoError(t, err)

			// Pass nil startTime — started_at should remain NULL
			err = actionRepo.UpdateActionPhase(ctx, actionID,
				phase, 1,
				core.CatalogCacheStatus_CACHE_DISABLED, nil, nil)
			require.NoError(t, err)

			action, err := actionRepo.GetAction(ctx, actionID)
			require.NoError(t, err)
			assert.False(t, action.StartedAt.Valid,
				"started_at should NOT be set when startTime is nil (phase=%s)", phase)
		})
	}
}

func TestUpdateActionPhase_StartedAtNotSetOnWaitingPhases(t *testing.T) {
	db := setupActionDB(t)
	defer func() { _ = db.Exec("DELETE FROM actions") }()
	actionRepo := NewActionRepo(db, database.DbConfig{})
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

	_, err := actionRepo.CreateAction(ctx, &workflow.ActionSpec{
		ActionId: actionID,
		InputUri: "s3://bucket/input",
	}, nil)
	require.NoError(t, err)

	// QUEUED should not set started_at
	err = actionRepo.UpdateActionPhase(ctx, actionID,
		common.ActionPhase_ACTION_PHASE_QUEUED, 1,
		core.CatalogCacheStatus_CACHE_DISABLED, nil)
	require.NoError(t, err)

	action, err := actionRepo.GetAction(ctx, actionID)
	require.NoError(t, err)
	assert.False(t, action.StartedAt.Valid, "started_at should not be set for QUEUED phase")
}

func TestUpdateActionPhase_StartedAtNotOverwritten(t *testing.T) {
	db := setupActionDB(t)
	defer func() { _ = db.Exec("DELETE FROM actions") }()
	actionRepo := NewActionRepo(db, database.DbConfig{})
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

	_, err := actionRepo.CreateAction(ctx, &workflow.ActionSpec{
		ActionId: actionID,
		InputUri: "s3://bucket/input",
	}, nil)
	require.NoError(t, err)

	// First RUNNING transition with explicit startTime
	firstStart := time.Now().Add(-10 * time.Second)
	err = actionRepo.UpdateActionPhase(ctx, actionID,
		common.ActionPhase_ACTION_PHASE_RUNNING, 1,
		core.CatalogCacheStatus_CACHE_DISABLED, &firstStart, nil)
	require.NoError(t, err)

	action, err := actionRepo.GetAction(ctx, actionID)
	require.NoError(t, err)
	assert.True(t, action.StartedAt.Valid)
	firstStartedAt := action.StartedAt.Time

	// Second RUNNING transition (e.g. retry) with a different startTime should
	// NOT overwrite started_at thanks to COALESCE(started_at, ?)
	err = actionRepo.UpdateActionPhase(ctx, actionID,
		common.ActionPhase_ACTION_PHASE_RUNNING, 2,
		core.CatalogCacheStatus_CACHE_DISABLED, nil)
	require.NoError(t, err)

	action, err = actionRepo.GetAction(ctx, actionID)
	require.NoError(t, err)
	assert.Equal(t, firstStartedAt, action.StartedAt.Time,
		"started_at should not be overwritten on subsequent RUNNING transitions")
}

func TestListRuns(t *testing.T) {
	db := setupActionDB(t)
	defer func() { _ = db.Exec("DELETE FROM actions") }()
	actionRepo := NewActionRepo(db, database.DbConfig{})
	ctx := context.Background()

	runsToCreate := []string{"run-1", "run-2", "run-3"}
	for _, runName := range runsToCreate {
		req := &workflow.CreateRunRequest{
			Id: &workflow.CreateRunRequest_RunId{RunId: &common.RunIdentifier{
				Org:     "org1",
				Project: "proj1",
				Domain:  "domain1",
				Name:    runName,
			}},
			RunSpec: nil,
		}
		_, err := actionRepo.CreateRun(ctx, req, "in://uri", "out://base")
		require.NoError(t, err)
	}

	// Table-driven list tests for ListRuns
	type listTestCase struct {
		name           string
		req            *workflow.ListRunsRequest
		expectLen      int
		expectTokenNil bool
		verify         func(t *testing.T, runs []*models.Run)
	}

	listTests := []listTestCase{
		{
			name:           "List by org should return 3 runs",
			req:            &workflow.ListRunsRequest{ScopeBy: &workflow.ListRunsRequest_Org{Org: "org1"}},
			expectLen:      3,
			expectTokenNil: true,
			verify: func(t *testing.T, runs []*models.Run) {
				runNames := map[string]bool{}
				for _, r := range runs {
					runNames[r.Name] = true
				}
				assert.True(t, runNames["run-1"])
				assert.True(t, runNames["run-2"])
				assert.True(t, runNames["run-3"])
			},
		},
	}

	for _, tt := range listTests {
		t.Run(tt.name, func(t *testing.T) {
			runs, nextToken, err := actionRepo.ListRuns(ctx, tt.req)
			require.NoError(t, err)
			assert.Len(t, runs, tt.expectLen)
			if tt.expectTokenNil {
				assert.Empty(t, nextToken)
			} else {
				assert.NotEmpty(t, nextToken)
			}
			if tt.verify != nil {
				tt.verify(t, runs)
			}
		})
	}

	// Pagination with limit and token results
	runsPage1, token1, err := actionRepo.ListRuns(ctx, &workflow.ListRunsRequest{
		Request: &common.ListRequest{Limit: 2},
		ScopeBy: &workflow.ListRunsRequest_Org{Org: "org1"},
	})
	require.NoError(t, err)
	assert.Len(t, runsPage1, 2)
	require.NotEmpty(t, token1)

	runsPage2, token2, err := actionRepo.ListRuns(ctx, &workflow.ListRunsRequest{
		Request: &common.ListRequest{Token: token1, Limit: 2},
		ScopeBy: &workflow.ListRunsRequest_Org{Org: "org1"},
	})
	require.NoError(t, err)
	assert.Len(t, runsPage2, 1)
	assert.Empty(t, token2)

	// Test project scope filtering doesn't include other org/project/domain
	_, err = actionRepo.CreateRun(ctx, &workflow.CreateRunRequest{
		Id: &workflow.CreateRunRequest_RunId{RunId: &common.RunIdentifier{
			Org:     "other-org",
			Project: "other-proj",
			Domain:  "domain1",
			Name:    "run-other",
		}},
	}, "in://x", "out://x")
	require.NoError(t, err)

	runsFiltered, _, err := actionRepo.ListRuns(ctx, &workflow.ListRunsRequest{
		ScopeBy: &workflow.ListRunsRequest_ProjectId{ProjectId: &common.ProjectIdentifier{
			Organization: "org1",
			Name:         "proj1",
			Domain:       "domain1",
		}},
	})
	require.NoError(t, err)
	assert.Len(t, runsFiltered, 3)
	for _, r := range runsFiltered {
		assert.Equal(t, "org1", r.Org)
		assert.Equal(t, "proj1", r.Project)
		assert.Equal(t, "domain1", r.Domain)
	}
}

func setupActionEventDB(t *testing.T) (*gorm.DB, *actionRepo) {
	db := setupActionDB(t)
	err := db.Exec(`CREATE TABLE action_events (
		org TEXT NOT NULL,
		project TEXT NOT NULL,
		domain TEXT NOT NULL,
		run_name TEXT NOT NULL,
		name TEXT NOT NULL,
		attempt INTEGER NOT NULL,
		phase INTEGER NOT NULL,
		version INTEGER NOT NULL,
		info BLOB,
		error_kind TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (org, project, domain, run_name, name, attempt, phase, version)
	)`).Error
	require.NoError(t, err)

	repo := NewActionRepo(db, database.DbConfig{}).(*actionRepo)
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
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
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
