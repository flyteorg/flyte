package implementations

import (
	"context"
	"crypto/tls"
	"errors"
	"net/smtp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"

	notification_interfaces "github.com/flyteorg/flyte/flyteadmin/pkg/async/notifications/interfaces"
	notification_mocks "github.com/flyteorg/flyte/flyteadmin/pkg/async/notifications/mocks"
	flyte_errors "github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

type StringWriter struct {
	buffer   string
	writeErr error
	closeErr error
}

func (s *StringWriter) Write(p []byte) (n int, err error) {
	s.buffer = s.buffer + string(p)
	return len(p), s.writeErr
}

func (s *StringWriter) Close() error {
	return s.closeErr
}

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
				ServiceName:            SMTP,
				SMTPServer:             "smtpServer",
				SMTPPort:               "smtpPort",
				SMTPUsername:           "smtpUsername",
				SMTPPasswordSecretName: "smtp_password",
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

	body := createMailBody("sender@sender.com", &email)
	assert.Contains(t, body, "From: sender@sender.com\r\n")
	assert.Contains(t, body, "To: john@doe.com,teresa@tester.com")
	assert.Contains(t, body, "Subject: subject\r\n")
	assert.Contains(t, body, "Content-Type: text/html; charset=\"UTF-8\"\r\n")
	assert.Contains(t, body, "Email Body")
}

func TestNewSmtpEmailer(t *testing.T) {
	secretManagerMock := mocks.SecretManager{}
	secretManagerMock.On("Get", mock.Anything, "smtp_password").Return("password", nil)

	notificationsConfig := getNotificationsEmailerConfig()

	smtpEmailer := NewSMTPEmailer(context.Background(), notificationsConfig, promutils.NewTestScope(), &secretManagerMock)

	assert.NotNil(t, smtpEmailer)
}

func TestCreateClient(t *testing.T) {
	auth := smtp.PlainAuth("", "user", "password", "localhost")

	tlsConf := tls.Config{
		InsecureSkipVerify: false,
		ServerName:         "localhost",
		MinVersion:         tls.VersionTLS13,
	}

	smtpClient := &notification_mocks.SMTPClient{}
	smtpClient.On("Hello", "localhost").Return(nil)
	smtpClient.On("Extension", "STARTTLS").Return(true, "")
	smtpClient.On("StartTLS", &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         "localhost",
		MinVersion:         tls.VersionTLS13,
	}).Return(nil)
	smtpClient.On("Extension", "AUTH").Return(true, "")
	smtpClient.On("Auth", auth).Return(nil)

	smtpEmailer := createSMTPEmailer(smtpClient, &tlsConf, &auth, nil)

	client, err := smtpEmailer.createClient(context.Background())

	assert.Nil(t, err)
	assert.NotNil(t, client)

}

func TestCreateClientErrorCreatingClient(t *testing.T) {
	auth := smtp.PlainAuth("", "user", "password", "localhost")

	tlsConf := tls.Config{
		InsecureSkipVerify: false,
		ServerName:         "localhost",
		MinVersion:         tls.VersionTLS13,
	}

	smtpClient := &notification_mocks.SMTPClient{}

	smtpEmailer := createSMTPEmailer(smtpClient, &tlsConf, &auth, errors.New("error creating client"))

	client, err := smtpEmailer.createClient(context.Background())

	assert.Equal(t, flyte_errors.NewFlyteAdminErrorf(codes.Internal, "errors were seen while sending emails"), err)
	assert.Nil(t, client)

}

func TestCreateClientErrorHello(t *testing.T) {
	auth := smtp.PlainAuth("", "user", "password", "localhost")

	tlsConf := tls.Config{
		InsecureSkipVerify: false,
		ServerName:         "localhost",
		MinVersion:         tls.VersionTLS13,
	}

	smtpClient := &notification_mocks.SMTPClient{}
	smtpClient.On("Hello", "localhost").Return(errors.New("Error with hello"))

	smtpEmailer := createSMTPEmailer(smtpClient, &tlsConf, &auth, nil)

	client, err := smtpEmailer.createClient(context.Background())

	assert.Equal(t, flyte_errors.NewFlyteAdminErrorf(codes.Internal, "errors were seen while sending emails"), err)
	assert.Nil(t, client)

}

func TestCreateClientErrorStartTLS(t *testing.T) {
	auth := smtp.PlainAuth("", "user", "password", "localhost")

	tlsConf := tls.Config{
		InsecureSkipVerify: false,
		ServerName:         "localhost",
		MinVersion:         tls.VersionTLS13,
	}

	smtpClient := &notification_mocks.SMTPClient{}
	smtpClient.On("Hello", "localhost").Return(nil).Times(1)
	smtpClient.On("Extension", "STARTTLS").Return(true, "").Times(1)
	smtpClient.On("StartTLS", &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         "localhost",
		MinVersion:         tls.VersionTLS13,
	}).Return(errors.New("Error with startls")).Times(1)

	smtpEmailer := createSMTPEmailer(smtpClient, &tlsConf, &auth, nil)

	client, err := smtpEmailer.createClient(context.Background())

	assert.Equal(t, flyte_errors.NewFlyteAdminErrorf(codes.Internal, "errors were seen while sending emails"), err)
	assert.Nil(t, client)

}

func TestCreateClientErrorAuth(t *testing.T) {
	auth := smtp.PlainAuth("", "user", "password", "localhost")

	tlsConf := tls.Config{
		InsecureSkipVerify: false,
		ServerName:         "localhost",
		MinVersion:         tls.VersionTLS13,
	}

	smtpClient := &notification_mocks.SMTPClient{}
	smtpClient.On("Hello", "localhost").Return(nil).Times(1)
	smtpClient.On("Extension", "STARTTLS").Return(true, "").Times(1)
	smtpClient.On("StartTLS", &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         "localhost",
		MinVersion:         tls.VersionTLS13,
	}).Return(nil).Times(1)
	smtpClient.On("Extension", "AUTH").Return(true, "").Times(1)
	smtpClient.On("Auth", auth).Return(errors.New("Error with hello")).Times(1)

	smtpEmailer := createSMTPEmailer(smtpClient, &tlsConf, &auth, nil)

	client, err := smtpEmailer.createClient(context.Background())

	assert.Equal(t, flyte_errors.NewFlyteAdminErrorf(codes.Internal, "errors were seen while sending emails"), err)
	assert.Nil(t, client)

}

func TestSendMail(t *testing.T) {
	auth := smtp.PlainAuth("", "user", "password", "localhost")

	tlsConf := tls.Config{
		InsecureSkipVerify: false,
		ServerName:         "localhost",
		MinVersion:         tls.VersionTLS13,
	}

	stringWriter := StringWriter{buffer: ""}

	smtpClient := &notification_mocks.SMTPClient{}
	smtpClient.On("Noop").Return(errors.New("no connection")).Times(1)
	smtpClient.On("Close").Return(nil).Times(1)
	smtpClient.On("Hello", "localhost").Return(nil).Times(1)
	smtpClient.On("Extension", "STARTTLS").Return(true, "").Times(1)
	smtpClient.On("StartTLS", &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         "localhost",
		MinVersion:         tls.VersionTLS13,
	}).Return(nil).Times(1)
	smtpClient.On("Extension", "AUTH").Return(true, "").Times(1)
	smtpClient.On("Auth", auth).Return(nil).Times(1)
	smtpClient.On("Mail", "flyte@flyte.org").Return(nil).Times(1)
	smtpClient.On("Rcpt", "alice@flyte.org").Return(nil).Times(1)
	smtpClient.On("Rcpt", "bob@flyte.org").Return(nil).Times(1)
	smtpClient.On("Data").Return(&stringWriter, nil).Times(1)

	smtpEmailer := createSMTPEmailer(smtpClient, &tlsConf, &auth, nil)

	err := smtpEmailer.SendEmail(context.Background(), &admin.EmailMessage{
		SubjectLine:     "subject",
		SenderEmail:     "flyte@flyte.org",
		RecipientsEmail: []string{"alice@flyte.org", "bob@flyte.org"},
		Body:            "This is an email.",
	})

	assert.True(t, strings.Contains(stringWriter.buffer, "From: sender"))
	assert.True(t, strings.Contains(stringWriter.buffer, "To: alice@flyte.org,bob@flyte.org"))
	assert.True(t, strings.Contains(stringWriter.buffer, "Subject: subject"))
	assert.True(t, strings.Contains(stringWriter.buffer, "This is an email."))
	assert.Nil(t, err)

}

func TestSendMailCreateClientError(t *testing.T) {
	auth := smtp.PlainAuth("", "user", "password", "localhost")

	tlsConf := tls.Config{
		InsecureSkipVerify: false,
		ServerName:         "localhost",
		MinVersion:         tls.VersionTLS13,
	}

	smtpClient := &notification_mocks.SMTPClient{}
	smtpClient.On("Noop").Return(errors.New("no connection")).Times(1)
	smtpClient.On("Close").Return(nil).Times(1)
	smtpClient.On("Hello", "localhost").Return(errors.New("error hello")).Times(1)

	smtpEmailer := createSMTPEmailer(smtpClient, &tlsConf, &auth, nil)

	err := smtpEmailer.SendEmail(context.Background(), &admin.EmailMessage{
		SubjectLine:     "subject",
		SenderEmail:     "flyte@flyte.org",
		RecipientsEmail: []string{"alice@flyte.org", "bob@flyte.org"},
		Body:            "This is an email.",
	})

	assert.Equal(t, flyte_errors.NewFlyteAdminErrorf(codes.Internal, "errors were seen while sending emails"), err)

}

func TestSendMailErrorMail(t *testing.T) {
	auth := smtp.PlainAuth("", "user", "password", "localhost")
	tlsConf := tls.Config{
		InsecureSkipVerify: false,
		ServerName:         "localhost",
		MinVersion:         tls.VersionTLS13,
	}

	smtpClient := &notification_mocks.SMTPClient{}
	smtpClient.On("Noop").Return(errors.New("no connection")).Times(1)
	smtpClient.On("Close").Return(nil).Times(1)
	smtpClient.On("Hello", "localhost").Return(nil).Times(1)
	smtpClient.On("Extension", "STARTTLS").Return(true, "").Times(1)
	smtpClient.On("StartTLS", &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         "localhost",
		MinVersion:         tls.VersionTLS13,
	}).Return(nil).Times(1)
	smtpClient.On("Extension", "AUTH").Return(true, "").Times(1)
	smtpClient.On("Auth", auth).Return(nil).Times(1)
	smtpClient.On("Mail", "flyte@flyte.org").Return(errors.New("error sending mail")).Times(1)

	smtpEmailer := createSMTPEmailer(smtpClient, &tlsConf, &auth, nil)

	err := smtpEmailer.SendEmail(context.Background(), &admin.EmailMessage{
		SubjectLine:     "subject",
		SenderEmail:     "flyte@flyte.org",
		RecipientsEmail: []string{"alice@flyte.org", "bob@flyte.org"},
		Body:            "This is an email.",
	})

	assert.Equal(t, flyte_errors.NewFlyteAdminErrorf(codes.Internal, "errors were seen while sending emails"), err)

}

func TestSendMailErrorRecipient(t *testing.T) {
	auth := smtp.PlainAuth("", "user", "password", "localhost")
	tlsConf := tls.Config{
		InsecureSkipVerify: false,
		ServerName:         "localhost",
		MinVersion:         tls.VersionTLS13,
	}

	smtpClient := &notification_mocks.SMTPClient{}
	smtpClient.On("Noop").Return(errors.New("no connection")).Times(1)
	smtpClient.On("Close").Return(nil).Times(1)
	smtpClient.On("Hello", "localhost").Return(nil).Times(1)
	smtpClient.On("Extension", "STARTTLS").Return(true, "").Times(1)
	smtpClient.On("StartTLS", &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         "localhost",
		MinVersion:         tls.VersionTLS13,
	}).Return(nil).Times(1)
	smtpClient.On("Extension", "AUTH").Return(true, "").Times(1)
	smtpClient.On("Auth", auth).Return(nil).Times(1)
	smtpClient.On("Mail", "flyte@flyte.org").Return(nil).Times(1)
	smtpClient.On("Rcpt", "alice@flyte.org").Return(errors.New("error adding recipient")).Times(1)

	smtpEmailer := createSMTPEmailer(smtpClient, &tlsConf, &auth, nil)

	err := smtpEmailer.SendEmail(context.Background(), &admin.EmailMessage{
		SubjectLine:     "subject",
		SenderEmail:     "flyte@flyte.org",
		RecipientsEmail: []string{"alice@flyte.org", "bob@flyte.org"},
		Body:            "This is an email.",
	})

	assert.Equal(t, flyte_errors.NewFlyteAdminErrorf(codes.Internal, "errors were seen while sending emails"), err)

}

func TestSendMailErrorData(t *testing.T) {
	auth := smtp.PlainAuth("", "user", "password", "localhost")
	tlsConf := tls.Config{
		InsecureSkipVerify: false,
		ServerName:         "localhost",
		MinVersion:         tls.VersionTLS13,
	}

	smtpClient := &notification_mocks.SMTPClient{}
	smtpClient.On("Noop").Return(errors.New("no connection")).Times(1)
	smtpClient.On("Close").Return(nil).Times(1)
	smtpClient.On("Hello", "localhost").Return(nil).Times(1)
	smtpClient.On("Extension", "STARTTLS").Return(true, "").Times(1)
	smtpClient.On("StartTLS", &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         "localhost",
		MinVersion:         tls.VersionTLS13,
	}).Return(nil).Times(1)
	smtpClient.On("Extension", "AUTH").Return(true, "").Times(1)
	smtpClient.On("Auth", auth).Return(nil).Times(1)
	smtpClient.On("Mail", "flyte@flyte.org").Return(nil).Times(1)
	smtpClient.On("Rcpt", "alice@flyte.org").Return(nil).Times(1)
	smtpClient.On("Rcpt", "bob@flyte.org").Return(nil).Times(1)
	smtpClient.On("Data").Return(nil, errors.New("error creating data writer")).Times(1)

	smtpEmailer := createSMTPEmailer(smtpClient, &tlsConf, &auth, nil)

	err := smtpEmailer.SendEmail(context.Background(), &admin.EmailMessage{
		SubjectLine:     "subject",
		SenderEmail:     "flyte@flyte.org",
		RecipientsEmail: []string{"alice@flyte.org", "bob@flyte.org"},
		Body:            "This is an email.",
	})

	assert.Equal(t, flyte_errors.NewFlyteAdminErrorf(codes.Internal, "errors were seen while sending emails"), err)

}

func TestSendMailErrorWriting(t *testing.T) {
	auth := smtp.PlainAuth("", "user", "password", "localhost")

	tlsConf := tls.Config{
		InsecureSkipVerify: false,
		ServerName:         "localhost",
		MinVersion:         tls.VersionTLS13,
	}

	stringWriter := StringWriter{buffer: "", writeErr: errors.New("error writing"), closeErr: nil}

	smtpClient := &notification_mocks.SMTPClient{}
	smtpClient.On("Noop").Return(errors.New("no connection")).Times(1)
	smtpClient.On("Close").Return(nil).Times(1)
	smtpClient.On("Hello", "localhost").Return(nil).Times(1)
	smtpClient.On("Extension", "STARTTLS").Return(true, "").Times(1)
	smtpClient.On("StartTLS", &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         "localhost",
		MinVersion:         tls.VersionTLS13,
	}).Return(nil).Times(1)
	smtpClient.On("Extension", "AUTH").Return(true, "").Times(1)
	smtpClient.On("Auth", auth).Return(nil).Times(1)
	smtpClient.On("Mail", "flyte@flyte.org").Return(nil).Times(1)
	smtpClient.On("Rcpt", "alice@flyte.org").Return(nil).Times(1)
	smtpClient.On("Rcpt", "bob@flyte.org").Return(nil).Times(1)
	smtpClient.On("Data").Return(&stringWriter, nil).Times(1)

	smtpEmailer := createSMTPEmailer(smtpClient, &tlsConf, &auth, nil)

	err := smtpEmailer.SendEmail(context.Background(), &admin.EmailMessage{
		SubjectLine:     "subject",
		SenderEmail:     "flyte@flyte.org",
		RecipientsEmail: []string{"alice@flyte.org", "bob@flyte.org"},
		Body:            "This is an email.",
	})

	assert.Equal(t, flyte_errors.NewFlyteAdminErrorf(codes.Internal, "errors were seen while sending emails"), err)

}

func TestSendMailErrorClose(t *testing.T) {
	auth := smtp.PlainAuth("", "user", "password", "localhost")

	tlsConf := tls.Config{
		InsecureSkipVerify: false,
		ServerName:         "localhost",
		MinVersion:         tls.VersionTLS13,
	}

	stringWriter := StringWriter{buffer: "", writeErr: nil, closeErr: errors.New("error writing")}

	smtpClient := &notification_mocks.SMTPClient{}
	smtpClient.On("Noop").Return(errors.New("no connection")).Times(1)
	smtpClient.On("Close").Return(nil).Times(1)
	smtpClient.On("Hello", "localhost").Return(nil).Times(1)
	smtpClient.On("Extension", "STARTTLS").Return(true, "").Times(1)
	smtpClient.On("StartTLS", &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         "localhost",
		MinVersion:         tls.VersionTLS13,
	}).Return(nil).Times(1)
	smtpClient.On("Extension", "AUTH").Return(true, "").Times(1)
	smtpClient.On("Auth", auth).Return(nil).Times(1)
	smtpClient.On("Mail", "flyte@flyte.org").Return(nil).Times(1)
	smtpClient.On("Rcpt", "alice@flyte.org").Return(nil).Times(1)
	smtpClient.On("Rcpt", "bob@flyte.org").Return(nil).Times(1)
	smtpClient.On("Data").Return(&stringWriter, nil).Times(1)

	smtpEmailer := createSMTPEmailer(smtpClient, &tlsConf, &auth, nil)

	err := smtpEmailer.SendEmail(context.Background(), &admin.EmailMessage{
		SubjectLine:     "subject",
		SenderEmail:     "flyte@flyte.org",
		RecipientsEmail: []string{"alice@flyte.org", "bob@flyte.org"},
		Body:            "This is an email.",
	})

	assert.True(t, strings.Contains(stringWriter.buffer, "From: sender"))
	assert.True(t, strings.Contains(stringWriter.buffer, "To: alice@flyte.org,bob@flyte.org"))
	assert.True(t, strings.Contains(stringWriter.buffer, "Subject: subject"))
	assert.True(t, strings.Contains(stringWriter.buffer, "This is an email."))
	assert.Equal(t, flyte_errors.NewFlyteAdminErrorf(codes.Internal, "errors were seen while sending emails"), err)

}

func createSMTPEmailer(smtpClient notification_interfaces.SMTPClient, tlsConf *tls.Config, auth *smtp.Auth, creationErr error) *SMTPEmailer {
	secretManagerMock := mocks.SecretManager{}
	secretManagerMock.On("Get", mock.Anything, "smtp_password").Return("password", nil)

	notificationsConfig := getNotificationsEmailerConfig()

	return &SMTPEmailer{
		config:        &notificationsConfig.NotificationsEmailerConfig,
		systemMetrics: newEmailMetrics(promutils.NewTestScope()),
		tlsConf:       tlsConf,
		auth:          auth,
		CreateSMTPClientFunc: func(connectString string) (notification_interfaces.SMTPClient, error) {
			return smtpClient, creationErr
		},
		smtpClient: smtpClient,
	}
}
