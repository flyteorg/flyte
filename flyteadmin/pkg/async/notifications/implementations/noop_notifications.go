package implementations

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/async/notifications/interfaces"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/logger"

	"strings"

	"github.com/golang/protobuf/proto"
)

// Email to use when there is no email configuration.
type NoopEmail struct{}

func (n *NoopEmail) SendEmail(ctx context.Context, email admin.EmailMessage) error {
	logger.Debugf(ctx, "received noop SendEmail request with subject [%s] and recipient [%s]",
		email.SubjectLine, strings.Join(email.RecipientsEmail, ","))
	return nil
}

func NewNoopEmail() interfaces.Emailer {
	return &NoopEmail{}
}

type NoopPublish struct{}

func (n *NoopPublish) Publish(ctx context.Context, notificationType string, msg proto.Message) error {
	logger.Debugf(ctx, "call to noop publish with notification type [%s] and proto message [%s]", notificationType, msg.String())
	return nil
}

func NewNoopPublish() interfaces.Publisher {
	return &NoopPublish{}
}

type NoopProcess struct{}

func (n *NoopProcess) StartProcessing() {
	logger.Debug(context.Background(), "call to noop start processing.")
}

func (n *NoopProcess) StopProcessing() error {
	logger.Debug(context.Background(), "call to noop stop processing.")
	return nil
}

func NewNoopProcess() interfaces.Processor {
	return &NoopProcess{}
}
