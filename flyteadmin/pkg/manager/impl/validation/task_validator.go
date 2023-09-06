// Miscellaneous functions to validate that required proto and spec fields are non empty when required for execution.
package validation

import (
	"context"
	"strings"

	repositoryInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/shared"
	runtime "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	workflowengineInterfaces "github.com/flyteorg/flyteadmin/pkg/workflowengine/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
	corev1 "k8s.io/api/core/v1"

	"google.golang.org/grpc/codes"
	"k8s.io/apimachinery/pkg/api/resource"
)

var whitelistedTaskErr = errors.NewFlyteAdminErrorf(codes.InvalidArgument, "task type must be whitelisted before use")

// This is called for a task with a non-nil container.
func validateContainer(task core.TaskTemplate, platformTaskResources workflowengineInterfaces.TaskResources) error {
	if err := ValidateEmptyStringField(task.GetContainer().Image, shared.Image); err != nil {
		return err
	}

	if task.GetContainer().Resources == nil {
		return nil
	}
	if err := validateTaskResources(task.Id, platformTaskResources.Limits, task.GetContainer().Resources.Requests,
		task.GetContainer().Resources.Limits); err != nil {
		logger.Debugf(context.Background(), "encountered errors validating task resources for [%+v]: %v",
			task.Id, err)
		return err
	}
	return nil
}

// This is called for a task with a non-nil k8s pod.
func validateK8sPod(task core.TaskTemplate, platformTaskResources workflowengineInterfaces.TaskResources) error {
	if task.GetK8SPod().PodSpec == nil {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"invalid TaskSpecification, pod tasks should specify their target as a K8sPod with a defined pod spec")
	}
	var podSpec corev1.PodSpec
	if err := utils.UnmarshalStructToObj(task.GetK8SPod().PodSpec, &podSpec); err != nil {
		logger.Debugf(context.Background(), "failed to unmarshal k8s podspec [%+v]: %v",
			task.GetK8SPod().PodSpec, err)
		return err
	}
	platformTaskResourceLimits := taskResourceSetToMap(platformTaskResources.Limits)
	for _, container := range podSpec.Containers {
		err := validateResource(task.Id, resourceListToQuantity(container.Resources.Requests),
			resourceListToQuantity(container.Resources.Limits), platformTaskResourceLimits)
		if err != nil {
			logger.Debugf(context.Background(), "encountered errors validating task resources for [%+v]: %v",
				task.Id, err)
			return err
		}
	}

	return nil
}

func validateRuntimeMetadata(metadata core.RuntimeMetadata) error {
	if err := ValidateEmptyStringField(metadata.Version, shared.RuntimeVersion); err != nil {
		return err
	}
	return nil
}

func validateTaskTemplate(taskID core.Identifier, task core.TaskTemplate,
	platformTaskResources workflowengineInterfaces.TaskResources, whitelistConfig runtime.WhitelistConfiguration) error {

	if err := ValidateEmptyStringField(task.Type, shared.Type); err != nil {
		return err
	}
	if err := validateTaskType(taskID, task.Type, whitelistConfig); err != nil {
		return err
	}
	if task.Metadata == nil {
		return shared.GetMissingArgumentError(shared.Metadata)
	}
	if task.Metadata.Runtime != nil {
		if err := validateRuntimeMetadata(*task.Metadata.Runtime); err != nil {
			return err
		}
	}
	if task.Interface == nil {
		// The actual interface proto has nothing to validate.
		return shared.GetMissingArgumentError(shared.TypedInterface)
	}

	if task.GetContainer() != nil {
		return validateContainer(task, platformTaskResources)
	}
	if task.GetK8SPod() != nil {
		return validateK8sPod(task, platformTaskResources)
	}
	return nil
}

func ValidateTask(
	ctx context.Context, request admin.TaskCreateRequest, db repositoryInterfaces.Repository,
	platformTaskResources workflowengineInterfaces.TaskResources, whitelistConfig runtime.WhitelistConfiguration,
	applicationConfig runtime.ApplicationConfiguration) error {
	if err := ValidateIdentifier(request.Id, common.Task); err != nil {
		return err
	}
	if err := ValidateProjectAndDomain(ctx, db, applicationConfig, request.Id.Project, request.Id.Domain); err != nil {
		return err
	}
	if request.Spec == nil || request.Spec.Template == nil {
		return shared.GetMissingArgumentError(shared.Spec)
	}
	return validateTaskTemplate(*request.Id, *request.Spec.Template, platformTaskResources, whitelistConfig)
}

func taskResourceSetToMap(
	resourceSet runtimeInterfaces.TaskResourceSet) map[core.Resources_ResourceName]*resource.Quantity {
	resourceMap := make(map[core.Resources_ResourceName]*resource.Quantity)
	if !resourceSet.CPU.IsZero() {
		resourceMap[core.Resources_CPU] = &resourceSet.CPU
	}
	if !resourceSet.Memory.IsZero() {
		resourceMap[core.Resources_MEMORY] = &resourceSet.Memory
	}
	if !resourceSet.GPU.IsZero() {
		resourceMap[core.Resources_GPU] = &resourceSet.GPU
	}
	if !resourceSet.EphemeralStorage.IsZero() {
		resourceMap[core.Resources_EPHEMERAL_STORAGE] = &resourceSet.EphemeralStorage
	}
	return resourceMap
}

func addResourceEntryToMap(
	identifier *core.Identifier, entry *core.Resources_ResourceEntry,
	resourceEntries *map[core.Resources_ResourceName]resource.Quantity) error {
	if _, ok := (*resourceEntries)[entry.Name]; ok {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"can't specify %v limit for task [%+v] multiple times", entry.Name, identifier)
	}
	quantity, err := resource.ParseQuantity(entry.Value)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"Parsing of %v request failed for value %v - reason  %v. "+
				"Please follow K8s conventions for resources "+
				"https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/", entry.Name, entry.Value, err)
	}
	(*resourceEntries)[entry.Name] = quantity
	return nil
}

func isWholeNumber(quantity resource.Quantity) bool {
	// Assert k8s quantity is a whole number
	return quantity.MilliValue()%1000 == 0
}

func resourceListToQuantity(resources corev1.ResourceList) map[core.Resources_ResourceName]resource.Quantity {
	var requestedToQuantity = make(map[core.Resources_ResourceName]resource.Quantity)
	for name, quantity := range resources {
		resourceName := core.Resources_ResourceName(core.Resources_ResourceName_value[strings.ToUpper(name.String())])
		requestedToQuantity[resourceName] = quantity
	}
	return requestedToQuantity
}

func requestedResourcesToQuantity(
	identifier *core.Identifier, resources []*core.Resources_ResourceEntry) (
	map[core.Resources_ResourceName]resource.Quantity, error) {

	var requestedToQuantity = make(map[core.Resources_ResourceName]resource.Quantity)
	for _, limitEntry := range resources {
		switch limitEntry.Name {
		case core.Resources_CPU:
			fallthrough
		case core.Resources_MEMORY:
			err := addResourceEntryToMap(identifier, limitEntry, &requestedToQuantity)
			if err != nil {
				return nil, err
			}
		case core.Resources_GPU:
			err := addResourceEntryToMap(identifier, limitEntry, &requestedToQuantity)
			if err != nil {
				return nil, err
			}
			if !isWholeNumber(requestedToQuantity[core.Resources_GPU]) {
				return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
					"gpu for [%+v] must be a whole number, got: %s instead", identifier, limitEntry.Value)
			}
		case core.Resources_EPHEMERAL_STORAGE:
			err := addResourceEntryToMap(identifier, limitEntry, &requestedToQuantity)
			if err != nil {
				return nil, err
			}
		default:
			continue
		}
	}
	return requestedToQuantity, nil
}

func validateTaskResources(
	identifier *core.Identifier, taskResourceLimits runtimeInterfaces.TaskResourceSet,
	requestedTaskResourceDefaults, requestedTaskResourceLimits []*core.Resources_ResourceEntry) error {
	requestedResourceDefaults, err := requestedResourcesToQuantity(identifier, requestedTaskResourceDefaults)
	if err != nil {
		return err
	}

	requestedResourceLimits, err := requestedResourcesToQuantity(identifier, requestedTaskResourceLimits)
	if err != nil {
		return err
	}

	platformTaskResourceLimits := taskResourceSetToMap(taskResourceLimits)

	return validateResource(identifier, requestedResourceDefaults, requestedResourceLimits, platformTaskResourceLimits)
}

func validateResource(identifier *core.Identifier, requestedResourceDefaults,
	requestedResourceLimits map[core.Resources_ResourceName]resource.Quantity,
	platformTaskResourceLimits map[core.Resources_ResourceName]*resource.Quantity) error {
	for resourceName, defaultQuantity := range requestedResourceDefaults {
		switch resourceName {
		case core.Resources_CPU:
			fallthrough
		case core.Resources_EPHEMERAL_STORAGE:
			fallthrough
		case core.Resources_MEMORY:
			limitQuantity, ok := requestedResourceLimits[resourceName]
			if ok && limitQuantity.Value() < defaultQuantity.Value() {
				// Only assert the requested limit is greater than the requested default when the limit is actually set
				return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
					"Requested %v default [%v] is greater than the limit [%v]."+
						" Please fix your configuration", resourceName, defaultQuantity.String(), limitQuantity.String())
			}
			platformLimit, platformLimitOk := platformTaskResourceLimits[resourceName]
			if ok && platformLimitOk && limitQuantity.Value() > platformLimit.Value() {
				// Also check that the requested limit is less than the platform task limit.
				return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
					"Requested %v limit [%v] is greater than current limit set in the platform configuration"+
						" [%v]. Please contact Flyte Admins to change these limits or consult the configuration",
					resourceName, limitQuantity.String(), platformLimit.String())
			}
			if platformLimitOk && defaultQuantity.Value() > platformTaskResourceLimits[resourceName].Value() {
				// Also check that the requested limit is less than the platform task limit.
				return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
					"Requested %v default [%v] is greater than current limit set in the platform configuration"+
						" [%v]. Please contact Flyte Admins to change these limits or consult the configuration",
					resourceName, defaultQuantity.String(), platformTaskResourceLimits[resourceName].String())
			}
		case core.Resources_GPU:
			limitQuantity, ok := requestedResourceLimits[resourceName]
			if ok && defaultQuantity.Value() != limitQuantity.Value() {
				return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
					"For extended resource 'gpu' the default value must equal the limit value for task [%+v]",
					identifier)
			}
			platformLimit, platformLimitOk := platformTaskResourceLimits[resourceName]
			if platformLimitOk && defaultQuantity.Value() > platformLimit.Value() {
				return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
					"Requested %v default [%v] is greater than current limit set in the platform configuration"+
						" [%v]. Please contact Flyte Admins to change these limits or consult the configuration",
					resourceName, defaultQuantity.String(), platformLimit.String())
			}
		}
	}
	return nil
}

func validateTaskType(taskID core.Identifier, taskType string, whitelistConfig runtime.WhitelistConfiguration) error {
	taskTypeWhitelist := whitelistConfig.GetTaskTypeWhitelist()
	if taskTypeWhitelist == nil {
		return nil
	}
	scopes, ok := taskTypeWhitelist[taskType]
	if !ok || scopes == nil || len(scopes) == 0 {
		return nil
	}
	for _, scope := range scopes {
		if scope.Project == "" {
			// All projects whitelisted
			return nil
		} else if scope.Project != taskID.Project {
			continue
		}
		// We have a potential match! Verify that this task type is approved given the specifity of the whitelist.
		if scope.Domain == "" {
			// All domains for this project are whitelisted
			return nil
		} else if scope.Domain == taskID.Domain {
			return nil
		}

	}
	return whitelistedTaskErr
}
