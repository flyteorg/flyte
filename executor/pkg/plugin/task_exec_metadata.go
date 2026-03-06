package plugin

import (
	"google.golang.org/protobuf/proto"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

var _ pluginsCore.TaskExecutionMetadata = &taskExecutionMetadata{}
var _ pluginsCore.TaskExecutionID = &taskExecutionID{}

type taskExecutionID struct {
	generatedName string
	id            core.TaskExecutionIdentifier
}

func (t *taskExecutionID) GetGeneratedName() string {
	return t.generatedName
}

func (t *taskExecutionID) GetGeneratedNameWith(minLength, maxLength int) (string, error) {
	name := t.generatedName
	if len(name) > maxLength {
		name = name[:maxLength]
	}
	return name, nil
}

func (t *taskExecutionID) GetID() *core.TaskExecutionIdentifier {
	return &t.id
}

func (t *taskExecutionID) GetUniqueNodeID() string {
	return t.generatedName
}

type taskExecutionMetadata struct {
	ownerID         types.NamespacedName
	taskExecutionID pluginsCore.TaskExecutionID
	namespace       string
	ownerReference  metav1.OwnerReference
	labels          map[string]string
	annotations     map[string]string
	maxAttempts     uint32
	overrides       pluginsCore.TaskOverrides
	envVars         map[string]string
}

// NewTaskExecutionMetadata creates a TaskExecutionMetadata from a TaskAction.
func NewTaskExecutionMetadata(ta *flyteorgv1.TaskAction) pluginsCore.TaskExecutionMetadata {
	// Extract resource requirements from the inline task template
	overrides := buildOverridesFromTaskTemplate(ta.Spec.TaskTemplate)

	// Build environment variables for the task pod
	envVars := map[string]string{
		"ACTION_NAME": ta.Spec.ActionName,
		"RUN_NAME":    ta.Spec.RunName,
		"_U_ORG_NAME": ta.Spec.Org,
		"_U_RUN_BASE": ta.Spec.RunOutputBase,
	}

	return &taskExecutionMetadata{
		ownerID: types.NamespacedName{
			Name:      ta.Name,
			Namespace: ta.Namespace,
		},
		taskExecutionID: &taskExecutionID{
			generatedName: ta.Name,
			id: core.TaskExecutionIdentifier{
				NodeExecutionId: &core.NodeExecutionIdentifier{
					ExecutionId: &core.WorkflowExecutionIdentifier{
						Project: ta.Spec.Project,
						Domain:  ta.Spec.Domain,
						Name:    ta.Spec.RunName,
						Org:     ta.Spec.Org,
					},
					NodeId: ta.Spec.ActionName,
				},
			},
		},
		namespace: ta.Namespace,
		ownerReference: metav1.OwnerReference{
			APIVersion: flyteorgv1.GroupVersion.String(),
			Kind:       "TaskAction",
			Name:       ta.Name,
			UID:        ta.UID,
		},
		labels:      ta.Labels,
		annotations: ta.Annotations,
		maxAttempts: 1,
		overrides:   overrides,
		envVars:     envVars,
	}
}

// buildOverridesFromTaskTemplate deserializes the task template and extracts resource requirements.
func buildOverridesFromTaskTemplate(data []byte) *taskOverrides {
	if len(data) == 0 {
		return &taskOverrides{}
	}
	tmpl := &core.TaskTemplate{}
	if err := proto.Unmarshal(data, tmpl); err != nil {
		return &taskOverrides{}
	}
	container := tmpl.GetContainer()
	if container == nil {
		return &taskOverrides{}
	}
	res, err := flytek8s.ToK8sResourceRequirements(container.Resources)
	if err != nil {
		return &taskOverrides{}
	}
	return &taskOverrides{resources: res}
}

func (m *taskExecutionMetadata) GetOwnerID() types.NamespacedName            { return m.ownerID }
func (m *taskExecutionMetadata) GetTaskExecutionID() pluginsCore.TaskExecutionID { return m.taskExecutionID }
func (m *taskExecutionMetadata) GetNamespace() string                         { return m.namespace }
func (m *taskExecutionMetadata) GetOwnerReference() metav1.OwnerReference     { return m.ownerReference }
func (m *taskExecutionMetadata) GetLabels() map[string]string                 { return m.labels }
func (m *taskExecutionMetadata) GetAnnotations() map[string]string            { return m.annotations }
func (m *taskExecutionMetadata) GetMaxAttempts() uint32                       { return m.maxAttempts }
func (m *taskExecutionMetadata) GetK8sServiceAccount() string                 { return "" }
func (m *taskExecutionMetadata) IsInterruptible() bool                        { return false }
func (m *taskExecutionMetadata) GetInterruptibleFailureThreshold() int32      { return 0 }
func (m *taskExecutionMetadata) GetEnvironmentVariables() map[string]string   { return m.envVars }
func (m *taskExecutionMetadata) GetConsoleURL() string                        { return "" }

func (m *taskExecutionMetadata) GetOverrides() pluginsCore.TaskOverrides {
	return m.overrides
}

func (m *taskExecutionMetadata) GetSecurityContext() *core.SecurityContext {
	return &core.SecurityContext{}
}

func (m *taskExecutionMetadata) GetPlatformResources() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{}
}

func (m *taskExecutionMetadata) GetExternalResourceAttributes() pluginsCore.ExternalResourceAttributes {
	return pluginsCore.ExternalResourceAttributes{}
}

// taskOverrides provides resource overrides extracted from the task template.
type taskOverrides struct {
	resources *v1.ResourceRequirements
}

func (t *taskOverrides) GetResources() *v1.ResourceRequirements     { return t.resources }
func (t *taskOverrides) GetExtendedResources() *core.ExtendedResources { return nil }
func (t *taskOverrides) GetContainerImage() string                   { return "" }
func (t *taskOverrides) GetConfigMap() *v1.ConfigMap                 { return nil }
func (t *taskOverrides) GetPodTemplate() *core.K8SPod               { return nil }
func (t *taskOverrides) GetConfig() map[string]string                { return nil }
