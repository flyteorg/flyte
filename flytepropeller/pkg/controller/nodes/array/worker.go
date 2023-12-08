package array

import (
	"context"
	"fmt"

	idlcore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
)

// TODO @hamersaw - docs
type nodeExecutionRequest struct {
	ctx                context.Context
	index              int
	nodePhase		   v1alpha1.NodePhase
	taskPhase		   int
    nodeExecutor       interfaces.Node
    executionContext   executors.ExecutionContext
    dagStructure       executors.DAGStructure
    nodeLookup         executors.NodeLookup
    subNodeSpec        *v1alpha1.NodeSpec
    subNodeStatus      *v1alpha1.NodeStatus
	arrayEventRecorder arrayEventRecorder
	responseChannel    chan struct {interfaces.NodeStatus; error}
}

// TODO @hamersaw - docs
type gatherOutputsRequest struct {
	ctx             context.Context
	reader          *ioutils.RemoteFileOutputReader
	responseChannel chan struct {literalMap map[string]*idlcore.Literal; error}
}

// TODO @hamersaw - docs
type worker struct {
	gatherOutputsRequestChannel chan *gatherOutputsRequest
	nodeExecutionRequestChannel chan *nodeExecutionRequest
}

// TODO @hamersaw - docs
func (w *worker) run() {
	for {
		select {
		case nodeExecutionRequest := <-w.nodeExecutionRequestChannel:
			nodeStatus, err := nodeExecutionRequest.nodeExecutor.RecursiveNodeHandler(nodeExecutionRequest.ctx, nodeExecutionRequest.executionContext,
				nodeExecutionRequest.dagStructure, nodeExecutionRequest.nodeLookup, nodeExecutionRequest.subNodeSpec)
			nodeExecutionRequest.responseChannel <- struct {interfaces.NodeStatus; error}{nodeStatus, err}
		case gatherOutputsRequest := <-w.gatherOutputsRequestChannel:
			// read outputs
			outputs, executionErr, err := gatherOutputsRequest.reader.Read(gatherOutputsRequest.ctx)
			if err != nil {
				gatherOutputsRequest.responseChannel <- struct{literalMap map[string]*idlcore.Literal; error}{nil, err}
				continue
			} else if executionErr != nil {
				gatherOutputsRequest.responseChannel <- struct{literalMap map[string]*idlcore.Literal; error}{nil, fmt.Errorf("%s", executionErr.String())}
				continue
			}

			gatherOutputsRequest.responseChannel <- struct{literalMap map[string]*idlcore.Literal; error}{outputs.GetLiterals(), nil}
		}
	}
}
