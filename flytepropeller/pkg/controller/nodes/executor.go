package nodes

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/lyft/flyteidl/clients/go/events"
	eventsErr "github.com/lyft/flyteidl/clients/go/events/errors"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/promutils/labeled"
	"github.com/lyft/flytestdlib/storage"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/controller/catalog"
	"github.com/lyft/flytepropeller/pkg/controller/executors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/errors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/subworkflow/launchplan"
	"github.com/lyft/flytepropeller/pkg/utils"
)

type nodeMetrics struct {
	FailureDuration    labeled.StopWatch
	SuccessDuration    labeled.StopWatch
	ResolutionFailure  labeled.Counter
	InputsWriteFailure labeled.Counter

	// Measures the latency between the last parent node stoppedAt time and current node's queued time.
	TransitionLatency labeled.StopWatch
	// Measures the latency between the time a node's been queued to the time the handler reported the executable moved
	// to running state
	QueuingLatency         labeled.StopWatch
	NodeExecutionTime      labeled.StopWatch
	NodeInputGatherLatency labeled.StopWatch
}

type nodeExecutor struct {
	nodeHandlerFactory HandlerFactory
	enqueueWorkflow    v1alpha1.EnqueueWorkflow
	store              *storage.DataStore
	nodeRecorder       events.NodeEventRecorder
	metrics            *nodeMetrics
}

// In this method we check if the queue is ready to be processed and if so, we prime it in Admin as queued
// Before we start the node execution, we need to transition this Node status to Queued.
// This is because a node execution has to exist before task/wf executions can start.
func (c *nodeExecutor) queueNodeIfReady(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.BaseNode, nodeStatus v1alpha1.ExecutableNodeStatus) (handler.Status, error) {
	logger.Debugf(ctx, "Node not yet started")
	// Query the nodes information to figure out if it can be executed.
	predicatePhase, err := CanExecute(ctx, w, node)
	if err != nil {
		logger.Debugf(ctx, "Node failed in CanExecute. Error [%s]", err)
		return handler.StatusUndefined, err
	}
	if predicatePhase == PredicatePhaseSkip {
		logger.Debugf(ctx, "Node upstream node was skipped. Skipping!")
		return handler.StatusSkipped, nil
	} else if predicatePhase == PredicatePhaseNotReady {
		logger.Debugf(ctx, "Node not ready for executing.")
		return handler.StatusNotStarted, nil
	}

	if len(nodeStatus.GetDataDir()) == 0 {
		// Predicate ready, lets Resolve the data
		dataDir, err := w.GetExecutionStatus().ConstructNodeDataDir(ctx, c.store, node.GetID())
		if err != nil {
			return handler.StatusUndefined, err
		}

		nodeStatus.SetDataDir(dataDir)
	}

	return handler.StatusQueued, nil
}

func (c *nodeExecutor) RecordTransitionLatency(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode, nodeStatus v1alpha1.ExecutableNodeStatus) {
	if nodeStatus.GetPhase() == v1alpha1.NodePhaseNotYetStarted || nodeStatus.GetPhase() == v1alpha1.NodePhaseQueued {
		// Log transition latency (The most recently finished parent node endAt time to this node's queuedAt time -now-)
		t, err := GetParentNodeMaxEndTime(ctx, w, node)
		if err != nil {
			logger.Warnf(ctx, "Failed to record transition latency for node. Error: %s", err.Error())
			return
		}
		if !t.IsZero() {
			c.metrics.TransitionLatency.Observe(ctx, t.Time, time.Now())
		}
	} else if nodeStatus.GetPhase() == v1alpha1.NodePhaseRetryableFailure && nodeStatus.GetLastUpdatedAt() != nil {
		c.metrics.TransitionLatency.Observe(ctx, nodeStatus.GetLastUpdatedAt().Time, time.Now())
	}
}

// Start the node execution. This implies that the node will start processing
func (c *nodeExecutor) startNode(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode, nodeStatus v1alpha1.ExecutableNodeStatus, h handler.IFace) (handler.Status, error) {

	// TODO: Performance problem, we may be in a retry loop and do not need to resolve the inputs again.
	// For now we will do this.
	dataDir := nodeStatus.GetDataDir()
	var nodeInputs *handler.Data
	if !node.IsStartNode() {
		t := c.metrics.NodeInputGatherLatency.Start(ctx)
		defer t.Stop()
		// Can execute
		var err error
		nodeInputs, err = Resolve(ctx, c.nodeHandlerFactory, w, node.GetID(), node.GetInputBindings(), c.store)
		// TODO we need to handle retryable, network errors here!!
		if err != nil {
			c.metrics.ResolutionFailure.Inc(ctx)
			logger.Warningf(ctx, "Failed to resolve inputs for Node. Error [%v]", err)
			return handler.StatusFailed(err), nil
		}

		if nodeInputs != nil {
			inputsFile := v1alpha1.GetInputsFile(dataDir)
			if err := c.store.WriteProtobuf(ctx, inputsFile, storage.Options{}, nodeInputs); err != nil {
				c.metrics.InputsWriteFailure.Inc(ctx)
				logger.Errorf(ctx, "Failed to store inputs for Node. Error [%v]. InputsFile [%s]", err, inputsFile)
				return handler.StatusUndefined, errors.Wrapf(
					errors.StorageError, node.GetID(), err, "Failed to store inputs for Node. InputsFile [%s]", inputsFile)
			}
		}

		logger.Debugf(ctx, "Node Data Directory [%s].", nodeStatus.GetDataDir())
	}

	// Now that we have resolved the inputs, we can record as a transition latency. This is because we have completed
	// all the overhead that we have to compute. Any failures after this will incur this penalty, but it could be due
	// to various external reasons - like queuing, overuse of quota, plugin overhead etc.
	c.RecordTransitionLatency(ctx, w, node, nodeStatus)

	// Start node execution
	return h.StartNode(ctx, w, node, nodeInputs)
}

func (c *nodeExecutor) handleNode(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode) (handler.Status, error) {
	logger.Debugf(ctx, "Handling Node [%s]", node.GetID())
	defer logger.Debugf(ctx, "Completed node [%s]", node.GetID())
	// Now depending on the node type decide
	h, err := c.nodeHandlerFactory.GetHandler(node.GetKind())
	if err != nil {
		return handler.StatusUndefined, err
	}

	// Important to note that we have special optimization for start node only (not end node)
	// We specifically ignore queueing of start node and directly move the start node to "starting"
	// This prevents an extra event to Admin and an extra write to etcD. This is also because of the fact that start node does not have tasks and do not need to send out task events.

	nodeStatus := w.GetNodeExecutionStatus(node.GetID())
	var status handler.Status
	if !node.IsStartNode() && !node.IsEndNode() && nodeStatus.GetPhase() == v1alpha1.NodePhaseNotYetStarted {
		// We only send the queued event to Admin in case the node was never started and when it is not either StartNode.
		// We do not do this for endNode, because endNode may still be not executable. This is because StartNode
		// completes as soon as started.
		return c.queueNodeIfReady(ctx, w, node, nodeStatus)
	} else if node.IsEndNode() {
		status, err = c.queueNodeIfReady(ctx, w, node, nodeStatus)
		if err == nil && status.Phase == handler.PhaseQueued {
			status, err = c.startNode(ctx, w, node, nodeStatus, h)
		}
	} else if node.IsStartNode() || nodeStatus.GetPhase() == v1alpha1.NodePhaseQueued ||
		nodeStatus.GetPhase() == v1alpha1.NodePhaseRetryableFailure {
		// If the node is either StartNode or was previously queued or failed in a previous attempt, we will call
		// the start method on the node handler
		status, err = c.startNode(ctx, w, node, nodeStatus, h)
	} else if nodeStatus.GetPhase() == v1alpha1.NodePhaseFailing {
		status, err = h.HandleFailingNode(ctx, w, node)
	} else {
		status, err = h.CheckNodeStatus(ctx, w, node, nodeStatus)
	}

	return status, err
}

func (c *nodeExecutor) IdempotentRecordEvent(ctx context.Context, nodeEvent *event.NodeExecutionEvent) error {
	err := c.nodeRecorder.RecordNodeEvent(ctx, nodeEvent)
	// TODO: add unit tests for this specific path
	if err != nil && eventsErr.IsAlreadyExists(err) {
		logger.Infof(ctx, "Node event phase: %s, nodeId %s already exist",
			nodeEvent.Phase.String(), nodeEvent.GetId().NodeId)
		return nil
	}
	return err
}

func (c *nodeExecutor) TransitionToPhase(ctx context.Context, execID *core.WorkflowExecutionIdentifier,
	node v1alpha1.ExecutableNode, nodeStatus v1alpha1.ExecutableNodeStatus, toStatus handler.Status) (executors.NodeStatus, error) {

	previousNodePhase := nodeStatus.GetPhase()
	// TODO GC analysis. We will create a ton of node-events but never publish them. We could first check for the PhaseChange and if so then do this processing

	nodeEvent := &event.NodeExecutionEvent{
		Id: &core.NodeExecutionIdentifier{
			NodeId:      node.GetID(),
			ExecutionId: execID,
		},
		InputUri: v1alpha1.GetInputsFile(nodeStatus.GetDataDir()).String(),
	}

	var returnStatus executors.NodeStatus
	errMsg := ""
	errCode := "NodeExecutionUnknownError"
	if toStatus.Err != nil {
		errMsg = toStatus.Err.Error()
		code, ok := errors.GetErrorCode(toStatus.Err)
		if ok {
			errCode = code.String()
		}
	}

	// If there is a child workflow, include the execution of the child workflow in the event
	if nodeStatus.GetWorkflowNodeStatus() != nil {
		nodeEvent.TargetMetadata = &event.NodeExecutionEvent_WorkflowNodeMetadata{
			WorkflowNodeMetadata: &event.WorkflowNodeMetadata{
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: execID.Project,
					Domain:  execID.Domain,
					Name:    nodeStatus.GetWorkflowNodeStatus().GetWorkflowExecutionName(),
				},
			},
		}
	}

	switch toStatus.Phase {
	case handler.PhaseNotStarted:
		return executors.NodeStatusPending, nil
	// TODO we should not need handler.PhaseQueued since we have added Task StateMachine. Remove it.
	case handler.PhaseQueued:
		nodeEvent.Phase = core.NodeExecution_QUEUED
		nodeStatus.UpdatePhase(v1alpha1.NodePhaseQueued, v1.NewTime(toStatus.OccurredAt), "")

		returnStatus = executors.NodeStatusQueued
		if !toStatus.OccurredAt.IsZero() {
			nodeEvent.OccurredAt = utils.GetProtoTime(&v1.Time{Time: toStatus.OccurredAt})
		} else {
			nodeEvent.OccurredAt = ptypes.TimestampNow() // TODO: add queueAt in nodeStatus
		}

	case handler.PhaseRunning:
		nodeEvent.Phase = core.NodeExecution_RUNNING
		nodeStatus.UpdatePhase(v1alpha1.NodePhaseRunning, v1.NewTime(toStatus.OccurredAt), "")
		nodeEvent.OccurredAt = utils.GetProtoTime(nodeStatus.GetStartedAt())
		returnStatus = executors.NodeStatusRunning

		if nodeStatus.GetQueuedAt() != nil && nodeStatus.GetStartedAt() != nil {
			c.metrics.QueuingLatency.Observe(ctx, nodeStatus.GetQueuedAt().Time, nodeStatus.GetStartedAt().Time)
		}
	case handler.PhaseRetryableFailure:
		maxAttempts := uint32(0)
		if node.GetRetryStrategy() != nil && node.GetRetryStrategy().MinAttempts != nil {
			maxAttempts = uint32(*node.GetRetryStrategy().MinAttempts)
		}

		nodeEvent.OutputResult = &event.NodeExecutionEvent_Error{
			Error: &core.ExecutionError{
				Code:     errCode,
				Message:  fmt.Sprintf("Retries [%d/%d], %s", nodeStatus.GetAttempts(), maxAttempts, errMsg),
				ErrorUri: v1alpha1.GetOutputErrorFile(nodeStatus.GetDataDir()).String(),
			},
		}

		if nodeStatus.IncrementAttempts() >= maxAttempts {
			logger.Debugf(ctx, "All retries have exhausted, failing node. [%d/%d]", nodeStatus.GetAttempts(), maxAttempts)
			// Failure
			nodeEvent.Phase = core.NodeExecution_FAILED
			nodeStatus.UpdatePhase(v1alpha1.NodePhaseFailed, v1.NewTime(toStatus.OccurredAt), errMsg)
			nodeEvent.OccurredAt = utils.GetProtoTime(nodeStatus.GetStoppedAt())
			returnStatus = executors.NodeStatusFailed(toStatus.Err)
			c.metrics.FailureDuration.Observe(ctx, nodeStatus.GetStartedAt().Time, nodeStatus.GetStoppedAt().Time)
		} else {
			// retry
			// TODO add a nodeEvent of retryableFailure (it is not a terminal event).
			// For now, we don't send an event for node retryable failures.
			nodeEvent = nil
			nodeStatus.UpdatePhase(v1alpha1.NodePhaseRetryableFailure, v1.NewTime(toStatus.OccurredAt), errMsg)
			returnStatus = executors.NodeStatusRunning

			// Reset all executors' state to start a fresh attempt.
			nodeStatus.ClearTaskStatus()
			nodeStatus.ClearWorkflowStatus()
			nodeStatus.ClearDynamicNodeStatus()

			// Required for transition (backwards compatibility)
			if nodeStatus.GetLastUpdatedAt() != nil {
				c.metrics.FailureDuration.Observe(ctx, nodeStatus.GetStartedAt().Time, nodeStatus.GetLastUpdatedAt().Time)
			}
		}

	case handler.PhaseSkipped:
		nodeEvent.Phase = core.NodeExecution_SKIPPED
		nodeStatus.UpdatePhase(v1alpha1.NodePhaseSkipped, v1.NewTime(toStatus.OccurredAt), "")
		nodeEvent.OccurredAt = utils.GetProtoTime(nodeStatus.GetStoppedAt())
		returnStatus = executors.NodeStatusSuccess

	case handler.PhaseSucceeding:
		nodeStatus.UpdatePhase(v1alpha1.NodePhaseSucceeding, v1.NewTime(toStatus.OccurredAt), "")
		// Currently we do not record events for this
		return executors.NodeStatusRunning, nil

	case handler.PhaseSuccess:
		nodeEvent.Phase = core.NodeExecution_SUCCEEDED
		reason := ""
		if nodeStatus.IsCached() {
			reason = "Task Skipped due to Discovery Cache Hit."
		}
		nodeStatus.UpdatePhase(v1alpha1.NodePhaseSucceeded, v1.NewTime(toStatus.OccurredAt), reason)
		nodeEvent.OccurredAt = utils.GetProtoTime(nodeStatus.GetStoppedAt())
		if metadata, err := c.store.Head(ctx, v1alpha1.GetOutputsFile(nodeStatus.GetDataDir())); err == nil && metadata.Exists() {
			nodeEvent.OutputResult = &event.NodeExecutionEvent_OutputUri{
				OutputUri: v1alpha1.GetOutputsFile(nodeStatus.GetDataDir()).String(),
			}
		}

		returnStatus = executors.NodeStatusSuccess
		c.metrics.SuccessDuration.Observe(ctx, nodeStatus.GetStartedAt().Time, nodeStatus.GetStoppedAt().Time)

	case handler.PhaseFailing:
		nodeEvent.Phase = core.NodeExecution_FAILING
		nodeStatus.UpdatePhase(v1alpha1.NodePhaseFailing, v1.NewTime(toStatus.OccurredAt), "")
		nodeEvent.OccurredAt = utils.GetProtoTime(nil)
		returnStatus = executors.NodeStatusRunning

	case handler.PhaseFailed:
		nodeEvent.Phase = core.NodeExecution_FAILED
		nodeStatus.UpdatePhase(v1alpha1.NodePhaseFailed, v1.NewTime(toStatus.OccurredAt), errMsg)
		nodeEvent.OccurredAt = utils.GetProtoTime(nodeStatus.GetStoppedAt())
		nodeEvent.OutputResult = &event.NodeExecutionEvent_Error{
			Error: &core.ExecutionError{
				Code:     errCode,
				Message:  errMsg,
				ErrorUri: v1alpha1.GetOutputErrorFile(nodeStatus.GetDataDir()).String(),
			},
		}
		returnStatus = executors.NodeStatusFailed(toStatus.Err)
		c.metrics.FailureDuration.Observe(ctx, nodeStatus.GetStartedAt().Time, nodeStatus.GetStoppedAt().Time)

	case handler.PhaseUndefined:
		return executors.NodeStatusUndefined, errors.Errorf(errors.IllegalStateError, node.GetID(), "unexpected undefined state received, without an error")
	}

	// We observe that the phase has changed, and so we will record this event.
	if nodeEvent != nil && previousNodePhase != nodeStatus.GetPhase() {
		if nodeStatus.GetParentTaskID() != nil {
			nodeEvent.ParentTaskMetadata = &event.ParentTaskExecutionMetadata{
				Id: nodeStatus.GetParentTaskID(),
			}
		}

		logger.Debugf(ctx, "Recording NodeEvent for Phase transition [%s] -> [%s]", previousNodePhase.String(), nodeStatus.GetPhase().String())
		err := c.IdempotentRecordEvent(ctx, nodeEvent)

		if err != nil && eventsErr.IsEventAlreadyInTerminalStateError(err) {
			logger.Warningf(ctx, "Failed to record nodeEvent, error [%s]", err.Error())
			return executors.NodeStatusFailed(errors.Wrapf(errors.IllegalStateError, node.GetID(), err,
				"phase mismatch between propeller and control plane; Propeller State: %s", returnStatus.NodePhase)), nil
		} else if err != nil {
			logger.Warningf(ctx, "Failed to record nodeEvent, error [%s]", err.Error())
			return executors.NodeStatusUndefined, errors.Wrapf(errors.EventRecordingFailed, node.GetID(), err, "failed to record node event")
		}
	}
	return returnStatus, nil
}

func (c *nodeExecutor) executeNode(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode) (executors.NodeStatus, error) {
	handlerStatus, err := c.handleNode(ctx, w, node)
	if err != nil {
		logger.Warningf(ctx, "Node handling failed with an error [%v]", err.Error())
		return executors.NodeStatusUndefined, err
	}
	nodeStatus := w.GetNodeExecutionStatus(node.GetID())
	return c.TransitionToPhase(ctx, w.GetExecutionID().WorkflowExecutionIdentifier, node, nodeStatus, handlerStatus)
}

// The space search for the next node to execute is implemented like a DFS algorithm. handleDownstream visits all the nodes downstream from
// the currentNode. Visit a node is the RecursiveNodeHandler. A visit may be partial, complete or may result in a failure.
func (c *nodeExecutor) handleDownstream(ctx context.Context, w v1alpha1.ExecutableWorkflow, currentNode v1alpha1.ExecutableNode) (executors.NodeStatus, error) {
	logger.Debugf(ctx, "Handling downstream Nodes")
	// This node is success. Handle all downstream nodes
	downstreamNodes, err := w.FromNode(currentNode.GetID())
	if err != nil {
		logger.Debugf(ctx, "Error when retrieving downstream nodes. Error [%v]", err)
		return executors.NodeStatusFailed(err), nil
	}
	if len(downstreamNodes) == 0 {
		logger.Debugf(ctx, "No downstream nodes found. Complete.")
		return executors.NodeStatusComplete, nil
	}
	// If any downstream node is failed, fail, all
	// Else if all are success then success
	// Else if any one is running then Downstream is still running
	allCompleted := true
	partialNodeCompletion := false
	for _, downstreamNodeName := range downstreamNodes {
		downstreamNode, ok := w.GetNode(downstreamNodeName)
		if !ok {
			return executors.NodeStatusFailed(errors.Errorf(errors.BadSpecificationError, currentNode.GetID(), "Unable to find Downstream Node [%v]", downstreamNodeName)), nil
		}
		state, err := c.RecursiveNodeHandler(ctx, w, downstreamNode)
		if err != nil {
			return executors.NodeStatusUndefined, err
		}
		if state.HasFailed() {
			logger.Debugf(ctx, "Some downstream node has failed, %s", state.Err.Error())
			return state, nil
		}
		if !state.IsComplete() {
			allCompleted = false
		}

		if state.PartiallyComplete() {
			// This implies that one of the downstream nodes has completed and workflow is ready for propagation
			// We do not propagate in current cycle to make it possible to store the state between transitions
			partialNodeCompletion = true
		}
	}
	if allCompleted {
		logger.Debugf(ctx, "All downstream nodes completed")
		return executors.NodeStatusComplete, nil
	}
	if partialNodeCompletion {
		return executors.NodeStatusSuccess, nil
	}
	return executors.NodeStatusPending, nil
}

func (c *nodeExecutor) SetInputsForStartNode(ctx context.Context, w v1alpha1.BaseWorkflowWithStatus, inputs *handler.Data) (executors.NodeStatus, error) {
	startNode := w.StartNode()
	if startNode == nil {
		return executors.NodeStatusFailed(errors.Errorf(errors.BadSpecificationError, v1alpha1.StartNodeID, "Start node not found")), nil
	}
	ctx = contextutils.WithNodeID(ctx, startNode.GetID())
	if inputs == nil {
		logger.Infof(ctx, "No inputs for the workflow. Skipping storing inputs")
		return executors.NodeStatusComplete, nil
	}
	// StartNode is special. It does not have any processing step. It just takes the workflow (or subworkflow) inputs and converts to its own outputs
	nodeStatus := w.GetNodeExecutionStatus(startNode.GetID())
	if nodeStatus.GetDataDir() == "" {
		return executors.NodeStatusUndefined, errors.Errorf(errors.IllegalStateError, startNode.GetID(), "no data-dir set, cannot store inputs")
	}
	outputFile := v1alpha1.GetOutputsFile(nodeStatus.GetDataDir())
	so := storage.Options{}
	if err := c.store.WriteProtobuf(ctx, outputFile, so, inputs); err != nil {
		logger.Errorf(ctx, "Failed to write protobuf (metadata). Error [%v]", err)
		return executors.NodeStatusUndefined, errors.Wrapf(errors.CausedByError, startNode.GetID(), err, "Failed to store workflow inputs (as start node)")
	}
	return executors.NodeStatusComplete, nil
}

func (c *nodeExecutor) RecursiveNodeHandler(ctx context.Context, w v1alpha1.ExecutableWorkflow, currentNode v1alpha1.ExecutableNode) (executors.NodeStatus, error) {
	currentNodeCtx := contextutils.WithNodeID(ctx, currentNode.GetID())
	nodeStatus := w.GetNodeExecutionStatus(currentNode.GetID())
	switch nodeStatus.GetPhase() {
	case v1alpha1.NodePhaseNotYetStarted, v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseRunning, v1alpha1.NodePhaseFailing, v1alpha1.NodePhaseRetryableFailure, v1alpha1.NodePhaseSucceeding:
		logger.Debugf(currentNodeCtx, "Handling node Status [%v]", nodeStatus.GetPhase().String())
		t := c.metrics.NodeExecutionTime.Start(ctx)
		defer t.Stop()
		return c.executeNode(currentNodeCtx, w, currentNode)
		// TODO we can optimize skip state handling by iterating down the graph and marking all as skipped
		// Currently we treat either Skip or Success the same way. In this approach only one node will be skipped
		// at a time. As we iterate down, further nodes will be skipped
	case v1alpha1.NodePhaseSucceeded, v1alpha1.NodePhaseSkipped:
		return c.handleDownstream(ctx, w, currentNode)
	case v1alpha1.NodePhaseFailed:
		logger.Debugf(currentNodeCtx, "Node Failed")
		return executors.NodeStatusFailed(errors.Errorf(errors.RuntimeExecutionError, currentNode.GetID(), "Node Failed.")), nil
	}
	return executors.NodeStatusUndefined, errors.Errorf(errors.IllegalStateError, currentNode.GetID(), "Should never reach here")
}

func (c *nodeExecutor) AbortHandler(ctx context.Context, w v1alpha1.ExecutableWorkflow, currentNode v1alpha1.ExecutableNode) error {
	ctx = contextutils.WithNodeID(ctx, currentNode.GetID())
	nodeStatus := w.GetNodeExecutionStatus(currentNode.GetID())
	switch nodeStatus.GetPhase() {
	case v1alpha1.NodePhaseRunning:
		// Abort this node
		h, err := c.nodeHandlerFactory.GetHandler(currentNode.GetKind())
		if err != nil {
			return err
		}
		return h.AbortNode(ctx, w, currentNode)
	case v1alpha1.NodePhaseSucceeded, v1alpha1.NodePhaseSkipped:
		// Abort downstream nodes
		downstreamNodes, err := w.FromNode(currentNode.GetID())
		if err != nil {
			logger.Debugf(ctx, "Error when retrieving downstream nodes. Error [%v]", err)
			return nil
		}
		for _, d := range downstreamNodes {
			downstreamNode, ok := w.GetNode(d)
			if !ok {
				return errors.Errorf(errors.BadSpecificationError, currentNode.GetID(), "Unable to find Downstream Node [%v]", d)
			}
			if err := c.AbortHandler(ctx, w, downstreamNode); err != nil {
				return err
			}
		}
		return nil
	}
	return nil
}

func (c *nodeExecutor) Initialize(ctx context.Context) error {
	logger.Infof(ctx, "Initializing Core Node Executor")
	return nil
}

func NewExecutor(ctx context.Context, store *storage.DataStore, enQWorkflow v1alpha1.EnqueueWorkflow,
	revalPeriod time.Duration, eventSink events.EventSink, workflowLauncher launchplan.Executor,
	catalogClient catalog.Client, kubeClient executors.Client, scope promutils.Scope) (executors.Node, error) {

	nodeScope := scope.NewSubScope("node")
	exec := &nodeExecutor{
		store:           store,
		enqueueWorkflow: enQWorkflow,
		nodeRecorder:    events.NewNodeEventRecorder(eventSink, nodeScope),
		metrics: &nodeMetrics{
			FailureDuration:        labeled.NewStopWatch("failure_duration", "Indicates the total execution time of a failed workflow.", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
			SuccessDuration:        labeled.NewStopWatch("success_duration", "Indicates the total execution time of a successful workflow.", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
			InputsWriteFailure:     labeled.NewCounter("inputs_write_fail", "Indicates failure in writing node inputs to metastore", nodeScope),
			ResolutionFailure:      labeled.NewCounter("input_resolve_fail", "Indicates failure in resolving node inputs", nodeScope),
			TransitionLatency:      labeled.NewStopWatch("transition_latency", "Measures the latency between the last parent node stoppedAt time and current node's queued time.", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
			QueuingLatency:         labeled.NewStopWatch("queueing_latency", "Measures the latency between the time a node's been queued to the time the handler reported the executable moved to running state", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
			NodeExecutionTime:      labeled.NewStopWatch("node_exec_latency", "Measures the time taken to execute one node, a node can be complex so it may encompass sub-node latency.", time.Microsecond, nodeScope, labeled.EmitUnlabeledMetric),
			NodeInputGatherLatency: labeled.NewStopWatch("node_input_latency", "Measures the latency to aggregate inputs and check readiness of a node", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
		},
	}
	nodeHandlerFactory, err := NewHandlerFactory(
		ctx,
		exec,
		eventSink,
		workflowLauncher,
		enQWorkflow,
		revalPeriod,
		store,
		catalogClient,
		kubeClient,
		nodeScope,
	)
	exec.nodeHandlerFactory = nodeHandlerFactory
	return exec, err
}
