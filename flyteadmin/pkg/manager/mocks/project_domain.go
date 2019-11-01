package mocks

import (
	"context"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

type UpdateProjectDomainFunc func(ctx context.Context, request admin.ProjectDomainAttributesUpdateRequest) (
	*admin.ProjectDomainAttributesUpdateResponse, error)

type MockProjectDomainManager struct {
	updateProjectDomainFunc UpdateProjectDomainFunc
}

func (m *MockProjectDomainManager) SetUpdateProjectDomainAttributes(updateProjectDomainFunc UpdateProjectDomainFunc) {
	m.updateProjectDomainFunc = updateProjectDomainFunc
}

func (m *MockProjectDomainManager) UpdateProjectDomain(
	ctx context.Context, request admin.ProjectDomainAttributesUpdateRequest) (
	*admin.ProjectDomainAttributesUpdateResponse, error) {
	if m.updateProjectDomainFunc != nil {
		return m.updateProjectDomainFunc(ctx, request)
	}
	return nil, nil
}
