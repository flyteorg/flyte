package interfaces

import (
	"context"
	"errors"

	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

var (
	ErrProjectNotFound      = errors.New("project not found")
	ErrProjectAlreadyExists = errors.New("project already exists")
)

type ProjectRepo interface {
	CreateProject(ctx context.Context, project *models.Project) error
	GetProject(ctx context.Context, identifier string) (*models.Project, error)
	UpdateProject(ctx context.Context, project *models.Project) error
	ListProjects(ctx context.Context, input ListResourceInput) ([]*models.Project, error)
}
