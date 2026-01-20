package transformers

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

func TestToTaskKey(t *testing.T) {
	taskId := &task.TaskIdentifier{
		Org:     "test-org",
		Project: "test-project",
		Domain:  "test-domain",
		Name:    "test-task",
		Version: "v1",
	}

	key := ToTaskKey(taskId)
	assert.Equal(t, "test-org", key.Org)
	assert.Equal(t, "test-project", key.Project)
	assert.Equal(t, "test-domain", key.Domain)
	assert.Equal(t, "test-task", key.Name)
	assert.Equal(t, "v1", key.Version)
}

func TestToTaskName(t *testing.T) {
	taskName := &task.TaskName{
		Org:     "test-org",
		Project: "test-project",
		Domain:  "test-domain",
		Name:    "test-task",
	}

	name := ToTaskName(taskName)
	assert.Equal(t, "test-org", name.Org)
	assert.Equal(t, "test-project", name.Project)
	assert.Equal(t, "test-domain", name.Domain)
	assert.Equal(t, "test-task", name.Name)
}

func TestNewTaskModel(t *testing.T) {
	ctx := context.Background()
	taskId := &task.TaskIdentifier{
		Org:     "test-org",
		Project: "test-project",
		Domain:  "test-domain",
		Name:    "prod.my_function",
		Version: "v1",
	}
	spec := &task.TaskSpec{
		Environment: &task.Environment{
			Name:        "prod",
			Description: "Production environment",
		},
		Documentation: &task.DocumentationEntity{
			ShortDescription: "A test task",
		},
	}

	model, err := NewTaskModel(ctx, taskId, spec)
	require.NoError(t, err)
	assert.Equal(t, "test-org", model.Org)
	assert.Equal(t, "prod", model.Environment)
	assert.Equal(t, "my_function", model.FunctionName)
	assert.Equal(t, "mock-subject", model.DeployedBy)
	assert.True(t, model.EnvDescription.Valid)
	assert.Equal(t, "Production environment", model.EnvDescription.String)
	assert.True(t, model.ShortDescription.Valid)
	assert.Equal(t, "A test task", model.ShortDescription.String)
}

func TestExtractFunctionName_WithEnvironment(t *testing.T) {
	ctx := context.Background()
	functionName := ExtractFunctionName(ctx, "prod.my_function", "prod")
	assert.Equal(t, "my_function", functionName)
}

func TestExtractFunctionName_WithoutEnvironment(t *testing.T) {
	ctx := context.Background()
	functionName := ExtractFunctionName(ctx, "my.nested.function", "")
	assert.Equal(t, "function", functionName)
}

func TestExtractFunctionName_Fallback(t *testing.T) {
	ctx := context.Background()
	functionName := ExtractFunctionName(ctx, "prod.my_function", "staging")
	assert.Equal(t, "my_function", functionName)
}

func TestTaskModelsToTaskDetails(t *testing.T) {
	ctx := context.Background()
	spec := &task.TaskSpec{
		Environment: &task.Environment{
			Name: "prod",
		},
	}
	specBytes, _ := proto.Marshal(spec)

	taskModels := []*models.Task{
		{
			TaskKey: models.TaskKey{
				Org:     "org1",
				Project: "proj1",
				Domain:  "domain1",
				Name:    "task1",
				Version: "v1",
			},
			Environment:      "prod",
			FunctionName:     "my_func",
			DeployedBy:       "user@example.com",
			TaskSpec:         specBytes,
			CreatedAt:        time.Now(),
			ShortDescription: sql.NullString{String: "Test task", Valid: true},
		},
	}

	details, err := TaskModelsToTaskDetails(ctx, taskModels)
	require.NoError(t, err)
	require.Len(t, details, 1)
	assert.Equal(t, "org1", details[0].TaskId.Org)
	assert.Equal(t, "my_func", details[0].Metadata.ShortName)
	assert.Equal(t, "prod", details[0].Metadata.EnvironmentName)
	assert.Equal(t, "Test task", details[0].Metadata.ShortDescription)
	assert.NotNil(t, details[0].Spec)
}

func TestTaskModelsToTasks(t *testing.T) {
	ctx := context.Background()
	taskModels := []*models.Task{
		{
			TaskKey: models.TaskKey{
				Org:     "org1",
				Project: "proj1",
				Domain:  "domain1",
				Name:    "task1",
				Version: "v1",
			},
			Environment:  "prod",
			FunctionName: "my_func",
			CreatedAt:    time.Now(),
		},
	}

	tasks, err := TaskModelsToTasks(ctx, taskModels, nil)
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, "org1", tasks[0].TaskId.Org)
	assert.Equal(t, "my_func", tasks[0].Metadata.ShortName)
	assert.NotNil(t, tasks[0].TaskSummary)
}

func TestTaskTriggersSummaryFromModel_NoTriggers(t *testing.T) {
	taskModel := &models.Task{
		TotalTriggers: 0,
	}

	summary := taskTriggersSummaryFromModel(taskModel)
	assert.Nil(t, summary)
}

func TestTaskTriggersSummaryFromModel_SingleTrigger(t *testing.T) {
	taskModel := &models.Task{
		TotalTriggers:  1,
		ActiveTriggers: 1,
		TriggerName:    sql.NullString{String: "my-trigger", Valid: true},
	}

	summary := taskTriggersSummaryFromModel(taskModel)
	require.NotNil(t, summary)
	details := summary.GetDetails()
	require.NotNil(t, details)
	assert.Equal(t, "my-trigger", details.Name)
	assert.True(t, details.Active)
}

func TestTaskTriggersSummaryFromModel_MultipleTriggers(t *testing.T) {
	taskModel := &models.Task{
		TotalTriggers:  5,
		ActiveTriggers: 3,
	}

	summary := taskTriggersSummaryFromModel(taskModel)
	require.NotNil(t, summary)
	stats := summary.GetStats()
	require.NotNil(t, stats)
	assert.Equal(t, uint32(5), stats.Total)
	assert.Equal(t, uint32(3), stats.Active)
}

func TestVersionModelsToVersionResponses(t *testing.T) {
	now := time.Now()
	versionModels := []*models.TaskVersion{
		{
			Version:   "v1",
			CreatedAt: now,
		},
		{
			Version:   "v2",
			CreatedAt: now.Add(time.Hour),
		},
	}

	responses := VersionModelsToVersionResponses(versionModels)
	require.Len(t, responses, 2)
	assert.Equal(t, "v1", responses[0].Version)
	assert.Equal(t, "v2", responses[1].Version)
	assert.NotNil(t, responses[0].DeployedAt)
	assert.NotNil(t, responses[1].DeployedAt)
}

func TestNewNullString(t *testing.T) {
	validStr := newNullString("test")
	assert.True(t, validStr.Valid)
	assert.Equal(t, "test", validStr.String)

	invalidStr := newNullString("")
	assert.False(t, invalidStr.Valid)
}

func TestTaskListResultToTasksAndMetadata(t *testing.T) {
	ctx := context.Background()
	result := &models.TaskListResult{
		Tasks: []*models.Task{
			{
				TaskKey: models.TaskKey{
					Org:     "org1",
					Project: "proj1",
					Domain:  "domain1",
					Name:    "task1",
					Version: "v1",
				},
				FunctionName: "func1",
				CreatedAt:    time.Now(),
			},
		},
		Total:         10,
		FilteredTotal: 1,
	}

	tasks, metadata, err := TaskListResultToTasksAndMetadata(ctx, result, nil, nil, false, false)
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, "func1", tasks[0].Metadata.ShortName)
	require.NotNil(t, metadata)
	assert.Equal(t, uint32(10), metadata.Total)
	assert.Equal(t, uint32(1), metadata.FilteredTotal)
}
