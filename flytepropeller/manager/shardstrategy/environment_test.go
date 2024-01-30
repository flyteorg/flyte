package shardstrategy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvironmentHashCode(t *testing.T) {
	expectedStrategy := &EnvironmentShardStrategy{
		EnvType: Project,
		PerShardIDs: [][]string{
			{"flytesnacks"},
			{"flytefoo", "flytebar"},
			{"*"},
		},
	}
	actualStrategy := &EnvironmentShardStrategy{
		EnvType: Project,
		PerShardIDs: [][]string{
			{"flytesnacks"},
			{"flytefoo", "flytebar"},
			{"*"},
		},
	}
	expectedHashcode, err := expectedStrategy.HashCode()
	assert.NoError(t, err)
	actualHashcode, err := actualStrategy.HashCode()
	assert.NoError(t, err)
	assert.Equal(t, expectedHashcode, actualHashcode)
}
