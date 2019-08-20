package logs

import (
	"context"

	logUtils "github.com/lyft/flyteidl/clients/go/coreutils/logs"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytestdlib/logger"
	"k8s.io/api/core/v1"
)

func GetLogsForContainerInPod(ctx context.Context, pod *v1.Pod, index uint32, nameSuffix string) ([]*core.TaskLog, error) {
	var logs []*core.TaskLog
	logConfig := GetLogConfig()

	logPlugins := map[string]logUtils.LogPlugin{}

	if logConfig.IsKubernetesEnabled {
		logPlugins["Kubernetes Logs"] = logUtils.NewKubernetesLogPlugin(logConfig.KubernetesURL)
	}
	if logConfig.IsCloudwatchEnabled {
		logPlugins["Cloudwatch Logs"] = logUtils.NewCloudwatchLogPlugin(logConfig.CloudwatchRegion, logConfig.CloudwatchLogGroup)
	}
	if logConfig.IsStackDriverEnabled {
		logPlugins["Stackdriver Logs"] = logUtils.NewStackdriverLogPlugin(logConfig.GCPProjectName, logConfig.StackdriverLogResourceName)
	}

	if len(logPlugins) == 0 {
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

	for name, plugin := range logPlugins {
		log, err := plugin.GetTaskLog(
			pod.Name,
			pod.Namespace,
			pod.Spec.Containers[index].Name,
			pod.Status.ContainerStatuses[index].ContainerID,
			name+nameSuffix,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, &log)
	}
	return logs, nil
}
