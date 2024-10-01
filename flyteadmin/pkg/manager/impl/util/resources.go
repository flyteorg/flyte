package util

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	workflowengineInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/workflowengine/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

// parseQuantityNoError parses the k8s defined resource quantity gracefully masking errors.
func parseQuantityNoError(ctx context.Context, ownerID, name, value string) resource.Quantity {
	q, err := resource.ParseQuantity(value)
	if err != nil {
		logger.Infof(ctx, "Failed to parse owner's [%s] resource [%s]'s value [%s] with err: %v", ownerID, name, value, err)
	}

	return q
}

// getTaskResourcesAsSet converts a list of flyteidl `ResourceEntry` messages into a singular `TaskResourceSet`.
func getTaskResourcesAsSet(ctx context.Context, identifier *core.Identifier,
	resourceEntries []*core.Resources_ResourceEntry, resourceName string) runtimeInterfaces.TaskResourceSet {

	result := runtimeInterfaces.TaskResourceSet{}
	for _, entry := range resourceEntries {
		switch entry.Name {
		case core.Resources_CPU:
			result.CPU = parseQuantityNoError(ctx, identifier.String(), fmt.Sprintf("%v.cpu", resourceName), entry.Value)
		case core.Resources_MEMORY:
			result.Memory = parseQuantityNoError(ctx, identifier.String(), fmt.Sprintf("%v.memory", resourceName), entry.Value)
		case core.Resources_EPHEMERAL_STORAGE:
			result.EphemeralStorage = parseQuantityNoError(ctx, identifier.String(),
				fmt.Sprintf("%v.ephemeral storage", resourceName), entry.Value)
		case core.Resources_GPU:
			result.GPU = parseQuantityNoError(ctx, identifier.String(), "gpu", entry.Value)
		}
	}

	return result
}

// GetCompleteTaskResourceRequirements parses the resource requests and limits from the `TaskTemplate` Container.
func GetCompleteTaskResourceRequirements(ctx context.Context, identifier *core.Identifier, task *core.CompiledTask) workflowengineInterfaces.TaskResources {
	return workflowengineInterfaces.TaskResources{
		Defaults: getTaskResourcesAsSet(ctx, identifier, task.GetTemplate().GetContainer().Resources.Requests, "requests"),
		Limits:   getTaskResourcesAsSet(ctx, identifier, task.GetTemplate().GetContainer().Resources.Limits, "limits"),
	}
}

// FromAdminProtoTaskResourceSpec parses the flyteidl `TaskResourceSpec` message into a `TaskResourceSet`.
func FromAdminProtoTaskResourceSpec(ctx context.Context, spec *admin.TaskResourceSpec) runtimeInterfaces.TaskResourceSet {
	result := runtimeInterfaces.TaskResourceSet{}
	if spec == nil {
		return result
	}

	if len(spec.Cpu) > 0 {
		result.CPU = parseQuantityNoError(ctx, "project", "cpu", spec.Cpu)
	}

	if len(spec.Memory) > 0 {
		result.Memory = parseQuantityNoError(ctx, "project", "memory", spec.Memory)
	}

	if len(spec.EphemeralStorage) > 0 {
		result.EphemeralStorage = parseQuantityNoError(ctx, "project", "ephemeral storage", spec.EphemeralStorage)
	}

	if len(spec.Gpu) > 0 {
		result.GPU = parseQuantityNoError(ctx, "project", "gpu", spec.Gpu)
	}

	return result
}

func ToAdminProtoTaskResourceSpec(taskResourceSet *runtimeInterfaces.TaskResourceSet) *admin.TaskResourceSpec {
	return &admin.TaskResourceSpec{
		Cpu:              taskResourceSet.CPU.String(),
		Memory:           taskResourceSet.Memory.String(),
		EphemeralStorage: taskResourceSet.EphemeralStorage.String(),
		Gpu:              taskResourceSet.GPU.String(),
	}
}

// GetTaskResources returns the most specific default and limit task resources for the specified id. This first checks
// if there is a matchable resource(s) defined, and uses the highest priority one, otherwise it falls back to using the
// flyteadmin default configured values.
func GetTaskResources(ctx context.Context, id *core.Identifier, resourceManager interfaces.ResourceInterface,
	taskResourceConfig runtimeInterfaces.TaskResourceConfiguration) workflowengineInterfaces.TaskResources {

	request := interfaces.ResourceRequest{
		ResourceType: admin.MatchableResource_TASK_RESOURCE,
	}
	if id != nil && len(id.Org) > 0 {
		request.Org = id.Org
	}
	if id != nil && len(id.Project) > 0 {
		request.Project = id.Project
	}
	if id != nil && len(id.Domain) > 0 {
		request.Domain = id.Domain
	}
	if id != nil && id.ResourceType == core.ResourceType_WORKFLOW && len(id.Name) > 0 {
		request.Workflow = id.Name
	}

	resource, err := resourceManager.GetResource(ctx, request)
	if err != nil && !errors.IsDoesNotExistError(err) {
		logger.Infof(ctx, "Failed to fetch override values when assigning task resource default values for [%+v]: %v",
			id, err)
	}

	logger.Debugf(ctx, "Assigning task requested resources for [%+v]", id)
	var taskResourceAttributes = workflowengineInterfaces.TaskResources{}
	if resource != nil && resource.Attributes != nil && resource.Attributes.GetTaskResourceAttributes() != nil {
		taskResourceAttributes.Defaults = FromAdminProtoTaskResourceSpec(ctx, resource.Attributes.GetTaskResourceAttributes().Defaults)
		taskResourceAttributes.Limits = FromAdminProtoTaskResourceSpec(ctx, resource.Attributes.GetTaskResourceAttributes().Limits)
	} else {
		taskResourceAttributes = workflowengineInterfaces.TaskResources{
			Defaults: taskResourceConfig.GetDefaults(),
			Limits:   taskResourceConfig.GetLimits(),
		}
	}

	return taskResourceAttributes
}
