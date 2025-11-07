package utils

import "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"

// GetVariable retrieves a variable from a VariableMap by key.
// Returns nil if the key is not found or if the VariableMap is nil.
func GetVariable(vm *core.VariableMap, key string) *core.Variable {
	if vm == nil {
		return nil
	}
	for _, entry := range vm.GetVariables() {
		if entry.GetKey() == key {
			return entry.GetValue()
		}
	}
	return nil
}

// SetVariable adds or updates a variable in a VariableMap.
// If the key already exists, it updates the value. Otherwise, it appends a new entry.
func SetVariable(vm *core.VariableMap, key string, variable *core.Variable) {
	if vm == nil {
		return
	}
	// Update the value of existing key
	for _, entry := range vm.Variables {
		if entry.GetKey() == key {
			entry.Value = variable
			return
		}
	}
	vm.Variables = append(vm.Variables, &core.VariableEntry{
		Key:   key,
		Value: variable,
	})
}

// VariableMapToMap converts a VariableMap to a Go map for easier access.
func VariableMapToMap(vm *core.VariableMap) map[string]*core.Variable {
	result := make(map[string]*core.Variable)
	if vm != nil {
		for _, entry := range vm.GetVariables() {
			result[entry.GetKey()] = entry.GetValue()
		}
	}
	return result
}

// MapToVariableMap converts a Go map to a VariableMap.
func MapToVariableMap(m map[string]*core.Variable) *core.VariableMap {
	entries := make([]*core.VariableEntry, 0, len(m))
	for key, value := range m {
		entries = append(entries, &core.VariableEntry{
			Key:   key,
			Value: value,
		})
	}
	return &core.VariableMap{Variables: entries}
}
