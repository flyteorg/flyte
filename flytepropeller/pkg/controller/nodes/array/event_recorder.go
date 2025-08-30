package array

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/golang/protobuf/ptypes"

	idlcore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/logs"
	pluginscore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/encoding"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/tasklog"
	mapplugin "github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/array/k8s"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/pod"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/common"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

type taskExecutionID struct {
	pluginscore.TaskExecutionID

	generatedName string
	id            *idlcore.TaskExecutionIdentifier
	nodeID        string
}

func (t *taskExecutionID) GetGeneratedName() string {
	return t.generatedName
}

func (t *taskExecutionID) GetID() idlcore.TaskExecutionIdentifier {
	return *t.id
}

func (t *taskExecutionID) GetGeneratedNameWith(int, int) (string, error) {
	return "", nil
}

func (t *taskExecutionID) GetUniqueNodeID() string {
	return t.nodeID
}

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

func (e *externalResourcesEventRecorder) RecordNodeEvent(_ context.Context, event *event.NodeExecutionEvent, _ *config.EventConfig) error {
	e.nodeEvents = append(e.nodeEvents, event)
	return nil
}

func (e *externalResourcesEventRecorder) RecordTaskEvent(_ context.Context, event *event.TaskExecutionEvent, _ *config.EventConfig) error {
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
		switch target := nodeExecutionEvent.GetTargetMetadata().(type) {
		case *event.NodeExecutionEvent_TaskNodeMetadata:
			if target.TaskNodeMetadata != nil {
				cacheStatus = target.TaskNodeMetadata.GetCacheStatus()
			}
		}
	}

	// fastcache will not emit task events for cache hits. we need to manually detect a
	// transition to `SUCCEEDED` and add an `ExternalResourceInfo` for it.
	if cacheStatus == idlcore.CatalogCacheStatus_CACHE_HIT && len(e.taskEvents) == 0 {
		e.externalResources = append(e.externalResources, &event.ExternalResourceInfo{
			ExternalId:   externalResourceID,
			Index:        uint32(index), // #nosec G115
			RetryAttempt: retryAttempt,
			Phase:        idlcore.TaskExecution_SUCCEEDED,
			CacheStatus:  cacheStatus,
		})
	}

	var mapLogPlugin tasklog.Plugin
	if config.GetConfig().ArrayNode.UseMapPluginLogs {
		mapLogPlugin, err = logs.InitializeLogPlugins(&mapplugin.GetConfig().LogConfig.Config)
		if err != nil {
			logger.Warnf(ctx, "failed to initialize log plugin with error:%v", err)
		}
	}

	for _, taskExecutionEvent := range e.taskEvents {
		if mapLogPlugin != nil && len(taskExecutionEvent.GetLogs()) > 0 {
			// override log links for subNode execution with map plugin
			logs, err := getPluginLogs(mapLogPlugin, nCtx, index, retryAttempt)
			if err != nil {
				logger.Warnf(ctx, "failed to compute logs for ArrayNode:%s index:%d retryAttempt:%d with error:%v", nCtx.NodeID(), index, retryAttempt, err)
			} else {
				taskExecutionEvent.Logs = logs
			}
		}

		for _, log := range taskExecutionEvent.GetLogs() {
			log.Name = fmt.Sprintf("%s-%d", log.GetName(), index)
		}

		externalResourceInfo := event.ExternalResourceInfo{
			ExternalId:   externalResourceID,
			Index:        uint32(index), // #nosec G115
			Logs:         taskExecutionEvent.GetLogs(),
			RetryAttempt: retryAttempt,
			Phase:        taskExecutionEvent.GetPhase(),
			CacheStatus:  cacheStatus,
			CustomInfo:   taskExecutionEvent.GetCustomInfo(),
		}

		if taskExecutionEvent.GetMetadata() != nil && len(taskExecutionEvent.GetMetadata().GetExternalResources()) == 1 {
			externalResourceInfo.CustomInfo = taskExecutionEvent.GetMetadata().GetExternalResources()[0].GetCustomInfo()
		}

		e.externalResources = append(e.externalResources, &externalResourceInfo)
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
		currentNodeUniqueID, err := common.GenerateUniqueID(nCtx.ExecutionContext().GetParentInfo(), nodeExecutionID.GetNodeId())
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

func (e *externalResourcesEventRecorder) finalizeRequired(context.Context) bool {
	return len(e.externalResources) > 0
}

type passThroughEventRecorder struct {
	interfaces.EventRecorder
}

func (*passThroughEventRecorder) process(context.Context, interfaces.NodeExecutionContext, int, uint32) error {
	return nil
}

func (*passThroughEventRecorder) finalize(context.Context, interfaces.NodeExecutionContext, idlcore.TaskExecution_Phase, uint32, *config.EventConfig) error {
	return nil
}

func (*passThroughEventRecorder) finalizeRequired(context.Context) bool {
	return false
}

func newArrayEventRecorder(eventRecorder interfaces.EventRecorder) arrayEventRecorder {
	if config.GetConfig().ArrayNode.EventVersion == 0 {
		return &externalResourcesEventRecorder{
			EventRecorder: eventRecorder,
		}
	}

	return &passThroughEventRecorder{
		EventRecorder: eventRecorder,
	}
}

func getPluginLogs(logPlugin tasklog.Plugin, nCtx interfaces.NodeExecutionContext, index int, retryAttempt uint32) ([]*idlcore.TaskLog, error) {
	subNodeSpec := nCtx.Node().GetArrayNode().GetSubNodeSpec()

	// retrieve taskTemplate from subNode
	taskID := subNodeSpec.GetTaskID()
	executableTask, err := nCtx.ExecutionContext().GetTask(*taskID)
	if err != nil {
		return nil, err
	}

	taskTemplate := executableTask.CoreTask()

	// build TaskExecutionID
	taskExecutionIdentifier := &idlcore.TaskExecutionIdentifier{
		TaskId:          taskTemplate.GetId(), // use taskID from subNodeSpec
		RetryAttempt:    nCtx.CurrentAttempt(),
		NodeExecutionId: nCtx.NodeExecutionMetadata().GetNodeExecutionID(), // use node metadata from ArrayNode
	}

	nodeID := nCtx.NodeID()
	if nCtx.ExecutionContext().GetEventVersion() != v1alpha1.EventVersion0 {
		var err error
		nodeID, err = common.GenerateUniqueID(nCtx.ExecutionContext().GetParentInfo(), nCtx.NodeID())
		if err != nil {
			return nil, err
		}
	}

	length := task.IDMaxLength
	if l := pod.DefaultPodPlugin.GetProperties().GeneratedNameMaxLength; l != nil {
		length = *l
	}

	uniqueID, err := encoding.FixedLengthUniqueIDForParts(length, []string{nCtx.NodeExecutionMetadata().GetOwnerID().Name, nodeID, strconv.Itoa(int(nCtx.CurrentAttempt()))})
	if err != nil {
		return nil, err
	}

	taskExecID := &taskExecutionID{
		generatedName: uniqueID,
		id:            taskExecutionIdentifier,
		nodeID:        nodeID,
	}

	// compute podName and containerName
	stCtx := mapplugin.NewSubTaskExecutionID(taskExecID, index, uint64(retryAttempt))

	podName := stCtx.GetGeneratedName()
	containerName := stCtx.GetGeneratedName()

	// initialize map plugin specific LogTemplateVars
	extraLogTemplateVars := []tasklog.TemplateVar{
		{
			Regex: mapplugin.LogTemplateRegexes.ExecutionIndex,
			Value: strconv.FormatUint(uint64(index), 10), // #nosec G115
		},
		{
			Regex: mapplugin.LogTemplateRegexes.RetryAttempt,
			Value: strconv.FormatUint(uint64(retryAttempt), 10),
		},
	}

	logs, err := logPlugin.GetTaskLogs(
		tasklog.Input{
			PodName:           podName,
			Namespace:         nCtx.NodeExecutionMetadata().GetNamespace(),
			ContainerName:     containerName,
			TaskExecutionID:   taskExecID,
			ExtraTemplateVars: extraLogTemplateVars,
			TaskTemplate:      taskTemplate,
		},
	)
	if err != nil {
		return nil, err
	}

	return logs.TaskLogs, nil
}

func sendEvents(ctx context.Context, nCtx interfaces.NodeExecutionContext, index int, retryAttempt uint32, nodePhase idlcore.NodeExecution_Phase,
	taskPhase idlcore.TaskExecution_Phase, eventRecorder interfaces.EventRecorder, eventConfig *config.EventConfig) error {

	subNodeID := buildSubNodeID(nCtx, index)
	timestamp := ptypes.TimestampNow()
	workflowExecutionID := nCtx.ExecutionContext().GetExecutionID().WorkflowExecutionIdentifier

	// Extract dynamic chain information.
	var dynamic = false
	if nCtx.ExecutionContext() != nil && nCtx.ExecutionContext().GetParentInfo() != nil && nCtx.ExecutionContext().GetParentInfo().IsInDynamicChain() {
		dynamic = true
	}
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
		ReportedAt:       timestamp,
		IsInDynamicChain: dynamic,
	}

	if err := eventRecorder.RecordNodeEvent(ctx, nodeExecutionEvent, eventConfig); err != nil {
		return err
	}

	// send TaskExecutionEvent
	taskExecutionEvent := &event.TaskExecutionEvent{
		TaskId: &idlcore.Identifier{
			ResourceType: idlcore.ResourceType_TASK,
			Project:      workflowExecutionID.GetProject(),
			Domain:       workflowExecutionID.GetDomain(),
			Name:         fmt.Sprintf("%s-%d", buildSubNodeID(nCtx, index), retryAttempt),
			Version:      "v1", // this value is irrelevant but necessary for the identifier to be valid
		},
		ParentNodeExecutionId: nodeExecutionEvent.GetId(),
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
