package awsbatch

import (
	"context"
	"regexp"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	pluginErrors "github.com/lyft/flyteplugins/go/tasks/errors"
	"github.com/lyft/flyteplugins/go/tasks/plugins/array/awsbatch/config"
	arrayCore "github.com/lyft/flyteplugins/go/tasks/plugins/array/core"
	"github.com/lyft/flytestdlib/errors"
	"github.com/lyft/flytestdlib/logger"

	pluginCore "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/lyft/flyteplugins/go/tasks/plugins/array/awsbatch/definition"
)

func getContainerImage(_ context.Context, task *core.TaskTemplate) string {
	if task.GetContainer() != nil && len(task.GetContainer().Image) > 0 {
		return task.GetContainer().Image
	}

	return ""
}

func getRole(_ context.Context, roleAnnotationKey string, annotations map[string]string) string {
	if len(roleAnnotationKey) > 0 {
		return annotations[roleAnnotationKey]
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
	definitionCache definition.Cache, currentState *State) (nextState *State, err error) {

	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return nil, err
	}

	containerImage := getContainerImage(ctx, taskTemplate)
	if len(containerImage) == 0 {
		logger.Infof(ctx, "Future task doesn't have an image specified. Failing.")
		return nil, errors.Errorf(pluginErrors.BadTaskSpecification, "Tasktemplate does not contain a container image.")
	}

	role := getRole(ctx, cfg.RoleAnnotationKey, tCtx.TaskExecutionMetadata().GetAnnotations())

	cacheKey := definition.NewCacheKey(role, containerImage)
	if existingArn, found := definitionCache.Get(cacheKey); found {
		logger.Infof(ctx, "Found an existing job definition for Image [%v] and Role [%v]. Arn [%v]",
			containerImage, role, existingArn)

		nextState = currentState.SetJobDefinitionArn(existingArn)
		nextState.State = nextState.SetPhase(arrayCore.PhaseLaunch, 0).SetReason("AWS job definition already exist.")
		return nextState, nil
	}

	name := definition.GetJobDefinitionSafeName(containerImageRepository(containerImage))

	arn, err := client.RegisterJobDefinition(ctx, name, containerImage, role)
	if err != nil {
		return currentState, err
	}

	err = definitionCache.Put(cacheKey, arn)
	if err != nil {
		logger.Warnf(ctx, "Failed to store job definition arn in cache. Will continue with the registered arn [%v]. Error: %v",
			arn, err)
	}

	nextState = currentState.SetJobDefinitionArn(arn)
	nextState.State = nextState.SetPhase(arrayCore.PhaseLaunch, 0).SetReason("Created AWS job definition")

	return nextState, nil
}
