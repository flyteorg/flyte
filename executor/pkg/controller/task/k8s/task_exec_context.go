package k8s

import (
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils/secrets"
	k8sUtils "github.com/flyteorg/flyte/flytepropeller/pkg/utils"
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
	if taskTmpl.GetSecurityContext() != nil && len(taskTmpl.GetSecurityContext().GetSecrets()) > 0 {
		secretsMap, err = secrets.MarshalSecretsToMapStrings(taskTmpl.GetSecurityContext().GetSecrets())
		if err != nil {
			return TaskExecutionMetadata{}, err
		}

		injectLabels[secrets.PodLabel] = secrets.PodLabelValue
	}

	id := tCtx.GetSecurityContext().RunAs.GetExecutionIdentity() //nolint:protogetter
	if len(id) > 0 {
		sanitizedID := k8sUtils.SanitizeLabelValue(id)
		injectLabels[executionIdentityVariable] = sanitizedID
	}

	return TaskExecutionMetadata{
		TaskExecutionMetadata: tCtx,
		annotations:           utils.UnionMaps(tCtx.GetAnnotations(), secretsMap),
		labels:                utils.UnionMaps(tCtx.GetLabels(), injectLabels),
	}, nil
}
