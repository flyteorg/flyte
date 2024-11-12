package implementations

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/async/notifications/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

type SMTPEmailer struct {
	config               *runtimeInterfaces.NotificationsEmailerConfig
	systemMetrics        emailMetrics
	tlsConf              *tls.Config
	auth                 *smtp.Auth
	smtpClient           interfaces.SMTPClient
	CreateSMTPClientFunc func(connectString string) (interfaces.SMTPClient, error)
}

func (s *SMTPEmailer) createClient(ctx context.Context) (interfaces.SMTPClient, error) {
	newClient, err := s.CreateSMTPClientFunc(s.config.EmailerConfig.SMTPServer + ":" + s.config.EmailerConfig.SMTPPort)

	if err != nil {
		return nil, s.emailError(ctx, fmt.Sprintf("Error creating email client: %s", err))
	}

	if err = newClient.Hello("localhost"); err != nil {
		return nil, s.emailError(ctx, fmt.Sprintf("Error initiating connection to SMTP server: %s", err))
	}

	if ok, _ := newClient.Extension("STARTTLS"); ok {
		if err = newClient.StartTLS(s.tlsConf); err != nil {
			return nil, s.emailError(ctx, fmt.Sprintf("Error initiating connection to SMTP server: %s", err))
		}
	}

	if ok, _ := newClient.Extension("AUTH"); ok {
		if err = newClient.Auth(*s.auth); err != nil {
			return nil, s.emailError(ctx, fmt.Sprintf("Error authenticating email client: %s", err))
		}
	}

	return newClient, nil
}

func (s *SMTPEmailer) SendEmail(ctx context.Context, email *admin.EmailMessage) error {

	if s.smtpClient == nil || s.smtpClient.Noop() != nil {

		if s.smtpClient != nil {
			err := s.smtpClient.Close()
			if err != nil {
				logger.Info(ctx, err)
			}
		}
		smtpClient, err := s.createClient(ctx)

		if err != nil {
			return s.emailError(ctx, fmt.Sprintf("Error creating SMTP email client: %s", err))
		}

		s.smtpClient = smtpClient
	}

	if err := s.smtpClient.Mail(email.SenderEmail); err != nil {
		return s.emailError(ctx, fmt.Sprintf("Error creating email instance: %s", err))
	}

	for _, recipient := range email.RecipientsEmail {
		if err := s.smtpClient.Rcpt(recipient); err != nil {
			return s.emailError(ctx, fmt.Sprintf("Error adding email recipient: %s", err))
		}
	}

	writer, err := s.smtpClient.Data()

	if err != nil {
		return s.emailError(ctx, fmt.Sprintf("Error adding email recipient: %s", err))
	}

	_, err = writer.Write([]byte(createMailBody(s.config.Sender, email)))

	if err != nil {
		return s.emailError(ctx, fmt.Sprintf("Error writing mail body: %s", err))
	}

	err = writer.Close()

	if err != nil {
		return s.emailError(ctx, fmt.Sprintf("Error closing mail body: %s", err))
	}

	s.systemMetrics.SendSuccess.Inc()
	return nil
}

func (s *SMTPEmailer) emailError(ctx context.Context, error string) error {
	s.systemMetrics.SendError.Inc()
	logger.Error(ctx, error)
	return errors.NewFlyteAdminErrorf(codes.Internal, "errors were seen while sending emails")
}

func createMailBody(emailSender string, email *admin.EmailMessage) string {
	headerMap := make(map[string]string)
	headerMap["From"] = emailSender
	headerMap["To"] = strings.Join(email.RecipientsEmail, ",")
	headerMap["Subject"] = email.SubjectLine
	headerMap["Content-Type"] = "text/html; charset=\"UTF-8\""

	mailMessage := ""

	for k, v := range headerMap {
		mailMessage += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	mailMessage += "\r\n" + email.Body

	return mailMessage
}

func NewSMTPEmailer(ctx context.Context, config runtimeInterfaces.NotificationsConfig, scope promutils.Scope, sm core.SecretManager) interfaces.Emailer {
	var tlsConfiguration *tls.Config
	emailConf := config.NotificationsEmailerConfig.EmailerConfig

	smtpPassword, err := sm.Get(ctx, emailConf.SMTPPasswordSecretName)
	if err != nil {
		logger.Debug(ctx, "No SMTP password found.")
		smtpPassword = ""
	}

	auth := smtp.PlainAuth("", emailConf.SMTPUsername, smtpPassword, emailConf.SMTPServer)

	// #nosec G402
	tlsConfiguration = &tls.Config{
		InsecureSkipVerify: emailConf.SMTPSkipTLSVerify,
		ServerName:         emailConf.SMTPServer,
	}

	return &SMTPEmailer{
		config:        &config.NotificationsEmailerConfig,
		systemMetrics: newEmailMetrics(scope.NewSubScope("smtp")),
		tlsConf:       tlsConfiguration,
		auth:          &auth,
		CreateSMTPClientFunc: func(connectString string) (interfaces.SMTPClient, error) {
			return smtp.Dial(connectString)
		},
	}
}
