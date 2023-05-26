package implementations

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/golang/protobuf/proto"
	"time"

	"github.com/NYTimes/gizmo/pubsub"
	"github.com/flyteorg/flyteadmin/pkg/async"
	"github.com/flyteorg/flyteadmin/pkg/async/notifications/interfaces"
	webhookInterfaces "github.com/flyteorg/flyteadmin/pkg/async/webhook/interfaces"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
)

type Processor struct {
	sub           pubsub.Subscriber
	webhook       webhookInterfaces.Webhook
	systemMetrics processorSystemMetrics
}

func (p *Processor) StartProcessing() {
	for {
		logger.Warningf(context.Background(), "Starting webhook processor")
		err := p.run()
		logger.Errorf(context.Background(), "error with running processor err: [%v] ", err)
		time.Sleep(async.RetryDelay)
	}
}

func (p *Processor) run() error {
	var payload admin.WebhookPayload
	var err error
	for msg := range p.sub.Start() {
		p.systemMetrics.MessageTotal.Inc()
		// Currently this is safe because Gizmo takes a string and casts it to a byte array.
		stringMsg := string(msg.Message())

		var snsJSONFormat map[string]interface{}

		// At Lyft, SNS populates SQS. This results in the message body of SQS having the SNS message format.
		// The message format is documented here: https://docs.aws.amazon.com/sns/latest/dg/sns-message-and-json-formats.html
		// The notification published is stored in the message field after unmarshalling the SQS message.
		if err := json.Unmarshal(msg.Message(), &snsJSONFormat); err != nil {
			p.systemMetrics.MessageDecodingError.Inc()
			logger.Errorf(context.Background(), "failed to unmarshall JSON message [%s] from processor with err: %v", stringMsg, err)
			p.markMessageDone(msg)
			continue
		}

		var value interface{}
		var ok bool
		var valueString string

		if value, ok = snsJSONFormat["Message"]; !ok {
			logger.Errorf(context.Background(), "failed to retrieve message from unmarshalled JSON object [%s]", stringMsg)
			p.systemMetrics.MessageDataError.Inc()
			p.markMessageDone(msg)
			continue
		}

		if valueString, ok = value.(string); !ok {
			p.systemMetrics.MessageDataError.Inc()
			logger.Errorf(context.Background(), "failed to retrieve notification message (in string format) from unmarshalled JSON object for message [%s]", stringMsg)
			p.markMessageDone(msg)
			continue
		}

		// The Publish method for SNS Encodes the notification using Base64 then stringifies it before
		// setting that as the message body for SNS. Do the inverse to retrieve the notification.
		notificationBytes, err := base64.StdEncoding.DecodeString(valueString)
		if err != nil {
			logger.Errorf(context.Background(), "failed to Base64 decode from message string [%s] from message [%s] with err: %v", valueString, stringMsg, err)
			p.systemMetrics.MessageDecodingError.Inc()
			p.markMessageDone(msg)
			continue
		}

		if err = proto.Unmarshal(notificationBytes, &payload); err != nil {
			logger.Debugf(context.Background(), "failed to unmarshal to notification object from decoded string[%s] from message [%s] with err: %v", valueString, stringMsg, err)
			p.systemMetrics.MessageDecodingError.Inc()
			p.markMessageDone(msg)
			continue
		}

		logger.Info(context.Background(), "Processor is sending message to webhook endpoint")
		// Send message to webhook
		if err = p.webhook.Post(context.Background(), payload); err != nil {
			p.systemMetrics.MessageProcessorError.Inc()
			logger.Errorf(context.Background(), "Error sending an message [%v] to webhook endpoint with err: %v", payload, err)
		} else {
			p.systemMetrics.MessageSuccess.Inc()
		}

		p.markMessageDone(msg)
	}

	if err = p.sub.Err(); err != nil {
		p.systemMetrics.ChannelClosedError.Inc()
		logger.Warningf(context.Background(), "The stream for the subscriber channel closed with err: %v", err)
	}

	return err
}

func (p *Processor) markMessageDone(message pubsub.SubscriberMessage) {
	if err := message.Done(); err != nil {
		p.systemMetrics.MessageDoneError.Inc()
		logger.Errorf(context.Background(), "failed to mark message as Done() in processor with err: %v", err)
	}
}

func (p *Processor) StopProcessing() error {
	// Note: If the underlying channel is already closed, then Stop() will return an error.
	err := p.sub.Stop()
	if err != nil {
		p.systemMetrics.StopError.Inc()
		logger.Errorf(context.Background(), "Failed to stop the subscriber channel gracefully with err: %v", err)
	}
	return err
}

func NewWebhookProcessor(sub pubsub.Subscriber, webhook webhookInterfaces.Webhook, scope promutils.Scope) interfaces.Processor {
	return &Processor{
		sub:           sub,
		webhook:       webhook,
		systemMetrics: newProcessorSystemMetrics(scope.NewSubScope("webhook_processor")),
	}
}
