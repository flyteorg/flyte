package flytek8s

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
)

// Wraps a regular TaskExecutionMetadata and overrides the IsInterruptible method to always return false
// This is useful as the runner and the scheduler pods should never be interruptible
type NonInterruptibleTaskExecutionMetadata struct {
	metadata pluginsCore.TaskExecutionMetadata
}

func (n NonInterruptibleTaskExecutionMetadata) GetOwnerID() types.NamespacedName {
	return n.metadata.GetOwnerID()
}

func (n NonInterruptibleTaskExecutionMetadata) GetTaskExecutionID() pluginsCore.TaskExecutionID {
	return n.metadata.GetTaskExecutionID()
}

func (n NonInterruptibleTaskExecutionMetadata) GetNamespace() string {
	return n.metadata.GetNamespace()
}

func (n NonInterruptibleTaskExecutionMetadata) GetOwnerReference() metav1.OwnerReference {
	return n.metadata.GetOwnerReference()
}

func (n NonInterruptibleTaskExecutionMetadata) GetOverrides() pluginsCore.TaskOverrides {
	return n.metadata.GetOverrides()
}

func (n NonInterruptibleTaskExecutionMetadata) GetLabels() map[string]string {
	return n.metadata.GetLabels()
}

func (n NonInterruptibleTaskExecutionMetadata) GetMaxAttempts() uint32 {
	return n.metadata.GetMaxAttempts()
}

func (n NonInterruptibleTaskExecutionMetadata) GetAnnotations() map[string]string {
	return n.metadata.GetAnnotations()
}

func (n NonInterruptibleTaskExecutionMetadata) GetK8sServiceAccount() string {
	return n.metadata.GetK8sServiceAccount()
}

func (n NonInterruptibleTaskExecutionMetadata) GetSecurityContext() core.SecurityContext {
	return n.metadata.GetSecurityContext()
}

func (n NonInterruptibleTaskExecutionMetadata) GetPlatformResources() *v1.ResourceRequirements {
	return n.metadata.GetPlatformResources()
}

func (n NonInterruptibleTaskExecutionMetadata) GetInterruptibleFailureThreshold() int32 {
	return n.metadata.GetInterruptibleFailureThreshold()
}

func (n NonInterruptibleTaskExecutionMetadata) GetEnvironmentVariables() map[string]string {
	return n.metadata.GetEnvironmentVariables()
}

func (n NonInterruptibleTaskExecutionMetadata) IsInterruptible() bool {
	return false
}

// A wrapper around a regular TaskExecutionContext allowing to inject a custom TaskExecutionMetadata which is
// non-interruptible
type NonInterruptibleTaskExecutionContext struct {
	pluginsCore.TaskExecutionContext
	metadata NonInterruptibleTaskExecutionMetadata
}

func (n NonInterruptibleTaskExecutionContext) TaskExecutionMetadata() pluginsCore.TaskExecutionMetadata {
	return n.metadata
}

func NewNonInterruptibleTaskExecutionContext(ctx pluginsCore.TaskExecutionContext) NonInterruptibleTaskExecutionContext {
	return NonInterruptibleTaskExecutionContext{
		TaskExecutionContext: ctx,
		metadata: NonInterruptibleTaskExecutionMetadata{
			metadata: ctx.TaskExecutionMetadata(),
		},
	}
}
