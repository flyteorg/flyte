package ext

import (
	"context"
	"fmt"
	"testing"

	"github.com/flyteorg/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var wResp *admin.WorkflowAttributesGetResponse

var pResp *admin.ProjectDomainAttributesGetResponse

func getAttributeMatchFetcherSetup() {
	ctx = context.Background()
	adminClient = new(mocks.AdminServiceClient)
	adminFetcherExt = AdminFetcherExtClient{AdminClient: adminClient}
	wResp = &admin.WorkflowAttributesGetResponse{Attributes: &admin.WorkflowAttributes{
		MatchingAttributes: &admin.MatchingAttributes{
			Target: nil,
		}}}
	pResp = &admin.ProjectDomainAttributesGetResponse{Attributes: &admin.ProjectDomainAttributes{
		Project: "dummyProject",
		Domain:  "dummyDomain",
		MatchingAttributes: &admin.MatchingAttributes{
			Target: nil,
		}}}
}

func TestFetchWorkflowAttributes(t *testing.T) {
	getAttributeMatchFetcherSetup()
	adminClient.OnGetWorkflowAttributesMatch(mock.Anything, mock.Anything).Return(wResp, nil)
	_, err := adminFetcherExt.FetchWorkflowAttributes(ctx, "dummyProject", "domainValue", "workflowName", admin.MatchableResource_TASK_RESOURCE)
	assert.Nil(t, err)
}

func TestFetchWorkflowAttributesError(t *testing.T) {
	t.Run("failed api", func(t *testing.T) {
		getAttributeMatchFetcherSetup()
		adminClient.OnGetWorkflowAttributesMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
		_, err := adminFetcherExt.FetchWorkflowAttributes(ctx, "dummyProject", "domainValue", "workflowName", admin.MatchableResource_TASK_RESOURCE)
		assert.Equal(t, fmt.Errorf("failed"), err)
	})
	t.Run("empty data from api", func(t *testing.T) {
		getAttributeMatchFetcherSetup()
		wResp := &admin.WorkflowAttributesGetResponse{}
		adminClient.OnGetWorkflowAttributesMatch(mock.Anything, mock.Anything).Return(wResp, nil)
		_, err := adminFetcherExt.FetchWorkflowAttributes(ctx, "dummyProject", "domainValue", "workflowName", admin.MatchableResource_TASK_RESOURCE)
		assert.NotNil(t, err)
		assert.True(t, IsNotFoundError(err))
		assert.EqualError(t, err, "attribute not found")
	})
}

func TestFetchProjectDomainAttributes(t *testing.T) {
	getAttributeMatchFetcherSetup()
	adminClient.OnGetProjectDomainAttributesMatch(mock.Anything, mock.Anything).Return(pResp, nil)
	_, err := adminFetcherExt.FetchProjectDomainAttributes(ctx, "dummyProject", "domainValue", admin.MatchableResource_TASK_RESOURCE)
	assert.Nil(t, err)
}

func TestFetchProjectDomainAttributesError(t *testing.T) {
	t.Run("failed api", func(t *testing.T) {
		getAttributeMatchFetcherSetup()
		adminClient.OnGetProjectDomainAttributesMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
		_, err := adminFetcherExt.FetchProjectDomainAttributes(ctx, "dummyProject", "domainValue", admin.MatchableResource_TASK_RESOURCE)
		assert.Equal(t, fmt.Errorf("failed"), err)
	})
	t.Run("empty data from api", func(t *testing.T) {
		getAttributeMatchFetcherSetup()
		pResp := &admin.ProjectDomainAttributesGetResponse{}
		adminClient.OnGetProjectDomainAttributesMatch(mock.Anything, mock.Anything).Return(pResp, nil)
		_, err := adminFetcherExt.FetchProjectDomainAttributes(ctx, "dummyProject", "domainValue", admin.MatchableResource_TASK_RESOURCE)
		assert.NotNil(t, err)
		assert.True(t, IsNotFoundError(err))
		assert.EqualError(t, err, "attribute not found")
	})
}

func TestFetchProjectAttributesError(t *testing.T) {
	t.Run("failed api", func(t *testing.T) {
		getAttributeMatchFetcherSetup()
		adminClient.OnGetProjectAttributesMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
		_, err := adminFetcherExt.FetchProjectAttributes(ctx, "dummyProject", admin.MatchableResource_TASK_RESOURCE)
		assert.Equal(t, fmt.Errorf("failed"), err)
	})
	t.Run("empty data from api", func(t *testing.T) {
		getAttributeMatchFetcherSetup()
		pResp := &admin.ProjectAttributesGetResponse{}
		adminClient.OnGetProjectAttributesMatch(mock.Anything, mock.Anything).Return(pResp, nil)
		_, err := adminFetcherExt.FetchProjectAttributes(ctx, "dummyProject", admin.MatchableResource_TASK_RESOURCE)
		assert.NotNil(t, err)
		assert.True(t, IsNotFoundError(err))
		assert.EqualError(t, err, "attribute not found")
	})
}
