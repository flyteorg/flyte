package mocks

import (
	"context"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

type UpdateProjectDomainFunc func(ctx context.Context, request admin.ProjectDomainAttributesUpdateRequest) (
	*admin.ProjectDomainAttributesUpdateResponse, error)
type GetProjectDomainFunc func(ctx context.Context, request admin.ProjectDomainAttributesGetRequest) (
	*admin.ProjectDomainAttributesGetResponse, error)
type DeleteProjectDomainFunc func(ctx context.Context, request admin.ProjectDomainAttributesDeleteRequest) (
	*admin.ProjectDomainAttributesDeleteResponse, error)

type MockProjectDomainAttributesManager struct {
	updateProjectDomainFunc UpdateProjectDomainFunc
	GetFunc                 GetProjectDomainFunc
	DeleteFunc              DeleteProjectDomainFunc
}

func (m *MockProjectDomainAttributesManager) SetUpdateProjectDomainAttributes(updateProjectDomainFunc UpdateProjectDomainFunc) {
	m.updateProjectDomainFunc = updateProjectDomainFunc
}

func (m *MockProjectDomainAttributesManager) UpdateProjectDomainAttributes(
	ctx context.Context, request admin.ProjectDomainAttributesUpdateRequest) (
	*admin.ProjectDomainAttributesUpdateResponse, error) {
	if m.updateProjectDomainFunc != nil {
		return m.updateProjectDomainFunc(ctx, request)
	}
	return nil, nil
}

func (m *MockProjectDomainAttributesManager) GetProjectDomainAttributes(
	ctx context.Context, request admin.ProjectDomainAttributesGetRequest) (
	*admin.ProjectDomainAttributesGetResponse, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, request)
	}
	return nil, nil
}

func (m *MockProjectDomainAttributesManager) DeleteProjectDomainAttributes(
	ctx context.Context, request admin.ProjectDomainAttributesDeleteRequest) (
	*admin.ProjectDomainAttributesDeleteResponse, error) {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, request)
	}
	return nil, nil
}
