package resourcemanager

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

type RedisLookasideBuffer struct {
	client      *redis.Client
	redisPrefix string
	expiry      time.Duration
}

func createKey(prefix, key string) string {
	return fmt.Sprintf("%s:%s", prefix, key)
}

func (r RedisLookasideBuffer) ConfirmExecution(ctx context.Context, executionKey string, executionValue string) error {
	err := r.client.Set(createKey(r.redisPrefix, executionKey), executionValue, r.expiry).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r RedisLookasideBuffer) RetrieveExecution(ctx context.Context, executionKey string) (string, error) {
	value, err := r.client.Get(createKey(r.redisPrefix, executionKey)).Result()
	if err == redis.Nil {
		return "", ExecutionNotFoundError
	} else if err != nil {
		return "", err
	}

	return value, nil
}

func NewRedisLookasideBuffer(ctx context.Context, client *redis.Client, redisPrefix string, expiry time.Duration) RedisLookasideBuffer {
	return RedisLookasideBuffer{
		client:      client,
		redisPrefix: redisPrefix,
		expiry:      expiry,
	}
}
