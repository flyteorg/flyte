package types

import (
	"strconv"
	"strings"
	"time"
)

type TaskPhase int

// NOTE: if we add a status here, we should make sure it converts correctly when reporting Task event
// See events_publisher.go
const (
	TaskPhaseQueued TaskPhase = iota
	TaskPhaseRunning
	TaskPhaseRetryableFailure
	TaskPhasePermanentFailure
	TaskPhaseSucceeded
	TaskPhaseUndefined
	TaskPhaseNotReady
	TaskPhaseUnknown
)

func (t TaskPhase) String() string {
	switch t {
	case TaskPhaseQueued:
		return "Queued"
	case TaskPhaseRunning:
		return "Running"
	case TaskPhaseRetryableFailure:
		return "RetryableFailure"
	case TaskPhasePermanentFailure:
		return "PermanentFailure"
	case TaskPhaseSucceeded:
		return "Succeeded"
	case TaskPhaseNotReady:
		return "NotReady"
	case TaskPhaseUndefined:
		return "Undefined"
	}
	return "Unknown"
}

func (t TaskPhase) IsTerminal() bool {
	return t.IsSuccess() || t.IsPermanentFailure()
}

func (t TaskPhase) IsSuccess() bool {
	return t == TaskPhaseSucceeded
}

func (t TaskPhase) IsPermanentFailure() bool {
	return t == TaskPhasePermanentFailure
}

func (t TaskPhase) IsRetryableFailure() bool {
	return t == TaskPhaseRetryableFailure
}

type TaskStatus struct {
	Phase        TaskPhase
	PhaseVersion uint32
	Err          error
	State        CustomState
	OccurredAt   time.Time
}

func (t TaskStatus) String() string {
	sb := strings.Builder{}
	sb.WriteString("{Phase: ")
	sb.WriteString(t.Phase.String())
	sb.WriteString(", Version: ")
	sb.WriteString(strconv.Itoa(int(t.PhaseVersion)))
	sb.WriteString(", At: ")
	sb.WriteString(t.OccurredAt.String())
	if t.Err != nil {
		sb.WriteString(", Err: ")
		sb.WriteString(t.Err.Error())
	}

	sb.WriteString(", CustomStateLen: ")
	sb.WriteString(strconv.Itoa(len(t.State)))

	sb.WriteString("}")

	return sb.String()
}

var TaskStatusNotReady = TaskStatus{Phase: TaskPhaseNotReady}
var TaskStatusQueued = TaskStatus{Phase: TaskPhaseQueued}
var TaskStatusRunning = TaskStatus{Phase: TaskPhaseRunning}
var TaskStatusSucceeded = TaskStatus{Phase: TaskPhaseSucceeded}
var TaskStatusUndefined = TaskStatus{Phase: TaskPhaseUndefined}
var TaskStatusUnknown = TaskStatus{Phase: TaskPhaseUnknown}

func (t TaskStatus) WithPhaseVersion(version uint32) TaskStatus {
	t.PhaseVersion = version
	return t
}

func (t TaskStatus) WithState(state CustomState) TaskStatus {
	t.State = state
	return t
}

func (t TaskStatus) WithOccurredAt(time time.Time) TaskStatus {
	t.OccurredAt = time
	return t
}

// This failure can be used to indicate that the task wasn't accepted due to resource quota or similar constraints.
func TaskStatusNotReadyFailure(err error) TaskStatus {
	return TaskStatus{Phase: TaskPhaseNotReady, Err: err}
}

// This failure can be used to indicate that the task failed with an error that is most probably transient
// and if the task retries (retry strategy) permits, it is safe to retry this task again.
// The same task execution will not be retried, but a new task execution will be created.
func TaskStatusRetryableFailure(err error) TaskStatus {
	return TaskStatus{Phase: TaskPhaseRetryableFailure, Err: err}
}

// PermanentFailure should be used to signal that either
//  1. The user wants to signal that the task has failed with something NON-RECOVERABLE
//  2. The plugin writer wants to signal that the task has failed with NON-RECOVERABLE
// Essentially a permanent failure will force the statemachine to shutdown and stop the task from being retried
// further, even if retries exist.
// If it is desirable to retry the task (a separate execution) then, use RetryableFailure
func TaskStatusPermanentFailure(err error) TaskStatus {
	return TaskStatus{Phase: TaskPhasePermanentFailure, Err: err}
}
