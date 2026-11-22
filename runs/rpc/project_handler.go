package rpc

import (
	"context"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/project"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/project/projectconnect"
	"github.com/flyteorg/flyte/v2/runs/service"
)

type projectHandler struct{ projects service.Project }

// NewProjectHandler exposes the project service through the Connect API.
func NewProjectHandler(projects service.Project) projectconnect.ProjectServiceHandler {
	return &projectHandler{projects: projects}
}

func (h *projectHandler) CreateProject(ctx context.Context, req *connect.Request[project.CreateProjectRequest]) (*connect.Response[project.CreateProjectResponse], error) {
	return rpcResponse(h.projects.CreateProject(ctx, req.Msg))
}

func (h *projectHandler) UpdateProject(ctx context.Context, req *connect.Request[project.UpdateProjectRequest]) (*connect.Response[project.UpdateProjectResponse], error) {
	return rpcResponse(h.projects.UpdateProject(ctx, req.Msg))
}

func (h *projectHandler) GetProject(ctx context.Context, req *connect.Request[project.GetProjectRequest]) (*connect.Response[project.GetProjectResponse], error) {
	return rpcResponse(h.projects.GetProject(ctx, req.Msg))
}

func (h *projectHandler) ListProjects(ctx context.Context, req *connect.Request[project.ListProjectsRequest]) (*connect.Response[project.ListProjectsResponse], error) {
	return rpcResponse(h.projects.ListProjects(ctx, req.Msg))
}
