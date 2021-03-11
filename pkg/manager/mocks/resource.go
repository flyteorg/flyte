package mocks

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

type UpdateProjectDomainFunc func(ctx context.Context, request admin.ProjectDomainAttributesUpdateRequest) (
	*admin.ProjectDomainAttributesUpdateResponse, error)
type GetProjectDomainFunc func(ctx context.Context, request admin.ProjectDomainAttributesGetRequest) (
	*admin.ProjectDomainAttributesGetResponse, error)
type DeleteProjectDomainFunc func(ctx context.Context, request admin.ProjectDomainAttributesDeleteRequest) (
	*admin.ProjectDomainAttributesDeleteResponse, error)
type ListResourceFunc func(ctx context.Context, request admin.ListMatchableAttributesRequest) (
	*admin.ListMatchableAttributesResponse, error)
type GetResourceFunc func(ctx context.Context, request interfaces.ResourceRequest) (*interfaces.ResourceResponse, error)

type MockResourceManager struct {
	updateProjectDomainFunc UpdateProjectDomainFunc
	GetFunc                 GetProjectDomainFunc
	DeleteFunc              DeleteProjectDomainFunc
	ListFunc                ListResourceFunc
	GetResourceFunc         GetResourceFunc
}

func (m *MockResourceManager) GetResource(ctx context.Context, request interfaces.ResourceRequest) (*interfaces.ResourceResponse, error) {
	if m.GetResourceFunc != nil {
		return m.GetResourceFunc(ctx, request)
	}
	return nil, nil
}

func (m *MockResourceManager) UpdateWorkflowAttributes(ctx context.Context, request admin.WorkflowAttributesUpdateRequest) (
	*admin.WorkflowAttributesUpdateResponse, error) {
	panic("implement me")
}

func (m *MockResourceManager) GetWorkflowAttributes(ctx context.Context, request admin.WorkflowAttributesGetRequest) (
	*admin.WorkflowAttributesGetResponse, error) {
	panic("implement me")
}

func (m *MockResourceManager) DeleteWorkflowAttributes(ctx context.Context, request admin.WorkflowAttributesDeleteRequest) (
	*admin.WorkflowAttributesDeleteResponse, error) {
	panic("implement me")
}

func (m *MockResourceManager) SetUpdateProjectDomainAttributes(updateProjectDomainFunc UpdateProjectDomainFunc) {
	m.updateProjectDomainFunc = updateProjectDomainFunc
}

func (m *MockResourceManager) UpdateProjectDomainAttributes(
	ctx context.Context, request admin.ProjectDomainAttributesUpdateRequest) (
	*admin.ProjectDomainAttributesUpdateResponse, error) {
	if m.updateProjectDomainFunc != nil {
		return m.updateProjectDomainFunc(ctx, request)
	}
	return nil, nil
}

func (m *MockResourceManager) GetProjectDomainAttributes(
	ctx context.Context, request admin.ProjectDomainAttributesGetRequest) (
	*admin.ProjectDomainAttributesGetResponse, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, request)
	}
	return nil, nil
}

func (m *MockResourceManager) DeleteProjectDomainAttributes(
	ctx context.Context, request admin.ProjectDomainAttributesDeleteRequest) (
	*admin.ProjectDomainAttributesDeleteResponse, error) {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, request)
	}
	return nil, nil
}

func (m *MockResourceManager) ListAll(ctx context.Context, request admin.ListMatchableAttributesRequest) (
	*admin.ListMatchableAttributesResponse, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, request)
	}
	return nil, nil
}
