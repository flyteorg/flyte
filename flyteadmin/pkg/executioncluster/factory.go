package executioncluster

import (
	"github.com/lyft/flytestdlib/promutils"

	runtime "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
)

type ClusterInterface interface {
	GetTarget(*ExecutionTargetSpec) (*ExecutionTarget, error)
	GetAllValidTargets() []ExecutionTarget
}

func GetEnabledClusters(clusterConfig runtime.ClusterConfiguration) []runtime.ClusterConfig {
	enabledClusters := make([]runtime.ClusterConfig, 0)
	for _, cluster := range clusterConfig.GetClusterConfigs() {
		if cluster.Enabled {
			enabledClusters = append(enabledClusters, cluster)
		}
	}
	return enabledClusters
}

func GetExecutionCluster(scope promutils.Scope, kubeConfig, master string, clusterConfig runtime.ClusterConfiguration) ClusterInterface {
	enabledClusters := GetEnabledClusters(clusterConfig)
	switch len(enabledClusters) {
	case 0:
		cluster, err := NewInCluster(scope, kubeConfig, master)
		if err != nil {
			panic(err)
		}
		return cluster
	default:
		cluster, err := NewRandomClusterSelector(scope, clusterConfig)
		if err != nil {
			panic(err)
		}
		return cluster
	}
}
