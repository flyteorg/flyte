package shardstrategy

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"hash/fnv"

	"github.com/flyteorg/flytepropeller/manager/config"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"

	v1 "k8s.io/api/core/v1"
)

//go:generate mockery -name ShardStrategy -case=underscore

// ShardStrategy defines necessary functionality for a sharding strategy.
type ShardStrategy interface {
	// GetPodCount returns the total number of pods for the sharding strategy.
	GetPodCount() int
	// HashCode generates a unique hash code to identify shard strategy updates.
	HashCode() (uint32, error)
	// UpdatePodSpec amends the PodSpec for the specified index to include label selectors.
	UpdatePodSpec(pod *v1.PodSpec, containerName string, podIndex int) error
}

// NewShardStrategy creates and validates a new ShardStrategy defined by the configuration.
func NewShardStrategy(ctx context.Context, shardConfig config.ShardConfig) (ShardStrategy, error) {
	switch shardConfig.Type {
	case config.ShardTypeHash:
		if shardConfig.ShardCount <= 0 {
			return nil, fmt.Errorf("configured ShardCount (%d) must be greater than zero", shardConfig.ShardCount)
		} else if shardConfig.ShardCount > v1alpha1.ShardKeyspaceSize {
			return nil, fmt.Errorf("configured ShardCount (%d) is larger than available keyspace size (%d)", shardConfig.ShardCount, v1alpha1.ShardKeyspaceSize)
		}

		return &HashShardStrategy{
			ShardCount: shardConfig.ShardCount,
		}, nil
	case config.ShardTypeProject, config.ShardTypeDomain:
		perShardIDs := make([][]string, 0)
		wildcardIDFound := false
		for _, perShardMapping := range shardConfig.PerShardMappings {
			if len(perShardMapping.IDs) == 0 {
				return nil, fmt.Errorf("unable to create shard with 0 configured ids")
			}

			// validate wildcard ID
			for _, id := range perShardMapping.IDs {
				if id == "*" {
					if len(perShardMapping.IDs) != 1 {
						return nil, fmt.Errorf("shards responsible for the wildcard id (ie. '*') may only contain one id")
					}

					if wildcardIDFound {
						return nil, fmt.Errorf("may only define one shard responsible for the wildcard id (ie. '*')")
					}

					wildcardIDFound = true
				}
			}

			perShardIDs = append(perShardIDs, perShardMapping.IDs)
		}

		var envType environmentType
		switch shardConfig.Type {
		case config.ShardTypeProject:
			envType = Project
		case config.ShardTypeDomain:
			envType = Domain
		}

		return &EnvironmentShardStrategy{
			EnvType:     envType,
			PerShardIDs: perShardIDs,
		}, nil
	}

	return nil, fmt.Errorf("shard strategy '%s' does not exist", shardConfig.Type)
}

func computeHashCode(data interface{}) (uint32, error) {
	hash := fnv.New32a()

	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	if err := encoder.Encode(data); err != nil {
		return 0, err
	}

	if _, err := hash.Write(buffer.Bytes()); err != nil {
		return 0, err
	}

	return hash.Sum32(), nil
}
