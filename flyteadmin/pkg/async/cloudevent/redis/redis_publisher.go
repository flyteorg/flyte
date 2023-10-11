package redis

import (
	"context"
	"fmt"
	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/golang/protobuf/proto"
	redisAPI "github.com/redis/go-redis/v9"
)

// Publisher satisfies the gizmo/pubsub.Publisher interface.
type Publisher struct {
	config    interfaces.RedisConfig
	client    *redisAPI.Client
	topicName string
}

func (p *Publisher) Publish(ctx context.Context, topic string, msg proto.Message) error {
	if len(topic) == 0 {
		topic = p.topicName
	}
	logger.Debugf(ctx, "Publishing message to topic [%s]", topic)

	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	res := p.client.Publish(ctx, topic, msgBytes)
	if res.Err() != nil {
		return fmt.Errorf("failed to publish message to topic [%s]: %w", topic, res.Err())
	}
	return nil
}

func (p *Publisher) PublishRaw(ctx context.Context, topic string, msgBytes []byte) error {
	if len(topic) == 0 {
		topic = p.topicName
	}
	logger.Debugf(ctx, "Publishing raw message to topic [%s]", topic)

	res := p.client.Publish(ctx, topic, msgBytes)
	if res.Err() != nil {
		return fmt.Errorf("failed to publish raw message to topic [%s]: %w", topic, res.Err())
	}
	return nil
}

func NewPublisher(config interfaces.RedisConfig) (*Publisher, error) {
	client := redisAPI.NewClient(&redisAPI.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	return &Publisher{
		config: config,
		client: client,
	}, nil
}
