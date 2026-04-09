package api

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
)

// createTestAction sets up a run and action in the DB via the InternalRunService API.
func createTestAction(
	t *testing.T,
	ctx context.Context,
	internalClient workflowconnect.InternalRunServiceClient,
	runID *common.RunIdentifier,
	actionName string,
	taskSpec *task.TaskSpec,
) {
	t.Helper()
	actionID := &common.ActionIdentifier{Run: runID, Name: actionName}

	_, err := internalClient.RecordAction(ctx, connect.NewRequest(&workflow.RecordActionRequest{
		ActionId: actionID,
		Spec: &workflow.RecordActionRequest_Task{
			Task: &workflow.TaskAction{
				Spec: taskSpec,
			},
		},
	}))
	require.NoError(t, err)
}

// recordActionEvent records a single action event via the InternalRunService API.
func recordActionEvent(
	t *testing.T,
	ctx context.Context,
	internalClient workflowconnect.InternalRunServiceClient,
	event *workflow.ActionEvent,
) {
	t.Helper()
	_, err := internalClient.RecordActionEvents(ctx, connect.NewRequest(&workflow.RecordActionEventsRequest{
		Events: []*workflow.ActionEvent{event},
	}))
	require.NoError(t, err)
}

func TestWatchActionDetails_StreamedUpdateIncludesFullDetails(t *testing.T) {
	t.Cleanup(func() { cleanupTestDB(t) })

	ctx := context.Background()
	httpClient := newClient()
	runClient := workflowconnect.NewRunServiceClient(httpClient, endpoint)
	internalClient := workflowconnect.NewInternalRunServiceClient(httpClient, endpoint)

	runID := &common.RunIdentifier{
		Org:     testOrg,
		Project: testProject,
		Domain:  testDomain,
		Name:    "r" + uniqueString(),
	}
	actionName := "action-1"
	actionID := &common.ActionIdentifier{Run: runID, Name: actionName}

	taskSpecProto := &task.TaskSpec{
		TaskTemplate: &core.TaskTemplate{
			Type: "python",
		},
	}
	createTestAction(t, ctx, internalClient, runID, actionName, taskSpecProto)

	// Record a RUNNING event
	now := time.Now()
	recordActionEvent(t, ctx, internalClient, &workflow.ActionEvent{
		Id:          actionID,
		Attempt:     0,
		Phase:       common.ActionPhase_ACTION_PHASE_RUNNING,
		Version:     0,
		UpdatedTime: timestamppb.New(now),
		LogInfo: []*core.TaskLog{
			{Name: "log-1", Uri: "s3://logs/1"},
		},
	})

	// Start watching
	watchCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	stream, err := runClient.WatchActionDetails(watchCtx, connect.NewRequest(&workflow.WatchActionDetailsRequest{
		ActionId: actionID,
	}))
	require.NoError(t, err)
	defer stream.Close()

	// First message: initial state
	require.True(t, stream.Receive())
	initial := stream.Msg().Details
	require.NotNil(t, initial)

	// Verify initial includes task spec
	assert.NotNil(t, initial.GetTask(), "initial should include task spec")
	assert.Equal(t, "python", initial.GetTask().GetTaskTemplate().GetType())

	// Verify initial includes attempts
	assert.NotEmpty(t, initial.Attempts, "initial should include attempts")
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_RUNNING, initial.Attempts[0].Phase)
}

func TestWatchActionDetails_StreamedUpdateIncludesErrorInfo(t *testing.T) {
	t.Cleanup(func() { cleanupTestDB(t) })

	ctx := context.Background()
	httpClient := newClient()
	runClient := workflowconnect.NewRunServiceClient(httpClient, endpoint)
	internalClient := workflowconnect.NewInternalRunServiceClient(httpClient, endpoint)

	runID := &common.RunIdentifier{
		Org:     testOrg,
		Project: testProject,
		Domain:  testDomain,
		Name:    "r" + uniqueString(),
	}
	actionName := "action-1"
	actionID := &common.ActionIdentifier{Run: runID, Name: actionName}

	createTestAction(t, ctx, internalClient, runID, actionName, nil)

	// Update action status to FAILED
	_, err := internalClient.UpdateActionStatus(ctx, connect.NewRequest(&workflow.UpdateActionStatusRequest{
		ActionId: actionID,
		Status: &workflow.ActionStatus{
			Phase:    common.ActionPhase_ACTION_PHASE_FAILED,
			Attempts: 1,
		},
	}))
	require.NoError(t, err)

	// Record a FAILED event with error info (attempt is 1-indexed to match status.Attempts)
	now := time.Now()
	recordActionEvent(t, ctx, internalClient, &workflow.ActionEvent{
		Id:      actionID,
		Attempt: 1,
		Phase:   common.ActionPhase_ACTION_PHASE_FAILED,
		Version: 0,
		ErrorInfo: &workflow.ErrorInfo{
			Message: "task failed with OOM",
			Kind:    workflow.ErrorInfo_KIND_SYSTEM,
		},
		UpdatedTime: timestamppb.New(now),
	})

	// Start watching
	watchCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	stream, err := runClient.WatchActionDetails(watchCtx, connect.NewRequest(&workflow.WatchActionDetailsRequest{
		ActionId: actionID,
	}))
	require.NoError(t, err)
	defer stream.Close()

	// First message: initial state should include error info
	require.True(t, stream.Receive())
	initial := stream.Msg().Details
	require.NotNil(t, initial)
	assert.NotNil(t, initial.GetErrorInfo())
	assert.Equal(t, "task failed with OOM", initial.GetErrorInfo().GetMessage())
}

func TestWatchActionDetails_GetActionNotFound(t *testing.T) {
	t.Cleanup(func() { cleanupTestDB(t) })

	ctx := context.Background()
	httpClient := newClient()
	runClient := workflowconnect.NewRunServiceClient(httpClient, endpoint)

	// Request details for an action that doesn't exist
	watchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	stream, err := runClient.WatchActionDetails(watchCtx, connect.NewRequest(&workflow.WatchActionDetailsRequest{
		ActionId: &common.ActionIdentifier{
			Run: &common.RunIdentifier{
				Org:     testOrg,
				Project: testProject,
				Domain:  testDomain,
				Name:    "nonexistent-run",
			},
			Name: "nonexistent-action",
		},
	}))
	require.NoError(t, err)
	defer stream.Close()

	// Should fail on first receive since action doesn't exist
	assert.False(t, stream.Receive())
	assert.Error(t, stream.Err())
}
