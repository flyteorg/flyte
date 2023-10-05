package update

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flytectl/cmd/config/subcommand/taskresourceattribute"
)

func updateTaskResourceAttributeSetup() {
	taskresourceattribute.DefaultUpdateConfig = &taskresourceattribute.AttrUpdateConfig{}
}

func TestUpdateTaskResourceAttributes(t *testing.T) {
	t.Run("no input file for update", func(t *testing.T) {
		s := setup()
		updateTaskResourceAttributeSetup()
		err := updateTaskResourceAttributesFunc(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("attrFile is mandatory while calling update for task resource attribute"), err)
		s.TearDownAndVerify(t, ``)
	})
	t.Run("successful update project domain attribute", func(t *testing.T) {
		s := setup()
		updateTaskResourceAttributeSetup()
		taskresourceattribute.DefaultUpdateConfig.AttrFile = "testdata/valid_project_domain_task_attribute.yaml"
		// No args implying project domain attribute deletion
		s.UpdaterExt.OnUpdateProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil)
		err := updateTaskResourceAttributesFunc(s.Ctx, []string{}, s.CmdCtx)
		assert.Nil(t, err)
		s.TearDownAndVerify(t, `Updated attributes from flytesnacks project and domain development`)
	})
	t.Run("failed update project domain attribute", func(t *testing.T) {
		s := setup()
		updateTaskResourceAttributeSetup()
		taskresourceattribute.DefaultUpdateConfig.AttrFile = "testdata/valid_project_domain_task_attribute.yaml"
		// No args implying project domain attribute deletion
		s.UpdaterExt.OnUpdateProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(fmt.Errorf("failed to update attributes"))
		err := updateTaskResourceAttributesFunc(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to update attributes"), err)
		s.TearDownAndVerify(t, ``)
	})
	t.Run("successful update workflow attribute", func(t *testing.T) {
		s := setup()
		updateTaskResourceAttributeSetup()
		taskresourceattribute.DefaultUpdateConfig.AttrFile = "testdata/valid_workflow_task_attribute.yaml"
		// No args implying project domain attribute deletion
		s.UpdaterExt.OnUpdateWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil)
		err := updateTaskResourceAttributesFunc(s.Ctx, []string{}, s.CmdCtx)
		assert.Nil(t, err)
		s.TearDownAndVerify(t, `Updated attributes from flytesnacks project and domain development and workflow core.control_flow.merge_sort.merge_sort`)
	})
	t.Run("failed update workflow attribute", func(t *testing.T) {
		s := setup()
		updateTaskResourceAttributeSetup()
		taskresourceattribute.DefaultUpdateConfig.AttrFile = "testdata/valid_workflow_task_attribute.yaml"
		// No args implying project domain attribute deletion
		s.UpdaterExt.OnUpdateWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(fmt.Errorf("failed to update attributes"))
		err := updateTaskResourceAttributesFunc(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to update attributes"), err)
		s.TearDownAndVerify(t, ``)
	})
	t.Run("non existent file", func(t *testing.T) {
		s := setup()
		updateTaskResourceAttributeSetup()
		taskresourceattribute.DefaultUpdateConfig.AttrFile = testDataNonExistentFile
		err := updateTaskResourceAttributesFunc(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("unable to read from testdata/non-existent-file yaml file"), err)
		s.TearDownAndVerify(t, ``)
	})
	t.Run("invalid update file", func(t *testing.T) {
		s := setup()
		updateTaskResourceAttributeSetup()
		taskresourceattribute.DefaultUpdateConfig.AttrFile = testDataInvalidAttrFile
		err := updateTaskResourceAttributesFunc(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("error unmarshaling JSON: while decoding JSON: json: unknown field \"InvalidDomain\""), err)
		s.TearDownAndVerify(t, ``)
	})
}
