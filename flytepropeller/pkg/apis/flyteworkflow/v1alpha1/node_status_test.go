package v1alpha1

import (
	"encoding/json"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

func TestIsPhaseTerminal(t *testing.T) {
	assert.True(t, IsPhaseTerminal(NodePhaseFailed))
	assert.True(t, IsPhaseTerminal(NodePhaseSkipped))
	assert.True(t, IsPhaseTerminal(NodePhaseSucceeded))

	assert.False(t, IsPhaseTerminal(NodePhaseFailing))
	assert.False(t, IsPhaseTerminal(NodePhaseRunning))
	assert.False(t, IsPhaseTerminal(NodePhaseNotYetStarted))
}

func TestNodeStatus_Equals(t *testing.T) {
	one := &NodeStatus{}
	var other *NodeStatus
	assert.False(t, one.Equals(other))

	other = &NodeStatus{}
	assert.True(t, one.Equals(other))

	one.Phase = NodePhaseRunning
	assert.False(t, one.Equals(other))

	other.Phase = one.Phase
	assert.True(t, one.Equals(other))

	one.DataDir = "data-dir"
	assert.False(t, one.Equals(other))

	other.DataDir = one.DataDir
	assert.True(t, one.Equals(other))

	parentNode := "x"
	one.ParentNode = &parentNode
	assert.False(t, one.Equals(other))

	parentNode2 := "y"
	other.ParentNode = &parentNode2
	assert.False(t, one.Equals(other))

	other.ParentNode = &parentNode
	assert.True(t, one.Equals(other))

	one.BranchStatus = &BranchNodeStatus{}
	assert.False(t, one.Equals(other))
	other.BranchStatus = &BranchNodeStatus{}
	assert.True(t, one.Equals(other))

	node := "x"
	one.SubNodeStatus = map[NodeID]*NodeStatus{
		node: {},
	}
	assert.False(t, one.Equals(other))
	other.SubNodeStatus = map[NodeID]*NodeStatus{
		node: {},
	}
	assert.True(t, one.Equals(other))

	one.SubNodeStatus[node].Phase = NodePhaseRunning
	assert.False(t, one.Equals(other))
	other.SubNodeStatus[node].Phase = NodePhaseRunning
	assert.True(t, one.Equals(other))
}

func TestBranchNodeStatus_Equals(t *testing.T) {
	var one *BranchNodeStatus
	var other *BranchNodeStatus
	assert.True(t, one.Equals(other))
	one = &BranchNodeStatus{}

	assert.False(t, one.Equals(other))
	other = &BranchNodeStatus{}

	assert.True(t, one.Equals(other))

	one.Phase = BranchNodeError
	assert.False(t, one.Equals(other))
	other.Phase = one.Phase

	assert.True(t, one.Equals(other))

	node := "x"
	one.FinalizedNodeID = &node
	assert.False(t, one.Equals(other))

	node2 := "y"
	other.FinalizedNodeID = &node2
	assert.False(t, one.Equals(other))

	node2 = node
	other.FinalizedNodeID = &node2
	assert.True(t, one.Equals(other))
}

func TestDynamicNodeStatus_Equals(t *testing.T) {
	var one *DynamicNodeStatus
	var other *DynamicNodeStatus
	assert.True(t, one.Equals(other))
	one = &DynamicNodeStatus{}

	assert.False(t, one.Equals(other))
	other = &DynamicNodeStatus{}

	assert.True(t, one.Equals(other))

	one.Phase = DynamicNodePhaseExecuting
	assert.False(t, one.Equals(other))
	other.Phase = one.Phase

	assert.True(t, one.Equals(other))
}

func TestCustomState_DeepCopyInto(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		var in CustomState
		var out CustomState
		in.DeepCopyInto(&out)
		assert.Nil(t, in)
		assert.Nil(t, out)
	})

	t.Run("Not nil in", func(t *testing.T) {
		in := CustomState(map[string]interface{}{
			"key1": "hello",
		})

		var out CustomState
		in.DeepCopyInto(&out)
		assert.NotNil(t, out)
		assert.Equal(t, 1, len(out))
	})
}

func TestCustomState_DeepCopy(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		var in CustomState
		assert.Nil(t, in)
		assert.Nil(t, in.DeepCopy())
	})

	t.Run("Not nil in", func(t *testing.T) {
		in := CustomState(map[string]interface{}{
			"key1": "hello",
		})

		out := in.DeepCopy()
		assert.NotNil(t, out)
		assert.Equal(t, 1, len(*out))
	})
}

func TestWorkflowStatus_Deserialize(t *testing.T) {
	raw := []byte(`{"phase":0,"dataDir":"/blah/bloh.pb","attempts":0,"cached":false}`)

	parsed := &NodeStatus{}
	err := json.Unmarshal(raw, parsed)
	assert.NoError(t, err)
}

func TestDynamicNodeStatus_SetExecutionError(t *testing.T) {
	tests := []struct {
		name     string
		Error    *ExecutionError
		NewError *core.ExecutionError
	}{
		{"preset", &ExecutionError{}, nil},
		{"preset-new", &ExecutionError{}, &core.ExecutionError{}},
		{"new", nil, &core.ExecutionError{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := &DynamicNodeStatus{
				Error: tt.Error,
			}
			in.SetExecutionError(tt.NewError)
			if tt.NewError == nil {
				assert.Nil(t, in.Error)
			} else {
				assert.NotNil(t, in.Error)
				assert.Equal(t, tt.NewError, in.Error.ExecutionError)
			}
		})
	}
}
