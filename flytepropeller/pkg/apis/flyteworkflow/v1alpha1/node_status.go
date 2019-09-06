package v1alpha1

import (
	"encoding/json"
	"reflect"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteplugins/go/tasks/v1/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BranchNodeStatus struct {
	Phase           BranchNodePhase `json:"phase"`
	FinalizedNodeID *NodeID         `json:"finalNodeId"`
}

func (in *BranchNodeStatus) GetPhase() BranchNodePhase {
	return in.Phase
}

func (in *BranchNodeStatus) SetBranchNodeError() {
	in.Phase = BranchNodeError
}

func (in *BranchNodeStatus) SetBranchNodeSuccess(id NodeID) {
	in.Phase = BranchNodeSuccess
	in.FinalizedNodeID = &id
}

func (in *BranchNodeStatus) GetFinalizedNode() *NodeID {
	return in.FinalizedNodeID
}

func (in *BranchNodeStatus) Equals(other *BranchNodeStatus) bool {
	if in == nil && other == nil {
		return true
	}
	if in != nil && other != nil {
		phaseEqual := in.Phase == other.Phase
		if phaseEqual {
			if in.FinalizedNodeID == nil && other.FinalizedNodeID == nil {
				return true
			}
			if in.FinalizedNodeID != nil && other.FinalizedNodeID != nil {
				return *in.FinalizedNodeID == *other.FinalizedNodeID
			}
			return false
		}
		return false
	}
	return false
}

type DynamicNodePhase int

const (
	DynamicNodePhaseNone DynamicNodePhase = iota
	DynamicNodePhaseExecuting
)

type DynamicNodeStatus struct {
	Phase DynamicNodePhase `json:"phase"`
}

func (s *DynamicNodeStatus) GetDynamicNodePhase() DynamicNodePhase {
	return s.Phase
}

func (s *DynamicNodeStatus) SetDynamicNodePhase(phase DynamicNodePhase) {
	s.Phase = phase
}

func (s *DynamicNodeStatus) Equals(o *DynamicNodeStatus) bool {
	if s == nil && o == nil {
		return true
	}
	if s != nil && o != nil {
		return s.Phase == o.Phase
	}
	return false
}

type SubWorkflowNodeStatus struct {
	Phase WorkflowPhase `json:"phase"`
}

func (s SubWorkflowNodeStatus) GetPhase() WorkflowPhase {
	return s.Phase
}

func (s *SubWorkflowNodeStatus) SetPhase(phase WorkflowPhase) {
	s.Phase = phase
}

type WorkflowNodeStatus struct {
	WorkflowName string `json:"name"`
}

func (in *WorkflowNodeStatus) SetWorkflowExecutionName(name string) {
	in.WorkflowName = name
}

func (in *WorkflowNodeStatus) GetWorkflowExecutionName() string {
	return in.WorkflowName
}

type NodeStatus struct {
	Phase         NodePhase     `json:"phase"`
	QueuedAt      *metav1.Time  `json:"queuedAt,omitempty"`
	StartedAt     *metav1.Time  `json:"startedAt,omitempty"`
	StoppedAt     *metav1.Time  `json:"stoppedAt,omitempty"`
	LastUpdatedAt *metav1.Time  `json:"lastUpdatedAt,omitempty"`
	Message       string        `json:"message,omitempty"`
	DataDir       DataReference `json:"dataDir,omitempty"`
	Attempts      uint32        `json:"attempts"`
	Cached        bool          `json:"cached"`
	dirty         bool
	// This is useful only for branch nodes. If this is set, then it can be used to determine if execution can proceed
	ParentNode    *NodeID                  `json:"parentNode,omitempty"`
	ParentTask    *TaskExecutionIdentifier `json:"parentTask,omitempty"`
	BranchStatus  *BranchNodeStatus        `json:"branchStatus,omitempty"`
	SubNodeStatus map[NodeID]*NodeStatus   `json:"subNodeStatus,omitempty"`
	// We can store the outputs at this layer

	WorkflowNodeStatus    *WorkflowNodeStatus    `json:"workflowNodeStatus,omitempty"`
	TaskNodeStatus        *TaskNodeStatus        `json:",omitempty"`
	SubWorkflowNodeStatus *SubWorkflowNodeStatus `json:"subWorkflowStatus,omitempty"`
	DynamicNodeStatus     *DynamicNodeStatus     `json:"dynamicNodeStatus,omitempty"`
}

func (in NodeStatus) VisitNodeStatuses(visitor NodeStatusVisitFn) {
	for n, s := range in.SubNodeStatus {
		visitor(n, s)
	}
}

func (in *NodeStatus) ClearWorkflowStatus() {
	in.WorkflowNodeStatus = nil
}

func (in *NodeStatus) ClearTaskStatus() {
	in.TaskNodeStatus = nil
}

func (in *NodeStatus) GetLastUpdatedAt() *metav1.Time {
	return in.LastUpdatedAt
}

func (in *NodeStatus) GetAttempts() uint32 {
	return in.Attempts
}

func (in *NodeStatus) SetCached() {
	in.Cached = true
	in.setDirty()
}

func (in *NodeStatus) setDirty() {
	in.dirty = true
}
func (in *NodeStatus) IsCached() bool {
	return in.Cached
}

func (in *NodeStatus) IsDirty() bool {
	return in.dirty
}

// ResetDirty is for unit tests, shouldn't be used in actual logic.
func (in *NodeStatus) ResetDirty() {
	in.dirty = false
}

func (in *NodeStatus) IncrementAttempts() uint32 {
	in.Attempts++
	in.setDirty()
	return in.Attempts
}

func (in *NodeStatus) GetOrCreateDynamicNodeStatus() MutableDynamicNodeStatus {
	if in.DynamicNodeStatus == nil {
		in.setDirty()
		in.DynamicNodeStatus = &DynamicNodeStatus{}
	}

	return in.DynamicNodeStatus
}

func (in *NodeStatus) ClearDynamicNodeStatus() {
	in.DynamicNodeStatus = nil
}

func (in *NodeStatus) GetOrCreateBranchStatus() MutableBranchNodeStatus {
	if in.BranchStatus == nil {
		in.BranchStatus = &BranchNodeStatus{}
	}

	in.setDirty()
	return in.BranchStatus
}

func (in *NodeStatus) GetWorkflowNodeStatus() ExecutableWorkflowNodeStatus {
	if in.WorkflowNodeStatus == nil {
		return nil
	}

	in.setDirty()
	return in.WorkflowNodeStatus
}

func (in *NodeStatus) GetPhase() NodePhase {
	return in.Phase
}

func (in *NodeStatus) GetMessage() string {
	return in.Message
}

func IsPhaseTerminal(phase NodePhase) bool {
	return phase == NodePhaseSucceeded || phase == NodePhaseFailed || phase == NodePhaseSkipped
}

func (in *NodeStatus) GetOrCreateTaskStatus() MutableTaskNodeStatus {
	if in.TaskNodeStatus == nil {
		in.TaskNodeStatus = &TaskNodeStatus{}
	}

	in.setDirty()
	return in.TaskNodeStatus
}

func (in *NodeStatus) UpdatePhase(p NodePhase, occurredAt metav1.Time, reason string) {
	if in.Phase == p {
		// We will not update the phase multiple times. This prevents the comparison from returning false positive
		return
	}

	in.Phase = p
	in.Message = reason
	if len(reason) > maxMessageSize {
		in.Message = reason[:maxMessageSize]
	}

	n := occurredAt
	if occurredAt.IsZero() {
		n = metav1.Now()
	}

	if p == NodePhaseQueued && in.QueuedAt == nil {
		in.QueuedAt = &n
	} else if p == NodePhaseRunning && in.StartedAt == nil {
		in.StartedAt = &n
	} else if IsPhaseTerminal(p) && in.StoppedAt == nil {
		if in.StartedAt == nil {
			in.StartedAt = &n
		}

		in.StoppedAt = &n
	}

	if in.Phase != p {
		in.LastUpdatedAt = &n
	}

	in.setDirty()
}

func (in *NodeStatus) GetStartedAt() *metav1.Time {
	return in.StartedAt
}

func (in *NodeStatus) GetStoppedAt() *metav1.Time {
	return in.StoppedAt
}

func (in *NodeStatus) GetQueuedAt() *metav1.Time {
	return in.QueuedAt
}

func (in *NodeStatus) GetParentNodeID() *NodeID {
	return in.ParentNode
}

func (in *NodeStatus) GetParentTaskID() *core.TaskExecutionIdentifier {
	if in.ParentTask != nil {
		return in.ParentTask.TaskExecutionIdentifier
	}
	return nil
}

func (in *NodeStatus) SetParentNodeID(n *NodeID) {
	in.ParentNode = n
	in.setDirty()
}

func (in *NodeStatus) SetParentTaskID(t *core.TaskExecutionIdentifier) {
	in.ParentTask = &TaskExecutionIdentifier{
		TaskExecutionIdentifier: t,
	}
	in.setDirty()
}

func (in *NodeStatus) GetOrCreateWorkflowStatus() MutableWorkflowNodeStatus {
	if in.WorkflowNodeStatus == nil {
		in.WorkflowNodeStatus = &WorkflowNodeStatus{}
	}

	in.setDirty()
	return in.WorkflowNodeStatus
}

func (in NodeStatus) GetTaskNodeStatus() ExecutableTaskNodeStatus {
	// Explicitly return nil here to avoid a misleading non-nil interface.
	if in.TaskNodeStatus == nil {
		return nil
	}

	return in.TaskNodeStatus
}

func (in NodeStatus) GetSubWorkflowNodeStatus() ExecutableSubWorkflowNodeStatus {
	if in.SubWorkflowNodeStatus == nil {
		return nil
	}

	return in.SubWorkflowNodeStatus
}

func (in NodeStatus) GetOrCreateSubWorkflowStatus() MutableSubWorkflowNodeStatus {
	if in.SubWorkflowNodeStatus == nil {
		in.SubWorkflowNodeStatus = &SubWorkflowNodeStatus{}
	}

	return in.SubWorkflowNodeStatus
}

func (in *NodeStatus) ClearSubWorkflowStatus() {
	in.SubWorkflowNodeStatus = nil
}

func (in *NodeStatus) GetNodeExecutionStatus(id NodeID) ExecutableNodeStatus {
	n, ok := in.SubNodeStatus[id]
	if ok {
		return n
	}
	if in.SubNodeStatus == nil {
		in.SubNodeStatus = make(map[NodeID]*NodeStatus)
	}
	newNodeStatus := &NodeStatus{}
	newNodeStatus.SetParentTaskID(in.GetParentTaskID())
	newNodeStatus.SetParentNodeID(in.GetParentNodeID())

	in.SubNodeStatus[id] = newNodeStatus
	return newNodeStatus
}

func (in *NodeStatus) IsTerminated() bool {
	return in.GetPhase() == NodePhaseFailed || in.GetPhase() == NodePhaseSkipped || in.GetPhase() == NodePhaseSucceeded
}

func (in *NodeStatus) GetDataDir() DataReference {
	return in.DataDir
}

func (in *NodeStatus) SetDataDir(d DataReference) {
	in.DataDir = d
	in.setDirty()
}

func (in *NodeStatus) Equals(other *NodeStatus) bool {
	// Assuming in is never nil
	if other == nil {
		return false
	}

	if in.Attempts != other.Attempts {
		return false
	}

	if in.Phase != other.Phase {
		return false
	}

	if !reflect.DeepEqual(in.TaskNodeStatus, other.TaskNodeStatus) {
		return false
	}

	if in.DataDir != other.DataDir {
		return false
	}

	if in.ParentNode != nil && other.ParentNode != nil {
		if *in.ParentNode != *other.ParentNode {
			return false
		}
	} else if !(in.ParentNode == other.ParentNode) {
		// Both are not nil
		return false
	}

	if !reflect.DeepEqual(in.ParentTask, other.ParentTask) {
		return false
	}

	if len(in.SubNodeStatus) != len(other.SubNodeStatus) {
		return false
	}

	for k, v := range in.SubNodeStatus {
		otherV, ok := other.SubNodeStatus[k]
		if !ok {
			return false
		}
		if !v.Equals(otherV) {
			return false
		}
	}

	return in.BranchStatus.Equals(other.BranchStatus) // && in.DynamicNodeStatus.Equals(other.DynamicNodeStatus)
}

// THIS IS NOT AUTO GENERATED
func (in *CustomState) DeepCopyInto(out *CustomState) {
	if in == nil || *in == nil {
		return
	}

	raw, err := json.Marshal(in)
	if err != nil {
		return
	}

	err = json.Unmarshal(raw, out)
	if err != nil {
		return
	}
}

func (in *CustomState) DeepCopy() *CustomState {
	if in == nil || *in == nil {
		return nil
	}

	out := &CustomState{}
	in.DeepCopyInto(out)
	return out
}

type TaskNodeStatus struct {
	Phase        types.TaskPhase   `json:"phase,omitempty"`
	PhaseVersion uint32            `json:"phaseVersion,omitempty"`
	CustomState  types.CustomState `json:"custom,omitempty"`
}

func (in *TaskNodeStatus) SetPhase(phase types.TaskPhase) {
	in.Phase = phase
}

func (in *TaskNodeStatus) SetPhaseVersion(version uint32) {
	in.PhaseVersion = version
}

func (in *TaskNodeStatus) SetCustomState(state types.CustomState) {
	in.CustomState = state
}

func (in TaskNodeStatus) GetPhase() types.TaskPhase {
	return in.Phase
}

func (in TaskNodeStatus) GetPhaseVersion() uint32 {
	return in.PhaseVersion
}

func (in TaskNodeStatus) GetCustomState() types.CustomState {
	return in.CustomState
}

func (in *TaskNodeStatus) UpdatePhase(phase types.TaskPhase, phaseVersion uint32) {
	in.Phase = phase
	in.PhaseVersion = phaseVersion
}

func (in *TaskNodeStatus) UpdateCustomState(state types.CustomState) {
	in.CustomState = state
}

func (in *TaskNodeStatus) DeepCopyInto(out *TaskNodeStatus) {
	if in == nil {
		return
	}

	raw, err := json.Marshal(in)
	if err != nil {
		return
	}

	err = json.Unmarshal(raw, out)
	if err != nil {
		return
	}
}

func (in *TaskNodeStatus) DeepCopy() *TaskNodeStatus {
	if in == nil {
		return nil
	}

	out := &TaskNodeStatus{}
	in.DeepCopyInto(out)
	return out
}
