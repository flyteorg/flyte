package tests

import (
	"context"
	"testing"

	"github.com/lyft/flyteadmin/pkg/manager/mocks"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
)

func TestUpdateProjectDomain(t *testing.T) {
	ctx := context.Background()

	mockProjectDomainManager := mocks.MockProjectDomainAttributesManager{}
	var updateCalled bool
	mockProjectDomainManager.SetUpdateProjectDomainAttributes(
		func(ctx context.Context,
			request admin.ProjectDomainAttributesUpdateRequest) (*admin.ProjectDomainAttributesUpdateResponse, error) {
			updateCalled = true
			return &admin.ProjectDomainAttributesUpdateResponse{}, nil
		},
	)
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		projectDomainAttributesManager: &mockProjectDomainManager,
	})

	resp, err := mockServer.UpdateProjectDomainAttributes(ctx, &admin.ProjectDomainAttributesUpdateRequest{})
	assert.NotNil(t, resp)
	assert.NoError(t, err)
	assert.True(t, updateCalled)
}
