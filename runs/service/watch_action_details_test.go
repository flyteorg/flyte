package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// newWatchActionDetailsTestServer creates an httptest server and connect client for testing WatchActionDetails.
func newWatchActionDetailsTestServer(t *testing.T, svc *RunService) (workflowconnect.RunServiceClient, func()) {
	t.Helper()
	mux := http.NewServeMux()
	path, handler := workflowconnect.NewRunServiceHandler(svc)
	mux.Handle(path, handler)
	server := httptest.NewUnstartedServer(mux)
	server.EnableHTTP2 = true
	server.StartTLS()
	client := workflowconnect.NewRunServiceClient(server.Client(), server.URL)
	return client, server.Close
}

func TestWatchActionDetails_StreamedUpdateIncludesFullDetails(t *testing.T) {
	actionRepo, taskRepo, svc := newTestServiceWithTaskRepo(t)

	// Set up task spec for the action
	runInfo := &workflow.RunInfo{
		TaskSpecDigest: "digest-abc",
		InputsUri:      "s3://bucket/inputs.pb",
	}
	runInfoBytes, err := proto.Marshal(runInfo)
	require.NoError(t, err)

	actionModel := &models.Action{
		Org:          "test-org",
		Project:      "test-project",
		Domain:       "test-domain",
		RunName:      "rtest12345",
		Name:         "action-1",
		Phase:        int32(common.ActionPhase_ACTION_PHASE_RUNNING),
		ActionType:   int32(workflow.ActionType_ACTION_TYPE_TASK),
		DetailedInfo: runInfoBytes,
	}

	taskSpecProto := &task.TaskSpec{
		TaskTemplate: &core.TaskTemplate{
			Type: "python",
		},
	}
	taskSpecBytes, err := proto.Marshal(taskSpecProto)
	require.NoError(t, err)

	now := time.Now()
	runningEvent := &workflow.ActionEvent{
		Id:          testActionID,
		Attempt:     0,
		Phase:       common.ActionPhase_ACTION_PHASE_RUNNING,
		Version:     0,
		UpdatedTime: timestamppb.New(now),
		LogInfo: []*core.TaskLog{
			{Name: "log-1", Uri: "s3://logs/1"},
		},
	}
	runningEventModel, err := models.NewActionEventModel(runningEvent)
	require.NoError(t, err)

	// GetAction is called only once for the initial state; the streamed update reuses the model from WatchActionUpdates.
	// Use mock.Anything for proto messages since they get re-serialized over HTTP.
	actionRepo.On("GetAction", mock.Anything, mock.Anything).Return(actionModel, nil).Once()
	taskRepo.On("GetTaskSpec", mock.Anything, "digest-abc").Return(&models.TaskSpec{
		Digest: "digest-abc",
		Spec:   taskSpecBytes,
	}, nil)
	actionRepo.On("ListEvents", mock.Anything, mock.Anything, 500).Return([]*models.ActionEvent{runningEventModel}, nil)

	// WatchActionUpdates sends one update for the matching action, then closes the channel
	actionRepo.On("WatchActionUpdates", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			updates := args.Get(2).(chan<- *models.Action)
			// Send an update for our action
			updates <- actionModel
			// Close channel to end the stream
			close(updates)
		}).Return()

	client, cleanup := newWatchActionDetailsTestServer(t, svc)
	defer cleanup()

	stream, err := client.WatchActionDetails(context.Background(), connect.NewRequest(&workflow.WatchActionDetailsRequest{
		ActionId: testActionID,
	}))
	require.NoError(t, err)
	defer stream.Close()

	// First message: initial state
	require.True(t, stream.Receive())
	initial := stream.Msg().Details
	require.NotNil(t, initial)

	// Second message: streamed update (should have full details like initial)
	require.True(t, stream.Receive())
	streamed := stream.Msg().Details
	require.NotNil(t, streamed)

	// Verify the streamed update includes task spec
	assert.NotNil(t, streamed.GetTask(), "streamed update should include task spec")
	assert.Equal(t, "python", streamed.GetTask().GetTaskTemplate().GetType())

	// Verify the streamed update includes attempts from events
	assert.NotEmpty(t, streamed.Attempts, "streamed update should include attempts")
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_RUNNING, streamed.Attempts[0].Phase)

	// Verify streamed response matches initial response
	assert.Equal(t, initial.GetTask().GetTaskTemplate().GetType(), streamed.GetTask().GetTaskTemplate().GetType())
	assert.Equal(t, len(initial.Attempts), len(streamed.Attempts))

	// No more messages
	assert.False(t, stream.Receive())
	assert.NoError(t, stream.Err())
}

func TestWatchActionDetails_StreamedUpdateIncludesErrorInfo(t *testing.T) {
	actionRepo, _, svc := newTestServiceWithTaskRepo(t)

	actionModel := &models.Action{
		Org:     "test-org",
		Project: "test-project",
		Domain:  "test-domain",
		RunName: "rtest12345",
		Name:    "action-1",
		Phase:   int32(common.ActionPhase_ACTION_PHASE_FAILED),
	}

	now := time.Now()
	failedEvent := &workflow.ActionEvent{
		Id:      testActionID,
		Attempt: 1,
		Phase:   common.ActionPhase_ACTION_PHASE_FAILED,
		Version: 1,
		ErrorInfo: &workflow.ErrorInfo{
			Message: "task failed with OOM",
			Kind:    workflow.ErrorInfo_KIND_SYSTEM,
		},
		UpdatedTime: timestamppb.New(now),
	}
	failedEventModel, err := models.NewActionEventModel(failedEvent)
	require.NoError(t, err)

	// GetAction is called only once for initial state; streamed update reuses the model from WatchActionUpdates.
	actionRepo.On("GetAction", mock.Anything, mock.Anything).Return(actionModel, nil).Once()
	actionRepo.On("ListEvents", mock.Anything, mock.Anything, 500).Return([]*models.ActionEvent{failedEventModel}, nil)

	// WatchActionUpdates sends one update then closes
	actionRepo.On("WatchActionUpdates", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			updates := args.Get(2).(chan<- *models.Action)
			updates <- actionModel
			close(updates)
		}).Return()

	client, cleanup := newWatchActionDetailsTestServer(t, svc)
	defer cleanup()

	stream, err := client.WatchActionDetails(context.Background(), connect.NewRequest(&workflow.WatchActionDetailsRequest{
		ActionId: testActionID,
	}))
	require.NoError(t, err)
	defer stream.Close()

	// First message: initial state
	require.True(t, stream.Receive())
	initial := stream.Msg().Details

	// Second message: streamed update
	require.True(t, stream.Receive())
	streamed := stream.Msg().Details

	// Verify both include error info
	assert.NotNil(t, initial.GetErrorInfo())
	assert.Equal(t, "task failed with OOM", initial.GetErrorInfo().GetMessage())
	assert.NotNil(t, streamed.GetErrorInfo(), "streamed update should include error info")
	assert.Equal(t, "task failed with OOM", streamed.GetErrorInfo().GetMessage())
}

func TestWatchActionDetails_FiltersOtherActions(t *testing.T) {
	actionRepo, _, svc := newTestServiceWithTaskRepo(t)

	actionModel := &models.Action{
		Org:     "test-org",
		Project: "test-project",
		Domain:  "test-domain",
		RunName: "rtest12345",
		Name:    "action-1",
		Phase:   int32(common.ActionPhase_ACTION_PHASE_RUNNING),
	}

	otherActionModel := &models.Action{
		Org:     "test-org",
		Project: "test-project",
		Domain:  "test-domain",
		RunName: "rtest12345",
		Name:    "action-other",
		Phase:   int32(common.ActionPhase_ACTION_PHASE_RUNNING),
	}

	// GetAction is called only once for initial state; the other-action update is filtered out.
	actionRepo.On("GetAction", mock.Anything, mock.Anything).Return(actionModel, nil).Once()
	actionRepo.On("ListEvents", mock.Anything, mock.Anything, 500).Return([]*models.ActionEvent{}, nil).Once()

	// Send an update for a different action, then close
	actionRepo.On("WatchActionUpdates", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			updates := args.Get(2).(chan<- *models.Action)
			// Send update for a different action - should be filtered out
			updates <- otherActionModel
			close(updates)
		}).Return()

	client, cleanup := newWatchActionDetailsTestServer(t, svc)
	defer cleanup()

	stream, err := client.WatchActionDetails(context.Background(), connect.NewRequest(&workflow.WatchActionDetailsRequest{
		ActionId: testActionID,
	}))
	require.NoError(t, err)
	defer stream.Close()

	// First message: initial state
	require.True(t, stream.Receive())
	assert.NotNil(t, stream.Msg().Details)

	// No more messages - the other action update should have been filtered
	assert.False(t, stream.Receive())
	assert.NoError(t, stream.Err())
}

func TestWatchActionDetails_GetActionDetailsErrorReturnsInternalError(t *testing.T) {
	actionRepo, _, svc := newTestServiceWithTaskRepo(t)

	// Initial call succeeds
	actionModel := &models.Action{
		Org:     "test-org",
		Project: "test-project",
		Domain:  "test-domain",
		RunName: "rtest12345",
		Name:    "action-1",
		Phase:   int32(common.ActionPhase_ACTION_PHASE_RUNNING),
	}

	// GetAction is called only once for the initial state; the streamed update reuses the model from WatchActionUpdates.
	actionRepo.On("GetAction", mock.Anything, mock.Anything).Return(actionModel, nil).Once()
	actionRepo.On("ListEvents", mock.Anything, mock.Anything, 500).Return([]*models.ActionEvent{}, nil).Once()

	// Second call to ListEvents fails (during streamed update via buildActionDetails)
	actionRepo.On("ListEvents", mock.Anything, mock.Anything, 500).Return(nil, assert.AnError).Once()

	actionRepo.On("WatchActionUpdates", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			updates := args.Get(2).(chan<- *models.Action)
			updates <- actionModel
			close(updates)
		}).Return()

	client, cleanup := newWatchActionDetailsTestServer(t, svc)
	defer cleanup()

	stream, err := client.WatchActionDetails(context.Background(), connect.NewRequest(&workflow.WatchActionDetailsRequest{
		ActionId: testActionID,
	}))
	require.NoError(t, err)
	defer stream.Close()

	// First message: initial state succeeds
	require.True(t, stream.Receive())

	// Second message: should fail due to ListEvents error
	assert.False(t, stream.Receive())
	assert.Error(t, stream.Err())
}
