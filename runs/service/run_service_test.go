package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/settings"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	storageMocks "github.com/flyteorg/flyte/v2/flytestdlib/storage/mocks"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions"
	actionsconnectmocks "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions/actionsconnect/mocks"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/project"
	projectMocks "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/project/projectconnect/mocks"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	triggerpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/trigger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	repoMocks "github.com/flyteorg/flyte/v2/runs/repository/mocks"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// noSettings returns a settings repo reporting no stored rows, so run creation
// resolves to empty settings and the applier changes nothing.
func noSettings(t *testing.T) *repoMocks.SettingsRepo {
	m := &repoMocks.SettingsRepo{}
	m.On("GetSettingsByKeys", mock.Anything, mock.Anything).Return(nil, nil)
	return m
}

// settingsRepoWith returns a settings repo holding one org-level row with the given
// settings. The row key must match what fetchLevels asks for, or the lookup aligns
// to nothing.
func settingsRepoWith(t *testing.T, s *settings.Settings) *repoMocks.SettingsRepo {
	t.Helper()
	data, err := protojson.Marshal(s)
	require.NoError(t, err)

	m := &repoMocks.SettingsRepo{}
	m.On("GetSettingsByKeys", mock.Anything, mock.Anything).
		Return([]*models.Settings{{Key: models.EncodeSettingsKey("", ""), Data: data, Version: 1}}, nil)
	return m
}

// settingsWithQueue returns a settings repo whose one row sets the default queue.
func settingsWithQueue(t *testing.T, queue string) *repoMocks.SettingsRepo {
	t.Helper()
	return settingsRepoWith(t, &settings.Settings{
		Run: &settings.RunSettings{
			DefaultQueue: &settings.StringSetting{State: stateValue, StringValue: queue},
		},
	})
}

// newMockProjectClientAlwaysOK returns a mock ProjectServiceClient whose GetProject always succeeds.
func newMockProjectClientAlwaysOK(t *testing.T) *projectMocks.ProjectServiceClient {
	pc := projectMocks.NewProjectServiceClient(t)
	pc.EXPECT().GetProject(mock.Anything, mock.Anything).
		Return(&project.GetProjectResponse{}, nil).Maybe()
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
	mux := http.NewServeMux()
	handler := connect.NewServer()
	workflowconnect.RegisterRunServiceHandler(handler, svc)
	connecthttp.Mount(mux, handler, connecthttp.WithReadMaxBytes(0))
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return workflowconnect.NewRunServiceClient(connect.NewClient(connecthttp.NewTransport(http.DefaultClient, server.URL, connecthttp.WithReadMaxBytes(0))))
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

	resp, err := svc.GetRunDetails(context.Background(), &workflow.GetRunDetailsRequest{
		RunId: runID,
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Details)
	require.NotNil(t, resp.Details.Action)
	require.NotNil(t, resp.Details.Action.GetTask())
	assert.Equal(t, "python", resp.Details.Action.GetTask().GetTaskTemplate().GetType())
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

	stream, err := client.WatchClusterEvents(context.Background(), &workflow.WatchClusterEventsRequest{
		Id:      actionID,
		Attempt: 0,
	})
	require.NoError(t, err)
	msg1, receiveErr1 := stream.Receive()
	require.NoError(t, receiveErr1)

	resp := msg1
	require.Len(t, resp.GetClusterEvents(), 1)
	assert.Equal(t, "Pod scheduled on node-a", resp.GetClusterEvents()[0].GetMessage())

	_, err = stream.Receive()
	assert.ErrorIs(t, err, io.EOF)
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

	stream, err := client.WatchClusterEvents(context.Background(), &workflow.WatchClusterEventsRequest{
		Id:      actionID,
		Attempt: 0,
	})
	require.NoError(t, err)

	msg3, receiveErr3 := stream.Receive()
	require.NoError(t, receiveErr3)
	first := msg3
	require.Len(t, first.GetClusterEvents(), 1)
	assert.Equal(t, "Pod created", first.GetClusterEvents()[0].GetMessage())

	msg4, receiveErr4 := stream.Receive()
	require.NoError(t, receiveErr4)
	second := msg4
	require.Len(t, second.GetClusterEvents(), 1)
	assert.Equal(t, "Pulled container image", second.GetClusterEvents()[0].GetMessage())

	_, err = stream.Receive()
	assert.ErrorIs(t, err, io.EOF)
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

	resp, err := svc.GetRunDetails(context.Background(), &workflow.GetRunDetailsRequest{
		RunId: runID,
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Details)
	require.NotNil(t, resp.Details.GetRunSpec())

	gotEnv := map[string]string{}
	for _, kv := range resp.Details.GetRunSpec().GetEnvs().GetValues() {
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

	resp, err := svc.GetRunDetails(context.Background(), &workflow.GetRunDetailsRequest{
		RunId: runID,
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Details)
	assert.Equal(t, core.CatalogCacheStatus_CACHE_HIT, resp.Details.Action.Status.GetCacheStatus())
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

	_, err := svc.GetRunDetails(context.Background(), &workflow.GetRunDetailsRequest{
		RunId: runID,
	})
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
		settingsRepo:  noSettings(t),
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
		Return(&actions.EnqueueResponse{}, nil).Once()

	resp, err := svc.CreateRun(context.Background(), &workflow.CreateRunRequest{
		Id: &workflow.CreateRunRequest_RunId{
			RunId: runID,
		},
		Task: &workflow.CreateRunRequest_TaskSpec{
			TaskSpec: &task.TaskSpec{},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.GetRun())
	assert.NotNil(t, resp.GetRun().GetAction())
	assert.NotNil(t, resp.GetRun().GetAction().GetId())
	assert.Equal(t, runID.Name, resp.GetRun().GetAction().GetId().GetName())
	assert.NotNil(t, resp.GetRun().GetAction().GetMetadata())

	status := resp.GetRun().GetAction().GetStatus()
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

		_, err := svc.AbortRun(context.Background(), &workflow.AbortRunRequest{RunId: runID})
		assert.NoError(t, err)
		actionsClient.AssertNotCalled(t, "Abort")
	})

	t.Run("success with custom reason", func(t *testing.T) {
		actionRepo, actionsClient, svc := newTestService(t)
		reason := "timeout exceeded"

		actionRepo.On("AbortRun", mock.Anything, runID, reason, (*common.EnrichedIdentity)(nil)).Return(nil)

		_, err := svc.AbortRun(context.Background(), &workflow.AbortRunRequest{RunId: runID, Reason: &reason})
		assert.NoError(t, err)
		actionsClient.AssertNotCalled(t, "Abort")
	})

	t.Run("db error returns error", func(t *testing.T) {
		actionRepo, actionsClient, svc := newTestService(t)

		actionRepo.On("AbortRun", mock.Anything, runID, mock.Anything, mock.Anything).Return(errors.New("db unavailable"))

		_, err := svc.AbortRun(context.Background(), &workflow.AbortRunRequest{RunId: runID})
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

		_, err := svc.AbortAction(context.Background(), &workflow.AbortActionRequest{ActionId: actionID})
		assert.NoError(t, err)
		actionsClient.AssertNotCalled(t, "Abort")
	})

	t.Run("success with custom reason", func(t *testing.T) {
		actionRepo, actionsClient, svc := newTestService(t)
		reason := "resource limit exceeded"

		actionRepo.On("AbortAction", mock.Anything, actionID, reason, (*common.EnrichedIdentity)(nil)).Return(nil)

		_, err := svc.AbortAction(context.Background(), &workflow.AbortActionRequest{ActionId: actionID, Reason: reason})
		assert.NoError(t, err)
		actionsClient.AssertNotCalled(t, "Abort")
	})

	t.Run("db error returns error", func(t *testing.T) {
		actionRepo, actionsClient, svc := newTestService(t)

		actionRepo.On("AbortAction", mock.Anything, actionID, mock.Anything, mock.Anything).Return(errors.New("db unavailable"))

		_, err := svc.AbortAction(context.Background(), &workflow.AbortActionRequest{ActionId: actionID})
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
	// The page token is the running offset (strconv.Itoa(offset+limit)); the incoming
	// token is decoded back into listInput.Offset. A request at offset 5 with limit 2 that
	// still has a next page yields token "7".
	testCases := []struct {
		name         string
		req          *common.ListRequest
		expectOffset int
		mockRes      mockListRes
		expect       *workflow.ListRunsResponse
	}{
		{
			"Empty Runs",
			&common.ListRequest{Limit: 2},
			0,
			mockListRes{runs: []*models.Run{}, err: nil},
			&workflow.ListRunsResponse{Runs: []*workflow.Run{}, Token: ""},
		},
		{
			// Service fetches Limit+1 rows to detect another page. With 3 rows returned
			// for a limit of 2, the slice is trimmed to the first 2 runs and the next-page
			// token is the next offset (5 + 2 = 7).
			"list with limit 2 and token",
			&common.ListRequest{Limit: 2, Token: "5"},
			5,
			mockListRes{runs: sqlRes[5:8], err: nil},
			&workflow.ListRunsResponse{Runs: runs[5:7], Token: "7"},
		},
		{
			// Only 2 rows returned for limit 3 means no next page — token empty.
			"list with limit 3 and token",
			&common.ListRequest{Limit: 3, Token: "8"},
			8,
			mockListRes{runs: sqlRes[8:10], err: nil},
			&workflow.ListRunsResponse{Runs: runs[8:10], Token: ""},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &workflow.ListRunsRequest{Request: tc.req}
			// Assert the incoming token was decoded into listInput.Offset.
			actionRepo.On("ListActions", mock.Anything, mock.MatchedBy(func(input interfaces.ListResourceInput) bool {
				return input.Offset == tc.expectOffset
			})).Return(tc.mockRes.runs, tc.mockRes.err).Once()
			got, err := svc.ListRuns(context.Background(), req)
			assert.NoError(t, err)
			assert.Equal(t, len(tc.expect.Runs), len(got.Runs))
			assert.Equal(t, tc.expect.Runs, got.Runs)
			assert.Equal(t, tc.expect.Token, got.Token)
		})
	}
}

// TestListRuns_SearchFilter verifies the console's synthetic "search" filter is resolved
// into the run-name-or-task-name predicate before the allow-list parser sees the request,
// that plain column filters still pass through alongside it, and that the caller's request
// is not mutated.
func TestListRuns_SearchFilter(t *testing.T) {
	actionRepo, _, svc := newTestService(t)

	var captured interfaces.ListResourceInput
	actionRepo.On("ListActions", mock.Anything, mock.MatchedBy(func(input interfaces.ListResourceInput) bool {
		captured = input
		return true
	})).Return([]*models.Run{}, nil).Once()

	listReq := &common.ListRequest{
		Limit: 10,
		Filters: []*common.Filter{
			{Field: "search", Function: common.Filter_CONTAINS_CASE_INSENSITIVE, Values: []string{"abc"}},
			{Field: "phase", Function: common.Filter_EQUAL, Values: []string{"2"}},
		},
	}
	_, err := svc.ListRuns(context.Background(), &workflow.ListRunsRequest{
		ScopeBy: &workflow.ListRunsRequest_ProjectId{ProjectId: &common.ProjectIdentifier{Name: "p", Domain: "d"}},
		Request: listReq,
	})
	require.NoError(t, err)

	require.NotNil(t, captured.Filter)
	expr, err := captured.Filter.QueryExpression("")
	require.NoError(t, err)
	assert.Contains(t, expr.Query, "(LOWER(run_name) LIKE LOWER(?)) OR (LOWER(task_name) LIKE LOWER(?))")
	assert.Contains(t, expr.Query, "phase = ?")
	assert.Contains(t, expr.Args, "%abc%")

	// The request handed in by the caller still carries the search filter.
	require.Len(t, listReq.Filters, 2)
	assert.Equal(t, "search", listReq.Filters[0].GetField())
}

// TestListRuns_RejectsBadSearchFilter verifies a search filter with an unsupported operator
// is an InvalidArgument and never reaches the repository. The service is built inline
// because newTestService requires the repository accessor to be used, and this path must
// fail before it is.
func TestListRuns_RejectsBadSearchFilter(t *testing.T) {
	actionRepo := &repoMocks.ActionRepo{} // no expectations: any ListActions call fails the test
	repo := &repoMocks.Repository{}
	repo.On("ActionRepo").Maybe().Return(actionRepo)
	svc := &RunService{repo: repo}
	t.Cleanup(func() { actionRepo.AssertExpectations(t) })

	_, err := svc.ListRuns(context.Background(), &workflow.ListRunsRequest{
		ScopeBy: &workflow.ListRunsRequest_ProjectId{ProjectId: &common.ProjectIdentifier{Name: "p", Domain: "d"}},
		Request: &common.ListRequest{
			Limit:   10,
			Filters: []*common.Filter{{Field: "search", Function: common.Filter_GREATER_THAN, Values: []string{"abc"}}},
		},
	})
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

// TestListRuns_LabelFilter verifies "labels.<key>" filters reach the repository as label
// predicates next to plain column filters, and that each listed run carries its labels.
func TestListRuns_LabelFilter(t *testing.T) {
	actionRepo, _, svc := newTestService(t)

	var captured interfaces.ListResourceInput
	actionRepo.On("ListActions", mock.Anything, mock.MatchedBy(func(input interfaces.ListResourceInput) bool {
		captured = input
		return true
	})).Return([]*models.Run{{
		Project: "p",
		Domain:  "d",
		RunName: "run-1",
		Name:    "a0",
		Labels:  []byte(`{"example.com/job-id":"j-1"}`),
	}}, nil).Once()

	got, err := svc.ListRuns(context.Background(), &workflow.ListRunsRequest{
		ScopeBy: &workflow.ListRunsRequest_ProjectId{ProjectId: &common.ProjectIdentifier{Name: "p", Domain: "d"}},
		Request: &common.ListRequest{
			Limit: 10,
			Filters: []*common.Filter{
				{Field: "labels.example.com/job-id", Function: common.Filter_EQUAL, Values: []string{"j-1"}},
				{Field: "phase", Function: common.Filter_EQUAL, Values: []string{"2"}},
			},
		},
	})
	require.NoError(t, err)

	require.NotNil(t, captured.Filter)
	expr, err := captured.Filter.QueryExpression("")
	require.NoError(t, err)
	assert.Contains(t, expr.Query, "labels @> ?::jsonb")
	assert.Contains(t, expr.Query, "phase = ?")
	assert.Contains(t, expr.Args, `{"example.com/job-id":"j-1"}`)

	require.Len(t, got.Runs, 1)
	assert.Equal(t, map[string]string{"example.com/job-id": "j-1"}, got.Runs[0].GetLabels())
}

// TestListRuns_LabelsFromRunSpec verifies a run stored before the labels column existed
// is still listed with the labels in its run spec.
func TestListRuns_LabelsFromRunSpec(t *testing.T) {
	actionRepo, _, svc := newTestService(t)

	runSpec, err := proto.Marshal(&task.RunSpec{
		Labels: &task.Labels{Values: map[string]string{"example.com/job-id": "j-1"}},
	})
	require.NoError(t, err)
	actionRepo.On("ListActions", mock.Anything, mock.Anything).Return([]*models.Run{{
		Project: "p",
		Domain:  "d",
		RunName: "run-1",
		Name:    "a0",
		RunSpec: runSpec,
	}}, nil).Once()

	got, err := svc.ListRuns(context.Background(), &workflow.ListRunsRequest{
		ScopeBy: &workflow.ListRunsRequest_ProjectId{ProjectId: &common.ProjectIdentifier{Name: "p", Domain: "d"}},
		Request: &common.ListRequest{Limit: 10},
	})
	require.NoError(t, err)

	require.Len(t, got.Runs, 1)
	assert.Equal(t, map[string]string{"example.com/job-id": "j-1"}, got.Runs[0].GetLabels())
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

	err = svc.listAndSendAllActions(context.Background(), runID, rsm, workflowconnect.RunServiceWatchActionsServerStream{})
	require.NoError(t, err)

	require.Len(t, captured.SortParameters, 2)
	assert.Equal(t, "created_at ASC", captured.SortParameters[0].GetOrderExpr())
	assert.Equal(t, "name ASC", captured.SortParameters[1].GetOrderExpr())
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

	t.Run("generated names pass run name validation", func(t *testing.T) {
		for seed := int64(0); seed < 100; seed++ {
			assert.NoError(t, validateRunName(generateRunName(seed)))
		}
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
		settingsRepo:  noSettings(t),
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
		Return(&actions.EnqueueResponse{}, nil).Once()

	_, err := svc.CreateRun(context.Background(), req)
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
		settingsRepo:  noSettings(t),
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
	actionsClient.On("Enqueue", mock.Anything, mock.Anything).Return(&actions.EnqueueResponse{}, nil).Once()

	resp, err := svc.CreateRun(context.Background(), req)
	require.NoError(t, err)

	run := resp.Run
	assert.Equal(t, "", run.Action.Id.Run.Org)
	assert.Equal(t, "myproj", run.Action.Id.Run.Project)
	assert.Equal(t, "production", run.Action.Id.Run.Domain)
	assert.Equal(t, "rtest99999", run.Action.Id.Run.Name)
	assert.Equal(t, "a0", run.Action.Id.Name)
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_QUEUED, run.Action.Status.Phase)
}

// TestCreateRun_TriggerFire_CarriesRunSpecEnvVars locks in that a run fired from a trigger
// carries the run-level env vars captured in the trigger's stored run spec. The scheduler fires
// with only a TriggerName (no run spec), so CreateRun must restore the run spec from the trigger
// — otherwise Trigger(env_vars=...) is silently dropped before the pod is built.
func TestCreateRun_TriggerFire_CarriesRunSpecEnvVars(t *testing.T) {
	actionRepo := &repoMocks.ActionRepo{}
	taskRepo := &repoMocks.TaskRepo{}
	triggerRepo := &repoMocks.TriggerRepo{}
	actionsClient := actionsconnectmocks.NewActionsServiceClient(t)
	repo := &repoMocks.Repository{}
	store := &storageMocks.ComposedProtobufStore{}
	dataStore := &storage.DataStore{ComposedProtobufStore: store}

	repo.On("ActionRepo").Return(actionRepo)
	repo.On("TaskRepo").Return(taskRepo)
	repo.On("TriggerRepo").Return(triggerRepo)

	svc := &RunService{
		repo:          repo,
		settingsRepo:  noSettings(t),
		actionsClient: actionsClient,
		projectClient: newMockProjectClientAlwaysOK(t),
		storagePrefix: "s3://flyte-data",
		dataStore:     dataStore,
	}

	// Trigger stored at registration with a run-level env var.
	triggerSpecBytes, err := proto.Marshal(&triggerpb.TriggerSpec{
		TaskVersion: "v1",
		RunSpec: &task.RunSpec{
			Envs: &task.Envs{
				Values: []*core.KeyValuePair{
					{Key: "TRIGGER_ENV", Value: "from-trigger"},
				},
			},
		},
	})
	require.NoError(t, err)

	triggerRepo.On("GetTrigger", mock.Anything, mock.Anything).Return(&models.Trigger{
		Project:        "myproj",
		Domain:         "production",
		TaskName:       "mytask",
		Name:           "mytrigger",
		TaskVersion:    "v1",
		LatestRevision: 1,
		Spec:           triggerSpecBytes,
	}, nil).Once()

	// Task referenced by the trigger.
	taskRepo.On("GetTask", mock.Anything, mock.Anything).Return(&models.Task{}, nil).Once()
	taskRepo.On("CreateTaskSpec", mock.Anything, mock.Anything).Return(nil).Maybe()

	// Capture the persisted action so we can inspect the run spec it carries.
	var persisted *models.Action
	actionRepo.On("CreateAction", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			persisted = args.Get(1).(*models.Action)
		}).
		Return(&models.Run{
			Project: "myproj",
			Domain:  "production",
			RunName: "rtrig1",
			Name:    "a0",
			Phase:   int32(common.ActionPhase_ACTION_PHASE_QUEUED),
		}, nil).Once()

	store.On("WriteProtobuf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	actionsClient.On("Enqueue", mock.Anything, mock.Anything).
		Return(&actions.EnqueueResponse{}, nil).Once()

	req := &workflow.CreateRunRequest{
		Id: &workflow.CreateRunRequest_RunId{
			RunId: &common.RunIdentifier{
				Project: "myproj",
				Domain:  "production",
				Name:    "rtrig1",
			},
		},
		Task: &workflow.CreateRunRequest_TriggerName{
			TriggerName: &common.TriggerName{
				Project:  "myproj",
				Domain:   "production",
				TaskName: "mytask",
				Name:     "mytrigger",
			},
		},
		Source: workflow.RunSource_RUN_SOURCE_SCHEDULE_TRIGGER,
	}

	_, err = svc.CreateRun(context.Background(), req)
	require.NoError(t, err)

	require.NotNil(t, persisted, "CreateAction should have been called")
	var gotRunSpec task.RunSpec
	require.NoError(t, proto.Unmarshal(persisted.RunSpec, &gotRunSpec))
	gotEnv := map[string]string{}
	for _, kv := range gotRunSpec.GetEnvs().GetValues() {
		gotEnv[kv.GetKey()] = kv.GetValue()
	}
	assert.Equal(t, "from-trigger", gotEnv["TRIGGER_ENV"],
		"trigger-fired run must carry the trigger's run-level env vars")
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
		settingsRepo:  noSettings(t),
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
	actionsClient.On("Enqueue", mock.Anything, mock.MatchedBy(func(req *actions.EnqueueRequest) bool {
		return req.Action.ActionId.Name == "a0" &&
			req.Action.ActionId.Run.Name == "rmy-run-123"
	})).Return(&actions.EnqueueResponse{}, nil).Once()

	_, err := svc.CreateRun(context.Background(), req)
	assert.NoError(t, err)

	actionsClient.AssertExpectations(t)
}

func TestCreateRun_RejectsInvalidRunName(t *testing.T) {
	// Invalid names must be rejected before any repo or project access, so no mocks are needed.
	svc := &RunService{}

	tests := []struct {
		name    string
		runName string
	}{
		{"dot", "my.run"},
		{"leading digit", "1run"},
		{"uppercase", "My-Run"},
		{"underscore", "my_run"},
		{"leading hyphen", "-run"},
		{"trailing hyphen", "run-"},
		{"too long", strings.Repeat("a", runNameMaxLength+1)},
		{"empty", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := &workflow.CreateRunRequest{
				Id: &workflow.CreateRunRequest_RunId{
					RunId: &common.RunIdentifier{
						Org:     "org",
						Project: "proj",
						Domain:  "dev",
						Name:    tc.runName,
					},
				},
				Task: &workflow.CreateRunRequest_TaskSpec{
					TaskSpec: &task.TaskSpec{},
				},
			}

			_, err := svc.CreateRun(context.Background(), req)
			require.Error(t, err)
			assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
		})
	}
}

func TestCreateRun_RejectsInvalidPodTemplateName(t *testing.T) {
	// Only the settings and project lookups run before the check. Everything else stays
	// nil on purpose: were the check missing, the write path would panic and fail the test.
	svc := &RunService{
		settingsRepo: settingsRepoWith(t, &settings.Settings{
			PodTemplateName: &settings.StringSetting{State: stateValue, StringValue: "GPU Template"},
		}),
		projectClient: newMockProjectClientAlwaysOK(t),
	}

	req := &workflow.CreateRunRequest{
		Id: &workflow.CreateRunRequest_RunId{
			RunId: &common.RunIdentifier{
				Org:     "org",
				Project: "proj",
				Domain:  "dev",
				Name:    "rq-123",
			},
		},
		InputWrapper: &workflow.CreateRunRequest_Inputs{Inputs: &task.Inputs{}},
		Task: &workflow.CreateRunRequest_TaskSpec{
			TaskSpec: &task.TaskSpec{},
		},
	}

	_, err := svc.CreateRun(context.Background(), req)
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
	assert.ErrorContains(t, err, "pod template name")
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
		settingsRepo:  noSettings(t),
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
			Inputs: &task.Inputs{
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

	actionsClient.On("Enqueue", mock.Anything, mock.MatchedBy(func(req *actions.EnqueueRequest) bool {
		return req.GetRunSpec().GetRawDataStorage().GetRawDataPrefix() == "s3://custom-raw"
	})).Return(&actions.EnqueueResponse{}, nil).Once()

	_, err := svc.CreateRun(context.Background(), req)
	require.NoError(t, err)
}

// Without a request or settings prefix, raw data stays under RunOutputBase, so the
// stored run spec must not be stamped with the cluster default.
func TestCreateRun_LeavesRawDataPrefixEmpty(t *testing.T) {
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
		settingsRepo:  noSettings(t),
		actionsClient: actionsClient,
		projectClient: newMockProjectClientAlwaysOK(t),
		storagePrefix: "s3://flyte-data",
		dataStore:     dataStore,
	}

	req := &workflow.CreateRunRequest{
		Id: &workflow.CreateRunRequest_RunId{
			RunId: &common.RunIdentifier{Org: "org", Project: "proj", Domain: "dev", Name: "rraw-123"},
		},
		RunSpec: &task.RunSpec{RunBaseDir: "s3://user-base"},
		Task:    &workflow.CreateRunRequest_TaskSpec{TaskSpec: &task.TaskSpec{}},
	}

	store.On("WriteProtobuf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	taskRepo.On("CreateTaskSpec", mock.Anything, mock.Anything).Return(nil).Once()
	actionRepo.On("CreateAction", mock.Anything, mock.MatchedBy(func(m *models.Run) bool {
		var rs task.RunSpec
		_ = proto.Unmarshal(m.RunSpec, &rs)
		return rs.GetRawDataStorage().GetRawDataPrefix() == ""
	}), mock.Anything).Return(&models.Run{Project: "proj", Domain: "dev", Name: "rraw-123"}, nil).Once()
	actionsClient.On("Enqueue", mock.Anything, mock.MatchedBy(func(req *actions.EnqueueRequest) bool {
		return req.GetRunSpec().GetRawDataStorage().GetRawDataPrefix() == ""
	})).Return(&actions.EnqueueResponse{}, nil).Once()

	_, err := svc.CreateRun(context.Background(), req)
	require.NoError(t, err)
}

// TestCreateRun_AppliesSettingsQueue proves the applier is actually wired into run
// creation: the request names no queue, and the value stored on the run comes from
// the org's settings row.
func TestCreateRun_AppliesSettingsQueue(t *testing.T) {
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
		settingsRepo:  settingsWithQueue(t, "fast-queue"),
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
				Name:    "rq-123",
			},
		},
		InputWrapper: &workflow.CreateRunRequest_Inputs{Inputs: &task.Inputs{}},
		Task: &workflow.CreateRunRequest_TaskSpec{
			TaskSpec: &task.TaskSpec{},
		},
	}

	store.On("WriteProtobuf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	taskRepo.On("CreateTaskSpec", mock.Anything, mock.Anything).Return(nil).Once()

	actionRepo.On("CreateAction", mock.Anything, mock.MatchedBy(func(m *models.Run) bool {
		var rs task.RunSpec
		_ = proto.Unmarshal(m.RunSpec, &rs)
		return rs.GetQueue() == "fast-queue"
	}), mock.Anything).Return(&models.Run{
		Project: "proj",
		Domain:  "dev",
		Name:    "rq-123",
	}, nil).Once()

	actionsClient.On("Enqueue", mock.Anything, mock.MatchedBy(func(req *actions.EnqueueRequest) bool {
		return req.GetRunSpec().GetQueue() == "fast-queue"
	})).Return(&actions.EnqueueResponse{}, nil).Once()

	_, err := svc.CreateRun(context.Background(), req)
	require.NoError(t, err)
}

// TestCreateRun_StoresLabels verifies the run's labels are stored in their queryable form,
// an empty object when it has none.
func TestCreateRun_StoresLabels(t *testing.T) {
	for _, tc := range []struct {
		name   string
		labels *task.Labels
		stored string
	}{
		{"labeled", &task.Labels{Values: map[string]string{"example.com/job-id": "j-1"}}, `{"example.com/job-id":"j-1"}`},
		{"unlabeled", nil, `{}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
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
				settingsRepo:  noSettings(t),
				actionsClient: actionsClient,
				projectClient: newMockProjectClientAlwaysOK(t),
				storagePrefix: "s3://flyte-data",
				dataStore:     dataStore,
			}

			req := &workflow.CreateRunRequest{
				Id: &workflow.CreateRunRequest_RunId{
					RunId: &common.RunIdentifier{Org: "org", Project: "proj", Domain: "dev", Name: "rl-123"},
				},
				InputWrapper: &workflow.CreateRunRequest_Inputs{Inputs: &task.Inputs{}},
				RunSpec:      &task.RunSpec{Labels: tc.labels},
				Task:         &workflow.CreateRunRequest_TaskSpec{TaskSpec: &task.TaskSpec{}},
			}

			store.On("WriteProtobuf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			taskRepo.On("CreateTaskSpec", mock.Anything, mock.Anything).Return(nil).Once()
			actionRepo.On("CreateAction", mock.Anything, mock.MatchedBy(func(m *models.Run) bool {
				return string(m.Labels) == tc.stored
			}), mock.Anything).Return(&models.Run{Project: "proj", Domain: "dev", Name: "rl-123"}, nil).Once()
			actionsClient.On("Enqueue", mock.Anything, mock.Anything).Return(&actions.EnqueueResponse{}, nil).Once()

			_, err := svc.CreateRun(context.Background(), req)
			require.NoError(t, err)
			actionRepo.AssertExpectations(t)
		})
	}
}

// TestCreateRun_RejectsNULLabel verifies a label the labels column cannot store is an
// InvalidArgument rather than a failed insert.
func TestCreateRun_RejectsNULLabel(t *testing.T) {
	actionRepo := &repoMocks.ActionRepo{} // no expectations: any CreateAction call fails the test
	repo := &repoMocks.Repository{}
	repo.On("ActionRepo").Maybe().Return(actionRepo)
	repo.On("TaskRepo").Maybe().Return(&repoMocks.TaskRepo{})
	t.Cleanup(func() { actionRepo.AssertExpectations(t) })

	svc := &RunService{
		repo:          repo,
		settingsRepo:  noSettings(t),
		projectClient: newMockProjectClientAlwaysOK(t),
		storagePrefix: "s3://flyte-data",
	}

	_, err := svc.CreateRun(context.Background(), &workflow.CreateRunRequest{
		Id: &workflow.CreateRunRequest_RunId{
			RunId: &common.RunIdentifier{Org: "org", Project: "proj", Domain: "dev", Name: "rl-nul"},
		},
		InputWrapper: &workflow.CreateRunRequest_Inputs{Inputs: &task.Inputs{}},
		RunSpec: &task.RunSpec{
			Labels: &task.Labels{Values: map[string]string{"k": "a\x00b"}},
		},
		Task: &workflow.CreateRunRequest_TaskSpec{TaskSpec: &task.TaskSpec{}},
	})
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
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
		settingsRepo:  noSettings(t),
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
		return info.InputsUri == offloadedURI+"/inputs.pb"
	}), mock.Anything).Return(expectedRun, nil).Once()
	actionsClient.On("Enqueue", mock.MatchedBy(func(_ context.Context) bool { return true }), mock.MatchedBy(func(req *actions.EnqueueRequest) bool {
		return req.Action.InputUri == offloadedURI
	})).Return(&actions.EnqueueResponse{}, nil).Once()

	resp, err := svc.CreateRun(context.Background(), req)
	require.NoError(t, err)

	run := resp.Run
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

		resp, err := svc.GetActionDataURIs(context.Background(), &workflow.GetActionDataURIsRequest{
			ActionId: actionID,
		})

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

		resp, err := svc.GetActionDataURIs(context.Background(), &workflow.GetActionDataURIsRequest{
			ActionId: actionID,
		})

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

		resp, err := svc.GetActionDataURIs(context.Background(), &workflow.GetActionDataURIsRequest{
			ActionId: actionID,
		})

		require.NoError(t, err)
		assert.Equal(t, "s3://bucket/inputs", resp.GetInputsUri())
		assert.Empty(t, resp.GetOutputsUri())
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

		resp, err := svc.GetActionDataURIs(context.Background(), &workflow.GetActionDataURIsRequest{
			ActionId: actionID,
		})

		require.NoError(t, err)
		assert.Equal(t, "s3://bucket/inputs", resp.GetInputsUri())
		assert.Equal(t, "s3://bucket/trace-outputs", resp.GetOutputsUri())
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

		resp, err := svc.GetActionDataURIs(context.Background(), &workflow.GetActionDataURIsRequest{
			ActionId: actionID,
		})

		require.NoError(t, err)
		assert.Equal(t, "s3://bucket/inputs", resp.GetInputsUri())
		assert.Equal(t, "s3://bucket/outputs/0/outputs.pb", resp.GetOutputsUri())
	})

	t.Run("recovered trace action returns outputs uri from RunInfo", func(t *testing.T) {
		actionRepo, _, svc := newTestService(t)
		actionRepo.On("GetAction", mock.Anything, matchActionID(actionID)).Return(&models.Action{
			Phase:      int32(common.ActionPhase_ACTION_PHASE_RECOVERED),
			ActionType: int32(workflow.ActionType_ACTION_TYPE_TRACE),
			DetailedInfo: mustMarshalRunInfo(&workflow.RunInfo{
				InputsUri:  "s3://bucket/inputs",
				OutputsUri: "s3://bucket/trace-outputs",
			}),
		}, nil)

		resp, err := svc.GetActionDataURIs(context.Background(), &workflow.GetActionDataURIsRequest{
			ActionId: actionID,
		})

		require.NoError(t, err)
		assert.Equal(t, "s3://bucket/inputs", resp.GetInputsUri())
		assert.Equal(t, "s3://bucket/trace-outputs", resp.GetOutputsUri())
	})

	t.Run("recovered task returns outputs from latest attempt", func(t *testing.T) {
		actionRepo, _, svc := newTestService(t)
		actionRepo.On("GetAction", mock.Anything, matchActionID(actionID)).Return(&models.Action{
			Phase:      int32(common.ActionPhase_ACTION_PHASE_RECOVERED),
			ActionType: int32(workflow.ActionType_ACTION_TYPE_TASK),
			DetailedInfo: mustMarshalRunInfo(&workflow.RunInfo{
				InputsUri: "s3://bucket/inputs",
			}),
		}, nil)

		eventBytes := mustMarshalEvent(&workflow.ActionEvent{
			Id:      actionID,
			Attempt: 0,
			Phase:   common.ActionPhase_ACTION_PHASE_RECOVERED,
			Outputs: &task.OutputReferences{
				OutputUri: "s3://bucket/outputs/0/outputs.pb",
			},
		})
		actionRepo.On("ListEvents", mock.Anything, matchActionID(actionID), 500).Return([]*models.ActionEvent{
			{Attempt: 0, Phase: int32(common.ActionPhase_ACTION_PHASE_RECOVERED), Info: eventBytes},
		}, nil)

		resp, err := svc.GetActionDataURIs(context.Background(), &workflow.GetActionDataURIsRequest{
			ActionId: actionID,
		})

		require.NoError(t, err)
		assert.Equal(t, "s3://bucket/inputs", resp.GetInputsUri())
		assert.Equal(t, "s3://bucket/outputs/0/outputs.pb", resp.GetOutputsUri())
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

		resp, err := svc.GetActionDataURIs(context.Background(), &workflow.GetActionDataURIsRequest{
			ActionId: actionID,
		})

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

		resp, err := svc.GetActionDataURIs(context.Background(), &workflow.GetActionDataURIsRequest{
			ActionId: actionID,
		})

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

		resp, err := svc.GetActionDataURIs(context.Background(), &workflow.GetActionDataURIsRequest{
			ActionId: actionID,
		})

		require.NoError(t, err)
		assert.Equal(t, "s3://bucket/outputs/1/outputs.pb", resp.GetOutputsUri())
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

		resp, err := svc.GetActionLogContext(context.Background(), &workflow.GetActionLogContextRequest{
			ActionId: actionID,
			Attempt:  1,
		})

		require.NoError(t, err)
		assert.Equal(t, "c1", resp.GetCluster())
		assert.Equal(t, "my-pod", resp.GetLogContext().GetPrimaryPodName())
	})

	t.Run("no event found returns NotFound", func(t *testing.T) {
		actionRepo, _, svc := newTestService(t)
		actionRepo.On("GetLatestEventByAttempt", mock.Anything, matchActionID(actionID), uint32(1)).Return(nil, sql.ErrNoRows)

		resp, err := svc.GetActionLogContext(context.Background(), &workflow.GetActionLogContextRequest{
			ActionId: actionID,
			Attempt:  1,
		})

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
	})

	t.Run("repo error returns Internal", func(t *testing.T) {
		actionRepo, _, svc := newTestService(t)
		actionRepo.On("GetLatestEventByAttempt", mock.Anything, matchActionID(actionID), uint32(1)).Return(nil, errors.New("db blew up"))

		resp, err := svc.GetActionLogContext(context.Background(), &workflow.GetActionLogContextRequest{
			ActionId: actionID,
			Attempt:  1,
		})

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

		resp, err := svc.GetActionLogContext(context.Background(), &workflow.GetActionLogContextRequest{
			ActionId: actionID,
			Attempt:  1,
		})

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
	})

	t.Run("undeserializable event returns Internal", func(t *testing.T) {
		actionRepo, _, svc := newTestService(t)
		actionRepo.On("GetLatestEventByAttempt", mock.Anything, matchActionID(actionID), uint32(1)).Return(&models.ActionEvent{
			Info: []byte("not-a-proto"),
		}, nil)

		resp, err := svc.GetActionLogContext(context.Background(), &workflow.GetActionLogContextRequest{
			ActionId: actionID,
			Attempt:  1,
		})

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, connect.CodeInternal, connect.CodeOf(err))
	})
}

func TestActionModelToDetails(t *testing.T) {
	svc := &RunService{}
	now := time.Now().UTC().Truncate(time.Millisecond)
	end := now.Add(5 * time.Second)

	tests := []struct {
		name     string
		action   *models.Action
		actionID *common.ActionIdentifier
		verify   func(t *testing.T, result *workflow.ActionDetails)
	}{
		{
			name: "BasicStatusFields",
			action: &models.Action{
				Phase:       int32(common.ActionPhase_ACTION_PHASE_RUNNING),
				CreatedAt:   now,
				Attempts:    2,
				CacheStatus: core.CatalogCacheStatus_CACHE_HIT,
			},
			actionID: testActionID,
			verify: func(t *testing.T, result *workflow.ActionDetails) {
				require.NotNil(t, result)
				assert.Equal(t, testActionID, result.Id)
				assert.Equal(t, common.ActionPhase_ACTION_PHASE_RUNNING, result.Status.Phase)
				assert.Equal(t, now, result.Status.StartTime.AsTime())
				assert.Equal(t, uint32(2), result.Status.Attempts)
				assert.Equal(t, core.CatalogCacheStatus_CACHE_HIT, result.Status.CacheStatus)
				assert.Nil(t, result.Status.EndTime)
				assert.Nil(t, result.Status.DurationMs)
			},
		},
		{
			name: "EndedAtSetsDurationFromTimestamps",
			action: &models.Action{
				Phase:     int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED),
				CreatedAt: now,
				EndedAt:   sql.NullTime{Time: end, Valid: true},
			},
			actionID: testActionID,
			verify: func(t *testing.T, result *workflow.ActionDetails) {
				require.NotNil(t, result.Status.EndTime)
				assert.Equal(t, end, result.Status.EndTime.AsTime())
				require.NotNil(t, result.Status.DurationMs)
				assert.Equal(t, uint64(5000), *result.Status.DurationMs)
			},
		},
		{
			name: "MetadataOptionalFields",
			action: &models.Action{
				Phase:            int32(common.ActionPhase_ACTION_PHASE_RUNNING),
				ActionType:       int32(workflow.ActionType_ACTION_TYPE_TASK),
				FunctionName:     "my_func",
				ParentActionName: sql.NullString{String: "parent-action", Valid: true},
				ActionGroup:      sql.NullString{String: "group-1", Valid: true},
				EnvironmentName:  sql.NullString{String: "prod", Valid: true},
			},
			actionID: testActionID,
			verify: func(t *testing.T, result *workflow.ActionDetails) {
				require.NotNil(t, result.Metadata)
				assert.Equal(t, workflow.ActionType_ACTION_TYPE_TASK, result.Metadata.ActionType)
				assert.Equal(t, "my_func", result.Metadata.FuntionName)
				assert.Equal(t, "parent-action", result.Metadata.Parent)
				assert.Equal(t, "group-1", result.Metadata.Group)
				assert.Equal(t, "prod", result.Metadata.EnvironmentName)
			},
		},
		{
			name: "TaskMetadataWithFullID",
			action: &models.Action{
				Phase:         int32(common.ActionPhase_ACTION_PHASE_RUNNING),
				ActionType:    int32(workflow.ActionType_ACTION_TYPE_TASK),
				TaskType:      "python",
				TaskProject:   sql.NullString{String: "proj", Valid: true},
				TaskDomain:    sql.NullString{String: "dev", Valid: true},
				TaskName:      sql.NullString{String: "my_task", Valid: true},
				TaskVersion:   sql.NullString{String: "v1", Valid: true},
				TaskShortName: sql.NullString{String: "short", Valid: true},
			},
			actionID: testActionID,
			verify: func(t *testing.T, result *workflow.ActionDetails) {
				taskMeta := result.Metadata.GetTask()
				require.NotNil(t, taskMeta)
				assert.Equal(t, "python", taskMeta.TaskType)
				assert.Equal(t, "short", taskMeta.ShortName)
				require.NotNil(t, taskMeta.Id)
				assert.Equal(t, "proj", taskMeta.Id.Project)
				assert.Equal(t, "dev", taskMeta.Id.Domain)
				assert.Equal(t, "my_task", taskMeta.Id.Name)
				assert.Equal(t, "v1", taskMeta.Id.Version)
			},
		},
		{
			name: "TraceMetadata",
			action: &models.Action{
				Phase:        int32(common.ActionPhase_ACTION_PHASE_RUNNING),
				ActionType:   int32(workflow.ActionType_ACTION_TYPE_TRACE),
				FunctionName: "trace_func",
			},
			actionID: testActionID,
			verify: func(t *testing.T, result *workflow.ActionDetails) {
				traceMeta := result.Metadata.GetTrace()
				require.NotNil(t, traceMeta)
				assert.Equal(t, "trace_func", traceMeta.Name)
				assert.Nil(t, result.Metadata.GetTask())
			},
		},
		{
			name: "NilActionID",
			action: &models.Action{
				Phase: int32(common.ActionPhase_ACTION_PHASE_QUEUED),
			},
			actionID: nil,
			verify: func(t *testing.T, result *workflow.ActionDetails) {
				require.NotNil(t, result)
				assert.Nil(t, result.Id)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.verify(t, svc.actionModelToDetails(tc.action, tc.actionID))
		})
	}
}

func TestBuildRunDetails(t *testing.T) {
	runID := &common.RunIdentifier{
		Org:     "test-org",
		Project: "test-project",
		Domain:  "test-domain",
		Name:    "test-run",
	}
	rootActionID := &common.ActionIdentifier{Run: runID, Name: runID.Name}

	// persistRunModel stores ActionSpec with proto.Marshal, so the fixture must too.
	actionSpecBytes, err := proto.Marshal(&workflow.ActionSpec{
		RunSpec: &task.RunSpec{
			Labels: &task.Labels{Values: map[string]string{"env": "prod"}},
		},
	})
	require.NoError(t, err)
	runSpecBytes, err := proto.Marshal(&task.RunSpec{
		Labels: &task.Labels{Values: map[string]string{"env": "staging"}},
	})
	require.NoError(t, err)

	tests := []struct {
		name   string
		run    *models.Run
		verify func(t *testing.T, result *workflow.RunDetails)
	}{
		{
			name: "EmptyActionSpecNilRunSpec",
			run:  &models.Run{Name: runID.Name, Phase: int32(common.ActionPhase_ACTION_PHASE_RUNNING)},
			verify: func(t *testing.T, result *workflow.RunDetails) {
				assert.Nil(t, result.RunSpec)
			},
		},
		{
			name: "ProtoActionSpecExtractsRunSpec",
			run: &models.Run{
				Name:       runID.Name,
				Phase:      int32(common.ActionPhase_ACTION_PHASE_RUNNING),
				ActionSpec: actionSpecBytes,
			},
			verify: func(t *testing.T, result *workflow.RunDetails) {
				require.NotNil(t, result.RunSpec)
				assert.Equal(t, "prod", result.RunSpec.GetLabels().GetValues()["env"])
			},
		},
		{
			name: "RunSpecColumnPreferredOverActionSpec",
			run: &models.Run{
				Name:       runID.Name,
				Phase:      int32(common.ActionPhase_ACTION_PHASE_RUNNING),
				ActionSpec: actionSpecBytes,
				RunSpec:    runSpecBytes,
			},
			verify: func(t *testing.T, result *workflow.RunDetails) {
				require.NotNil(t, result.RunSpec)
				assert.Equal(t, "staging", result.RunSpec.GetLabels().GetValues()["env"])
			},
		},
		{
			name: "MalformedActionSpecNilRunSpec",
			run: &models.Run{
				Name:       runID.Name,
				Phase:      int32(common.ActionPhase_ACTION_PHASE_RUNNING),
				ActionSpec: []byte("not-valid-proto{{"),
			},
			verify: func(t *testing.T, result *workflow.RunDetails) {
				assert.Nil(t, result.RunSpec)
			},
		},
		{
			name: "RootActionIDPropagated",
			run:  &models.Run{Name: runID.Name, Phase: int32(common.ActionPhase_ACTION_PHASE_QUEUED)},
			verify: func(t *testing.T, result *workflow.RunDetails) {
				require.NotNil(t, result.Action)
				assert.True(t, proto.Equal(rootActionID, result.Action.Id))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actionRepo, _, svc := newTestService(t)
			actionRepo.On("ListEvents", mock.Anything, matchActionID(rootActionID), 500).Return([]*models.ActionEvent{}, nil)

			result, err := svc.buildRunDetails(context.Background(), tc.run, runID)
			require.NoError(t, err)
			require.NotNil(t, result)
			tc.verify(t, result)
		})
	}
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
				stream, err := client.WatchRunDetails(context.Background(), &workflow.WatchRunDetailsRequest{
					RunId: runID,
				})
				require.NoError(t, err)
				_, err = stream.Receive()
				require.Error(t, err)
				var connectErr *connect.Error
				require.True(t, errors.As(err, &connectErr))
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

				stream, err := client.WatchRunDetails(ctx, &workflow.WatchRunDetailsRequest{
					RunId: runID,
				})
				require.NoError(t, err)

				msg7, receiveErr7 := stream.Receive()
				require.NoError(t, receiveErr7)
				details := msg7.Details
				require.NotNil(t, details)
				assert.Equal(t, "prod", details.GetRunSpec().GetLabels().GetValues()["env"])
				assert.True(t, proto.Equal(rootActionID, details.GetAction().GetId()))
				assert.Equal(t, common.ActionPhase_ACTION_PHASE_RUNNING, details.GetAction().GetStatus().GetPhase())

				cancel()
				_, err = stream.Receive()
				assert.Error(t, err)
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
				stream, err := client.WatchRunDetails(context.Background(), &workflow.WatchRunDetailsRequest{
					RunId: runID,
				})
				require.NoError(t, err)

				_, receiveErr9 := stream.Receive()
				require.NoError(t, receiveErr9)
				_, err = stream.Receive()
				require.Error(t, err)
				var connectErr *connect.Error
				require.True(t, errors.As(err, &connectErr))
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

				stream, err := client.WatchRunDetails(ctx, &workflow.WatchRunDetailsRequest{
					RunId: runID,
				})
				require.NoError(t, err)

				_, receiveErr11 := stream.Receive()
				require.NoError(t, receiveErr11)
				msg12, receiveErr12 := stream.Receive()
				require.NoError(t, receiveErr12)
				details := msg12.Details
				require.NotNil(t, details)
				assert.Equal(t, "prod", details.GetRunSpec().GetLabels().GetValues()["env"])
				action := details.GetAction()
				assert.Equal(t, common.ActionPhase_ACTION_PHASE_FAILED, action.GetStatus().GetPhase())
				require.Len(t, action.GetAttempts(), 1)
				assert.Equal(t, workflow.ErrorInfo_KIND_USER, action.GetErrorInfo().GetKind())
				assert.Equal(t, "boom", action.GetErrorInfo().GetMessage())

				cancel()
				_, err = stream.Receive()
				assert.Error(t, err)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actionRepo, _, svc := newTestService(t)
			tc.setupMocks(actionRepo)
			client := newRunServiceTestClient(t, svc)
			tc.test(t, client)
		})
	}
}

// The console decides an action is a signalable condition from metadata.spec
// alone. The watch stream builds its own metadata rather than reusing
// actionMetadataFromModel, so the condition arm has to be kept in step: without
// it a paused condition streams as an ordinary action and the UI renders no way
// to signal it.
func TestConvertNodeUpdateToEnrichedProto_ConditionMetadata(t *testing.T) {
	svc := &RunService{}
	runID := &common.RunIdentifier{Project: "p", Domain: "d", Name: "r1"}

	condType := &core.LiteralType{
		Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
	}
	specBytes, err := proto.Marshal(&workflow.ActionSpec{
		Spec: &workflow.ActionSpec_Condition{
			Condition: &workflow.ConditionAction{
				Name: "approve-item",
				Type: condType,
			},
		},
	})
	require.NoError(t, err)

	enriched := svc.convertNodeUpdateToEnrichedProto(runID, &nodeUpdate{
		Node: &node{
			Action: &models.Action{
				Project:      "p",
				Domain:       "d",
				RunName:      "r1",
				Name:         "cond1",
				Phase:        int32(common.ActionPhase_ACTION_PHASE_PAUSED),
				ActionType:   int32(workflow.ActionType_ACTION_TYPE_CONDITION),
				FunctionName: "approve-item",
				ActionSpec:   specBytes,
			},
		},
		MeetsFilter: true,
	})

	require.NotNil(t, enriched)
	condMeta := enriched.GetAction().GetMetadata().GetCondition()
	require.NotNil(t, condMeta, "condition metadata must be set, otherwise the console cannot classify the action")
	assert.Equal(t, "approve-item", condMeta.GetName())
	// The declared type drives which input the console renders, and the payload
	// type check before signaling.
	assert.Equal(t, core.SimpleType_STRING, condMeta.GetType().GetSimple())
}

func TestWatchRuns_RelistsOnResync(t *testing.T) {
	actionRepo, _, svc := newTestService(t)
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
	stream, err := client.WatchRuns(ctx, &workflow.WatchRunsRequest{})
	require.NoError(t, err)

	msg14, receiveErr14 := stream.Receive()
	require.NoError(t, receiveErr14)
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_RUNNING, msg14.GetRuns()[0].GetAction().GetStatus().GetPhase())
	msg15, receiveErr15 := stream.Receive()
	require.NoError(t, receiveErr15)
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_SUCCEEDED, msg15.GetRuns()[0].GetAction().GetStatus().GetPhase())
}

func TestWatchActions_RelistsOnResync(t *testing.T) {
	actionRepo, _, svc := newTestService(t)
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
	stream, err := client.WatchActions(ctx, &workflow.WatchActionsRequest{RunId: runID})
	require.NoError(t, err)

	msg16, receiveErr16 := stream.Receive()
	require.NoError(t, receiveErr16)
	assert.Len(t, msg16.GetEnrichedActions(), 2)
	msg17, receiveErr17 := stream.Receive()
	require.NoError(t, receiveErr17)
	resent := msg17.GetEnrichedActions()
	require.Len(t, resent, 1, "the unchanged child must not be sent again")
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_SUCCEEDED, resent[0].GetAction().GetStatus().GetPhase())
}

func TestWatchActions_KeepsStreamOnResyncListError(t *testing.T) {
	actionRepo, _, svc := newTestService(t)
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
	stream, err := client.WatchActions(ctx, &workflow.WatchActionsRequest{RunId: runID})
	require.NoError(t, err)

	_, receiveErr18 := stream.Receive()
	require.NoError(t, receiveErr18)
	msg19, receiveErr19 := stream.Receive()
	require.NoError(t, receiveErr19)
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_SUCCEEDED, msg19.GetEnrichedActions()[0].GetAction().GetStatus().GetPhase())
}
