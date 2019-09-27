package executioncluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func getRandomClusterSelectorForTest() RandomClusterSelector {
	return RandomClusterSelector{
		executionTargetMap: map[string]ExecutionTarget{
			"cluster-1": {
				ID:      "t1",
				Enabled: true,
			},
			"cluster-2": {
				ID:      "t2",
				Enabled: true,
			},
			"cluster-disabled": {
				ID: "t4",
			},
		},
		totalEnabledClusterCount: 2,
	}
}

func TestRandomClusterSelectorGetTarget(t *testing.T) {
	cluster := getRandomClusterSelectorForTest()
	target, err := cluster.GetTarget(&ExecutionTargetSpec{TargetID: "cluster-1"})
	assert.Nil(t, err)
	assert.Equal(t, "t1", target.ID)
	assert.True(t, target.Enabled)
	target, err = cluster.GetTarget(&ExecutionTargetSpec{TargetID: "cluster-disabled"})
	assert.Nil(t, err)
	assert.Equal(t, "t4", target.ID)
	assert.False(t, target.Enabled)
}

func TestRandomClusterSelectorGetRandomTarget(t *testing.T) {
	cluster := getRandomClusterSelectorForTest()
	target, err := cluster.GetTarget(nil)
	assert.Nil(t, err)
	assert.NotNil(t, target)
	assert.NotEmpty(t, target.ID)
}

func TestRandomClusterSelectorGetRemoteTarget(t *testing.T) {
	cluster := getRandomClusterSelectorForTest()
	_, err := cluster.GetTarget(&ExecutionTargetSpec{TargetID: "cluster-3"})
	assert.NotNil(t, err)
	assert.EqualError(t, err, "invalid cluster target cluster-3")
}

func TestRandomClusterSelectorGetAllValidTargets(t *testing.T) {
	cluster := getRandomClusterSelectorForTest()
	targets := cluster.GetAllValidTargets()
	assert.Equal(t, 2, len(targets))
}
