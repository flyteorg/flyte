package array

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/golang/protobuf/ptypes"

	idlcore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/encoding"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/common"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task"
)

type arrayEventRecorder interface {
	interfaces.EventRecorder
	process(ctx context.Context, nCtx interfaces.NodeExecutionContext, index int, retryAttempt uint32) error
	finalize(ctx context.Context, nCtx interfaces.NodeExecutionContext, taskPhase idlcore.TaskExecution_Phase, taskPhaseVersion uint32, eventConfig *config.EventConfig) error
	finalizeRequired(ctx context.Context) bool
}

type externalResourcesEventRecorder struct {
	interfaces.EventRecorder
	externalResources []*event.ExternalResourceInfo
	nodeEvents        []*event.NodeExecutionEvent
	taskEvents        []*event.TaskExecutionEvent
}

func (e *externalResourcesEventRecorder) RecordNodeEvent(ctx context.Context, event *event.NodeExecutionEvent, eventConfig *config.EventConfig) error {
	e.nodeEvents = append(e.nodeEvents, event)
	return nil
}

func (e *externalResourcesEventRecorder) RecordTaskEvent(ctx context.Context, event *event.TaskExecutionEvent, eventConfig *config.EventConfig) error {
	e.taskEvents = append(e.taskEvents, event)
	return nil
}

func (e *externalResourcesEventRecorder) process(ctx context.Context, nCtx interfaces.NodeExecutionContext, index int, retryAttempt uint32) error {
	// generate externalResourceID
	currentNodeUniqueID := nCtx.NodeID()
	if nCtx.ExecutionContext().GetEventVersion() != v1alpha1.EventVersion0 {
		var err error
		currentNodeUniqueID, err = common.GenerateUniqueID(nCtx.ExecutionContext().GetParentInfo(), nCtx.NodeID())
		if err != nil {
			return err
		}
	}

	uniqueID, err := encoding.FixedLengthUniqueIDForParts(task.IDMaxLength, []string{nCtx.NodeExecutionMetadata().GetOwnerID().Name, currentNodeUniqueID, strconv.Itoa(int(retryAttempt))})
	if err != nil {
		return err
	}

	externalResourceID := fmt.Sprintf("%s-n%d-%d", uniqueID, index, retryAttempt)

	// process events
	cacheStatus := idlcore.CatalogCacheStatus_CACHE_DISABLED
	for _, nodeExecutionEvent := range e.nodeEvents {
		switch target := nodeExecutionEvent.TargetMetadata.(type) {
		case *event.NodeExecutionEvent_TaskNodeMetadata:
			if target.TaskNodeMetadata != nil {
				cacheStatus = target.TaskNodeMetadata.CacheStatus
			}
		}
	}

	// fastcache will not emit task events for cache hits. we need to manually detect a
	// transition to `SUCCEEDED` and add an `ExternalResourceInfo` for it.
	if cacheStatus == idlcore.CatalogCacheStatus_CACHE_HIT && len(e.taskEvents) == 0 {
		e.externalResources = append(e.externalResources, &event.ExternalResourceInfo{
			ExternalId:   externalResourceID,
			Index:        uint32(index),
			RetryAttempt: retryAttempt,
			Phase:        idlcore.TaskExecution_SUCCEEDED,
			CacheStatus:  cacheStatus,
		})
	}

	for _, taskExecutionEvent := range e.taskEvents {
		for _, log := range taskExecutionEvent.Logs {
			log.Name = fmt.Sprintf("%s-%d", log.Name, index)
		}

		e.externalResources = append(e.externalResources, &event.ExternalResourceInfo{
			ExternalId:   externalResourceID,
			Index:        uint32(index),
			Logs:         taskExecutionEvent.Logs,
			RetryAttempt: retryAttempt,
			Phase:        taskExecutionEvent.Phase,
			CacheStatus:  cacheStatus,
		})
	}

	// clear nodeEvents and taskEvents
	e.nodeEvents = e.nodeEvents[:0]
	e.taskEvents = e.taskEvents[:0]

	return nil
}

func (e *externalResourcesEventRecorder) finalize(ctx context.Context, nCtx interfaces.NodeExecutionContext,
	taskPhase idlcore.TaskExecution_Phase, taskPhaseVersion uint32, eventConfig *config.EventConfig) error {

	// build TaskExecutionEvent
	occurredAt, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		return err
	}

	var taskID *idlcore.Identifier
	subNode := nCtx.Node().GetArrayNode().GetSubNodeSpec()
	if subNode != nil && subNode.Kind == v1alpha1.NodeKindTask {
		executableTask, err := nCtx.ExecutionContext().GetTask(*subNode.GetTaskID())
		if err != nil {
			return err
		}

		taskID = executableTask.CoreTask().GetId()
	}

	nodeExecutionID := *nCtx.NodeExecutionMetadata().GetNodeExecutionID()
	if nCtx.ExecutionContext().GetEventVersion() != v1alpha1.EventVersion0 {
		currentNodeUniqueID, err := common.GenerateUniqueID(nCtx.ExecutionContext().GetParentInfo(), nodeExecutionID.NodeId)
		if err != nil {
			return err
		}
		nodeExecutionID.NodeId = currentNodeUniqueID
	}

	taskExecutionEvent := &event.TaskExecutionEvent{
		TaskId:                taskID,
		ParentNodeExecutionId: &nodeExecutionID,
		RetryAttempt:          0, // ArrayNode will never retry
		Phase:                 taskPhase,
		PhaseVersion:          taskPhaseVersion,
		OccurredAt:            occurredAt,
		Metadata: &event.TaskExecutionMetadata{
			ExternalResources: e.externalResources,
			PluginIdentifier:  "k8s-array",
		},
		TaskType:     "container_array",
		EventVersion: 1,
	}

	// only attach input values if taskPhase is QUEUED meaning this the first evaluation
	if taskPhase == idlcore.TaskExecution_QUEUED {
		if eventConfig.RawOutputPolicy == config.RawOutputPolicyInline {
			// pass inputs by value
			literalMap, err := nCtx.InputReader().Get(ctx)
			if err != nil {
				return err
			}

			taskExecutionEvent.InputValue = &event.TaskExecutionEvent_InputData{
				InputData: literalMap,
			}
		} else {
			// pass inputs by reference
			taskExecutionEvent.InputValue = &event.TaskExecutionEvent_InputUri{
				InputUri: nCtx.InputReader().GetInputPath().String(),
			}
		}
	}

	// only attach output uri if taskPhase is SUCCEEDED
	if taskPhase == idlcore.TaskExecution_SUCCEEDED {
		taskExecutionEvent.OutputResult = &event.TaskExecutionEvent_OutputUri{
			OutputUri: v1alpha1.GetOutputsFile(nCtx.NodeStatus().GetOutputDir()).String(),
		}
	}

	// record TaskExecutionEvent
	return e.EventRecorder.RecordTaskEvent(ctx, taskExecutionEvent, eventConfig)
}

func (e *externalResourcesEventRecorder) finalizeRequired(ctx context.Context) bool {
	return len(e.externalResources) > 0
}

type passThroughEventRecorder struct {
	interfaces.EventRecorder
}

func (*passThroughEventRecorder) process(ctx context.Context, nCtx interfaces.NodeExecutionContext, index int, retryAttempt uint32) error {
	return nil
}

func (*passThroughEventRecorder) finalize(ctx context.Context, nCtx interfaces.NodeExecutionContext,
	taskPhase idlcore.TaskExecution_Phase, taskPhaseVersion uint32, eventConfig *config.EventConfig) error {
	return nil
}

func (*passThroughEventRecorder) finalizeRequired(ctx context.Context) bool {
	return false
}

func newArrayEventRecorder(eventRecorder interfaces.EventRecorder) arrayEventRecorder {
	if config.GetConfig().ArrayNodeEventVersion == 0 {
		return &externalResourcesEventRecorder{
			EventRecorder: eventRecorder,
		}
	}

	return &passThroughEventRecorder{
		EventRecorder: eventRecorder,
	}
}

func sendEvents(ctx context.Context, nCtx interfaces.NodeExecutionContext, index int, retryAttempt uint32, nodePhase idlcore.NodeExecution_Phase,
	taskPhase idlcore.TaskExecution_Phase, eventRecorder interfaces.EventRecorder, eventConfig *config.EventConfig) error {

	subNodeID := buildSubNodeID(nCtx, index)
	timestamp := ptypes.TimestampNow()
	workflowExecutionID := nCtx.ExecutionContext().GetExecutionID().WorkflowExecutionIdentifier

	// send NodeExecutionEvent
	nodeExecutionEvent := &event.NodeExecutionEvent{
		Id: &idlcore.NodeExecutionIdentifier{
			NodeId:      subNodeID,
			ExecutionId: workflowExecutionID,
		},
		Phase:      nodePhase,
		OccurredAt: timestamp,
		ParentNodeMetadata: &event.ParentNodeExecutionMetadata{
			NodeId: nCtx.NodeID(),
		},
		ReportedAt: timestamp,
	}

	if err := eventRecorder.RecordNodeEvent(ctx, nodeExecutionEvent, eventConfig); err != nil {
		return err
	}

	// send TaskExeucutionEvent
	taskExecutionEvent := &event.TaskExecutionEvent{
		TaskId: &idlcore.Identifier{
			ResourceType: idlcore.ResourceType_TASK,
			Project:      workflowExecutionID.Project,
			Domain:       workflowExecutionID.Domain,
			Name:         fmt.Sprintf("%s-%d", buildSubNodeID(nCtx, index), retryAttempt),
			Version:      "v1", // this value is irrelevant but necessary for the identifier to be valid
		},
		ParentNodeExecutionId: nodeExecutionEvent.Id,
		Phase:                 taskPhase,
		TaskType:              "k8s-array",
		OccurredAt:            timestamp,
		ReportedAt:            timestamp,
	}

	if err := eventRecorder.RecordTaskEvent(ctx, taskExecutionEvent, eventConfig); err != nil {
		return err
	}

	return nil
}
