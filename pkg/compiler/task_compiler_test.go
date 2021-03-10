package compiler

import (
	"testing"

	"github.com/flyteorg/flytepropeller/pkg/compiler/errors"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

func MakeResource(name core.Resources_ResourceName, v string) *core.Resources_ResourceEntry {
	return &core.Resources_ResourceEntry{
		Name:  name,
		Value: v,
	}
}

func TestValidateContainerCommand(t *testing.T) {
	task := core.TaskTemplate{
		Id: &core.Identifier{Name: "task_123"},
		Interface: &core.TypedInterface{
			Inputs: createVariableMap(map[string]*core.Variable{
				"foo": {},
			}),
			Outputs: createEmptyVariableMap(),
		},
		Target: &core.TaskTemplate_Container{
			Container: &core.Container{
				Image: "image://",
			},
		},
	}
	errs := errors.NewCompileErrors()
	assert.False(t, validateContainerCommand(&task, errs))
	assert.Contains(t, errs.Error(), "Node Id: container, Description: Value required [command]")

	task.GetContainer().Command = []string{"cmd"}
	errs = errors.NewCompileErrors()
	assert.True(t, validateContainerCommand(&task, errs))
	assert.False(t, errs.HasErrors())
}

func TestCompileTask(t *testing.T) {
	task, err := CompileTask(&core.TaskTemplate{
		Id: &core.Identifier{Name: "task_123"},
		Interface: &core.TypedInterface{
			Inputs:  createEmptyVariableMap(),
			Outputs: createEmptyVariableMap(),
		},
		Target: &core.TaskTemplate_Container{
			Container: &core.Container{
				Image:   "image://",
				Command: []string{"cmd"},
				Args:    []string{"args"},
				Resources: &core.Resources{
					Requests: []*core.Resources_ResourceEntry{
						MakeResource(core.Resources_CPU, "5"),
					},
					Limits: []*core.Resources_ResourceEntry{
						MakeResource(core.Resources_MEMORY, "100Gi"),
					},
				},
				Env: []*core.KeyValuePair{
					{
						Key:   "Env_Var",
						Value: "Env_Val",
					},
				},
				Config: []*core.KeyValuePair{
					{
						Key:   "config_key",
						Value: "config_value",
					},
				},
			},
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, task)
}
