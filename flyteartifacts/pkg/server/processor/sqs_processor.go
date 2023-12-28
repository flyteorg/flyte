package processor

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/NYTimes/gizmo/pubsub"
	pbcloudevents "github.com/cloudevents/sdk-go/binding/format/protobuf/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	flyteEvents "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

type PubSubProcessor struct {
	sub pubsub.Subscriber
}

func (p *PubSubProcessor) StartProcessing(ctx context.Context, handler EventsHandlerInterface) {
	for {
		logger.Warningf(context.Background(), "Starting notifications processor")
		err := p.run(ctx, handler)
		logger.Errorf(context.Background(), "error with running processor err: [%v] ", err)
		time.Sleep(1000 * 1000 * 1000 * 5)
	}
}

func (p *PubSubProcessor) run(ctx context.Context, handler EventsHandlerInterface) error {
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
			logger.Errorf(context.Background(), "failed to unmarshal JSON message [%s] from processor with err: %v", stringMsg, err)
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
		err = p.handleMessage(ctx, incomingMessageBytes, handler)
		if err != nil {
			logger.Errorf(context.Background(), "error handling message [%v] with err: %v", snsJSONFormat, err)
			//p.systemMetrics.MessageHandlingError.Inc()
			p.markMessageDone(msg)
			continue
		}
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

func (p *PubSubProcessor) handleMessage(ctx context.Context, msgBytes []byte, handler EventsHandlerInterface) error {
	ce := &event.Event{}
	err := pbcloudevents.Protobuf.Unmarshal(msgBytes, ce)
	if err != nil {
		logger.Errorf(context.Background(), "error with unmarshalling message [%v]", err)
		return err
	}

	logger.Debugf(ctx, "Cloud event received message type [%s]", ce.Type())
	// ce data should be a jsonpb Marshaled proto message, one of
	// - event.CloudEventTaskExecution
	// - event.CloudEventNodeExecution
	// - event.CloudEventWorkflowExecution
	// - event.CloudEventExecutionStart
	ceData := bytes.NewReader(ce.Data())
	unmarshaler := jsonpb.Unmarshaler{}
	// Use the type to determine which proto message to unmarshal to.
	var flyteEvent proto.Message
	if ce.Type() == "com.flyte.resource.cloudevents.TaskExecution" {
		flyteEvent = &flyteEvents.CloudEventTaskExecution{}
		err = unmarshaler.Unmarshal(ceData, flyteEvent)
	} else if ce.Type() == "com.flyte.resource.cloudevents.WorkflowExecution" {
		flyteEvent = &flyteEvents.CloudEventWorkflowExecution{}
		err = unmarshaler.Unmarshal(ceData, flyteEvent)
	} else if ce.Type() == "com.flyte.resource.cloudevents.NodeExecution" {
		flyteEvent = &flyteEvents.CloudEventNodeExecution{}
		err = unmarshaler.Unmarshal(ceData, flyteEvent)
	} else if ce.Type() == "com.flyte.resource.cloudevents.ExecutionStart" {
		flyteEvent = &flyteEvents.CloudEventExecutionStart{}
		err = unmarshaler.Unmarshal(ceData, flyteEvent)
	} else {
		logger.Infof(ctx, "Ignoring cloud event type [%s]", ce.Type())
		return nil
	}

	if err != nil {
		logger.Errorf(ctx, "error unmarshalling message [%s] [%v]", ce.Type(), err)
		return err
	}
	err = handler.HandleEvent(ctx, ce, flyteEvent)
	if err != nil {
		logger.Errorf(context.Background(), "error handling event on topic [%v]", err)
		return err
	}

	return nil
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
