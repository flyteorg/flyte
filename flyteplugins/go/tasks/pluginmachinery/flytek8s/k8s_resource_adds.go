package flytek8s

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	propellerCfg "github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
)

const (
	flyteExecutionURL = "FLYTE_EXECUTION_URL"

	FlyteInternalWorkerNameEnvVarKey        = "_F_WN"  // "FLYTE_INTERNAL_WORKER_NAME"
	FlyteInternalDistErrorStrategyEnvVarKey = "_F_DES" // "FLYTE_INTERNAL_DIST_ERROR_STRATEGY"
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

func GetExecutionEnvVars(id pluginsCore.TaskExecutionID, consoleURL string) []v1.EnvVar {

	//nolint:protogetter
	if id == nil || id.GetID().NodeExecutionId == nil || id.GetID().NodeExecutionId.GetExecutionId() == nil {
		return []v1.EnvVar{}
	}

	// Execution level env variables.
	nodeExecutionID := id.GetID().NodeExecutionId.GetExecutionId() //nolint:protogetter
	attemptNumber := strconv.Itoa(int(id.GetID().RetryAttempt))    //nolint:protogetter
	envVars := []v1.EnvVar{
		{
			Name:  "FLYTE_INTERNAL_EXECUTION_ID",
			Value: nodeExecutionID.GetName(),
		},
		{
			Name:  "FLYTE_INTERNAL_EXECUTION_PROJECT",
			Value: nodeExecutionID.GetProject(),
		},
		{
			Name:  "FLYTE_INTERNAL_EXECUTION_DOMAIN",
			Value: nodeExecutionID.GetDomain(),
		},
		{
			Name:  "FLYTE_ATTEMPT_NUMBER",
			Value: attemptNumber,
		},
	}

	if len(consoleURL) > 0 {
		consoleURL = strings.TrimRight(consoleURL, "/")
		envVars = append(envVars, v1.EnvVar{
			Name:  flyteExecutionURL,
			Value: fmt.Sprintf("%s/projects/%s/domains/%s/executions/%s/nodeId/%s/nodes", consoleURL, nodeExecutionID.GetProject(), nodeExecutionID.GetDomain(), nodeExecutionID.GetName(), id.GetUniqueNodeID()),
		})
	}

	// Task definition Level env variables.
	if id.GetID().TaskId != nil { //nolint:protogetter
		taskID := id.GetID().TaskId //nolint:protogetter

		envVars = append(envVars,
			v1.EnvVar{
				Name:  "FLYTE_INTERNAL_TASK_PROJECT",
				Value: taskID.GetProject(),
			},
			v1.EnvVar{
				Name:  "FLYTE_INTERNAL_TASK_DOMAIN",
				Value: taskID.GetDomain(),
			},
			v1.EnvVar{
				Name:  "FLYTE_INTERNAL_TASK_NAME",
				Value: taskID.GetName(),
			},
			v1.EnvVar{
				Name:  "FLYTE_INTERNAL_TASK_VERSION",
				Value: taskID.GetVersion(),
			})

	}
	return envVars
}

func GetLiteralOffloadingEnvVars() []v1.EnvVar {
	propellerConfig := propellerCfg.GetConfig()
	if !propellerConfig.LiteralOffloadingConfig.Enabled {
		return []v1.EnvVar{}
	}

	envVars := []v1.EnvVar{}
	if propellerConfig.LiteralOffloadingConfig.MinSizeInMBForOffloading > 0 {
		envVars = append(envVars,
			v1.EnvVar{
				Name:  "_F_L_MIN_SIZE_MB",
				Value: strconv.FormatInt(propellerConfig.LiteralOffloadingConfig.MinSizeInMBForOffloading, 10),
			},
		)
	}
	if propellerConfig.LiteralOffloadingConfig.MaxSizeInMBForOffloading > 0 {
		envVars = append(envVars,
			v1.EnvVar{
				Name:  "_F_L_MAX_SIZE_MB",
				Value: strconv.FormatInt(propellerConfig.LiteralOffloadingConfig.MaxSizeInMBForOffloading, 10),
			},
		)
	}
	return envVars
}

func DecorateEnvVars(ctx context.Context, envVars []v1.EnvVar, envFroms []v1.EnvFromSource, taskEnvironmentVariables map[string]string, id pluginsCore.TaskExecutionID, consoleURL string) ([]v1.EnvVar, []v1.EnvFromSource) {
	envVars = append(envVars, GetContextEnvVars(ctx)...)
	envVars = append(envVars, GetExecutionEnvVars(id, consoleURL)...)
	envVars = append(envVars, GetLiteralOffloadingEnvVars()...)

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

	for _, secretName := range config.GetK8sPluginConfig().DefaultEnvFromSecrets {
		optional := true
		secretRef := v1.SecretEnvSource{LocalObjectReference: v1.LocalObjectReference{Name: secretName}, Optional: &optional}
		envFroms = append(envFroms, v1.EnvFromSource{SecretRef: &secretRef})
	}

	for _, cmName := range config.GetK8sPluginConfig().DefaultEnvFromConfigMaps {
		optional := true
		cmRef := v1.ConfigMapEnvSource{LocalObjectReference: v1.LocalObjectReference{Name: cmName}, Optional: &optional}
		envFroms = append(envFroms, v1.EnvFromSource{ConfigMapRef: &cmRef})
	}

	return envVars, envFroms
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
