package task

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/event"
	pluginCore "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io"

	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
)

func ToTransitionType(ttype pluginCore.TransitionType) handler.TransitionType {
	if ttype == pluginCore.TransitionTypeBarrier {
		return handler.TransitionTypeBarrier
	}
	return handler.TransitionTypeEphemeral
}

func ToTaskEventPhase(p pluginCore.Phase) core.TaskExecution_Phase {
	switch p {
	case pluginCore.PhaseQueued:
		return core.TaskExecution_QUEUED
	case pluginCore.PhaseInitializing:
		// TODO add initializing phase
		return core.TaskExecution_QUEUED
	case pluginCore.PhaseWaitingForResources:
		// TODO add waiting for resources
		return core.TaskExecution_QUEUED
	case pluginCore.PhaseRunning:
		return core.TaskExecution_RUNNING
	case pluginCore.PhaseSuccess:
		return core.TaskExecution_SUCCEEDED
	case pluginCore.PhasePermanentFailure:
		return core.TaskExecution_FAILED
	case pluginCore.PhaseRetryableFailure:
		return core.TaskExecution_FAILED
	case pluginCore.PhaseNotReady:
		fallthrough
	case pluginCore.PhaseUndefined:
		fallthrough
	default:
		return core.TaskExecution_UNDEFINED
	}
}

func ToTaskExecutionEvent(taskExecID *core.TaskExecutionIdentifier, in io.InputFilePaths, out io.OutputFilePaths, info pluginCore.PhaseInfo) (*event.TaskExecutionEvent, error) {
	// Transitions to a new phase

	tm := ptypes.TimestampNow()
	var err error
	if i := info.Info(); i != nil && i.OccurredAt != nil {
		tm, err = ptypes.TimestampProto(*i.OccurredAt)
		if err != nil {
			return nil, err
		}
	}

	tev := &event.TaskExecutionEvent{
		TaskId:                taskExecID.TaskId,
		ParentNodeExecutionId: taskExecID.NodeExecutionId,
		RetryAttempt:          taskExecID.RetryAttempt,
		Phase:                 ToTaskEventPhase(info.Phase()),
		PhaseVersion:          info.Version(),
		ProducerId:            "propeller",
		OccurredAt:            tm,
		InputUri:              in.GetInputPath().String(),
	}

	if info.Phase().IsSuccess() && out != nil {
		if out.GetOutputPath() != "" {
			tev.OutputResult = &event.TaskExecutionEvent_OutputUri{OutputUri: out.GetOutputPath().String()}
		}
	}

	if info.Phase().IsFailure() && info.Err() != nil {
		tev.OutputResult = &event.TaskExecutionEvent_Error{
			Error: info.Err(),
		}
	}

	if info.Info() != nil {
		tev.Logs = info.Info().Logs
		tev.CustomInfo = info.Info().CustomInfo
	}

	return tev, nil
}

func GetTaskExecutionIdentifier(nCtx handler.NodeExecutionContext) *core.TaskExecutionIdentifier {
	return &core.TaskExecutionIdentifier{
		TaskId:       nCtx.TaskReader().GetTaskID(),
		RetryAttempt: nCtx.CurrentAttempt(),
		NodeExecutionId: &core.NodeExecutionIdentifier{
			NodeId:      nCtx.NodeID(),
			ExecutionId: nCtx.NodeExecutionMetadata().GetExecutionID().WorkflowExecutionIdentifier,
		},
	}
}
