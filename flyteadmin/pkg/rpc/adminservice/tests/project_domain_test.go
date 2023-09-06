package tests

import (
	"context"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/manager/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
)

func TestUpdateProjectDomain(t *testing.T) {
	ctx := context.Background()

	mockProjectDomainManager := mocks.MockResourceManager{}
	var updateCalled bool
	mockProjectDomainManager.SetUpdateProjectDomainAttributes(
		func(ctx context.Context,
			request admin.ProjectDomainAttributesUpdateRequest) (*admin.ProjectDomainAttributesUpdateResponse, error) {
			updateCalled = true
			return &admin.ProjectDomainAttributesUpdateResponse{}, nil
		},
	)
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		resourceManager: &mockProjectDomainManager,
	})

	resp, err := mockServer.UpdateProjectDomainAttributes(ctx, &admin.ProjectDomainAttributesUpdateRequest{
		Attributes: &admin.ProjectDomainAttributes{
			Project: "project",
			Domain:  "domain",
		},
	})
	assert.NotNil(t, resp)
	assert.NoError(t, err)
	assert.True(t, updateCalled)
}

func TestUpdateProjectAttr(t *testing.T) {
	ctx := context.Background()

	mockProjectDomainManager := mocks.MockResourceManager{}
	var updateCalled bool
	mockProjectDomainManager.SetUpdateProjectAttributes(
		func(ctx context.Context,
			request admin.ProjectAttributesUpdateRequest) (*admin.ProjectAttributesUpdateResponse, error) {
			updateCalled = true
			return &admin.ProjectAttributesUpdateResponse{}, nil
		},
	)
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		resourceManager: &mockProjectDomainManager,
	})

	resp, err := mockServer.UpdateProjectAttributes(ctx, &admin.ProjectAttributesUpdateRequest{
		Attributes: &admin.ProjectAttributes{
			Project: "project",
		},
	})
	assert.NotNil(t, resp)
	assert.NoError(t, err)
	assert.True(t, updateCalled)
}

func TestDeleteProjectAttr(t *testing.T) {
	ctx := context.Background()

	mockProjectDomainManager := mocks.MockResourceManager{}
	var deleteCalled bool
	mockProjectDomainManager.SetDeleteProjectAttributes(
		func(ctx context.Context,
			request admin.ProjectAttributesDeleteRequest) (*admin.ProjectAttributesDeleteResponse, error) {
			deleteCalled = true
			return &admin.ProjectAttributesDeleteResponse{}, nil
		},
	)
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		resourceManager: &mockProjectDomainManager,
	})

	resp, err := mockServer.DeleteProjectAttributes(ctx, &admin.ProjectAttributesDeleteRequest{
		Project:      "project",
		ResourceType: admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG,
	})
	assert.NotNil(t, resp)
	assert.NoError(t, err)
	assert.True(t, deleteCalled)
}

func TestGetProjectAttr(t *testing.T) {
	ctx := context.Background()

	mockProjectDomainManager := mocks.MockResourceManager{}
	var getCalled bool
	mockProjectDomainManager.SetGetProjectAttributes(
		func(ctx context.Context,
			request admin.ProjectAttributesGetRequest) (*admin.ProjectAttributesGetResponse, error) {
			getCalled = true
			return &admin.ProjectAttributesGetResponse{}, nil
		},
	)
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		resourceManager: &mockProjectDomainManager,
	})

	resp, err := mockServer.GetProjectAttributes(ctx, &admin.ProjectAttributesGetRequest{
		Project:      "project",
		ResourceType: admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG,
	})
	assert.NotNil(t, resp)
	assert.NoError(t, err)
	assert.True(t, getCalled)
}
