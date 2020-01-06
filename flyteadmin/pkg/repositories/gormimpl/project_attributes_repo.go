package gormimpl

import (
	"context"

	"github.com/jinzhu/gorm"
	"github.com/lyft/flyteadmin/pkg/repositories/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flytestdlib/promutils"
	"google.golang.org/grpc/codes"

	flyteAdminErrors "github.com/lyft/flyteadmin/pkg/errors"
)

type ProjectAttributesRepo struct {
	db               *gorm.DB
	errorTransformer errors.ErrorTransformer
	metrics          gormMetrics
}

func (r *ProjectAttributesRepo) CreateOrUpdate(ctx context.Context, input models.ProjectAttributes) error {
	timer := r.metrics.GetDuration.Start()
	var record models.ProjectAttributes
	tx := r.db.FirstOrCreate(&record, models.ProjectAttributes{
		Project:  input.Project,
		Resource: input.Resource,
	})
	timer.Stop()
	if tx.Error != nil {
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}

	timer = r.metrics.UpdateDuration.Start()
	record.Attributes = input.Attributes
	tx = r.db.Save(&record)
	timer.Stop()
	if tx.Error != nil {
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return nil
}

func (r *ProjectAttributesRepo) Get(ctx context.Context, project, resource string) (models.ProjectAttributes, error) {
	var model models.ProjectAttributes
	timer := r.metrics.GetDuration.Start()
	tx := r.db.Where(&models.ProjectAttributes{
		Project:  project,
		Resource: resource,
	}).First(&model)
	timer.Stop()
	if tx.Error != nil {
		return models.ProjectAttributes{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	if tx.RecordNotFound() {
		return models.ProjectAttributes{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound,
			"project [%s] not found", project)
	}
	return model, nil
}

func (r *ProjectAttributesRepo) Delete(ctx context.Context, project, resource string) error {
	var tx *gorm.DB
	r.metrics.DeleteDuration.Time(func() {
		tx = r.db.Where(&models.ProjectAttributes{
			Project:  project,
			Resource: resource,
		}).Unscoped().Delete(models.ProjectAttributes{})
	})
	if tx.Error != nil {
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	if tx.RecordNotFound() {
		return flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound,
			"project [%s] not found", project)
	}
	return nil
}

func NewProjectAttributesRepo(db *gorm.DB, errorTransformer errors.ErrorTransformer,
	scope promutils.Scope) interfaces.ProjectAttributesRepoInterface {
	metrics := newMetrics(scope)
	return &ProjectAttributesRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}
