package mocks

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/executioncluster"
)

type GetTargetFunc func(context.Context, *executioncluster.ExecutionTargetSpec) (*executioncluster.ExecutionTarget, error)
type GetAllValidTargetsFunc func() map[string]*executioncluster.ExecutionTarget

type MockCluster struct {
	getTargetFunc          GetTargetFunc
	getAllTargetsFunc      GetAllValidTargetsFunc
	getAllValidTargetsFunc GetAllValidTargetsFunc
}

func (m *MockCluster) SetGetTargetCallback(getTargetFunc GetTargetFunc) {
	m.getTargetFunc = getTargetFunc
}

func (m *MockCluster) SetGetAllValidTargetsCallback(getAllValidTargetsFunc GetAllValidTargetsFunc) {
	m.getAllValidTargetsFunc = getAllValidTargetsFunc
}

func (m *MockCluster) SetGetAllTargetsCallback(getAllTargetsFunc GetAllValidTargetsFunc) {
	m.getAllTargetsFunc = getAllTargetsFunc
}

func (m *MockCluster) GetTarget(ctx context.Context, execCluster *executioncluster.ExecutionTargetSpec) (*executioncluster.ExecutionTarget, error) {
	if m.getTargetFunc != nil {
		return m.getTargetFunc(ctx, execCluster)
	}
	return nil, nil
}

func (m *MockCluster) GetValidTargets() map[string]*executioncluster.ExecutionTarget {
	if m.getAllValidTargetsFunc != nil {
		return m.getAllValidTargetsFunc()
	}
	return nil
}

func (m *MockCluster) GetAllTargets() map[string]*executioncluster.ExecutionTarget {
	if m.getAllTargetsFunc != nil {
		return m.getAllTargetsFunc()
	}
	return nil
}
