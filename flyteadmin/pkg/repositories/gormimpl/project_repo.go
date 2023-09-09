package gormimpl

import (
	"context"
	"errors"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/promutils"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"

	flyteAdminErrors "github.com/flyteorg/flyteadmin/pkg/errors"
	flyteAdminDbErrors "github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

type ProjectRepo struct {
	db               *gorm.DB
	errorTransformer flyteAdminDbErrors.ErrorTransformer
	metrics          gormMetrics
}

func (r *ProjectRepo) Create(ctx context.Context, project models.Project) error {
	timer := r.metrics.CreateDuration.Start()
	tx := r.db.Omit("id").Create(&project)
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
	}).Take(&project)
	timer.Stop()
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return models.Project{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound, "project [%s] not found", projectID)
	}

	if tx.Error != nil {
		return models.Project{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}

	return project, nil
}

func (r *ProjectRepo) List(ctx context.Context, input interfaces.ListResourceInput) ([]models.Project, error) {
	var projects []models.Project

	tx := r.db.Offset(input.Offset)
	if input.Limit != 0 {
		tx = tx.Limit(input.Limit)
	}

	// Apply filters
	// If no filter provided, default to filtering out archived projects
	if len(input.InlineFilters) == 0 && len(input.MapFilters) == 0 {
		tx = tx.Where("state != ?", int32(admin.Project_ARCHIVED))
	} else {
		var err error
		tx, err = applyFilters(tx, input.InlineFilters, input.MapFilters)
		if err != nil {
			return nil, err
		}
	}

	// Apply sort ordering
	if input.SortParameter != nil {
		tx = tx.Order(input.SortParameter.GetGormOrderExpr())
	}

	timer := r.metrics.ListDuration.Start()
	tx.Find(&projects)
	timer.Stop()

	if tx.Error != nil {
		return nil, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return projects, nil
}

func NewProjectRepo(db *gorm.DB, errorTransformer flyteAdminDbErrors.ErrorTransformer,
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
