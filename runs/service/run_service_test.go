package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	storageMocks "github.com/flyteorg/flyte/v2/flytestdlib/storage/mocks"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
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
	taskRepo := &repoMocks.TaskRepo{}
	actionsClient := &mockActionsClient{}
	repo := &repoMocks.Repository{}
	repo.On("ActionRepo").Return(actionRepo)
	repo.On("TaskRepo").Maybe().Return(taskRepo)

	svc := &RunService{repo: repo, actionsClient: actionsClient}

	t.Cleanup(func() {
		repo.AssertExpectations(t)
		actionRepo.AssertExpectations(t)
		taskRepo.AssertExpectations(t)
		actionsClient.AssertExpectations(t)
	})

	return actionRepo, actionsClient, svc
}

func matchActionID(expected *common.ActionIdentifier) interface{} {
	return mock.MatchedBy(func(actual *common.ActionIdentifier) bool {
		return proto.Equal(actual, expected)
	})
}

func TestGetRunDetails_WithTaskSpec(t *testing.T) {
	actionRepo := &repoMocks.ActionRepo{}
	taskRepo := &repoMocks.TaskRepo{}
	actionsClient := &mockActionsClient{}
	repo := &repoMocks.Repository{}
	repo.On("ActionRepo").Return(actionRepo)
	repo.On("TaskRepo").Maybe().Return(taskRepo)

	svc := &RunService{repo: repo, actionsClient: actionsClient}

	t.Cleanup(func() {
		repo.AssertExpectations(t)
		actionRepo.AssertExpectations(t)
		taskRepo.AssertExpectations(t)
		actionsClient.AssertExpectations(t)
	})

	runID := &common.RunIdentifier{
		Org:     "test-org",
		Project: "test-project",
		Domain:  "test-domain",
		Name:    "rtest12345",
	}
	rootActionID := &common.ActionIdentifier{Run: runID, Name: runID.Name}

	runInfo := &workflow.RunInfo{
		TaskSpecDigest: "abc123",
		InputsUri:      "s3://bucket/inputs.pb",
	}
	runInfoBytes, _ := proto.Marshal(runInfo)

	runModel := &models.Run{
		Org:          runID.Org,
		Project:      runID.Project,
		Domain:       runID.Domain,
		RunName:      runID.Name,
		Name:         runID.Name,
		Phase:        int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED),
		ActionType:   int32(workflow.ActionType_ACTION_TYPE_TASK),
		DetailedInfo: runInfoBytes,
	}

	taskSpecProto := &task.TaskSpec{
		TaskTemplate: &core.TaskTemplate{
			Type: "python",
		},
	}
	taskSpecBytes, _ := proto.Marshal(taskSpecProto)

	actionRepo.On("GetRun", mock.Anything, runID).Return(runModel, nil)
	taskRepo.On("GetTaskSpec", mock.Anything, "abc123").Return(&models.TaskSpec{
		Digest: "abc123",
		Spec:   taskSpecBytes,
	}, nil)
	actionRepo.On("ListEvents", mock.Anything, matchActionID(rootActionID), 500).Return([]*models.ActionEvent{}, nil)

	resp, err := svc.GetRunDetails(context.Background(), connect.NewRequest(&workflow.GetRunDetailsRequest{
		RunId: runID,
	}))
	require.NoError(t, err)
	require.NotNil(t, resp.Msg.Details)
	require.NotNil(t, resp.Msg.Details.Action)
	require.NotNil(t, resp.Msg.Details.Action.GetTask())
	assert.Equal(t, "python", resp.Msg.Details.Action.GetTask().GetTaskTemplate().GetType())
}

func TestGetRunDetails_UsesActionCacheStatus(t *testing.T) {
	actionRepo := &repoMocks.ActionRepo{}
	taskRepo := &repoMocks.TaskRepo{}
	actionsClient := &mockActionsClient{}
	repo := &repoMocks.Repository{}
	repo.On("ActionRepo").Return(actionRepo)
	repo.On("TaskRepo").Maybe().Return(taskRepo)

	svc := &RunService{repo: repo, actionsClient: actionsClient}

	t.Cleanup(func() {
		repo.AssertExpectations(t)
		actionRepo.AssertExpectations(t)
		taskRepo.AssertExpectations(t)
		actionsClient.AssertExpectations(t)
	})

	runID := &common.RunIdentifier{
		Org:     "test-org",
		Project: "test-project",
		Domain:  "test-domain",
		Name:    "rtest12345",
	}
	rootActionID := &common.ActionIdentifier{Run: runID, Name: runID.Name}

	runModel := &models.Run{
		Org:         runID.Org,
		Project:     runID.Project,
		Domain:      runID.Domain,
		RunName:     runID.Name,
		Name:        runID.Name,
		Phase:       int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED),
		ActionType:  int32(workflow.ActionType_ACTION_TYPE_TASK),
		Attempts:    1,
		CacheStatus: core.CatalogCacheStatus_CACHE_HIT,
	}

	now := time.Now()
	event := &workflow.ActionEvent{
		Id:          rootActionID,
		Attempt:     1,
		Phase:       common.ActionPhase_ACTION_PHASE_SUCCEEDED,
		Version:     1,
		UpdatedTime: timestamppb.New(now),
	}
	eventModel, _ := models.NewActionEventModel(event)

	actionRepo.On("GetRun", mock.Anything, runID).Return(runModel, nil)
	actionRepo.On("ListEvents", mock.Anything, matchActionID(rootActionID), 500).Return([]*models.ActionEvent{eventModel}, nil)

	resp, err := svc.GetRunDetails(context.Background(), connect.NewRequest(&workflow.GetRunDetailsRequest{
		RunId: runID,
	}))
	require.NoError(t, err)
	require.NotNil(t, resp.Msg.Details)
	assert.Equal(t, core.CatalogCacheStatus_CACHE_HIT, resp.Msg.Details.Action.Status.GetCacheStatus())
}

func TestGetRunDetails_TaskSpecLookupFails(t *testing.T) {
	actionRepo := &repoMocks.ActionRepo{}
	taskRepo := &repoMocks.TaskRepo{}
	actionsClient := &mockActionsClient{}
	repo := &repoMocks.Repository{}
	repo.On("ActionRepo").Return(actionRepo)
	repo.On("TaskRepo").Maybe().Return(taskRepo)

	svc := &RunService{repo: repo, actionsClient: actionsClient}

	t.Cleanup(func() {
		repo.AssertExpectations(t)
		actionRepo.AssertExpectations(t)
		taskRepo.AssertExpectations(t)
		actionsClient.AssertExpectations(t)
	})

	runID := &common.RunIdentifier{
		Org:     "test-org",
		Project: "test-project",
		Domain:  "test-domain",
		Name:    "rtest12345",
	}
	runInfo := &workflow.RunInfo{TaskSpecDigest: "bad-digest"}
	runInfoBytes, _ := proto.Marshal(runInfo)

	runModel := &models.Run{
		Org:          runID.Org,
		Project:      runID.Project,
		Domain:       runID.Domain,
		RunName:      runID.Name,
		Name:         runID.Name,
		Phase:        int32(common.ActionPhase_ACTION_PHASE_RUNNING),
		ActionType:   int32(workflow.ActionType_ACTION_TYPE_TASK),
		DetailedInfo: runInfoBytes,
	}

	actionRepo.On("GetRun", mock.Anything, runID).Return(runModel, nil)
	taskRepo.On("GetTaskSpec", mock.Anything, "bad-digest").Return(nil, errors.New("not found"))
	actionRepo.On("ListEvents", mock.Anything, matchActionID(&common.ActionIdentifier{Run: runID, Name: runID.Name}), 500).Return([]*models.ActionEvent{}, nil)

	_, err := svc.GetRunDetails(context.Background(), connect.NewRequest(&workflow.GetRunDetailsRequest{
		RunId: runID,
	}))
	assert.Error(t, err)
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
	baseTime := time.Date(2026, time.March, 24, 10, 0, 0, 0, time.UTC)
	for i := 0; i < 10; i++ {
		startTime := timestamppb.New(baseTime.Add(time.Duration(i) * time.Minute))
		endTime := timestamppb.New(startTime.AsTime().Add(2 * time.Minute))
		durationMs := uint64(2 * time.Minute / time.Millisecond)
		envName := fmt.Sprintf("env-%d", i)
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
				Metadata: &workflow.ActionMetadata{
					EnvironmentName: envName,
				},
				Status: &workflow.ActionStatus{
					Phase:      common.ActionPhase_ACTION_PHASE_SUCCEEDED,
					StartTime:  startTime,
					EndTime:    endTime,
					DurationMs: &durationMs,
				},
			},
		})
		sqlRes = append(sqlRes, &models.Run{
			ID:         uint(i),
			Org:        "test-org",
			Project:    "test-project",
			Domain:     "test-domain",
			Name:       fmt.Sprintf("run-%d", i),
			Phase:      int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED),
			CreatedAt:  startTime.AsTime(),
			EndedAt:    sql.NullTime{Time: endTime.AsTime(), Valid: true},
			DurationMs: sql.NullInt64{Int64: int64(durationMs), Valid: true},
			EnvironmentName: sql.NullString{
				String: envName,
				Valid:  true,
			},
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
	startTime := timestamppb.New(time.Date(2026, time.March, 24, 9, 0, 0, 0, time.UTC))
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
				CreatedAt:     startTime.AsTime(),
				ActionDetails: detailJson,
				EnvironmentName: sql.NullString{
					String: "prod",
					Valid:  true,
				},
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
					Metadata: &workflow.ActionMetadata{
						EnvironmentName: "prod",
					},
					Status: status,
				},
			},
		},
		{
			"run with missing details",
			&models.Run{
				ID:         uint(0),
				Org:        org,
				Project:    project,
				Domain:     domain,
				Name:       name,
				Phase:      int32(common.ActionPhase_ACTION_PHASE_QUEUED),
				CreatedAt:  startTime.AsTime(),
				EndedAt:    sql.NullTime{Time: endTime.AsTime(), Valid: true},
				DurationMs: sql.NullInt64{Int64: int64(durationMs), Valid: true},
				EnvironmentName: sql.NullString{
					String: "staging",
					Valid:  true,
				},
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
					Metadata: &workflow.ActionMetadata{
						EnvironmentName: "staging",
					},
					Status: &workflow.ActionStatus{
						Phase:      common.ActionPhase_ACTION_PHASE_QUEUED,
						StartTime:  startTime,
						EndTime:    endTime,
						DurationMs: &durationMs,
					},
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

func newStringLiteral(s string) *core.Literal {
	return &core.Literal{Value: &core.Literal_Scalar{
		Scalar: &core.Scalar{Value: &core.Scalar_Primitive{
			Primitive: &core.Primitive{Value: &core.Primitive_StringValue{StringValue: s}},
		}},
	}}
}

func newIntLiteral(n int64) *core.Literal {
	return &core.Literal{Value: &core.Literal_Scalar{
		Scalar: &core.Scalar{Value: &core.Scalar_Primitive{
			Primitive: &core.Primitive{Value: &core.Primitive_Integer{Integer: n}},
		}},
	}}
}

// TestInputsProtoCompat verifies that task.Inputs and core.LiteralMap are wire-compatible,
// ensuring components that still read/write inputs.pb as LiteralMap work as expect.
func TestInputsProtoCompat(t *testing.T) {
	t.Run("task.Inputs roundtrip preserves literals and context", func(t *testing.T) {
		inputs := &task.Inputs{
			Literals: []*task.NamedLiteral{
				{Name: "x", Value: newStringLiteral("hello")},
				{Name: "y", Value: newIntLiteral(42)},
			},
			Context: []*core.KeyValuePair{
				{Key: "env", Value: "prod"},
				{Key: "region", Value: "us-east-1"},
			},
		}

		data, err := proto.Marshal(inputs)
		require.NoError(t, err)

		got := &task.Inputs{}
		require.NoError(t, proto.Unmarshal(data, got))

		assert.Len(t, got.Literals, 2)
		assert.Equal(t, "x", got.Literals[0].Name)
		assert.Equal(t, "y", got.Literals[1].Name)
		assert.Len(t, got.Context, 2)
		assert.Equal(t, "env", got.Context[0].Key)
		assert.Equal(t, "prod", got.Context[0].Value)
	})

	t.Run("task.Inputs read as core.LiteralMap preserves literals", func(t *testing.T) {
		inputs := &task.Inputs{
			Literals: []*task.NamedLiteral{
				{Name: "x", Value: newStringLiteral("hello")},
				{Name: "y", Value: newIntLiteral(42)},
			},
			Context: []*core.KeyValuePair{
				{Key: "env", Value: "prod"},
			},
		}

		data, err := proto.Marshal(inputs)
		require.NoError(t, err)

		literalMap := &core.LiteralMap{}
		require.NoError(t, proto.Unmarshal(data, literalMap))

		assert.Len(t, literalMap.Literals, 2)
		assert.Contains(t, literalMap.Literals, "x")
		assert.Contains(t, literalMap.Literals, "y")
		assert.True(t, proto.Equal(newStringLiteral("hello"), literalMap.Literals["x"]))
		assert.True(t, proto.Equal(newIntLiteral(42), literalMap.Literals["y"]))
	})

	t.Run("core.LiteralMap read as task.Inputs preserves literals", func(t *testing.T) {
		literalMap := &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"a": newStringLiteral("old_value"),
			},
		}

		data, err := proto.Marshal(literalMap)
		require.NoError(t, err)

		got := &task.Inputs{}
		require.NoError(t, proto.Unmarshal(data, got))

		assert.Len(t, got.Literals, 1)
		assert.Equal(t, "a", got.Literals[0].Name)
		assert.True(t, proto.Equal(newStringLiteral("old_value"), got.Literals[0].Value))
		assert.Empty(t, got.Context)
	})
}

func TestCreateRun_WritesEmptyInputsProto(t *testing.T) {
	actionRepo := &repoMocks.ActionRepo{}
	actionsClient := &mockActionsClient{}
	repo := &repoMocks.Repository{}
	store := &storageMocks.ComposedProtobufStore{}
	dataStore := &storage.DataStore{ComposedProtobufStore: store}

	repo.On("ActionRepo").Return(actionRepo)

	svc := &RunService{
		repo:          repo,
		actionsClient: actionsClient,
		storagePrefix: "s3://flyte-data",
		dataStore:     dataStore,
	}

	req := &workflow.CreateRunRequest{
		Id: &workflow.CreateRunRequest_ProjectId{
			ProjectId: &common.ProjectIdentifier{
				Organization: "testorg",
				Domain:       "development",
				Name:         "testproject",
			},
		},
		Task: &workflow.CreateRunRequest_TaskSpec{
			TaskSpec: &task.TaskSpec{},
		},
	}

	expectedRun := &models.Run{
		Org:     "testorg",
		Project: "testproject",
		Domain:  "development",
		Name:    "generated-run",
	}

	store.On("WriteProtobuf", mock.Anything, mock.AnythingOfType("storage.DataReference"), storage.Options{}, mock.MatchedBy(func(msg proto.Message) bool {
		inputs, ok := msg.(*task.Inputs)
		return ok && len(inputs.Literals) == 0 && len(inputs.Context) == 0
	})).Return(nil).Once()

	actionRepo.On("CreateRun", mock.Anything, mock.AnythingOfType("*workflow.CreateRunRequest"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
		Return(expectedRun, nil).Once()

	actionsClient.On("Enqueue", mock.Anything, mock.Anything).
		Return(connect.NewResponse(&actions.EnqueueResponse{}), nil).Once()

	_, err := svc.CreateRun(context.Background(), connect.NewRequest(req))
	assert.NoError(t, err)

	repo.AssertExpectations(t)
	actionRepo.AssertExpectations(t)
	actionsClient.AssertExpectations(t)
	store.AssertExpectations(t)
}

func TestCreateRun_ResponseUsesRunModel(t *testing.T) {
	actionRepo := &repoMocks.ActionRepo{}
	actionsClient := &mockActionsClient{}
	repo := &repoMocks.Repository{}
	store := &storageMocks.ComposedProtobufStore{}
	dataStore := &storage.DataStore{ComposedProtobufStore: store}

	repo.On("ActionRepo").Return(actionRepo)

	svc := &RunService{
		repo:          repo,
		actionsClient: actionsClient,
		storagePrefix: "s3://flyte-data",
		dataStore:     dataStore,
	}

	req := &workflow.CreateRunRequest{
		Id: &workflow.CreateRunRequest_RunId{
			RunId: &common.RunIdentifier{
				Org:     "myorg",
				Project: "myproj",
				Domain:  "production",
				Name:    "rtest99999",
			},
		},
		Task: &workflow.CreateRunRequest_TaskSpec{
			TaskSpec: &task.TaskSpec{},
		},
	}

	expectedRun := &models.Run{
		Org:     "myorg",
		Project: "myproj",
		Domain:  "production",
		Name:    "rtest99999",
		Phase:   int32(common.ActionPhase_ACTION_PHASE_QUEUED),
	}

	store.On("WriteProtobuf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	actionRepo.On("CreateRun", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedRun, nil).Once()
	actionsClient.On("Enqueue", mock.Anything, mock.Anything).Return(connect.NewResponse(&actions.EnqueueResponse{}), nil).Once()

	resp, err := svc.CreateRun(context.Background(), connect.NewRequest(req))
	require.NoError(t, err)

	run := resp.Msg.Run
	assert.Equal(t, "myorg", run.Action.Id.Run.Org)
	assert.Equal(t, "myproj", run.Action.Id.Run.Project)
	assert.Equal(t, "production", run.Action.Id.Run.Domain)
	assert.Equal(t, "rtest99999", run.Action.Id.Run.Name)
	assert.Equal(t, "rtest99999", run.Action.Id.Name)
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_QUEUED, run.Action.Status.Phase)
}

func TestCreateRun_ActionIDUsesRunName(t *testing.T) {
	actionRepo := &repoMocks.ActionRepo{}
	actionsClient := &mockActionsClient{}
	repo := &repoMocks.Repository{}
	store := &storageMocks.ComposedProtobufStore{}
	dataStore := &storage.DataStore{ComposedProtobufStore: store}

	repo.On("ActionRepo").Return(actionRepo)

	svc := &RunService{
		repo:          repo,
		actionsClient: actionsClient,
		storagePrefix: "s3://flyte-data",
		dataStore:     dataStore,
	}

	req := &workflow.CreateRunRequest{
		Id: &workflow.CreateRunRequest_RunId{
			RunId: &common.RunIdentifier{
				Org:     "org",
				Project: "proj",
				Domain:  "dev",
				Name:    "rmy-run-123",
			},
		},
		Task: &workflow.CreateRunRequest_TaskSpec{
			TaskSpec: &task.TaskSpec{},
		},
	}

	store.On("WriteProtobuf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	actionRepo.On("CreateRun", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&models.Run{Org: "org", Project: "proj", Domain: "dev", Name: "rmy-run-123"}, nil).Once()

	// Verify the enqueue request uses the run name as the action name
	actionsClient.On("Enqueue", mock.Anything, mock.MatchedBy(func(req *connect.Request[actions.EnqueueRequest]) bool {
		return req.Msg.Action.ActionId.Name == "rmy-run-123" &&
			req.Msg.Action.ActionId.Run.Name == "rmy-run-123"
	})).Return(connect.NewResponse(&actions.EnqueueResponse{}), nil).Once()

	_, err := svc.CreateRun(context.Background(), connect.NewRequest(req))
	assert.NoError(t, err)

	actionsClient.AssertExpectations(t)
}

func TestCreateRun_PreservesInputContextAndRawDataPath(t *testing.T) {
	actionRepo := &repoMocks.ActionRepo{}
	actionsClient := &mockActionsClient{}
	repo := &repoMocks.Repository{}
	store := &storageMocks.ComposedProtobufStore{}
	dataStore := &storage.DataStore{ComposedProtobufStore: store}

	repo.On("ActionRepo").Return(actionRepo)

	svc := &RunService{
		repo:          repo,
		actionsClient: actionsClient,
		storagePrefix: "s3://flyte-data",
		dataStore:     dataStore,
	}

	req := &workflow.CreateRunRequest{
		Id: &workflow.CreateRunRequest_RunId{
			RunId: &common.RunIdentifier{
				Org:     "org",
				Project: "proj",
				Domain:  "dev",
				Name:    "rctx-123",
			},
		},
		Inputs: &task.Inputs{
			Context: []*core.KeyValuePair{
				{Key: "trace_id", Value: "root-abc"},
			},
		},
		RunSpec: &task.RunSpec{
			RawDataStorage: &task.RawDataStorage{RawDataPrefix: "s3://custom-raw"},
		},
		Task: &workflow.CreateRunRequest_TaskSpec{
			TaskSpec: &task.TaskSpec{},
		},
	}

	store.On("WriteProtobuf", mock.Anything, storage.DataReference("s3://flyte-data/org/proj/dev/rctx-123/inputs/inputs.pb"), storage.Options{}, mock.MatchedBy(func(msg proto.Message) bool {
		inputs, ok := msg.(*task.Inputs)
		return ok &&
			len(inputs.Context) == 1 &&
			inputs.Context[0].GetKey() == "trace_id" &&
			inputs.Context[0].GetValue() == "root-abc"
	})).Return(nil).Once()

	actionRepo.On("CreateRun", mock.Anything, mock.MatchedBy(func(actual *workflow.CreateRunRequest) bool {
		return actual.GetRunSpec().GetRawDataStorage().GetRawDataPrefix() == "s3://custom-raw"
	}), "s3://flyte-data/org/proj/dev/rctx-123/inputs", "s3://flyte-data/org/proj/dev/rctx-123/").Return(&models.Run{
		Org:     "org",
		Project: "proj",
		Domain:  "dev",
		Name:    "rctx-123",
	}, nil).Once()

	actionsClient.On("Enqueue", mock.Anything, mock.MatchedBy(func(req *connect.Request[actions.EnqueueRequest]) bool {
		return req.Msg.GetRunSpec().GetRawDataStorage().GetRawDataPrefix() == "s3://custom-raw"
	})).Return(connect.NewResponse(&actions.EnqueueResponse{}), nil).Once()

	_, err := svc.CreateRun(context.Background(), connect.NewRequest(req))
	require.NoError(t, err)
}

func TestGetActionData_ReadsOutputFromAttempts(t *testing.T) {
	actionRepo := &repoMocks.ActionRepo{}
	actionsClient := &mockActionsClient{}
	repo := &repoMocks.Repository{}
	store := &storageMocks.ComposedProtobufStore{}
	dataStore := &storage.DataStore{ComposedProtobufStore: store}

	repo.On("ActionRepo").Return(actionRepo)

	svc := &RunService{
		repo:          repo,
		actionsClient: actionsClient,
		dataStore:     dataStore,
	}

	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     "test-org",
			Project: "test-project",
			Domain:  "test-domain",
			Name:    "rtest12345",
		},
		Name: "action-1",
	}

	// Build action spec with input URI
	actionSpec := &workflow.ActionSpec{
		InputUri: "s3://bucket/inputs/inputs.pb",
	}
	actionSpecBytes, _ := proto.Marshal(actionSpec)

	runInfo := &workflow.RunInfo{}
	runInfoBytes, _ := proto.Marshal(runInfo)

	actionModel := &models.Action{
		Org:          "test-org",
		Project:      "test-project",
		Domain:       "test-domain",
		RunName:      "rtest12345",
		Name:         "action-1",
		Phase:        int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED),
		ActionType:   int32(workflow.ActionType_ACTION_TYPE_TASK),
		ActionSpec:   actionSpecBytes,
		DetailedInfo: runInfoBytes,
		Attempts:     1,
	}

	// Build event with output URI
	successEvent := &workflow.ActionEvent{
		Id:          actionID,
		Attempt:     0,
		Phase:       common.ActionPhase_ACTION_PHASE_SUCCEEDED,
		Version:     1,
		UpdatedTime: timestamppb.Now(),
		Outputs: &task.OutputReferences{
			OutputUri: "s3://bucket/outputs/action-1/outputs.pb",
		},
	}
	eventModel, _ := models.NewActionEventModel(successEvent)

	actionRepo.On("GetAction", mock.Anything, actionID).Return(actionModel, nil)
	actionRepo.On("ListEvents", mock.Anything, matchActionID(actionID), 500).Return([]*models.ActionEvent{eventModel}, nil)

	// Mock reading inputs
	store.On("ReadProtobuf", mock.Anything, storage.DataReference("s3://bucket/inputs/inputs.pb"), mock.AnythingOfType("*task.Inputs")).
		Return(nil).Once()

	// Mock reading outputs — verify it reads from the attempt's output URI
	store.On("ReadProtobuf", mock.Anything, storage.DataReference("s3://bucket/outputs/action-1/outputs.pb"), mock.AnythingOfType("*core.LiteralMap")).
		Run(func(args mock.Arguments) {
			lm := args.Get(2).(*core.LiteralMap)
			lm.Literals = map[string]*core.Literal{
				"result": newStringLiteral("success"),
			}
		}).
		Return(nil).Once()

	resp, err := svc.GetActionData(context.Background(), connect.NewRequest(&workflow.GetActionDataRequest{
		ActionId: actionID,
	}))
	require.NoError(t, err)
	assert.Len(t, resp.Msg.Outputs.Literals, 1)
	assert.Equal(t, "result", resp.Msg.Outputs.Literals[0].Name)

	store.AssertExpectations(t)
}

func TestGetActionData_NonSucceededSkipsOutputs(t *testing.T) {
	actionRepo := &repoMocks.ActionRepo{}
	repo := &repoMocks.Repository{}
	store := &storageMocks.ComposedProtobufStore{}
	dataStore := &storage.DataStore{ComposedProtobufStore: store}

	repo.On("ActionRepo").Return(actionRepo)

	svc := &RunService{
		repo:      repo,
		dataStore: dataStore,
	}

	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     "test-org",
			Project: "test-project",
			Domain:  "test-domain",
			Name:    "rtest12345",
		},
		Name: "action-1",
	}

	actionSpec := &workflow.ActionSpec{
		InputUri: "s3://bucket/inputs/inputs.pb",
	}
	actionSpecBytes, _ := proto.Marshal(actionSpec)

	runInfo := &workflow.RunInfo{}
	runInfoBytes, _ := proto.Marshal(runInfo)

	actionModel := &models.Action{
		Org:          "test-org",
		Project:      "test-project",
		Domain:       "test-domain",
		RunName:      "rtest12345",
		Name:         "action-1",
		Phase:        int32(common.ActionPhase_ACTION_PHASE_RUNNING),
		ActionSpec:   actionSpecBytes,
		DetailedInfo: runInfoBytes,
	}

	actionRepo.On("GetAction", mock.Anything, actionID).Return(actionModel, nil)

	// Mock reading inputs
	store.On("ReadProtobuf", mock.Anything, storage.DataReference("s3://bucket/inputs/inputs.pb"), mock.AnythingOfType("*task.Inputs")).
		Return(nil).Once()

	resp, err := svc.GetActionData(context.Background(), connect.NewRequest(&workflow.GetActionDataRequest{
		ActionId: actionID,
	}))
	require.NoError(t, err)
	// Outputs should be empty since action is still running
	assert.Empty(t, resp.Msg.Outputs.Literals)

	store.AssertExpectations(t)
	// Verify ReadProtobuf was only called once (for inputs, not outputs)
	store.AssertNumberOfCalls(t, "ReadProtobuf", 1)
}

func TestConvertActionToEnrichedProto_IncludesDuration(t *testing.T) {
	svc := &RunService{}
	now := time.Now()
	startedAt := now.Add(-3 * time.Second)
	endedAt := now

	t.Run("includes duration, start time, and end time", func(t *testing.T) {
		action := &models.Action{
			Org:              "org",
			Project:          "proj",
			Domain:           "dev",
			RunName:          "run1",
			Name:             "action1",
			Phase:            int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED),
			StartedAt:        sql.NullTime{Time: startedAt, Valid: true},
			EndedAt:          sql.NullTime{Time: endedAt, Valid: true},
			DurationMs:       sql.NullInt64{Int64: 3000, Valid: true},
			Attempts:         1,
			CacheStatus:      core.CatalogCacheStatus_CACHE_DISABLED,
			ParentActionName: sql.NullString{String: "run1", Valid: true},
		}

		enriched := svc.convertActionToEnrichedProto(action)
		require.NotNil(t, enriched)
		require.NotNil(t, enriched.Action)
		require.NotNil(t, enriched.Action.Status)

		status := enriched.Action.Status
		assert.Equal(t, common.ActionPhase_ACTION_PHASE_SUCCEEDED, status.Phase)
		assert.Equal(t, startedAt.Unix(), status.StartTime.AsTime().Unix())
		assert.Equal(t, endedAt.Unix(), status.EndTime.AsTime().Unix())
		require.NotNil(t, status.DurationMs)
		assert.Equal(t, uint64(3000), *status.DurationMs)
		assert.Equal(t, uint32(1), status.Attempts)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_DISABLED, status.CacheStatus)
		assert.True(t, enriched.MeetsFilter)
	})

	t.Run("uses created_at when started_at is not set", func(t *testing.T) {
		createdAt := now.Add(-5 * time.Second)
		action := &models.Action{
			Org:       "org",
			Project:   "proj",
			Domain:    "dev",
			RunName:   "run1",
			Name:      "action2",
			Phase:     int32(common.ActionPhase_ACTION_PHASE_RUNNING),
			CreatedAt: createdAt,
		}

		enriched := svc.convertActionToEnrichedProto(action)
		require.NotNil(t, enriched)
		assert.Equal(t, createdAt.Unix(), enriched.Action.Status.StartTime.AsTime().Unix())
		assert.Nil(t, enriched.Action.Status.EndTime)
		assert.Nil(t, enriched.Action.Status.DurationMs)
	})

	t.Run("nil action returns nil", func(t *testing.T) {
		assert.Nil(t, svc.convertActionToEnrichedProto(nil))
	})

	t.Run("zero duration is not included", func(t *testing.T) {
		action := &models.Action{
			Org:        "org",
			Project:    "proj",
			Domain:     "dev",
			RunName:    "run1",
			Name:       "action3",
			Phase:      int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED),
			DurationMs: sql.NullInt64{Int64: 0, Valid: true},
		}

		enriched := svc.convertActionToEnrichedProto(action)
		require.NotNil(t, enriched)
		assert.Nil(t, enriched.Action.Status.DurationMs)
	})
}

func TestRecordEvents_AdvancesActionPhase(t *testing.T) {
	actionRepo, _, svc := newTestService(t)

	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     "org",
			Project: "proj",
			Domain:  "dev",
			Name:    "run1",
		},
		Name: "action1",
	}

	eventTime := timestamppb.Now()
	events := []*workflow.ActionEvent{
		{
			Id:          actionID,
			Attempt:     1,
			Phase:       common.ActionPhase_ACTION_PHASE_RUNNING,
			Version:     0,
			UpdatedTime: eventTime,
		},
	}

	actionRepo.On("InsertEvents", mock.Anything, mock.Anything).Return(nil).Once()
	actionRepo.On("UpdateActionPhase",
		mock.Anything,
		actionID,
		common.ActionPhase_ACTION_PHASE_RUNNING,
		uint32(1),
		core.CatalogCacheStatus_CACHE_DISABLED,
		mock.AnythingOfType("*time.Time"),
		(*time.Time)(nil), // non-terminal, no endTime
	).Return(nil).Once()

	err := svc.recordEvents(context.Background(), events)
	require.NoError(t, err)
}

func TestRecordEvents_AdvancesPhaseToTerminal(t *testing.T) {
	actionRepo, _, svc := newTestService(t)

	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     "org",
			Project: "proj",
			Domain:  "dev",
			Name:    "run1",
		},
		Name: "action1",
	}

	eventTime := timestamppb.Now()
	events := []*workflow.ActionEvent{
		{
			Id:          actionID,
			Attempt:     1,
			Phase:       common.ActionPhase_ACTION_PHASE_SUCCEEDED,
			Version:     0,
			UpdatedTime: eventTime,
		},
	}

	actionRepo.On("InsertEvents", mock.Anything, mock.Anything).Return(nil).Once()
	actionRepo.On("UpdateActionPhase",
		mock.Anything,
		actionID,
		common.ActionPhase_ACTION_PHASE_SUCCEEDED,
		uint32(1),
		core.CatalogCacheStatus_CACHE_DISABLED,
		mock.AnythingOfType("*time.Time"),
		mock.AnythingOfType("*time.Time"), // terminal → endTime is set
	).Return(nil).Once()

	err := svc.recordEvents(context.Background(), events)
	require.NoError(t, err)
}

func TestRecordEvents_SkipsQueuedPhase(t *testing.T) {
	actionRepo, _, svc := newTestService(t)

	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     "org",
			Project: "proj",
			Domain:  "dev",
			Name:    "run1",
		},
		Name: "action1",
	}

	events := []*workflow.ActionEvent{
		{
			Id:      actionID,
			Attempt: 1,
			Phase:   common.ActionPhase_ACTION_PHASE_QUEUED,
			Version: 0,
		},
	}

	actionRepo.On("InsertEvents", mock.Anything, mock.Anything).Return(nil).Once()
	// No UpdateActionPhase call expected for QUEUED phase

	err := svc.recordEvents(context.Background(), events)
	require.NoError(t, err)
}

func TestRecordEvents_PicksHighestPhasePerAction(t *testing.T) {
	actionRepo, _, svc := newTestService(t)

	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     "org",
			Project: "proj",
			Domain:  "dev",
			Name:    "run1",
		},
		Name: "action1",
	}

	eventTime := timestamppb.Now()
	events := []*workflow.ActionEvent{
		{
			Id:          actionID,
			Attempt:     1,
			Phase:       common.ActionPhase_ACTION_PHASE_RUNNING,
			Version:     0,
			UpdatedTime: eventTime,
		},
		{
			Id:          actionID,
			Attempt:     1,
			Phase:       common.ActionPhase_ACTION_PHASE_SUCCEEDED,
			Version:     0,
			UpdatedTime: eventTime,
		},
	}

	actionRepo.On("InsertEvents", mock.Anything, mock.Anything).Return(nil).Once()
	// Only the highest phase (SUCCEEDED) should trigger UpdateActionPhase
	actionRepo.On("UpdateActionPhase",
		mock.Anything,
		actionID,
		common.ActionPhase_ACTION_PHASE_SUCCEEDED,
		uint32(1),
		core.CatalogCacheStatus_CACHE_DISABLED,
		mock.AnythingOfType("*time.Time"),
		mock.AnythingOfType("*time.Time"),
	).Return(nil).Once()

	err := svc.recordEvents(context.Background(), events)
	require.NoError(t, err)
}
