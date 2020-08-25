package impl

import (
	"context"
	"fmt"
	"hash/fnv"
	"math/rand"

	"github.com/lyft/flytestdlib/logger"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/manager/impl/resources"
	managerInterfaces "github.com/lyft/flyteadmin/pkg/manager/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories"
	"google.golang.org/grpc/codes"

	"github.com/lyft/flyteadmin/pkg/executioncluster"
	"github.com/lyft/flyteadmin/pkg/executioncluster/interfaces"

	"github.com/lyft/flytestdlib/random"

	runtime "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flytestdlib/promutils"
)

// Implementation of Random cluster selector
// Selects cluster based on weights and domains.
type RandomClusterSelector struct {
	equalWeightedAllClusters random.WeightedRandomList
	labelWeightedRandomMap   map[string]random.WeightedRandomList
	executionTargetMap       map[string]executioncluster.ExecutionTarget
	resourceManager          managerInterfaces.ResourceInterface
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

func getExecutionTargets(ctx context.Context, scope promutils.Scope, executionTargetProvider interfaces.ExecutionTargetProvider,
	clusterConfig runtime.ClusterConfiguration) (random.WeightedRandomList, map[string]executioncluster.ExecutionTarget, error) {
	executionTargetMap := make(map[string]executioncluster.ExecutionTarget)
	entries := make([]random.Entry, 0)
	for _, cluster := range clusterConfig.GetClusterConfigs() {
		if _, ok := executionTargetMap[cluster.Name]; ok {
			return nil, nil, fmt.Errorf("duplicate clusters for name %s", cluster.Name)
		}
		executionTarget, err := executionTargetProvider.GetExecutionTarget(scope, cluster)
		if err != nil {
			return nil, nil, err
		}
		executionTargetMap[cluster.Name] = *executionTarget
		if executionTarget.Enabled {
			targetEntry := random.Entry{
				Item: *executionTarget,
			}
			entries = append(entries, targetEntry)
		}
	}
	weightedRandomList, err := random.NewWeightedRandom(ctx, entries)
	if err != nil {
		return nil, nil, err
	}
	return weightedRandomList, executionTargetMap, nil
}

func getLabeledWeightedRandomForCluster(ctx context.Context,
	clusterConfig runtime.ClusterConfiguration, executionTargetMap map[string]executioncluster.ExecutionTarget) (map[string]random.WeightedRandomList, error) {
	labeledWeightedRandomMap := make(map[string]random.WeightedRandomList)
	for label, clusterEntities := range clusterConfig.GetLabelClusterMap() {
		entries := make([]random.Entry, 0)
		for _, clusterEntity := range clusterEntities {
			cluster := executionTargetMap[clusterEntity.ID]
			// If cluster is not enabled, it is not eligible for selection
			if !cluster.Enabled {
				continue
			}
			targetEntry := random.Entry{
				Item:   cluster,
				Weight: clusterEntity.Weight,
			}
			entries = append(entries, targetEntry)
		}
		if len(entries) > 0 {
			weightedRandomList, err := random.NewWeightedRandom(ctx, entries)
			if err != nil {
				return nil, err
			}
			labeledWeightedRandomMap[label] = weightedRandomList
		}
	}
	return labeledWeightedRandomMap, nil
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

func (s RandomClusterSelector) GetTarget(ctx context.Context, spec *executioncluster.ExecutionTargetSpec) (*executioncluster.ExecutionTarget, error) {
	if spec == nil {
		return nil, fmt.Errorf("empty executionTargetSpec")
	}
	if spec.TargetID != "" {
		if val, ok := s.executionTargetMap[spec.TargetID]; ok {
			return &val, nil
		}
		return nil, fmt.Errorf("invalid cluster target %s", spec.TargetID)
	}
	resource, err := s.resourceManager.GetResource(ctx, managerInterfaces.ResourceRequest{
		Project:      spec.Project,
		Domain:       spec.Domain,
		Workflow:     spec.Workflow,
		LaunchPlan:   spec.LaunchPlan,
		ResourceType: admin.MatchableResource_EXECUTION_CLUSTER_LABEL,
	})
	if err != nil {
		if flyteAdminError, ok := err.(errors.FlyteAdminError); !ok || flyteAdminError.Code() != codes.NotFound {
			return nil, err
		}
	}
	var weightedRandomList random.WeightedRandomList
	if resource != nil && resource.Attributes.GetExecutionClusterLabel() != nil {
		label := resource.Attributes.GetExecutionClusterLabel().Value

		if _, ok := s.labelWeightedRandomMap[label]; ok {
			weightedRandomList = s.labelWeightedRandomMap[label]
		} else {
			logger.Debugf(ctx, "No cluster mapping found for the label %s", label)
		}
	} else {
		logger.Debugf(ctx, "No override found for the spec %v", spec)
	}
	// If there is no label associated (or) if the label is invalid, choose from all enabled clusters.
	// Note that if there is a valid label with zero "Enabled" clusters, we still choose from all enabled ones.
	if weightedRandomList == nil {
		weightedRandomList = s.equalWeightedAllClusters
	}

	executionName := spec.ExecutionID
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

func NewRandomClusterSelector(scope promutils.Scope, config runtime.Configuration, executionTargetProvider interfaces.ExecutionTargetProvider, db repositories.RepositoryInterface) (interfaces.ClusterInterface, error) {
	equalWeightedAllClusters, executionTargetMap, err := getExecutionTargets(context.Background(), scope, executionTargetProvider, config.ClusterConfiguration())
	if err != nil {
		return nil, err
	}
	labelWeightedRandomMap, err := getLabeledWeightedRandomForCluster(context.Background(), config.ClusterConfiguration(), executionTargetMap)
	if err != nil {
		return nil, err
	}
	return &RandomClusterSelector{
		labelWeightedRandomMap:   labelWeightedRandomMap,
		executionTargetMap:       executionTargetMap,
		resourceManager:          resources.NewResourceManager(db, config.ApplicationConfiguration()),
		equalWeightedAllClusters: equalWeightedAllClusters,
	}, nil
}
