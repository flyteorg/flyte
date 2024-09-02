package mocks

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

type CreateProjectFunc func(ctx context.Context, request *admin.ProjectRegisterRequest) (*admin.ProjectRegisterResponse, error)
type ListProjectFunc func(ctx context.Context, request *admin.ProjectListRequest) (*admin.Projects, error)
type UpdateProjectFunc func(ctx context.Context, request *admin.Project) (*admin.ProjectUpdateResponse, error)
type GetProjectFunc func(ctx context.Context, request *admin.ProjectGetRequest) (*admin.Project, error)
type GetDomainsFunc func(ctx context.Context, request *admin.GetDomainRequest) *admin.GetDomainsResponse

type MockProjectManager struct {
	listProjectFunc   ListProjectFunc
	createProjectFunc CreateProjectFunc
	updateProjectFunc UpdateProjectFunc
	getProjectFunc    GetProjectFunc
	getDomainsFunc    GetDomainsFunc
}

func (m *MockProjectManager) SetCreateProject(createProjectFunc CreateProjectFunc) {
	m.createProjectFunc = createProjectFunc
}

func (m *MockProjectManager) CreateProject(ctx context.Context, request *admin.ProjectRegisterRequest) (*admin.ProjectRegisterResponse, error) {
	if m.createProjectFunc != nil {
		return m.createProjectFunc(ctx, request)
	}
	return nil, nil
}

func (m *MockProjectManager) UpdateProject(ctx context.Context, request *admin.Project) (*admin.ProjectUpdateResponse, error) {
	if m.updateProjectFunc != nil {
		return m.updateProjectFunc(ctx, request)
	}
	return nil, nil
}

func (m *MockProjectManager) SetListCallback(listProjectFunc ListProjectFunc) {
	m.listProjectFunc = listProjectFunc
}

func (m *MockProjectManager) ListProjects(
	ctx context.Context, request *admin.ProjectListRequest) (*admin.Projects, error) {
	if m.listProjectFunc != nil {
		return m.listProjectFunc(ctx, request)
	}
	return nil, nil
}

func (m *MockProjectManager) SetGetCallBack(getProjectFunc GetProjectFunc) {
	m.getProjectFunc = getProjectFunc
}

func (m *MockProjectManager) GetProject(ctx context.Context, request *admin.ProjectGetRequest) (*admin.Project, error) {
	if m.getProjectFunc != nil {
		return m.getProjectFunc(ctx, request)
	}
	return nil, nil
}

func (m *MockProjectManager) GetDomains(ctx context.Context, request *admin.GetDomainRequest) *admin.GetDomainsResponse {
	if m.getDomainsFunc != nil {
		return m.getDomainsFunc(ctx, request)
	}
	return nil
}
