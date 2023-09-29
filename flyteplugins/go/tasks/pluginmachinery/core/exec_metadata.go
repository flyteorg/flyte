package core

import (
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// TaskOverrides interface to expose any overrides that have been set for this task (like resource overrides etc)
type TaskOverrides interface {
	GetResources() *v1.ResourceRequirements
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
	GetInterruptibleFailureThreshold() uint32
	GetEnvironmentVariables() map[string]string
}
