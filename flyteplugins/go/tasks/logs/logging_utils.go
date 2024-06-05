package logs

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/tasklog"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

// Internal
func GetLogsForContainerInPod(ctx context.Context, logPlugin tasklog.Plugin, taskExecID pluginsCore.TaskExecutionID, pod *v1.Pod, index uint32, nameSuffix string, extraLogTemplateVars []tasklog.TemplateVar, taskTemplate *core.TaskTemplate) ([]*core.TaskLog, error) {
	if logPlugin == nil {
		return nil, nil
	}

	if pod == nil {
		logger.Error(ctx, "cannot extract logs for a nil container")
		return nil, nil
	}

	if uint32(len(pod.Spec.Containers)) <= index {
		logger.Errorf(ctx, "container IndexOutOfBound, requested [%d], but total containers [%d] in pod phase [%v]", index, len(pod.Spec.Containers), pod.Status.Phase)
		return nil, nil
	}

	containerID := v1.ContainerStatus{}.ContainerID
	if uint32(len(pod.Status.ContainerStatuses)) <= index {
		logger.Errorf(ctx, "containerStatus IndexOutOfBound, requested [%d], but total containerStatuses [%d] in pod phase [%v]", index, len(pod.Status.ContainerStatuses), pod.Status.Phase)
	} else {
		containerID = pod.Status.ContainerStatuses[index].ContainerID
	}

	startTime := pod.CreationTimestamp.Unix()
	finishTime := time.Now().Unix()

	logs, err := logPlugin.GetTaskLogs(
		tasklog.Input{
			PodName:              pod.Name,
			PodUID:               string(pod.GetUID()),
			Namespace:            pod.Namespace,
			ContainerName:        pod.Spec.Containers[index].Name,
			ContainerID:          containerID,
			LogName:              nameSuffix,
			PodRFC3339StartTime:  time.Unix(startTime, 0).Format(time.RFC3339),
			PodRFC3339FinishTime: time.Unix(finishTime, 0).Format(time.RFC3339),
			PodUnixStartTime:     startTime,
			PodUnixFinishTime:    finishTime,
			TaskExecutionID:      taskExecID,
			ExtraTemplateVars:    extraLogTemplateVars,
			TaskTemplate:         taskTemplate,
			HostName:             pod.Spec.Hostname,
		},
	)

	if err != nil {
		return nil, err
	}

	return logs.TaskLogs, nil
}

type templateLogPluginCollection struct {
	plugins        []tasklog.TemplateLogPlugin
	dynamicPlugins []tasklog.TemplateLogPlugin
}

func (t templateLogPluginCollection) GetTaskLogs(input tasklog.Input) (tasklog.Output, error) {
	var taskLogs []*core.TaskLog

	for _, plugin := range append(t.plugins, t.dynamicPlugins...) {
		o, err := plugin.GetTaskLogs(input)
		if err != nil {
			return tasklog.Output{}, err
		}
		taskLogs = append(taskLogs, o.TaskLogs...)
	}

	return tasklog.Output{TaskLogs: taskLogs}, nil
}

// InitializeLogPlugins initializes log plugin based on config.
func InitializeLogPlugins(cfg *LogConfig) (tasklog.Plugin, error) {
	// Use a list to maintain order.
	var plugins []tasklog.TemplateLogPlugin
	var dynamicPlugins []tasklog.TemplateLogPlugin

	if cfg.IsKubernetesEnabled {
		if len(cfg.KubernetesTemplateURI) > 0 {
			plugins = append(plugins, tasklog.TemplateLogPlugin{DisplayName: "Kubernetes Logs", TemplateURIs: []tasklog.TemplateURI{cfg.KubernetesTemplateURI}, MessageFormat: core.TaskLog_JSON})
		} else {
			plugins = append(plugins, tasklog.TemplateLogPlugin{DisplayName: "Kubernetes Logs", TemplateURIs: []tasklog.TemplateURI{fmt.Sprintf("%s/#!/log/{{ .namespace }}/{{ .podName }}/pod?namespace={{ .namespace }}", cfg.KubernetesURL)}, MessageFormat: core.TaskLog_JSON})
		}
	}

	if cfg.IsCloudwatchEnabled {
		if len(cfg.CloudwatchTemplateURI) > 0 {
			plugins = append(plugins, tasklog.TemplateLogPlugin{DisplayName: "Cloudwatch Logs", TemplateURIs: []tasklog.TemplateURI{cfg.CloudwatchTemplateURI}, MessageFormat: core.TaskLog_JSON})
		} else {
			plugins = append(plugins, tasklog.TemplateLogPlugin{DisplayName: "Cloudwatch Logs", TemplateURIs: []tasklog.TemplateURI{fmt.Sprintf("https://console.aws.amazon.com/cloudwatch/home?region=%s#logEventViewer:group=%s;stream=var.log.containers.{{ .podName }}_{{ .namespace }}_{{ .containerName }}-{{ .containerId }}.log", cfg.CloudwatchRegion, cfg.CloudwatchLogGroup)}, MessageFormat: core.TaskLog_JSON})
		}
	}

	if cfg.IsStackDriverEnabled {
		if len(cfg.StackDriverTemplateURI) > 0 {
			plugins = append(plugins, tasklog.TemplateLogPlugin{DisplayName: "Stackdriver Logs", TemplateURIs: []tasklog.TemplateURI{cfg.StackDriverTemplateURI}, MessageFormat: core.TaskLog_JSON})
		} else {
			plugins = append(plugins, tasklog.TemplateLogPlugin{DisplayName: "Stackdriver Logs", TemplateURIs: []tasklog.TemplateURI{fmt.Sprintf("https://console.cloud.google.com/logs/viewer?project=%s&angularJsUrl=%%2Flogs%%2Fviewer%%3Fproject%%3D%s&resource=%s&advancedFilter=resource.labels.pod_name%%3D{{ .podName }}", cfg.GCPProjectName, cfg.GCPProjectName, cfg.StackdriverLogResourceName)}, MessageFormat: core.TaskLog_JSON})
		}
	}

	for logLinkType, dynamicLogLink := range cfg.DynamicLogLinks {
		dynamicPlugins = append(
			dynamicPlugins,
			tasklog.TemplateLogPlugin{
				Name:                logLinkType,
				DisplayName:         dynamicLogLink.DisplayName,
				DynamicTemplateURIs: dynamicLogLink.TemplateURIs,
				MessageFormat:       core.TaskLog_JSON,
			})
	}

	plugins = append(plugins, cfg.Templates...)
	return templateLogPluginCollection{plugins: plugins, dynamicPlugins: dynamicPlugins}, nil
}
