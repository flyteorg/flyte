package gormimpl

import (
	"context"

	"google.golang.org/grpc/codes"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flytestdlib/promutils"

	"github.com/jinzhu/gorm"

	"github.com/lyft/flyteadmin/pkg/common"
	flyteAdminErrors "github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
)

type ProjectRepo struct {
	db               *gorm.DB
	errorTransformer errors.ErrorTransformer
	metrics          gormMetrics
}

func (r *ProjectRepo) Create(ctx context.Context, project models.Project) error {
	timer := r.metrics.CreateDuration.Start()
	tx := r.db.Create(&project)
	timer.Stop()
	if tx.Error != nil {
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return nil
}

func (r *ProjectRepo) Get(ctx context.Context, projectID string) (models.Project, error) {
	var project models.Project
	timer := r.metrics.GetDuration.Start()
	tx := r.db.Where(&models.Project{
		Identifier: projectID,
	}).First(&project)
	timer.Stop()
	if tx.Error != nil {
		return models.Project{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	if tx.RecordNotFound() {
		return models.Project{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound, "project [%s] not found", projectID)
	}
	return project, nil
}

func (r *ProjectRepo) ListAll(ctx context.Context, sortParameter common.SortParameter) ([]models.Project, error) {
	var projects []models.Project
	var tx = r.db.Where("state != ?", int32(admin.Project_ARCHIVED))
	if sortParameter != nil {
		tx = tx.Order(sortParameter.GetGormOrderExpr())
	}
	tx = tx.Find(&projects)
	if tx.Error != nil {
		return nil, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return projects, nil
}

func NewProjectRepo(db *gorm.DB, errorTransformer errors.ErrorTransformer,
	scope promutils.Scope) interfaces.ProjectRepoInterface {
	metrics := newMetrics(scope)
	return &ProjectRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}

func (r *ProjectRepo) UpdateProject(ctx context.Context, projectUpdate models.Project) error {
	// Use gorm client to update the two fields that are changed.
	writeTx := r.db.Model(&projectUpdate).Updates(projectUpdate)

	// Return error if applies.
	if writeTx.Error != nil {
		return r.errorTransformer.ToFlyteAdminError(writeTx.Error)
	}

	return nil
}
