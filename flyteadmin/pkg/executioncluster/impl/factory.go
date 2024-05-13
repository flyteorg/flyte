package impl

import (
	executioncluster_interface "github.com/flyteorg/flyte/flyteadmin/pkg/executioncluster/interfaces"
	managerInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

func GetExecutionCluster(scope promutils.Scope, kubeConfig, master string, config interfaces.Configuration, resourceManager managerInterfaces.ResourceInterface) executioncluster_interface.ClusterInterface {
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
		cluster, err := NewRandomClusterSelector(listTargetsProvider, config, resourceManager)
		if err != nil {
			panic(err)
		}
		return cluster
	}
}
