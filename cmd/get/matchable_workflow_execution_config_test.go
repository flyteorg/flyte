package get

import (
	"fmt"
	"os"
	"testing"

	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/workflowexecutionconfig"
	u "github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getWorkflowExecutionConfigSetup() {
	ctx = u.Ctx
	cmdCtx = u.CmdCtx
	mockClient = u.MockClient
	workflowexecutionconfig.DefaultFetchConfig = &workflowexecutionconfig.AttrFetchConfig{}
	// Clean up the temp directory.
	_ = os.Remove(testDataTempFile)
}

func TestGetWorkflowExecutionConfig(t *testing.T) {
	executionClusterLabel := &admin.WorkflowExecutionConfig{
		MaxParallelism: 5,
	}
	projectDomainResp := &admin.ProjectDomainAttributesGetResponse{
		Attributes: &admin.ProjectDomainAttributes{
			Project: config.GetConfig().Project,
			Domain:  config.GetConfig().Domain,
			MatchingAttributes: &admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_WorkflowExecutionConfig{
					WorkflowExecutionConfig: executionClusterLabel,
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
				Target: &admin.MatchingAttributes_WorkflowExecutionConfig{
					WorkflowExecutionConfig: executionClusterLabel,
				},
			},
		},
	}
	t.Run("successful get project domain attribute", func(t *testing.T) {
		var args []string
		setup()
		getWorkflowExecutionConfigSetup()
		// No args implying project domain attribute deletion
		u.FetcherExt.OnFetchProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(projectDomainResp, nil)
		err = getWorkflowExecutionConfigFunc(ctx, args, cmdCtx)
		assert.Nil(t, err)
		u.FetcherExt.AssertCalled(t, "FetchProjectDomainAttributes",
			ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG)
		tearDownAndVerify(t, `{"project":"dummyProject","domain":"dummyDomain","max_parallelism":5}`)
	})
	t.Run("successful get project domain attribute and write to file", func(t *testing.T) {
		var args []string
		setup()
		getWorkflowExecutionConfigSetup()
		workflowexecutionconfig.DefaultFetchConfig.AttrFile = testDataTempFile
		// No args implying project domain attribute deletion
		u.FetcherExt.OnFetchProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(projectDomainResp, nil)
		err = getWorkflowExecutionConfigFunc(ctx, args, cmdCtx)
		assert.Nil(t, err)
		u.FetcherExt.AssertCalled(t, "FetchProjectDomainAttributes",
			ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG)
		tearDownAndVerify(t, `wrote the config to file temp-output-file`)
	})
	t.Run("successful get project domain attribute and write to file failure", func(t *testing.T) {
		var args []string
		setup()
		getWorkflowExecutionConfigSetup()
		workflowexecutionconfig.DefaultFetchConfig.AttrFile = testDataNotExistentTempFile
		// No args implying project domain attribute deletion
		u.FetcherExt.OnFetchProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(projectDomainResp, nil)
		err = getWorkflowExecutionConfigFunc(ctx, args, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("error dumping in file due to open non-existent-dir/temp-output-file: no such file or directory"), err)
		u.FetcherExt.AssertCalled(t, "FetchProjectDomainAttributes",
			ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG)
		tearDownAndVerify(t, ``)
	})
	t.Run("failed get project domain attribute", func(t *testing.T) {
		var args []string
		setup()
		getWorkflowExecutionConfigSetup()
		// No args implying project domain attribute deletion
		u.FetcherExt.OnFetchProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil, fmt.Errorf("failed to fetch response"))
		err = getWorkflowExecutionConfigFunc(ctx, args, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to fetch response"), err)
		u.FetcherExt.AssertCalled(t, "FetchProjectDomainAttributes",
			ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG)
		tearDownAndVerify(t, ``)
	})
	t.Run("successful get workflow attribute", func(t *testing.T) {
		var args []string
		setup()
		getWorkflowExecutionConfigSetup()
		args = []string{"workflow"}
		u.FetcherExt.OnFetchWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(workflowResp, nil)
		err = getWorkflowExecutionConfigFunc(ctx, args, cmdCtx)
		assert.Nil(t, err)
		u.FetcherExt.AssertCalled(t, "FetchWorkflowAttributes",
			ctx, config.GetConfig().Project, config.GetConfig().Domain, "workflow",
			admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG)
		tearDownAndVerify(t, `{"project":"dummyProject","domain":"dummyDomain","workflow":"workflow","max_parallelism":5}`)
	})
	t.Run("failed get workflow attribute", func(t *testing.T) {
		var args []string
		setup()
		getWorkflowExecutionConfigSetup()
		args = []string{"workflow"}
		u.FetcherExt.OnFetchWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed to fetch response"))
		err = getWorkflowExecutionConfigFunc(ctx, args, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to fetch response"), err)
		u.FetcherExt.AssertCalled(t, "FetchWorkflowAttributes",
			ctx, config.GetConfig().Project, config.GetConfig().Domain, "workflow",
			admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG)
		tearDownAndVerify(t, ``)
	})
}
