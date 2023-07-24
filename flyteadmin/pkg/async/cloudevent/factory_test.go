package cloudevent

import (
	"context"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/async/cloudevent/implementations"
	"github.com/flyteorg/flyteadmin/pkg/common"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

func TestGetCloudEventPublisher(t *testing.T) {
	cfg := runtimeInterfaces.CloudEventsConfig{
		Enable:                true,
		EventsPublisherConfig: runtimeInterfaces.EventsPublisherConfig{TopicName: "topic"},
	}

	t.Run("local publisher", func(t *testing.T) {
		cfg.Type = common.Local
		assert.NotNil(t, NewCloudEventsPublisher(context.Background(), cfg, promutils.NewTestScope()))
	})

	t.Run("disable cloud event publisher", func(t *testing.T) {
		cfg.Enable = false
		assert.NotNil(t, NewCloudEventsPublisher(context.Background(), cfg, promutils.NewTestScope()))
	})
}

func TestInvalidAwsConfig(t *testing.T) {
	defer func() { r := recover(); assert.NotNil(t, r) }()
	cfg := runtimeInterfaces.CloudEventsConfig{
		Enable:                true,
		Type:                  common.AWS,
		EventsPublisherConfig: runtimeInterfaces.EventsPublisherConfig{TopicName: "topic"},
	}
	NewCloudEventsPublisher(context.Background(), cfg, promutils.NewTestScope())
	t.Errorf("did not panic")
}

func TestInvalidGcpConfig(t *testing.T) {
	defer func() { r := recover(); assert.NotNil(t, r) }()
	cfg := runtimeInterfaces.CloudEventsConfig{
		Enable:                true,
		Type:                  common.GCP,
		EventsPublisherConfig: runtimeInterfaces.EventsPublisherConfig{TopicName: "topic"},
	}
	NewCloudEventsPublisher(context.Background(), cfg, promutils.NewTestScope())
	t.Errorf("did not panic")
}

func TestInvalidKafkaConfig(t *testing.T) {
	defer func() { r := recover(); assert.NotNil(t, r) }()
	cfg := runtimeInterfaces.CloudEventsConfig{
		Enable:                true,
		Type:                  implementations.Kafka,
		EventsPublisherConfig: runtimeInterfaces.EventsPublisherConfig{TopicName: "topic"},
		KafkaConfig:           runtimeInterfaces.KafkaConfig{Version: "0.8.2.0"},
	}
	NewCloudEventsPublisher(context.Background(), cfg, promutils.NewTestScope())
	cfg.KafkaConfig = runtimeInterfaces.KafkaConfig{Version: "2.1.0"}
	NewCloudEventsPublisher(context.Background(), cfg, promutils.NewTestScope())
	t.Errorf("did not panic")
}
