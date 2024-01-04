package k8s

import (
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils/secrets"
)

const executionIdentityVariable = "execution-identity"

// TaskExecutionContext provides a layer on top of core TaskExecutionContext with a custom TaskExecutionMetadata.
type TaskExecutionContext struct {
	pluginsCore.TaskExecutionContext
	metadataOverride pluginsCore.TaskExecutionMetadata
}

func (t TaskExecutionContext) TaskExecutionMetadata() pluginsCore.TaskExecutionMetadata {
	return t.metadataOverride
}

func newTaskExecutionContext(tCtx pluginsCore.TaskExecutionContext, metadataOverride pluginsCore.TaskExecutionMetadata) TaskExecutionContext {
	return TaskExecutionContext{
		TaskExecutionContext: tCtx,
		metadataOverride:     metadataOverride,
	}
}

// TaskExecutionMetadata provides a layer on top of the core TaskExecutionMetadata with customized annotations and labels
// for k8s plugins.
type TaskExecutionMetadata struct {
	pluginsCore.TaskExecutionMetadata

	annotations map[string]string
	labels      map[string]string
}

func (t TaskExecutionMetadata) GetLabels() map[string]string {
	return t.labels
}

func (t TaskExecutionMetadata) GetAnnotations() map[string]string {
	return t.annotations
}

// newTaskExecutionMetadata creates a TaskExecutionMetadata with secrets serialized as annotations and a label added
// to trigger the flyte pod webhook. If known, the execution identity is injected as a label.
func newTaskExecutionMetadata(tCtx pluginsCore.TaskExecutionMetadata, taskTmpl *core.TaskTemplate) (TaskExecutionMetadata, error) {
	var err error
	secretsMap := make(map[string]string)
	injectLabels := make(map[string]string)
	if taskTmpl.SecurityContext != nil && len(taskTmpl.SecurityContext.Secrets) > 0 {
		secretsMap, err = secrets.MarshalSecretsToMapStrings(taskTmpl.SecurityContext.Secrets)
		if err != nil {
			return TaskExecutionMetadata{}, err
		}

		injectLabels[secrets.PodLabel] = secrets.PodLabelValue
	}

	id := tCtx.GetSecurityContext().RunAs.ExecutionIdentity
	if len(id) > 0 {
		injectLabels[executionIdentityVariable] = id
	}

	return TaskExecutionMetadata{
		TaskExecutionMetadata: tCtx,
		annotations:           utils.UnionMaps(tCtx.GetAnnotations(), secretsMap),
		labels:                utils.UnionMaps(tCtx.GetLabels(), injectLabels),
	}, nil
}
