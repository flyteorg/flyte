package implementations

import (
	"crypto/tls"
	"fmt"
	"github.com/flyteorg/flyte/flyteadmin/pkg/async/notifications/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"net/smtp"
	"strings"
)

type SmtpEmailer struct {
	config        *runtimeInterfaces.NotificationsEmailerConfig
	systemMetrics emailMetrics
	tlsConf       *tls.Config
	auth          *smtp.Auth
}

func (s *SmtpEmailer) SendEmail(ctx context.Context, email admin.EmailMessage) error {

	newClient, err := smtp.Dial(s.config.EmailerConfig.SmtpServer + ":" + s.config.EmailerConfig.SmtpPort)

	if err != nil {
		return s.emailError(ctx, fmt.Sprintf("Error creating email client: %s", err))
	}

	defer newClient.Close()

	if err = newClient.Hello("localhost"); err != nil {
		return s.emailError(ctx, fmt.Sprintf("Error initiating connection to SMTP server: %s", err))
	}

	if ok, _ := newClient.Extension("STARTTLS"); ok {
		if err = newClient.StartTLS(s.tlsConf); err != nil {
			return err
		}
	}

	if ok, _ := newClient.Extension("AUTH"); ok {
		if err = newClient.Auth(*s.auth); err != nil {
			return s.emailError(ctx, fmt.Sprintf("Error authenticating email client: %s", err))
		}
	}

	if err = newClient.Mail(email.SenderEmail); err != nil {
		return s.emailError(ctx, fmt.Sprintf("Error creating email instance: %s", err))
	}

	for _, recipient := range email.RecipientsEmail {
		if err = newClient.Rcpt(recipient); err != nil {
			logger.Errorf(ctx, "Error adding email recipient: %s", err)
		}
	}

	writer, err := newClient.Data()

	if err != nil {
		return s.emailError(ctx, fmt.Sprintf("Error adding email recipient: %s", err))
	}

	mailBody, err := createMailBody(s.config.Sender, email)

	if err != nil {
		return s.emailError(ctx, fmt.Sprintf("Error creating email body: %s", err))
	}

	_, err = writer.Write([]byte(mailBody))

	if err != nil {
		return s.emailError(ctx, fmt.Sprintf("Error writing mail body: %s", err))
	}

	err = writer.Close()

	if err != nil {
		return s.emailError(ctx, fmt.Sprintf("Error closing mail body: %s", err))
	}

	err = newClient.Quit()

	if err != nil {
		return s.emailError(ctx, fmt.Sprintf("Error quitting mail agent: %s", err))
	}

	s.systemMetrics.SendSuccess.Inc()
	return nil
}

func (s *SmtpEmailer) emailError(ctx context.Context, error string) error {
	s.systemMetrics.SendError.Inc()
	logger.Error(ctx, error)
	return errors.NewFlyteAdminErrorf(codes.Internal, "errors were seen while sending emails")
}

func createMailBody(emailSender string, email admin.EmailMessage) (string, error) {
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

	return mailMessage, nil
}

func NewSmtpEmailer(ctx context.Context, config runtimeInterfaces.NotificationsConfig, scope promutils.Scope, sm core.SecretManager) interfaces.Emailer {
	var tlsConfiguration *tls.Config
	emailConf := config.NotificationsEmailerConfig.EmailerConfig

	smtpPassword, err := sm.Get(ctx, emailConf.SmtpPasswordSecretName)
	if err != nil {
		logger.Debug(ctx, "No SMTP password found.")
		smtpPassword = ""
	}

	auth := smtp.PlainAuth("", emailConf.SmtpUsername, smtpPassword, emailConf.SmtpServer)

	tlsConfiguration = &tls.Config{
		InsecureSkipVerify: emailConf.SmtpSkipTLSVerify,
		ServerName:         emailConf.SmtpServer,
	}

	return &SmtpEmailer{
		config:        &config.NotificationsEmailerConfig,
		systemMetrics: newEmailMetrics(scope.NewSubScope("smtp")),
		tlsConf:       tlsConfiguration,
		auth:          &auth,
	}
}
