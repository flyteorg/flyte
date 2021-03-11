package interfaces

import (
	"github.com/flyteorg/flyteadmin/pkg/executioncluster"
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/promutils"
)

type ExecutionTargetProvider interface {
	GetExecutionTarget(scope promutils.Scope, k8sCluster interfaces.ClusterConfig) (*executioncluster.ExecutionTarget, error)
}
