package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
)

func TestTruncateShortDescription_UnderLimit(t *testing.T) {
	desc := "Short description"
	result := truncateShortDescription(desc)
	assert.Equal(t, desc, result)
}

func TestTruncateShortDescription_AtLimit(t *testing.T) {
	desc := strings.Repeat("a", 255)
	result := truncateShortDescription(desc)
	assert.Equal(t, desc, result)
	assert.Len(t, result, 255)
}

func TestTruncateShortDescription_OverLimit(t *testing.T) {
	desc := strings.Repeat("a", 300)
	result := truncateShortDescription(desc)
	assert.Len(t, result, 255)
	assert.Equal(t, strings.Repeat("a", 255), result)
}

func TestTruncateLongDescription_UnderLimit(t *testing.T) {
	desc := "Long description"
	result := truncateLongDescription(desc)
	assert.Equal(t, desc, result)
}

func TestTruncateLongDescription_AtLimit(t *testing.T) {
	desc := strings.Repeat("b", 2048)
	result := truncateLongDescription(desc)
	assert.Equal(t, desc, result)
	assert.Len(t, result, 2048)
}

func TestTruncateLongDescription_OverLimit(t *testing.T) {
	desc := strings.Repeat("b", 3000)
	result := truncateLongDescription(desc)
	assert.Len(t, result, 2048)
	assert.Equal(t, strings.Repeat("b", 2048), result)
}

func TestFillDefaultInputs(t *testing.T) {
	t.Run("nil inputs gets defaults", func(t *testing.T) {
		defaults := []*task.NamedParameter{
			{
				Name: "x",
				Parameter: &core.Parameter{
					Behavior: &core.Parameter_Default{Default: newStringLiteral("default_val")},
				},
			},
		}
		result := fillDefaultInputs(nil, defaults)
		require.Len(t, result.Literals, 1)
		assert.Equal(t, "x", result.Literals[0].Name)
	})

	t.Run("existing input not overridden by default", func(t *testing.T) {
		inputs := &task.Inputs{
			Literals: []*task.NamedLiteral{
				{Name: "x", Value: newStringLiteral("user_val")},
			},
		}
		defaults := []*task.NamedParameter{
			{
				Name: "x",
				Parameter: &core.Parameter{
					Behavior: &core.Parameter_Default{Default: newStringLiteral("default_val")},
				},
			},
		}
		result := fillDefaultInputs(inputs, defaults)
		require.Len(t, result.Literals, 1)
		assert.Equal(t, "x", result.Literals[0].Name)
		// Should keep the user-provided value
		assert.Equal(t, "user_val", result.Literals[0].Value.GetScalar().GetPrimitive().GetStringValue())
	})

	t.Run("missing input filled with default", func(t *testing.T) {
		inputs := &task.Inputs{
			Literals: []*task.NamedLiteral{
				{Name: "x", Value: newStringLiteral("user_val")},
			},
		}
		defaults := []*task.NamedParameter{
			{
				Name: "y",
				Parameter: &core.Parameter{
					Behavior: &core.Parameter_Default{Default: newIntLiteral(42)},
				},
			},
		}
		result := fillDefaultInputs(inputs, defaults)
		require.Len(t, result.Literals, 2)
		assert.Equal(t, "x", result.Literals[0].Name)
		assert.Equal(t, "y", result.Literals[1].Name)
	})

	t.Run("default without value is skipped", func(t *testing.T) {
		inputs := &task.Inputs{}
		defaults := []*task.NamedParameter{
			{
				Name:      "z",
				Parameter: &core.Parameter{},
			},
		}
		result := fillDefaultInputs(inputs, defaults)
		assert.Empty(t, result.Literals)
	})

	t.Run("nil defaults returns inputs unchanged", func(t *testing.T) {
		inputs := &task.Inputs{
			Literals: []*task.NamedLiteral{
				{Name: "a", Value: newStringLiteral("val")},
			},
		}
		result := fillDefaultInputs(inputs, nil)
		require.Len(t, result.Literals, 1)
		assert.Equal(t, "a", result.Literals[0].Name)
	})
}

func TestTaskIdFromTaskSpec(t *testing.T) {
	t.Run("nil spec returns nil", func(t *testing.T) {
		assert.Nil(t, taskIdFromTaskSpec(nil))
	})

	t.Run("extracts id from task spec", func(t *testing.T) {
		spec := &task.TaskSpec{
			TaskTemplate: &core.TaskTemplate{
				Id: &core.Identifier{
					Org:     "org1",
					Project: "proj1",
					Domain:  "dev",
					Name:    "my-task",
					Version: "v1",
				},
			},
		}
		id := taskIdFromTaskSpec(spec)
		assert.Equal(t, "", id.Org)
		assert.Equal(t, "proj1", id.Project)
		assert.Equal(t, "dev", id.Domain)
		assert.Equal(t, "my-task", id.Name)
		assert.Equal(t, "v1", id.Version)
	})

	t.Run("spec without template returns empty id", func(t *testing.T) {
		spec := &task.TaskSpec{}
		id := taskIdFromTaskSpec(spec)
		assert.NotNil(t, id)
		assert.Empty(t, id.Name)
	})
}

func TestGenerateCacheKeyForTask(t *testing.T) {
	t.Run("deterministic output", func(t *testing.T) {
		tmpl := &core.TaskTemplate{
			Id:   &core.Identifier{Name: "my-task"},
			Type: "python",
			Metadata: &core.TaskMetadata{
				Discoverable:     true,
				DiscoveryVersion: "1.0",
			},
			Interface: &core.TypedInterface{},
		}
		inputs := &task.Inputs{
			Literals: []*task.NamedLiteral{
				{Name: "x", Value: newStringLiteral("hello")},
			},
		}

		hash1, err := computeFilteredInputsHash(tmpl, inputs)
		require.NoError(t, err)
		key1, err := generateCacheKeyForTask(tmpl, hash1)
		require.NoError(t, err)
		assert.NotEmpty(t, key1)

		hash2, err := computeFilteredInputsHash(tmpl, inputs)
		require.NoError(t, err)
		key2, err := generateCacheKeyForTask(tmpl, hash2)
		require.NoError(t, err)
		assert.Equal(t, key1, key2)
	})

	t.Run("different inputs produce different keys", func(t *testing.T) {
		tmpl := &core.TaskTemplate{
			Id:   &core.Identifier{Name: "my-task"},
			Type: "python",
			Metadata: &core.TaskMetadata{
				Discoverable:     true,
				DiscoveryVersion: "1.0",
			},
		}

		hash1, err := computeFilteredInputsHash(tmpl, &task.Inputs{
			Literals: []*task.NamedLiteral{
				{Name: "x", Value: newStringLiteral("hello")},
			},
		})
		require.NoError(t, err)
		key1, err := generateCacheKeyForTask(tmpl, hash1)
		require.NoError(t, err)

		hash2, err := computeFilteredInputsHash(tmpl, &task.Inputs{
			Literals: []*task.NamedLiteral{
				{Name: "x", Value: newStringLiteral("world")},
			},
		})
		require.NoError(t, err)
		key2, err := generateCacheKeyForTask(tmpl, hash2)
		require.NoError(t, err)

		assert.NotEqual(t, key1, key2)
	})

	t.Run("ignored input vars are excluded", func(t *testing.T) {
		tmpl := &core.TaskTemplate{
			Id:   &core.Identifier{Name: "my-task"},
			Type: "python",
			Metadata: &core.TaskMetadata{
				Discoverable:       true,
				DiscoveryVersion:   "1.0",
				CacheIgnoreInputVars: []string{"y"},
			},
		}
		inputs := &task.Inputs{
			Literals: []*task.NamedLiteral{
				{Name: "x", Value: newStringLiteral("hello")},
				{Name: "y", Value: newStringLiteral("ignored")},
			},
		}

		hashWithIgnored, err := computeFilteredInputsHash(tmpl, inputs)
		require.NoError(t, err)
		keyWithIgnored, err := generateCacheKeyForTask(tmpl, hashWithIgnored)
		require.NoError(t, err)

		// Same template without the ignored var should produce same key
		inputsWithoutY := &task.Inputs{
			Literals: []*task.NamedLiteral{
				{Name: "x", Value: newStringLiteral("hello")},
			},
		}
		hashWithout, err := computeFilteredInputsHash(tmpl, inputsWithoutY)
		require.NoError(t, err)
		keyWithout, err := generateCacheKeyForTask(tmpl, hashWithout)
		require.NoError(t, err)

		assert.Equal(t, keyWithIgnored, keyWithout)
	})

	t.Run("nil inputs", func(t *testing.T) {
		tmpl := &core.TaskTemplate{
			Id:   &core.Identifier{Name: "my-task"},
			Type: "python",
			Metadata: &core.TaskMetadata{
				Discoverable:     true,
				DiscoveryVersion: "1.0",
			},
		}
		hash, err := computeFilteredInputsHash(tmpl, nil)
		require.NoError(t, err)
		key, err := generateCacheKeyForTask(tmpl, hash)
		require.NoError(t, err)
		assert.NotEmpty(t, key)
	})
}

func TestBuildRunOutputBase(t *testing.T) {
	t.Run("does not end with trailing slash", func(t *testing.T) {
		// Regression: a trailing slash caused SDK consumers doing
		// f"{run_output_base}/{action_name}/..." to produce a doubled slash like
		// "s3://bucket/proj/dev/run//action/1/error.pb".
		got := buildRunOutputBase("s3://flyte-data", "flytesnacks", "development", "r782")
		assert.Equal(t, "s3://flyte-data/flytesnacks/development/r782", got)
		assert.False(t, strings.HasSuffix(got, "/"), "run output base must not end in '/', got %q", got)
	})

	t.Run("storagePrefix with trailing slash is normalized", func(t *testing.T) {
		got := buildRunOutputBase("s3://flyte-data/", "flytesnacks", "development", "r782")
		assert.Equal(t, "s3://flyte-data/flytesnacks/development/r782", got)
	})

	t.Run("appending an action segment yields single slash separator", func(t *testing.T) {
		base := buildRunOutputBase("s3://flyte-data", "flytesnacks", "development", "r782")
		// Simulate the SDK's f-string construction: f"{run_output_base}/{action_name}/1/error.pb"
		errorPath := base + "/action-0/1/error.pb"
		rest := strings.TrimPrefix(errorPath, "s3://")
		assert.NotContains(t, rest, "//", "constructed error path must not contain '//', got %q", errorPath)
	})
}
