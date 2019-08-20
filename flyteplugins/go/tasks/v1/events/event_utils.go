package events

import (
	"time"

	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/lyft/flyteplugins/go/tasks/v1/errors"
	"github.com/lyft/flyteplugins/go/tasks/v1/types"

	"github.com/golang/protobuf/ptypes"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/event"
)

// Additional info that should be sent to the front end. The Information is sent to the front-end if it meets certain
// criterion, for example currently, it is sent only if an event was not already sent for
type TaskEventInfo struct {
	// log information for the task execution
	Logs []*core.TaskLog
	// Set this value to the intended time when the status occurred at. If not provided, will be defaulted to the current
	// time at the time of publishing the event.
	OccurredAt *time.Time
	// Custom Event information that the plugin would like to expose to the front-end
	CustomInfo *structpb.Struct
}

// Convert all TaskStatus to an ExecutionPhase that is common to Flyte and Admin understands
// NOTE: if we add a TaskStatus entry, we should add it here too
func convertTaskPhaseToExecutionStatus(status types.TaskPhase) core.TaskExecution_Phase {
	switch status {
	case types.TaskPhaseRunning:
		return core.TaskExecution_RUNNING
	case types.TaskPhaseSucceeded:
		return core.TaskExecution_SUCCEEDED
	case types.TaskPhaseRetryableFailure, types.TaskPhasePermanentFailure:
		return core.TaskExecution_FAILED
	case types.TaskPhaseQueued:
		return core.TaskExecution_QUEUED
	default:
		return core.TaskExecution_UNDEFINED
	}
}

func CreateEvent(taskCtx types.TaskContext, taskStatus types.TaskStatus, info *TaskEventInfo) *event.TaskExecutionEvent {

	newTaskExecutionPhase := convertTaskPhaseToExecutionStatus(taskStatus.Phase)
	taskExecutionID := taskCtx.GetTaskExecutionID().GetID()

	occurredAt := ptypes.TimestampNow()
	logs := make([]*core.TaskLog, 0)
	var customInfo *structpb.Struct

	if info != nil {
		customInfo = info.CustomInfo
		if info.OccurredAt != nil {
			t, err := ptypes.TimestampProto(*info.OccurredAt)
			if err != nil {
				occurredAt = t
			}
		}

		logs = append(logs, info.Logs...)
	}

	taskEvent := &event.TaskExecutionEvent{
		Phase:                 newTaskExecutionPhase,
		PhaseVersion:          taskStatus.PhaseVersion,
		RetryAttempt:          taskCtx.GetTaskExecutionID().GetID().RetryAttempt,
		InputUri:              taskCtx.GetInputsFile().String(),
		OccurredAt:            occurredAt,
		Logs:                  logs,
		CustomInfo:            customInfo,
		TaskId:                taskExecutionID.TaskId,
		ParentNodeExecutionId: taskExecutionID.NodeExecutionId,
	}

	if newTaskExecutionPhase == core.TaskExecution_FAILED {
		errorCode := "UnknownTaskError"
		message := "unknown reason"
		if taskStatus.Err != nil {
			ec, ok := errors.GetErrorCode(taskStatus.Err)
			if ok {
				errorCode = ec
			}
			message = taskStatus.Err.Error()
		}
		taskEvent.OutputResult = &event.TaskExecutionEvent_Error{
			Error: &core.ExecutionError{
				Code:     errorCode,
				Message:  message,
				ErrorUri: taskCtx.GetErrorFile().String(),
			},
		}
	} else if newTaskExecutionPhase == core.TaskExecution_SUCCEEDED {
		taskEvent.OutputResult = &event.TaskExecutionEvent_OutputUri{
			OutputUri: taskCtx.GetOutputsFile().String(),
		}
	}
	return taskEvent
}
