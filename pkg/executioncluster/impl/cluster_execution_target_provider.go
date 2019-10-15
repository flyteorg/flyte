package impl

import (
	"github.com/lyft/flyteadmin/pkg/executioncluster"
	"github.com/lyft/flyteadmin/pkg/flytek8s"
	runtime "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	flyteclient "github.com/lyft/flytepropeller/pkg/client/clientset/versioned"
	"github.com/lyft/flytestdlib/promutils"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type clusterExecutionTargetProvider struct{}

// Creates a new Execution target for a cluster based on config passed in.
func (c *clusterExecutionTargetProvider) GetExecutionTarget(scope promutils.Scope, k8sCluster runtime.ClusterConfig) (*executioncluster.ExecutionTarget, error) {
	kubeConf, err := flytek8s.GetRestClientConfigForCluster(k8sCluster)
	if err != nil {
		return nil, err
	}
	flyteClient, err := getRestClientFromKubeConfig(scope, kubeConf)
	if err != nil {
		return nil, err
	}
	client, err := client.New(kubeConf, client.Options{})
	if err != nil {
		return nil, err
	}
	return &executioncluster.ExecutionTarget{
		FlyteClient: flyteClient,
		Client:      client,
		ID:          k8sCluster.Name,
		Enabled:     k8sCluster.Enabled,
	}, nil
}

func getRestClientFromKubeConfig(scope promutils.Scope, kubeConfiguration *rest.Config) (*flyteclient.Clientset, error) {
	fc, err := flyteclient.NewForConfig(kubeConfiguration)
	if err != nil {
		scope.MustNewCounter(
			"flyteclient_initialization_error",
			"count of errors encountered initializing a flyte client from kube config").Inc()
		return nil, err
	}
	return fc, nil
}
