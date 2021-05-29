package implementations

import (
	"context"
	"io/ioutil"
	"os"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"

	"github.com/flyteorg/flyteadmin/pkg/async/notifications/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
)

type SendgridEmailer struct {
	client        *sendgrid.Client
	systemMetrics emailMetrics
}

func getEmailAddresses(addresses []string) []*mail.Email {
	sendgridAddresses := make([]*mail.Email, len(addresses))
	for idx, email := range addresses {
		// No name needed
		sendgridAddresses[idx] = mail.NewEmail("", email)
	}
	return sendgridAddresses
}

func getSendgridEmail(adminEmail admin.EmailMessage) *mail.SGMailV3 {
	m := mail.NewV3Mail()
	// This from email address is really here as a formality. For sendgrid specifically, the sender email is determined
	// from the api key that's used, not what you send along here.
	from := mail.NewEmail("Flyte Notifications", adminEmail.SenderEmail)
	content := mail.NewContent("text/html", adminEmail.Body)
	m.SetFrom(from)
	m.AddContent(content)

	personalization := mail.NewPersonalization()
	emailAddresses := getEmailAddresses(adminEmail.RecipientsEmail)
	personalization.AddTos(emailAddresses...)
	personalization.Subject = adminEmail.SubjectLine
	m.AddPersonalizations(personalization)

	return m
}

func getAPIKey(config runtimeInterfaces.EmailServerConfig) string {
	if config.APIKeyEnvVar != "" {
		return os.Getenv(config.APIKeyEnvVar)
	}
	// If environment variable not specified, assume the file is there.
	apiKeyFile, err := ioutil.ReadFile(config.APIKeyFilePath)
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(apiKeyFile))
}

func (s SendgridEmailer) SendEmail(ctx context.Context, email admin.EmailMessage) error {
	m := getSendgridEmail(email)
	s.systemMetrics.SendTotal.Inc()
	response, err := s.client.Send(m)
	if err != nil {
		logger.Errorf(ctx, "Sendgrid error sending %s", err)
		s.systemMetrics.SendError.Inc()
		return err
	}
	s.systemMetrics.SendSuccess.Inc()
	logger.Debugf(ctx, "Sendgrid sent email %s", response.Body)

	return nil
}

func NewSendGridEmailer(config runtimeInterfaces.NotificationsConfig, scope promutils.Scope) interfaces.Emailer {
	return &SendgridEmailer{
		client:        sendgrid.NewSendClient(getAPIKey(config.NotificationsEmailerConfig.EmailerConfig)),
		systemMetrics: newEmailMetrics(scope.NewSubScope("sendgrid")),
	}
}
