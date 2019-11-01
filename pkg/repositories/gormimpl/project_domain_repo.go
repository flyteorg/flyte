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

type ProjectDomainRepo struct {
	db               *gorm.DB
	errorTransformer errors.ErrorTransformer
	metrics          gormMetrics
}

func (r *ProjectDomainRepo) CreateOrUpdate(ctx context.Context, input models.ProjectDomain) error {
	timer := r.metrics.GetDuration.Start()
	var record models.ProjectDomain
	tx := r.db.FirstOrCreate(&record, models.ProjectDomain{
		Project: input.Project,
		Domain:  input.Domain,
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

func (r *ProjectDomainRepo) Get(ctx context.Context, project, domain string) (models.ProjectDomain, error) {
	var model models.ProjectDomain
	timer := r.metrics.GetDuration.Start()
	tx := r.db.Where(&models.ProjectDomain{
		Project: project,
		Domain:  domain,
	}).First(&model)
	timer.Stop()
	if tx.Error != nil {
		return models.ProjectDomain{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	if tx.RecordNotFound() {
		return models.ProjectDomain{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound,
			"project-domain [%s-%s] not found", project, domain)
	}
	return model, nil
}

func NewProjectDomainRepo(db *gorm.DB, errorTransformer errors.ErrorTransformer,
	scope promutils.Scope) interfaces.ProjectDomainRepoInterface {
	metrics := newMetrics(scope)
	return &ProjectDomainRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}
