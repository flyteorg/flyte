package presto

import (
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/promutils/labeled"
)

type ExecutorMetrics struct {
	Scope                 promutils.Scope
	ReleaseResourceFailed labeled.Counter
	AllocationGranted     labeled.Counter
	AllocationNotGranted  labeled.Counter
}

func getPrestoExecutorMetrics(scope promutils.Scope) ExecutorMetrics {
	return ExecutorMetrics{
		Scope: scope,
		ReleaseResourceFailed: labeled.NewCounter("presto_released_resource_failed",
			"Error releasing allocation token for Presto", scope),
		AllocationGranted: labeled.NewCounter("presto_allocation_granted",
			"Allocation request granted for Presto", scope),
		AllocationNotGranted: labeled.NewCounter("presto_allocation_not_granted",
			"Allocation request did not fail but not granted for Presto", scope),
	}
}
