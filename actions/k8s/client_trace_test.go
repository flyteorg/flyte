package k8s

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	runmocks "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect/mocks"
)

const (
	traceProject = "proj"
	traceDomain  = "dev"
	traceRun     = "run1"
	traceParent  = "parent"
)

func newTraceAction() *actions.Action {
	parent := traceParent
	return &actions.Action{
		ActionId: &common.ActionIdentifier{
			Run:  &common.RunIdentifier{Project: traceProject, Domain: traceDomain, Name: traceRun},
			Name: "trace1",
		},
		ParentActionName: &parent,
		Spec:             &actions.Action_Trace{Trace: &workflow.TraceAction{}},
	}
}

func newTraceClient(runClient *runmocks.InternalRunServiceClient) *ActionsClient {
	return &ActionsClient{
		runClient:   runClient,
		subscribers: make(map[string]map[chan *ActionUpdate]struct{}),
		bufferSize:  1,
	}
}

func TestEnqueueTrace_RejectedRecordIsReturned(t *testing.T) {
	for _, tc := range []struct {
		name     string
		response *connect.Response[workflow.RecordActionResponse]
	}{
		{
			name: "rejected in the body",
			response: connect.NewResponse(&workflow.RecordActionResponse{
				Status: &status.Status{Code: int32(code.Code_INTERNAL), Message: "failed to create action"},
			}),
		},
		{name: "no status", response: connect.NewResponse(&workflow.RecordActionResponse{})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := runmocks.NewInternalRunServiceClient(t)
			mockClient.On("RecordAction", mock.Anything, mock.Anything).Return(tc.response, nil).Once()
			c := newTraceClient(mockClient)
			updates := c.Subscribe(traceRun, traceParent)

			err := c.Enqueue(context.Background(), newTraceAction(), nil)

			t.Logf("Enqueue error: %v; updates delivered: %d", err, len(updates))
			require.Error(t, err)
			assert.Equal(t, connect.CodeInternal, connect.CodeOf(err))
			assert.Empty(t, updates, "a trace the run service rejected must not be reported as succeeded")
		})
	}
}

func TestEnqueueTrace_AcceptedRecordNotifiesSubscribers(t *testing.T) {
	mockClient := runmocks.NewInternalRunServiceClient(t)
	mockClient.On("RecordAction", mock.Anything, mock.Anything).
		Return(connect.NewResponse(&workflow.RecordActionResponse{
			Status: &status.Status{Code: int32(code.Code_OK)},
		}), nil).Once()
	c := newTraceClient(mockClient)
	updates := c.Subscribe(traceRun, traceParent)

	require.NoError(t, c.Enqueue(context.Background(), newTraceAction(), nil))

	require.Len(t, updates, 1)
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_SUCCEEDED, (<-updates).Phase)
}
