package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

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
		lm, ok := msg.(*core.LiteralMap)
		return ok && len(lm.Literals) == 0
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
