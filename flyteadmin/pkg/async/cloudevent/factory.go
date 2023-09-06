package cloudevent

import (
	"context"
	"time"

	"github.com/NYTimes/gizmo/pubsub"
	gizmoAWS "github.com/NYTimes/gizmo/pubsub/aws"
	gizmoGCP "github.com/NYTimes/gizmo/pubsub/gcp"
	"github.com/Shopify/sarama"
	"github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/flyteorg/flyteadmin/pkg/async"
	cloudEventImplementations "github.com/flyteorg/flyteadmin/pkg/async/cloudevent/implementations"
	"github.com/flyteorg/flyteadmin/pkg/async/cloudevent/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/async/notifications/implementations"
	"github.com/flyteorg/flyteadmin/pkg/common"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
)

func NewCloudEventsPublisher(ctx context.Context, config runtimeInterfaces.CloudEventsConfig, scope promutils.Scope) interfaces.Publisher {
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
		return cloudEventImplementations.NewCloudEventsPublisher(&cloudEventImplementations.PubSubSender{Pub: publisher}, scope, config.EventsPublisherConfig.EventTypes)
	case common.GCP:
		pubsubConfig := gizmoGCP.Config{
			Topic: config.EventsPublisherConfig.TopicName,
		}
		pubsubConfig.ProjectID = config.GCPConfig.ProjectID
		var publisher pubsub.MultiPublisher
		var err error
		err = async.Retry(reconnectAttempts, reconnectDelay, func() error {
			publisher, err = gizmoGCP.NewPublisher(ctx, pubsubConfig)
			return err
		})

		if err != nil {
			panic(err)
		}
		return cloudEventImplementations.NewCloudEventsPublisher(&cloudEventImplementations.PubSubSender{Pub: publisher}, scope, config.EventsPublisherConfig.EventTypes)
	case cloudEventImplementations.Kafka:
		saramaConfig := sarama.NewConfig()
		var err error
		saramaConfig.Version, err = sarama.ParseKafkaVersion(config.KafkaConfig.Version)
		if err != nil {
			logger.Fatalf(ctx, "failed to parse kafka version, %v", err)
			panic(err)
		}
		sender, err := kafka_sarama.NewSender(config.KafkaConfig.Brokers, saramaConfig, config.EventsPublisherConfig.TopicName)
		if err != nil {
			panic(err)
		}
		client, err := cloudevents.NewClient(sender, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
		if err != nil {
			logger.Fatalf(ctx, "failed to create kafka client, %v", err)
			panic(err)
		}
		return cloudEventImplementations.NewCloudEventsPublisher(&cloudEventImplementations.KafkaSender{Client: client}, scope, config.EventsPublisherConfig.EventTypes)
	case common.Local:
		fallthrough
	default:
		logger.Infof(ctx,
			"Using default noop cloud events publisher implementation for config type [%s]", config.Type)
		return implementations.NewNoopPublish()
	}
}
