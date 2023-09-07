package handler

import (
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	pluginCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"

	"github.com/flyteorg/flytestdlib/bitarray"
	"github.com/flyteorg/flytestdlib/storage"
)

// This is the legacy state structure that gets translated to node status
// TODO eventually we could just convert this to be binary node state encoded into the node status

type TaskNodeState struct {
	PluginPhase                        pluginCore.Phase
	PluginPhaseVersion                 uint32
	PluginState                        []byte
	PluginStateVersion                 uint32
	LastPhaseUpdatedAt                 time.Time
	PreviousNodeExecutionCheckpointURI storage.DataReference
	CleanupOnFailure                   bool
}

type BranchNodeState struct {
	FinalizedNodeID *v1alpha1.NodeID
	Phase           v1alpha1.BranchNodePhase
}

type DynamicNodePhase uint8

type DynamicNodeState struct {
	Phase              v1alpha1.DynamicNodePhase
	Reason             string
	Error              *core.ExecutionError
	IsFailurePermanent bool
}

type WorkflowNodeState struct {
	Phase v1alpha1.WorkflowNodePhase
	Error *core.ExecutionError
}

type GateNodeState struct {
	Phase     v1alpha1.GateNodePhase
	StartedAt time.Time
}

type ArrayNodeState struct {
	Phase                 v1alpha1.ArrayNodePhase
	TaskPhaseVersion      uint32
	Error                 *core.ExecutionError
	SubNodePhases         bitarray.CompactArray
	SubNodeTaskPhases     bitarray.CompactArray
	SubNodeRetryAttempts  bitarray.CompactArray
	SubNodeSystemFailures bitarray.CompactArray
}
