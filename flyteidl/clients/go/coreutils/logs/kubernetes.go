package logs

import (
	"fmt"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

type kubernetesLogPlugin struct {
	k8sURL string
}

func (s kubernetesLogPlugin) GetTaskLog(podName, namespace, containerName, containerID, logName string) (core.TaskLog, error) {
	return core.TaskLog{
		Uri:           fmt.Sprintf("%s/#!/log/%s/%s/pod?namespace=%s", s.k8sURL, namespace, podName, namespace),
		Name:          logName,
		MessageFormat: core.TaskLog_UNKNOWN,
	}, nil
}

// Deprecated: Please use NewTemplateLogPlugin from github.com/lyft/flyteplugins/go/tasks/pluginmachinery/tasklog instead.
func NewKubernetesLogPlugin(k8sURL string) LogPlugin {
	return &kubernetesLogPlugin{
		k8sURL: k8sURL,
	}
}
