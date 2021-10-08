package mocks

import (
	"github.com/flyteorg/flyteadmin/pkg/executioncluster"
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/prometheus/client_golang/prometheus"
)

type MockExecutionTargetProvider struct{}

// Creates a new Execution target for a cluster based on config passed in.
func (c *MockExecutionTargetProvider) GetExecutionTarget(_ prometheus.Counter, k8sCluster interfaces.ClusterConfig) (*executioncluster.ExecutionTarget, error) {
	return &executioncluster.ExecutionTarget{
		ID:      k8sCluster.Name,
		Enabled: k8sCluster.Enabled,
	}, nil
}
