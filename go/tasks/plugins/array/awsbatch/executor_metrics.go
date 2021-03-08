package awsbatch

import (
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
)

type ExecutorMetrics struct {
	Scope              promutils.Scope
	SubTasksSubmitted  labeled.Counter
	SubTasksSucceeded  labeled.Counter
	SubTasksFailed     labeled.Counter
	SubTasksQueued     labeled.Counter
	BatchJobTerminated labeled.Counter
}

func getAwsBatchExecutorMetrics(scope promutils.Scope) ExecutorMetrics {
	return ExecutorMetrics{
		Scope: scope,
		SubTasksSubmitted: labeled.NewCounter("sub_task_submitted",
			"Sub tasks submitted", scope),
		SubTasksSucceeded: labeled.NewCounter("sub_task_success",
			"Batch sub tasks succeeded", scope),
		SubTasksQueued: labeled.NewCounter("sub_task_queued",
			"Batch sub tasks queued", scope),
		SubTasksFailed: labeled.NewCounter("sub_task_failure",
			"Batch sub tasks failed", scope),
		BatchJobTerminated: labeled.NewCounter("batch_job_terminated",
			"Batch job terminated", scope),
	}
}
