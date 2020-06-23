package container

import (
	"context"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/flytek8s"

	v1 "k8s.io/api/core/v1"

	"github.com/lyft/flyteplugins/go/tasks/logs"
	pluginsCore "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/k8s"
)

const (
	containerTaskType = "container"
)

type Plugin struct {
}

func (Plugin) GetTaskPhase(ctx context.Context, pluginContext k8s.PluginContext, r k8s.Resource) (pluginsCore.PhaseInfo, error) {

	pod := r.(*v1.Pod)

	t := flytek8s.GetLastTransitionOccurredAt(pod).Time
	info := pluginsCore.TaskInfo{
		OccurredAt: &t,
	}
	if pod.Status.Phase != v1.PodPending && pod.Status.Phase != v1.PodUnknown {
		taskLogs, err := logs.GetLogsForContainerInPod(ctx, pod, 0, " (User)")
		if err != nil {
			return pluginsCore.PhaseInfoUndefined, err
		}
		info.Logs = taskLogs
	}
	switch pod.Status.Phase {
	case v1.PodSucceeded:
		return flytek8s.DemystifySuccess(pod.Status, info)
	case v1.PodFailed:
		code, message := flytek8s.ConvertPodFailureToError(pod.Status)
		return pluginsCore.PhaseInfoRetryableFailure(code, message, &info), nil
	case v1.PodPending:
		return flytek8s.DemystifyPending(pod.Status)
	case v1.PodUnknown:
		return pluginsCore.PhaseInfoUndefined, nil
	}
	if len(info.Logs) > 0 {
		return pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion+1, &info), nil
	}
	return pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion, &info), nil
}

// Creates a new Pod that will Exit on completion. The pods have no retries by design
func (Plugin) BuildResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (k8s.Resource, error) {

	podSpec, err := flytek8s.ToK8sPodSpec(ctx, taskCtx.TaskExecutionMetadata(), taskCtx.TaskReader(), taskCtx.InputReader(), taskCtx.OutputWriter())
	if err != nil {
		return nil, err
	}

	pod := flytek8s.BuildPodWithSpec(podSpec)

	// We want to Also update the serviceAccount to the serviceaccount of the workflow
	pod.Spec.ServiceAccountName = taskCtx.TaskExecutionMetadata().GetK8sServiceAccount()

	return pod, nil
}

func (Plugin) BuildIdentityResource(_ context.Context, _ pluginsCore.TaskExecutionMetadata) (k8s.Resource, error) {
	return flytek8s.BuildIdentityPod(), nil
}

func init() {
	pluginmachinery.PluginRegistry().RegisterK8sPlugin(
		k8s.PluginEntry{
			ID:                  containerTaskType,
			RegisteredTaskTypes: []pluginsCore.TaskType{containerTaskType},
			ResourceToWatch:     &v1.Pod{},
			Plugin:              Plugin{},
			IsDefault:           true,
		})
}
