package awsbatch

import (
	"context"

	core2 "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/storage"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array"

	"github.com/flyteorg/flytestdlib/logger"

	arrayCore "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/core"

	"github.com/flyteorg/flytestdlib/bitarray"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/arraystatus"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/awsbatch/config"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/errorcollector"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
)

func createSubJobList(count int) []*Job {
	res := make([]*Job, count)
	for i := range res {
		res[i] = &Job{
			Status: JobStatus{Phase: core.PhaseNotReady},
		}
	}

	return res
}

func CheckSubTasksState(ctx context.Context, taskMeta core.TaskExecutionMetadata, outputPrefix, baseOutputSandbox storage.DataReference, jobStore *JobStore,
	dataStore *storage.DataStore, cfg *config.Config, currentState *State, metrics ExecutorMetrics) (newState *State, err error) {
	newState = currentState
	parentState := currentState.State
	jobName := taskMeta.GetTaskExecutionID().GetGeneratedName()
	job := jobStore.Get(jobName)
	// If job isn't currently being monitored (recovering from a restart?), add it to the sync-cache and return
	if job == nil {
		logger.Info(ctx, "Job not found in cache, adding it. [%v]", jobName)

		_, err = jobStore.GetOrCreate(jobName, &Job{
			ID:             *currentState.ExternalJobID,
			OwnerReference: taskMeta.GetOwnerID(),
			SubJobs:        createSubJobList(currentState.GetExecutionArraySize()),
		})

		if err != nil {
			return nil, err
		}

		return currentState, nil
	}

	msg := errorcollector.NewErrorMessageCollector()
	newArrayStatus := arraystatus.ArrayStatus{
		Summary:  arraystatus.ArraySummary{},
		Detailed: arrayCore.NewPhasesCompactArray(uint(currentState.GetExecutionArraySize())),
	}

	queued := 0
	for childIdx, subJob := range job.SubJobs {
		actualPhase := subJob.Status.Phase
		originalIdx := arrayCore.CalculateOriginalIndex(childIdx, currentState.GetIndexesToCache())
		if subJob.Status.Phase == core.PhaseQueued {
			queued++
		}
		if subJob.Status.Phase.IsFailure() {
			if len(subJob.Status.Message) > 0 {
				// If the service reported an error but there is no error.pb written, write one with the
				// service-provided error message.
				msg.Collect(childIdx, subJob.Status.Message)
				or, err := array.ConstructOutputReader(ctx, dataStore, outputPrefix, baseOutputSandbox, originalIdx)
				if err != nil {
					return nil, err
				}

				if hasErr, err := or.IsError(ctx); err != nil {
					return nil, err
				} else if !hasErr {
					// The subtask has not produced an error.pb, write one.
					ow, err := array.ConstructOutputWriter(ctx, dataStore, outputPrefix, baseOutputSandbox, originalIdx)
					if err != nil {
						return nil, err
					}

					if err = ow.Put(ctx, ioutils.NewInMemoryOutputReader(nil, &io.ExecutionError{
						ExecutionError: &core2.ExecutionError{
							Code:     "",
							Message:  subJob.Status.Message,
							ErrorUri: "",
						},
						IsRecoverable: false,
					})); err != nil {
						return nil, err
					}
				}
			} else {
				msg.Collect(childIdx, "Job failed")
			}
		} else if subJob.Status.Phase.IsSuccess() {
			actualPhase, err = array.CheckTaskOutput(ctx, dataStore, outputPrefix, baseOutputSandbox, childIdx, originalIdx)
			if err != nil {
				return nil, err
			}
		}

		newArrayStatus.Detailed.SetItem(childIdx, bitarray.Item(actualPhase))
		newArrayStatus.Summary.Inc(actualPhase)
		parentState.RetryAttempts.SetItem(childIdx, bitarray.Item(len(subJob.Attempts)))
	}

	if queued > 0 {
		metrics.SubTasksQueued.Add(ctx, float64(queued))
	}

	parentState = parentState.SetArrayStatus(newArrayStatus)
	// Based on the summary produced above, deduce the overall phase of the task.
	phase := arrayCore.SummaryToPhase(ctx, currentState.GetOriginalMinSuccesses()-currentState.GetOriginalArraySize()+int64(currentState.GetExecutionArraySize()), newArrayStatus.Summary)

	if phase != arrayCore.PhaseCheckingSubTaskExecutions {
		metrics.SubTasksSucceeded.Add(ctx, float64(newArrayStatus.Summary[core.PhaseSuccess]))
		totalFailed := newArrayStatus.Summary[core.PhasePermanentFailure] + newArrayStatus.Summary[core.PhaseRetryableFailure]
		metrics.SubTasksFailed.Add(ctx, float64(totalFailed))
	}
	if phase == arrayCore.PhaseWriteToDiscoveryThenFail {
		errorMsg := msg.Summary(cfg.MaxErrorStringLength)
		parentState = parentState.SetReason(errorMsg)
	}
	if phase == arrayCore.PhaseCheckingSubTaskExecutions {
		newPhaseVersion := uint32(0)
		// For now, the only changes to PhaseVersion and PreviousSummary occur for running array jobs.
		for phase, count := range parentState.GetArrayStatus().Summary {
			newPhaseVersion += uint32(phase) * uint32(count)
		}

		parentState = parentState.SetPhase(phase, newPhaseVersion).SetReason("Task is still running.")
	} else {
		parentState = parentState.SetPhase(phase, core.DefaultPhaseVersion)
	}

	p, v := parentState.GetPhase()
	logger.Debugf(ctx, "Current phase [phase: %v, version: %v]. Summary: %+v", p, v, newArrayStatus.Summary)
	newState.State = parentState

	return newState, nil
}
