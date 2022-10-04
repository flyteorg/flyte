package impl

import (
	"context"
	"fmt"
	"hash/fnv"
	"math/rand"

	repositoryInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/resources"
	managerInterfaces "github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyteadmin/pkg/executioncluster"
	"github.com/flyteorg/flyteadmin/pkg/executioncluster/interfaces"

	"github.com/flyteorg/flytestdlib/random"

	runtime "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
)

// Implementation of Random cluster selector
// Selects cluster based on weights and domains.
type RandomClusterSelector struct {
	interfaces.ListTargetsInterface
	equalWeightedAllClusters random.WeightedRandomList
	labelWeightedRandomMap   map[string]random.WeightedRandomList
	resourceManager          managerInterfaces.ResourceInterface
	defaultExecutionLabel    string
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

func convertToRandomWeightedList(ctx context.Context, targets map[string]*executioncluster.ExecutionTarget) (random.WeightedRandomList, error) {
	entries := make([]random.Entry, 0)
	for _, executionTarget := range targets {
		if executionTarget.Enabled {
			targetEntry := random.Entry{
				Item: *executionTarget,
			}
			entries = append(entries, targetEntry)
		}
	}
	weightedRandomList, err := random.NewWeightedRandom(ctx, entries)
	if err != nil {
		return nil, err
	}
	return weightedRandomList, nil
}

func getLabeledWeightedRandomForCluster(ctx context.Context,
	clusterConfig runtime.ClusterConfiguration, executionTargetMap map[string]*executioncluster.ExecutionTarget) (map[string]random.WeightedRandomList, error) {
	labeledWeightedRandomMap := make(map[string]random.WeightedRandomList)
	for label, clusterEntities := range clusterConfig.GetLabelClusterMap() {
		entries := make([]random.Entry, 0)
		for _, clusterEntity := range clusterEntities {
			cluster, found := executionTargetMap[clusterEntity.ID]
			// If cluster is not found, it was never enabled. Non-enabled clusters are not eligible for selection
			if !(found && cluster.Enabled) {
				continue
			}
			targetEntry := random.Entry{
				Item:   *cluster,
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

func (s RandomClusterSelector) GetTarget(ctx context.Context, spec *executioncluster.ExecutionTargetSpec) (*executioncluster.ExecutionTarget, error) {
	if spec == nil {
		return nil, fmt.Errorf("empty executionTargetSpec")
	}
	if spec.TargetID != "" {
		if val, ok := s.GetAllTargets()[spec.TargetID]; ok {
			return val, nil
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

	if weightedRandomList == nil {
		if s.defaultExecutionLabel != "" {
			if _, ok := s.labelWeightedRandomMap[s.defaultExecutionLabel]; ok {
				weightedRandomList = s.labelWeightedRandomMap[s.defaultExecutionLabel]
			} else {
				logger.Warnf(ctx, "No cluster mapping found for the default execution label %s", s.defaultExecutionLabel)
			}
		}
	}

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

func NewRandomClusterSelector(listTargets interfaces.ListTargetsInterface, config runtime.Configuration,
	db repositoryInterfaces.Repository) (interfaces.ClusterInterface, error) {

	defaultExecutionLabel := config.ClusterConfiguration().GetDefaultExecutionLabel()

	equalWeightedAllClusters, err := convertToRandomWeightedList(context.Background(), listTargets.GetValidTargets())
	if err != nil {
		return nil, err
	}
	labelWeightedRandomMap, err := getLabeledWeightedRandomForCluster(context.Background(), config.ClusterConfiguration(), listTargets.GetValidTargets())
	if err != nil {
		return nil, err
	}
	return &RandomClusterSelector{
		labelWeightedRandomMap:   labelWeightedRandomMap,
		resourceManager:          resources.NewResourceManager(db, config.ApplicationConfiguration()),
		equalWeightedAllClusters: equalWeightedAllClusters,
		ListTargetsInterface:     listTargets,
		defaultExecutionLabel:    defaultExecutionLabel,
	}, nil
}
