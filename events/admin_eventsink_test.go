package events

import (
	"context"
	"reflect"
	"testing"

	"github.com/flyteorg/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flytepropeller/events/errors"
	fastcheckMocks "github.com/flyteorg/flytestdlib/fastcheck/mocks"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	wfEvent = &event.WorkflowExecutionEvent{
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "p",
			Domain:  "d",
			Name:    "n",
		},
		Phase:        core.WorkflowExecution_RUNNING,
		OccurredAt:   ptypes.TimestampNow(),
		ProducerId:   "",
		OutputResult: &event.WorkflowExecutionEvent_OutputUri{OutputUri: ""},
	}

	nodeEvent = &event.NodeExecutionEvent{
		Id: &core.NodeExecutionIdentifier{
			NodeId: "node-id",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    "n",
			},
		},
		Phase:      core.NodeExecution_FAILED,
		OccurredAt: ptypes.TimestampNow(),
		ProducerId: "",
		InputValue: &event.NodeExecutionEvent_InputUri{
			InputUri: "input-uri",
		},
		DeckUri:      deckURI,
		OutputResult: &event.NodeExecutionEvent_OutputUri{OutputUri: ""},
	}

	taskEvent = &event.TaskExecutionEvent{
		Phase:        core.TaskExecution_SUCCEEDED,
		OccurredAt:   ptypes.TimestampNow(),
		TaskId:       &core.Identifier{ResourceType: core.ResourceType_TASK, Name: "task-id"},
		RetryAttempt: 1,
		ParentNodeExecutionId: &core.NodeExecutionIdentifier{
			NodeId: "node-id",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    "n",
			},
		},
		Logs: []*core.TaskLog{{Uri: "logs.txt"}},
	}
)

func CreateMockAdminEventSink(t *testing.T, rate int64, capacity int) (EventSink, *mocks.AdminServiceClient, *fastcheckMocks.Filter) {
	mockClient := &mocks.AdminServiceClient{}
	filter := &fastcheckMocks.Filter{}
	eventSink, _ := NewAdminEventSink(context.Background(), mockClient, &Config{Rate: rate, Capacity: capacity}, filter)
	return eventSink, mockClient, filter
}

func TestAdminWorkflowEvent(t *testing.T) {
	ctx := context.Background()
	adminEventSink, adminClient, filter := CreateMockAdminEventSink(t, 100, 1000)
	filter.OnAddMatch(mock.Anything, mock.Anything).Return(true)
	filter.OnContainsMatch(mock.Anything, mock.Anything).Return(false)

	adminClient.On(
		"CreateWorkflowEvent",
		ctx,
		mock.MatchedBy(func(req *admin.WorkflowExecutionEventRequest) bool {
			return req.Event == wfEvent
		},
		)).Return(&admin.WorkflowExecutionEventResponse{}, nil)

	err := adminEventSink.Sink(ctx, wfEvent)
	assert.NoError(t, err)
}

func TestAdminNodeEvent(t *testing.T) {
	ctx := context.Background()
	adminEventSink, adminClient, filter := CreateMockAdminEventSink(t, 100, 1000)
	filter.OnAddMatch(mock.Anything, mock.Anything).Return(true)
	filter.OnContainsMatch(mock.Anything, mock.Anything).Return(false)

	adminClient.On(
		"CreateNodeEvent",
		ctx,
		mock.MatchedBy(func(req *admin.NodeExecutionEventRequest) bool {
			return req.Event == nodeEvent
		}),
	).Return(&admin.NodeExecutionEventResponse{}, nil)

	err := adminEventSink.Sink(ctx, nodeEvent)
	assert.NoError(t, err)
}

func TestAdminTaskEvent(t *testing.T) {
	ctx := context.Background()
	adminEventSink, adminClient, filter := CreateMockAdminEventSink(t, 100, 1000)
	filter.OnAddMatch(mock.Anything, mock.Anything).Return(true)
	filter.OnContainsMatch(mock.Anything, mock.Anything).Return(false)

	adminClient.On(
		"CreateTaskEvent",
		ctx,
		mock.MatchedBy(func(req *admin.TaskExecutionEventRequest) bool {
			return req.Event == taskEvent
		}),
	).Return(&admin.TaskExecutionEventResponse{}, nil)

	err := adminEventSink.Sink(ctx, taskEvent)
	assert.NoError(t, err)
}

func TestAdminAlreadyExistsError(t *testing.T) {
	ctx := context.Background()
	adminEventSink, adminClient, filter := CreateMockAdminEventSink(t, 100, 1000)
	filter.OnAddMatch(mock.Anything, mock.Anything).Return(true)
	filter.OnContainsMatch(mock.Anything, mock.Anything).Return(false)

	alreadyExistErr := status.Error(codes.AlreadyExists, "Grpc AlreadyExists error")

	adminClient.On(
		"CreateTaskEvent",
		ctx,
		mock.MatchedBy(func(req *admin.TaskExecutionEventRequest) bool { return true }),
	).Return(nil, alreadyExistErr)

	err := adminEventSink.Sink(ctx, taskEvent)
	assert.Error(t, err)
	assert.True(t, errors.IsAlreadyExists(err))
}

func TestAdminRateLimitError(t *testing.T) {
	ctx := context.Background()
	adminEventSink, adminClient, filter := CreateMockAdminEventSink(t, 1, 1)
	filter.OnAddMatch(mock.Anything, mock.Anything).Return(true)
	filter.OnContainsMatch(mock.Anything, mock.Anything).Return(false)

	adminClient.On(
		"CreateTaskEvent",
		ctx,
		mock.MatchedBy(func(req *admin.TaskExecutionEventRequest) bool {
			return req.Event == taskEvent
		}),
	).Return(&admin.TaskExecutionEventResponse{}, nil)

	var rateLimitedErr error
	for i := 0; i < 10; i++ {
		err := adminEventSink.Sink(ctx, taskEvent)

		if err != nil {
			rateLimitedErr = err
			break
		}
	}

	assert.Error(t, rateLimitedErr)
	assert.True(t, errors.IsResourceExhausted(rateLimitedErr))
}

func TestAdminFilterContains(t *testing.T) {
	ctx := context.Background()
	adminEventSink, _, filter := CreateMockAdminEventSink(t, 1, 1)
	filter.OnAddMatch(mock.Anything, mock.Anything).Return(true)
	filter.OnContainsMatch(mock.Anything, mock.Anything).Return(true)

	wfErr := adminEventSink.Sink(ctx, wfEvent)
	assert.NoError(t, wfErr)

	nodeErr := adminEventSink.Sink(ctx, nodeEvent)
	assert.NoError(t, nodeErr)

	taskErr := adminEventSink.Sink(ctx, taskEvent)
	assert.NoError(t, taskErr)
}

func TestIDFromMessage(t *testing.T) {
	nodeEventRetryGroup := &event.NodeExecutionEvent{
		Id: &core.NodeExecutionIdentifier{
			NodeId: "node-id",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    "n",
			},
		},
		Phase:      core.NodeExecution_FAILED,
		OccurredAt: ptypes.TimestampNow(),
		ProducerId: "",
		InputValue: &event.NodeExecutionEvent_InputUri{
			InputUri: "input-uri",
		},
		OutputResult: &event.NodeExecutionEvent_OutputUri{OutputUri: ""},
		RetryGroup:   "1",
	}

	retry0 := &event.TaskExecutionEvent{
		Phase:        core.TaskExecution_SUCCEEDED,
		OccurredAt:   ptypes.TimestampNow(),
		TaskId:       &core.Identifier{ResourceType: core.ResourceType_TASK, Name: "task-id"},
		RetryAttempt: 0,
		ParentNodeExecutionId: &core.NodeExecutionIdentifier{
			NodeId: "node-id",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    "n",
			},
		},
		Logs: []*core.TaskLog{{Uri: "logs.txt"}},
	}

	pv1 := &event.TaskExecutionEvent{
		Phase:        core.TaskExecution_SUCCEEDED,
		PhaseVersion: 1,
		OccurredAt:   ptypes.TimestampNow(),
		TaskId:       &core.Identifier{ResourceType: core.ResourceType_TASK, Name: "task-id"},
		RetryAttempt: 0,
		ParentNodeExecutionId: &core.NodeExecutionIdentifier{
			NodeId: "node-id",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    "n",
			},
		},
		Logs: []*core.TaskLog{{Uri: "logs.txt"}},
	}

	tests := []struct {
		name    string
		message proto.Message
		want    string
	}{
		{"workflow", wfEvent, "p:d:n:2"},
		{"node", nodeEvent, "p:d:n:node-id::5"},
		{"node", nodeEventRetryGroup, "p:d:n:node-id:1:5"},
		{"task", taskEvent, "p:d:n:node-id:task-id::1:3:0"},
		{"task", retry0, "p:d:n:node-id:task-id::0:3:0"},
		{"task", pv1, "p:d:n:node-id:task-id::0:3:1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IDFromMessage(tt.message)
			assert.NoError(t, err)

			if !reflect.DeepEqual(got, []byte(tt.want)) {
				t.Errorf("IDFromMessage() = %s, want %s", string(got), tt.want)
			}
		})
	}
}
