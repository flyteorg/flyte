package gormimpl

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/jinzhu/gorm"
	"google.golang.org/grpc/codes"

	flyteAdminErrors "github.com/flyteorg/flyteadmin/pkg/errors"
)

type ResourceRepo struct {
	db               *gorm.DB
	errorTransformer errors.ErrorTransformer
	metrics          gormMetrics
}

const priorityDescending = "priority desc"

/*
	The data in the Resource repo maps to the following rules:
	* Domain and ResourceType can never be empty.
	* Empty string can be interpreted as all. Example: "" for Project field can be interpreted as all Projects for a domain.
	* One cannot provide specific value for Project, unless a specific value for Domain is provided.
	** Project is always scoped within a domain.
	**	Example: Domain="" Project="Lyft" is invalid.
	* One cannot provide specific value for Workflow, unless a specific value for Domain and Project is provided.
	** Workflow is always scoped within a domain and project.
	**	Example: Domain="staging" Project="" Workflow="W1" is invalid.
	* One cannot provide specific value for Launch plan, unless a specific value for Domain, Project and Workflow is provided.
	** Launch plan is always scoped within a domain, project and workflow.
	**	Example: Domain="staging" Project="Lyft" Workflow="" LaunchPlan= "l1" is invalid.
*/
func validateCreateOrUpdateResourceInput(project, domain, workflow, launchPlan, resourceType string) bool {
	if domain == "" || resourceType == "" {
		return false
	}
	if project == "" && (workflow != "" || launchPlan != "") {
		return false
	}
	if workflow == "" && launchPlan != "" {
		return false
	}
	return true
}

func (r *ResourceRepo) CreateOrUpdate(ctx context.Context, input models.Resource) error {
	if !validateCreateOrUpdateResourceInput(input.Project, input.Domain, input.Workflow, input.LaunchPlan, input.ResourceType) {
		return errors.GetInvalidInputError(fmt.Sprintf("%v", input))
	}
	if input.Priority == 0 {
		return errors.GetInvalidInputError(fmt.Sprintf("invalid priority %v", input))
	}
	timer := r.metrics.GetDuration.Start()
	var record models.Resource
	tx := r.db.FirstOrCreate(&record, models.Resource{
		Project:      input.Project,
		Domain:       input.Domain,
		Workflow:     input.Workflow,
		LaunchPlan:   input.LaunchPlan,
		ResourceType: input.ResourceType,
		Priority:     input.Priority,
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

func (r *ResourceRepo) Get(ctx context.Context, ID interfaces.ResourceID) (models.Resource, error) {
	if !validateCreateOrUpdateResourceInput(ID.Project, ID.Domain, ID.Workflow, ID.LaunchPlan, ID.ResourceType) {
		return models.Resource{}, r.errorTransformer.ToFlyteAdminError(errors.GetInvalidInputError(fmt.Sprintf("%v", ID)))
	}
	var resources []models.Resource
	timer := r.metrics.GetDuration.Start()

	txWhereClause := "resource_type = ? AND domain = ? AND project IN (?) AND workflow IN (?) AND launch_plan IN (?)"
	project := []string{""}
	if ID.Project != "" {
		project = append(project, ID.Project)
	}

	workflow := []string{""}
	if ID.Workflow != "" {
		workflow = append(workflow, ID.Workflow)
	}

	launchPlan := []string{""}
	if ID.LaunchPlan != "" {
		launchPlan = append(launchPlan, ID.LaunchPlan)
	}

	tx := r.db.Where(txWhereClause, ID.ResourceType, ID.Domain, project, workflow, launchPlan)
	tx.Order(priorityDescending).First(&resources)
	timer.Stop()

	if tx.Error != nil {
		return models.Resource{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	if tx.RecordNotFound() || len(resources) == 0 {
		return models.Resource{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound,
			"Resource [%+v] not found", ID)
	}
	return resources[0], nil
}

func (r *ResourceRepo) GetRaw(ctx context.Context, ID interfaces.ResourceID) (models.Resource, error) {
	if ID.Domain == "" || ID.ResourceType == "" {
		return models.Resource{}, r.errorTransformer.ToFlyteAdminError(errors.GetInvalidInputError(fmt.Sprintf("%v", ID)))
	}
	var model models.Resource
	timer := r.metrics.GetDuration.Start()
	tx := r.db.Where(&models.Resource{
		Project:      ID.Project,
		Domain:       ID.Domain,
		Workflow:     ID.Workflow,
		LaunchPlan:   ID.LaunchPlan,
		ResourceType: ID.ResourceType,
	}).First(&model)
	timer.Stop()
	if tx.Error != nil {
		return models.Resource{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	if tx.RecordNotFound() {
		return models.Resource{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound,
			"%v", ID)
	}
	return model, nil
}

func (r *ResourceRepo) ListAll(ctx context.Context, resourceType string) ([]models.Resource, error) {
	var resources []models.Resource
	timer := r.metrics.ListDuration.Start()

	tx := r.db.Where(&models.Resource{ResourceType: resourceType}).Order(priorityDescending).Find(&resources)
	timer.Stop()

	if tx.Error != nil {
		return nil, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return resources, nil
}

func (r *ResourceRepo) Delete(ctx context.Context, ID interfaces.ResourceID) error {
	var tx *gorm.DB
	r.metrics.DeleteDuration.Time(func() {
		tx = r.db.Where(&models.Resource{
			Project:      ID.Project,
			Domain:       ID.Domain,
			Workflow:     ID.Workflow,
			LaunchPlan:   ID.LaunchPlan,
			ResourceType: ID.ResourceType,
		}).Unscoped().Delete(models.Resource{})
	})
	if tx.Error != nil {
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	if tx.RecordNotFound() {
		return flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound, "%v", ID)
	}
	return nil
}

func NewResourceRepo(db *gorm.DB, errorTransformer errors.ErrorTransformer,
	scope promutils.Scope) interfaces.ResourceRepoInterface {
	metrics := newMetrics(scope)
	return &ResourceRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}
