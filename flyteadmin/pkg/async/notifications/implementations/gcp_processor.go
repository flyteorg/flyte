package implementations

import (
	"context"
	"time"

	"github.com/NYTimes/gizmo/pubsub"
	"github.com/flyteorg/flyteadmin/pkg/async"
	"github.com/flyteorg/flyteadmin/pkg/async/notifications/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/golang/protobuf/proto"
)

// TODO: Add a counter that encompasses the publisher stats grouped by project and domain.
type GcpProcessor struct {
	email interfaces.Emailer
	interfaces.BaseProcessor
}

func NewGcpProcessor(sub pubsub.Subscriber, emailer interfaces.Emailer, scope promutils.Scope) interfaces.Processor {
	return &GcpProcessor{
		email: emailer,
		BaseProcessor: interfaces.BaseProcessor{
			Sub:           sub,
			SystemMetrics: interfaces.NewProcessorSystemMetrics(scope.NewSubScope("processor")),
		},
	}
}

func (p *GcpProcessor) StartProcessing() {
	for {
		logger.Warningf(context.Background(), "Starting GCP notifications processor")
		err := p.run()
		logger.Errorf(context.Background(), "error with running GCP processor err: [%v] ", err)
		time.Sleep(async.RetryDelay)
	}
}

func (p *GcpProcessor) run() error {
	var emailMessage admin.EmailMessage

	for msg := range p.Sub.Start() {
		p.SystemMetrics.MessageTotal.Inc()

		if err := proto.Unmarshal(msg.Message(), &emailMessage); err != nil {
			logger.Debugf(context.Background(), "failed to unmarshal to notification object message [%s] with err: %v", string(msg.Message()), err)
			p.SystemMetrics.MessageDecodingError.Inc()
			p.MarkMessageDone(msg)
			continue
		}

		if err := p.email.SendEmail(context.Background(), emailMessage); err != nil {
			p.SystemMetrics.MessageProcessorError.Inc()
			logger.Errorf(context.Background(), "Error sending an email message for message [%s] with emailM with err: %v", emailMessage.String(), err)
		} else {
			p.SystemMetrics.MessageSuccess.Inc()
		}

		p.MarkMessageDone(msg)
	}

	// According to https://github.com/NYTimes/gizmo/blob/f2b3deec03175b11cdfb6642245a49722751357f/pubsub/pubsub.go#L36-L39,
	// the channel backing the subscriber will just close if there is an error. The call to Err() is needed to identify
	// there was an error in the channel or there are no more messages left (resulting in no errors when calling Err()).
	if err := p.Sub.Err(); err != nil {
		p.SystemMetrics.ChannelClosedError.Inc()
		logger.Warningf(context.Background(), "The stream for the subscriber channel closed with err: %v", err)
		return err
	}

	return nil
}
