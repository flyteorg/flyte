package shardstrategy

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"

	"github.com/flyteorg/flyte/flytepropeller/manager/config"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

var (
	hashShardStrategy = &HashShardStrategy{
		ShardCount: 3,
	}

	projectShardStrategy = &EnvironmentShardStrategy{
		EnvType: Project,
		PerShardIDs: [][]string{
			[]string{"flytesnacks"},
			[]string{"flytefoo", "flytebar"},
		},
	}

	projectShardStrategyWildcard = &EnvironmentShardStrategy{
		EnvType: Project,
		PerShardIDs: [][]string{
			[]string{"flytesnacks"},
			[]string{"flytefoo", "flytebar"},
			[]string{"*"},
		},
	}

	domainShardStrategy = &EnvironmentShardStrategy{
		EnvType: Domain,
		PerShardIDs: [][]string{
			[]string{"production"},
			[]string{"foo", "bar"},
		},
	}

	domainShardStrategyWildcard = &EnvironmentShardStrategy{
		EnvType: Domain,
		PerShardIDs: [][]string{
			[]string{"production"},
			[]string{"foo", "bar"},
			[]string{"*"},
		},
	}
)

func TestGetPodCount(t *testing.T) {
	tests := []struct {
		name          string
		shardStrategy ShardStrategy
		podCount      int
	}{
		{"hash", hashShardStrategy, 3},
		{"project", projectShardStrategy, 2},
		{"project_wildcard", projectShardStrategyWildcard, 3},
		{"domain", domainShardStrategy, 2},
		{"domain_wildcard", domainShardStrategyWildcard, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.podCount, tt.shardStrategy.GetPodCount())
		})
	}
}

func TestUpdatePodSpec(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		shardStrategy ShardStrategy
	}{
		{"hash", hashShardStrategy},
		{"project", projectShardStrategy},
		{"project_wildcard", projectShardStrategyWildcard},
		{"domain", domainShardStrategy},
		{"domain_wildcard", domainShardStrategyWildcard},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for podIndex := 0; podIndex < tt.shardStrategy.GetPodCount(); podIndex++ {
				podSpec := v1.PodSpec{
					Containers: []v1.Container{
						v1.Container{
							Name: "flytepropeller",
						},
					},
				}

				err := tt.shardStrategy.UpdatePodSpec(&podSpec, "flytepropeller", podIndex)
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdatePodSpecInvalidPodIndex(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		shardStrategy ShardStrategy
	}{
		{"hash", hashShardStrategy},
		{"project", projectShardStrategy},
		{"project_wildcard", projectShardStrategyWildcard},
		{"domain", domainShardStrategy},
		{"domain_wildcard", domainShardStrategyWildcard},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			podSpec := v1.PodSpec{
				Containers: []v1.Container{
					v1.Container{
						Name: "flytepropeller",
					},
				},
			}

			lowerErr := tt.shardStrategy.UpdatePodSpec(&podSpec, "flytepropeller", -1)
			assert.Error(t, lowerErr)

			upperErr := tt.shardStrategy.UpdatePodSpec(&podSpec, "flytepropeller", tt.shardStrategy.GetPodCount())
			assert.Error(t, upperErr)
		})
	}
}

func TestUpdatePodSpecInvalidPodSpec(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		shardStrategy ShardStrategy
	}{
		{"hash", hashShardStrategy},
		{"project", projectShardStrategy},
		{"project_wildcard", projectShardStrategyWildcard},
		{"domain", domainShardStrategy},
		{"domain_wildcard", domainShardStrategyWildcard},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			podSpec := v1.PodSpec{
				Containers: []v1.Container{
					v1.Container{
						Name: "flytefoo",
					},
				},
			}

			err := tt.shardStrategy.UpdatePodSpec(&podSpec, "flytepropeller", 0)
			assert.Error(t, err)
		})
	}
}

var (
	hashShardConfig = config.ShardConfig{
		Type:       config.ShardTypeHash,
		ShardCount: 3,
	}
	projectShardConfig = config.ShardConfig{
		Type: config.ShardTypeProject,
		PerShardMappings: []config.PerShardMappingsConfig{
			{IDs: []string{"flytesnacks"}},
			{IDs: []string{"flytefoo", "flytebar"}},
		},
	}
	projectShardWildcardConfig = config.ShardConfig{
		Type: config.ShardTypeProject,
		PerShardMappings: []config.PerShardMappingsConfig{
			{IDs: []string{"flytesnacks"}},
			{IDs: []string{"flytefoo", "flytebar"}},
			{IDs: []string{"*"}},
		},
	}
	domainShardConfig = config.ShardConfig{
		Type: config.ShardTypeDomain,
		PerShardMappings: []config.PerShardMappingsConfig{
			{IDs: []string{"production"}},
			{IDs: []string{"foo", "bar"}},
		},
	}
	domainShardWildcardConfig = config.ShardConfig{
		Type: config.ShardTypeDomain,
		PerShardMappings: []config.PerShardMappingsConfig{
			{IDs: []string{"production"}},
			{IDs: []string{"foo", "bar"}},
			{IDs: []string{"*"}},
		},
	}
)

func TestNewShardStrategy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		ShardStrategy ShardStrategy
		shardConfig   config.ShardConfig
	}{
		{"hash", hashShardStrategy, hashShardConfig},
		{"project", projectShardStrategy, projectShardConfig},
		{"project_wildcard", projectShardStrategyWildcard, projectShardWildcardConfig},
		{"domain", domainShardStrategy, domainShardConfig},
		{"domain_wildcard", domainShardStrategyWildcard, domainShardWildcardConfig},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy, err := NewShardStrategy(context.TODO(), tt.shardConfig)
			assert.NoError(t, err)
			assert.Equal(t, tt.ShardStrategy, strategy)
		})
	}
}

var (
	errorHashShardConfig1 = config.ShardConfig{
		Type:       config.ShardTypeHash,
		ShardCount: 0,
	}
	errorHashShardConfig2 = config.ShardConfig{
		Type:       config.ShardTypeHash,
		ShardCount: v1alpha1.ShardKeyspaceSize + 1,
	}
	errorProjectShardConfig = config.ShardConfig{
		Type: config.ShardTypeProject,
		PerShardMappings: []config.PerShardMappingsConfig{
			{IDs: []string{}},
		},
	}
	errorProjectShardWildcardConfig = config.ShardConfig{
		Type: config.ShardTypeProject,
		PerShardMappings: []config.PerShardMappingsConfig{
			{IDs: []string{"flytesnacks", "*"}},
		},
	}
	errorDomainShardWildcardConfig = config.ShardConfig{
		Type: config.ShardTypeDomain,
		PerShardMappings: []config.PerShardMappingsConfig{
			{IDs: []string{"production"}},
			{IDs: []string{"*"}},
			{IDs: []string{"*"}},
		},
	}
	errorShardType = config.ShardConfig{
		Type: -1,
	}
)

func TestNewShardStrategyErrorConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		shardConfig config.ShardConfig
	}{
		{"hash_shard_cnt_zero", errorHashShardConfig1},
		{"hash_shard_cnt_larger_than_keyspace", errorHashShardConfig2},
		{"project_with_zero_config_ids", errorProjectShardConfig},
		{"project_wildcard_with_other_ids", errorProjectShardWildcardConfig},
		{"domain_multi_wildcard_ids", errorDomainShardWildcardConfig},
		{"error_shard_type", errorShardType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewShardStrategy(context.TODO(), tt.shardConfig)
			assert.Error(t, err)
		})
	}
}

func TestComputeHashCode(t *testing.T) {
	expectedData := struct {
		Field1 string
		Field2 int
	}{
		Field1: "flytesnacks",
		Field2: 42,
	}
	actualData := struct {
		Field1 string
		Field2 int
	}{
		Field1: "flytesnacks",
		Field2: 42,
	}
	expectedHashcode, err := computeHashCode(expectedData)
	assert.NoError(t, err)
	actualHashcode, err := computeHashCode(actualData)
	assert.NoError(t, err)
	assert.Equal(t, expectedHashcode, actualHashcode)
}
