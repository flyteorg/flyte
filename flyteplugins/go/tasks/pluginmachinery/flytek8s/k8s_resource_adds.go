package flytek8s

import (
	"context"
	"os"
	"strconv"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"

	"github.com/flyteorg/flytestdlib/contextutils"
	v1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/util/sets"

	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
)

func GetContextEnvVars(ownerCtx context.Context) []v1.EnvVar {
	var envVars []v1.EnvVar

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

func GetExecutionEnvVars(id pluginsCore.TaskExecutionID) []v1.EnvVar {

	if id == nil || id.GetID().NodeExecutionId == nil || id.GetID().NodeExecutionId.ExecutionId == nil {
		return []v1.EnvVar{}
	}

	// Execution level env variables.
	nodeExecutionID := id.GetID().NodeExecutionId.ExecutionId
	attemptNumber := strconv.Itoa(int(id.GetID().RetryAttempt))
	envVars := []v1.EnvVar{
		{
			Name:  "FLYTE_INTERNAL_EXECUTION_ID",
			Value: nodeExecutionID.Name,
		},
		{
			Name:  "FLYTE_INTERNAL_EXECUTION_PROJECT",
			Value: nodeExecutionID.Project,
		},
		{
			Name:  "FLYTE_INTERNAL_EXECUTION_DOMAIN",
			Value: nodeExecutionID.Domain,
		},
		{
			Name:  "FLYTE_ATTEMPT_NUMBER",
			Value: attemptNumber,
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
		taskID := id.GetID().TaskId

		envVars = append(envVars,
			v1.EnvVar{
				Name:  "FLYTE_INTERNAL_TASK_PROJECT",
				Value: taskID.Project,
			},
			v1.EnvVar{
				Name:  "FLYTE_INTERNAL_TASK_DOMAIN",
				Value: taskID.Domain,
			},
			v1.EnvVar{
				Name:  "FLYTE_INTERNAL_TASK_NAME",
				Value: taskID.Name,
			},
			v1.EnvVar{
				Name:  "FLYTE_INTERNAL_TASK_VERSION",
				Value: taskID.Version,
			},
			// Historic Task Definition Level env variables.
			// Remove these once SDK is migrated to use the new ones.
			v1.EnvVar{
				Name:  "FLYTE_INTERNAL_PROJECT",
				Value: taskID.Project,
			},
			v1.EnvVar{
				Name:  "FLYTE_INTERNAL_DOMAIN",
				Value: taskID.Domain,
			},
			v1.EnvVar{
				Name:  "FLYTE_INTERNAL_NAME",
				Value: taskID.Name,
			},
			v1.EnvVar{
				Name:  "FLYTE_INTERNAL_VERSION",
				Value: taskID.Version,
			})

	}
	return envVars
}

func DecorateEnvVars(ctx context.Context, envVars []v1.EnvVar, taskEnvironmentVariables map[string]string, id pluginsCore.TaskExecutionID) []v1.EnvVar {
	envVars = append(envVars, GetContextEnvVars(ctx)...)
	envVars = append(envVars, GetExecutionEnvVars(id)...)

	for k, v := range taskEnvironmentVariables {
		envVars = append(envVars, v1.EnvVar{Name: k, Value: v})
	}
	for k, v := range config.GetK8sPluginConfig().DefaultEnvVars {
		envVars = append(envVars, v1.EnvVar{Name: k, Value: v})
	}
	for k, envVarName := range config.GetK8sPluginConfig().DefaultEnvVarsFromEnv {
		value := os.Getenv(envVarName)
		envVars = append(envVars, v1.EnvVar{Name: k, Value: value})
	}

	return envVars
}

func GetPodTolerations(interruptible bool, resourceRequirements ...v1.ResourceRequirements) []v1.Toleration {
	// 1. Get the tolerations for the resources requested
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

	// 2. Get the tolerations for interruptible pods
	if interruptible {
		tolerations = append(tolerations, config.GetK8sPluginConfig().InterruptibleTolerations...)
	}

	// 3. Add default tolerations
	tolerations = append(tolerations, config.GetK8sPluginConfig().DefaultTolerations...)

	return tolerations
}
