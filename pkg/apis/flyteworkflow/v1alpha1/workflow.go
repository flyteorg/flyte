package v1alpha1

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/flyteorg/flytestdlib/storage"

	"k8s.io/apimachinery/pkg/types"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/golang/protobuf/jsonpb"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Defines a non-configurable keyspace size for shard keys. This needs to be a small value because we use label
// selectors to define shard key ranges which do not support range queries. It should only be modified increasingly
// to ensure backward compatibility.
const ShardKeyspaceSize = 32

const StartNodeID = "start-node"
const EndNodeID = "end-node"

type WorkflowDefinitionVersion uint32

var LatestWorkflowDefinitionVersion = WorkflowDefinitionVersion1

const (
	WorkflowDefinitionVersion0 WorkflowDefinitionVersion = iota
	WorkflowDefinitionVersion1
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FlyteWorkflow: represents one Execution Workflow object
type FlyteWorkflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	*WorkflowSpec     `json:"spec"`
	WorkflowMeta      *WorkflowMeta                `json:"workflowMeta,omitempty"`
	Inputs            *Inputs                      `json:"inputs,omitempty"`
	ExecutionID       ExecutionID                  `json:"executionId"`
	Tasks             map[TaskID]*TaskSpec         `json:"tasks"`
	SubWorkflows      map[WorkflowID]*WorkflowSpec `json:"subWorkflows,omitempty"`
	// StartTime before the system will actively try to mark it failed and kill associated containers.
	// Value must be a positive integer.
	// +optional
	ActiveDeadlineSeconds *int64 `json:"activeDeadlineSeconds,omitempty"`
	// Defaults value of parameters to be used for nodes if not set by the node.
	NodeDefaults NodeDefaults `json:"node-defaults,omitempty"`
	// Specifies the time when the workflow has been accepted into the system.
	AcceptedAt *metav1.Time `json:"acceptedAt,omitempty"`
	// [DEPRECATED] ServiceAccountName is the name of the ServiceAccount to use to run this pod.
	// [DEPRECATED] More info: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/
	// [DEPRECATED] +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty" protobuf:"bytes,8,opt,name=serviceAccountName"`
	// Security context fields to define privilege and access control settings
	// +optional
	SecurityContext core.SecurityContext `json:"securityContext,omitempty" protobuf:"bytes,12,rep,name=securityContext"`
	// Status is the only mutable section in the workflow. It holds all the execution information
	Status WorkflowStatus `json:"status,omitempty"`
	// RawOutputDataConfig defines the configurations to use for generating raw outputs (e.g. blobs, schemas).
	RawOutputDataConfig RawOutputDataConfig `json:"rawOutputDataConfig,omitempty"`
	// Workflow-execution specifications and overrides
	ExecutionConfig ExecutionConfig `json:"executionConfig,omitempty"`

	// non-Serialized fields (these will not get written to etcd)
	// As of 2020-07, the only real implementation of this interface is a URLPathConstructor, which is just an empty
	// struct. However, because this field is an interface, we create it once when the crd is hydrated from etcd,
	// so that it can be used downstream without any confusion.
	// This field is here because it's easier to put it here than pipe through a new object through all of propeller.
	DataReferenceConstructor storage.ReferenceConstructor `json:"-"`

	// WorkflowClosureReference is the location containing an offloaded WorkflowClosure. This is used to offload
	// portions of the CRD to an external data store to reduce CRD size. If this exists, FlytePropeller must retrieve
	// and parse the static data prior to processing.
	WorkflowClosureReference DataReference `json:"workflowClosureReference,omitempty"`
}

func (in *FlyteWorkflow) GetSecurityContext() core.SecurityContext {
	return in.SecurityContext
}

func (in *FlyteWorkflow) GetEventVersion() EventVersion {
	if in.WorkflowMeta != nil {
		return in.WorkflowMeta.EventVersion
	}
	return EventVersion0
}

func (in *FlyteWorkflow) GetDefinitionVersion() WorkflowDefinitionVersion {
	if in.Status.DefinitionVersion != nil {
		return *in.Status.DefinitionVersion
	}

	return WorkflowDefinitionVersion0
}

func (in *FlyteWorkflow) GetExecutionConfig() ExecutionConfig {
	return in.ExecutionConfig
}

type WorkflowMeta struct {
	EventVersion EventVersion `json:"eventVersion,omitempty"`
}

type EventVersion int

const (
	EventVersion0 EventVersion = iota
	EventVersion1
	EventVersion2
)

type NodeDefaults struct {
	// Default behaviour for Interruptible for nodes unless explicitly set at the node level.
	Interruptible bool `json:"interruptible,omitempty"`
}

var FlyteWorkflowGVK = SchemeGroupVersion.WithKind(FlyteWorkflowKind)

func (in *FlyteWorkflow) GetOwnerReference() metav1.OwnerReference {
	// TODO Open Issue - https://github.com/kubernetes/client-go/issues/308
	// For some reason the CRD does not have the GVK correctly populated. So we will fake it.
	if len(in.GroupVersionKind().Group) == 0 || len(in.GroupVersionKind().Kind) == 0 || len(in.GroupVersionKind().Version) == 0 {
		return *metav1.NewControllerRef(in, FlyteWorkflowGVK)
	}
	return *metav1.NewControllerRef(in, in.GroupVersionKind())
}

func (in *FlyteWorkflow) GetTask(id TaskID) (ExecutableTask, error) {
	t, ok := in.Tasks[id]
	if !ok {
		return nil, errors.Errorf("Unable to find task with Id [%v]", id)
	}
	return t, nil
}

func (in *FlyteWorkflow) GetExecutionStatus() ExecutableWorkflowStatus {
	s := &in.Status
	s.DataReferenceConstructor = in.DataReferenceConstructor
	return s
}

func (in *FlyteWorkflow) GetK8sWorkflowID() types.NamespacedName {
	return types.NamespacedName{
		Name:      in.GetName(),
		Namespace: in.GetNamespace(),
	}
}

func (in *FlyteWorkflow) GetExecutionID() ExecutionID {
	return in.ExecutionID
}

func (in *FlyteWorkflow) FindSubWorkflow(subID WorkflowID) ExecutableSubWorkflow {
	s, ok := in.SubWorkflows[subID]
	if !ok {
		return nil
	}
	return s
}

func (in *FlyteWorkflow) GetNodeExecutionStatus(ctx context.Context, id NodeID) ExecutableNodeStatus {
	return in.GetExecutionStatus().GetNodeExecutionStatus(ctx, id)
}

func (in *FlyteWorkflow) GetServiceAccountName() string {
	return in.ServiceAccountName
}

func (in *FlyteWorkflow) IsInterruptible() bool {
	// use execution config override if set (can enable/disable interruptible flag for a single execution)
	if in.ExecutionConfig.Interruptible != nil {
		return *in.ExecutionConfig.Interruptible
	}

	// fall back to node defaults if no override was provided
	return in.NodeDefaults.Interruptible
}

func (in *FlyteWorkflow) GetRawOutputDataConfig() RawOutputDataConfig {
	return in.RawOutputDataConfig
}

type Inputs struct {
	*core.LiteralMap
}

func (in *Inputs) UnmarshalJSON(b []byte) error {
	in.LiteralMap = &core.LiteralMap{}
	return jsonpb.Unmarshal(bytes.NewReader(b), in.LiteralMap)
}

func (in *Inputs) MarshalJSON() ([]byte, error) {
	if in == nil || in.LiteralMap == nil {
		return nilJSON, nil
	}

	var buf bytes.Buffer
	if err := marshaler.Marshal(&buf, in.LiteralMap); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Inputs) DeepCopyInto(out *Inputs) {
	*out = *in
	// We do not manipulate the object, so its ok
	// Once we figure out the autogenerate story we can replace this
}

// Deprecated: Please use Connections instead
type DeprecatedConnections struct {
	DownstreamEdges map[NodeID][]NodeID
	UpstreamEdges   map[NodeID][]NodeID
}

func (in *DeprecatedConnections) UnmarshalJSON(b []byte) error {
	in.DownstreamEdges = map[NodeID][]NodeID{}
	err := json.Unmarshal(b, &in.DownstreamEdges)
	if err != nil {
		return err
	}
	in.UpstreamEdges = map[NodeID][]NodeID{}
	for from, nodes := range in.DownstreamEdges {
		for _, to := range nodes {
			if _, ok := in.UpstreamEdges[to]; !ok {
				in.UpstreamEdges[to] = []NodeID{}
			}
			in.UpstreamEdges[to] = append(in.UpstreamEdges[to], from)
		}
	}
	return nil
}

func (in *DeprecatedConnections) MarshalJSON() ([]byte, error) {
	return json.Marshal(in.DownstreamEdges)
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DeprecatedConnections) DeepCopyInto(out *DeprecatedConnections) {
	*out = *in
	// We do not manipulate the object, so its ok
	// Once we figure out the autogenerate story we can replace this
}

// Connections keep track of downstream and upstream dependencies (including data and execution dependencies).
type Connections struct {
	Downstream map[NodeID][]NodeID `json:"downstream"`
	Upstream   map[NodeID][]NodeID `json:"upstream"`
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Connections) DeepCopyInto(out *Connections) {
	*out = *in
	// We do not manipulate the object, so its ok
	// Once we figure out the autogenerate story we can replace this
}

// WorkflowSpec is the spec for the actual Flyte Workflow (DAG)
type WorkflowSpec struct {
	ID    WorkflowID           `json:"id"`
	Nodes map[NodeID]*NodeSpec `json:"nodes"`

	// Defines the set of connections (both data dependencies and execution dependencies) that the graph is
	// formed of. The execution engine will respect and follow these connections as it determines which nodes
	// can and should be executed.
	// Deprecated: Please use Connections
	DeprecatedConnections DeprecatedConnections `json:"connections"`

	// Defines the set of connections (both data dependencies and execution dependencies) that the graph is
	// formed of. The execution engine will respect and follow these connections as it determines which nodes
	// can and should be executed.
	Connections Connections `json:"edges"`

	// Defines a single node to execute in case the system determined the Workflow has failed.
	OnFailure *NodeSpec `json:"onFailure,omitempty"`

	// Defines the declaration of the outputs types and names this workflow is expected to generate.
	Outputs *OutputVarMap `json:"outputs,omitempty"`

	// Defines the data links used to construct the final outputs of the workflow. Bindings will typically
	// refer to specific outputs of a subset of the nodes executed in the Workflow. When executing the end-node,
	// the execution engine will traverse these bindings and assemble the final set of outputs of the workflow.
	OutputBindings []*Binding `json:"outputBindings,omitempty"`

	// Defines the policy for handling failures whether it's to fail immediately, or let the nodes run
	// to completion.
	OnFailurePolicy WorkflowOnFailurePolicy `json:"onFailurePolicy,omitempty"`
}

func (in *WorkflowSpec) StartNode() ExecutableNode {
	n, ok := in.Nodes[StartNodeID]
	if !ok {
		return nil
	}
	return n
}

func (in *WorkflowSpec) GetID() WorkflowID {
	return in.ID
}

func (in *WorkflowSpec) ToNode(name NodeID) ([]NodeID, error) {
	if _, ok := in.Nodes[name]; !ok {
		return nil, errors.Errorf("Bad Node [%v], is not defined in the Workflow [%v]", name, in.ID)
	}
	upstreamNodes := in.GetConnections().Upstream[name]
	return upstreamNodes, nil
}

func (in *WorkflowSpec) FromNode(name NodeID) ([]NodeID, error) {
	if _, ok := in.Nodes[name]; !ok {
		return nil, errors.Errorf("Bad Node [%v], is not defined in the Workflow [%v]", name, in.ID)
	}

	downstreamNodes := in.GetConnections().Downstream[name]
	return downstreamNodes, nil
}

func (in *WorkflowSpec) GetOnFailurePolicy() WorkflowOnFailurePolicy {
	return in.OnFailurePolicy
}

func (in *WorkflowSpec) GetOutputs() *OutputVarMap {
	return in.Outputs
}

func (in *WorkflowSpec) GetNode(nodeID NodeID) (ExecutableNode, bool) {
	n, ok := in.Nodes[nodeID]
	return n, ok
}

func (in *WorkflowSpec) GetConnections() *Connections {
	// For backward compatibility, if the new Connections field is not yet populated then copy the connections from the
	// deprecated field. This will happen in one of two cases:
	//  1. If an old Admin generated the CRD
	//  2. If new propeller is deployed and is unmarshalling an old CRD.
	if len(in.Connections.Upstream) == 0 && len(in.Connections.Downstream) == 0 {
		in.Connections.Upstream = in.DeprecatedConnections.UpstreamEdges
		in.Connections.Downstream = in.DeprecatedConnections.DownstreamEdges
	}

	return &in.Connections
}

func (in *WorkflowSpec) GetOutputBindings() []*Binding {
	return in.OutputBindings
}

func (in *WorkflowSpec) GetOnFailureNode() ExecutableNode {
	if in.OnFailure == nil {
		return nil
	}
	return in.OnFailure
}

func (in *WorkflowSpec) GetNodes() []NodeID {
	nodeIds := make([]NodeID, 0, len(in.Nodes))
	for id := range in.Nodes {
		nodeIds = append(nodeIds, id)
	}
	return nodeIds
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// FlyteWorkflowList is a list of FlyteWorkflow resources
type FlyteWorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []FlyteWorkflow `json:"items"`
}
