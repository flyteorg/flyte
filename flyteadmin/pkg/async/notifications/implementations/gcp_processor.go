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
	sub           pubsub.Subscriber
	email         interfaces.Emailer
	systemMetrics processorSystemMetrics
}

func NewGcpProcessor(sub pubsub.Subscriber, emailer interfaces.Emailer, scope promutils.Scope) interfaces.Processor {
	return &GcpProcessor{
		sub:           sub,
		email:         emailer,
		systemMetrics: newProcessorSystemMetrics(scope.NewSubScope("gcp_processor")),
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

	for msg := range p.sub.Start() {
		p.systemMetrics.MessageTotal.Inc()

		if err := proto.Unmarshal(msg.Message(), &emailMessage); err != nil {
			logger.Debugf(context.Background(), "failed to unmarshal to notification object message [%s] with err: %v", string(msg.Message()), err)
			p.systemMetrics.MessageDecodingError.Inc()
			p.markMessageDone(msg)
			continue
		}

		if err := p.email.SendEmail(context.Background(), emailMessage); err != nil {
			p.systemMetrics.MessageProcessorError.Inc()
			logger.Errorf(context.Background(), "Error sending an email message for message [%s] with emailM with err: %v", emailMessage.String(), err)
		} else {
			p.systemMetrics.MessageSuccess.Inc()
		}

		p.markMessageDone(msg)
	}

	// According to https://github.com/NYTimes/gizmo/blob/f2b3deec03175b11cdfb6642245a49722751357f/pubsub/pubsub.go#L36-L39,
	// the channel backing the subscriber will just close if there is an error. The call to Err() is needed to identify
	// there was an error in the channel or there are no more messages left (resulting in no errors when calling Err()).
	if err := p.sub.Err(); err != nil {
		p.systemMetrics.ChannelClosedError.Inc()
		logger.Warningf(context.Background(), "The stream for the subscriber channel closed with err: %v", err)
		return err
	}

	return nil
}

func (p *GcpProcessor) markMessageDone(message pubsub.SubscriberMessage) {
	if err := message.Done(); err != nil {
		p.systemMetrics.MessageDoneError.Inc()
		logger.Errorf(context.Background(), "failed to mark message as Done() in processor with err: %v", err)
	}
}

func (p *GcpProcessor) StopProcessing() error {
	// Note: If the underlying channel is already closed, then Stop() will return an error.
	if err := p.sub.Stop(); err != nil {
		p.systemMetrics.StopError.Inc()
		logger.Errorf(context.Background(), "Failed to stop the subscriber channel gracefully with err: %v", err)
		return err
	}

	return nil
}
