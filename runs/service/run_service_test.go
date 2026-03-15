package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/datatypes"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	repoMocks "github.com/flyteorg/flyte/v2/runs/repository/mocks"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// mockActionsClient implements actionsconnect.ActionsServiceClient for testing.
type mockActionsClient struct {
	mock.Mock
}

func (m *mockActionsClient) Enqueue(ctx context.Context, req *connect.Request[actions.EnqueueRequest]) (*connect.Response[actions.EnqueueResponse], error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*connect.Response[actions.EnqueueResponse]), args.Error(1)
}

func (m *mockActionsClient) GetLatestState(ctx context.Context, req *connect.Request[actions.GetLatestStateRequest]) (*connect.Response[actions.GetLatestStateResponse], error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*connect.Response[actions.GetLatestStateResponse]), args.Error(1)
}

func (m *mockActionsClient) WatchForUpdates(ctx context.Context, req *connect.Request[actions.WatchForUpdatesRequest]) (*connect.ServerStreamForClient[actions.WatchForUpdatesResponse], error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*connect.ServerStreamForClient[actions.WatchForUpdatesResponse]), args.Error(1)
}

func (m *mockActionsClient) Update(ctx context.Context, req *connect.Request[actions.UpdateRequest]) (*connect.Response[actions.UpdateResponse], error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*connect.Response[actions.UpdateResponse]), args.Error(1)
}

func (m *mockActionsClient) Abort(ctx context.Context, req *connect.Request[actions.AbortRequest]) (*connect.Response[actions.AbortResponse], error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*connect.Response[actions.AbortResponse]), args.Error(1)
}

func newTestService(t *testing.T) (*repoMocks.ActionRepo, *mockActionsClient, *RunService) {
	actionRepo := &repoMocks.ActionRepo{}
	actionsClient := &mockActionsClient{}
	repo := &repoMocks.Repository{}
	repo.On("ActionRepo").Return(actionRepo)

	svc := &RunService{repo: repo, actionsClient: actionsClient}

	t.Cleanup(func() {
		repo.AssertExpectations(t)
		actionRepo.AssertExpectations(t)
		actionsClient.AssertExpectations(t)
	})

	return actionRepo, actionsClient, svc
}

func TestAbortRun(t *testing.T) {
	runID := &common.RunIdentifier{
		Org:     "test-org",
		Project: "test-project",
		Domain:  "test-domain",
		Name:    "rtest12345",
	}

	t.Run("success with default reason", func(t *testing.T) {
		actionRepo, actionsClient, svc := newTestService(t)

		actionRepo.On("AbortRun", mock.Anything, runID, "User requested abort", (*common.EnrichedIdentity)(nil)).Return(nil)
		actionsClient.On("Abort", mock.Anything, mock.MatchedBy(func(req *connect.Request[actions.AbortRequest]) bool {
			return req.Msg.ActionId.Run.Name == runID.Name &&
				req.Msg.ActionId.Name == "a0" &&
				req.Msg.Reason != nil && *req.Msg.Reason == "User requested abort"
		})).Return(connect.NewResponse(&actions.AbortResponse{}), nil)

		_, err := svc.AbortRun(context.Background(), connect.NewRequest(&workflow.AbortRunRequest{RunId: runID}))
		assert.NoError(t, err)
	})

	t.Run("success with custom reason", func(t *testing.T) {
		actionRepo, actionsClient, svc := newTestService(t)
		reason := "timeout exceeded"

		actionRepo.On("AbortRun", mock.Anything, runID, reason, (*common.EnrichedIdentity)(nil)).Return(nil)
		actionsClient.On("Abort", mock.Anything, mock.MatchedBy(func(req *connect.Request[actions.AbortRequest]) bool {
			return req.Msg.Reason != nil && *req.Msg.Reason == reason
		})).Return(connect.NewResponse(&actions.AbortResponse{}), nil)

		_, err := svc.AbortRun(context.Background(), connect.NewRequest(&workflow.AbortRunRequest{RunId: runID, Reason: &reason}))
		assert.NoError(t, err)
	})

	t.Run("db error stops before actions service", func(t *testing.T) {
		actionRepo, actionsClient, svc := newTestService(t)

		actionRepo.On("AbortRun", mock.Anything, runID, mock.Anything, mock.Anything).Return(errors.New("db unavailable"))

		_, err := svc.AbortRun(context.Background(), connect.NewRequest(&workflow.AbortRunRequest{RunId: runID}))
		assert.Error(t, err)
		actionsClient.AssertNotCalled(t, "Abort")
	})

	t.Run("actions service error is returned", func(t *testing.T) {
		actionRepo, actionsClient, svc := newTestService(t)

		actionRepo.On("AbortRun", mock.Anything, runID, mock.Anything, mock.Anything).Return(nil)
		actionsClient.On("Abort", mock.Anything, mock.Anything).Return(nil, errors.New("actions service unavailable"))

		_, err := svc.AbortRun(context.Background(), connect.NewRequest(&workflow.AbortRunRequest{RunId: runID}))
		assert.Error(t, err)
	})
}

func TestAbortAction(t *testing.T) {
	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     "test-org",
			Project: "test-project",
			Domain:  "test-domain",
			Name:    "rtest12345",
		},
		Name: "action-1",
	}

	t.Run("success with default reason", func(t *testing.T) {
		actionRepo, actionsClient, svc := newTestService(t)

		actionRepo.On("AbortAction", mock.Anything, actionID, "User requested abort", (*common.EnrichedIdentity)(nil)).Return(nil)
		actionsClient.On("Abort", mock.Anything, mock.MatchedBy(func(req *connect.Request[actions.AbortRequest]) bool {
			return req.Msg.ActionId.Name == actionID.Name &&
				req.Msg.Reason != nil && *req.Msg.Reason == "User requested abort"
		})).Return(connect.NewResponse(&actions.AbortResponse{}), nil)

		_, err := svc.AbortAction(context.Background(), connect.NewRequest(&workflow.AbortActionRequest{ActionId: actionID}))
		assert.NoError(t, err)
	})

	t.Run("success with custom reason", func(t *testing.T) {
		actionRepo, actionsClient, svc := newTestService(t)
		reason := "resource limit exceeded"

		actionRepo.On("AbortAction", mock.Anything, actionID, reason, (*common.EnrichedIdentity)(nil)).Return(nil)
		actionsClient.On("Abort", mock.Anything, mock.MatchedBy(func(req *connect.Request[actions.AbortRequest]) bool {
			return req.Msg.Reason != nil && *req.Msg.Reason == reason
		})).Return(connect.NewResponse(&actions.AbortResponse{}), nil)

		_, err := svc.AbortAction(context.Background(), connect.NewRequest(&workflow.AbortActionRequest{ActionId: actionID, Reason: reason}))
		assert.NoError(t, err)
	})

	t.Run("db error stops before actions service", func(t *testing.T) {
		actionRepo, actionsClient, svc := newTestService(t)

		actionRepo.On("AbortAction", mock.Anything, actionID, mock.Anything, mock.Anything).Return(errors.New("db unavailable"))

		_, err := svc.AbortAction(context.Background(), connect.NewRequest(&workflow.AbortActionRequest{ActionId: actionID}))
		assert.Error(t, err)
		actionsClient.AssertNotCalled(t, "Abort")
	})

	t.Run("actions service error is returned", func(t *testing.T) {
		actionRepo, actionsClient, svc := newTestService(t)

		actionRepo.On("AbortAction", mock.Anything, actionID, mock.Anything, mock.Anything).Return(nil)
		actionsClient.On("Abort", mock.Anything, mock.Anything).Return(nil, errors.New("actions service unavailable"))

		_, err := svc.AbortAction(context.Background(), connect.NewRequest(&workflow.AbortActionRequest{ActionId: actionID}))
		assert.Error(t, err)
	})
}

func TestListRuns(t *testing.T) {
	actionRepo, _, svc := newTestService(t)
	runs := []*workflow.Run{}
	sqlRes := []*models.Run{}
	for i := 0; i < 10; i++ {
		runs = append(runs, &workflow.Run{
			Action: &workflow.Action{
				Id: &common.ActionIdentifier{
					Run: &common.RunIdentifier{
						Org:     "test-org",
						Project: "test-project",
						Domain:  "test-domain",
						Name:    fmt.Sprintf("run-%d", i),
					},
					Name: fmt.Sprintf("run-%d", i),
				},
				Metadata: &workflow.ActionMetadata{},
				Status:   &workflow.ActionStatus{Phase: common.ActionPhase_ACTION_PHASE_SUCCEEDED},
			},
		})
		sqlRes = append(sqlRes, &models.Run{
			ID:      uint(i),
			Org:     "test-org",
			Project: "test-project",
			Domain:  "test-domain",
			Name:    fmt.Sprintf("run-%d", i),
			Phase:   int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED),
		})
	}
	type mockListRes struct {
		runs  []*models.Run
		token string
		err   error
	}
	testCases := []struct {
		name    string
		req     *common.ListRequest
		mockRes mockListRes
		expect  *workflow.ListRunsResponse
	}{
		{
			"Empty Runs",
			&common.ListRequest{Limit: 2},
			mockListRes{runs: []*models.Run{}, err: nil},
			&workflow.ListRunsResponse{Runs: []*workflow.Run{}, Token: ""},
		},
		{
			"list with limit 2 and token",
			&common.ListRequest{Limit: 2, Token: "5"},
			mockListRes{runs: sqlRes[5:7], token: "7", err: nil},
			&workflow.ListRunsResponse{Runs: runs[5:7], Token: "7"},
		},
		{
			"list with limit 3 and token",
			&common.ListRequest{Limit: 3, Token: "8"},
			mockListRes{runs: sqlRes[8:10], token: "", err: nil},
			&workflow.ListRunsResponse{Runs: runs[8:10], Token: ""},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := connect.NewRequest(&workflow.ListRunsRequest{Request: tc.req})
			actionRepo.On("ListRuns", mock.Anything, req.Msg).Return(tc.mockRes.runs, tc.mockRes.token, tc.mockRes.err)
			got, err := svc.ListRuns(context.Background(), req)
			assert.NoError(t, err)
			assert.Equal(t, len(tc.expect.Runs), len(got.Msg.Runs))
			assert.Equal(t, tc.expect.Runs, got.Msg.Runs)
			assert.Equal(t, tc.expect.Token, got.Msg.Token)
		})
	}
}

func TestGenerateRunName(t *testing.T) {
	t.Run("starts with r prefix", func(t *testing.T) {
		name := generateRunName(42)
		assert.True(t, strings.HasPrefix(name, "r"), "run name should start with 'r', got: %s", name)
	})

	t.Run("has correct length", func(t *testing.T) {
		name := generateRunName(42)
		assert.Equal(t, runIDLength, len(name))
	})

	t.Run("different seeds produce different names", func(t *testing.T) {
		name1 := generateRunName(1)
		name2 := generateRunName(2)
		assert.NotEqual(t, name1, name2)
	})

	t.Run("same seed produces same name", func(t *testing.T) {
		name1 := generateRunName(12345)
		name2 := generateRunName(12345)
		assert.Equal(t, name1, name2)
	})
}

func TestConvertRunToProto(t *testing.T) {
	// Define the test cases as a slice of structs.
	s := &RunService{}
	name := generateRunName(int64(0))
	org := "test_org"
	project := "test_project"
	domain := "test_domain"
	startTime := timestamppb.Now()
	endTime := timestamppb.New(startTime.AsTime().Add(time.Minute))
	durationMs := uint64(endTime.AsTime().Sub(startTime.AsTime()).Milliseconds())
	status := &workflow.ActionStatus{
		Phase:       common.ActionPhase_ACTION_PHASE_SUCCEEDED,
		StartTime:   startTime,
		EndTime:     endTime,
		Attempts:    uint32(1),
		CacheStatus: core.CatalogCacheStatus_CACHE_DISABLED,
		DurationMs:  &durationMs,
	}
	detail := &workflow.ActionDetails{
		Status: status,
	}
	detailJson, _ := json.Marshal(detail)
	testCases := []struct {
		name   string
		run    *models.Run
		expect *workflow.Run
	}{
		{"empty run", nil, nil},
		{"valid run",
			&models.Run{
				ID:            uint(0),
				Org:           org,
				Project:       project,
				Domain:        domain,
				Name:          name,
				Phase:         int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED),
				ActionDetails: datatypes.JSON(detailJson),
			},
			&workflow.Run{
				Action: &workflow.Action{
					Id: &common.ActionIdentifier{
						Run: &common.RunIdentifier{
							Org:     org,
							Project: project,
							Domain:  domain,
							Name:    name,
						},
						Name: name,
					},
					Metadata: &workflow.ActionMetadata{},
					Status:   status,
				},
			},
		},
		{
			"run with missing details",
			&models.Run{
				ID:      uint(0),
				Org:     org,
				Project: project,
				Domain:  domain,
				Name:    name,
				Phase:   int32(common.ActionPhase_ACTION_PHASE_QUEUED),
			},
			&workflow.Run{
				Action: &workflow.Action{
					Id: &common.ActionIdentifier{
						Run: &common.RunIdentifier{
							Org:     org,
							Project: project,
							Domain:  domain,
							Name:    name,
						},
						Name: name,
					},
					Metadata: &workflow.ActionMetadata{},
					Status:   &workflow.ActionStatus{Phase: common.ActionPhase_ACTION_PHASE_QUEUED},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := s.convertRunToProto(tc.run)
			if tc.expect == nil {
				assert.Nil(t, res)
				return
			}
			assert.Equal(t, res.Action.Id, tc.expect.Action.Id)
			assert.Equal(t, res.Action.Metadata, tc.expect.Action.Metadata)
			assert.Equal(t, res.Action.Status, tc.expect.Action.Status)
		})
	}
}
