package workflowstore

import (
	"context"
	"fmt"

	flyteworkflowv1alpha1 "github.com/flyteorg/flytepropeller/pkg/client/clientset/versioned/typed/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/client/listers/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytestdlib/promutils"
)

func NewWorkflowStore(ctx context.Context, cfg *Config, lister v1alpha1.FlyteWorkflowLister,
	workflows flyteworkflowv1alpha1.FlyteworkflowV1alpha1Interface, scope promutils.Scope) (FlyteWorkflow, error) {

	var workflowStore FlyteWorkflow
	var err error

	switch cfg.Policy {
	case PolicyInMemory:
		workflowStore = NewInMemoryWorkflowStore()
	case PolicyPassThrough:
		workflowStore = NewPassthroughWorkflowStore(ctx, scope, workflows, lister)
	case PolicyTrackTerminated:
		workflowStore = NewPassthroughWorkflowStore(ctx, scope, workflows, lister)
		workflowStore, err = NewTerminatedTrackingStore(ctx, scope, workflowStore)
	case PolicyResourceVersionCache:
		workflowStore = NewPassthroughWorkflowStore(ctx, scope, workflows, lister)
		workflowStore, err = NewTerminatedTrackingStore(ctx, scope, workflowStore)
		workflowStore = NewResourceVersionCachingStore(ctx, scope, workflowStore)
	}

	if err != nil {
		return nil, err
	} else if workflowStore == nil {
		return nil, fmt.Errorf("empty workflow store config")
	}

	return workflowStore, err
}
