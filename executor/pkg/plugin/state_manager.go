package plugin

import (
	"bytes"
	"encoding/gob"
	"fmt"

	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
)

var (
	_ pluginsCore.PluginStateReader = &PluginStateManager{}
	_ pluginsCore.PluginStateWriter = &PluginStateManager{}
)

// PluginStateManager implements PluginStateReader and PluginStateWriter using Gob encoding
// over byte buffers. It is initialized with the previous state from TaskAction.Status.PluginState
// and captures new state written by the plugin during Handle.
type PluginStateManager struct {
	prevStateBytes   []byte
	prevStateVersion uint8

	newStateBytes   []byte
	newStateVersion uint8
	stateWritten    bool
}

// NewPluginStateManager creates a PluginStateManager initialized with the previous round's state.
func NewPluginStateManager(prevState []byte, prevVersion uint8) *PluginStateManager {
	return &PluginStateManager{
		prevStateBytes:   prevState,
		prevStateVersion: prevVersion,
	}
}

// GetStateVersion returns the version of the previous state.
func (m *PluginStateManager) GetStateVersion() uint8 {
	return m.prevStateVersion
}

// Get deserializes the previous state into t using Gob decoding.
// If there is no previous state, t remains its zero value and version 0 is returned.
func (m *PluginStateManager) Get(t interface{}) (uint8, error) {
	if len(m.prevStateBytes) == 0 {
		return 0, nil
	}
	buf := bytes.NewBuffer(m.prevStateBytes)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(t); err != nil {
		return 0, fmt.Errorf("failed to decode plugin state: %w", err)
	}
	return m.prevStateVersion, nil
}

// Put serializes v using Gob encoding and stores it as the new state.
// Only the last call to Put is recorded; all previous calls are overwritten.
func (m *PluginStateManager) Put(stateVersion uint8, v interface{}) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("failed to encode plugin state: %w", err)
	}
	m.newStateBytes = buf.Bytes()
	m.newStateVersion = stateVersion
	m.stateWritten = true
	return nil
}

// Reset clears the state to empty.
func (m *PluginStateManager) Reset() error {
	m.newStateBytes = nil
	m.newStateVersion = 0
	m.stateWritten = true
	return nil
}

// GetNewState returns the new state bytes, version, and whether state was written during this round.
// The controller uses this to persist the state back to TaskAction.Status.
func (m *PluginStateManager) GetNewState() (stateBytes []byte, version uint8, written bool) {
	return m.newStateBytes, m.newStateVersion, m.stateWritten
}
