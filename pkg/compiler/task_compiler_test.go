package compiler

import (
	"testing"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"
	v1 "k8s.io/api/core/v1"

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

func TestValidateK8sPod(t *testing.T) {
	var podSpec = v1.PodSpec{
		Containers: []v1.Container{
			{
				Name: "primary",
			},
			{
				Name: "secondary",
			},
		},
	}
	structObj, _ := utils.MarshalObjToStruct(&podSpec)
	task := core.TaskTemplate{
		Id: &core.Identifier{Name: "task_123"},
		Interface: &core.TypedInterface{
			Inputs: createVariableMap(map[string]*core.Variable{
				"foo": {},
			}),
			Outputs: createEmptyVariableMap(),
		},
		Target: &core.TaskTemplate_K8SPod{
			K8SPod: &core.K8SPod{
				PodSpec: structObj,
			},
		},
	}
	errs := errors.NewCompileErrors()
	assert.True(t, validateK8sPod(&task, errs))

	podSpec = v1.PodSpec{
		Containers: []v1.Container{
			{
				Name: "primary",
			},
			{
				Name: "$Up3R+Invalid",
			},
		},
	}
	structObj, _ = utils.MarshalObjToStruct(&podSpec)
	task.Target = &core.TaskTemplate_K8SPod{
		K8SPod: &core.K8SPod{
			PodSpec: structObj,
		},
	}
	errs = errors.NewCompileErrors()
	assert.False(t, validateK8sPod(&task, errs))
	assert.Contains(t, errs.Error(), "InvalidValue, Node Id: root, Description: Invalid value [k8s pod spec container name]")
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
