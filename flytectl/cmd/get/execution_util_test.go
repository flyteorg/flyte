package get

import (
	"testing"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestTaskInputs(t *testing.T) {
	taskInputs := map[string]*core.Variable{}
	t.Run("nil task", func(t *testing.T) {
		retValue := TaskInputs(nil)
		assert.Equal(t, taskInputs, retValue)
	})
	t.Run("valid inputs", func(t *testing.T) {
		task := createTask()
		retValue := TaskInputs(task)
		assert.Equal(t, task.Closure.CompiledTask.Template.Interface.Inputs.Variables, retValue)
	})
	t.Run("closure  compiled task nil", func(t *testing.T) {
		task := createTask()
		task.Closure.CompiledTask = nil
		retValue := TaskInputs(task)
		assert.Equal(t, taskInputs, retValue)
	})
	t.Run("closure  compiled task template nil", func(t *testing.T) {
		task := createTask()
		task.Closure.CompiledTask.Template = nil
		retValue := TaskInputs(task)
		assert.Equal(t, taskInputs, retValue)
	})
	t.Run("closure  compiled task template interface nil", func(t *testing.T) {
		task := createTask()
		task.Closure.CompiledTask.Template.Interface = nil
		retValue := TaskInputs(task)
		assert.Equal(t, taskInputs, retValue)
	})
	t.Run("closure  compiled task template interface input nil", func(t *testing.T) {
		task := createTask()
		task.Closure.CompiledTask.Template.Interface.Inputs = nil
		retValue := TaskInputs(task)
		assert.Equal(t, taskInputs, retValue)
	})
}

func createTask() *admin.Task {
	sortedListLiteralType := core.Variable{
		Type: &core.LiteralType{
			Type: &core.LiteralType_CollectionType{
				CollectionType: &core.LiteralType{
					Type: &core.LiteralType_Simple{
						Simple: core.SimpleType_INTEGER,
					},
				},
			},
		},
	}

	variableMap := map[string]*core.Variable{
		"sorted_list1": &sortedListLiteralType,
		"sorted_list2": &sortedListLiteralType,
	}

	inputs := &core.VariableMap{
		Variables: variableMap,
	}
	typedInterface := &core.TypedInterface{
		Inputs: inputs,
	}
	taskTemplate := &core.TaskTemplate{
		Interface: typedInterface,
	}
	compiledTask := &core.CompiledTask{
		Template: taskTemplate,
	}
	return &admin.Task{
		Id: &core.Identifier{
			Name:    "task1",
			Version: "v2",
		},
		Closure: &admin.TaskClosure{
			CreatedAt:    &timestamppb.Timestamp{Seconds: 1, Nanos: 0},
			CompiledTask: compiledTask,
		},
	}
}
