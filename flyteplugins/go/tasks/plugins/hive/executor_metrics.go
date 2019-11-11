package hive

import (
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/promutils/labeled"
)

type QuboleHiveExecutorMetrics struct {
	Scope                 promutils.Scope
	ReleaseResourceFailed labeled.Counter
	AllocationGranted     labeled.Counter
	AllocationNotGranted  labeled.Counter
}

func getQuboleHiveExecutorMetrics(scope promutils.Scope) QuboleHiveExecutorMetrics {
	return QuboleHiveExecutorMetrics{
		Scope: scope,
		ReleaseResourceFailed: labeled.NewCounter("released_resource_failed",
			"Error releasing allocation token", scope),
		AllocationGranted: labeled.NewCounter("allocation_granted",
			"Allocation request granted", scope),
		AllocationNotGranted: labeled.NewCounter("allocation_not_granted",
			"Allocation request did not fail but not granted", scope),
	}
}
