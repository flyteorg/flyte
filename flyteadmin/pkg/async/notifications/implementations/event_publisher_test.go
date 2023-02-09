package implementations

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/golang/protobuf/ptypes"

	"github.com/NYTimes/gizmo/pubsub"
	"github.com/NYTimes/gizmo/pubsub/pubsubtest"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

var testEventPublisher pubsubtest.TestPublisher
var mockEventPublisher pubsub.Publisher = &testEventPublisher

var executionID = core.WorkflowExecutionIdentifier{
	Project: "project",
	Domain:  "domain",
	Name:    "name",
}
var nodeExecutionID = core.NodeExecutionIdentifier{
	NodeId:      "node id",
	ExecutionId: &executionID,
}

var taskID = &core.Identifier{
	ResourceType: core.ResourceType_TASK,
	Project:      "p",
	Domain:       "d",
	Version:      "v",
	Name:         "n",
}

var occurredAt = time.Now().UTC()
var occurredAtProto, _ = ptypes.TimestampProto(occurredAt)

var taskPhase = core.TaskExecution_RUNNING

var retryAttempt = uint32(1)

const requestID = "request id"

var taskRequest = &admin.TaskExecutionEventRequest{
	RequestId: requestID,
	Event: &event.TaskExecutionEvent{
		TaskId:                taskID,
		ParentNodeExecutionId: &nodeExecutionID,
		RetryAttempt:          retryAttempt,
		Phase:                 taskPhase,
		OccurredAt:            occurredAtProto,
	},
}

var nodeRequest = &admin.NodeExecutionEventRequest{
	RequestId: requestID,
	Event: &event.NodeExecutionEvent{
		ProducerId: "propeller",
		Id:         &nodeExecutionID,
		OccurredAt: occurredAtProto,
		Phase:      core.NodeExecution_RUNNING,
		InputValue: &event.NodeExecutionEvent_InputUri{
			InputUri: "input uri",
		},
	},
}

var workflowRequest = &admin.WorkflowExecutionEventRequest{
	Event: &event.WorkflowExecutionEvent{
		Phase: core.WorkflowExecution_SUCCEEDED,
		OutputResult: &event.WorkflowExecutionEvent_OutputUri{
			OutputUri: "somestring",
		},
		ExecutionId: &executionID,
	},
}

// This method should be invoked before every test around Publisher.
func initializeEventPublisher() {
	testEventPublisher.Published = nil
	testEventPublisher.GivenError = nil
	testEventPublisher.FoundError = nil
}

func TestNewEventsPublisher_EventTypes(t *testing.T) {
	{
		tests := []struct {
			name            string
			eventTypes      []string
			events          []proto.Message
			shouldSendEvent []bool
			expectedSendCnt int
		}{
			{"eventTypes as workflow,node", []string{"workflow", "node"},
				[]proto.Message{workflowRequest, nodeRequest, taskRequest},
				[]bool{true, true, false},
				2},
			{"eventTypes as workflow,task", []string{"workflow", "task"},
				[]proto.Message{workflowRequest, nodeRequest, taskRequest},
				[]bool{true, false, true},
				2},
			{"eventTypes as workflow,task", []string{"node", "task"},
				[]proto.Message{workflowRequest, nodeRequest, taskRequest},
				[]bool{false, true, true},
				2},
			{"eventTypes as task", []string{"task"},
				[]proto.Message{taskRequest},
				[]bool{true},
				1},
			{"eventTypes as node", []string{"node"},
				[]proto.Message{nodeRequest},
				[]bool{true},
				1},
			{"eventTypes as workflow", []string{"workflow"},
				[]proto.Message{workflowRequest},
				[]bool{true},
				1},
			{"eventTypes as workflow", []string{"workflow"},
				[]proto.Message{nodeRequest, taskRequest},
				[]bool{false, false},
				0},
			{"eventTypes as task", []string{"task"},
				[]proto.Message{workflowRequest, nodeRequest},
				[]bool{false, false},
				0},
			{"eventTypes as node", []string{"node"},
				[]proto.Message{workflowRequest, taskRequest},
				[]bool{false, false},
				0},
			{"eventTypes as all", []string{"all"},
				[]proto.Message{workflowRequest, nodeRequest, taskRequest},
				[]bool{true, true, true},
				3},
			{"eventTypes as *", []string{"*"},
				[]proto.Message{workflowRequest, nodeRequest, taskRequest},
				[]bool{true, true, true},
				3},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				initializeEventPublisher()
				var currentEventPublisher = NewEventsPublisher(mockEventPublisher, promutils.NewTestScope(), test.eventTypes)
				var cnt = 0
				for id, event := range test.events {
					assert.Nil(t, currentEventPublisher.Publish(context.Background(), proto.MessageName(event),
						event))
					if test.shouldSendEvent[id] {
						assert.Equal(t, proto.MessageName(event), testEventPublisher.Published[cnt].Key)
						marshalledData, err := proto.Marshal(event)
						assert.Nil(t, err)
						assert.Equal(t, marshalledData, testEventPublisher.Published[cnt].Body)
						cnt++
					}
				}
				assert.Equal(t, test.expectedSendCnt, len(testEventPublisher.Published))
			})
		}
	}
}

func TestEventPublisher_PublishError(t *testing.T) {
	initializeEventPublisher()
	currentEventPublisher := NewEventsPublisher(mockEventPublisher, promutils.NewTestScope(), []string{"*"})
	var publishError = errors.New("publish() returns an error")
	testEventPublisher.GivenError = publishError
	assert.Equal(t, publishError, currentEventPublisher.Publish(context.Background(),
		proto.MessageName(taskRequest), taskRequest))
}
