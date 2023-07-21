package interfaces

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/NYTimes/gizmo/pubsub"
	gizmoGCP "github.com/NYTimes/gizmo/pubsub/gcp"
	"github.com/flyteorg/flytestdlib/logger"
)

// Exposes the common methods required for a subscriber.
// There is one ProcessNotification per type.
type Processor interface {

	// Starts processing messages from the underlying subscriber.
	// If the channel closes gracefully, no error will be returned.
	// If the underlying channel experiences errors,
	// an error is returned and the channel is closed.
	StartProcessing()

	// This should be invoked when the application is shutting down.
	// If StartProcessing() returned an error, StopProcessing() will return an error because
	// the channel was already closed.
	StopProcessing() error
}

type BaseProcessor struct {
	Sub           pubsub.Subscriber
	SystemMetrics ProcessorSystemMetrics
}

// FromPubSubMessage Parse the message from GCP PubSub and return the message subject and the message body.
func (p *BaseProcessor) FromPubSubMessage(msg pubsub.SubscriberMessage) (string, []byte, bool) {
	var gcpMsg *gizmoGCP.SubMessage
	var ok bool
	p.SystemMetrics.MessageTotal.Inc()
	if gcpMsg, ok = msg.(*gizmoGCP.SubMessage); !ok {
		logger.Errorf(context.Background(), "failed to cast message [%v] to gizmoGCP.SubMessage", msg)
		p.SystemMetrics.MessageDataError.Inc()
		p.MarkMessageDone(msg)
		return "", nil, false
	}
	subject := gcpMsg.Attributes["key"]
	return subject, gcpMsg.Message(), true
}

// FromSQSMessage Parse the message from AWS SQS and return the message subject and the message body.
func (p *BaseProcessor) FromSQSMessage(msg pubsub.SubscriberMessage) (string, []byte, bool) {
	// Currently this is safe because Gizmo takes a string and casts it to a byte array.
	stringMsg := string(msg.Message())
	var snsJSONFormat map[string]interface{}

	if err := json.Unmarshal(msg.Message(), &snsJSONFormat); err != nil {
		p.SystemMetrics.MessageDecodingError.Inc()
		logger.Errorf(context.Background(), "failed to unmarshall JSON message [%s] from processor with err: %v", stringMsg, err)
		p.MarkMessageDone(msg)
		return "", nil, false
	}

	var value interface{}
	var ok bool
	var valueString string
	var subject string

	if value, ok = snsJSONFormat["Message"]; !ok {
		logger.Errorf(context.Background(), "failed to retrieve message from unmarshalled JSON object [%s]", stringMsg)
		p.SystemMetrics.MessageDataError.Inc()
		p.MarkMessageDone(msg)
		return "", nil, false
	}

	if valueString, ok = value.(string); !ok {
		p.SystemMetrics.MessageDataError.Inc()
		logger.Errorf(context.Background(), "failed to retrieve notification message (in string format) from unmarshalled JSON object for message [%s]", stringMsg)
		p.MarkMessageDone(msg)
		return "", nil, false
	}

	if value, ok = snsJSONFormat["Subject"]; !ok {
		logger.Errorf(context.Background(), "failed to retrieve message type from unmarshalled JSON object [%s]", stringMsg)
		p.SystemMetrics.MessageDataError.Inc()
		p.MarkMessageDone(msg)
		return "", nil, false
	}

	if subject, ok = value.(string); !ok {
		p.SystemMetrics.MessageDataError.Inc()
		logger.Errorf(context.Background(), "failed to retrieve notification message type (in string format) from unmarshalled JSON object for message [%s]", stringMsg)
		p.MarkMessageDone(msg)
		return "", nil, false
	}

	// The Publish method for SNS Encodes the notification using Base64 then stringifies it before
	// setting that as the message body for SNS. Do the inverse to retrieve the notification.
	messageBytes, err := base64.StdEncoding.DecodeString(valueString)
	if err != nil {
		logger.Errorf(context.Background(), "failed to Base64 decode from message string [%s] from message [%s] with err: %v", valueString, stringMsg, err)
		p.SystemMetrics.MessageDecodingError.Inc()
		p.MarkMessageDone(msg)
		return "", nil, false
	}
	return subject, messageBytes, true
}

func (p *BaseProcessor) StopProcessing() error {
	// Note: If the underlying channel is already closed, then Stop() will return an error.
	err := p.Sub.Stop()
	if err != nil {
		p.SystemMetrics.StopError.Inc()
		logger.Errorf(context.Background(), "Failed to stop the subscriber channel gracefully with err: %v", err)
	}
	return err
}

func (p *BaseProcessor) MarkMessageDone(message pubsub.SubscriberMessage) {
	if err := message.Done(); err != nil {
		p.SystemMetrics.MessageDoneError.Inc()
		logger.Errorf(context.Background(), "failed to mark message as Done() in processor with err: %v", err)
	}
}
