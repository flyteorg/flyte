// Package nodes contains the Core Nodes Executor implementation and a subpackage for every node kind
// This module implements the core Nodes executor.
// This executor is the starting point for executing any node in the workflow. Since Nodes in a workflow are composable,
// i.e., one node may contain other nodes, the Node Handler is recursive in nature.
// This executor handles the core logic for all nodes, but specific logic for handling different kinds of nodes is delegated
// to the respective node handlers
//
// Available node handlers are
//   - Task: Arguably the most important handler as it handles all tasks. These include all plugins. The goal of the workflow is
//     is to run tasks, thus every workflow will contain at least one TaskNode (except for the case, where the workflow
//     is purely a meta-workflow and can run other workflows
//   - SubWorkflow: This is one of the most important handlers. It can execute Workflows that are nested inside a workflow
//   - DynamicTask Handler: This is just a decorator on the Task Handler. It handles cases, in which the Task returns a futures
//     file. Every Task is actually executed through the DynamicTaskHandler
//   - Branch Handler: This handler is used to execute branches
//   - Start & End Node handler: these are nominal handlers for the start and end node and do no really carry a lot of logic
package nodes

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/flytepropeller/events"
	eventsErr "github.com/flyteorg/flyte/flytepropeller/events/errors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/common"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/errors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/recovery"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/subworkflow/launchplan"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	errors2 "github.com/flyteorg/flyte/flytestdlib/errors"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

const cacheSerializedReason = "waiting on serialized cache"

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

	catalogPutFailureCount         labeled.Counter
	catalogGetFailureCount         labeled.Counter
	catalogPutSuccessCount         labeled.Counter
	catalogMissCount               labeled.Counter
	catalogHitCount                labeled.Counter
	catalogSkipCount               labeled.Counter
	reservationGetSuccessCount     labeled.Counter
	reservationGetFailureCount     labeled.Counter
	reservationReleaseSuccessCount labeled.Counter
	reservationReleaseFailureCount labeled.Counter
}

// recursiveNodeExector implements the executors.Node interfaces and is the starting point for
// executing any node in the workflow.
type recursiveNodeExecutor struct {
	nodeExecutor       interfaces.NodeExecutor
	nCtxBuilder        interfaces.NodeExecutionContextBuilder
	enqueueWorkflow    v1alpha1.EnqueueWorkflow
	nodeHandlerFactory interfaces.HandlerFactory
	store              *storage.DataStore
	metrics            *nodeMetrics
}

func (c *recursiveNodeExecutor) SetInputsForStartNode(ctx context.Context, execContext executors.ExecutionContext, dag executors.DAGStructureWithStartNode, nl executors.NodeLookup, inputs *core.LiteralMap) (interfaces.NodeStatus, error) {
	startNode := dag.StartNode()
	ctx = contextutils.WithNodeID(ctx, startNode.GetID())
	if inputs == nil {
		logger.Infof(ctx, "No inputs for the workflow. Skipping storing inputs")
		return interfaces.NodeStatusComplete, nil
	}

	// StartNode is special. It does not have any processing step. It just takes the workflow (or subworkflow) inputs and converts to its own outputs
	nodeStatus := nl.GetNodeExecutionStatus(ctx, startNode.GetID())

	if len(nodeStatus.GetDataDir()) == 0 {
		return interfaces.NodeStatusUndefined, errors.Errorf(errors.IllegalStateError, startNode.GetID(), "no data-dir set, cannot store inputs")
	}
	outputFile := v1alpha1.GetOutputsFile(nodeStatus.GetOutputDir())

	so := storage.Options{}
	if err := c.store.WriteProtobuf(ctx, outputFile, so, inputs); err != nil {
		logger.Errorf(ctx, "Failed to write protobuf (metadata). Error [%v]", err)
		return interfaces.NodeStatusUndefined, errors.Wrapf(errors.CausedByError, startNode.GetID(), err, "Failed to store workflow inputs (as start node)")
	}

	return interfaces.NodeStatusComplete, nil
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

// IsMaxParallelismAchieved checks if we have already achieved max parallelism. It returns true, if the desired max parallelism
// value is achieved, false otherwise
// MaxParallelism is defined as the maximum number of TaskNodes and LaunchPlans (together) that can be executed concurrently
// by one workflow execution. A setting of `0` indicates that it is disabled.
func IsMaxParallelismAchieved(ctx context.Context, currentNode v1alpha1.ExecutableNode, currentPhase v1alpha1.NodePhase,
	execContext executors.ExecutionContext) bool {
	maxParallelism := execContext.GetExecutionConfig().MaxParallelism
	if maxParallelism == 0 {
		logger.Debugf(ctx, "Parallelism control disabled")
		return false
	}

	if currentNode.GetKind() == v1alpha1.NodeKindTask || currentNode.GetKind() == v1alpha1.NodeKindArray ||
		(currentNode.GetKind() == v1alpha1.NodeKindWorkflow && currentNode.GetWorkflowNode() != nil && currentNode.GetWorkflowNode().GetLaunchPlanRefID() != nil) {
		// If we are queued, let us see if we can proceed within the node parallelism bounds
		if execContext.CurrentParallelism() >= maxParallelism {
			logger.Infof(ctx, "Maximum Parallelism for task/launch-plan nodes achieved [%d] >= Max [%d], Round will be short-circuited.", execContext.CurrentParallelism(), maxParallelism)
			return true
		}
		// We know that Propeller goes through each workflow in a single thread, thus every node is really processed
		// sequentially. So, we can continue - now that we know we are under the parallelism limits and increment the
		// parallelism if the node, enters a running state
		logger.Debugf(ctx, "Parallelism criteria not met, Current [%d], Max [%d]", execContext.CurrentParallelism(), maxParallelism)
	} else {
		logger.Debugf(ctx, "NodeKind: %s in status [%s]. Parallelism control is not applicable. Current Parallelism [%d]",
			currentNode.GetKind().String(), currentPhase.String(), execContext.CurrentParallelism())
	}
	return false
}

// RecursiveNodeHandler This is the entrypoint of executing a node in a workflow. A workflow consists of nodes, that are
// nested within other nodes. The system follows an actor model, where the parent nodes control the execution of nested nodes
// The recursive node-handler uses a modified depth-first type of algorithm to execute non-blocked nodes.
func (c *recursiveNodeExecutor) RecursiveNodeHandler(ctx context.Context, execContext executors.ExecutionContext,
	dag executors.DAGStructure, nl executors.NodeLookup, currentNode v1alpha1.ExecutableNode) (
	interfaces.NodeStatus, error) {

	currentNodeCtx := contextutils.WithNodeID(ctx, currentNode.GetID())
	nodeStatus := nl.GetNodeExecutionStatus(ctx, currentNode.GetID())
	nodePhase := nodeStatus.GetPhase()

	if nodePhase == v1alpha1.NodePhaseRunning && execContext != nil {
		execContext.IncrementNodeExecutionCount()
		if currentNode.GetKind() == v1alpha1.NodeKindTask {
			execContext.IncrementTaskExecutionCount()
		}
		logger.Debugf(currentNodeCtx, "recursive handler - node execution count [%v], task execution count [%v], phase [%v], ",
			execContext.CurrentNodeExecutionCount(), execContext.CurrentTaskExecutionCount(), nodePhase.String())
	}

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
			return interfaces.NodeStatusRunning, nil
		}

		if IsMaxParallelismAchieved(ctx, currentNode, nodePhase, execContext) {
			return interfaces.NodeStatusRunning, nil
		}

		nCtx, err := c.nCtxBuilder.BuildNodeExecutionContext(ctx, execContext, nl, currentNode.GetID())
		if err != nil {
			// NodeExecution creation failure is a permanent fail / system error.
			// Should a system failure always return an err?
			return interfaces.NodeStatusFailed(&core.ExecutionError{
				Code:    "InternalError",
				Message: err.Error(),
				Kind:    core.ExecutionError_SYSTEM,
			}), nil
		}

		// Now depending on the node type decide
		h, err := c.nodeHandlerFactory.GetHandler(nCtx.Node().GetKind())
		if err != nil {
			return interfaces.NodeStatusUndefined, err
		}

		return c.nodeExecutor.HandleNode(currentNodeCtx, dag, nCtx, h)

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
			return interfaces.NodeStatusUndefined, err
		}

		return interfaces.NodeStatusFailed(nodeStatus.GetExecutionError()), nil
	} else if nodePhase == v1alpha1.NodePhaseTimedOut {
		logger.Debugf(currentNodeCtx, "Node has timed out, traversing downstream.")
		_, err := c.handleDownstream(ctx, execContext, dag, nl, currentNode)
		if err != nil {
			return interfaces.NodeStatusUndefined, err
		}

		return interfaces.NodeStatusTimedOut, nil
	}

	return interfaces.NodeStatusUndefined, errors.Errorf(errors.IllegalStateError, currentNode.GetID(),
		"Should never reach here. Current Phase: %v", nodePhase)
}

// The space search for the next node to execute is implemented like a DFS algorithm. handleDownstream visits all the nodes downstream from
// the currentNode. Visit a node is the RecursiveNodeHandler. A visit may be partial, complete or may result in a failure.
func (c *recursiveNodeExecutor) handleDownstream(ctx context.Context, execContext executors.ExecutionContext, dag executors.DAGStructure, nl executors.NodeLookup, currentNode v1alpha1.ExecutableNode) (interfaces.NodeStatus, error) {
	logger.Debugf(ctx, "Handling downstream Nodes")
	// This node is success. Handle all downstream nodes
	downstreamNodes, err := dag.FromNode(currentNode.GetID())
	if err != nil {
		logger.Debugf(ctx, "Error when retrieving downstream nodes, [%s]", err)
		return interfaces.NodeStatusFailed(&core.ExecutionError{
			Code:    errors.BadSpecificationError,
			Message: fmt.Sprintf("failed to retrieve downstream nodes for [%s]", currentNode.GetID()),
			Kind:    core.ExecutionError_SYSTEM,
		}), nil
	}
	if len(downstreamNodes) == 0 {
		logger.Debugf(ctx, "No downstream nodes found. Complete.")
		return interfaces.NodeStatusComplete, nil
	}
	// If any downstream node is failed, fail, all
	// Else if all are success then success
	// Else if any one is running then Downstream is still running
	allCompleted := true
	partialNodeCompletion := false
	onFailurePolicy := execContext.GetOnFailurePolicy()
	stateOnComplete := interfaces.NodeStatusComplete
	for _, downstreamNodeName := range downstreamNodes {
		downstreamNode, ok := nl.GetNode(downstreamNodeName)
		if !ok {
			return interfaces.NodeStatusFailed(&core.ExecutionError{
				Code:    errors.BadSpecificationError,
				Message: fmt.Sprintf("failed to retrieve downstream node [%s] for [%s]", downstreamNodeName, currentNode.GetID()),
				Kind:    core.ExecutionError_SYSTEM,
			}), nil
		}

		logger.Debugf(ctx, "downstream handler starting node id %v, ", downstreamNode.GetID())
		state, err := c.RecursiveNodeHandler(ctx, execContext, dag, nl, downstreamNode)
		if err != nil {
			return interfaces.NodeStatusUndefined, err
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
		return interfaces.NodeStatusSuccess, nil
	}

	return interfaces.NodeStatusPending, nil
}

func (c *recursiveNodeExecutor) FinalizeHandler(ctx context.Context, execContext executors.ExecutionContext, dag executors.DAGStructure, nl executors.NodeLookup, currentNode v1alpha1.ExecutableNode) error {
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

		nCtx, err := c.nCtxBuilder.BuildNodeExecutionContext(ctx, execContext, nl, currentNode.GetID())
		if err != nil {
			return err
		}
		// Abort this node
		err = c.nodeExecutor.Finalize(ctx, h, nCtx)
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

func (c *recursiveNodeExecutor) AbortHandler(ctx context.Context, execContext executors.ExecutionContext, dag executors.DAGStructure, nl executors.NodeLookup, currentNode v1alpha1.ExecutableNode, reason string) error {
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

		nCtx, err := c.nCtxBuilder.BuildNodeExecutionContext(ctx, execContext, nl, currentNode.GetID())
		if err != nil {
			return err
		}
		// Abort this node
		return c.nodeExecutor.Abort(ctx, h, nCtx, reason, true)
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

func (c *recursiveNodeExecutor) Initialize(ctx context.Context) error {
	logger.Infof(ctx, "Initializing Core Node Executor")
	s := c.newSetupContext(ctx)
	return c.nodeHandlerFactory.Setup(ctx, c, s)
}

// GetNodeExecutionContextBuilder returns the current NodeExecutionContextBuilder
func (c *recursiveNodeExecutor) GetNodeExecutionContextBuilder() interfaces.NodeExecutionContextBuilder {
	return c.nCtxBuilder
}

// WithNodeExecutionContextBuilder returns a new Node with the given NodeExecutionContextBuilder
func (c *recursiveNodeExecutor) WithNodeExecutionContextBuilder(nCtxBuilder interfaces.NodeExecutionContextBuilder) interfaces.Node {
	return &recursiveNodeExecutor{
		nodeExecutor:       c.nodeExecutor,
		nCtxBuilder:        nCtxBuilder,
		enqueueWorkflow:    c.enqueueWorkflow,
		nodeHandlerFactory: c.nodeHandlerFactory,
		store:              c.store,
		metrics:            c.metrics,
	}
}

// nodeExecutor implements the NodeExecutor interface and is responsible for executing a single node.
type nodeExecutor struct {
	catalog                         catalog.Client
	clusterID                       string
	enableCRDebugMetadata           bool
	defaultActiveDeadline           time.Duration
	defaultDataSandbox              storage.DataReference
	defaultExecutionDeadline        time.Duration
	enqueueWorkflow                 v1alpha1.EnqueueWorkflow
	eventConfig                     *config.EventConfig
	interruptibleFailureThreshold   int32
	maxNodeRetriesForSystemFailures uint32
	metrics                         *nodeMetrics
	nodeRecorder                    events.NodeEventRecorder
	outputResolver                  OutputResolver
	recoveryClient                  recovery.Client
	shardSelector                   ioutils.ShardSelector
	store                           *storage.DataStore
	taskRecorder                    events.TaskEventRecorder
	consoleURL                      string
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

func (c *nodeExecutor) recoverInputs(ctx context.Context, nCtx interfaces.NodeExecutionContext,
	recovered *admin.NodeExecution, recoveredData *admin.NodeExecutionGetDataResponse) (*core.LiteralMap, error) {

	nodeInputs := recoveredData.FullInputs
	if nodeInputs != nil {
		if err := c.store.WriteProtobuf(ctx, nCtx.InputReader().GetInputPath(), storage.Options{}, nodeInputs); err != nil {
			c.metrics.InputsWriteFailure.Inc(ctx)
			logger.Errorf(ctx, "Failed to move recovered inputs for Node. Error [%v]. InputsFile [%s]", err, nCtx.InputReader().GetInputPath())
			return nil, errors.Wrapf(errors.StorageError, nCtx.NodeID(), err, "Failed to store inputs for Node. InputsFile [%s]", nCtx.InputReader().GetInputPath())
		}
	} else if len(recovered.InputUri) > 0 {
		// If the inputs are too large they won't be returned inline in the RecoverData call. We must fetch them before copying them.
		nodeInputs = &core.LiteralMap{}
		if recoveredData.FullInputs == nil {
			if err := c.store.ReadProtobuf(ctx, storage.DataReference(recovered.InputUri), nodeInputs); err != nil {
				return nil, errors.Wrapf(errors.InputsNotFoundError, nCtx.NodeID(), err, "failed to read data from dataDir [%v].", recovered.InputUri)
			}
		}

		if err := c.store.WriteProtobuf(ctx, nCtx.InputReader().GetInputPath(), storage.Options{}, nodeInputs); err != nil {
			c.metrics.InputsWriteFailure.Inc(ctx)
			logger.Errorf(ctx, "Failed to move recovered inputs for Node. Error [%v]. InputsFile [%s]", err, nCtx.InputReader().GetInputPath())
			return nil, errors.Wrapf(errors.StorageError, nCtx.NodeID(), err, "Failed to store inputs for Node. InputsFile [%s]", nCtx.InputReader().GetInputPath())
		}
	}

	return nodeInputs, nil
}

func (c *nodeExecutor) attemptRecovery(ctx context.Context, nCtx interfaces.NodeExecutionContext) (handler.PhaseInfo, error) {
	fullyQualifiedNodeID := nCtx.NodeExecutionMetadata().GetNodeExecutionID().NodeId
	if nCtx.ExecutionContext().GetEventVersion() != v1alpha1.EventVersion0 {
		// compute fully qualified node id (prefixed with parent id and retry attempt) to ensure uniqueness
		var err error
		fullyQualifiedNodeID, err = common.GenerateUniqueID(nCtx.ExecutionContext().GetParentInfo(), nCtx.NodeExecutionMetadata().GetNodeExecutionID().NodeId)
		if err != nil {
			return handler.PhaseInfoUndefined, err
		}
	}

	recovered, err := c.recoveryClient.RecoverNodeExecution(ctx, nCtx.ExecutionContext().GetExecutionConfig().RecoveryExecution.WorkflowExecutionIdentifier, fullyQualifiedNodeID)
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
		return handler.PhaseInfoUndefined, nil
	case core.NodeExecution_SUCCEEDED:
		fallthrough
	case core.NodeExecution_RECOVERED:
		logger.Debugf(ctx, "Node [%+v] can be recovered. Proceeding to copy inputs and outputs", nCtx.NodeExecutionMetadata().GetNodeExecutionID())
	default:
		// The node execution may be partially recoverable through intra task checkpointing. Save the checkpoint
		// uri in the task node state to pass to the task handler later on.
		if metadata, ok := recovered.Closure.TargetMetadata.(*admin.NodeExecutionClosure_TaskNodeMetadata); ok {
			state := nCtx.NodeStateReader().GetTaskNodeState()
			state.PreviousNodeExecutionCheckpointURI = storage.DataReference(metadata.TaskNodeMetadata.CheckpointUri)
			err = nCtx.NodeStateWriter().PutTaskNodeState(state)
			if err != nil {
				logger.Warnf(ctx, "failed to save recovered checkpoint uri for [%+v]: [%+v]",
					nCtx.NodeExecutionMetadata().GetNodeExecutionID(), err)
			}
		}

		// if this node is a dynamic task we attempt to recover the compiled workflow from instances where the parent
		// task succeeded but the dynamic task did not complete. this is important to ensure correctness since node ids
		// within the compiled closure may not be generated deterministically.
		if recovered.Metadata != nil && recovered.Metadata.IsDynamic && len(recovered.Closure.DynamicJobSpecUri) > 0 {
			// recover node inputs
			recoveredData, err := c.recoveryClient.RecoverNodeExecutionData(ctx,
				nCtx.ExecutionContext().GetExecutionConfig().RecoveryExecution.WorkflowExecutionIdentifier, fullyQualifiedNodeID)
			if err != nil || recoveredData == nil {
				return handler.PhaseInfoUndefined, nil
			}

			if _, err := c.recoverInputs(ctx, nCtx, recovered, recoveredData); err != nil {
				return handler.PhaseInfoUndefined, err
			}

			// copy previous DynamicJobSpec file
			f, err := task.NewRemoteFutureFileReader(ctx, nCtx.NodeStatus().GetOutputDir(), nCtx.DataStore())
			if err != nil {
				return handler.PhaseInfoUndefined, err
			}

			dynamicJobSpecReference := storage.DataReference(recovered.Closure.DynamicJobSpecUri)
			if err := nCtx.DataStore().CopyRaw(ctx, dynamicJobSpecReference, f.GetLoc(), storage.Options{}); err != nil {
				return handler.PhaseInfoUndefined, errors.Wrapf(errors.StorageError, nCtx.NodeID(), err,
					"failed to store dynamic job spec for node. source file [%s] destination file [%s]", dynamicJobSpecReference, f.GetLoc())
			}

			// transition node phase to 'Running' and dynamic task phase to 'DynamicNodePhaseParentFinalized'
			state := nCtx.NodeStateReader().GetDynamicNodeState()
			state.Phase = v1alpha1.DynamicNodePhaseParentFinalized
			if err := nCtx.NodeStateWriter().PutDynamicNodeState(state); err != nil {
				return handler.PhaseInfoUndefined, errors.Wrapf(errors.UnknownError, nCtx.NodeID(), err, "failed to store dynamic node state")
			}

			return handler.PhaseInfoRunning(&handler.ExecutionInfo{}), nil
		}

		logger.Debugf(ctx, "Node [%+v] phase [%v] is not recoverable", nCtx.NodeExecutionMetadata().GetNodeExecutionID(), recovered.Closure.Phase)
		return handler.PhaseInfoUndefined, nil
	}

	recoveredData, err := c.recoveryClient.RecoverNodeExecutionData(ctx, nCtx.ExecutionContext().GetExecutionConfig().RecoveryExecution.WorkflowExecutionIdentifier, fullyQualifiedNodeID)
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
	nodeInputs, err := c.recoverInputs(ctx, nCtx, recovered, recoveredData)
	if err != nil {
		return handler.PhaseInfoUndefined, err
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
	oi := &handler.OutputInfo{
		OutputURI: outputFile,
	}

	deckFile := storage.DataReference(recovered.Closure.GetDeckUri())
	if len(deckFile) > 0 {
		metadata, err := nCtx.DataStore().Head(ctx, deckFile)
		if err != nil {
			logger.Errorf(ctx, "Failed to check the existence of deck file. Error: %v", err)
			return handler.PhaseInfoUndefined, errors.Wrapf(errors.CausedByError, nCtx.NodeID(), err, "Failed to check the existence of deck file.")
		}

		if metadata.Exists() {
			oi.DeckURI = &deckFile
		}
	}

	if err := c.store.WriteProtobuf(ctx, outputFile, so, outputs); err != nil {
		logger.Errorf(ctx, "Failed to write protobuf (metadata). Error [%v]", err)
		return handler.PhaseInfoUndefined, errors.Wrapf(errors.CausedByError, nCtx.NodeID(), err, "Failed to store recovered node execution outputs")
	}

	info := &handler.ExecutionInfo{
		Inputs:     nodeInputs,
		OutputInfo: oi,
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
func (c *nodeExecutor) preExecute(ctx context.Context, dag executors.DAGStructure, nCtx interfaces.NodeExecutionContext) (
	handler.PhaseInfo, error) {
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
				if err != nil || phaseInfo.GetPhase() != handler.EPhaseUndefined {
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

		return handler.PhaseInfoQueued("node queued", nodeInputs), nil
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

func (c *nodeExecutor) isEligibleForRetry(nCtx interfaces.NodeExecutionContext, nodeStatus v1alpha1.ExecutableNodeStatus, err *core.ExecutionError) (currentAttempt uint32, maxAttempts uint32, isEligible bool) {
	if config.GetConfig().NodeConfig.IgnoreRetryCause {
		currentAttempt = nodeStatus.GetAttempts() + 1
	} else {
		if err.Kind == core.ExecutionError_SYSTEM {
			currentAttempt = nodeStatus.GetSystemFailures()
			maxAttempts = c.maxNodeRetriesForSystemFailures
			isEligible = currentAttempt < c.maxNodeRetriesForSystemFailures
			return
		}

		currentAttempt = (nodeStatus.GetAttempts() + 1) - nodeStatus.GetSystemFailures()
	}
	maxAttempts = uint32(config.GetConfig().NodeConfig.DefaultMaxAttempts)
	if nCtx.Node().GetRetryStrategy() != nil && nCtx.Node().GetRetryStrategy().MinAttempts != nil && *nCtx.Node().GetRetryStrategy().MinAttempts != 1 {
		maxAttempts = uint32(*nCtx.Node().GetRetryStrategy().MinAttempts)
	}
	isEligible = currentAttempt < maxAttempts
	return
}

func (c *nodeExecutor) execute(ctx context.Context, h interfaces.NodeHandler, nCtx interfaces.NodeExecutionContext, nodeStatus v1alpha1.ExecutableNodeStatus) (handler.PhaseInfo, error) {
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
			logger.Infof(ctx, "Node has timed out; timeout configured: %v", activeDeadline)
			return handler.PhaseInfoTimedOut(nil, fmt.Sprintf("task active timeout [%s] expired", activeDeadline.String())), nil
		}

		// Execution timeout is a retry-able error
		executionDeadline := c.defaultExecutionDeadline
		if nCtx.Node().GetExecutionDeadline() != nil && *nCtx.Node().GetExecutionDeadline() > 0 {
			executionDeadline = *nCtx.Node().GetExecutionDeadline()
		}
		if isTimeoutExpired(nodeStatus.GetLastAttemptStartedAt(), executionDeadline) {
			logger.Infof(ctx, "Current execution for the node timed out; timeout configured: %v", executionDeadline)
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
		nCtx.NodeStateWriter().ClearNodeStatus()
	}

	return phase, nil
}

func (c *nodeExecutor) Abort(ctx context.Context, h interfaces.NodeHandler, nCtx interfaces.NodeExecutionContext, reason string, finalTransition bool) error {
	logger.Debugf(ctx, "Calling aborting & finalize")
	if err := h.Abort(ctx, nCtx, reason); err != nil {
		finalizeErr := h.Finalize(ctx, nCtx)
		if finalizeErr != nil {
			return errors.ErrorCollection{Errors: []error{err, finalizeErr}}
		}
		return err
	}

	if err := h.Finalize(ctx, nCtx); err != nil {
		return err
	}

	// only send event if this is the final transition for this node
	if finalTransition {
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

		err := nCtx.EventsRecorder().RecordNodeEvent(ctx, &event.NodeExecutionEvent{
			Id:         nodeExecutionID,
			Phase:      core.NodeExecution_ABORTED,
			OccurredAt: ptypes.TimestampNow(),
			OutputResult: &event.NodeExecutionEvent_Error{
				Error: &core.ExecutionError{
					Code:    "NodeAborted",
					Message: reason,
				},
			},
			ProducerId: c.clusterID,
			ReportedAt: ptypes.TimestampNow(),
		}, c.eventConfig)
		if err != nil && !eventsErr.IsNotFound(err) && !eventsErr.IsEventIncompatibleClusterError(err) {
			if errors2.IsCausedBy(err, errors.IllegalStateError) {
				logger.Debugf(ctx, "Failed to record abort event due to illegal state transition. Ignoring the error. Error: %v", err)
			} else {
				logger.Warningf(ctx, "Failed to record nodeEvent, error [%s]", err.Error())
				return errors.Wrapf(errors.EventRecordingFailed, nCtx.NodeID(), err, "failed to record node event")
			}
		}
	}

	return nil
}

func (c *nodeExecutor) Finalize(ctx context.Context, h interfaces.NodeHandler, nCtx interfaces.NodeExecutionContext) error {
	return h.Finalize(ctx, nCtx)
}

func (c *nodeExecutor) handleNotYetStartedNode(ctx context.Context, dag executors.DAGStructure, nCtx interfaces.NodeExecutionContext, h interfaces.NodeHandler) (interfaces.NodeStatus, error) {
	nodeStatus := nCtx.NodeStatus()

	logger.Debugf(ctx, "Node not yet started, running pre-execute")
	defer logger.Debugf(ctx, "Node pre-execute completed")
	occurredAt := time.Now()
	p, err := c.preExecute(ctx, dag, nCtx)
	if err != nil {
		logger.Errorf(ctx, "failed preExecute for node. Error: %s", err.Error())
		return interfaces.NodeStatusUndefined, err
	}

	if p.GetPhase() == handler.EPhaseUndefined {
		return interfaces.NodeStatusUndefined, errors.Errorf(errors.IllegalStateError, nCtx.NodeID(), "received undefined phase.")
	}

	if p.GetPhase() == handler.EPhaseNotReady {
		return interfaces.NodeStatusPending, nil
	}

	var cacheStatus *catalog.Status
	if cacheHandler, ok := h.(interfaces.CacheableNodeHandler); p.GetPhase() != handler.EPhaseSkip && ok {
		cacheable, _, err := cacheHandler.IsCacheable(ctx, nCtx)
		if err != nil {
			logger.Errorf(ctx, "failed to determine if node is cacheable with err '%s'", err.Error())
			return interfaces.NodeStatusUndefined, err
		} else if cacheable && nCtx.ExecutionContext().GetExecutionConfig().OverwriteCache {
			logger.Info(ctx, "execution config forced cache skip, not checking catalog")
			status := catalog.NewStatus(core.CatalogCacheStatus_CACHE_SKIPPED, nil)
			cacheStatus = &status
			c.metrics.catalogSkipCount.Inc(ctx)
		} else if cacheable {
			entry, err := c.CheckCatalogCache(ctx, nCtx, cacheHandler)
			if err != nil {
				logger.Errorf(ctx, "failed to check the catalog cache with err '%s'", err.Error())
				return interfaces.NodeStatusUndefined, err
			}

			status := entry.GetStatus()
			cacheStatus = &status
			if entry.GetStatus().GetCacheStatus() == core.CatalogCacheStatus_CACHE_HIT {
				// if cache hit we immediately transition the node to successful
				outputFile := v1alpha1.GetOutputsFile(nCtx.NodeStatus().GetOutputDir())
				p = handler.PhaseInfoSuccess(&handler.ExecutionInfo{
					OutputInfo: &handler.OutputInfo{
						OutputURI: outputFile,
					},
					TaskNodeInfo: &handler.TaskNodeInfo{
						TaskNodeMetadata: &event.TaskNodeMetadata{
							CacheStatus: entry.GetStatus().GetCacheStatus(),
							CatalogKey:  entry.GetStatus().GetMetadata(),
						},
					},
				})
			}
		}
	}

	np, err := ToNodePhase(p.GetPhase())
	if err != nil {
		return interfaces.NodeStatusUndefined, errors.Wrapf(errors.IllegalStateError, nCtx.NodeID(), err, "failed to move from queued")
	}

	if np == v1alpha1.NodePhaseSucceeding && (!h.FinalizeRequired() || (cacheStatus != nil && cacheStatus.GetCacheStatus() == core.CatalogCacheStatus_CACHE_HIT)) {
		logger.Infof(ctx, "Finalize not required, moving node to Succeeded")
		np = v1alpha1.NodePhaseSucceeded
	}

	if np != nodeStatus.GetPhase() {
		// assert np == Queued!
		logger.Infof(ctx, "Change in node state detected from [%s] -> [%s]", nodeStatus.GetPhase().String(), np.String())
		p = p.WithOccuredAt(occurredAt)

		nev, err := ToNodeExecutionEvent(nCtx.NodeExecutionMetadata().GetNodeExecutionID(),
			p, nCtx.InputReader().GetInputPath().String(), nodeStatus, nCtx.ExecutionContext().GetEventVersion(),
			nCtx.ExecutionContext().GetParentInfo(), nCtx.Node(), c.clusterID, nCtx.NodeStateReader().GetDynamicNodeState().Phase,
			c.eventConfig)
		if err != nil {
			return interfaces.NodeStatusUndefined, errors.Wrapf(errors.IllegalStateError, nCtx.NodeID(), err, "could not convert phase info to event")
		}
		err = nCtx.EventsRecorder().RecordNodeEvent(ctx, nev, c.eventConfig)
		if err != nil {
			logger.Warningf(ctx, "Failed to record nodeEvent, error [%s]", err.Error())
			return interfaces.NodeStatusUndefined, errors.Wrapf(errors.EventRecordingFailed, nCtx.NodeID(), err, "failed to record node event")
		}
		UpdateNodeStatus(np, p, nCtx.NodeStateReader(), nodeStatus, c.enableCRDebugMetadata)
		c.RecordTransitionLatency(ctx, dag, nCtx.ContextualNodeLookup(), nCtx.Node(), nodeStatus)
	}

	if np == v1alpha1.NodePhaseQueued {
		if nCtx.NodeExecutionMetadata().IsInterruptible() {
			c.metrics.InterruptibleNodesRunning.Inc(ctx)
		}
		return interfaces.NodeStatusQueued, nil
	} else if np == v1alpha1.NodePhaseSkipped {
		return interfaces.NodeStatusSuccess, nil
	}

	return interfaces.NodeStatusPending, nil
}

func (c *nodeExecutor) handleQueuedOrRunningNode(ctx context.Context, nCtx interfaces.NodeExecutionContext, h interfaces.NodeHandler) (interfaces.NodeStatus, error) {
	nodeStatus := nCtx.NodeStatus()
	currentPhase := nodeStatus.GetPhase()
	p := handler.PhaseInfoUndefined

	// case v1alpha1.NodePhaseQueued, v1alpha1.NodePhaseRunning:
	logger.Debugf(ctx, "node executing, current phase [%s]", currentPhase)
	defer logger.Debugf(ctx, "node execution completed")

	var cacheStatus *catalog.Status
	var catalogReservationStatus *core.CatalogReservation_Status
	cacheHandler, cacheHandlerOk := h.(interfaces.CacheableNodeHandler)
	if cacheHandlerOk {
		// if node is cacheable we attempt to check the cache if in queued phase or get / extend a
		// catalog reservation
		_, cacheSerializable, err := cacheHandler.IsCacheable(ctx, nCtx)
		if err != nil {
			logger.Errorf(ctx, "failed to determine if node is cacheable with err '%s'", err.Error())
			return interfaces.NodeStatusUndefined, err
		}

		if cacheSerializable && nCtx.ExecutionContext().GetExecutionConfig().OverwriteCache {
			status := catalog.NewStatus(core.CatalogCacheStatus_CACHE_SKIPPED, nil)
			cacheStatus = &status
		} else if cacheSerializable && currentPhase == v1alpha1.NodePhaseQueued && nodeStatus.GetMessage() == cacheSerializedReason {
			// since we already check the cache before transitioning to Phase Queued we only need to check it again if
			// the cache is serialized and that causes the node to stay in the Queued phase. the easiest way to detect
			// this is verifying the NodeStatus Reason is what we set it during cache serialization.

			entry, err := c.CheckCatalogCache(ctx, nCtx, cacheHandler)
			if err != nil {
				logger.Errorf(ctx, "failed to check the catalog cache with err '%s'", err.Error())
				return interfaces.NodeStatusUndefined, err
			}

			status := entry.GetStatus()
			cacheStatus = &status
			if entry.GetStatus().GetCacheStatus() == core.CatalogCacheStatus_CACHE_HIT {
				// if cache hit we immediately transition the node to successful
				outputFile := v1alpha1.GetOutputsFile(nCtx.NodeStatus().GetOutputDir())
				p = handler.PhaseInfoSuccess(&handler.ExecutionInfo{
					OutputInfo: &handler.OutputInfo{
						OutputURI: outputFile,
					},
					TaskNodeInfo: &handler.TaskNodeInfo{
						TaskNodeMetadata: &event.TaskNodeMetadata{
							CacheStatus: entry.GetStatus().GetCacheStatus(),
							CatalogKey:  entry.GetStatus().GetMetadata(),
						},
					},
				})
			}
		}

		if cacheSerializable && !nCtx.ExecutionContext().GetExecutionConfig().OverwriteCache &&
			(cacheStatus == nil || (cacheStatus.GetCacheStatus() != core.CatalogCacheStatus_CACHE_HIT)) {

			entry, err := c.GetOrExtendCatalogReservation(ctx, nCtx, cacheHandler, config.GetConfig().WorkflowReEval.Duration)
			if err != nil {
				logger.Errorf(ctx, "failed to check for catalog reservation with err '%s'", err.Error())
				return interfaces.NodeStatusUndefined, err
			}

			status := entry.GetStatus()
			catalogReservationStatus = &status
			if status == core.CatalogReservation_RESERVATION_ACQUIRED && currentPhase == v1alpha1.NodePhaseQueued {
				logger.Infof(ctx, "acquired cache reservation")
			} else if status == core.CatalogReservation_RESERVATION_EXISTS {
				// if reservation is held by another owner we stay in the queued phase
				p = handler.PhaseInfoQueued(cacheSerializedReason, nil)
			}
		}
	}

	// Since we reset node status inside execute for retryable failure, we use lastAttemptStartTime to carry that information
	// across execute which is used to emit metrics
	lastAttemptStartTime := nodeStatus.GetLastAttemptStartedAt()

	// we execute the node if:
	//  (1) caching is disabled (ie. cacheStatus == nil)
	//  (2) there was no cache hit and the cache is not blocked by a cache reservation
	//  (3) the node is already running, this covers the scenario where the node held the cache
	//      reservation, but it expired and was captured by a different node
	if currentPhase != v1alpha1.NodePhaseQueued ||
		((cacheStatus == nil || cacheStatus.GetCacheStatus() != core.CatalogCacheStatus_CACHE_HIT) &&
			(catalogReservationStatus == nil || *catalogReservationStatus != core.CatalogReservation_RESERVATION_EXISTS)) {

		var err error
		p, err = c.execute(ctx, h, nCtx, nodeStatus)
		if err != nil {
			logger.Errorf(ctx, "failed Execute for node. Error: %s", err.Error())
			return interfaces.NodeStatusUndefined, err
		}
	}

	if p.GetPhase() == handler.EPhaseUndefined {
		return interfaces.NodeStatusUndefined, errors.Errorf(errors.IllegalStateError, nCtx.NodeID(), "received undefined phase.")
	}

	if p.GetPhase() == handler.EPhaseSuccess && cacheHandlerOk {
		// if node is cacheable we attempt to write outputs to the cache and release catalog reservation
		cacheable, cacheSerializable, err := cacheHandler.IsCacheable(ctx, nCtx)
		if err != nil {
			logger.Errorf(ctx, "failed to determine if node is cacheable with err '%s'", err.Error())
			return interfaces.NodeStatusUndefined, err
		}

		if cacheable && (cacheStatus == nil || cacheStatus.GetCacheStatus() != core.CatalogCacheStatus_CACHE_HIT) {
			status, err := c.WriteCatalogCache(ctx, nCtx, cacheHandler)
			if err != nil {
				// ignore failure to write to catalog
				logger.Warnf(ctx, "failed to write to the catalog cache with err '%s'", err.Error())
			}

			cacheStatus = &status
		}

		if cacheSerializable && (cacheStatus == nil || cacheStatus.GetCacheStatus() != core.CatalogCacheStatus_CACHE_HIT) {
			entry, err := c.ReleaseCatalogReservation(ctx, nCtx, cacheHandler)
			if err != nil {
				// ignore failure to release the catalog reservation
				logger.Warnf(ctx, "failed to write to the catalog cache with err '%s'", err.Error())
			} else {
				status := entry.GetStatus()
				catalogReservationStatus = &status
			}
		}
	}

	// update phase info with catalog cache and reservation information while maintaining all
	// other metadata
	p = updatePhaseCacheInfo(p, cacheStatus, catalogReservationStatus)

	np, err := ToNodePhase(p.GetPhase())
	if err != nil {
		return interfaces.NodeStatusUndefined, errors.Wrapf(errors.IllegalStateError, nCtx.NodeID(), err, "failed to move from queued")
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
	finalStatus := interfaces.NodeStatusRunning
	if np == v1alpha1.NodePhaseFailing && !h.FinalizeRequired() {
		logger.Infof(ctx, "Finalize not required, moving node to Failed")
		np = v1alpha1.NodePhaseFailed
		finalStatus = interfaces.NodeStatusFailed(p.GetErr())
	}

	if np == v1alpha1.NodePhaseTimingOut && !h.FinalizeRequired() {
		logger.Infof(ctx, "Finalize not required, moving node to TimedOut")
		np = v1alpha1.NodePhaseTimedOut
		finalStatus = interfaces.NodeStatusTimedOut
	}

	if np == v1alpha1.NodePhaseSucceeding && (!h.FinalizeRequired() || (cacheStatus != nil && cacheStatus.GetCacheStatus() == core.CatalogCacheStatus_CACHE_HIT)) {
		logger.Infof(ctx, "Finalize not required, moving node to Succeeded")
		np = v1alpha1.NodePhaseSucceeded
		finalStatus = interfaces.NodeStatusSuccess
	}

	if np == v1alpha1.NodePhaseRecovered {
		logger.Infof(ctx, "Finalize not required, moving node to Recovered")
		finalStatus = interfaces.NodeStatusRecovered
	}

	// If it is retryable failure, we do no want to send any events, as the node is essentially still running
	// Similarly if the phase has not changed from the last time, events do not need to be sent
	if np != nodeStatus.GetPhase() && np != v1alpha1.NodePhaseRetryableFailure {
		// assert np == skipped, succeeding, failing or recovered
		logger.Infof(ctx, "Change in node state detected from [%s] -> [%s], (handler phase [%s])", nodeStatus.GetPhase().String(), np.String(), p.GetPhase().String())

		nev, err := ToNodeExecutionEvent(nCtx.NodeExecutionMetadata().GetNodeExecutionID(),
			p, nCtx.InputReader().GetInputPath().String(), nCtx.NodeStatus(), nCtx.ExecutionContext().GetEventVersion(),
			nCtx.ExecutionContext().GetParentInfo(), nCtx.Node(), c.clusterID, nCtx.NodeStateReader().GetDynamicNodeState().Phase,
			c.eventConfig)
		if err != nil {
			return interfaces.NodeStatusUndefined, errors.Wrapf(errors.IllegalStateError, nCtx.NodeID(), err, "could not convert phase info to event")
		}

		err = nCtx.EventsRecorder().RecordNodeEvent(ctx, nev, c.eventConfig)
		if err != nil {
			if eventsErr.IsTooLarge(err) || eventsErr.IsInvalidArguments(err) {
				// we immediately transition to failing if one of two scenarios occur during node event recording:
				// (1) the event is too large to be sent over gRPC. this can occur if, for example, a dynamic task
				// has a very large fanout and the compiled workflow closure causes the event to exceed the gRPC
				// message size limit.
				// (2) the event is invalid. this can occur if, for example, a dynamic task compiles a workflow
				// which is invalid per admin limits (ex. maximum resources exceeded).
				np = v1alpha1.NodePhaseFailing
				p = handler.PhaseInfoFailure(core.ExecutionError_USER, "NodeFailed", err.Error(), p.GetInfo())

				err = nCtx.EventsRecorder().RecordNodeEvent(ctx, &event.NodeExecutionEvent{
					Id:         nCtx.NodeExecutionMetadata().GetNodeExecutionID(),
					Phase:      core.NodeExecution_FAILED,
					OccurredAt: ptypes.TimestampNow(),
					OutputResult: &event.NodeExecutionEvent_Error{
						Error: &core.ExecutionError{
							Code:    "NodeFailed",
							Message: err.Error(),
						},
					},
					ReportedAt: ptypes.TimestampNow(),
				}, c.eventConfig)

				if err != nil {
					return interfaces.NodeStatusUndefined, errors.Wrapf(errors.EventRecordingFailed, nCtx.NodeID(), err, "failed to record node event")
				}
			} else {
				logger.Warningf(ctx, "Failed to record nodeEvent, error [%s]", err.Error())
				return interfaces.NodeStatusUndefined, errors.Wrapf(errors.EventRecordingFailed, nCtx.NodeID(), err, "failed to record node event")
			}
		}

		// We reach here only when transitioning from Queued to Running. In this case, the startedAt is not set.
		if np == v1alpha1.NodePhaseRunning {
			if nodeStatus.GetQueuedAt() != nil {
				c.metrics.QueuingLatency.Observe(ctx, nodeStatus.GetQueuedAt().Time, time.Now())
			}
		}
	}

	UpdateNodeStatus(np, p, nCtx.NodeStateReader(), nodeStatus, c.enableCRDebugMetadata)
	return finalStatus, nil
}

func (c *nodeExecutor) handleRetryableFailure(ctx context.Context, nCtx interfaces.NodeExecutionContext, h interfaces.NodeHandler) (interfaces.NodeStatus, error) {
	nodeStatus := nCtx.NodeStatus()
	logger.Debugf(ctx, "node failed with retryable failure, aborting and finalizing, message: %s", nodeStatus.GetMessage())
	if err := c.Abort(ctx, h, nCtx, nodeStatus.GetMessage(), false); err != nil {
		return interfaces.NodeStatusUndefined, err
	}

	// NOTE: It is important to increment attempts only after abort has been called. Increment attempt mutates the state
	// Attempt is used throughout the system to determine the idempotent resource version.
	nodeStatus.IncrementAttempts()
	nodeStatus.UpdatePhase(v1alpha1.NodePhaseRunning, metav1.Now(), "retrying", c.enableCRDebugMetadata, nil)
	// We are going to retry in the next round, so we should clear all current state
	nodeStatus.ClearSubNodeStatus()
	nodeStatus.ClearTaskStatus()
	nodeStatus.ClearWorkflowStatus()
	nodeStatus.ClearDynamicNodeStatus()
	nodeStatus.ClearGateNodeStatus()
	nodeStatus.ClearArrayNodeStatus()
	return interfaces.NodeStatusPending, nil
}

func (c *nodeExecutor) HandleNode(ctx context.Context, dag executors.DAGStructure, nCtx interfaces.NodeExecutionContext, h interfaces.NodeHandler) (interfaces.NodeStatus, error) {
	ctx, span := otelutils.NewSpan(ctx, otelutils.FlytePropellerTracer, "pkg.controller.nodes.NodeExecutor/handleNode")
	defer span.End()

	logger.Debugf(ctx, "Handling Node [%s]", nCtx.NodeID())
	defer logger.Debugf(ctx, "Completed node [%s]", nCtx.NodeID())

	nodeStatus := nCtx.NodeStatus()
	currentPhase := nodeStatus.GetPhase()

	// Optimization!
	// If it is start node we directly move it to Queued without needing to run preExecute
	if currentPhase == v1alpha1.NodePhaseNotYetStarted && !nCtx.Node().IsStartNode() {
		p, err := c.handleNotYetStartedNode(ctx, dag, nCtx, h)
		if err != nil {
			return p, err
		}
		if p.NodePhase == interfaces.NodePhaseQueued {
			logger.Infof(ctx, "Node was queued, parallelism is now [%d]", nCtx.ExecutionContext().IncrementParallelism())
		}
		return p, err
	}

	if currentPhase == v1alpha1.NodePhaseFailing {
		logger.Debugf(ctx, "node failing")
		if err := c.Abort(ctx, h, nCtx, "node failing", false); err != nil {
			return interfaces.NodeStatusUndefined, err
		}
		t := metav1.Now()

		startedAt := nodeStatus.GetStartedAt()
		if startedAt == nil {
			startedAt = &t
		}
		nodeStatus.UpdatePhase(v1alpha1.NodePhaseFailed, t, nodeStatus.GetMessage(), c.enableCRDebugMetadata, nodeStatus.GetExecutionError())
		c.metrics.FailureDuration.Observe(ctx, startedAt.Time, nodeStatus.GetStoppedAt().Time)
		if nCtx.NodeExecutionMetadata().IsInterruptible() {
			c.metrics.InterruptibleNodesTerminated.Inc(ctx)
		}
		return interfaces.NodeStatusFailed(nodeStatus.GetExecutionError()), nil
	}

	if currentPhase == v1alpha1.NodePhaseTimingOut {
		logger.Debugf(ctx, "node timing out")
		if err := c.Abort(ctx, h, nCtx, "node timed out", false); err != nil {
			return interfaces.NodeStatusUndefined, err
		}

		nodeStatus.UpdatePhase(v1alpha1.NodePhaseTimedOut, metav1.Now(), nodeStatus.GetMessage(), c.enableCRDebugMetadata, nodeStatus.GetExecutionError())
		c.metrics.TimedOutFailure.Inc(ctx)
		if nCtx.NodeExecutionMetadata().IsInterruptible() {
			c.metrics.InterruptibleNodesTerminated.Inc(ctx)
		}
		return interfaces.NodeStatusTimedOut, nil
	}

	if currentPhase == v1alpha1.NodePhaseSucceeding {
		logger.Debugf(ctx, "node succeeding")
		if err := c.Finalize(ctx, h, nCtx); err != nil {
			return interfaces.NodeStatusUndefined, err
		}
		t := metav1.Now()

		started := nodeStatus.GetStartedAt()
		if started == nil {
			started = &t
		}
		stopped := nodeStatus.GetStoppedAt()
		if stopped == nil {
			stopped = &t
		}
		c.metrics.SuccessDuration.Observe(ctx, started.Time, stopped.Time)
		nodeStatus.UpdatePhase(v1alpha1.NodePhaseSucceeded, t, "completed successfully", c.enableCRDebugMetadata, nil)
		if nCtx.NodeExecutionMetadata().IsInterruptible() {
			c.metrics.InterruptibleNodesTerminated.Inc(ctx)
		}
		return interfaces.NodeStatusSuccess, nil
	}

	if currentPhase == v1alpha1.NodePhaseRetryableFailure {
		return c.handleRetryableFailure(ctx, nCtx, h)
	}

	if currentPhase == v1alpha1.NodePhaseFailed {
		// This should never happen
		return interfaces.NodeStatusFailed(nodeStatus.GetExecutionError()), nil
	}

	return c.handleQueuedOrRunningNode(ctx, nCtx, h)
}

func NewExecutor(ctx context.Context, nodeConfig config.NodeConfig, store *storage.DataStore, enQWorkflow v1alpha1.EnqueueWorkflow, eventSink events.EventSink,
	workflowLauncher launchplan.Executor, launchPlanReader launchplan.Reader, defaultRawOutputPrefix storage.DataReference, kubeClient executors.Client,
	catalogClient catalog.Client, recoveryClient recovery.Client, eventConfig *config.EventConfig, clusterID string, signalClient service.SignalServiceClient,
	nodeHandlerFactory interfaces.HandlerFactory, scope promutils.Scope) (interfaces.Node, error) {

	// TODO we may want to make this configurable.
	shardSelector, err := ioutils.NewBase36PrefixShardSelector(ctx)
	if err != nil {
		return nil, err
	}

	nodeScope := scope.NewSubScope("node")
	metrics := &nodeMetrics{
		Scope:                          nodeScope,
		FailureDuration:                labeled.NewStopWatch("failure_duration", "Indicates the total execution time of a failed workflow.", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
		SuccessDuration:                labeled.NewStopWatch("success_duration", "Indicates the total execution time of a successful workflow.", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
		RecoveryDuration:               labeled.NewStopWatch("recovery_duration", "Indicates the total execution time of a recovered workflow.", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
		UserErrorDuration:              labeled.NewStopWatch("user_error_duration", "Indicates the total execution time before user error", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
		SystemErrorDuration:            labeled.NewStopWatch("system_error_duration", "Indicates the total execution time before system error", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
		UnknownErrorDuration:           labeled.NewStopWatch("unknown_error_duration", "Indicates the total execution time before unknown error", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
		PermanentUserErrorDuration:     labeled.NewStopWatch("perma_user_error_duration", "Indicates the total execution time before non recoverable user error", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
		PermanentSystemErrorDuration:   labeled.NewStopWatch("perma_system_error_duration", "Indicates the total execution time before non recoverable system error", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
		PermanentUnknownErrorDuration:  labeled.NewStopWatch("perma_unknown_error_duration", "Indicates the total execution time before non recoverable unknown error", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
		InputsWriteFailure:             labeled.NewCounter("inputs_write_fail", "Indicates failure in writing node inputs to metastore", nodeScope),
		TimedOutFailure:                labeled.NewCounter("timeout_fail", "Indicates failure due to timeout", nodeScope),
		InterruptedThresholdHit:        labeled.NewCounter("interrupted_threshold", "Indicates the node interruptible disabled because it hit max failure count", nodeScope),
		InterruptibleNodesRunning:      labeled.NewCounter("interruptible_nodes_running", "number of interruptible nodes running", nodeScope),
		InterruptibleNodesTerminated:   labeled.NewCounter("interruptible_nodes_terminated", "number of interruptible nodes finished running", nodeScope),
		ResolutionFailure:              labeled.NewCounter("input_resolve_fail", "Indicates failure in resolving node inputs", nodeScope),
		TransitionLatency:              labeled.NewStopWatch("transition_latency", "Measures the latency between the last parent node stoppedAt time and current node's queued time.", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
		QueuingLatency:                 labeled.NewStopWatch("queueing_latency", "Measures the latency between the time a node's been queued to the time the handler reported the executable moved to running state", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
		NodeExecutionTime:              labeled.NewStopWatch("node_exec_latency", "Measures the time taken to execute one node, a node can be complex so it may encompass sub-node latency.", time.Microsecond, nodeScope, labeled.EmitUnlabeledMetric),
		NodeInputGatherLatency:         labeled.NewStopWatch("node_input_latency", "Measures the latency to aggregate inputs and check readiness of a node", time.Millisecond, nodeScope, labeled.EmitUnlabeledMetric),
		catalogHitCount:                labeled.NewCounter("discovery_hit_count", "Task cached in Discovery", scope),
		catalogMissCount:               labeled.NewCounter("discovery_miss_count", "Task not cached in Discovery", scope),
		catalogSkipCount:               labeled.NewCounter("discovery_skip_count", "Task cached skipped in Discovery", scope),
		catalogPutSuccessCount:         labeled.NewCounter("discovery_put_success_count", "Discovery Put success count", scope),
		catalogPutFailureCount:         labeled.NewCounter("discovery_put_failure_count", "Discovery Put failure count", scope),
		catalogGetFailureCount:         labeled.NewCounter("discovery_get_failure_count", "Discovery Get failure count", scope),
		reservationGetFailureCount:     labeled.NewCounter("reservation_get_failure_count", "Reservation GetOrExtend failure count", scope),
		reservationGetSuccessCount:     labeled.NewCounter("reservation_get_success_count", "Reservation GetOrExtend success count", scope),
		reservationReleaseFailureCount: labeled.NewCounter("reservation_release_failure_count", "Reservation Release failure count", scope),
		reservationReleaseSuccessCount: labeled.NewCounter("reservation_release_success_count", "Reservation Release success count", scope),
	}

	nodeExecutor := &nodeExecutor{
		catalog:                         catalogClient,
		clusterID:                       clusterID,
		enableCRDebugMetadata:           nodeConfig.EnableCRDebugMetadata,
		defaultActiveDeadline:           nodeConfig.DefaultDeadlines.DefaultNodeActiveDeadline.Duration,
		defaultDataSandbox:              defaultRawOutputPrefix,
		defaultExecutionDeadline:        nodeConfig.DefaultDeadlines.DefaultNodeExecutionDeadline.Duration,
		enqueueWorkflow:                 enQWorkflow,
		eventConfig:                     eventConfig,
		interruptibleFailureThreshold:   nodeConfig.InterruptibleFailureThreshold,
		maxNodeRetriesForSystemFailures: uint32(nodeConfig.MaxNodeRetriesOnSystemFailures),
		metrics:                         metrics,
		nodeRecorder:                    events.NewNodeEventRecorder(eventSink, nodeScope, store),
		outputResolver:                  NewRemoteFileOutputResolver(store),
		recoveryClient:                  recoveryClient,
		shardSelector:                   shardSelector,
		store:                           store,
		taskRecorder:                    events.NewTaskEventRecorder(eventSink, scope.NewSubScope("task"), store),
	}

	exec := &recursiveNodeExecutor{
		nodeExecutor:       nodeExecutor,
		nCtxBuilder:        nodeExecutor,
		nodeHandlerFactory: nodeHandlerFactory,
		enqueueWorkflow:    enQWorkflow,
		store:              store,
		metrics:            metrics,
	}
	return exec, err
}
