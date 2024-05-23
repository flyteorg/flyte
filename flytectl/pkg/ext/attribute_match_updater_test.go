package ext

import (
	"context"
	"fmt"
	"testing"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var adminUpdaterExt AdminUpdaterExtClient

func updateAttributeMatchFetcherSetup() {
	ctx = context.Background()
	adminClient = new(mocks.AdminServiceClient)
	adminUpdaterExt = AdminUpdaterExtClient{AdminClient: adminClient}
}

func TestUpdateWorkflowAttributes(t *testing.T) {
	updateAttributeMatchFetcherSetup()
	matchingAttr := &admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_TaskResourceAttributes{},
	}
	adminClient.OnUpdateWorkflowAttributesMatch(mock.Anything, mock.Anything).Return(nil, nil)
	err := adminUpdaterExt.UpdateWorkflowAttributes(ctx, "dummyProject", "domainValue", "workflowName", matchingAttr)
	assert.Nil(t, err)
}

func TestUpdateWorkflowAttributesError(t *testing.T) {
	updateAttributeMatchFetcherSetup()
	adminClient.OnUpdateWorkflowAttributesMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
	err := adminUpdaterExt.UpdateWorkflowAttributes(ctx, "dummyProject", "domainValue", "workflowName", nil)
	assert.Equal(t, fmt.Errorf("failed"), err)
}

func TestUpdateProjectDomainAttributes(t *testing.T) {
	updateAttributeMatchFetcherSetup()
	matchingAttr := &admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_TaskResourceAttributes{},
	}
	adminClient.OnUpdateProjectDomainAttributesMatch(mock.Anything, mock.Anything).Return(nil, nil)
	err := adminUpdaterExt.UpdateProjectDomainAttributes(ctx, "dummyProject", "domainValue", matchingAttr)
	assert.Nil(t, err)
}

func TestUpdateProjectDomainAttributesError(t *testing.T) {
	updateAttributeMatchFetcherSetup()
	adminClient.OnUpdateProjectDomainAttributesMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
	err := adminUpdaterExt.UpdateProjectDomainAttributes(ctx, "dummyProject", "domainValue", nil)
	assert.Equal(t, fmt.Errorf("failed"), err)
}

func TestUpdateProjectAttributes(t *testing.T) {
	updateAttributeMatchFetcherSetup()
	matchingAttr := &admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_TaskResourceAttributes{},
	}
	adminClient.OnUpdateProjectAttributesMatch(mock.Anything, mock.Anything).Return(nil, nil)
	err := adminUpdaterExt.UpdateProjectAttributes(ctx, "dummyProject", matchingAttr)
	assert.Nil(t, err)
}

func TestUpdateProjectAttributesError(t *testing.T) {
	updateAttributeMatchFetcherSetup()
	adminClient.OnUpdateProjectAttributesMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
	err := adminUpdaterExt.UpdateProjectAttributes(ctx, "dummyProject", nil)
	assert.Equal(t, fmt.Errorf("failed"), err)
}
