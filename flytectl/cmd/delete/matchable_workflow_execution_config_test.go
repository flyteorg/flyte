package delete

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/config/subcommand/workflowexecutionconfig"

	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func deleteWorkflowExecutionConfigSetup() {
	workflowexecutionconfig.DefaultDelConfig = &workflowexecutionconfig.AttrDeleteConfig{}
	args = []string{}
}

func TestDeleteWorkflowExecutionConfig(t *testing.T) {
	t.Run("successful project domain attribute deletion commandline", func(t *testing.T) {
		s := setup()
		deleteWorkflowExecutionConfigSetup()
		// Empty attribute file
		workflowexecutionconfig.DefaultDelConfig.AttrFile = ""
		// No args implying project domain attribute deletion
		s.DeleterExt.OnDeleteProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil)
		err := deleteWorkflowExecutionConfig(s.Ctx, []string{}, s.CmdCtx)
		assert.Nil(t, err)
		s.DeleterExt.AssertCalled(t, "DeleteProjectDomainAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG)
	})
	t.Run("failed project domain attribute deletion", func(t *testing.T) {
		s := setup()
		deleteWorkflowExecutionConfigSetup()
		// No args implying project domain attribute deletion
		s.DeleterExt.OnDeleteProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(fmt.Errorf("failed to delete project domain attributes"))
		err := deleteWorkflowExecutionConfig(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to delete project domain attributes"), err)
		s.DeleterExt.AssertCalled(t, "DeleteProjectDomainAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG)
	})
	t.Run("successful project domain attribute deletion file", func(t *testing.T) {
		s := setup()
		deleteWorkflowExecutionConfigSetup()
		// Empty attribute file
		workflowexecutionconfig.DefaultDelConfig.AttrFile = "testdata/valid_project_domain_workflow_execution_config.yaml"
		// No args implying project domain attribute deletion
		s.DeleterExt.OnDeleteProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil)
		err := deleteWorkflowExecutionConfig(s.Ctx, []string{}, s.CmdCtx)
		assert.Nil(t, err)
		s.DeleterExt.AssertCalled(t, "DeleteProjectDomainAttributes",
			s.Ctx, "flytesnacks", "development", admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG)
	})
	t.Run("successful workflow attribute deletion", func(t *testing.T) {
		s := setup()
		deleteWorkflowExecutionConfigSetup()
		// Empty attribute file
		workflowexecutionconfig.DefaultDelConfig.AttrFile = ""
		args := []string{"workflow1"}
		s.DeleterExt.OnDeleteWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil)
		err := deleteWorkflowExecutionConfig(s.Ctx, args, s.CmdCtx)
		assert.Nil(t, err)
		s.DeleterExt.AssertCalled(t, "DeleteWorkflowAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, "workflow1",
			admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG)
	})
	t.Run("failed workflow attribute deletion", func(t *testing.T) {
		s := setup()
		deleteWorkflowExecutionConfigSetup()
		// Empty attribute file
		workflowexecutionconfig.DefaultDelConfig.AttrFile = ""
		args := []string{"workflow1"}
		s.DeleterExt.OnDeleteWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(fmt.Errorf("failed to delete workflow attribute"))
		err := deleteWorkflowExecutionConfig(s.Ctx, args, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to delete workflow attribute"), err)
		s.DeleterExt.AssertCalled(t, "DeleteWorkflowAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, "workflow1",
			admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG)
	})
	t.Run("successful workflow attribute deletion file", func(t *testing.T) {
		s := setup()
		deleteWorkflowExecutionConfigSetup()
		// Empty attribute file
		workflowexecutionconfig.DefaultDelConfig.AttrFile = "testdata/valid_workflow_workflow_execution_config.yaml"
		// No args implying project domain attribute deletion
		s.DeleterExt.OnDeleteWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil)
		err := deleteWorkflowExecutionConfig(s.Ctx, []string{}, s.CmdCtx)
		assert.Nil(t, err)
		s.DeleterExt.AssertCalled(t, "DeleteWorkflowAttributes",
			s.Ctx, "flytesnacks", "development", "core.control_flow.merge_sort.merge_sort",
			admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG)
	})
	t.Run("workflow attribute deletion non existent file", func(t *testing.T) {
		s := setup()
		deleteWorkflowExecutionConfigSetup()
		// Empty attribute file
		workflowexecutionconfig.DefaultDelConfig.AttrFile = testDataNonExistentFile
		// No args implying project domain attribute deletion
		s.DeleterExt.OnDeleteWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil)
		err := deleteWorkflowExecutionConfig(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		s.DeleterExt.AssertNotCalled(t, "DeleteWorkflowAttributes",
			s.Ctx, "flytesnacks", "development", "core.control_flow.merge_sort.merge_sort",
			admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG)
	})
	t.Run("attribute deletion invalid file", func(t *testing.T) {
		s := setup()
		deleteWorkflowExecutionConfigSetup()
		// Empty attribute file
		workflowexecutionconfig.DefaultDelConfig.AttrFile = testDataInvalidAttrFile
		// No args implying project domain attribute deletion
		err := deleteWorkflowExecutionConfig(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t,
			fmt.Errorf("error unmarshaling JSON: while decoding JSON: json: unknown field \"InvalidDomain\""),
			err)
		s.DeleterExt.AssertNotCalled(t, "DeleteProjectDomainAttributes",
			s.Ctx, "flytesnacks", "development", admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG)
	})
}
