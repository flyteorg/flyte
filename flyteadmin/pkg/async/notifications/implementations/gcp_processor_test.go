package implementations

import (
	"context"
	"testing"

	"github.com/NYTimes/gizmo/pubsub/pubsubtest"
	"github.com/flyteorg/flyteadmin/pkg/async/notifications/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/pkg/errors"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

var (
	testGcpSubscriber pubsubtest.TestSubscriber
	mockGcpEmailer    mocks.MockEmailer
)

// This method should be invoked before every test to Subscriber.
func initializeGcpSubscriber() {
	testGcpSubscriber.GivenStopError = nil
	testGcpSubscriber.GivenErrError = nil
	testGcpSubscriber.FoundError = nil
	testGcpSubscriber.ProtoMessages = nil
	testGcpSubscriber.JSONMessages = nil
}

func TestGcpProcessor_StartProcessing(t *testing.T) {
	initializeGcpSubscriber()
	testGcpSubscriber.ProtoMessages = append(testGcpSubscriber.ProtoMessages, testSubscriberProtoMessages...)

	testGcpProcessor := NewGcpProcessor(&testGcpSubscriber, &mockGcpEmailer, promutils.NewTestScope())

	sendEmailValidationFunc := func(ctx context.Context, email admin.EmailMessage) error {
		assert.Equal(t, email.Body, testEmail.Body)
		assert.Equal(t, email.RecipientsEmail, testEmail.RecipientsEmail)
		assert.Equal(t, email.SubjectLine, testEmail.SubjectLine)
		assert.Equal(t, email.SenderEmail, testEmail.SenderEmail)
		return nil
	}
	mockGcpEmailer.SetSendEmailFunc(sendEmailValidationFunc)
	assert.Nil(t, testGcpProcessor.(*GcpProcessor).run())

	// Check fornumber of messages processed.
	m := &dto.Metric{}
	err := testGcpProcessor.(*GcpProcessor).systemMetrics.MessageSuccess.Write(m)
	assert.Nil(t, err)
	assert.Equal(t, "counter:<value:1 > ", m.String())
}

func TestGcpProcessor_StartProcessingNoMessages(t *testing.T) {
	initializeGcpSubscriber()

	testGcpProcessor := NewGcpProcessor(&testGcpSubscriber, &mockGcpEmailer, promutils.NewTestScope())

	// Expect no errors are returned.
	assert.Nil(t, testGcpProcessor.(*GcpProcessor).run())

	// Check fornumber of messages processed.
	m := &dto.Metric{}
	err := testGcpProcessor.(*GcpProcessor).systemMetrics.MessageSuccess.Write(m)
	assert.Nil(t, err)
	assert.Equal(t, "counter:<value:0 > ", m.String())
}

func TestGcpProcessor_StartProcessingError(t *testing.T) {
	initializeGcpSubscriber()

	ret := errors.New("err() returned an error")
	// The error set by GivenErrError is returned by Err().
	// Err() is checked before Run() returning.
	testGcpSubscriber.GivenErrError = ret

	testGcpProcessor := NewGcpProcessor(&testGcpSubscriber, &mockGcpEmailer, promutils.NewTestScope())
	assert.Equal(t, ret, testGcpProcessor.(*GcpProcessor).run())
}

func TestGcpProcessor_StartProcessingEmailError(t *testing.T) {
	initializeGcpSubscriber()
	emailError := errors.New("error sending email")
	sendEmailErrorFunc := func(ctx context.Context, email admin.EmailMessage) error {
		return emailError
	}
	mockGcpEmailer.SetSendEmailFunc(sendEmailErrorFunc)
	testGcpSubscriber.ProtoMessages = append(testGcpSubscriber.ProtoMessages, testSubscriberProtoMessages...)

	testGcpProcessor := NewGcpProcessor(&testGcpSubscriber, &mockGcpEmailer, promutils.NewTestScope())

	// Even if there is an error in sending an email StartProcessing will return no errors.
	assert.Nil(t, testGcpProcessor.(*GcpProcessor).run())

	// Check for an email error stat.
	m := &dto.Metric{}
	err := testGcpProcessor.(*GcpProcessor).systemMetrics.MessageProcessorError.Write(m)
	assert.Nil(t, err)
	assert.Equal(t, "counter:<value:1 > ", m.String())
}

func TestGcpProcessor_StopProcessing(t *testing.T) {
	initializeGcpSubscriber()
	testGcpProcessor := NewGcpProcessor(&testGcpSubscriber, &mockGcpEmailer, promutils.NewTestScope())
	assert.Nil(t, testGcpProcessor.StopProcessing())
}

func TestGcpProcessor_StopProcessingError(t *testing.T) {
	initializeGcpSubscriber()
	stopError := errors.New("stop() returns an error")
	testGcpSubscriber.GivenStopError = stopError
	testGcpProcessor := NewGcpProcessor(&testGcpSubscriber, &mockGcpEmailer, promutils.NewTestScope())
	assert.Equal(t, stopError, testGcpProcessor.StopProcessing())
}
