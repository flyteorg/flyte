package interfaces

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

// Interface for managing projects (and domains).
type ProjectInterface interface {
	CreateProject(ctx context.Context, request admin.ProjectRegisterRequest) (*admin.ProjectRegisterResponse, error)
	ListProjects(ctx context.Context, request admin.ProjectListRequest) (*admin.Projects, error)
	UpdateProject(ctx context.Context, request admin.Project) (*admin.ProjectUpdateResponse, error)
}
