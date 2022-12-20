package impl

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/flyteorg/flyteadmin/pkg/executioncluster"
	"github.com/flyteorg/flyteadmin/pkg/executioncluster/interfaces"
	runtime "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
)

type listTargets struct {
	allTargets   map[string]*executioncluster.ExecutionTarget
	validTargets map[string]*executioncluster.ExecutionTarget
}

func (l *listTargets) GetAllTargets() map[string]*executioncluster.ExecutionTarget {
	return l.allTargets
}

func (l *listTargets) GetValidTargets() map[string]*executioncluster.ExecutionTarget {
	return l.validTargets
}

func NewListTargets(initializationErrorCounter prometheus.Counter, executionTargetProvider interfaces.ExecutionTargetProvider,
	clusterConfig runtime.ClusterConfiguration) (interfaces.ListTargetsInterface, error) {
	allTargets := make(map[string]*executioncluster.ExecutionTarget)
	validTargets := make(map[string]*executioncluster.ExecutionTarget)

	for _, cluster := range clusterConfig.GetClusterConfigs() {
		if _, ok := allTargets[cluster.Name]; ok {
			return nil, fmt.Errorf("duplicate clusters for name %s", cluster.Name)
		}
		executionTarget, err := executionTargetProvider.GetExecutionTarget(initializationErrorCounter, cluster)
		if err != nil {
			return nil, err
		}
		allTargets[cluster.Name] = executionTarget
		if executionTarget.Enabled {
			validTargets[cluster.Name] = executionTarget
		}
	}
	return &listTargets{
		allTargets:   allTargets,
		validTargets: validTargets,
	}, nil
}
