package pod

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyteplugins/go/tasks/logs"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery"
	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/tasklog"

	v1 "k8s.io/api/core/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	podTaskType         = "pod"
	PrimaryContainerKey = "primary_container_name"
)

var (
	DefaultPodPlugin = plugin{
		defaultPodBuilder: containerPodBuilder{},
		podBuilders: map[string]podBuilder{
			SidecarTaskType: sidecarPodBuilder{},
		},
	}
)

type podBuilder interface {
	buildPodSpec(ctx context.Context, task *core.TaskTemplate, taskCtx pluginsCore.TaskExecutionContext) (*v1.PodSpec, error)
	getPrimaryContainerName(task *core.TaskTemplate, taskCtx pluginsCore.TaskExecutionContext) (string, error)
	updatePodMetadata(ctx context.Context, pod *v1.Pod, task *core.TaskTemplate, taskCtx pluginsCore.TaskExecutionContext) error
}

type plugin struct {
	defaultPodBuilder podBuilder
	podBuilders       map[string]podBuilder
}

func (plugin) BuildIdentityResource(_ context.Context, _ pluginsCore.TaskExecutionMetadata) (
	client.Object, error) {
	return flytek8s.BuildIdentityPod(), nil
}

func (p plugin) BuildResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (client.Object, error) {
	// read TaskTemplate
	task, err := taskCtx.TaskReader().Read(ctx)
	if err != nil {
		return nil, errors.Errorf(errors.BadTaskSpecification,
			"TaskSpecification cannot be read, Err: [%v]", err.Error())
	}

	// initialize PodBuilder
	builder, exists := p.podBuilders[task.Type]
	if !exists {
		builder = p.defaultPodBuilder
	}

	podSpec, err := builder.buildPodSpec(ctx, task, taskCtx)
	if err != nil {
		return nil, err
	}

	podSpec.ServiceAccountName = flytek8s.GetServiceAccountNameFromTaskExecutionMetadata(taskCtx.TaskExecutionMetadata())

	podTemplate := flytek8s.DefaultPodTemplateStore.LoadOrDefault(taskCtx.TaskExecutionMetadata().GetNamespace())
	primaryContainerName, err := builder.getPrimaryContainerName(task, taskCtx)
	if err != nil {
		return nil, err
	}

	pod, err := flytek8s.BuildPodWithSpec(podTemplate, podSpec, primaryContainerName)
	if err != nil {
		return nil, err
	}

	// update pod metadata
	if err = builder.updatePodMetadata(ctx, pod, task, taskCtx); err != nil {
		return nil, err
	}

	return pod, nil
}

func (p plugin) GetTaskPhase(ctx context.Context, pluginContext k8s.PluginContext, r client.Object) (pluginsCore.PhaseInfo, error) {
	logPlugin, err := logs.InitializeLogPlugins(logs.GetLogConfig())
	if err != nil {
		return pluginsCore.PhaseInfoUndefined, err
	}

	return p.GetTaskPhaseWithLogs(ctx, pluginContext, r, logPlugin, " (User)")
}

func (plugin) GetTaskPhaseWithLogs(ctx context.Context, pluginContext k8s.PluginContext, r client.Object, logPlugin tasklog.Plugin, logSuffix string) (pluginsCore.PhaseInfo, error) {
	pod := r.(*v1.Pod)

	transitionOccurredAt := flytek8s.GetLastTransitionOccurredAt(pod).Time
	info := pluginsCore.TaskInfo{
		OccurredAt: &transitionOccurredAt,
	}

	if pod.Status.Phase != v1.PodPending && pod.Status.Phase != v1.PodUnknown {
		taskLogs, err := logs.GetLogsForContainerInPod(ctx, logPlugin, pod, 0, logSuffix)
		if err != nil {
			return pluginsCore.PhaseInfoUndefined, err
		}
		info.Logs = taskLogs
	}

	switch pod.Status.Phase {
	case v1.PodSucceeded:
		return flytek8s.DemystifySuccess(pod.Status, info)
	case v1.PodFailed:
		return flytek8s.DemystifyFailure(pod.Status, info)
	case v1.PodPending:
		return flytek8s.DemystifyPending(pod.Status)
	case v1.PodReasonUnschedulable:
		return pluginsCore.PhaseInfoQueued(transitionOccurredAt, pluginsCore.DefaultPhaseVersion, "pod unschedulable"), nil
	case v1.PodUnknown:
		return pluginsCore.PhaseInfoUndefined, nil
	}

	primaryContainerName, exists := r.GetAnnotations()[PrimaryContainerKey]
	if !exists {
		// if the primary container annotation dos not exist, then the task requires all containers
		// to succeed to declare success. therefore, if the pod is not in one of the above states we
		// fallback to declaring the task as 'running'.
		if len(info.Logs) > 0 {
			return pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion+1, &info), nil
		}
		return pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion, &info), nil
	}

	// if the primary container annotation exists, we use the status of the specified container
	primaryContainerPhase := flytek8s.DeterminePrimaryContainerPhase(primaryContainerName, pod.Status.ContainerStatuses, &info)
	if primaryContainerPhase.Phase() == pluginsCore.PhaseRunning && len(info.Logs) > 0 {
		return pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion+1, primaryContainerPhase.Info()), nil
	}
	return primaryContainerPhase, nil
}

func (plugin) GetProperties() k8s.PluginProperties {
	return k8s.PluginProperties{}
}

func init() {
	// Register ContainerTaskType and SidecarTaskType plugin entries. These separate task types
	// still exist within the system, only now both are evaluated using the same internal pod plugin
	// instance. This simplifies migration as users may keep the same configuration but are
	// seamlessly transitioned from separate container and sidecar plugins to a single pod plugin.
	pluginmachinery.PluginRegistry().RegisterK8sPlugin(
		k8s.PluginEntry{
			ID:                  ContainerTaskType,
			RegisteredTaskTypes: []pluginsCore.TaskType{ContainerTaskType},
			ResourceToWatch:     &v1.Pod{},
			Plugin:              DefaultPodPlugin,
			IsDefault:           true,
		})

	pluginmachinery.PluginRegistry().RegisterK8sPlugin(
		k8s.PluginEntry{
			ID:                  SidecarTaskType,
			RegisteredTaskTypes: []pluginsCore.TaskType{SidecarTaskType},
			ResourceToWatch:     &v1.Pod{},
			Plugin:              DefaultPodPlugin,
			IsDefault:           false,
		})

	// register podTaskType plugin entry
	pluginmachinery.PluginRegistry().RegisterK8sPlugin(
		k8s.PluginEntry{
			ID:                  podTaskType,
			RegisteredTaskTypes: []pluginsCore.TaskType{ContainerTaskType, SidecarTaskType},
			ResourceToWatch:     &v1.Pod{},
			Plugin:              DefaultPodPlugin,
			IsDefault:           true,
		})
}
