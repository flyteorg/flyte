package k8s

import (
	"context"
	"fmt"
	"time"

	idlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyteplugins/go/tasks/logs"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/arraystatus"
	arrayCore "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/core"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/errorcollector"

	"github.com/flyteorg/flytestdlib/bitarray"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/storage"
)

// allocateResource attempts to allot resources for the specified parameter with the
// TaskExecutionContexts ResourceManager.
func allocateResource(ctx context.Context, tCtx core.TaskExecutionContext, config *Config, podName string) (core.AllocationStatus, error) {
	if !IsResourceConfigSet(config.ResourceConfig) {
		return core.AllocationStatusGranted, nil
	}

	resourceNamespace := core.ResourceNamespace(config.ResourceConfig.PrimaryLabel)
	resourceConstraintSpec := core.ResourceConstraintsSpec{
		ProjectScopeResourceConstraint:   nil,
		NamespaceScopeResourceConstraint: nil,
	}

	allocationStatus, err := tCtx.ResourceManager().AllocateResource(ctx, resourceNamespace, podName, resourceConstraintSpec)
	if err != nil {
		return core.AllocationUndefined, err
	}

	return allocationStatus, nil
}

// deallocateResource attempts to release resources for the specified parameter with the
// TaskExecutionContexts ResourceManager.
func deallocateResource(ctx context.Context, tCtx core.TaskExecutionContext, config *Config, podName string) error {
	if !IsResourceConfigSet(config.ResourceConfig) {
		return nil
	}
	resourceNamespace := core.ResourceNamespace(config.ResourceConfig.PrimaryLabel)

	err := tCtx.ResourceManager().ReleaseResource(ctx, resourceNamespace, podName)
	if err != nil {
		logger.Errorf(ctx, "Error releasing token [%s]. error %s", podName, err)
		return err
	}

	return nil
}

// LaunchAndCheckSubTasksState iterates over each subtask performing operations to transition them
// to a terminal state. This may include creating new k8s resources, monitoring existing k8s
// resources, retrying failed attempts, or declaring a permanent failure among others.
func LaunchAndCheckSubTasksState(ctx context.Context, tCtx core.TaskExecutionContext, kubeClient core.KubeClient,
	config *Config, dataStore *storage.DataStore, outputPrefix, baseOutputDataSandbox storage.DataReference, currentState *arrayCore.State) (
	newState *arrayCore.State, externalResources []*core.ExternalResource, err error) {

	newState = currentState
	messageCollector := errorcollector.NewErrorMessageCollector()
	newArrayStatus := &arraystatus.ArrayStatus{
		Summary:  arraystatus.ArraySummary{},
		Detailed: arrayCore.NewPhasesCompactArray(uint(currentState.GetExecutionArraySize())),
	}
	externalResources = make([]*core.ExternalResource, 0, len(currentState.GetArrayStatus().Detailed.GetItems()))

	// If we have arrived at this state for the first time then currentState has not been
	// initialized with number of sub tasks.
	if len(currentState.GetArrayStatus().Detailed.GetItems()) == 0 {
		currentState.ArrayStatus = *newArrayStatus
	}

	// If the current State is newly minted then we must initialize RetryAttempts to track how many
	// times each subtask is executed.
	if len(currentState.RetryAttempts.GetItems()) == 0 {
		count := uint(currentState.GetExecutionArraySize())
		maxValue := bitarray.Item(tCtx.TaskExecutionMetadata().GetMaxAttempts())

		retryAttemptsArray, err := bitarray.NewCompactArray(count, maxValue)
		if err != nil {
			logger.Errorf(ctx, "Failed to create attempts compact array with [count: %v, maxValue: %v]", count, maxValue)
			return currentState, externalResources, nil
		}

		// Initialize subtask retryAttempts to 0 so that, in tandem with the podName logic, we
		// maintain backwards compatibility.
		for i := 0; i < currentState.GetExecutionArraySize(); i++ {
			retryAttemptsArray.SetItem(i, 0)
		}

		currentState.RetryAttempts = retryAttemptsArray
	}

	// If the current State is newly minted then we must initialize SystemFailures to track how many
	// times the subtask failed due to system issues, this is necessary to correctly evaluate
	// interruptible subtasks.
	if len(currentState.SystemFailures.GetItems()) == 0 {
		count := uint(currentState.GetExecutionArraySize())
		maxValue := bitarray.Item(tCtx.TaskExecutionMetadata().GetMaxAttempts())

		systemFailuresArray, err := bitarray.NewCompactArray(count, maxValue)
		if err != nil {
			logger.Errorf(ctx, "Failed to create system failures array with [count: %v, maxValue: %v]", count, maxValue)
			return currentState, externalResources, err
		}

		for i := 0; i < currentState.GetExecutionArraySize(); i++ {
			systemFailuresArray.SetItem(i, 0)
		}

		currentState.SystemFailures = systemFailuresArray
	}

	// initialize log plugin
	logPlugin, err := logs.InitializeLogPlugins(&config.LogConfig.Config)
	if err != nil {
		return currentState, externalResources, err
	}

	// identify max parallelism
	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return currentState, externalResources, err
	} else if taskTemplate == nil {
		return currentState, externalResources, errors.Errorf(errors.BadTaskSpecification, "Required value not set, taskTemplate is nil")
	}

	arrayJob, err := arrayCore.ToArrayJob(taskTemplate.GetCustom(), taskTemplate.TaskTypeVersion)
	if err != nil {
		return currentState, externalResources, err
	}

	currentParallelism := 0
	maxParallelism := int(arrayJob.Parallelism)

	currentSubTaskPhaseHash, err := currentState.GetArrayStatus().HashCode()
	if err != nil {
		return currentState, externalResources, err
	}

	for childIdx, existingPhaseIdx := range currentState.GetArrayStatus().Detailed.GetItems() {
		existingPhase := core.Phases[existingPhaseIdx]
		retryAttempt := currentState.RetryAttempts.GetItem(childIdx)

		if existingPhase == core.PhaseRetryableFailure {
			retryAttempt++
			newState.RetryAttempts.SetItem(childIdx, retryAttempt)
		} else if existingPhase.IsTerminal() {
			newArrayStatus.Detailed.SetItem(childIdx, bitarray.Item(existingPhase))
			continue
		}

		originalIdx := arrayCore.CalculateOriginalIndex(childIdx, newState.GetIndexesToCache())
		systemFailures := currentState.SystemFailures.GetItem(childIdx)
		stCtx, err := NewSubTaskExecutionContext(ctx, tCtx, taskTemplate, childIdx, originalIdx, retryAttempt, systemFailures)
		if err != nil {
			return currentState, externalResources, err
		}
		podName := stCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()

		// depending on the existing subtask phase we either a launch new k8s resource or monitor
		// an existing instance
		var phaseInfo core.PhaseInfo
		var perr error
		if existingPhase == core.PhaseUndefined || existingPhase == core.PhaseWaitingForResources || existingPhase == core.PhaseRetryableFailure {
			// attempt to allocateResource
			allocationStatus, err := allocateResource(ctx, stCtx, config, podName)
			if err != nil {
				logger.Errorf(ctx, "Resource manager failed for TaskExecId [%s] token [%s]. error %s",
					stCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID(), podName, err)
				return currentState, externalResources, err
			}

			logger.Infof(ctx, "Allocation result for [%s] is [%s]", podName, allocationStatus)
			if allocationStatus != core.AllocationStatusGranted {
				phaseInfo = core.PhaseInfoWaitingForResourcesInfo(time.Now(), core.DefaultPhaseVersion, "Exceeded ResourceManager quota", nil)
			} else {
				phaseInfo, perr = launchSubtask(ctx, stCtx, config, kubeClient)

				// if launchSubtask fails we attempt to deallocate the (previously allocated)
				// resource to mitigate leaks
				if perr != nil {
					perr = deallocateResource(ctx, stCtx, config, podName)
					if perr != nil {
						logger.Errorf(ctx, "Error releasing allocation token [%s] in Finalize [%s]", podName, err)
					}
				}
			}
		} else {
			phaseInfo, perr = getSubtaskPhaseInfo(ctx, stCtx, config, kubeClient, logPlugin)
		}

		if perr != nil {
			return currentState, externalResources, perr
		}

		if phaseInfo.Err() != nil {
			messageCollector.Collect(childIdx, phaseInfo.Err().String())

			// If the service reported an error but there is no error.pb written, write one with the
			// service-provided error message.
			or, err := array.ConstructOutputReader(ctx, dataStore, outputPrefix, baseOutputDataSandbox, originalIdx)
			if err != nil {
				return currentState, externalResources, err
			}

			if hasErr, err := or.IsError(ctx); err != nil {
				return currentState, externalResources, err
			} else if !hasErr {
				// The subtask has not produced an error.pb, write one.
				ow, err := array.ConstructOutputWriter(ctx, dataStore, outputPrefix, baseOutputDataSandbox, originalIdx)
				if err != nil {
					return currentState, externalResources, err
				}

				if err = ow.Put(ctx, ioutils.NewInMemoryOutputReader(nil, nil, &io.ExecutionError{
					ExecutionError: phaseInfo.Err(),
					IsRecoverable:  phaseInfo.Phase() != core.PhasePermanentFailure,
				})); err != nil {
					return currentState, externalResources, err
				}
			}
		}

		if phaseInfo.Err() != nil && phaseInfo.Err().GetKind() == idlCore.ExecutionError_SYSTEM {
			newState.SystemFailures.SetItem(childIdx, systemFailures+1)
		} else {
			newState.SystemFailures.SetItem(childIdx, systemFailures)
		}

		// process subtask phase
		actualPhase := phaseInfo.Phase()
		if actualPhase.IsSuccess() {
			actualPhase, err = array.CheckTaskOutput(ctx, dataStore, outputPrefix, baseOutputDataSandbox, childIdx, originalIdx)
			if err != nil {
				return currentState, externalResources, err
			}
		}

		if actualPhase == core.PhaseRetryableFailure && uint32(retryAttempt+1) >= stCtx.TaskExecutionMetadata().GetMaxAttempts() {
			// If we see a retryable failure we must check if the number of retries exceeds the maximum
			// attempts. If so, transition to a permanent failure so that is not attempted again.
			actualPhase = core.PhasePermanentFailure
		}
		newArrayStatus.Detailed.SetItem(childIdx, bitarray.Item(actualPhase))

		if actualPhase.IsTerminal() {
			err = deallocateResource(ctx, stCtx, config, podName)
			if err != nil {
				logger.Errorf(ctx, "Error releasing allocation token [%s] in Finalize [%s]", podName, err)
				return currentState, externalResources, err
			}

			err = finalizeSubtask(ctx, stCtx, config, kubeClient)
			if err != nil {
				logger.Errorf(ctx, "Error finalizing resource [%s] in Finalize [%s]", podName, err)
				return currentState, externalResources, err
			}
		}

		// process phaseInfo
		var logLinks []*idlCore.TaskLog
		if phaseInfo.Info() != nil {
			logLinks = phaseInfo.Info().Logs
		}

		externalResources = append(externalResources, &core.ExternalResource{
			ExternalID:   podName,
			Index:        uint32(originalIdx),
			Logs:         logLinks,
			RetryAttempt: uint32(retryAttempt),
			Phase:        actualPhase,
		})

		// validate parallelism
		if !actualPhase.IsTerminal() || actualPhase == core.PhaseRetryableFailure {
			currentParallelism++
		}

		if maxParallelism != 0 && currentParallelism >= maxParallelism {
			break
		}
	}

	// compute task phase from array status summary
	for _, phaseIdx := range newArrayStatus.Detailed.GetItems() {
		newArrayStatus.Summary.Inc(core.Phases[phaseIdx])
	}

	phase := arrayCore.SummaryToPhase(ctx, currentState.GetOriginalMinSuccesses()-currentState.GetOriginalArraySize()+int64(currentState.GetExecutionArraySize()), newArrayStatus.Summary)

	// process new state
	newState = newState.SetArrayStatus(*newArrayStatus)
	if phase == arrayCore.PhaseWriteToDiscoveryThenFail {
		errorMsg := messageCollector.Summary(GetConfig().MaxErrorStringLength)
		newState = newState.SetReason(errorMsg)
	}

	_, version := currentState.GetPhase()
	if phase == arrayCore.PhaseCheckingSubTaskExecutions {
		newSubTaskPhaseHash, err := newState.GetArrayStatus().HashCode()
		if err != nil {
			return currentState, externalResources, err
		}

		if newSubTaskPhaseHash != currentSubTaskPhaseHash {
			version++
		}

		newState = newState.SetPhase(phase, version).SetReason("Task is still running")
	} else {
		newState = newState.SetPhase(phase, version+1)
	}

	return newState, externalResources, nil
}

// TerminateSubTasks performs operations to gracefully terminate all subtasks. This may include
// aborting and finalizing active k8s resources.
func TerminateSubTasks(ctx context.Context, tCtx core.TaskExecutionContext, kubeClient core.KubeClient, config *Config,
	terminateFunction func(context.Context, SubTaskExecutionContext, *Config, core.KubeClient) error, currentState *arrayCore.State) error {

	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return err
	} else if taskTemplate == nil {
		return errors.Errorf(errors.BadTaskSpecification, "Required value not set, taskTemplate is nil")
	}

	messageCollector := errorcollector.NewErrorMessageCollector()
	for childIdx, existingPhaseIdx := range currentState.GetArrayStatus().Detailed.GetItems() {
		existingPhase := core.Phases[existingPhaseIdx]
		retryAttempt := uint64(0)
		if childIdx < len(currentState.RetryAttempts.GetItems()) {
			// we can use RetryAttempts if it has been initialized, otherwise stay with default 0
			retryAttempt = currentState.RetryAttempts.GetItem(childIdx)
		}

		// return immediately if subtask has completed or not yet started
		if existingPhase.IsTerminal() || existingPhase == core.PhaseUndefined {
			continue
		}

		originalIdx := arrayCore.CalculateOriginalIndex(childIdx, currentState.GetIndexesToCache())
		stCtx, err := NewSubTaskExecutionContext(ctx, tCtx, taskTemplate, childIdx, originalIdx, retryAttempt, 0)
		if err != nil {
			return err
		}

		err = terminateFunction(ctx, stCtx, config, kubeClient)
		if err != nil {
			messageCollector.Collect(childIdx, err.Error())
		}
	}

	if messageCollector.Length() > 0 {
		return fmt.Errorf(messageCollector.Summary(config.MaxErrorStringLength))
	}

	return nil
}
