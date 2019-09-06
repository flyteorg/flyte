package v1alpha1

import (
	"context"

	"github.com/lyft/flytestdlib/storage"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const maxMessageSize = 1024

type WorkflowStatus struct {
	Phase           WorkflowPhase `json:"phase"`
	StartedAt       *metav1.Time  `json:"startedAt,omitempty"`
	StoppedAt       *metav1.Time  `json:"stoppedAt,omitempty"`
	LastUpdatedAt   *metav1.Time  `json:"lastUpdatedAt,omitempty"`
	Message         string        `json:"message,omitempty"`
	DataDir         DataReference `json:"dataDir,omitempty"`
	OutputReference DataReference `json:"outputRef,omitempty"`

	// We can store the outputs at this layer
	// We can also store a cross section of nodes being executed currently here. This could be an optimization

	NodeStatus map[NodeID]*NodeStatus `json:"nodeStatus,omitempty"`

	// Number of Attempts completed with rounds resulting in error. this is used to cap out poison pill workflows
	// that spin in an error loop. The value should be set at the global level and will be enforced. At the end of
	// the retries the workflow will fail
	FailedAttempts uint32 `json:"failedAttempts,omitempty"`
}

func IsWorkflowPhaseTerminal(p WorkflowPhase) bool {
	return p == WorkflowPhaseFailed || p == WorkflowPhaseSuccess || p == WorkflowPhaseAborted
}

func (in *WorkflowStatus) SetMessage(msg string) {
	in.Message = msg
}

func (in *WorkflowStatus) UpdatePhase(p WorkflowPhase, msg string) {
	in.Phase = p
	in.Message = msg
	if len(msg) > maxMessageSize {
		in.Message = msg[:maxMessageSize]
	}

	n := metav1.Now()
	if in.StartedAt == nil {
		in.StartedAt = &n
	}

	if IsWorkflowPhaseTerminal(p) && in.StoppedAt == nil {
		in.StoppedAt = &n
	}

	in.LastUpdatedAt = &n
}

func (in *WorkflowStatus) IncFailedAttempts() {
	in.FailedAttempts++
}

func (in *WorkflowStatus) GetPhase() WorkflowPhase {
	return in.Phase
}

func (in *WorkflowStatus) GetStartedAt() *metav1.Time {
	return in.StartedAt
}

func (in *WorkflowStatus) GetStoppedAt() *metav1.Time {
	return in.StoppedAt
}

func (in *WorkflowStatus) GetLastUpdatedAt() *metav1.Time {
	return in.LastUpdatedAt
}

func (in *WorkflowStatus) IsTerminated() bool {
	return in.Phase == WorkflowPhaseSuccess || in.Phase == WorkflowPhaseFailed || in.Phase == WorkflowPhaseAborted
}

func (in *WorkflowStatus) GetMessage() string {
	return in.Message
}

func (in *WorkflowStatus) GetNodeExecutionStatus(id NodeID) ExecutableNodeStatus {
	n, ok := in.NodeStatus[id]
	if ok {
		return n
	}
	if in.NodeStatus == nil {
		in.NodeStatus = make(map[NodeID]*NodeStatus)
	}
	newNodeStatus := &NodeStatus{}
	in.NodeStatus[id] = newNodeStatus
	return newNodeStatus
}

func (in *WorkflowStatus) ConstructNodeDataDir(ctx context.Context, constructor storage.ReferenceConstructor, name NodeID) (storage.DataReference, error) {
	return constructor.ConstructReference(ctx, in.GetDataDir(), name, "data")
}

func (in *WorkflowStatus) GetDataDir() DataReference {
	return in.DataDir
}

func (in *WorkflowStatus) SetDataDir(d DataReference) {
	in.DataDir = d
}

func (in *WorkflowStatus) GetOutputReference() DataReference {
	return in.OutputReference
}

func (in *WorkflowStatus) SetOutputReference(reference DataReference) {
	in.OutputReference = reference
}

func (in *WorkflowStatus) Equals(other *WorkflowStatus) bool {
	// Assuming in is never nil!
	if other == nil {
		return false
	}
	if in.FailedAttempts != other.FailedAttempts {
		return false
	}
	if in.Phase != other.Phase {
		return false
	}
	// We will not compare the time and message
	if in.DataDir != other.DataDir {
		return false
	}

	if in.OutputReference != other.OutputReference {
		return false
	}

	if len(in.NodeStatus) != len(other.NodeStatus) {
		return false
	}

	for k, v := range in.NodeStatus {
		otherV, ok := other.NodeStatus[k]
		if !ok {
			return false
		}
		if !v.Equals(otherV) {
			return false
		}
	}
	return true
}
