package k8s

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/lyft/flytestdlib/storage"

	"github.com/lyft/flyteplugins/go/tasks/plugins/array"

	arrayCore "github.com/lyft/flyteplugins/go/tasks/plugins/array/core"

	"github.com/lyft/flyteplugins/go/tasks/plugins/array/arraystatus"
	"github.com/lyft/flyteplugins/go/tasks/plugins/array/errorcollector"
	"github.com/lyft/flytestdlib/bitarray"

	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sTypes "k8s.io/apimachinery/pkg/types"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/flytek8s"

	idlCore "github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytestdlib/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/lyft/flyteplugins/go/tasks/logs"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
)

const (
	ErrCheckPodStatus errors.ErrorCode = "CHECK_POD_FAILED"
)

func CheckSubTasksState(ctx context.Context, tCtx core.TaskExecutionContext, kubeClient core.KubeClient,
	dataStore *storage.DataStore, outputPrefix storage.DataReference, currentState *arrayCore.State) (
	newState *arrayCore.State, logLinks []*idlCore.TaskLog, err error) {

	logLinks = make([]*idlCore.TaskLog, 0, 4)
	newState = currentState

	msg := errorcollector.NewErrorMessageCollector()
	newArrayStatus := arraystatus.ArrayStatus{
		Summary:  arraystatus.ArraySummary{},
		Detailed: arrayCore.NewPhasesCompactArray(uint(currentState.GetExecutionArraySize())),
	}

	for childIdx, existingPhaseIdx := range currentState.GetArrayStatus().Detailed.GetItems() {
		existingPhase := core.Phases[existingPhaseIdx]
		if existingPhase.IsTerminal() {
			// If we get here it means we have already "processed" this terminal phase since we will only persist
			// the phase after all processing is done (e.g. check outputs/errors file, record events... etc.).
			newArrayStatus.Summary.Inc(existingPhase)
			newArrayStatus.Detailed.SetItem(childIdx, bitarray.Item(existingPhase))

			// TODO: collect log links before doing this
			continue
		}

		phaseInfo, err := CheckPodStatus(ctx, kubeClient,
			k8sTypes.NamespacedName{
				Name:      formatSubTaskName(ctx, tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), strconv.Itoa(childIdx)),
				Namespace: tCtx.TaskExecutionMetadata().GetNamespace(),
			})
		if err != nil {
			return currentState, logLinks, errors.Wrapf(ErrCheckPodStatus, err, "Failed to check pod status")
		}

		if phaseInfo.Info() != nil {
			logLinks = append(logLinks, phaseInfo.Info().Logs...)
		}

		if phaseInfo.Err() != nil {
			msg.Collect(childIdx, phaseInfo.Err().String())
		}

		actualPhase := phaseInfo.Phase()
		if phaseInfo.Phase().IsSuccess() {
			originalIdx := arrayCore.CalculateOriginalIndex(childIdx, currentState.GetIndexesToCache())
			actualPhase, err = array.CheckTaskOutput(ctx, dataStore, outputPrefix, childIdx, originalIdx)
			if err != nil {
				return nil, nil, err
			}
		}

		newArrayStatus.Detailed.SetItem(childIdx, bitarray.Item(actualPhase))
		newArrayStatus.Summary.Inc(actualPhase)
	}

	newState = newState.SetArrayStatus(newArrayStatus)

	// Check that the taskTemplate is valid
	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return currentState, logLinks, err
	} else if taskTemplate == nil {
		return currentState, logLinks, fmt.Errorf("required value not set, taskTemplate is nil")
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

	return newState, logLinks, nil
}

func CheckPodStatus(ctx context.Context, client core.KubeClient, name k8sTypes.NamespacedName) (
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
				Kind: idlCore.ExecutionError_SYSTEM,
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
		taskLogs, err := logs.GetLogsForContainerInPod(ctx, pod, 0, " (User)")
		if err != nil {
			return core.PhaseInfoUndefined, err
		}
		taskInfo.Logs = taskLogs
	}
	switch pod.Status.Phase {
	case v1.PodSucceeded:
		return flytek8s.DemystifySuccess(pod.Status, taskInfo)
	case v1.PodFailed:
		code, message := flytek8s.ConvertPodFailureToError(pod.Status)
		return core.PhaseInfoRetryableFailure(code, message, &taskInfo), nil
	case v1.PodPending:
		return flytek8s.DemystifyPending(pod.Status)
	case v1.PodUnknown:
		return core.PhaseInfoUndefined, nil
	}
	if len(taskInfo.Logs) > 0 {
		return core.PhaseInfoRunning(core.DefaultPhaseVersion+1, &taskInfo), nil
	}
	return core.PhaseInfoRunning(core.DefaultPhaseVersion, &taskInfo), nil

}
