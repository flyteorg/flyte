package executioncluster

import (
	"fmt"

	"github.com/lyft/flyteadmin/pkg/flytek8s"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type InCluster struct {
	target ExecutionTarget
}

func (i InCluster) GetTarget(spec *ExecutionTargetSpec) (*ExecutionTarget, error) {
	if spec != nil && spec.TargetID != "" {
		return nil, errors.New(fmt.Sprintf("remote target %s is not supported", spec.TargetID))
	}
	return &i.target, nil
}

func (i InCluster) GetAllValidTargets() []ExecutionTarget {
	return []ExecutionTarget{
		i.target,
	}
}

func NewInCluster(scope promutils.Scope, kubeConfig, master string) (ClusterInterface, error) {
	clientConfig, err := flytek8s.GetRestClientConfig(kubeConfig, master, nil)
	if err != nil {
		return nil, err
	}
	flyteClient, err := getRestClientFromKubeConfig(scope, clientConfig)
	if err != nil {
		return nil, err
	}
	client, err := client.New(clientConfig, client.Options{})
	if err != nil {
		return nil, err
	}
	return &InCluster{
		target: ExecutionTarget{
			Client:      client,
			FlyteClient: flyteClient,
		},
	}, nil
}
