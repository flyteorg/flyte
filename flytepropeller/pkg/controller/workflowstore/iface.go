package workflowstore

import (
	"context"

	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)


// FlyteWorkflow store interface provides an abstraction of accessing the actual FlyteWorkflow object.
type FlyteWorkflow interface {
	Get(ctx context.Context, namespace, name string) (*v1alpha1.FlyteWorkflow, error)
	Update(ctx context.Context, workflow *v1alpha1.FlyteWorkflow) (newWF *v1alpha1.FlyteWorkflow, err error)
}
