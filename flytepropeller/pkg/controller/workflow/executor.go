package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"

	"github.com/flyteorg/flytepropeller/events"
	eventsErr "github.com/flyteorg/flytepropeller/events/errors"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/interfaces"
	"github.com/flyteorg/flytepropeller/pkg/controller/workflow/errors"
	"github.com/flyteorg/flytepropeller/pkg/utils"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flytestdlib/storage"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

type workflowMetrics struct {
	AcceptedWorkflows         labeled.Counter
	FailureDuration           labeled.StopWatch
	SuccessDuration           labeled.StopWatch
	IncompleteWorkflowAborted labeled.Counter

	// Measures the time between when we receive service call to create an execution and when it has moved to running state.
	AcceptanceLatency labeled.StopWatch
	// Measures the time between when the WF moved to succeeding/failing state and when it finally moved to a terminal state.
	CompletionLatency labeled.StopWatch
}

type Status struct {
	TransitionToPhase v1alpha1.WorkflowPhase
	Err               *core.ExecutionError
}

var StatusReady = Status{TransitionToPhase: v1alpha1.WorkflowPhaseReady}
var StatusRunning = Status{TransitionToPhase: v1alpha1.WorkflowPhaseRunning}
var StatusSucceeding = Status{TransitionToPhase: v1alpha1.WorkflowPhaseSucceeding}
var StatusSuccess = Status{TransitionToPhase: v1alpha1.WorkflowPhaseSuccess}

func StatusFailureNode(originalErr *core.ExecutionError) Status {
	return Status{TransitionToPhase: v1alpha1.WorkflowPhaseHandlingFailureNode, Err: originalErr}
}

func StatusFailing(err *core.ExecutionError) Status {
	return Status{TransitionToPhase: v1alpha1.WorkflowPhaseFailing, Err: err}
}

func StatusFailed(err *core.ExecutionError) Status {
	return Status{TransitionToPhase: v1alpha1.WorkflowPhaseFailed, Err: err}
}

type workflowExecutor struct {
	enqueueWorkflow v1alpha1.EnqueueWorkflow
	store           *storage.DataStore
	wfRecorder      events.WorkflowEventRecorder
	k8sRecorder     record.EventRecorder
	metadataPrefix  storage.DataReference
	nodeExecutor    interfaces.Node
	metrics         *workflowMetrics
	eventConfig     *config.EventConfig
	clusterID       string
}

func (c *workflowExecutor) constructWorkflowMetadataPrefix(ctx context.Context, w *v1alpha1.FlyteWorkflow) (storage.DataReference, error) {
	if w.GetExecutionID().WorkflowExecutionIdentifier != nil {
		execID := fmt.Sprintf("%v-%v-%v", w.GetExecutionID().GetProject(), w.GetExecutionID().GetDomain(), w.GetExecutionID().GetName())
		return c.store.ConstructReference(ctx, c.metadataPrefix, execID)
	}
	// TODO should we use a random guid as the prefix? Otherwise we may get collisions
	logger.Warningf(ctx, "Workflow has no ExecutionID. Using the name as the storage-prefix. This maybe unsafe!")
	return c.store.ConstructReference(ctx, c.metadataPrefix, w.Name)
}

func (c *workflowExecutor) handleReadyWorkflow(ctx context.Context, w *v1alpha1.FlyteWorkflow) (Status, error) {

	startNode := w.StartNode()
	if startNode == nil {
		return StatusFailing(&core.ExecutionError{
			Kind:    core.ExecutionError_SYSTEM,
			Code:    errors.BadSpecificationError.String(),
			Message: "StartNode not found."}), nil
	}

	ref, err := c.constructWorkflowMetadataPrefix(ctx, w)
	if err != nil {
		return StatusFailing(&core.ExecutionError{
			Kind:    core.ExecutionError_SYSTEM,
			Code:    "MetadataPrefixCreationFailure",
			Message: err.Error()}), nil
	}
	w.GetExecutionStatus().SetDataDir(ref)
	var inputs *core.LiteralMap
	if w.Inputs != nil {
		inputs = w.Inputs.LiteralMap
	}
	// Before starting the subworkflow, lets set the inputs for the Workflow. The inputs for a SubWorkflow are essentially
	// Copy of the inputs to the Node
	nodeStatus := w.GetNodeExecutionStatus(ctx, startNode.GetID())
	dataDir, err := c.store.ConstructReference(ctx, ref, startNode.GetID(), "data")
	if err != nil {
		return StatusFailing(&core.ExecutionError{
			Kind:    core.ExecutionError_SYSTEM,
			Code:    "MetadataPrefixCreationFailure",
			Message: err.Error()}), nil
	}
	outputDir, err := c.store.ConstructReference(ctx, dataDir, "0")
	if err != nil {
		return StatusFailing(&core.ExecutionError{
			Kind:    core.ExecutionError_SYSTEM,
			Code:    "MetadataPrefixCreationFailure",
			Message: err.Error()}), nil
	}

	logger.Infof(ctx, "Setting the MetadataDir for StartNode [%v]", dataDir)
	nodeStatus.SetDataDir(dataDir)
	nodeStatus.SetOutputDir(outputDir)
	execcontext := executors.NewExecutionContext(w, w, w, nil, executors.InitializeControlFlow())
	s, err := c.nodeExecutor.SetInputsForStartNode(ctx, execcontext, w, executors.NewNodeLookup(w, w.GetExecutionStatus(), w), inputs)
	if err != nil {
		return StatusReady, err
	}

	if s.HasFailed() {
		return StatusFailing(s.Err), nil
	}
	return StatusRunning, nil
}

func (c *workflowExecutor) handleRunningWorkflow(ctx context.Context, w *v1alpha1.FlyteWorkflow) (Status, error) {
	startNode := w.StartNode()
	if startNode == nil {
		return StatusFailing(&core.ExecutionError{
			Kind:    core.ExecutionError_SYSTEM,
			Code:    errors.IllegalStateError.String(),
			Message: "Start node not found"}), nil
	}
	execcontext := executors.NewExecutionContext(w, w, w, nil, executors.InitializeControlFlow())
	state, err := c.nodeExecutor.RecursiveNodeHandler(ctx, execcontext, w, w, startNode)
	if err != nil {
		return StatusRunning, err
	}
	if state.HasFailed() {
		logger.Infof(ctx, "Workflow has failed. Error [%s]", state.Err.String())
		return StatusFailing(state.Err), nil
	}
	if state.HasTimedOut() {
		return StatusFailing(&core.ExecutionError{
			Kind:    core.ExecutionError_USER,
			Code:    "Timeout",
			Message: "Timeout in node"}), nil
	}
	if state.IsComplete() {
		return StatusSucceeding, nil
	}
	if state.PartiallyComplete() {
		c.enqueueWorkflow(w.GetK8sWorkflowID().String())
	}
	return StatusRunning, nil
}

func (c *workflowExecutor) handleFailureNode(ctx context.Context, w *v1alpha1.FlyteWorkflow) (Status, error) {
	execErr := executionErrorOrDefault(w.GetExecutionStatus().GetExecutionError(), w.GetExecutionStatus().GetMessage())
	errorNode := w.GetOnFailureNode()
	execcontext := executors.NewExecutionContext(w, w, w, nil, executors.InitializeControlFlow())
	state, err := c.nodeExecutor.RecursiveNodeHandler(ctx, execcontext, w, w, errorNode)
	if err != nil {
		return StatusFailureNode(execErr), err
	}

	if state.HasFailed() {
		return StatusFailed(state.Err), nil
	}

	if state.HasTimedOut() {
		return StatusFailed(&core.ExecutionError{
			Kind:    core.ExecutionError_USER,
			Code:    "TimedOut",
			Message: "FailureNode Timed-out"}), nil
	}

	if state.PartiallyComplete() {
		// Re-enqueue the workflow
		c.enqueueWorkflow(w.GetK8sWorkflowID().String())
		return StatusFailureNode(execErr), nil
	}

	// If the failure node finished executing, transition to failed.
	return StatusFailed(execErr), nil
}

func executionErrorOrDefault(execError *core.ExecutionError, fallbackMessage string) *core.ExecutionError {
	if execError == nil {
		return &core.ExecutionError{
			Code:    "UnknownError",
			Message: fmt.Sprintf("Unknown error, last seen message [%s]", fallbackMessage),
			Kind:    core.ExecutionError_UNKNOWN,
		}
	}

	return execError
}

func (c *workflowExecutor) handleFailingWorkflow(ctx context.Context, w *v1alpha1.FlyteWorkflow) (Status, error) {
	execErr := executionErrorOrDefault(w.GetExecutionStatus().GetExecutionError(), w.GetExecutionStatus().GetMessage())

	// Best effort clean-up.
	if err := c.cleanupRunningNodes(ctx, w, "Some node execution failed, auto-abort."); err != nil {
		logger.Errorf(ctx, "Failed to propagate Abort for workflow:%v. Error: %v",
			w.ExecutionID.WorkflowExecutionIdentifier, err)
		return StatusFailing(execErr), err
	}

	errorNode := w.GetOnFailureNode()
	if errorNode != nil {
		return StatusFailureNode(execErr), nil
	}

	return StatusFailed(execErr), nil
}

func (c *workflowExecutor) handleSucceedingWorkflow(ctx context.Context, w *v1alpha1.FlyteWorkflow) Status {
	logger.Infof(ctx, "Workflow completed successfully")
	endNodeStatus := w.GetNodeExecutionStatus(ctx, v1alpha1.EndNodeID)
	if endNodeStatus.GetPhase() == v1alpha1.NodePhaseSucceeded {
		if endNodeStatus.GetOutputDir() != "" {
			w.Status.SetOutputReference(v1alpha1.GetOutputsFile(endNodeStatus.GetOutputDir()))
		}
	}
	return StatusSuccess
}

func convertToExecutionError(err *core.ExecutionError, alternateErr *core.ExecutionError) *event.WorkflowExecutionEvent_Error {
	if err == nil {
		if alternateErr != nil {
			err = alternateErr
		} else {
			err = &core.ExecutionError{
				Code:    errors.RuntimeExecutionError.String(),
				Message: "Unknown error",
				Kind:    core.ExecutionError_UNKNOWN,
			}
		}
	}
	return &event.WorkflowExecutionEvent_Error{
		Error: err,
	}
}

func (c *workflowExecutor) IdempotentReportEvent(ctx context.Context, e *event.WorkflowExecutionEvent) error {
	err := c.wfRecorder.RecordWorkflowEvent(ctx, e, c.eventConfig)
	if err != nil && eventsErr.IsAlreadyExists(err) {
		logger.Infof(ctx, "Workflow event phase: %s, executionId %s already exist",
			e.Phase.String(), e.ExecutionId)
		return nil
	}
	return err
}

func (c *workflowExecutor) TransitionToPhase(ctx context.Context, execID *core.WorkflowExecutionIdentifier, wStatus v1alpha1.ExecutableWorkflowStatus, toStatus Status) error {
	if wStatus.GetPhase() != toStatus.TransitionToPhase {
		logger.Debugf(ctx, "Transitioning/Recording event for workflow state transition [%s] -> [%s]", wStatus.GetPhase().String(), toStatus.TransitionToPhase.String())

		wfEvent := &event.WorkflowExecutionEvent{
			ExecutionId: execID,
			ProducerId:  c.clusterID,
		}
		previousError := wStatus.GetExecutionError()
		switch toStatus.TransitionToPhase {
		case v1alpha1.WorkflowPhaseReady:
			// Do nothing
			return nil
		case v1alpha1.WorkflowPhaseRunning:
			wfEvent.Phase = core.WorkflowExecution_RUNNING
			wStatus.UpdatePhase(v1alpha1.WorkflowPhaseRunning, "Workflow Started", nil)
			wfEvent.OccurredAt = utils.GetProtoTime(wStatus.GetStartedAt())
		case v1alpha1.WorkflowPhaseHandlingFailureNode:
			fallthrough
		case v1alpha1.WorkflowPhaseFailing:
			wfEvent.Phase = core.WorkflowExecution_FAILING
			wfEvent.OutputResult = convertToExecutionError(toStatus.Err, previousError)
			wStatus.UpdatePhase(v1alpha1.WorkflowPhaseFailing, "", wfEvent.GetError())
			wfEvent.OccurredAt = utils.GetProtoTime(nil)
		case v1alpha1.WorkflowPhaseFailed:
			wfEvent.Phase = core.WorkflowExecution_FAILED
			wfEvent.OutputResult = convertToExecutionError(toStatus.Err, previousError)
			wStatus.UpdatePhase(v1alpha1.WorkflowPhaseFailed, "", wfEvent.GetError())
			wfEvent.OccurredAt = utils.GetProtoTime(wStatus.GetStoppedAt())
			// Completion latency is only observed when a workflow completes successfully
			c.metrics.FailureDuration.Observe(ctx, wStatus.GetStartedAt().Time, wStatus.GetStoppedAt().Time)
		case v1alpha1.WorkflowPhaseSucceeding:
			wfEvent.Phase = core.WorkflowExecution_SUCCEEDING
			endNodeStatus := wStatus.GetNodeExecutionStatus(ctx, v1alpha1.EndNodeID)
			// Workflow completion latency is recorded as the time it takes for the workflow to transition from end
			// node started time to workflow success being sent to the control plane.
			if endNodeStatus != nil && endNodeStatus.GetStartedAt() != nil {
				c.metrics.CompletionLatency.Observe(ctx, endNodeStatus.GetStartedAt().Time, time.Now())
			}

			wStatus.UpdatePhase(v1alpha1.WorkflowPhaseSucceeding, "", nil)
			wfEvent.OccurredAt = utils.GetProtoTime(nil)
		case v1alpha1.WorkflowPhaseSuccess:
			wfEvent.Phase = core.WorkflowExecution_SUCCEEDED
			wStatus.UpdatePhase(v1alpha1.WorkflowPhaseSuccess, "", nil)
			// Not all workflows have outputs
			if wStatus.GetOutputReference() != "" {
				wfEvent.OutputResult = &event.WorkflowExecutionEvent_OutputUri{
					OutputUri: wStatus.GetOutputReference().String(),
				}
			}
			wfEvent.OccurredAt = utils.GetProtoTime(wStatus.GetStoppedAt())
			c.metrics.SuccessDuration.Observe(ctx, wStatus.GetStartedAt().Time, wStatus.GetStoppedAt().Time)
		case v1alpha1.WorkflowPhaseAborted:
			wfEvent.Phase = core.WorkflowExecution_ABORTED
			if wStatus.GetLastUpdatedAt() != nil {
				c.metrics.CompletionLatency.Observe(ctx, wStatus.GetLastUpdatedAt().Time, time.Now())
			}
			wStatus.UpdatePhase(v1alpha1.WorkflowPhaseAborted, "", nil)
			wfEvent.OccurredAt = utils.GetProtoTime(wStatus.GetStoppedAt())
		default:
			return errors.Errorf(errors.IllegalStateError, "", "Illegal transition from [%v] -> [%v]", wStatus.GetPhase().String(), toStatus.TransitionToPhase.String())
		}

		if recordingErr := c.IdempotentReportEvent(ctx, wfEvent); recordingErr != nil {
			if eventsErr.IsAlreadyExists(recordingErr) {
				logger.Warningf(ctx, "Failed to record workflowEvent, error [%s]. Trying to record state: %s. Ignoring this error!", recordingErr.Error(), wfEvent.Phase)
				return nil
			}
			if eventsErr.IsEventAlreadyInTerminalStateError(recordingErr) {
				// Move to WorkflowPhaseFailed for state mis-match
				msg := fmt.Sprintf("workflow state mismatch between propeller and control plane; Propeller State: %s, ExecutionId %s", wfEvent.Phase.String(), wfEvent.ExecutionId)
				logger.Warningf(ctx, msg)
				wStatus.UpdatePhase(v1alpha1.WorkflowPhaseFailed, msg, nil)
				return nil
			}
			if (wfEvent.Phase == core.WorkflowExecution_FAILING || wfEvent.Phase == core.WorkflowExecution_FAILED) &&
				(eventsErr.IsNotFound(recordingErr) || eventsErr.IsEventIncompatibleClusterError(recordingErr)) {
				// Don't stall the workflow transition to terminated (so that resources can be cleaned up) since these events
				// are being discarded by the back-end anyways.
				logger.Infof(ctx, "Failed to record %s workflowEvent, error [%s]. Ignoring this error!", wfEvent.Phase.String(), recordingErr.Error())
				return nil
			}
			logger.Warningf(ctx, "Event recording failed. Error [%s]", recordingErr.Error())
			return errors.Wrapf(errors.EventRecordingError, "", recordingErr, "failed to publish event")
		}
	}
	return nil
}

func (c *workflowExecutor) Initialize(ctx context.Context) error {
	logger.Infof(ctx, "Initializing Core Workflow Executor")
	return c.nodeExecutor.Initialize(ctx)
}

func (c *workflowExecutor) HandleFlyteWorkflow(ctx context.Context, w *v1alpha1.FlyteWorkflow) error {
	logger.Infof(ctx, "Handling Workflow [%s], id: [%s], p [%s]", w.GetName(), w.GetExecutionID(), w.GetExecutionStatus().GetPhase().String())
	defer logger.Infof(ctx, "Handling Workflow [%s] Done", w.GetName())

	w.DataReferenceConstructor = c.store

	wStatus := w.GetExecutionStatus()
	// Initialize the Status if not already initialized
	switch wStatus.GetPhase() {
	case v1alpha1.WorkflowPhaseReady:
		newStatus, err := c.handleReadyWorkflow(ctx, w)
		if err != nil {
			return err
		}
		c.metrics.AcceptedWorkflows.Inc(ctx)
		if err := c.TransitionToPhase(ctx, w.ExecutionID.WorkflowExecutionIdentifier, wStatus, newStatus); err != nil {
			return err
		}
		c.k8sRecorder.Event(w, corev1.EventTypeNormal, v1alpha1.WorkflowPhaseRunning.String(), "Workflow began execution")

		// TODO: Consider annotating with the newStatus.
		acceptedAt := w.GetCreationTimestamp().Time
		if w.AcceptedAt != nil && !w.AcceptedAt.IsZero() {
			acceptedAt = w.AcceptedAt.Time
		}

		c.metrics.AcceptanceLatency.Observe(ctx, acceptedAt, time.Now())
		return nil

	case v1alpha1.WorkflowPhaseRunning:
		newStatus, err := c.handleRunningWorkflow(ctx, w)
		if err != nil {
			logger.Warningf(ctx, "Error in handling running workflow [%v]", err.Error())
			return err
		}
		if err := c.TransitionToPhase(ctx, w.ExecutionID.WorkflowExecutionIdentifier, wStatus, newStatus); err != nil {
			return err
		}
		return nil
	case v1alpha1.WorkflowPhaseSucceeding:
		newStatus := c.handleSucceedingWorkflow(ctx, w)

		if err := c.TransitionToPhase(ctx, w.ExecutionID.WorkflowExecutionIdentifier, wStatus, newStatus); err != nil {
			return err
		}
		c.k8sRecorder.Event(w, corev1.EventTypeNormal, v1alpha1.WorkflowPhaseSuccess.String(), "Workflow completed.")
		return nil
	case v1alpha1.WorkflowPhaseFailing:
		newStatus, err := c.handleFailingWorkflow(ctx, w)
		if err != nil {
			return err
		}
		failingErr := c.TransitionToPhase(ctx, w.ExecutionID.WorkflowExecutionIdentifier, wStatus, newStatus)
		// Ignore ExecutionNotFound and IncompatibleCluster errors to allow graceful failure
		if failingErr != nil && !(eventsErr.IsNotFound(failingErr) || eventsErr.IsEventIncompatibleClusterError(failingErr)) {
			return failingErr
		}
		c.k8sRecorder.Event(w, corev1.EventTypeWarning, v1alpha1.WorkflowPhaseFailed.String(), "Workflow failed.")
		return nil
	case v1alpha1.WorkflowPhaseHandlingFailureNode:
		newStatus, err := c.handleFailureNode(ctx, w)
		if err != nil {
			return err
		}
		failureErr := c.TransitionToPhase(ctx, w.ExecutionID.WorkflowExecutionIdentifier, wStatus, newStatus)
		// Ignore ExecutionNotFound and IncompatibleCluster errors to allow graceful failure
		if failureErr != nil && !(eventsErr.IsNotFound(failureErr) || eventsErr.IsEventIncompatibleClusterError(failureErr)) {
			return failureErr
		}
		c.k8sRecorder.Event(w, corev1.EventTypeWarning, v1alpha1.WorkflowPhaseFailed.String(), "Workflow failed.")
		return nil
	default:
		return errors.Errorf(errors.IllegalStateError, w.ID, "Unsupported state [%s] for workflow", w.GetExecutionStatus().GetPhase().String())
	}
}

func (c *workflowExecutor) HandleAbortedWorkflow(ctx context.Context, w *v1alpha1.FlyteWorkflow, maxRetries uint32) error {
	w.DataReferenceConstructor = c.store
	if !w.Status.IsTerminated() {
		reason := fmt.Sprintf("max number of system retry attempts [%d/%d] exhausted - system failure.", w.Status.FailedAttempts, maxRetries)
		c.metrics.IncompleteWorkflowAborted.Inc(ctx)
		// Check of the workflow was deleted and that caused the abort
		if w.GetDeletionTimestamp() != nil {
			reason = "Workflow aborted."
		}

		// We will always try to cleanup, even if we have extinguished all our retries
		// TODO ABORT should have its separate set of retries
		err := c.cleanupRunningNodes(ctx, w, reason)
		// Best effort clean-up.
		if err != nil && w.Status.FailedAttempts <= maxRetries {
			logger.Errorf(ctx, "Failed to propagate Abort for workflow:%v. Error: %v", w.ExecutionID.WorkflowExecutionIdentifier, err)
			return err
		}

		if w.Status.FailedAttempts > maxRetries {
			err = errors.Errorf(errors.RuntimeExecutionError, w.GetID(), "max number of system retry attempts [%d/%d] exhausted. Last known status message: %v", w.Status.FailedAttempts, maxRetries, w.Status.Message)
		}

		var status Status
		if err != nil {
			// This workflow failed, record that phase and corresponding error message.
			status = StatusFailed(&core.ExecutionError{
				Code:    "Workflow abort failed",
				Message: err.Error(),
				Kind:    core.ExecutionError_SYSTEM,
			})
		} else {
			// Otherwise, this workflow is aborted.
			status = Status{
				TransitionToPhase: v1alpha1.WorkflowPhaseAborted,
			}
		}
		if err := c.TransitionToPhase(ctx, w.ExecutionID.WorkflowExecutionIdentifier, w.GetExecutionStatus(), status); err != nil {
			return err
		}
	}
	return nil
}

func (c *workflowExecutor) cleanupRunningNodes(ctx context.Context, w v1alpha1.ExecutableWorkflow, reason string) error {
	startNode := w.StartNode()
	if startNode == nil {
		return errors.Errorf(errors.IllegalStateError, w.GetID(), "StartNode not found in running workflow?")
	}

	execcontext := executors.NewExecutionContext(w, w, w, nil, executors.InitializeControlFlow())
	if err := c.nodeExecutor.AbortHandler(ctx, execcontext, w, w, startNode, reason); err != nil {
		return errors.Errorf(errors.CausedByError, w.GetID(), "Failed to propagate Abort for workflow. Error: %v", err)
	}

	return nil
}

func NewExecutor(ctx context.Context, store *storage.DataStore, enQWorkflow v1alpha1.EnqueueWorkflow, eventSink events.EventSink,
	k8sEventRecorder record.EventRecorder, metadataPrefix string, nodeExecutor interfaces.Node, eventConfig *config.EventConfig,
	clusterID string, scope promutils.Scope) (executors.Workflow, error) {
	basePrefix := store.GetBaseContainerFQN(ctx)
	if metadataPrefix != "" {
		var err error
		basePrefix, err = store.ConstructReference(ctx, basePrefix, metadataPrefix)
		if err != nil {
			return nil, err
		}
	}
	logger.Infof(ctx, "Metadata will be stored in container path: [%s]", basePrefix)

	workflowScope := scope.NewSubScope("workflow")

	return &workflowExecutor{
		nodeExecutor:    nodeExecutor,
		store:           store,
		enqueueWorkflow: enQWorkflow,
		wfRecorder:      events.NewWorkflowEventRecorder(eventSink, workflowScope, store),
		k8sRecorder:     k8sEventRecorder,
		metadataPrefix:  basePrefix,
		metrics:         newMetrics(workflowScope),
		eventConfig:     eventConfig,
		clusterID:       clusterID,
	}, nil
}

func newMetrics(workflowScope promutils.Scope) *workflowMetrics {
	return &workflowMetrics{
		AcceptedWorkflows:         labeled.NewCounter("accepted", "Number of workflows accepted by propeller", workflowScope),
		FailureDuration:           labeled.NewStopWatch("failure_duration", "Indicates the total execution time of a failed workflow.", time.Millisecond, workflowScope, labeled.EmitUnlabeledMetric),
		SuccessDuration:           labeled.NewStopWatch("success_duration", "Indicates the total execution time of a successful workflow.", time.Millisecond, workflowScope, labeled.EmitUnlabeledMetric),
		IncompleteWorkflowAborted: labeled.NewCounter("workflow_aborted", "Indicates an inprogress execution was aborted", workflowScope, labeled.EmitUnlabeledMetric),
		AcceptanceLatency:         labeled.NewStopWatch("acceptance_latency", "Delay between workflow creation and moving it to running state.", time.Millisecond, workflowScope, labeled.EmitUnlabeledMetric),
		CompletionLatency:         labeled.NewStopWatch("completion_latency", "Measures the time between when the WF moved to succeeding/failing state and when it finally moved to a terminal state.", time.Millisecond, workflowScope, labeled.EmitUnlabeledMetric),
	}
}
