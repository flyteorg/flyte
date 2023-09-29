package mocks

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

type CreateProjectFunc func(ctx context.Context, request admin.ProjectRegisterRequest) (*admin.ProjectRegisterResponse, error)
type ListProjectFunc func(ctx context.Context, request admin.ProjectListRequest) (*admin.Projects, error)
type UpdateProjectFunc func(ctx context.Context, request admin.Project) (*admin.ProjectUpdateResponse, error)

type MockProjectManager struct {
	listProjectFunc   ListProjectFunc
	createProjectFunc CreateProjectFunc
	updateProjectFunc UpdateProjectFunc
}

func (m *MockProjectManager) SetCreateProject(createProjectFunc CreateProjectFunc) {
	m.createProjectFunc = createProjectFunc
}

func (m *MockProjectManager) CreateProject(ctx context.Context, request admin.ProjectRegisterRequest) (*admin.ProjectRegisterResponse, error) {
	if m.createProjectFunc != nil {
		return m.createProjectFunc(ctx, request)
	}
	return nil, nil
}

func (m *MockProjectManager) UpdateProject(ctx context.Context, request admin.Project) (*admin.ProjectUpdateResponse, error) {
	if m.updateProjectFunc != nil {
		return m.updateProjectFunc(ctx, request)
	}
	return nil, nil
}

func (m *MockProjectManager) SetListCallback(listProjectFunc ListProjectFunc) {
	m.listProjectFunc = listProjectFunc
}

func (m *MockProjectManager) ListProjects(
	ctx context.Context, request admin.ProjectListRequest) (*admin.Projects, error) {
	if m.listProjectFunc != nil {
		return m.listProjectFunc(ctx, request)
	}
	return nil, nil
}
