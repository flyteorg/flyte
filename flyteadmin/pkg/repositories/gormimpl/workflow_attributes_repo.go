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

type WorkflowAttributesRepo struct {
	db               *gorm.DB
	errorTransformer errors.ErrorTransformer
	metrics          gormMetrics
}

func (r *WorkflowAttributesRepo) CreateOrUpdate(ctx context.Context, input models.WorkflowAttributes) error {
	timer := r.metrics.GetDuration.Start()
	var record models.WorkflowAttributes
	tx := r.db.FirstOrCreate(&record, models.WorkflowAttributes{
		Project:  input.Project,
		Domain:   input.Domain,
		Workflow: input.Workflow,
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

func (r *WorkflowAttributesRepo) Get(ctx context.Context, project, domain, workflow, resource string) (
	models.WorkflowAttributes, error) {
	var model models.WorkflowAttributes
	timer := r.metrics.GetDuration.Start()
	tx := r.db.Where(&models.WorkflowAttributes{
		Project:  project,
		Domain:   domain,
		Workflow: workflow,
		Resource: resource,
	}).First(&model)
	timer.Stop()
	if tx.Error != nil {
		return models.WorkflowAttributes{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	if tx.RecordNotFound() {
		return models.WorkflowAttributes{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound,
			"project-domain-workflow [%s-%s-%s] not found", project, domain, workflow)
	}
	return model, nil
}

func (r *WorkflowAttributesRepo) Delete(ctx context.Context, project, domain, workflow, resource string) error {
	var tx *gorm.DB
	r.metrics.DeleteDuration.Time(func() {
		tx = r.db.Where(&models.WorkflowAttributes{
			Project:  project,
			Domain:   domain,
			Workflow: workflow,
			Resource: resource,
		}).Unscoped().Delete(models.WorkflowAttributes{})
	})
	if tx.Error != nil {
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	if tx.RecordNotFound() {
		return flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound,
			"project-domain-workflow [%s-%s-%s] not found", project, domain, workflow)
	}
	return nil
}

func NewWorkflowAttributesRepo(db *gorm.DB, errorTransformer errors.ErrorTransformer,
	scope promutils.Scope) interfaces.WorkflowAttributesRepoInterface {
	metrics := newMetrics(scope)
	return &WorkflowAttributesRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}
