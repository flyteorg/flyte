package get

import (
	"fmt"
	"os"
	"testing"

	"github.com/flyteorg/flytectl/cmd/testutils"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/executionclusterlabel"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getExecutionClusterLabelSetup() {
	executionclusterlabel.DefaultFetchConfig = &executionclusterlabel.AttrFetchConfig{}
	// Clean up the temp directory.
	_ = os.Remove(testDataTempFile)
}

func TestGetExecutionClusterLabel(t *testing.T) {
	executionClusterLabel := &admin.ExecutionClusterLabel{
		Value: "foo",
	}
	projectDomainResp := &admin.ProjectDomainAttributesGetResponse{
		Attributes: &admin.ProjectDomainAttributes{
			Project: config.GetConfig().Project,
			Domain:  config.GetConfig().Domain,
			MatchingAttributes: &admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_ExecutionClusterLabel{
					ExecutionClusterLabel: executionClusterLabel,
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
				Target: &admin.MatchingAttributes_ExecutionClusterLabel{
					ExecutionClusterLabel: executionClusterLabel,
				},
			},
		},
	}
	t.Run("successful get project domain attribute", func(t *testing.T) {
		s := testutils.SetupWithExt()
		getExecutionClusterLabelSetup()
		// No args implying project domain attribute deletion
		s.FetcherExt.OnFetchProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(projectDomainResp, nil)
		err := getExecutionClusterLabel(s.Ctx, []string{}, s.CmdCtx)
		assert.Nil(t, err)
		s.FetcherExt.AssertCalled(t, "FetchProjectDomainAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_EXECUTION_CLUSTER_LABEL)
		s.TearDownAndVerify(t, `{"project":"dummyProject","domain":"dummyDomain","value":"foo"}`)
	})
	t.Run("successful get project domain attribute and write to file", func(t *testing.T) {
		s := testutils.SetupWithExt()
		getExecutionClusterLabelSetup()
		executionclusterlabel.DefaultFetchConfig.AttrFile = testDataTempFile
		// No args implying project domain attribute deletion
		s.FetcherExt.OnFetchProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(projectDomainResp, nil)
		err := getExecutionClusterLabel(s.Ctx, []string{}, s.CmdCtx)
		assert.Nil(t, err)
		s.FetcherExt.AssertCalled(t, "FetchProjectDomainAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_EXECUTION_CLUSTER_LABEL)
		s.TearDownAndVerify(t, `wrote the config to file temp-output-file`)
	})
	t.Run("successful get project domain attribute and write to file failure", func(t *testing.T) {
		s := testutils.SetupWithExt()
		getExecutionClusterLabelSetup()
		executionclusterlabel.DefaultFetchConfig.AttrFile = testDataNotExistentTempFile
		// No args implying project domain attribute deletion
		s.FetcherExt.OnFetchProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(projectDomainResp, nil)
		err := getExecutionClusterLabel(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("error dumping in file due to open non-existent-dir/temp-output-file: no such file or directory"), err)
		s.FetcherExt.AssertCalled(t, "FetchProjectDomainAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_EXECUTION_CLUSTER_LABEL)
		s.TearDownAndVerify(t, ``)
	})
	t.Run("failed to get project domain attribute", func(t *testing.T) {
		s := testutils.SetupWithExt()
		getExecutionClusterLabelSetup()
		// No args implying project domain attribute deletion
		s.FetcherExt.OnFetchProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil, fmt.Errorf("failed to fetch response"))
		err := getExecutionClusterLabel(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to fetch response"), err)
		s.FetcherExt.AssertCalled(t, "FetchProjectDomainAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_EXECUTION_CLUSTER_LABEL)
		s.TearDownAndVerify(t, ``)
	})
	t.Run("successful get workflow attribute", func(t *testing.T) {
		s := testutils.SetupWithExt()
		getExecutionClusterLabelSetup()
		args := []string{"workflow"}
		s.FetcherExt.OnFetchWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(workflowResp, nil)
		err := getExecutionClusterLabel(s.Ctx, args, s.CmdCtx)
		assert.Nil(t, err)
		s.FetcherExt.AssertCalled(t, "FetchWorkflowAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, "workflow",
			admin.MatchableResource_EXECUTION_CLUSTER_LABEL)
		s.TearDownAndVerify(t, `{"project":"dummyProject","domain":"dummyDomain","workflow":"workflow","value":"foo"}`)
	})
	t.Run("failed to get workflow attribute", func(t *testing.T) {
		s := testutils.SetupWithExt()
		getExecutionClusterLabelSetup()
		args := []string{"workflow"}
		s.FetcherExt.OnFetchWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed to fetch response"))
		err := getExecutionClusterLabel(s.Ctx, args, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to fetch response"), err)
		s.FetcherExt.AssertCalled(t, "FetchWorkflowAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, "workflow",
			admin.MatchableResource_EXECUTION_CLUSTER_LABEL)
		s.TearDownAndVerify(t, ``)
	})
}
