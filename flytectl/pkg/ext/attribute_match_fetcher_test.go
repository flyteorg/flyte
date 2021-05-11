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

func getAttributeMatchFetcherSetup() {
	ctx = context.Background()
	adminClient = new(mocks.AdminServiceClient)
	adminFetcherExt = AdminFetcherExtClient{AdminClient: adminClient}
}

func TestFetchWorkflowAttributes(t *testing.T) {
	getAttributeMatchFetcherSetup()
	resp := &admin.WorkflowAttributesGetResponse{}
	adminClient.OnGetWorkflowAttributesMatch(mock.Anything, mock.Anything).Return(resp, nil)
	_, err := adminFetcherExt.FetchWorkflowAttributes(ctx, "dummyProject", "domainValue", "workflowName", admin.MatchableResource_TASK_RESOURCE)
	assert.Nil(t, err)
}

func TestFetchWorkflowAttributesError(t *testing.T) {
	getAttributeMatchFetcherSetup()
	adminClient.OnGetWorkflowAttributesMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
	_, err := adminFetcherExt.FetchWorkflowAttributes(ctx, "dummyProject", "domainValue", "workflowName", admin.MatchableResource_TASK_RESOURCE)
	assert.Equal(t, fmt.Errorf("failed"), err)
}

func TestFetchProjectDomainAttributes(t *testing.T) {
	getAttributeMatchFetcherSetup()
	resp := &admin.ProjectDomainAttributesGetResponse{}
	adminClient.OnGetProjectDomainAttributesMatch(mock.Anything, mock.Anything).Return(resp, nil)
	_, err := adminFetcherExt.FetchProjectDomainAttributes(ctx, "dummyProject", "domainValue", admin.MatchableResource_TASK_RESOURCE)
	assert.Nil(t, err)
}

func TestFetchProjectDomainAttributesError(t *testing.T) {
	getAttributeMatchFetcherSetup()
	adminClient.OnGetProjectDomainAttributesMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
	_, err := adminFetcherExt.FetchProjectDomainAttributes(ctx, "dummyProject", "domainValue", admin.MatchableResource_TASK_RESOURCE)
	assert.Equal(t, fmt.Errorf("failed"), err)
}
