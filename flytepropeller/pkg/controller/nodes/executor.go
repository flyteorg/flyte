// Core Nodes Executor implementation
// This module implements the core Nodes executor.
// This executor is the starting point for executing any node in the workflow. Since Nodes in a workflow are composable,
// i.e., one node may contain other nodes, the Node Handler is recursive in nature.
// This executor handles the core logic for all nodes, but specific logic for handling different kinds of nodes is delegated
// to the respective node handlers
//
// Available node handlers are
// - Task: Arguably the most important handler as it handles all tasks. These include all plugins. The goal of the workflow is
//         is to run tasks, thus every workflow will contain atleast one TaskNode (except for the case, where the workflow
//          is purely a meta-workflow and can run other workflows
// - SubWorkflow: This is one of the most important handlers. It can executes Workflows that are nested inside a workflow
// - DynamicTask Handler: This is just a decorator on the Task Handler. It handles cases, in which the Task returns a futures
//                        file. Every Task is actually executed through the DynamicTaskHandler
// - Branch Handler: This handler is used to execute branches
// - Start & End Node handler: these are nominal handlers for the start and end node and do no really carry a lot of logic
package nodes

import (
	"context"
	"fmt"
	"time"

	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/recovery"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/common"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"
	errors2 "github.com/flyteorg/flytestdlib/errors"

	"github.com/flyteorg/flyteidl/clients/go/events"
	eventsErr "github.com/flyteorg/flyteidl/clients/go/events/errors"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	controllerEvents "github.com/flyteorg/flytepropeller/pkg/controller/events"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/golang/protobuf/ptypes"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/catalog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flytepropeller/pkg/controller/config"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/errors"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/subworkflow/launchplan"
)

type nodeMetrics struct {
	Scope                         promutils.Scope
	FailureDuration               labeled.StopWatch
	SuccessDuration               labeled.StopWatch
	RecoveryDuration              labeled.StopWatch
	UserErrorDuration             labeled.StopWatch
	SystemErrorDuration           labeled.StopWatch
	UnknownErrorDuration          labeled.StopWatch
	PermanentUserErrorDuration    labeled.StopWatch
	PermanentSystemErrorDuration  labeled.StopWatch
	PermanentUnknownErrorDuration labeled.StopWatch
	ResolutionFailure             labeled.Counter
	InputsWriteFailure            labeled.Counter
	TimedOutFailure               labeled.Counter

	InterruptedThresholdHit      labeled.Counter
	InterruptibleNodesRunning    labeled.Counter
	InterruptibleNodesTerminated labeled.Counter

	// Measures the latency between the last parent node stoppedAt time and current node's queued time.
	TransitionLatency labeled.StopWatch
	// Measures the latency between the time a node's been queued to the time the handler reported the executable moved
	// to running state
	QueuingLatency         labeled.StopWatch
	NodeExecutionTime      labeled.StopWatch
	NodeInputGatherLatency labeled.StopWatch
}

// Implements the executors.Node interface
type nodeExecutor struct {
	nodeHandlerFactory              HandlerFactory
	enqueueWorkflow                 v1alpha1.EnqueueWorkflow
	store                           *storage.DataStore
	nodeRecorder                    controllerEvents.NodeEventRecorder
	taskRecorder                    controllerEvents.TaskEventRecorder
	metrics                         *nodeMetrics
	maxDatasetSizeBytes             int64
	outputResolver                  OutputResolver
	defaultExecutionDeadline        time.Duration
	defaultActiveDeadline           time.Duration
	maxNodeRetriesForSystemFailures uint32
	interruptibleFailureThreshold   uint32
	defaultDataSandbox              storage.DataReference
	shardSelector                   ioutils.ShardSelector
	recoveryClient                  recovery.Client
	eventConfig                     *config.EventConfig
}

func (c *nodeExecutor) RecordTransitionLatency(ctx context.Context, dag executors.DAGStructure, nl executors.NodeLookup, node v1alpha1.ExecutableNode, nodeStatus v1alpha1.ExecutableNodeStatus) {
	if nodeStatus.GetPhase() == v1alpha1.NodePhaseNotYetStarted || nodeStatus.GetPhase() == v1alpha1.NodePhaseQueued {
		// Log transition latency (The most recently finished parent node endAt time to this node's queuedAt time -now-)
		t, err := GetParentNodeMaxEndTime(ctx, dag, nl, node)
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

func (c *nodeExecutor) IdempotentRecordEvent(ctx context.Context, nodeEvent *event.NodeExecutionEvent) error {
	if nodeEvent == nil {
		return fmt.Errorf("event recording attempt of Nil Node execution event")
	}

	if nodeEvent.Id == nil {
		return fmt.Errorf("event recording attempt of with nil node Event ID")
	}

	logger.Infof(ctx, "Recording event p[%+v]", nodeEvent)
	err := c.nodeRecorder.RecordNodeEvent(ctx, nodeEvent, c.eventConfig)
	if err != nil {
		if nodeEvent.GetId().NodeId == v1alpha1.EndNodeID {
			return nil
		}

		if eventsErr.IsAlreadyExists(err) {
			logger.Infof(ctx, "Node event phase: %s, nodeId %s already exist",
				nodeEvent.Phase.String(), nodeEvent.GetId().NodeId)
			return nil
		} else if eventsErr.IsEventAlreadyInTerminalStateError(err) {
			if IsTerminalNodePhase(nodeEvent.Phase) {
				// Event was trying to record a different terminal phase for an already terminal event. ignoring.
				logger.Infof(ctx, "Node event phase: %s, nodeId %s already in terminal phase. err: %s",
					nodeEvent.Phase.String(), nodeEvent.GetId().NodeId, err.Error())
				return nil
			}
			logger.Warningf(ctx, "Failed to record nodeEvent, error [%s]", err.Error())
			return errors.Wrapf(errors.IllegalStateError, nodeEvent.Id.NodeId, err, "phase mis-match mismatch between propeller and control plane; Trying to record Node p: %s", nodeEvent.Phase)
		}
	}
	return err
}

func (c *nodeExecutor) attemptRecovery(ctx context.Context, nCtx handler.NodeExecutionContext) (handler.PhaseInfo, error) {
	recovered, err := c.recoveryClient.RecoverNodeExecution(ctx, nCtx.ExecutionContext().GetExecutionConfig().RecoveryExecution.WorkflowExecutionIdentifier, nCtx.NodeExecutionMetadata().GetNodeExecutionID())
	if err != nil {
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.NotFound {
			logger.Warnf(ctx, "Failed to recover node [%+v] with err [%+v]", nCtx.NodeExecutionMetadata().GetNodeExecutionID(), err)
		}
		// The node is not recoverable when it's not found in the parent execution
		return handler.PhaseInfoUndefined, nil
	}
	if recovered == nil {
		logger.Warnf(ctx, "call to recover node [%+v] returned no error but also no node", nCtx.NodeExecutionMetadata().GetNodeExecutionID())
		return handler.PhaseInfoUndefined, nil
	}
	if recovered.Closure == nil {
		logger.Warnf(ctx, "Fetched node execution [%+v] data but was missing closure. Will not attempt to recover",
			nCtx.NodeExecutionMetadata().GetNodeExecutionID())
		return handler.PhaseInfoUndefined, nil
	}
	// A recoverable node execution should always be in a terminal phase
	switch recovered.Closure.Phase {
	case core.NodeExecution_SKIPPED:
		return handler.PhaseInfoSkip(nil, "node execution recovery indicated original node was skipped"), nil
	case core.NodeExecution_SUCCEEDED:
		logger.Debugf(ctx, "Node [%+v] can be recovered. Proceeding to copy inputs and outputs", nCtx.NodeExecutionMetadata().GetNodeExecutionID())
	default:
		logger.Debugf(ctx, "Node [%+v] phase [%v] is not recoverable", nCtx.NodeExecutionMetadata().GetNodeExecutionID(), recovered.Closure.Phase)
		return handler.PhaseInfoUndefined, nil
	}

	recoveredData, err := c.recoveryClient.RecoverNodeExecutionData(ctx, nCtx.ExecutionContext().GetExecutionConfig().RecoveryExecution.WorkflowExecutionIdentifier, nCtx.NodeExecutionMetadata().GetNodeExecutionID())
	if err != nil {
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.NotFound {
			logger.Warnf(ctx, "Failed to attemptRecovery node execution data for [%+v] although back-end indicated node was recoverable with err [%+v]",
				nCtx.NodeExecutionMetadata().GetNodeExecutionID(), err)
		}
		return handler.PhaseInfoUndefined, nil
	}
	if recoveredData == nil {
		logger.Warnf(ctx, "call to attemptRecovery node [%+v] data returned no error but also no data", nCtx.NodeExecutionMetadata().GetNodeExecutionID())
		return handler.PhaseInfoUndefined, nil
	}
	// Copy inputs to this node's expected location
	if recoveredData.FullInputs != nil {
		if err := c.store.WriteProtobuf(ctx, nCtx.InputReader().GetInputPath(), storage.Options{}, recoveredData.FullInputs); err != nil {
			c.metrics.InputsWriteFailure.Inc(ctx)
			logger.Errorf(ctx, "Failed to move recovered inputs for Node. Error [%v]. InputsFile [%s]", err, nCtx.InputReader().GetInputPath())
			return handler.PhaseInfoUndefined, errors.Wrapf(
				errors.StorageError, nCtx.NodeID(), err, "Failed to store inputs for Node. InputsFile [%s]", nCtx.InputReader().GetInputPath())
		}
	} else if len(recovered.InputUri) > 0 {
		// If the inputs are too large they won't be returned inline in the RecoverData call. We must fetch them before copying them.
		nodeInputs := &core.LiteralMap{}
		if recoveredData.FullInputs == nil {
			if err := c.store.ReadProtobuf(ctx, storage.DataReference(recovered.InputUri), nodeInputs); err != nil {
				return handler.PhaseInfoUndefined, errors.Wrapf(errors.InputsNotFoundError, nCtx.NodeID(), err, "failed to read data from dataDir [%v].", recovered.InputUri)
			}
		}

		if err := c.store.WriteProtobuf(ctx, nCtx.InputReader().GetInputPath(), storage.Options{}, nodeInputs); err != nil {
			c.metrics.InputsWriteFailure.Inc(ctx)
			logger.Errorf(ctx, "Failed to move recovered inputs for Node. Error [%v]. InputsFile [%s]", err, nCtx.InputReader().GetInputPath())
			return handler.PhaseInfoUndefined, errors.Wrapf(
				errors.StorageError, nCtx.NodeID(), err, "Failed to store inputs for Node. InputsFile [%s]", nCtx.InputReader().GetInputPath())
		}
	}
	// Similarly, copy outputs' reference
	so := storage.Options{}
	var outputs = &core.LiteralMap{}
	if recoveredData.FullOutputs != nil {
		outputs = recoveredData.FullOutputs
	} else if recovered.Closure.GetOutputData() != nil {
		outputs = recovered.Closure.GetOutputData()
	} else if len(recovered.Closure.GetOutputUri()) > 0 {
		if err := c.store.ReadProtobuf(ctx, storage.DataReference(recovered.Closure.GetOutputUri()), outputs); err != nil {
			return handler.PhaseInfoUndefined, errors.Wrapf(errors.InputsNotFoundError, nCtx.NodeID(), err, "failed to read output data [%v].", recovered.Closure.GetOutputUri())
		}
	} else {
		logger.Debugf(ctx, "No outputs found for recovered node [%+v]", nCtx.NodeExecutionMetadata().GetNodeExecutionID())
	}
	outputFile := v1alpha1.GetOutputsFile(nCtx.NodeStatus().GetOutputDir())
	if err := c.store.WriteProtobuf(ctx, outputFile, so, outputs); err != nil {
		logger.Errorf(ctx, "Failed to write protobuf (metadata). Error [%v]", err)
		return handler.PhaseInfoUndefined, errors.Wrapf(errors.CausedByError, nCtx.NodeID(), err, "Failed to store recovered node execution outputs")
	}

	info := &handler.ExecutionInfo{
		OutputInfo: &handler.OutputInfo{
			OutputURI: outputFile,
		},
	}
	if recovered.Closure.GetTaskNodeMetadata() != nil {
		taskNodeInfo := &handler.TaskNodeInfo{
			TaskNodeMetadata: &event.TaskNodeMetadata{
				CatalogKey:  recovered.Closure.GetTaskNodeMetadata().CatalogKey,
				CacheStatus: recovered.Closure.GetTaskNodeMetadata().CacheStatus,
			},
		}
		if recoveredData.DynamicWorkflow != nil {
			taskNodeInfo.TaskNodeMetadata.DynamicWorkflow = &event.DynamicWorkflowNodeMetadata{
				Id:               recoveredData.DynamicWorkflow.Id,
				CompiledWorkflow: recoveredData.DynamicWorkflow.CompiledWorkflow,
			}
		}
		info.TaskNodeInfo = taskNodeInfo
	} else if recovered.Closure.GetWorkflowNodeMetadata() != nil {
		logger.Warnf(ctx, "Attempted to recover node")
		info.WorkflowNodeInfo = &handler.WorkflowNodeInfo{
			LaunchedWorkflowID: recovered.Closure.GetWorkflowNodeMetadata().ExecutionId,
		}
	}
	return handler.PhaseInfoRecovered(info), nil
}

// In this method we check if the queue is ready to be processed and if so, we prime it in Admin as queued
// Before we start the node execution, we need to transition this Node status to Queued.
// This is because a node execution has to exist before task/wf executions can start.
func (c *nodeExecutor) preExecute(ctx context.Context, dag executors.DAGStructure, nCtx handler.NodeExecutionContext) (handler.PhaseInfo, error) {
	logger.Debugf(ctx, "Node not yet started")
	// Query the nodes information to figure out if it can be executed.
	predicatePhase, err := CanExecute(ctx, dag, nCtx.ContextualNodeLookup(), nCtx.Node())
	if err != nil {
		logger.Debugf(ctx, "Node failed in CanExecute. Error [%s]", err)
		return handler.PhaseInfoUndefined, err
	}

	if predicatePhase == PredicatePhaseReady {
		// TODO: Performance problem, we maybe in a retry loop and do not need to resolve the inputs again.
		// For now we will do this
		node := nCtx.Node()
		var nodeInputs *core.LiteralMap
		if !node.IsStartNode() {
			if nCtx.ExecutionContext().GetExecutionConfig().RecoveryExecution.WorkflowExecutionIdentifier != nil {
				phaseInfo, err := c.attemptRecovery(ctx, nCtx)
				if err != nil || phaseInfo.GetPhase() == handler.EPhaseRecovered {
					return phaseInfo, err
				}
			}
			nodeStatus := nCtx.NodeStatus()
			dataDir := nodeStatus.GetDataDir()
			t := c.metrics.NodeInputGatherLatency.Start(ctx)
			defer t.Stop()
			// Can execute
			var err error
			nodeInputs, err = Resolve(ctx, c.outputResolver, nCtx.ContextualNodeLookup(), node.GetID(), node.GetInputBindings())
			// TODO we need to handle retryable, network errors here!!
			if err != nil {
				c.metrics.ResolutionFailure.Inc(ctx)
				logger.Warningf(ctx, "Failed to resolve inputs for Node. Error [%v]", err)
				return handler.PhaseInfoFailure(core.ExecutionError_SYSTEM, "BindingResolutionFailure", err.Error(), nil), nil
			}

			if nodeInputs != nil {
				inputsFile := v1alpha1.GetInputsFile(dataDir)
				if err := c.store.WriteProtobuf(ctx, inputsFile, storage.Options{}, nodeInputs); err != nil {
					c.metrics.InputsWriteFailure.Inc(ctx)
					logger.Errorf(ctx, "Failed to store inputs for Node. Error [%v]. InputsFile [%s]", err, inputsFile)
					return handler.PhaseInfoUndefined, errors.Wrapf(
						errors.StorageError, node.GetID(), err, "Failed to store inputs for Node. InputsFile [%s]", inputsFile)
				}
			}

			logger.Debugf(ctx, "Node Data Directory [%s].", nodeStatus.GetDataDir())
		}

		return handler.PhaseInfoQueued("node queued"), nil
	}

	// Now that we have resolved the inputs, we can record as a transition latency. This is because we have completed
	// all the overhead that we have to compute. Any failures after this will incur this penalty, but it could be due
	// to various external reasons - like queuing, overuse of quota, plugin overhead etc.
	logger.Debugf(ctx, "preExecute completed in phase [%s]", predicatePhase.String())
	if predicatePhase == PredicatePhaseSkip {
		return handler.PhaseInfoSkip(nil, "Node Skipped as parent node was skipped"), nil
	}

	return handler.PhaseInfoNotReady("predecessor node not yet complete"), nil
}

func isTimeoutExpired(queuedAt *metav1.Time, timeout time.Duration) bool {
	if !queuedAt.IsZero() && timeout != 0 {
		deadline := queuedAt.Add(timeout)
		if deadline.Before(time.Now()) {
			return true
		}
	}
	return false
}

func (c *nodeExecutor) isEligibleForRetry(nCtx *nodeExecContext, nodeStatus v1alpha1.ExecutableNodeStatus, err *core.ExecutionError) (currentAttempt, maxAttempts uint32, isEligible bool) {
	if err.Kind == core.ExecutionError_SYSTEM {
		currentAttempt = nodeStatus.GetSystemFailures()
		maxAttempts = c.maxNodeRetriesForSystemFailures
		isEligible = currentAttempt < c.maxNodeRetriesForSystemFailures
		return
	}

	currentAttempt = (nodeStatus.GetAttempts() + 1) - nodeStatus.GetSystemFailures()
	if nCtx.Node().GetRetryStrategy() != nil && nCtx.Node().GetRetryStrategy().MinAttempts != nil {
		maxAttempts = uint32(*nCtx.Node().GetRetryStrategy().MinAttempts)
	}
	isEligible = currentAttempt < maxAttempts
	return
}

func (c *nodeExecutor) execute(ctx context.Context, h handler.Node, nCtx *nodeExecContext, nodeStatus v1alpha1.ExecutableNodeStatus) (handler.PhaseInfo, error) {
	logger.Debugf(ctx, "Executing node")
	defer logger.Debugf(ctx, "Node execution round complete")

	t, err := h.Handle(ctx, nCtx)
	if err != nil {
		return handler.PhaseInfoUndefined, err
	}

	phase := t.Info()
	// check for timeout for non-terminal phases
	if !phase.GetPhase().IsTerminal() {
		activeDeadline := c.defaultActiveDeadline
		if nCtx.Node().GetActiveDeadline() != nil && *nCtx.Node().GetActiveDeadline() > 0 {
			activeDeadline = *nCtx.Node().GetActiveDeadline()
		}
		if isTimeoutExpired(nodeStatus.GetQueuedAt(), activeDeadline) {
			logger.Errorf(ctx, "Node has timed out; timeout configured: %v", activeDeadline)
			return handler.PhaseInfoTimedOut(nil, fmt.Sprintf("task active timeout [%s] expired", activeDeadline.String())), nil
		}

		// Execution timeout is a retry-able error
		executionDeadline := c.defaultExecutionDeadline
		if nCtx.Node().GetExecutionDeadline() != nil && *nCtx.Node().GetExecutionDeadline() > 0 {
			executionDeadline = *nCtx.Node().GetExecutionDeadline()
		}
		if isTimeoutExpired(nodeStatus.GetLastAttemptStartedAt(), executionDeadline) {
			logger.Errorf(ctx, "Current execution for the node timed out; timeout configured: %v", executionDeadline)
			executionErr := &core.ExecutionError{Code: "TimeoutExpired", Message: fmt.Sprintf("task execution timeout [%s] expired", executionDeadline.String()), Kind: core.ExecutionError_USER}
			phase = handler.PhaseInfoRetryableFailureErr(executionErr, nil)
		}
	}

	if phase.GetPhase() == handler.EPhaseRetryableFailure {
		currentAttempt, maxAttempts, isEligible := c.isEligibleForRetry(nCtx, nodeStatus, phase.GetErr())
		if !isEligible {
			return handler.PhaseInfoFailure(
				core.ExecutionError_USER,
				fmt.Sprintf("RetriesExhausted|%s", phase.GetErr().Code),
				fmt.Sprintf("[%d/%d] currentAttempt done. Last Error: %s::%s", currentAttempt, maxAttempts, phase.GetErr().Kind.String(), phase.GetErr().Message),
				phase.GetInfo(),
			), nil
		}

		// Retrying to clearing all status
		nCtx.nsm.clearNodeStatus()
	}

	return phase, nil
}

func (c *nodeExecutor) abort(ctx context.Context, h handler.Node, nCtx handler.NodeExecutionContext, reason string) error {
	logger.Debugf(ctx, "Calling aborting & finalize")
	if err := h.Abort(ctx, nCtx, reason); err != nil {
		finalizeErr := h.Finalize(ctx, nCtx)
		if finalizeErr != nil {
			return errors.ErrorCollection{Errors: []error{err, finalizeErr}}
		}
		return err
	}

	return h.Finalize(ctx, nCtx)
}

func (c *nodeExecutor) finalize(ctx context.Context, h handler.Node, nCtx handler.NodeExecutionContext) error {
	return h.Finalize(ctx, nCtx)
}

func (c *nodeExecutor) handleNotYetStartedNode(ctx context.Context, dag executors.DAGStructure, nCtx *nodeExecContext, _ handler.Node) (executors.NodeStatus, error) {
	logger.Debugf(ctx, "Node not yet started, running pre-execute")
	defer logger.Debugf(ctx, "Node pre-execute completed")
	p, err := c.preExecute(ctx, dag, nCtx)
	if err != nil {
		logger.Errorf(ctx, "failed preExecute for node. Error: %s", err.Error())
		return executors.NodeStatusUndefined, err
	}

	if p.GetPhase() == handler.EPhaseUndefined {
		return executors.NodeStatusUndefined, errors.Errorf(errors.IllegalStateError, nCtx.NodeID(), "received undefined phase.")
	}

	if p.GetPhase() == handler.EPhaseNotReady {
		return executors.NodeStatusPending, nil
	}

	np, err := ToNodePhase(p.GetPhase())
	if err != nil {
		return executors.NodeStatusUndefined, errors.Wrapf(errors.IllegalStateError, nCtx.NodeID(), err, "failed to move from queued")
	}

	nodeStatus := nCtx.NodeStatus()
	if np != nodeStatus.GetPhase() {
		// assert np == Queued!
		logger.Infof(ctx, "Change in node state detected from [%s] -> [%s]", nodeStatus.GetPhase().String(), np.String())
		nev, err := ToNodeExecutionEvent(nCtx.NodeExecutionMetadata().GetNodeExecutionID(),
			p, nCtx.InputReader().GetInputPath().String(), nodeStatus, nCtx.ExecutionContext().GetEventVersion(),
			nCtx.ExecutionContext().GetParentInfo(), nCtx.node)
		if err != nil {
			return executors.NodeStatusUndefined, errors.Wrapf(errors.IllegalStateError, nCtx.NodeID(), err, "could not convert phase info to event")
		}
		err = c.IdempotentRecordEvent(ctx, nev)
		if err != nil {
			logger.Warningf(ctx, "Failed to record nodeEvent, error [%s]", err.Error())
			return executors.NodeStatusUndefined, errors.Wrapf(errors.EventRecordingFailed, nCtx.NodeID(), err, "failed to record node event")
		}
		UpdateNodeStatus(np, p, nCtx.nsm, nodeStatus)
		c.RecordTransitionLatency(ctx, dag, nCtx.ContextualNodeLookup(), nCtx.Node(), nodeStatus)
	}

	if np == v1alpha1.NodePhaseQueued {
		if nCtx.md.IsInterruptible() {
			c.metrics.InterruptibleNodesRunning.Inc(ctx)
		}
		return executors.NodeStatusQueued, nil
	} else if np == v1alpha1.NodePhaseSkipped {
		return executors.NodeStatusSuccess, nil
	}

	return executors.NodeStatusPending, nil
}

func (c *nodeExecutor) handleQueuedOrRunningNode(ctx context.Context, nCtx *nodeExecContext, h handler.Node) (executors.NodeStatus, error) {
	nodeStatus := nCtx.NodeStatus()
	currentPhase := nodeStatus.GetPhase()

	// case v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseRunning:
	logger.Debugf(ctx, "node executing, current phase [%s]", currentPhase)
	defer logger.Debugf(ctx, "node execution completed")

	// Since we reset node status inside execute for retryable failure, we use lastAttemptStartTime to carry that information
	// across execute which is used to emit metrics
	lastAttemptStartTime := nodeStatus.GetLastAttemptStartedAt()

	p, err := c.execute(ctx, h, nCtx, nodeStatus)
	if err != nil {
		logger.Errorf(ctx, "failed Execute for node. Error: %s", err.Error())
		return executors.NodeStatusUndefined, err
	}

	if p.GetPhase() == handler.EPhaseUndefined {
		return executors.NodeStatusUndefined, errors.Errorf(errors.IllegalStateError, nCtx.NodeID(), "received undefined phase.")
	}

	np, err := ToNodePhase(p.GetPhase())
	if err != nil {
		return executors.NodeStatusUndefined, errors.Wrapf(errors.IllegalStateError, nCtx.NodeID(), err, "failed to move from queued")
	}

	// execErr in phase-info 'p' is only available if node has failed to execute, and the current phase at that time
	// will be v1alpha1.NodePhaseRunning
	execErr := p.GetErr()
	if execErr != nil && (currentPhase == v1alpha1.NodePhaseRunning || currentPhase == v1alpha1.NodePhaseQueued ||
		currentPhase == v1alpha1.NodePhaseDynamicRunning) {
		endTime := time.Now()
		startTime := endTime
		if lastAttemptStartTime != nil {
			startTime = lastAttemptStartTime.Time
		}

		if execErr.GetKind() == core.ExecutionError_SYSTEM {
			nodeStatus.IncrementSystemFailures()
			c.metrics.SystemErrorDuration.Observe(ctx, startTime, endTime)
		} else if execErr.GetKind() == core.ExecutionError_USER {
			c.metrics.UserErrorDuration.Observe(ctx, startTime, endTime)
		} else {
			c.metrics.UnknownErrorDuration.Observe(ctx, startTime, endTime)
		}
		// When a node fails, we fail the workflow. Independent of number of nodes succeeding/failing, whenever a first node fails,
		// the entire workflow is failed.
		if np == v1alpha1.NodePhaseFailing {
			if execErr.GetKind() == core.ExecutionError_SYSTEM {
				nodeStatus.IncrementSystemFailures()
				c.metrics.PermanentSystemErrorDuration.Observe(ctx, startTime, endTime)
			} else if execErr.GetKind() == core.ExecutionError_USER {
				c.metrics.PermanentUserErrorDuration.Observe(ctx, startTime, endTime)
			} else {
				c.metrics.PermanentUnknownErrorDuration.Observe(ctx, startTime, endTime)
			}
		}
	}
	finalStatus := executors.NodeStatusRunning
	if np == v1alpha1.NodePhaseFailing && !h.FinalizeRequired() {
		logger.Infof(ctx, "Finalize not required, moving node to Failed")
		np = v1alpha1.NodePhaseFailed
		finalStatus = executors.NodeStatusFailed(p.GetErr())
	}

	if np == v1alpha1.NodePhaseTimingOut && !h.FinalizeRequired() {
		logger.Infof(ctx, "Finalize not required, moving node to TimedOut")
		np = v1alpha1.NodePhaseTimedOut
		finalStatus = executors.NodeStatusTimedOut
	}

	if np == v1alpha1.NodePhaseSucceeding && !h.FinalizeRequired() {
		logger.Infof(ctx, "Finalize not required, moving node to Succeeded")
		np = v1alpha1.NodePhaseSucceeded
		finalStatus = executors.NodeStatusSuccess
	}
	if np == v1alpha1.NodePhaseRecovered {
		logger.Infof(ctx, "Finalize not required, moving node to Recovered")
		finalStatus = executors.NodeStatusRecovered
	}

	// If it is retryable failure, we do no want to send any events, as the node is essentially still running
	// Similarly if the phase has not changed from the last time, events do not need to be sent
	if np != nodeStatus.GetPhase() && np != v1alpha1.NodePhaseRetryableFailure {
		// assert np == skipped, succeeding, failing or recovered
		logger.Infof(ctx, "Change in node state detected from [%s] -> [%s], (handler phase [%s])", nodeStatus.GetPhase().String(), np.String(), p.GetPhase().String())
		nev, err := ToNodeExecutionEvent(nCtx.NodeExecutionMetadata().GetNodeExecutionID(),
			p, nCtx.InputReader().GetInputPath().String(), nCtx.NodeStatus(), nCtx.ExecutionContext().GetEventVersion(),
			nCtx.ExecutionContext().GetParentInfo(), nCtx.node)
		if err != nil {
			return executors.NodeStatusUndefined, errors.Wrapf(errors.IllegalStateError, nCtx.NodeID(), err, "could not convert phase info to event")
		}

		err = c.IdempotentRecordEvent(ctx, nev)
		if err != nil {
			logger.Warningf(ctx, "Failed to record nodeEvent, error [%s]", err.Error())
			return executors.NodeStatusUndefined, errors.Wrapf(errors.EventRecordingFailed, nCtx.NodeID(), err, "failed to record node event")
		}

		// We reach here only when transitioning from Queued to Running. In this case, the startedAt is not set.
		if np == v1alpha1.NodePhaseRunning {
			if nodeStatus.GetQueuedAt() != nil {
				c.metrics.QueuingLatency.Observe(ctx, nodeStatus.GetQueuedAt().Time, time.Now())
			}
		}
	}

	UpdateNodeStatus(np, p, nCtx.nsm, nodeStatus)
	return finalStatus, nil
}

func (c *nodeExecutor) handleRetryableFailure(ctx context.Context, nCtx *nodeExecContext, h handler.Node) (executors.NodeStatus, error) {
	nodeStatus := nCtx.NodeStatus()
	logger.Debugf(ctx, "node failed with retryable failure, aborting and finalizing, message: %s", nodeStatus.GetMessage())
	if err := c.abort(ctx, h, nCtx, nodeStatus.GetMessage()); err != nil {
		return executors.NodeStatusUndefined, err
	}

	// NOTE: It is important to increment attempts only after abort has been called. Increment attempt mutates the state
	// Attempt is used throughout the system to determine the idempotent resource version.
	nodeStatus.IncrementAttempts()
	nodeStatus.UpdatePhase(v1alpha1.NodePhaseRunning, v1.Now(), "retrying", nil)
	// We are going to retry in the next round, so we should clear all current state
	nodeStatus.ClearSubNodeStatus()
	nodeStatus.ClearTaskStatus()
	nodeStatus.ClearWorkflowStatus()
	nodeStatus.ClearDynamicNodeStatus()
	return executors.NodeStatusPending, nil
}

func (c *nodeExecutor) handleNode(ctx context.Context, dag executors.DAGStructure, nCtx *nodeExecContext, h handler.Node) (executors.NodeStatus, error) {
	logger.Debugf(ctx, "Handling Node [%s]", nCtx.NodeID())
	defer logger.Debugf(ctx, "Completed node [%s]", nCtx.NodeID())

	nodeStatus := nCtx.NodeStatus()
	currentPhase := nodeStatus.GetPhase()

	// Optimization!
	// If it is start node we directly move it to Queued without needing to run preExecute
	if currentPhase == v1alpha1.NodePhaseNotYetStarted && !nCtx.Node().IsStartNode() {
		return c.handleNotYetStartedNode(ctx, dag, nCtx, h)
	}

	if currentPhase == v1alpha1.NodePhaseFailing {
		logger.Debugf(ctx, "node failing")
		if err := c.finalize(ctx, h, nCtx); err != nil {
			return executors.NodeStatusUndefined, err
		}
		nodeStatus.UpdatePhase(v1alpha1.NodePhaseFailed, v1.Now(), nodeStatus.GetMessage(), nodeStatus.GetExecutionError())
		c.metrics.FailureDuration.Observe(ctx, nodeStatus.GetStartedAt().Time, nodeStatus.GetStoppedAt().Time)
		if nCtx.md.IsInterruptible() {
			c.metrics.InterruptibleNodesTerminated.Inc(ctx)
		}
		return executors.NodeStatusFailed(nodeStatus.GetExecutionError()), nil
	}

	if currentPhase == v1alpha1.NodePhaseTimingOut {
		logger.Debugf(ctx, "node timing out")
		if err := c.abort(ctx, h, nCtx, "node timed out"); err != nil {
			return executors.NodeStatusUndefined, err
		}

		nodeStatus.ClearSubNodeStatus()
		nodeStatus.UpdatePhase(v1alpha1.NodePhaseTimedOut, v1.Now(), nodeStatus.GetMessage(), nodeStatus.GetExecutionError())
		c.metrics.TimedOutFailure.Inc(ctx)
		if nCtx.md.IsInterruptible() {
			c.metrics.InterruptibleNodesTerminated.Inc(ctx)
		}
		return executors.NodeStatusTimedOut, nil
	}

	if currentPhase == v1alpha1.NodePhaseSucceeding {
		logger.Debugf(ctx, "node succeeding")
		if err := c.finalize(ctx, h, nCtx); err != nil {
			return executors.NodeStatusUndefined, err
		}

		nodeStatus.ClearSubNodeStatus()
		nodeStatus.UpdatePhase(v1alpha1.NodePhaseSucceeded, v1.Now(), "completed successfully", nil)
		c.metrics.SuccessDuration.Observe(ctx, nodeStatus.GetStartedAt().Time, nodeStatus.GetStoppedAt().Time)
		if nCtx.md.IsInterruptible() {
			c.metrics.InterruptibleNodesTerminated.Inc(ctx)
		}
		return executors.NodeStatusSuccess, nil
	}

	if currentPhase == v1alpha1.NodePhaseRetryableFailure {
		return c.handleRetryableFailure(ctx, nCtx, h)
	}

	if currentPhase == v1alpha1.NodePhaseFailed {
		// This should never happen
		return executors.NodeStatusFailed(nodeStatus.GetExecutionError()), nil
	}

	return c.handleQueuedOrRunningNode(ctx, nCtx, h)
}

// The space search for the next node to execute is implemented like a DFS algorithm. handleDownstream visits all the nodes downstream from
// the currentNode. Visit a node is the RecursiveNodeHandler. A visit may be partial, complete or may result in a failure.
func (c *nodeExecutor) handleDownstream(ctx context.Context, execContext executors.ExecutionContext, dag executors.DAGStructure, nl executors.NodeLookup, currentNode v1alpha1.ExecutableNode) (executors.NodeStatus, error) {
	logger.Debugf(ctx, "Handling downstream Nodes")
	// This node is success. Handle all downstream nodes
	downstreamNodes, err := dag.FromNode(currentNode.GetID())
	if err != nil {
		logger.Debugf(ctx, "Error when retrieving downstream nodes, [%s]", err)
		return executors.NodeStatusFailed(&core.ExecutionError{
			Code:    errors.BadSpecificationError,
			Message: fmt.Sprintf("failed to retrieve downstream nodes for [%s]", currentNode.GetID()),
			Kind:    core.ExecutionError_SYSTEM,
		}), nil
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
	onFailurePolicy := execContext.GetOnFailurePolicy()
	stateOnComplete := executors.NodeStatusComplete
	for _, downstreamNodeName := range downstreamNodes {
		downstreamNode, ok := nl.GetNode(downstreamNodeName)
		if !ok {
			return executors.NodeStatusFailed(&core.ExecutionError{
				Code:    errors.BadSpecificationError,
				Message: fmt.Sprintf("failed to retrieve downstream node [%s] for [%s]", downstreamNodeName, currentNode.GetID()),
				Kind:    core.ExecutionError_SYSTEM,
			}), nil
		}

		state, err := c.RecursiveNodeHandler(ctx, execContext, dag, nl, downstreamNode)
		if err != nil {
			return executors.NodeStatusUndefined, err
		}

		if state.HasFailed() || state.HasTimedOut() {
			logger.Debugf(ctx, "Some downstream node has failed. Failed: [%v]. TimedOut: [%v]. Error: [%s]", state.HasFailed(), state.HasTimedOut(), state.Err)
			if onFailurePolicy == v1alpha1.WorkflowOnFailurePolicy(core.WorkflowMetadata_FAIL_AFTER_EXECUTABLE_NODES_COMPLETE) {
				// If the failure policy allows other nodes to continue running, do not exit the loop,
				// Keep track of the last failed state in the loop since it'll be the one to return.
				// TODO: If multiple nodes fail (which this mode allows), consolidate/summarize failure states in one.
				stateOnComplete = state
			} else {
				return state, nil
			}
		} else if !state.IsComplete() {
			// A Failed/Timedout node is implicitly considered "complete" this means none of the downstream nodes from
			// that node will ever be allowed to run.
			// This else block, therefore, deals with all other states. IsComplete will return true if and only if this
			// node as well as all of its downstream nodes have finished executing with success statuses. Otherwise we
			// mark this node's state as not completed to ensure we will visit it again later.
			allCompleted = false
		}

		if state.PartiallyComplete() {
			// This implies that one of the downstream nodes has just succeeded and workflow is ready for propagation
			// We do not propagate in current cycle to make it possible to store the state between transitions
			partialNodeCompletion = true
		}
	}

	if allCompleted {
		logger.Debugf(ctx, "All downstream nodes completed")
		return stateOnComplete, nil
	}

	if partialNodeCompletion {
		return executors.NodeStatusSuccess, nil
	}

	return executors.NodeStatusPending, nil
}

func (c *nodeExecutor) SetInputsForStartNode(ctx context.Context, execContext executors.ExecutionContext, dag executors.DAGStructureWithStartNode, nl executors.NodeLookup, inputs *core.LiteralMap) (executors.NodeStatus, error) {
	startNode := dag.StartNode()
	ctx = contextutils.WithNodeID(ctx, startNode.GetID())
	if inputs == nil {
		logger.Infof(ctx, "No inputs for the workflow. Skipping storing inputs")
		return executors.NodeStatusComplete, nil
	}

	// StartNode is special. It does not have any processing step. It just takes the workflow (or subworkflow) inputs and converts to its own outputs
	nodeStatus := nl.GetNodeExecutionStatus(ctx, startNode.GetID())

	if len(nodeStatus.GetDataDir()) == 0 {
		return executors.NodeStatusUndefined, errors.Errorf(errors.IllegalStateError, startNode.GetID(), "no data-dir set, cannot store inputs")
	}
	outputFile := v1alpha1.GetOutputsFile(nodeStatus.GetOutputDir())

	so := storage.Options{}
	if err := c.store.WriteProtobuf(ctx, outputFile, so, inputs); err != nil {
		logger.Errorf(ctx, "Failed to write protobuf (metadata). Error [%v]", err)
		return executors.NodeStatusUndefined, errors.Wrapf(errors.CausedByError, startNode.GetID(), err, "Failed to store workflow inputs (as start node)")
	}

	return executors.NodeStatusComplete, nil
}

func canHandleNode(phase v1alpha1.NodePhase) bool {
	return phase == v1alpha1.NodePhaseNotYetStarted ||
		phase == v1alpha1.NodePhaseQueued ||
		phase == v1alpha1.NodePhaseRunning ||
		phase == v1alpha1.NodePhaseFailing ||
		phase == v1alpha1.NodePhaseTimingOut ||
		phase == v1alpha1.NodePhaseRetryableFailure ||
		phase == v1alpha1.NodePhaseSucceeding ||
		phase == v1alpha1.NodePhaseDynamicRunning
}

func (c *nodeExecutor) RecursiveNodeHandler(ctx context.Context, execContext executors.ExecutionContext,
	dag executors.DAGStructure, nl executors.NodeLookup, currentNode v1alpha1.ExecutableNode) (
	executors.NodeStatus, error) {

	currentNodeCtx := contextutils.WithNodeID(ctx, currentNode.GetID())
	nodeStatus := nl.GetNodeExecutionStatus(ctx, currentNode.GetID())
	nodePhase := nodeStatus.GetPhase()

	if canHandleNode(nodePhase) {
		// TODO Follow up Pull Request,
		// 1. Rename this method to DAGTraversalHandleNode (accepts a DAGStructure along-with) the remaining arguments
		// 2. Create a new method called HandleNode (part of the interface) (remaining all args as the previous method, but no DAGStructure
		// 3. Additional both methods will receive inputs reader
		// 4. The Downstream nodes handler will Resolve the Inputs
		// 5. the method will delegate all other node handling to HandleNode.
		// 6. Thus we can get rid of SetInputs for StartNode as well
		logger.Debugf(currentNodeCtx, "Handling node Status [%v]", nodeStatus.GetPhase().String())

		t := c.metrics.NodeExecutionTime.Start(ctx)
		defer t.Stop()

		// This is an optimization to avoid creating the nodeContext object in case the node has already been looked at.
		// If the overhead was zero, we would just do the isDirtyCheck after the nodeContext is created
		if nodeStatus.IsDirty() {
			return executors.NodeStatusRunning, nil
		}

		// Now if the node is of type task, then let us check if we are within the parallelism limit, only if the node
		// has been queued already
		if currentNode.GetKind() == v1alpha1.NodeKindTask && nodeStatus.GetPhase() == v1alpha1.NodePhaseQueued {
			maxParallelism := execContext.GetExecutionConfig().MaxParallelism
			if maxParallelism > 0 {
				// If we are queued, let us see if we can proceed within the node parallelism bounds
				if execContext.CurrentParallelism() >= maxParallelism {
					logger.Infof(ctx, "Maximum Parallelism for task nodes achieved [%d] >= Max [%d], Round will be short-circuited.", execContext.CurrentParallelism(), maxParallelism)
					return executors.NodeStatusRunning, nil
				}
				// We know that Propeller goes through each workflow in a single thread, thus every node is really processed
				// sequentially. So, we can continue - now that we know we are under the parallelism limits and increment the
				// parallelism if the node, enters a running state
				logger.Debugf(ctx, "Parallelism criteria not met, Current [%d], Max [%d]", execContext.CurrentParallelism(), maxParallelism)
			} else {
				logger.Debugf(ctx, "Parallelism control disabled")
			}
		} else {
			logger.Debugf(ctx, "NodeKind: %s in status [%s]. Parallelism control is not applicable. Current Parallelism [%d]",
				currentNode.GetKind().String(), nodeStatus.GetPhase().String(), execContext.CurrentParallelism())
		}

		nCtx, err := c.newNodeExecContextDefault(ctx, currentNode.GetID(), execContext, nl)
		if err != nil {
			// NodeExecution creation failure is a permanent fail / system error.
			// Should a system failure always return an err?
			return executors.NodeStatusFailed(&core.ExecutionError{
				Code:    "InternalError",
				Message: err.Error(),
				Kind:    core.ExecutionError_SYSTEM,
			}), nil
		}

		// Now depending on the node type decide
		h, err := c.nodeHandlerFactory.GetHandler(nCtx.Node().GetKind())
		if err != nil {
			return executors.NodeStatusUndefined, err
		}

		return c.handleNode(currentNodeCtx, dag, nCtx, h)

		// TODO we can optimize skip state handling by iterating down the graph and marking all as skipped
		// Currently we treat either Skip or Success the same way. In this approach only one node will be skipped
		// at a time. As we iterate down, further nodes will be skipped
	} else if nodePhase == v1alpha1.NodePhaseSucceeded || nodePhase == v1alpha1.NodePhaseSkipped || nodePhase == v1alpha1.NodePhaseRecovered {
		logger.Debugf(currentNodeCtx, "Node has [%v], traversing downstream.", nodePhase)
		return c.handleDownstream(ctx, execContext, dag, nl, currentNode)
	} else if nodePhase == v1alpha1.NodePhaseFailed {
		logger.Debugf(currentNodeCtx, "Node has failed, traversing downstream.")
		_, err := c.handleDownstream(ctx, execContext, dag, nl, currentNode)
		if err != nil {
			return executors.NodeStatusUndefined, err
		}

		return executors.NodeStatusFailed(nodeStatus.GetExecutionError()), nil
	} else if nodePhase == v1alpha1.NodePhaseTimedOut {
		logger.Debugf(currentNodeCtx, "Node has timed out, traversing downstream.")
		_, err := c.handleDownstream(ctx, execContext, dag, nl, currentNode)
		if err != nil {
			return executors.NodeStatusUndefined, err
		}

		return executors.NodeStatusTimedOut, nil
	}

	return executors.NodeStatusUndefined, errors.Errorf(errors.IllegalStateError, currentNode.GetID(),
		"Should never reach here. Current Phase: %v", nodePhase)
}

func (c *nodeExecutor) FinalizeHandler(ctx context.Context, execContext executors.ExecutionContext, dag executors.DAGStructure, nl executors.NodeLookup, currentNode v1alpha1.ExecutableNode) error {
	nodeStatus := nl.GetNodeExecutionStatus(ctx, currentNode.GetID())
	nodePhase := nodeStatus.GetPhase()

	if nodePhase == v1alpha1.NodePhaseNotYetStarted {
		logger.Infof(ctx, "Node not yet started, will not finalize")
		// Nothing to be aborted
		return nil
	}

	if canHandleNode(nodePhase) {
		ctx = contextutils.WithNodeID(ctx, currentNode.GetID())

		// Now depending on the node type decide
		h, err := c.nodeHandlerFactory.GetHandler(currentNode.GetKind())
		if err != nil {
			return err
		}

		nCtx, err := c.newNodeExecContextDefault(ctx, currentNode.GetID(), execContext, nl)
		if err != nil {
			return err
		}
		// Abort this node
		err = c.finalize(ctx, h, nCtx)
		if err != nil {
			return err
		}
	} else {
		// Abort downstream nodes
		downstreamNodes, err := dag.FromNode(currentNode.GetID())
		if err != nil {
			logger.Debugf(ctx, "Error when retrieving downstream nodes. Error [%v]", err)
			return nil
		}

		errs := make([]error, 0, len(downstreamNodes))
		for _, d := range downstreamNodes {
			downstreamNode, ok := nl.GetNode(d)
			if !ok {
				return errors.Errorf(errors.BadSpecificationError, currentNode.GetID(), "Unable to find Downstream Node [%v]", d)
			}

			if err := c.FinalizeHandler(ctx, execContext, dag, nl, downstreamNode); err != nil {
				logger.Infof(ctx, "Failed to abort node [%v]. Error: %v", d, err)
				errs = append(errs, err)
			}
		}

		if len(errs) > 0 {
			return errors.ErrorCollection{Errors: errs}
		}

		return nil
	}

	return nil
}

func (c *nodeExecutor) AbortHandler(ctx context.Context, execContext executors.ExecutionContext, dag executors.DAGStructure, nl executors.NodeLookup, currentNode v1alpha1.ExecutableNode, reason string) error {
	nodeStatus := nl.GetNodeExecutionStatus(ctx, currentNode.GetID())
	nodePhase := nodeStatus.GetPhase()

	if nodePhase == v1alpha1.NodePhaseNotYetStarted {
		logger.Infof(ctx, "Node not yet started, will not finalize")
		// Nothing to be aborted
		return nil
	}

	if canHandleNode(nodePhase) {
		ctx = contextutils.WithNodeID(ctx, currentNode.GetID())

		// Now depending on the node type decide
		h, err := c.nodeHandlerFactory.GetHandler(currentNode.GetKind())
		if err != nil {
			return err
		}

		nCtx, err := c.newNodeExecContextDefault(ctx, currentNode.GetID(), execContext, nl)
		if err != nil {
			return err
		}
		// Abort this node
		err = c.abort(ctx, h, nCtx, reason)
		if err != nil {
			return err
		}
		nodeExecutionID := &core.NodeExecutionIdentifier{
			ExecutionId: nCtx.NodeExecutionMetadata().GetNodeExecutionID().ExecutionId,
			NodeId:      nCtx.NodeExecutionMetadata().GetNodeExecutionID().NodeId,
		}
		if nCtx.ExecutionContext().GetEventVersion() != v1alpha1.EventVersion0 {
			currentNodeUniqueID, err := common.GenerateUniqueID(nCtx.ExecutionContext().GetParentInfo(), nodeExecutionID.NodeId)
			if err != nil {
				return err
			}
			nodeExecutionID.NodeId = currentNodeUniqueID
		}

		err = c.IdempotentRecordEvent(ctx, &event.NodeExecutionEvent{
			Id:         nodeExecutionID,
			Phase:      core.NodeExecution_ABORTED,
			OccurredAt: ptypes.TimestampNow(),
			OutputResult: &event.NodeExecutionEvent_Error{
				Error: &core.ExecutionError{
					Code:    "NodeAborted",
					Message: reason,
				},
			},
		})
		if err != nil {
			if errors2.IsCausedBy(err, errors.IllegalStateError) {
				logger.Debugf(ctx, "Failed to record abort event due to illegal state transition. Ignoring the error. Error: %v", err)
			} else {
				logger.Warningf(ctx, "Failed to record nodeEvent, error [%s]", err.Error())
				return errors.Wrapf(errors.EventRecordingFailed, nCtx.NodeID(), err, "failed to record node event")
			}
		}
	} else if nodePhase == v1alpha1.NodePhaseSucceeded || nodePhase == v1alpha1.NodePhaseSkipped || nodePhase == v1alpha1.NodePhaseRecovered {
		// Abort downstream nodes
		downstreamNodes, err := dag.FromNode(currentNode.GetID())
		if err != nil {
			logger.Debugf(ctx, "Error when retrieving downstream nodes. Error [%v]", err)
			return nil
		}

		errs := make([]error, 0, len(downstreamNodes))
		for _, d := range downstreamNodes {
			downstreamNode, ok := nl.GetNode(d)
			if !ok {
				return errors.Errorf(errors.BadSpecificationError, currentNode.GetID(), "Unable to find Downstream Node [%v]", d)
			}

			if err := c.AbortHandler(ctx, execContext, dag, nl, downstreamNode, reason); err != nil {
				logger.Infof(ctx, "Failed to abort node [%v]. Error: %v", d, err)
				errs = append(errs, err)
			}
		}

		if len(errs) > 0 {
			return errors.ErrorCollection{Errors: errs}
		}

		return nil
	} else {
		ctx = contextutils.WithNodeID(ctx, currentNode.GetID())
		logger.Warnf(ctx, "Trying to abort a node in state [%s]", nodeStatus.GetPhase().String())
	}

	return nil
}

func (c *nodeExecutor) Initialize(ctx context.Context) error {
	logger.Infof(ctx, "Initializing Core Node Executor")
	s := c.newSetupContext(ctx)
	return c.nodeHandlerFactory.Setup(ctx, s)
}

func NewExecutor(ctx context.Context, nodeConfig config.NodeConfig, store *storage.DataStore, enQWorkflow v1alpha1.EnqueueWorkflow, eventSink events.EventSink,
	workflowLauncher launchplan.Executor, launchPlanReader launchplan.Reader, maxDatasetSize int64,
	defaultRawOutputPrefix storage.DataReference, kubeClient executors.Client,
	catalogClient catalog.Client, recoveryClient recovery.Client, eventConfig *config.EventConfig, scope promutils.Scope) (executors.Node, error) {

	// TODO we may want to make this configurable.
	shardSelector, err := ioutils.NewBase36PrefixShardSelector(ctx)
	if err != nil {
		return nil, err
	}

	nodeScope := scope.NewSubScope("node")
	exec := &nodeExecutor{
		store:               store,
		enqueueWorkflow:     enQWorkflow,
		nodeRecorder:        controllerEvents.NewNodeEventRecorder(eventSink, nodeScope, store),
		taskRecorder:        controllerEvents.NewTaskEventRecorder(eventSink, scope.NewSubScope("task"), store),
		maxDatasetSizeBytes: maxDatasetSize,
		metrics: &nodeMetrics{
			Scope:                         nodeScope,
			FailureDuration:               labeled.NewStopWatch("failure_duration", "Indicates the total execution time of a failed workflow.", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
			SuccessDuration:               labeled.NewStopWatch("success_duration", "Indicates the total execution time of a successful workflow.", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
			RecoveryDuration:              labeled.NewStopWatch("recovery_duration", "Indicates the total execution time of a recovered workflow.", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
			UserErrorDuration:             labeled.NewStopWatch("user_error_duration", "Indicates the total execution time before user error", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
			SystemErrorDuration:           labeled.NewStopWatch("system_error_duration", "Indicates the total execution time before system error", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
			UnknownErrorDuration:          labeled.NewStopWatch("unknown_error_duration", "Indicates the total execution time before unknown error", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
			PermanentUserErrorDuration:    labeled.NewStopWatch("perma_user_error_duration", "Indicates the total execution time before non recoverable user error", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
			PermanentSystemErrorDuration:  labeled.NewStopWatch("perma_system_error_duration", "Indicates the total execution time before non recoverable system error", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
			PermanentUnknownErrorDuration: labeled.NewStopWatch("perma_unknown_error_duration", "Indicates the total execution time before non recoverable unknown error", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
			InputsWriteFailure:            labeled.NewCounter("inputs_write_fail", "Indicates failure in writing node inputs to metastore", nodeScope),
			TimedOutFailure:               labeled.NewCounter("timeout_fail", "Indicates failure due to timeout", nodeScope),
			InterruptedThresholdHit:       labeled.NewCounter("interrupted_threshold", "Indicates the node interruptible disabled because it hit max failure count", nodeScope),
			InterruptibleNodesRunning:     labeled.NewCounter("interruptible_nodes_running", "number of interruptible nodes running", nodeScope),
			InterruptibleNodesTerminated:  labeled.NewCounter("interruptible_nodes_terminated", "number of interruptible nodes finished running", nodeScope),
			ResolutionFailure:             labeled.NewCounter("input_resolve_fail", "Indicates failure in resolving node inputs", nodeScope),
			TransitionLatency:             labeled.NewStopWatch("transition_latency", "Measures the latency between the last parent node stoppedAt time and current node's queued time.", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
			QueuingLatency:                labeled.NewStopWatch("queueing_latency", "Measures the latency between the time a node's been queued to the time the handler reported the executable moved to running state", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
			NodeExecutionTime:             labeled.NewStopWatch("node_exec_latency", "Measures the time taken to execute one node, a node can be complex so it may encompass sub-node latency.", time.Microsecond, nodeScope, labeled.EmitUnlabeledMetric),
			NodeInputGatherLatency:        labeled.NewStopWatch("node_input_latency", "Measures the latency to aggregate inputs and check readiness of a node", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
		},
		outputResolver:                  NewRemoteFileOutputResolver(store),
		defaultExecutionDeadline:        nodeConfig.DefaultDeadlines.DefaultNodeExecutionDeadline.Duration,
		defaultActiveDeadline:           nodeConfig.DefaultDeadlines.DefaultNodeActiveDeadline.Duration,
		maxNodeRetriesForSystemFailures: uint32(nodeConfig.MaxNodeRetriesOnSystemFailures),
		interruptibleFailureThreshold:   uint32(nodeConfig.InterruptibleFailureThreshold),
		defaultDataSandbox:              defaultRawOutputPrefix,
		shardSelector:                   shardSelector,
		recoveryClient:                  recoveryClient,
		eventConfig:                     eventConfig,
	}
	nodeHandlerFactory, err := NewHandlerFactory(ctx, exec, workflowLauncher, launchPlanReader, kubeClient, catalogClient, recoveryClient, eventConfig, nodeScope)
	exec.nodeHandlerFactory = nodeHandlerFactory
	return exec, err
}

func init() {
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey,
		contextutils.TaskIDKey)
}
