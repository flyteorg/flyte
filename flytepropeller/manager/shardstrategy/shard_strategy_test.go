package shardstrategy

import (
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "k8s.io/api/core/v1"
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
