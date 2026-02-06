package service

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	interfaces "github.com/flyteorg/flyte/v2/runs/repository/mocks"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// =============================================================================
// RunService Tests
//
// Conventions:
// - Tests follow Arrange / Act / Assert.
// - Test names use the pattern: Test<Subject>_<Scenario>_<ExpectedResult>.
// - Comments describe intent (what/why) rather than restating code (how).
// =============================================================================

// TestNewRunService verifies the service can be constructed
// and returns a non-nil instance with minimal dependencies.
func TestNewRunService(t *testing.T) {
	service := NewRunService(nil, nil)
	assert.NotNil(t, service)
	t.Log("✅ RunService can be created")
}

// TestCreateRunRequest_WithProjectId verifies the request
// can be constructed using the ProjectId oneof branch and the fields are readable via getters.
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

	assert.NotNil(t, req)
	assert.NotNil(t, req.GetProjectId())
	assert.Equal(t, "test-org", req.GetProjectId().Organization)
	assert.Equal(t, "test-project", req.GetProjectId().Name)
	assert.Equal(t, "development", req.GetProjectId().Domain)

	t.Log("✅ CreateRunRequest with ProjectId works")
}

// TestCreateRunRequest_WithRunId verifies the request
// can be constructed using the RunId oneof branch and references a TaskId variant.
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

	assert.NotNil(t, req)
	assert.NotNil(t, req.GetRunId())
	assert.Equal(t, "my-run-123", req.GetRunId().Name)
	assert.NotNil(t, req.GetTaskId())
	assert.Equal(t, "my-task", req.GetTaskId().Name)

	t.Log("✅ CreateRunRequest with RunId works")
}

// TestRunService_CreateRun_WithProjectId verifies:
// - The service translates ProjectId + TaskSpec into a models.Run persisted via repo.
// - The returned API response includes an Action with identifiers, status, and metadata populated.
func TestRunService_CreateRun_WithProjectId(t *testing.T) {
	mockRepo := interfaces.NewRepository(t)
	mockActionRepo := interfaces.NewActionRepo(t)
	mockRepo.EXPECT().ActionRepo().Return(mockActionRepo)

	service := NewRunService(mockRepo, nil)

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

	req := connect.NewRequest(reqMsg)

	mockActionRepo.EXPECT().
		CreateRun(mock.Anything, mock.MatchedBy(func(run *models.Run) bool {
			return run.Org == "test-org" &&
				run.Project == "test-project" &&
				run.Domain == "development" &&
				run.Phase == "PHASE_QUEUED" &&
				run.ParentActionName == nil &&
				run.Name != ""
		})).
		Return(&models.Run{
			ID:        123,
			Org:       "test-org",
			Project:   "test-project",
			Domain:    "development",
			Name:      "generated-run-name",
			Phase:     "PHASE_QUEUED",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil)

	resp, err := service.CreateRun(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Msg)
	assert.NotNil(t, resp.Msg.Run)
	assert.NotNil(t, resp.Msg.Run.Action)

	// Validate identifier projection into the API response.
	assert.NotNil(t, resp.Msg.Run.Action.Id)
	assert.Equal(t, "test-org", resp.Msg.Run.Action.Id.Run.Org)
	assert.Equal(t, "test-project", resp.Msg.Run.Action.Id.Run.Project)
	assert.Equal(t, "development", resp.Msg.Run.Action.Id.Run.Domain)
	assert.Equal(t, "generated-run-name", resp.Msg.Run.Action.Id.Run.Name)

	// Validate default queued status semantics for newly-created runs.
	assert.NotNil(t, resp.Msg.Run.Action.Status)
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_QUEUED, resp.Msg.Run.Action.Status.Phase)
	assert.NotNil(t, resp.Msg.Run.Action.Status.StartTime)
	assert.Equal(t, uint32(0), resp.Msg.Run.Action.Status.Attempts)

	// Validate request-derived metadata is preserved.
	assert.NotNil(t, resp.Msg.Run.Action.Metadata)
	assert.Equal(t, workflow.RunSource_RUN_SOURCE_CLI, resp.Msg.Run.Action.Metadata.Source)

	t.Log("✅ CreateRun with ProjectId successfully called and returned complete response")
}

// TestRunService_CreateRun_WithRunId verifies:
// - The service honors caller-specified RunId and TaskId references.
// - The run is persisted in queued phase and the response reflects status + metadata.
func TestRunService_CreateRun_WithRunId(t *testing.T) {
	mockRepo := interfaces.NewRepository(t)
	mockActionRepo := interfaces.NewActionRepo(t)
	mockRepo.EXPECT().ActionRepo().Return(mockActionRepo)

	service := NewRunService(mockRepo, nil)

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

	req := connect.NewRequest(reqMsg)

	mockActionRepo.EXPECT().
		CreateRun(mock.Anything, mock.MatchedBy(func(run *models.Run) bool {
			return run.Org == "test-org" &&
				run.Project == "test-project" &&
				run.Domain == "development" &&
				run.Name == "my-custom-run-123" &&
				run.Phase == "PHASE_QUEUED" &&
				run.ParentActionName == nil
		})).
		Return(&models.Run{
			ID:        456,
			Org:       "test-org",
			Project:   "test-project",
			Domain:    "development",
			Name:      "my-custom-run-123",
			Phase:     "PHASE_QUEUED",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil)

	resp, err := service.CreateRun(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Msg.Run)
	assert.Equal(t, "my-custom-run-123", resp.Msg.Run.Action.Id.Run.Name)

	// Validate default queued status semantics for newly-created runs.
	assert.NotNil(t, resp.Msg.Run.Action.Status)
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_QUEUED, resp.Msg.Run.Action.Status.Phase)
	assert.NotNil(t, resp.Msg.Run.Action.Status.StartTime)
	if resp.Msg.Run.Action.Status != nil && resp.Msg.Run.Action.Status.StartTime != nil {
		startTime := resp.Msg.Run.Action.Status.StartTime.AsTime()
		t.Logf("📅 Created time: %v", startTime)
		t.Logf("📅 Created time (formatted): %s", startTime.Format(time.RFC3339))
	}

	// Validate request-derived metadata is preserved.
	assert.NotNil(t, resp.Msg.Run.Action.Metadata)
	assert.Equal(t, workflow.RunSource_RUN_SOURCE_WEB, resp.Msg.Run.Action.Metadata.Source)

	t.Log("✅ CreateRun with RunId successfully called and returned complete response")
}
