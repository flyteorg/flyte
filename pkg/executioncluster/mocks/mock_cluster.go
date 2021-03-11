package mocks

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/executioncluster"
)

type GetTargetFunc func(context.Context, *executioncluster.ExecutionTargetSpec) (*executioncluster.ExecutionTarget, error)
type GetAllValidTargetsFunc func() []executioncluster.ExecutionTarget

type MockCluster struct {
	getTargetFunc          GetTargetFunc
	getAllValidTargetsFunc GetAllValidTargetsFunc
}

func (m *MockCluster) SetGetTargetCallback(getTargetFunc GetTargetFunc) {
	m.getTargetFunc = getTargetFunc
}

func (m *MockCluster) SetGetAllValidTargetsCallback(getAllValidTargetsFunc GetAllValidTargetsFunc) {
	m.getAllValidTargetsFunc = getAllValidTargetsFunc
}

func (m *MockCluster) GetTarget(ctx context.Context, execCluster *executioncluster.ExecutionTargetSpec) (*executioncluster.ExecutionTarget, error) {
	if m.getTargetFunc != nil {
		return m.getTargetFunc(ctx, execCluster)
	}
	return nil, nil
}

func (m *MockCluster) GetAllValidTargets() []executioncluster.ExecutionTarget {
	if m.getAllValidTargetsFunc != nil {
		return m.getAllValidTargetsFunc()
	}
	return nil
}
