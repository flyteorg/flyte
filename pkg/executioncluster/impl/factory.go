package impl

import (
	executioncluster_interface "github.com/flyteorg/flyteadmin/pkg/executioncluster/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories"
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/promutils"
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
		cluster, err := NewRandomClusterSelector(scope, config, &clusterExecutionTargetProvider{}, db)
		if err != nil {
			panic(err)
		}
		return cluster
	}
}
