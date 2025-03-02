package v1alpha1

import (
	"time"

	typesv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/utils"
)

type OutputVarMap struct {
	*core.VariableMap
}

func (in *OutputVarMap) MarshalJSON() ([]byte, error) {
	return utils.MarshalPbToBytes(in.VariableMap)
}

func (in *OutputVarMap) UnmarshalJSON(b []byte) error {
	in.VariableMap = &core.VariableMap{}
	return utils.UnmarshalBytesToPb(b, in.VariableMap)
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OutputVarMap) DeepCopyInto(out *OutputVarMap) {
	*out = *in
	// We do not manipulate the object, so its ok
	// Once we figure out the autogenerate story we can replace this
}

type Binding struct {
	*core.Binding
}

func (in *Binding) MarshalJSON() ([]byte, error) {
	return utils.MarshalPbToBytes(in.Binding)
}

func (in *Binding) UnmarshalJSON(b []byte) error {
	in.Binding = &core.Binding{}
	return utils.UnmarshalBytesToPb(b, in.Binding)
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Binding) DeepCopyInto(out *Binding) {
	*out = *in
	// We do not manipulate the object, so its ok
	// Once we figure out the autogenerate story we can replace this
}

type ExponentialBackoff struct {
	MaxExponent uint32       `json:"maxExponent"`
	Max         *v1.Duration `json:"max"`
}

type RetryOnOOM struct {
	Factor  float32             `json:"factor"`
	Limit   string              `json:"limit"`
	Backoff *ExponentialBackoff `json:"backoff,omitempty"`
}

func (ro *RetryOnOOM) GetBackoff() *ExponentialBackoff {
	return ro.Backoff
}

// Strategy to be used to Retry a node that is in RetryableFailure state
type RetryStrategy struct {
	OnOOM *RetryOnOOM `json:"onOOM,omitempty"`
	// MinAttempts implies the at least n attempts to try this node before giving up. The at least here is because we may
	// fail to write the attempt information and end up retrying again.
	// Also `0` and `1` both mean at least one attempt will be done. 0 is a degenerate case.
	MinAttempts *int `json:"minAttempts"`
	// TODO Add retrydelay?
}

func (rs *RetryStrategy) GetOnOOM() *RetryOnOOM {
	return rs.OnOOM
}

type Alias struct {
	core.Alias
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Alias) DeepCopyInto(out *Alias) {
	*out = *in
	// We do not manipulate the object, so its ok
	// Once we figure out the autogenerate story we can replace this
}

type NodeMetadata struct {
	core.NodeMetadata
}

func (in *NodeMetadata) DeepCopyInto(out *NodeMetadata) {
	*out = *in
	// We do not manipulate the object, so its ok
	// Once we figure out the autogenerate story we can replace this
}

type ExtendedResources struct {
	*core.ExtendedResources
}

func (in *ExtendedResources) MarshalJSON() ([]byte, error) {
	return utils.MarshalPbToBytes(in.ExtendedResources)
}

func (in *ExtendedResources) UnmarshalJSON(b []byte) error {
	in.ExtendedResources = &core.ExtendedResources{}
	return utils.UnmarshalBytesToPb(b, in.ExtendedResources)
}

func (in *ExtendedResources) DeepCopyInto(out *ExtendedResources) {
	*out = *in
	// We do not manipulate the object, so its ok
	// Once we figure out the autogenerate story we can replace this
}

type NodeSpec struct {
	ID            NodeID                        `json:"id"`
	Name          string                        `json:"name,omitempty"`
	Resources     *typesv1.ResourceRequirements `json:"resources,omitempty"`
	Kind          NodeKind                      `json:"kind"`
	BranchNode    *BranchNodeSpec               `json:"branch,omitempty"`
	TaskRef       *TaskID                       `json:"task,omitempty"`
	WorkflowNode  *WorkflowNodeSpec             `json:"workflow,omitempty"`
	GateNode      *GateNodeSpec                 `json:"gate,omitempty"`
	ArrayNode     *ArrayNodeSpec                `json:"array,omitempty"`
	InputBindings []*Binding                    `json:"inputBindings,omitempty"`
	Config        *typesv1.ConfigMap            `json:"config,omitempty"`
	RetryStrategy *RetryStrategy                `json:"retry,omitempty"`
	OutputAliases []Alias                       `json:"outputAlias,omitempty"`

	// SecurityContext holds pod-level security attributes and common container settings.
	// Optional: Defaults to empty.  See type description for default values of each field.
	// +optional
	SecurityContext *typesv1.PodSecurityContext `json:"securityContext,omitempty" protobuf:"bytes,14,opt,name=securityContext"`
	// ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
	// If specified, these secrets will be passed to individual puller implementations for them to use. For example,
	// in the case of docker, only DockerConfig type secrets are honored.
	// More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	ImagePullSecrets []typesv1.LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,15,rep,name=imagePullSecrets"`
	// Specifies the hostname of the Pod
	// If not specified, the pod's hostname will be set to a system-defined value.
	// +optional
	Hostname string `json:"hostname,omitempty" protobuf:"bytes,16,opt,name=hostname"`
	// If specified, the fully qualified Pod hostname will be "<hostname>.<subdomain>.<pod namespace>.svc.<cluster domain>".
	// If not specified, the pod will not have a domainname at all.
	// +optional
	Subdomain string `json:"subdomain,omitempty" protobuf:"bytes,17,opt,name=subdomain"`
	// If specified, the pod's scheduling constraints
	// +optional
	Affinity *typesv1.Affinity `json:"affinity,omitempty" protobuf:"bytes,18,opt,name=affinity"`
	// If specified, the pod will be dispatched by specified scheduler.
	// If not specified, the pod will be dispatched by default scheduler.
	// +optional
	SchedulerName string `json:"schedulerName,omitempty" protobuf:"bytes,19,opt,name=schedulerName"`
	// If specified, includes overrides for extended resources to allocate to the
	// node.
	// +optional
	ExtendedResources *ExtendedResources `json:"extendedResources,omitempty" protobuf:"bytes,20,opt,name=extendedResources"`
	// If specified, the pod's tolerations.
	// +optional
	Tolerations []typesv1.Toleration `json:"tolerations,omitempty" protobuf:"bytes,22,opt,name=tolerations"`
	// Node execution timeout
	ExecutionDeadline *v1.Duration `json:"executionDeadline,omitempty"`
	// StartTime before the system will actively try to mark it failed and kill associated containers.
	// Value must be a positive integer. This includes time spent waiting in the queue.
	// +optional
	ActiveDeadline *v1.Duration `json:"activeDeadline,omitempty"`
	// The value set to True means task is OK with getting interrupted
	// +optional
	Interruptible *bool `json:"interruptible,omitempty"`

	ContainerImage string `json:"containerImage,omitempty"`

	PodTemplate *core.K8SPod `json:"podTemplate,omitempty" protobuf:"bytes,23,opt,name=podTemplate"`
}

func (in *NodeSpec) GetName() string {
	return in.Name
}

func (in *NodeSpec) GetRetryStrategy() *RetryStrategy {
	return in.RetryStrategy
}

func (in *NodeSpec) GetExecutionDeadline() *time.Duration {
	if in.ExecutionDeadline != nil {
		return &in.ExecutionDeadline.Duration
	}
	return nil
}

func (in *NodeSpec) GetActiveDeadline() *time.Duration {
	if in.ActiveDeadline != nil {
		return &in.ActiveDeadline.Duration
	}
	return nil
}

func (in *NodeSpec) IsInterruptible() *bool {
	return in.Interruptible
}

func (in *NodeSpec) GetConfig() *typesv1.ConfigMap {
	return in.Config
}

func (in *NodeSpec) GetResources() *typesv1.ResourceRequirements {
	return in.Resources
}

func (in *NodeSpec) GetExtendedResources() *core.ExtendedResources {
	if in.ExtendedResources == nil {
		return nil
	}
	return in.ExtendedResources.ExtendedResources
}

func (in *NodeSpec) GetOutputAlias() []Alias {
	return in.OutputAliases
}

// In functions below, explicitly strip out nil type information because NodeSpec's WorkflowNode is a struct type,
// not interface and downstream nil checks will not pass.
// See the test in TestPointersForNodeSpec for more information.

func (in *NodeSpec) GetWorkflowNode() ExecutableWorkflowNode {
	if in != nil {
		if in.WorkflowNode == nil {
			return nil
		}
		return in.WorkflowNode
	}
	return nil
}

func (in *NodeSpec) GetBranchNode() ExecutableBranchNode {
	if in.BranchNode == nil {
		return nil
	}
	return in.BranchNode
}

func (in *NodeSpec) GetGateNode() ExecutableGateNode {
	if in.GateNode == nil {
		return nil
	}
	return in.GateNode
}

func (in *NodeSpec) GetArrayNode() ExecutableArrayNode {
	if in.ArrayNode == nil {
		return nil
	}
	return in.ArrayNode
}

func (in *NodeSpec) GetTaskID() *TaskID {
	return in.TaskRef
}

func (in *NodeSpec) GetKind() NodeKind {
	return in.Kind
}

func (in *NodeSpec) GetID() NodeID {
	return in.ID
}

func (in *NodeSpec) IsStartNode() bool {
	return in.ID == StartNodeID
}

func (in *NodeSpec) IsEndNode() bool {
	return in.ID == EndNodeID
}

func (in *NodeSpec) GetInputBindings() []*Binding {
	return in.InputBindings
}

func (in *NodeSpec) GetContainerImage() string {
	return in.ContainerImage
}

func (in *NodeSpec) GetPodTemplate() *core.K8SPod {
	return in.PodTemplate
}
