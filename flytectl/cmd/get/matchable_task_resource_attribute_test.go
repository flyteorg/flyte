package get

import (
	"fmt"
	"os"
	"testing"

	"github.com/flyteorg/flytectl/cmd/testutils"

	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/taskresourceattribute"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getTaskResourceAttributeSetup() {
	taskresourceattribute.DefaultFetchConfig = &taskresourceattribute.AttrFetchConfig{}
	// Clean up the temp directory.
	_ = os.Remove(testDataTempFile)
}

func TestGetTaskResourceAttributes(t *testing.T) {
	taskResourceAttr := &admin.TaskResourceAttributes{
		Defaults: &admin.TaskResourceSpec{
			Cpu:    "1",
			Memory: "150Mi",
		},
		Limits: &admin.TaskResourceSpec{
			Cpu:    "2",
			Memory: "350Mi",
		},
	}
	projectDomainResp := &admin.ProjectDomainAttributesGetResponse{
		Attributes: &admin.ProjectDomainAttributes{
			Project: config.GetConfig().Project,
			Domain:  config.GetConfig().Domain,
			MatchingAttributes: &admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_TaskResourceAttributes{
					TaskResourceAttributes: taskResourceAttr,
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
				Target: &admin.MatchingAttributes_TaskResourceAttributes{
					TaskResourceAttributes: taskResourceAttr,
				},
			},
		},
	}
	t.Run("successful get project domain attribute", func(t *testing.T) {
		s := testutils.SetupWithExt()
		getTaskResourceAttributeSetup()
		// No args implying project domain attribute deletion
		s.FetcherExt.OnFetchProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(projectDomainResp, nil)
		err := getTaskResourceAttributes(s.Ctx, []string{}, s.CmdCtx)
		assert.Nil(t, err)
		s.FetcherExt.AssertCalled(t, "FetchProjectDomainAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_TASK_RESOURCE)
		s.TearDownAndVerify(t, `{"project":"dummyProject","domain":"dummyDomain","defaults":{"cpu":"1","memory":"150Mi"},"limits":{"cpu":"2","memory":"350Mi"}}`)
	})
	t.Run("successful get project domain attribute and write to file", func(t *testing.T) {
		s := testutils.SetupWithExt()
		getTaskResourceAttributeSetup()
		taskresourceattribute.DefaultFetchConfig.AttrFile = testDataTempFile
		// No args implying project domain attribute deletion
		s.FetcherExt.OnFetchProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(projectDomainResp, nil)
		err := getTaskResourceAttributes(s.Ctx, []string{}, s.CmdCtx)
		assert.Nil(t, err)
		s.FetcherExt.AssertCalled(t, "FetchProjectDomainAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_TASK_RESOURCE)
		s.TearDownAndVerify(t, `wrote the config to file temp-output-file`)
	})
	t.Run("successful get project domain attribute and write to file failure", func(t *testing.T) {
		s := testutils.SetupWithExt()
		getTaskResourceAttributeSetup()
		taskresourceattribute.DefaultFetchConfig.AttrFile = testDataNotExistentTempFile
		// No args implying project domain attribute deletion
		s.FetcherExt.OnFetchProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(projectDomainResp, nil)
		err := getTaskResourceAttributes(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("error dumping in file due to open non-existent-dir/temp-output-file: no such file or directory"), err)
		s.FetcherExt.AssertCalled(t, "FetchProjectDomainAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_TASK_RESOURCE)
		s.TearDownAndVerify(t, ``)
	})
	t.Run("failed get project domain attribute", func(t *testing.T) {
		s := testutils.SetupWithExt()
		getTaskResourceAttributeSetup()
		// No args implying project domain attribute deletion
		s.FetcherExt.OnFetchProjectDomainAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil, fmt.Errorf("failed to fetch response"))
		err := getTaskResourceAttributes(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to fetch response"), err)
		s.FetcherExt.AssertCalled(t, "FetchProjectDomainAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_TASK_RESOURCE)
		s.TearDownAndVerify(t, ``)
	})
	t.Run("successful get workflow attribute", func(t *testing.T) {
		s := testutils.SetupWithExt()
		getTaskResourceAttributeSetup()
		args := []string{"workflow"}
		s.FetcherExt.OnFetchWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(workflowResp, nil)
		err := getTaskResourceAttributes(s.Ctx, args, s.CmdCtx)
		assert.Nil(t, err)
		s.FetcherExt.AssertCalled(t, "FetchWorkflowAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, "workflow",
			admin.MatchableResource_TASK_RESOURCE)
		s.TearDownAndVerify(t, `{"project":"dummyProject","domain":"dummyDomain","workflow":"workflow","defaults":{"cpu":"1","memory":"150Mi"},"limits":{"cpu":"2","memory":"350Mi"}}`)
	})
	t.Run("failed get workflow attribute", func(t *testing.T) {
		s := testutils.SetupWithExt()
		getTaskResourceAttributeSetup()
		args := []string{"workflow"}
		s.FetcherExt.OnFetchWorkflowAttributesMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed to fetch response"))
		err := getTaskResourceAttributes(s.Ctx, args, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to fetch response"), err)
		s.FetcherExt.AssertCalled(t, "FetchWorkflowAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, "workflow",
			admin.MatchableResource_TASK_RESOURCE)
		s.TearDownAndVerify(t, ``)
	})
}
