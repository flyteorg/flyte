package rpc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	repoMocks "github.com/flyteorg/flyte/v2/runs/repository/mocks"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
	"github.com/flyteorg/flyte/v2/runs/service"
)

func newTestRun(t *testing.T) (*repoMocks.ActionRepo, *service.RunService) {
	t.Helper()
	actionRepo := repoMocks.NewActionRepo(t)
	repo := repoMocks.NewRepository(t)
	repo.On("ActionRepo").Return(actionRepo)
	return actionRepo, service.NewRun(repo, nil, nil, nil, "", nil, nil)
}

func matchActionID(expected *common.ActionIdentifier) interface{} {
	return mock.MatchedBy(func(actual *common.ActionIdentifier) bool {
		return proto.Equal(actual, expected)
	})
}

func matchRunID(expected *common.RunIdentifier) interface{} {
	return mock.MatchedBy(func(actual *common.RunIdentifier) bool {
		return proto.Equal(actual, expected)
	})
}

func newRunServiceTestClient(t *testing.T, svc *service.RunService) workflowconnect.RunServiceClient {
	path, handler := workflowconnect.NewRunServiceHandler(&runHandler{run: svc})

	mux := http.NewServeMux()
	mux.Handle(path, handler)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return workflowconnect.NewRunServiceClient(http.DefaultClient, server.URL)
}

func TestWatchClusterEvents_UsesPersistedClusterEvents(t *testing.T) {
	actionRepo, svc := newTestRun(t)
	client := newRunServiceTestClient(t, svc)

	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     "test-org",
			Project: "test-project",
			Domain:  "test-domain",
			Name:    "rtest12345",
		},
		Name: "action-1",
	}

	actionModel := &models.Action{
		Project: actionID.Run.Project,
		Domain:  actionID.Run.Domain,
		RunName: actionID.Run.Name,
		Name:    actionID.Name,
		Phase:   int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED),
	}

	clusterEvent := &workflow.ClusterEvent{
		OccurredAt: timestamppb.New(time.Unix(100, 0)),
		Message:    "Pod scheduled on node-a",
	}
	eventModel, err := models.NewActionEventModel(&workflow.ActionEvent{
		Id:            actionID,
		Attempt:       0,
		Phase:         common.ActionPhase_ACTION_PHASE_RUNNING,
		Version:       1,
		UpdatedTime:   timestamppb.New(time.Unix(101, 0)),
		ClusterEvents: []*workflow.ClusterEvent{clusterEvent},
	})
	require.NoError(t, err)
	eventModel.UpdatedAt = time.Unix(101, 0)

	actionRepo.On("GetAction", mock.Anything, matchActionID(actionID)).Return(actionModel, nil).Once()
	actionRepo.On("WatchActionUpdates", mock.Anything, matchActionID(actionID), mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			updates := args.Get(2).(chan<- *models.Action)
			errs := args.Get(3).(chan<- error)
			close(updates)
			close(errs)
		}).Once()
	actionRepo.On("ListEventsSince", mock.Anything, matchActionID(actionID), uint32(0), time.Time{}, 0, 500).
		Return([]*models.ActionEvent{eventModel}, nil).Once()

	stream, err := client.WatchClusterEvents(context.Background(), connect.NewRequest(&workflow.WatchClusterEventsRequest{
		Id:      actionID,
		Attempt: 0,
	}))
	require.NoError(t, err)
	require.True(t, stream.Receive())

	resp := stream.Msg()
	require.Len(t, resp.GetClusterEvents(), 1)
	assert.Equal(t, "Pod scheduled on node-a", resp.GetClusterEvents()[0].GetMessage())

	assert.False(t, stream.Receive())
	assert.NoError(t, stream.Err())
}

func TestWatchClusterEvents_StreamsNewPersistedClusterEventsWithoutReplay(t *testing.T) {
	actionRepo, svc := newTestRun(t)
	client := newRunServiceTestClient(t, svc)

	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     "test-org",
			Project: "test-project",
			Domain:  "test-domain",
			Name:    "rtest12345",
		},
		Name: "action-1",
	}

	runningAction := &models.Action{
		Project: actionID.Run.Project,
		Domain:  actionID.Run.Domain,
		RunName: actionID.Run.Name,
		Name:    actionID.Name,
		Phase:   int32(common.ActionPhase_ACTION_PHASE_RUNNING),
	}
	succeededAction := &models.Action{
		Project: actionID.Run.Project,
		Domain:  actionID.Run.Domain,
		RunName: actionID.Run.Name,
		Name:    actionID.Name,
		Phase:   int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED),
	}

	event1, err := models.NewActionEventModel(&workflow.ActionEvent{
		Id:          actionID,
		Attempt:     0,
		Phase:       common.ActionPhase_ACTION_PHASE_RUNNING,
		Version:     1,
		UpdatedTime: timestamppb.New(time.Unix(101, 0)),
		ClusterEvents: []*workflow.ClusterEvent{{
			OccurredAt: timestamppb.New(time.Unix(100, 0)),
			Message:    "Pod created",
		}},
	})
	require.NoError(t, err)
	event1.UpdatedAt = time.Unix(101, 0)

	event2, err := models.NewActionEventModel(&workflow.ActionEvent{
		Id:          actionID,
		Attempt:     0,
		Phase:       common.ActionPhase_ACTION_PHASE_SUCCEEDED,
		Version:     2,
		UpdatedTime: timestamppb.New(time.Unix(103, 0)),
		ClusterEvents: []*workflow.ClusterEvent{{
			OccurredAt: timestamppb.New(time.Unix(102, 0)),
			Message:    "Pulled container image",
		}},
	})
	require.NoError(t, err)
	event2.UpdatedAt = time.Unix(103, 0)

	actionRepo.On("GetAction", mock.Anything, matchActionID(actionID)).Return(runningAction, nil).Once()
	actionRepo.On("ListEventsSince", mock.Anything, matchActionID(actionID), uint32(0), time.Time{}, 0, 500).
		Return([]*models.ActionEvent{event1}, nil).Once()
	actionRepo.On("WatchActionUpdates", mock.Anything, matchActionID(actionID), mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			updates := args.Get(2).(chan<- *models.Action)
			updates <- succeededAction
			close(updates)
		}).Once()
	actionRepo.On("ListEventsSince", mock.Anything, matchActionID(actionID), uint32(0), event1.UpdatedAt, 0, 500).
		Return([]*models.ActionEvent{event2}, nil).Once()

	stream, err := client.WatchClusterEvents(context.Background(), connect.NewRequest(&workflow.WatchClusterEventsRequest{
		Id:      actionID,
		Attempt: 0,
	}))
	require.NoError(t, err)

	require.True(t, stream.Receive())
	first := stream.Msg()
	require.Len(t, first.GetClusterEvents(), 1)
	assert.Equal(t, "Pod created", first.GetClusterEvents()[0].GetMessage())

	require.True(t, stream.Receive())
	second := stream.Msg()
	require.Len(t, second.GetClusterEvents(), 1)
	assert.Equal(t, "Pulled container image", second.GetClusterEvents()[0].GetMessage())

	assert.False(t, stream.Receive())
	assert.NoError(t, stream.Err())
}

func TestWatchRunDetails(t *testing.T) {
	runID := &common.RunIdentifier{
		Org:     "test-org",
		Project: "test-project",
		Domain:  "test-domain",
		Name:    "rtest-watch-1",
	}
	rootActionID := &common.ActionIdentifier{Run: runID, Name: runID.Name}

	actionSpecBytes, err := proto.Marshal(&workflow.ActionSpec{
		RunSpec: &task.RunSpec{Labels: &task.Labels{Values: map[string]string{"env": "prod"}}},
	})
	require.NoError(t, err)

	runModel := &models.Run{
		Project:    runID.Project,
		Domain:     runID.Domain,
		Name:       runID.Name,
		Phase:      int32(common.ActionPhase_ACTION_PHASE_RUNNING),
		Attempts:   1,
		ActionSpec: actionSpecBytes,
	}

	// The terminal update: a FAILED run whose last attempt carries error info.
	failedModel := &models.Run{
		Project:    runID.Project,
		Domain:     runID.Domain,
		Name:       runID.Name,
		Phase:      int32(common.ActionPhase_ACTION_PHASE_FAILED),
		Attempts:   1,
		ActionSpec: actionSpecBytes,
	}
	failedEventBytes, err := proto.Marshal(&workflow.ActionEvent{
		Id:      rootActionID,
		Attempt: 1,
		Phase:   common.ActionPhase_ACTION_PHASE_FAILED,
		ErrorInfo: &workflow.ErrorInfo{
			Kind:    workflow.ErrorInfo_KIND_USER,
			Message: "boom",
		},
	})
	require.NoError(t, err)
	failedEvents := []*models.ActionEvent{{Attempt: 1, Info: failedEventBytes}}

	tests := []struct {
		name       string
		setupMocks func(actionRepo *repoMocks.ActionRepo)
		test       func(t *testing.T, client workflowconnect.RunServiceClient)
	}{
		{
			name: "GetRun_Error_Returns_NotFound",
			setupMocks: func(actionRepo *repoMocks.ActionRepo) {
				actionRepo.EXPECT().GetRun(mock.Anything, matchRunID(runID)).Return(nil, fmt.Errorf("run not found"))
			},
			test: func(t *testing.T, client workflowconnect.RunServiceClient) {
				stream, err := client.WatchRunDetails(context.Background(), connect.NewRequest(&workflow.WatchRunDetailsRequest{
					RunId: runID,
				}))
				require.NoError(t, err)
				assert.False(t, stream.Receive())
				require.Error(t, stream.Err())
				var connectErr *connect.Error
				require.True(t, errors.As(stream.Err(), &connectErr))
				assert.Equal(t, connect.CodeNotFound, connectErr.Code())
			},
		},
		{
			name: "Initial_State_Has_RunSpec_And_Attempts",
			setupMocks: func(actionRepo *repoMocks.ActionRepo) {
				actionRepo.EXPECT().GetRun(mock.Anything, matchRunID(runID)).Return(runModel, nil)
				actionRepo.EXPECT().ListEvents(mock.Anything, matchActionID(rootActionID), 500).Return([]*models.ActionEvent{}, nil)
				actionRepo.EXPECT().WatchRunUpdates(mock.Anything, matchRunID(runID), mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, _ *common.RunIdentifier, _ chan<- *models.Run, _ chan<- error) {
						<-ctx.Done()
					}).Call.Maybe()
			},
			test: func(t *testing.T, client workflowconnect.RunServiceClient) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				stream, err := client.WatchRunDetails(ctx, connect.NewRequest(&workflow.WatchRunDetailsRequest{
					RunId: runID,
				}))
				require.NoError(t, err)

				require.True(t, stream.Receive())
				details := stream.Msg().Details
				require.NotNil(t, details)
				assert.Equal(t, "prod", details.GetRunSpec().GetLabels().GetValues()["env"])
				assert.True(t, proto.Equal(rootActionID, details.GetAction().GetId()))
				assert.Equal(t, common.ActionPhase_ACTION_PHASE_RUNNING, details.GetAction().GetStatus().GetPhase())

				cancel()
				assert.False(t, stream.Receive())
			},
		},
		{
			name: "Watch_Error_Returns_Internal_Error",
			setupMocks: func(actionRepo *repoMocks.ActionRepo) {
				actionRepo.EXPECT().GetRun(mock.Anything, matchRunID(runID)).Return(runModel, nil)
				actionRepo.EXPECT().ListEvents(mock.Anything, matchActionID(rootActionID), 500).Return([]*models.ActionEvent{}, nil)
				actionRepo.EXPECT().WatchRunUpdates(mock.Anything, matchRunID(runID), mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ *common.RunIdentifier, _ chan<- *models.Run, errs chan<- error) {
						errs <- fmt.Errorf("db watch failed")
					})
			},
			test: func(t *testing.T, client workflowconnect.RunServiceClient) {
				stream, err := client.WatchRunDetails(context.Background(), connect.NewRequest(&workflow.WatchRunDetailsRequest{
					RunId: runID,
				}))
				require.NoError(t, err)

				assert.True(t, stream.Receive())
				assert.False(t, stream.Receive())
				require.Error(t, stream.Err())
				var connectErr *connect.Error
				require.True(t, errors.As(stream.Err(), &connectErr))
				assert.Equal(t, connect.CodeInternal, connectErr.Code())
			},
		},
		{
			name: "Run_Update_Streams_Attempts_And_ErrorInfo",
			setupMocks: func(actionRepo *repoMocks.ActionRepo) {
				actionRepo.EXPECT().GetRun(mock.Anything, matchRunID(runID)).Return(runModel, nil)
				actionRepo.EXPECT().ListEvents(mock.Anything, matchActionID(rootActionID), 500).Return([]*models.ActionEvent{}, nil).Once()
				actionRepo.EXPECT().ListEvents(mock.Anything, matchActionID(rootActionID), 500).Return(failedEvents, nil).Once()
				actionRepo.EXPECT().WatchRunUpdates(mock.Anything, matchRunID(runID), mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, _ *common.RunIdentifier, updates chan<- *models.Run, _ chan<- error) {
						updates <- failedModel
						<-ctx.Done()
					})
			},
			test: func(t *testing.T, client workflowconnect.RunServiceClient) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				stream, err := client.WatchRunDetails(ctx, connect.NewRequest(&workflow.WatchRunDetailsRequest{
					RunId: runID,
				}))
				require.NoError(t, err)

				require.True(t, stream.Receive())
				require.True(t, stream.Receive())
				details := stream.Msg().Details
				require.NotNil(t, details)
				assert.Equal(t, "prod", details.GetRunSpec().GetLabels().GetValues()["env"])
				action := details.GetAction()
				assert.Equal(t, common.ActionPhase_ACTION_PHASE_FAILED, action.GetStatus().GetPhase())
				require.Len(t, action.GetAttempts(), 1)
				assert.Equal(t, workflow.ErrorInfo_KIND_USER, action.GetErrorInfo().GetKind())
				assert.Equal(t, "boom", action.GetErrorInfo().GetMessage())

				cancel()
				assert.False(t, stream.Receive())
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actionRepo, svc := newTestRun(t)
			tc.setupMocks(actionRepo)
			client := newRunServiceTestClient(t, svc)
			tc.test(t, client)
		})
	}
}

func TestWatchRuns_RelistsOnResync(t *testing.T) {
	actionRepo, svc := newTestRun(t)
	client := newRunServiceTestClient(t, svc)

	run := func(phase common.ActionPhase) *models.Run {
		return &models.Run{Project: "p", Domain: "d", RunName: "r1", Name: "a0", Phase: int32(phase)}
	}
	actionRepo.On("WatchAllRunUpdates", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			args.Get(1).(chan<- *models.Run) <- nil
			<-args.Get(0).(context.Context).Done()
		}).Once()
	actionRepo.On("ListActions", mock.Anything, mock.Anything).
		Return([]*models.Run{run(common.ActionPhase_ACTION_PHASE_RUNNING)}, nil).Once()
	actionRepo.On("ListActions", mock.Anything, mock.Anything).
		Return([]*models.Run{run(common.ActionPhase_ACTION_PHASE_SUCCEEDED)}, nil).Once()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.WatchRuns(ctx, connect.NewRequest(&workflow.WatchRunsRequest{}))
	require.NoError(t, err)

	require.True(t, stream.Receive())
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_RUNNING, stream.Msg().GetRuns()[0].GetAction().GetStatus().GetPhase())
	require.True(t, stream.Receive())
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_SUCCEEDED, stream.Msg().GetRuns()[0].GetAction().GetStatus().GetPhase())
}

func TestWatchActions_RelistsOnResync(t *testing.T) {
	actionRepo, svc := newTestRun(t)
	client := newRunServiceTestClient(t, svc)

	runID := &common.RunIdentifier{Project: "p", Domain: "d", Name: "r1"}
	root := func(phase common.ActionPhase, updatedAt time.Time) *models.Action {
		return &models.Action{Project: "p", Domain: "d", RunName: "r1", Name: "a0", Phase: int32(phase), UpdatedAt: updatedAt}
	}
	child := &models.Action{
		Project: "p", Domain: "d", RunName: "r1", Name: "a1",
		ParentActionName: sql.NullString{String: "a0", Valid: true},
		Phase:            int32(common.ActionPhase_ACTION_PHASE_RUNNING),
		UpdatedAt:        time.Unix(100, 0),
	}
	actionRepo.On("WatchAllActionUpdates", mock.Anything, matchRunID(runID), mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			args.Get(2).(chan<- *models.Action) <- nil
			<-args.Get(0).(context.Context).Done()
		}).Once()
	actionRepo.On("ListActions", mock.Anything, mock.Anything).
		Return([]*models.Action{root(common.ActionPhase_ACTION_PHASE_RUNNING, time.Unix(100, 0)), child}, nil).Once()
	actionRepo.On("ListActions", mock.Anything, mock.Anything).
		Return([]*models.Action{root(common.ActionPhase_ACTION_PHASE_SUCCEEDED, time.Unix(200, 0)), child}, nil).Once()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.WatchActions(ctx, connect.NewRequest(&workflow.WatchActionsRequest{RunId: runID}))
	require.NoError(t, err)

	require.True(t, stream.Receive())
	assert.Len(t, stream.Msg().GetEnrichedActions(), 2)
	require.True(t, stream.Receive())
	resent := stream.Msg().GetEnrichedActions()
	require.Len(t, resent, 1, "the unchanged child must not be sent again")
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_SUCCEEDED, resent[0].GetAction().GetStatus().GetPhase())
}

func TestWatchActions_KeepsStreamOnResyncListError(t *testing.T) {
	actionRepo, svc := newTestRun(t)
	client := newRunServiceTestClient(t, svc)

	runID := &common.RunIdentifier{Project: "p", Domain: "d", Name: "r1"}
	root := func(phase common.ActionPhase, updatedAt time.Time) *models.Action {
		return &models.Action{Project: "p", Domain: "d", RunName: "r1", Name: "a0", Phase: int32(phase), UpdatedAt: updatedAt}
	}
	actionRepo.On("WatchAllActionUpdates", mock.Anything, matchRunID(runID), mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			updates := args.Get(2).(chan<- *models.Action)
			updates <- nil
			updates <- root(common.ActionPhase_ACTION_PHASE_SUCCEEDED, time.Unix(200, 0))
			<-args.Get(0).(context.Context).Done()
		}).Once()
	actionRepo.On("ListActions", mock.Anything, mock.Anything).
		Return([]*models.Action{root(common.ActionPhase_ACTION_PHASE_RUNNING, time.Unix(100, 0))}, nil).Once()
	actionRepo.On("ListActions", mock.Anything, mock.Anything).
		Return(nil, errors.New("db down")).Once()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.WatchActions(ctx, connect.NewRequest(&workflow.WatchActionsRequest{RunId: runID}))
	require.NoError(t, err)

	require.True(t, stream.Receive())
	require.True(t, stream.Receive(), "a failed re-list must not end the stream: %v", stream.Err())
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_SUCCEEDED, stream.Msg().GetEnrichedActions()[0].GetAction().GetStatus().GetPhase())
}
