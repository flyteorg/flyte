package flytek8s

import (
	"context"

	"github.com/lyft/flytestdlib/contextutils"
	v1 "k8s.io/api/core/v1"

	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/lyft/flyteplugins/go/tasks/v1/flytek8s/config"
	"github.com/lyft/flyteplugins/go/tasks/v1/types"
	"github.com/lyft/flyteplugins/go/tasks/v1/utils"
)

func GetContextEnvVars(ownerCtx context.Context) []v1.EnvVar {
	envVars := []v1.EnvVar{}

	if ownerCtx == nil {
		return envVars
	}

	// Injecting useful env vars from the context
	if wfName := contextutils.Value(ownerCtx, contextutils.WorkflowIDKey); wfName != "" {
		envVars = append(envVars,
			v1.EnvVar{
				Name:  "FLYTE_INTERNAL_EXECUTION_WORKFLOW",
				Value: wfName,
			},
		)
	}
	return envVars
}

func GetExecutionEnvVars(id types.TaskExecutionID) []v1.EnvVar {

	if id == nil || id.GetID().NodeExecutionId == nil || id.GetID().NodeExecutionId.ExecutionId == nil {
		return []v1.EnvVar{}
	}

	// Execution level env variables.
	nodeExecutionId := id.GetID().NodeExecutionId.ExecutionId
	envVars := []v1.EnvVar{
		{
			Name:  "FLYTE_INTERNAL_EXECUTION_ID",
			Value: nodeExecutionId.Name,
		},
		{
			Name:  "FLYTE_INTERNAL_EXECUTION_PROJECT",
			Value: nodeExecutionId.Project,
		},
		{
			Name:  "FLYTE_INTERNAL_EXECUTION_DOMAIN",
			Value: nodeExecutionId.Domain,
		},
		// TODO: Fill in these
		// {
		// 	Name:  "FLYTE_INTERNAL_EXECUTION_WORKFLOW",
		// 	Value: "",
		// },
		// {
		// 	Name:  "FLYTE_INTERNAL_EXECUTION_LAUNCHPLAN",
		// 	Value: "",
		// },
	}

	// Task definition Level env variables.
	if id.GetID().TaskId != nil {
		taskId := id.GetID().TaskId

		envVars = append(envVars,
			v1.EnvVar{
				Name:  "FLYTE_INTERNAL_TASK_PROJECT",
				Value: taskId.Project,
			},
			v1.EnvVar{
				Name:  "FLYTE_INTERNAL_TASK_DOMAIN",
				Value: taskId.Domain,
			},
			v1.EnvVar{
				Name:  "FLYTE_INTERNAL_TASK_NAME",
				Value: taskId.Name,
			},
			v1.EnvVar{
				Name:  "FLYTE_INTERNAL_TASK_VERSION",
				Value: taskId.Version,
			},
			// Historic Task Definition Level env variables.
			// Remove these once SDK is migrated to use the new ones.
			v1.EnvVar{
				Name:  "FLYTE_INTERNAL_PROJECT",
				Value: taskId.Project,
			},
			v1.EnvVar{
				Name:  "FLYTE_INTERNAL_DOMAIN",
				Value: taskId.Domain,
			},
			v1.EnvVar{
				Name:  "FLYTE_INTERNAL_NAME",
				Value: taskId.Name,
			},
			v1.EnvVar{
				Name:  "FLYTE_INTERNAL_VERSION",
				Value: taskId.Version,
			})

	}
	return envVars
}

func DecorateEnvVars(ctx context.Context, envVars []v1.EnvVar, id types.TaskExecutionID) []v1.EnvVar {
	// Injecting workflow name into the container's env vars
	envVars = append(envVars, GetContextEnvVars(ctx)...)

	envVars = append(envVars, GetExecutionEnvVars(id)...)

	for k, v := range config.GetK8sPluginConfig().DefaultEnvVars {
		envVars = append(envVars, v1.EnvVar{Name: k, Value: v})
	}
	return envVars
}

func GetTolerationsForResources(resourceRequirements ...v1.ResourceRequirements) []v1.Toleration {
	var tolerations []v1.Toleration
	resourceNames := sets.NewString()
	for _, resources := range resourceRequirements {
		for r := range resources.Limits {
			resourceNames.Insert(r.String())
		}
		for r := range resources.Requests {
			resourceNames.Insert(r.String())
		}
	}
	resourceTols := config.GetK8sPluginConfig().ResourceTolerations
	for _, r := range resourceNames.UnsortedList() {
		if v, ok := resourceTols[v1.ResourceName(r)]; ok {
			tolerations = append(tolerations, v...)
		}
	}
	return tolerations
}

func AddObjectMetadata(taskCtx types.TaskContext, o K8sResource) {
	o.SetNamespace(taskCtx.GetNamespace())
	o.SetAnnotations(UnionMaps(config.GetK8sPluginConfig().DefaultAnnotations, o.GetAnnotations(), utils.CopyMap(taskCtx.GetAnnotations())))
	o.SetLabels(UnionMaps(o.GetLabels(), utils.CopyMap(taskCtx.GetLabels()), config.GetK8sPluginConfig().DefaultLabels))
	o.SetOwnerReferences([]v12.OwnerReference{taskCtx.GetOwnerReference()})
	o.SetName(taskCtx.GetTaskExecutionID().GetGeneratedName())
	if config.GetK8sPluginConfig().InjectFinalizer {
		f := append(o.GetFinalizers(), finalizer)
		o.SetFinalizers(f)
	}
}
