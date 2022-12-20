package update

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/config/subcommand/clusterresourceattribute"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func updateClusterResourceAttributeSetup() {
	clusterresourceattribute.DefaultUpdateConfig = &clusterresourceattribute.AttrUpdateConfig{}
}

func TestUpdateClusterResourceAttributes(t *testing.T) {
	t.Run("no input file for update", func(t *testing.T) {
		s := setup()
		updateClusterResourceAttributeSetup()
		err := updateClusterResourceAttributesFunc(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("attrFile is mandatory while calling update for cluster resource attribute"), err)
		tearDownAndVerify(t, s.Writer, ``)
	})
	t.Run("successfully updated project domain attribute", func(t *testing.T) {
		s := setup()
		updateClusterResourceAttributeSetup()
		clusterresourceattribute.DefaultUpdateConfig.AttrFile = "testdata/valid_project_domain_cluster_attribute.yaml"
		// No args implying project domain attribute deletion
		s.UpdaterExt.OnUpdateProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil)
		err := updateClusterResourceAttributesFunc(s.Ctx, []string{}, s.CmdCtx)
		assert.Nil(t, err)
		tearDownAndVerify(t, s.Writer, `Updated attributes from flytesnacks project and domain development`)
	})
	t.Run("failed to update project domain attribute", func(t *testing.T) {
		s := setup()
		updateClusterResourceAttributeSetup()
		clusterresourceattribute.DefaultUpdateConfig.AttrFile = "testdata/valid_project_domain_cluster_attribute.yaml"
		// No args implying project domain attribute deletion
		s.UpdaterExt.OnUpdateProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(fmt.Errorf("failed to update attributes"))
		err := updateClusterResourceAttributesFunc(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to update attributes"), err)
		tearDownAndVerify(t, s.Writer, ``)
	})
	t.Run("successfully updated workflow attribute", func(t *testing.T) {
		s := setup()
		updateClusterResourceAttributeSetup()
		clusterresourceattribute.DefaultUpdateConfig.AttrFile = "testdata/valid_workflow_cluster_attribute.yaml"
		// No args implying project domain attribute deletion
		s.UpdaterExt.OnUpdateWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil)
		err := updateClusterResourceAttributesFunc(s.Ctx, nil, s.CmdCtx)
		assert.Nil(t, err)
		tearDownAndVerify(t, s.Writer, `Updated attributes from flytesnacks project and domain development and workflow core.control_flow.merge_sort.merge_sort`)
	})
	t.Run("failed to update workflow attribute", func(t *testing.T) {
		s := setup()
		updateClusterResourceAttributeSetup()
		clusterresourceattribute.DefaultUpdateConfig.AttrFile = "testdata/valid_workflow_cluster_attribute.yaml"
		// No args implying project domain attribute deletion
		s.UpdaterExt.OnUpdateWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(fmt.Errorf("failed to update attributes"))
		err := updateClusterResourceAttributesFunc(s.Ctx, nil, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to update attributes"), err)
		tearDownAndVerify(t, s.Writer, ``)
	})
	t.Run("non existent file", func(t *testing.T) {
		s := setup()
		updateClusterResourceAttributeSetup()
		clusterresourceattribute.DefaultUpdateConfig.AttrFile = testDataNonExistentFile
		err := updateClusterResourceAttributesFunc(s.Ctx, nil, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("unable to read from testdata/non-existent-file yaml file"), err)
		tearDownAndVerify(t, s.Writer, ``)
	})
	t.Run("invalid update file", func(t *testing.T) {
		s := setup()
		updateClusterResourceAttributeSetup()
		clusterresourceattribute.DefaultUpdateConfig.AttrFile = testDataInvalidAttrFile
		err := updateClusterResourceAttributesFunc(s.Ctx, nil, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("error unmarshaling JSON: while decoding JSON: json: unknown field \"InvalidDomain\""), err)
		tearDownAndVerify(t, s.Writer, ``)
	})
}
