package resourcemanager

import (
	"context"

	"github.com/lyft/flytepropeller/pkg/controller/nodes/task/resourcemanager/config"

	"github.com/go-redis/redis"
	"github.com/lyft/flytestdlib/logger"
)

type RedisClient interface {
	SCard(string) *redis.IntCmd
	SIsMember(string, interface{}) *redis.BoolCmd
	SAdd(string, interface{}) *redis.IntCmd
	SRem(string, interface{}) *redis.IntCmd
	Ping() *redis.StatusCmd
}

type Redis struct {
	c *redis.Client
}

func (r *Redis) SCard(key string) *redis.IntCmd {
	return r.c.SCard(key)
}

func (r *Redis) SIsMember(key string, member interface{}) *redis.BoolCmd {
	return r.c.SIsMember(key, member)
}

func (r *Redis) SAdd(key string, member interface{}) *redis.IntCmd {
	return r.c.SAdd(key, member)
}

func (r *Redis) SRem(key string, member interface{}) *redis.IntCmd {
	return r.c.SRem(key, member)
}

func (r *Redis) Ping() *redis.StatusCmd {
	return r.c.Ping()
}

func NewRedisClient(ctx context.Context, config config.RedisConfig) (RedisClient, error) {
	client := &Redis{
		c: redis.NewClient(&redis.Options{
			Addr:       config.HostPath,
			Password:   config.HostKey,
			DB:         0, // use default DB
			MaxRetries: config.MaxRetries,
		}),
	}

	_, err := client.Ping().Result()
	if err != nil {
		logger.Errorf(ctx, "Error creating Redis client at [%s]. Error: %v", config.HostPath, err)
		return nil, err
	}

	logger.Infof(ctx, "Created Redis client with host [%s]...", config.HostPath)
	return client, nil
}
