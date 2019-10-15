package interfaces

import (
	"github.com/lyft/flyteadmin/pkg/executioncluster"
)

// Interface for the Execution Cluster
type ClusterInterface interface {
	GetTarget(*executioncluster.ExecutionTargetSpec) (*executioncluster.ExecutionTarget, error)
	GetAllValidTargets() []executioncluster.ExecutionTarget
}
