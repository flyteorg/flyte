package workflowstore

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytestdlib/fastcheck"

	"github.com/flyteorg/flytestdlib/promutils"
)

func workflowKey(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

// A specialized store that stores a LRU cache of all the workflows that are in a terminal phase.
// Terminated workflows are ignored (Get returns a nil).
// Processing terminated FlyteWorkflows can occur when workflow updates are reported after a workflow has already completed.
type terminatedTracking struct {
	w                FlyteWorkflow
	terminatedFilter fastcheck.Filter
}

func (t *terminatedTracking) Get(ctx context.Context, namespace, name string) (*v1alpha1.FlyteWorkflow, error) {
	if t.terminatedFilter.Contains(ctx, []byte(workflowKey(namespace, name))) {
		return nil, ErrWorkflowTerminated
	}

	return t.w.Get(ctx, namespace, name)
}

func (t *terminatedTracking) UpdateStatus(ctx context.Context, workflow *v1alpha1.FlyteWorkflow, priorityClass PriorityClass) (
	newWF *v1alpha1.FlyteWorkflow, err error) {
	newWF, err = t.w.UpdateStatus(ctx, workflow, priorityClass)
	if err != nil {
		return nil, err
	}

	if newWF != nil {
		if newWF.GetExecutionStatus().IsTerminated() {
			t.terminatedFilter.Add(ctx, []byte(workflowKey(workflow.Namespace, workflow.Name)))
		}
	}

	return newWF, nil
}

func (t *terminatedTracking) Update(ctx context.Context, workflow *v1alpha1.FlyteWorkflow, priorityClass PriorityClass) (
	newWF *v1alpha1.FlyteWorkflow, err error) {
	newWF, err = t.w.Update(ctx, workflow, priorityClass)
	if err != nil {
		return nil, err
	}

	if newWF != nil {
		if newWF.GetExecutionStatus().IsTerminated() {
			t.terminatedFilter.Add(ctx, []byte(workflowKey(workflow.Namespace, workflow.Name)))
		}
	}

	return newWF, nil
}

func NewTerminatedTrackingStore(_ context.Context, scope promutils.Scope, workflowStore FlyteWorkflow) (FlyteWorkflow, error) {
	filter, err := fastcheck.NewLRUCacheFilter(1000, scope.NewSubScope("terminated_filter"))
	if err != nil {
		return nil, err
	}

	return &terminatedTracking{
		w:                workflowStore,
		terminatedFilter: filter,
	}, nil
}
