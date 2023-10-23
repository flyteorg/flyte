package delete

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytectl/cmd/config"
	pluginoverride "github.com/flyteorg/flytectl/cmd/config/subcommand/plugin_override"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func deletePluginOverrideSetup() {
	pluginoverride.DefaultDelConfig = &pluginoverride.AttrDeleteConfig{}
	args = []string{}
}

func TestPluginOverride(t *testing.T) {
	t.Run("successful project domain attribute deletion commandline", func(t *testing.T) {
		s := setup()
		deletePluginOverrideSetup()
		// Empty attribute file
		pluginoverride.DefaultDelConfig.AttrFile = ""
		// No args implying project domain attribute deletion
		s.DeleterExt.OnDeleteProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil)
		err := deletePluginOverride(s.Ctx, []string{}, s.CmdCtx)
		assert.Nil(t, err)
		s.DeleterExt.AssertCalled(t, "DeleteProjectDomainAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_PLUGIN_OVERRIDE)
	})
	t.Run("failed project domain attribute deletion", func(t *testing.T) {
		s := setup()
		deletePluginOverrideSetup()
		// No args implying project domain attribute deletion
		s.DeleterExt.OnDeleteProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(fmt.Errorf("failed to delte project domain attributes"))
		err := deletePluginOverride(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to delte project domain attributes"), err)
		s.DeleterExt.AssertCalled(t, "DeleteProjectDomainAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_PLUGIN_OVERRIDE)
	})
	t.Run("successful project domain attribute deletion file", func(t *testing.T) {
		s := setup()
		deletePluginOverrideSetup()
		// Empty attribute file
		pluginoverride.DefaultDelConfig.AttrFile = "testdata/valid_project_domain_plugin_override.yaml"
		// No args implying project domain attribute deletion
		s.DeleterExt.OnDeleteProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil)
		err := deletePluginOverride(s.Ctx, []string{}, s.CmdCtx)
		assert.Nil(t, err)
		s.DeleterExt.AssertCalled(t, "DeleteProjectDomainAttributes",
			s.Ctx, "flytesnacks", "development", admin.MatchableResource_PLUGIN_OVERRIDE)
	})
	t.Run("successful workflow attribute deletion", func(t *testing.T) {
		s := setup()
		deletePluginOverrideSetup()
		// Empty attribute file
		pluginoverride.DefaultDelConfig.AttrFile = ""
		args := []string{"workflow1"}
		s.DeleterExt.OnDeleteWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil)
		err := deletePluginOverride(s.Ctx, args, s.CmdCtx)
		assert.Nil(t, err)
		s.DeleterExt.AssertCalled(t, "DeleteWorkflowAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, "workflow1",
			admin.MatchableResource_PLUGIN_OVERRIDE)
	})
	t.Run("failed workflow attribute deletion", func(t *testing.T) {
		s := setup()
		deletePluginOverrideSetup()
		// Empty attribute file
		pluginoverride.DefaultDelConfig.AttrFile = ""
		args := []string{"workflow1"}
		s.DeleterExt.OnDeleteWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(fmt.Errorf("failed to delete workflow attribute"))
		err := deletePluginOverride(s.Ctx, args, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to delete workflow attribute"), err)
		s.DeleterExt.AssertCalled(t, "DeleteWorkflowAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, "workflow1",
			admin.MatchableResource_PLUGIN_OVERRIDE)
	})
	t.Run("successful workflow attribute deletion file", func(t *testing.T) {
		s := setup()
		deletePluginOverrideSetup()
		// Empty attribute file
		pluginoverride.DefaultDelConfig.AttrFile = "testdata/valid_workflow_plugin_override.yaml"
		// No args implying project domain attribute deletion
		s.DeleterExt.OnDeleteWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil)
		err := deletePluginOverride(s.Ctx, []string{}, s.CmdCtx)
		assert.Nil(t, err)
		s.DeleterExt.AssertCalled(t, "DeleteWorkflowAttributes",
			s.Ctx, "flytesnacks", "development", "core.control_flow.merge_sort.merge_sort",
			admin.MatchableResource_PLUGIN_OVERRIDE)
	})
	t.Run("workflow attribute deletion non existent file", func(t *testing.T) {
		s := setup()
		deletePluginOverrideSetup()
		// Empty attribute file
		pluginoverride.DefaultDelConfig.AttrFile = testDataNonExistentFile
		// No args implying project domain attribute deletion
		s.DeleterExt.OnDeleteWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil)
		err := deletePluginOverride(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		s.DeleterExt.AssertNotCalled(t, "DeleteWorkflowAttributes",
			s.Ctx, "flytesnacks", "development", "core.control_flow.merge_sort.merge_sort",
			admin.MatchableResource_PLUGIN_OVERRIDE)
	})
	t.Run("attribute deletion invalid file", func(t *testing.T) {
		s := setup()
		deletePluginOverrideSetup()
		// Empty attribute file
		pluginoverride.DefaultDelConfig.AttrFile = testDataInvalidAttrFile
		// No args implying project domain attribute deletion
		err := deletePluginOverride(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t,
			fmt.Errorf("error unmarshaling JSON: while decoding JSON: json: unknown field \"InvalidDomain\""),
			err)
		s.DeleterExt.AssertNotCalled(t, "DeleteProjectDomainAttributes",
			s.Ctx, "flytesnacks", "development", admin.MatchableResource_PLUGIN_OVERRIDE)
	})
}
