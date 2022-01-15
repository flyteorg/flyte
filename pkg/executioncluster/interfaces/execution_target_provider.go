package interfaces

import (
	"github.com/flyteorg/flyteadmin/pkg/executioncluster"
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/prometheus/client_golang/prometheus"
)

//go:generate mockery -all -case=underscore -output=../mocks -case=underscore

type ExecutionTargetProvider interface {
	GetExecutionTarget(initializationErrorCounter prometheus.Counter, k8sCluster interfaces.ClusterConfig) (*executioncluster.ExecutionTarget, error)
}
