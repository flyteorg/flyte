package taskresourceattribute

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
)

func TestFileConfig(t *testing.T) {
	taskAttrFileConfig := TaskResourceAttrFileConfig{
		Project: "dummyProject",
		Domain:  "dummyDomain",
		TaskResourceAttributes: &admin.TaskResourceAttributes{
			Defaults: &admin.TaskResourceSpec{
				Cpu:    "1",
				Memory: "150Mi",
			},
			Limits: &admin.TaskResourceSpec{
				Cpu:    "2",
				Memory: "350Mi",
			},
		},
	}
	matchingAttr := &admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_TaskResourceAttributes{
			TaskResourceAttributes: taskAttrFileConfig.TaskResourceAttributes,
		},
	}
	t.Run("decorate", func(t *testing.T) {
		assert.Equal(t, matchingAttr, taskAttrFileConfig.Decorate())
	})

	t.Run("decorate", func(t *testing.T) {
		taskAttrFileConfigNew := TaskResourceAttrFileConfig{
			Project: "dummyProject",
			Domain:  "dummyDomain",
		}
		taskAttrFileConfigNew.UnDecorate(matchingAttr)
		assert.Equal(t, taskAttrFileConfig, taskAttrFileConfigNew)
	})
	t.Run("get project domain workflow", func(t *testing.T) {
		taskAttrFileConfigNew := TaskResourceAttrFileConfig{
			Project:  "dummyProject",
			Domain:   "dummyDomain",
			Workflow: "workflow",
		}
		assert.Equal(t, "dummyProject", taskAttrFileConfigNew.GetProject())
		assert.Equal(t, "dummyDomain", taskAttrFileConfigNew.GetDomain())
		assert.Equal(t, "workflow", taskAttrFileConfigNew.GetWorkflow())
	})
}
