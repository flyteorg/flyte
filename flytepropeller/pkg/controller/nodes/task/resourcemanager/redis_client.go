package resourcemanager

import (
	"context"

	"github.com/lyft/flytepropeller/pkg/controller/nodes/task/resourcemanager/config"

	"github.com/go-redis/redis"
	"github.com/lyft/flytestdlib/logger"
)

func NewRedisClient(ctx context.Context, config config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:       config.HostPath,
		Password:   config.HostKey,
		DB:         0, // use default DB
		MaxRetries: config.MaxRetries,
	})

	_, err := client.Ping().Result()
	if err != nil {
		logger.Errorf(ctx, "Error creating Redis client at [%s]. Error: %v", config.HostPath, err)
		return nil, err
	}

	logger.Infof(ctx, "Created Redis client with host [%s]...", config.HostPath)
	return client, nil
}
