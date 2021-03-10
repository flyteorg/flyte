package resourcemanager

import (
	"context"

	rmConfig "github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/resourcemanager/config"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
)

const (
	resourceManagerPrometheusScope      = "resourcemanager"
	redisResourceManagerPrometheusScope = "redis"
)

func GetResourceManagerBuilderByType(ctx context.Context, managerType rmConfig.Type, scope promutils.Scope) (
	Builder, error) {
	rmScope := scope.NewSubScope(resourceManagerPrometheusScope)

	switch managerType {
	case rmConfig.TypeNoop:
		logger.Infof(ctx, "Using the NOOP resource manager")
		return &NoopResourceManagerBuilder{}, nil
	case rmConfig.TypeRedis:
		logger.Infof(ctx, "Using Redis based resource manager")
		config := rmConfig.GetConfig()
		redisClient, err := NewRedisClient(ctx, config.RedisConfig)
		if err != nil {
			logger.Errorf(ctx, "Unable to initialize a redis client for the resource manager: [%v]", err)
			return nil, err
		}
		return NewRedisResourceManagerBuilder(ctx, redisClient, rmScope.NewSubScope(redisResourceManagerPrometheusScope))
	}
	logger.Infof(ctx, "Using the NOOP resource manager by default")
	return &NoopResourceManagerBuilder{}, nil
}
