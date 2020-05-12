package controller

import (
	"context"
	"fmt"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/promutils/labeled"

	"github.com/lyft/flytepropeller/pkg/controller/config"
	"github.com/lyft/flytepropeller/pkg/controller/workflowstore"

	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/lyft/flytepropeller/pkg/controller/executors"
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
	}
}

type Propeller struct {
	wfStore          workflowstore.FlyteWorkflow
	workflowExecutor executors.Workflow
	metrics          *propellerMetrics
	cfg              *config.Config
}

func (p *Propeller) Initialize(ctx context.Context) error {
	return p.workflowExecutor.Initialize(ctx)
}

// reconciler compares the actual state with the desired, and attempts to
// converge the two. It then updates the GetExecutionStatus block of the FlyteWorkflow resource
// with the current status of the resource.
// Every FlyteWorkflow transitions through the following
//
// The Workflow to be worked on is identified for the given namespace and executionID (which is the name of the workflow)
// The return value should be an error, in the case, we wish to retry this workflow
// <pre>
//
//     +--------+        +--------+        +--------+     +--------+
//     |        |        |        |        |        |     |        |
//     | Ready  +--------> Running+--------> Succeeding---> Success|
//     |        |        |        |        |        |     |        |
//     +--------+        +--------+        +---------     +--------+
//         |                  |
//         |                  |
//         |             +----v---+        +--------+
//         |             |        |        |        |
//         +-------------> Failing+--------> Failed |
//                       |        |        |        |
//                       +--------+        +--------+
// </pre>
func (p *Propeller) Handle(ctx context.Context, namespace, name string) error {
	logger.Infof(ctx, "Processing Workflow.")
	defer logger.Infof(ctx, "Completed processing workflow.")

	// Get the FlyteWorkflow resource with this namespace/name
	w, err := p.wfStore.Get(ctx, namespace, name)
	if err != nil {
		if workflowstore.IsNotFound(err) {
			p.metrics.WorkflowNotFound.Inc()
			logger.Warningf(ctx, "Workflow namespace[%v]/name[%v] not found, may be deleted.", namespace, name)
			return nil
		}
		if workflowstore.IsWorkflowStale(err) {
			p.metrics.RoundSkipped.Inc()
			logger.Warningf(ctx, "Workflow namespace[%v]/name[%v] Stale.", namespace, name)
			return nil
		}
		logger.Warningf(ctx, "Failed to GetWorkflow, retrying with back-off", err)
		return err
	}

	t := p.metrics.DeepCopyTime.Start()
	wfDeepCopy := w.DeepCopy()
	t.Stop()
	ctx = contextutils.WithWorkflowID(ctx, wfDeepCopy.GetID())
	if execID := wfDeepCopy.GetExecutionID(); execID.WorkflowExecutionIdentifier != nil {
		ctx = contextutils.WithProjectDomain(ctx, wfDeepCopy.GetExecutionID().Project, wfDeepCopy.GetExecutionID().Domain)
	}
	ctx = contextutils.WithResourceVersion(ctx, wfDeepCopy.GetResourceVersion())

	maxRetries := uint32(p.cfg.MaxWorkflowRetries)
	if IsDeleted(wfDeepCopy) || (wfDeepCopy.Status.FailedAttempts > maxRetries) {
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
			err = p.workflowExecutor.HandleAbortedWorkflow(ctx, wfDeepCopy, maxRetries)
		}()
		if err != nil {
			p.metrics.AbortError.Inc(ctx)
			return err
		}
	} else {
		if wfDeepCopy.GetExecutionStatus().IsTerminated() {
			if HasCompletedLabel(wfDeepCopy) && !HasFinalizer(wfDeepCopy) {
				logger.Debugf(ctx, "Workflow is terminated.")
				return nil
			}
			// NOTE: This should never really happen, but in case we externally mark the workflow as terminated
			// We should allow cleanup
			logger.Warn(ctx, "Workflow is marked as terminated but doesn't have the completed label, marking it as completed.")
		} else {
			SetFinalizerIfEmpty(wfDeepCopy, FinalizerKey)

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
				err = p.workflowExecutor.HandleFlyteWorkflow(ctx, wfDeepCopy)
			}()

			if err != nil {
				logger.Errorf(ctx, "Error when trying to reconcile workflow. Error [%v]. Error Type[%v]. Is nill [%v]",
					err, reflect.TypeOf(err))

				// Let's mark these as system errors.
				// We only want to increase failed attempts and discard any other partial changes to the CRD.
				wfDeepCopy = w.DeepCopy()
				wfDeepCopy.GetExecutionStatus().IncFailedAttempts()
				wfDeepCopy.GetExecutionStatus().SetMessage(err.Error())
				p.metrics.SystemError.Inc(ctx)
			} else {
				// No updates in the status we detected, we will skip writing to KubeAPI
				if wfDeepCopy.Status.Equals(&w.Status) {
					logger.Info(ctx, "WF hasn't been updated in this round.")
					return nil
				}
			}
		}
	}
	// If the end result is a terminated workflow, we remove the labels
	if wfDeepCopy.GetExecutionStatus().IsTerminated() {
		// We add a completed label so that we can avoid polling for this workflow
		SetCompletedLabel(wfDeepCopy, time.Now())
		ResetFinalizers(wfDeepCopy)
	}
	// TODO we will need to call updatestatus when it is supported. But to preserve metadata like (label/finalizer) we will need to use update

	// update the GetExecutionStatus block of the FlyteWorkflow resource. UpdateStatus will not
	// allow changes to the Spec of the resource, which is ideal for ensuring
	// nothing other than resource status has been updated.
	_, err = p.wfStore.Update(ctx, wfDeepCopy, workflowstore.PriorityClassCritical)
	return err
}

func NewPropellerHandler(_ context.Context, cfg *config.Config, wfStore workflowstore.FlyteWorkflow, executor executors.Workflow, scope promutils.Scope) *Propeller {

	metrics := newPropellerMetrics(scope)
	return &Propeller{
		metrics:          metrics,
		wfStore:          wfStore,
		workflowExecutor: executor,
		cfg:              cfg,
	}
}
