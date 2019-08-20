package resourcemanager

import (
	"context"
	"github.com/go-redis/redis"
	"github.com/lyft/flytestdlib/logger"
)

func NewRedisClient(ctx context.Context, host string, key string, maxRetries int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:       host,
		Password:   key,
		DB:         0, // use default DB
		MaxRetries: maxRetries,
	})

	_, err := client.Ping().Result()
	if err != nil {
		logger.Errorf(ctx, "Error creating Redis client at %s %v", host, err)
		return nil, err
	}
	logger.Infof(ctx, "Created Redis client with host %s key %s ...", host, key)
	return client, nil
}
