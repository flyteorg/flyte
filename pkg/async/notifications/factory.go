package notifications

import (
	"context"
	"time"

	"github.com/lyft/flyteadmin/pkg/async"

	"github.com/lyft/flyteadmin/pkg/async/notifications/implementations"
	"github.com/lyft/flyteadmin/pkg/async/notifications/interfaces"
	runtimeInterfaces "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flytestdlib/logger"

	"github.com/NYTimes/gizmo/pubsub"
	gizmoAWS "github.com/NYTimes/gizmo/pubsub/aws"
	gizmoGCP "github.com/NYTimes/gizmo/pubsub/gcp"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"

	"github.com/lyft/flyteadmin/pkg/common"
	"github.com/lyft/flytestdlib/promutils"
)

const maxRetries = 3

var enable64decoding = false

type PublisherConfig struct {
	TopicName string
}

type ProcessorConfig struct {
	QueueName string
	AccountID string
}

type EmailerConfig struct {
	SenderEmail string
	BaseURL     string
}

func GetEmailer(config runtimeInterfaces.NotificationsConfig, scope promutils.Scope) interfaces.Emailer {
	switch config.Type {
	case common.AWS:
		awsConfig := aws.NewConfig().WithRegion(config.Region).WithMaxRetries(maxRetries)
		awsSession, err := session.NewSession(awsConfig)
		if err != nil {
			panic(err)
		}
		sesClient := ses.New(awsSession)
		return implementations.NewAwsEmailer(
			config,
			scope,
			sesClient,
		)
	case common.Local:
		fallthrough
	default:
		logger.Infof(context.Background(), "Using default noop emailer implementation for config type [%s]", config.Type)
		return implementations.NewNoopEmail()
	}
}

func NewNotificationsProcessor(config runtimeInterfaces.NotificationsConfig, scope promutils.Scope) interfaces.Processor {
	reconnectAttempts := config.ReconnectAttempts
	reconnectDelay := time.Duration(config.ReconnectDelaySeconds) * time.Second
	var sub pubsub.Subscriber
	var emailer interfaces.Emailer
	switch config.Type {
	case common.AWS:
		sqsConfig := gizmoAWS.SQSConfig{
			QueueName:           config.NotificationsProcessorConfig.QueueName,
			QueueOwnerAccountID: config.NotificationsProcessorConfig.AccountID,
			// The AWS configuration type uses SNS to SQS for notifications.
			// Gizmo by default will decode the SQS message using Base64 decoding.
			// However, the message body of SQS is the SNS message format which isn't Base64 encoded.
			ConsumeBase64: &enable64decoding,
		}
		sqsConfig.Region = config.Region
		var err error
		err = async.Retry(reconnectAttempts, reconnectDelay, func() error {
			sub, err = gizmoAWS.NewSubscriber(sqsConfig)
			return err
		})

		if err != nil {
			panic(err)
		}
		emailer = GetEmailer(config, scope)
	case common.Local:
		fallthrough
	default:
		logger.Infof(context.Background(),
			"Using default noop notifications processor implementation for config type [%s]", config.Type)
		return implementations.NewNoopProcess()
	}
	return implementations.NewProcessor(sub, emailer, scope)
}

func NewNotificationsPublisher(config runtimeInterfaces.NotificationsConfig, scope promutils.Scope) interfaces.Publisher {
	reconnectAttempts := config.ReconnectAttempts
	reconnectDelay := time.Duration(config.ReconnectDelaySeconds) * time.Second
	switch config.Type {
	case common.AWS:
		snsConfig := gizmoAWS.SNSConfig{
			Topic: config.NotificationsPublisherConfig.TopicName,
		}
		if config.AWSConfig.Region != "" {
			snsConfig.Region = config.AWSConfig.Region
		} else {
			snsConfig.Region = config.Region
		}

		var publisher pubsub.Publisher
		var err error
		err = async.Retry(reconnectAttempts, reconnectDelay, func() error {
			publisher, err = gizmoAWS.NewPublisher(snsConfig)
			return err
		})

		// Any persistent errors initiating Publisher with Amazon configurations results in a failed start up.
		if err != nil {
			panic(err)
		}
		return implementations.NewPublisher(publisher, scope)
	case common.GCP:
		pubsubConfig := gizmoGCP.Config{
			Topic: config.NotificationsPublisherConfig.TopicName,
		}
		pubsubConfig.ProjectID = config.GCPConfig.ProjectID
		var publisher pubsub.MultiPublisher
		var err error
		err = async.Retry(reconnectAttempts, reconnectDelay, func() error {
			publisher, err = gizmoGCP.NewPublisher(context.TODO(), pubsubConfig)
			return err
		})

		if err != nil {
			panic(err)
		}
		return implementations.NewPublisher(publisher, scope)
	case common.Local:
		fallthrough
	default:
		logger.Infof(context.Background(),
			"Using default noop notifications publisher implementation for config type [%s]", config.Type)
		return implementations.NewNoopPublish()
	}
}

func NewEventsPublisher(config runtimeInterfaces.ExternalEventsConfig, scope promutils.Scope) interfaces.Publisher {
	if !config.Enable {
		return implementations.NewNoopPublish()
	}
	reconnectAttempts := config.ReconnectAttempts
	reconnectDelay := time.Duration(config.ReconnectDelaySeconds) * time.Second
	switch config.Type {
	case common.AWS:
		snsConfig := gizmoAWS.SNSConfig{
			Topic: config.EventsPublisherConfig.TopicName,
		}
		snsConfig.Region = config.AWSConfig.Region

		var publisher pubsub.Publisher
		var err error
		err = async.Retry(reconnectAttempts, reconnectDelay, func() error {
			publisher, err = gizmoAWS.NewPublisher(snsConfig)
			return err
		})

		// Any persistent errors initiating Publisher with Amazon configurations results in a failed start up.
		if err != nil {
			panic(err)
		}
		return implementations.NewEventsPublisher(publisher, scope, config.EventsPublisherConfig.EventTypes)
	case common.GCP:
		pubsubConfig := gizmoGCP.Config{
			Topic: config.EventsPublisherConfig.TopicName,
		}
		pubsubConfig.ProjectID = config.GCPConfig.ProjectID
		var publisher pubsub.MultiPublisher
		var err error
		err = async.Retry(reconnectAttempts, reconnectDelay, func() error {
			publisher, err = gizmoGCP.NewPublisher(context.TODO(), pubsubConfig)
			return err
		})

		if err != nil {
			panic(err)
		}
		return implementations.NewEventsPublisher(publisher, scope, config.EventsPublisherConfig.EventTypes)
	case common.Local:
		fallthrough
	default:
		logger.Infof(context.Background(),
			"Using default noop events publisher implementation for config type [%s]", config.Type)
		return implementations.NewNoopPublish()
	}
}
