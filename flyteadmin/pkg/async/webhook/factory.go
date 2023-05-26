package webhook

import (
	"context"
	"time"

	"github.com/NYTimes/gizmo/pubsub"
	gizmoAWS "github.com/NYTimes/gizmo/pubsub/aws"
	"github.com/flyteorg/flyteadmin/pkg/async"
	"github.com/flyteorg/flyteadmin/pkg/async/notifications/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/async/webhook/implementations"
	"github.com/flyteorg/flytestdlib/logger"

	webhookInterfaces "github.com/flyteorg/flyteadmin/pkg/async/webhook/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/promutils"
)

var enable64decoding = false

func GetWebhook(config runtimeInterfaces.WebHookConfig, scope promutils.Scope) webhookInterfaces.Webhook {
	// TODO: Get others webhooks
	return implementations.NewSlackWebhook(config, scope)
}

func NewWebhookProcessors(config runtimeInterfaces.WebhookNotificationsConfig, scope promutils.Scope) []interfaces.Processor {
	reconnectAttempts := config.ReconnectAttempts
	reconnectDelay := time.Duration(config.ReconnectDelaySeconds) * time.Second
	var sub pubsub.Subscriber
	var processors []interfaces.Processor

	for _, cfg := range config.WebhooksConfig {
		if len(config.AWSConfig.Region) != 0 {
			sqsConfig := gizmoAWS.SQSConfig{
				QueueName:           cfg.NotificationsProcessorConfig.QueueName,
				QueueOwnerAccountID: cfg.NotificationsProcessorConfig.AccountID,
				// The AWS configuration type uses SNS to SQS for notifications.
				// Gizmo by default will decode the SQS message using Base64 decoding.
				// However, the message body of SQS is the SNS message format which isn't Base64 encoded.
				ConsumeBase64: &enable64decoding,
			}
			sqsConfig.Region = config.AWSConfig.Region
			var err error
			err = async.Retry(reconnectAttempts, reconnectDelay, func() error {
				sub, err = gizmoAWS.NewSubscriber(sqsConfig)
				if err != nil {
					logger.Warnf(context.TODO(), "Failed to initialize new gizmo aws subscriber with config [%+v] and err: %v", sqsConfig, err)
				}
				return err
			})
			if err != nil {
				panic(err)
			}
		}

		webhook := GetWebhook(cfg, scope)
		processors = append(processors, implementations.NewWebhookProcessor(sub, webhook, scope))
	}
	return processors
}
