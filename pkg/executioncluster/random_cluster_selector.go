package executioncluster

import (
	"fmt"

	runtime "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flytestdlib/promutils"
	"k8s.io/apimachinery/pkg/util/rand"
)

type RandomClusterSelector struct {
	executionTargetMap       map[string]ExecutionTarget
	totalEnabledClusterCount int
}

func getExecutionTargetMap(scope promutils.Scope, clusterConfig runtime.ClusterConfiguration) (map[string]ExecutionTarget, error) {
	executionTargetMap := make(map[string]ExecutionTarget)
	for _, cluster := range clusterConfig.GetClusterConfigs() {
		if _, ok := executionTargetMap[cluster.Name]; ok {
			return nil, fmt.Errorf("duplicate clusters for name %s", cluster.Name)
		}
		executionTarget, err := NewExecutionTarget(scope, cluster)
		if err != nil {
			return nil, err
		}
		executionTargetMap[cluster.Name] = *executionTarget
	}
	return executionTargetMap, nil
}

func (s RandomClusterSelector) GetAllValidTargets() []ExecutionTarget {
	v := make([]ExecutionTarget, 0, len(s.executionTargetMap))
	for _, value := range s.executionTargetMap {
		if value.Enabled {
			v = append(v, value)
		}
	}
	return v
}

func (s RandomClusterSelector) GetTarget(spec *ExecutionTargetSpec) (*ExecutionTarget, error) {
	if spec != nil && spec.TargetID != "" {
		if val, ok := s.executionTargetMap[spec.TargetID]; ok {
			return &val, nil
		}
		return nil, fmt.Errorf("invalid cluster target %s", spec.TargetID)
	}
	targetIdx := rand.Intn(s.totalEnabledClusterCount)
	index := 0
	for _, val := range s.executionTargetMap {
		if val.Enabled {
			if index == targetIdx {
				return &val, nil
			}
			index++
		}
	}
	return nil, nil
}

func NewRandomClusterSelector(scope promutils.Scope, clusterConfig runtime.ClusterConfiguration) (ClusterInterface, error) {
	executionTargetMap, err := getExecutionTargetMap(scope, clusterConfig)
	if err != nil {
		return nil, err
	}
	enabledClusters := GetEnabledClusters(clusterConfig)
	return &RandomClusterSelector{
		executionTargetMap:       executionTargetMap,
		totalEnabledClusterCount: len(enabledClusters),
	}, nil
}
