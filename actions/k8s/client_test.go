package k8s

import (
	"context"
	"fmt"
	"testing"
	"time"

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
		Project: "proj",
		Domain:  "dev",
		Name:    "run1",
	}
	ta := &executorv1.TaskAction{
		Spec: executorv1.TaskActionSpec{
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

func TestBuildTaskActionName(t *testing.T) {
	runID := &common.RunIdentifier{
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

func TestFlyteNamespace(t *testing.T) {
	assert.Equal(t, "flyte", flyteNamespace)
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

func TestInheritRunContextFromParentTaskAction(t *testing.T) {
	interruptible := true
	parent := &executorv1.TaskAction{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"team": "platform",
			},
			Annotations: map[string]string{
				"owner": "sdk",
			},
		},
		Spec: executorv1.TaskActionSpec{
			EnvVars: map[string]string{
				"TRACE_ID": "abc123",
				"TEAM":     "platform",
			},
			Interruptible: &interruptible,
		},
	}

	child := &executorv1.TaskAction{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"flyte.org/run": "run1",
			},
		},
		Spec: executorv1.TaskActionSpec{},
	}

	inheritRunContextFromParentTaskAction(child, parent)

	require.NotNil(t, child.Spec.Interruptible)
	assert.True(t, *child.Spec.Interruptible)
	assert.Equal(t, parent.Spec.EnvVars, child.Spec.EnvVars)
	assert.Equal(t, "platform", child.Labels["team"])
	assert.Equal(t, "run1", child.Labels["flyte.org/run"])
	assert.Equal(t, "sdk", child.Annotations["owner"])

	// Verify deep copy (child mutation must not mutate parent map).
	child.Spec.EnvVars["TRACE_ID"] = "mutated"
	assert.Equal(t, "abc123", parent.Spec.EnvVars["TRACE_ID"])
	child.Annotations["owner"] = "mutated"
	assert.Equal(t, "sdk", parent.Annotations["owner"])
}

func TestInheritRunContextFromParentTaskAction_DoesNotOverrideExistingLabels(t *testing.T) {
	parentInterruptible := true
	parent := &executorv1.TaskAction{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"flyte.org/run": "should-not-override",
				"team":          "platform",
			},
		},
		Spec: executorv1.TaskActionSpec{
			EnvVars: map[string]string{
				"TRACE_ID": "parent",
				"TEAM":     "platform",
			},
			Interruptible: &parentInterruptible,
		},
	}

	child := &executorv1.TaskAction{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"flyte.org/run": "run1",
			},
		},
		Spec: executorv1.TaskActionSpec{},
	}

	inheritRunContextFromParentTaskAction(child, parent)

	require.NotNil(t, child.Spec.Interruptible)
	assert.True(t, *child.Spec.Interruptible)
	assert.Equal(t, map[string]string{"TRACE_ID": "parent", "TEAM": "platform"}, child.Spec.EnvVars)
	assert.Equal(t, "run1", child.Labels["flyte.org/run"])
	assert.Equal(t, "platform", child.Labels["team"])
}

func TestNotifyRunService_ChildAddedPromotesParentToRunning(t *testing.T) {
	ctx := context.Background()

	mockClient := runmocks.NewInternalRunServiceClient(t)
	c := &ActionsClient{
		runClient:   mockClient,
		subscribers: make(map[string]map[chan *ActionUpdate]struct{}),
	}

	runID := &common.RunIdentifier{
		Project: "proj",
		Domain:  "dev",
		Name:    "run1",
	}

	ta := &executorv1.TaskAction{
		Spec: executorv1.TaskActionSpec{
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
			Run:  &common.RunIdentifier{Project: "proj", Domain: "dev", Name: "run1"},
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
			Run:  &common.RunIdentifier{Project: "proj", Domain: "dev", Name: "run1"},
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

func newWorkerTestClient(numWorkers, bufSize int) *ActionsClient {
	c := &ActionsClient{
		numWorkers: numWorkers,
		workerChs:  make([]chan watch.Event, numWorkers),
	}
	for i := range c.workerChs {
		c.workerChs[i] = make(chan watch.Event, bufSize)
	}
	return c
}

func newTaskActionEvent(name string, eventType watch.EventType) watch.Event {
	ta := &executorv1.TaskAction{}
	ta.Name = name
	return watch.Event{Type: eventType, Object: ta}
}

func newTaskAction(name string) *executorv1.TaskAction {
	ta := &executorv1.TaskAction{}
	ta.Name = name
	return ta
}

func TestDispatchEvent_ConsistentSharding(t *testing.T) {
	c := newWorkerTestClient(4, 10)
	taskAction := newTaskAction("run1-action1")

	c.dispatchEvent(taskAction, watch.Modified)
	c.dispatchEvent(taskAction, watch.Modified)

	// Both events must land in exactly one shard (the same one each time)
	var total int
	for _, ch := range c.workerChs {
		total += len(ch)
	}
	assert.Equal(t, 2, total)
}

func TestDispatchEvent_DifferentNamesCanLandOnDifferentShards(t *testing.T) {
	c := newWorkerTestClient(4, 10)

	// Send many differently-named events and confirm they spread across shards
	names := []string{"run1-a", "run1-b", "run1-c", "run1-d", "run1-e", "run1-f", "run1-g", "run1-h"}
	for _, name := range names {
		c.dispatchEvent(newTaskAction(name), watch.Modified)
	}

	var nonEmpty int
	for _, ch := range c.workerChs {
		if len(ch) > 0 {
			nonEmpty++
		}
	}
	assert.Greater(t, nonEmpty, 1, "expected events to spread across more than one shard")
}

func TestDispatchEvent_NilTaskActionIsIgnored(t *testing.T) {
	c := newWorkerTestClient(2, 10)

	c.dispatchEvent(nil, watch.Modified)

	for i, ch := range c.workerChs {
		assert.Equal(t, 0, len(ch), "worker %d should have received nothing", i)
	}
}

func TestDispatchEvent_FullChannelBlocks(t *testing.T) {
	c := newWorkerTestClient(1, 1) // single worker, capacity 1
	taskAction := newTaskAction("run1-action1")

	c.dispatchEvent(taskAction, watch.Modified) // fills the channel

	// Second dispatch must block because the channel is full.
	// Run it in a goroutine and verify it unblocks once the channel is drained.
	done := make(chan struct{})
	go func() {
		c.dispatchEvent(taskAction, watch.Modified)
		close(done)
	}()

	// Drain the channel to unblock the sender.
	<-c.workerChs[0]

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("dispatchEvent did not unblock after channel was drained")
	}
}

func TestDispatchEvent_DeepCopiesTaskAction(t *testing.T) {
	c := newWorkerTestClient(1, 10)
	taskAction := newTaskAction("run1-action1")

	c.dispatchEvent(taskAction, watch.Modified)
	taskAction.Name = "mutated-after-dispatch"

	event := <-c.workerChs[0]
	dispatched, ok := event.Object.(*executorv1.TaskAction)
	require.True(t, ok)
	assert.Equal(t, "run1-action1", dispatched.Name)
}

func TestWorker_ExitsOnStopCh(t *testing.T) {
	c := &ActionsClient{}
	ch := make(chan watch.Event, 10)
	stopCh := make(chan struct{})

	done := make(chan struct{})
	go func() {
		c.worker(context.Background(), ch, stopCh)
		close(done)
	}()

	close(stopCh)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not exit after stopCh was closed")
	}
}

func TestWorker_ExitsOnContextCancel(t *testing.T) {
	c := &ActionsClient{}
	ch := make(chan watch.Event, 10)
	stopCh := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		c.worker(ctx, ch, stopCh)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not exit after context was cancelled")
	}
}

func TestBuildActionUpdate_PropagatesErrorState(t *testing.T) {
	ta := &executorv1.TaskAction{
		Spec: executorv1.TaskActionSpec{
			Project: "flytesnacks", Domain: "development", RunName: "r1", ActionName: "child",
			TaskType: "python",
		},
		Status: executorv1.TaskActionStatus{
			Conditions: []metav1.Condition{
				{Type: string(executorv1.ConditionTypeFailed), Status: metav1.ConditionTrue},
			},
			ErrorState: &executorv1.ErrorState{
				Code: "OOMKilled", Kind: "USER", Message: "container oom",
			},
		},
	}

	upd := buildActionUpdate(context.Background(), ta, watch.Modified)

	if assert.NotNil(t, upd.ErrorState, "buildActionUpdate must propagate Status.ErrorState onto the channel update") {
		assert.Equal(t, "OOMKilled", upd.ErrorState.Code)
		assert.Equal(t, "USER", upd.ErrorState.Kind)
		assert.Equal(t, "container oom", upd.ErrorState.Message)
	}
}

func TestBuildActionUpdate_NilErrorStateWhenAbsent(t *testing.T) {
	ta := &executorv1.TaskAction{
		Spec: executorv1.TaskActionSpec{
			Project: "p", Domain: "d", RunName: "r", ActionName: "a", TaskType: "python",
		},
	}

	upd := buildActionUpdate(context.Background(), ta, watch.Added)

	assert.Nil(t, upd.ErrorState)
}

// A delete event (real or DeletedFinalStateUnknown tombstone delivered after
// informer resync) on a TaskAction that already reached a terminal phase must
// NOT overwrite the recorded phase with ABORTED.
func TestBuildActionUpdate_DeleteAfterTerminalPreservesPhase(t *testing.T) {
	cases := []struct {
		name      string
		condition string
		want      common.ActionPhase
	}{
		{"succeeded", string(executorv1.ConditionTypeSucceeded), common.ActionPhase_ACTION_PHASE_SUCCEEDED},
		{"failed", string(executorv1.ConditionTypeFailed), common.ActionPhase_ACTION_PHASE_FAILED},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ta := &executorv1.TaskAction{
				Spec: executorv1.TaskActionSpec{
					Project: "p", Domain: "d", RunName: "r", ActionName: "a",
				},
				Status: executorv1.TaskActionStatus{
					Conditions: []metav1.Condition{
						{Type: tc.condition, Status: metav1.ConditionTrue},
					},
				},
			}

			upd := buildActionUpdate(context.Background(), ta, watch.Deleted)

			require.NotNil(t, upd)
			assert.Equal(t, tc.want, upd.Phase, "delete event must not overwrite a terminal phase with ABORTED")
			assert.True(t, upd.IsDeleted)
		})
	}
}

// A delete event on a TaskAction that is still running (no terminal condition)
// should still mark the action as ABORTED.
func TestBuildActionUpdate_DeleteOnNonTerminalForcesAborted(t *testing.T) {
	ta := &executorv1.TaskAction{
		Spec: executorv1.TaskActionSpec{
			Project: "p", Domain: "d", RunName: "r", ActionName: "a",
		},
		Status: executorv1.TaskActionStatus{
			Conditions: []metav1.Condition{
				{
					Type:   string(executorv1.ConditionTypeProgressing),
					Status: metav1.ConditionTrue,
					Reason: string(executorv1.ConditionReasonExecuting),
				},
			},
		},
	}

	upd := buildActionUpdate(context.Background(), ta, watch.Deleted)

	require.NotNil(t, upd)
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_ABORTED, upd.Phase)
	assert.True(t, upd.IsDeleted)
}
