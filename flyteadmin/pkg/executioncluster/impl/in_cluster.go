package impl

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/flyteadmin/pkg/executioncluster"
	"github.com/flyteorg/flyte/flyteadmin/pkg/executioncluster/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/flytek8s"
)

// DO NOT USE: only for backwards compatibility
const defaultInClusterTargetID = "id"

type InCluster struct {
	target    executioncluster.ExecutionTarget
	asTargets map[string]*executioncluster.ExecutionTarget
}

func (i InCluster) GetTarget(ctx context.Context, spec *executioncluster.ExecutionTargetSpec) (*executioncluster.ExecutionTarget, error) {
	if spec != nil && !(spec.TargetID == "" || spec.TargetID == defaultInClusterTargetID) {
		return nil, errors.New(fmt.Sprintf("remote target %s is not supported", spec.TargetID))
	}
	if spec != nil && spec.ExecutionClusterLabel != nil && spec.ExecutionClusterLabel.Value != "" {
		return nil, errors.New(fmt.Sprintf("execution cluster label %s is not supported", spec.ExecutionClusterLabel.Value))
	}
	return &i.target, nil
}

func (i InCluster) GetAllTargets() map[string]*executioncluster.ExecutionTarget {
	return i.asTargets
}

func (i InCluster) GetValidTargets() map[string]*executioncluster.ExecutionTarget {
	return i.asTargets
}

func NewInCluster(initializationErrorCounter prometheus.Counter, kubeConfig, master string) (interfaces.ClusterInterface, error) {
	clientConfig, err := flytek8s.GetRestClientConfig(kubeConfig, master, nil)
	if err != nil {
		return nil, err
	}
	flyteClient, err := getRestClientFromKubeConfig(initializationErrorCounter, clientConfig)
	if err != nil {
		return nil, err
	}
	kubeClient, err := client.New(clientConfig, client.Options{})
	if err != nil {
		return nil, err
	}
	dynamicClient, err := dynamic.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}
	target := executioncluster.ExecutionTarget{
		Client:        kubeClient,
		FlyteClient:   flyteClient,
		DynamicClient: dynamicClient,
		Config:        *clientConfig,
	}
	return &InCluster{
		target: target,
		asTargets: map[string]*executioncluster.ExecutionTarget{
			target.ID: &target,
		},
	}, nil
}
