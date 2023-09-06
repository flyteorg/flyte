package awsbatch

import (
	"context"
	"regexp"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	pluginErrors "github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/awsbatch/config"
	arrayCore "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/core"
	awsUtils "github.com/flyteorg/flyteplugins/go/tasks/plugins/awsutils"
	"github.com/flyteorg/flytestdlib/errors"
	"github.com/flyteorg/flytestdlib/logger"

	pluginCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/awsbatch/definition"
)

const defaultComputeEngine = "EC2"
const platformCapabilitiesConfigKey = "platformCapabilities"

func getContainerImage(_ context.Context, task *core.TaskTemplate) string {
	if task.GetContainer() != nil && len(task.GetContainer().Image) > 0 {
		return task.GetContainer().Image
	}

	return ""
}

var urlRegex = regexp.MustCompile(`^(?:([^/]+)/)?(?:([^/]+)/)*?([^@:/]+)(?:[@:][^/]+)?$`)

// Gets the repository part of the container image url
func containerImageRepository(containerImage string) string {
	parts := urlRegex.FindAllStringSubmatch(containerImage, -1)
	if len(parts) > 0 && len(parts[0]) > 3 {
		return parts[0][3]
	}

	return ""
}

func EnsureJobDefinition(ctx context.Context, tCtx pluginCore.TaskExecutionContext, cfg *config.Config, client Client,
	definitionCache definition.Cache, currentState *State, terminalVersion uint32) (nextState *State, err error) {

	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return nil, err
	}

	containerImage := getContainerImage(ctx, taskTemplate)
	if len(containerImage) == 0 {
		logger.Infof(ctx, "Future task doesn't have an image specified. Failing.")
		return nil, errors.Errorf(pluginErrors.BadTaskSpecification, "Tasktemplate does not contain a container image.")
	}

	role := awsUtils.GetRoleFromSecurityContext(cfg.RoleAnnotationKey, tCtx.TaskExecutionMetadata())
	platformCapabilities := taskTemplate.GetConfig()[platformCapabilitiesConfigKey]
	if len(platformCapabilities) == 0 {
		platformCapabilities = defaultComputeEngine
	}

	cacheKey := definition.NewCacheKey(role, containerImage, platformCapabilities)
	if existingArn, found := definitionCache.Get(cacheKey); found {
		logger.Infof(ctx, "Found an existing job definition for Image [%v], Role [%v], JobDefinitionInput [%v]. Arn [%v]",
			containerImage, role, platformCapabilities, existingArn)

		nextState = currentState.SetJobDefinitionArn(existingArn)
		nextState.State = nextState.SetPhase(arrayCore.PhaseLaunch, terminalVersion).SetReason("AWS job definition already exist.")
		return nextState, nil
	}

	name := definition.GetJobDefinitionSafeName(containerImageRepository(containerImage))

	arn, err := client.RegisterJobDefinition(ctx, name, containerImage, role, platformCapabilities)
	if err != nil {
		return currentState, err
	}

	err = definitionCache.Put(cacheKey, arn)
	if err != nil {
		logger.Warnf(ctx, "Failed to store job definition arn in cache. Will continue with the registered arn [%v]. Error: %v",
			arn, err)
	}

	nextState = currentState.SetJobDefinitionArn(arn)
	nextState.State = nextState.SetPhase(arrayCore.PhaseLaunch, terminalVersion).SetReason("Created AWS job definition")

	return nextState, nil
}
