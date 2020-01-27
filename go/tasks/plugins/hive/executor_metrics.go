package hive

import (
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/promutils/labeled"
	"github.com/prometheus/client_golang/prometheus"
)

type QuboleHiveExecutorMetrics struct {
	Scope                 promutils.Scope
	ReleaseResourceFailed labeled.Counter
	AllocationGranted     labeled.Counter
	AllocationNotGranted  labeled.Counter
	TokenAge              prometheus.Summary
}

var (
	tokenAgeObjectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 1.0: 0.0}
)


func getQuboleHiveExecutorMetrics(scope promutils.Scope) QuboleHiveExecutorMetrics {
	return QuboleHiveExecutorMetrics{
		Scope: scope,
		ReleaseResourceFailed: labeled.NewCounter("released_resource_failed",
			"Error releasing allocation token", scope),
		AllocationGranted: labeled.NewCounter("allocation_granted",
			"Allocation request granted", scope),
		AllocationNotGranted: labeled.NewCounter("allocation_not_granted",
			"Allocation request did not fail but not granted", scope),
		TokenAge:
		// scope.MustNewSummary("token_age", "The age of the resource manager tokens"),
	}
}
