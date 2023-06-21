package implementations

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/util"
	repoInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/golang/protobuf/proto"
	"time"

	"github.com/NYTimes/gizmo/pubsub"
	"github.com/flyteorg/flyteadmin/pkg/async"
	"github.com/flyteorg/flyteadmin/pkg/async/notifications"
	"github.com/flyteorg/flyteadmin/pkg/async/notifications/interfaces"
	webhookInterfaces "github.com/flyteorg/flyteadmin/pkg/async/webhook/interfaces"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
)

type Processor struct {
	sub           pubsub.Subscriber
	webhook       webhookInterfaces.Webhook
	db            repoInterfaces.Repository
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
	var request admin.WorkflowExecutionEventRequest
	var err error
	for msg := range p.sub.Start() {
		p.systemMetrics.MessageTotal.Inc()
		// Currently this is safe because Gizmo takes a string and casts it to a byte array.
		stringMsg := string(msg.Message())
		var snsJSONFormat map[string]interface{}

		if err := json.Unmarshal(msg.Message(), &snsJSONFormat); err != nil {
			p.systemMetrics.MessageDecodingError.Inc()
			logger.Errorf(context.Background(), "failed to unmarshall JSON message [%s] from processor with err: %v", stringMsg, err)
			p.markMessageDone(msg)
			continue
		}

		var value interface{}
		var ok bool
		var valueString string
		var messageType string

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

		if value, ok = snsJSONFormat["Subject"]; !ok {
			logger.Errorf(context.Background(), "failed to retrieve message type from unmarshalled JSON object [%s]", stringMsg)
			p.systemMetrics.MessageDataError.Inc()
			p.markMessageDone(msg)
			continue
		}

		if messageType, ok = value.(string); !ok {
			p.systemMetrics.MessageDataError.Inc()
			logger.Errorf(context.Background(), "failed to retrieve notification message type (in string format) from unmarshalled JSON object for message [%s]", stringMsg)
			p.markMessageDone(msg)
			continue
		}

		if messageType != proto.MessageName(&admin.WorkflowExecutionEventRequest{}) {
			p.markMessageDone(msg)
			continue
		}

		// The Publish method for SNS Encodes the notification using Base64 then stringifies it before
		// setting that as the message body for SNS. Do the inverse to retrieve the notification.
		requestBytes, err := base64.StdEncoding.DecodeString(valueString)
		if err != nil {
			logger.Errorf(context.Background(), "failed to Base64 decode from message string [%s] from message [%s] with err: %v", valueString, stringMsg, err)
			p.systemMetrics.MessageDecodingError.Inc()
			p.markMessageDone(msg)
			continue
		}

		if err = proto.Unmarshal(requestBytes, &request); err != nil {
			logger.Errorf(context.Background(), "failed to unmarshal to notification object from decoded string[%s] from message [%s] with err: %v", valueString, stringMsg, err)
			p.systemMetrics.MessageDecodingError.Inc()
			p.markMessageDone(msg)
			continue
		}

		if common.IsExecutionTerminal(request.Event.Phase) == false {
			p.markMessageDone(msg)
			continue
		}

		executionModel, err := util.GetExecutionModel(context.Background(), p.db, *request.Event.ExecutionId)
		adminExecution, err := transformers.FromExecutionModel(*executionModel, transformers.DefaultExecutionTransformerOptions)
		payload.Message = notifications.SubstituteParameters(p.webhook.GetConfig().Payload, request, adminExecution)
		logger.Info(context.Background(), "Processor is sending message to webhook endpoint")
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
		logger.Errorf(context.Background(), "The stream for the subscriber channel closed with err: %v", err)
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

func NewWebhookProcessor(sub pubsub.Subscriber, webhook webhookInterfaces.Webhook, db repoInterfaces.Repository, scope promutils.Scope) interfaces.Processor {
	return &Processor{
		sub:           sub,
		webhook:       webhook,
		db:            db,
		systemMetrics: newProcessorSystemMetrics(scope.NewSubScope("webhook_processor")),
	}
}
