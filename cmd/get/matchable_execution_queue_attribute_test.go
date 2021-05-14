package get

import (
	"fmt"
	"os"
	"testing"

	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/executionqueueattribute"
	u "github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getExecutionQueueAttributeSetup() {
	ctx = u.Ctx
	cmdCtx = u.CmdCtx
	mockClient = u.MockClient
	executionqueueattribute.DefaultFetchConfig = &executionqueueattribute.AttrFetchConfig{}
	// Clean up the temp directory.
	_ = os.Remove(testDataTempFile)
}

func TestGetExecutionQueueAttributes(t *testing.T) {
	executionQueueAttr := &admin.ExecutionQueueAttributes{
		Tags: []string{"foo", "bar"},
	}
	projectDomainResp := &admin.ProjectDomainAttributesGetResponse{
		Attributes: &admin.ProjectDomainAttributes{
			Project: config.GetConfig().Project,
			Domain:  config.GetConfig().Domain,
			MatchingAttributes: &admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
					ExecutionQueueAttributes: executionQueueAttr,
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
				Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
					ExecutionQueueAttributes: executionQueueAttr,
				},
			},
		},
	}
	t.Run("successful get project domain attribute", func(t *testing.T) {
		var args []string
		setup()
		getExecutionQueueAttributeSetup()
		// No args implying project domain attribute deletion
		u.FetcherExt.OnFetchProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(projectDomainResp, nil)
		err = getExecutionQueueAttributes(ctx, args, cmdCtx)
		assert.Nil(t, err)
		u.FetcherExt.AssertCalled(t, "FetchProjectDomainAttributes",
			ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_EXECUTION_QUEUE)
		tearDownAndVerify(t, `{"project":"dummyProject","domain":"dummyDomain","tags":["foo","bar"]}`)
	})
	t.Run("successful get project domain attribute and write to file", func(t *testing.T) {
		var args []string
		setup()
		getExecutionQueueAttributeSetup()
		executionqueueattribute.DefaultFetchConfig.AttrFile = testDataTempFile
		// No args implying project domain attribute deletion
		u.FetcherExt.OnFetchProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(projectDomainResp, nil)
		err = getExecutionQueueAttributes(ctx, args, cmdCtx)
		assert.Nil(t, err)
		u.FetcherExt.AssertCalled(t, "FetchProjectDomainAttributes",
			ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_EXECUTION_QUEUE)
		tearDownAndVerify(t, `wrote the config to file temp-output-file`)
	})
	t.Run("successful get project domain attribute and write to file failure", func(t *testing.T) {
		var args []string
		setup()
		getExecutionQueueAttributeSetup()
		executionqueueattribute.DefaultFetchConfig.AttrFile = testDataNotExistentTempFile
		// No args implying project domain attribute deletion
		u.FetcherExt.OnFetchProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(projectDomainResp, nil)
		err = getExecutionQueueAttributes(ctx, args, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("error dumping in file due to open non-existent-dir/temp-output-file: no such file or directory"), err)
		u.FetcherExt.AssertCalled(t, "FetchProjectDomainAttributes",
			ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_EXECUTION_QUEUE)
		tearDownAndVerify(t, ``)
	})
	t.Run("failed get project domain attribute", func(t *testing.T) {
		var args []string
		setup()
		getExecutionQueueAttributeSetup()
		// No args implying project domain attribute deletion
		u.FetcherExt.OnFetchProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil, fmt.Errorf("failed to fetch response"))
		err = getExecutionQueueAttributes(ctx, args, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to fetch response"), err)
		u.FetcherExt.AssertCalled(t, "FetchProjectDomainAttributes",
			ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_EXECUTION_QUEUE)
		tearDownAndVerify(t, ``)
	})
	t.Run("successful get workflow attribute", func(t *testing.T) {
		var args []string
		setup()
		getExecutionQueueAttributeSetup()
		args = []string{"workflow"}
		u.FetcherExt.OnFetchWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(workflowResp, nil)
		err = getExecutionQueueAttributes(ctx, args, cmdCtx)
		assert.Nil(t, err)
		u.FetcherExt.AssertCalled(t, "FetchWorkflowAttributes",
			ctx, config.GetConfig().Project, config.GetConfig().Domain, "workflow",
			admin.MatchableResource_EXECUTION_QUEUE)
		tearDownAndVerify(t, `{"project":"dummyProject","domain":"dummyDomain","workflow":"workflow","tags":["foo","bar"]}`)
	})
	t.Run("failed get workflow attribute", func(t *testing.T) {
		var args []string
		setup()
		getExecutionQueueAttributeSetup()
		args = []string{"workflow"}
		u.FetcherExt.OnFetchWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed to fetch response"))
		err = getExecutionQueueAttributes(ctx, args, cmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to fetch response"), err)
		u.FetcherExt.AssertCalled(t, "FetchWorkflowAttributes",
			ctx, config.GetConfig().Project, config.GetConfig().Domain, "workflow",
			admin.MatchableResource_EXECUTION_QUEUE)
		tearDownAndVerify(t, ``)
	})
}
