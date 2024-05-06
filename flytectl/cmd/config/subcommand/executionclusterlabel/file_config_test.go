package executionclusterlabel

import (
	"testing"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
)

func TestFileConfig(t *testing.T) {
	execClusterLabelFileConfig := FileConfig{
		Project: "dummyProject",
		Domain:  "dummyDomain",
		ExecutionClusterLabel: &admin.ExecutionClusterLabel{
			Value: "foo",
		},
	}
	matchingAttr := &admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_ExecutionClusterLabel{
			ExecutionClusterLabel: execClusterLabelFileConfig.ExecutionClusterLabel,
		},
	}
	t.Run("decorate", func(t *testing.T) {
		assert.Equal(t, matchingAttr, execClusterLabelFileConfig.Decorate())
	})

	t.Run("decorate", func(t *testing.T) {
		taskAttrFileConfigNew := FileConfig{
			Project: "dummyProject",
			Domain:  "dummyDomain",
		}
		taskAttrFileConfigNew.UnDecorate(matchingAttr)
		assert.Equal(t, execClusterLabelFileConfig, taskAttrFileConfigNew)
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
