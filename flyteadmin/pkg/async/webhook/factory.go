package webhook

import (
	"context"
	"fmt"
	gizmoGCP "github.com/NYTimes/gizmo/pubsub/gcp"
	"github.com/flyteorg/flyteadmin/pkg/common"
	"time"

	repoInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"

	"github.com/NYTimes/gizmo/pubsub"
	gizmoAWS "github.com/NYTimes/gizmo/pubsub/aws"
	"github.com/flyteorg/flyteadmin/pkg/async"
	notificationsImplementations "github.com/flyteorg/flyteadmin/pkg/async/notifications/implementations"
	"github.com/flyteorg/flyteadmin/pkg/async/notifications/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/async/webhook/implementations"
	"github.com/flyteorg/flytestdlib/logger"

	webhookInterfaces "github.com/flyteorg/flyteadmin/pkg/async/webhook/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/promutils"
)

var enable64decoding = false

func GetWebhook(config runtimeInterfaces.WebHookConfig, scope promutils.Scope) webhookInterfaces.Webhook {
	switch config.Name {
	case implementations.Slack:
		return implementations.NewSlackWebhook(config, scope)
	default:
		panic(fmt.Errorf("no matching webhook implementation for %s", config.Name))
	}
}

func NewWebhookProcessors(db repoInterfaces.Repository, config runtimeInterfaces.WebhookNotificationsConfig, scope promutils.Scope) []interfaces.Processor {
	reconnectAttempts := config.ReconnectAttempts
	reconnectDelay := time.Duration(config.ReconnectDelaySeconds) * time.Second
	var sub pubsub.Subscriber
	var processors []interfaces.Processor
	var err error

	for _, cfg := range config.WebhooksConfig {
		var processor interfaces.Processor
		switch config.Type {
		case common.AWS:
			sqsConfig := gizmoAWS.SQSConfig{
				QueueName:           cfg.NotificationsProcessorConfig.QueueName,
				QueueOwnerAccountID: cfg.NotificationsProcessorConfig.AccountID,
				// The AWS configuration type uses SNS to SQS for notifications.
				// Gizmo by default will decode the SQS message using Base64 decoding.
				// However, the message body of SQS is the SNS message format which isn't Base64 encoded.
				ConsumeBase64: &enable64decoding,
			}
			sqsConfig.Region = config.AWSConfig.Region
			err = async.Retry(reconnectAttempts, reconnectDelay, func() error {
				sub, err = gizmoAWS.NewSubscriber(sqsConfig)
				if err != nil {
					logger.Errorf(context.TODO(), "Failed to initialize new gizmo aws subscriber with config [%+v] and err: %v", sqsConfig, err)
				}
				return err
			})
			if err != nil {
				panic(err)
			}
			processor = implementations.NewWebhookProcessor(common.AWS, sub, GetWebhook(cfg, scope), db, scope)
		case common.GCP:
			projectID := config.GCPConfig.ProjectID
			subscription := cfg.NotificationsProcessorConfig.QueueName
			err = async.Retry(reconnectAttempts, reconnectDelay, func() error {
				sub, err = gizmoGCP.NewSubscriber(context.TODO(), projectID, subscription)
				if err != nil {
					logger.Warnf(context.TODO(), "Failed to initialize new gizmo gcp subscriber with config [ProjectID: %s, Subscription: %s] and err: %v", projectID, subscription, err)
				}
				return err
			})
			if err != nil {
				panic(err)
			}
			processor = implementations.NewWebhookProcessor(common.GCP, sub, GetWebhook(cfg, scope), db, scope)
		case common.Local:
			fallthrough
		default:
			logger.Infof(context.Background(),
				"Using default noop notifications processor implementation for config type [%s]", config.Type)
			processor = notificationsImplementations.NewNoopProcess()
		}
		processors = append(processors, processor)
	}
	return processors
}
