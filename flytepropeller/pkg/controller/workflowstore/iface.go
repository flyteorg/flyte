package workflowstore

import (
	"context"

	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

//go:generate mockery-v2 --all --with-expecter

type PriorityClass int

const (
	PriorityClassCritical PriorityClass = iota
	PriorityClassRegular
)

// FlyteWorkflow store interface provides an abstraction of accessing the actual FlyteWorkflow object.
type FlyteWorkflow interface {
	Get(ctx context.Context, namespace, name string) (*v1alpha1.FlyteWorkflow, error)
	UpdateStatus(ctx context.Context, workflow *v1alpha1.FlyteWorkflow, priorityClass PriorityClass) (
		newWF *v1alpha1.FlyteWorkflow, err error)
	Update(ctx context.Context, workflow *v1alpha1.FlyteWorkflow, priorityClass PriorityClass) (
		newWF *v1alpha1.FlyteWorkflow, err error)
}
