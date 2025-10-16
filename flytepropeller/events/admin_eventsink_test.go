package events

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flytepropeller/events/errors"
	fastcheckMocks "github.com/flyteorg/flyte/flytestdlib/fastcheck/mocks"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
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
	scope := promutils.NewTestScope()
	eventSink, _ := NewAdminEventSink(context.Background(), mockClient, &Config{Rate: rate, Capacity: capacity, EventQueueSize: 1000}, filter, scope)
	return eventSink, mockClient, filter
}

func TestAdminWorkflowEvent(t *testing.T) {
	ctx := context.Background()
	adminEventSink, adminClient, filter := CreateMockAdminEventSink(t, 100, 1000)
	filter.EXPECT().Add(mock.Anything, mock.Anything).Return(true)
	filter.EXPECT().Contains(mock.Anything, mock.Anything).Return(false)

	adminClient.On(
		"CreateWorkflowEvent",
		ctx,
		mock.MatchedBy(func(req *admin.WorkflowExecutionEventRequest) bool {
			return req.GetEvent() == wfEvent
		},
		)).Return(&admin.WorkflowExecutionEventResponse{}, nil)

	err := adminEventSink.Sink(ctx, wfEvent)
	assert.NoError(t, err)
}

func TestAdminNodeEvent(t *testing.T) {
	ctx := context.Background()
	adminEventSink, adminClient, filter := CreateMockAdminEventSink(t, 100, 1000)
	filter.EXPECT().Add(mock.Anything, mock.Anything).Return(true)
	filter.EXPECT().Contains(mock.Anything, mock.Anything).Return(false)

	adminClient.On(
		"CreateNodeEvent",
		ctx,
		mock.MatchedBy(func(req *admin.NodeExecutionEventRequest) bool {
			return req.GetEvent() == nodeEvent
		}),
	).Return(&admin.NodeExecutionEventResponse{}, nil)

	err := adminEventSink.Sink(ctx, nodeEvent)
	assert.NoError(t, err)
}

func TestAdminTaskEvent(t *testing.T) {
	ctx := context.Background()
	adminEventSink, adminClient, filter := CreateMockAdminEventSink(t, 100, 1000)
	filter.EXPECT().Add(mock.Anything, mock.Anything).Return(true)
	filter.EXPECT().Contains(mock.Anything, mock.Anything).Return(false)

	adminClient.On(
		"CreateTaskEvent",
		ctx,
		mock.MatchedBy(func(req *admin.TaskExecutionEventRequest) bool {
			return req.GetEvent() == taskEvent
		}),
	).Return(&admin.TaskExecutionEventResponse{}, nil)

	err := adminEventSink.Sink(ctx, taskEvent)
	assert.NoError(t, err)
}

func TestAdminFilterContains(t *testing.T) {
	ctx := context.Background()
	adminEventSink, _, filter := CreateMockAdminEventSink(t, 1, 1)
	filter.EXPECT().Add(mock.Anything, mock.Anything).Return(true)
	filter.EXPECT().Contains(mock.Anything, mock.Anything).Return(true)

	wfErr := adminEventSink.Sink(ctx, wfEvent)
	assert.Error(t, wfErr)
	assert.True(t, errors.IsAlreadyExists(wfErr))

	nodeErr := adminEventSink.Sink(ctx, nodeEvent)
	assert.Error(t, nodeErr)
	assert.True(t, errors.IsAlreadyExists(nodeErr))

	taskErr := adminEventSink.Sink(ctx, taskEvent)
	assert.Error(t, taskErr)
	assert.True(t, errors.IsAlreadyExists(taskErr))
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

func TestAsyncEventProcessing(t *testing.T) {
	ctx := context.Background()
	mockClient := &mocks.AdminServiceClient{}
	filter := &fastcheckMocks.Filter{}
	cfg := &Config{
		Rate:           500,
		Capacity:       1000,
		EventQueueSize: 1000,
	}

	scope := promutils.NewTestScope()
	eventSink, err := NewAdminEventSink(ctx, mockClient, cfg, filter, scope)
	assert.NoError(t, err)
	defer eventSink.Close()

	filter.On("Add", mock.Anything, mock.Anything).Return(true)
	filter.On("Contains", mock.Anything, mock.Anything).Return(false)

	wfEvent := &event.WorkflowExecutionEvent{
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "p",
			Domain:  "d",
			Name:    "n",
		},
		Phase:      core.WorkflowExecution_RUNNING,
		OccurredAt: ptypes.TimestampNow(),
	}

	mockClient.On("CreateWorkflowEvent", mock.Anything, mock.Anything).
		Return(&admin.WorkflowExecutionEventResponse{}, nil)

	// Enqueue event
	err = eventSink.Sink(ctx, wfEvent)
	assert.NoError(t, err)

	// Wait for worker to process
	time.Sleep(200 * time.Millisecond)

	// Verify event was sent
	mockClient.AssertExpectations(t)
}

func TestAsyncNoEventsDroppedUnderRateLimit(t *testing.T) {
	ctx := context.Background()
	mockClient := &mocks.AdminServiceClient{}
	filter := &fastcheckMocks.Filter{}

	cfg := &Config{
		Rate:           2,
		Capacity:       2,
		EventQueueSize: 1000,
	}

	scope := promutils.NewTestScope()
	eventSink, err := NewAdminEventSink(ctx, mockClient, cfg, filter, scope)
	assert.NoError(t, err)
	defer eventSink.Close()

	filter.On("Add", mock.Anything, mock.Anything).Return(true)
	filter.On("Contains", mock.Anything, mock.Anything).Return(false)

	mockClient.On("CreateWorkflowEvent", mock.Anything, mock.Anything).
		Return(&admin.WorkflowExecutionEventResponse{}, nil)

	// Send 10 UNIQUE events (different execution names) - this will hit rate limit
	numEvents := 10
	for i := 0; i < numEvents; i++ {
		wfEvent := &event.WorkflowExecutionEvent{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    fmt.Sprintf("execution-%d", i), // Unique execution name
			},
			Phase:      core.WorkflowExecution_RUNNING,
			OccurredAt: ptypes.TimestampNow(),
		}
		err := eventSink.Sink(ctx, wfEvent)
		assert.NoError(t, err, "Sink should never fail even under rate limiting")
	}

	// Wait for all events to be processed (at 2/sec, 10 events = 5 seconds)
	time.Sleep(6 * time.Second)

	// Critical assertion: ALL events must be processed, NONE dropped
	mockClient.AssertNumberOfCalls(t, "CreateWorkflowEvent", numEvents)
}

func TestAsyncQueueBackpressure(t *testing.T) {
	ctx := context.Background()
	mockClient := &mocks.AdminServiceClient{}
	filter := &fastcheckMocks.Filter{}

	cfg := &Config{
		Rate:           1,
		Capacity:       1,
		EventQueueSize: 1000,
	}

	scope := promutils.NewTestScope()
	eventSink, err := NewAdminEventSink(ctx, mockClient, cfg, filter, scope)
	assert.NoError(t, err)
	defer eventSink.Close()

	filter.On("Add", mock.Anything, mock.Anything).Return(true)
	filter.On("Contains", mock.Anything, mock.Anything).Return(false)

	mockClient.On("CreateWorkflowEvent", mock.Anything, mock.Anything).
		Return(&admin.WorkflowExecutionEventResponse{}, nil)

	// Send UNIQUE events in goroutine to test backpressure
	numEvents := 5
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numEvents; i++ {
			wfEvent := &event.WorkflowExecutionEvent{
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "p",
					Domain:  "d",
					Name:    fmt.Sprintf("execution-%d", i), // Unique execution name
				},
				Phase:      core.WorkflowExecution_RUNNING,
				OccurredAt: ptypes.TimestampNow(),
			}
			err := eventSink.Sink(ctx, wfEvent)
			assert.NoError(t, err, "Sink should block but not error when queue is full")
		}
	}()

	// Wait for all events to be enqueued and processed
	wg.Wait()
	time.Sleep(6 * time.Second)

	// All events should be processed, none dropped
	mockClient.AssertNumberOfCalls(t, "CreateWorkflowEvent", numEvents)
}

func TestAsyncGracefulShutdown(t *testing.T) {
	ctx := context.Background()
	mockClient := &mocks.AdminServiceClient{}
	filter := &fastcheckMocks.Filter{}
	cfg := &Config{
		Rate:           500,
		Capacity:       1000,
		EventQueueSize: 1000,
	}

	scope := promutils.NewTestScope()
	eventSink, err := NewAdminEventSink(ctx, mockClient, cfg, filter, scope)
	assert.NoError(t, err)

	filter.On("Add", mock.Anything, mock.Anything).Return(true)
	filter.On("Contains", mock.Anything, mock.Anything).Return(false)

	mockClient.On("CreateWorkflowEvent", mock.Anything, mock.Anything).
		Return(&admin.WorkflowExecutionEventResponse{}, nil)

	// Enqueue multiple UNIQUE events
	numEvents := 5
	for i := 0; i < numEvents; i++ {
		wfEvent := &event.WorkflowExecutionEvent{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    fmt.Sprintf("execution-%d", i), // Unique execution name
			},
			Phase:      core.WorkflowExecution_RUNNING,
			OccurredAt: ptypes.TimestampNow(),
		}
		err := eventSink.Sink(ctx, wfEvent)
		assert.NoError(t, err)
	}

	// Close immediately - should drain queue
	err = eventSink.Close()
	assert.NoError(t, err)

	// All events should have been processed before shutdown
	mockClient.AssertNumberOfCalls(t, "CreateWorkflowEvent", numEvents)
}

func TestAsyncMultipleEventTypes(t *testing.T) {
	ctx := context.Background()
	mockClient := &mocks.AdminServiceClient{}
	filter := &fastcheckMocks.Filter{}
	cfg := &Config{
		Rate:           500,
		Capacity:       1000,
		EventQueueSize: 1000,
	}

	scope := promutils.NewTestScope()
	eventSink, err := NewAdminEventSink(ctx, mockClient, cfg, filter, scope)
	assert.NoError(t, err)
	defer eventSink.Close()

	filter.On("Add", mock.Anything, mock.Anything).Return(true)
	filter.On("Contains", mock.Anything, mock.Anything).Return(false)

	// Workflow event
	wfEvent := &event.WorkflowExecutionEvent{
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "p",
			Domain:  "d",
			Name:    "n",
		},
		Phase:      core.WorkflowExecution_RUNNING,
		OccurredAt: ptypes.TimestampNow(),
	}

	// Node event
	nodeEvent := &event.NodeExecutionEvent{
		Id: &core.NodeExecutionIdentifier{
			NodeId: "node1",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    "n",
			},
		},
		Phase:      core.NodeExecution_RUNNING,
		OccurredAt: ptypes.TimestampNow(),
	}

	// Task event
	taskEvent := &event.TaskExecutionEvent{
		TaskId: &core.Identifier{
			ResourceType: core.ResourceType_TASK,
			Name:         "task1",
		},
		ParentNodeExecutionId: &core.NodeExecutionIdentifier{
			NodeId: "node1",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    "n",
			},
		},
		Phase:      core.TaskExecution_RUNNING,
		OccurredAt: ptypes.TimestampNow(),
	}

	mockClient.On("CreateWorkflowEvent", mock.Anything, mock.Anything).
		Return(&admin.WorkflowExecutionEventResponse{}, nil)
	mockClient.On("CreateNodeEvent", mock.Anything, mock.Anything).
		Return(&admin.NodeExecutionEventResponse{}, nil)
	mockClient.On("CreateTaskEvent", mock.Anything, mock.Anything).
		Return(&admin.TaskExecutionEventResponse{}, nil)

	// Send all three event types
	assert.NoError(t, eventSink.Sink(ctx, wfEvent))
	assert.NoError(t, eventSink.Sink(ctx, nodeEvent))
	assert.NoError(t, eventSink.Sink(ctx, taskEvent))

	// Wait for processing
	time.Sleep(500 * time.Millisecond)

	// All should be processed
	mockClient.AssertExpectations(t)
}

func TestAsyncHighVolume(t *testing.T) {
	ctx := context.Background()
	mockClient := &mocks.AdminServiceClient{}
	filter := &fastcheckMocks.Filter{}

	cfg := &Config{
		Rate:           100,
		Capacity:       100,
		EventQueueSize: 1000,
	}

	scope := promutils.NewTestScope()
	eventSink, err := NewAdminEventSink(ctx, mockClient, cfg, filter, scope)
	assert.NoError(t, err)
	defer eventSink.Close()

	filter.On("Add", mock.Anything, mock.Anything).Return(true)
	filter.On("Contains", mock.Anything, mock.Anything).Return(false)

	mockClient.On("CreateWorkflowEvent", mock.Anything, mock.Anything).
		Return(&admin.WorkflowExecutionEventResponse{}, nil)

	numEvents := 1000
	var wg sync.WaitGroup

	numWorkers := 10
	eventsPerWorker := numEvents / numWorkers

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		workerID := w
		go func() {
			defer wg.Done()
			for i := 0; i < eventsPerWorker; i++ {
				wfEvent := &event.WorkflowExecutionEvent{
					ExecutionId: &core.WorkflowExecutionIdentifier{
						Project: "stress-test",
						Domain:  "production",
						Name:    fmt.Sprintf("worker-%d-execution-%d", workerID, i),
					},
					Phase:      core.WorkflowExecution_RUNNING,
					OccurredAt: ptypes.TimestampNow(),
				}
				err := eventSink.Sink(ctx, wfEvent)
				assert.NoError(t, err, "Sink should never fail under high load")
			}
		}()
	}

	wg.Wait()

	time.Sleep(12 * time.Second)

	mockClient.AssertNumberOfCalls(t, "CreateWorkflowEvent", numEvents)
}

func TestAsyncEventFailureRetry(t *testing.T) {
	ctx := context.Background()
	mockClient := &mocks.AdminServiceClient{}
	filter := &fastcheckMocks.Filter{}
	cfg := &Config{
		Rate:           50,
		Capacity:       100,
		EventQueueSize: 1000,
	}

	scope := promutils.NewTestScope()
	eventSink, err := NewAdminEventSink(ctx, mockClient, cfg, filter, scope)
	assert.NoError(t, err)
	defer eventSink.Close()

	filter.On("Add", mock.Anything, mock.Anything).Return(true)
	filter.On("Contains", mock.Anything, mock.Anything).Return(false)

	wfEvent := &event.WorkflowExecutionEvent{
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "p",
			Domain:  "d",
			Name:    "retry-test",
		},
		Phase:      core.WorkflowExecution_RUNNING,
		OccurredAt: ptypes.TimestampNow(),
	}

	// First call fails
	mockClient.On("CreateWorkflowEvent", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("network error - simulated failure")).
		Once()

	// Second call succeeds
	mockClient.On("CreateWorkflowEvent", mock.Anything, mock.Anything).
		Return(&admin.WorkflowExecutionEventResponse{}, nil).
		Once()

	// First attempt - enqueue event
	err = eventSink.Sink(ctx, wfEvent)
	assert.NoError(t, err)

	// Wait for processing (will fail)
	time.Sleep(200 * time.Millisecond)

	// Second attempt - should be allowed to retry since first failed
	err = eventSink.Sink(ctx, wfEvent)
	assert.NoError(t, err, "Should allow retry after failure")

	// Wait for second processing (should succeed)
	time.Sleep(200 * time.Millisecond)

	// Verify both attempts were made (failure + success)
	mockClient.AssertExpectations(t)
}

func TestAsyncInFlightDuplicate(t *testing.T) {
	ctx := context.Background()
	mockClient := &mocks.AdminServiceClient{}
	filter := &fastcheckMocks.Filter{}

	cfg := &Config{
		Rate:           1,
		Capacity:       1,
		EventQueueSize: 1000,
	}

	scope := promutils.NewTestScope()
	eventSink, err := NewAdminEventSink(ctx, mockClient, cfg, filter, scope)
	assert.NoError(t, err)
	defer eventSink.Close()

	filter.On("Add", mock.Anything, mock.Anything).Return(true)
	filter.On("Contains", mock.Anything, mock.Anything).Return(false)

	wfEvent := &event.WorkflowExecutionEvent{
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "p",
			Domain:  "d",
			Name:    "in-flight-duplicate-test",
		},
		Phase:      core.WorkflowExecution_RUNNING,
		OccurredAt: ptypes.TimestampNow(),
	}

	// Mock should only be called once (for the first event)
	mockClient.On("CreateWorkflowEvent", mock.Anything, mock.Anything).
		Return(&admin.WorkflowExecutionEventResponse{}, nil).
		Once()

	// First submission - should succeed and start processing
	err = eventSink.Sink(ctx, wfEvent)
	assert.NoError(t, err, "First submission should succeed")

	// Immediately submit same event again while first is still in-flight
	// This should be rejected because the event is already queued for processing
	err = eventSink.Sink(ctx, wfEvent)
	assert.Error(t, err, "Second submission should fail - event already in-flight")
	assert.True(t, errors.IsAlreadyExists(err), "Error should be AlreadyExists")

	// Wait for the first event to finish processing
	time.Sleep(2 * time.Second)

	// Verify only one call was made to admin client (no duplicate sent)
	mockClient.AssertExpectations(t)
	mockClient.AssertNumberOfCalls(t, "CreateWorkflowEvent", 1)
}

func TestAsyncNetworkFailures(t *testing.T) {
	t.Run("NetworkTimeout", func(t *testing.T) {
		ctx := context.Background()
		mockClient := &mocks.AdminServiceClient{}
		filter := &fastcheckMocks.Filter{}
		cfg := &Config{
			Rate:           50,
			Capacity:       100,
			EventQueueSize: 1000,
		}

		scope := promutils.NewTestScope()
		eventSink, err := NewAdminEventSink(ctx, mockClient, cfg, filter, scope)
		assert.NoError(t, err)
		defer eventSink.Close()

		filter.On("Add", mock.Anything, mock.Anything).Return(true)
		filter.On("Contains", mock.Anything, mock.Anything).Return(false)

		wfEvent := &event.WorkflowExecutionEvent{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    "network-timeout-test",
			},
			Phase:      core.WorkflowExecution_RUNNING,
			OccurredAt: ptypes.TimestampNow(),
		}

		// Simulate network timeout
		mockClient.On("CreateWorkflowEvent", mock.Anything, mock.Anything).
			Return(nil, fmt.Errorf("context deadline exceeded")).
			Once()

		// Retry should succeed
		mockClient.On("CreateWorkflowEvent", mock.Anything, mock.Anything).
			Return(&admin.WorkflowExecutionEventResponse{}, nil).
			Once()

		// First attempt - will timeout
		err = eventSink.Sink(ctx, wfEvent)
		assert.NoError(t, err)
		time.Sleep(200 * time.Millisecond)

		// Second attempt - should be allowed to retry
		err = eventSink.Sink(ctx, wfEvent)
		assert.NoError(t, err)
		time.Sleep(200 * time.Millisecond)

		mockClient.AssertExpectations(t)
	})

	t.Run("ConnectionRefused", func(t *testing.T) {
		ctx := context.Background()
		mockClient := &mocks.AdminServiceClient{}
		filter := &fastcheckMocks.Filter{}
		cfg := &Config{
			Rate:           50,
			Capacity:       100,
			EventQueueSize: 1000,
		}

		scope := promutils.NewTestScope()
		eventSink, err := NewAdminEventSink(ctx, mockClient, cfg, filter, scope)
		assert.NoError(t, err)
		defer eventSink.Close()

		filter.On("Add", mock.Anything, mock.Anything).Return(true)
		filter.On("Contains", mock.Anything, mock.Anything).Return(false)

		wfEvent := &event.WorkflowExecutionEvent{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    "connection-refused-test",
			},
			Phase:      core.WorkflowExecution_RUNNING,
			OccurredAt: ptypes.TimestampNow(),
		}

		// Simulate connection refused
		mockClient.On("CreateWorkflowEvent", mock.Anything, mock.Anything).
			Return(nil, fmt.Errorf("dial tcp 127.0.0.1:8089: connect: connection refused")).
			Once()

		// After admin restarts, retry should succeed
		mockClient.On("CreateWorkflowEvent", mock.Anything, mock.Anything).
			Return(&admin.WorkflowExecutionEventResponse{}, nil).
			Once()

		// First attempt - connection refused
		err = eventSink.Sink(ctx, wfEvent)
		assert.NoError(t, err)
		time.Sleep(200 * time.Millisecond)

		// Second attempt after admin restart
		err = eventSink.Sink(ctx, wfEvent)
		assert.NoError(t, err)
		time.Sleep(200 * time.Millisecond)

		mockClient.AssertExpectations(t)
	})

	t.Run("IntermittentFailures", func(t *testing.T) {
		ctx := context.Background()
		mockClient := &mocks.AdminServiceClient{}
		filter := &fastcheckMocks.Filter{}
		cfg := &Config{
			Rate:           50,
			Capacity:       100,
			EventQueueSize: 1000,
		}

		scope := promutils.NewTestScope()
		eventSink, err := NewAdminEventSink(ctx, mockClient, cfg, filter, scope)
		assert.NoError(t, err)
		defer eventSink.Close()

		filter.On("Add", mock.Anything, mock.Anything).Return(true)
		filter.On("Contains", mock.Anything, mock.Anything).Return(false)

		wfEvent := &event.WorkflowExecutionEvent{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    "intermittent-test",
			},
			Phase:      core.WorkflowExecution_RUNNING,
			OccurredAt: ptypes.TimestampNow(),
		}

		// Multiple intermittent failures before success
		mockClient.On("CreateWorkflowEvent", mock.Anything, mock.Anything).
			Return(nil, fmt.Errorf("temporary network error")).
			Once()

		mockClient.On("CreateWorkflowEvent", mock.Anything, mock.Anything).
			Return(nil, fmt.Errorf("service unavailable")).
			Once()

		mockClient.On("CreateWorkflowEvent", mock.Anything, mock.Anything).
			Return(&admin.WorkflowExecutionEventResponse{}, nil).
			Once()

		// Three attempts, last one succeeds
		for i := 0; i < 3; i++ {
			err = eventSink.Sink(ctx, wfEvent)
			assert.NoError(t, err, "Sink should allow retry attempt %d", i)
			time.Sleep(200 * time.Millisecond)
		}

		mockClient.AssertExpectations(t)
	})

	t.Run("PartialBatchFailure", func(t *testing.T) {
		ctx := context.Background()
		mockClient := &mocks.AdminServiceClient{}
		filter := &fastcheckMocks.Filter{}
		cfg := &Config{
			Rate:           50,
			Capacity:       100,
			EventQueueSize: 1000,
		}

		scope := promutils.NewTestScope()
		eventSink, err := NewAdminEventSink(ctx, mockClient, cfg, filter, scope)
		assert.NoError(t, err)
		defer eventSink.Close()

		filter.On("Add", mock.Anything, mock.Anything).Return(true)
		filter.On("Contains", mock.Anything, mock.Anything).Return(false)

		// Event 0: success
		mockClient.On("CreateWorkflowEvent", mock.Anything, mock.MatchedBy(func(req *admin.WorkflowExecutionEventRequest) bool {
			return req.GetEvent().GetExecutionId().GetName() == "batch-event-0"
		})).Return(&admin.WorkflowExecutionEventResponse{}, nil).Once()

		// Event 1: fail first time, then succeed on retry
		mockClient.On("CreateWorkflowEvent", mock.Anything, mock.MatchedBy(func(req *admin.WorkflowExecutionEventRequest) bool {
			return req.GetEvent().GetExecutionId().GetName() == "batch-event-1"
		})).Return(nil, fmt.Errorf("network error for event 1")).Once()

		mockClient.On("CreateWorkflowEvent", mock.Anything, mock.MatchedBy(func(req *admin.WorkflowExecutionEventRequest) bool {
			return req.GetEvent().GetExecutionId().GetName() == "batch-event-1"
		})).Return(&admin.WorkflowExecutionEventResponse{}, nil).Once()

		// Event 2: success
		mockClient.On("CreateWorkflowEvent", mock.Anything, mock.MatchedBy(func(req *admin.WorkflowExecutionEventRequest) bool {
			return req.GetEvent().GetExecutionId().GetName() == "batch-event-2"
		})).Return(&admin.WorkflowExecutionEventResponse{}, nil).Once()

		// Event 3: fail first time, then succeed on retry
		mockClient.On("CreateWorkflowEvent", mock.Anything, mock.MatchedBy(func(req *admin.WorkflowExecutionEventRequest) bool {
			return req.GetEvent().GetExecutionId().GetName() == "batch-event-3"
		})).Return(nil, fmt.Errorf("network error for event 3")).Once()

		mockClient.On("CreateWorkflowEvent", mock.Anything, mock.MatchedBy(func(req *admin.WorkflowExecutionEventRequest) bool {
			return req.GetEvent().GetExecutionId().GetName() == "batch-event-3"
		})).Return(&admin.WorkflowExecutionEventResponse{}, nil).Once()

		// Event 4: success
		mockClient.On("CreateWorkflowEvent", mock.Anything, mock.MatchedBy(func(req *admin.WorkflowExecutionEventRequest) bool {
			return req.GetEvent().GetExecutionId().GetName() == "batch-event-4"
		})).Return(&admin.WorkflowExecutionEventResponse{}, nil).Once()

		// Send 5 events - events 1 and 3 will fail initially
		for i := 0; i < 5; i++ {
			wfEvent := &event.WorkflowExecutionEvent{
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "p",
					Domain:  "d",
					Name:    fmt.Sprintf("batch-event-%d", i),
				},
				Phase:      core.WorkflowExecution_RUNNING,
				OccurredAt: ptypes.TimestampNow(),
			}
			err = eventSink.Sink(ctx, wfEvent)
			assert.NoError(t, err)
		}

		time.Sleep(500 * time.Millisecond)

		// Retry failed events - they should be allowed
		for i := 1; i <= 3; i += 2 {
			wfEvent := &event.WorkflowExecutionEvent{
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "p",
					Domain:  "d",
					Name:    fmt.Sprintf("batch-event-%d", i),
				},
				Phase:      core.WorkflowExecution_RUNNING,
				OccurredAt: ptypes.TimestampNow(),
			}
			err = eventSink.Sink(ctx, wfEvent)
			assert.NoError(t, err, "Should allow retry for failed event")
		}

		time.Sleep(500 * time.Millisecond)

		// All events should eventually succeed
		mockClient.AssertExpectations(t)
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		mockClient := &mocks.AdminServiceClient{}
		filter := &fastcheckMocks.Filter{}
		cfg := &Config{
			Rate:           1,
			Capacity:       1,
			EventQueueSize: 1000,
		}

		scope := promutils.NewTestScope()
		eventSink, err := NewAdminEventSink(ctx, mockClient, cfg, filter, scope)
		assert.NoError(t, err)
		defer eventSink.Close()

		filter.On("Add", mock.Anything, mock.Anything).Return(true)
		filter.On("Contains", mock.Anything, mock.Anything).Return(false)

		mockClient.On("CreateWorkflowEvent", mock.Anything, mock.Anything).
			Return(&admin.WorkflowExecutionEventResponse{}, nil)

		wfEvent := &event.WorkflowExecutionEvent{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    "context-cancel-test",
			},
			Phase:      core.WorkflowExecution_RUNNING,
			OccurredAt: ptypes.TimestampNow(),
		}

		// Enqueue event
		err = eventSink.Sink(ctx, wfEvent)
		assert.NoError(t, err)

		// Cancel context while processing
		time.Sleep(100 * time.Millisecond)
		cancel()

		// Should handle cancellation gracefully
		time.Sleep(200 * time.Millisecond)
	})
}

func TestSinkAfterClose(t *testing.T) {
	ctx := context.Background()
	adminEventSink, _, filter := CreateMockAdminEventSink(t, 100, 1000)
	filter.On("Contains", mock.Anything, mock.Anything).Return(false)

	err := adminEventSink.Close()
	assert.NoError(t, err)

	testEvent := &event.WorkflowExecutionEvent{
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "p",
			Domain:  "d",
			Name:    "close-test",
		},
		Phase:      core.WorkflowExecution_RUNNING,
		OccurredAt: ptypes.TimestampNow(),
	}

	err = adminEventSink.Sink(ctx, testEvent)
	assert.Error(t, err, "Sink should fail after Close()")
	assert.Contains(t, err.Error(), "closed", "Error should mention sink is closed")
}

func TestQueueFullNonBlocking(t *testing.T) {
	ctx := context.Background()
	mockClient := &mocks.AdminServiceClient{}
	filter := &fastcheckMocks.Filter{}

	cfg := &Config{
		Rate:           100,
		Capacity:       100,
		EventQueueSize: 2,
	}

	scope := promutils.NewTestScope()
	eventSink, err := NewAdminEventSink(ctx, mockClient, cfg, filter, scope)
	assert.NoError(t, err)

	filter.On("Contains", mock.Anything, mock.Anything).Return(false)
	filter.On("Add", mock.Anything, mock.Anything).Return(true)

	blockProcessing := make(chan struct{})
	processingStarted := make(chan bool, 10)
	processedCount := make(chan int, 10)
	var count int

	mockClient.On("CreateWorkflowEvent", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			processingStarted <- true
			<-blockProcessing
			count++
			processedCount <- count
		}).
		Return(&admin.WorkflowExecutionEventResponse{}, nil)

	// Enqueue 3 events - queue capacity is 2, so one is being processed and 2 are queued
	for i := 0; i < 3; i++ {
		wfEvent := &event.WorkflowExecutionEvent{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    fmt.Sprintf("fill-%d", i),
			},
			Phase:      core.WorkflowExecution_RUNNING,
			OccurredAt: ptypes.TimestampNow(),
		}
		err := eventSink.Sink(ctx, wfEvent)
		assert.NoError(t, err)
	}

	// Wait for worker to start processing first event
	<-processingStarted
	time.Sleep(50 * time.Millisecond)

	// Now the queue should be full: 1 being processed + 2 in queue
	wfEvent4 := &event.WorkflowExecutionEvent{
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "p",
			Domain:  "d",
			Name:    "overflow",
		},
		Phase:      core.WorkflowExecution_RUNNING,
		OccurredAt: ptypes.TimestampNow(),
	}

	err = eventSink.Sink(ctx, wfEvent4)
	assert.Error(t, err, "Should reject event when queue is full")
	assert.True(t, errors.IsResourceExhausted(err), "Should return ResourceExhausted error")

	// Unblock processing and wait for all 3 events to complete
	close(blockProcessing)
	for i := 0; i < 3; i++ {
		<-processedCount
	}

	// Now close gracefully
	err = eventSink.Close()
	assert.NoError(t, err)

	// Verify exactly 3 events were processed
	mockClient.AssertNumberOfCalls(t, "CreateWorkflowEvent", 3)
}
