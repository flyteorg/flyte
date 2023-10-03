package implementations

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/common"

	"github.com/NYTimes/gizmo/pubsub"
	"github.com/NYTimes/gizmo/pubsub/pubsubtest"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/flyteorg/flyteadmin/pkg/async/webhook/mocks"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	repositoryMocks "github.com/flyteorg/flyteadmin/pkg/repositories/mocks"
)

var (
	mockWebhook     = mocks.MockWebhook{}
	repo            = repositoryMocks.NewMockRepository()
	testWebhook     = admin.WebhookPayload{Message: "hello world"}
	workflowRequest = &admin.WorkflowExecutionEventRequest{
		Event: &event.WorkflowExecutionEvent{
			Phase: core.WorkflowExecution_FAILED,
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
		},
	}
	msg, _                = proto.Marshal(workflowRequest)
	testSubscriberMessage = map[string]interface{}{
		"Type":             "Notification",
		"MessageId":        "1-a-3-c",
		"TopicArn":         "arn:aws:sns:my-region:123:flyte-test-notifications",
		"Subject":          "flyteidl.admin.WorkflowExecutionEventRequest",
		"Message":          aws.String(base64.StdEncoding.EncodeToString(msg)),
		"Timestamp":        "2019-01-04T22:59:32.849Z",
		"SignatureVersion": "1",
		"Signature":        "some&ignature==",
		"SigningCertURL":   "https://sns.my-region.amazonaws.com/afdaf",
		"UnsubscribeURL":   "https://sns.my-region.amazonaws.com/sns:my-region:123:flyte-test-notifications:1-2-3-4-5"}
	testSubscriber pubsubtest.TestSubscriber
	mockSub        pubsub.Subscriber = &testSubscriber
)

// This method should be invoked before every test to Subscriber.
func initializeProcessor() {
	testSubscriber.GivenStopError = nil
	testSubscriber.GivenErrError = nil
	testSubscriber.FoundError = nil
	testSubscriber.ProtoMessages = nil
	testSubscriber.JSONMessages = nil
}

func TestProcessor_StartProcessing(t *testing.T) {
	initializeProcessor()
	testSubscriber.JSONMessages = append(testSubscriber.JSONMessages, testSubscriberMessage)

	sendWebhookValidationFunc := func(ctx context.Context, payload admin.WebhookPayload) error {
		assert.Equal(t, payload.Message, testWebhook.Message)
		return nil
	}
	mockWebhook.SetWebhookPostFunc(sendWebhookValidationFunc)
	occurredAt := time.Now().UTC()
	closure := &admin.ExecutionClosure{
		Phase: core.WorkflowExecution_RUNNING,
		StateChangeDetails: &admin.ExecutionStateChangeDetails{
			State:      admin.ExecutionState_EXECUTION_ACTIVE,
			OccurredAt: testutils.MockCreatedAtProto,
		},
		WorkflowId: &core.Identifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
			Version: "version",
		},
	}
	closureBytes, err := proto.Marshal(closure)
	assert.Nil(t, err)
	spec := &admin.ExecutionSpec{
		LaunchPlan: &core.Identifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
			Version: "version",
		},
	}
	specBytes, err := proto.Marshal(spec)
	assert.Nil(t, err)
	repo.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(
		func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
			return models.Execution{
				ExecutionKey: models.ExecutionKey{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
				BaseModel: models.BaseModel{
					ID: uint(8),
				},
				Spec:         specBytes,
				Phase:        core.WorkflowExecution_SUCCEEDED.String(),
				Closure:      closureBytes,
				LaunchPlanID: uint(1),
				WorkflowID:   uint(2),
				StartedAt:    &occurredAt,
			}, nil
		},
	)
	testProcessor := NewWebhookProcessor(common.AWS, mockSub, &mockWebhook, repo, promutils.NewTestScope())
	assert.Nil(t, testProcessor.(*Processor).run())
}

func TestProcessor_StartProcessingNoMessages(t *testing.T) {
	initializeProcessor()
	testProcessor := NewWebhookProcessor(common.AWS, mockSub, &mockWebhook, repo, promutils.NewTestScope())
	assert.Nil(t, testProcessor.(*Processor).run())
}

func TestProcessor_StartProcessingNoNotificationMessage(t *testing.T) {
	var testMessage = map[string]interface{}{
		"Type":      "Not a real notification",
		"MessageId": "1234",
	}
	initializeProcessor()
	testSubscriber.JSONMessages = append(testSubscriber.JSONMessages, testMessage)
	testProcessor := NewWebhookProcessor(common.AWS, mockSub, &mockWebhook, repo, promutils.NewTestScope())
	assert.Nil(t, testProcessor.(*Processor).run())
}

func TestProcessor_StartProcessingMessageWrongDataType(t *testing.T) {
	var testMessage = map[string]interface{}{
		"Type":      "Not a real notification",
		"MessageId": "1234",
		"Message":   12,
	}
	initializeProcessor()
	testSubscriber.JSONMessages = append(testSubscriber.JSONMessages, testMessage)
	testProcessor := NewWebhookProcessor(common.AWS, mockSub, &mockWebhook, repo, promutils.NewTestScope())
	assert.Nil(t, testProcessor.(*Processor).run())
}

func TestProcessor_StartProcessingBase64DecodeError(t *testing.T) {
	var testMessage = map[string]interface{}{
		"Type":      "Not a real notification",
		"MessageId": "1234",
		"Message":   "NotBase64encoded",
	}
	initializeProcessor()
	testSubscriber.JSONMessages = append(testSubscriber.JSONMessages, testMessage)
	testProcessor := NewWebhookProcessor(common.AWS, mockSub, &mockWebhook, repo, promutils.NewTestScope())
	assert.Nil(t, testProcessor.(*Processor).run())
}

func TestProcessor_StartProcessingProtoMarshallError(t *testing.T) {
	var badByte = []byte("atreyu")
	var testMessage = map[string]interface{}{
		"Type":      "Not a real notification",
		"MessageId": "1234",
		"Message":   aws.String(base64.StdEncoding.EncodeToString(badByte)),
	}
	initializeProcessor()
	testSubscriber.JSONMessages = append(testSubscriber.JSONMessages, testMessage)
	testProcessor := NewWebhookProcessor(common.AWS, mockSub, &mockWebhook, repo, promutils.NewTestScope())
	assert.Nil(t, testProcessor.(*Processor).run())
}

func TestProcessor_StartProcessingError(t *testing.T) {
	initializeProcessor()
	var ret = errors.New("err() returned an error")
	testSubscriber.GivenErrError = ret
	testProcessor := NewWebhookProcessor(common.AWS, mockSub, &mockWebhook, repo, promutils.NewTestScope())
	assert.Equal(t, ret, testProcessor.(*Processor).run())
}

func TestProcessor_StartProcessingWebhookError(t *testing.T) {
	initializeProcessor()
	webhookError := errors.New("webhook error")
	sendWebhookErrorFunc := func(ctx context.Context, payload admin.WebhookPayload) error {
		return webhookError
	}
	mockWebhook.SetWebhookPostFunc(sendWebhookErrorFunc)
	testSubscriber.JSONMessages = append(testSubscriber.JSONMessages, testSubscriberMessage)
	testProcessor := NewWebhookProcessor(common.AWS, mockSub, &mockWebhook, repo, promutils.NewTestScope())
	assert.Nil(t, testProcessor.(*Processor).run())
}

func TestProcessor_StopProcessing(t *testing.T) {
	initializeProcessor()
	testProcessor := NewWebhookProcessor(common.AWS, mockSub, &mockWebhook, repo, promutils.NewTestScope())
	assert.Nil(t, testProcessor.StopProcessing())
}

func TestProcessor_StopProcessingError(t *testing.T) {
	initializeProcessor()
	var stopError = errors.New("stop() returns an error")
	testSubscriber.GivenStopError = stopError
	testProcessor := NewWebhookProcessor(common.AWS, mockSub, &mockWebhook, repo, promutils.NewTestScope())
	assert.Equal(t, stopError, testProcessor.StopProcessing())
}
