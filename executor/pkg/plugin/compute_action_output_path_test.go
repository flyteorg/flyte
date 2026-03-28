package plugin

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeActionOutputPath(t *testing.T) {
	ctx := context.Background()

	t.Run("shard is inserted after bucket not after full base path", func(t *testing.T) {
		path, err := ComputeActionOutputPath(ctx, "flyte", "my-action-abc", "s3://my-bucket/org/proj/dev/run123", "action-0", 1)
		require.NoError(t, err)

		u := string(path)
		// shard must come before org/proj/dev, not after run123
		shardIdx := strings.Index(u, "s3://my-bucket/") + len("s3://my-bucket/")
		rest := u[shardIdx:]
		parts := strings.SplitN(rest, "/", 2)
		shard := parts[0]
		assert.Len(t, shard, 2, "shard should be a 2-char prefix")
		assert.True(t, strings.HasPrefix(u, "s3://my-bucket/"+shard+"/org/proj/dev/run123/"),
			"shard %q should appear right after the bucket, got: %s", shard, u)
	})

	t.Run("attempt is the last path segment", func(t *testing.T) {
		path, err := ComputeActionOutputPath(ctx, "flyte", "my-action-abc", "s3://my-bucket/org/proj/dev/run123", "action-0", 3)
		require.NoError(t, err)

		parts := strings.Split(strings.TrimRight(string(path), "/"), "/")
		assert.Equal(t, "3", parts[len(parts)-1])
	})

	t.Run("action name is the second-to-last path segment", func(t *testing.T) {
		path, err := ComputeActionOutputPath(ctx, "flyte", "my-action-abc", "s3://my-bucket/org/proj/dev/run123", "action-0", 1)
		require.NoError(t, err)

		parts := strings.Split(strings.TrimRight(string(path), "/"), "/")
		assert.Equal(t, "action-0", parts[len(parts)-2])
	})

	t.Run("trailing slash on RunOutputBase is handled", func(t *testing.T) {
		withSlash, err := ComputeActionOutputPath(ctx, "flyte", "my-action-abc", "s3://my-bucket/org/proj/dev/run123/", "action-0", 1)
		require.NoError(t, err)
		withoutSlash, err := ComputeActionOutputPath(ctx, "flyte", "my-action-abc", "s3://my-bucket/org/proj/dev/run123", "action-0", 1)
		require.NoError(t, err)

		assert.Equal(t, string(withoutSlash), string(withSlash))
	})

	t.Run("same namespace and name always produces the same shard", func(t *testing.T) {
		p1, err := ComputeActionOutputPath(ctx, "flyte", "my-action-abc", "s3://my-bucket/run", "action-0", 1)
		require.NoError(t, err)
		p2, err := ComputeActionOutputPath(ctx, "flyte", "my-action-abc", "s3://my-bucket/run", "action-0", 1)
		require.NoError(t, err)

		assert.Equal(t, p1, p2)
	})

	t.Run("different namespace or name produces different shards", func(t *testing.T) {
		// Collect shards from a handful of distinct namespace/name pairs and
		// verify we see more than one unique value — confirming distribution.
		inputs := [][2]string{
			{"ns-a", "action-1"},
			{"ns-b", "action-1"},
			{"ns-a", "action-2"},
			{"ns-c", "action-99"},
			{"prod", "long-running-task-xyz"},
		}
		shards := make(map[string]struct{})
		for _, in := range inputs {
			p, err := ComputeActionOutputPath(ctx, in[0], in[1], "s3://my-bucket/run", "action-0", 1)
			require.NoError(t, err)
			// extract shard: segment right after bucket
			u := string(p)
			after := strings.TrimPrefix(u, "s3://my-bucket/")
			shard := strings.SplitN(after, "/", 2)[0]
			shards[shard] = struct{}{}
		}
		assert.Greater(t, len(shards), 1, "expected multiple distinct shards across different namespace/name pairs")
	})

	t.Run("different attempts produce different paths", func(t *testing.T) {
		p1, err := ComputeActionOutputPath(ctx, "flyte", "my-action-abc", "s3://my-bucket/run", "action-0", 1)
		require.NoError(t, err)
		p2, err := ComputeActionOutputPath(ctx, "flyte", "my-action-abc", "s3://my-bucket/run", "action-0", 2)
		require.NoError(t, err)

		assert.NotEqual(t, p1, p2)
	})

	t.Run("shard does not change across attempts for the same action", func(t *testing.T) {
		extractShard := func(path string) string {
			after := strings.TrimPrefix(path, "s3://my-bucket/")
			return strings.SplitN(after, "/", 2)[0]
		}

		p1, err := ComputeActionOutputPath(ctx, "flyte", "my-action-abc", "s3://my-bucket/run", "action-0", 1)
		require.NoError(t, err)
		p2, err := ComputeActionOutputPath(ctx, "flyte", "my-action-abc", "s3://my-bucket/run", "action-0", 2)
		require.NoError(t, err)

		assert.Equal(t, extractShard(string(p1)), extractShard(string(p2)),
			"shard should be stable across retries since it only depends on namespace/name")
	})

	t.Run("invalid RunOutputBase returns error", func(t *testing.T) {
		_, err := ComputeActionOutputPath(ctx, "flyte", "my-action-abc", "://bad url", "action-0", 1)
		assert.Error(t, err)
	})
}