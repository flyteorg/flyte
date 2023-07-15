package pod

import (
	"context"

	pluginserrors "github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyteplugins/go/tasks/logs"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery"
	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/tasklog"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"

	"github.com/flyteorg/flytestdlib/logger"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ContainerTaskType    = "container"
	podTaskType          = "pod"
	pythonTaskType       = "python-task"
	rawContainerTaskType = "raw-container"
	SidecarTaskType      = "sidecar"
)

// Why, you might wonder do we recreate the generated go struct generated from the plugins.SidecarJob proto? Because
// although we unmarshal the task custom json, the PodSpec itself is not generated from a proto definition,
// but a proper go struct defined in k8s libraries. Therefore we only unmarshal the sidecar as a json, rather than jsonpb.
type sidecarJob struct {
	PodSpec              *v1.PodSpec
	PrimaryContainerName string
	Annotations          map[string]string
	Labels               map[string]string
}

var DefaultPodPlugin = plugin{}

type plugin struct {
}

func (plugin) BuildIdentityResource(_ context.Context, _ pluginsCore.TaskExecutionMetadata) (client.Object, error) {
	return flytek8s.BuildIdentityPod(), nil
}

func (p plugin) BuildResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (client.Object, error) {
	taskTemplate, err := taskCtx.TaskReader().Read(ctx)
	if err != nil {
		logger.Warnf(ctx, "failed to read task information when trying to construct Pod, err: %s", err.Error())
		return nil, err
	}

	var podSpec *v1.PodSpec
	objectMeta := &metav1.ObjectMeta{
		Annotations: make(map[string]string),
		Labels:      make(map[string]string),
	}
	primaryContainerName := ""

	if taskTemplate.Type == SidecarTaskType && taskTemplate.TaskTypeVersion == 0 {
		// handles pod tasks when they are defined as Sidecar tasks and marshal the podspec using k8s proto.
		sidecarJob := sidecarJob{}
		err := utils.UnmarshalStructToObj(taskTemplate.GetCustom(), &sidecarJob)
		if err != nil {
			return nil, pluginserrors.Errorf(pluginserrors.BadTaskSpecification, "invalid TaskSpecification [%v], Err: [%v]", taskTemplate.GetCustom(), err.Error())
		}

		if sidecarJob.PodSpec == nil {
			return nil, pluginserrors.Errorf(pluginserrors.BadTaskSpecification, "invalid TaskSpecification, nil PodSpec [%v]", taskTemplate.GetCustom())
		}

		podSpec = sidecarJob.PodSpec

		// get primary container name
		primaryContainerName = sidecarJob.PrimaryContainerName

		// update annotations and labels
		objectMeta.Annotations = utils.UnionMaps(objectMeta.Annotations, sidecarJob.Annotations)
		objectMeta.Labels = utils.UnionMaps(objectMeta.Labels, sidecarJob.Labels)
	} else if taskTemplate.Type == SidecarTaskType && taskTemplate.TaskTypeVersion == 1 {
		// handles pod tasks that marshal the pod spec to the task custom.
		err := utils.UnmarshalStructToObj(taskTemplate.GetCustom(), &podSpec)
		if err != nil {
			return nil, pluginserrors.Errorf(pluginserrors.BadTaskSpecification,
				"Unable to unmarshal task custom [%v], Err: [%v]", taskTemplate.GetCustom(), err.Error())
		}

		// get primary container name
		if len(taskTemplate.GetConfig()) == 0 {
			return nil, pluginserrors.Errorf(pluginserrors.BadTaskSpecification,
				"invalid TaskSpecification, config needs to be non-empty and include missing [%s] key", flytek8s.PrimaryContainerKey)
		}

		var ok bool
		if primaryContainerName, ok = taskTemplate.GetConfig()[flytek8s.PrimaryContainerKey]; !ok {
			return nil, pluginserrors.Errorf(pluginserrors.BadTaskSpecification,
				"invalid TaskSpecification, config missing [%s] key in [%v]", flytek8s.PrimaryContainerKey, taskTemplate.GetConfig())
		}

		// update annotations and labels
		if taskTemplate.GetK8SPod() != nil && taskTemplate.GetK8SPod().Metadata != nil {
			objectMeta.Annotations = utils.UnionMaps(objectMeta.Annotations, taskTemplate.GetK8SPod().Metadata.Annotations)
			objectMeta.Labels = utils.UnionMaps(objectMeta.Labels, taskTemplate.GetK8SPod().Metadata.Labels)
		}
	} else {
		// handles both container / pod tasks that use the TaskTemplate Container and K8sPod fields
		var err error
		podSpec, objectMeta, primaryContainerName, err = flytek8s.BuildRawPod(ctx, taskCtx)
		if err != nil {
			return nil, err
		}
	}

	// update podSpec and objectMeta with Flyte customizations
	podSpec, objectMeta, err = flytek8s.ApplyFlytePodConfiguration(ctx, taskCtx, podSpec, objectMeta, primaryContainerName)
	if err != nil {
		return nil, err
	}

	// set primaryContainerKey annotation if this is a Sidecar task or, as an optimization, if there is only a single
	// container. this plugin marks the task complete if the primary Container is complete, so if there is only one
	// container we can mark the task as complete before the Pod has been marked complete.
	if taskTemplate.Type == SidecarTaskType || len(podSpec.Containers) == 1 {
		objectMeta.Annotations[flytek8s.PrimaryContainerKey] = primaryContainerName
	}

	podSpec.ServiceAccountName = flytek8s.GetServiceAccountNameFromTaskExecutionMetadata(taskCtx.TaskExecutionMetadata())

	pod := flytek8s.BuildIdentityPod()
	pod.ObjectMeta = *objectMeta
	pod.Spec = *podSpec

	return pod, nil
}

func (p plugin) GetTaskPhase(ctx context.Context, pluginContext k8s.PluginContext, r client.Object) (pluginsCore.PhaseInfo, error) {
	logPlugin, err := logs.InitializeLogPlugins(logs.GetLogConfig())
	if err != nil {
		return pluginsCore.PhaseInfoUndefined, err
	}

	return p.GetTaskPhaseWithLogs(ctx, pluginContext, r, logPlugin, " (User)", nil)
}

func (plugin) GetTaskPhaseWithLogs(ctx context.Context, pluginContext k8s.PluginContext, r client.Object, logPlugin tasklog.Plugin, logSuffix string, extraLogTemplateVarsByScheme *tasklog.TemplateVarsByScheme) (pluginsCore.PhaseInfo, error) {
	pluginState := k8s.PluginState{}
	_, err := pluginContext.PluginStateReader().Get(&pluginState)
	if err != nil {
		return pluginsCore.PhaseInfoUndefined, err
	}

	pod := r.(*v1.Pod)

	transitionOccurredAt := flytek8s.GetLastTransitionOccurredAt(pod).Time
	reportedAt := flytek8s.GetReportedAt(pod).Time
	if reportedAt.IsZero() {
		reportedAt = transitionOccurredAt
	}

	info := pluginsCore.TaskInfo{
		OccurredAt: &transitionOccurredAt,
		ReportedAt: &reportedAt,
	}

	taskExecID := pluginContext.TaskExecutionMetadata().GetTaskExecutionID().GetID()
	if pod.Status.Phase != v1.PodPending && pod.Status.Phase != v1.PodUnknown {
		taskLogs, err := logs.GetLogsForContainerInPod(ctx, logPlugin, &taskExecID, pod, 0, logSuffix, extraLogTemplateVarsByScheme)
		if err != nil {
			return pluginsCore.PhaseInfoUndefined, err
		}
		info.Logs = taskLogs
	}

	phaseInfo := pluginsCore.PhaseInfoUndefined
	switch pod.Status.Phase {
	case v1.PodSucceeded:
		phaseInfo, err = flytek8s.DemystifySuccess(pod.Status, info)
	case v1.PodFailed:
		phaseInfo, err = flytek8s.DemystifyFailure(pod.Status, info)
	case v1.PodPending:
		phaseInfo, err = flytek8s.DemystifyPending(pod.Status)
	case v1.PodReasonUnschedulable:
		phaseInfo = pluginsCore.PhaseInfoQueued(transitionOccurredAt, pluginsCore.DefaultPhaseVersion, "pod unschedulable")
	case v1.PodUnknown:
		// DO NOTHING
	default:
		primaryContainerName, exists := r.GetAnnotations()[flytek8s.PrimaryContainerKey]
		if !exists {
			// if all of the containers in the Pod are complete, as an optimization, we can declare the task as
			// succeeded rather than waiting for the Pod to be marked completed.
			allSuccessfullyTerminated := len(pod.Status.ContainerStatuses) > 0
			for _, s := range pod.Status.ContainerStatuses {
				if s.State.Waiting != nil || s.State.Running != nil || (s.State.Terminated != nil && s.State.Terminated.ExitCode != 0) {
					allSuccessfullyTerminated = false
				}
			}

			if allSuccessfullyTerminated {
				return flytek8s.DemystifySuccess(pod.Status, info)
			}

			// if the primary container annotation does not exist, then the task requires all containers
			// to succeed to declare success. therefore, if the pod is not in one of the above states we
			// fallback to declaring the task as 'running'.
			phaseInfo = pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion, &info)
			if len(info.Logs) > 0 {
				phaseInfo = phaseInfo.WithVersion(pluginsCore.DefaultPhaseVersion + 1)
			}
		} else {
			// if the primary container annotation exists, we use the status of the specified container
			phaseInfo = flytek8s.DeterminePrimaryContainerPhase(primaryContainerName, pod.Status.ContainerStatuses, &info)
			if phaseInfo.Phase() == pluginsCore.PhaseRunning && len(info.Logs) > 0 {
				phaseInfo = phaseInfo.WithVersion(pluginsCore.DefaultPhaseVersion + 1)
			}
		}
	}

	if err != nil {
		return pluginsCore.PhaseInfoUndefined, err
	} else if phaseInfo.Phase() != pluginsCore.PhaseRunning && phaseInfo.Phase() == pluginState.Phase &&
		phaseInfo.Version() <= pluginState.PhaseVersion && phaseInfo.Reason() != pluginState.Reason {

		// if we have the same Phase as the previous evaluation and updated the Reason but not the PhaseVersion we must
		// update the PhaseVersion so an event is sent to reflect the Reason update. this does not handle the Running
		// Phase because the legacy used `DefaultPhaseVersion + 1` which will only increment to 1.
		phaseInfo = phaseInfo.WithVersion(pluginState.PhaseVersion + 1)
	}

	return phaseInfo, err
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
			RegisteredTaskTypes: []pluginsCore.TaskType{ContainerTaskType, pythonTaskType, rawContainerTaskType},
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
			RegisteredTaskTypes: []pluginsCore.TaskType{ContainerTaskType, pythonTaskType, rawContainerTaskType, SidecarTaskType},
			ResourceToWatch:     &v1.Pod{},
			Plugin:              DefaultPodPlugin,
			IsDefault:           true,
		})
}
