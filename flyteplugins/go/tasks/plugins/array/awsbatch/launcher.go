package awsbatch

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyteplugins/go/tasks/errors"

	"github.com/flyteorg/flytestdlib/bitarray"
	"github.com/flyteorg/flytestdlib/logger"

	arrayCore "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/core"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/arraystatus"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/awsbatch/config"
)

func LaunchSubTasks(ctx context.Context, tCtx core.TaskExecutionContext, batchClient Client, pluginConfig *config.Config,
	currentState *State, metrics ExecutorMetrics, terminalVersion uint32) (nextState *State, err error) {
	size := currentState.GetExecutionArraySize()

	jobDefinition := currentState.GetJobDefinitionArn()
	if len(jobDefinition) == 0 {
		return nil, fmt.Errorf("system error; no job definition created")
	}

	batchInput, err := FlyteTaskToBatchInput(ctx, tCtx, jobDefinition, pluginConfig)
	if err != nil {
		return nil, err
	}

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

	metrics.SubTasksSubmitted.Add(ctx, float64(size))

	retryAttemptsArray, err := bitarray.NewCompactArray(uint(size), bitarray.Item(pluginConfig.MaxRetries))
	if err != nil {
		logger.Errorf(context.Background(), "Failed to create attempts compact array with [count: %v, maxValue: %v]", size, pluginConfig.MaxRetries)
		return nil, err
	}

	parentState := currentState.
		SetPhase(arrayCore.PhaseCheckingSubTaskExecutions, terminalVersion).
		SetArrayStatus(arraystatus.ArrayStatus{
			Summary: arraystatus.ArraySummary{
				core.PhaseQueued: int64(size),
			},
			Detailed: arrayCore.NewPhasesCompactArray(uint(size)),
		}).
		SetReason("Successfully launched subtasks.").
		SetRetryAttempts(retryAttemptsArray)

	nextState = currentState.SetExternalJobID(j)
	nextState.State = parentState

	return nextState, nil
}

// Attempts to terminate the AWS Job if one is recorded in the pluginState. This API is idempotent and should be safe
// to call multiple times on the same job. It'll result in multiple calls to AWS Batch in that case, however.
func TerminateSubTasks(ctx context.Context, tCtx core.TaskExecutionContext, batchClient Client, reason string, metrics ExecutorMetrics) error {
	pluginState := &State{}
	if _, err := tCtx.PluginStateReader().Get(pluginState); err != nil {
		return errors.Wrapf(errors.CorruptedPluginState, err, "Failed to unmarshal custom state")
	}

	// This only makes sense if the task has "just" been kicked off. Assigning state here is meant to make subsequent
	// code simpler.
	if pluginState.State == nil {
		pluginState.State = &arrayCore.State{}
	}

	p, _ := pluginState.GetPhase()
	logger.Infof(ctx, "TerminateSubTasks is called with phase [%v] and reason [%v]", p, reason)

	if pluginState.GetExternalJobID() != nil {
		jobID := *pluginState.GetExternalJobID()
		logger.Infof(ctx, "Cancelling AWS Job [%v] because [%v].", jobID, reason)
		err := batchClient.TerminateJob(ctx, jobID, reason)
		if err != nil {
			return err
		}
		metrics.BatchJobTerminated.Inc(ctx)
	}

	return nil
}
