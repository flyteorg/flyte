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

var adminDeleterExt AdminDeleterExtClient

func deleteAttributeMatchFetcherSetup() {
	ctx = context.Background()
	adminClient = new(mocks.AdminServiceClient)
	adminDeleterExt = AdminDeleterExtClient{AdminClient: adminClient}
}

func TestDeleteWorkflowAttributes(t *testing.T) {
	deleteAttributeMatchFetcherSetup()
	adminClient.OnDeleteWorkflowAttributesMatch(mock.Anything, mock.Anything).Return(nil, nil)
	err := adminDeleterExt.DeleteWorkflowAttributes(ctx, "dummyProject", "domainValue", "workflowName", admin.MatchableResource_TASK_RESOURCE)
	assert.Nil(t, err)
}

func TestDeleteWorkflowAttributesError(t *testing.T) {
	deleteAttributeMatchFetcherSetup()
	adminClient.OnDeleteWorkflowAttributesMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
	err := adminDeleterExt.DeleteWorkflowAttributes(ctx, "dummyProject", "domainValue", "workflowName", admin.MatchableResource_TASK_RESOURCE)
	assert.Equal(t, fmt.Errorf("failed"), err)
}

func TestDeleteProjectDomainAttributes(t *testing.T) {
	deleteAttributeMatchFetcherSetup()
	adminClient.OnDeleteProjectDomainAttributesMatch(mock.Anything, mock.Anything).Return(nil, nil)
	err := adminDeleterExt.DeleteProjectDomainAttributes(ctx, "dummyProject", "domainValue", admin.MatchableResource_TASK_RESOURCE)
	assert.Nil(t, err)
}

func TestDeleteProjectDomainAttributesError(t *testing.T) {
	deleteAttributeMatchFetcherSetup()
	adminClient.OnDeleteProjectDomainAttributesMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
	err := adminDeleterExt.DeleteProjectDomainAttributes(ctx, "dummyProject", "domainValue", admin.MatchableResource_TASK_RESOURCE)
	assert.Equal(t, fmt.Errorf("failed"), err)
}

func TestDeleteProjectAttributes(t *testing.T) {
	deleteAttributeMatchFetcherSetup()
	adminClient.OnDeleteProjectAttributesMatch(mock.Anything, mock.Anything).Return(nil, nil)
	err := adminDeleterExt.DeleteProjectAttributes(ctx, "dummyProject", admin.MatchableResource_TASK_RESOURCE)
	assert.Nil(t, err)
}

func TestDeleteProjectAttributesError(t *testing.T) {
	deleteAttributeMatchFetcherSetup()
	adminClient.OnDeleteProjectAttributesMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
	err := adminDeleterExt.DeleteProjectAttributes(ctx, "dummyProject", admin.MatchableResource_TASK_RESOURCE)
	assert.Equal(t, fmt.Errorf("failed"), err)
}
