package executionqueueattribute

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
)

func TestFileConfig(t *testing.T) {
	executionQueueAttrFileConfig := AttrFileConfig{
		Project: "dummyProject",
		Domain:  "dummyDomain",
		ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
			Tags: []string{"foo", "bar"},
		},
	}
	matchingAttr := &admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
			ExecutionQueueAttributes: executionQueueAttrFileConfig.ExecutionQueueAttributes,
		},
	}
	t.Run("decorate", func(t *testing.T) {
		assert.Equal(t, matchingAttr, executionQueueAttrFileConfig.Decorate())
	})

	t.Run("decorate", func(t *testing.T) {
		executionAttrFileConfigNew := AttrFileConfig{
			Project: "dummyProject",
			Domain:  "dummyDomain",
		}
		executionAttrFileConfigNew.UnDecorate(matchingAttr)
		assert.Equal(t, executionQueueAttrFileConfig, executionAttrFileConfigNew)
	})
	t.Run("get project domain workflow", func(t *testing.T) {
		executionQueueAttrFileConfigNew := AttrFileConfig{
			Project:  "dummyProject",
			Domain:   "dummyDomain",
			Workflow: "workflow",
		}
		assert.Equal(t, "dummyProject", executionQueueAttrFileConfigNew.GetProject())
		assert.Equal(t, "dummyDomain", executionQueueAttrFileConfigNew.GetDomain())
		assert.Equal(t, "workflow", executionQueueAttrFileConfigNew.GetWorkflow())
	})
}
