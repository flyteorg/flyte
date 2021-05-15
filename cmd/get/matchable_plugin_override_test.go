package get

import (
	"fmt"
	"os"
	"testing"

	"github.com/flyteorg/flytectl/cmd/config"
	pluginoverride "github.com/flyteorg/flytectl/cmd/config/subcommand/plugin_override"
	u "github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getPluginOverrideSetup() {
	ctx = u.Ctx
	cmdCtx = u.CmdCtx
	mockClient = u.MockClient
	pluginoverride.DefaultFetchConfig = &pluginoverride.AttrFetchConfig{}
	// Clean up the temp directory.
	_ = os.Remove(testDataTempFile)
}

func TestGetPluginOverride(t *testing.T) {
	pluginOverride1 := &admin.PluginOverride{
		TaskType:              "python_task",
		PluginId:              []string{"plugin-override1", "plugin-override2"},
		MissingPluginBehavior: admin.PluginOverride_FAIL,
	}
	pluginOverride2 := &admin.PluginOverride{
		TaskType:              "java_task",
		PluginId:              []string{"plugin-override3", "plugin-override3"},
		MissingPluginBehavior: admin.PluginOverride_USE_DEFAULT,
	}
	pluginOverrides := []*admin.PluginOverride{pluginOverride1, pluginOverride2}
	projectDomainResp := &admin.ProjectDomainAttributesGetResponse{
		Attributes: &admin.ProjectDomainAttributes{
			Project: config.GetConfig().Project,
			Domain:  config.GetConfig().Domain,
			MatchingAttributes: &admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_PluginOverrides{
					PluginOverrides: &admin.PluginOverrides{
						Overrides: pluginOverrides,
					},
				},
			},
		},
	}
	workflowResp := &admin.WorkflowAttributesGetResponse{
		Attributes: &admin.WorkflowAttributes{
			Project:  config.GetConfig().Project,
			Domain:   config.GetConfig().Domain,
			Workflow: "workflow",
			MatchingAttributes: &admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_PluginOverrides{
					PluginOverrides: &admin.PluginOverrides{
						Overrides: pluginOverrides,
					},
				},
			},
		},
	}
	t.Run("successful get project domain attribute", func(t *testing.T) {
		var args []string
		setup()
		getPluginOverrideSetup()
		// No args implying project domain attribute deletion
		u.FetcherExt.OnFetchProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(projectDomainResp, nil)
		err = getPluginOverridesFunc(ctx, args, cmdCtx)
		assert.Nil(t, err)
		u.FetcherExt.AssertCalled(t, "FetchProjectDomainAttributes",
			ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_PLUGIN_OVERRIDE)
		tearDownAndVerify(t, `{"project":"dummyProject","domain":"dummyDomain","overrides":[{"task_type":"python_task","plugin_id":["plugin-override1","plugin-override2"]},{"task_type":"java_task","plugin_id":["plugin-override3","plugin-override3"],"missing_plugin_behavior":1}]}`)
	})
	t.Run("successful get project domain attribute and write to file", func(t *testing.T) {
		var args []string
		setup()
		getPluginOverrideSetup()
		pluginoverride.DefaultFetchConfig.AttrFile = testDataTempFile
		// No args implying project domain attribute deletion
		u.FetcherExt.OnFetchProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(projectDomainResp, nil)
		err = getPluginOverridesFunc(ctx, args, cmdCtx)
		assert.Nil(t, err)
		u.FetcherExt.AssertCalled(t, "FetchProjectDomainAttributes",
			ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_PLUGIN_OVERRIDE)
		tearDownAndVerify(t, `wrote the config to file temp-output-file`)
	})
	t.Run("successful get project domain attribute and write to file failure", func(t *testing.T) {
		var args []string
		setup()
		getPluginOverrideSetup()
		pluginoverride.DefaultFetchConfig.AttrFile = testDataNotExistentTempFile
		// No args implying project domain attribute deletion
		u.FetcherExt.OnFetchProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(projectDomainResp, nil)
		err = getPluginOverridesFunc(ctx, args, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("error dumping in file due to open non-existent-dir/temp-output-file: no such file or directory"), err)
		u.FetcherExt.AssertCalled(t, "FetchProjectDomainAttributes",
			ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_PLUGIN_OVERRIDE)
		tearDownAndVerify(t, ``)
	})
	t.Run("failed get project domain attribute", func(t *testing.T) {
		var args []string
		setup()
		getPluginOverrideSetup()
		// No args implying project domain attribute deletion
		u.FetcherExt.OnFetchProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil, fmt.Errorf("failed to fetch response"))
		err = getPluginOverridesFunc(ctx, args, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to fetch response"), err)
		u.FetcherExt.AssertCalled(t, "FetchProjectDomainAttributes",
			ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_PLUGIN_OVERRIDE)
		tearDownAndVerify(t, ``)
	})
	t.Run("successful get workflow attribute", func(t *testing.T) {
		var args []string
		setup()
		getPluginOverrideSetup()
		args = []string{"workflow"}
		u.FetcherExt.OnFetchWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(workflowResp, nil)
		err = getPluginOverridesFunc(ctx, args, cmdCtx)
		assert.Nil(t, err)
		u.FetcherExt.AssertCalled(t, "FetchWorkflowAttributes",
			ctx, config.GetConfig().Project, config.GetConfig().Domain, "workflow", admin.MatchableResource_PLUGIN_OVERRIDE)
		tearDownAndVerify(t, `{"project":"dummyProject","domain":"dummyDomain","workflow":"workflow","overrides":[{"task_type":"python_task","plugin_id":["plugin-override1","plugin-override2"]},{"task_type":"java_task","plugin_id":["plugin-override3","plugin-override3"],"missing_plugin_behavior":1}]}`)
	})
	t.Run("failed get workflow attribute", func(t *testing.T) {
		var args []string
		setup()
		getPluginOverrideSetup()
		args = []string{"workflow"}
		u.FetcherExt.OnFetchWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed to fetch response"))
		err = getPluginOverridesFunc(ctx, args, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to fetch response"), err)
		u.FetcherExt.AssertCalled(t, "FetchWorkflowAttributes",
			ctx, config.GetConfig().Project, config.GetConfig().Domain, "workflow", admin.MatchableResource_PLUGIN_OVERRIDE)
		tearDownAndVerify(t, ``)
	})
}
