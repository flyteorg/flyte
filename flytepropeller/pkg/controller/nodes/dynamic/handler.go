package dynamic

import (
	"context"
	"time"

	"github.com/lyft/flytestdlib/promutils/labeled"

	"github.com/lyft/flytepropeller/pkg/compiler"
	common2 "github.com/lyft/flytepropeller/pkg/compiler/common"
	"github.com/lyft/flytepropeller/pkg/controller/executors"

	"github.com/lyft/flytepropeller/pkg/controller/nodes/common"

	"github.com/lyft/flytestdlib/promutils"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/storage"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/compiler/transformers/k8s"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/errors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
)

type dynamicNodeHandler struct {
	handler.IFace
	metrics        metrics
	simpleResolver common.SimpleOutputsResolver
	store          *storage.DataStore
	nodeExecutor   executors.Node
	enQWorkflow    v1alpha1.EnqueueWorkflow
}

type metrics struct {
	buildDynamicWorkflow   labeled.StopWatch
	retrieveDynamicJobSpec labeled.StopWatch
	CacheHit               labeled.StopWatch
	CacheError             labeled.Counter
}

func newMetrics(scope promutils.Scope) metrics {
	return metrics{
		buildDynamicWorkflow:   labeled.NewStopWatch("build_dynamic_workflow", "Overhead for building a dynamic workflow in memory.", time.Microsecond, scope),
		retrieveDynamicJobSpec: labeled.NewStopWatch("retrieve_dynamic_spec", "Overhead of downloading and unmarshaling dynamic job spec", time.Microsecond, scope),
		CacheHit:               labeled.NewStopWatch("dynamic_workflow_cache_hit", "A dynamic workflow was loaded from store.", time.Microsecond, scope),
		CacheError:             labeled.NewCounter("cache_err", "A dynamic workflow failed to store or load from data store.", scope),
	}
}

func (e dynamicNodeHandler) ExtractOutput(ctx context.Context, w v1alpha1.ExecutableWorkflow, n v1alpha1.ExecutableNode,
	bindToVar handler.VarName) (values *core.Literal, err error) {
	outputResolver, casted := e.IFace.(handler.OutputResolver)
	if !casted {
		return e.simpleResolver.ExtractOutput(ctx, w, n, bindToVar)
	}

	return outputResolver.ExtractOutput(ctx, w, n, bindToVar)
}

func (e dynamicNodeHandler) getDynamicJobSpec(ctx context.Context, node v1alpha1.ExecutableNode, dataDir storage.DataReference) (*core.DynamicJobSpec, error) {
	t := e.metrics.retrieveDynamicJobSpec.Start(ctx)
	defer t.Stop()

	futuresFilePath, err := e.store.ConstructReference(ctx, dataDir, v1alpha1.GetFutureFile())
	if err != nil {
		logger.Warnf(ctx, "Failed to construct data path for futures file. Error: %v", err)
		return nil, err
	}

	// If no futures file produced, then declare success and return.
	if metadata, err := e.store.Head(ctx, futuresFilePath); err != nil {
		logger.Warnf(ctx, "Failed to read futures file. Error: %v", err)
		return nil, errors.Wrapf(errors.CausedByError, node.GetID(), err, "Failed to do HEAD on futures file.")
	} else if !metadata.Exists() {
		return nil, nil
	}

	djSpec := &core.DynamicJobSpec{}
	if err := e.store.ReadProtobuf(ctx, futuresFilePath, djSpec); err != nil {
		logger.Warnf(ctx, "Failed to read futures file. Error: %v", err)
		return nil, errors.Wrapf(errors.CausedByError, node.GetID(), err, "Failed to read futures protobuf file.")
	}

	return djSpec, nil
}

func (e dynamicNodeHandler) buildDynamicWorkflowTemplate(ctx context.Context, djSpec *core.DynamicJobSpec,
	w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode, nodeStatus v1alpha1.ExecutableNodeStatus) (
	*core.WorkflowTemplate, error) {

	iface, err := underlyingInterface(w, node)
	if err != nil {
		return nil, err
	}

	// Modify node IDs to include lineage, the entire system assumes node IDs are unique per parent WF.
	// We keep track of the original node ids because that's where inputs are written to.
	parentNodeID := node.GetID()
	for _, n := range djSpec.Nodes {
		newID, err := hierarchicalNodeID(parentNodeID, n.Id)
		if err != nil {
			return nil, err
		}

		// Instantiate a nodeStatus using the modified name but set its data directory using the original name.
		subNodeStatus := nodeStatus.GetNodeExecutionStatus(newID)
		originalNodePath, err := e.store.ConstructReference(ctx, nodeStatus.GetDataDir(), n.Id)
		if err != nil {
			return nil, err
		}

		subNodeStatus.SetDataDir(originalNodePath)
		subNodeStatus.ResetDirty()

		n.Id = newID
	}

	if node.GetTaskID() != nil {
		// If the parent is a task, pass down data children nodes should inherit.
		parentTask, err := w.GetTask(*node.GetTaskID())
		if err != nil {
			return nil, errors.Wrapf(errors.CausedByError, node.GetID(), err, "Failed to find task [%v].", node.GetTaskID())
		}

		for _, t := range djSpec.Tasks {
			if t.GetContainer() != nil && parentTask.CoreTask().GetContainer() != nil {
				t.GetContainer().Config = append(t.GetContainer().Config, parentTask.CoreTask().GetContainer().Config...)
			}
		}
	}

	for _, o := range djSpec.Outputs {
		err = updateBindingNodeIDsWithLineage(parentNodeID, o.Binding)
		if err != nil {
			return nil, err
		}
	}

	return &core.WorkflowTemplate{
		Id: &core.Identifier{
			Project:      w.GetExecutionID().Project,
			Domain:       w.GetExecutionID().Domain,
			Version:      rand.String(10),
			Name:         rand.String(10),
			ResourceType: core.ResourceType_WORKFLOW,
		},
		Nodes:     djSpec.Nodes,
		Outputs:   djSpec.Outputs,
		Interface: iface,
	}, nil
}

// For any node that is not in a NEW/READY state in the recording, CheckNodeStatus will be invoked. The implementation should handle
// idempotency and return the current observed state of the node
func (e dynamicNodeHandler) CheckNodeStatus(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode,
	previousNodeStatus v1alpha1.ExecutableNodeStatus) (handler.Status, error) {

	var status handler.Status
	var err error
	switch previousNodeStatus.GetOrCreateDynamicNodeStatus().GetDynamicNodePhase() {
	case v1alpha1.DynamicNodePhaseExecuting:
		// If the node succeeded, check if it generated a futures.pb file to execute.
		dynamicWF, nStatus, _, err := e.buildContextualDynamicWorkflow(ctx, w, node, previousNodeStatus)
		if err != nil {
			return handler.StatusFailed(err), nil
		}

		s, err := e.progressDynamicWorkflow(ctx, nStatus, dynamicWF)
		if err == nil && s == handler.StatusSuccess {
			// After the dynamic node completes we need to copy the outputs from its end nodes to the parent nodes status
			endNode := dynamicWF.GetNodeExecutionStatus(v1alpha1.EndNodeID)
			outputPath := v1alpha1.GetOutputsFile(endNode.GetDataDir())
			destinationPath := v1alpha1.GetOutputsFile(previousNodeStatus.GetDataDir())
			logger.Infof(ctx, "Dynamic workflow completed, copying outputs from the end-node [%s] to the parent node data dir [%s]", outputPath, destinationPath)
			if err := e.store.CopyRaw(ctx, outputPath, destinationPath, storage.Options{}); err != nil {
				logger.Errorf(ctx, "Failed to copy outputs from dynamic sub-wf [%s] to [%s]. Error: %s", outputPath, destinationPath, err.Error())
				return handler.StatusUndefined, errors.Wrapf(errors.StorageError, node.GetID(), err, "Failed to copy outputs from dynamic sub-wf [%s] to [%s]. Error: %s", outputPath, destinationPath, err.Error())
			}
			if successHandler, ok := e.IFace.(handler.PostNodeSuccessHandler); ok {
				return successHandler.HandleNodeSuccess(ctx, w, node)
			}
			logger.Warnf(ctx, "Bad configuration for dynamic node, no post node success handler found!")
		}
		return s, err
	default:
		// Invoke the underlying check node status.
		status, err = e.IFace.CheckNodeStatus(ctx, w, node, previousNodeStatus)

		if err != nil {
			return status, err
		}

		if status.Phase != handler.PhaseSuccess {
			return status, err
		}

		// If the node succeeded, check if it generated a futures.pb file to execute.
		_, _, isDynamic, err := e.buildContextualDynamicWorkflow(ctx, w, node, previousNodeStatus)
		if err != nil {
			return handler.StatusFailed(err), nil
		}

		if !isDynamic {
			if successHandler, ok := e.IFace.(handler.PostNodeSuccessHandler); ok {
				return successHandler.HandleNodeSuccess(ctx, w, node)
			}
			logger.Warnf(ctx, "Bad configuration for dynamic node, no post node success handler found!")
			return status, err
		}

		// Mark the node as a dynamic node executing its child nodes. Next time check node status is called, it'll go
		// directly to progress the dynamically generated workflow.
		previousNodeStatus.GetOrCreateDynamicNodeStatus().SetDynamicNodePhase(v1alpha1.DynamicNodePhaseExecuting)

		return handler.StatusRunning, nil
	}
}

func (e dynamicNodeHandler) buildContextualDynamicWorkflow(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode,
	previousNodeStatus v1alpha1.ExecutableNodeStatus) (dynamicWf v1alpha1.ExecutableWorkflow, status v1alpha1.ExecutableNodeStatus, isDynamic bool, err error) {

	var nStatus v1alpha1.ExecutableNodeStatus
	rootNodeStatus := w.GetNodeExecutionStatus(node.GetID())
	if node.GetTaskID() != nil {
		// TODO: This is a hack to set parent task execution id, we should move to node-node relationship.
		execID, err := e.getTaskExecutionIdentifier(ctx, w, node)
		if err != nil {
			return nil, nil, false, err
		}

		dynamicNode := &v1alpha1.NodeSpec{
			ID: "dynamic-node",
		}

		nStatus = rootNodeStatus.GetNodeExecutionStatus(dynamicNode.GetID())
		nStatus.SetDataDir(rootNodeStatus.GetDataDir())
		nStatus.SetParentTaskID(execID)
	} else {
		nStatus = w.GetNodeExecutionStatus(node.GetID())
	}

	subwf, isDynamic, err := e.loadOrBuildDynamicWorkflow(ctx, w, node, previousNodeStatus.GetDataDir(), nStatus)
	if err != nil {
		return nil, nStatus, false, err
	}

	if !isDynamic {
		return nil, nil, false, nil
	}

	return newContextualWorkflow(w, subwf, nStatus, subwf.Tasks, subwf.SubWorkflows), nStatus, isDynamic, nil
}

func (e dynamicNodeHandler) buildFlyteWorkflow(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode,
	dataDir storage.DataReference, nStatus v1alpha1.ExecutableNodeStatus) (compiledWf *v1alpha1.FlyteWorkflow, isDynamic bool, err error) {
	t := e.metrics.buildDynamicWorkflow.Start(ctx)
	defer t.Stop()

	// We will only get here if the Phase is success. The downside is that this is an overhead for all nodes that are
	// not dynamic. But given that we will only check once, it should be ok.
	// TODO: Check for node.is_dynamic once the IDL changes are in and SDK migration has happened.
	djSpec, err := e.getDynamicJobSpec(ctx, node, dataDir)
	if err != nil {
		return nil, false, err
	}

	if djSpec == nil {
		return nil, false, nil
	}

	var closure *core.CompiledWorkflowClosure
	wf, err := e.buildDynamicWorkflowTemplate(ctx, djSpec, w, node, nStatus)
	if err != nil {
		return nil, true, err
	}

	compiledTasks, err := compileTasks(ctx, djSpec.Tasks)
	if err != nil {
		return nil, true, err
	}

	// TODO: This will currently fail if the WF references any launch plans
	closure, err = compiler.CompileWorkflow(wf, djSpec.Subworkflows, compiledTasks, []common2.InterfaceProvider{})
	if err != nil {
		return nil, true, err
	}

	subwf, err := k8s.BuildFlyteWorkflow(closure, &core.LiteralMap{}, nil, "")
	if err != nil {
		return nil, false, err
	}

	return subwf, true, nil
}

func (e dynamicNodeHandler) loadOrBuildDynamicWorkflow(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode,
	dataDir storage.DataReference, nodeStatus v1alpha1.ExecutableNodeStatus) (compiledWf *v1alpha1.FlyteWorkflow, isDynamic bool, err error) {

	cacheHitStopWatch := e.metrics.CacheHit.Start(ctx)
	// Check if we have compiled the workflow before:
	compiledFuturesFilePath, err := e.store.ConstructReference(ctx, nodeStatus.GetDataDir(), v1alpha1.GetCompiledFutureFile())
	if err != nil {
		logger.Warnf(ctx, "Failed to construct data path for futures file. Error: %v", err)
		return nil, false, errors.Wrapf(errors.CausedByError, node.GetID(), err, "Failed to construct data path for futures file.")
	}

	// If there is a cached compiled Workflow, load and return it.
	if metadata, err := e.store.Head(ctx, compiledFuturesFilePath); err != nil {
		logger.Warnf(ctx, "Failed to call head on compiled futures file. Error: %v", err)
		return nil, false, errors.Wrapf(errors.CausedByError, node.GetID(), err, "Failed to do HEAD on compiled futures file.")
	} else if metadata.Exists() {
		// It exists, load and return it
		compiledWf, err = loadCachedFlyteWorkflow(ctx, e.store, compiledFuturesFilePath)
		if err != nil {
			logger.Warnf(ctx, "Failed to load cached flyte workflow from [%v], this will cause the dynamic workflow to be recompiled. Error: %v",
				compiledFuturesFilePath, err)
			e.metrics.CacheError.Inc(ctx)
		} else {
			cacheHitStopWatch.Stop()
			return compiledWf, true, nil
		}
	}

	// If we have not build this spec before, build it now and cache it.
	compiledWf, isDynamic, err = e.buildFlyteWorkflow(ctx, w, node, dataDir, nodeStatus)
	if err != nil {
		return compiledWf, isDynamic, err
	}

	if !isDynamic {
		return compiledWf, isDynamic, err
	}

	// Cache the built WF. Errors are swallowed.
	err = cacheFlyteWorkflow(ctx, e.store, compiledWf, compiledFuturesFilePath)
	if err != nil {
		logger.Warnf(ctx, "Failed to cache flyte workflow, this will cause a cache miss next time and cause the dynamic workflow to be recompiled. Error: %v", err)
		e.metrics.CacheError.Inc(ctx)
	}

	return compiledWf, true, nil
}

func (e dynamicNodeHandler) progressDynamicWorkflow(ctx context.Context, parentNodeStatus v1alpha1.ExecutableNodeStatus,
	w v1alpha1.ExecutableWorkflow) (handler.Status, error) {

	state, err := e.nodeExecutor.RecursiveNodeHandler(ctx, w, w.StartNode())
	if err != nil {
		return handler.StatusUndefined, err
	}

	if state.HasFailed() {
		if w.GetOnFailureNode() != nil {
			return handler.StatusFailing(state.Err), nil
		}
		return handler.StatusFailed(state.Err), nil
	}

	if state.IsComplete() {
		nodeID := ""
		if parentNodeStatus.GetParentNodeID() != nil {
			nodeID = *parentNodeStatus.GetParentNodeID()
		}

		// If the WF interface has outputs, validate that the outputs file was written.
		if outputBindings := w.GetOutputBindings(); len(outputBindings) > 0 {
			endNodeStatus := w.GetNodeExecutionStatus(v1alpha1.EndNodeID)
			if endNodeStatus == nil {
				return handler.StatusFailed(errors.Errorf(errors.SubWorkflowExecutionFailed, nodeID,
					"No end node found in subworkflow.")), nil
			}

			sourcePath := v1alpha1.GetOutputsFile(endNodeStatus.GetDataDir())
			if metadata, err := e.store.Head(ctx, sourcePath); err == nil {
				if !metadata.Exists() {
					return handler.StatusFailed(errors.Errorf(errors.SubWorkflowExecutionFailed, nodeID,
						"Subworkflow is expected to produce outputs but no outputs file was written to %v.",
						sourcePath)), nil
				}
			} else {
				return handler.StatusUndefined, err
			}

			destinationPath := v1alpha1.GetOutputsFile(parentNodeStatus.GetDataDir())
			if err := e.store.CopyRaw(ctx, sourcePath, destinationPath, storage.Options{}); err != nil {
				return handler.StatusFailed(errors.Wrapf(errors.OutputsNotFoundError, nodeID,
					err, "Failed to copy subworkflow outputs from [%v] to [%v]",
					sourcePath, destinationPath)), nil
			}
		}

		return handler.StatusSuccess, nil
	}

	if state.PartiallyComplete() {
		// Re-enqueue the workflow
		e.enQWorkflow(w.GetK8sWorkflowID().String())
	}

	return handler.StatusRunning, nil
}

func (e dynamicNodeHandler) getTaskExecutionIdentifier(_ context.Context, w v1alpha1.ExecutableWorkflow,
	node v1alpha1.ExecutableNode) (*core.TaskExecutionIdentifier, error) {

	taskID := node.GetTaskID()
	task, err := w.GetTask(*taskID)
	if err != nil {
		return nil, errors.Wrapf(errors.BadSpecificationError, node.GetID(), err, "Unable to find task for taskId: [%v]", *taskID)
	}

	nodeStatus := w.GetNodeExecutionStatus(node.GetID())
	return &core.TaskExecutionIdentifier{
		TaskId:       task.CoreTask().Id,
		RetryAttempt: nodeStatus.GetAttempts(),
		NodeExecutionId: &core.NodeExecutionIdentifier{
			NodeId:      node.GetID(),
			ExecutionId: w.GetExecutionID().WorkflowExecutionIdentifier,
		},
	}, nil
}

func (e dynamicNodeHandler) AbortNode(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode) error {

	previousNodeStatus := w.GetNodeExecutionStatus(node.GetID())
	switch previousNodeStatus.GetOrCreateDynamicNodeStatus().GetDynamicNodePhase() {
	case v1alpha1.DynamicNodePhaseExecuting:
		dynamicWF, _, isDynamic, err := e.buildContextualDynamicWorkflow(ctx, w, node, previousNodeStatus)
		if err != nil {
			return err
		}

		if !isDynamic {
			return nil
		}

		return e.nodeExecutor.AbortHandler(ctx, dynamicWF, dynamicWF.StartNode())
	default:
		// Invoke the underlying abort node.
		return e.IFace.AbortNode(ctx, w, node)
	}
}

func New(underlying handler.IFace, nodeExecutor executors.Node, enQWorkflow v1alpha1.EnqueueWorkflow, store *storage.DataStore,
	scope promutils.Scope) handler.IFace {

	return dynamicNodeHandler{
		IFace:        underlying,
		metrics:      newMetrics(scope),
		nodeExecutor: nodeExecutor,
		enQWorkflow:  enQWorkflow,
		store:        store,
	}
}
