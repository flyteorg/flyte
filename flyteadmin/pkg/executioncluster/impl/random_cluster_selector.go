package impl

import (
	"context"
	"fmt"
	"hash/fnv"
	"math/rand"

	"github.com/lyft/flyteadmin/pkg/executioncluster"
	"github.com/lyft/flyteadmin/pkg/executioncluster/interfaces"

	"github.com/lyft/flytestdlib/random"

	runtime "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flytestdlib/promutils"
)

// Implementation of Random cluster selector
// Selects cluster based on weights and domains.
type RandomClusterSelector struct {
	domainWeightedRandomMap map[string]random.WeightedRandomList
	executionTargetMap      map[string]executioncluster.ExecutionTarget
}

func getRandSource(seed string) (rand.Source, error) {
	h := fnv.New64a()
	_, err := h.Write([]byte(seed))
	if err != nil {
		return nil, err
	}
	hashedSeed := int64(h.Sum64())
	return rand.NewSource(hashedSeed), nil
}

func getValidDomainMap(validDomains runtime.DomainsConfig) map[string]runtime.Domain {
	domainMap := make(map[string]runtime.Domain)
	for _, domain := range validDomains {
		domainMap[domain.ID] = domain
	}
	return domainMap
}

func getExecutionTargetMap(scope promutils.Scope, executionTargetProvider interfaces.ExecutionTargetProvider, clusterConfig runtime.ClusterConfiguration) (map[string]executioncluster.ExecutionTarget, error) {
	executionTargetMap := make(map[string]executioncluster.ExecutionTarget)
	for _, cluster := range clusterConfig.GetClusterConfigs() {
		if _, ok := executionTargetMap[cluster.Name]; ok {
			return nil, fmt.Errorf("duplicate clusters for name %s", cluster.Name)
		}
		executionTarget, err := executionTargetProvider.GetExecutionTarget(scope, cluster)
		if err != nil {
			return nil, err
		}
		executionTargetMap[cluster.Name] = *executionTarget
	}
	return executionTargetMap, nil
}

func getDomainsForCluster(cluster runtime.ClusterConfig, domainMap map[string]runtime.Domain) ([]string, error) {
	if len(cluster.AllowedDomains) == 0 {
		allDomains := make([]string, len(domainMap))
		index := 0
		for id := range domainMap {
			allDomains[index] = id
			index++
		}
		return allDomains, nil
	}
	for _, allowedDomain := range cluster.AllowedDomains {
		if _, ok := domainMap[allowedDomain]; !ok {
			return nil, fmt.Errorf("invalid domain %s", allowedDomain)
		}
	}
	return cluster.AllowedDomains, nil
}

func getDomainWeightedRandomForCluster(ctx context.Context, scope promutils.Scope, executionTargetProvider interfaces.ExecutionTargetProvider,
	clusterConfig runtime.ClusterConfiguration,
	domainMap map[string]runtime.Domain) (map[string]random.WeightedRandomList, error) {
	domainEntriesMap := make(map[string][]random.Entry)
	for _, cluster := range clusterConfig.GetClusterConfigs() {
		// If cluster is not enabled, it is not eligible for selection
		if !cluster.Enabled {
			continue
		}
		executionTarget, err := executionTargetProvider.GetExecutionTarget(scope, cluster)
		if err != nil {
			return nil, err
		}
		targetEntry := random.Entry{
			Item:   *executionTarget,
			Weight: cluster.Weight,
		}
		clusterDomains, err := getDomainsForCluster(cluster, domainMap)
		if err != nil {
			return nil, err
		}
		for _, domain := range clusterDomains {
			if _, ok := domainEntriesMap[domain]; ok {
				domainEntriesMap[domain] = append(domainEntriesMap[domain], targetEntry)
			} else {
				domainEntriesMap[domain] = []random.Entry{targetEntry}
			}
		}
	}
	domainWeightedRandomMap := make(map[string]random.WeightedRandomList)
	for domain, entries := range domainEntriesMap {
		weightedRandomList, err := random.NewWeightedRandom(ctx, entries)
		if err != nil {
			return nil, err
		}
		domainWeightedRandomMap[domain] = weightedRandomList
	}
	return domainWeightedRandomMap, nil
}

func (s RandomClusterSelector) GetAllValidTargets() []executioncluster.ExecutionTarget {
	v := make([]executioncluster.ExecutionTarget, 0)
	for _, value := range s.executionTargetMap {
		if value.Enabled {
			v = append(v, value)
		}
	}
	return v
}

func (s RandomClusterSelector) GetTarget(spec *executioncluster.ExecutionTargetSpec) (*executioncluster.ExecutionTarget, error) {
	if spec == nil || (spec.TargetID == "" && spec.ExecutionID == nil) {
		return nil, fmt.Errorf("invalid executionTargetSpec %v", spec)
	}
	if spec.TargetID != "" {
		if val, ok := s.executionTargetMap[spec.TargetID]; ok {
			return &val, nil
		}
		return nil, fmt.Errorf("invalid cluster target %s", spec.TargetID)
	}
	if spec.ExecutionID != nil {
		if weightedRandomList, ok := s.domainWeightedRandomMap[spec.ExecutionID.GetDomain()]; ok {
			executionName := spec.ExecutionID.GetName()
			if executionName != "" {
				randSrc, err := getRandSource(executionName)
				if err != nil {
					return nil, err
				}
				result, err := weightedRandomList.GetWithSeed(randSrc)
				if err != nil {
					return nil, err
				}
				execTarget := result.(executioncluster.ExecutionTarget)
				return &execTarget, nil
			}
			execTarget := weightedRandomList.Get().(executioncluster.ExecutionTarget)
			return &execTarget, nil
		}
	}
	return nil, fmt.Errorf("invalid executionTargetSpec %v", *spec)

}

func NewRandomClusterSelector(scope promutils.Scope, clusterConfig runtime.ClusterConfiguration, executionTargetProvider interfaces.ExecutionTargetProvider, domainConfig *runtime.DomainsConfig) (interfaces.ClusterInterface, error) {
	executionTargetMap, err := getExecutionTargetMap(scope, executionTargetProvider, clusterConfig)
	if err != nil {
		return nil, err
	}
	domainMap := getValidDomainMap(*domainConfig)
	domainWeightedRandomMap, err := getDomainWeightedRandomForCluster(context.Background(), scope, executionTargetProvider, clusterConfig, domainMap)
	if err != nil {
		return nil, err
	}
	return &RandomClusterSelector{
		domainWeightedRandomMap: domainWeightedRandomMap,
		executionTargetMap:      executionTargetMap,
	}, nil
}
