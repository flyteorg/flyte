package sidecar

import (
	"context"
	"fmt"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/plugins"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"

	pluginsCore "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/k8s"

	"github.com/lyft/flyteplugins/go/tasks/errors"
	"github.com/lyft/flyteplugins/go/tasks/logs"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/utils"

	k8sv1 "k8s.io/api/core/v1"
)

const (
	sidecarTaskType     = "sidecar"
	primaryContainerKey = "primary"
)

type sidecarResourceHandler struct{}

// This method handles templatizing primary container input args, env variables and adds a GPU toleration to the pod
// spec if necessary.
func validateAndFinalizePod(
	ctx context.Context, taskCtx pluginsCore.TaskExecutionContext, primaryContainerName string, pod k8sv1.Pod) (*k8sv1.Pod, error) {
	var hasPrimaryContainer bool

	finalizedContainers := make([]k8sv1.Container, len(pod.Spec.Containers))
	resReqs := make([]k8sv1.ResourceRequirements, 0, len(pod.Spec.Containers))
	for index, container := range pod.Spec.Containers {
		if container.Name == primaryContainerName {
			hasPrimaryContainer = true
		}
		modifiedCommand, err := utils.ReplaceTemplateCommandArgs(ctx, container.Command, taskCtx.InputReader(), taskCtx.OutputWriter())
		if err != nil {
			return nil, err
		}
		container.Command = modifiedCommand

		modifiedArgs, err := utils.ReplaceTemplateCommandArgs(ctx, container.Args, taskCtx.InputReader(), taskCtx.OutputWriter())
		if err != nil {
			return nil, err
		}
		container.Args = modifiedArgs
		container.Env = flytek8s.DecorateEnvVars(ctx, container.Env, taskCtx.TaskExecutionMetadata().GetTaskExecutionID())
		resources := flytek8s.ApplyResourceOverrides(ctx, container.Resources)
		resReqs = append(resReqs, *resources)
		finalizedContainers[index] = container
	}
	if !hasPrimaryContainer {
		return nil, errors.Errorf(errors.BadTaskSpecification,
			"invalid Sidecar task, primary container [%s] not defined", primaryContainerName)

	}
	pod.Spec.Containers = finalizedContainers
	if pod.Spec.Tolerations == nil {
		pod.Spec.Tolerations = make([]k8sv1.Toleration, 0)
	}
	pod.Spec.Tolerations = append(
		flytek8s.GetPodTolerations(taskCtx.TaskExecutionMetadata().IsInterruptible(), resReqs...), pod.Spec.Tolerations...)
	if taskCtx.TaskExecutionMetadata().IsInterruptible() && len(config.GetK8sPluginConfig().InterruptibleNodeSelector) > 0 {
		pod.Spec.NodeSelector = config.GetK8sPluginConfig().InterruptibleNodeSelector
	}

	return &pod, nil
}

func (sidecarResourceHandler) BuildResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (k8s.Resource, error) {
	sidecarJob := plugins.SidecarJob{}
	task, err := taskCtx.TaskReader().Read(ctx)
	if err != nil {
		return nil, errors.Errorf(errors.BadTaskSpecification,
			"TaskSpecification cannot be read, Err: [%v]", err.Error())
	}
	err = utils.UnmarshalStruct(task.GetCustom(), &sidecarJob)
	if err != nil {
		return nil, errors.Errorf(errors.BadTaskSpecification,
			"invalid TaskSpecification [%v], Err: [%v]", task.GetCustom(), err.Error())
	}

	pod := flytek8s.BuildPodWithSpec(sidecarJob.PodSpec)
	// Set the restart policy to *not* inherit from the default so that a completed pod doesn't get caught in a
	// CrashLoopBackoff after the initial job completion.
	pod.Spec.RestartPolicy = k8sv1.RestartPolicyNever

	// We want to Also update the serviceAccount to the serviceaccount of the workflow
	pod.Spec.ServiceAccountName = taskCtx.TaskExecutionMetadata().GetK8sServiceAccount()

	pod, err = validateAndFinalizePod(ctx, taskCtx, sidecarJob.PrimaryContainerName, *pod)
	if err != nil {
		return nil, err
	}

	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string, 1)
	}

	pod.Annotations[primaryContainerKey] = sidecarJob.PrimaryContainerName

	return pod, nil
}

func (sidecarResourceHandler) BuildIdentityResource(_ context.Context, _ pluginsCore.TaskExecutionMetadata) (
	k8s.Resource, error) {
	return flytek8s.BuildIdentityPod(), nil
}

func determinePrimaryContainerPhase(primaryContainerName string, statuses []k8sv1.ContainerStatus, info *pluginsCore.TaskInfo) pluginsCore.PhaseInfo {
	for _, s := range statuses {
		if s.Name == primaryContainerName {
			if s.State.Waiting != nil || s.State.Running != nil {
				return pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion, info)
			}

			if s.State.Terminated != nil {
				if s.State.Terminated.ExitCode != 0 {
					return pluginsCore.PhaseInfoRetryableFailure(
						s.State.Terminated.Reason, s.State.Terminated.Message, info)
				}
				return pluginsCore.PhaseInfoSuccess(info)
			}
		}
	}

	// If for some reason we can't find the primary container, always just return a permanent failure
	return pluginsCore.PhaseInfoFailure("PrimaryContainerMissing",
		fmt.Sprintf("Primary container [%s] not found in pod's container statuses", primaryContainerName), info)
}

func (sidecarResourceHandler) GetTaskPhase(ctx context.Context, pluginContext k8s.PluginContext, r k8s.Resource) (pluginsCore.PhaseInfo, error) {
	pod := r.(*k8sv1.Pod)

	transitionOccurredAt := flytek8s.GetLastTransitionOccurredAt(pod).Time
	info := pluginsCore.TaskInfo{
		OccurredAt: &transitionOccurredAt,
	}
	if pod.Status.Phase != k8sv1.PodPending && pod.Status.Phase != k8sv1.PodUnknown {
		taskLogs, err := logs.GetLogsForContainerInPod(ctx, pod, 0, " (User)")
		if err != nil {
			return pluginsCore.PhaseInfoUndefined, err
		}
		info.Logs = taskLogs
	}
	switch pod.Status.Phase {
	case k8sv1.PodSucceeded:
		return flytek8s.DemystifySuccess(pod.Status, info)
	case k8sv1.PodFailed:
		code, message := flytek8s.ConvertPodFailureToError(pod.Status)
		return pluginsCore.PhaseInfoRetryableFailure(code, message, &info), nil
	case k8sv1.PodPending:
		return flytek8s.DemystifyPending(pod.Status)
	case k8sv1.PodReasonUnschedulable:
		return pluginsCore.PhaseInfoQueued(transitionOccurredAt, pluginsCore.DefaultPhaseVersion, "pod unschedulable"), nil
	case k8sv1.PodUnknown:
		return pluginsCore.PhaseInfoUndefined, nil
	}

	// Otherwise, assume the pod is running.
	primaryContainerName, ok := r.GetAnnotations()[primaryContainerKey]
	if !ok {
		return pluginsCore.PhaseInfoUndefined, errors.Errorf(errors.BadTaskSpecification,
			"missing primary container annotation for pod")
	}
	primaryContainerPhase := determinePrimaryContainerPhase(primaryContainerName, pod.Status.ContainerStatuses, &info)

	if primaryContainerPhase.Phase() == pluginsCore.PhaseRunning && len(info.Logs) > 0 {
		return pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion+1, primaryContainerPhase.Info()), nil
	}
	return primaryContainerPhase, nil
}

func init() {
	pluginmachinery.PluginRegistry().RegisterK8sPlugin(
		k8s.PluginEntry{
			ID:                  sidecarTaskType,
			RegisteredTaskTypes: []pluginsCore.TaskType{sidecarTaskType},
			ResourceToWatch:     &k8sv1.Pod{},
			Plugin:              sidecarResourceHandler{},
			IsDefault:           false,
		})
}
