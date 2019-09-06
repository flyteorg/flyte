package workflowstore

import (
	"context"
	"fmt"
	"sync"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/prometheus/client_golang/prometheus"
)

// TODO - optimization maybe? we can move this to predicate check, before we add it to the queue?
type resourceVersionMetrics struct {
	workflowStaleCount   prometheus.Counter
	workflowEvictedCount prometheus.Counter
}

// Simple function that covnerts the namespace and name to a string
func resourceVersionKey(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

// A specialized store that stores a inmemory cache of all the workflows that are currently executing and their last observed version numbers
// If the version numbers between the last update and the next Get have not been updated then the Get returns a nil (ignores the workflow)
// Propeller round will then just ignore the workflow
type resourceVersionCaching struct {
	w                               FlyteWorkflow
	metrics                         *resourceVersionMetrics
	lastUpdatedResourceVersionCache sync.Map
}

func (r *resourceVersionCaching) updateRevisionCache(_ context.Context, namespace, name, resourceVersion string, isTerminated bool) {
	if isTerminated {
		r.metrics.workflowEvictedCount.Inc()
		r.lastUpdatedResourceVersionCache.Delete(resourceVersionKey(namespace, name))
	} else {
		r.lastUpdatedResourceVersionCache.Store(resourceVersionKey(namespace, name), resourceVersion)
	}
}

func (r *resourceVersionCaching) isResourceVersionSameAsPrevious(namespace, name, resourceVersion string) bool {
	if v, ok := r.lastUpdatedResourceVersionCache.Load(resourceVersionKey(namespace, name)); ok {
		strV := v.(string)
		if strV == resourceVersion {
			r.metrics.workflowStaleCount.Inc()
			return true
		}
	}
	return false
}

func (r *resourceVersionCaching) Get(ctx context.Context, namespace, name string) (*v1alpha1.FlyteWorkflow, error) {
	w, err := r.w.Get(ctx, namespace, name)
	if err != nil {
		return nil, err
	}
	if w != nil {
		if r.isResourceVersionSameAsPrevious(namespace, name, w.ResourceVersion) {
			return nil, errStaleWorkflowError
		}
	}
	return w, nil
}

func (r *resourceVersionCaching) UpdateStatus(ctx context.Context, workflow *v1alpha1.FlyteWorkflow, priorityClass PriorityClass) error {
	err := r.w.UpdateStatus(ctx, workflow, priorityClass)
	if err != nil {
		return err
	}

	r.updateRevisionCache(ctx, workflow.Namespace, workflow.Name, workflow.ResourceVersion, workflow.Status.IsTerminated())
	return nil
}

func (r *resourceVersionCaching) Update(ctx context.Context, workflow *v1alpha1.FlyteWorkflow, priorityClass PriorityClass) error {
	err := r.w.Update(ctx, workflow, priorityClass)
	if err != nil {
		return err
	}

	r.updateRevisionCache(ctx, workflow.Namespace, workflow.Name, workflow.ResourceVersion, workflow.Status.IsTerminated())
	return nil
}

func NewResourceVersionCachingStore(ctx context.Context, scope promutils.Scope, workflowStore FlyteWorkflow) FlyteWorkflow {

	return &resourceVersionCaching{
		w: workflowStore,
		metrics: &resourceVersionMetrics{
			workflowStaleCount:   scope.MustNewCounter("wf_stale", "Found stale workflow in cache"),
			workflowEvictedCount: scope.MustNewCounter("wf_evict", "removed workflow from resource version cache"),
		},
		lastUpdatedResourceVersionCache: sync.Map{},
	}
}
