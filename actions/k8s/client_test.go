package k8s

import (
	"context"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

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

func ptr[T any](v T) *T {
	return &v
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

// mockWatcher implements watch.Interface for testing
type mockWatcher struct {
	ch chan watch.Event
}

func (m *mockWatcher) Stop()                         {}
func (m *mockWatcher) ResultChan() <-chan watch.Event { return m.ch }

// newFakeK8sClient creates a fake client.WithWatch with executorv1 scheme registered.
func newFakeK8sClient(objs ...client.Object) client.WithWatch {
	scheme := runtime.NewScheme()
	_ = executorv1.AddToScheme(scheme)
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
}

// watchOverrideClient wraps a fake client but overrides Watch for testing doWatch behavior.
type watchOverrideClient struct {
	client.WithWatch
	watchFn func(ctx context.Context, obj client.ObjectList, opts ...client.ListOption) (watch.Interface, error)
}

func (w *watchOverrideClient) Watch(ctx context.Context, obj client.ObjectList, opts ...client.ListOption) (watch.Interface, error) {
	return w.watchFn(ctx, obj, opts...)
}

func TestHandleWatchEvent_UpdatesResourceVersion(t *testing.T) {
	c := &ActionsClient{
		subscribers: make(map[string]map[chan *ActionUpdate]struct{}),
	}

	ta := &executorv1.TaskAction{
		Spec: executorv1.TaskActionSpec{
			Org:        "org",
			Project:    "proj",
			Domain:     "dev",
			RunName:    "run1",
			ActionName: "action-1",
		},
	}
	ta.ResourceVersion = "500"

	c.handleWatchEvent(context.Background(), watch.Event{
		Type:   watch.Modified,
		Object: ta,
	})

	c.mu.RLock()
	assert.Equal(t, "500", c.lastResourceVersion)
	c.mu.RUnlock()
}

func TestDoWatch_PassesResourceVersionOnReconnect(t *testing.T) {
	var capturedOpts *client.ListOptions

	eventCh := make(chan watch.Event)
	close(eventCh)

	woc := &watchOverrideClient{
		WithWatch: newFakeK8sClient(),
		watchFn: func(ctx context.Context, obj client.ObjectList, opts ...client.ListOption) (watch.Interface, error) {
			capturedOpts = &client.ListOptions{}
			for _, opt := range opts {
				opt.ApplyToList(capturedOpts)
			}
			return &mockWatcher{ch: eventCh}, nil
		},
	}

	c := &ActionsClient{
		k8sClient:           woc,
		subscribers:         make(map[string]map[chan *ActionUpdate]struct{}),
		stopCh:              make(chan struct{}),
		lastResourceVersion: "100",
	}

	_ = c.doWatch(context.Background())

	assert.NotNil(t, capturedOpts)
	assert.NotNil(t, capturedOpts.Raw)
	assert.Equal(t, "100", capturedOpts.Raw.ResourceVersion)
}

func TestDoWatch_HandlesGoneError(t *testing.T) {
	taskAction := &executorv1.TaskAction{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "taskaction-1",
			Namespace:       "default",
			ResourceVersion: "200",
		},
		Spec: executorv1.TaskActionSpec{
			Org:              "org",
			Project:          "proj",
			Domain:           "dev",
			RunName:          "run1",
			ActionName:       "action-1",
			ParentActionName: ptr("parent-1"),
		},
	}

	eventCh := make(chan watch.Event, 1)
	eventCh <- watch.Event{
		Type: watch.Error,
		Object: &metav1.Status{
			Code:   410,
			Reason: metav1.StatusReasonGone,
		},
	}

	woc := &watchOverrideClient{
		WithWatch: newFakeK8sClient(taskAction),
		watchFn: func(ctx context.Context, obj client.ObjectList, opts ...client.ListOption) (watch.Interface, error) {
			return &mockWatcher{ch: eventCh}, nil
		},
	}

	updateCh := make(chan *ActionUpdate, 1)
	c := &ActionsClient{
		k8sClient:           woc,
		subscribers:         map[string]map[chan *ActionUpdate]struct{}{"parent-1": {updateCh: {}}},
		stopCh:              make(chan struct{}),
		lastResourceVersion: "100",
	}

	err := c.doWatch(context.Background())

	// resyncFromList replays the current snapshot and doWatch returns nil.
	assert.NoError(t, err)

	select {
	case update := <-updateCh:
		assert.Equal(t, "action-1", update.ActionID.Name)
		assert.Equal(t, "parent-1", update.ParentActionName)
		assert.False(t, update.IsDeleted)
	default:
		t.Fatal("expected relist replay to notify subscribers")
	}

	c.mu.RLock()
	assert.Equal(t, "200", c.lastResourceVersion)
	c.mu.RUnlock()
}

func TestDoWatch_PreservesResourceVersionOnOtherErrors(t *testing.T) {
	eventCh := make(chan watch.Event, 1)
	eventCh <- watch.Event{
		Type: watch.Error,
		Object: &metav1.Status{
			Code:   500,
			Reason: metav1.StatusReasonInternalError,
		},
	}

	woc := &watchOverrideClient{
		WithWatch: newFakeK8sClient(),
		watchFn: func(ctx context.Context, obj client.ObjectList, opts ...client.ListOption) (watch.Interface, error) {
			return &mockWatcher{ch: eventCh}, nil
		},
	}

	c := &ActionsClient{
		k8sClient:           woc,
		subscribers:         make(map[string]map[chan *ActionUpdate]struct{}),
		stopCh:              make(chan struct{}),
		lastResourceVersion: "100",
	}

	err := c.doWatch(context.Background())

	assert.Error(t, err)

	c.mu.RLock()
	assert.Equal(t, "100", c.lastResourceVersion)
	c.mu.RUnlock()
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
