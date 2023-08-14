package notifications

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/async"

	"github.com/flyteorg/flyteadmin/pkg/async/notifications/implementations"
	"github.com/flyteorg/flyteadmin/pkg/async/notifications/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/logger"

	"github.com/NYTimes/gizmo/pubsub"
	gizmoAWS "github.com/NYTimes/gizmo/pubsub/aws"
	gizmoGCP "github.com/NYTimes/gizmo/pubsub/gcp"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flytestdlib/promutils"
)

const maxRetries = 3

var enable64decoding = false

var msgChan chan []byte
var once sync.Once

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

// For sandbox only
func CreateMsgChan() {
	once.Do(func() {
		msgChan = make(chan []byte)
	})
}

func GetEmailer(config runtimeInterfaces.NotificationsConfig, scope promutils.Scope) interfaces.Emailer {
	// If an external email service is specified use that instead.
	// TODO: Handling of this is messy, see https://github.com/flyteorg/flyte/issues/1063
	if config.NotificationsEmailerConfig.EmailerConfig.ServiceName != "" {
		switch config.NotificationsEmailerConfig.EmailerConfig.ServiceName {
		case implementations.Sendgrid:
			return implementations.NewSendGridEmailer(config, scope)
		default:
			panic(fmt.Errorf("No matching email implementation for %s", config.NotificationsEmailerConfig.EmailerConfig.ServiceName))
		}
	}

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
			if err != nil {
				logger.Warnf(context.TODO(), "Failed to initialize new gizmo aws subscriber with config [%+v] and err: %v", sqsConfig, err)
			}
			return err
		})

		if err != nil {
			panic(err)
		}
		emailer = GetEmailer(config, scope)
		return implementations.NewProcessor(sub, emailer, scope)
	case common.GCP:
		projectID := config.GCPConfig.ProjectID
		subscription := config.NotificationsProcessorConfig.QueueName
		var err error
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
		emailer = GetEmailer(config, scope)
		return implementations.NewGcpProcessor(sub, emailer, scope)
	case common.Sandbox:
		emailer = GetEmailer(config, scope)
		return implementations.NewSandboxProcessor(msgChan, emailer)
	case common.Local:
		fallthrough
	default:
		logger.Infof(context.Background(),
			"Using default noop notifications processor implementation for config type [%s]", config.Type)
		return implementations.NewNoopProcess()
	}
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
	case common.Sandbox:
		CreateMsgChan()
		return implementations.NewSandboxPublisher(msgChan)
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
