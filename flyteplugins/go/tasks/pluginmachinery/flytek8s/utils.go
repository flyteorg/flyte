package flytek8s

import (
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	pluginmachinery_core "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
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
		case core.Resources_GPU:
			if !v.IsZero() {
				k8sResources[resourceGPU] = v
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

// getNormalizedAcceleratorDevice returns the normalized name for the given device.
// This should map to the node label that the corresponding nodes are provisioned with.
// Falls back to the original device name if the device is not configured.
func GetNormalizedAcceleratorDevice(device string) string {
	cfg := config.GetK8sPluginConfig()
	if normalized, ok := cfg.AcceleratorDevices[strings.ToUpper(device)]; ok {
		return normalized
	}
	return device
}

// getAcceleratorResourceName returns the Kubernetes resource name for the given device class.
// Falls back to the legacy GpuResourceName if the device class is not configured.
func getAcceleratorResourceName(accelerator *core.GPUAccelerator) v1.ResourceName {
	// Use the shared helper function to get the accelerator config
	accelConfig := getAcceleratorConfig(accelerator)
	return accelConfig.ResourceName
}

// getAllAcceleratorResourceNames returns the Kubernetes resource names for all accelerator devices.
func getAllAcceleratorResourceNames() map[v1.ResourceName]struct{} {
	cfg := config.GetK8sPluginConfig()
	acceleratorResourceNames := make(map[v1.ResourceName]struct{})

	// Add the legacy GPU resource name for backward compatibility
	acceleratorResourceNames[cfg.GpuResourceName] = struct{}{}

	// Add resource names from all configured accelerator device classes
	for _, deviceClassConfig := range cfg.AcceleratorDeviceClasses {
		if deviceClassConfig.ResourceName != "" {
			acceleratorResourceNames[deviceClassConfig.ResourceName] = struct{}{}
		}
	}
	return acceleratorResourceNames
}

// podRequiresAccelerator returns true if any container in the pod requires any accelerator devices.
func podRequiresAccelerator(podSpec *v1.PodSpec) bool {
	acceleratorResourceNames := getAllAcceleratorResourceNames()
	for _, cnt := range podSpec.Containers {
		for resourceName := range acceleratorResourceNames {
			if _, ok := cnt.Resources.Limits[resourceName]; ok {
				return true
			}
		}
	}
	return false
}

// getConfiguredDeviceClasses returns a list of configured device class names for logging purposes.
func getConfiguredDeviceClasses(deviceClasses map[string]config.AcceleratorDeviceClassConfig) []string {
	classes := make([]string, 0, len(deviceClasses))
	for k := range deviceClasses {
		classes = append(classes, k)
	}
	return classes
}
