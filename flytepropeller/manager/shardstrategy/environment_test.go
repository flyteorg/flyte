package shardstrategy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvironmentHashCode(t *testing.T) {
	strategy1 := &EnvironmentShardStrategy{
		EnvType: Project,
		PerShardIDs: [][]string{
			{"flytesnacks"},
			{"flytefoo", "flytebar"},
			{"*"},
		},
	}
	strategy2 := &EnvironmentShardStrategy{
		EnvType: Project,
		PerShardIDs: [][]string{
			{"flytesnacks"},
			{"flytefoo", "flytebar"},
			{"*"},
		},
	}
	hashcode1, err := strategy1.HashCode()
	assert.NoError(t, err)
	hashcode2, err := strategy2.HashCode()
	assert.NoError(t, err)
	assert.Equal(t, hashcode1, hashcode2)
}
