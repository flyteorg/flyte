package impl

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.Exec(`CREATE TABLE tasks (
		org TEXT NOT NULL,
		project TEXT NOT NULL,
		domain TEXT NOT NULL,
		name TEXT NOT NULL,
		version TEXT NOT NULL,
		environment TEXT,
		function_name TEXT,
		deployed_by TEXT,
		trigger_name TEXT,
		total_triggers INTEGER DEFAULT 0,
		active_triggers INTEGER DEFAULT 0,
		trigger_automation_spec BLOB,
		trigger_types INTEGER,
		task_spec BLOB,
		env_description TEXT,
		short_description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (org, project, domain, name, version)
	)`).Error
	require.NoError(t, err)

	return db
}

func TestCreateTask(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTaskRepo(db)
	ctx := context.Background()

	task := &models.Task{
		TaskKey: models.TaskKey{
			Org:     "test-org",
			Project: "test-project",
			Domain:  "test-domain",
			Name:    "test-task",
			Version: "v1",
		},
		Environment:  "production",
		FunctionName: "my_function",
		DeployedBy:   "user@example.com",
		TaskSpec:     []byte(`{"type": "python"}`),
	}

	err := repo.CreateTask(ctx, task)
	assert.NoError(t, err)

	retrieved, err := repo.GetTask(ctx, task.TaskKey)
	require.NoError(t, err)
	assert.Equal(t, task.Environment, retrieved.Environment)
	assert.Equal(t, task.FunctionName, retrieved.FunctionName)
    assert.False(t, retrieved.CreatedAt.IsZero(), "created_at should not be zero")
    assert.False(t, retrieved.UpdatedAt.IsZero(), "updated_at should not be zero")
}

func TestGetTask_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTaskRepo(db)
	ctx := context.Background()

	key := models.TaskKey{
		Org:     "non-existent",
		Project: "test",
		Domain:  "test",
		Name:    "test",
		Version: "v1",
	}

	_, err := repo.GetTask(ctx, key)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestListTasks(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTaskRepo(db)
	ctx := context.Background()

	task1 := &models.Task{
		TaskKey: models.TaskKey{
			Org:     "org1",
			Project: "proj1",
			Domain:  "domain1",
			Name:    "task1",
			Version: "v1",
		},
		Environment: "prod",
		TaskSpec:    []byte("{}"),
	}

	task2 := &models.Task{
		TaskKey: models.TaskKey{
			Org:     "org1",
			Project: "proj1",
			Domain:  "domain1",
			Name:    "task2",
			Version: "v1",
		},
		Environment: "dev",
		TaskSpec:    []byte("{}"),
	}

	require.NoError(t, repo.CreateTask(ctx, task1))
	require.NoError(t, repo.CreateTask(ctx, task2))

	result, err := repo.ListTasks(ctx, interfaces.ListResourceInput{
		Limit:  10,
		Offset: 0,
	})

	require.NoError(t, err)
	assert.Len(t, result.Tasks, 2)
	assert.Equal(t, uint32(2), result.Total)
}

func TestCreateTaskSpec(t *testing.T) {
	db := setupTestDB(t)

	err := db.Exec(`CREATE TABLE task_specs (
		digest TEXT PRIMARY KEY,
		spec BLOB NOT NULL
	)`).Error
	require.NoError(t, err)

	repo := NewTaskRepo(db)
	ctx := context.Background()

	spec := &models.TaskSpec{
		Digest: "abc123",
		Spec:   []byte(`{"task": "spec"}`),
	}

	err = repo.CreateTaskSpec(ctx, spec)
	assert.NoError(t, err)

	retrieved, err := repo.GetTaskSpec(ctx, "abc123")
	require.NoError(t, err)
	assert.Equal(t, spec.Digest, retrieved.Digest)
	assert.Equal(t, spec.Spec, retrieved.Spec)
}

func TestCreateTask_Timestamps(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTaskRepo(db)
	ctx := context.Background()

	task := &models.Task{
		TaskKey: models.TaskKey{
			Org:     "test-org",
			Project: "test-project",
			Domain:  "test-domain",
			Name:    "test-task",
			Version: "v1",
		},
		Environment:  "production",
		FunctionName: "my_function",
		TaskSpec:     []byte(`{}`),
	}

	err := repo.CreateTask(ctx, task)
	assert.NoError(t, err)

	retrieved, err := repo.GetTask(ctx, task.TaskKey)
	require.NoError(t, err)

	now := time.Now()
	assert.True(t, retrieved.CreatedAt.After(now.Add(-5*time.Second)))
	assert.True(t, retrieved.CreatedAt.Before(now.Add(time.Second)))

	diff := retrieved.UpdatedAt.Sub(retrieved.CreatedAt)
	assert.True(t, diff >= 0 && diff < time.Second)
}

func TestCreateTask_UpdatePreservesCreatedAt(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTaskRepo(db)
	ctx := context.Background()

	task := &models.Task{
		TaskKey: models.TaskKey{
			Org:     "test-org",
			Project: "test-project",
			Domain:  "test-domain",
			Name:    "test-task",
			Version: "v1",
		},
		Environment:  "production",
		FunctionName: "original_function",
		TaskSpec:     []byte(`{}`),
	}

	err := repo.CreateTask(ctx, task)
	require.NoError(t, err)

	original, err := repo.GetTask(ctx, task.TaskKey)
	require.NoError(t, err)

	originalCreatedAt := original.CreatedAt
	originalUpdatedAt := original.UpdatedAt

	time.Sleep(100 * time.Millisecond)

	task.FunctionName = "updated_function"
	err = repo.CreateTask(ctx, task)
	require.NoError(t, err)

	updated, err := repo.GetTask(ctx, task.TaskKey)
	require.NoError(t, err)

	assert.Equal(t, originalCreatedAt, updated.CreatedAt)
	assert.True(t, updated.UpdatedAt.After(originalUpdatedAt))
	assert.Equal(t, "updated_function", updated.FunctionName)
}
