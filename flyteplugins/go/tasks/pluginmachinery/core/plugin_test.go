package core_test

import (
	"context"
	"testing"

	"gotest.tools/assert"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
)

func TestLoadPlugin(t *testing.T) {
	corePluginType := "core"

	t.Run("valid", func(t *testing.T) {
		corePlugin := &mocks.Plugin{}
		corePlugin.On("GetID").Return(corePluginType)
		corePlugin.EXPECT().GetProperties().Return(core.PluginProperties{})

		corePluginEntry := core.PluginEntry{
			ID:                  corePluginType,
			RegisteredTaskTypes: []core.TaskType{corePluginType},
			LoadPlugin: func(ctx context.Context, iCtx core.SetupContext) (core.Plugin, error) {
				return corePlugin, nil
			},
		}
		setupCtx := mocks.SetupContext{}
		p, err := core.LoadPlugin(context.TODO(), &setupCtx, corePluginEntry)
		assert.NilError(t, err)
		assert.Equal(t, corePluginType, p.GetID())
	})

	t.Run("valid GeneratedNameMaxLength", func(t *testing.T) {
		corePlugin := &mocks.Plugin{}
		corePlugin.On("GetID").Return(corePluginType)
		length := 10
		corePlugin.EXPECT().GetProperties().Return(core.PluginProperties{
			GeneratedNameMaxLength: &length,
		})

		corePluginEntry := core.PluginEntry{
			ID:                  corePluginType,
			RegisteredTaskTypes: []core.TaskType{corePluginType},
			LoadPlugin: func(ctx context.Context, iCtx core.SetupContext) (core.Plugin, error) {
				return corePlugin, nil
			},
		}
		setupCtx := mocks.SetupContext{}
		p, err := core.LoadPlugin(context.TODO(), &setupCtx, corePluginEntry)
		assert.NilError(t, err)
		assert.Equal(t, corePluginType, p.GetID())
	})

	t.Run("valid GeneratedNameMaxLength", func(t *testing.T) {
		corePlugin := &mocks.Plugin{}
		corePlugin.On("GetID").Return(corePluginType)
		length := 10
		corePlugin.EXPECT().GetProperties().Return(core.PluginProperties{
			GeneratedNameMaxLength: &length,
		})

		corePluginEntry := core.PluginEntry{
			ID:                  corePluginType,
			RegisteredTaskTypes: []core.TaskType{corePluginType},
			LoadPlugin: func(ctx context.Context, iCtx core.SetupContext) (core.Plugin, error) {
				return corePlugin, nil
			},
		}
		setupCtx := mocks.SetupContext{}
		_, err := core.LoadPlugin(context.TODO(), &setupCtx, corePluginEntry)
		assert.NilError(t, err)
	})

	t.Run("invalid GeneratedNameMaxLength", func(t *testing.T) {
		corePlugin := &mocks.Plugin{}
		corePlugin.On("GetID").Return(corePluginType)
		length := 5
		corePlugin.EXPECT().GetProperties().Return(core.PluginProperties{
			GeneratedNameMaxLength: &length,
		})

		corePluginEntry := core.PluginEntry{
			ID:                  corePluginType,
			RegisteredTaskTypes: []core.TaskType{corePluginType},
			LoadPlugin: func(ctx context.Context, iCtx core.SetupContext) (core.Plugin, error) {
				return corePlugin, nil
			},
		}
		setupCtx := mocks.SetupContext{}
		_, err := core.LoadPlugin(context.TODO(), &setupCtx, corePluginEntry)
		assert.Error(t, err, "GeneratedNameMaxLength needs to be greater then 8")
	})

}

func TestConnectorService(t *testing.T) {
	connectorService := core.ConnectorService{}
	taskTypes := []core.TaskType{"sensor", "chatgpt"}

	for _, taskType := range taskTypes {
		assert.Equal(t, false, connectorService.ContainTaskType(taskType))
	}

	connectorService.SetSupportedTaskType(taskTypes)
	for _, taskType := range taskTypes {
		assert.Equal(t, true, connectorService.ContainTaskType(taskType))
	}
}
