package core

import (
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

// TaskOverrides interface to expose any overrides that have been set for this task (like resource overrides etc)
type TaskOverrides interface {
	GetResources() *v1.ResourceRequirements
	GetExtendedResources() *core.ExtendedResources
	GetContainerImage() string
	GetConfig() *v1.ConfigMap
}

// TaskExecutionID is a simple Interface to expose the ExecutionID of the running Task
type TaskExecutionID interface {
	// GetGeneratedName returns the computed/generated name for the task id
	// deprecated: use GetGeneratedNameWithLength
	GetGeneratedName() string

	// GetGeneratedNameWith returns the generated name within a bounded length. If the name is smaller than minLength,
	// it'll get right-padded with character '0'. If the name is bigger than maxLength, it'll get hashed to fit within.
	GetGeneratedNameWith(minLength, maxLength int) (string, error)

	// GetID returns the underlying idl task identifier.
	GetID() core.TaskExecutionIdentifier

	// GetUniqueNodeID returns the fully-qualified Node ID that is unique within a
	// given workflow execution.
	GetUniqueNodeID() string

	GetConsoleURL() string
}

// TaskExecutionMetadata represents any execution information for a Task. It is used to communicate meta information about the
// execution or any previously stored information
type TaskExecutionMetadata interface {
	// GetOwnerID returns the owning Kubernetes object
	GetOwnerID() types.NamespacedName
	// GetTaskExecutionID is a specially generated task execution id, that is guaranteed to be unique and consistent for subsequent calls
	GetTaskExecutionID() TaskExecutionID
	GetNamespace() string
	GetOwnerReference() v12.OwnerReference
	GetOverrides() TaskOverrides
	GetLabels() map[string]string
	GetMaxAttempts() uint32
	GetAnnotations() map[string]string
	GetK8sServiceAccount() string
	GetSecurityContext() core.SecurityContext
	IsInterruptible() bool
	GetPlatformResources() *v1.ResourceRequirements
	GetInterruptibleFailureThreshold() int32
	GetEnvironmentVariables() map[string]string
}
