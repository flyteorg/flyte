package ioutils

import (
	"context"
	"testing"

	core2 "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
)

func TestNewOutputSandbox(t *testing.T) {
	assert.Equal(t, NewRawOutputPaths(context.TODO(), "x").GetRawOutputPrefix(), storage.DataReference("x"))
}

func TestNewShardedDeterministicRawOutputPath(t *testing.T) {
	ctx := context.TODO()

	t.Run("success-path", func(t *testing.T) {
		ss := NewConstantShardSelector([]string{"x"})
		sd, err := NewShardedDeterministicRawOutputPath(ctx, ss, "s3://bucket", "m", storage.URLPathConstructor{})
		assert.NoError(t, err)
		assert.Equal(t, storage.DataReference("s3://bucket/x/6b0d31c0d563223024da45691584643ac78c96e8"), sd.GetRawOutputPrefix())
	})

	t.Run("error", func(t *testing.T) {
		ss := NewConstantShardSelector([]string{"s3:// abc"})
		sd, err := NewShardedDeterministicRawOutputPath(ctx, ss, "s3://bucket", "m", storage.URLPathConstructor{})
		assert.Error(t, err, "%s", sd)
	})
}

func TestNewShardedRawOutputPath(t *testing.T) {
	ctx := context.TODO()
	t.Run("", func(t *testing.T) {
		ss := NewConstantShardSelector([]string{"x"})
		sd, err := NewShardedRawOutputPath(ctx, ss, "s3://flyte", "unique", storage.URLPathConstructor{})
		assert.NoError(t, err)
		assert.Equal(t, storage.DataReference("s3://flyte/x/unique"), sd.GetRawOutputPrefix())
	})

	t.Run("error", func(t *testing.T) {
		ss := NewConstantShardSelector([]string{"s3:// abc"})
		sd, err := NewShardedRawOutputPath(ctx, ss, "s3://bucket", "m", storage.URLPathConstructor{})
		assert.Error(t, err, "%s", sd)
	})
}

func TestNewDeterministicUniqueRawOutputPath(t *testing.T) {
	ctx := context.TODO()

	t.Run("success-path", func(t *testing.T) {
		sd, err := NewDeterministicUniqueRawOutputPath(ctx, "s3://bucket", "m", storage.URLPathConstructor{})
		assert.NoError(t, err)
		assert.Equal(t, storage.DataReference("s3://bucket/6b0d31c0d563223024da45691584643ac78c96e8"), sd.GetRawOutputPrefix())
	})

	t.Run("error-not-possible", func(t *testing.T) {
		sd, err := NewDeterministicUniqueRawOutputPath(ctx, "bucket", "m", storage.URLPathConstructor{})
		assert.NoError(t, err)
		assert.Equal(t, "/bucket/6b0d31c0d563223024da45691584643ac78c96e8", sd.GetRawOutputPrefix().String())
	})
}

func TestNewTaskIDRawOutputPath(t *testing.T) {
	p, err := NewTaskIDRawOutputPath(context.TODO(), "s3://bucket", &core2.TaskExecutionIdentifier{
		NodeExecutionId: &core2.NodeExecutionIdentifier{
			NodeId: "n1",
			ExecutionId: &core2.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "exec",
			},
		},
		RetryAttempt: 0,
		TaskId: &core2.Identifier{
			Name: "task1",
		},
	}, storage.URLPathConstructor{})
	assert.NoError(t, err)
	assert.Equal(t, "s3://bucket/project/domain/exec/n1/0/task1", p.GetRawOutputPrefix().String())
}
