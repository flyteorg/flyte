package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
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
	actionsconnectmocks "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions/actionsconnect/mocks"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/project"
	projectMocks "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/project/projectconnect/mocks"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	repoMocks "github.com/flyteorg/flyte/v2/runs/repository/mocks"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// newMockProjectClientAlwaysOK returns a mock ProjectServiceClient whose GetProject always succeeds.
func newMockProjectClientAlwaysOK(t *testing.T) *projectMocks.ProjectServiceClient {
	pc := projectMocks.NewProjectServiceClient(t)
	pc.EXPECT().GetProject(mock.Anything, mock.Anything).
		Return(connect.NewResponse(&project.GetProjectResponse{}), nil).Maybe()
	return pc
}

func newTestService(t *testing.T) (*repoMocks.ActionRepo, *actionsconnectmocks.ActionsServiceClient, *RunService) {
	actionRepo := &repoMocks.ActionRepo{}
	taskRepo := &repoMocks.TaskRepo{}
	actionsClient := actionsconnectmocks.NewActionsServiceClient(t)
	repo := &repoMocks.Repository{}
	repo.On("ActionRepo").Return(actionRepo)
	repo.On("TaskRepo").Maybe().Return(taskRepo)

	svc := &RunService{repo: repo, actionsClient: actionsClient}

	t.Cleanup(func() {
		repo.AssertExpectations(t)
		actionRepo.AssertExpectations(t)
		taskRepo.AssertExpectations(t)
	})

	return actionRepo, actionsClient, svc
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

func newRunServiceTestClient(t *testing.T, svc *RunService) workflowconnect.RunServiceClient {
	path, handler := workflowconnect.NewRunServiceHandler(svc)

	mux := http.NewServeMux()
	mux.Handle(path, handler)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return workflowconnect.NewRunServiceClient(http.DefaultClient, server.URL)
}

type mockDataProxyClient struct {
	mock.Mock
}

func (m *mockDataProxyClient) GetActionData(
	ctx context.Context,
	req *connect.Request[dataproxy.GetActionDataRequest],
) (*connect.Response[dataproxy.GetActionDataResponse], error) {
	args := m.Called(ctx, req)
	resp, _ := args.Get(0).(*connect.Response[dataproxy.GetActionDataResponse])
	return resp, args.Error(1)
}

func TestGetRunDetails_WithTaskSpec(t *testing.T) {
	actionRepo := &repoMocks.ActionRepo{}
	taskRepo := &repoMocks.TaskRepo{}
	actionsClient := actionsconnectmocks.NewActionsServiceClient(t)
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

func TestWatchClusterEvents_UsesPersistedClusterEvents(t *testing.T) {
	actionRepo, _, svc := newTestService(t)
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
	actionRepo, _, svc := newTestService(t)
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

func TestGetRunDetails_ReturnsRunSpecEnvVars(t *testing.T) {
	actionRepo := &repoMocks.ActionRepo{}
	taskRepo := &repoMocks.TaskRepo{}
	actionsClient := actionsconnectmocks.NewActionsServiceClient(t)
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

	runSpecBytes, err := proto.Marshal(&task.RunSpec{
		Envs: &task.Envs{
			Values: []*core.KeyValuePair{
				{Key: "foo", Value: "boo"},
				{Key: "LOG_LEVEL", Value: "30"},
			},
		},
	})
	require.NoError(t, err)

	runModel := &models.Run{
		Project:    runID.Project,
		Domain:     runID.Domain,
		RunName:    runID.Name,
		Name:       runID.Name,
		Phase:      int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED),
		ActionType: int32(workflow.ActionType_ACTION_TYPE_TASK),
		RunSpec:    runSpecBytes,
	}

	actionRepo.On("GetRun", mock.Anything, runID).Return(runModel, nil)
	actionRepo.On("ListEvents", mock.Anything, matchActionID(rootActionID), 500).Return([]*models.ActionEvent{}, nil)

	resp, err := svc.GetRunDetails(context.Background(), connect.NewRequest(&workflow.GetRunDetailsRequest{
		RunId: runID,
	}))
	require.NoError(t, err)
	require.NotNil(t, resp.Msg.Details)
	require.NotNil(t, resp.Msg.Details.GetRunSpec())

	gotEnv := map[string]string{}
	for _, kv := range resp.Msg.Details.GetRunSpec().GetEnvs().GetValues() {
		gotEnv[kv.GetKey()] = kv.GetValue()
	}
	assert.Equal(t, "boo", gotEnv["foo"])
	assert.Equal(t, "30", gotEnv["LOG_LEVEL"])
}

func TestGetRunDetails_UsesActionCacheStatus(t *testing.T) {
	actionRepo := &repoMocks.ActionRepo{}
	taskRepo := &repoMocks.TaskRepo{}
	actionsClient := actionsconnectmocks.NewActionsServiceClient(t)
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
	actionsClient := actionsconnectmocks.NewActionsServiceClient(t)
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

func TestFillDefaultInputsForCreateRun(t *testing.T) {
	inputs := &task.Inputs{
		Literals: []*task.NamedLiteral{
			{
				Name: "x",
				Value: &core.Literal{
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Primitive{
								Primitive: &core.Primitive{Value: &core.Primitive_Integer{Integer: 7}},
							},
						},
					},
				},
			},
		},
	}

	defaultInputs := []*task.NamedParameter{
		{
			Name: "x",
			Parameter: &core.Parameter{
				Behavior: &core.Parameter_Default{
					Default: &core.Literal{
						Value: &core.Literal_Scalar{
							Scalar: &core.Scalar{
								Value: &core.Scalar_Primitive{
									Primitive: &core.Primitive{Value: &core.Primitive_Integer{Integer: 42}},
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "y",
			Parameter: &core.Parameter{
				Behavior: &core.Parameter_Default{
					Default: &core.Literal{
						Value: &core.Literal_Scalar{
							Scalar: &core.Scalar{
								Value: &core.Scalar_Primitive{
									Primitive: &core.Primitive{Value: &core.Primitive_StringValue{StringValue: "default"}},
								},
							},
						},
					},
				},
			},
		},
	}

	gotInputs := fillDefaultInputs(inputs, defaultInputs)

	assert.Len(t, gotInputs.Literals, 2)
	got := make(map[string]*core.Literal, len(gotInputs.Literals))
	for _, nl := range gotInputs.Literals {
		got[nl.Name] = nl.Value
	}
	assert.Equal(t, int64(7), got["x"].GetScalar().GetPrimitive().GetInteger(), "provided input should not be overwritten")
	assert.Equal(t, "default", got["y"].GetScalar().GetPrimitive().GetStringValue(), "missing input should be filled from default")
}

func TestCreateRunResponseIncludesMetadataAndStatus(t *testing.T) {
	actionRepo := &repoMocks.ActionRepo{}
	taskRepo := &repoMocks.TaskRepo{}
	actionsClient := actionsconnectmocks.NewActionsServiceClient(t)
	repo := &repoMocks.Repository{}
	store := &storageMocks.ComposedProtobufStore{}
	dataStore := &storage.DataStore{ComposedProtobufStore: store}

	repo.On("ActionRepo").Return(actionRepo)
	repo.On("TaskRepo").Maybe().Return(taskRepo)

	svc := &RunService{
		repo:          repo,
		actionsClient: actionsClient,
		projectClient: newMockProjectClientAlwaysOK(t),
		storagePrefix: "s3://flyte-data",
		dataStore:     dataStore,
	}

	runID := &common.RunIdentifier{
		Org:     "test-org",
		Project: "test-project",
		Domain:  "test-domain",
		Name:    "rtest12345",
	}
	createdAt := time.Now().UTC().Truncate(time.Second)

	store.On("WriteProtobuf", mock.Anything, mock.Anything, storage.Options{}, mock.Anything).Return(nil).Once()

	taskRepo.On("CreateTaskSpec", mock.Anything, mock.Anything).Return(nil).Once()
	actionRepo.On("CreateAction", mock.Anything, mock.Anything, mock.Anything).
		Return(&models.Run{
			Project:     runID.Project,
			Domain:      runID.Domain,
			Name:        runID.Name,
			Phase:       int32(common.ActionPhase_ACTION_PHASE_QUEUED),
			CreatedAt:   createdAt,
			Attempts:    1,
			CacheStatus: core.CatalogCacheStatus_CACHE_DISABLED,
		}, nil).Once()

	actionsClient.On("Enqueue", mock.Anything, mock.Anything).
		Return(connect.NewResponse(&actions.EnqueueResponse{}), nil).Once()

	resp, err := svc.CreateRun(context.Background(), connect.NewRequest(&workflow.CreateRunRequest{
		Id: &workflow.CreateRunRequest_RunId{
			RunId: runID,
		},
		Task: &workflow.CreateRunRequest_TaskSpec{
			TaskSpec: &task.TaskSpec{},
		},
	}))
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Msg.GetRun())
	assert.NotNil(t, resp.Msg.GetRun().GetAction())
	assert.NotNil(t, resp.Msg.GetRun().GetAction().GetId())
	assert.Equal(t, runID.Name, resp.Msg.GetRun().GetAction().GetId().GetName())
	assert.NotNil(t, resp.Msg.GetRun().GetAction().GetMetadata())

	status := resp.Msg.GetRun().GetAction().GetStatus()
	assert.NotNil(t, status)
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_QUEUED, status.GetPhase())
	assert.NotNil(t, status.GetStartTime())
	assert.True(t, status.GetStartTime().AsTime().Equal(createdAt))
	assert.Equal(t, uint32(1), status.GetAttempts())
	assert.Equal(t, core.CatalogCacheStatus_CACHE_DISABLED, status.GetCacheStatus())
	assert.Nil(t, status.EndTime)
	assert.Nil(t, status.DurationMs)

	repo.AssertExpectations(t)
	actionRepo.AssertExpectations(t)
	taskRepo.AssertExpectations(t)
	actionsClient.AssertExpectations(t)
	store.AssertExpectations(t)
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

		_, err := svc.AbortRun(context.Background(), connect.NewRequest(&workflow.AbortRunRequest{RunId: runID}))
		assert.NoError(t, err)
		actionsClient.AssertNotCalled(t, "Abort")
	})

	t.Run("success with custom reason", func(t *testing.T) {
		actionRepo, actionsClient, svc := newTestService(t)
		reason := "timeout exceeded"

		actionRepo.On("AbortRun", mock.Anything, runID, reason, (*common.EnrichedIdentity)(nil)).Return(nil)

		_, err := svc.AbortRun(context.Background(), connect.NewRequest(&workflow.AbortRunRequest{RunId: runID, Reason: &reason}))
		assert.NoError(t, err)
		actionsClient.AssertNotCalled(t, "Abort")
	})

	t.Run("db error returns error", func(t *testing.T) {
		actionRepo, actionsClient, svc := newTestService(t)

		actionRepo.On("AbortRun", mock.Anything, runID, mock.Anything, mock.Anything).Return(errors.New("db unavailable"))

		_, err := svc.AbortRun(context.Background(), connect.NewRequest(&workflow.AbortRunRequest{RunId: runID}))
		assert.Error(t, err)
		actionsClient.AssertNotCalled(t, "Abort")
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

		_, err := svc.AbortAction(context.Background(), connect.NewRequest(&workflow.AbortActionRequest{ActionId: actionID}))
		assert.NoError(t, err)
		actionsClient.AssertNotCalled(t, "Abort")
	})

	t.Run("success with custom reason", func(t *testing.T) {
		actionRepo, actionsClient, svc := newTestService(t)
		reason := "resource limit exceeded"

		actionRepo.On("AbortAction", mock.Anything, actionID, reason, (*common.EnrichedIdentity)(nil)).Return(nil)

		_, err := svc.AbortAction(context.Background(), connect.NewRequest(&workflow.AbortActionRequest{ActionId: actionID, Reason: reason}))
		assert.NoError(t, err)
		actionsClient.AssertNotCalled(t, "Abort")
	})

	t.Run("db error returns error", func(t *testing.T) {
		actionRepo, actionsClient, svc := newTestService(t)

		actionRepo.On("AbortAction", mock.Anything, actionID, mock.Anything, mock.Anything).Return(errors.New("db unavailable"))

		_, err := svc.AbortAction(context.Background(), connect.NewRequest(&workflow.AbortActionRequest{ActionId: actionID}))
		assert.Error(t, err)
		actionsClient.AssertNotCalled(t, "Abort")
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
						Project: "test-project",
						Domain:  "test-domain",
						Name:    fmt.Sprintf("run-%d", i),
					},
					Name: "a0",
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
			Project:    "test-project",
			Domain:     "test-domain",
			RunName:    fmt.Sprintf("run-%d", i),
			Name:       "a0",
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
		runs []*models.Run
		err  error
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
			// Service fetches Limit+1 rows to detect another page. With 3 rows
			// returned for a limit of 2, the slice is trimmed to the first 2
			// runs and the cursor token is the trimmed last row's created_at.
			"list with limit 2 and token",
			&common.ListRequest{Limit: 2, Token: sqlRes[5].CreatedAt.UTC().Format(time.RFC3339Nano)},
			mockListRes{runs: sqlRes[5:8], err: nil},
			&workflow.ListRunsResponse{Runs: runs[5:7], Token: sqlRes[6].CreatedAt.UTC().Format(time.RFC3339Nano)},
		},
		{
			// Only 2 rows returned for limit 3 means no next page — token empty.
			"list with limit 3 and token",
			&common.ListRequest{Limit: 3, Token: sqlRes[8].CreatedAt.UTC().Format(time.RFC3339Nano)},
			mockListRes{runs: sqlRes[8:10], err: nil},
			&workflow.ListRunsResponse{Runs: runs[8:10], Token: ""},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := connect.NewRequest(&workflow.ListRunsRequest{Request: tc.req})
			actionRepo.On("ListActions", mock.Anything, mock.Anything).Return(tc.mockRes.runs, tc.mockRes.err).Once()
			got, err := svc.ListRuns(context.Background(), req)
			assert.NoError(t, err)
			assert.Equal(t, len(tc.expect.Runs), len(got.Msg.Runs))
			assert.Equal(t, tc.expect.Runs, got.Msg.Runs)
			assert.Equal(t, tc.expect.Token, got.Msg.Token)
		})
	}
}

// TestListAndSendAllActionsUsesAscendingSort guards against regressing the
// default repo sort order leaking into the WatchActions seed path. Children
// have a later created_at than their parents; if ListActions returns rows in
// descending order, the run state manager's insertAction fails because the
// parent node is not yet in the tree. This test asserts we always pass an
// ascending created_at sort parameter.
func TestListAndSendAllActionsUsesAscendingSort(t *testing.T) {
	actionRepo, _, svc := newTestService(t)

	runID := &common.RunIdentifier{Project: "p", Domain: "d", Name: "run-1"}

	var captured interfaces.ListResourceInput
	actionRepo.On("ListActions", mock.Anything, mock.MatchedBy(func(input interfaces.ListResourceInput) bool {
		captured = input
		return true
	})).Return([]*models.Action{}, nil).Once()

	rsm, err := newRunStateManager(nil)
	require.NoError(t, err)

	err = svc.listAndSendAllActions(context.Background(), runID, rsm, nil)
	require.NoError(t, err)

	require.Len(t, captured.SortParameters, 1)
	assert.Equal(t, "created_at ASC", captured.SortParameters[0].GetOrderExpr())
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
		{
			"valid run",
			&models.Run{
				Project:       project,
				Domain:        domain,
				RunName:       name,
				Name:          "a0",
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
							Project: project,
							Domain:  domain,
							Name:    name,
						},
						Name: "a0",
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
				Project:    project,
				Domain:     domain,
				RunName:    name,
				Name:       "a0",
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
							Project: project,
							Domain:  domain,
							Name:    name,
						},
						Name: "a0",
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
	taskRepo := &repoMocks.TaskRepo{}
	actionsClient := actionsconnectmocks.NewActionsServiceClient(t)
	repo := &repoMocks.Repository{}
	store := &storageMocks.ComposedProtobufStore{}
	dataStore := &storage.DataStore{ComposedProtobufStore: store}

	repo.On("ActionRepo").Return(actionRepo)
	repo.On("TaskRepo").Return(taskRepo)

	svc := &RunService{
		repo:          repo,
		actionsClient: actionsClient,
		projectClient: newMockProjectClientAlwaysOK(t),
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
		Project: "testproject",
		Domain:  "development",
		Name:    "generated-run",
	}

	store.On("WriteProtobuf", mock.Anything, mock.AnythingOfType("storage.DataReference"), storage.Options{}, mock.MatchedBy(func(msg proto.Message) bool {
		inputs, ok := msg.(*task.Inputs)
		return ok && len(inputs.Literals) == 0 && len(inputs.Context) == 0
	})).Return(nil).Once()

	taskRepo.On("CreateTaskSpec", mock.Anything, mock.Anything).Return(nil).Once()
	actionRepo.On("CreateAction", mock.Anything, mock.Anything, mock.Anything).Return(expectedRun, nil).Once()

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
	taskRepo := &repoMocks.TaskRepo{}
	actionsClient := actionsconnectmocks.NewActionsServiceClient(t)
	repo := &repoMocks.Repository{}
	store := &storageMocks.ComposedProtobufStore{}
	dataStore := &storage.DataStore{ComposedProtobufStore: store}

	repo.On("ActionRepo").Return(actionRepo)
	repo.On("TaskRepo").Return(taskRepo)

	svc := &RunService{
		repo:          repo,
		actionsClient: actionsClient,
		projectClient: newMockProjectClientAlwaysOK(t),
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
		Project: "myproj",
		Domain:  "production",
		RunName: "rtest99999",
		Name:    "a0",
		Phase:   int32(common.ActionPhase_ACTION_PHASE_QUEUED),
	}

	store.On("WriteProtobuf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	taskRepo.On("CreateTaskSpec", mock.Anything, mock.Anything).Return(nil).Once()
	actionRepo.On("CreateAction", mock.Anything, mock.Anything, mock.Anything).Return(expectedRun, nil).Once()
	actionsClient.On("Enqueue", mock.Anything, mock.Anything).Return(connect.NewResponse(&actions.EnqueueResponse{}), nil).Once()

	resp, err := svc.CreateRun(context.Background(), connect.NewRequest(req))
	require.NoError(t, err)

	run := resp.Msg.Run
	assert.Equal(t, "", run.Action.Id.Run.Org)
	assert.Equal(t, "myproj", run.Action.Id.Run.Project)
	assert.Equal(t, "production", run.Action.Id.Run.Domain)
	assert.Equal(t, "rtest99999", run.Action.Id.Run.Name)
	assert.Equal(t, "a0", run.Action.Id.Name)
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_QUEUED, run.Action.Status.Phase)
}

func TestCreateRun_ActionIDUsesRunName(t *testing.T) {
	actionRepo := &repoMocks.ActionRepo{}
	taskRepo := &repoMocks.TaskRepo{}
	actionsClient := actionsconnectmocks.NewActionsServiceClient(t)
	repo := &repoMocks.Repository{}
	store := &storageMocks.ComposedProtobufStore{}
	dataStore := &storage.DataStore{ComposedProtobufStore: store}

	repo.On("ActionRepo").Return(actionRepo)
	repo.On("TaskRepo").Return(taskRepo)

	svc := &RunService{
		repo:          repo,
		actionsClient: actionsClient,
		projectClient: newMockProjectClientAlwaysOK(t),
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
	taskRepo.On("CreateTaskSpec", mock.Anything, mock.Anything).Return(nil).Once()
	actionRepo.On("CreateAction", mock.Anything, mock.Anything, mock.Anything).
		Return(&models.Run{Project: "proj", Domain: "dev", RunName: "rmy-run-123", Name: "a0"}, nil).Once()

	// Verify the enqueue request uses "a0" as the action name and the run name in the run identifier
	actionsClient.On("Enqueue", mock.Anything, mock.MatchedBy(func(req *connect.Request[actions.EnqueueRequest]) bool {
		return req.Msg.Action.ActionId.Name == "a0" &&
			req.Msg.Action.ActionId.Run.Name == "rmy-run-123"
	})).Return(connect.NewResponse(&actions.EnqueueResponse{}), nil).Once()

	_, err := svc.CreateRun(context.Background(), connect.NewRequest(req))
	assert.NoError(t, err)

	actionsClient.AssertExpectations(t)
}

func TestCreateRun_PreservesInputContextAndRawDataPath(t *testing.T) {
	actionRepo := &repoMocks.ActionRepo{}
	taskRepo := &repoMocks.TaskRepo{}
	actionsClient := actionsconnectmocks.NewActionsServiceClient(t)
	repo := &repoMocks.Repository{}
	store := &storageMocks.ComposedProtobufStore{}
	dataStore := &storage.DataStore{ComposedProtobufStore: store}

	repo.On("ActionRepo").Return(actionRepo)
	repo.On("TaskRepo").Return(taskRepo)

	svc := &RunService{
		repo:          repo,
		actionsClient: actionsClient,
		projectClient: newMockProjectClientAlwaysOK(t),
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
		InputWrapper: &workflow.CreateRunRequest_Inputs{
			&task.Inputs{
				Context: []*core.KeyValuePair{
					{Key: "trace_id", Value: "root-abc"},
				},
			},
		},
		RunSpec: &task.RunSpec{
			RawDataStorage: &task.RawDataStorage{RawDataPrefix: "s3://custom-raw"},
		},
		Task: &workflow.CreateRunRequest_TaskSpec{
			TaskSpec: &task.TaskSpec{},
		},
	}

	store.On("WriteProtobuf", mock.Anything, storage.DataReference("s3://flyte-data/proj/dev/rctx-123/inputs/inputs.pb"), storage.Options{}, mock.MatchedBy(func(msg proto.Message) bool {
		inputs, ok := msg.(*task.Inputs)
		return ok &&
			len(inputs.Context) == 1 &&
			inputs.Context[0].GetKey() == "trace_id" &&
			inputs.Context[0].GetValue() == "root-abc"
	})).Return(nil).Once()

	taskRepo.On("CreateTaskSpec", mock.Anything, mock.Anything).Return(nil).Once()
	actionRepo.On("CreateAction", mock.Anything, mock.MatchedBy(func(m *models.Run) bool {
		var rs task.RunSpec
		_ = proto.Unmarshal(m.RunSpec, &rs)
		return rs.GetRawDataStorage().GetRawDataPrefix() == "s3://custom-raw"
	}), mock.Anything).Return(&models.Run{
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

func TestGetActionData_DelegatesToDataProxy(t *testing.T) {
	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     "test-org",
			Project: "test-project",
			Domain:  "test-domain",
			Name:    "rtest12345",
		},
		Name: "action-1",
	}

	dpClient := &mockDataProxyClient{}
	svc := &RunService{dataProxyClient: dpClient}
	dpClient.On("GetActionData", mock.Anything, mock.MatchedBy(func(req *connect.Request[dataproxy.GetActionDataRequest]) bool {
		return proto.Equal(req.Msg.GetActionId(), actionID)
	})).Return(connect.NewResponse(&dataproxy.GetActionDataResponse{
		Inputs: &task.Inputs{
			Literals: []*task.NamedLiteral{
				{Name: "x", Value: newStringLiteral("input")},
			},
		},
		Outputs: &task.Outputs{
			Literals: []*task.NamedLiteral{
				{Name: "result", Value: newStringLiteral("success")},
			},
		},
	}), nil).Once()

	resp, err := svc.GetActionData(context.Background(), connect.NewRequest(&workflow.GetActionDataRequest{
		ActionId: actionID,
	}))
	require.NoError(t, err)
	assert.Len(t, resp.Msg.Inputs.Literals, 1)
	assert.Len(t, resp.Msg.Outputs.Literals, 1)
	assert.Equal(t, "x", resp.Msg.Inputs.Literals[0].Name)
	assert.Equal(t, "result", resp.Msg.Outputs.Literals[0].Name)
	dpClient.AssertExpectations(t)
}

func TestGetActionData_PropagatesDataProxyError(t *testing.T) {
	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     "test-org",
			Project: "test-project",
			Domain:  "test-domain",
			Name:    "rtest12345",
		},
		Name: "action-1",
	}

	dpClient := &mockDataProxyClient{}
	svc := &RunService{dataProxyClient: dpClient}
	dpClient.On("GetActionData", mock.Anything, mock.Anything).Return(
		nil, connect.NewError(connect.CodeNotFound, errors.New("action not found")),
	).Once()

	resp, err := svc.GetActionData(context.Background(), connect.NewRequest(
		&workflow.GetActionDataRequest{
		ActionId: actionID,
	}))
	assert.Nil(t, resp)
	require.Error(t, err)
	assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
	dpClient.AssertExpectations(t)
}

func TestCreateRun_WithOffloadedInputData(t *testing.T) {
	actionRepo := &repoMocks.ActionRepo{}
	taskRepo := &repoMocks.TaskRepo{}
	actionsClient := actionsconnectmocks.NewActionsServiceClient(t)
	repo := &repoMocks.Repository{}
	store := &storageMocks.ComposedProtobufStore{}
	dataStore := &storage.DataStore{ComposedProtobufStore: store}

	repo.On("ActionRepo").Return(actionRepo)
	repo.On("TaskRepo").Return(taskRepo)

	svc := &RunService{
		repo:          repo,
		actionsClient: actionsClient,
		projectClient: newMockProjectClientAlwaysOK(t),
		storagePrefix: "s3://flyte-data",
		dataStore:     dataStore,
	}

	offloadedURI := "s3://test-bucket/uploads/testorg/testproject/development/offloaded-inputs/somehash"

	req := &workflow.CreateRunRequest{
		Id: &workflow.CreateRunRequest_RunId{
			RunId: &common.RunIdentifier{
				Org:     "testorg",
				Project: "testproject",
				Domain:  "development",
				Name:    "rtest-offloaded",
			},
		},
		Task: &workflow.CreateRunRequest_TaskSpec{
			TaskSpec: &task.TaskSpec{},
		},
		InputWrapper: &workflow.CreateRunRequest_OffloadedInputData{
			OffloadedInputData: &common.OffloadedInputData{
				Uri:        offloadedURI,
				InputsHash: "testhash123",
			},
		},
	}

	expectedRun := &models.Run{
		Project: "testproject",
		Domain:  "development",
		RunName: "rtest-offloaded",
		Name:    "a0",
		Phase:   int32(common.ActionPhase_ACTION_PHASE_QUEUED),
	}

	// No ReadProtobuf or WriteProtobuf expected — offloaded inputs are used directly.
	taskRepo.On("CreateTaskSpec", mock.Anything, mock.Anything).Return(nil).Once()
	actionRepo.On("CreateAction", mock.Anything, mock.MatchedBy(func(m *models.Run) bool {
		var info workflow.RunInfo
		_ = proto.Unmarshal(m.DetailedInfo, &info)
		return info.InputsUri == offloadedURI
	}), mock.Anything).Return(expectedRun, nil).Once()
	actionsClient.On("Enqueue", mock.MatchedBy(func(_ context.Context) bool { return true }), mock.MatchedBy(func(req *connect.Request[actions.EnqueueRequest]) bool {
		return req.Msg.Action.InputUri == offloadedURI
	})).Return(connect.NewResponse(&actions.EnqueueResponse{}), nil).Once()

	resp, err := svc.CreateRun(context.Background(), connect.NewRequest(req))
	require.NoError(t, err)

	run := resp.Msg.Run
	assert.Equal(t, "", run.Action.Id.Run.Org)
	assert.Equal(t, "testproject", run.Action.Id.Run.Project)
	assert.Equal(t, "rtest-offloaded", run.Action.Id.Run.Name)

	store.AssertExpectations(t)
	store.AssertNumberOfCalls(t, "ReadProtobuf", 0)
	store.AssertNumberOfCalls(t, "WriteProtobuf", 0)
}

func TestGetActionDataURIs(t *testing.T) {
	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org: "test-org", Project: "test-project", Domain: "test-domain", Name: "rtest12345",
		},
		Name: "a0",
	}

	mustMarshalRunInfo := func(info *workflow.RunInfo) []byte {
		b, err := proto.Marshal(info)
		require.NoError(t, err)
		return b
	}

	mustMarshalEvent := func(event *workflow.ActionEvent) []byte {
		b, err := proto.Marshal(event)
		require.NoError(t, err)
		return b
	}

	t.Run("action not found returns NotFound", func(t *testing.T) {
		actionRepo, _, svc := newTestService(t)
		actionRepo.On("GetAction", mock.Anything, matchActionID(actionID)).Return(nil, errors.New("missing"))

		resp, err := svc.GetActionDataURIs(context.Background(), connect.NewRequest(&workflow.GetActionDataURIsRequest{
			ActionId: actionID,
		}))

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
	})

	t.Run("missing detailed_info returns NotFound", func(t *testing.T) {
		actionRepo, _, svc := newTestService(t)
		actionRepo.On("GetAction", mock.Anything, matchActionID(actionID)).Return(&models.Action{
			Phase:        int32(common.ActionPhase_ACTION_PHASE_RUNNING),
			DetailedInfo: nil,
		}, nil)

		resp, err := svc.GetActionDataURIs(context.Background(), connect.NewRequest(&workflow.GetActionDataURIsRequest{
			ActionId: actionID,
		}))

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
	})

	t.Run("non-terminal phase returns inputs only", func(t *testing.T) {
		actionRepo, _, svc := newTestService(t)
		actionRepo.On("GetAction", mock.Anything, matchActionID(actionID)).Return(&models.Action{
			Phase:      int32(common.ActionPhase_ACTION_PHASE_RUNNING),
			ActionType: int32(workflow.ActionType_ACTION_TYPE_TASK),
			DetailedInfo: mustMarshalRunInfo(&workflow.RunInfo{
				InputsUri:  "s3://bucket/inputs",
				OutputsUri: "s3://bucket/outputs",
			}),
		}, nil)

		resp, err := svc.GetActionDataURIs(context.Background(), connect.NewRequest(&workflow.GetActionDataURIsRequest{
			ActionId: actionID,
		}))

		require.NoError(t, err)
		assert.Equal(t, "s3://bucket/inputs", resp.Msg.GetInputsUri())
		assert.Empty(t, resp.Msg.GetOutputsUri())
	})

	t.Run("trace action returns outputs uri from RunInfo", func(t *testing.T) {
		actionRepo, _, svc := newTestService(t)
		actionRepo.On("GetAction", mock.Anything, matchActionID(actionID)).Return(&models.Action{
			Phase:      int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED),
			ActionType: int32(workflow.ActionType_ACTION_TYPE_TRACE),
			DetailedInfo: mustMarshalRunInfo(&workflow.RunInfo{
				InputsUri:  "s3://bucket/inputs",
				OutputsUri: "s3://bucket/trace-outputs",
			}),
		}, nil)

		resp, err := svc.GetActionDataURIs(context.Background(), connect.NewRequest(&workflow.GetActionDataURIsRequest{
			ActionId: actionID,
		}))

		require.NoError(t, err)
		assert.Equal(t, "s3://bucket/inputs", resp.Msg.GetInputsUri())
		assert.Equal(t, "s3://bucket/trace-outputs", resp.Msg.GetOutputsUri())
	})

	t.Run("succeeded task returns outputs from latest attempt", func(t *testing.T) {
		actionRepo, _, svc := newTestService(t)
		actionRepo.On("GetAction", mock.Anything, matchActionID(actionID)).Return(&models.Action{
			Phase:      int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED),
			ActionType: int32(workflow.ActionType_ACTION_TYPE_TASK),
			DetailedInfo: mustMarshalRunInfo(&workflow.RunInfo{
				InputsUri: "s3://bucket/inputs",
			}),
		}, nil)

		eventBytes := mustMarshalEvent(&workflow.ActionEvent{
			Id:      actionID,
			Attempt: 0,
			Phase:   common.ActionPhase_ACTION_PHASE_SUCCEEDED,
			Outputs: &task.OutputReferences{
				OutputUri: "s3://bucket/outputs/0/outputs.pb",
			},
		})
		actionRepo.On("ListEvents", mock.Anything, matchActionID(actionID), 500).Return([]*models.ActionEvent{
			{Attempt: 0, Phase: int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED), Info: eventBytes},
		}, nil)

		resp, err := svc.GetActionDataURIs(context.Background(), connect.NewRequest(&workflow.GetActionDataURIsRequest{
			ActionId: actionID,
		}))

		require.NoError(t, err)
		assert.Equal(t, "s3://bucket/inputs", resp.Msg.GetInputsUri())
		assert.Equal(t, "s3://bucket/outputs/0/outputs.pb", resp.Msg.GetOutputsUri())
	})

	t.Run("succeeded task with no attempts returns NotFound", func(t *testing.T) {
		actionRepo, _, svc := newTestService(t)
		actionRepo.On("GetAction", mock.Anything, matchActionID(actionID)).Return(&models.Action{
			Phase:      int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED),
			ActionType: int32(workflow.ActionType_ACTION_TYPE_TASK),
			DetailedInfo: mustMarshalRunInfo(&workflow.RunInfo{
				InputsUri: "s3://bucket/inputs",
			}),
		}, nil)
		actionRepo.On("ListEvents", mock.Anything, matchActionID(actionID), 500).Return([]*models.ActionEvent{}, nil)

		resp, err := svc.GetActionDataURIs(context.Background(), connect.NewRequest(&workflow.GetActionDataURIsRequest{
			ActionId: actionID,
		}))

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
	})

	t.Run("succeeded task with empty output_uri returns NotFound", func(t *testing.T) {
		actionRepo, _, svc := newTestService(t)
		actionRepo.On("GetAction", mock.Anything, matchActionID(actionID)).Return(&models.Action{
			Phase:      int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED),
			ActionType: int32(workflow.ActionType_ACTION_TYPE_TASK),
			DetailedInfo: mustMarshalRunInfo(&workflow.RunInfo{
				InputsUri: "s3://bucket/inputs",
			}),
		}, nil)

		eventBytes := mustMarshalEvent(&workflow.ActionEvent{
			Id:      actionID,
			Attempt: 0,
			Phase:   common.ActionPhase_ACTION_PHASE_SUCCEEDED,
		})
		actionRepo.On("ListEvents", mock.Anything, matchActionID(actionID), 500).Return([]*models.ActionEvent{
			{Attempt: 0, Phase: int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED), Info: eventBytes},
		}, nil)

		resp, err := svc.GetActionDataURIs(context.Background(), connect.NewRequest(&workflow.GetActionDataURIsRequest{
			ActionId: actionID,
		}))

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
	})

	t.Run("succeeded task uses last attempt outputs when multiple attempts", func(t *testing.T) {
		actionRepo, _, svc := newTestService(t)
		actionRepo.On("GetAction", mock.Anything, matchActionID(actionID)).Return(&models.Action{
			Phase:        int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED),
			ActionType:   int32(workflow.ActionType_ACTION_TYPE_TASK),
			DetailedInfo: mustMarshalRunInfo(&workflow.RunInfo{InputsUri: "s3://bucket/inputs"}),
		}, nil)

		now := time.Now()
		event0 := mustMarshalEvent(&workflow.ActionEvent{
			Id: actionID, Attempt: 0,
			Phase:        common.ActionPhase_ACTION_PHASE_FAILED,
			ReportedTime: timestamppb.New(now),
			Outputs:      &task.OutputReferences{OutputUri: "s3://bucket/outputs/0/outputs.pb"},
		})
		event1 := mustMarshalEvent(&workflow.ActionEvent{
			Id: actionID, Attempt: 1,
			Phase:        common.ActionPhase_ACTION_PHASE_SUCCEEDED,
			ReportedTime: timestamppb.New(now.Add(time.Second)),
			Outputs:      &task.OutputReferences{OutputUri: "s3://bucket/outputs/1/outputs.pb"},
		})
		actionRepo.On("ListEvents", mock.Anything, matchActionID(actionID), 500).Return([]*models.ActionEvent{
			{Attempt: 0, Info: event0},
			{Attempt: 1, Info: event1},
		}, nil)

		resp, err := svc.GetActionDataURIs(context.Background(), connect.NewRequest(&workflow.GetActionDataURIsRequest{
			ActionId: actionID,
		}))

		require.NoError(t, err)
		assert.Equal(t, "s3://bucket/outputs/1/outputs.pb", resp.Msg.GetOutputsUri())
	})
}

func TestGetActionLogContext(t *testing.T) {
	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org: "test-org", Project: "test-project", Domain: "test-domain", Name: "rtest12345",
		},
		Name: "a0",
	}

	logContext := &core.LogContext{
		PrimaryPodName: "my-pod",
		Pods: []*core.PodLogContext{
			{PodName: "my-pod", Namespace: "ns"},
		},
	}

	mustMarshalEvent := func(event *workflow.ActionEvent) []byte {
		b, err := proto.Marshal(event)
		require.NoError(t, err)
		return b
	}

	t.Run("success returns log context and cluster", func(t *testing.T) {
		actionRepo, _, svc := newTestService(t)
		actionRepo.On("GetLatestEventByAttempt", mock.Anything, matchActionID(actionID), uint32(1)).Return(&models.ActionEvent{
			Info: mustMarshalEvent(&workflow.ActionEvent{
				Id:         actionID,
				Attempt:    1,
				LogContext: logContext,
				Cluster:    "c1",
			}),
		}, nil)

		resp, err := svc.GetActionLogContext(context.Background(), connect.NewRequest(&workflow.GetActionLogContextRequest{
			ActionId: actionID,
			Attempt:  1,
		}))

		require.NoError(t, err)
		assert.Equal(t, "c1", resp.Msg.GetCluster())
		assert.Equal(t, "my-pod", resp.Msg.GetLogContext().GetPrimaryPodName())
	})

	t.Run("no event found returns NotFound", func(t *testing.T) {
		actionRepo, _, svc := newTestService(t)
		actionRepo.On("GetLatestEventByAttempt", mock.Anything, matchActionID(actionID), uint32(1)).Return(nil, sql.ErrNoRows)

		resp, err := svc.GetActionLogContext(context.Background(), connect.NewRequest(&workflow.GetActionLogContextRequest{
			ActionId: actionID,
			Attempt:  1,
		}))

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
	})

	t.Run("repo error returns Internal", func(t *testing.T) {
		actionRepo, _, svc := newTestService(t)
		actionRepo.On("GetLatestEventByAttempt", mock.Anything, matchActionID(actionID), uint32(1)).Return(nil, errors.New("db blew up"))

		resp, err := svc.GetActionLogContext(context.Background(), connect.NewRequest(&workflow.GetActionLogContextRequest{
			ActionId: actionID,
			Attempt:  1,
		}))

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, connect.CodeInternal, connect.CodeOf(err))
	})

	t.Run("event without log context returns NotFound", func(t *testing.T) {
		actionRepo, _, svc := newTestService(t)
		actionRepo.On("GetLatestEventByAttempt", mock.Anything, matchActionID(actionID), uint32(1)).Return(&models.ActionEvent{
			Info: mustMarshalEvent(&workflow.ActionEvent{
				Id:      actionID,
				Attempt: 1,
				Cluster: "c1",
			}),
		}, nil)

		resp, err := svc.GetActionLogContext(context.Background(), connect.NewRequest(&workflow.GetActionLogContextRequest{
			ActionId: actionID,
			Attempt:  1,
		}))

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
	})

	t.Run("undeserializable event returns Internal", func(t *testing.T) {
		actionRepo, _, svc := newTestService(t)
		actionRepo.On("GetLatestEventByAttempt", mock.Anything, matchActionID(actionID), uint32(1)).Return(&models.ActionEvent{
			Info: []byte("not-a-proto"),
		}, nil)

		resp, err := svc.GetActionLogContext(context.Background(), connect.NewRequest(&workflow.GetActionLogContextRequest{
			ActionId: actionID,
			Attempt:  1,
		}))

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, connect.CodeInternal, connect.CodeOf(err))
	})
}
