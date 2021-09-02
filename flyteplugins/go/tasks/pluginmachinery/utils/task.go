package utils

import "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

// This function extracts the variable named "results" from the TaskTemplate
func GetResultsVariable(taskTemplate *core.TaskTemplate) (results *core.Variable, exists bool) {
	if taskTemplate == nil {
		return nil, false
	}
	if taskTemplate.Interface == nil {
		return nil, false
	}
	if taskTemplate.Interface.Outputs == nil {
		return nil, false
	}
	if taskTemplate.Interface.Outputs.Variables == nil {
		return nil, false
	}
	for _, e := range taskTemplate.Interface.Outputs.Variables {
		if e.Name == "results" {
			return e.Var, true
		}
	}
	return nil, false
}
