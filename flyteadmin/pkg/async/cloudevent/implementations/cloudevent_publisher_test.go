package implementations

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/NYTimes/gizmo/pubsub"
	"github.com/NYTimes/gizmo/pubsub/pubsubtest"
	pbcloudevents "github.com/cloudevents/sdk-go/binding/format/protobuf/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flyte/flyteadmin/pkg/data/mocks"
	repositoryInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	repoMocks "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
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

type DummyRepositories struct {
	repositoryInterfaces.Repository
	RepoExecution repositoryInterfaces.ExecutionRepoInterface
}

func (r *DummyRepositories) ExecutionRepo() repositoryInterfaces.ExecutionRepoInterface {
	return r.RepoExecution
}

func getMockSingleTaskSpec() *admin.ExecutionSpec {
	return &admin.ExecutionSpec{
		LaunchPlan: &core.Identifier{
			ResourceType: core.ResourceType_TASK,
			Project:      "project",
			Domain:       "domain",
			Name:         "name",
			Version:      "version",
		},
		RawOutputDataConfig: &admin.RawOutputDataConfig{OutputLocationPrefix: "default_raw_output"},
	}
}

func getMockExecutionModel() models.Execution {
	spec := getMockSingleTaskSpec()
	specBytes, _ := proto.Marshal(spec)
	startedAt := time.Date(2018, 8, 30, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2022, 01, 18, 0, 0, 0, 0, time.UTC)
	startedAtProto := timestamppb.New(startedAt)
	createdAtProto := timestamppb.New(createdAt)

	closure := admin.ExecutionClosure{
		Phase:     core.WorkflowExecution_RUNNING,
		StartedAt: startedAtProto,
		StateChangeDetails: &admin.ExecutionStateChangeDetails{
			State:      admin.ExecutionState_EXECUTION_ACTIVE,
			OccurredAt: createdAtProto,
		},
		WorkflowId: &core.Identifier{
			ResourceType: core.ResourceType_WORKFLOW,
			Project:      "project",
			Domain:       "domain",
			Name:         "name",
			Version:      "version",
			Org:          "",
		},
		ResolvedSpec: spec,
	}
	closureBytes, _ := proto.Marshal(&closure)
	stateInt := int32(admin.ExecutionState_EXECUTION_ACTIVE)
	executionModel := models.Execution{
		Spec:         specBytes,
		Phase:        core.WorkflowExecution_SUCCEEDED.String(),
		Closure:      closureBytes,
		LaunchPlanID: uint(1),
		WorkflowID:   uint(2),
		StartedAt:    &startedAt,
		State:        &stateInt,
	}
	return executionModel
}

func TestCloudEventsPublisher_TransformWorkflow(t *testing.T) {
	testScope := promutils.NewTestScope()
	ctx := context.Background()

	mockURLData := mocks.NewMockRemoteURL()
	dummyDataConfig := interfaces.RemoteDataConfig{}
	dummyEventPublisherConfig := interfaces.EventsPublisherConfig{}
	cloudEventPublisher := NewCloudEventsWrappedPublisher(nil, mockPubSubSender, testScope, nil, mockURLData, dummyDataConfig, dummyEventPublisherConfig)

	t.Run("single task should skip", func(t *testing.T) {
		mockExecutionRepo := repoMocks.NewMockExecutionRepo()
		mockDB := &DummyRepositories{RepoExecution: mockExecutionRepo}

		mockExecutionRepo.(*repoMocks.MockExecutionRepo).SetGetCallback(func(ctx context.Context, input repositoryInterfaces.Identifier) (models.Execution, error) {
			assert.Equal(t, input.Org, executionID.Org)
			assert.Equal(t, input.Project, executionID.Project)
			assert.Equal(t, input.Domain, executionID.Domain)
			assert.Equal(t, input.Name, executionID.Name)
			dummyModel := getMockExecutionModel()

			return dummyModel, nil
		})

		rawEvent := &event.WorkflowExecutionEvent{
			Phase:       core.WorkflowExecution_SUCCEEDED,
			ExecutionId: &executionID,
		}

		casted := cloudEventPublisher.(*CloudEventWrappedPublisher)
		casted.db = mockDB
		ceEvent, err := casted.TransformWorkflowExecutionEvent(ctx, rawEvent)
		assert.Nil(t, err)
		assert.NotNil(t, ceEvent)
		assert.Nil(t, ceEvent.GetOutputInterface())
	})
}
