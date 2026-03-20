package impl

import (
	"context"
	"testing"

	"github.com/flyteorg/flyte/v2/flytestdlib/database"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		ended_at DATETIME,
		duration_ms INTEGER,
		attempts INTEGER NOT NULL DEFAULT 1,
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
