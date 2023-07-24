package hive

import (
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"github.com/prometheus/client_golang/prometheus"
)

type QuboleHiveExecutorMetrics struct {
	Scope                 promutils.Scope
	ResourceReleased      labeled.Counter
	ResourceReleaseFailed labeled.Counter
	AllocationGranted     labeled.Counter
	AllocationNotGranted  labeled.Counter
	ResourceWaitTime      prometheus.Summary
}

var (
	tokenAgeObjectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 1.0: 0.0}
)

func getQuboleHiveExecutorMetrics(scope promutils.Scope) QuboleHiveExecutorMetrics {
	return QuboleHiveExecutorMetrics{
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
	}
}
