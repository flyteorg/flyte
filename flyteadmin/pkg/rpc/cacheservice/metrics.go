package cacheservice

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/flyteorg/flyteadmin/pkg/rpc/adminservice/util"
	"github.com/flyteorg/flytestdlib/promutils"
)

type cacheEndpointMetrics struct {
	scope promutils.Scope

	evictExecution     util.RequestMetrics
	evictTaskExecution util.RequestMetrics
}

type CacheMetrics struct {
	Scope        promutils.Scope
	PanicCounter prometheus.Counter

	cacheEndpointMetrics cacheEndpointMetrics
}

func InitMetrics(cacheScope promutils.Scope) CacheMetrics {
	return CacheMetrics{
		Scope: cacheScope,
		PanicCounter: cacheScope.MustNewCounter("handler_panic",
			"panics encountered while handling requests to the cache service"),

		cacheEndpointMetrics: cacheEndpointMetrics{
			scope:              cacheScope,
			evictExecution:     util.NewRequestMetrics(cacheScope, "evict_execution"),
			evictTaskExecution: util.NewRequestMetrics(cacheScope, "evict_task_execution"),
		},
	}
}
