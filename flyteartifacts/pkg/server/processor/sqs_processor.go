package processor

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/NYTimes/gizmo/pubsub"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"time"
)

type PubSubProcessor struct {
	sub pubsub.Subscriber
}

func (p *PubSubProcessor) StartProcessing(ctx context.Context) {
	for {
		logger.Warningf(context.Background(), "Starting notifications processor")
		err := p.run()
		logger.Errorf(context.Background(), "error with running processor err: [%v] ", err)
		time.Sleep(1000 * 1000 * 1000 * 5)
	}
}

// todo: i think we can easily add context
func (p *PubSubProcessor) run() error {
	var err error
	for msg := range p.sub.Start() {
		// gatepr: add metrics
		// Currently this is safe because Gizmo takes a string and casts it to a byte array.
		stringMsg := string(msg.Message())

		var snsJSONFormat map[string]interface{}

		// Typically, SNS populates SQS. This results in the message body of SQS having the SNS message format.
		// The message format is documented here: https://docs.aws.amazon.com/sns/latest/dg/sns-message-and-json-formats.html
		// The notification published is stored in the message field after unmarshalling the SQS message.
		if err := json.Unmarshal(msg.Message(), &snsJSONFormat); err != nil {
			//p.systemMetrics.MessageDecodingError.Inc()
			logger.Errorf(context.Background(), "failed to unmarshall JSON message [%s] from processor with err: %v", stringMsg, err)
			p.markMessageDone(msg)
			continue
		}

		var value interface{}
		var ok bool
		var valueString string

		if value, ok = snsJSONFormat["Message"]; !ok {
			logger.Errorf(context.Background(), "failed to retrieve message from unmarshalled JSON object [%s]", stringMsg)
			// p.systemMetrics.MessageDataError.Inc()
			p.markMessageDone(msg)
			continue
		}

		if valueString, ok = value.(string); !ok {
			// p.systemMetrics.MessageDataError.Inc()
			logger.Errorf(context.Background(), "failed to retrieve notification message (in string format) from unmarshalled JSON object for message [%s]", stringMsg)
			p.markMessageDone(msg)
			continue
		}

		// The Publish method for SNS Encodes the notification using Base64 then stringifies it before
		// setting that as the message body for SNS. Do the inverse to retrieve the notification.
		incomingMessageBytes, err := base64.StdEncoding.DecodeString(valueString)
		if err != nil {
			logger.Errorf(context.Background(), "failed to Base64 decode from message string [%s] from message [%s] with err: %v", valueString, stringMsg, err)
			//p.systemMetrics.MessageDecodingError.Inc()
			p.markMessageDone(msg)
			continue
		}
		logger.Infof(context.Background(), "incoming message bytes [%v]", incomingMessageBytes)
		// handle message
		p.markMessageDone(msg)
	}

	// According to https://github.com/NYTimes/gizmo/blob/f2b3deec03175b11cdfb6642245a49722751357f/pubsub/pubsub.go#L36-L39,
	// the channel backing the subscriber will just close if there is an error. The call to Err() is needed to identify
	// there was an error in the channel or there are no more messages left (resulting in no errors when calling Err()).
	if err = p.sub.Err(); err != nil {
		//p.systemMetrics.ChannelClosedError.Inc()
		logger.Warningf(context.Background(), "The stream for the subscriber channel closed with err: %v", err)
	}

	// If there are no errors, nil will be returned.
	return err
}

func (p *PubSubProcessor) markMessageDone(message pubsub.SubscriberMessage) {
	if err := message.Done(); err != nil {
		//p.systemMetrics.MessageDoneError.Inc()
		logger.Errorf(context.Background(), "failed to mark message as Done() in processor with err: %v", err)
	}
}

func (p *PubSubProcessor) StopProcessing() error {
	// Note: If the underlying channel is already closed, then Stop() will return an error.
	err := p.sub.Stop()
	if err != nil {
		//p.systemMetrics.StopError.Inc()
		logger.Errorf(context.Background(), "Failed to stop the subscriber channel gracefully with err: %v", err)
	}
	return err
}

func NewPubSubProcessor(sub pubsub.Subscriber, _ promutils.Scope) *PubSubProcessor {
	return &PubSubProcessor{
		sub: sub,
	}
}
