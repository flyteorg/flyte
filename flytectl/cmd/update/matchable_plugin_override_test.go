package update

import (
	"fmt"
	"testing"

	pluginoverride "github.com/flyteorg/flytectl/cmd/config/subcommand/plugin_override"
	u "github.com/flyteorg/flytectl/cmd/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func updatePluginOverrideSetup() {
	ctx = u.Ctx
	cmdCtx = u.CmdCtx
	mockClient = u.MockClient
	pluginoverride.DefaultUpdateConfig = &pluginoverride.AttrUpdateConfig{}
}

func TestPluginOverride(t *testing.T) {
	t.Run("no input file for update", func(t *testing.T) {
		setup()
		updatePluginOverrideSetup()
		err = updatePluginOverridesFunc(ctx, args, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("attrFile is mandatory while calling update for plugin override"), err)
		tearDownAndVerify(t, ``)
	})
	t.Run("successful update project domain attribute", func(t *testing.T) {
		setup()
		updatePluginOverrideSetup()
		pluginoverride.DefaultUpdateConfig.AttrFile = "testdata/valid_project_domain_plugin_override.yaml"
		// No args implying project domain attribute deletion
		u.UpdaterExt.OnUpdateProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil)
		err = updatePluginOverridesFunc(ctx, args, cmdCtx)
		assert.Nil(t, err)
		tearDownAndVerify(t, `Updated attributes from flytectldemo project and domain development`)
	})
	t.Run("failed update project domain attribute", func(t *testing.T) {
		setup()
		updatePluginOverrideSetup()
		pluginoverride.DefaultUpdateConfig.AttrFile = "testdata/valid_project_domain_plugin_override.yaml"
		// No args implying project domain attribute deletion
		u.UpdaterExt.OnUpdateProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(fmt.Errorf("failed to update attributes"))
		err = updatePluginOverridesFunc(ctx, args, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to update attributes"), err)
		tearDownAndVerify(t, ``)
	})
	t.Run("successful update workflow attribute", func(t *testing.T) {
		setup()
		updatePluginOverrideSetup()
		pluginoverride.DefaultUpdateConfig.AttrFile = "testdata/valid_workflow_plugin_override.yaml"
		// No args implying project domain attribute deletion
		u.UpdaterExt.OnUpdateWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil)
		err = updatePluginOverridesFunc(ctx, nil, cmdCtx)
		assert.Nil(t, err)
		tearDownAndVerify(t, `Updated attributes from flytectldemo project and domain development and workflow core.control_flow.run_merge_sort.merge_sort`)
	})
	t.Run("failed update workflow attribute", func(t *testing.T) {
		setup()
		updatePluginOverrideSetup()
		pluginoverride.DefaultUpdateConfig.AttrFile = "testdata/valid_workflow_plugin_override.yaml"
		// No args implying project domain attribute deletion
		u.UpdaterExt.OnUpdateWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(fmt.Errorf("failed to update attributes"))
		err = updatePluginOverridesFunc(ctx, nil, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to update attributes"), err)
		tearDownAndVerify(t, ``)
	})
	t.Run("non existent file", func(t *testing.T) {
		setup()
		updatePluginOverrideSetup()
		pluginoverride.DefaultUpdateConfig.AttrFile = testDataNonExistentFile
		err = updatePluginOverridesFunc(ctx, nil, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("unable to read from testdata/non-existent-file yaml file"), err)
		tearDownAndVerify(t, ``)
	})
	t.Run("invalid update file", func(t *testing.T) {
		setup()
		updatePluginOverrideSetup()
		pluginoverride.DefaultUpdateConfig.AttrFile = testDataInvalidAttrFile
		err = updatePluginOverridesFunc(ctx, nil, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("error unmarshaling JSON: while decoding JSON: json: unknown field \"InvalidDomain\""), err)
		tearDownAndVerify(t, ``)
	})
}
