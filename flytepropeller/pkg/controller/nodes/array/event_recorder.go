package array

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/golang/protobuf/ptypes"

	idlcore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	events "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
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

func (t *taskExecutionID) GetGeneratedNameWith(minLength, maxLength int) (string, error) {
	return "", nil
}

func (t *taskExecutionID) GetUniqueNodeID() string {
	return t.nodeID
}

//go:generate mockery -all -case=underscore

type ArrayEventRecorder interface {
	interfaces.EventRecorder
	process(ctx context.Context, nCtx interfaces.NodeExecutionContext, index int, retryAttempt uint32) error
	finalize(ctx context.Context, nCtx interfaces.NodeExecutionContext, taskPhase idlcore.TaskExecution_Phase, taskPhaseVersion uint32, eventConfig *config.EventConfig, arrayNodeExecutionError *idlcore.ExecutionError) error
	finalizeRequired(ctx context.Context) bool
}

func mapNodeExecutionPhaseToTaskExecutionPhase(nodePhase idlcore.NodeExecution_Phase) idlcore.TaskExecution_Phase {
	switch nodePhase {
	case idlcore.NodeExecution_UNDEFINED:
		return idlcore.TaskExecution_UNDEFINED
	case idlcore.NodeExecution_QUEUED:
		return idlcore.TaskExecution_QUEUED
	case idlcore.NodeExecution_RUNNING, idlcore.NodeExecution_DYNAMIC_RUNNING:
		return idlcore.TaskExecution_RUNNING
	case idlcore.NodeExecution_SUCCEEDED:
		return idlcore.TaskExecution_SUCCEEDED
	case idlcore.NodeExecution_FAILING, idlcore.NodeExecution_FAILED, idlcore.NodeExecution_TIMED_OUT:
		return idlcore.TaskExecution_FAILED
	case idlcore.NodeExecution_ABORTED:
		return idlcore.TaskExecution_ABORTED
	case idlcore.NodeExecution_SKIPPED:
		return idlcore.TaskExecution_UNDEFINED
	case idlcore.NodeExecution_RECOVERED:
		return idlcore.TaskExecution_SUCCEEDED
	default:
		return idlcore.TaskExecution_UNDEFINED
	}
}

type externalResourcesEventRecorder struct {
	interfaces.EventRecorder
	externalResources []*events.ExternalResourceInfo
	nodeEvents        []*events.NodeExecutionEvent
	taskEvents        []*events.TaskExecutionEvent
	expectTaskEvents  bool
}

func (e *externalResourcesEventRecorder) RecordNodeEvent(ctx context.Context, event *events.NodeExecutionEvent, eventConfig *config.EventConfig) error {
	e.nodeEvents = append(e.nodeEvents, event)
	return nil
}

func (e *externalResourcesEventRecorder) RecordTaskEvent(ctx context.Context, event *events.TaskExecutionEvent, eventConfig *config.EventConfig) error {
	e.taskEvents = append(e.taskEvents, event)
	return nil
}

func (e *externalResourcesEventRecorder) process(ctx context.Context, nCtx interfaces.NodeExecutionContext, index int, retryAttempt uint32) error {
	externalResourceID, err := generateExternalResourceID(nCtx, index, retryAttempt)
	if err != nil {
		return err
	}

	// process events
	cacheStatus := idlcore.CatalogCacheStatus_CACHE_DISABLED
	for _, nodeExecutionEvent := range e.nodeEvents {
		switch target := nodeExecutionEvent.TargetMetadata.(type) {
		case *events.NodeExecutionEvent_TaskNodeMetadata:
			if target.TaskNodeMetadata != nil {
				cacheStatus = target.TaskNodeMetadata.CacheStatus
			}
		}
	}

	if len(e.taskEvents) == 0 {
		// fastcache will not emit task events for cache hits. we need to manually detect a
		// transition to `SUCCEEDED` and add an `ExternalResourceInfo` for it.
		if cacheStatus == idlcore.CatalogCacheStatus_CACHE_HIT {
			e.externalResources = append(e.externalResources, &events.ExternalResourceInfo{
				ExternalId:   externalResourceID,
				Index:        uint32(index),
				RetryAttempt: retryAttempt,
				Phase:        idlcore.TaskExecution_SUCCEEDED,
				CacheStatus:  cacheStatus,
			})
		} else if !e.expectTaskEvents {
			// if no TaskExecutionEvents are expected we need to manually process the node events
			// to populate ExternalResourceInfo
			for _, nodeExecutionEvent := range e.nodeEvents {
				var targetMetadata *events.ExternalResourceInfo_WorkflowNodeMetadata
				if nodeExecutionEvent.GetWorkflowNodeMetadata() != nil && nodeExecutionEvent.GetWorkflowNodeMetadata().ExecutionId != nil {
					targetMetadata = &events.ExternalResourceInfo_WorkflowNodeMetadata{
						WorkflowNodeMetadata: nodeExecutionEvent.GetWorkflowNodeMetadata(),
					}
				}

				e.externalResources = append(e.externalResources, &events.ExternalResourceInfo{
					ExternalId: externalResourceID,
					Index:      uint32(index),
					Logs: []*idlcore.TaskLog{
						{
							Name: nodeExecutionEvent.NodeName,
						},
					},
					RetryAttempt:   retryAttempt,
					Phase:          mapNodeExecutionPhaseToTaskExecutionPhase(nodeExecutionEvent.Phase),
					CacheStatus:    cacheStatus,
					TargetMetadata: targetMetadata,
				})
			}
		}
	}

	var mapLogPlugin tasklog.Plugin
	if config.GetConfig().ArrayNode.UseMapPluginLogs {
		mapLogPlugin, err = logs.InitializeLogPlugins(&mapplugin.GetConfig().LogConfig.Config)
		if err != nil {
			logger.Warnf(ctx, "failed to initialize log plugin with error:%v", err)
		}
	}

	for _, taskExecutionEvent := range e.taskEvents {
		if mapLogPlugin != nil && len(taskExecutionEvent.Logs) > 0 {
			// override log links for subNode execution with map plugin
			var taskExecGeneratedName string
			if taskExecutionEvent.Metadata != nil {
				taskExecGeneratedName = taskExecutionEvent.Metadata.GeneratedName
			}
			containerName := findPodContainerInLogContext(taskExecutionEvent.GetLogContext(), taskExecGeneratedName)
			logs, err := getPluginLogs(mapLogPlugin, nCtx, index, retryAttempt, taskExecGeneratedName, containerName)
			if err != nil {
				logger.Warnf(ctx, "failed to compute logs for ArrayNode:%s index:%d retryAttempt:%d with error:%v", nCtx.NodeID(), index, retryAttempt, err)
			} else {
				taskExecutionEvent.Logs = logs
			}
		}

		for _, log := range taskExecutionEvent.Logs {
			log.Name = fmt.Sprintf("%s-%d", log.Name, index)
		}

		externalResourceInfo := events.ExternalResourceInfo{
			ExternalId:   externalResourceID,
			Index:        uint32(index),
			Logs:         taskExecutionEvent.Logs,
			LogContext:   taskExecutionEvent.LogContext,
			RetryAttempt: retryAttempt,
			Phase:        taskExecutionEvent.Phase,
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
	taskPhase idlcore.TaskExecution_Phase, taskPhaseVersion uint32, eventConfig *config.EventConfig, arrayNodeExecutionError *idlcore.ExecutionError) error {

	// build task Identifier
	var taskID *idlcore.Identifier
	subNode := nCtx.Node().GetArrayNode().GetSubNodeSpec()
	if subNode == nil {
		return nil
	}

	switch subNode.Kind {
	case v1alpha1.NodeKindTask:
		executableTask, err := nCtx.ExecutionContext().GetTask(*subNode.GetTaskID())
		if err != nil {
			return err
		}

		taskID = executableTask.CoreTask().GetId()
	case v1alpha1.NodeKindWorkflow:
		launchPlanRefID := subNode.GetWorkflowNode().GetLaunchPlanRefID()
		taskID = &idlcore.Identifier{
			ResourceType: idlcore.ResourceType_TASK,
			Project:      launchPlanRefID.Project,
			Domain:       launchPlanRefID.Domain,
			Name:         launchPlanRefID.Name,
			Version:      launchPlanRefID.Version,
			Org:          launchPlanRefID.Org,
		}
	default:
		return fmt.Errorf("unsupported subNode kind [%v]", subNode.Kind)
	}

	// build TaskExecutionEvent
	occurredAt, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		return err
	}

	nodeExecutionID := *nCtx.NodeExecutionMetadata().GetNodeExecutionID()
	if nCtx.ExecutionContext().GetEventVersion() != v1alpha1.EventVersion0 {
		currentNodeUniqueID, err := common.GenerateUniqueID(nCtx.ExecutionContext().GetParentInfo(), nodeExecutionID.NodeId)
		if err != nil {
			return err
		}
		nodeExecutionID.NodeId = currentNodeUniqueID
	}

	taskExecutionEvent := &events.TaskExecutionEvent{
		TaskId:                taskID,
		ParentNodeExecutionId: &nodeExecutionID,
		RetryAttempt:          0, // ArrayNode will never retry
		Phase:                 taskPhase,
		PhaseVersion:          taskPhaseVersion,
		OccurredAt:            occurredAt,
		Metadata: &events.TaskExecutionMetadata{
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

			taskExecutionEvent.InputValue = &events.TaskExecutionEvent_InputData{
				InputData: literalMap,
			}
		} else {
			// pass inputs by reference
			taskExecutionEvent.InputValue = &events.TaskExecutionEvent_InputUri{
				InputUri: nCtx.InputReader().GetInputPath().String(),
			}
		}
	}

	// only attach output uri if taskPhase is SUCCEEDED
	if taskPhase == idlcore.TaskExecution_SUCCEEDED {
		taskExecutionEvent.OutputResult = &events.TaskExecutionEvent_OutputUri{
			OutputUri: v1alpha1.GetOutputsFile(nCtx.NodeStatus().GetOutputDir()).String(),
		}
	} else if taskPhase == idlcore.TaskExecution_FAILED {
		// attach first evaluated error(s) if taskPhase is FAILED
		taskExecutionEvent.OutputResult = &events.TaskExecutionEvent_Error{
			Error: arrayNodeExecutionError,
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
	taskPhase idlcore.TaskExecution_Phase, taskPhaseVersion uint32, eventConfig *config.EventConfig, arrayNodeExecutionError *idlcore.ExecutionError) error {
	return nil
}

func (*passThroughEventRecorder) finalizeRequired(ctx context.Context) bool {
	return false
}

func newArrayEventRecorder(eventRecorder interfaces.EventRecorder, expectTaskEvents bool) ArrayEventRecorder {
	if config.GetConfig().ArrayNode.EventVersion == 0 {
		return &externalResourcesEventRecorder{
			EventRecorder:    eventRecorder,
			expectTaskEvents: expectTaskEvents,
		}
	}

	return &passThroughEventRecorder{
		EventRecorder: eventRecorder,
	}
}

func getPluginLogs(logPlugin tasklog.Plugin, nCtx interfaces.NodeExecutionContext, index int, retryAttempt uint32, podName, containerName string) ([]*idlcore.TaskLog, error) {
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

	taskExecutionId := &taskExecutionID{
		generatedName: uniqueID,
		id:            taskExecutionIdentifier,
		nodeID:        nodeID,
	}

	if podName == "" {
		subNodeID := fmt.Sprintf("n%v", strconv.Itoa(index))
		subNodeRetryAttemptStr := strconv.FormatUint(uint64(retryAttempt), 10)
		podName, err = encoding.FixedLengthUniqueIDForParts(length, []string{uniqueID, subNodeID, subNodeRetryAttemptStr})
		if err != nil {
			return nil, err
		}
	}
	if containerName == "" {
		containerName = podName
	}

	// initialize map plugin specific LogTemplateVars
	extraLogTemplateVars := []tasklog.TemplateVar{
		{
			Regex: mapplugin.LogTemplateRegexes.ExecutionIndex,
			Value: strconv.FormatUint(uint64(index), 10),
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
			TaskExecutionID:   taskExecutionId,
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
	nodeExecutionEvent := &events.NodeExecutionEvent{
		Id: &idlcore.NodeExecutionIdentifier{
			NodeId:      subNodeID,
			ExecutionId: workflowExecutionID,
		},
		Phase:      nodePhase,
		OccurredAt: timestamp,
		ParentNodeMetadata: &events.ParentNodeExecutionMetadata{
			NodeId: nCtx.NodeID(),
		},
		ReportedAt:       timestamp,
		IsInDynamicChain: dynamic,
	}

	if err := eventRecorder.RecordNodeEvent(ctx, nodeExecutionEvent, eventConfig); err != nil {
		return err
	}

	// send TaskExecutionEvent
	taskExecutionEvent := &events.TaskExecutionEvent{
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

func findPodContainerInLogContext(logCtx *idlcore.LogContext, podName string) string {
	for _, podLogCtx := range logCtx.GetPods() {
		if podLogCtx.PodName == podName {
			return podLogCtx.GetPrimaryContainerName()
		}
	}
	return ""
}

func generateExternalResourceID(nCtx interfaces.NodeExecutionContext, index int, retryAttempt uint32) (string, error) {
	currentNodeUniqueID := nCtx.NodeID()
	if nCtx.ExecutionContext().GetEventVersion() != v1alpha1.EventVersion0 {
		var err error
		currentNodeUniqueID, err = common.GenerateUniqueID(nCtx.ExecutionContext().GetParentInfo(), nCtx.NodeID())
		if err != nil {
			return "", err
		}
	}

	uniqueID, err := encoding.FixedLengthUniqueIDForParts(task.IDMaxLength, []string{nCtx.NodeExecutionMetadata().GetOwnerID().Name, currentNodeUniqueID, strconv.Itoa(int(nCtx.CurrentAttempt()))})
	if err != nil {
		return "", err
	}

	// Note: if the subNode is a task, then this value should map to the subNode's pod name. Services such as
	// usage utilize ExternalResourceInfo.ExternalId to identify the pod.

	externalResourceID := fmt.Sprintf("%s-n%d-%d", uniqueID, index, retryAttempt)

	return externalResourceID, nil
}
