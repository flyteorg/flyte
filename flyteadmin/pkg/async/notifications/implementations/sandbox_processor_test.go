package implementations

import (
	"context"
	"testing"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/async/notifications/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var mockSandboxEmailer mocks.MockEmailer

func TestSandboxProcessor_StartProcessingSuccess(t *testing.T) {
	msgChan := make(chan []byte, 1)
	msgChan <- msg
	testSandboxProcessor := NewSandboxProcessor(msgChan, &mockSandboxEmailer)

	sendEmailValidationFunc := func(ctx context.Context, email admin.EmailMessage) error {
		assert.Equal(t, testEmail.Body, email.Body)
		assert.Equal(t, testEmail.RecipientsEmail, email.RecipientsEmail)
		assert.Equal(t, testEmail.SubjectLine, email.SubjectLine)
		assert.Equal(t, testEmail.SenderEmail, email.SenderEmail)
		return nil
	}

	mockSandboxEmailer.SetSendEmailFunc(sendEmailValidationFunc)
	assert.Nil(t, testSandboxProcessor.(*SandboxProcessor).run())
}

func TestSandboxProcessor_StartProcessingNoMessage(t *testing.T) {
	msgChan := make(chan []byte, 1)
	testSandboxProcessor := NewSandboxProcessor(msgChan, &mockSandboxEmailer)
	go testSandboxProcessor.StartProcessing()
	time.Sleep(1 * time.Second)
}

func TestSandboxProcessor_StartProcessingError(t *testing.T) {
	msgChan := make(chan []byte, 1)
	msgChan <- msg

	emailError := errors.New("error running processor")
	sendEmailValidationFunc := func(ctx context.Context, email admin.EmailMessage) error {
		return emailError
	}
	mockSandboxEmailer.SetSendEmailFunc(sendEmailValidationFunc)

	testSandboxProcessor := NewSandboxProcessor(msgChan, &mockSandboxEmailer)
	go testSandboxProcessor.StartProcessing()

	// give time to receive the err in StartProcessing
	time.Sleep(1 * time.Second)
	assert.Zero(t, len(msgChan))
}

func TestSandboxProcessor_StartProcessingMessageError(t *testing.T) {
	msgChan := make(chan []byte, 1)
	invalidProtoMessage := []byte("invalid message")
	msgChan <- invalidProtoMessage
	testSandboxProcessor := NewSandboxProcessor(msgChan, &mockSandboxEmailer)
	assert.NotNil(t, testSandboxProcessor.(*SandboxProcessor).run())
}

func TestSandboxProcessor_StartProcessingEmailError(t *testing.T) {
	msgChan := make(chan []byte, 1)
	msgChan <- msg
	testSandboxProcessor := NewSandboxProcessor(msgChan, &mockSandboxEmailer)

	emailError := errors.New("error sending email")
	sendEmailValidationFunc := func(ctx context.Context, email admin.EmailMessage) error {
		return emailError
	}

	mockSandboxEmailer.SetSendEmailFunc(sendEmailValidationFunc)
	assert.NotNil(t, testSandboxProcessor.(*SandboxProcessor).run())
}

func TestSandboxProcessor_StopProcessing(t *testing.T) {
	msgChan := make(chan []byte, 1)
	testSandboxProcessor := NewSandboxProcessor(msgChan, &mockSandboxEmailer)
	assert.Nil(t, testSandboxProcessor.StopProcessing())
}
