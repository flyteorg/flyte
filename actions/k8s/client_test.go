package k8s

import (
	"context"
	"fmt"
	"testing"

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
	// UpdateActionStatus won't be called because phase is UNSPECIFIED
	// RecordActionEvents will be called twice (once per event)
	mockClient.On("RecordActionEvents", mock.Anything, mock.Anything).
		Return(&connect.Response[workflow.RecordActionEventsResponse]{}, nil)

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

	mockClient.On("RecordActionEvents", mock.Anything, mock.Anything).
		Return(&connect.Response[workflow.RecordActionEventsResponse]{}, nil)

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
	mockClient.On("RecordActionEvents", mock.Anything, mock.Anything).
		Return(&connect.Response[workflow.RecordActionEventsResponse]{}, nil)

	c.notifyRunService(ctx, ta, update, watch.Added)
	c.notifyRunService(ctx, ta, update, watch.Added)

	mockClient.AssertNumberOfCalls(t, "RecordAction", 2)
}
