package logs

import (
	"context"
	"fmt"
	"strconv"
	"time"

	v1 "k8s.io/api/core/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/tasklog"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

const (
	kubernetesLogsDisplayName     = "Kubernetes Logs"
	cloudwatchLoggingDisplayName  = "Cloudwatch Logs"
	googleCloudLoggingDisplayName = "Google Cloud Logs"
	flyteEnableVscode             = "_F_E_VS"
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

	var containerID string
	if uint32(len(pod.Spec.Containers)) <= index {
		logger.Errorf(ctx, "container IndexOutOfBound, requested [%d], but total containers [%d] in pod phase [%v]", index, len(pod.Spec.Containers), pod.Status.Phase)
		return nil, nil
	} else {
		containerID = pod.Status.ContainerStatuses[index].ContainerID
	}

	if uint32(len(pod.Status.ContainerStatuses)) <= index {
		logger.Errorf(ctx, "containerStatus IndexOutOfBound, requested [%d], but total containerStatuses [%d] in pod phase [%v]", index, len(pod.Status.ContainerStatuses), pod.Status.Phase)
		return nil, nil
	}

	startTime := pod.CreationTimestamp.Unix()
	finishTime := time.Now().Unix()

	enableVscode := false
	for _, env := range pod.Spec.Containers[index].Env {
		if env.Name != flyteEnableVscode {
			continue
		}
		var err error
		enableVscode, err = strconv.ParseBool(env.Value)
		if err != nil {
			logger.Errorf(ctx, "failed to parse %s env var [%s] for pod [%s]", flyteEnableVscode, env.Value, pod.Name)
		}
	}

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
			EnableVscode:         enableVscode,
		},
	)

	if err != nil {
		return nil, err
	}

	return logs.TaskLogs, nil
}

type templateLogPluginCollection struct {
	plugins        []tasklog.Plugin
	dynamicPlugins []tasklog.Plugin
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
	var plugins []tasklog.Plugin
	var dynamicPlugins []tasklog.Plugin

	if cfg.IsKubernetesEnabled {
		if len(cfg.KubernetesTemplateURI) > 0 {
			plugins = append(plugins, tasklog.TemplateLogPlugin{DisplayName: kubernetesLogsDisplayName, TemplateURIs: []tasklog.TemplateURI{cfg.KubernetesTemplateURI}, MessageFormat: core.TaskLog_JSON})
		} else {
			plugins = append(plugins, tasklog.TemplateLogPlugin{DisplayName: kubernetesLogsDisplayName, TemplateURIs: []tasklog.TemplateURI{fmt.Sprintf("%s/#!/log/{{ .namespace }}/{{ .podName }}/pod?namespace={{ .namespace }}", cfg.KubernetesURL)}, MessageFormat: core.TaskLog_JSON})
		}
	}

	if cfg.IsCloudwatchEnabled {
		if len(cfg.CloudwatchTemplateURI) > 0 {
			plugins = append(plugins, tasklog.TemplateLogPlugin{DisplayName: cloudwatchLoggingDisplayName, TemplateURIs: []tasklog.TemplateURI{cfg.CloudwatchTemplateURI}, MessageFormat: core.TaskLog_JSON})
		} else {
			plugins = append(plugins, tasklog.TemplateLogPlugin{DisplayName: cloudwatchLoggingDisplayName, TemplateURIs: []tasklog.TemplateURI{fmt.Sprintf("https://console.aws.amazon.com/cloudwatch/home?region=%s#logEventViewer:group=%s;stream=var.log.containers.{{ .podName }}_{{ .namespace }}_{{ .containerName }}-{{ .containerId }}.log", cfg.CloudwatchRegion, cfg.CloudwatchLogGroup)}, MessageFormat: core.TaskLog_JSON})
		}
	}

	if cfg.IsStackDriverEnabled {
		if len(cfg.StackDriverTemplateURI) > 0 {
			plugins = append(plugins, tasklog.TemplateLogPlugin{DisplayName: googleCloudLoggingDisplayName, TemplateURIs: []tasklog.TemplateURI{cfg.StackDriverTemplateURI}, MessageFormat: core.TaskLog_JSON})
		} else {
			plugins = append(plugins, tasklog.TemplateLogPlugin{DisplayName: googleCloudLoggingDisplayName, TemplateURIs: []tasklog.TemplateURI{fmt.Sprintf("https://console.cloud.google.com/logs/viewer?project=%s&angularJsUrl=%%2Flogs%%2Fviewer%%3Fproject%%3D%s&resource=%s&advancedFilter=resource.labels.pod_name%%3D{{ .podName }}", cfg.GCPProjectName, cfg.GCPProjectName, cfg.StackdriverLogResourceName)}, MessageFormat: core.TaskLog_JSON})
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
				ShowWhilePending:    dynamicLogLink.ShowWhilePending,
				HideOnceFinished:    dynamicLogLink.HideOnceFinished,
			})
	}

	plugins = append(plugins, azureTemplatePluginsToPluginSlice(cfg.AzureLogTemplates)...)
	plugins = append(plugins, templatePluginToPluginSlice(cfg.Templates)...)
	return templateLogPluginCollection{plugins: plugins, dynamicPlugins: dynamicPlugins}, nil
}

func templatePluginToPluginSlice(templatePlugins []tasklog.TemplateLogPlugin) []tasklog.Plugin {
	plugins := make([]tasklog.Plugin, len(templatePlugins))
	for i := range templatePlugins {
		plugins[i] = &templatePlugins[i]
	}
	return plugins
}

func azureTemplatePluginsToPluginSlice(templatePlugins []tasklog.AzureLogsTemplatePlugin) []tasklog.Plugin {
	plugins := make([]tasklog.Plugin, len(templatePlugins))
	for i := range templatePlugins {
		plugins[i] = &templatePlugins[i]
	}
	return plugins
}
