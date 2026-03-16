package k8s

import (
	"context"
	"fmt"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/watch"

	executorv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/flytestdlib/fastcheck"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
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

func TestNotifyRunService_FailedRecordEnqueuesRetry(t *testing.T) {
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

	// First event — RecordAction fails, should enqueue for retry
	c.notifyRunService(ctx, ta, update, watch.Added)

	mockClient.AssertNumberOfCalls(t, "RecordAction", 1)
	assert.Len(t, c.pendingRecords, 1, "failed record should be enqueued for retry")

	// Second Added event — not in bloom filter, so it retries inline too
	mockClient.On("RecordAction", mock.Anything, mock.Anything).
		Return(&connect.Response[workflow.RecordActionResponse]{}, nil).Once()
	c.notifyRunService(ctx, ta, update, watch.Added)

	mockClient.AssertNumberOfCalls(t, "RecordAction", 2)

	// Third event — now it's in the filter, should be skipped
	c.notifyRunService(ctx, ta, update, watch.Added)
	mockClient.AssertNumberOfCalls(t, "RecordAction", 2)
}

func TestProcessRetries_SuccessfulRetry(t *testing.T) {
	ctx := context.Background()

	mockClient := runmocks.NewInternalRunServiceClient(t)

	filter, err := fastcheck.NewOppoBloomFilter(128, promutils.NewTestScope())
	assert.NoError(t, err)

	c := &ActionsClient{
		runClient:      mockClient,
		recordedFilter: filter,
		subscribers:    make(map[string]map[chan *ActionUpdate]struct{}),
	}

	actionID := &common.ActionIdentifier{
		Run:  &common.RunIdentifier{Org: "org", Project: "proj", Domain: "dev", Name: "run1"},
		Name: "action-retry-ok",
	}
	actionKey := []byte(buildTaskActionName(actionID))
	req := &workflow.RecordActionRequest{ActionId: actionID}

	// Enqueue a pending record with nextRetry in the past so it's due
	c.pendingRecords = []*pendingRecord{{
		req:       req,
		actionKey: actionKey,
		attempts:  0,
		nextRetry: time.Now().Add(-1 * time.Second),
	}}

	mockClient.On("RecordAction", mock.Anything, mock.Anything).
		Return(&connect.Response[workflow.RecordActionResponse]{}, nil).Once()

	c.processRetries(ctx)

	mockClient.AssertNumberOfCalls(t, "RecordAction", 1)
	assert.Empty(t, c.pendingRecords, "successful retry should clear pending")
	assert.True(t, c.recordedFilter.Contains(ctx, actionKey), "successful retry should add to bloom filter")
}

func TestProcessRetries_ExhaustsRetries(t *testing.T) {
	ctx := context.Background()

	mockClient := runmocks.NewInternalRunServiceClient(t)

	c := &ActionsClient{
		runClient:   mockClient,
		subscribers: make(map[string]map[chan *ActionUpdate]struct{}),
	}

	actionID := &common.ActionIdentifier{
		Run:  &common.RunIdentifier{Org: "org", Project: "proj", Domain: "dev", Name: "run1"},
		Name: "action-exhaust",
	}
	req := &workflow.RecordActionRequest{ActionId: actionID}

	// Enqueue at max retries - 1, so next failure is the last attempt
	c.pendingRecords = []*pendingRecord{{
		req:       req,
		actionKey: []byte(buildTaskActionName(actionID)),
		attempts:  maxRecordRetries - 1,
		nextRetry: time.Now().Add(-1 * time.Second),
	}}

	mockClient.On("RecordAction", mock.Anything, mock.Anything).
		Return((*connect.Response[workflow.RecordActionResponse])(nil), fmt.Errorf("permanent error")).Once()

	c.processRetries(ctx)

	mockClient.AssertNumberOfCalls(t, "RecordAction", 1)
	assert.Empty(t, c.pendingRecords, "exhausted retries should drop the record")
}

func TestProcessRetries_RequeuesOnFailure(t *testing.T) {
	ctx := context.Background()

	mockClient := runmocks.NewInternalRunServiceClient(t)

	c := &ActionsClient{
		runClient:   mockClient,
		subscribers: make(map[string]map[chan *ActionUpdate]struct{}),
	}

	actionID := &common.ActionIdentifier{
		Run:  &common.RunIdentifier{Org: "org", Project: "proj", Domain: "dev", Name: "run1"},
		Name: "action-requeue",
	}
	req := &workflow.RecordActionRequest{ActionId: actionID}

	c.pendingRecords = []*pendingRecord{{
		req:       req,
		actionKey: []byte(buildTaskActionName(actionID)),
		attempts:  1,
		nextRetry: time.Now().Add(-1 * time.Second),
	}}

	mockClient.On("RecordAction", mock.Anything, mock.Anything).
		Return((*connect.Response[workflow.RecordActionResponse])(nil), fmt.Errorf("transient error")).Once()

	c.processRetries(ctx)

	mockClient.AssertNumberOfCalls(t, "RecordAction", 1)
	assert.Len(t, c.pendingRecords, 1, "failed retry should re-enqueue")
	assert.Equal(t, 2, c.pendingRecords[0].attempts, "attempt count should increment")
	assert.True(t, c.pendingRecords[0].nextRetry.After(time.Now()), "next retry should be in the future")
}

func TestProcessRetries_SkipsNotYetDue(t *testing.T) {
	ctx := context.Background()

	mockClient := runmocks.NewInternalRunServiceClient(t)

	c := &ActionsClient{
		runClient:   mockClient,
		subscribers: make(map[string]map[chan *ActionUpdate]struct{}),
	}

	actionID := &common.ActionIdentifier{
		Run:  &common.RunIdentifier{Org: "org", Project: "proj", Domain: "dev", Name: "run1"},
		Name: "action-not-due",
	}
	req := &workflow.RecordActionRequest{ActionId: actionID}

	// nextRetry is far in the future
	c.pendingRecords = []*pendingRecord{{
		req:       req,
		actionKey: []byte(buildTaskActionName(actionID)),
		attempts:  0,
		nextRetry: time.Now().Add(1 * time.Hour),
	}}

	c.processRetries(ctx)

	// Should not have called RecordAction
	mockClient.AssertNumberOfCalls(t, "RecordAction", 0)
	assert.Len(t, c.pendingRecords, 1, "not-yet-due record should remain in queue")
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

func TestBuildTaskActionName(t *testing.T) {
	runID := &common.RunIdentifier{
		Org:     "org",
		Project: "project",
		Domain:  "development",
		Name:    "rabc123",
	}

	t.Run("root action uses a0-0 suffix", func(t *testing.T) {
		// Root: action name == run name
		actionID := &common.ActionIdentifier{
			Run:  runID,
			Name: runID.Name,
		}
		assert.Equal(t, "rabc123-a0-0", buildTaskActionName(actionID))
	})

	t.Run("child action includes action name", func(t *testing.T) {
		actionID := &common.ActionIdentifier{
			Run:  runID,
			Name: "train",
		}
		assert.Equal(t, "rabc123-train-0", buildTaskActionName(actionID))
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
