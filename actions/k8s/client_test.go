package k8s

import (
	"context"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	executorv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/flytestdlib/fastcheck"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	runmocks "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect/mocks"
)

func newTestActionUpdate(actionName string) (*executorv1.TaskAction, *ActionUpdate) {
	runID := &common.RunIdentifier{
		Org:     "org",
		Project: "proj",
		Domain:  "dev",
		Name:    "run1",
	}
	ta := &executorv1.TaskAction{
		Spec: executorv1.TaskActionSpec{
			Org:        runID.Org,
			Project:    runID.Project,
			Domain:     runID.Domain,
			RunName:    runID.Name,
			ActionName: actionName,
		},
	}
	update := &ActionUpdate{
		ActionID: &common.ActionIdentifier{
			Run:  runID,
			Name: actionName,
		},
	}
	return ta, update
}

func TestNotifyRunService_DeduplicateRecordAction(t *testing.T) {
	ctx := context.Background()

	mockClient := runmocks.NewInternalRunServiceClient(t)

	filter, err := fastcheck.NewOppoBloomFilter(128, promutils.NewTestScope())
	assert.NoError(t, err)

	c := &ActionsClient{
		runClient:      mockClient,
		recordedFilter: filter,
		subscribers:    make(map[string]map[chan *ActionUpdate]struct{}),
	}

	ta, update := newTestActionUpdate("action-1")

	// Expect RecordAction called exactly once
	mockClient.On("RecordAction", mock.Anything, mock.Anything).
		Return(&connect.Response[workflow.RecordActionResponse]{}, nil).Once()

	// First Added event — should call RecordAction
	c.notifyRunService(ctx, ta, update, watch.Added)

	// Second Added event (replay) — should be deduplicated
	c.notifyRunService(ctx, ta, update, watch.Added)

	mockClient.AssertNumberOfCalls(t, "RecordAction", 1)
}

func TestNotifyRunService_FailedRecordAllowsRetry(t *testing.T) {
	ctx := context.Background()

	mockClient := runmocks.NewInternalRunServiceClient(t)

	filter, err := fastcheck.NewOppoBloomFilter(128, promutils.NewTestScope())
	assert.NoError(t, err)

	c := &ActionsClient{
		runClient:      mockClient,
		recordedFilter: filter,
		subscribers:    make(map[string]map[chan *ActionUpdate]struct{}),
	}

	ta, update := newTestActionUpdate("action-2")

	// First call fails
	mockClient.On("RecordAction", mock.Anything, mock.Anything).
		Return((*connect.Response[workflow.RecordActionResponse])(nil), fmt.Errorf("transient error")).Once()
	// Second call succeeds
	mockClient.On("RecordAction", mock.Anything, mock.Anything).
		Return(&connect.Response[workflow.RecordActionResponse]{}, nil).Once()

	// First event — RecordAction fails, should NOT add to filter
	c.notifyRunService(ctx, ta, update, watch.Added)

	// Second event — should retry RecordAction since first failed
	c.notifyRunService(ctx, ta, update, watch.Added)

	mockClient.AssertNumberOfCalls(t, "RecordAction", 2)

	// Third event — now it's in the filter, should be skipped
	c.notifyRunService(ctx, ta, update, watch.Added)
	mockClient.AssertNumberOfCalls(t, "RecordAction", 2)
}

func TestNotifyRunService_NilFilter(t *testing.T) {
	ctx := context.Background()

	mockClient := runmocks.NewInternalRunServiceClient(t)

	// No filter — should always call RecordAction
	c := &ActionsClient{
		runClient:   mockClient,
		subscribers: make(map[string]map[chan *ActionUpdate]struct{}),
	}

	ta, update := newTestActionUpdate("action-3")

	mockClient.On("RecordAction", mock.Anything, mock.Anything).
		Return(&connect.Response[workflow.RecordActionResponse]{}, nil)

	c.notifyRunService(ctx, ta, update, watch.Added)
	c.notifyRunService(ctx, ta, update, watch.Added)

	mockClient.AssertNumberOfCalls(t, "RecordAction", 2)
}

func TestNotifyRunService_UpdateActionStatusIncludesAttemptsAndCacheStatus(t *testing.T) {
	ctx := context.Background()

	mockClient := runmocks.NewInternalRunServiceClient(t)
	c := &ActionsClient{
		runClient:   mockClient,
		subscribers: make(map[string]map[chan *ActionUpdate]struct{}),
	}

	ta, update := newTestActionUpdate("action-4")
	ta.Status.Attempts = 3
	ta.Status.CacheStatus = core.CatalogCacheStatus_CACHE_HIT
	update.Phase = common.ActionPhase_ACTION_PHASE_SUCCEEDED

	mockClient.On("UpdateActionStatus", mock.Anything, mock.MatchedBy(func(req *connect.Request[workflow.UpdateActionStatusRequest]) bool {
		status := req.Msg.GetStatus()
		return status.GetPhase() == common.ActionPhase_ACTION_PHASE_SUCCEEDED &&
			status.GetAttempts() == 3 &&
			status.GetCacheStatus() == core.CatalogCacheStatus_CACHE_HIT
	})).Return(&connect.Response[workflow.UpdateActionStatusResponse]{}, nil).Once()

	c.notifyRunService(ctx, ta, update, watch.Modified)

	mockClient.AssertNumberOfCalls(t, "UpdateActionStatus", 1)
}

func TestNotifyRunService_UpdateActionStatusIncludesEndTime(t *testing.T) {
	ctx := context.Background()

	mockClient := runmocks.NewInternalRunServiceClient(t)
	c := &ActionsClient{
		runClient:   mockClient,
		subscribers: make(map[string]map[chan *ActionUpdate]struct{}),
	}

	completionTime := metav1.Now()
	ta, update := newTestActionUpdate("action-endtime")
	ta.Status.Attempts = 1
	ta.Status.PhaseHistory = []executorv1.PhaseTransition{
		{Phase: "Queued", OccurredAt: metav1.NewTime(completionTime.Add(-30 * 1e9))},
		{Phase: "Executing", OccurredAt: metav1.NewTime(completionTime.Add(-20 * 1e9))},
		{Phase: string(executorv1.ConditionReasonCompleted), OccurredAt: completionTime},
	}
	update.Phase = common.ActionPhase_ACTION_PHASE_SUCCEEDED

	mockClient.On("UpdateActionStatus", mock.Anything, mock.MatchedBy(func(req *connect.Request[workflow.UpdateActionStatusRequest]) bool {
		status := req.Msg.GetStatus()
		return status.GetPhase() == common.ActionPhase_ACTION_PHASE_SUCCEEDED &&
			status.GetEndTime() != nil &&
			status.GetEndTime().AsTime().Equal(completionTime.Time)
	})).Return(&connect.Response[workflow.UpdateActionStatusResponse]{}, nil).Once()

	c.notifyRunService(ctx, ta, update, watch.Modified)

	mockClient.AssertNumberOfCalls(t, "UpdateActionStatus", 1)
}

func TestBuildTaskActionName(t *testing.T) {
	runID := &common.RunIdentifier{
		Org:     "org",
		Project: "project",
		Domain:  "development",
		Name:    "rabc123",
	}

	t.Run("root action uses a0 suffix", func(t *testing.T) {
		// Root: action name == run name
		actionID := &common.ActionIdentifier{
			Run:  runID,
			Name: runID.Name,
		}
		assert.Equal(t, "rabc123-a0", buildTaskActionName(actionID))
	})

	t.Run("child action includes action name", func(t *testing.T) {
		actionID := &common.ActionIdentifier{
			Run:  runID,
			Name: "train",
		}
		assert.Equal(t, "rabc123-train", buildTaskActionName(actionID))
	})
}

func TestBuildNamespace(t *testing.T) {
	t.Run("combines project and domain", func(t *testing.T) {
		runID := &common.RunIdentifier{
			Project: "flytesnacks",
			Domain:  "development",
		}
		assert.Equal(t, "flytesnacks-development", buildNamespace(runID))
	})

	t.Run("different project and domain", func(t *testing.T) {
		runID := &common.RunIdentifier{
			Project: "myproject",
			Domain:  "production",
		}
		assert.Equal(t, "myproject-production", buildNamespace(runID))
	})
}

func TestExtractTaskCacheKey(t *testing.T) {
	t.Run("returns cache key for task action", func(t *testing.T) {
		action := &actions.Action{
			Spec: &actions.Action_Task{
				Task: &workflow.TaskAction{
					CacheKey: wrapperspb.String("cache-v1"),
				},
			},
		}

		assert.Equal(t, "cache-v1", extractTaskCacheKey(action))
	})

	t.Run("returns empty for non-task action", func(t *testing.T) {
		assert.Empty(t, extractTaskCacheKey(&actions.Action{}))
	})
}

func TestApplyRunSpecToTaskAction_ProjectsRuntimeSettings(t *testing.T) {
	taskAction := &executorv1.TaskAction{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"flyte.org/run": "run1",
			},
		},
		Spec: executorv1.TaskActionSpec{},
	}

	applyRunSpecToTaskAction(taskAction, &task.RunSpec{
		Envs: &task.Envs{
			Values: []*core.KeyValuePair{
				{Key: "TRACE_ID", Value: "abc123"},
			},
		},
		Interruptible: wrapperspb.Bool(true),
		Labels: &task.Labels{
			Values: map[string]string{
				"flyte.org/run": "should-not-override",
				"team":          "platform",
			},
		},
		Annotations: &task.Annotations{
			Values: map[string]string{
				"owner": "sdk",
			},
		},
	})

	require.NotNil(t, taskAction.Spec.Interruptible)
	assert.True(t, *taskAction.Spec.Interruptible)
	assert.Len(t, taskAction.Spec.EnvVars, 1)
	assert.Equal(t, "abc123", taskAction.Spec.EnvVars["TRACE_ID"])
	assert.Equal(t, "run1", taskAction.Labels["flyte.org/run"])
	assert.Equal(t, "platform", taskAction.Labels["team"])
	assert.Equal(t, "sdk", taskAction.Annotations["owner"])
}

func TestNotifyRunService_ChildAddedPromotesParentToRunning(t *testing.T) {
	ctx := context.Background()

	mockClient := runmocks.NewInternalRunServiceClient(t)
	c := &ActionsClient{
		runClient:   mockClient,
		subscribers: make(map[string]map[chan *ActionUpdate]struct{}),
	}

	runID := &common.RunIdentifier{
		Org:     "org",
		Project: "proj",
		Domain:  "dev",
		Name:    "run1",
	}

	ta := &executorv1.TaskAction{
		Spec: executorv1.TaskActionSpec{
			Org:        runID.Org,
			Project:    runID.Project,
			Domain:     runID.Domain,
			RunName:    runID.Name,
			ActionName: "child-1",
		},
	}
	update := &ActionUpdate{
		ActionID: &common.ActionIdentifier{
			Run:  runID,
			Name: "child-1",
		},
		ParentActionName: "run1",
	}

	// Expect RecordAction for the child
	mockClient.On("RecordAction", mock.Anything, mock.Anything).
		Return(&connect.Response[workflow.RecordActionResponse]{}, nil).Once()

	// Expect UpdateActionStatus for the PARENT with RUNNING phase
	mockClient.On("UpdateActionStatus", mock.Anything, mock.MatchedBy(func(req *connect.Request[workflow.UpdateActionStatusRequest]) bool {
		return req.Msg.GetActionId().GetName() == "run1" &&
			req.Msg.GetStatus().GetPhase() == common.ActionPhase_ACTION_PHASE_RUNNING
	})).Return(&connect.Response[workflow.UpdateActionStatusResponse]{}, nil).Once()

	c.notifyRunService(ctx, ta, update, watch.Added)

	mockClient.AssertNumberOfCalls(t, "RecordAction", 1)
	mockClient.AssertNumberOfCalls(t, "UpdateActionStatus", 1)
}

func TestNotifyRunService_SkipsTerminalAddedEventsOnlyWhenInBloomFilter(t *testing.T) {
	ctx := context.Background()

	mockClient := runmocks.NewInternalRunServiceClient(t)
	filter, err := fastcheck.NewOppoBloomFilter(128, promutils.NewTestScope())
	require.NoError(t, err)

	c := &ActionsClient{
		runClient:      mockClient,
		recordedFilter: filter,
		subscribers:    make(map[string]map[chan *ActionUpdate]struct{}),
	}

	// Create a terminal TaskAction CRD (SUCCEEDED)
	ta := &executorv1.TaskAction{
		Spec: executorv1.TaskActionSpec{
			Org:        "org",
			Project:    "proj",
			Domain:     "dev",
			RunName:    "run1",
			ActionName: "completed-action",
		},
		Status: executorv1.TaskActionStatus{
			Conditions: []metav1.Condition{
				{Type: string(executorv1.ConditionTypeSucceeded), Status: "True"},
			},
		},
	}
	update := &ActionUpdate{
		ActionID: &common.ActionIdentifier{
			Run:  &common.RunIdentifier{Org: "org", Project: "proj", Domain: "dev", Name: "run1"},
			Name: "completed-action",
		},
		Phase: common.ActionPhase_ACTION_PHASE_SUCCEEDED,
	}

	// First ADDED event (cold start, not in bloom filter): should process normally
	mockClient.On("RecordAction", mock.Anything, mock.Anything).
		Return(&connect.Response[workflow.RecordActionResponse]{}, nil).Once()
	mockClient.On("UpdateActionStatus", mock.Anything, mock.Anything).
		Return(&connect.Response[workflow.UpdateActionStatusResponse]{}, nil)
	c.notifyRunService(ctx, ta, update, watch.Added)

	mockClient.AssertNumberOfCalls(t, "RecordAction", 1)
	mockClient.AssertNumberOfCalls(t, "UpdateActionStatus", 1)

	// Action should now be in the bloom filter
	actionKey := []byte(buildTaskActionName(update.ActionID))
	assert.True(t, filter.Contains(ctx, actionKey))

	// Second ADDED event (reconnect, in bloom filter): should skip RecordAction
	// but still call UpdateActionStatus.
	c.notifyRunService(ctx, ta, update, watch.Added)
	mockClient.AssertNumberOfCalls(t, "RecordAction", 1)       // no new RecordAction
	mockClient.AssertNumberOfCalls(t, "UpdateActionStatus", 2) // one more UpdateActionStatus
}

func TestNotifyRunService_ProcessesNonTerminalAddedEvents(t *testing.T) {
	ctx := context.Background()

	mockClient := runmocks.NewInternalRunServiceClient(t)
	filter, err := fastcheck.NewOppoBloomFilter(128, promutils.NewTestScope())
	require.NoError(t, err)

	c := &ActionsClient{
		runClient:      mockClient,
		recordedFilter: filter,
		subscribers:    make(map[string]map[chan *ActionUpdate]struct{}),
	}

	// Create a non-terminal TaskAction CRD (QUEUED)
	ta := &executorv1.TaskAction{
		Spec: executorv1.TaskActionSpec{
			Org:        "org",
			Project:    "proj",
			Domain:     "dev",
			RunName:    "run1",
			ActionName: "queued-action",
		},
		Status: executorv1.TaskActionStatus{
			Conditions: []metav1.Condition{
				{Type: string(executorv1.ConditionTypeProgressing), Status: "True", Reason: string(executorv1.ConditionReasonQueued)},
			},
		},
	}
	update := &ActionUpdate{
		ActionID: &common.ActionIdentifier{
			Run:  &common.RunIdentifier{Org: "org", Project: "proj", Domain: "dev", Name: "run1"},
			Name: "queued-action",
		},
		Phase: common.ActionPhase_ACTION_PHASE_QUEUED,
	}

	// Non-terminal ADDED events should be processed normally
	mockClient.On("RecordAction", mock.Anything, mock.Anything).
		Return(&connect.Response[workflow.RecordActionResponse]{}, nil).Once()
	mockClient.On("UpdateActionStatus", mock.Anything, mock.Anything).
		Return(&connect.Response[workflow.UpdateActionStatusResponse]{}, nil).Once()

	c.notifyRunService(ctx, ta, update, watch.Added)

	mockClient.AssertNumberOfCalls(t, "RecordAction", 1)
	mockClient.AssertNumberOfCalls(t, "UpdateActionStatus", 1)
}

func TestNotifyRunService_DuplicateAddedSkipsRecordAction(t *testing.T) {
	ctx := context.Background()

	mockClient := runmocks.NewInternalRunServiceClient(t)
	filter, err := fastcheck.NewOppoBloomFilter(128, promutils.NewTestScope())
	require.NoError(t, err)

	c := &ActionsClient{
		runClient:      mockClient,
		recordedFilter: filter,
		subscribers:    make(map[string]map[chan *ActionUpdate]struct{}),
	}

	ta, update := newTestActionUpdate("action-dup")
	update.Phase = common.ActionPhase_ACTION_PHASE_RUNNING

	// First call — should process normally
	mockClient.On("RecordAction", mock.Anything, mock.Anything).
		Return(&connect.Response[workflow.RecordActionResponse]{}, nil).Once()
	mockClient.On("UpdateActionStatus", mock.Anything, mock.Anything).
		Return(&connect.Response[workflow.UpdateActionStatusResponse]{}, nil)
	c.notifyRunService(ctx, ta, update, watch.Added)

	// Second call (duplicate ADDED) — should skip RecordAction but still call UpdateActionStatus
	c.notifyRunService(ctx, ta, update, watch.Added)

	mockClient.AssertNumberOfCalls(t, "RecordAction", 1)
	mockClient.AssertNumberOfCalls(t, "UpdateActionStatus", 2)
}

func TestNotifyRunService_TerminalDuplicateRepairsTimestamps(t *testing.T) {
	ctx := context.Background()

	mockClient := runmocks.NewInternalRunServiceClient(t)
	filter, err := fastcheck.NewOppoBloomFilter(128, promutils.NewTestScope())
	require.NoError(t, err)

	c := &ActionsClient{
		runClient:      mockClient,
		recordedFilter: filter,
		subscribers:    make(map[string]map[chan *ActionUpdate]struct{}),
	}

	ta, update := newTestActionUpdate("action-terminal-dup")
	update.Phase = common.ActionPhase_ACTION_PHASE_SUCCEEDED

	// First call — should process normally (RecordAction + UpdateActionStatus)
	mockClient.On("RecordAction", mock.Anything, mock.Anything).
		Return(&connect.Response[workflow.RecordActionResponse]{}, nil).Once()
	mockClient.On("UpdateActionStatus", mock.Anything, mock.Anything).
		Return(&connect.Response[workflow.UpdateActionStatusResponse]{}, nil).Times(2)
	c.notifyRunService(ctx, ta, update, watch.Added)

	// Second call (terminal duplicate ADDED) — should skip RecordAction but
	// still call UpdateActionStatus to repair missing timestamps.
	c.notifyRunService(ctx, ta, update, watch.Added)

	mockClient.AssertNumberOfCalls(t, "RecordAction", 1)       // no new RecordAction
	mockClient.AssertNumberOfCalls(t, "UpdateActionStatus", 2) // one more UpdateActionStatus
}

func TestNotifyRunService_RootActionAddedDoesNotPromoteParent(t *testing.T) {
	ctx := context.Background()

	mockClient := runmocks.NewInternalRunServiceClient(t)
	c := &ActionsClient{
		runClient:   mockClient,
		subscribers: make(map[string]map[chan *ActionUpdate]struct{}),
	}

	// Root action has no parent
	ta, update := newTestActionUpdate("action-root")

	mockClient.On("RecordAction", mock.Anything, mock.Anything).
		Return(&connect.Response[workflow.RecordActionResponse]{}, nil).Once()

	c.notifyRunService(ctx, ta, update, watch.Added)

	mockClient.AssertNumberOfCalls(t, "RecordAction", 1)
	// No UpdateActionStatus should be called for root (no parent to promote)
	mockClient.AssertNumberOfCalls(t, "UpdateActionStatus", 0)
}
