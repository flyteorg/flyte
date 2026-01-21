package cache

import (
	"context"
	"net"

	"github.com/redis/go-redis/v9"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	flyteMetrics "github.com/flyteorg/flyte/v2/flytestdlib/metrics"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

// A hook that adds metrics around redis operations.
//
// Usage:
//
//	redisClient := redis.NewClient(...)
//	redisClient.AddHook(NewRedisInstrumentationHook(scope.NewSubScope("redis_client")))
type RedisInstrumentationHook struct {
	operation flyteMetrics.Operation
}

func (hook RedisInstrumentationHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (_ net.Conn, err error) {
		defer stopRedisTimer(ctx, hook.operation.Start("dial"), "dial", &err)
		return next(ctx, network, addr)
	}
}

func (hook RedisInstrumentationHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) (err error) {
		defer stopRedisTimer(ctx, hook.operation.Start(cmd.Name()), cmd.Name(), &err)
		return next(ctx, cmd)
	}
}

func (hook RedisInstrumentationHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) (err error) {
		defer stopRedisTimer(ctx, hook.operation.Start("pipeline"), "pipeline", &err)
		return next(ctx, cmds)
	}
}

func stopRedisTimer(ctx context.Context, timer flyteMetrics.OperationTimer, command string, err *error) {
	realError := *err
	if *err == redis.Nil {
		realError = nil
	}

	if realError != nil {
		logger.Errorf(ctx, "redis command %q failed: %v", command, realError)
	}

	timer.Stop(&realError)
}

func NewRedisInstrumentationHook(scope promutils.Scope) *RedisInstrumentationHook {
	return &RedisInstrumentationHook{
		operation: flyteMetrics.NewOperationHistogram(scope),
	}
}
