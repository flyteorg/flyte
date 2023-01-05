package flytek8s

import (
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	pluginmachinery_core "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func ToK8sEnvVar(env []*core.KeyValuePair) []v1.EnvVar {
	envVars := make([]v1.EnvVar, 0, len(env))
	for _, kv := range env {
		envVars = append(envVars, v1.EnvVar{Name: kv.Key, Value: kv.Value})
	}
	return envVars
}

// TODO we should modify the container resources to contain a map of enum values?
// Also we should probably create tolerations / taints, but we could do that as a post process
func ToK8sResourceList(resources []*core.Resources_ResourceEntry) (v1.ResourceList, error) {
	k8sResources := make(v1.ResourceList, len(resources))
	for _, r := range resources {
		rVal := r.Value
		v, err := resource.ParseQuantity(rVal)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to parse resource as a valid quantity.")
		}
		switch r.Name {
		case core.Resources_CPU:
			if !v.IsZero() {
				k8sResources[v1.ResourceCPU] = v
			}
		case core.Resources_MEMORY:
			if !v.IsZero() {
				k8sResources[v1.ResourceMemory] = v
			}
		case core.Resources_STORAGE:
			if !v.IsZero() {
				k8sResources[v1.ResourceStorage] = v
			}
		case core.Resources_GPU:
			if !v.IsZero() {
				k8sResources[ResourceNvidiaGPU] = v
			}
		case core.Resources_EPHEMERAL_STORAGE:
			if !v.IsZero() {
				k8sResources[v1.ResourceEphemeralStorage] = v
			}
		}
	}
	return k8sResources, nil
}

func ToK8sResourceRequirements(resources *core.Resources) (*v1.ResourceRequirements, error) {
	res := &v1.ResourceRequirements{}
	if resources == nil {
		return res, nil
	}
	req, err := ToK8sResourceList(resources.Requests)
	if err != nil {
		return res, err
	}
	lim, err := ToK8sResourceList(resources.Limits)
	if err != nil {
		return res, err
	}
	res.Limits = lim
	res.Requests = req
	return res, nil
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
