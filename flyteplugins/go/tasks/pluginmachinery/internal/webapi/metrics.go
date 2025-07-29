package webapi

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
)

// Global metrics cache to avoid recreating metrics for the same scope
var (
	metricsCache = make(map[string]*Metrics)
	metricsMutex sync.RWMutex
)

type Metrics struct {
	Scope                   promutils.Scope
	ResourceReleased        labeled.Counter
	ResourceReleaseFailed   labeled.Counter
	AllocationGranted       labeled.Counter
	AllocationNotGranted    labeled.Counter
	ResourceWaitTime        prometheus.Summary
	SucceededUnmarshalState labeled.StopWatch
	FailedUnmarshalState    labeled.Counter
}

var (
	tokenAgeObjectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 1.0: 0.0}
)

func newMetrics(scope promutils.Scope) Metrics {
	scopeName := scope.CurrentScope()
	
	// Check if we already have metrics for this scope
	metricsMutex.RLock()
	if cachedMetrics, exists := metricsCache[scopeName]; exists {
		defer metricsMutex.RUnlock()
		return *cachedMetrics
	}
	metricsMutex.RUnlock()
	
	// Create new metrics and store globally
	metricsMutex.Lock()
	defer metricsMutex.Unlock()
	
	// Double-check in case another goroutine created metrics while we were acquiring the lock
	if cachedMetrics, exists := metricsCache[scopeName]; exists {
		return *cachedMetrics
	}
	
	newMetrics := Metrics{
		Scope: scope,
		ResourceReleased: labeled.NewCounter("resource_release_success",
			"Resource allocation token released", scope, labeled.EmitUnlabeledMetric),
		ResourceReleaseFailed: labeled.NewCounter("resource_release_failed",
			"Error releasing allocation token", scope, labeled.EmitUnlabeledMetric),
		AllocationGranted: labeled.NewCounter("allocation_grant_success",
			"Allocation request granted", scope, labeled.EmitUnlabeledMetric),
		AllocationNotGranted: labeled.NewCounter("allocation_grant_failed",
			"Allocation request did not fail but not granted", scope, labeled.EmitUnlabeledMetric),
		ResourceWaitTime: scope.MustNewSummaryWithOptions("resource_wait_time", "Duration the execution has been waiting for a resource allocation token",
			promutils.SummaryOptions{Objectives: tokenAgeObjectives}),
		SucceededUnmarshalState: labeled.NewStopWatch("unmarshal_state_success", "Successfully unmarshaled state",
			time.Millisecond, scope),
		FailedUnmarshalState: labeled.NewCounter("unmarshal_state_failed",
			"Failed to unmarshal state", scope, labeled.EmitUnlabeledMetric),
	}
	
	metricsCache[scopeName] = &newMetrics
	return newMetrics
}
