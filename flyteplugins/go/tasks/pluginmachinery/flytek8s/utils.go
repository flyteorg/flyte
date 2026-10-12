package flytek8s

import (
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	pluginmachinery_core "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
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

// partitionPlaceholder is replaced with the task's partition size when rendering a device
// class's PartitionResourceName template.
const partitionPlaceholder = "{partition}"

// getAcceleratorResourceName returns the Kubernetes resource name for the given accelerator.
// A partitioned accelerator whose device class sets PartitionResourceName gets the rendered
// per-partition name. Otherwise this is the device class's ResourceName, falling back to the
// legacy GpuResourceName if the device class is not configured.
func getAcceleratorResourceName(accelerator *core.GPUAccelerator) v1.ResourceName {
	// Use the shared helper function to get the accelerator config
	accelConfig := getAcceleratorConfig(accelerator)
	if name, ok := partitionResourceName(accelConfig, accelerator); ok {
		return name
	}
	return accelConfig.ResourceName
}

// partitionResourceName renders the device class's PartitionResourceName template for the
// accelerator's partition size. It returns false when the template is unset or the accelerator
// does not ask for a partition size, in which case callers use ResourceName as before.
func partitionResourceName(accelConfig config.AcceleratorDeviceClassConfig, accelerator *core.GPUAccelerator) (v1.ResourceName, bool) {
	partition := accelerator.GetPartitionSize()
	if accelConfig.PartitionResourceName == "" || partition == "" {
		return "", false
	}
	return v1.ResourceName(strings.ReplaceAll(accelConfig.PartitionResourceName, partitionPlaceholder, partition)), true
}

// matchesPartitionResourceName reports whether name could have been rendered from the
// PartitionResourceName template tmpl for some non-empty partition size.
func matchesPartitionResourceName(tmpl string, name v1.ResourceName) bool {
	prefix, suffix, found := strings.Cut(tmpl, partitionPlaceholder)
	if !found {
		return string(name) == tmpl
	}
	s := string(name)
	return len(s) > len(prefix)+len(suffix) && strings.HasPrefix(s, prefix) && strings.HasSuffix(s, suffix)
}

// acceleratorBaseResourceName maps an accelerator resource name to its device class's resource
// name. A name rendered from a device class's PartitionResourceName template, such as
// nvidia.com/mig-1g.5gb, maps to that class's ResourceName, and a plain accelerator resource
// name maps to itself. The second result is false for names that are not accelerator resources.
func acceleratorBaseResourceName(name v1.ResourceName) (v1.ResourceName, bool) {
	if _, ok := getAllAcceleratorResourceNames()[name]; ok {
		return name, true
	}
	cfg := config.GetK8sPluginConfig()
	for _, deviceClassConfig := range cfg.AcceleratorDeviceClasses {
		if deviceClassConfig.PartitionResourceName == "" || !matchesPartitionResourceName(deviceClassConfig.PartitionResourceName, name) {
			continue
		}
		if deviceClassConfig.ResourceName != "" {
			return deviceClassConfig.ResourceName, true
		}
		return cfg.GpuResourceName, true //nolint: staticcheck
	}
	return "", false
}

// getAllAcceleratorResourceNames returns the Kubernetes resource names for all accelerator devices.
func getAllAcceleratorResourceNames() map[v1.ResourceName]struct{} {
	cfg := config.GetK8sPluginConfig()
	acceleratorResourceNames := make(map[v1.ResourceName]struct{})

	// Add the legacy GPU resource name for backward compatibility
	acceleratorResourceNames[cfg.GpuResourceName] = struct{}{} //nolint: staticcheck

	// Add resource names from all configured accelerator device classes
	for _, deviceClassConfig := range cfg.AcceleratorDeviceClasses {
		if deviceClassConfig.ResourceName != "" {
			acceleratorResourceNames[deviceClassConfig.ResourceName] = struct{}{}
		}
	}
	return acceleratorResourceNames
}

// podRequiresAccelerator returns true if any container in the pod requires any accelerator
// devices, including per-partition accelerator resources.
func podRequiresAccelerator(podSpec *v1.PodSpec) bool {
	for _, cnt := range podSpec.Containers {
		for resourceName := range cnt.Resources.Limits {
			if _, ok := acceleratorBaseResourceName(resourceName); ok {
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
