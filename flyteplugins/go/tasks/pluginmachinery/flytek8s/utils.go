package flytek8s

import (
	"math"
	"strconv"

	"github.com/pkg/errors" // For decimal arithmetic
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
func ToK8sResourceList(resources []*core.Resources_ResourceEntry, onOOMConfig pluginmachinery_core.OnOOMConfig) (v1.ResourceList, error) {
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
				if onOOMConfig != nil && onOOMConfig.GetExponent() > 0 && onOOMConfig.GetFactor() > 1.0 {
					// Convert to multiplier into string and parse it to quantity
					multiplier := math.Pow(float64(onOOMConfig.GetFactor()), float64(onOOMConfig.GetExponent()))
					multiplier_str := strconv.FormatFloat(multiplier, 'f', -1, 64)
					multiplier_quantity, err := resource.ParseQuantity(multiplier_str)
					if err != nil {
						return nil, errors.Wrap(err, "Failed to parse resource as a valid quantity.")
					}

					// Multiply the value by the multiplier
					// Note: this does not update the string in resource.Quantity
					v.AsDec().Mul(v.AsDec(), multiplier_quantity.AsDec())

					// Convert the new value to a string and parse it to quantity
					// Create a new resource.Quantity to have matching string and numerical value
					v, err = resource.ParseQuantity(v.AsDec().String())
					if err != nil {
						return nil, errors.Wrap(err, "Failed to parse resource as a valid quantity.")
					}

					limitVal, err := resource.ParseQuantity(onOOMConfig.GetLimit())
					if err != nil {
						return nil, errors.Wrap(err, "Failed to parse resource as a valid quantity.")
					}
					if !limitVal.IsZero() && v.Cmp(limitVal) > 0 {
						v = limitVal
					}
				}
				k8sResources[v1.ResourceMemory] = v
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

func ToK8sResourceRequirements(resources *core.Resources, onOOMConfig pluginmachinery_core.OnOOMConfig) (*v1.ResourceRequirements, error) {
	res := &v1.ResourceRequirements{}
	if resources == nil {
		return res, nil
	}
	req, err := ToK8sResourceList(resources.GetRequests(), onOOMConfig)
	if err != nil {
		return res, err
	}
	lim, err := ToK8sResourceList(resources.GetLimits(), onOOMConfig)
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
