package workflowexecutionconfig

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
)

func TestFileConfig(t *testing.T) {
	workflowExecutionConfigFileConfig := FileConfig{
		Project: "dummyProject",
		Domain:  "dummyDomain",
		WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{
			MaxParallelism: 5,
		},
	}
	matchingAttr := &admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_WorkflowExecutionConfig{
			WorkflowExecutionConfig: workflowExecutionConfigFileConfig.WorkflowExecutionConfig,
		},
	}
	t.Run("decorate", func(t *testing.T) {
		assert.Equal(t, matchingAttr, workflowExecutionConfigFileConfig.Decorate())
	})

	t.Run("decorate", func(t *testing.T) {
		workflowExecutionConfigFileConfigNew := FileConfig{
			Project: "dummyProject",
			Domain:  "dummyDomain",
		}
		workflowExecutionConfigFileConfigNew.UnDecorate(matchingAttr)
		assert.Equal(t, workflowExecutionConfigFileConfig, workflowExecutionConfigFileConfigNew)
	})
	t.Run("get project domain workflow", func(t *testing.T) {
		workflowExecutionConfigFileConfigNew := FileConfig{
			Project:  "dummyProject",
			Domain:   "dummyDomain",
			Workflow: "workflow",
		}
		assert.Equal(t, "dummyProject", workflowExecutionConfigFileConfigNew.GetProject())
		assert.Equal(t, "dummyDomain", workflowExecutionConfigFileConfigNew.GetDomain())
		assert.Equal(t, "workflow", workflowExecutionConfigFileConfigNew.GetWorkflow())
	})
}
