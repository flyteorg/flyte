package get

import (
	"fmt"
	"os"
	"testing"

	"github.com/flyteorg/flytectl/cmd/testutils"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/executionqueueattribute"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getExecutionQueueAttributeSetup() {
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
		s := testutils.SetupWithExt()
		getExecutionQueueAttributeSetup()
		// No args implying project domain attribute deletion
		s.FetcherExt.OnFetchProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(projectDomainResp, nil)
		err := getExecutionQueueAttributes(s.Ctx, []string{}, s.CmdCtx)
		assert.Nil(t, err)
		s.FetcherExt.AssertCalled(t, "FetchProjectDomainAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_EXECUTION_QUEUE)
		s.TearDownAndVerify(t, `{"project":"dummyProject","domain":"dummyDomain","tags":["foo","bar"]}`)
	})
	t.Run("successful get project domain attribute and write to file", func(t *testing.T) {
		s := testutils.SetupWithExt()
		getExecutionQueueAttributeSetup()
		executionqueueattribute.DefaultFetchConfig.AttrFile = testDataTempFile
		// No args implying project domain attribute deletion
		s.FetcherExt.OnFetchProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(projectDomainResp, nil)
		err := getExecutionQueueAttributes(s.Ctx, []string{}, s.CmdCtx)
		assert.Nil(t, err)
		s.FetcherExt.AssertCalled(t, "FetchProjectDomainAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_EXECUTION_QUEUE)
		s.TearDownAndVerify(t, `wrote the config to file temp-output-file`)
	})
	t.Run("successful get project domain attribute and write to file failure", func(t *testing.T) {
		s := testutils.SetupWithExt()
		getExecutionQueueAttributeSetup()
		executionqueueattribute.DefaultFetchConfig.AttrFile = testDataNotExistentTempFile
		// No args implying project domain attribute deletion
		s.FetcherExt.OnFetchProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(projectDomainResp, nil)
		err := getExecutionQueueAttributes(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("error dumping in file due to open non-existent-dir/temp-output-file: no such file or directory"), err)
		s.FetcherExt.AssertCalled(t, "FetchProjectDomainAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_EXECUTION_QUEUE)
		s.TearDownAndVerify(t, ``)
	})
	t.Run("failed get project domain attribute", func(t *testing.T) {
		s := testutils.SetupWithExt()
		getExecutionQueueAttributeSetup()
		// No args implying project domain attribute deletion
		s.FetcherExt.OnFetchProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil, fmt.Errorf("failed to fetch response"))
		err := getExecutionQueueAttributes(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to fetch response"), err)
		s.FetcherExt.AssertCalled(t, "FetchProjectDomainAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_EXECUTION_QUEUE)
		s.TearDownAndVerify(t, ``)
	})
	t.Run("successful get workflow attribute", func(t *testing.T) {
		s := testutils.SetupWithExt()
		getExecutionQueueAttributeSetup()
		args := []string{"workflow"}
		s.FetcherExt.OnFetchWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(workflowResp, nil)
		err := getExecutionQueueAttributes(s.Ctx, args, s.CmdCtx)
		assert.Nil(t, err)
		s.FetcherExt.AssertCalled(t, "FetchWorkflowAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, "workflow",
			admin.MatchableResource_EXECUTION_QUEUE)
		s.TearDownAndVerify(t, `{"project":"dummyProject","domain":"dummyDomain","workflow":"workflow","tags":["foo","bar"]}`)
	})
	t.Run("failed get workflow attribute", func(t *testing.T) {
		s := testutils.SetupWithExt()
		getExecutionQueueAttributeSetup()
		args := []string{"workflow"}
		s.FetcherExt.OnFetchWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed to fetch response"))
		err := getExecutionQueueAttributes(s.Ctx, args, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to fetch response"), err)
		s.FetcherExt.AssertCalled(t, "FetchWorkflowAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, "workflow",
			admin.MatchableResource_EXECUTION_QUEUE)
		s.TearDownAndVerify(t, ``)
	})
}
