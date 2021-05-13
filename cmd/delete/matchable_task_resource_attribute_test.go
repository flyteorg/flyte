package delete

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/taskresourceattribute"
	u "github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func deleteTaskResourceAttributeSetup() {
	ctx = u.Ctx
	cmdCtx = u.CmdCtx
	mockClient = u.MockClient
	taskresourceattribute.DefaultDelConfig = &taskresourceattribute.AttrDeleteConfig{}
	args = []string{}
}

func TestDeleteTaskResourceAttributes(t *testing.T) {
	t.Run("successful project domain attribute deletion commandline", func(t *testing.T) {
		setup()
		deleteTaskResourceAttributeSetup()
		// Empty attribute file
		taskresourceattribute.DefaultDelConfig.AttrFile = ""
		// No args implying project domain attribute deletion
		u.DeleterExt.OnDeleteProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil)
		err = deleteTaskResourceAttributes(ctx, args, cmdCtx)
		assert.Nil(t, err)
		u.DeleterExt.AssertCalled(t, "DeleteProjectDomainAttributes",
			ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_TASK_RESOURCE)
	})
	t.Run("failed project domain attribute deletion", func(t *testing.T) {
		setup()
		deleteTaskResourceAttributeSetup()
		// No args implying project domain attribute deletion
		u.DeleterExt.OnDeleteProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(fmt.Errorf("failed to delte project domain attributes"))
		err = deleteTaskResourceAttributes(ctx, args, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to delte project domain attributes"), err)
		u.DeleterExt.AssertCalled(t, "DeleteProjectDomainAttributes",
			ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_TASK_RESOURCE)
	})
	t.Run("successful project domain attribute deletion file", func(t *testing.T) {
		setup()
		deleteTaskResourceAttributeSetup()
		// Empty attribute file
		taskresourceattribute.DefaultDelConfig.AttrFile = "testdata/valid_project_domain_task_attribute.yaml"
		// No args implying project domain attribute deletion
		u.DeleterExt.OnDeleteProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil)
		err = deleteTaskResourceAttributes(ctx, args, cmdCtx)
		assert.Nil(t, err)
		u.DeleterExt.AssertCalled(t, "DeleteProjectDomainAttributes",
			ctx, "flytectldemo", "development", admin.MatchableResource_TASK_RESOURCE)
	})
	t.Run("successful workflow attribute deletion", func(t *testing.T) {
		setup()
		deleteTaskResourceAttributeSetup()
		// Empty attribute file
		taskresourceattribute.DefaultDelConfig.AttrFile = ""
		args := []string{"workflow1"}
		u.DeleterExt.OnDeleteWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil)
		err = deleteTaskResourceAttributes(ctx, args, cmdCtx)
		assert.Nil(t, err)
		u.DeleterExt.AssertCalled(t, "DeleteWorkflowAttributes",
			ctx, config.GetConfig().Project, config.GetConfig().Domain, "workflow1",
			admin.MatchableResource_TASK_RESOURCE)
	})
	t.Run("failed workflow attribute deletion", func(t *testing.T) {
		setup()
		deleteTaskResourceAttributeSetup()
		// Empty attribute file
		taskresourceattribute.DefaultDelConfig.AttrFile = ""
		args := []string{"workflow1"}
		u.DeleterExt.OnDeleteWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(fmt.Errorf("failed to delete workflow attribute"))
		err = deleteTaskResourceAttributes(ctx, args, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to delete workflow attribute"), err)
		u.DeleterExt.AssertCalled(t, "DeleteWorkflowAttributes",
			ctx, config.GetConfig().Project, config.GetConfig().Domain, "workflow1",
			admin.MatchableResource_TASK_RESOURCE)
	})
	t.Run("successful workflow attribute deletion file", func(t *testing.T) {
		setup()
		deleteTaskResourceAttributeSetup()
		// Empty attribute file
		taskresourceattribute.DefaultDelConfig.AttrFile = "testdata/valid_workflow_task_attribute.yaml"
		// No args implying project domain attribute deletion
		u.DeleterExt.OnDeleteWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil)
		err = deleteTaskResourceAttributes(ctx, args, cmdCtx)
		assert.Nil(t, err)
		u.DeleterExt.AssertCalled(t, "DeleteWorkflowAttributes",
			ctx, "flytectldemo", "development", "core.control_flow.run_merge_sort.merge_sort",
			admin.MatchableResource_TASK_RESOURCE)
	})
	t.Run("workflow attribute deletion non existent file", func(t *testing.T) {
		setup()
		deleteTaskResourceAttributeSetup()
		// Empty attribute file
		taskresourceattribute.DefaultDelConfig.AttrFile = "testdata/non-existent"
		// No args implying project domain attribute deletion
		u.DeleterExt.OnDeleteWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil)
		err = deleteTaskResourceAttributes(ctx, args, cmdCtx)
		assert.NotNil(t, err)
		u.DeleterExt.AssertNotCalled(t, "DeleteWorkflowAttributes",
			ctx, "flytectldemo", "development", "core.control_flow.run_merge_sort.merge_sort",
			admin.MatchableResource_TASK_RESOURCE)
	})
	t.Run("attribute deletion invalid file", func(t *testing.T) {
		setup()
		deleteTaskResourceAttributeSetup()
		// Empty attribute file
		taskresourceattribute.DefaultDelConfig.AttrFile = "testdata/invalid_attribute.yaml"
		// No args implying project domain attribute deletion
		err = deleteTaskResourceAttributes(ctx, args, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t,
			fmt.Errorf("error unmarshaling JSON: while decoding JSON: json: unknown field \"InvalidDomain\""),
			err)
		u.DeleterExt.AssertNotCalled(t, "DeleteProjectDomainAttributes",
			ctx, "flytectldemo", "development", admin.MatchableResource_TASK_RESOURCE)
	})
}
