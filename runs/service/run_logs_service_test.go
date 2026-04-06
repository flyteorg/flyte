package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"fmt"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
	"database/sql"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/logs/dataplane"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	repoMocks "github.com/flyteorg/flyte/v2/runs/repository/mocks"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// mockLogStreamer is a test double for LogStreamer.
type mockLogStreamer struct {
	mock.Mock
}

func (m *mockLogStreamer) TailLogs(ctx context.Context, logContext *core.LogContext, stream *connect.ServerStream[workflow.TailLogsResponse]) error {
	args := m.Called(ctx, logContext, stream)
	return args.Error(0)
}

var tailLogsActionID = &common.ActionIdentifier{
	Run: &common.RunIdentifier{
		Org:     "test-org",
		Project: "test-project",
		Domain:  "test-domain",
		Name:    "rtest12345",
	},
	Name: "action-1",
}

func newTailLogsTestClient(t *testing.T, actionRepo *repoMocks.ActionRepo, streamer *mockLogStreamer) workflowconnect.RunLogsServiceClient {
	repo := &repoMocks.Repository{}
	repo.On("ActionRepo").Return(actionRepo)

	svc := NewRunLogsService(repo, streamer)
	path, handler := workflowconnect.NewRunLogsServiceHandler(svc)

	mux := http.NewServeMux()
	mux.Handle(path, handler)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := workflowconnect.NewRunLogsServiceClient(http.DefaultClient, server.URL)
	return client
}

func makeEventWithLogContext(actionID *common.ActionIdentifier, attempt uint32, phase common.ActionPhase, logCtx *core.LogContext) *models.ActionEvent {
	event := &workflow.ActionEvent{
		Id:          actionID,
		Attempt:     attempt,
		Phase:       phase,
		Version:     0,
		UpdatedTime: timestamppb.Now(),
		LogContext:  logCtx,
	}
	m, _ := models.NewActionEventModel(event)
	return m
}

func TestTailLogs_HappyPath(t *testing.T) {
	actionRepo := &repoMocks.ActionRepo{}
	streamer := &mockLogStreamer{}

	logCtx := &core.LogContext{
		PrimaryPodName: "my-pod",
		Pods: []*core.PodLogContext{
			{PodName: "my-pod", Namespace: "ns"},
		},
	}

	eventModel := makeEventWithLogContext(tailLogsActionID, 1, common.ActionPhase_ACTION_PHASE_RUNNING, logCtx)
	actionRepo.On("GetLatestEventByAttempt", mock.Anything, mock.Anything, uint32(1)).Return(eventModel, nil)

	// The streamer should be called with the logContext and send some response.
	streamer.On("TailLogs", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		stream := args.Get(2).(*connect.ServerStream[workflow.TailLogsResponse])
		_ = stream.Send(&workflow.TailLogsResponse{
			Logs: []*workflow.TailLogsResponse_Logs{
				{
					Lines: []*dataplane.LogLine{
						{Message: "hello world", Originator: dataplane.LogLineOriginator_USER},
					},
				},
			},
		})
	}).Return(nil)

	client := newTailLogsTestClient(t, actionRepo, streamer)

	stream, err := client.TailLogs(context.Background(), connect.NewRequest(&workflow.TailLogsRequest{
		ActionId: tailLogsActionID,
		Attempt:  1,
	}))
	assert.NoError(t, err)

	assert.True(t, stream.Receive())
	resp := stream.Msg()
	assert.Len(t, resp.Logs, 1)
	assert.Len(t, resp.Logs[0].Lines, 1)
	assert.Equal(t, "hello world", resp.Logs[0].Lines[0].Message)

	assert.False(t, stream.Receive())
	assert.NoError(t, stream.Err())

	actionRepo.AssertExpectations(t)
	streamer.AssertExpectations(t)
}

func TestTailLogs_NoLogContext(t *testing.T) {
	actionRepo := &repoMocks.ActionRepo{}
	streamer := &mockLogStreamer{}

	// Event without LogContext.
	eventModel := makeEventWithLogContext(tailLogsActionID, 1, common.ActionPhase_ACTION_PHASE_RUNNING, nil)
	actionRepo.On("GetLatestEventByAttempt", mock.Anything, mock.Anything, uint32(1)).Return(eventModel, nil)

	client := newTailLogsTestClient(t, actionRepo, streamer)

	stream, err := client.TailLogs(context.Background(), connect.NewRequest(&workflow.TailLogsRequest{
		ActionId: tailLogsActionID,
		Attempt:  1,
	}))
	assert.NoError(t, err)

	assert.False(t, stream.Receive())
	assert.Error(t, stream.Err())
	assert.Equal(t, connect.CodeNotFound, connect.CodeOf(stream.Err()))

	actionRepo.AssertExpectations(t)
}

func TestTailLogs_GetLatestEventError(t *testing.T) {
	actionRepo := &repoMocks.ActionRepo{}
	streamer := &mockLogStreamer{}

	actionRepo.On("GetLatestEventByAttempt", mock.Anything, mock.Anything, uint32(1)).Return(nil, assert.AnError)

	client := newTailLogsTestClient(t, actionRepo, streamer)

	stream, err := client.TailLogs(context.Background(), connect.NewRequest(&workflow.TailLogsRequest{
		ActionId: tailLogsActionID,
		Attempt:  1,
	}))
	assert.NoError(t, err)

	assert.False(t, stream.Receive())
	assert.Error(t, stream.Err())
	assert.Equal(t, connect.CodeInternal, connect.CodeOf(stream.Err()))

	actionRepo.AssertExpectations(t)
}

func TestTailLogs_StreamerError(t *testing.T) {
	actionRepo := &repoMocks.ActionRepo{}
	streamer := &mockLogStreamer{}

	logCtx := &core.LogContext{
		PrimaryPodName: "my-pod",
		Pods: []*core.PodLogContext{
			{PodName: "my-pod", Namespace: "ns"},
		},
	}

	eventModel := makeEventWithLogContext(tailLogsActionID, 1, common.ActionPhase_ACTION_PHASE_RUNNING, logCtx)
	actionRepo.On("GetLatestEventByAttempt", mock.Anything, mock.Anything, uint32(1)).Return(eventModel, nil)

	streamerErr := connect.NewError(connect.CodeInternal, assert.AnError)
	streamer.On("TailLogs", mock.Anything, mock.Anything, mock.Anything).Return(streamerErr)

	client := newTailLogsTestClient(t, actionRepo, streamer)

	stream, err := client.TailLogs(context.Background(), connect.NewRequest(&workflow.TailLogsRequest{
		ActionId: tailLogsActionID,
		Attempt:  1,
	}))
	assert.NoError(t, err)

	assert.False(t, stream.Receive())
	assert.Error(t, stream.Err())

	actionRepo.AssertExpectations(t)
	streamer.AssertExpectations(t)
}

func TestTailLogs_ConcurrencyLimit(t *testing.T) {
	actionRepo := &repoMocks.ActionRepo{}
	streamer := &mockLogStreamer{}

	logCtx := &core.LogContext{
		PrimaryPodName: "my-pod",
		Pods: []*core.PodLogContext{
			{PodName: "my-pod", Namespace: "ns"},
		},
	}

	eventModel := makeEventWithLogContext(tailLogsActionID, 1, common.ActionPhase_ACTION_PHASE_RUNNING, logCtx)
	actionRepo.On("GetLatestEventByAttempt", mock.Anything, mock.Anything, mock.Anything).Return(eventModel, nil)

	// Block in streamer to hold semaphore slots.
	blocker := make(chan struct{})
	streamer.On("TailLogs", mock.Anything, mock.Anything, mock.Anything).Run(func(_ mock.Arguments) {
		<-blocker
	}).Return(nil)

	repo := &repoMocks.Repository{}
	repo.On("ActionRepo").Return(actionRepo)

	svc := NewRunLogsService(repo, streamer)
	// Exhaust the semaphore by acquiring all slots.
	svc.sem.Acquire(context.Background(), defaultMaxConcurrentStreams)

	// The next TailLogs should be rejected.
	path, handler := workflowconnect.NewRunLogsServiceHandler(svc)
	mux := http.NewServeMux()
	mux.Handle(path, handler)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := workflowconnect.NewRunLogsServiceClient(http.DefaultClient, server.URL)
	stream, err := client.TailLogs(context.Background(), connect.NewRequest(&workflow.TailLogsRequest{
		ActionId: tailLogsActionID,
		Attempt:  1,
	}))
	assert.NoError(t, err)

	assert.False(t, stream.Receive())
	assert.Error(t, stream.Err())
	assert.Equal(t, connect.CodeResourceExhausted, connect.CodeOf(stream.Err()))

	// Release so cleanup doesn't hang.
	svc.sem.Release(defaultMaxConcurrentStreams)
	close(blocker)
}

func TestTailLogs_AttemptZero(t *testing.T) {
	actionRepo := &repoMocks.ActionRepo{}
	streamer := &mockLogStreamer{}

	logCtx := &core.LogContext{
		PrimaryPodName: "my-pod",
		Pods: []*core.PodLogContext{
			{PodName: "my-pod", Namespace: "ns"},
		},
	}

	eventModel := makeEventWithLogContext(tailLogsActionID, 0, common.ActionPhase_ACTION_PHASE_RUNNING, logCtx)
	actionRepo.On("GetLatestEventByAttempt", mock.Anything, mock.Anything, uint32(0)).Return(eventModel, nil)

	streamer.On("TailLogs", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		stream := args.Get(2).(*connect.ServerStream[workflow.TailLogsResponse])
		_ = stream.Send(&workflow.TailLogsResponse{
			Logs: []*workflow.TailLogsResponse_Logs{
				{
					Lines: []*dataplane.LogLine{
						{Message: "attempt zero log", Originator: dataplane.LogLineOriginator_USER},
					},
				},
			},
		})
	}).Return(nil)

	client := newTailLogsTestClient(t, actionRepo, streamer)

	stream, err := client.TailLogs(context.Background(), connect.NewRequest(&workflow.TailLogsRequest{
		ActionId: tailLogsActionID,
		Attempt:  0,
	}))
	assert.NoError(t, err)

	assert.True(t, stream.Receive())
	resp := stream.Msg()
	assert.Len(t, resp.Logs, 1)
	assert.Equal(t, "attempt zero log", resp.Logs[0].Lines[0].Message)

	assert.False(t, stream.Receive())
	assert.NoError(t, stream.Err())

	actionRepo.AssertExpectations(t)
	streamer.AssertExpectations(t)
}

func TestTailLogs_EventNotFound(t *testing.T) {
	actionRepo := &repoMocks.ActionRepo{}
	streamer := &mockLogStreamer{}

	actionRepo.On("GetLatestEventByAttempt", mock.Anything, mock.Anything, uint32(5)).
		Return(nil, fmt.Errorf("event not found for attempt 5: %w", sql.ErrNoRows))

	client := newTailLogsTestClient(t, actionRepo, streamer)

	stream, err := client.TailLogs(context.Background(), connect.NewRequest(&workflow.TailLogsRequest{
		ActionId: tailLogsActionID,
		Attempt:  5,
	}))
	assert.NoError(t, err)

	assert.False(t, stream.Receive())
	assert.Error(t, stream.Err())
	assert.Equal(t, connect.CodeNotFound, connect.CodeOf(stream.Err()))

	actionRepo.AssertExpectations(t)
}
