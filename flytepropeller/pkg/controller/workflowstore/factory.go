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

	switch cfg.Policy {
	case PolicyInMemory:
		return NewInMemoryWorkflowStore(), nil
	case PolicyPassThrough:
		return NewPassthroughWorkflowStore(ctx, scope, workflows, lister), nil
	case PolicyResourceVersionCache:
		return NewResourceVersionCachingStore(ctx, scope, NewPassthroughWorkflowStore(ctx, scope, workflows, lister)), nil
	}

	return nil, fmt.Errorf("empty workflow store config")
}
