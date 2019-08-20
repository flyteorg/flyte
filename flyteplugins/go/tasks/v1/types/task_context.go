package types

import (
	"github.com/lyft/flytestdlib/storage"
	"k8s.io/apimachinery/pkg/types"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	typesv1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockery -name TaskContext

// Interface to expose any overrides that have been set for this task (like resource overrides etc)
type TaskOverrides interface {
	GetResources() *typesv1.ResourceRequirements
	GetConfig() *typesv1.ConfigMap
}

// Simple Interface to expose the ExecutionID of the running Task
type TaskExecutionID interface {
	GetGeneratedName() string
	GetID() core.TaskExecutionIdentifier
}

// TaskContext represents any execution information for a Task. It is used to communicate meta information about the
// execution or any previously stored information
type TaskContext interface {
	GetOwnerID() types.NamespacedName
	GetTaskExecutionID() TaskExecutionID
	GetDataDir() storage.DataReference
	GetInputsFile() storage.DataReference
	GetOutputsFile() storage.DataReference
	GetErrorFile() storage.DataReference
	GetNamespace() string
	GetOwnerReference() metaV1.OwnerReference
	GetOverrides() TaskOverrides
	GetLabels() map[string]string
	GetAnnotations() map[string]string
	GetCustomState() CustomState
	GetK8sServiceAccount() string
	GetPhase() TaskPhase
	GetPhaseVersion() uint32
}
