package k8splugins

import (
	"context"
	"fmt"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/plugins"

	v1 "github.com/lyft/flyteplugins/go/tasks/v1"
	"github.com/lyft/flyteplugins/go/tasks/v1/errors"
	"github.com/lyft/flyteplugins/go/tasks/v1/events"
	"github.com/lyft/flyteplugins/go/tasks/v1/logs"
	"github.com/lyft/flyteplugins/go/tasks/v1/utils"

	"github.com/lyft/flyteplugins/go/tasks/v1/flytek8s"

	k8sv1 "k8s.io/api/core/v1"

	"github.com/lyft/flyteplugins/go/tasks/v1/types"
)

const (
	sidecarTaskType     = "sidecar"
	primaryContainerKey = "primary"
)

type sidecarResourceHandler struct{}

// This method handles templatizing primary container input args, env variables and adds a GPU toleration to the pod
// spec if necessary.
func validateAndFinalizeContainers(
	ctx context.Context, taskCtx types.TaskContext, primaryContainerName string, pod k8sv1.Pod,
	inputs *core.LiteralMap) (*k8sv1.Pod, error) {
	var hasPrimaryContainer bool

	finalizedContainers := make([]k8sv1.Container, len(pod.Spec.Containers))
	resReqs := make([]k8sv1.ResourceRequirements, 0, len(pod.Spec.Containers))
	for index, container := range pod.Spec.Containers {
		if container.Name == primaryContainerName {
			hasPrimaryContainer = true
		}
		modifiedCommand, err := utils.ReplaceTemplateCommandArgs(ctx,
			container.Command,
			utils.CommandLineTemplateArgs{
				Input:        taskCtx.GetInputsFile().String(),
				OutputPrefix: taskCtx.GetDataDir().String(),
				Inputs:       utils.LiteralMapToTemplateArgs(ctx, inputs),
			})

		if err != nil {
			return nil, err
		}
		container.Command = modifiedCommand

		modifiedArgs, err := utils.ReplaceTemplateCommandArgs(ctx,
			container.Args,
			utils.CommandLineTemplateArgs{
				Input:        taskCtx.GetInputsFile().String(),
				OutputPrefix: taskCtx.GetDataDir().String(),
				Inputs:       utils.LiteralMapToTemplateArgs(ctx, inputs),
			})

		if err != nil {
			return nil, err
		}
		container.Args = modifiedArgs

		container.Env = flytek8s.DecorateEnvVars(ctx, container.Env, taskCtx.GetTaskExecutionID())
		resources := flytek8s.ApplyResourceOverrides(ctx, container.Resources)
		resReqs = append(resReqs, *resources)
		finalizedContainers[index] = container
	}
	if !hasPrimaryContainer {
		return nil, errors.Errorf(errors.BadTaskSpecification,
			"invalid Sidecar task, primary container [%s] not defined", primaryContainerName)

	}
	pod.Spec.Containers = finalizedContainers
	pod.Spec.Tolerations = flytek8s.GetTolerationsForResources(resReqs...)
	return &pod, nil
}

func (sidecarResourceHandler) BuildResource(
	ctx context.Context, taskCtx types.TaskContext, task *core.TaskTemplate, inputs *core.LiteralMap) (
	flytek8s.K8sResource, error) {
	sidecarJob := plugins.SidecarJob{}
	err := utils.UnmarshalStruct(task.GetCustom(), &sidecarJob)
	if err != nil {
		return nil, errors.Errorf(errors.BadTaskSpecification,
			"invalid TaskSpecification [%v], Err: [%v]", task.GetCustom(), err.Error())
	}

	pod := flytek8s.BuildPodWithSpec(sidecarJob.PodSpec)
	// Set the restart policy to *not* inherit from the default so that a completed pod doesn't get caught in a
	// CrashLoopBackoff after the initial job completion.
	pod.Spec.RestartPolicy = k8sv1.RestartPolicyNever

	// We want to Also update the serviceAccount to the serviceaccount of the workflow
	pod.Spec.ServiceAccountName = taskCtx.GetK8sServiceAccount()

	pod, err = validateAndFinalizeContainers(ctx, taskCtx, sidecarJob.PrimaryContainerName, *pod, inputs)
	if err != nil {
		return nil, err
	}

	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string, 1)
	}

	pod.Annotations[primaryContainerKey] = sidecarJob.PrimaryContainerName

	return pod, nil
}

func (sidecarResourceHandler) BuildIdentityResource(ctx context.Context, taskCtx types.TaskContext) (
	flytek8s.K8sResource, error) {
	return flytek8s.BuildIdentityPod(), nil
}

func determinePrimaryContainerStatus(primaryContainerName string, statuses []k8sv1.ContainerStatus) (
	types.TaskStatus, error) {
	for _, s := range statuses {
		if s.Name == primaryContainerName {
			if s.State.Waiting != nil || s.State.Running != nil {
				return types.TaskStatusRunning, nil
			}

			if s.State.Terminated != nil {
				if s.State.Terminated.ExitCode != 0 {
					return types.TaskStatusRetryableFailure(errors.Errorf(
						s.State.Terminated.Reason, s.State.Terminated.Message)), nil
				}
				return types.TaskStatusSucceeded, nil
			}
		}
	}

	// If for some reason we can't find the primary container, always just return a permanent failure
	return types.TaskStatusPermanentFailure(errors.Errorf("PrimaryContainerMissing",
		"Primary container [%s] not found in pod's container statuses", primaryContainerName)), nil
}

func (sidecarResourceHandler) GetTaskStatus(
	ctx context.Context, taskCtx types.TaskContext, resource flytek8s.K8sResource) (
	types.TaskStatus, *events.TaskEventInfo, error) {
	pod := resource.(*k8sv1.Pod)

	var info *events.TaskEventInfo
	if pod.Status.Phase != k8sv1.PodPending && pod.Status.Phase != k8sv1.PodUnknown {
		taskLogs := make([]*core.TaskLog, 0)
		for idx, container := range pod.Spec.Containers {
			containerLogs, err := logs.GetLogsForContainerInPod(ctx, pod, uint32(idx), fmt.Sprintf(" (%s)", container.Name))
			if err != nil {
				return types.TaskStatusUndefined, nil, err
			}
			taskLogs = append(taskLogs, containerLogs...)
		}

		t := GetLastTransitionOccurredAt(pod).Time
		info = &events.TaskEventInfo{
			Logs:       taskLogs,
			OccurredAt: &t,
		}
	}
	switch pod.Status.Phase {
	case k8sv1.PodSucceeded:
		return types.TaskStatusSucceeded, info, nil
	case k8sv1.PodFailed:
		return types.TaskStatusRetryableFailure(ConvertPodFailureToError(pod.Status)), info, nil
	case k8sv1.PodPending:
		status, err := flytek8s.DemystifyPending(pod.Status)
		return status, info, err
	case k8sv1.PodReasonUnschedulable:
		return types.TaskStatusQueued, info, nil
	case k8sv1.PodUnknown:
		return types.TaskStatusUnknown, info, nil
	}

	// Otherwise, assume the pod is running.
	primaryContainerName, ok := resource.GetAnnotations()[primaryContainerKey]
	if !ok {
		return types.TaskStatusUndefined, nil, errors.Errorf(errors.BadTaskSpecification,
			"missing primary container annotation for pod")
	}

	status, err := determinePrimaryContainerStatus(primaryContainerName, pod.Status.ContainerStatuses)
	return status, info, err
}

func init() {
	v1.RegisterLoader(func(ctx context.Context) error {
		return v1.K8sRegisterForTaskTypes(sidecarTaskType, &k8sv1.Pod{}, flytek8s.DefaultInformerResyncDuration,
			sidecarResourceHandler{}, sidecarTaskType)
	})
}
