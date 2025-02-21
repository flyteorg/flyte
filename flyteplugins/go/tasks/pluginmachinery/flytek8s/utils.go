package flytek8s

import (
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	pluginmachinery_core "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
)

func ToK8sEnvVar(env []*core.KeyValuePair) []v1.EnvVar {
	envVars := make([]v1.EnvVar, 0, len(env))
	for _, kv := range env {
		envVars = append(envVars, v1.EnvVar{Name: kv.GetKey(), Value: kv.GetValue()})
	}
	return envVars
}

// TODO we should modify the container resources to contain a map of enum values?
// Also we should probably create tolerations / taints, but we could do that as a post process
func ToK8sResourceList(resources []*core.Resources_ResourceEntry, OOMCount uint32) (v1.ResourceList, error) {
	k8sResources := make(v1.ResourceList, len(resources))
	for _, r := range resources {
		rVal := r.GetValue()
		v, err := resource.ParseQuantity(rVal)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to parse resource as a valid quantity.")
		}
		switch r.GetName() {
		case core.Resources_CPU:
			if !v.IsZero() {
				k8sResources[v1.ResourceCPU] = v
			}
		case core.Resources_MEMORY:
			if !v.IsZero() {
				memQuantity := k8sResources[v1.ResourceMemory]
				memQuantity.Add(v)
				k8sResources[v1.ResourceMemory] = memQuantity
			}
		case core.Resources_GPU:
			if !v.IsZero() {
				k8sResources[ResourceNvidiaGPU] = v
			}
		case core.Resources_EPHEMERAL_STORAGE:
			if !v.IsZero() {
				k8sResources[v1.ResourceEphemeralStorage] = v
			}
		case core.Resources_OOM_RESERVED_MEMORY:
			if !v.IsZero() {
				memQuantity := k8sResources[v1.ResourceMemory]
				for i := uint32(0); i < OOMCount; i++ {
					memQuantity.Add(v)
				}
				k8sResources[v1.ResourceMemory] = memQuantity
			}
		}
	}
	return k8sResources, nil
}

func ToK8sResourceRequirements(resources *core.Resources, OOMCount uint32) (*v1.ResourceRequirements, error) {
	// TODO: HERE!!!!!!!!!
	res := &v1.ResourceRequirements{}
	if resources == nil {
		return res, nil
	}
	req, err := ToK8sResourceList(resources.GetRequests(), OOMCount)
	if err != nil {
		return res, err
	}
	lim, err := ToK8sResourceList(resources.GetLimits(), OOMCount)
	if err != nil {
		return res, err
	}
	res.Limits = lim
	res.Requests = req
	return res, nil
}

// ApplyK8sResourceOverrides ensures that both resource requests and limits are set.
// This is required because we run user executions in namespaces bound with a project quota and the Kubernetes scheduler will reject requests omitting these.
// This function is called by plugins that don't necessarily construct a default flyte container (container and k8s pod tasks)
// and therefore don't already receive the ApplyResourceOverrides treatment and subsequent validation which handles adding sensible defaults for requests and limits.
func ApplyK8sResourceOverrides(teMetadata pluginmachinery_core.TaskExecutionMetadata, resources *v1.ResourceRequirements) v1.ResourceRequirements {
	platformResources := teMetadata.GetPlatformResources()
	if platformResources == nil {
		platformResources = &v1.ResourceRequirements{}
	}

	return ApplyResourceOverrides(*resources, *platformResources, assignIfUnset)
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
