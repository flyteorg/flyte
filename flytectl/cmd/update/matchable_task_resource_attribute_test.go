package update

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/config/subcommand/taskresourceattribute"
	u "github.com/flyteorg/flytectl/cmd/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func updateTaskResourceAttributeSetup() {
	ctx = u.Ctx
	cmdCtx = u.CmdCtx
	mockClient = u.MockClient
	taskresourceattribute.DefaultUpdateConfig = &taskresourceattribute.AttrUpdateConfig{}
}

func TestUpdateTaskResourceAttributes(t *testing.T) {
	t.Run("no input file for update", func(t *testing.T) {
		setup()
		updateTaskResourceAttributeSetup()
		err = updateTaskResourceAttributesFunc(ctx, args, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("attrFile is mandatory while calling update for task resource attribute"), err)
		tearDownAndVerify(t, ``)
	})
	t.Run("successful update project domain attribute", func(t *testing.T) {
		setup()
		updateTaskResourceAttributeSetup()
		taskresourceattribute.DefaultUpdateConfig.AttrFile = "testdata/valid_project_domain_task_attribute.yaml"
		// No args implying project domain attribute deletion
		u.UpdaterExt.OnUpdateProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil)
		err = updateTaskResourceAttributesFunc(ctx, args, cmdCtx)
		assert.Nil(t, err)
		tearDownAndVerify(t, ``)
	})
	t.Run("failed update project domain attribute", func(t *testing.T) {
		setup()
		updateTaskResourceAttributeSetup()
		taskresourceattribute.DefaultUpdateConfig.AttrFile = "testdata/valid_project_domain_task_attribute.yaml"
		// No args implying project domain attribute deletion
		u.UpdaterExt.OnUpdateProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(fmt.Errorf("failed to update attributes"))
		err = updateTaskResourceAttributesFunc(ctx, args, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to update attributes"), err)
		tearDownAndVerify(t, ``)
	})
	t.Run("successful update workflow attribute", func(t *testing.T) {
		setup()
		updateTaskResourceAttributeSetup()
		taskresourceattribute.DefaultUpdateConfig.AttrFile = "testdata/valid_workflow_task_attribute.yaml"
		// No args implying project domain attribute deletion
		u.UpdaterExt.OnUpdateWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil)
		err = updateTaskResourceAttributesFunc(ctx, nil, cmdCtx)
		assert.Nil(t, err)
		tearDownAndVerify(t, ``)
	})
	t.Run("failed update workflow attribute", func(t *testing.T) {
		setup()
		updateTaskResourceAttributeSetup()
		taskresourceattribute.DefaultUpdateConfig.AttrFile = "testdata/valid_workflow_task_attribute.yaml"
		// No args implying project domain attribute deletion
		u.UpdaterExt.OnUpdateWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(fmt.Errorf("failed to update attributes"))
		err = updateTaskResourceAttributesFunc(ctx, nil, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to update attributes"), err)
		tearDownAndVerify(t, ``)
	})
	t.Run("non existent file", func(t *testing.T) {
		setup()
		updateTaskResourceAttributeSetup()
		taskresourceattribute.DefaultUpdateConfig.AttrFile = "testdata/non-existent-filel"
		err = updateTaskResourceAttributesFunc(ctx, nil, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("unable to read from testdata/non-existent-filel yaml file"), err)
		tearDownAndVerify(t, ``)
	})
	t.Run("invalid update file", func(t *testing.T) {
		setup()
		updateTaskResourceAttributeSetup()
		taskresourceattribute.DefaultUpdateConfig.AttrFile = "testdata/invalid_attribute.yaml"
		err = updateTaskResourceAttributesFunc(ctx, nil, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("error unmarshaling JSON: while decoding JSON: json: unknown field \"InvalidDomain\""), err)
		tearDownAndVerify(t, ``)
	})
}
