package implementations

import (
	"context"
	"errors"
	"testing"

	"encoding/base64"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/flyteorg/flyteadmin/pkg/async/notifications/mocks"
	"github.com/stretchr/testify/assert"
)

var mockEmailer mocks.MockEmailer

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

	// Because the message stored in Amazon SQS is a JSON of the SNS output, store the test output in the JSON Messages.
	testSubscriber.JSONMessages = append(testSubscriber.JSONMessages, testSubscriberMessage)

	sendEmailValidationFunc := func(ctx context.Context, email admin.EmailMessage) error {
		assert.Equal(t, email.Body, testEmail.Body)
		assert.Equal(t, email.RecipientsEmail, testEmail.RecipientsEmail)
		assert.Equal(t, email.SubjectLine, testEmail.SubjectLine)
		assert.Equal(t, email.SenderEmail, testEmail.SenderEmail)
		return nil
	}
	mockEmailer.SetSendEmailFunc(sendEmailValidationFunc)
	// TODO Add test for metric inc for number of messages processed.
	// Assert 1 message processed and 1 total.
	assert.Nil(t, testProcessor.(*Processor).run())
}

func TestProcessor_StartProcessingNoMessages(t *testing.T) {
	initializeProcessor()
	// Expect no errors are returned.
	assert.Nil(t, testProcessor.(*Processor).run())
	// TODO add test for metric inc() for number of messages processed.
	// Assert 0 messages processed and 0 total.
}

func TestProcessor_StartProcessingNoNotificationMessage(t *testing.T) {
	var testMessage = map[string]interface{}{
		"Type":      "Not a real notification",
		"MessageId": "1234",
	}
	initializeProcessor()
	testSubscriber.JSONMessages = append(testSubscriber.JSONMessages, testMessage)
	assert.Nil(t, testProcessor.(*Processor).run())
	// TODO add test for metric inc() for number of messages processed.
	// Assert 1 messages error and 1 total.
}

func TestProcessor_StartProcessingMessageWrongDataType(t *testing.T) {
	var testMessage = map[string]interface{}{
		"Type":      "Not a real notification",
		"MessageId": "1234",
		"Message":   12,
	}
	initializeProcessor()
	testSubscriber.JSONMessages = append(testSubscriber.JSONMessages, testMessage)
	assert.Nil(t, testProcessor.(*Processor).run())
	// TODO add test for metric inc() for number of messages processed.
	// Assert 1 messages error and 1 total.
}

func TestProcessor_StartProcessingBase64DecodeError(t *testing.T) {
	var testMessage = map[string]interface{}{
		"Type":      "Not a real notification",
		"MessageId": "1234",
		"Message":   "NotBase64encoded",
	}
	initializeProcessor()
	testSubscriber.JSONMessages = append(testSubscriber.JSONMessages, testMessage)
	assert.Nil(t, testProcessor.(*Processor).run())
	// TODO add test for metric inc() for number of messages processed.
	// Assert 1 messages error and 1 total.
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
	assert.Nil(t, testProcessor.(*Processor).run())
	// TODO add test for metric inc() for number of messages processed.
	// Assert 1 messages error and 1 total.
}

func TestProcessor_StartProcessingError(t *testing.T) {
	initializeProcessor()
	var ret = errors.New("err() returned an error")
	// The error set by GivenErrError is returned by Err().
	// Err() is checked before Run() returning.
	testSubscriber.GivenErrError = ret
	assert.Equal(t, ret, testProcessor.(*Processor).run())
}

func TestProcessor_StartProcessingEmailError(t *testing.T) {
	initializeProcessor()
	emailError := errors.New("error sending email")
	sendEmailErrorFunc := func(ctx context.Context, email admin.EmailMessage) error {
		return emailError
	}
	mockEmailer.SetSendEmailFunc(sendEmailErrorFunc)
	testSubscriber.JSONMessages = append(testSubscriber.JSONMessages, testSubscriberMessage)

	// Even if there is an error in sending an email StartProcessing will return no errors.
	// TODO: Once stats have been added check for an email error stat.
	assert.Nil(t, testProcessor.(*Processor).run())
}

func TestProcessor_StopProcessing(t *testing.T) {
	initializeProcessor()
	assert.Nil(t, testProcessor.StopProcessing())
}

func TestProcessor_StopProcessingError(t *testing.T) {
	initializeProcessor()
	var stopError = errors.New("stop() returns an error")
	testSubscriber.GivenStopError = stopError
	assert.Equal(t, stopError, testProcessor.StopProcessing())
}
