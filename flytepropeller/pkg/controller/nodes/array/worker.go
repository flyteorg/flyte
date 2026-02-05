package array

import (
	"context"
	"fmt"
	"runtime/debug"

	idlcore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

// nodeExecutionRequest is a request to execute an ArrayNode subNode
type nodeExecutionRequest struct {
	ctx                context.Context
	index              int
	nodePhase          v1alpha1.NodePhase
	taskPhase          int
	nodeExecutor       interfaces.Node
	executionContext   executors.ExecutionContext
	dagStructure       executors.DAGStructure
	nodeLookup         executors.NodeLookup
	subNodeSpec        *v1alpha1.NodeSpec
	subNodeStatus      *v1alpha1.NodeStatus
	arrayEventRecorder arrayEventRecorder
	responseChannel    chan struct {
		interfaces.NodeStatus
		error
	}
}

// gatherOutputsRequest is a request to read outputs from an ArrayNode subNode
type gatherOutputsRequest struct {
	ctx             context.Context
	reader          *ioutils.RemoteFileOutputReader
	responseChannel chan struct {
		literalMap map[string]*idlcore.Literal
		error
	}
}

// worker is an entity that is used to parallelize I/O bound operations for ArrayNode execution
type worker struct {
	gatherOutputsRequestChannel chan *gatherOutputsRequest
	nodeExecutionRequestChannel chan *nodeExecutionRequest
}

// run starts the main handle loop for the worker
func (w *worker) run() {
	for {
		select {
		case nodeExecutionRequest := <-w.nodeExecutionRequestChannel:
			var nodeStatus interfaces.NodeStatus
			var err error
			func() {
				defer func() {
					if r := recover(); r != nil {
						stack := debug.Stack()
						err = fmt.Errorf("panic when executing ArrayNode subNode, Stack: [%s]", string(stack))
						logger.Errorf(nodeExecutionRequest.ctx, err.Error())
					}
				}()

				// execute RecurseNodeHandler on node
				nodeStatus, err = nodeExecutionRequest.nodeExecutor.RecursiveNodeHandler(nodeExecutionRequest.ctx, nodeExecutionRequest.executionContext,
					nodeExecutionRequest.dagStructure, nodeExecutionRequest.nodeLookup, nodeExecutionRequest.subNodeSpec)
			}()

			nodeExecutionRequest.responseChannel <- struct {
				interfaces.NodeStatus
				error
			}{nodeStatus, err}
		case gatherOutputsRequest := <-w.gatherOutputsRequestChannel:
			var literalMap map[string]*idlcore.Literal
			var err error
			func() {
				defer func() {
					if r := recover(); r != nil {
						stack := debug.Stack()
						err = fmt.Errorf("panic when executing ArrayNode subNode, Stack: [%s]", string(stack))
						logger.Errorf(gatherOutputsRequest.ctx, err.Error())
					}
				}()

				// read outputs
				outputs, executionErr, gatherErr := gatherOutputsRequest.reader.Read(gatherOutputsRequest.ctx)
				if gatherErr != nil {
					err = gatherErr
				} else if executionErr != nil {
					err = fmt.Errorf("%s", executionErr.String())
				} else {
					literalMap = outputs.GetLiterals()
				}
			}()

			gatherOutputsRequest.responseChannel <- struct {
				literalMap map[string]*idlcore.Literal
				error
			}{literalMap, err}
		}
	}
}
