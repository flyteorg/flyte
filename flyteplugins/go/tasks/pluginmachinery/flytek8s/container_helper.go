package flytek8s

import (
	"context"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/template"
	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
)

const resourceGPU = "gpu"

// ResourceNvidiaGPU is the name of the Nvidia GPU resource.
// Copied from: k8s.io/autoscaler/cluster-autoscaler/utils/gpu/gpu.go
const ResourceNvidiaGPU = "nvidia.com/gpu"

// Specifies whether resource resolution should assign unset resource requests or limits from platform defaults
// or existing container values.
const assignIfUnset = true

func MergeResources(in v1.ResourceRequirements, out *v1.ResourceRequirements) {
	if out.Limits == nil {
		out.Limits = in.Limits
	} else if in.Limits != nil {
		for key, val := range in.Limits {
			out.Limits[key] = val
		}
	}
	if out.Requests == nil {
		out.Requests = in.Requests
	} else if in.Requests != nil {
		for key, val := range in.Requests {
			out.Requests[key] = val
		}
	}
}

type ResourceRequirement struct {
	Request resource.Quantity
	Limit   resource.Quantity
}

func resolvePlatformDefaults(platformResources v1.ResourceRequirements, configCPU, configMemory resource.Quantity) v1.ResourceRequirements {
	if len(platformResources.Requests) == 0 {
		platformResources.Requests = make(v1.ResourceList)
	}

	if _, ok := platformResources.Requests[v1.ResourceCPU]; !ok {
		platformResources.Requests[v1.ResourceCPU] = configCPU
	}

	if _, ok := platformResources.Requests[v1.ResourceMemory]; !ok {
		platformResources.Requests[v1.ResourceMemory] = configMemory
	}

	if len(platformResources.Limits) == 0 {
		platformResources.Limits = make(v1.ResourceList)
	}

	return platformResources
}

// AdjustOrDefaultResource validates resources conform to platform limits and assigns defaults for Request and Limit values by
// using the Request when the Limit is unset, and vice versa.
func AdjustOrDefaultResource(request, limit, platformDefault, platformLimit resource.Quantity) ResourceRequirement {
	if request.IsZero() {
		if !limit.IsZero() {
			request = limit
		} else {
			request = platformDefault
		}
	}

	if limit.IsZero() {
		limit = request
	}

	return ensureResourceRange(request, limit, platformLimit)
}

func ensureResourceLimit(value, limit resource.Quantity) resource.Quantity {
	if value.IsZero() || limit.IsZero() {
		return value
	}

	if value.Cmp(limit) == 1 {
		return limit
	}

	return value
}

// ensureResourceRange doesn't assign resources unless they need to be adjusted downwards
func ensureResourceRange(request, limit, platformLimit resource.Quantity) ResourceRequirement {
	// Ensure request is < platformLimit
	request = ensureResourceLimit(request, platformLimit)
	// Ensure limit is < platformLimit
	limit = ensureResourceLimit(limit, platformLimit)
	// Ensure request is < limit
	request = ensureResourceLimit(request, limit)

	return ResourceRequirement{
		Request: request,
		Limit:   limit,
	}
}

func adjustResourceRequirement(resourceName v1.ResourceName, resourceRequirements,
	platformResources v1.ResourceRequirements, assignIfUnset bool) {

	var resourceValue ResourceRequirement
	if assignIfUnset {
		resourceValue = AdjustOrDefaultResource(resourceRequirements.Requests[resourceName],
			resourceRequirements.Limits[resourceName], platformResources.Requests[resourceName],
			platformResources.Limits[resourceName])
	} else {
		resourceValue = ensureResourceRange(resourceRequirements.Requests[resourceName],
			resourceRequirements.Limits[resourceName], platformResources.Limits[resourceName])
	}

	resourceRequirements.Requests[resourceName] = resourceValue.Request
	resourceRequirements.Limits[resourceName] = resourceValue.Limit
}

// ApplyResourceOverrides handles resource resolution, allocation and validation. Primarily, it ensures that container
// resources do not exceed defined platformResource limits and in the case of assignIfUnset, ensures that limits and
// requests are sensibly set for resources of all types.
// Furthermore, this function handles some clean-up such as converting GPU resources to the recognized Nvidia gpu
// resource name and deleting unsupported Storage-type resources.
func ApplyResourceOverrides(resources, platformResources v1.ResourceRequirements, assignIfUnset bool) v1.ResourceRequirements {
	if len(resources.Requests) == 0 {
		resources.Requests = make(v1.ResourceList)
	}

	if len(resources.Limits) == 0 {
		resources.Limits = make(v1.ResourceList)
	}

	// As a fallback, in the case the Flyte workflow object does not have platformResource defaults set, the defaults
	// come from the plugin config.
	platformResources = resolvePlatformDefaults(platformResources, config.GetK8sPluginConfig().DefaultCPURequest,
		config.GetK8sPluginConfig().DefaultMemoryRequest)

	adjustResourceRequirement(v1.ResourceCPU, resources, platformResources, assignIfUnset)
	adjustResourceRequirement(v1.ResourceMemory, resources, platformResources, assignIfUnset)

	_, ephemeralStorageRequested := resources.Requests[v1.ResourceEphemeralStorage]
	_, ephemeralStorageLimited := resources.Limits[v1.ResourceEphemeralStorage]

	if ephemeralStorageRequested || ephemeralStorageLimited {
		adjustResourceRequirement(v1.ResourceEphemeralStorage, resources, platformResources, assignIfUnset)
	}

	// TODO: Make configurable. 1/15/2019 Flyte Cluster doesn't support setting storage requests/limits.
	// https://github.com/kubernetes/enhancements/issues/362
	delete(resources.Requests, v1.ResourceStorage)
	delete(resources.Limits, v1.ResourceStorage)

	gpuResourceName := config.GetK8sPluginConfig().GpuResourceName
	shouldAdjustGPU := false
	_, gpuRequested := resources.Requests[gpuResourceName]
	_, gpuLimited := resources.Limits[gpuResourceName]
	if gpuRequested || gpuLimited {
		shouldAdjustGPU = true
	}

	// Override GPU
	if res, found := resources.Requests[resourceGPU]; found {
		resources.Requests[gpuResourceName] = res
		delete(resources.Requests, resourceGPU)
		shouldAdjustGPU = true
	}

	if res, found := resources.Limits[resourceGPU]; found {
		resources.Limits[gpuResourceName] = res
		delete(resources.Limits, resourceGPU)
		shouldAdjustGPU = true
	}

	if shouldAdjustGPU {
		adjustResourceRequirement(gpuResourceName, resources, platformResources, assignIfUnset)
	}

	return resources
}

// ToK8sContainer transforms a task template target of type core.Container into a bare-bones kubernetes container, which
// can be further modified with flyte-specific customizations specified by various static and run-time attributes.
func ToK8sContainer(ctx context.Context, taskContainer *core.Container, iFace *core.TypedInterface, parameters template.Parameters) (*v1.Container, error) {
	// Perform preliminary validations
	if parameters.TaskExecMetadata.GetOverrides() == nil {
		return nil, errors.Errorf(errors.BadTaskSpecification, "platform/compiler error, overrides not set for task")
	}
	if parameters.TaskExecMetadata.GetOverrides() == nil || parameters.TaskExecMetadata.GetOverrides().GetResources() == nil {
		return nil, errors.Errorf(errors.BadTaskSpecification, "resource requirements not found for container task, required!")
	}
	// Make the container name the same as the pod name, unless it violates K8s naming conventions
	// Container names are subject to the DNS-1123 standard
	containerName := parameters.TaskExecMetadata.GetTaskExecutionID().GetGeneratedName()
	if errs := validation.IsDNS1123Label(containerName); len(errs) > 0 {
		containerName = rand.String(4)
	}
	container := &v1.Container{
		Name:                     containerName,
		Image:                    taskContainer.GetImage(),
		Args:                     taskContainer.GetArgs(),
		Command:                  taskContainer.GetCommand(),
		Env:                      ToK8sEnvVar(taskContainer.GetEnv()),
		TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
	}
	if err := AddCoPilotToContainer(ctx, config.GetK8sPluginConfig().CoPilot, container, iFace, taskContainer.DataConfig); err != nil {
		return nil, err
	}
	if container.SecurityContext == nil && config.GetK8sPluginConfig().DefaultSecurityContext != nil {
		container.SecurityContext = config.GetK8sPluginConfig().DefaultSecurityContext.DeepCopy()
	}
	return container, nil
}

//go:generate enumer -type=ResourceCustomizationMode -trimprefix=ResourceCustomizationMode

type ResourceCustomizationMode int

const (
	// ResourceCustomizationModeAssignResources is used for container tasks where resources are validated and assigned if necessary.
	ResourceCustomizationModeAssignResources ResourceCustomizationMode = iota
	// ResourceCustomizationModeMergeExistingResources is used for primary containers in pod tasks where container requests and limits are
	// merged, validated and assigned if necessary.
	ResourceCustomizationModeMergeExistingResources
	// ResourceCustomizationModeEnsureExistingResourcesInRange is used for secondary containers in pod tasks where requests and limits are only
	// adjusted if needed (downwards).
	ResourceCustomizationModeEnsureExistingResourcesInRange
)

// AddFlyteCustomizationsToContainer takes a container definition which specifies how to run a Flyte task and fills in
// templated command and argument values, updates resources and decorates environment variables with platform and
// task-specific customizations.
func AddFlyteCustomizationsToContainer(ctx context.Context, parameters template.Parameters,
	mode ResourceCustomizationMode, container *v1.Container) error {
	modifiedCommand, err := template.Render(ctx, container.Command, parameters)
	if err != nil {
		return err
	}
	container.Command = modifiedCommand

	modifiedArgs, err := template.Render(ctx, container.Args, parameters)
	if err != nil {
		return err
	}
	container.Args = modifiedArgs

	container.Env = DecorateEnvVars(ctx, container.Env, parameters.TaskExecMetadata.GetTaskExecutionID())

	if parameters.TaskExecMetadata.GetOverrides() != nil && parameters.TaskExecMetadata.GetOverrides().GetResources() != nil {
		res := parameters.TaskExecMetadata.GetOverrides().GetResources()
		platformResources := parameters.TaskExecMetadata.GetPlatformResources()
		if platformResources == nil {
			platformResources = &v1.ResourceRequirements{}
		}

		logger.Infof(ctx, "ApplyResourceOverrides with Resources [%v], Platform Resources [%v] and Container"+
			" Resources [%v] with mode [%v]", res, platformResources, container.Resources, mode)

		switch mode {
		case ResourceCustomizationModeAssignResources:
			container.Resources = ApplyResourceOverrides(*res, *platformResources, assignIfUnset)
		case ResourceCustomizationModeMergeExistingResources:
			MergeResources(*res, &container.Resources)
			container.Resources = ApplyResourceOverrides(container.Resources, *platformResources, assignIfUnset)
		case ResourceCustomizationModeEnsureExistingResourcesInRange:
			container.Resources = ApplyResourceOverrides(container.Resources, *platformResources, !assignIfUnset)
		}

		logger.Infof(ctx, "Adjusted container resources [%v]", container.Resources)
	}
	return nil
}
