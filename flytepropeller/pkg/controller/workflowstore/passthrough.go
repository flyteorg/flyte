package workflowstore

import (
	"context"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	v1alpha12 "github.com/flyteorg/flytepropeller/pkg/client/clientset/versioned/typed/flyteworkflow/v1alpha1"
	listers "github.com/flyteorg/flytepropeller/pkg/client/listers/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/prometheus/client_golang/prometheus"

	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
)

type workflowstoreMetrics struct {
	workflowUpdateCount         prometheus.Counter
	workflowUpdateFailedCount   prometheus.Counter
	workflowUpdateSuccessCount  prometheus.Counter
	workflowUpdateConflictCount prometheus.Counter
	workflowTooLarge            prometheus.Counter
	workflowUpdateLatency       promutils.StopWatch
}

type passthroughWorkflowStore struct {
	wfLister    listers.FlyteWorkflowLister
	wfClientSet v1alpha12.FlyteworkflowV1alpha1Interface
	metrics     *workflowstoreMetrics
}

func (p *passthroughWorkflowStore) Get(ctx context.Context, namespace, name string) (*v1alpha1.FlyteWorkflow, error) {
	w, err := p.wfLister.FlyteWorkflows(namespace).Get(name)
	if err != nil {
		// The FlyteWorkflow resource may no longer exist, in which case we stop
		// processing.
		if kubeerrors.IsNotFound(err) {
			logger.Warningf(ctx, "Workflow not found in cache.")
			return nil, ErrWorkflowNotFound
		}
		return nil, err
	}
	return w, nil
}

func (p *passthroughWorkflowStore) UpdateStatus(ctx context.Context, workflow *v1alpha1.FlyteWorkflow, priorityClass PriorityClass) (
	newWF *v1alpha1.FlyteWorkflow, err error) {
	p.metrics.workflowUpdateCount.Inc()
	// Something has changed. Lets save
	logger.Debugf(ctx, "Observed FlyteWorkflow State change. [%v] -> [%v]", workflow.Status.Phase.String(), workflow.Status.Phase.String())
	t := p.metrics.workflowUpdateLatency.Start()
	newWF, err = p.wfClientSet.FlyteWorkflows(workflow.Namespace).Update(ctx, workflow, v1.UpdateOptions{})
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil, nil
		}

		if kubeerrors.IsConflict(err) {
			p.metrics.workflowUpdateConflictCount.Inc()
		}
		if kubeerrors.IsRequestEntityTooLargeError(err) {
			p.metrics.workflowTooLarge.Inc()
			return nil, ErrWorkflowToLarge
		}
		p.metrics.workflowUpdateFailedCount.Inc()
		logger.Errorf(ctx, "Failed to update workflow status. Error [%v]", err)
		return nil, err
	}
	t.Stop()
	p.metrics.workflowUpdateSuccessCount.Inc()
	logger.Debugf(ctx, "Updated workflow status.")
	return newWF, nil
}

func (p *passthroughWorkflowStore) Update(ctx context.Context, workflow *v1alpha1.FlyteWorkflow, priorityClass PriorityClass) (
	newWF *v1alpha1.FlyteWorkflow, err error) {
	// If the workflow has any managed fields setting the array to one empty ManagedField clears them in the CRD.
	// FlyteWorkflow CRDs are only managed by a single FlytePropeller instance and therefore the managed fields paradigm
	// does not add useful functionality. Clearing them reduces CRD size, improving etcd I/O performance.
	if len(workflow.ObjectMeta.ManagedFields) > 0 {
		workflow.ObjectMeta.ManagedFields = workflow.ObjectMeta.ManagedFields[:1]
		workflow.ObjectMeta.ManagedFields[0] = v1.ManagedFieldsEntry{}
	}

	p.metrics.workflowUpdateCount.Inc()
	// Something has changed. Lets save
	logger.Debugf(ctx, "Observed FlyteWorkflow Update (maybe finalizer)")
	t := p.metrics.workflowUpdateLatency.Start()
	newWF, err = p.wfClientSet.FlyteWorkflows(workflow.Namespace).Update(ctx, workflow, v1.UpdateOptions{})
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil, nil
		}
		if kubeerrors.IsConflict(err) {
			p.metrics.workflowUpdateConflictCount.Inc()
		}
		if kubeerrors.IsRequestEntityTooLargeError(err) {
			p.metrics.workflowTooLarge.Inc()
			return nil, ErrWorkflowToLarge
		}
		p.metrics.workflowUpdateFailedCount.Inc()
		logger.Errorf(ctx, "Failed to update workflow. Error [%v]", err)
		return nil, err
	}
	t.Stop()
	p.metrics.workflowUpdateSuccessCount.Inc()
	logger.Debugf(ctx, "Updated workflow.")
	return newWF, nil
}

func NewPassthroughWorkflowStore(_ context.Context, scope promutils.Scope, wfClient v1alpha12.FlyteworkflowV1alpha1Interface,
	flyteworkflowLister listers.FlyteWorkflowLister) FlyteWorkflow {

	metrics := &workflowstoreMetrics{
		workflowUpdateCount:         scope.MustNewCounter("wf_updated", "Total number of status updates"),
		workflowUpdateFailedCount:   scope.MustNewCounter("wf_update_failed", "Failure to update ETCd"),
		workflowUpdateConflictCount: scope.MustNewCounter("wf_update_conflict", "Failure to update ETCd because of conflict"),
		workflowUpdateSuccessCount:  scope.MustNewCounter("wf_update_success", "Success in updating ETCd"),
		workflowUpdateLatency:       scope.MustNewStopWatch("wf_update_latency", "Time taken to complete update/updatestatus", time.Millisecond),
		workflowTooLarge:            scope.MustNewCounter("wf_too_large", "Failure to update ETCd because of size of the workflow is too large."),
	}

	return &passthroughWorkflowStore{
		wfLister:    flyteworkflowLister,
		wfClientSet: wfClient,
		metrics:     metrics,
	}
}
