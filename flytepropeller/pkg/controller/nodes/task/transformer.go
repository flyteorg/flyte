package task

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/event"
	pluginCore "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/controller/executors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/common"

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
		return core.TaskExecution_INITIALIZING
	case pluginCore.PhaseWaitingForResources:
		return core.TaskExecution_WAITING_FOR_RESOURCES
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

func trimErrorMessage(original string, maxLength int) string {
	if len(original) <= maxLength {
		return original
	}

	return original[0:maxLength/2] + original[len(original)-maxLength/2:]
}

func ToTaskExecutionEvent(taskExecID *core.TaskExecutionIdentifier, in io.InputFilePaths, out io.OutputFilePaths, info pluginCore.PhaseInfo,
	nodeExecutionMetadata handler.NodeExecutionMetadata, execContext executors.ExecutionContext) (*event.TaskExecutionEvent, error) {
	// Transitions to a new phase

	tm := ptypes.TimestampNow()
	var err error
	if i := info.Info(); i != nil && i.OccurredAt != nil {
		tm, err = ptypes.TimestampProto(*i.OccurredAt)
		if err != nil {
			return nil, err
		}
	}

	nodeExecutionID := &core.NodeExecutionIdentifier{
		ExecutionId: taskExecID.NodeExecutionId.ExecutionId,
	}
	if execContext.GetEventVersion() != v1alpha1.EventVersion0 {
		currentNodeUniqueID, err := common.GenerateUniqueID(execContext.GetParentInfo(), taskExecID.NodeExecutionId.NodeId)
		if err != nil {
			return nil, err
		}
		nodeExecutionID.NodeId = currentNodeUniqueID
	} else {
		nodeExecutionID.NodeId = taskExecID.NodeExecutionId.NodeId
	}
	tev := &event.TaskExecutionEvent{
		TaskId:                taskExecID.TaskId,
		ParentNodeExecutionId: nodeExecutionID,
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

	if nodeExecutionMetadata.IsInterruptible() {
		tev.Metadata = &event.TaskExecutionMetadata{InstanceClass: event.TaskExecutionMetadata_INTERRUPTIBLE}
	} else {
		tev.Metadata = &event.TaskExecutionMetadata{InstanceClass: event.TaskExecutionMetadata_DEFAULT}
	}

	return tev, nil
}

func GetTaskExecutionIdentifier(nCtx handler.NodeExecutionContext) *core.TaskExecutionIdentifier {
	return &core.TaskExecutionIdentifier{
		TaskId:          nCtx.TaskReader().GetTaskID(),
		RetryAttempt:    nCtx.CurrentAttempt(),
		NodeExecutionId: nCtx.NodeExecutionMetadata().GetNodeExecutionID(),
	}
}
