package dynamic

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/semver"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/catalog"
	pluginCore "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/promutils/labeled"

	"github.com/lyft/flytepropeller/pkg/compiler"
	common2 "github.com/lyft/flytepropeller/pkg/compiler/common"
	"github.com/lyft/flytepropeller/pkg/controller/executors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/task"

	"github.com/lyft/flytestdlib/promutils"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytestdlib/storage"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/compiler/transformers/k8s"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/errors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
)

//go:generate mockery -all -case=underscore

const dynamicNodeID = "dynamic-node"

// TODO: Remove after deploying version 0.2.4
var arrayFlyteKitVersion = semver.MustParse("0.2.3")

type TaskNodeHandler interface {
	handler.Node
	ValidateOutputAndCacheAdd(ctx context.Context, nodeID v1alpha1.NodeID, i io.InputReader, r io.OutputReader, outputCommitter io.OutputWriter,
		tr pluginCore.TaskReader, m catalog.Metadata) (*io.ExecutionError, error)
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
		retrieveDynamicJobSpec: labeled.NewStopWatch("retrieve_dynamic_spec", "Overhead of downloading and un-marshaling dynamic job spec", time.Microsecond, scope),
		CacheHit:               labeled.NewStopWatch("dynamic_workflow_cache_hit", "A dynamic workflow was loaded from store.", time.Microsecond, scope),
		CacheError:             labeled.NewCounter("cache_err", "A dynamic workflow failed to store or load from data store.", scope),
	}
}

type dynamicNodeTaskNodeHandler struct {
	TaskNodeHandler
	metrics      metrics
	nodeExecutor executors.Node
}

func (d dynamicNodeTaskNodeHandler) handleParentNode(ctx context.Context, prevState handler.DynamicNodeState, nCtx handler.NodeExecutionContext) (handler.Transition, handler.DynamicNodeState, error) {
	// It seems parent node is still running, lets call handle for parent node
	trns, err := d.TaskNodeHandler.Handle(ctx, nCtx)
	if err != nil {
		return trns, prevState, err
	}

	if trns.Info().GetPhase() == handler.EPhaseSuccess {
		f, err := task.NewRemoteFutureFileReader(ctx, nCtx.NodeStatus().GetDataDir(), nCtx.DataStore())
		if err != nil {
			return handler.UnknownTransition, prevState, err
		}
		// If the node succeeded, check if it generated a futures.pb file to execute.
		ok, err := f.Exists(ctx)
		if err != nil {
			return handler.UnknownTransition, prevState, err
		}
		if ok {
			// Mark the node that parent node has completed and a dynamic node executing its child nodes. Next time check node status is called, it'll go
			// directly to progress the dynamically generated workflow.
			logger.Infof(ctx, "future file detected, assuming dynamic node")
			// There is a futures file, so we need to continue running the node with the modified state
			return trns.WithInfo(handler.PhaseInfoRunning(trns.Info().GetInfo())), handler.DynamicNodeState{Phase: v1alpha1.DynamicNodePhaseExecuting}, nil
		}
	}

	logger.Infof(ctx, "regular node detected, (no future file found)")
	return trns, prevState, nil
}

func (d dynamicNodeTaskNodeHandler) handleDynamicSubNodes(ctx context.Context, nCtx handler.NodeExecutionContext, prevState handler.DynamicNodeState) (handler.Transition, handler.DynamicNodeState, error) {
	dynamicWF, _, err := d.buildContextualDynamicWorkflow(ctx, nCtx)
	if err != nil {
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(
			"DynamicWorkflowBuildFailed", err.Error(), nil)), handler.DynamicNodeState{Phase: v1alpha1.DynamicNodePhaseFailing, Reason: err.Error()}, nil
	}

	trns, newState, err := d.progressDynamicWorkflow(ctx, dynamicWF, nCtx, prevState)
	if err != nil {
		return handler.UnknownTransition, prevState, err
	}

	if trns.Info().GetPhase() == handler.EPhaseSuccess {
		logger.Infof(ctx, "dynamic workflow node has succeeded, will call on success handler for parent node [%s]", nCtx.NodeID())
		outputPaths := ioutils.NewRemoteFileOutputPaths(ctx, nCtx.DataStore(), nCtx.NodeStatus().GetDataDir())
		execID := task.GetTaskExecutionIdentifier(nCtx)
		outputReader := ioutils.NewRemoteFileOutputReader(ctx, nCtx.DataStore(), outputPaths, nCtx.MaxDatasetSizeBytes())
		ee, err := d.TaskNodeHandler.ValidateOutputAndCacheAdd(ctx, nCtx.NodeID(), nCtx.InputReader(), outputReader, nil, nCtx.TaskReader(), catalog.Metadata{
			TaskExecutionIdentifier: execID,
		})

		if err != nil {
			return handler.UnknownTransition, prevState, err
		}

		if ee != nil {
			if ee.IsRecoverable {
				return trns.WithInfo(handler.PhaseInfoRetryableFailureErr(ee.ExecutionError, trns.Info().GetInfo())), handler.DynamicNodeState{Phase: v1alpha1.DynamicNodePhaseFailing, Reason: ee.ExecutionError.String()}, nil
			}

			return trns.WithInfo(handler.PhaseInfoFailureErr(ee.ExecutionError, trns.Info().GetInfo())), handler.DynamicNodeState{Phase: v1alpha1.DynamicNodePhaseFailing, Reason: ee.ExecutionError.String()}, nil
		}
	}

	return trns, newState, nil
}

func (d dynamicNodeTaskNodeHandler) Handle(ctx context.Context, nCtx handler.NodeExecutionContext) (handler.Transition, error) {
	ds := nCtx.NodeStateReader().GetDynamicNodeState()
	var err error
	var trns handler.Transition
	newState := ds
	logger.Infof(ctx, "Dynamic handler.Handle's called with phase %v.", ds.Phase)
	if ds.Phase == v1alpha1.DynamicNodePhaseExecuting {
		trns, newState, err = d.handleDynamicSubNodes(ctx, nCtx, ds)
		if err != nil {
			logger.Errorf(ctx, "handling dynamic subnodes failed with error: %s", err.Error())
			return trns, err
		}
	} else if ds.Phase == v1alpha1.DynamicNodePhaseFailing {
		err = d.Abort(ctx, nCtx, ds.Reason)
		if err != nil {
			logger.Errorf(ctx, "Failing to abort dynamic workflow")
			return trns, err
		}

		trns = handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRetryableFailure("DynamicNodeFailed", ds.Reason, nil))
	} else {
		trns, newState, err = d.handleParentNode(ctx, ds, nCtx)
		if err != nil {
			logger.Errorf(ctx, "handling parent node failed with error: %s", err.Error())
			return trns, err
		}
	}

	if err := nCtx.NodeStateWriter().PutDynamicNodeState(newState); err != nil {
		return handler.UnknownTransition, err
	}

	return trns, nil
}

func (d dynamicNodeTaskNodeHandler) Abort(ctx context.Context, nCtx handler.NodeExecutionContext, reason string) error {
	ds := nCtx.NodeStateReader().GetDynamicNodeState()
	switch ds.Phase {
	case v1alpha1.DynamicNodePhaseFailing:
		fallthrough
	case v1alpha1.DynamicNodePhaseExecuting:
		dynamicWF, isDynamic, err := d.buildContextualDynamicWorkflow(ctx, nCtx)
		if err != nil {
			return err
		}

		if !isDynamic {
			return nil
		}

		return d.nodeExecutor.AbortHandler(ctx, dynamicWF, dynamicWF.StartNode(), reason)
	default:
		// The parent node has not yet completed, so we will abort the parent node
		return d.TaskNodeHandler.Abort(ctx, nCtx, reason)
	}
}

// This is a weird method. We should always finalize before we set the dynamic parent node phase as complete?
func (d dynamicNodeTaskNodeHandler) Finalize(ctx context.Context, nCtx handler.NodeExecutionContext) error {
	ds := nCtx.NodeStateReader().GetDynamicNodeState()
	switch ds.Phase {
	case v1alpha1.DynamicNodePhaseFailing:
		fallthrough
	case v1alpha1.DynamicNodePhaseExecuting:
		dynamicWF, isDynamic, err := d.buildContextualDynamicWorkflow(ctx, nCtx)
		if err != nil {
			return err
		}

		if !isDynamic {
			return nil
		}

		return d.nodeExecutor.FinalizeHandler(ctx, dynamicWF, dynamicWF.StartNode())
	default:
		return d.TaskNodeHandler.Finalize(ctx, nCtx)
	}
}

func (d dynamicNodeTaskNodeHandler) buildDynamicWorkflowTemplate(ctx context.Context, djSpec *core.DynamicJobSpec, nCtx handler.NodeExecutionContext, parentNodeStatus v1alpha1.ExecutableNodeStatus) (
	*core.WorkflowTemplate, error) {

	iface, err := underlyingInterface(ctx, nCtx.TaskReader())
	if err != nil {
		return nil, err
	}

	// Modify node IDs to include lineage, the entire system assumes node IDs are unique per parent WF.
	// We keep track of the original node ids because that's where inputs are written to.
	parentNodeID := nCtx.NodeID()
	for _, n := range djSpec.Nodes {
		newID, err := hierarchicalNodeID(parentNodeID, n.Id)
		if err != nil {
			return nil, err
		}

		// Instantiate a nodeStatus using the modified name but set its data directory using the original name.
		subNodeStatus := parentNodeStatus.GetNodeExecutionStatus(newID)
		originalNodePath, err := nCtx.DataStore().ConstructReference(ctx, nCtx.NodeStatus().GetDataDir(), n.Id)
		if err != nil {
			return nil, err
		}

		subNodeStatus.SetDataDir(originalNodePath)
		subNodeStatus.ResetDirty()

		n.Id = newID
	}

	if nCtx.TaskReader().GetTaskID() != nil {
		// If the parent is a task, pass down data children nodes should inherit.
		parentTask, err := nCtx.TaskReader().Read(ctx)
		if err != nil {
			return nil, errors.Wrapf(errors.CausedByError, nCtx.NodeID(), err, "Failed to find task [%v].", nCtx.TaskReader().GetTaskID())
		}

		for _, t := range djSpec.Tasks {
			if t.GetContainer() != nil && parentTask.GetContainer() != nil {
				t.GetContainer().Config = append(t.GetContainer().Config, parentTask.GetContainer().Config...)
			}

			// TODO: This is a hack since array tasks' interfaces are malformed. Remove after
			// FlyteKit version that generates the right interfaces is deployed.
			if t.Type == "container_array" {
				isBelow, err := isFlyteKitVersionBelow(t.GetMetadata().Runtime, arrayFlyteKitVersion)
				if err != nil {
					logger.Warnf(ctx, "Failed to validate flytekit version of task [%v] will skip array"+
						" interface manipulation.", t.GetId())
					continue
				}

				if isBelow {
					iface := t.GetInterface()
					iface.Outputs = makeArrayInterface(iface.Outputs)
				}
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
			Project:      nCtx.NodeExecutionMetadata().GetExecutionID().Project,
			Domain:       nCtx.NodeExecutionMetadata().GetExecutionID().Domain,
			Version:      rand.String(10),
			Name:         rand.String(10),
			ResourceType: core.ResourceType_WORKFLOW,
		},
		Nodes:     djSpec.Nodes,
		Outputs:   djSpec.Outputs,
		Interface: iface,
	}, nil
}

func (d dynamicNodeTaskNodeHandler) buildContextualDynamicWorkflow(ctx context.Context, nCtx handler.NodeExecutionContext) (dynamicWf v1alpha1.ExecutableWorkflow, isDynamic bool, err error) {

	t := d.metrics.buildDynamicWorkflow.Start(ctx)
	defer t.Stop()

	f, err := task.NewRemoteFutureFileReader(ctx, nCtx.NodeStatus().GetDataDir(), nCtx.DataStore())
	if err != nil {
		return nil, false, err
	}

	// TODO: This is a hack to set parent task execution id, we should move to node-node relationship.
	execID := task.GetTaskExecutionIdentifier(nCtx)
	nStatus := nCtx.NodeStatus().GetNodeExecutionStatus(dynamicNodeID)
	nStatus.SetDataDir(nCtx.NodeStatus().GetDataDir())
	nStatus.SetParentTaskID(execID)

	// cacheHitStopWatch := d.metrics.CacheHit.Start(ctx)
	// Check if we have compiled the workflow before:
	// If there is a cached compiled Workflow, load and return it.
	// if ok, err := f.CacheExists(ctx); err != nil {
	//	logger.Warnf(ctx, "Failed to call head on compiled futures file. Error: %v", err)
	//	return nil, false, errors.Wrapf(errors.CausedByError, nCtx.NodeID(), err, "Failed to do HEAD on compiled futures file.")
	// } else if ok {
	//	// It exists, load and return it
	//	compiledWf, err := f.RetrieveCache(ctx)
	//	if err != nil {
	//		logger.Warnf(ctx, "Failed to load cached flyte workflow , this will cause the dynamic workflow to be recompiled. Error: %v", err)
	//		d.metrics.CacheError.Inc(ctx)
	//	} else {
	//		cacheHitStopWatch.Stop()
	//		return newContextualWorkflow(nCtx.Workflow(), compiledWf, nStatus, compiledWf.Tasks, compiledWf.SubWorkflows), true, nil
	//	}
	// }

	// We know for sure that futures file was generated. Lets read it
	djSpec, err := f.Read(ctx)
	if err != nil {
		return nil, false, errors.Wrapf(errors.RuntimeExecutionError, nCtx.NodeID(), err, "unable to read futures file, maybe corrupted")
	}

	var closure *core.CompiledWorkflowClosure
	wf, err := d.buildDynamicWorkflowTemplate(ctx, djSpec, nCtx, nStatus)
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
		return nil, true, err
	}

	if err := f.Cache(ctx, subwf); err != nil {
		logger.Errorf(ctx, "Failed to cache Dynamic workflow [%s]", err.Error())
	}

	return newContextualWorkflow(nCtx.Workflow(), subwf, nStatus, subwf.Tasks, subwf.SubWorkflows), true, nil
}

func (d dynamicNodeTaskNodeHandler) progressDynamicWorkflow(ctx context.Context, dynamicWorkflow v1alpha1.ExecutableWorkflow,
	nCtx handler.NodeExecutionContext, prevState handler.DynamicNodeState) (handler.Transition, handler.DynamicNodeState, error) {

	state, err := d.nodeExecutor.RecursiveNodeHandler(ctx, dynamicWorkflow, dynamicWorkflow.StartNode())
	if err != nil {
		return handler.UnknownTransition, prevState, err
	}

	if state.HasFailed() {
		if dynamicWorkflow.GetOnFailureNode() != nil {
			// TODO Once we migrate to closure node we need to handle subworkflow using the subworkflow handler
			logger.Errorf(ctx, "We do not support failure nodes in dynamic workflow today")
		}

		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRunning(nil)),
			handler.DynamicNodeState{Phase: v1alpha1.DynamicNodePhaseFailing, Reason: state.Err.Error()},
			nil
	}

	if state.IsComplete() {
		var o *handler.OutputInfo
		// If the WF interface has outputs, validate that the outputs file was written.
		if outputBindings := dynamicWorkflow.GetOutputBindings(); len(outputBindings) > 0 {
			endNodeStatus := dynamicWorkflow.GetNodeExecutionStatus(v1alpha1.EndNodeID)
			if endNodeStatus == nil {
				return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure("MalformedDynamicWorkflow", "no end-node found in dynamic workflow", nil)),
					handler.DynamicNodeState{Phase: v1alpha1.DynamicNodePhaseFailing, Reason: "no end-node found in dynamic workflow"},
					nil
			}

			sourcePath := v1alpha1.GetOutputsFile(endNodeStatus.GetDataDir())
			if metadata, err := nCtx.DataStore().Head(ctx, sourcePath); err == nil {
				if !metadata.Exists() {
					return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRetryableFailure("DynamicWorkflowOutputsNotFound", " is expected to produce outputs but no outputs file was written to %v.", nil)),
						handler.DynamicNodeState{Phase: v1alpha1.DynamicNodePhaseFailing, Reason: "DynamicWorkflow is expected to produce outputs but no outputs file was written"},
						nil
				}
			} else {
				return handler.UnknownTransition, prevState, err
			}

			destinationPath := v1alpha1.GetOutputsFile(nCtx.NodeStatus().GetDataDir())
			if err := nCtx.DataStore().CopyRaw(ctx, sourcePath, destinationPath, storage.Options{}); err != nil {
				return handler.DoTransition(handler.TransitionTypeEphemeral,
						handler.PhaseInfoFailure(errors.OutputsNotFoundError,
							fmt.Sprintf("Failed to copy subworkflow outputs from [%v] to [%v]", sourcePath, destinationPath), nil),
					), handler.DynamicNodeState{Phase: v1alpha1.DynamicNodePhaseFailing, Reason: "Failed to copy subworkflow outputs"},
					nil
			}
			o = &handler.OutputInfo{OutputURI: destinationPath}
		}

		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoSuccess(&handler.ExecutionInfo{
			OutputInfo: o,
		})), prevState, nil
	}

	if state.PartiallyComplete() {
		if err := nCtx.EnqueueOwnerFunc()(); err != nil {
			return handler.UnknownTransition, prevState, err
		}
	}

	return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRunning(nil)), prevState, nil
}

func New(underlying TaskNodeHandler, nodeExecutor executors.Node, scope promutils.Scope) handler.Node {

	return &dynamicNodeTaskNodeHandler{
		TaskNodeHandler: underlying,
		metrics:         newMetrics(scope),
		nodeExecutor:    nodeExecutor,
	}
}
