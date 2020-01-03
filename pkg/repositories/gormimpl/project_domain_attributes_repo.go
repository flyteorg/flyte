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

type ProjectDomainAttributesRepo struct {
	db               *gorm.DB
	errorTransformer errors.ErrorTransformer
	metrics          gormMetrics
}

func (r *ProjectDomainAttributesRepo) CreateOrUpdate(ctx context.Context, input models.ProjectDomainAttributes) error {
	timer := r.metrics.GetDuration.Start()
	var record models.ProjectDomainAttributes
	tx := r.db.FirstOrCreate(&record, models.ProjectDomainAttributes{
		Project:  input.Project,
		Domain:   input.Domain,
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

func (r *ProjectDomainAttributesRepo) Get(ctx context.Context, project, domain, resource string) (
	models.ProjectDomainAttributes, error) {
	var model models.ProjectDomainAttributes
	timer := r.metrics.GetDuration.Start()
	tx := r.db.Where(&models.ProjectDomainAttributes{
		Project:  project,
		Domain:   domain,
		Resource: resource,
	}).First(&model)
	timer.Stop()
	if tx.Error != nil {
		return models.ProjectDomainAttributes{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	if tx.RecordNotFound() {
		return models.ProjectDomainAttributes{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound,
			"project-domain [%s-%s] not found", project, domain)
	}
	return model, nil
}

func NewProjectDomainAttributesRepo(db *gorm.DB, errorTransformer errors.ErrorTransformer,
	scope promutils.Scope) interfaces.ProjectDomainAttributesRepoInterface {
	metrics := newMetrics(scope)
	return &ProjectDomainAttributesRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}
