package flytek8s

import (
	"context"
	"regexp"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core/template"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytestdlib/logger"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/lyft/flyteplugins/go/tasks/errors"
	pluginsCore "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io"
)

var isAcceptableK8sName, _ = regexp.Compile("[a-z0-9]([-a-z0-9]*[a-z0-9])?")

const resourceGPU = "gpu"

// ResourceNvidiaGPU is the name of the Nvidia GPU resource.
// Copied from: k8s.io/autoscaler/cluster-autoscaler/utils/gpu/gpu.go
const ResourceNvidiaGPU = "nvidia.com/gpu"

func ApplyResourceOverrides(ctx context.Context, resources v1.ResourceRequirements) *v1.ResourceRequirements {
	// set memory and cpu to default if not provided by user.
	if len(resources.Requests) == 0 {
		resources.Requests = make(v1.ResourceList)
	}

	if len(resources.Limits) == 0 {
		resources.Limits = make(v1.ResourceList)
	}

	if _, found := resources.Requests[v1.ResourceCPU]; !found {
		// use cpu limit if set else default to config
		if _, limitSet := resources.Limits[v1.ResourceCPU]; limitSet {
			resources.Requests[v1.ResourceCPU] = resources.Limits[v1.ResourceCPU]
		} else {
			resources.Requests[v1.ResourceCPU] = resource.MustParse(config.GetK8sPluginConfig().DefaultCPURequest)
		}
	}

	if _, found := resources.Requests[v1.ResourceMemory]; !found {
		// use memory limit if set else default to config
		if _, limitSet := resources.Limits[v1.ResourceCPU]; limitSet {
			resources.Requests[v1.ResourceMemory] = resources.Limits[v1.ResourceMemory]
		} else {
			resources.Requests[v1.ResourceMemory] = resource.MustParse(config.GetK8sPluginConfig().DefaultMemoryRequest)
		}
	}

	if _, found := resources.Limits[v1.ResourceCPU]; !found {
		logger.Infof(ctx, "found cpu limit missing, setting limit to the requested value %v", resources.Requests[v1.ResourceCPU])
		resources.Limits[v1.ResourceCPU] = resources.Requests[v1.ResourceCPU]
	}

	if _, found := resources.Limits[v1.ResourceMemory]; !found {
		logger.Infof(ctx, "found memory limit missing, setting limit to the requested value %v", resources.Requests[v1.ResourceMemory])
		resources.Limits[v1.ResourceMemory] = resources.Requests[v1.ResourceMemory]
	}

	// TODO: Make configurable. 1/15/2019 Flyte Cluster doesn't support setting storage requests/limits.
	// https://github.com/kubernetes/enhancements/issues/362
	delete(resources.Requests, v1.ResourceStorage)
	delete(resources.Requests, v1.ResourceEphemeralStorage)

	delete(resources.Limits, v1.ResourceStorage)
	delete(resources.Limits, v1.ResourceEphemeralStorage)

	// Override GPU
	if res, found := resources.Requests[resourceGPU]; found {
		resources.Requests[ResourceNvidiaGPU] = res
		delete(resources.Requests, resourceGPU)
	}
	if res, found := resources.Limits[resourceGPU]; found {
		resources.Limits[ResourceNvidiaGPU] = res
		delete(resources.Requests, resourceGPU)
	}

	return &resources
}

// Returns a K8s Container for the execution
func ToK8sContainer(ctx context.Context, taskExecutionMetadata pluginsCore.TaskExecutionMetadata, taskContainer *core.Container, iFace *core.TypedInterface,
	inputReader io.InputReader, outputPaths io.OutputFilePaths) (*v1.Container, error) {
	modifiedCommand, err := template.ReplaceTemplateCommandArgs(ctx, taskExecutionMetadata, taskContainer.GetCommand(), inputReader, outputPaths)
	if err != nil {
		return nil, err
	}

	modifiedArgs, err := template.ReplaceTemplateCommandArgs(ctx, taskExecutionMetadata, taskContainer.GetArgs(), inputReader, outputPaths)
	if err != nil {
		return nil, err
	}

	envVars := DecorateEnvVars(ctx, ToK8sEnvVar(taskContainer.GetEnv()), taskExecutionMetadata.GetTaskExecutionID())

	if taskExecutionMetadata.GetOverrides() == nil {
		return nil, errors.Errorf(errors.BadTaskSpecification, "platform/compiler error, overrides not set for task")
	}
	if taskExecutionMetadata.GetOverrides() == nil || taskExecutionMetadata.GetOverrides().GetResources() == nil {
		return nil, errors.Errorf(errors.BadTaskSpecification, "resource requirements not found for container task, required!")
	}

	res := taskExecutionMetadata.GetOverrides().GetResources()
	if res != nil {
		res = ApplyResourceOverrides(ctx, *res)
	}

	// Make the container name the same as the pod name, unless it violates K8s naming conventions
	// Container names are subject to the DNS-1123 standard
	containerName := taskExecutionMetadata.GetTaskExecutionID().GetGeneratedName()
	if !isAcceptableK8sName.MatchString(containerName) || len(containerName) > 63 {
		containerName = rand.String(4)
	}
	c := &v1.Container{
		Name:                     containerName,
		Image:                    taskContainer.GetImage(),
		Args:                     modifiedArgs,
		Command:                  modifiedCommand,
		Env:                      envVars,
		TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
	}

	if res != nil {
		c.Resources = *res
	}

	if err := AddCoPilotToContainer(ctx, config.GetK8sPluginConfig().CoPilot, c, iFace, taskContainer.DataConfig); err != nil {
		return nil, err
	}
	return c, nil
}
