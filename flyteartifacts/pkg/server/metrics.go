package server

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

type RequestMetrics struct {
	scope           promutils.Scope
	requestDuration promutils.StopWatch
	errCount        prometheus.Counter
	successCount    prometheus.Counter
}

func NewRequestMetrics(scope promutils.Scope, method string) RequestMetrics {
	methodScope := scope.NewSubScope(method)

	return RequestMetrics{
		scope: methodScope,
		requestDuration: methodScope.MustNewStopWatch("duration",
			"recorded response time duration for endpoint", time.Millisecond),
		errCount:     methodScope.MustNewCounter("errors", "count of errors returned by endpoint"),
		successCount: methodScope.MustNewCounter("success", "count of successful responses returned by endpoint"),
	}
}

type endpointMetrics struct {
	scope promutils.Scope

	create           RequestMetrics
	list             RequestMetrics
	registerConsumer RequestMetrics
	registerProducer RequestMetrics
	createTrigger    RequestMetrics
	deleteTrigger    RequestMetrics
	addTag           RequestMetrics
	search           RequestMetrics
}

type ServiceMetrics struct {
	Scope        promutils.Scope
	PanicCounter prometheus.Counter

	executionEndpointMetrics endpointMetrics
}

func InitMetrics(scope promutils.Scope) ServiceMetrics {
	return ServiceMetrics{
		Scope: scope,
		PanicCounter: scope.MustNewCounter("handler_panic",
			"panics encountered while handling requests to the artifact service"),

		executionEndpointMetrics: endpointMetrics{
			scope:            scope,
			create:           NewRequestMetrics(scope, "create_artifact"),
			list:             NewRequestMetrics(scope, "list_artifacts"),
			registerConsumer: NewRequestMetrics(scope, "register_consumer"),
			registerProducer: NewRequestMetrics(scope, "register_producer"),
			createTrigger:    NewRequestMetrics(scope, "create_trigger"),
			deleteTrigger:    NewRequestMetrics(scope, "delete_trigger"),
			addTag:           NewRequestMetrics(scope, "add_tag"),
			search:           NewRequestMetrics(scope, "search"),
		},
	}
}
