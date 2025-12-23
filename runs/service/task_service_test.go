package service

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	mocks "github.com/flyteorg/flyte/v2/runs/repository/mocks"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

func TestDeployTask(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewRepository(t)
	mockTaskRepo := mocks.NewTaskRepo(t)

	mockRepo.EXPECT().TaskRepo().Return(mockTaskRepo)
	mockTaskRepo.EXPECT().CreateTask(ctx, mock.AnythingOfType("*models.Task")).Return(nil)

	service := NewTaskService(mockRepo)

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

	resp, err := service.DeployTask(ctx, connect.NewRequest(req))
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetTaskDetails(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewRepository(t)
	mockTaskRepo := mocks.NewTaskRepo(t)

	spec := &task.TaskSpec{
		Environment: &task.Environment{Name: "prod"},
	}
	specBytes, _ := proto.Marshal(spec)

	taskModel := &models.Task{
		TaskKey: models.TaskKey{
			Org:     "test-org",
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

	mockRepo.EXPECT().TaskRepo().Return(mockTaskRepo)
	mockTaskRepo.EXPECT().GetTask(ctx, mock.AnythingOfType("models.TaskKey")).Return(taskModel, nil)

	service := NewTaskService(mockRepo)

	req := &task.GetTaskDetailsRequest{
		TaskId: &task.TaskIdentifier{
			Org:     "test-org",
			Project: "test-project",
			Domain:  "test-domain",
			Name:    "test-task",
			Version: "v1",
		},
	}

	resp, err := service.GetTaskDetails(ctx, connect.NewRequest(req))
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Msg.Details)
	assert.Equal(t, "test-org", resp.Msg.Details.TaskId.Org)
}

func TestListTasks(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewRepository(t)
	mockTaskRepo := mocks.NewTaskRepo(t)

	taskModels := []*models.Task{
		{
			TaskKey: models.TaskKey{
				Org:     "test-org",
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

	mockRepo.EXPECT().TaskRepo().Return(mockTaskRepo)
	mockTaskRepo.EXPECT().ListTasks(ctx, mock.AnythingOfType("interfaces.ListResourceInput")).Return(result, nil)

	service := NewTaskService(mockRepo)

	req := &task.ListTasksRequest{
		ScopeBy: &task.ListTasksRequest_Org{
			Org: "test-org",
		},
		Request: &common.ListRequest{
			Limit: 10,
		},
	}

	resp, err := service.ListTasks(ctx, connect.NewRequest(req))
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Msg.Tasks, 1)
	assert.NotNil(t, resp.Msg.Metadata)
	assert.Equal(t, uint32(1), resp.Msg.Metadata.Total)
}

func TestListVersions(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewRepository(t)
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

	mockRepo.EXPECT().TaskRepo().Return(mockTaskRepo)
	mockTaskRepo.EXPECT().ListVersions(ctx, mock.AnythingOfType("interfaces.ListResourceInput")).Return(versions, nil)

	service := NewTaskService(mockRepo)

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

	resp, err := service.ListVersions(ctx, connect.NewRequest(req))
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Msg.Versions, 2)
	assert.Equal(t, "v1", resp.Msg.Versions[0].Version)
	assert.Equal(t, "v2", resp.Msg.Versions[1].Version)
}