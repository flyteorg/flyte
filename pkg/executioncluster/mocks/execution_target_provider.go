package mocks

import (
	"github.com/flyteorg/flyteadmin/pkg/executioncluster"
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/promutils"
)

type MockExecutionTargetProvider struct{}

// Creates a new Execution target for a cluster based on config passed in.
func (c *MockExecutionTargetProvider) GetExecutionTarget(scope promutils.Scope, k8sCluster interfaces.ClusterConfig) (*executioncluster.ExecutionTarget, error) {
	return &executioncluster.ExecutionTarget{
		ID:      k8sCluster.Name,
		Enabled: k8sCluster.Enabled,
	}, nil
}
