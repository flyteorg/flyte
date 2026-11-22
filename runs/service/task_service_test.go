package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	triggerpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/trigger"
	mocks "github.com/flyteorg/flyte/v2/runs/repository/mocks"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

func cronAutomation(kickoffArg string) *task.TriggerAutomationSpec {
	return &task.TriggerAutomationSpec{
		Automation: &task.TriggerAutomationSpec_Schedule{
			Schedule: &task.Schedule{
				Expression:          &task.Schedule_Cron{Cron: &task.Cron{Expression: "0 0 * * *"}},
				KickoffTimeInputArg: kickoffArg,
			},
		},
	}
}

// buildTriggerModels maps an offloaded embedded trigger's inputs into the TriggerSpec input_wrapper.
func TestBuildTriggerModels_OffloadedInputs(t *testing.T) {
	taskId := &task.TaskIdentifier{Project: "p", Domain: "d", Name: "t", Version: "v1"}
	offloaded := &common.OffloadedInputData{Uri: "offloaded-inputs/hash123", InputsHash: "hash123"}
	tts := []*task.TaskTrigger{{
		Name:           "trig",
		Spec:           &task.TaskTriggerSpec{InputWrapper: &task.TaskTriggerSpec_OffloadedInputData{OffloadedInputData: offloaded}},
		AutomationSpec: cronAutomation(""),
	}}

	triggerModels, err := buildTriggerModels(context.Background(), taskId, tts)
	require.NoError(t, err)
	require.Len(t, triggerModels, 1)

	var spec triggerpb.TriggerSpec
	require.NoError(t, proto.Unmarshal(triggerModels[0].Spec, &spec))
	assert.Equal(t, "offloaded-inputs/hash123", spec.GetOffloadedInputData().GetUri())
	assert.Nil(t, spec.GetInputs())
}

func TestDeployTask(t *testing.T) {
	ctx := context.Background()
	mockTaskRepo := mocks.NewTaskRepo(t)

	mockTaskRepo.EXPECT().CreateTask(ctx, mock.AnythingOfType("*models.Task"), mock.Anything).Return(nil)
	projectRepo := mocks.NewProjectRepo(t)
	projectRepo.EXPECT().GetProject(ctx, "test-project").Return(&models.Project{Identifier: "test-project"}, nil).Once()
	service := NewTaskService(mockTaskRepo, NewProjectService(projectRepo, nil))

	req := &task.DeployTaskRequest{
		TaskId: &task.TaskIdentifier{
			Org:     "test-org",
			Project: "test-project",
			Domain:  "test-domain",
			Name:    "test-task",
			Version: "v1",
		},
		Spec: &task.TaskSpec{
			Environment: &task.Environment{
				Name: "prod",
			},
		},
	}

	resp, err := service.DeployTask(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetTaskDetails(t *testing.T) {
	ctx := context.Background()
	mockTaskRepo := mocks.NewTaskRepo(t)

	spec := &task.TaskSpec{
		Environment: &task.Environment{Name: "prod"},
	}
	specBytes, _ := proto.Marshal(spec)

	taskModel := &models.Task{
		TaskKey: models.TaskKey{
			Project: "test-project",
			Domain:  "test-domain",
			Name:    "test-task",
			Version: "v1",
		},
		Environment:  "prod",
		FunctionName: "test_func",
		TaskSpec:     specBytes,
		CreatedAt:    time.Now(),
	}

	mockTaskRepo.EXPECT().GetTask(ctx, mock.AnythingOfType("models.TaskKey")).Return(taskModel, nil)

	service := NewTaskService(mockTaskRepo, nil)

	req := &task.GetTaskDetailsRequest{
		TaskId: &task.TaskIdentifier{
			Org:     "test-org",
			Project: "test-project",
			Domain:  "test-domain",
			Name:    "test-task",
			Version: "v1",
		},
	}

	resp, err := service.GetTaskDetails(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Details)
	assert.Equal(t, "", resp.Details.TaskId.Org)
}

func TestListTasks(t *testing.T) {
	ctx := context.Background()
	mockTaskRepo := mocks.NewTaskRepo(t)

	taskModels := []*models.Task{
		{
			TaskKey: models.TaskKey{
				Project: "test-project",
				Domain:  "test-domain",
				Name:    "task1",
				Version: "v1",
			},
			FunctionName: "func1",
			CreatedAt:    time.Now(),
		},
	}

	result := &models.TaskListResult{
		Tasks:         taskModels,
		Total:         1,
		FilteredTotal: 1,
	}

	mockTaskRepo.EXPECT().ListTasks(ctx, mock.AnythingOfType("interfaces.ListResourceInput")).Return(result, nil)

	service := NewTaskService(mockTaskRepo, nil)

	req := &task.ListTasksRequest{
		ScopeBy: &task.ListTasksRequest_Org{
			Org: "test-org",
		},
		Request: &common.ListRequest{
			Limit: 10,
		},
	}

	resp, err := service.ListTasks(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Tasks, 1)
	assert.NotNil(t, resp.Metadata)
	assert.Equal(t, uint32(1), resp.Metadata.Total)
}

func TestListVersions(t *testing.T) {
	ctx := context.Background()
	mockTaskRepo := mocks.NewTaskRepo(t)

	now := time.Now()
	versions := []*models.TaskVersion{
		{
			Version:   "v1",
			CreatedAt: now,
		},
		{
			Version:   "v2",
			CreatedAt: now.Add(time.Hour),
		},
	}

	mockTaskRepo.EXPECT().ListVersions(ctx, mock.AnythingOfType("interfaces.ListResourceInput")).Return(versions, nil)

	service := NewTaskService(mockTaskRepo, nil)

	req := &task.ListVersionsRequest{
		TaskName: &task.TaskName{
			Org:     "test-org",
			Project: "test-project",
			Domain:  "test-domain",
			Name:    "test-task",
		},
		Request: &common.ListRequest{
			Limit: 10,
		},
	}

	resp, err := service.ListVersions(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Versions, 2)
	assert.Equal(t, "v1", resp.Versions[0].Version)
	assert.Equal(t, "v2", resp.Versions[1].Version)
}

func TestTruncateShortDescription_UnderLimit(t *testing.T) {
	desc := "Short description"
	result := truncateShortDescription(desc)
	assert.Equal(t, desc, result)
}

func TestTruncateShortDescription_AtLimit(t *testing.T) {
	desc := strings.Repeat("a", 255)
	result := truncateShortDescription(desc)
	assert.Equal(t, desc, result)
	assert.Len(t, result, 255)
}

func TestTruncateShortDescription_OverLimit(t *testing.T) {
	desc := strings.Repeat("a", 300)
	result := truncateShortDescription(desc)
	assert.Len(t, result, 255)
	assert.Equal(t, strings.Repeat("a", 255), result)
}

func TestTruncateLongDescription_UnderLimit(t *testing.T) {
	desc := "Long description"
	result := truncateLongDescription(desc)
	assert.Equal(t, desc, result)
}

func TestTruncateLongDescription_AtLimit(t *testing.T) {
	desc := strings.Repeat("b", 2048)
	result := truncateLongDescription(desc)
	assert.Equal(t, desc, result)
	assert.Len(t, result, 2048)
}

func TestTruncateLongDescription_OverLimit(t *testing.T) {
	desc := strings.Repeat("b", 3000)
	result := truncateLongDescription(desc)
	assert.Len(t, result, 2048)
	assert.Equal(t, strings.Repeat("b", 2048), result)
}
