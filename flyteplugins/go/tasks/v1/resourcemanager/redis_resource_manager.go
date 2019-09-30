package resourcemanager

import (
	"context"
	"time"

	"github.com/go-redis/redis"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/lyft/flyteplugins/go/tasks/v1/qubole/config"
)

// This is the key that will point to the Redis Set.
// https://redis.io/commands#set
const RedisSetKey = "qubole"

type RedisResourceManager struct {
	client      *redis.Client
	redisSetKey string
	Metrics     RedisResourceManagerMetrics
}

type RedisResourceManagerMetrics struct {
	Scope                promutils.Scope
	RedisSizeCheckTime   promutils.StopWatch
	AllocatedTokensGauge prometheus.Gauge
}

func NewRedisResourceManagerMetrics(scope promutils.Scope) RedisResourceManagerMetrics {
	return RedisResourceManagerMetrics{
		Scope: scope,
		RedisSizeCheckTime: scope.MustNewStopWatch("redis:size_check_time_ms",
			"The time it takes to measure the size of the Redis Set where all the queries are stored", time.Millisecond),

		AllocatedTokensGauge: scope.MustNewGauge("size",
			"The number of allocation tokens currently in the Redis set"),
	}
}

func (r RedisResourceManager) AllocateResource(ctx context.Context, namespace string, allocationToken string) (
	AllocationStatus, error) {

	// Check to see if the allocation token is already in the set
	found, err := r.client.SIsMember(r.redisSetKey, allocationToken).Result()
	if err != nil {
		logger.Errorf(ctx, "Error getting size of Redis set %v", err)
		return AllocationUndefined, err
	}
	if found {
		logger.Infof(ctx, "Already allocated [%s:%s]", namespace, allocationToken)
		return AllocationStatusGranted, nil
	}

	size, err := r.client.SCard(r.redisSetKey).Result()
	if err != nil {
		logger.Errorf(ctx, "Error getting size of Redis set %v", err)
		return AllocationUndefined, err
	}

	if size >= int64(config.GetQuboleConfig().QuboleLimit) {
		logger.Infof(ctx, "Too many allocations (total [%d]), rejecting [%s:%s]", size, namespace, allocationToken)
		return AllocationStatusExhausted, nil
	}

	countAdded, err := r.client.SAdd(r.redisSetKey, allocationToken).Result()
	if err != nil {
		logger.Errorf(ctx, "Error adding token [%s:%s] %v", namespace, allocationToken, err)
		return AllocationUndefined, err
	}
	logger.Infof(ctx, "Added %d to the Redis Qubole set", countAdded)

	return AllocationStatusGranted, err
}

func (r RedisResourceManager) ReleaseResource(ctx context.Context, namespace string, allocationToken string) error {
	countRemoved, err := r.client.SRem(r.redisSetKey, allocationToken).Result()
	if err != nil {
		logger.Errorf(ctx, "Error removing token [%s:%s] %v", namespace, allocationToken, err)
		return err
	}
	logger.Infof(ctx, "Removed %d token: %s", countRemoved, allocationToken)

	return nil
}

func (r *RedisResourceManager) pollRedis(ctx context.Context) {
	stopWatch := r.Metrics.RedisSizeCheckTime.Start()
	defer stopWatch.Stop()
	size, err := r.client.SCard(r.redisSetKey).Result()
	if err != nil {
		logger.Errorf(ctx, "Error getting size of Redis set in metrics poller %v", err)
		return
	}
	r.Metrics.AllocatedTokensGauge.Set(float64(size))
}

func (r *RedisResourceManager) startMetricsGathering(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.pollRedis(ctx)
			}
		}
	}()
}

func NewRedisResourceManager(ctx context.Context, client *redis.Client, scope promutils.Scope) (*RedisResourceManager, error) {
	rm := &RedisResourceManager{
		client:      client,
		Metrics:     NewRedisResourceManagerMetrics(scope),
		redisSetKey: RedisSetKey,
	}
	rm.startMetricsGathering(ctx)

	return rm, nil
}
