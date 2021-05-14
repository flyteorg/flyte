package update

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/config/subcommand/executionqueueattribute"
	u "github.com/flyteorg/flytectl/cmd/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func updateExecutionQueueAttributeSetup() {
	ctx = u.Ctx
	cmdCtx = u.CmdCtx
	mockClient = u.MockClient
	executionqueueattribute.DefaultUpdateConfig = &executionqueueattribute.AttrUpdateConfig{}
}

func TestExecutionQueueAttributes(t *testing.T) {
	t.Run("no input file for update", func(t *testing.T) {
		setup()
		updateExecutionQueueAttributeSetup()
		err = updateExecutionQueueAttributesFunc(ctx, args, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("attrFile is mandatory while calling update for execution queue attribute"), err)
		tearDownAndVerify(t, ``)
	})
	t.Run("successful update project domain attribute", func(t *testing.T) {
		setup()
		updateExecutionQueueAttributeSetup()
		executionqueueattribute.DefaultUpdateConfig.AttrFile = "testdata/valid_project_domain_execution_queue_attribute.yaml"
		// No args implying project domain attribute deletion
		u.UpdaterExt.OnUpdateProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil)
		err = updateExecutionQueueAttributesFunc(ctx, args, cmdCtx)
		assert.Nil(t, err)
		tearDownAndVerify(t, ``)
	})
	t.Run("failed update project domain attribute", func(t *testing.T) {
		setup()
		updateExecutionQueueAttributeSetup()
		executionqueueattribute.DefaultUpdateConfig.AttrFile = "testdata/valid_project_domain_execution_queue_attribute.yaml"
		// No args implying project domain attribute deletion
		u.UpdaterExt.OnUpdateProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(fmt.Errorf("failed to update attributes"))
		err = updateExecutionQueueAttributesFunc(ctx, args, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to update attributes"), err)
		tearDownAndVerify(t, ``)
	})
	t.Run("successful update workflow attribute", func(t *testing.T) {
		setup()
		updateExecutionQueueAttributeSetup()
		executionqueueattribute.DefaultUpdateConfig.AttrFile = "testdata/valid_workflow_execution_queue_attribute.yaml"
		// No args implying project domain attribute deletion
		u.UpdaterExt.OnUpdateWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil)
		err = updateExecutionQueueAttributesFunc(ctx, nil, cmdCtx)
		assert.Nil(t, err)
		tearDownAndVerify(t, ``)
	})
	t.Run("failed update workflow attribute", func(t *testing.T) {
		setup()
		updateExecutionQueueAttributeSetup()
		executionqueueattribute.DefaultUpdateConfig.AttrFile = "testdata/valid_workflow_execution_queue_attribute.yaml"
		// No args implying project domain attribute deletion
		u.UpdaterExt.OnUpdateWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(fmt.Errorf("failed to update attributes"))
		err = updateExecutionQueueAttributesFunc(ctx, nil, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to update attributes"), err)
		tearDownAndVerify(t, ``)
	})
	t.Run("non existent file", func(t *testing.T) {
		setup()
		updateExecutionQueueAttributeSetup()
		executionqueueattribute.DefaultUpdateConfig.AttrFile = testDataNonExistentFile
		err = updateExecutionQueueAttributesFunc(ctx, nil, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("unable to read from testdata/non-existent-file yaml file"), err)
		tearDownAndVerify(t, ``)
	})
	t.Run("invalid update file", func(t *testing.T) {
		setup()
		updateExecutionQueueAttributeSetup()
		executionqueueattribute.DefaultUpdateConfig.AttrFile = testDataInvalidAttrFile
		err = updateExecutionQueueAttributesFunc(ctx, nil, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("error unmarshaling JSON: while decoding JSON: json: unknown field \"InvalidDomain\""), err)
		tearDownAndVerify(t, ``)
	})
}
