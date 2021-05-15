package update

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/config/subcommand/executionclusterlabel"
	u "github.com/flyteorg/flytectl/cmd/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func updateExecutionClusterLabelSetup() {
	ctx = u.Ctx
	cmdCtx = u.CmdCtx
	mockClient = u.MockClient
	executionclusterlabel.DefaultUpdateConfig = &executionclusterlabel.AttrUpdateConfig{}
}

func TestExecutionClusterLabel(t *testing.T) {
	t.Run("no input file for update", func(t *testing.T) {
		setup()
		updateExecutionClusterLabelSetup()
		err = updateExecutionClusterLabelFunc(ctx, args, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("attrFile is mandatory while calling update for execution cluster label"), err)
		tearDownAndVerify(t, ``)
	})
	t.Run("successful update project domain attribute", func(t *testing.T) {
		setup()
		updateExecutionClusterLabelSetup()
		executionclusterlabel.DefaultUpdateConfig.AttrFile = "testdata/valid_project_domain_execution_cluster_label.yaml"
		// No args implying project domain attribute deletion
		u.UpdaterExt.OnUpdateProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil)
		err = updateExecutionClusterLabelFunc(ctx, args, cmdCtx)
		assert.Nil(t, err)
		tearDownAndVerify(t, ``)
	})
	t.Run("failed update project domain attribute", func(t *testing.T) {
		setup()
		updateExecutionClusterLabelSetup()
		executionclusterlabel.DefaultUpdateConfig.AttrFile = "testdata/valid_project_domain_execution_cluster_label.yaml"
		// No args implying project domain attribute deletion
		u.UpdaterExt.OnUpdateProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(fmt.Errorf("failed to update attributes"))
		err = updateExecutionClusterLabelFunc(ctx, args, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to update attributes"), err)
		tearDownAndVerify(t, ``)
	})
	t.Run("successful update workflow attribute", func(t *testing.T) {
		setup()
		updateExecutionClusterLabelSetup()
		executionclusterlabel.DefaultUpdateConfig.AttrFile = "testdata/valid_workflow_execution_cluster_label.yaml"
		// No args implying project domain attribute deletion
		u.UpdaterExt.OnUpdateWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil)
		err = updateExecutionClusterLabelFunc(ctx, nil, cmdCtx)
		assert.Nil(t, err)
		tearDownAndVerify(t, ``)
	})
	t.Run("failed update workflow attribute", func(t *testing.T) {
		setup()
		updateExecutionClusterLabelSetup()
		executionclusterlabel.DefaultUpdateConfig.AttrFile = "testdata/valid_workflow_execution_cluster_label.yaml"
		// No args implying project domain attribute deletion
		u.UpdaterExt.OnUpdateWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(fmt.Errorf("failed to update attributes"))
		err = updateExecutionClusterLabelFunc(ctx, nil, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to update attributes"), err)
		tearDownAndVerify(t, ``)
	})
	t.Run("non existent file", func(t *testing.T) {
		setup()
		updateExecutionClusterLabelSetup()
		executionclusterlabel.DefaultUpdateConfig.AttrFile = testDataNonExistentFile
		err = updateExecutionClusterLabelFunc(ctx, nil, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("unable to read from testdata/non-existent-file yaml file"), err)
		tearDownAndVerify(t, ``)
	})
	t.Run("invalid update file", func(t *testing.T) {
		setup()
		updateExecutionClusterLabelSetup()
		executionclusterlabel.DefaultUpdateConfig.AttrFile = testDataInvalidAttrFile
		err = updateExecutionClusterLabelFunc(ctx, nil, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("error unmarshaling JSON: while decoding JSON: json: unknown field \"InvalidDomain\""), err)
		tearDownAndVerify(t, ``)
	})
}
