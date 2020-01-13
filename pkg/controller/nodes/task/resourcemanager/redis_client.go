package resourcemanager

import (
	"context"

	"github.com/lyft/flytepropeller/pkg/controller/nodes/task/resourcemanager/config"

	"github.com/go-redis/redis"
	"github.com/lyft/flytestdlib/logger"
)

//go:generate mockery -name RedisClient -case=underscore

type RedisClient interface {
	// A pass-through method. Getting the cardinality of the Redis set
	SCard(string) *redis.IntCmd
	// A pass-through method. Checking if an entity is a member of the set specified by the key
	SIsMember(string, interface{}) *redis.BoolCmd
	// A pass-through method. Adding an entity to the set specified by the key
	SAdd(string, interface{}) *redis.IntCmd
	// A pass-through method. Removing an entity from the set specified by the key
	SRem(string, interface{}) *redis.IntCmd
	// A pass-through method. Pinging the Redis client
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
