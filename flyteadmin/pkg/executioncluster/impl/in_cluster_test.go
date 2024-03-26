package impl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteadmin/pkg/executioncluster"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

func TestInClusterGetTarget(t *testing.T) {
	cluster := InCluster{
		target: executioncluster.ExecutionTarget{
			ID: "t1",
		},
	}
	target, err := cluster.GetTarget(context.Background(), nil)
	assert.Nil(t, err)
	assert.Equal(t, "t1", target.ID)
}

func TestInClusterGetTarget_AllowableSpecIDs(t *testing.T) {
	cluster := InCluster{
		target: executioncluster.ExecutionTarget{},
	}
	target, err := cluster.GetTarget(context.Background(), nil)
	assert.Nil(t, err)
	assert.Equal(t, *target, cluster.target)

	target, err = cluster.GetTarget(context.Background(), &executioncluster.ExecutionTargetSpec{})
	assert.Nil(t, err)
	assert.Equal(t, *target, cluster.target)

	target, err = cluster.GetTarget(context.Background(), &executioncluster.ExecutionTargetSpec{
		TargetID: defaultInClusterTargetID,
	})
	assert.Nil(t, err)
	assert.Equal(t, *target, cluster.target)

	_, err = cluster.GetTarget(context.Background(), &executioncluster.ExecutionTargetSpec{
		TargetID: "t1",
	})
	assert.Error(t, err)
}

func TestInClusterGetRemoteTarget(t *testing.T) {
	cluster := InCluster{
		target: executioncluster.ExecutionTarget{},
	}
	_, err := cluster.GetTarget(context.Background(), &executioncluster.ExecutionTargetSpec{TargetID: "t1"})
	assert.EqualError(t, err, "remote target t1 is not supported")
}

func TestInClusterGetTargetWithExecutionClusterLabel(t *testing.T) {
	cluster := InCluster{
		target: executioncluster.ExecutionTarget{},
	}
	_, err := cluster.GetTarget(context.Background(), &executioncluster.ExecutionTargetSpec{ExecutionClusterLabel: &admin.ExecutionClusterLabel{
		Value: "l1",
	}})
	assert.EqualError(t, err, "execution cluster label l1 is not supported")
}

func TestInClusterGetAllValidTargets(t *testing.T) {
	target := &executioncluster.ExecutionTarget{
		Enabled: true,
	}
	cluster := InCluster{
		target: *target,
		asTargets: map[string]*executioncluster.ExecutionTarget{
			target.ID: target,
		},
	}
	targets := cluster.GetValidTargets()
	assert.Equal(t, 1, len(targets))
	assert.Equal(t, targets[target.ID], target)

	targets = cluster.GetAllTargets()
	assert.Equal(t, 1, len(targets))
	assert.Equal(t, targets[target.ID], target)
}
