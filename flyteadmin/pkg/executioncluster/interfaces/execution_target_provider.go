package interfaces

import (
	"github.com/lyft/flyteadmin/pkg/executioncluster"
	"github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flytestdlib/promutils"
)

type ExecutionTargetProvider interface {
	GetExecutionTarget(scope promutils.Scope, k8sCluster interfaces.ClusterConfig) (*executioncluster.ExecutionTarget, error)
}
