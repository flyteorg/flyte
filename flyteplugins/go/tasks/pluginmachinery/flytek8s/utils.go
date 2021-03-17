package flytek8s

import (
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	pluginmachinery_core "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	v1 "k8s.io/api/core/v1"
)

func ToK8sEnvVar(env []*core.KeyValuePair) []v1.EnvVar {
	envVars := make([]v1.EnvVar, 0, len(env))
	for _, kv := range env {
		envVars = append(envVars, v1.EnvVar{Name: kv.Key, Value: kv.Value})
	}
	return envVars
}

func GetServiceAccountNameFromTaskExecutionMetadata(taskExecutionMetadata pluginmachinery_core.TaskExecutionMetadata) string {
	var serviceAccount string
	securityContext := taskExecutionMetadata.GetSecurityContext()
	if securityContext.GetRunAs() != nil {
		serviceAccount = securityContext.GetRunAs().GetK8SServiceAccount()
	}

	// TO BE DEPRECATED
	if len(serviceAccount) == 0 {
		serviceAccount = taskExecutionMetadata.GetK8sServiceAccount()
	}

	return serviceAccount
}
