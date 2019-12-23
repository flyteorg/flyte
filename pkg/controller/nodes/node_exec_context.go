package nodes

import (
	"context"
	"fmt"

	"github.com/lyft/flyteidl/clients/go/events"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/lyft/flytestdlib/storage"
	"k8s.io/apimachinery/pkg/types"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
	"github.com/lyft/flytepropeller/pkg/utils"
)

const NodeIDLabel = "node-id"
const TaskNameLabel = "task-name"

type execMetadata struct {
	v1alpha1.WorkflowMeta
}

func (e execMetadata) GetK8sServiceAccount() string {
	return e.WorkflowMeta.GetServiceAccountName()
}

func (e execMetadata) GetOwnerID() types.NamespacedName {
	return types.NamespacedName{Name: e.GetName(), Namespace: e.GetNamespace()}
}

type execContext struct {
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
	w                   v1alpha1.ExecutableWorkflow
	nodeLabels          map[string]string
}

func (e execContext) EnqueueOwnerFunc() func() error {
	return e.enqueueOwner
}

func (e execContext) Workflow() v1alpha1.ExecutableWorkflow {
	return e.w
}

func (e execContext) TaskReader() handler.TaskReader {
	return e.tr
}

func (e execContext) NodeStateReader() handler.NodeStateReader {
	return e.nsm
}

func (e execContext) NodeStateWriter() handler.NodeStateWriter {
	return e.nsm
}

func (e execContext) DataStore() *storage.DataStore {
	return e.store
}

func (e execContext) InputReader() io.InputReader {
	return e.inputs
}

func (e execContext) EventsRecorder() events.TaskEventRecorder {
	return e.er
}

func (e execContext) NodeID() v1alpha1.NodeID {
	return e.node.GetID()
}

func (e execContext) Node() v1alpha1.ExecutableNode {
	return e.node
}

func (e execContext) CurrentAttempt() uint32 {
	return e.nodeStatus.GetAttempts()
}

func (e execContext) NodeStatus() v1alpha1.ExecutableNodeStatus {
	return e.nodeStatus
}

func (e execContext) NodeExecutionMetadata() handler.NodeExecutionMetadata {
	return e.md
}

func (e execContext) MaxDatasetSizeBytes() int64 {
	return e.maxDatasetSizeBytes
}

func (e execContext) GetLabels() map[string]string {
	return e.nodeLabels
}

func newNodeExecContext(_ context.Context, store *storage.DataStore, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode, nodeStatus v1alpha1.ExecutableNodeStatus, inputs io.InputReader, maxDatasetSize int64, er events.TaskEventRecorder, tr handler.TaskReader, nsm *nodeStateManager, enqueueOwner func() error) *execContext {
	md := execMetadata{WorkflowMeta: w}

	// Copying the labels before updating it for this node
	nodeLabels := make(map[string]string)
	for k, v := range md.GetLabels() {
		nodeLabels[k] = v
	}
	nodeLabels[NodeIDLabel] = utils.SanitizeLabelValue(node.GetID())
	if tr != nil && tr.GetTaskID() != nil {
		nodeLabels[TaskNameLabel] = utils.SanitizeLabelValue(tr.GetTaskID().Name)
	}

	return &execContext{
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
		w:                   w,
		nodeLabels:          nodeLabels,
	}
}

func (c *nodeExecutor) newNodeExecContextDefault(ctx context.Context, w v1alpha1.ExecutableWorkflow, n v1alpha1.ExecutableNode, s v1alpha1.ExecutableNodeStatus) (*execContext, error) {
	var tr handler.TaskReader
	if n.GetKind() == v1alpha1.NodeKindTask {
		if n.GetTaskID() == nil {
			return nil, fmt.Errorf("bad state, no task-id defined for node [%s]", n.GetID())
		}
		tk, err := w.GetTask(*n.GetTaskID())
		if err != nil {
			return nil, err
		}
		tr = &taskReader{TaskTemplate: tk.CoreTask()}
	}

	workflowEnqueuer := func() error {
		c.enqueueWorkflow(w.GetID())
		return nil
	}

	return newNodeExecContext(ctx, c.store, w, n, s,
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
		c.maxDatasetSizeBytes,
		&taskEventRecorder{TaskEventRecorder: c.taskRecorder},
		tr,
		newNodeStateManager(ctx, s),
		workflowEnqueuer,
	), nil
}
