package update

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/config/subcommand/workflowexecutionconfig"

	u "github.com/flyteorg/flytectl/cmd/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func updateWorkflowExecutionConfigSetup() {
	ctx = u.Ctx
	cmdCtx = u.CmdCtx
	mockClient = u.MockClient
	workflowexecutionconfig.DefaultUpdateConfig = &workflowexecutionconfig.AttrUpdateConfig{}
}

func TestWorkflowExecutionConfigs(t *testing.T) {
	t.Run("no input file for update", func(t *testing.T) {
		setup()
		updateWorkflowExecutionConfigSetup()
		err = updateWorkflowExecutionConfigFunc(ctx, args, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("attrFile is mandatory while calling update for workflow execution config"), err)
		tearDownAndVerify(t, ``)
	})
	t.Run("successful update project domain attribute", func(t *testing.T) {
		setup()
		updateWorkflowExecutionConfigSetup()
		workflowexecutionconfig.DefaultUpdateConfig.AttrFile = "testdata/valid_project_domain_workflow_execution_config.yaml"
		// No args implying project domain attribute deletion
		u.UpdaterExt.OnUpdateProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil)
		err = updateWorkflowExecutionConfigFunc(ctx, args, cmdCtx)
		assert.Nil(t, err)
		tearDownAndVerify(t, `Updated attributes from flytectldemo project and domain development`)
	})
	t.Run("failed update project domain attribute", func(t *testing.T) {
		setup()
		updateWorkflowExecutionConfigSetup()
		workflowexecutionconfig.DefaultUpdateConfig.AttrFile = "testdata/valid_project_domain_workflow_execution_config.yaml"
		// No args implying project domain attribute deletion
		u.UpdaterExt.OnUpdateProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(fmt.Errorf("failed to update attributes"))
		err = updateWorkflowExecutionConfigFunc(ctx, args, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to update attributes"), err)
		tearDownAndVerify(t, ``)
	})
	t.Run("successful update workflow attribute", func(t *testing.T) {
		setup()
		updateWorkflowExecutionConfigSetup()
		workflowexecutionconfig.DefaultUpdateConfig.AttrFile = "testdata/valid_workflow_workflow_execution_config.yaml"
		// No args implying project domain attribute deletion
		u.UpdaterExt.OnUpdateWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil)
		err = updateWorkflowExecutionConfigFunc(ctx, nil, cmdCtx)
		assert.Nil(t, err)
		tearDownAndVerify(t, `Updated attributes from flytectldemo project and domain development and workflow core.control_flow.run_merge_sort.merge_sort`)
	})
	t.Run("failed update workflow attribute", func(t *testing.T) {
		setup()
		updateWorkflowExecutionConfigSetup()
		workflowexecutionconfig.DefaultUpdateConfig.AttrFile = "testdata/valid_workflow_workflow_execution_config.yaml"
		// No args implying project domain attribute deletion
		u.UpdaterExt.OnUpdateWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(fmt.Errorf("failed to update attributes"))
		err = updateWorkflowExecutionConfigFunc(ctx, nil, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to update attributes"), err)
		tearDownAndVerify(t, ``)
	})
	t.Run("non existent file", func(t *testing.T) {
		setup()
		updateWorkflowExecutionConfigSetup()
		workflowexecutionconfig.DefaultUpdateConfig.AttrFile = testDataNonExistentFile
		err = updateWorkflowExecutionConfigFunc(ctx, nil, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("unable to read from testdata/non-existent-file yaml file"), err)
		tearDownAndVerify(t, ``)
	})
	t.Run("invalid update file", func(t *testing.T) {
		setup()
		updateWorkflowExecutionConfigSetup()
		workflowexecutionconfig.DefaultUpdateConfig.AttrFile = testDataInvalidAttrFile
		err = updateWorkflowExecutionConfigFunc(ctx, nil, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("error unmarshaling JSON: while decoding JSON: json: unknown field \"InvalidDomain\""), err)
		tearDownAndVerify(t, ``)
	})
}
