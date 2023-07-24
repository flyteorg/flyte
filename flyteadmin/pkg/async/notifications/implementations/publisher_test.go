package implementations

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/NYTimes/gizmo/pubsub"
	"github.com/NYTimes/gizmo/pubsub/pubsubtest"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/flyteorg/flyteadmin/pkg/async/notifications/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

var (
	testPublisher    pubsubtest.TestPublisher
	mockPublisher    pubsub.Publisher = &testPublisher
	currentPublisher                  = NewPublisher(mockPublisher, promutils.NewTestScope())
	testEmail                         = admin.EmailMessage{
		RecipientsEmail: []string{
			"a@example.com",
			"b@example.com",
		},
		SenderEmail: "no-reply@example.com",
		SubjectLine: "Test email",
		Body:        "This is a sample email.",
	}
)

var msg, _ = proto.Marshal(&testEmail)

var testSubscriberMessage = map[string]interface{}{
	"Type":             "Notification",
	"MessageId":        "1-a-3-c",
	"TopicArn":         "arn:aws:sns:my-region:123:flyte-test-notifications",
	"Subject":          "flyteidl.admin.EmailNotification",
	"Message":          aws.String(base64.StdEncoding.EncodeToString(msg)),
	"Timestamp":        "2019-01-04T22:59:32.849Z",
	"SignatureVersion": "1",
	"Signature":        "some&ignature==",
	"SigningCertURL":   "https://sns.my-region.amazonaws.com/afdaf",
	"UnsubscribeURL":   "https://sns.my-region.amazonaws.com/sns:my-region:123:flyte-test-notifications:1-2-3-4-5",
}

var testSubscriberProtoMessages = []proto.Message{
	&testEmail,
}

var (
	testSubscriber pubsubtest.TestSubscriber
	mockSub        pubsub.Subscriber = &testSubscriber
	mockEmail      mocks.MockEmailer
	testProcessor  = NewProcessor(mockSub, &mockEmail, promutils.NewTestScope())
)

// This method should be invoked before every test around Publisher.
func initializePublisher() {
	testPublisher.Published = nil
	testPublisher.GivenError = nil
	testPublisher.FoundError = nil
}

func TestPublisher_PublishSuccess(t *testing.T) {
	initializePublisher()
	assert.Nil(t, currentPublisher.Publish(context.Background(), proto.MessageName(&testEmail), &testEmail))
	assert.Equal(t, 1, len(testPublisher.Published))
	assert.Equal(t, proto.MessageName(&testEmail), testPublisher.Published[0].Key)
	marshalledData, err := proto.Marshal(&testEmail)
	assert.Nil(t, err)
	assert.Equal(t, marshalledData, testPublisher.Published[0].Body)
}

func TestPublisher_PublishError(t *testing.T) {
	initializePublisher()
	publishError := errors.New("publish() returns an error")
	testPublisher.GivenError = publishError
	assert.Equal(t, publishError, currentPublisher.Publish(context.Background(), "test", &testEmail))
}
