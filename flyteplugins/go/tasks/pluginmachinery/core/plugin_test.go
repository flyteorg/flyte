package core_test

import (
	"context"
	"testing"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"gotest.tools/assert"
)

func TestLoadPlugin(t *testing.T) {
	corePluginType := "core"

	t.Run("valid", func(t *testing.T) {
		corePlugin := &mocks.Plugin{}
		corePlugin.On("GetID").Return(corePluginType)
		corePlugin.OnGetProperties().Return(core.PluginProperties{})

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
		corePlugin.OnGetProperties().Return(core.PluginProperties{
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
		corePlugin.OnGetProperties().Return(core.PluginProperties{
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
		corePlugin.OnGetProperties().Return(core.PluginProperties{
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
