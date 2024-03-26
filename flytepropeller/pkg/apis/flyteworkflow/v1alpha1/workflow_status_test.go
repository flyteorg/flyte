package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsWorkflowPhaseTerminal(t *testing.T) {
	assert.True(t, IsWorkflowPhaseTerminal(WorkflowPhaseFailed))
	assert.True(t, IsWorkflowPhaseTerminal(WorkflowPhaseSuccess))

	assert.False(t, IsWorkflowPhaseTerminal(WorkflowPhaseFailing))
	assert.False(t, IsWorkflowPhaseTerminal(WorkflowPhaseSucceeding))
	assert.False(t, IsWorkflowPhaseTerminal(WorkflowPhaseReady))
	assert.False(t, IsWorkflowPhaseTerminal(WorkflowPhaseRunning))
}

func TestWorkflowStatus_Equals(t *testing.T) {
	one := &WorkflowStatus{}
	other := &WorkflowStatus{}
	assert.True(t, one.Equals(other))

	one.Phase = WorkflowPhaseRunning
	assert.False(t, one.Equals(other))

	other.Phase = one.Phase
	assert.True(t, one.Equals(other))

	one.DataDir = "data-dir"
	assert.False(t, one.Equals(other))
	other.DataDir = one.DataDir
	assert.True(t, one.Equals(other))

	node := "x"
	one.NodeStatus = map[NodeID]*NodeStatus{
		node: {},
	}
	assert.False(t, one.Equals(other))
	other.NodeStatus = map[NodeID]*NodeStatus{
		node: {},
	}
	assert.True(t, one.Equals(other))

	one.NodeStatus[node].Phase = NodePhaseRunning
	assert.False(t, one.Equals(other))
	other.NodeStatus[node].Phase = NodePhaseRunning
	assert.True(t, one.Equals(other))

	one.OutputReference = "out"
	assert.False(t, one.Equals(other))
	other.OutputReference = "out"
	assert.True(t, one.Equals(other))
}
