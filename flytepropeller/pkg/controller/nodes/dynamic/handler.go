package dynamic

import (
	"context"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/errors"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/interfaces"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/subworkflow/launchplan"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task"
	"github.com/flyteorg/flytepropeller/pkg/utils"

	stdErrors "github.com/flyteorg/flytestdlib/errors"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
)

//go:generate mockery -all -case=underscore

const dynamicNodeID = "dynamic-node"

type TaskNodeHandler interface {
	interfaces.CacheableNodeHandler
	ValidateOutput(ctx context.Context, nodeID v1alpha1.NodeID, i io.InputReader,
		r io.OutputReader, outputCommitter io.OutputWriter, executionConfig v1alpha1.ExecutionConfig,
		tr ioutils.SimpleTaskReader) (*io.ExecutionError, error)
}

type metrics struct {
	buildDynamicWorkflow   labeled.StopWatch
	retrieveDynamicJobSpec labeled.StopWatch
	CacheHit               labeled.StopWatch
	CacheError             labeled.Counter
	CacheMiss              labeled.Counter
}

func newMetrics(scope promutils.Scope) metrics {
	return metrics{
		buildDynamicWorkflow:   labeled.NewStopWatch("build_dynamic_workflow", "Overhead for building a dynamic workflow in memory.", time.Microsecond, scope),
		retrieveDynamicJobSpec: labeled.NewStopWatch("retrieve_dynamic_spec", "Overhead of downloading and un-marshaling dynamic job spec", time.Microsecond, scope),
		CacheHit:               labeled.NewStopWatch("dynamic_workflow_cache_hit", "A dynamic workflow was loaded from store.", time.Microsecond, scope),
		CacheError:             labeled.NewCounter("cache_err", "A dynamic workflow failed to store or load from data store.", scope),
		CacheMiss:              labeled.NewCounter("cache_miss", "A dynamic workflow did not already exist in the data store.", scope),
	}
}

type dynamicNodeTaskNodeHandler struct {
	TaskNodeHandler
	metrics      metrics
	nodeExecutor interfaces.Node
	lpReader     launchplan.Reader
	eventConfig  *config.EventConfig
}

func (d dynamicNodeTaskNodeHandler) handleParentNode(ctx context.Context, prevState handler.DynamicNodeState, nCtx interfaces.NodeExecutionContext) (handler.Transition, handler.DynamicNodeState, error) {
	// It seems parent node is still running, lets call handle for parent node
	trns, err := d.TaskNodeHandler.Handle(ctx, nCtx)
	if err != nil {
		return trns, prevState, err
	}

	if trns.Info().GetPhase() == handler.EPhaseSuccess {
		f, err := task.NewRemoteFutureFileReader(ctx, nCtx.NodeStatus().GetOutputDir(), nCtx.DataStore())
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
			// directly to record, and then progress the dynamically generated workflow.
			logger.Infof(ctx, "future file detected, assuming dynamic node")
			// There is a futures file, so we need to continue running the node with the modified state
			return trns.WithInfo(handler.PhaseInfoRunning(trns.Info().GetInfo())), handler.DynamicNodeState{Phase: v1alpha1.DynamicNodePhaseParentFinalizing}, nil
		}
	}

	logger.Infof(ctx, "regular node detected, (no future file found)")
	return trns, prevState, nil
}

func (d dynamicNodeTaskNodeHandler) produceDynamicWorkflow(ctx context.Context, nCtx interfaces.NodeExecutionContext) (
	handler.Transition, handler.DynamicNodeState, error) {
	// The first time this is called we go ahead and evaluate the dynamic node to build the workflow. We then cache
	// this workflow definition and send it to be persisted by flyteadmin so that users can observe the structure.
	dCtx, err := d.buildContextualDynamicWorkflow(ctx, nCtx)
	if err != nil {
		if stdErrors.IsCausedBy(err, utils.ErrorCodeUser) {
			return handler.DoTransition(handler.TransitionTypeEphemeral,
				handler.PhaseInfoFailure(core.ExecutionError_USER, "DynamicWorkflowBuildFailed", err.Error(), nil),
			), handler.DynamicNodeState{Phase: v1alpha1.DynamicNodePhaseFailing, Reason: err.Error()}, nil
		}
		return handler.Transition{}, handler.DynamicNodeState{}, err
	}
	taskNodeInfoMetadata := &event.TaskNodeMetadata{}
	if dCtx.subWorkflowClosure != nil && dCtx.subWorkflowClosure.Primary != nil && dCtx.subWorkflowClosure.Primary.Template != nil {
		taskNodeInfoMetadata.DynamicWorkflow = &event.DynamicWorkflowNodeMetadata{
			Id:                dCtx.subWorkflowClosure.Primary.Template.Id,
			CompiledWorkflow:  dCtx.subWorkflowClosure,
			DynamicJobSpecUri: dCtx.dynamicJobSpecURI,
		}
	}

	nextState := handler.DynamicNodeState{Phase: v1alpha1.DynamicNodePhaseExecuting}
	return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoDynamicRunning(&handler.ExecutionInfo{
		TaskNodeInfo: &handler.TaskNodeInfo{
			TaskNodeMetadata: taskNodeInfoMetadata,
		},
	})), nextState, nil
}

func (d dynamicNodeTaskNodeHandler) handleDynamicSubNodes(ctx context.Context, nCtx interfaces.NodeExecutionContext, prevState handler.DynamicNodeState) (handler.Transition, handler.DynamicNodeState, error) {
	dCtx, err := d.buildContextualDynamicWorkflow(ctx, nCtx)
	if err != nil {
		if stdErrors.IsCausedBy(err, utils.ErrorCodeUser) {
			return handler.DoTransition(handler.TransitionTypeEphemeral,
				handler.PhaseInfoFailure(core.ExecutionError_USER, "DynamicWorkflowBuildFailed", err.Error(), nil),
			), handler.DynamicNodeState{Phase: v1alpha1.DynamicNodePhaseFailing, Reason: err.Error()}, nil
		}
		// Mostly a system error or unknown
		return handler.Transition{}, handler.DynamicNodeState{}, err
	}

	trns, newState, err := d.progressDynamicWorkflow(ctx, dCtx.execContext, dCtx.subWorkflow, dCtx.nodeLookup, nCtx, prevState)
	if err != nil {
		return handler.UnknownTransition, prevState, err
	}

	if trns.Info().GetPhase() == handler.EPhaseSuccess {
		logger.Infof(ctx, "dynamic workflow node has succeeded, will call on success handler for parent node [%s]", nCtx.NodeID())
		// These outputPaths only reads the output metadata. So the sandbox is completely optional here and hence it is nil.
		// The sandbox creation as it uses hashing can be expensive and we skip that expense.
		outputPaths := ioutils.NewReadOnlyOutputFilePaths(ctx, nCtx.DataStore(), nCtx.NodeStatus().GetOutputDir())
		outputReader := ioutils.NewRemoteFileOutputReader(ctx, nCtx.DataStore(), outputPaths, nCtx.MaxDatasetSizeBytes())
		ee, err := d.TaskNodeHandler.ValidateOutput(ctx, nCtx.NodeID(), nCtx.InputReader(),
			outputReader, nil, nCtx.ExecutionContext().GetExecutionConfig(), nCtx.TaskReader())

		if err != nil {
			return handler.UnknownTransition, prevState, err
		}

		if ee != nil {
			if ee.IsRecoverable {
				return trns.WithInfo(handler.PhaseInfoRetryableFailureErr(ee.ExecutionError, trns.Info().GetInfo())), handler.DynamicNodeState{Phase: v1alpha1.DynamicNodePhaseFailing, Reason: ee.ExecutionError.String()}, nil
			}

			return trns.WithInfo(handler.PhaseInfoFailureErr(ee.ExecutionError, trns.Info().GetInfo())), handler.DynamicNodeState{Phase: v1alpha1.DynamicNodePhaseFailing, Reason: ee.ExecutionError.String()}, nil
		}

		trns = trns.WithInfo(trns.Info().WithInfo(&handler.ExecutionInfo{
			OutputInfo: trns.Info().GetInfo().OutputInfo,
		}))
	}

	return trns, newState, nil
}

// Handle method is the entry point for handling a node, which may or maynot be a Dynamic node
// The State machine for a dynamic node is as follows
// DynamicNodePhaseNone: The parent node is being handled
// DynamicNodePhaseParentFinalizing: The parent node has completes successfully and sub-nodes exist (futures file found). Parent node is being finalized.
// DynamicNodePhaseParentFinalized: The parent has node completed successfully and the generated dynamic sub workflow has been serialized and sent as an event.
// DynamicNodePhaseExecuting: The parent node has completed and finalized successfully, the sub-nodes are being handled
// DynamicNodePhaseFailing: one or more of sub-nodes have failed and the failure is being handled
func (d dynamicNodeTaskNodeHandler) Handle(ctx context.Context, nCtx interfaces.NodeExecutionContext) (handler.Transition, error) {
	ds := nCtx.NodeStateReader().GetDynamicNodeState()
	var err error
	var trns handler.Transition
	newState := ds
	logger.Infof(ctx, "Dynamic handler.Handle's called with phase %v.", ds.Phase)
	switch ds.Phase {
	case v1alpha1.DynamicNodePhaseExecuting:
		trns, newState, err = d.handleDynamicSubNodes(ctx, nCtx, ds)
		if err != nil {
			logger.Errorf(ctx, "handling dynamic subnodes failed with error: %s", err.Error())
			return trns, err
		}
	case v1alpha1.DynamicNodePhaseFailing:
		err = d.Abort(ctx, nCtx, ds.Reason)
		if err != nil {
			logger.Errorf(ctx, "Failing to abort dynamic workflow")
			return trns, err
		}

		// if DynamicNodeStatus is noted with permanent failures we report a non-recoverable failure
		phaseInfoFailureFunc := handler.PhaseInfoRetryableFailure
		phaseInfoFailureFuncErr := handler.PhaseInfoRetryableFailureErr
		if ds.IsFailurePermanent {
			phaseInfoFailureFunc = handler.PhaseInfoFailure
			phaseInfoFailureFuncErr = handler.PhaseInfoFailureErr
		}

		if ds.Error != nil {
			trns = handler.DoTransition(handler.TransitionTypeEphemeral, phaseInfoFailureFuncErr(ds.Error, nil))
		} else {
			trns = handler.DoTransition(handler.TransitionTypeEphemeral, phaseInfoFailureFunc(core.ExecutionError_UNKNOWN, "DynamicNodeFailing", ds.Reason, nil))
		}
	case v1alpha1.DynamicNodePhaseParentFinalizing:
		if err := d.finalizeParentNode(ctx, nCtx); err != nil {
			return handler.UnknownTransition, err
		}
		newState = handler.DynamicNodeState{Phase: v1alpha1.DynamicNodePhaseParentFinalized}
		trns = handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRunning(trns.Info().GetInfo()))
	case v1alpha1.DynamicNodePhaseParentFinalized:
		trns, newState, err = d.produceDynamicWorkflow(ctx, nCtx)
		if err != nil {
			logger.Errorf(ctx, "handling producing dynamic workflow definition failed with error: %s", err.Error())
			return trns, err
		}
	default:
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

func (d dynamicNodeTaskNodeHandler) Abort(ctx context.Context, nCtx interfaces.NodeExecutionContext, reason string) error {
	ds := nCtx.NodeStateReader().GetDynamicNodeState()
	switch ds.Phase {
	case v1alpha1.DynamicNodePhaseFailing:
		fallthrough
	case v1alpha1.DynamicNodePhaseExecuting:
		logger.Infof(ctx, "Aborting dynamic workflow at RetryAttempt [%d]", nCtx.CurrentAttempt())
		dCtx, err := d.buildContextualDynamicWorkflow(ctx, nCtx)
		if err != nil {
			if stdErrors.IsCausedBy(err, utils.ErrorCodeUser) {
				logger.Errorf(ctx, "failed to build dynamic workflow, user error: %s", err)
			}
			return err
		}

		if !dCtx.isDynamic {
			return nil
		}

		return d.nodeExecutor.AbortHandler(ctx, dCtx.execContext, dCtx.subWorkflow, dCtx.nodeLookup, dCtx.subWorkflow.StartNode(), reason)
	default:
		logger.Infof(ctx, "Aborting regular node RetryAttempt [%d]", nCtx.CurrentAttempt())
		// The parent node has not yet completed, so we will abort the parent node
		return d.TaskNodeHandler.Abort(ctx, nCtx, reason)
	}
}

func (d dynamicNodeTaskNodeHandler) finalizeParentNode(ctx context.Context, nCtx interfaces.NodeExecutionContext) error {
	logger.Infof(ctx, "Finalizing Parent node RetryAttempt [%d]", nCtx.CurrentAttempt())
	if err := d.TaskNodeHandler.Finalize(ctx, nCtx); err != nil {
		logger.Errorf(ctx, "Failed to finalize Dynamic Nodes Parent.")
		return err
	}
	return nil
}

// This is a weird method. We should always finalize before we set the dynamic parent node phase as complete?
func (d dynamicNodeTaskNodeHandler) Finalize(ctx context.Context, nCtx interfaces.NodeExecutionContext) error {
	errs := make([]error, 0, 2)

	ds := nCtx.NodeStateReader().GetDynamicNodeState()
	if ds.Phase == v1alpha1.DynamicNodePhaseFailing || ds.Phase == v1alpha1.DynamicNodePhaseExecuting {
		logger.Infof(ctx, "Finalizing dynamic workflow RetryAttempt [%d]", nCtx.CurrentAttempt())
		dCtx, err := d.buildContextualDynamicWorkflow(ctx, nCtx)
		if err != nil {
			errs = append(errs, err)
		} else {
			if dCtx.isDynamic {
				if err := d.nodeExecutor.FinalizeHandler(ctx, dCtx.execContext, dCtx.subWorkflow, dCtx.nodeLookup, dCtx.subWorkflow.StartNode()); err != nil {
					logger.Errorf(ctx, "failed to finalize dynamic workflow, err: %s", err)
					errs = append(errs, err)
				}
			}
		}
	}

	// We should always finalize the parent node success or failure.
	// If we use the phase to decide when to finalize in the case where Dynamic node is in phase Executiing
	// (i.e. child nodes are now being executed) and Finalize is invoked, we will never invoke the finalizer for the parent.
	if err := d.finalizeParentNode(ctx, nCtx); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.ErrorCollection{Errors: errs}
	}

	return nil
}

func New(underlying TaskNodeHandler, nodeExecutor interfaces.Node, launchPlanReader launchplan.Reader, eventConfig *config.EventConfig, scope promutils.Scope) interfaces.NodeHandler {

	return &dynamicNodeTaskNodeHandler{
		TaskNodeHandler: underlying,
		metrics:         newMetrics(scope),
		nodeExecutor:    nodeExecutor,
		lpReader:        launchPlanReader,
		eventConfig:     eventConfig,
	}
}
