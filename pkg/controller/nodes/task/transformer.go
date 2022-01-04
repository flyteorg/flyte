package task

import (
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	pluginCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/common"
	"github.com/golang/protobuf/ptypes"

	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/handler"
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

func getParentNodeExecIDForTask(taskExecID *core.TaskExecutionIdentifier, execContext executors.ExecutionContext) (*core.NodeExecutionIdentifier, error) {
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
	return nodeExecutionID, nil
}

type ToTaskExecutionEventInputs struct {
	TaskExecContext       pluginCore.TaskExecutionContext
	InputReader           io.InputFilePaths
	OutputWriter          io.OutputFilePaths
	Info                  pluginCore.PhaseInfo
	NodeExecutionMetadata handler.NodeExecutionMetadata
	ExecContext           executors.ExecutionContext
	TaskType              string
	PluginID              string
	ResourcePoolInfo      []*event.ResourcePoolInfo
	ClusterID             string
}

func ToTaskExecutionEvent(input ToTaskExecutionEventInputs) (*event.TaskExecutionEvent, error) {
	// Transitions to a new phase

	tm := ptypes.TimestampNow()
	var err error
	if i := input.Info.Info(); i != nil && i.OccurredAt != nil {
		tm, err = ptypes.TimestampProto(*i.OccurredAt)
		if err != nil {
			return nil, err
		}
	}

	taskExecID := input.TaskExecContext.TaskExecutionMetadata().GetTaskExecutionID().GetID()
	nodeExecutionID, err := getParentNodeExecIDForTask(&taskExecID, input.ExecContext)
	if err != nil {
		return nil, err
	}
	metadata := input.Info.Info().Metadata
	if metadata == nil {
		metadata = &event.TaskExecutionMetadata{}
	}
	metadata.PluginIdentifier = input.PluginID
	metadata.GeneratedName = input.TaskExecContext.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()
	metadata.ResourcePoolInfo = input.ResourcePoolInfo
	tev := &event.TaskExecutionEvent{
		TaskId:                taskExecID.TaskId,
		ParentNodeExecutionId: nodeExecutionID,
		RetryAttempt:          taskExecID.RetryAttempt,
		Phase:                 ToTaskEventPhase(input.Info.Phase()),
		PhaseVersion:          input.Info.Version(),
		ProducerId:            input.ClusterID,
		OccurredAt:            tm,
		InputUri:              input.InputReader.GetInputPath().String(),
		TaskType:              input.TaskType,
		Reason:                input.Info.Reason(),
		Metadata:              metadata,
	}

	if input.Info.Phase().IsSuccess() && input.OutputWriter != nil {
		if input.OutputWriter.GetOutputPath() != "" {
			tev.OutputResult = &event.TaskExecutionEvent_OutputUri{OutputUri: input.OutputWriter.GetOutputPath().String()}
		}
	}

	if input.Info.Phase().IsFailure() && input.Info.Err() != nil {
		tev.OutputResult = &event.TaskExecutionEvent_Error{
			Error: input.Info.Err(),
		}
	}

	if input.Info.Info() != nil {
		tev.Logs = input.Info.Info().Logs
		tev.CustomInfo = input.Info.Info().CustomInfo
	}

	if input.NodeExecutionMetadata.IsInterruptible() {
		tev.Metadata.InstanceClass = event.TaskExecutionMetadata_INTERRUPTIBLE
	} else {
		tev.Metadata.InstanceClass = event.TaskExecutionMetadata_DEFAULT
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
