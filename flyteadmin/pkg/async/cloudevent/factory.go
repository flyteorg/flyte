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

	"github.com/flyteorg/flyte/flyteadmin/pkg/async"
	cloudEventImplementations "github.com/flyteorg/flyte/flyteadmin/pkg/async/cloudevent/implementations"
	"github.com/flyteorg/flyte/flyteadmin/pkg/async/cloudevent/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/async/notifications/implementations"
	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	dataInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/data/interfaces"
	repositoryInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/sandboxutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

func NewCloudEventsPublisher(ctx context.Context, db repositoryInterfaces.Repository, storageClient *storage.DataStore, urlData dataInterfaces.RemoteURLInterface, cloudEventsConfig runtimeInterfaces.CloudEventsConfig, remoteDataConfig runtimeInterfaces.RemoteDataConfig, scope promutils.Scope) interfaces.Publisher {
	if !cloudEventsConfig.Enable {
		logger.Infof(ctx, "CloudEvents are disabled, config is [+%v]", cloudEventsConfig)
		return implementations.NewNoopPublish()
	}
	reconnectAttempts := cloudEventsConfig.ReconnectAttempts
	reconnectDelay := time.Duration(cloudEventsConfig.ReconnectDelaySeconds) * time.Second

	var sender interfaces.Sender
	switch cloudEventsConfig.Type {
	case common.AWS:
		snsConfig := gizmoAWS.SNSConfig{
			Topic: cloudEventsConfig.EventsPublisherConfig.TopicName,
		}
		snsConfig.Region = cloudEventsConfig.AWSConfig.Region

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
		sender = &cloudEventImplementations.PubSubSender{Pub: publisher}

	case common.GCP:
		pubsubConfig := gizmoGCP.Config{
			Topic: cloudEventsConfig.EventsPublisherConfig.TopicName,
		}
		pubsubConfig.ProjectID = cloudEventsConfig.GCPConfig.ProjectID
		var publisher pubsub.MultiPublisher
		var err error
		err = async.Retry(reconnectAttempts, reconnectDelay, func() error {
			publisher, err = gizmoGCP.NewPublisher(ctx, pubsubConfig)
			return err
		})

		if err != nil {
			panic(err)
		}
		sender = &cloudEventImplementations.PubSubSender{Pub: publisher}

	case cloudEventImplementations.Kafka:
		saramaConfig := sarama.NewConfig()
		var err error
		saramaConfig.Version, err = sarama.ParseKafkaVersion(cloudEventsConfig.KafkaConfig.Version)
		if err != nil {
			logger.Fatalf(ctx, "failed to parse kafka version, %v", err)
			panic(err)
		}
		kafkaSender, err := kafka_sarama.NewSender(cloudEventsConfig.KafkaConfig.Brokers, saramaConfig, cloudEventsConfig.EventsPublisherConfig.TopicName)
		if err != nil {
			panic(err)
		}
		client, err := cloudevents.NewClient(kafkaSender, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
		if err != nil {
			logger.Fatalf(ctx, "failed to create kafka client, %v", err)
			panic(err)
		}
		sender = &cloudEventImplementations.KafkaSender{Client: client}

	case common.Sandbox:
		var publisher pubsub.Publisher
		publisher = sandboxutils.NewCloudEventsPublisher()
		sender = &cloudEventImplementations.PubSubSender{
			Pub: publisher,
		}

	case common.Local:
		fallthrough
	default:
		logger.Infof(ctx,
			"Using default noop cloud events publisher implementation for config type [%s]", cloudEventsConfig.Type)
		return implementations.NewNoopPublish()
	}

	if cloudEventsConfig.CloudEventVersion == runtimeInterfaces.CloudEventVersionv2 {
		return cloudEventImplementations.NewCloudEventsWrappedPublisher(db, sender, scope, storageClient, urlData, remoteDataConfig)
	}

	return cloudEventImplementations.NewCloudEventsPublisher(sender, scope, cloudEventsConfig.EventsPublisherConfig.EventTypes)

}
