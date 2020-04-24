package notifications

import (
	"context"

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
		process, err := gizmoAWS.NewSubscriber(sqsConfig)
		if err != nil {
			panic(err)
		}
		sub = process
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
		publisher, err := gizmoAWS.NewPublisher(snsConfig)
		// Any errors initiating Publisher with Amazon configurations results in a failed start up.
		if err != nil {
			panic(err)
		}
		return implementations.NewPublisher(publisher, scope)
	case common.GCP:
		pubsubConfig := gizmoGCP.Config{
			Topic: config.NotificationsPublisherConfig.TopicName,
		}
		pubsubConfig.ProjectID = config.GCPConfig.ProjectID
		publisher, err := gizmoGCP.NewPublisher(context.TODO(), pubsubConfig)
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
