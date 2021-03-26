package sidecar

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/template"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery"
	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s"

	"github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyteplugins/go/tasks/logs"
	k8sv1 "k8s.io/api/core/v1"
)

const (
	sidecarTaskType     = "sidecar"
	primaryContainerKey = "primary_container_name"
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
		modifiedCommand, err := template.Render(ctx, container.Command, template.Parameters{
			TaskExecMetadata: taskCtx.TaskExecutionMetadata(),
			Inputs:           taskCtx.InputReader(),
			OutputPath:       taskCtx.OutputWriter(),
			Task:             taskCtx.TaskReader(),
		})
		if err != nil {
			return nil, err
		}
		container.Command = modifiedCommand

		modifiedArgs, err := template.Render(ctx, container.Args, template.Parameters{
			TaskExecMetadata: taskCtx.TaskExecutionMetadata(),
			Inputs:           taskCtx.InputReader(),
			OutputPath:       taskCtx.OutputWriter(),
			Task:             taskCtx.TaskReader(),
		})
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
	flytek8s.UpdatePod(taskCtx.TaskExecutionMetadata(), resReqs, &pod.Spec)
	return &pod, nil
}

// Why, you might wonder do we recreate the generated go struct generated from the plugins.SidecarJob proto? Because
// although we unmarshal the task custom json, the PodSpec itself is not generated from a  proto definition,
// but a proper go struct defined in k8s libraries. Therefore we only unmarshal the sidecar as a json, rather than jsonpb.
type sidecarJob struct {
	PodSpec              *k8sv1.PodSpec
	PrimaryContainerName string
}

func (sidecarResourceHandler) GetProperties() k8s.PluginProperties {
	return k8s.PluginProperties{}
}

func (sidecarResourceHandler) BuildResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (client.Object, error) {
	var podSpec k8sv1.PodSpec
	var primaryContainerName string

	task, err := taskCtx.TaskReader().Read(ctx)
	if err != nil {
		return nil, errors.Errorf(errors.BadTaskSpecification,
			"TaskSpecification cannot be read, Err: [%v]", err.Error())
	}
	if task.TaskTypeVersion == 0 {
		sidecarJob := sidecarJob{}
		err := utils.UnmarshalStructToObj(task.GetCustom(), &sidecarJob)
		if err != nil {
			return nil, errors.Errorf(errors.BadTaskSpecification,
				"invalid TaskSpecification [%v], Err: [%v]", task.GetCustom(), err.Error())
		}
		if sidecarJob.PodSpec == nil {
			return nil, errors.Errorf(errors.BadTaskSpecification,
				"invalid TaskSpecification, nil PodSpec [%v]", task.GetCustom())
		}
		podSpec = *sidecarJob.PodSpec
		primaryContainerName = sidecarJob.PrimaryContainerName
	} else {
		err := utils.UnmarshalStructToObj(task.GetCustom(), &podSpec)
		if err != nil {
			return nil, errors.Errorf(errors.BadTaskSpecification,
				"Unable to unmarshal task custom [%v], Err: [%v]", task.GetCustom(), err.Error())
		}
		if len(task.GetConfig()) == 0 {
			return nil, errors.Errorf(errors.BadTaskSpecification,
				"invalid TaskSpecification, config needs to be non-empty and include missing [%s] key", primaryContainerKey)
		}
		var ok bool
		primaryContainerName, ok = task.GetConfig()[primaryContainerKey]
		if !ok {
			return nil, errors.Errorf(errors.BadTaskSpecification,
				"invalid TaskSpecification, config missing [%s] key in [%v]", primaryContainerKey, task.GetConfig())
		}
	}

	pod := flytek8s.BuildPodWithSpec(&podSpec)
	// Set the restart policy to *not* inherit from the default so that a completed pod doesn't get caught in a
	// CrashLoopBackoff after the initial job completion.
	pod.Spec.RestartPolicy = k8sv1.RestartPolicyNever

	pod.Spec.ServiceAccountName = flytek8s.GetServiceAccountNameFromTaskExecutionMetadata(taskCtx.TaskExecutionMetadata())

	pod, err = validateAndFinalizePod(ctx, taskCtx, primaryContainerName, *pod)
	if err != nil {
		return nil, err
	}

	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string, 1)
	}

	pod.Annotations[primaryContainerKey] = primaryContainerName

	return pod, nil
}

func (sidecarResourceHandler) BuildIdentityResource(_ context.Context, _ pluginsCore.TaskExecutionMetadata) (
	client.Object, error) {
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

func (sidecarResourceHandler) GetTaskPhase(ctx context.Context, pluginContext k8s.PluginContext, r client.Object) (pluginsCore.PhaseInfo, error) {
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
			DefaultForTaskTypes: []pluginsCore.TaskType{sidecarTaskType},
		})
}
