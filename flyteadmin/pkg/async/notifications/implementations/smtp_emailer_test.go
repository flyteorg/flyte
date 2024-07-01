package implementations

import (
	"context"
	"github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func getNotificationsEmailerConfig() interfaces.NotificationsConfig {
	return interfaces.NotificationsConfig{
		Type:                         "",
		Region:                       "",
		AWSConfig:                    interfaces.AWSConfig{},
		GCPConfig:                    interfaces.GCPConfig{},
		NotificationsPublisherConfig: interfaces.NotificationsPublisherConfig{},
		NotificationsProcessorConfig: interfaces.NotificationsProcessorConfig{},
		NotificationsEmailerConfig: interfaces.NotificationsEmailerConfig{
			EmailerConfig: interfaces.EmailServerConfig{
				ServiceName:            Smtp,
				SmtpServer:             "smtpServer",
				SmtpPort:               "smtpPort",
				SmtpUsername:           "smtpUsername",
				SmtpPasswordSecretName: "smtp_password",
			},
			Subject: "subject",
			Sender:  "sender",
			Body:    "body"},
		ReconnectAttempts:     1,
		ReconnectDelaySeconds: 2}
}

func TestEmailCreation(t *testing.T) {
	email := admin.EmailMessage{
		RecipientsEmail: []string{"john@doe.com", "teresa@tester.com"},
		SubjectLine:     "subject",
		Body:            "Email Body",
		SenderEmail:     "sender@sender.com",
	}

	body, err := createMailBody("sender@sender.com", email)

	assert.NoError(t, err)
	assert.Equal(t, "From: sender@sender.com\r\nTo: john@doe.com,teresa@tester.com\r\nSubject: subject\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\nEmail Body", body)
}

func TestNewSmtpEmailer(t *testing.T) {
	secretManagerMock := mocks.SecretManager{}
	secretManagerMock.On("Get", mock.Anything, "smtp_password").Return("password", nil)

	notificationsConfig := getNotificationsEmailerConfig()

	smtpEmailer := NewSmtpEmailer(context.Background(), notificationsConfig, promutils.NewTestScope(), &secretManagerMock)

	assert.NotNil(t, smtpEmailer)
}
