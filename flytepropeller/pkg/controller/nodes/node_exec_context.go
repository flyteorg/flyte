package nodes

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/flytepropeller/events"
	eventsErr "github.com/flyteorg/flyte/flytepropeller/events/errors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors"
	nodeerrors "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/errors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
	"github.com/flyteorg/flyte/flytepropeller/pkg/utils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

const NodeIDLabel = "node-id"
const TaskNameLabel = "task-name"
const NodeInterruptibleLabel = "interruptible"

type eventRecorder struct {
	taskEventRecorder events.TaskEventRecorder
	nodeEventRecorder events.NodeEventRecorder
}

func (e eventRecorder) RecordTaskEvent(ctx context.Context, ev *event.TaskExecutionEvent, eventConfig *config.EventConfig) error {
	if err := e.taskEventRecorder.RecordTaskEvent(ctx, ev, eventConfig); err != nil {
		if eventsErr.IsAlreadyExists(err) {
			logger.Warningf(ctx, "Failed to record taskEvent, error [%s]. Trying to record state: %s. Ignoring this error!", err.Error(), ev.Phase)
			return nil
		} else if eventsErr.IsEventAlreadyInTerminalStateError(err) {
			if IsTerminalTaskPhase(ev.Phase) {
				// Event is terminal and the stored value in flyteadmin is already terminal. This implies aborted case. So ignoring
				logger.Warningf(ctx, "Failed to record taskEvent, error [%s]. Trying to record state: %s. Ignoring this error!", err.Error(), ev.Phase)
				return nil
			}
			logger.Warningf(ctx, "Failed to record taskEvent in state: %s, error: %s", ev.Phase, err)
			return errors.Wrapf(err, "failed to record task event, as it already exists in terminal state. Event state: %s", ev.Phase)
		}
		return err
	}
	return nil
}

func (e eventRecorder) RecordNodeEvent(ctx context.Context, nodeEvent *event.NodeExecutionEvent, eventConfig *config.EventConfig) error {
	if nodeEvent == nil {
		return fmt.Errorf("event recording attempt of Nil Node execution event")
	}

	if nodeEvent.Id == nil {
		return fmt.Errorf("event recording attempt of with nil node Event ID")
	}

	logger.Infof(ctx, "Recording NodeEvent [%s] phase[%s]", nodeEvent.GetId().String(), nodeEvent.Phase.String())
	err := e.nodeEventRecorder.RecordNodeEvent(ctx, nodeEvent, eventConfig)
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
			return nodeerrors.Wrapf(nodeerrors.IllegalStateError, nodeEvent.Id.NodeId, err, "phase mismatch mismatch between propeller and control plane; Trying to record Node p: %s", nodeEvent.Phase)
		}
	}
	return err
}

type nodeExecMetadata struct {
	v1alpha1.Meta
	nodeExecID                    *core.NodeExecutionIdentifier
	interruptible                 bool
	interruptibleFailureThreshold int32
	nodeLabels                    map[string]string
}

func (e nodeExecMetadata) GetNodeExecutionID() *core.NodeExecutionIdentifier {
	return e.nodeExecID
}

func (e nodeExecMetadata) GetK8sServiceAccount() string {
	return e.Meta.GetServiceAccountName()
}

func (e nodeExecMetadata) GetOwnerID() types.NamespacedName {
	return types.NamespacedName{Name: e.GetName(), Namespace: e.GetNamespace()}
}

func (e nodeExecMetadata) IsInterruptible() bool {
	return e.interruptible
}

func (e nodeExecMetadata) GetInterruptibleFailureThreshold() int32 {
	return e.interruptibleFailureThreshold
}

func (e nodeExecMetadata) GetLabels() map[string]string {
	return e.nodeLabels
}

func (e nodeExecMetadata) GetConsoleURL() string { return e.Meta.GetConsoleURL() }

type nodeExecContext struct {
	store           *storage.DataStore
	tr              interfaces.TaskReader
	md              interfaces.NodeExecutionMetadata
	eventRecorder   interfaces.EventRecorder
	inputs          io.InputReader
	node            v1alpha1.ExecutableNode
	nodeStatus      v1alpha1.ExecutableNodeStatus
	nsm             *nodeStateManager
	enqueueOwner    func() error
	rawOutputPrefix storage.DataReference
	shardSelector   ioutils.ShardSelector
	nl              executors.NodeLookup
	ic              executors.ExecutionContext
}

func (e nodeExecContext) ExecutionContext() executors.ExecutionContext {
	return e.ic
}

func (e nodeExecContext) ContextualNodeLookup() executors.NodeLookup {
	return e.nl
}

func (e nodeExecContext) OutputShardSelector() ioutils.ShardSelector {
	return e.shardSelector
}

func (e nodeExecContext) RawOutputPrefix() storage.DataReference {
	return e.rawOutputPrefix
}

func (e nodeExecContext) EnqueueOwnerFunc() func() error {
	return e.enqueueOwner
}

func (e nodeExecContext) TaskReader() interfaces.TaskReader {
	return e.tr
}

func (e nodeExecContext) NodeStateReader() interfaces.NodeStateReader {
	return e.nsm
}

func (e nodeExecContext) NodeStateWriter() interfaces.NodeStateWriter {
	return e.nsm
}

func (e nodeExecContext) DataStore() *storage.DataStore {
	return e.store
}

func (e nodeExecContext) InputReader() io.InputReader {
	return e.inputs
}

func (e nodeExecContext) EventsRecorder() interfaces.EventRecorder {
	return e.eventRecorder
}

func (e nodeExecContext) NodeID() v1alpha1.NodeID {
	return e.node.GetID()
}

func (e nodeExecContext) Node() v1alpha1.ExecutableNode {
	return e.node
}

func (e nodeExecContext) CurrentAttempt() uint32 {
	return e.nodeStatus.GetAttempts()
}

func (e nodeExecContext) NodeStatus() v1alpha1.ExecutableNodeStatus {
	return e.nodeStatus
}

func (e nodeExecContext) NodeExecutionMetadata() interfaces.NodeExecutionMetadata {
	return e.md
}

func newNodeExecContext(_ context.Context, store *storage.DataStore, execContext executors.ExecutionContext, nl executors.NodeLookup,
	node v1alpha1.ExecutableNode, nodeStatus v1alpha1.ExecutableNodeStatus, inputs io.InputReader, interruptible bool, interruptibleFailureThreshold int32,
	taskEventRecorder events.TaskEventRecorder, nodeEventRecorder events.NodeEventRecorder, tr interfaces.TaskReader, nsm *nodeStateManager,
	enqueueOwner func() error, rawOutputPrefix storage.DataReference, outputShardSelector ioutils.ShardSelector) *nodeExecContext {

	md := nodeExecMetadata{
		Meta: execContext,
		nodeExecID: &core.NodeExecutionIdentifier{
			NodeId:      node.GetID(),
			ExecutionId: execContext.GetExecutionID().WorkflowExecutionIdentifier,
		},
		interruptible:                 interruptible,
		interruptibleFailureThreshold: interruptibleFailureThreshold,
	}

	// Copy the wf labels before adding node specific labels.
	nodeLabels := make(map[string]string)
	for k, v := range execContext.GetLabels() {
		nodeLabels[k] = v
	}
	nodeLabels[NodeIDLabel] = utils.SanitizeLabelValue(node.GetID())
	if tr != nil && tr.GetTaskID() != nil {
		nodeLabels[TaskNameLabel] = utils.SanitizeLabelValue(tr.GetTaskID().Name)
	}
	nodeLabels[NodeInterruptibleLabel] = strconv.FormatBool(interruptible)
	md.nodeLabels = nodeLabels

	return &nodeExecContext{
		md:         md,
		store:      store,
		node:       node,
		nodeStatus: nodeStatus,
		inputs:     inputs,
		eventRecorder: &eventRecorder{
			taskEventRecorder: taskEventRecorder,
			nodeEventRecorder: nodeEventRecorder,
		},
		tr:              tr,
		nsm:             nsm,
		enqueueOwner:    enqueueOwner,
		rawOutputPrefix: rawOutputPrefix,
		shardSelector:   outputShardSelector,
		nl:              nl,
		ic:              execContext,
	}
}

func isAboveInterruptibleFailureThreshold(numFailures uint32, maxAttempts uint32, interruptibleThreshold int32) bool {
	if interruptibleThreshold > 0 {
		return numFailures >= uint32(interruptibleThreshold)
	}

	return numFailures >= maxAttempts-uint32(-interruptibleThreshold)
}

func (c *nodeExecutor) BuildNodeExecutionContext(ctx context.Context, executionContext executors.ExecutionContext,
	nl executors.NodeLookup, currentNodeID v1alpha1.NodeID) (interfaces.NodeExecutionContext, error) {
	n, ok := nl.GetNode(currentNodeID)
	if !ok {
		return nil, fmt.Errorf("failed to find node with ID [%s] in execution [%s]", currentNodeID, executionContext.GetID())
	}

	var tr interfaces.TaskReader
	if n.GetKind() == v1alpha1.NodeKindTask {
		if n.GetTaskID() == nil {
			return nil, fmt.Errorf("bad state, no task-id defined for node [%s]", n.GetID())
		}
		tk, err := executionContext.GetTask(*n.GetTaskID())
		if err != nil {
			return nil, err
		}
		tr = taskReader{TaskTemplate: tk.CoreTask()}
	}

	workflowEnqueuer := func() error {
		c.enqueueWorkflow(executionContext.GetID())
		return nil
	}

	interruptible := executionContext.IsInterruptible()
	if n.IsInterruptible() != nil {
		interruptible = *n.IsInterruptible()
	}

	s := nl.GetNodeExecutionStatus(ctx, currentNodeID)

	if config.GetConfig().NodeConfig.IgnoreRetryCause {
		// For the unified retry behavior we execute the last interruptibleFailureThreshold attempts on a non
		// interruptible machine
		maxAttempts := uint32(config.GetConfig().NodeConfig.DefaultMaxAttempts)
		if n.GetRetryStrategy() != nil && n.GetRetryStrategy().MinAttempts != nil && *n.GetRetryStrategy().MinAttempts != 1 {
			maxAttempts = uint32(*n.GetRetryStrategy().MinAttempts)
		}

		// For interruptible nodes run at least one attempt on an interruptible machine (thus s.GetAttempts() > 0) even if there won't be any retries
		if interruptible && s.GetAttempts() > 0 && isAboveInterruptibleFailureThreshold(s.GetAttempts(), maxAttempts, c.interruptibleFailureThreshold) {
			interruptible = false
			c.metrics.InterruptedThresholdHit.Inc(ctx)
		}
	} else {
		// Else a node is not considered interruptible if the system failures have exceeded the configured threshold
		if interruptible && isAboveInterruptibleFailureThreshold(s.GetSystemFailures(), c.maxNodeRetriesForSystemFailures+1, c.interruptibleFailureThreshold) {
			interruptible = false
			c.metrics.InterruptedThresholdHit.Inc(ctx)
		}
	}

	rawOutputPrefix := c.defaultDataSandbox
	if executionContext.GetRawOutputDataConfig().RawOutputDataConfig != nil && len(executionContext.GetRawOutputDataConfig().OutputLocationPrefix) > 0 {
		rawOutputPrefix = storage.DataReference(executionContext.GetRawOutputDataConfig().OutputLocationPrefix)
	}

	return newNodeExecContext(ctx, c.store, executionContext, nl, n, s,
		ioutils.NewCachedInputReader(
			ctx,
			ioutils.NewRemoteFileInputReader(
				ctx,
				c.store,
				ioutils.NewInputFilePaths(
					ctx,
					c.store,
					s.GetDataDir(),
				),
			),
		),
		interruptible,
		c.interruptibleFailureThreshold,
		c.taskRecorder,
		c.nodeRecorder,
		tr,
		newNodeStateManager(ctx, s),
		workflowEnqueuer,
		rawOutputPrefix,
		c.shardSelector,
	), nil
}
