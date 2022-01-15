package interfaces

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/executioncluster"
)

type GetTargetInterface interface {
	GetTarget(context.Context, *executioncluster.ExecutionTargetSpec) (*executioncluster.ExecutionTarget, error)
}

type ListTargetsInterface interface {
	GetAllTargets() map[string]*executioncluster.ExecutionTarget
	// Returns all enabled targets.
	GetValidTargets() map[string]*executioncluster.ExecutionTarget
}

// Interface for the Execution Cluster
type ClusterInterface interface {
	GetTargetInterface
	ListTargetsInterface
}
