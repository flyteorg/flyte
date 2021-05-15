package pluginoverride

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
)

func TestFileConfig(t *testing.T) {
	pluginOverride1 := &admin.PluginOverride{
		TaskType:              "python_task",
		PluginId:              []string{"plugin-override1", "plugin-override2"},
		MissingPluginBehavior: admin.PluginOverride_FAIL,
	}
	pluginOverride2 := &admin.PluginOverride{
		TaskType:              "java_task",
		PluginId:              []string{"plugin-override3", "plugin-override3"},
		MissingPluginBehavior: admin.PluginOverride_USE_DEFAULT,
	}
	pluginOverrides := []*admin.PluginOverride{pluginOverride1, pluginOverride2}

	pluginOverrideFileConfig := FileConfig{
		Project: "dummyProject",
		Domain:  "dummyDomain",
		PluginOverrides: &admin.PluginOverrides{
			Overrides: pluginOverrides,
		},
	}
	matchingAttr := &admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_PluginOverrides{
			PluginOverrides: pluginOverrideFileConfig.PluginOverrides,
		},
	}
	t.Run("decorate", func(t *testing.T) {
		assert.Equal(t, matchingAttr, pluginOverrideFileConfig.Decorate())
	})

	t.Run("decorate", func(t *testing.T) {
		taskAttrFileConfigNew := FileConfig{
			Project: "dummyProject",
			Domain:  "dummyDomain",
		}
		taskAttrFileConfigNew.UnDecorate(matchingAttr)
		assert.Equal(t, pluginOverrideFileConfig, taskAttrFileConfigNew)
	})
	t.Run("get project domain workflow", func(t *testing.T) {
		taskAttrFileConfigNew := FileConfig{
			Project:  "dummyProject",
			Domain:   "dummyDomain",
			Workflow: "workflow",
		}
		assert.Equal(t, "dummyProject", taskAttrFileConfigNew.GetProject())
		assert.Equal(t, "dummyDomain", taskAttrFileConfigNew.GetDomain())
		assert.Equal(t, "workflow", taskAttrFileConfigNew.GetWorkflow())
	})
}
