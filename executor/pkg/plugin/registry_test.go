package plugin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	executorConfig "github.com/flyteorg/flyte/v2/executor/pkg/config"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	k8sPlugin "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

// mockCorePlugin is a minimal pluginsCore.Plugin for use in tests.
type mockCorePlugin struct{ id string }

func (m *mockCorePlugin) GetID() string { return m.id }
func (m *mockCorePlugin) GetProperties() pluginsCore.PluginProperties {
	return pluginsCore.PluginProperties{}
}
func (m *mockCorePlugin) Handle(_ context.Context, _ pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {
	return pluginsCore.Transition{}, nil
}
func (m *mockCorePlugin) Abort(_ context.Context, _ pluginsCore.TaskExecutionContext) error {
	return nil
}
func (m *mockCorePlugin) Finalize(_ context.Context, _ pluginsCore.TaskExecutionContext) error {
	return nil
}

// mockPluginRegistry implements PluginRegistryIface with configurable entries.
type mockPluginRegistry struct {
	corePlugins []pluginsCore.PluginEntry
	k8sPlugins  []k8sPlugin.PluginEntry
}

func (m *mockPluginRegistry) GetCorePlugins() []pluginsCore.PluginEntry { return m.corePlugins }
func (m *mockPluginRegistry) GetK8sPlugins() []k8sPlugin.PluginEntry    { return m.k8sPlugins }

// coreEntry builds a PluginEntry backed by a mockCorePlugin.
func coreEntry(id string, taskTypes ...string) pluginsCore.PluginEntry {
	p := &mockCorePlugin{id: id}
	return pluginsCore.PluginEntry{
		ID:                  id,
		RegisteredTaskTypes: taskTypes,
		LoadPlugin: func(_ context.Context, _ pluginsCore.SetupContext) (pluginsCore.Plugin, error) {
			return p, nil
		},
	}
}

func newTestRegistry(reg PluginRegistryIface) *Registry {
	setupCtx := NewSetupContext(nil, nil, nil, nil, nil, "TaskAction", promutils.NewTestScope())
	return NewRegistry(setupCtx, reg)
}

func resetConfig(t *testing.T, enabled []string) {
	t.Helper()
	require.NoError(t, executorConfig.SetTaskPluginConfig(&executorConfig.TaskPluginConfig{
		EnabledPlugins: enabled,
	}))
	t.Cleanup(func() {
		_ = executorConfig.SetTaskPluginConfig(&executorConfig.TaskPluginConfig{})
	})
}

func TestPluginsConfigMeta(t *testing.T) {
	tests := []struct {
		name           string
		configPlugins  []string
		checkAllowed   string
		wantAllowed    bool
		checkGetConfig bool
		wantGetPlugins []string
	}{
		{
			name:           "GetEnabledPlugins reflects config",
			configPlugins:  []string{"container", "spark"},
			checkGetConfig: true,
			wantGetPlugins: []string{"container", "spark"},
		},
		{
			name:           "GetEnabledPlugins empty config",
			configPlugins:  nil,
			checkGetConfig: true,
			wantGetPlugins: nil,
		},
		{
			name:          "empty EnabledPlugins allows any id",
			configPlugins: nil,
			checkAllowed:  "anything",
			wantAllowed:   true,
		},
		{
			name:          "empty EnabledPlugins allows empty string",
			configPlugins: nil,
			checkAllowed:  "",
			wantAllowed:   true,
		},
		{
			name:          "listed id is allowed",
			configPlugins: []string{"container", "spark"},
			checkAllowed:  "container",
			wantAllowed:   true,
		},
		{
			name:          "unlisted id is blocked",
			configPlugins: []string{"container"},
			checkAllowed:  "ray",
			wantAllowed:   false,
		},
		{
			name:          "empty string blocked when list non-empty",
			configPlugins: []string{"container"},
			checkAllowed:  "",
			wantAllowed:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resetConfig(t, tc.configPlugins)
			if tc.checkGetConfig {
				meta := GetEnabledPlugins()
				assert.Equal(t, tc.wantGetPlugins, meta.EnabledPlugins)
				return
			}
			pcm := &PluginsConfigMeta{EnabledPlugins: tc.configPlugins}
			assert.Equal(t, tc.wantAllowed, pcm.isAllowed(tc.checkAllowed))
		})
	}
}

// 3.1 Empty allowlist loads all plugins.
func TestRegistry_EmptyAllowlist_LoadsAll(t *testing.T) {
	resetConfig(t, nil)

	reg := &mockPluginRegistry{
		corePlugins: []pluginsCore.PluginEntry{
			coreEntry("container", "container"),
			coreEntry("sidecar", "sidecar"),
		},
	}
	r := newTestRegistry(reg)
	require.NoError(t, r.Initialize(context.Background()))

	p, err := r.ResolvePlugin("container")
	require.NoError(t, err)
	assert.Equal(t, "container", p.GetID())

	p, err = r.ResolvePlugin("sidecar")
	require.NoError(t, err)
	assert.Equal(t, "sidecar", p.GetID())
}

// 3.2 Non-empty allowlist filters out plugins not in the list.
func TestRegistry_Allowlist_FiltersPlugins(t *testing.T) {
	resetConfig(t, []string{"container"})

	reg := &mockPluginRegistry{
		corePlugins: []pluginsCore.PluginEntry{
			coreEntry("container", "container"),
			coreEntry("ray", "ray"),
		},
	}
	r := newTestRegistry(reg)
	require.NoError(t, r.Initialize(context.Background()))

	p, err := r.ResolvePlugin("container")
	require.NoError(t, err)
	assert.Equal(t, "container", p.GetID())

	// "ray" was filtered out — no default plugin, so ResolvePlugin should error
	_, err = r.ResolvePlugin("ray")
	assert.Error(t, err)
}

// 3.3 ResolvePlugin returns error for task types belonging to filtered plugins.
func TestRegistry_ResolvePlugin_ErrorForFilteredTaskType(t *testing.T) {
	resetConfig(t, []string{"container"})

	reg := &mockPluginRegistry{
		corePlugins: []pluginsCore.PluginEntry{
			coreEntry("container", "container"),
			coreEntry("spark", "spark"),
		},
	}
	r := newTestRegistry(reg)
	require.NoError(t, r.Initialize(context.Background()))

	_, err := r.ResolvePlugin("spark")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "spark")
}

// 3.4 Allowlist with unknown ID does not affect loading of other plugins.
func TestRegistry_Allowlist_UnknownIDIgnored(t *testing.T) {
	resetConfig(t, []string{"container", "nonexistent-plugin"})

	reg := &mockPluginRegistry{
		corePlugins: []pluginsCore.PluginEntry{
			coreEntry("container", "container"),
		},
	}
	r := newTestRegistry(reg)
	require.NoError(t, r.Initialize(context.Background()))

	p, err := r.ResolvePlugin("container")
	require.NoError(t, err)
	assert.Equal(t, "container", p.GetID())
}

// 3.5 DefaultForTaskTypes routes a task type to a different loaded plugin.
func TestRegistry_DefaultForTaskTypes_Override(t *testing.T) {
	require.NoError(t, executorConfig.SetTaskPluginConfig(&executorConfig.TaskPluginConfig{
		DefaultForTaskTypes: map[string]string{"sidecar": "container"},
	}))
	t.Cleanup(func() { _ = executorConfig.SetTaskPluginConfig(&executorConfig.TaskPluginConfig{}) })

	reg := &mockPluginRegistry{
		corePlugins: []pluginsCore.PluginEntry{
			coreEntry("container", "container"),
			coreEntry("sidecar", "sidecar"),
		},
	}
	r := newTestRegistry(reg)
	require.NoError(t, r.Initialize(context.Background()))

	// sidecar task type must now resolve to the container plugin
	p, err := r.ResolvePlugin("sidecar")
	require.NoError(t, err)
	assert.Equal(t, "container", p.GetID())

	// container task type is unaffected
	p, err = r.ResolvePlugin("container")
	require.NoError(t, err)
	assert.Equal(t, "container", p.GetID())
}

// 3.6 DefaultForTaskTypes with an unknown plugin ID does not panic and leaves routing unchanged.
func TestRegistry_DefaultForTaskTypes_UnknownPluginID(t *testing.T) {
	require.NoError(t, executorConfig.SetTaskPluginConfig(&executorConfig.TaskPluginConfig{
		DefaultForTaskTypes: map[string]string{"container": "nonexistent-plugin"},
	}))
	t.Cleanup(func() { _ = executorConfig.SetTaskPluginConfig(&executorConfig.TaskPluginConfig{}) })

	reg := &mockPluginRegistry{
		corePlugins: []pluginsCore.PluginEntry{
			coreEntry("container", "container"),
		},
	}
	r := newTestRegistry(reg)
	require.NoError(t, r.Initialize(context.Background()))

	// routing for container must remain as the original container plugin
	p, err := r.ResolvePlugin("container")
	require.NoError(t, err)
	assert.Equal(t, "container", p.GetID())
}
