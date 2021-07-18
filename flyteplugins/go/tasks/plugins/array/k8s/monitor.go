package k8s

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/tasklog"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/storage"

	arrayCore "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/core"

	"github.com/flyteorg/flytestdlib/bitarray"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/arraystatus"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/errorcollector"

	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sTypes "k8s.io/apimachinery/pkg/types"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s"

	idlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	errors2 "github.com/flyteorg/flytestdlib/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/flyteorg/flyteplugins/go/tasks/logs"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
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

	logPlugin, err := logs.InitializeLogPlugins(&config.LogConfig.Config)
	if err != nil {
		logger.Errorf(ctx, "Error initializing LogPlugins: [%s]", err)
		return currentState, logLinks, subTaskIDs, err
	}

	for childIdx, existingPhaseIdx := range currentState.GetArrayStatus().Detailed.GetItems() {
		existingPhase := core.Phases[existingPhaseIdx]
		indexStr := strconv.Itoa(childIdx)
		podName := formatSubTaskName(ctx, tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), indexStr)

		if existingPhase.IsTerminal() {
			// If we get here it means we have already "processed" this terminal phase since we will only persist
			// the phase after all processing is done (e.g. check outputs/errors file, record events... etc.).

			// Since we know we have already "processed" this terminal phase we can safely deallocate resource
			err = deallocateResource(ctx, tCtx, config, childIdx)
			if err != nil {
				logger.Errorf(ctx, "Error releasing allocation token [%s] in LaunchAndCheckSubTasks [%s]", podName, err)
				return currentState, logLinks, subTaskIDs, errors2.Wrapf(ErrCheckPodStatus, err, "Error releasing allocation token.")
			}
			newArrayStatus.Summary.Inc(existingPhase)
			newArrayStatus.Detailed.SetItem(childIdx, bitarray.Item(existingPhase))
			originalIdx := arrayCore.CalculateOriginalIndex(childIdx, newState.GetIndexesToCache())

			phaseInfo, err := FetchPodStatusAndLogs(ctx, kubeClient,
				k8sTypes.NamespacedName{
					Name:      podName,
					Namespace: GetNamespaceForExecution(tCtx, config.NamespaceTemplate),
				},
				originalIdx,
				tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID().RetryAttempt,
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

		var monitorResult MonitorResult
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
	}

	newState = newState.SetArrayStatus(*newArrayStatus)

	// Check that the taskTemplate is valid
	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return currentState, logLinks, subTaskIDs, err
	} else if taskTemplate == nil {
		return currentState, logLinks, subTaskIDs, fmt.Errorf("required value not set, taskTemplate is nil")
	}

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

func FetchPodStatusAndLogs(ctx context.Context, client core.KubeClient, name k8sTypes.NamespacedName, index int, retryAttempt uint32, logPlugin tasklog.Plugin) (
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

		if logPlugin != nil {
			o, err := logPlugin.GetTaskLogs(tasklog.Input{
				PodName:          pod.Name,
				Namespace:        pod.Namespace,
				LogName:          fmt.Sprintf(" #%d-%d", retryAttempt, index),
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
		code, message := flytek8s.ConvertPodFailureToError(pod.Status)
		phaseInfo = core.PhaseInfoRetryableFailure(code, message, &taskInfo)
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
