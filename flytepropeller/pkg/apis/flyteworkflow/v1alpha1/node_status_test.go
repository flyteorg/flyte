package v1alpha1

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/bitarray"
	"github.com/flyteorg/flyte/flytestdlib/storage"
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

	now := time.Now()

	one.QueuedAt = &metav1.Time{Time: now}
	assert.False(t, one.Equals(other))

	other.QueuedAt = &metav1.Time{Time: now}
	assert.True(t, one.Equals(other))

	one.StartedAt = &metav1.Time{Time: now}
	assert.False(t, one.Equals(other))

	other.StartedAt = &metav1.Time{Time: now}
	assert.True(t, one.Equals(other))

	one.StoppedAt = &metav1.Time{Time: now}
	assert.False(t, one.Equals(other))

	other.StoppedAt = &metav1.Time{Time: now}
	assert.True(t, one.Equals(other))

	one.LastUpdatedAt = &metav1.Time{Time: now}
	assert.False(t, one.Equals(other))

	other.LastUpdatedAt = &metav1.Time{Time: now}
	assert.True(t, one.Equals(other))

	one.LastAttemptStartedAt = &metav1.Time{Time: now}
	assert.False(t, one.Equals(other))

	other.LastAttemptStartedAt = &metav1.Time{Time: now}
	assert.True(t, one.Equals(other))

	one.Message = "test"
	assert.False(t, one.Equals(other))

	other.Message = "test"
	assert.True(t, one.Equals(other))

	one.IncrementAttempts()
	assert.False(t, one.Equals(other))

	other.IncrementAttempts()
	assert.True(t, one.Equals(other))

	one.IncrementSystemFailures()
	assert.False(t, one.Equals(other))

	other.IncrementSystemFailures()
	assert.True(t, one.Equals(other))

	one.SetCached()
	assert.False(t, one.Equals(other))

	other.SetCached()
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

	one.TaskNodeStatus = &TaskNodeStatus{
		Phase: 1,
	}
	assert.False(t, one.Equals(other))

	other.TaskNodeStatus = &TaskNodeStatus{
		Phase: 1,
	}
	assert.True(t, one.Equals(other))

	one.DynamicNodeStatus = &DynamicNodeStatus{
		Phase: 1,
	}
	assert.False(t, one.Equals(other))

	other.DynamicNodeStatus = &DynamicNodeStatus{
		Phase: 1,
	}
	assert.True(t, one.Equals(other))

	one.GateNodeStatus = &GateNodeStatus{
		Phase: 1,
	}
	assert.False(t, one.Equals(other))

	other.GateNodeStatus = &GateNodeStatus{
		Phase: 1,
	}
	assert.True(t, one.Equals(other))

	one.ArrayNodeStatus = &ArrayNodeStatus{
		Phase: 1,
	}
	assert.False(t, one.Equals(other))

	other.ArrayNodeStatus = &ArrayNodeStatus{
		Phase: 1,
	}
	assert.True(t, one.Equals(other))

	one.Error = &ExecutionError{
		ExecutionError: &core.ExecutionError{
			Code: "1",
		},
	}
	assert.False(t, one.Equals(other))
	other.Error = &ExecutionError{
		ExecutionError: &core.ExecutionError{
			Code: "1",
		},
	}
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

	one.IsFailurePermanent = true
	assert.False(t, one.Equals(other))
	other.IsFailurePermanent = true

	assert.True(t, one.Equals(other))

	one.Error = &ExecutionError{
		ExecutionError: &core.ExecutionError{
			Code: "1",
		},
	}
	assert.False(t, one.Equals(other))
	other.Error = &ExecutionError{
		ExecutionError: &core.ExecutionError{
			Code: "1",
		},
	}
	assert.True(t, one.Equals(other))
}

func TestTaskNodeStatus_Equals(t *testing.T) {
	var one *TaskNodeStatus
	var other *TaskNodeStatus
	assert.True(t, one.Equals(other))

	one = &TaskNodeStatus{}
	assert.False(t, one.Equals(other))

	other = &TaskNodeStatus{}
	assert.True(t, one.Equals(other))

	one.Phase = 5
	assert.False(t, one.Equals(other))

	other.Phase = 5
	assert.True(t, one.Equals(other))

	one.PhaseVersion = 5
	assert.False(t, one.Equals(other))

	other.PhaseVersion = 5
	assert.True(t, one.Equals(other))

	one.PluginState = []byte{1, 2, 3}
	assert.False(t, one.Equals(other))

	other.PluginState = []byte{1, 2, 3}
	assert.True(t, one.Equals(other))

	one.PluginStateVersion = 5
	assert.False(t, one.Equals(other))

	other.PluginStateVersion = 5
	assert.True(t, one.Equals(other))

	one.BarrierClockTick = 5
	assert.False(t, one.Equals(other))

	other.BarrierClockTick = 5
	assert.True(t, one.Equals(other))

	now := time.Now()
	one.LastPhaseUpdatedAt = now
	assert.False(t, one.Equals(other))

	other.LastPhaseUpdatedAt = now
	assert.True(t, one.Equals(other))

	one.PreviousNodeExecutionCheckpointPath = "/test"
	assert.False(t, one.Equals(other))

	other.PreviousNodeExecutionCheckpointPath = "/test"
	assert.True(t, one.Equals(other))

	one.CleanupOnFailure = true
	assert.False(t, one.Equals(other))

	other.CleanupOnFailure = true
	assert.True(t, one.Equals(other))
}

func TestGateNodeStatus_Equals(t *testing.T) {
	var one *GateNodeStatus
	var other *GateNodeStatus

	assert.True(t, one.Equals(other))

	one = &GateNodeStatus{}
	assert.False(t, one.Equals(other))

	other = &GateNodeStatus{}
	assert.True(t, one.Equals(other))

	one.Phase = 5
	assert.False(t, one.Equals(other))

	other.Phase = 5
	assert.True(t, one.Equals(other))
}

func TestArrayNodeStatus_Equals(t *testing.T) {
	var one *ArrayNodeStatus
	var other *ArrayNodeStatus

	assert.True(t, one.Equals(other))

	one = &ArrayNodeStatus{}
	assert.False(t, one.Equals(other))

	other = &ArrayNodeStatus{}
	assert.True(t, one.Equals(other))

	one.Phase = 5
	assert.False(t, one.Equals(other))

	other.Phase = 5
	assert.True(t, one.Equals(other))

	one.ExecutionError = &core.ExecutionError{
		Code: "1",
	}
	assert.False(t, one.Equals(other))

	other.ExecutionError = &core.ExecutionError{
		Code: "1",
	}
	assert.True(t, one.Equals(other))

	compactArray, err := bitarray.NewCompactArray(10, bitarray.Item(10))
	require.NoError(t, err)
	compactArray.SetItem(0, 5)

	one.SubNodePhases = compactArray
	assert.False(t, one.Equals(other))

	other.SubNodePhases = *compactArray.DeepCopy()
	assert.True(t, one.Equals(other))

	one.SubNodeTaskPhases = compactArray
	assert.False(t, one.Equals(other))

	other.SubNodeTaskPhases = *compactArray.DeepCopy()
	assert.True(t, one.Equals(other))

	one.SubNodeRetryAttempts = compactArray
	assert.False(t, one.Equals(other))

	other.SubNodeRetryAttempts = *compactArray.DeepCopy()
	assert.True(t, one.Equals(other))

	one.SubNodeSystemFailures = compactArray
	assert.False(t, one.Equals(other))

	other.SubNodeSystemFailures = *compactArray.DeepCopy()
	assert.True(t, one.Equals(other))

	one.SubNodeDeltaTimestamps = compactArray
	assert.False(t, one.Equals(other))

	other.SubNodeDeltaTimestamps = *compactArray.DeepCopy()
	assert.True(t, one.Equals(other))

	one.TaskPhaseVersion = 5
	assert.False(t, one.Equals(other))

	other.TaskPhaseVersion = 5
	assert.True(t, one.Equals(other))

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
	const success = "success"
	for _, enableCRDebugMetadata := range []bool{false, true} {
		t.Run("identical-phase", func(t *testing.T) {
			p := NodePhaseQueued
			ns := NodeStatus{
				Phase:   p,
				Message: queued,
			}
			msg := queued
			ns.UpdatePhase(p, n, msg, enableCRDebugMetadata, nil)
			assert.Nil(t, ns.QueuedAt)
		})

		t.Run("zero", func(t *testing.T) {
			p := NodePhaseQueued
			ns := NodeStatus{}
			msg := queued
			ns.UpdatePhase(p, metav1.NewTime(time.Time{}), msg, enableCRDebugMetadata, nil)
			assert.NotNil(t, ns.QueuedAt)
		})

		t.Run("non-terminal", func(t *testing.T) {
			ns := NodeStatus{}
			p := NodePhaseQueued
			msg := queued
			ns.UpdatePhase(p, n, msg, enableCRDebugMetadata, nil)

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
			ns.UpdatePhase(p, n, msg, enableCRDebugMetadata, nil)

			assert.Equal(t, *ns.LastUpdatedAt, n)
			assert.Nil(t, ns.QueuedAt)
			assert.Equal(t, *ns.LastAttemptStartedAt, n)
			assert.Equal(t, *ns.StartedAt, n)
			assert.Nil(t, ns.StoppedAt)
			assert.Equal(t, p, ns.Phase)
			assert.Equal(t, msg, ns.Message)
			assert.Nil(t, ns.Error)
		})

		t.Run("non-terminal-timing-out", func(t *testing.T) {
			ns := NodeStatus{}
			p := NodePhaseTimingOut
			msg := "timing-out"
			ns.UpdatePhase(p, n, msg, enableCRDebugMetadata, nil)

			assert.Equal(t, *ns.LastUpdatedAt, n)
			assert.Nil(t, ns.QueuedAt)
			assert.Nil(t, ns.LastAttemptStartedAt)
			assert.Nil(t, ns.StartedAt)
			assert.Nil(t, ns.StoppedAt)
			assert.Equal(t, p, ns.Phase)
			assert.Equal(t, msg, ns.Message)
			assert.Nil(t, ns.Error)
		})

		t.Run("terminal-success", func(t *testing.T) {
			ns := NodeStatus{}
			p := NodePhaseSucceeded
			msg := success
			ns.UpdatePhase(p, n, msg, enableCRDebugMetadata, nil)

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
			ns.UpdatePhase(p, n, msg, enableCRDebugMetadata, nil)

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
			ns.UpdatePhase(p, n, msg, enableCRDebugMetadata, nil)

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
			ns.UpdatePhase(p, n2, msg, enableCRDebugMetadata, nil)

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

	t.Run("terminal-fail", func(t *testing.T) {
		ns := NodeStatus{}
		p := NodePhaseFailed
		msg := "failed"
		err := &core.ExecutionError{}
		ns.UpdatePhase(p, n, msg, true, err)

		assert.Equal(t, *ns.LastUpdatedAt, n)
		assert.Nil(t, ns.QueuedAt)
		assert.Equal(t, *ns.LastAttemptStartedAt, n)
		assert.Equal(t, *ns.StartedAt, n)
		assert.Equal(t, *ns.StoppedAt, n)
		assert.Equal(t, p, ns.Phase)
		assert.Equal(t, msg, ns.Message)
		assert.Equal(t, ns.Error.ExecutionError, err)
	})

	t.Run("terminal-fail-clear-state-on-any-termination", func(t *testing.T) {
		ns := NodeStatus{}
		p := NodePhaseFailed
		msg := "failed"
		err := &core.ExecutionError{}
		ns.UpdatePhase(p, n, msg, false, err)

		assert.Nil(t, ns.LastUpdatedAt)
		assert.Nil(t, ns.QueuedAt)
		assert.Nil(t, ns.LastAttemptStartedAt)
		assert.Nil(t, ns.StartedAt)
		assert.Equal(t, *ns.StoppedAt, n)
		assert.Equal(t, p, ns.Phase)
		assert.Equal(t, ns.Message, "")
		assert.Equal(t, ns.Error.ExecutionError, err)
	})

	t.Run("terminal-timeout", func(t *testing.T) {
		ns := NodeStatus{}
		p := NodePhaseTimedOut
		msg := "tm"
		err := &core.ExecutionError{}
		ns.UpdatePhase(p, n, msg, true, err)

		assert.Equal(t, *ns.LastUpdatedAt, n)
		assert.Nil(t, ns.QueuedAt)
		assert.Equal(t, *ns.LastAttemptStartedAt, n)
		assert.Equal(t, *ns.StartedAt, n)
		assert.Equal(t, *ns.StoppedAt, n)
		assert.Equal(t, p, ns.Phase)
		assert.Equal(t, msg, ns.Message)
		assert.Equal(t, ns.Error.ExecutionError, err)
	})

	t.Run("terminal-timeout-clear-state-on-any-termination", func(t *testing.T) {
		ns := NodeStatus{}
		p := NodePhaseTimedOut
		msg := "tm"
		err := &core.ExecutionError{}
		ns.UpdatePhase(p, n, msg, false, err)

		assert.Nil(t, ns.LastUpdatedAt)
		assert.Nil(t, ns.QueuedAt)
		assert.Nil(t, ns.LastAttemptStartedAt)
		assert.Nil(t, ns.StartedAt)
		assert.Equal(t, *ns.StoppedAt, n)
		assert.Equal(t, p, ns.Phase)
		assert.Equal(t, ns.Message, "")
		assert.Equal(t, ns.Error.ExecutionError, err)
	})
}
