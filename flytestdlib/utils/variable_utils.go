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
