package delete

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/executionqueueattribute"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func deleteExecutionQueueAttributeSetup() {
	executionqueueattribute.DefaultDelConfig = &executionqueueattribute.AttrDeleteConfig{}
	args = []string{}
}

func TestDeleteExecutionQueueAttributes(t *testing.T) {
	t.Run("successful project domain attribute deletion commandline", func(t *testing.T) {
		s := setup()
		deleteExecutionQueueAttributeSetup()
		// Empty attribute file
		executionqueueattribute.DefaultDelConfig.AttrFile = ""
		// No args implying project domain attribute deletion
		s.DeleterExt.OnDeleteProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil)
		err := deleteExecutionQueueAttributes(s.Ctx, []string{}, s.CmdCtx)
		assert.Nil(t, err)
		s.DeleterExt.AssertCalled(t, "DeleteProjectDomainAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_EXECUTION_QUEUE)
	})
	t.Run("failed project domain attribute deletion", func(t *testing.T) {
		s := setup()
		deleteTaskResourceAttributeSetup()
		// No args implying project domain attribute deletion
		s.DeleterExt.OnDeleteProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(fmt.Errorf("failed to delte project domain attributes"))
		err := deleteExecutionQueueAttributes(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to delte project domain attributes"), err)
		s.DeleterExt.AssertCalled(t, "DeleteProjectDomainAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_EXECUTION_QUEUE)
	})
	t.Run("successful project domain attribute deletion file", func(t *testing.T) {
		s := setup()
		deleteTaskResourceAttributeSetup()
		// Empty attribute file
		executionqueueattribute.DefaultDelConfig.AttrFile = "testdata/valid_project_domain_execution_queue_attribute.yaml"
		// No args implying project domain attribute deletion
		s.DeleterExt.OnDeleteProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil)
		err := deleteExecutionQueueAttributes(s.Ctx, []string{}, s.CmdCtx)
		assert.Nil(t, err)
		s.DeleterExt.AssertCalled(t, "DeleteProjectDomainAttributes",
			s.Ctx, "flytectldemo", "development", admin.MatchableResource_EXECUTION_QUEUE)
	})
	t.Run("successful workflow attribute deletion", func(t *testing.T) {
		s := setup()
		deleteTaskResourceAttributeSetup()
		// Empty attribute file
		executionqueueattribute.DefaultDelConfig.AttrFile = ""
		args := []string{"workflow1"}
		s.DeleterExt.OnDeleteWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil)
		err := deleteExecutionQueueAttributes(s.Ctx, args, s.CmdCtx)
		assert.Nil(t, err)
		s.DeleterExt.AssertCalled(t, "DeleteWorkflowAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, "workflow1",
			admin.MatchableResource_EXECUTION_QUEUE)
	})
	t.Run("failed workflow attribute deletion", func(t *testing.T) {
		s := setup()
		deleteTaskResourceAttributeSetup()
		// Empty attribute file
		executionqueueattribute.DefaultDelConfig.AttrFile = ""
		args := []string{"workflow1"}
		s.DeleterExt.OnDeleteWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(fmt.Errorf("failed to delete workflow attribute"))
		err := deleteExecutionQueueAttributes(s.Ctx, args, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to delete workflow attribute"), err)
		s.DeleterExt.AssertCalled(t, "DeleteWorkflowAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, "workflow1",
			admin.MatchableResource_EXECUTION_QUEUE)
	})
	t.Run("successful workflow attribute deletion file", func(t *testing.T) {
		s := setup()
		deleteTaskResourceAttributeSetup()
		// Empty attribute file
		executionqueueattribute.DefaultDelConfig.AttrFile = "testdata/valid_workflow_execution_queue_attribute.yaml"
		// No args implying project domain attribute deletion
		s.DeleterExt.OnDeleteWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil)
		err := deleteExecutionQueueAttributes(s.Ctx, []string{}, s.CmdCtx)
		assert.Nil(t, err)
		s.DeleterExt.AssertCalled(t, "DeleteWorkflowAttributes",
			s.Ctx, "flytectldemo", "development", "core.control_flow.run_merge_sort.merge_sort",
			admin.MatchableResource_EXECUTION_QUEUE)
	})
	t.Run("workflow attribute deletion non existent file", func(t *testing.T) {
		s := setup()
		deleteTaskResourceAttributeSetup()
		// Empty attribute file
		executionqueueattribute.DefaultDelConfig.AttrFile = testDataNonExistentFile
		// No args implying project domain attribute deletion
		s.DeleterExt.OnDeleteWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil)
		err := deleteExecutionQueueAttributes(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		s.DeleterExt.AssertNotCalled(t, "DeleteWorkflowAttributes",
			s.Ctx, "flytectldemo", "development", "core.control_flow.run_merge_sort.merge_sort",
			admin.MatchableResource_EXECUTION_QUEUE)
	})
	t.Run("attribute deletion invalid file", func(t *testing.T) {
		s := setup()
		deleteTaskResourceAttributeSetup()
		// Empty attribute file
		executionqueueattribute.DefaultDelConfig.AttrFile = testDataInvalidAttrFile
		// No args implying project domain attribute deletion
		err := deleteExecutionQueueAttributes(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t,
			fmt.Errorf("error unmarshaling JSON: while decoding JSON: json: unknown field \"InvalidDomain\""),
			err)
		s.DeleterExt.AssertNotCalled(t, "DeleteProjectDomainAttributes",
			s.Ctx, "flytectldemo", "development", admin.MatchableResource_EXECUTION_QUEUE)
	})
}
