package update

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/config/subcommand/clusterresourceattribute"
	u "github.com/flyteorg/flytectl/cmd/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func updateClusterResourceAttributeSetup() {
	ctx = u.Ctx
	cmdCtx = u.CmdCtx
	mockClient = u.MockClient
	clusterresourceattribute.DefaultUpdateConfig = &clusterresourceattribute.AttrUpdateConfig{}
}

func TestUpdateClusterResourceAttributes(t *testing.T) {
	t.Run("no input file for update", func(t *testing.T) {
		setup()
		updateClusterResourceAttributeSetup()
		err = updateClusterResourceAttributesFunc(ctx, args, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("attrFile is mandatory while calling update for cluster resource attribute"), err)
		tearDownAndVerify(t, ``)
	})
	t.Run("successfully updated project domain attribute", func(t *testing.T) {
		setup()
		updateClusterResourceAttributeSetup()
		clusterresourceattribute.DefaultUpdateConfig.AttrFile = "testdata/valid_project_domain_cluster_attribute.yaml"
		// No args implying project domain attribute deletion
		u.UpdaterExt.OnUpdateProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil)
		err = updateClusterResourceAttributesFunc(ctx, args, cmdCtx)
		assert.Nil(t, err)
		tearDownAndVerify(t, `Updated attributes from flytectldemo project and domain development`)
	})
	t.Run("failed to update project domain attribute", func(t *testing.T) {
		setup()
		updateClusterResourceAttributeSetup()
		clusterresourceattribute.DefaultUpdateConfig.AttrFile = "testdata/valid_project_domain_cluster_attribute.yaml"
		// No args implying project domain attribute deletion
		u.UpdaterExt.OnUpdateProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(fmt.Errorf("failed to update attributes"))
		err = updateClusterResourceAttributesFunc(ctx, args, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to update attributes"), err)
		tearDownAndVerify(t, ``)
	})
	t.Run("successfully updated workflow attribute", func(t *testing.T) {
		setup()
		updateClusterResourceAttributeSetup()
		clusterresourceattribute.DefaultUpdateConfig.AttrFile = "testdata/valid_workflow_cluster_attribute.yaml"
		// No args implying project domain attribute deletion
		u.UpdaterExt.OnUpdateWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil)
		err = updateClusterResourceAttributesFunc(ctx, nil, cmdCtx)
		assert.Nil(t, err)
		tearDownAndVerify(t, `Updated attributes from flytectldemo project and domain development and workflow core.control_flow.run_merge_sort.merge_sort`)
	})
	t.Run("failed to update workflow attribute", func(t *testing.T) {
		setup()
		updateClusterResourceAttributeSetup()
		clusterresourceattribute.DefaultUpdateConfig.AttrFile = "testdata/valid_workflow_cluster_attribute.yaml"
		// No args implying project domain attribute deletion
		u.UpdaterExt.OnUpdateWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(fmt.Errorf("failed to update attributes"))
		err = updateClusterResourceAttributesFunc(ctx, nil, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to update attributes"), err)
		tearDownAndVerify(t, ``)
	})
	t.Run("non existent file", func(t *testing.T) {
		setup()
		updateClusterResourceAttributeSetup()
		clusterresourceattribute.DefaultUpdateConfig.AttrFile = testDataNonExistentFile
		err = updateClusterResourceAttributesFunc(ctx, nil, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("unable to read from testdata/non-existent-file yaml file"), err)
		tearDownAndVerify(t, ``)
	})
	t.Run("invalid update file", func(t *testing.T) {
		setup()
		updateClusterResourceAttributeSetup()
		clusterresourceattribute.DefaultUpdateConfig.AttrFile = testDataInvalidAttrFile
		err = updateClusterResourceAttributesFunc(ctx, nil, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("error unmarshaling JSON: while decoding JSON: json: unknown field \"InvalidDomain\""), err)
		tearDownAndVerify(t, ``)
	})
}
