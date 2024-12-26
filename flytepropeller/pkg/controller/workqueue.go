package controller

import (
	"context"

	"golang.org/x/time/rate"
	"k8s.io/client-go/util/workqueue"

	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	_ "github.com/flyteorg/flyte/flytestdlib/promutils" // Setup workqueue metrics
)

func NewWorkQueue(ctx context.Context, cfg config.WorkqueueConfig, name string) (workqueue.RateLimitingInterface, error) {
	// TODO introduce bounds checks
	logger.Infof(ctx, "WorkQueue type [%v] configured", cfg.Type)
	switch cfg.Type {
	case config.WorkqueueTypeBucketRateLimiter:
		logger.Infof(ctx, "Using Bucket Ratelimited Workqueue, Rate [%v] Capacity [%v]", cfg.Rate, cfg.Capacity)
		return workqueue.NewNamedRateLimitingQueue(
			NewDedupingBucketRateLimiter(NewLimiter(rate.Limit(cfg.Rate), cfg.Capacity)),
			name), nil
	case config.WorkqueueTypeExponentialFailureRateLimiter:
		logger.Infof(ctx, "Using Exponential failure backoff Ratelimited Workqueue, Base Delay [%v], max Delay [%v]", cfg.BaseDelay, cfg.MaxDelay)
		return workqueue.NewNamedRateLimitingQueue(
			workqueue.NewItemExponentialFailureRateLimiter(cfg.BaseDelay.Duration, cfg.MaxDelay.Duration),
			name), nil
	case config.WorkqueueTypeMaxOfRateLimiter:
		logger.Infof(ctx, "Using Max-of Ratelimited Workqueue, Bucket {Rate [%v] Capacity [%v]} | FailureBackoff {Base Delay [%v], max Delay [%v]}", cfg.Rate, cfg.Capacity, cfg.BaseDelay, cfg.MaxDelay)
		return workqueue.NewNamedRateLimitingQueue(
			workqueue.NewMaxOfRateLimiter(
				NewDedupingBucketRateLimiter(NewLimiter(rate.Limit(cfg.Rate), cfg.Capacity)),
				workqueue.NewItemExponentialFailureRateLimiter(cfg.BaseDelay.Duration,
					cfg.MaxDelay.Duration),
			), name), nil

	case config.WorkqueueTypeDefault:
		fallthrough
	default:
		logger.Infof(ctx, "Using Default Workqueue")
		return workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), name), nil
	}
}
