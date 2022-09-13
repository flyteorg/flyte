package pod

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteplugins/go/tasks/errors"
	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s"

	v1 "k8s.io/api/core/v1"
)

const (
	ContainerTaskType = "container"
)

type containerPodBuilder struct {
}

func (containerPodBuilder) buildPodSpec(ctx context.Context, task *core.TaskTemplate, taskCtx pluginsCore.TaskExecutionContext) (*v1.PodSpec, error) {
	podSpec, err := flytek8s.ToK8sPodSpec(ctx, taskCtx)
	if err != nil {
		return nil, err
	}

	return podSpec, nil
}

func (containerPodBuilder) getPrimaryContainerName(task *core.TaskTemplate, taskCtx pluginsCore.TaskExecutionContext) (string, error) {
	primaryContainerName := taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()
	if primaryContainerName == "" {
		return "", errors.Errorf(errors.BadTaskSpecification, "invalid TaskSpecification, missing generated name")
	}
	return primaryContainerName, nil
}

func (containerPodBuilder) updatePodMetadata(ctx context.Context, pod *v1.Pod, task *core.TaskTemplate, taskCtx pluginsCore.TaskExecutionContext) error {
	return nil
}
