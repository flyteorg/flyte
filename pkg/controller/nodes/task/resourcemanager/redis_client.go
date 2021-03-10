package resourcemanager

import (
	"context"

	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/resourcemanager/config"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/go-redis/redis"
)

//go:generate mockery -name RedisClient -case=underscore

type RedisClient interface {
	// A pass-through method. Getting the cardinality of the Redis set
	SCard(string) (int64, error)
	// A pass-through method. Checking if an entity is a member of the set specified by the key
	SIsMember(string, interface{}) (bool, error)
	// A pass-through method. Adding an entity to the set specified by the key
	SAdd(string, interface{}) (int64, error)
	// A pass-through method. Removing an entity from the set specified by the key
	SRem(string, interface{}) (int64, error)
	// A pass-through method. Getting the complete list of MEMBERS of the set
	SMembers(string) ([]string, error)
	// A pass-through method. Pinging the Redis client
	Ping() (string, error)
}

type Redis struct {
	c redis.UniversalClient
}

func (r *Redis) SCard(key string) (int64, error) {
	return r.c.SCard(key).Result()
}

func (r *Redis) SIsMember(key string, member interface{}) (bool, error) {
	return r.c.SIsMember(key, member).Result()
}

func (r *Redis) SAdd(key string, member interface{}) (int64, error) {
	return r.c.SAdd(key, member).Result()
}

func (r *Redis) SRem(key string, member interface{}) (int64, error) {
	return r.c.SRem(key, member).Result()
}

func (r *Redis) SMembers(key string) ([]string, error) {
	return r.c.SMembers(key).Result()
}

func (r *Redis) Ping() (string, error) {
	return r.c.Ping().Result()
}

func NewRedisClient(ctx context.Context, config config.RedisConfig) (RedisClient, error) {
	// Backward compatibility
	if len(config.HostPaths) == 0 && len(config.HostPath) > 0 {
		config.HostPaths = []string{config.HostPath}
	}

	client := &Redis{
		c: redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:      config.HostPaths,
			MasterName: config.PrimaryName,
			Password:   config.HostKey,
			DB:         0, // use default DB
			MaxRetries: config.MaxRetries,
		}),
	}

	_, err := client.Ping()
	if err != nil {
		logger.Errorf(ctx, "Error creating Redis client at [%+v]. Error: %v", config.HostPaths, err)
		return nil, err
	}

	logger.Infof(ctx, "Created Redis client with host [%+v]...", config.HostPaths)
	return client, nil
}
