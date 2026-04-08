package impl

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

func setupDB(t *testing.T) *sqlx.DB {
	t.Helper()
	return testDB
}

func setupTestDB(t *testing.T) *sqlx.DB {
	db := setupDB(t)
	t.Cleanup(func() {
		db.Exec("DELETE FROM task_specs")
		db.Exec("DELETE FROM tasks")
	})
	return db
}

func TestCreateTask(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTaskRepo(db)
	ctx := context.Background()

	task := &models.Task{
		TaskKey: models.TaskKey{
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

	startTime := time.Now()
	err := repo.CreateTask(ctx, task, nil)
	assert.NoError(t, err)

	retrieved, err := repo.GetTask(ctx, task.TaskKey)
	require.NoError(t, err)
	assert.Equal(t, task.Environment, retrieved.Environment)
	assert.Equal(t, task.FunctionName, retrieved.FunctionName)
	assert.WithinDuration(t, startTime, retrieved.CreatedAt, 1*time.Second, "created_at should be close to now")
	assert.Equal(t, retrieved.CreatedAt, retrieved.UpdatedAt, "created_at and updated_at should be equal")
}

func TestGetTask_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTaskRepo(db)
	ctx := context.Background()

	key := models.TaskKey{
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
			Project: "proj1",
			Domain:  "domain1",
			Name:    "task2",
			Version: "v1",
		},
		Environment: "dev",
		TaskSpec:    []byte("{}"),
	}

	require.NoError(t, repo.CreateTask(ctx, task1, nil))
	require.NoError(t, repo.CreateTask(ctx, task2, nil))

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
	repo := NewTaskRepo(db)
	ctx := context.Background()

	spec := &models.TaskSpec{
		Digest: "abc123",
		Spec:   []byte(`{"task": "spec"}`),
	}

	err := repo.CreateTaskSpec(ctx, spec)
	assert.NoError(t, err)

	retrieved, err := repo.GetTaskSpec(ctx, "abc123")
	require.NoError(t, err)
	assert.Equal(t, spec.Digest, retrieved.Digest)
	assert.Equal(t, spec.Spec, retrieved.Spec)
}

func TestCreateTask_UpdatePreservesCreatedAt(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTaskRepo(db)
	ctx := context.Background()

	task := &models.Task{
		TaskKey: models.TaskKey{
			Project: "test-project",
			Domain:  "test-domain",
			Name:    "test-task",
			Version: "v1",
		},
		Environment:  "production",
		FunctionName: "original_function",
		TaskSpec:     []byte(`{}`),
	}

	err := repo.CreateTask(ctx, task, nil)
	require.NoError(t, err)

	original, err := repo.GetTask(ctx, task.TaskKey)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	task.FunctionName = "updated_function"
	err = repo.CreateTask(ctx, task, nil)
	require.NoError(t, err)

	updated, err := repo.GetTask(ctx, task.TaskKey)
	require.NoError(t, err)

	assert.Equal(t, original.CreatedAt, updated.CreatedAt, "create time shouldn't change")
	assert.True(t, updated.UpdatedAt.After(original.UpdatedAt))
	assert.Equal(t, "updated_function", updated.FunctionName)
}
