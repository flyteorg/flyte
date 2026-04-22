package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	actionsconnectmocks "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions/actionsconnect/mocks"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	repoMocks "github.com/flyteorg/flyte/v2/runs/repository/mocks"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

var testActionID = &common.ActionIdentifier{
	Run: &common.RunIdentifier{
		Org:     "test-org",
		Project: "test-project",
		Domain:  "test-domain",
		Name:    "rtest12345",
	},
	Name: "action-1",
}

func newTestServiceWithTaskRepo(t *testing.T) (*repoMocks.ActionRepo, *repoMocks.TaskRepo, *RunService) {
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

	return actionRepo, taskRepo, svc
}

func TestGetActionDetails_ActionNotFound(t *testing.T) {
	actionRepo, _, svc := newTestServiceWithTaskRepo(t)

	actionRepo.On("GetAction", mock.Anything, testActionID).Return(nil, errors.New("action not found"))

	_, err := svc.GetActionDetails(context.Background(), connect.NewRequest(&workflow.GetActionDetailsRequest{
		ActionId: testActionID,
	}))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "action not found")
}

func TestGetActionDetails_NoDetailedInfo(t *testing.T) {
	actionRepo, _, svc := newTestServiceWithTaskRepo(t)

	actionModel := &models.Action{
		Project: "test-project",
		Domain:  "test-domain",
		RunName: "rtest12345",
		Name:    "action-1",
		Phase:   int32(common.ActionPhase_ACTION_PHASE_RUNNING),
	}

	actionRepo.On("GetAction", mock.Anything, testActionID).Return(actionModel, nil)
	actionRepo.On("ListEvents", mock.Anything, testActionID, 500).Return([]*models.ActionEvent{}, nil)

	resp, err := svc.GetActionDetails(context.Background(), connect.NewRequest(&workflow.GetActionDetailsRequest{
		ActionId: testActionID,
	}))
	assert.NoError(t, err)
	assert.NotNil(t, resp.Msg.Details)
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_RUNNING, resp.Msg.Details.Status.Phase)
}

func TestGetActionDetails_WithTaskSpec(t *testing.T) {
	actionRepo, taskRepo, svc := newTestServiceWithTaskRepo(t)

	runInfo := &workflow.RunInfo{
		TaskSpecDigest: "abc123",
		InputsUri:      "s3://bucket/inputs.pb",
	}
	runInfoBytes, _ := proto.Marshal(runInfo)

	actionModel := &models.Action{
		Project:      "test-project",
		Domain:       "test-domain",
		RunName:      "rtest12345",
		Name:         "action-1",
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

	actionRepo.On("GetAction", mock.Anything, testActionID).Return(actionModel, nil)
	taskRepo.On("GetTaskSpec", mock.Anything, "abc123").Return(&models.TaskSpec{
		Digest: "abc123",
		Spec:   taskSpecBytes,
	}, nil)
	actionRepo.On("ListEvents", mock.Anything, testActionID, 500).Return([]*models.ActionEvent{}, nil)

	resp, err := svc.GetActionDetails(context.Background(), connect.NewRequest(&workflow.GetActionDetailsRequest{
		ActionId: testActionID,
	}))
	assert.NoError(t, err)
	assert.NotNil(t, resp.Msg.Details)
	assert.NotNil(t, resp.Msg.Details.GetTask())
	assert.Equal(t, "python", resp.Msg.Details.GetTask().GetTaskTemplate().GetType())
}

func TestGetActionDetails_WithTraceSpec(t *testing.T) {
	actionRepo, taskRepo, svc := newTestServiceWithTaskRepo(t)

	runInfo := &workflow.RunInfo{
		TaskSpecDigest: "trace-digest",
	}
	runInfoBytes, _ := proto.Marshal(runInfo)

	actionModel := &models.Action{
		Project:      "test-project",
		Domain:       "test-domain",
		RunName:      "rtest12345",
		Name:         "action-1",
		Phase:        int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED),
		ActionType:   int32(workflow.ActionType_ACTION_TYPE_TRACE),
		DetailedInfo: runInfoBytes,
	}

	traceSpecProto := &task.TraceSpec{
		Interface: &core.TypedInterface{},
	}
	traceSpecBytes, _ := proto.Marshal(traceSpecProto)

	actionRepo.On("GetAction", mock.Anything, testActionID).Return(actionModel, nil)
	taskRepo.On("GetTaskSpec", mock.Anything, "trace-digest").Return(&models.TaskSpec{
		Digest: "trace-digest",
		Spec:   traceSpecBytes,
	}, nil)
	actionRepo.On("ListEvents", mock.Anything, testActionID, 500).Return([]*models.ActionEvent{}, nil)

	resp, err := svc.GetActionDetails(context.Background(), connect.NewRequest(&workflow.GetActionDetailsRequest{
		ActionId: testActionID,
	}))
	assert.NoError(t, err)
	assert.NotNil(t, resp.Msg.Details.GetTrace())
}

func TestGetActionDetails_TaskSpecLookupFails(t *testing.T) {
	actionRepo, taskRepo, svc := newTestServiceWithTaskRepo(t)

	runInfo := &workflow.RunInfo{TaskSpecDigest: "bad-digest"}
	runInfoBytes, _ := proto.Marshal(runInfo)

	actionModel := &models.Action{
		Project:      "test-project",
		Domain:       "test-domain",
		RunName:      "rtest12345",
		Name:         "action-1",
		Phase:        int32(common.ActionPhase_ACTION_PHASE_RUNNING),
		ActionType:   int32(workflow.ActionType_ACTION_TYPE_TASK),
		DetailedInfo: runInfoBytes,
	}

	actionRepo.On("GetAction", mock.Anything, testActionID).Return(actionModel, nil)
	taskRepo.On("GetTaskSpec", mock.Anything, "bad-digest").Return(nil, errors.New("not found"))
	actionRepo.On("ListEvents", mock.Anything, testActionID, 500).Return([]*models.ActionEvent{}, nil)

	_, err := svc.GetActionDetails(context.Background(), connect.NewRequest(&workflow.GetActionDetailsRequest{
		ActionId: testActionID,
	}))
	assert.Error(t, err)
}

func TestGetActionDetails_FailedActionSetsErrorInfo(t *testing.T) {
	actionRepo, _, svc := newTestServiceWithTaskRepo(t)

	actionModel := &models.Action{
		Project:  "test-project",
		Domain:   "test-domain",
		RunName:  "rtest12345",
		Name:     "action-1",
		Phase:    int32(common.ActionPhase_ACTION_PHASE_FAILED),
		Attempts: 1,
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
	failedEventModel, _ := models.NewActionEventModel(failedEvent)

	actionRepo.On("GetAction", mock.Anything, testActionID).Return(actionModel, nil)
	actionRepo.On("ListEvents", mock.Anything, testActionID, 500).Return([]*models.ActionEvent{failedEventModel}, nil)

	resp, err := svc.GetActionDetails(context.Background(), connect.NewRequest(&workflow.GetActionDetailsRequest{
		ActionId: testActionID,
	}))
	assert.NoError(t, err)
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_FAILED, resp.Msg.Details.Status.Phase)
	assert.NotNil(t, resp.Msg.Details.GetErrorInfo())
	assert.Equal(t, "task failed with OOM", resp.Msg.Details.GetErrorInfo().GetMessage())
}

func TestGetActionDetails_WithAttempts(t *testing.T) {
	actionRepo, _, svc := newTestServiceWithTaskRepo(t)

	actionModel := &models.Action{
		Project: "test-project",
		Domain:  "test-domain",
		RunName: "rtest12345",
		Name:    "action-1",
		Phase:   int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED),
	}

	now := time.Now()

	// Attempt 0: failed
	event1 := &workflow.ActionEvent{
		Id:          testActionID,
		Attempt:     0,
		Phase:       common.ActionPhase_ACTION_PHASE_RUNNING,
		Version:     0,
		UpdatedTime: timestamppb.New(now.Add(-3 * time.Minute)),
	}
	event2 := &workflow.ActionEvent{
		Id:      testActionID,
		Attempt: 0,
		Phase:   common.ActionPhase_ACTION_PHASE_FAILED,
		Version: 1,
		ErrorInfo: &workflow.ErrorInfo{
			Message: "timeout",
			Kind:    workflow.ErrorInfo_KIND_SYSTEM,
		},
		UpdatedTime: timestamppb.New(now.Add(-2 * time.Minute)),
	}

	// Attempt 1: succeeded
	event3 := &workflow.ActionEvent{
		Id:          testActionID,
		Attempt:     1,
		Phase:       common.ActionPhase_ACTION_PHASE_RUNNING,
		Version:     0,
		UpdatedTime: timestamppb.New(now.Add(-1 * time.Minute)),
	}
	event4 := &workflow.ActionEvent{
		Id:          testActionID,
		Attempt:     1,
		Phase:       common.ActionPhase_ACTION_PHASE_SUCCEEDED,
		Version:     1,
		UpdatedTime: timestamppb.New(now),
	}

	eventModels := make([]*models.ActionEvent, 0, 4)
	for _, e := range []*workflow.ActionEvent{event1, event2, event3, event4} {
		m, _ := models.NewActionEventModel(e)
		eventModels = append(eventModels, m)
	}

	actionRepo.On("GetAction", mock.Anything, testActionID).Return(actionModel, nil)
	actionRepo.On("ListEvents", mock.Anything, testActionID, 500).Return(eventModels, nil)

	resp, err := svc.GetActionDetails(context.Background(), connect.NewRequest(&workflow.GetActionDetailsRequest{
		ActionId: testActionID,
	}))
	assert.NoError(t, err)

	attempts := resp.Msg.Details.Attempts
	assert.Equal(t, 2, len(attempts))
	assert.Equal(t, uint32(0), attempts[0].Attempt)
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_FAILED, attempts[0].Phase)
	assert.Equal(t, uint32(1), attempts[1].Attempt)
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_SUCCEEDED, attempts[1].Phase)
}

func TestIsTerminalPhase(t *testing.T) {
	assert.True(t, IsTerminalPhase(common.ActionPhase_ACTION_PHASE_FAILED))
	assert.True(t, IsTerminalPhase(common.ActionPhase_ACTION_PHASE_SUCCEEDED))
	assert.True(t, IsTerminalPhase(common.ActionPhase_ACTION_PHASE_TIMED_OUT))
	assert.True(t, IsTerminalPhase(common.ActionPhase_ACTION_PHASE_ABORTED))

	assert.False(t, IsTerminalPhase(common.ActionPhase_ACTION_PHASE_UNSPECIFIED))
	assert.False(t, IsTerminalPhase(common.ActionPhase_ACTION_PHASE_QUEUED))
	assert.False(t, IsTerminalPhase(common.ActionPhase_ACTION_PHASE_RUNNING))
	assert.False(t, IsTerminalPhase(common.ActionPhase_ACTION_PHASE_WAITING_FOR_RESOURCES))
}

func TestMergeEvents_Empty(t *testing.T) {
	result := mergeEvents(0, []*workflow.ActionEvent{})
	assert.Equal(t, uint32(0), result.Attempt)
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_UNSPECIFIED, result.Phase)
}

func TestMergeEvents_SingleEvent(t *testing.T) {
	now := time.Now()
	events := []*workflow.ActionEvent{
		{
			Phase:       common.ActionPhase_ACTION_PHASE_RUNNING,
			UpdatedTime: timestamppb.New(now),
		},
	}

	result := mergeEvents(0, events)
	assert.Equal(t, uint32(0), result.Attempt)
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_RUNNING, result.Phase)
	assert.Nil(t, result.EndTime)
}

func TestMergeEvents_PhaseTransitions(t *testing.T) {
	now := time.Now()
	events := []*workflow.ActionEvent{
		{
			Phase:       common.ActionPhase_ACTION_PHASE_QUEUED,
			UpdatedTime: timestamppb.New(now),
		},
		{
			Phase:       common.ActionPhase_ACTION_PHASE_RUNNING,
			UpdatedTime: timestamppb.New(now.Add(1 * time.Second)),
		},
		{
			Phase:       common.ActionPhase_ACTION_PHASE_SUCCEEDED,
			UpdatedTime: timestamppb.New(now.Add(2 * time.Second)),
		},
	}

	result := mergeEvents(1, events)
	assert.Equal(t, uint32(1), result.Attempt)
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_SUCCEEDED, result.Phase)
	assert.NotNil(t, result.EndTime)
	assert.Equal(t, 3, len(result.PhaseTransitions))
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_QUEUED, result.PhaseTransitions[0].Phase)
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_RUNNING, result.PhaseTransitions[1].Phase)
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_SUCCEEDED, result.PhaseTransitions[2].Phase)
}

func TestMergeEvents_SortsByReportedTime(t *testing.T) {
	now := time.Now()
	// Events arrive with different ReportedTime and UpdatedTime.
	// mergeEvents sorts by ReportedTime when both events have it.
	events := []*workflow.ActionEvent{
		{
			Phase:        common.ActionPhase_ACTION_PHASE_QUEUED,
			UpdatedTime:  timestamppb.New(now),
			ReportedTime: timestamppb.New(now),
		},
		{
			Phase:        common.ActionPhase_ACTION_PHASE_RUNNING,
			UpdatedTime:  timestamppb.New(now.Add(1 * time.Second)),
			ReportedTime: timestamppb.New(now.Add(1 * time.Second)),
		},
		{
			Phase:        common.ActionPhase_ACTION_PHASE_SUCCEEDED,
			UpdatedTime:  timestamppb.New(now.Add(2 * time.Second)),
			ReportedTime: timestamppb.New(now.Add(2 * time.Second)),
		},
	}

	result := mergeEvents(1, events)
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_SUCCEEDED, result.Phase)
	assert.Equal(t, 3, len(result.PhaseTransitions))
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_QUEUED, result.PhaseTransitions[0].Phase)
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_RUNNING, result.PhaseTransitions[1].Phase)
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_SUCCEEDED, result.PhaseTransitions[2].Phase)
	assert.True(t, result.PhaseTransitions[0].GetStartTime().AsTime().Equal(now))
	assert.True(t, result.PhaseTransitions[1].GetStartTime().AsTime().Equal(now.Add(1*time.Second)))
}

func TestMergeEvents_MergesLogs(t *testing.T) {
	now := time.Now()
	events := []*workflow.ActionEvent{
		{
			Phase:       common.ActionPhase_ACTION_PHASE_RUNNING,
			UpdatedTime: timestamppb.New(now),
			LogInfo: []*core.TaskLog{
				{Name: "log-a", Uri: "s3://logs/a"},
			},
		},
		{
			Phase:       common.ActionPhase_ACTION_PHASE_RUNNING,
			UpdatedTime: timestamppb.New(now.Add(1 * time.Second)),
			Version:     1,
			LogInfo: []*core.TaskLog{
				{Name: "log-b", Uri: "s3://logs/b"},
			},
		},
	}

	result := mergeEvents(0, events)
	assert.Equal(t, 2, len(result.LogInfo))
}

func TestMergeEvents_SetsErrorInfo(t *testing.T) {
	now := time.Now()
	events := []*workflow.ActionEvent{
		{
			Phase:       common.ActionPhase_ACTION_PHASE_RUNNING,
			UpdatedTime: timestamppb.New(now),
		},
		{
			Phase:       common.ActionPhase_ACTION_PHASE_FAILED,
			UpdatedTime: timestamppb.New(now.Add(1 * time.Second)),
			ErrorInfo: &workflow.ErrorInfo{
				Message: "out of memory",
				Kind:    workflow.ErrorInfo_KIND_SYSTEM,
			},
		},
	}

	result := mergeEvents(0, events)
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_FAILED, result.Phase)
	assert.NotNil(t, result.ErrorInfo)
	assert.Equal(t, "out of memory", result.ErrorInfo.GetMessage())
}

func TestMergeEvents_ClusterEvents(t *testing.T) {
	now := time.Now()
	events := []*workflow.ActionEvent{
		{
			Phase:       common.ActionPhase_ACTION_PHASE_RUNNING,
			UpdatedTime: timestamppb.New(now),
			ClusterEvents: []*workflow.ClusterEvent{
				{Message: "Scheduled"},
			},
		},
		{
			Phase:       common.ActionPhase_ACTION_PHASE_RUNNING,
			UpdatedTime: timestamppb.New(now.Add(1 * time.Second)),
			Version:     1,
			ClusterEvents: []*workflow.ClusterEvent{
				{Message: "Pulling image"},
			},
		},
	}

	result := mergeEvents(0, events)
	assert.Equal(t, 2, len(result.ClusterEvents))
}

func TestMergeEvents_LogContext(t *testing.T) {
	now := time.Now()
	events := []*workflow.ActionEvent{
		{
			Phase:       common.ActionPhase_ACTION_PHASE_RUNNING,
			UpdatedTime: timestamppb.New(now),
		},
		{
			Phase:       common.ActionPhase_ACTION_PHASE_RUNNING,
			UpdatedTime: timestamppb.New(now.Add(1 * time.Second)),
			Version:     1,
			LogContext: &core.LogContext{
				PrimaryPodName: "my-pod",
			},
		},
	}

	result := mergeEvents(0, events)
	assert.True(t, result.LogsAvailable)
	assert.Equal(t, "my-pod", result.LogContext.GetPrimaryPodName())
}

func TestMergeEvents_ReportURI(t *testing.T) {
	now := time.Now()
	events := []*workflow.ActionEvent{
		{
			Phase:       common.ActionPhase_ACTION_PHASE_RUNNING,
			UpdatedTime: timestamppb.New(now),
		},
		{
			Phase:       common.ActionPhase_ACTION_PHASE_SUCCEEDED,
			UpdatedTime: timestamppb.New(now.Add(1 * time.Second)),
			Outputs: &task.OutputReferences{
				ReportUri: "s3://reports/report.html",
			},
		},
	}

	result := mergeEvents(0, events)
	assert.Equal(t, "s3://reports/report.html", result.GetOutputs().GetReportUri())
}

func TestBuildActionDetails_CanceledContextSuppressesErrors(t *testing.T) {
	actionRepo, taskRepo, svc := newTestServiceWithTaskRepo(t)

	runInfo := &workflow.RunInfo{TaskSpecDigest: "some-digest"}
	runInfoBytes, _ := proto.Marshal(runInfo)

	actionModel := &models.Action{
		Project:      "test-project",
		Domain:       "test-domain",
		RunName:      "rtest12345",
		Name:         "action-1",
		Phase:        int32(common.ActionPhase_ACTION_PHASE_RUNNING),
		ActionType:   int32(workflow.ActionType_ACTION_TYPE_TASK),
		DetailedInfo: runInfoBytes,
	}

	// Cancel context before calling buildActionDetails
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Both task spec and events queries will fail due to canceled context
	taskRepo.On("GetTaskSpec", mock.Anything, "some-digest").Return(nil, context.Canceled)
	actionRepo.On("ListEvents", mock.Anything, testActionID, 500).Return(nil, context.Canceled)

	_, err := svc.buildActionDetails(ctx, actionModel, testActionID)
	// Should return an error but not panic or log excessively
	assert.Error(t, err)
}

func TestGetActionDetails_SplitIntoGetAndBuild(t *testing.T) {
	// Verify that getActionDetails calls GetAction then buildActionDetails
	actionRepo, _, svc := newTestServiceWithTaskRepo(t)

	actionModel := &models.Action{
		Project: "test-project",
		Domain:  "test-domain",
		RunName: "rtest12345",
		Name:    "action-1",
		Phase:   int32(common.ActionPhase_ACTION_PHASE_RUNNING),
	}

	actionRepo.On("GetAction", mock.Anything, testActionID).Return(actionModel, nil)
	actionRepo.On("ListEvents", mock.Anything, testActionID, 500).Return([]*models.ActionEvent{}, nil)

	details, err := svc.getActionDetails(context.Background(), testActionID)
	assert.NoError(t, err)
	assert.NotNil(t, details)
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_RUNNING, details.Status.Phase)
}

func TestMergeEvents_EndTimeNeverBeforeStartTime(t *testing.T) {
	earlyTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	lateTime := time.Date(2026, 1, 1, 0, 0, 10, 0, time.UTC)

	events := []*workflow.ActionEvent{
		{
			Phase:       common.ActionPhase_ACTION_PHASE_RUNNING,
			UpdatedTime: timestamppb.New(lateTime),
		},
		{
			Phase:       common.ActionPhase_ACTION_PHASE_SUCCEEDED,
			UpdatedTime: timestamppb.New(earlyTime),
		},
	}

	result := mergeEvents(0, events)
	// endTime should be clamped to startTime
	assert.Equal(t, result.StartTime.AsTime(), result.EndTime.AsTime())
}
