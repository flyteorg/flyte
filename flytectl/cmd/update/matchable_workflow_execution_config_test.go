package update

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/config/subcommand/workflowexecutionconfig"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func updateWorkflowExecutionConfigSetup() {
	workflowexecutionconfig.DefaultUpdateConfig = &workflowexecutionconfig.AttrUpdateConfig{}
}

func TestWorkflowExecutionConfigs(t *testing.T) {
	t.Run("no input file for update", func(t *testing.T) {
		s := setup()
		updateWorkflowExecutionConfigSetup()
		err := updateWorkflowExecutionConfigFunc(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("attrFile is mandatory while calling update for workflow execution config"), err)
		s.TearDownAndVerify(t, ``)
	})
	t.Run("successful update project domain attribute", func(t *testing.T) {
		s := setup()
		updateWorkflowExecutionConfigSetup()
		workflowexecutionconfig.DefaultUpdateConfig.AttrFile = "testdata/valid_project_domain_workflow_execution_config.yaml"
		// No args implying project domain attribute deletion
		s.UpdaterExt.OnUpdateProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil)
		err := updateWorkflowExecutionConfigFunc(s.Ctx, []string{}, s.CmdCtx)
		assert.Nil(t, err)
		s.TearDownAndVerify(t, `Updated attributes from flytesnacks project and domain development`)
	})
	t.Run("failed update project domain attribute", func(t *testing.T) {
		s := setup()
		updateWorkflowExecutionConfigSetup()
		workflowexecutionconfig.DefaultUpdateConfig.AttrFile = "testdata/valid_project_domain_workflow_execution_config.yaml"
		// No args implying project domain attribute deletion
		s.UpdaterExt.OnUpdateProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(fmt.Errorf("failed to update attributes"))
		err := updateWorkflowExecutionConfigFunc(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to update attributes"), err)
		s.TearDownAndVerify(t, ``)
	})
	t.Run("successful update workflow attribute", func(t *testing.T) {
		s := setup()
		updateWorkflowExecutionConfigSetup()
		workflowexecutionconfig.DefaultUpdateConfig.AttrFile = "testdata/valid_workflow_workflow_execution_config.yaml"
		// No args implying project domain attribute deletion
		s.UpdaterExt.OnUpdateWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil)
		err := updateWorkflowExecutionConfigFunc(s.Ctx, []string{}, s.CmdCtx)
		assert.Nil(t, err)
		s.TearDownAndVerify(t, `Updated attributes from flytesnacks project and domain development and workflow core.control_flow.merge_sort.merge_sort`)
	})
	t.Run("failed update workflow attribute", func(t *testing.T) {
		s := setup()
		updateWorkflowExecutionConfigSetup()
		workflowexecutionconfig.DefaultUpdateConfig.AttrFile = "testdata/valid_workflow_workflow_execution_config.yaml"
		// No args implying project domain attribute deletion
		s.UpdaterExt.OnUpdateWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(fmt.Errorf("failed to update attributes"))
		err := updateWorkflowExecutionConfigFunc(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to update attributes"), err)
		s.TearDownAndVerify(t, ``)
	})
	t.Run("non existent file", func(t *testing.T) {
		s := setup()
		updateWorkflowExecutionConfigSetup()
		workflowexecutionconfig.DefaultUpdateConfig.AttrFile = testDataNonExistentFile
		err := updateWorkflowExecutionConfigFunc(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("unable to read from testdata/non-existent-file yaml file"), err)
		s.TearDownAndVerify(t, ``)
	})
	t.Run("invalid update file", func(t *testing.T) {
		s := setup()
		updateWorkflowExecutionConfigSetup()
		workflowexecutionconfig.DefaultUpdateConfig.AttrFile = testDataInvalidAttrFile
		err := updateWorkflowExecutionConfigFunc(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("error unmarshaling JSON: while decoding JSON: json: unknown field \"InvalidDomain\""), err)
		s.TearDownAndVerify(t, ``)
	})
}
