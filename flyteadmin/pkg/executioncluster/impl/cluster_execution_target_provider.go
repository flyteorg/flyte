package impl

import (
	"github.com/flyteorg/flyteadmin/pkg/executioncluster"
	"github.com/flyteorg/flyteadmin/pkg/executioncluster/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/flytek8s"
	runtime "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	flyteclient "github.com/flyteorg/flytepropeller/pkg/client/clientset/versioned"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type clusterExecutionTargetProvider struct{}

// Creates a new Execution target for a cluster based on config passed in.
func (c *clusterExecutionTargetProvider) GetExecutionTarget(initializationErrorCounter prometheus.Counter, k8sCluster runtime.ClusterConfig) (*executioncluster.ExecutionTarget, error) {
	kubeConf, err := flytek8s.GetRestClientConfig("", "", &k8sCluster)
	if err != nil {
		return nil, err
	}
	flyteClient, err := getRestClientFromKubeConfig(initializationErrorCounter, kubeConf)
	if err != nil {
		return nil, err
	}
	client, err := client.New(kubeConf, client.Options{})
	if err != nil {
		return nil, err
	}
	dynamicClient, err := dynamic.NewForConfig(kubeConf)
	if err != nil {
		return nil, err
	}
	return &executioncluster.ExecutionTarget{
		FlyteClient:   flyteClient,
		Client:        client,
		DynamicClient: dynamicClient,
		ID:            k8sCluster.Name,
		Enabled:       k8sCluster.Enabled,
		Config:        *kubeConf,
	}, nil
}

func getRestClientFromKubeConfig(initializationErrorCounter prometheus.Counter, kubeConfiguration *rest.Config) (*flyteclient.Clientset, error) {
	fc, err := flyteclient.NewForConfig(kubeConfiguration)
	if err != nil {
		initializationErrorCounter.Inc()
		return nil, err
	}
	return fc, nil
}

func NewExecutionTargetProvider() interfaces.ExecutionTargetProvider {
	return &clusterExecutionTargetProvider{}
}
