package awsbatch

import (
	"context"
	"fmt"

	"github.com/lyft/flytestdlib/logger"

	arrayCore "github.com/lyft/flyteplugins/go/tasks/plugins/array/core"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/lyft/flyteplugins/go/tasks/plugins/array/arraystatus"
	"github.com/lyft/flyteplugins/go/tasks/plugins/array/awsbatch/config"
)

func LaunchSubTasks(ctx context.Context, tCtx core.TaskExecutionContext, batchClient Client, pluginConfig *config.Config,
	currentState *State) (nextState *State, err error) {

	if int64(currentState.GetExecutionArraySize()) > pluginConfig.MaxArrayJobSize {
		ee := fmt.Errorf("array size > max allowed. Requested [%v]. Allowed [%v]", currentState.GetExecutionArraySize(), pluginConfig.MaxArrayJobSize)
		logger.Info(ctx, ee)
		currentState.State = currentState.SetPhase(arrayCore.PhasePermanentFailure, 0).SetReason(ee.Error())
		return currentState, nil
	}

	jobDefinition := currentState.GetJobDefinitionArn()
	if len(jobDefinition) == 0 {
		return nil, fmt.Errorf("system error; no job definition created")
	}

	batchInput, err := FlyteTaskToBatchInput(ctx, tCtx, jobDefinition, pluginConfig)
	if err != nil {
		return nil, err
	}

	size := currentState.GetExecutionArraySize()
	t, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return nil, err
	}

	// If the original job was marked as an array (not a single job), then make sure to set it up correctly.
	if t.Type == arrayTaskType {
		logger.Debugf(ctx, "Task is of type [%v]. Will setup task index env vars.", t.Type)
		batchInput = UpdateBatchInputForArray(ctx, batchInput, int64(size))
	}

	j, err := batchClient.SubmitJob(ctx, batchInput)
	if err != nil {
		logger.Errorf(ctx, "Failed to submit job [%+v]. Error: %v", batchInput, err)
		return nil, err
	}

	parentState := currentState.
		SetPhase(arrayCore.PhaseCheckingSubTaskExecutions, 0).
		SetArrayStatus(arraystatus.ArrayStatus{
			Summary: arraystatus.ArraySummary{
				core.PhaseQueued: int64(size),
			},
			Detailed: arrayCore.NewPhasesCompactArray(uint(size)),
		}).
		SetReason("Successfully launched subtasks.")

	nextState = currentState.SetExternalJobID(j)
	nextState.State = parentState

	return nextState, nil
}

func TerminateSubTasks(ctx context.Context, batchClient Client, jobID string) error {
	return batchClient.TerminateJob(ctx, jobID, "aborted")
}
