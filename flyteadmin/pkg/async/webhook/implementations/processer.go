package implementations

import (
	"context"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/util"
	repoInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/golang/protobuf/proto"

	"github.com/NYTimes/gizmo/pubsub"
	"github.com/flyteorg/flyteadmin/pkg/async"
	"github.com/flyteorg/flyteadmin/pkg/async/notifications"
	"github.com/flyteorg/flyteadmin/pkg/async/notifications/interfaces"
	webhookInterfaces "github.com/flyteorg/flyteadmin/pkg/async/webhook/interfaces"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
)

type Processor struct {
	subType string
	webhook webhookInterfaces.Webhook
	db      repoInterfaces.Repository
	interfaces.BaseProcessor
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
	var subject string
	var messageByte []byte
	var ok bool

	for msg := range p.Sub.Start() {
		p.SystemMetrics.MessageTotal.Inc()
		stringMsg := string(msg.Message())
		if p.subType == common.AWS {
			subject, messageByte, ok = p.FromSQSMessage(msg)
		} else {
			subject, messageByte, ok = p.FromPubSubMessage(msg)
		}
		if !ok {
			continue
		}

		if subject != proto.MessageName(&admin.WorkflowExecutionEventRequest{}) {
			p.MarkMessageDone(msg)
			continue
		}

		if err = proto.Unmarshal(messageByte, &request); err != nil {
			logger.Errorf(context.Background(), "failed to unmarshal to notification object from decoded string from message [%s] with err: %v", stringMsg, err)
			p.SystemMetrics.MessageDecodingError.Inc()
			p.MarkMessageDone(msg)
			continue
		}

		if !common.IsExecutionTerminal(request.Event.Phase) {
			p.MarkMessageDone(msg)
			continue
		}

		executionModel, err := util.GetExecutionModel(context.Background(), p.db, *request.Event.ExecutionId)
		if err != nil {
			p.SystemMetrics.MessageProcessorError.Inc()
			logger.Errorf(context.Background(), "failed to retrieve execution model for execution [%+v] from message [%s] with err: %v", request.Event.ExecutionId, stringMsg, err)
			p.MarkMessageDone(msg)
			continue
		}
		adminExecution, err := transformers.FromExecutionModel(context.Background(), *executionModel, transformers.DefaultExecutionTransformerOptions)
		if err != nil {
			p.SystemMetrics.MessageProcessorError.Inc()
			logger.Errorf(context.Background(), "failed to transform execution model [%+v] from message [%s] with err: %v", executionModel, stringMsg, err)
			p.MarkMessageDone(msg)
			continue
		}

		payload.Message = notifications.SubstituteParameters(p.webhook.GetConfig().Payload, request, adminExecution)
		logger.Info(context.Background(), "Processor is sending message to webhook endpoint")
		if err = p.webhook.Post(context.Background(), payload); err != nil {
			p.SystemMetrics.MessageProcessorError.Inc()
			logger.Errorf(context.Background(), "Error sending an message [%v] to webhook endpoint with err: %v", payload, err)
		} else {
			p.SystemMetrics.MessageSuccess.Inc()
		}

		p.MarkMessageDone(msg)
	}

	if err = p.Sub.Err(); err != nil {
		p.SystemMetrics.ChannelClosedError.Inc()
		logger.Errorf(context.Background(), "The stream for the subscriber channel closed with err: %v", err)
	}

	return err
}

func NewWebhookProcessor(subType string, sub pubsub.Subscriber, webhook webhookInterfaces.Webhook, db repoInterfaces.Repository, scope promutils.Scope) interfaces.Processor {
	if subType != common.AWS && subType != common.GCP {
		panic("unknown subscriber type [" + subType + "]")
	}

	return &Processor{
		subType: subType,
		webhook: webhook,
		db:      db,
		BaseProcessor: interfaces.BaseProcessor{
			Sub:           sub,
			SystemMetrics: interfaces.NewProcessorSystemMetrics(scope.NewSubScope("webhook_processor")),
		},
	}
}
