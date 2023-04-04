package implementations

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang/protobuf/jsonpb"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/golang/protobuf/ptypes"

	"github.com/NYTimes/gizmo/pubsub"
	"github.com/NYTimes/gizmo/pubsub/pubsubtest"
	pbcloudevents "github.com/cloudevents/sdk-go/binding/format/protobuf/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

type mockKafkaSender struct{}

func (s mockKafkaSender) Send(ctx context.Context, notificationType string, event cloudevents.Event) error {
	return errorPublish
}

var errorPublish = errors.New("publish() returns an error")
var testCloudEventPublisher pubsubtest.TestPublisher
var mockCloudEventPublisher pubsub.Publisher = &testCloudEventPublisher
var mockPubSubSender = &PubSubSender{Pub: mockCloudEventPublisher}

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
func initializeCloudEventPublisher() {
	testCloudEventPublisher.Published = nil
	testCloudEventPublisher.GivenError = nil
	testCloudEventPublisher.FoundError = nil
}

func TestNewCloudEventsPublisher_EventTypes(t *testing.T) {
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
				initializeCloudEventPublisher()
				var currentEventPublisher = NewCloudEventsPublisher(mockPubSubSender, promutils.NewTestScope(), test.eventTypes)
				var cnt = 0
				for id, event := range test.events {
					assert.Nil(t, currentEventPublisher.Publish(context.Background(), proto.MessageName(event),
						event))
					if test.shouldSendEvent[id] {
						assert.Equal(t, proto.MessageName(event), testCloudEventPublisher.Published[cnt].Key)
						body := testCloudEventPublisher.Published[cnt].Body
						cloudEvent := cloudevents.NewEvent()
						err := pbcloudevents.Protobuf.Unmarshal(body, &cloudEvent)
						assert.Nil(t, err)

						assert.Equal(t, cloudEvent.DataContentType(), cloudevents.ApplicationJSON)
						assert.Equal(t, cloudEvent.SpecVersion(), cloudevents.VersionV1)
						assert.Equal(t, cloudEvent.Type(), fmt.Sprintf("%v.%v", cloudEventTypePrefix, proto.MessageName(event)))
						assert.Equal(t, cloudEvent.Source(), cloudEventSource)
						assert.Equal(t, cloudEvent.Extensions(), map[string]interface{}{jsonSchemaURLKey: jsonSchemaURL})

						e, err := (&jsonpb.Marshaler{}).MarshalToString(event)
						assert.Nil(t, err)
						assert.Equal(t, string(cloudEvent.Data()), e)
						cnt++
					}
				}
				assert.Equal(t, test.expectedSendCnt, len(testCloudEventPublisher.Published))
			})
		}
	}
}

func TestCloudEventPublisher_PublishError(t *testing.T) {
	initializeCloudEventPublisher()
	currentEventPublisher := NewCloudEventsPublisher(mockPubSubSender, promutils.NewTestScope(), []string{"*"})
	testCloudEventPublisher.GivenError = errorPublish
	assert.Equal(t, errorPublish, currentEventPublisher.Publish(context.Background(),
		proto.MessageName(taskRequest), taskRequest))

	currentEventPublisher = NewCloudEventsPublisher(&mockKafkaSender{}, promutils.NewTestScope(), []string{"*"})
	assert.Equal(t, errorPublish, currentEventPublisher.Publish(context.Background(),
		proto.MessageName(taskRequest), taskRequest))
}
