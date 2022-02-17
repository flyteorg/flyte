package impl

import (
	executioncluster_interface "github.com/flyteorg/flyteadmin/pkg/executioncluster/interfaces"
	repositoryInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/promutils"
)

func GetExecutionCluster(scope promutils.Scope, kubeConfig, master string, config interfaces.Configuration, db repositoryInterfaces.Repository) executioncluster_interface.ClusterInterface {
	initializationErrorCounter := scope.MustNewCounter(
		"flyteclient_initialization_error",
		"count of errors encountered initializing a flyte client from kube config")
	switch len(config.ClusterConfiguration().GetClusterConfigs()) {
	case 0:
		cluster, err := NewInCluster(initializationErrorCounter, kubeConfig, master)
		if err != nil {
			panic(err)
		}
		return cluster
	default:
		listTargetsProvider, err := NewListTargets(initializationErrorCounter, NewExecutionTargetProvider(), config.ClusterConfiguration())
		if err != nil {
			panic(err)
		}
		cluster, err := NewRandomClusterSelector(listTargetsProvider, config, db)
		if err != nil {
			panic(err)
		}
		return cluster
	}
}
