// runs/service/run_service_test.go

package service

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	interfaces "github.com/flyteorg/flyte/v2/runs/repository/mocks"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test 1: Verify RunService structure can be created
func TestNewRunService(t *testing.T) {
	service := NewRunService(nil, nil)

	assert.NotNil(t, service)
	t.Log("✅ RunService can be created")
}

// Test 2: Test CreateRunRequest structure (with ProjectId)
func TestCreateRunRequest_WithProjectId(t *testing.T) {
	projectId := &common.ProjectIdentifier{
		Organization: "test-org",
		Name:         "test-project",
		Domain:       "development",
	}

	taskSpec := &task.TaskSpec{}

	req := &workflow.CreateRunRequest{
		Id: &workflow.CreateRunRequest_ProjectId{
			ProjectId: projectId,
		},
		Task: &workflow.CreateRunRequest_TaskSpec{
			TaskSpec: taskSpec,
		},
		Source: workflow.RunSource_RUN_SOURCE_CLI,
	}

	// Verify we can read the data
	assert.NotNil(t, req)
	assert.NotNil(t, req.GetProjectId())
	assert.Equal(t, "test-org", req.GetProjectId().Organization)
	assert.Equal(t, "test-project", req.GetProjectId().Name)
	assert.Equal(t, "development", req.GetProjectId().Domain)

	t.Log("✅ CreateRunRequest with ProjectId works")
}

// Test 3: Test CreateRunRequest structure (with RunId)
func TestCreateRunRequest_WithRunId(t *testing.T) {
	runId := &common.RunIdentifier{
		Org:     "test-org",
		Project: "test-project",
		Domain:  "development",
		Name:    "my-run-123",
	}

	taskId := &task.TaskIdentifier{
		Org:     "test-org",
		Project: "test-project",
		Domain:  "development",
		Name:    "my-task",
		Version: "v1.0",
	}

	req := &workflow.CreateRunRequest{
		Id: &workflow.CreateRunRequest_RunId{
			RunId: runId,
		},
		Task: &workflow.CreateRunRequest_TaskId{
			TaskId: taskId,
		},
		Source: workflow.RunSource_RUN_SOURCE_WEB,
	}

	// Verify
	assert.NotNil(t, req)
	assert.NotNil(t, req.GetRunId())
	assert.Equal(t, "my-run-123", req.GetRunId().Name)
	assert.NotNil(t, req.GetTaskId())
	assert.Equal(t, "my-task", req.GetTaskId().Name)

	t.Log("✅ CreateRunRequest with RunId works")
}

// ⭐⭐⭐ NEW TEST - Actually calls service.CreateRun() ⭐⭐⭐
// Test 4: Test actual CreateRun function with ProjectId
func TestRunService_CreateRun_WithProjectId(t *testing.T) {
	// Setup mocks
	mockRepo := interfaces.NewRepository(t)
	mockActionRepo := interfaces.NewActionRepo(t)

	// Configure mock repository
	mockRepo.EXPECT().ActionRepo().Return(mockActionRepo)

	// Create service with mocks (no queue client for now)
	service := NewRunService(mockRepo, nil)

	// Create the protobuf request
	reqMsg := &workflow.CreateRunRequest{
		Id: &workflow.CreateRunRequest_ProjectId{
			ProjectId: &common.ProjectIdentifier{
				Organization: "test-org",
				Name:         "test-project",
				Domain:       "development",
			},
		},
		Task: &workflow.CreateRunRequest_TaskSpec{
			TaskSpec: &task.TaskSpec{
				ShortName: "my-task",
			},
		},
		Source: workflow.RunSource_RUN_SOURCE_CLI,
	}

	// ⭐ Wrap in connect.Request
	req := connect.NewRequest(reqMsg)

	// Setup mock expectations
	mockActionRepo.EXPECT().
		CreateRun(mock.Anything, reqMsg). // Expects the unwrapped protobuf
		Return(&models.Run{
			ID:      123,
			Org:     "test-org",
			Project: "test-project",
			Domain:  "development",
			Name:    "generated-run-name",
		}, nil)

	// ⭐ Actually call CreateRun with connect.Request!
	resp, err := service.CreateRun(context.Background(), req)

	// Verify the response
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Msg)
	assert.NotNil(t, resp.Msg.Run)
	assert.NotNil(t, resp.Msg.Run.Action)
	assert.NotNil(t, resp.Msg.Run.Action.Id)
	assert.Equal(t, "test-org", resp.Msg.Run.Action.Id.Run.Org)
	assert.Equal(t, "test-project", resp.Msg.Run.Action.Id.Run.Project)
	assert.Equal(t, "development", resp.Msg.Run.Action.Id.Run.Domain)
	assert.Equal(t, "generated-run-name", resp.Msg.Run.Action.Id.Run.Name)

	t.Log("✅ CreateRun with ProjectId successfully called and returned response")
}

// ⭐⭐⭐ NEW TEST - Test with RunId ⭐⭐⭐
// Test 5: Test actual CreateRun function with RunId
func TestRunService_CreateRun_WithRunId(t *testing.T) {
	// Setup mocks
	mockRepo := interfaces.NewRepository(t)
	mockActionRepo := interfaces.NewActionRepo(t)

	mockRepo.EXPECT().ActionRepo().Return(mockActionRepo)

	service := NewRunService(mockRepo, nil)

	// Create the protobuf request
	reqMsg := &workflow.CreateRunRequest{
		Id: &workflow.CreateRunRequest_RunId{
			RunId: &common.RunIdentifier{
				Org:     "test-org",
				Project: "test-project",
				Domain:  "development",
				Name:    "my-custom-run-123",
			},
		},
		Task: &workflow.CreateRunRequest_TaskId{
			TaskId: &task.TaskIdentifier{
				Org:     "test-org",
				Project: "test-project",
				Domain:  "development",
				Name:    "my-task",
				Version: "v1.0",
			},
		},
		Source: workflow.RunSource_RUN_SOURCE_WEB,
	}

	// ⭐ Wrap in connect.Request
	req := connect.NewRequest(reqMsg)

	// Setup mock expectations
	mockActionRepo.EXPECT().
		CreateRun(mock.Anything, reqMsg).
		Return(&models.Run{
			ID:      456,
			Org:     "test-org",
			Project: "test-project",
			Domain:  "development",
			Name:    "my-custom-run-123",
		}, nil)

	// ⭐ Actually call CreateRun!
	resp, err := service.CreateRun(context.Background(), req)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Msg.Run)
	assert.Equal(t, "my-custom-run-123", resp.Msg.Run.Action.Id.Run.Name)

	t.Log("✅ CreateRun with RunId successfully called and returned response")
}
