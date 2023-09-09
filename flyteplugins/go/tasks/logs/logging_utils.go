package logs

import (
	"context"
	"fmt"
	"time"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/tasklog"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
	v1 "k8s.io/api/core/v1"
)

type logPlugin struct {
	Name   string
	Plugin tasklog.Plugin
}

// Internal
func GetLogsForContainerInPod(ctx context.Context, logPlugin tasklog.Plugin, taskExecID *core.TaskExecutionIdentifier, pod *v1.Pod, index uint32, nameSuffix string, extraLogTemplateVarsByScheme *tasklog.TemplateVarsByScheme) ([]*core.TaskLog, error) {
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

	if uint32(len(pod.Status.ContainerStatuses)) <= index {
		logger.Errorf(ctx, "containerStatus IndexOutOfBound, requested [%d], but total containerStatuses [%d] in pod phase [%v]", index, len(pod.Status.ContainerStatuses), pod.Status.Phase)
		return nil, nil
	}

	startTime := pod.CreationTimestamp.Unix()
	finishTime := time.Now().Unix()

	logs, err := logPlugin.GetTaskLogs(
		tasklog.Input{
			PodName:                   pod.Name,
			PodUID:                    string(pod.GetUID()),
			Namespace:                 pod.Namespace,
			ContainerName:             pod.Spec.Containers[index].Name,
			ContainerID:               pod.Status.ContainerStatuses[index].ContainerID,
			LogName:                   nameSuffix,
			PodRFC3339StartTime:       time.Unix(startTime, 0).Format(time.RFC3339),
			PodRFC3339FinishTime:      time.Unix(finishTime, 0).Format(time.RFC3339),
			PodUnixStartTime:          startTime,
			PodUnixFinishTime:         finishTime,
			TaskExecutionIdentifier:   taskExecID,
			ExtraTemplateVarsByScheme: extraLogTemplateVarsByScheme,
		},
	)

	if err != nil {
		return nil, err
	}

	return logs.TaskLogs, nil
}

type taskLogPluginWrapper struct {
	logPlugins []logPlugin
}

func (t taskLogPluginWrapper) GetTaskLogs(input tasklog.Input) (logOutput tasklog.Output, err error) {
	logs := make([]*core.TaskLog, 0, len(t.logPlugins))
	suffix := input.LogName

	for _, plugin := range t.logPlugins {
		input.LogName = plugin.Name + suffix
		o, err := plugin.Plugin.GetTaskLogs(input)
		if err != nil {
			return tasklog.Output{}, err
		}

		logs = append(logs, o.TaskLogs...)
	}

	return tasklog.Output{
		TaskLogs: logs,
	}, nil
}

// InitializeLogPlugins initializes log plugin based on config.
func InitializeLogPlugins(cfg *LogConfig) (tasklog.Plugin, error) {
	// Use a list to maintain order.
	logPlugins := make([]logPlugin, 0, 2)

	if cfg.IsKubernetesEnabled {
		if len(cfg.KubernetesTemplateURI) > 0 {
			logPlugins = append(logPlugins, logPlugin{Name: "Kubernetes Logs", Plugin: tasklog.NewTemplateLogPlugin(tasklog.TemplateSchemePod, []string{cfg.KubernetesTemplateURI}, core.TaskLog_JSON)})
		} else {
			logPlugins = append(logPlugins, logPlugin{Name: "Kubernetes Logs", Plugin: tasklog.NewTemplateLogPlugin(tasklog.TemplateSchemePod, []string{fmt.Sprintf("%s/#!/log/{{ .namespace }}/{{ .podName }}/pod?namespace={{ .namespace }}", cfg.KubernetesURL)}, core.TaskLog_JSON)})
		}
	}

	if cfg.IsCloudwatchEnabled {
		if len(cfg.CloudwatchTemplateURI) > 0 {
			logPlugins = append(logPlugins, logPlugin{Name: "Cloudwatch Logs", Plugin: tasklog.NewTemplateLogPlugin(tasklog.TemplateSchemePod, []string{cfg.CloudwatchTemplateURI}, core.TaskLog_JSON)})
		} else {
			logPlugins = append(logPlugins, logPlugin{Name: "Cloudwatch Logs", Plugin: tasklog.NewTemplateLogPlugin(tasklog.TemplateSchemePod, []string{fmt.Sprintf("https://console.aws.amazon.com/cloudwatch/home?region=%s#logEventViewer:group=%s;stream=var.log.containers.{{ .podName }}_{{ .namespace }}_{{ .containerName }}-{{ .containerId }}.log", cfg.CloudwatchRegion, cfg.CloudwatchLogGroup)}, core.TaskLog_JSON)})
		}
	}

	if cfg.IsStackDriverEnabled {
		if len(cfg.StackDriverTemplateURI) > 0 {
			logPlugins = append(logPlugins, logPlugin{Name: "Stackdriver Logs", Plugin: tasklog.NewTemplateLogPlugin(tasklog.TemplateSchemePod, []string{cfg.StackDriverTemplateURI}, core.TaskLog_JSON)})
		} else {
			logPlugins = append(logPlugins, logPlugin{Name: "Stackdriver Logs", Plugin: tasklog.NewTemplateLogPlugin(tasklog.TemplateSchemePod, []string{fmt.Sprintf("https://console.cloud.google.com/logs/viewer?project=%s&angularJsUrl=%%2Flogs%%2Fviewer%%3Fproject%%3D%s&resource=%s&advancedFilter=resource.labels.pod_name%%3D{{ .podName }}", cfg.GCPProjectName, cfg.GCPProjectName, cfg.StackdriverLogResourceName)}, core.TaskLog_JSON)})
		}
	}

	if len(cfg.Templates) > 0 {
		for _, cfg := range cfg.Templates {
			logPlugins = append(logPlugins, logPlugin{Name: cfg.DisplayName, Plugin: tasklog.NewTemplateLogPlugin(cfg.Scheme, cfg.TemplateURIs, cfg.MessageFormat)})
		}
	}

	if len(logPlugins) == 0 {
		return nil, nil
	}

	return taskLogPluginWrapper{logPlugins: logPlugins}, nil
}
