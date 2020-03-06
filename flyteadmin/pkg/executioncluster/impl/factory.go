package impl

import (
	executioncluster_interface "github.com/lyft/flyteadmin/pkg/executioncluster/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories"
	"github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flytestdlib/promutils"
)

func GetExecutionCluster(scope promutils.Scope, kubeConfig, master string, config interfaces.Configuration, db repositories.RepositoryInterface) executioncluster_interface.ClusterInterface {
	switch len(config.ClusterConfiguration().GetClusterConfigs()) {
	case 0:
		cluster, err := NewInCluster(scope, kubeConfig, master)
		if err != nil {
			panic(err)
		}
		return cluster
	default:
		cluster, err := NewRandomClusterSelector(scope, config.ClusterConfiguration(), &clusterExecutionTargetProvider{}, db)
		if err != nil {
			panic(err)
		}
		return cluster
	}
}
