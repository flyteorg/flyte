package nodes

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/flyteorg/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/common"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/golang/protobuf/ptypes"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/handler"
)

// This is used by flyteadmin to indicate that the events will now contain populated IsParent and IsDynamic bits.
var nodeExecutionEventVersion = int32(1)

func ToNodeExecOutput(info *handler.OutputInfo) *event.NodeExecutionEvent_OutputUri {
	if info == nil || info.OutputURI == "" {
		return nil
	}

	return &event.NodeExecutionEvent_OutputUri{
		OutputUri: info.OutputURI.String(),
	}
}

func ToNodeExecWorkflowNodeMetadata(info *handler.WorkflowNodeInfo) *event.NodeExecutionEvent_WorkflowNodeMetadata {
	if info == nil || info.LaunchedWorkflowID == nil {
		return nil
	}
	return &event.NodeExecutionEvent_WorkflowNodeMetadata{
		WorkflowNodeMetadata: &event.WorkflowNodeMetadata{
			ExecutionId: info.LaunchedWorkflowID,
		},
	}
}

func ToNodeExecTaskNodeMetadata(info *handler.TaskNodeInfo) *event.NodeExecutionEvent_TaskNodeMetadata {
	if info == nil || info.TaskNodeMetadata == nil {
		return nil
	}
	return &event.NodeExecutionEvent_TaskNodeMetadata{
		TaskNodeMetadata: info.TaskNodeMetadata,
	}
}

func ToNodeExecEventPhase(p handler.EPhase) core.NodeExecution_Phase {
	switch p {
	case handler.EPhaseQueued:
		return core.NodeExecution_QUEUED
	case handler.EPhaseRunning, handler.EPhaseRetryableFailure:
		return core.NodeExecution_RUNNING
	case handler.EPhaseDynamicRunning:
		return core.NodeExecution_DYNAMIC_RUNNING
	case handler.EPhaseSkip:
		return core.NodeExecution_SKIPPED
	case handler.EPhaseSuccess:
		return core.NodeExecution_SUCCEEDED
	case handler.EPhaseFailed:
		return core.NodeExecution_FAILED
	case handler.EPhaseRecovered:
		return core.NodeExecution_RECOVERED
	default:
		return core.NodeExecution_UNDEFINED
	}
}

func ToNodeExecutionEvent(nodeExecID *core.NodeExecutionIdentifier,
	info handler.PhaseInfo,
	inputPath string,
	status v1alpha1.ExecutableNodeStatus,
	eventVersion v1alpha1.EventVersion,
	parentInfo executors.ImmutableParentInfo,
	node v1alpha1.ExecutableNode, clusterID string, dynamicNodePhase v1alpha1.DynamicNodePhase) (*event.NodeExecutionEvent, error) {
	if info.GetPhase() == handler.EPhaseNotReady {
		return nil, nil
	}
	if info.GetPhase() == handler.EPhaseUndefined {
		return nil, fmt.Errorf("illegal state, undefined phase received for node [%s]", nodeExecID.NodeId)
	}
	occurredTime, err := ptypes.TimestampProto(info.GetOccurredAt())
	if err != nil {
		return nil, err
	}

	phase := ToNodeExecEventPhase(info.GetPhase())
	if eventVersion < v1alpha1.EventVersion2 && phase == core.NodeExecution_DYNAMIC_RUNNING {
		// For older workflow event versions we lump dynamic running with running.
		phase = core.NodeExecution_RUNNING
	}

	var nev *event.NodeExecutionEvent
	// Start node is special case where the Inputs and Outputs are the same and hence here we copy the Output file
	// into the OutputResult and in admin we copy it over into input aswell.
	if nodeExecID.NodeId == v1alpha1.StartNodeID {
		outputsFile := v1alpha1.GetOutputsFile(status.GetOutputDir())
		nev = &event.NodeExecutionEvent{
			Id:    nodeExecID,
			Phase: phase,
			OutputResult: ToNodeExecOutput(&handler.OutputInfo{
				OutputURI: outputsFile,
			}),
			OccurredAt:   occurredTime,
			ProducerId:   clusterID,
			EventVersion: nodeExecutionEventVersion,
		}
	} else {
		nev = &event.NodeExecutionEvent{
			Id:           nodeExecID,
			Phase:        phase,
			InputUri:     inputPath,
			OccurredAt:   occurredTime,
			ProducerId:   clusterID,
			EventVersion: nodeExecutionEventVersion,
		}
	}

	if eventVersion == v1alpha1.EventVersion0 && status.GetParentTaskID() != nil {
		nev.ParentTaskMetadata = &event.ParentTaskExecutionMetadata{
			Id: status.GetParentTaskID(),
		}
	}

	if eventVersion != v1alpha1.EventVersion0 {
		currentNodeUniqueID, err := common.GenerateUniqueID(parentInfo, nev.Id.NodeId)
		if err != nil {
			return nil, err
		}
		if parentInfo != nil {
			nev.ParentNodeMetadata = &event.ParentNodeExecutionMetadata{
				NodeId: parentInfo.GetUniqueID(),
			}
			nev.RetryGroup = strconv.Itoa(int(parentInfo.CurrentAttempt()))
		}
		nev.SpecNodeId = node.GetID()
		nev.Id.NodeId = currentNodeUniqueID
		nev.NodeName = node.GetName()
	}

	eInfo := info.GetInfo()
	if eInfo != nil {
		if eInfo.WorkflowNodeInfo != nil {
			v := ToNodeExecWorkflowNodeMetadata(eInfo.WorkflowNodeInfo)
			if v != nil {
				nev.TargetMetadata = v
			}
		} else if eInfo.TaskNodeInfo != nil {
			v := ToNodeExecTaskNodeMetadata(eInfo.TaskNodeInfo)
			if v != nil {
				nev.TargetMetadata = v
			}
		}
	}
	if eInfo != nil && eInfo.OutputInfo != nil {
		nev.OutputResult = ToNodeExecOutput(eInfo.OutputInfo)
	} else if info.GetErr() != nil {
		nev.OutputResult = &event.NodeExecutionEvent_Error{
			Error: info.GetErr(),
		}
	}
	if node.GetKind() == v1alpha1.NodeKindWorkflow && node.GetWorkflowNode() != nil && node.GetWorkflowNode().GetSubWorkflowRef() != nil {
		nev.IsParent = true
	} else if dynamicNodePhase != v1alpha1.DynamicNodePhaseNone {
		nev.IsDynamic = true
		if nev.GetTaskNodeMetadata() != nil && nev.GetTaskNodeMetadata().DynamicWorkflow != nil {
			nev.IsParent = true
		}
	}
	return nev, nil
}

func ToNodePhase(p handler.EPhase) (v1alpha1.NodePhase, error) {
	switch p {
	case handler.EPhaseNotReady:
		return v1alpha1.NodePhaseNotYetStarted, nil
	case handler.EPhaseQueued:
		return v1alpha1.NodePhaseQueued, nil
	case handler.EPhaseRunning:
		return v1alpha1.NodePhaseRunning, nil
	case handler.EPhaseDynamicRunning:
		return v1alpha1.NodePhaseDynamicRunning, nil
	case handler.EPhaseRetryableFailure:
		return v1alpha1.NodePhaseRetryableFailure, nil
	case handler.EPhaseSkip:
		return v1alpha1.NodePhaseSkipped, nil
	case handler.EPhaseSuccess:
		return v1alpha1.NodePhaseSucceeding, nil
	case handler.EPhaseFailed:
		return v1alpha1.NodePhaseFailing, nil
	case handler.EPhaseTimedout:
		return v1alpha1.NodePhaseTimingOut, nil
	case handler.EPhaseRecovered:
		return v1alpha1.NodePhaseRecovered, nil
	}
	return v1alpha1.NodePhaseNotYetStarted, fmt.Errorf("no known conversion from handlerPhase[%d] to NodePhase", p)
}

func ToK8sTime(t time.Time) v1.Time {
	return v1.Time{Time: t}
}

func UpdateNodeStatus(np v1alpha1.NodePhase, p handler.PhaseInfo, n *nodeStateManager, s v1alpha1.ExecutableNodeStatus) {
	// We update the phase only if it is not already updated
	if np != s.GetPhase() {
		s.UpdatePhase(np, ToK8sTime(p.GetOccurredAt()), p.GetReason(), p.GetErr())
	}
	// Update TaskStatus
	if n.t != nil {
		t := s.GetOrCreateTaskStatus()
		t.SetPhaseVersion(n.t.PluginPhaseVersion)
		t.SetPhase(int(n.t.PluginPhase))
		t.SetLastPhaseUpdatedAt(n.t.LastPhaseUpdatedAt)
		t.SetPluginState(n.t.PluginState)
		t.SetPluginStateVersion(n.t.PluginStateVersion)
		t.SetBarrierClockTick(n.t.BarrierClockTick)
	}

	// Update dynamic node status
	if n.d != nil {
		t := s.GetOrCreateDynamicNodeStatus()
		t.SetDynamicNodePhase(n.d.Phase)
		t.SetDynamicNodeReason(n.d.Reason)
		t.SetExecutionError(n.d.Error)
	}

	// Update branch node status
	if n.b != nil {
		t := s.GetOrCreateBranchStatus()
		if n.b.Phase == v1alpha1.BranchNodeError {
			t.SetBranchNodeError()
		} else if n.b.FinalizedNodeID != nil {
			t.SetBranchNodeSuccess(*n.b.FinalizedNodeID)
		} else {
			logger.Warnf(context.TODO(), "branch node status neither success nor error set")
		}
	}

	// Update workflow node status
	if n.w != nil {
		t := s.GetOrCreateWorkflowStatus()
		t.SetWorkflowNodePhase(n.w.Phase)
		t.SetExecutionError(n.w.Error)
	}
}
