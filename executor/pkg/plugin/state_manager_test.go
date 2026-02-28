package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
)

func TestPluginStateManager_EmptyState(t *testing.T) {
	mgr := NewPluginStateManager(nil, 0)

	assert.Equal(t, uint8(0), mgr.GetStateVersion())

	var state k8s.PluginState
	version, err := mgr.Get(&state)
	require.NoError(t, err)
	assert.Equal(t, uint8(0), version)
	assert.Equal(t, k8s.PluginState{}, state)
}

func TestPluginStateManager_RoundTrip(t *testing.T) {
	original := k8s.PluginState{
		Phase:        pluginsCore.PhaseRunning,
		PhaseVersion: 3,
		Reason:       "task is running",
	}

	// Write state
	writeMgr := NewPluginStateManager(nil, 0)
	require.NoError(t, writeMgr.Put(1, &original))

	newBytes, newVersion, written := writeMgr.GetNewState()
	assert.True(t, written)
	assert.Equal(t, uint8(1), newVersion)
	assert.NotEmpty(t, newBytes)

	// Read state back in a new manager (simulates next reconciliation round)
	readMgr := NewPluginStateManager(newBytes, newVersion)
	assert.Equal(t, uint8(1), readMgr.GetStateVersion())

	var restored k8s.PluginState
	version, err := readMgr.Get(&restored)
	require.NoError(t, err)
	assert.Equal(t, uint8(1), version)
	assert.Equal(t, original, restored)
}

func TestPluginStateManager_PutOverwrites(t *testing.T) {
	mgr := NewPluginStateManager(nil, 0)

	first := k8s.PluginState{Phase: pluginsCore.PhaseQueued}
	second := k8s.PluginState{Phase: pluginsCore.PhaseRunning}

	require.NoError(t, mgr.Put(1, &first))
	require.NoError(t, mgr.Put(2, &second))

	newBytes, newVersion, written := mgr.GetNewState()
	assert.True(t, written)
	assert.Equal(t, uint8(2), newVersion)

	readMgr := NewPluginStateManager(newBytes, newVersion)
	var restored k8s.PluginState
	_, err := readMgr.Get(&restored)
	require.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, restored.Phase)
}

func TestPluginStateManager_Reset(t *testing.T) {
	original := k8s.PluginState{Phase: pluginsCore.PhaseRunning}

	writeMgr := NewPluginStateManager(nil, 0)
	require.NoError(t, writeMgr.Put(1, &original))
	require.NoError(t, writeMgr.Reset())

	newBytes, newVersion, written := writeMgr.GetNewState()
	assert.True(t, written)
	assert.Equal(t, uint8(0), newVersion)
	assert.Nil(t, newBytes)
}

func TestPluginStateManager_NoWriteReturnsNotWritten(t *testing.T) {
	mgr := NewPluginStateManager(nil, 0)

	_, _, written := mgr.GetNewState()
	assert.False(t, written)
}
