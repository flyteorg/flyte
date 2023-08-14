package v1alpha1

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flytestdlib/storage"

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

func TestNodeStatus_GetNodeExecutionStatus(t *testing.T) {
	ctx := context.Background()
	t.Run("First Level", func(t *testing.T) {
		t.Run("Not cached", func(t *testing.T) {
			n := NodeStatus{
				SubNodeStatus:            map[NodeID]*NodeStatus{},
				DataReferenceConstructor: storage.URLPathConstructor{},
			}

			newNode := n.GetNodeExecutionStatus(ctx, "abc")
			assert.Equal(t, storage.DataReference("/abc/0"), newNode.GetOutputDir())
			assert.Equal(t, storage.DataReference("/abc"), newNode.GetDataDir())
		})

		t.Run("cached", func(t *testing.T) {
			n := NodeStatus{
				SubNodeStatus:            map[NodeID]*NodeStatus{},
				DataReferenceConstructor: storage.URLPathConstructor{},
			}

			newNode := n.GetNodeExecutionStatus(ctx, "abc")
			assert.Equal(t, storage.DataReference("/abc/0"), newNode.GetOutputDir())
			assert.Equal(t, storage.DataReference("/abc"), newNode.GetDataDir())

			newNode = n.GetNodeExecutionStatus(ctx, "abc")
			assert.Equal(t, storage.DataReference("/abc/0"), newNode.GetOutputDir())
			assert.Equal(t, storage.DataReference("/abc"), newNode.GetDataDir())
		})

		t.Run("cached but datadir not populated", func(t *testing.T) {
			n := NodeStatus{
				SubNodeStatus: map[NodeID]*NodeStatus{
					"abc": {},
				},
				DataReferenceConstructor: storage.URLPathConstructor{},
			}

			newNode := n.GetNodeExecutionStatus(ctx, "abc")
			assert.Equal(t, storage.DataReference("/abc/0"), newNode.GetOutputDir())
			assert.Equal(t, storage.DataReference("/abc"), newNode.GetDataDir())
		})
	})

	t.Run("Nested", func(t *testing.T) {
		n := NodeStatus{
			SubNodeStatus:            map[NodeID]*NodeStatus{},
			DataReferenceConstructor: storage.URLPathConstructor{},
		}

		newNode := n.GetNodeExecutionStatus(ctx, "abc")
		assert.Equal(t, storage.DataReference("/abc/0"), newNode.GetOutputDir())
		assert.Equal(t, storage.DataReference("/abc"), newNode.GetDataDir())

		subsubNode := newNode.GetNodeExecutionStatus(ctx, "xyz")
		assert.Equal(t, storage.DataReference("/abc/0/xyz/0"), subsubNode.GetOutputDir())
		assert.Equal(t, storage.DataReference("/abc/0/xyz"), subsubNode.GetDataDir())
	})
}

func TestNodeStatus_UpdatePhase(t *testing.T) {
	n := metav1.NewTime(time.Now())

	const queued = "queued"
	t.Run("identical-phase", func(t *testing.T) {
		p := NodePhaseQueued
		ns := NodeStatus{
			Phase:   p,
			Message: queued,
		}
		msg := queued
		ns.UpdatePhase(p, n, msg, nil)
		assert.Nil(t, ns.QueuedAt)
	})

	t.Run("zero", func(t *testing.T) {
		p := NodePhaseQueued
		ns := NodeStatus{}
		msg := queued
		ns.UpdatePhase(p, metav1.NewTime(time.Time{}), msg, nil)
		assert.NotNil(t, ns.QueuedAt)
	})

	t.Run("non-terminal", func(t *testing.T) {
		ns := NodeStatus{}
		p := NodePhaseQueued
		msg := queued
		ns.UpdatePhase(p, n, msg, nil)

		assert.Equal(t, *ns.LastUpdatedAt, n)
		assert.Equal(t, *ns.QueuedAt, n)
		assert.Nil(t, ns.LastAttemptStartedAt)
		assert.Nil(t, ns.StartedAt)
		assert.Nil(t, ns.StoppedAt)
		assert.Equal(t, p, ns.Phase)
		assert.Equal(t, msg, ns.Message)
		assert.Nil(t, ns.Error)
	})

	t.Run("non-terminal-running", func(t *testing.T) {
		ns := NodeStatus{}
		p := NodePhaseRunning
		msg := "running"
		ns.UpdatePhase(p, n, msg, nil)

		assert.Equal(t, *ns.LastUpdatedAt, n)
		assert.Nil(t, ns.QueuedAt)
		assert.Equal(t, *ns.LastAttemptStartedAt, n)
		assert.Equal(t, *ns.StartedAt, n)
		assert.Nil(t, ns.StoppedAt)
		assert.Equal(t, p, ns.Phase)
		assert.Equal(t, msg, ns.Message)
		assert.Nil(t, ns.Error)
	})

	t.Run("terminal-fail", func(t *testing.T) {
		ns := NodeStatus{}
		p := NodePhaseFailed
		msg := "failed"
		err := &core.ExecutionError{}
		ns.UpdatePhase(p, n, msg, err)

		assert.Equal(t, *ns.LastUpdatedAt, n)
		assert.Nil(t, ns.QueuedAt)
		assert.Equal(t, *ns.LastAttemptStartedAt, n)
		assert.Equal(t, *ns.StartedAt, n)
		assert.Equal(t, *ns.StoppedAt, n)
		assert.Equal(t, p, ns.Phase)
		assert.Equal(t, msg, ns.Message)
		assert.Equal(t, ns.Error.ExecutionError, err)
	})

	t.Run("terminal-timeout", func(t *testing.T) {
		ns := NodeStatus{}
		p := NodePhaseTimedOut
		msg := "tm"
		err := &core.ExecutionError{}
		ns.UpdatePhase(p, n, msg, err)

		assert.Equal(t, *ns.LastUpdatedAt, n)
		assert.Nil(t, ns.QueuedAt)
		assert.Equal(t, *ns.LastAttemptStartedAt, n)
		assert.Equal(t, *ns.StartedAt, n)
		assert.Equal(t, *ns.StoppedAt, n)
		assert.Equal(t, p, ns.Phase)
		assert.Equal(t, msg, ns.Message)
		assert.Equal(t, ns.Error.ExecutionError, err)
	})

	const success = "success"
	t.Run("terminal-success", func(t *testing.T) {
		ns := NodeStatus{}
		p := NodePhaseSucceeded
		msg := success
		ns.UpdatePhase(p, n, msg, nil)

		assert.Nil(t, ns.LastUpdatedAt)
		assert.Nil(t, ns.QueuedAt)
		assert.Nil(t, ns.LastAttemptStartedAt)
		assert.Nil(t, ns.StartedAt)
		assert.Equal(t, *ns.StoppedAt, n)
		assert.Equal(t, p, ns.Phase)
		assert.Empty(t, ns.Message)
		assert.Nil(t, ns.Error)
	})

	t.Run("terminal-skipped", func(t *testing.T) {
		ns := NodeStatus{}
		p := NodePhaseSucceeded
		msg := success
		ns.UpdatePhase(p, n, msg, nil)

		assert.Nil(t, ns.LastUpdatedAt)
		assert.Nil(t, ns.QueuedAt)
		assert.Nil(t, ns.LastAttemptStartedAt)
		assert.Nil(t, ns.StartedAt)
		assert.Equal(t, *ns.StoppedAt, n)
		assert.Equal(t, p, ns.Phase)
		assert.Empty(t, ns.Message)
		assert.Nil(t, ns.Error)
	})

	t.Run("terminal-success-preset", func(t *testing.T) {
		ns := NodeStatus{
			QueuedAt:             &n,
			StartedAt:            &n,
			LastUpdatedAt:        &n,
			LastAttemptStartedAt: &n,
			WorkflowNodeStatus:   &WorkflowNodeStatus{},
			BranchStatus:         &BranchNodeStatus{},
			DynamicNodeStatus:    &DynamicNodeStatus{},
			TaskNodeStatus:       &TaskNodeStatus{},
			SubNodeStatus:        map[NodeID]*NodeStatus{},
		}
		p := NodePhaseSucceeded
		msg := success
		ns.UpdatePhase(p, n, msg, nil)

		assert.Nil(t, ns.LastUpdatedAt)
		assert.Nil(t, ns.QueuedAt)
		assert.Nil(t, ns.LastAttemptStartedAt)
		assert.Nil(t, ns.StartedAt)
		assert.Equal(t, *ns.StoppedAt, n)
		assert.Equal(t, p, ns.Phase)
		assert.Empty(t, ns.Message)
		assert.Nil(t, ns.Error)
		assert.Nil(t, ns.SubNodeStatus)
		assert.Nil(t, ns.DynamicNodeStatus)
		assert.Nil(t, ns.WorkflowNodeStatus)
		assert.Nil(t, ns.BranchStatus)
		assert.Nil(t, ns.TaskNodeStatus)
	})

	t.Run("non-terminal-preset", func(t *testing.T) {
		ns := NodeStatus{
			QueuedAt:             &n,
			StartedAt:            &n,
			LastUpdatedAt:        &n,
			LastAttemptStartedAt: &n,
			WorkflowNodeStatus:   &WorkflowNodeStatus{},
			BranchStatus:         &BranchNodeStatus{},
			DynamicNodeStatus:    &DynamicNodeStatus{},
			TaskNodeStatus:       &TaskNodeStatus{},
			SubNodeStatus:        map[NodeID]*NodeStatus{},
		}
		n2 := metav1.NewTime(time.Now())
		p := NodePhaseRunning
		msg := "running"
		ns.UpdatePhase(p, n2, msg, nil)

		assert.Equal(t, *ns.LastUpdatedAt, n2)
		assert.Equal(t, *ns.QueuedAt, n)
		assert.Equal(t, *ns.LastAttemptStartedAt, n)
		assert.Equal(t, *ns.StartedAt, n)
		assert.Nil(t, ns.StoppedAt)
		assert.Equal(t, p, ns.Phase)
		assert.Equal(t, msg, ns.Message)
		assert.Nil(t, ns.Error)
		assert.NotNil(t, ns.SubNodeStatus)
		assert.NotNil(t, ns.DynamicNodeStatus)
		assert.NotNil(t, ns.WorkflowNodeStatus)
		assert.NotNil(t, ns.BranchStatus)
		assert.NotNil(t, ns.TaskNodeStatus)
	})
}
