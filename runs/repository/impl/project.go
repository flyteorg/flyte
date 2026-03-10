package impl

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

type projectRepo struct {
	db *gorm.DB
}

func NewProjectRepo(db *gorm.DB) interfaces.ProjectRepo {
	return &projectRepo{
		db: db,
	}
}

func (r *projectRepo) CreateProject(ctx context.Context, project *models.Project) error {
	now := time.Now().UTC()
	project.CreatedAt = now
	project.UpdatedAt = now

	result := r.db.WithContext(ctx).Create(project)
	if result.Error != nil {
		logger.Errorf(ctx, "failed to create project %s/%s: %v", project.Org, project.ID, result.Error)
		return fmt.Errorf("%w: %v", interfaces.ErrProjectAlreadyExists, result.Error)
	}

	return nil
}

func (r *projectRepo) GetProject(ctx context.Context, key models.ProjectKey) (*models.Project, error) {
	var project models.Project
	result := r.db.WithContext(ctx).
		Where("org = ? AND id = ?", key.Org, key.ID).
		First(&project)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w: %s/%s", interfaces.ErrProjectNotFound, key.Org, key.ID)
		}
		logger.Errorf(ctx, "failed to get project %s/%s: %v", key.Org, key.ID, result.Error)
		return nil, fmt.Errorf("failed to get project %s/%s: %w", key.Org, key.ID, result.Error)
	}

	return &project, nil
}

func (r *projectRepo) UpdateProject(ctx context.Context, project *models.Project) error {
	updates := map[string]interface{}{
		"name":        project.Name,
		"description": project.Description,
		"labels":      project.Labels,
		"state":       project.State,
		"updated_at":  time.Now().UTC(),
	}

	result := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("org = ? AND id = ?", project.Org, project.ID).
		Updates(updates)
	if result.Error != nil {
		logger.Errorf(ctx, "failed to update project %s/%s: %v", project.Org, project.ID, result.Error)
		return fmt.Errorf("failed to update project %s/%s: %w", project.Org, project.ID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: %s/%s", interfaces.ErrProjectNotFound, project.Org, project.ID)
	}

	return nil
}

func (r *projectRepo) ListProjects(ctx context.Context, input interfaces.ListResourceInput) ([]*models.Project, error) {
	var projects []*models.Project
	query := r.db.WithContext(ctx).Model(&models.Project{})

	if input.Filter != nil {
		expr, err := input.Filter.GormQueryExpression("")
		if err != nil {
			return nil, fmt.Errorf("failed to build filter: %w", err)
		}
		query = query.Where(expr.Query, expr.Args...)
	}

	if input.ScopeByFilter != nil {
		expr, err := input.ScopeByFilter.GormQueryExpression("")
		if err != nil {
			return nil, fmt.Errorf("failed to build scope filter: %w", err)
		}
		query = query.Where(expr.Query, expr.Args...)
	}

	if len(input.SortParameters) > 0 {
		for _, sp := range input.SortParameters {
			query = query.Order(sp.GetGormOrderExpr())
		}
	} else {
		query = query.Order("id DESC")
	}

	query = query.Offset(input.Offset)
	if input.Limit > 0 {
		query = query.Limit(input.Limit)
	}
	if err := query.Find(&projects).Error; err != nil {
		logger.Errorf(ctx, "failed to list projects: %v", err)
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	return projects, nil
}
