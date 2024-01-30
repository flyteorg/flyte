package controller

import (
	"context"
	"fmt"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	eventsErr "github.com/flyteorg/flyte/flytepropeller/events/errors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/transformers/k8s"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/workflowstore"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

// TODO Lets move everything to use controller runtime

type propellerMetrics struct {
	Scope                    promutils.Scope
	DeepCopyTime             promutils.StopWatch
	RawWorkflowTraversalTime labeled.StopWatch
	SystemError              labeled.Counter
	AbortError               labeled.Counter
	PanicObserved            labeled.Counter
	RoundSkipped             prometheus.Counter
	WorkflowNotFound         prometheus.Counter
	WorkflowClosureReadTime  labeled.StopWatch
	StreakLength             labeled.Counter
	RoundTime                labeled.StopWatch
}

func newPropellerMetrics(scope promutils.Scope) *propellerMetrics {
	roundScope := scope.NewSubScope("round")
	return &propellerMetrics{
		Scope:                    scope,
		DeepCopyTime:             roundScope.MustNewStopWatch("deepcopy", "Total time to deep copy wf object", time.Millisecond),
		RawWorkflowTraversalTime: labeled.NewStopWatch("raw", "Total time to traverse the workflow", time.Millisecond, roundScope, labeled.EmitUnlabeledMetric),
		SystemError:              labeled.NewCounter("system_error", "Failure to reconcile a workflow, system error", roundScope, labeled.EmitUnlabeledMetric),
		AbortError:               labeled.NewCounter("abort_error", "Failure to abort a workflow, system error", roundScope, labeled.EmitUnlabeledMetric),
		PanicObserved:            labeled.NewCounter("panic", "Panic during handling or aborting workflow", roundScope, labeled.EmitUnlabeledMetric),
		RoundSkipped:             roundScope.MustNewCounter("skipped", "Round Skipped because of stale workflow"),
		WorkflowNotFound:         roundScope.MustNewCounter("not_found", "workflow not found in the cache"),
		WorkflowClosureReadTime:  labeled.NewStopWatch("closure_read", "Total time taken to read and parse the offloaded WorkflowClosure", time.Millisecond, roundScope, labeled.EmitUnlabeledMetric),
		StreakLength:             labeled.NewCounter("streak_length", "Number of consecutive rounds used in fast follow mode", roundScope, labeled.EmitUnlabeledMetric),
		RoundTime:                labeled.NewStopWatch("round_time", "Total time taken by one round traversing, copying and storing a workflow", time.Millisecond, roundScope, labeled.EmitUnlabeledMetric),
	}
}

// Helper method to record system error in the workflow.
func RecordSystemError(w *v1alpha1.FlyteWorkflow, err error) *v1alpha1.FlyteWorkflow {
	// Let's mark these as system errors.
	// We only want to increase failed attempts and discard any other partial changes to the CRD.
	wfDeepCopy := w.DeepCopy()
	wfDeepCopy.GetExecutionStatus().IncFailedAttempts()
	wfDeepCopy.GetExecutionStatus().SetMessage(err.Error())
	return wfDeepCopy
}

// Core Propeller structure that houses the Reconciliation loop for Flytepropeller
type Propeller struct {
	store            *storage.DataStore
	wfStore          workflowstore.FlyteWorkflow
	workflowExecutor executors.Workflow
	metrics          *propellerMetrics
	cfg              *config.Config
}

// Initialize initializes all downstream executors
func (p *Propeller) Initialize(ctx context.Context) error {
	return p.workflowExecutor.Initialize(ctx)
}

func SetDefinitionVersionIfEmpty(wf *v1alpha1.FlyteWorkflow, version v1alpha1.WorkflowDefinitionVersion) {
	if wf.Status.DefinitionVersion == nil {
		wf.Status.DefinitionVersion = &version
	}
}

// TryMutateWorkflow will try to mutate the workflow by traversing it and reconciling the desired and actual state.
// The desired state here is the entire workflow is completed, actual state is each nodes current execution state.
func (p *Propeller) TryMutateWorkflow(ctx context.Context, originalW *v1alpha1.FlyteWorkflow) (*v1alpha1.FlyteWorkflow, error) {

	t := p.metrics.DeepCopyTime.Start()
	mutableW := originalW.DeepCopy()
	t.Stop()
	ctx = contextutils.WithWorkflowID(ctx, mutableW.GetID())
	if execID := mutableW.GetExecutionID(); execID.WorkflowExecutionIdentifier != nil {
		ctx = contextutils.WithProjectDomain(ctx, mutableW.GetExecutionID().Project, mutableW.GetExecutionID().Domain)
	}
	ctx = contextutils.WithResourceVersion(ctx, mutableW.GetResourceVersion())

	maxRetries := uint32(p.cfg.MaxWorkflowRetries)
	if IsDeleted(mutableW) || (mutableW.Status.FailedAttempts > maxRetries) {
		var err error
		func() {
			defer func() {
				if r := recover(); r != nil {
					stack := debug.Stack()
					err = fmt.Errorf("panic when aborting workflow, Stack: [%s]", string(stack))
					logger.Errorf(ctx, err.Error())
					p.metrics.PanicObserved.Inc(ctx)
				}
			}()
			err = p.workflowExecutor.HandleAbortedWorkflow(ctx, mutableW, maxRetries)
		}()
		if err != nil {
			p.metrics.AbortError.Inc(ctx)
			return nil, err
		}
		return mutableW, nil
	}

	if !mutableW.GetExecutionStatus().IsTerminated() {
		var err error
		SetFinalizerIfEmpty(mutableW, FinalizerKey)
		SetDefinitionVersionIfEmpty(mutableW, v1alpha1.LatestWorkflowDefinitionVersion)

		func() {
			t := p.metrics.RawWorkflowTraversalTime.Start(ctx)
			defer func() {
				t.Stop()
				if r := recover(); r != nil {
					stack := debug.Stack()
					err = fmt.Errorf("panic when reconciling workflow, Stack: [%s]", string(stack))
					logger.Errorf(ctx, err.Error())
					p.metrics.PanicObserved.Inc(ctx)
				}
			}()
			err = p.workflowExecutor.HandleFlyteWorkflow(ctx, mutableW)
		}()
		if err != nil {
			logger.Errorf(ctx, "Error when trying to reconcile workflow. Error [%v]. Error Type[%v]",
				err, reflect.TypeOf(err))
			p.metrics.SystemError.Inc(ctx)
			return nil, err
		}
	} else {
		logger.Warn(ctx, "Workflow is marked as terminated but doesn't have the completed label, marking it as completed.")
	}
	return mutableW, nil
}

// Handle method is the entry point for the reconciler.
// It compares the actual state with the desired, and attempts to
// converge the two. It then updates the GetExecutionStatus block of the FlyteWorkflow resource
// with the current status of the resource.
// Every FlyteWorkflow transitions through the following
//
// The Workflow to be worked on is identified for the given namespace and executionID (which is the name of the workflow)
// The return value should be an error, in the case, we wish to retry this workflow
// <pre>
//
//	+--------+        +---------+        +------------+     +---------+
//	|        |        |         |        |            |     |         |
//	| Ready  +--------> Running +--------> Succeeding +-----> Success |
//	|        |        |         |        |            |     |         |
//	+--------+        +---------+        +------------+     +---------+
//	    |                  |
//	    |                  |
//	    |             +----v----+        +---------------------+        +--------+
//	    |             |         |        |     (optional)      |        |        |
//	    +-------------> Failing +--------> HandlingFailureNode +--------> Failed |
//	                  |         |        |                     |        |        |
//	                  +---------+        +---------------------+        +--------+
//
// </pre>
func (p *Propeller) Handle(ctx context.Context, namespace, name string) error {
	var span trace.Span
	ctx, span = otelutils.NewSpan(ctx, otelutils.FlytePropellerTracer, "pkg.controller.Propeller/Handle")
	defer span.End()

	logger.Infof(ctx, "Processing Workflow.")
	defer logger.Infof(ctx, "Completed processing workflow.")

	// Get the FlyteWorkflow resource with this namespace/name
	_, wfStoreGetSpan := otelutils.NewSpan(ctx, otelutils.FlytePropellerTracer, "WorkflowStore.Get")
	w, fetchErr := p.wfStore.Get(ctx, namespace, name)
	wfStoreGetSpan.End()
	if fetchErr != nil {
		if workflowstore.IsNotFound(fetchErr) {
			p.metrics.WorkflowNotFound.Inc()
			logger.Warningf(ctx, "Workflow namespace[%v]/name[%v] not found, may be deleted.", namespace, name)
			return nil
		}
		if workflowstore.IsWorkflowTerminated(fetchErr) {
			p.metrics.RoundSkipped.Inc()
			logger.Warningf(ctx, "Workflow namespace[%v]/name[%v] has already been terminated.", namespace, name)
			return nil
		}
		if workflowstore.IsWorkflowStale(fetchErr) {
			p.metrics.RoundSkipped.Inc()
			logger.Warningf(ctx, "Workflow namespace[%v]/name[%v] Stale.", namespace, name)
			return nil
		}
		logger.Warningf(ctx, "Failed to GetWorkflow, retrying with back-off", fetchErr)
		return fetchErr
	}

	if w.GetExecutionStatus().IsTerminated() {
		if HasCompletedLabel(w) && !HasFinalizer(w) {
			logger.Debugf(ctx, "Workflow is terminated.")
			// This workflow had previously completed, let us ignore it
			return nil
		}
	}

	// if the FlyteWorkflow CRD has the WorkflowClosureReference set then we have offloaded the
	// static fields to the blobstore to reduce CRD size. we must read and parse the workflow
	// closure so that these fields may be temporarily repopulated.
	var wfClosureCrdFields *k8s.WfClosureCrdFields
	var err error
	if len(w.WorkflowClosureReference) > 0 {
		wfClosureCrdFields, err = p.parseWorkflowClosureCrdFields(ctx, w.WorkflowClosureReference)
		if err != nil {
			return err
		}
	}

	streak := 0
	defer p.metrics.StreakLength.Add(ctx, float64(streak))

	maxLength := p.cfg.MaxStreakLength
	if maxLength <= 0 {
		maxLength = 1
	}

	for streak = 0; streak < maxLength; streak++ {
		w, err = p.streak(ctx, w, wfClosureCrdFields)
		if err != nil {
			return err
		} else if w == nil {
			break
		}

		logger.Infof(ctx, "FastFollow Enabled. Detected State change, we will try another round. StreakLength [%d]", streak)
	}
	logger.Infof(ctx, "Streak ended at [%d]/Max: [%d]", streak, maxLength)
	return nil
}

// parseWorkflowClosureCrdFields attempts to retrieve offloaded static workflow closure data from the specified
// DataReference.
func (p *Propeller) parseWorkflowClosureCrdFields(ctx context.Context, dataReference storage.DataReference) (*k8s.WfClosureCrdFields, error) {
	_, span := otelutils.NewSpan(ctx, otelutils.FlytePropellerTracer, "pkg.controller.Propeller/parseWorkflowClosureCrdFields")
	defer span.End()

	t := p.metrics.WorkflowClosureReadTime.Start(ctx)
	defer t.Stop()

	wfClosure := &admin.WorkflowClosure{}
	err := p.store.ReadProtobuf(ctx, dataReference, wfClosure)
	if err != nil {
		logger.Errorf(ctx, "Failed to retrieve workflow closure data from '%s' with error '%s'", dataReference, err)
		return nil, err
	}

	wfClosureCrdFields, err := k8s.BuildWfClosureCrdFields(wfClosure.CompiledWorkflow)
	if err != nil {
		logger.Errorf(ctx, "Failed to parse workflow closure data from '%s' with error '%s'", dataReference, err)
		return nil, err
	}

	return wfClosureCrdFields, nil
}

// streak performs a single iteration of mutating a workflow returning the newly mutated workflow on success or nil if
// the workflow was not updated.
func (p *Propeller) streak(ctx context.Context, w *v1alpha1.FlyteWorkflow, wfClosureCrdFields *k8s.WfClosureCrdFields) (*v1alpha1.FlyteWorkflow, error) {
	ctx, span := otelutils.NewSpan(ctx, otelutils.FlytePropellerTracer, "pkg.controller.Propeller/streak")
	defer span.End()

	t := p.metrics.RoundTime.Start(ctx)
	defer t.Stop()

	// if the wfClosureCrdFields struct is not nil then it contains static workflow data which
	// has been offloaded to the blobstore. we must set these fields so they're available
	// during workflow processing and immediately remove them afterwards so they do not
	// accidentally get written to the workflow store once the new state is stored.
	if wfClosureCrdFields != nil {
		w.WorkflowSpec = wfClosureCrdFields.WorkflowSpec
		w.Tasks = wfClosureCrdFields.Tasks
		w.SubWorkflows = wfClosureCrdFields.SubWorkflows
	}

	mutatedWf, err := p.TryMutateWorkflow(ctx, w)

	if wfClosureCrdFields != nil {
		// strip data populated from WorkflowClosureReference
		w.SubWorkflows, w.Tasks, w.WorkflowSpec = nil, nil, nil
		if mutatedWf != nil {
			mutatedWf.SubWorkflows, mutatedWf.Tasks, mutatedWf.WorkflowSpec = nil, nil, nil
		}
	}

	if err != nil {
		// NOTE We are overriding the deepcopy here, as we are essentially ignoring all mutations
		// We only want to increase failed attempts and discard any other partial changes to the CRD.
		mutatedWf = RecordSystemError(w, err)
		p.metrics.SystemError.Inc(ctx)
	} else if mutatedWf == nil {
		logger.Errorf(ctx, "Should not happen! Mutation resulted in a nil workflow!")
		return nil, nil
	} else {
		if !w.GetExecutionStatus().IsTerminated() {
			// No updates in the status we detected, we will skip writing to KubeAPI
			if mutatedWf.Status.Equals(&w.Status) {
				logger.Info(ctx, "WF hasn't been updated in this round.")
				return nil, nil
			}
		}
		if mutatedWf.GetExecutionStatus().IsTerminated() {
			// If the end result is a terminated workflow, we remove the labels
			// We add a completed label so that we can avoid polling for this workflow
			SetCompletedLabel(mutatedWf, time.Now())
			ResetFinalizers(mutatedWf)
		}
	}

	// ExecutionNotFound error is returned when flyteadmin is missing the workflow. This is not
	// a valid state unless we are experiencing a race condition where the workflow has not yet
	// been inserted into the db (ie. workflow phase is WorkflowPhaseReady).
	if err != nil && eventsErr.IsNotFound(err) && w.GetExecutionStatus().GetPhase() != v1alpha1.WorkflowPhaseReady {
		logger.Errorf(ctx, "Failed to process workflow, failing: %s", err)

		// We set the workflow status to failing to abort any active tasks in the next round.
		mutableW := w.DeepCopy()
		mutableW.Status.UpdatePhase(v1alpha1.WorkflowPhaseFailing, "Workflow execution is missing in flyteadmin, aborting", &core.ExecutionError{
			Kind:    core.ExecutionError_SYSTEM,
			Code:    "ExecutionNotFound",
			Message: "Workflow execution not found in flyteadmin.",
		})
		if _, e := p.wfStore.Update(ctx, mutableW, workflowstore.PriorityClassCritical); e != nil {
			logger.Errorf(ctx, "Failed to record an ExecutionNotFound workflow as failed, reason: %s. Retrying...", e)
			return nil, e
		}
		return nil, nil
	}

	// Incompatible cluster means that another cluster has been designated to handle this workflow execution.
	// We should early abort in this case, since any events originating from this cluster for this execution will
	// be rejected.
	if err != nil && eventsErr.IsEventIncompatibleClusterError(err) {
		logger.Errorf(ctx, "No longer designated to process workflow, failing: %s", err)

		// We set the workflow status to failing to abort any active tasks in the next round.
		mutableW := w.DeepCopy()
		mutableW.Status.UpdatePhase(v1alpha1.WorkflowPhaseFailing, "Workflow execution cluster reassigned, aborting", &core.ExecutionError{
			Kind:    core.ExecutionError_SYSTEM,
			Code:    string(eventsErr.EventIncompatibleCusterError),
			Message: fmt.Sprintf("Workflow execution cluster reassigned: %v", err),
		})
		if _, e := p.wfStore.Update(ctx, mutableW, workflowstore.PriorityClassCritical); e != nil {
			logger.Errorf(ctx, "Failed to record an EventIncompatibleClusterError workflow as failed, reason: %s. Retrying...", e)
			return nil, e
		}
		return nil, nil
	}

	// TODO we will need to call updatestatus when it is supported. But to preserve metadata like (label/finalizer) we will need to use update

	// update the GetExecutionStatus block of the FlyteWorkflow resource. UpdateStatus will not
	// allow changes to the Spec of the resource, which is ideal for ensuring
	// nothing other than resource status has been updated.
	_, wfStoreUpdateSpan := otelutils.NewSpan(ctx, otelutils.FlytePropellerTracer, "WorkflowStore.Update")
	newWf, updateErr := p.wfStore.Update(ctx, mutatedWf, workflowstore.PriorityClassCritical)
	wfStoreUpdateSpan.End()
	if updateErr != nil {
		// The update has failed, lets check if this is because the size is too large. If so
		if workflowstore.IsWorkflowTooLarge(updateErr) {
			logger.Errorf(ctx, "Failed storing workflow to the store, reason: %s", updateErr)
			p.metrics.SystemError.Inc(ctx)
			// Workflow is too large, we will mark the workflow as failing and record it. This will automatically
			// propagate the failure in the next round.
			mutableW := w.DeepCopy()
			// catch potential indefinite update loop
			if mutatedWf.GetExecutionStatus().IsTerminated() {
				ResetFinalizers(mutableW)
				SetDefinitionVersionIfEmpty(mutableW, v1alpha1.LatestWorkflowDefinitionVersion)
				SetCompletedLabel(mutableW, time.Now())
				msg := fmt.Sprintf("Workflow size has breached threshold. Finalized with status: %v", mutatedWf.GetExecutionStatus().GetPhase())
				mutableW.Status.UpdatePhase(v1alpha1.WorkflowPhaseFailed, msg, &core.ExecutionError{
					Kind:    core.ExecutionError_SYSTEM,
					Code:    "WorkflowTooLarge",
					Message: "Workflow execution state is too large for Flyte to handle.",
				})
			} else {
				mutableW.Status.UpdatePhase(v1alpha1.WorkflowPhaseFailing, "Workflow size has breached threshold, aborting", &core.ExecutionError{
					Kind:    core.ExecutionError_SYSTEM,
					Code:    "WorkflowTooLarge",
					Message: "Workflow execution state is too large for Flyte to handle.",
				})
			}
			if _, e := p.wfStore.Update(ctx, mutableW, workflowstore.PriorityClassCritical); e != nil {
				logger.Errorf(ctx, "Failed recording a large workflow as failed, reason: %s. Retrying...", e)
				return nil, e
			}
			return nil, nil
		}
		return nil, updateErr
	}
	if err != nil {
		// An error was encountered during the round. Let us return, so that we can back-off gracefully
		return nil, err
	}
	if mutatedWf.GetExecutionStatus().IsTerminated() || newWf.ResourceVersion == mutatedWf.ResourceVersion {
		// Workflow is terminated (no need to continue) or no status was changed, we can wait
		logger.Infof(ctx, "Will not fast follow, Reason: Wf terminated? %v, Version matched? %v",
			mutatedWf.GetExecutionStatus().IsTerminated(), newWf.ResourceVersion == mutatedWf.ResourceVersion)
		return nil, nil
	}
	return newWf, nil
}

// NewPropellerHandler creates a new Propeller and initializes metrics
func NewPropellerHandler(_ context.Context, cfg *config.Config, store *storage.DataStore, wfStore workflowstore.FlyteWorkflow, executor executors.Workflow, scope promutils.Scope) *Propeller {

	metrics := newPropellerMetrics(scope)
	return &Propeller{
		metrics:          metrics,
		store:            store,
		wfStore:          wfStore,
		workflowExecutor: executor,
		cfg:              cfg,
	}
}
