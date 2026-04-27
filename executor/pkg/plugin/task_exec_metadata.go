package plugin

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	flytesecret "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret"
	pluginsUtils "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/utils/secrets"
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
	interruptible   bool
	securityContext *core.SecurityContext
}

// NewTaskExecutionMetadata creates a TaskExecutionMetadata from a TaskAction.
func NewTaskExecutionMetadata(ta *flyteorgv1.TaskAction) (pluginsCore.TaskExecutionMetadata, error) {
	// Extract resource requirements from the inline task template
	overrides := buildOverridesFromTaskTemplate(ta.Spec.TaskTemplate)

	// Handling secrets
	var err error
	securityContext := extractSecurityContextFromTaskTemplate(ta.Spec.TaskTemplate)
	secretsMap := make(map[string]string)
	injectLabels := map[string]string{
		// Flyte OSS v2 has no organization concept, but the secret fetcher
		// still requires an organization label on the pod. Inject a stable
		// dummy value so scoped secret lookups succeed.
		flytesecret.OrganizationLabel: "flyte",
		flytesecret.ProjectLabel:      ta.Spec.Project,
		flytesecret.DomainLabel:       ta.Spec.Domain,
	}
	if securityContext != nil && len(securityContext.Secrets) > 0 {
		secretsMap, err = secrets.MarshalSecretsToMapStrings(securityContext.Secrets)
		if err != nil {
			return nil, err
		}
		injectLabels[secrets.PodLabel] = secrets.PodLabelValue
	}

	// Build environment variables for the task pod
	envVars := map[string]string{
		"ACTION_NAME": ta.Spec.ActionName,
		"RUN_NAME":    ta.Spec.RunName,
		"_U_RUN_BASE": ta.Spec.RunOutputBase,
		"_U_ORG_NAME": "local",
	}
	for key, value := range ta.Spec.EnvVars {
		if _, exists := envVars[key]; !exists {
			envVars[key] = value
		}
	}
	generatedName := buildGeneratedName(ta)
	retryAttempt := attemptToRetry(ta.Status.Attempts)
	maxAttempts := maxAttemptsFromTaskTemplate(ta.Spec.TaskTemplate)

	return &taskExecutionMetadata{
		ownerID: types.NamespacedName{
			Name:      ta.Name,
			Namespace: ta.Namespace,
		},
		taskExecutionID: &taskExecutionID{
			generatedName: generatedName,
			id: core.TaskExecutionIdentifier{
				NodeExecutionId: &core.NodeExecutionIdentifier{
					ExecutionId: &core.WorkflowExecutionIdentifier{
						Project: ta.Spec.Project,
						Domain:  ta.Spec.Domain,
						Name:    ta.Spec.RunName,
					},
					NodeId: ta.Spec.ActionName,
				},
				RetryAttempt: retryAttempt,
			},
		},
		namespace: ta.Namespace,
		ownerReference: metav1.OwnerReference{
			APIVersion: flyteorgv1.GroupVersion.String(),
			Kind:       "TaskAction",
			Name:       ta.Name,
			UID:        ta.UID,
			Controller: ptr.To(true),
		},
		labels:          pluginsUtils.UnionMaps(ta.Labels, injectLabels),
		annotations:     pluginsUtils.UnionMaps(ta.Annotations, secretsMap),
		maxAttempts:     maxAttempts,
		overrides:       overrides,
		envVars:         envVars,
		interruptible:   ta.Spec.Interruptible != nil && *ta.Spec.Interruptible,
		securityContext: securityContext,
	}, nil
}

func buildGeneratedName(ta *flyteorgv1.TaskAction) string {
	return fmt.Sprintf("%s-%d", ta.Name, attemptToRetry(ta.Status.Attempts))
}

// attemptToRetry convert attempt to retry count
func attemptToRetry(attempt uint32) uint32 {
	if attempt <= 1 {
		return 0
	}
	return attempt - 1
}

// maxAttemptsFromTaskTemplate give the max attempts (retries + 1) from the task template.
func maxAttemptsFromTaskTemplate(data []byte) uint32 {
	if len(data) == 0 {
		return 1
	}

	tmpl := &core.TaskTemplate{}
	if err := proto.Unmarshal(data, tmpl); err != nil {
		return 1
	}

	md := tmpl.GetMetadata()
	if md == nil || md.GetRetries() == nil {
		return 1
	}

	return md.GetRetries().GetRetries() + 1
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

// extractSecurityContextFromTaskTemplate deserializes the task template and extracts SecurityContext.
func extractSecurityContextFromTaskTemplate(data []byte) *core.SecurityContext {
	if len(data) == 0 {
		return &core.SecurityContext{}
	}
	tmpl := &core.TaskTemplate{}
	if err := proto.Unmarshal(data, tmpl); err != nil {
		return &core.SecurityContext{}
	}
	if tmpl.SecurityContext == nil {
		return &core.SecurityContext{}
	}
	return tmpl.GetSecurityContext()
}

func (m *taskExecutionMetadata) GetOwnerID() types.NamespacedName { return m.ownerID }
func (m *taskExecutionMetadata) GetTaskExecutionID() pluginsCore.TaskExecutionID {
	return m.taskExecutionID
}
func (m *taskExecutionMetadata) GetNamespace() string                       { return m.namespace }
func (m *taskExecutionMetadata) GetOwnerReference() metav1.OwnerReference   { return m.ownerReference }
func (m *taskExecutionMetadata) GetLabels() map[string]string               { return m.labels }
func (m *taskExecutionMetadata) GetAnnotations() map[string]string          { return m.annotations }
func (m *taskExecutionMetadata) GetMaxAttempts() uint32                     { return m.maxAttempts }
func (m *taskExecutionMetadata) GetK8sServiceAccount() string               { return "" }
func (m *taskExecutionMetadata) IsInterruptible() bool                      { return m.interruptible }
func (m *taskExecutionMetadata) GetInterruptibleFailureThreshold() int32    { return 0 }
func (m *taskExecutionMetadata) GetEnvironmentVariables() map[string]string { return m.envVars }
func (m *taskExecutionMetadata) GetConsoleURL() string                      { return "" }

func (m *taskExecutionMetadata) GetOverrides() pluginsCore.TaskOverrides {
	return m.overrides
}

func (m *taskExecutionMetadata) GetSecurityContext() *core.SecurityContext {
	return m.securityContext
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

func (t *taskOverrides) GetResources() *v1.ResourceRequirements        { return t.resources }
func (t *taskOverrides) GetExtendedResources() *core.ExtendedResources { return nil }
func (t *taskOverrides) GetContainerImage() string                     { return "" }
func (t *taskOverrides) GetConfigMap() *v1.ConfigMap                   { return nil }
func (t *taskOverrides) GetPodTemplate() *core.K8SPod                  { return nil }
func (t *taskOverrides) GetConfig() map[string]string                  { return nil }
