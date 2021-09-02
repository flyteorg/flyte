package utils

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

func TestGetResultsVariable(t *testing.T) {
	t.Run("variable not found with nil pointer", func(t *testing.T) {
		emptyTaskTemplate := &core.TaskTemplate{
			Interface: nil,
		}
		_, found := GetResultsVariable(emptyTaskTemplate)
		assert.False(t, found)
	})

	t.Run("variable not found", func(t *testing.T) {
		taskTemplate := &core.TaskTemplate{
			Interface: &core.TypedInterface{
				Outputs: &core.VariableMap{
					Variables: []*core.VariableMapEntry{
						{
							Name: "o0",
						},
					},
				},
			},
		}
		_, found := GetResultsVariable(taskTemplate)
		assert.False(t, found)
	})

	t.Run("happy case", func(t *testing.T) {
		taskTemplate := &core.TaskTemplate{
			Interface: &core.TypedInterface{
				Outputs: &core.VariableMap{
					Variables: []*core.VariableMapEntry{
						{
							Name: "results",
							Var: &core.Variable{
								Description: "athena result",
							},
						},
					},
				},
			},
		}
		v, found := GetResultsVariable(taskTemplate)
		assert.True(t, found)
		assert.Equal(t, "athena result", v.Description)
	})
}
