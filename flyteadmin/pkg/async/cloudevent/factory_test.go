package cloudevent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyte/flyteadmin/pkg/async/cloudevent/implementations"
	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	dataMocks "github.com/flyteorg/flyte/flyteadmin/pkg/data/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/mocks"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
	storageMocks "github.com/flyteorg/flyte/flytestdlib/storage/mocks"
)

func getMockStore() *storage.DataStore {
	pbStore := &storageMocks.ComposedProtobufStore{}
	pbStore.OnReadProtobufMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(_ mock.Arguments) {

	})

	mockStore := &storage.DataStore{
		ComposedProtobufStore: pbStore,
		ReferenceConstructor:  &storageMocks.ReferenceConstructor{},
	}
	return mockStore
}

var remoteCfg = runtimeInterfaces.RemoteDataConfig{
	Scheme: "",
}

func TestGetCloudEventPublisher(t *testing.T) {
	cfg := runtimeInterfaces.CloudEventsConfig{
		Enable:                true,
		EventsPublisherConfig: runtimeInterfaces.EventsPublisherConfig{TopicName: "topic"},
	}

	db := mocks.NewMockRepository()
	mockStore := getMockStore()
	url := dataMocks.NewMockRemoteURL()

	t.Run("local publisher", func(t *testing.T) {
		cfg.Type = common.Local
		assert.NotNil(t, NewCloudEventsPublisher(context.Background(), db, mockStore, url, cfg, remoteCfg, promutils.NewTestScope()))
	})

	t.Run("disable cloud event publisher", func(t *testing.T) {
		cfg.Enable = false
		assert.NotNil(t, NewCloudEventsPublisher(context.Background(), db, mockStore, url, cfg, remoteCfg, promutils.NewTestScope()))
	})
}

func TestInvalidAwsConfig(t *testing.T) {
	defer func() { r := recover(); assert.NotNil(t, r) }()
	cfg := runtimeInterfaces.CloudEventsConfig{
		Enable:                true,
		Type:                  common.AWS,
		EventsPublisherConfig: runtimeInterfaces.EventsPublisherConfig{TopicName: "topic"},
	}
	db := mocks.NewMockRepository()
	mockStore := getMockStore()
	url := dataMocks.NewMockRemoteURL()

	NewCloudEventsPublisher(context.Background(), db, mockStore, url, cfg, remoteCfg, promutils.NewTestScope())
	t.Errorf("did not panic")
}

func TestInvalidGcpConfig(t *testing.T) {
	defer func() { r := recover(); assert.NotNil(t, r) }()
	cfg := runtimeInterfaces.CloudEventsConfig{
		Enable:                true,
		Type:                  common.GCP,
		EventsPublisherConfig: runtimeInterfaces.EventsPublisherConfig{TopicName: "topic"},
	}
	db := mocks.NewMockRepository()
	mockStore := getMockStore()
	url := dataMocks.NewMockRemoteURL()

	NewCloudEventsPublisher(context.Background(), db, mockStore, url, cfg, remoteCfg, promutils.NewTestScope())
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
	db := mocks.NewMockRepository()
	mockStore := getMockStore()
	url := dataMocks.NewMockRemoteURL()

	NewCloudEventsPublisher(context.Background(), db, mockStore, url, cfg, remoteCfg, promutils.NewTestScope())
	cfg.KafkaConfig = runtimeInterfaces.KafkaConfig{Version: "2.1.0"}
	NewCloudEventsPublisher(context.Background(), db, mockStore, url, cfg, remoteCfg, promutils.NewTestScope())
	t.Errorf("did not panic")
}
