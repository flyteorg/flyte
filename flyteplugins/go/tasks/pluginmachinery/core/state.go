package core

// Write new plugin state for a plugin
type PluginStateWriter interface {
	// Only the last call to this method is recorded. All previous calls are overwritten
	// This data is also not accessible until the next round.
	Put(stateVersion uint8, v interface{}) error
	// Resets the state to empty or zero value
	Reset() error
}

// Read previously written plugin state (previous round)
type PluginStateReader interface {
	// Retrieve state version that is currently stored
	GetStateVersion() uint8
	// Retrieve the typed state in t from the stored value. It also returns the stateversion.
	// If there is no state, t will be zero value, stateversion will be 0
	Get(t interface{}) (stateVersion uint8, err error)
}
