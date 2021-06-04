package bigquery

import (
	"testing"
	"time"

	flyteIdlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/bigquery/v2"
	"google.golang.org/api/googleapi"
)

func TestFormatJobReference(t *testing.T) {
	t.Run("format job reference", func(t *testing.T) {
		jobReference := bigquery.JobReference{
			JobId:     "my-job-id",
			Location:  "EU",
			ProjectId: "flyte-test",
		}

		str := formatJobReference(jobReference)

		assert.Equal(t, "flyte-test:EU.my-job-id", str)
	})
}

func TestCreateTaskInfo(t *testing.T) {
	t.Run("create task info", func(t *testing.T) {
		resourceMeta := ResourceMetaWrapper{
			JobReference: bigquery.JobReference{
				JobId:     "my-job-id",
				Location:  "EU",
				ProjectId: "flyte-test",
			},
		}

		taskInfo := createTaskInfo(&resourceMeta)

		assert.Equal(t, 1, len(taskInfo.Logs))
		assert.Equal(t, flyteIdlCore.TaskLog{
			Uri:  "https://console.cloud.google.com/bigquery?project=flyte-test&j=bq:EU:my-job-id&page=queryresults",
			Name: "BigQuery Console",
		}, *taskInfo.Logs[0])
	})
}

func TestHandleCreateError(t *testing.T) {
	occurredAt := time.Now()
	taskInfo := core.TaskInfo{OccurredAt: &occurredAt}

	t.Run("handle 401", func(t *testing.T) {
		createError := googleapi.Error{
			Code:    401,
			Message: "user xxx is not authorized",
		}

		phase := handleCreateError(&createError, &taskInfo)

		assert.Equal(t, flyteIdlCore.ExecutionError{
			Code:    "http401",
			Message: "user xxx is not authorized",
			Kind:    flyteIdlCore.ExecutionError_USER,
		}, *phase.Err())
		assert.Equal(t, taskInfo, *phase.Info())
	})

	t.Run("handle 500", func(t *testing.T) {
		createError := googleapi.Error{
			Code:    500,
			Message: "oops",
		}

		phase := handleCreateError(&createError, &taskInfo)

		assert.Equal(t, flyteIdlCore.ExecutionError{
			Code:    "http500",
			Message: "oops",
			Kind:    flyteIdlCore.ExecutionError_SYSTEM,
		}, *phase.Err())
		assert.Equal(t, taskInfo, *phase.Info())
	})
}

func TestHandleErrorResult(t *testing.T) {
	occurredAt := time.Now()
	taskInfo := core.TaskInfo{OccurredAt: &occurredAt}

	type args struct {
		reason    string
		phase     core.Phase
		errorKind flyteIdlCore.ExecutionError_ErrorKind
	}

	tests := []args{
		{
			reason:    "accessDenied",
			phase:     pluginsCore.PhasePermanentFailure,
			errorKind: flyteIdlCore.ExecutionError_USER,
		},
		{
			reason:    "backendError",
			phase:     pluginsCore.PhaseRetryableFailure,
			errorKind: flyteIdlCore.ExecutionError_SYSTEM,
		},
		{
			reason:    "billingNotEnabled",
			phase:     pluginsCore.PhasePermanentFailure,
			errorKind: flyteIdlCore.ExecutionError_USER,
		},
		{
			reason:    "blocked",
			phase:     pluginsCore.PhasePermanentFailure,
			errorKind: flyteIdlCore.ExecutionError_USER,
		},
		{
			reason:    "duplicate",
			phase:     pluginsCore.PhasePermanentFailure,
			errorKind: flyteIdlCore.ExecutionError_USER,
		},
		{
			reason:    "internalError",
			phase:     pluginsCore.PhaseRetryableFailure,
			errorKind: flyteIdlCore.ExecutionError_SYSTEM,
		},
		{
			reason:    "invalid",
			phase:     pluginsCore.PhasePermanentFailure,
			errorKind: flyteIdlCore.ExecutionError_USER,
		},
		{
			reason:    "invalidQuery",
			phase:     pluginsCore.PhasePermanentFailure,
			errorKind: flyteIdlCore.ExecutionError_USER,
		},
		{
			reason:    "invalidUser",
			phase:     pluginsCore.PhaseRetryableFailure,
			errorKind: flyteIdlCore.ExecutionError_SYSTEM,
		},
		{
			reason:    "notFound",
			phase:     pluginsCore.PhasePermanentFailure,
			errorKind: flyteIdlCore.ExecutionError_USER,
		},
		{
			reason:    "notImplemented",
			phase:     pluginsCore.PhasePermanentFailure,
			errorKind: flyteIdlCore.ExecutionError_USER,
		},
		{
			reason:    "quotaExceeded",
			phase:     pluginsCore.PhaseRetryableFailure,
			errorKind: flyteIdlCore.ExecutionError_USER,
		},
		{
			reason:    "rateLimitExceeded",
			phase:     pluginsCore.PhaseRetryableFailure,
			errorKind: flyteIdlCore.ExecutionError_USER,
		},
		{
			reason:    "resourceInUse",
			phase:     pluginsCore.PhaseRetryableFailure,
			errorKind: flyteIdlCore.ExecutionError_SYSTEM,
		},

		{
			reason:    "resourcesExceeded",
			phase:     pluginsCore.PhasePermanentFailure,
			errorKind: flyteIdlCore.ExecutionError_USER,
		},

		{
			reason:    "responseTooLarge",
			phase:     pluginsCore.PhasePermanentFailure,
			errorKind: flyteIdlCore.ExecutionError_USER,
		},
		{
			reason:    "stopped",
			phase:     pluginsCore.PhasePermanentFailure,
			errorKind: flyteIdlCore.ExecutionError_USER,
		},
		{
			reason:    "tableUnavailable",
			phase:     pluginsCore.PhaseRetryableFailure,
			errorKind: flyteIdlCore.ExecutionError_SYSTEM,
		},
		{
			reason:    "timeout",
			phase:     pluginsCore.PhasePermanentFailure,
			errorKind: flyteIdlCore.ExecutionError_USER,
		},
	}

	for _, test := range tests {
		t.Run(test.reason, func(t *testing.T) {
			phaseInfo := handleErrorResult(test.reason, "message", &taskInfo)

			assert.Equal(t, test.phase, phaseInfo.Phase())
			assert.Equal(t, test.reason, phaseInfo.Err().Code)
			assert.Equal(t, test.errorKind, phaseInfo.Err().Kind)
			assert.Equal(t, "message", phaseInfo.Err().Message)
		})
	}
}
