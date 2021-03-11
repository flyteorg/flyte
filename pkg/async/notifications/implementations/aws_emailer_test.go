package implementations

import (
	"testing"

	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	"github.com/flyteorg/flyteadmin/pkg/async/notifications/mocks"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func getNotificationsConfig() runtimeInterfaces.NotificationsConfig {
	return runtimeInterfaces.NotificationsConfig{
		NotificationsEmailerConfig: runtimeInterfaces.NotificationsEmailerConfig{
			Body: "Execution \"{{ name }}\" has succeeded in \"{{ domain }}\". View details at " +
				"<a href=\"https://example.com/executions/{{ project }}/{{ domain }}/{{ name }}\">" +
				"https://example.com/executions/{{ project }}/{{ domain }}/{{ name }}</a>.",
			Sender:  "no-reply@example.com",
			Subject: "Notice: Execution \"{{ name }}\" has succeeded in \"{{ domain }}\".",
		},
	}
}

func TestAwsEmailer_SendEmail(t *testing.T) {
	mockAwsEmail := mocks.SESClient{}
	var awsSES sesiface.SESAPI = &mockAwsEmail
	expectedSenderEmail := "no-reply@example.com"
	emailNotification := admin.EmailMessage{
		SubjectLine: "Notice: Execution \"name\" has succeeded in \"domain\".",
		SenderEmail: "no-reply@example.com",
		RecipientsEmail: []string{
			"my@example.com",
			"john@example.com",
		},
		Body: "Execution \"name\" has succeeded in \"domain\". View details at " +
			"<a href=\"https://example.com/executions/T/B/D\">" +
			"https://example.com/executions/T/B/D</a>.",
	}

	sendEmailValidationFunc := func(input *ses.SendEmailInput) (*ses.SendEmailOutput, error) {
		assert.Equal(t, *input.Source, expectedSenderEmail)
		assert.Equal(t, *input.Message.Body.Html.Data, emailNotification.Body)
		assert.Equal(t, *input.Message.Subject.Data, emailNotification.SubjectLine)
		for _, toEmail := range input.Destination.ToAddresses {
			var foundEmail = false
			for _, verifyToEmail := range emailNotification.RecipientsEmail {
				if *toEmail == verifyToEmail {
					foundEmail = true
				}
			}
			assert.Truef(t, foundEmail, "To Email address [%s] wasn't apart of original inputs.", *toEmail)
		}
		assert.Equal(t, len(input.Destination.ToAddresses), len(emailNotification.RecipientsEmail))
		return &ses.SendEmailOutput{}, nil
	}
	mockAwsEmail.SetSendEmailFunc(sendEmailValidationFunc)
	testEmail := NewAwsEmailer(getNotificationsConfig(), promutils.NewTestScope(), awsSES)

	assert.Nil(t, testEmail.SendEmail(context.Background(), emailNotification))
}

func TestFlyteEmailToSesEmailInput(t *testing.T) {
	emailNotification := admin.EmailMessage{
		SubjectLine: "Notice: Execution \"name\" has succeeded in \"domain\".",
		SenderEmail: "no-reply@example.com",
		RecipientsEmail: []string{
			"my@example.com",
			"john@example.com",
		},
		Body: "Execution \"name\" has succeeded in \"domain\". View details at " +
			"<a href=\"https://example.com/executions/T/B/D\">" +
			"https://example.com/executions/T/B/D</a>.",
	}

	sesEmailInput := FlyteEmailToSesEmailInput(emailNotification)
	assert.Equal(t, *sesEmailInput.Destination.ToAddresses[0], emailNotification.RecipientsEmail[0])
	assert.Equal(t, *sesEmailInput.Destination.ToAddresses[1], emailNotification.RecipientsEmail[1])
	assert.Equal(t, *sesEmailInput.Message.Subject.Data, "Notice: Execution \"name\" has succeeded in \"domain\".")
}

func TestAwsEmailer_SendEmailError(t *testing.T) {
	mockAwsEmail := mocks.SESClient{}
	var awsSES sesiface.SESAPI
	emailError := errors.New("error sending email")
	sendEmailErrorFunc := func(input *ses.SendEmailInput) (*ses.SendEmailOutput, error) {
		return nil, emailError
	}
	mockAwsEmail.SetSendEmailFunc(sendEmailErrorFunc)
	awsSES = &mockAwsEmail

	testEmail := NewAwsEmailer(getNotificationsConfig(), promutils.NewTestScope(), awsSES)

	emailNotification := admin.EmailMessage{
		SubjectLine: "Notice: Execution \"name\" has succeeded in \"domain\".",
		SenderEmail: "no-reply@example.com",
		RecipientsEmail: []string{
			"my@example.com",
			"john@example.com",
		},
		Body: "Execution \"name\" has succeeded in \"domain\". View details at " +
			"<a href=\"https://example.com/executions/T/B/D\">" +
			"https://example.com/executions/T/B/D</a>.",
	}
	assert.EqualError(t, testEmail.SendEmail(context.Background(), emailNotification), "errors were seen while sending emails")
}

func TestAwsEmailer_SendEmailEmailOutput(t *testing.T) {
	mockAwsEmail := mocks.SESClient{}
	var awsSES sesiface.SESAPI
	emailOutput := ses.SendEmailOutput{
		MessageId: aws.String("1234"),
	}
	sendEmailErrorFunc := func(input *ses.SendEmailInput) (*ses.SendEmailOutput, error) {
		return &emailOutput, nil
	}
	mockAwsEmail.SetSendEmailFunc(sendEmailErrorFunc)
	awsSES = &mockAwsEmail

	testEmail := NewAwsEmailer(getNotificationsConfig(), promutils.NewTestScope(), awsSES)

	emailNotification := admin.EmailMessage{
		SubjectLine: "Notice: Execution \"name\" has succeeded in \"domain\".",
		SenderEmail: "no-reply@example.com",
		RecipientsEmail: []string{
			"my@example.com",
			"john@example.com",
		},
		Body: "Execution \"name\" has succeeded in \"domain\". View details at " +
			"<a href=\"https://example.com/executions/T/B/D\">" +
			"https://example.com/executions/T/B/D</a>.",
	}
	assert.Nil(t, testEmail.SendEmail(context.Background(), emailNotification))
}
