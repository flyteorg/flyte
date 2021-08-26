package nodes

import (
	"context"
	"fmt"
	"strconv"

	"github.com/flyteorg/flytepropeller/pkg/controller/events"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/storage"
	"k8s.io/apimachinery/pkg/types"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flytepropeller/pkg/utils"
)

const NodeIDLabel = "node-id"
const TaskNameLabel = "task-name"
const NodeInterruptibleLabel = "interruptible"

type nodeExecMetadata struct {
	v1alpha1.Meta
	nodeExecID     *core.NodeExecutionIdentifier
	interrutptible bool
	nodeLabels     map[string]string
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
	return e.interrutptible
}

func (e nodeExecMetadata) GetLabels() map[string]string {
	return e.nodeLabels
}

type nodeExecContext struct {
	store               *storage.DataStore
	tr                  handler.TaskReader
	md                  handler.NodeExecutionMetadata
	er                  events.TaskEventRecorder
	inputs              io.InputReader
	node                v1alpha1.ExecutableNode
	nodeStatus          v1alpha1.ExecutableNodeStatus
	maxDatasetSizeBytes int64
	nsm                 *nodeStateManager
	enqueueOwner        func() error
	rawOutputPrefix     storage.DataReference
	shardSelector       ioutils.ShardSelector
	nl                  executors.NodeLookup
	ic                  executors.ExecutionContext
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

func (e nodeExecContext) TaskReader() handler.TaskReader {
	return e.tr
}

func (e nodeExecContext) NodeStateReader() handler.NodeStateReader {
	return e.nsm
}

func (e nodeExecContext) NodeStateWriter() handler.NodeStateWriter {
	return e.nsm
}

func (e nodeExecContext) DataStore() *storage.DataStore {
	return e.store
}

func (e nodeExecContext) InputReader() io.InputReader {
	return e.inputs
}

func (e nodeExecContext) EventsRecorder() events.TaskEventRecorder {
	return e.er
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

func (e nodeExecContext) NodeExecutionMetadata() handler.NodeExecutionMetadata {
	return e.md
}

func (e nodeExecContext) MaxDatasetSizeBytes() int64 {
	return e.maxDatasetSizeBytes
}

func newNodeExecContext(_ context.Context, store *storage.DataStore, execContext executors.ExecutionContext, nl executors.NodeLookup,
	node v1alpha1.ExecutableNode, nodeStatus v1alpha1.ExecutableNodeStatus, inputs io.InputReader, interruptible bool,
	maxDatasetSize int64, er events.TaskEventRecorder, tr handler.TaskReader, nsm *nodeStateManager,
	enqueueOwner func() error, rawOutputPrefix storage.DataReference, outputShardSelector ioutils.ShardSelector) *nodeExecContext {

	md := nodeExecMetadata{
		Meta: execContext,
		nodeExecID: &core.NodeExecutionIdentifier{
			NodeId:      node.GetID(),
			ExecutionId: execContext.GetExecutionID().WorkflowExecutionIdentifier,
		},
		interrutptible: interruptible,
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
		md:                  md,
		store:               store,
		node:                node,
		nodeStatus:          nodeStatus,
		inputs:              inputs,
		er:                  er,
		maxDatasetSizeBytes: maxDatasetSize,
		tr:                  tr,
		nsm:                 nsm,
		enqueueOwner:        enqueueOwner,
		rawOutputPrefix:     rawOutputPrefix,
		shardSelector:       outputShardSelector,
		nl:                  nl,
		ic:                  execContext,
	}
}

func (c *nodeExecutor) newNodeExecContextDefault(ctx context.Context, currentNodeID v1alpha1.NodeID,
	executionContext executors.ExecutionContext, nl executors.NodeLookup) (*nodeExecContext, error) {
	n, ok := nl.GetNode(currentNodeID)
	if !ok {
		return nil, fmt.Errorf("failed to find node with ID [%s] in execution [%s]", currentNodeID, executionContext.GetID())
	}

	var tr handler.TaskReader
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

	// a node is not considered interruptible if the system failures have exceeded the configured threshold
	if interruptible && s.GetSystemFailures() >= c.interruptibleFailureThreshold {
		interruptible = false
		c.metrics.InterruptedThresholdHit.Inc(ctx)
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
		c.maxDatasetSizeBytes,
		&taskEventRecorder{TaskEventRecorder: c.taskRecorder},
		tr,
		newNodeStateManager(ctx, s),
		workflowEnqueuer,
		rawOutputPrefix,
		c.shardSelector,
	), nil
}
