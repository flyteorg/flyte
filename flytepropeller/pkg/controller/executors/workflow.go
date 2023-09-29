package executors

import (
	"context"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

//go:generate mockery -name Workflow -case=underscore

type Workflow interface {
	Initialize(ctx context.Context) error
	HandleFlyteWorkflow(ctx context.Context, w *v1alpha1.FlyteWorkflow) error
	HandleAbortedWorkflow(ctx context.Context, w *v1alpha1.FlyteWorkflow, maxRetries uint32) error
}
