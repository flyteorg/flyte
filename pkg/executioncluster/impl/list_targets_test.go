package impl

import (
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/executioncluster"
	"github.com/flyteorg/flyteadmin/pkg/executioncluster/mocks"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	runtimeMocks "github.com/flyteorg/flyteadmin/pkg/runtime/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewListTargets(t *testing.T) {
	enabledCluster := "EC"
	disabledCluster := "DC"
	execTargetProvider := mocks.ExecutionTargetProvider{}
	execTargetProvider.OnGetExecutionTargetMatch(mock.Anything,
		mock.MatchedBy(func(cluster runtimeInterfaces.ClusterConfig) bool {
			return cluster.Name == enabledCluster
		})).Return(
		&executioncluster.ExecutionTarget{
			Enabled: true,
			ID:      enabledCluster,
		}, nil)
	execTargetProvider.OnGetExecutionTargetMatch(mock.Anything,
		mock.MatchedBy(func(cluster runtimeInterfaces.ClusterConfig) bool {
			return cluster.Name == disabledCluster
		})).Return(
		&executioncluster.ExecutionTarget{
			Enabled: false,
			ID:      disabledCluster,
		}, nil)
	conf := runtimeMocks.ClusterConfiguration{}
	conf.OnGetClusterConfigs().Return([]runtimeInterfaces.ClusterConfig{
		{
			Name:    enabledCluster,
			Enabled: true,
		},
		{
			Name:    disabledCluster,
			Enabled: false,
		},
	})
	listTargetsProvider, err := NewListTargets(
		nil,
		&execTargetProvider, &conf)
	assert.NoError(t, err)
	validTargets := listTargetsProvider.GetValidTargets()
	assert.Len(t, validTargets, 1)
	assert.Contains(t, validTargets, enabledCluster)

	allTargets := listTargetsProvider.GetAllTargets()
	assert.Len(t, allTargets, 2)
	assert.Contains(t, allTargets, enabledCluster)
	assert.Contains(t, allTargets, disabledCluster)
}
