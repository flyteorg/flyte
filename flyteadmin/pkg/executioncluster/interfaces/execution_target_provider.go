package interfaces

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/flyteorg/flyte/flyteadmin/pkg/executioncluster"
	"github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
)

//go:generate mockery --all --case=underscore --output=../mocks --case=underscore --with-expecter

type ExecutionTargetProvider interface {
	GetExecutionTarget(initializationErrorCounter prometheus.Counter, k8sCluster interfaces.ClusterConfig) (*executioncluster.ExecutionTarget, error)
}
