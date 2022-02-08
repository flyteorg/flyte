package k8s

import (
	"context"
	"fmt"
	"time"

	idlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyteplugins/go/tasks/logs"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/tasklog"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/arraystatus"
	arrayCore "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/core"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/errorcollector"

	"github.com/flyteorg/flytestdlib/bitarray"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/storage"

	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sTypes "k8s.io/apimachinery/pkg/types"

	errors2 "github.com/flyteorg/flytestdlib/errors"
)

const (
	ErrCheckPodStatus errors2.ErrorCode = "CHECK_POD_FAILED"
)

func LaunchAndCheckSubTasksState(ctx context.Context, tCtx core.TaskExecutionContext, kubeClient core.KubeClient,
	config *Config, dataStore *storage.DataStore, outputPrefix, baseOutputDataSandbox storage.DataReference, currentState *arrayCore.State) (
	newState *arrayCore.State, logLinks []*idlCore.TaskLog, subTaskIDs []*string, err error) {
	if int64(currentState.GetExecutionArraySize()) > config.MaxArrayJobSize {
		ee := fmt.Errorf("array size > max allowed. Requested [%v]. Allowed [%v]", currentState.GetExecutionArraySize(), config.MaxArrayJobSize)
		logger.Info(ctx, ee)
		currentState = currentState.SetPhase(arrayCore.PhasePermanentFailure, 0).SetReason(ee.Error())
		return currentState, logLinks, subTaskIDs, nil
	}

	logLinks = make([]*idlCore.TaskLog, 0, 4)
	newState = currentState
	msg := errorcollector.NewErrorMessageCollector()
	newArrayStatus := &arraystatus.ArrayStatus{
		Summary:  arraystatus.ArraySummary{},
		Detailed: arrayCore.NewPhasesCompactArray(uint(currentState.GetExecutionArraySize())),
	}
	subTaskIDs = make([]*string, 0, len(currentState.GetArrayStatus().Detailed.GetItems()))

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
			logger.Errorf(context.Background(), "Failed to create attempts compact array with [count: %v, maxValue: %v]", count, maxValue)
			return currentState, logLinks, subTaskIDs, nil
		}

		// Set subtask retryAttempts using the existing task context retry attempt. For new tasks
		// this will initialize to 0, but running tasks will use the existing retry attempt.
		retryAttempt := bitarray.Item(tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID().RetryAttempt)
		for i := 0; i < currentState.GetExecutionArraySize(); i++ {
			retryAttemptsArray.SetItem(i, retryAttempt)
		}

		currentState.RetryAttempts = retryAttemptsArray
	}

	logPlugin, err := logs.InitializeLogPlugins(&config.LogConfig.Config)
	if err != nil {
		logger.Errorf(ctx, "Error initializing LogPlugins: [%s]", err)
		return currentState, logLinks, subTaskIDs, err
	}

	// identify max parallelism
	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return currentState, logLinks, subTaskIDs, err
	} else if taskTemplate == nil {
		return currentState, logLinks, subTaskIDs, errors.Errorf(errors.BadTaskSpecification, "Required value not set, taskTemplate is nil")
	}

	arrayJob, err := arrayCore.ToArrayJob(taskTemplate.GetCustom(), taskTemplate.TaskTypeVersion)
	if err != nil {
		return currentState, logLinks, subTaskIDs, err
	}

	currentParallelism := 0
	maxParallelism := int(arrayJob.Parallelism)

	for childIdx, existingPhaseIdx := range currentState.GetArrayStatus().Detailed.GetItems() {
		existingPhase := core.Phases[existingPhaseIdx]
		originalIdx := arrayCore.CalculateOriginalIndex(childIdx, newState.GetIndexesToCache())

		retryAttempt := currentState.RetryAttempts.GetItem(childIdx)
		podName := formatSubTaskName(ctx, tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), childIdx, retryAttempt)

		if existingPhase.IsTerminal() {
			// If we get here it means we have already "processed" this terminal phase since we will only persist
			// the phase after all processing is done (e.g. check outputs/errors file, record events... etc.).

			// Since we know we have already "processed" this terminal phase we can safely deallocate resource
			err = deallocateResource(ctx, tCtx, config, podName)
			if err != nil {
				logger.Errorf(ctx, "Error releasing allocation token [%s] in LaunchAndCheckSubTasks [%s]", podName, err)
				return currentState, logLinks, subTaskIDs, errors2.Wrapf(ErrCheckPodStatus, err, "Error releasing allocation token.")
			}

			// If a subtask is marked as a retryable failure we check if the number of retries
			// exceeds the maximum attempts. If so, transition the task to a permanent failure
			// so that is not attempted again. If it can be retried, increment the retry attempts
			// value and transition the task to "Undefined" so that it is reevaluated.
			if existingPhase == core.PhaseRetryableFailure {
				if uint32(retryAttempt+1) < tCtx.TaskExecutionMetadata().GetMaxAttempts() {
					newState.RetryAttempts.SetItem(childIdx, retryAttempt+1)

					newArrayStatus.Summary.Inc(core.PhaseUndefined)
					newArrayStatus.Detailed.SetItem(childIdx, bitarray.Item(core.PhaseUndefined))
					continue
				} else {
					existingPhase = core.PhasePermanentFailure
				}
			}

			newArrayStatus.Summary.Inc(existingPhase)
			newArrayStatus.Detailed.SetItem(childIdx, bitarray.Item(existingPhase))

			phaseInfo, err := FetchPodStatusAndLogs(ctx, kubeClient,
				k8sTypes.NamespacedName{
					Name:      podName,
					Namespace: GetNamespaceForExecution(tCtx, config.NamespaceTemplate),
				},
				originalIdx,
				tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID().RetryAttempt,
				retryAttempt,
				logPlugin)

			if err != nil {
				return currentState, logLinks, subTaskIDs, err
			}

			if phaseInfo.Info() != nil {
				logLinks = append(logLinks, phaseInfo.Info().Logs...)
			}

			continue
		}

		task := &Task{
			State:            newState,
			NewArrayStatus:   newArrayStatus,
			Config:           config,
			ChildIdx:         childIdx,
			OriginalIndex:    originalIdx,
			MessageCollector: &msg,
			SubTaskIDs:       subTaskIDs,
		}

		// The first time we enter this state we will launch every subtask. On subsequent rounds, the pod
		// has already been created so we return a Success value and continue with the Monitor step.
		var launchResult LaunchResult
		launchResult, err = task.Launch(ctx, tCtx, kubeClient)
		if err != nil {
			logger.Errorf(ctx, "K8s array - Launch error %v", err)
			return currentState, logLinks, subTaskIDs, err
		}

		switch launchResult {
		case LaunchSuccess:
			// Continue with execution if successful
		case LaunchError:
			return currentState, logLinks, subTaskIDs, err
		// If Resource manager is enabled and there are currently not enough resources we can skip this round
		// for a subtask and wait until there are enough resources.
		case LaunchWaiting:
			continue
		case LaunchReturnState:
			return currentState, logLinks, subTaskIDs, nil
		}

		monitorResult, taskLogs, err := task.Monitor(ctx, tCtx, kubeClient, dataStore, outputPrefix, baseOutputDataSandbox, logPlugin)

		if len(taskLogs) > 0 {
			logLinks = append(logLinks, taskLogs...)
		}
		subTaskIDs = task.SubTaskIDs

		if monitorResult != MonitorSuccess {
			if err != nil {
				logger.Errorf(ctx, "K8s array - Monitor error %v", err)
			}
			return currentState, logLinks, subTaskIDs, err
		}

		// validate map task parallelism
		newSubtaskPhase := core.Phases[newArrayStatus.Detailed.GetItem(childIdx)]
		if !newSubtaskPhase.IsTerminal() || newSubtaskPhase == core.PhaseRetryableFailure {
			currentParallelism++
		}

		if maxParallelism != 0 && currentParallelism >= maxParallelism {
			// If max parallelism has been achieved we need to fill the subtask phase summary with
			// the remaining subtasks so the overall map task phase can be accurately identified.
			for i := childIdx + 1; i < len(currentState.GetArrayStatus().Detailed.GetItems()); i++ {
				childSubtaskPhase := core.Phases[newArrayStatus.Detailed.GetItem(i)]
				newArrayStatus.Summary.Inc(childSubtaskPhase)
			}

			break
		}
	}

	newState = newState.SetArrayStatus(*newArrayStatus)

	phase := arrayCore.SummaryToPhase(ctx, currentState.GetOriginalMinSuccesses()-currentState.GetOriginalArraySize()+int64(currentState.GetExecutionArraySize()), newArrayStatus.Summary)
	if phase == arrayCore.PhaseWriteToDiscoveryThenFail {
		errorMsg := msg.Summary(GetConfig().MaxErrorStringLength)
		newState = newState.SetReason(errorMsg)
	}

	if phase == arrayCore.PhaseCheckingSubTaskExecutions {
		newPhaseVersion := uint32(0)

		// For now, the only changes to PhaseVersion and PreviousSummary occur for running array jobs.
		for phase, count := range newState.GetArrayStatus().Summary {
			newPhaseVersion += uint32(phase) * uint32(count)
		}

		newState = newState.SetPhase(phase, newPhaseVersion).SetReason("Task is still running.")
	} else {
		newState = newState.SetPhase(phase, core.DefaultPhaseVersion)
	}

	return newState, logLinks, subTaskIDs, nil
}

func FetchPodStatusAndLogs(ctx context.Context, client core.KubeClient, name k8sTypes.NamespacedName, index int, retryAttempt uint32, subtaskRetryAttempt uint64, logPlugin tasklog.Plugin) (
	info core.PhaseInfo, err error) {

	pod := &v1.Pod{
		TypeMeta: metaV1.TypeMeta{
			Kind:       PodKind,
			APIVersion: v1.SchemeGroupVersion.String(),
		},
	}

	err = client.GetClient().Get(ctx, name, pod)
	now := time.Now()

	if err != nil {
		if k8serrors.IsNotFound(err) {
			// If the object disappeared at this point, it means it was manually removed or garbage collected.
			// Mark it as a failure.
			return core.PhaseInfoFailed(core.PhaseRetryableFailure, &idlCore.ExecutionError{
				Code:    string(k8serrors.ReasonForError(err)),
				Message: err.Error(),
				Kind:    idlCore.ExecutionError_SYSTEM,
			}, &core.TaskInfo{
				OccurredAt: &now,
			}), nil
		}

		return info, err
	}

	t := flytek8s.GetLastTransitionOccurredAt(pod).Time
	taskInfo := core.TaskInfo{
		OccurredAt: &t,
	}

	if pod.Status.Phase != v1.PodPending && pod.Status.Phase != v1.PodUnknown {
		// We append the subtaskRetryAttempt to the log name only when it is > 0 to ensure backwards
		// compatibility when dynamically transitioning running map tasks to use subtask retry attempts.
		var logName string
		if subtaskRetryAttempt == 0 {
			logName = fmt.Sprintf(" #%d-%d", retryAttempt, index)
		} else {
			logName = fmt.Sprintf(" #%d-%d-%d", retryAttempt, index, subtaskRetryAttempt)
		}

		if logPlugin != nil {
			o, err := logPlugin.GetTaskLogs(tasklog.Input{
				PodName:          pod.Name,
				Namespace:        pod.Namespace,
				LogName:          logName,
				PodUnixStartTime: pod.CreationTimestamp.Unix(),
			})

			if err != nil {
				return core.PhaseInfoUndefined, err
			}
			taskInfo.Logs = o.TaskLogs
		}
	}

	var phaseInfo core.PhaseInfo
	var err2 error

	switch pod.Status.Phase {
	case v1.PodSucceeded:
		phaseInfo, err2 = flytek8s.DemystifySuccess(pod.Status, taskInfo)
	case v1.PodFailed:
		phaseInfo, err2 = flytek8s.DemystifyFailure(pod.Status, taskInfo)
	case v1.PodPending:
		phaseInfo, err2 = flytek8s.DemystifyPending(pod.Status)
	case v1.PodUnknown:
		phaseInfo = core.PhaseInfoUndefined
	default:
		primaryContainerName, ok := pod.GetAnnotations()[primaryContainerKey]
		if ok {
			// Special handling for determining the phase of an array job for a Pod task.
			phaseInfo = flytek8s.DeterminePrimaryContainerPhase(primaryContainerName, pod.Status.ContainerStatuses, &taskInfo)
			if phaseInfo.Phase() == core.PhaseRunning && len(taskInfo.Logs) > 0 {
				return core.PhaseInfoRunning(core.DefaultPhaseVersion+1, phaseInfo.Info()), nil
			}
			return phaseInfo, nil
		}

		if len(taskInfo.Logs) > 0 {
			phaseInfo = core.PhaseInfoRunning(core.DefaultPhaseVersion+1, &taskInfo)
		} else {
			phaseInfo = core.PhaseInfoRunning(core.DefaultPhaseVersion, &taskInfo)
		}
	}

	if err2 == nil && phaseInfo.Info() != nil {
		// Append sub-job status in Log Name for viz.
		for _, log := range phaseInfo.Info().Logs {
			log.Name += fmt.Sprintf(" (%s)", phaseInfo.Phase().String())
		}
	}

	return phaseInfo, err2

}
