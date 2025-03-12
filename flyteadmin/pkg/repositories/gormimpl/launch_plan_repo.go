package gormimpl

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/flyteorg/flyte/flyteadmin/auth/isolation"
	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/util"
	adminErrors "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

var (
	launchPlanColumnNames = common.ResourceColumns{Project: "launch_plans.project", Domain: "launch_plans.domain"}
)

const launchPlanTableName = "launch_plans"

type launchPlanMetrics struct {
	SetActiveDuration promutils.StopWatch
}

// Implementation of LaunchPlanRepoInterface.
type LaunchPlanRepo struct {
	db                *gorm.DB
	errorTransformer  adminErrors.ErrorTransformer
	metrics           gormMetrics
	launchPlanMetrics launchPlanMetrics
}

func (r *LaunchPlanRepo) Create(ctx context.Context, input models.LaunchPlan) error {
	if err := util.FilterResourceMutation(ctx, input.Project, input.Domain); err != nil {
		return err
	}
	timer := r.metrics.CreateDuration.Start()
	tx := r.db.WithContext(ctx).Omit("id").Create(&input)
	timer.Stop()
	if tx.Error != nil {
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return nil
}

func (r *LaunchPlanRepo) Update(ctx context.Context, input models.LaunchPlan) error {
	if err := util.FilterResourceMutation(ctx, input.Project, input.Domain); err != nil {
		return err
	}
	timer := r.metrics.UpdateDuration.Start()
	tx := r.db.WithContext(ctx).Model(&input).Updates(input)
	timer.Stop()
	if err := tx.Error; err != nil {
		return r.errorTransformer.ToFlyteAdminError(err)
	}
	return nil
}

func (r *LaunchPlanRepo) Get(ctx context.Context, input interfaces.Identifier) (models.LaunchPlan, error) {
	var launchPlan models.LaunchPlan
	timer := r.metrics.GetDuration.Start()
	isolationFilter := util.GetIsolationFilter(ctx, isolation.DomainTargetResourceScopeDepth, launchPlanColumnNames)
	tx := r.db.WithContext(ctx).Where(&models.LaunchPlan{
		LaunchPlanKey: models.LaunchPlanKey{
			Project: input.Project,
			Domain:  input.Domain,
			Name:    input.Name,
			Version: input.Version,
		},
	})
	if isolationFilter != nil {
		cleanSession := tx.Session(&gorm.Session{NewDB: true})
		tx = tx.Where(cleanSession.Scopes(isolationFilter.GetScopes()...))
	}
	tx = tx.Take(&launchPlan)
	timer.Stop()

	if tx.Error != nil && errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return models.LaunchPlan{},
			adminErrors.GetMissingEntityError(core.ResourceType_LAUNCH_PLAN.String(), &core.Identifier{
				Project: input.Project,
				Domain:  input.Domain,
				Name:    input.Name,
				Version: input.Version,
			})
	} else if tx.Error != nil {
		return models.LaunchPlan{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return launchPlan, nil
}

// This operation is performed as a two-step transaction because only one launch plan version can be active at a time.
// Transactional semantics are used to guarantee that setting the desired launch plan to active also disables
// the existing launch plan version (if any).
func (r *LaunchPlanRepo) SetActive(
	ctx context.Context, toEnable models.LaunchPlan, toDisable *models.LaunchPlan) error {
	if err := util.FilterResourceMutation(ctx, toEnable.Project, toEnable.Domain); err != nil {
		return err
	}
	if toDisable != nil {
		if err := util.FilterResourceMutation(ctx, toDisable.Project, toDisable.Domain); err != nil {
			return err
		}
	}
	timer := r.launchPlanMetrics.SetActiveDuration.Start()
	defer timer.Stop()
	// Use a transaction to guarantee no partial updates.
	tx := r.db.WithContext(ctx).Begin()

	// There is a launch plan to disable as part of this transaction
	if toDisable != nil {
		tx.Model(&toDisable).UpdateColumns(toDisable)
		if err := tx.Error; err != nil {
			tx.Rollback()
			return r.errorTransformer.ToFlyteAdminError(err)
		}
	}

	// And update the desired version.
	tx.Model(&toEnable).UpdateColumns(toEnable)
	if err := tx.Error; err != nil {
		tx.Rollback()
		return r.errorTransformer.ToFlyteAdminError(err)
	}
	if err := tx.Commit().Error; err != nil {
		return r.errorTransformer.ToFlyteAdminError(err)
	}
	return nil
}

func (r *LaunchPlanRepo) List(ctx context.Context, input interfaces.ListResourceInput) (
	interfaces.LaunchPlanCollectionOutput, error) {
	// First validate input.
	if err := ValidateListInput(input); err != nil {
		return interfaces.LaunchPlanCollectionOutput{}, err
	}
	isolationFilter := util.GetIsolationFilter(ctx, isolation.DomainTargetResourceScopeDepth, launchPlanColumnNames)
	var launchPlans []models.LaunchPlan
	tx := r.db.WithContext(ctx).Limit(input.Limit).Offset(input.Offset)

	// Add join conditions
	tx = tx.Joins("inner join workflows on launch_plans.workflow_id = workflows.id")

	// Apply filters
	tx, err := applyScopedFilters(tx, input.InlineFilters, input.MapFilters, isolationFilter)
	if err != nil {
		return interfaces.LaunchPlanCollectionOutput{}, err
	}
	// Apply sort ordering.
	if input.SortParameter != nil {
		tx = tx.Order(input.SortParameter.GetGormOrderExpr())
	}

	timer := r.metrics.ListDuration.Start()
	tx.Find(&launchPlans)
	timer.Stop()
	if tx.Error != nil {
		logger.Warningf(ctx,
			"Failed to list launch plans by workflow with input [%+v] with err: %+v", input, tx.Error)
		return interfaces.LaunchPlanCollectionOutput{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}

	return interfaces.LaunchPlanCollectionOutput{
		LaunchPlans: launchPlans,
	}, nil
}

func (r *LaunchPlanRepo) ListLaunchPlanIdentifiers(ctx context.Context, input interfaces.ListResourceInput) (
	interfaces.LaunchPlanCollectionOutput, error) {

	// Validate input, input must have a limit
	if err := ValidateListInput(input); err != nil {
		return interfaces.LaunchPlanCollectionOutput{}, err
	}

	isolationFilter := util.GetIsolationFilter(ctx, isolation.DomainTargetResourceScopeDepth, launchPlanColumnNames)
	tx := r.db.WithContext(ctx).Model(models.LaunchPlan{}).Limit(input.Limit).Offset(input.Offset)

	// Apply filters
	tx, err := applyFilters(tx, input.InlineFilters, input.MapFilters, isolationFilter)
	if err != nil {
		return interfaces.LaunchPlanCollectionOutput{}, err
	}
	// Apply sort ordering.
	if input.SortParameter != nil {
		tx = tx.Order(input.SortParameter.GetGormOrderExpr())
	}

	// Scan the results into a list of launch plans
	var launchPlans []models.LaunchPlan
	timer := r.metrics.ListIdentifiersDuration.Start()
	tx.Select([]string{Project, Domain, Name}).Group(identifierGroupBy).Scan(&launchPlans)
	timer.Stop()
	if tx.Error != nil {
		return interfaces.LaunchPlanCollectionOutput{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}

	return interfaces.LaunchPlanCollectionOutput{
		LaunchPlans: launchPlans,
	}, nil

}

// Returns an instance of LaunchPlanRepoInterface
func NewLaunchPlanRepo(
	db *gorm.DB, errorTransformer adminErrors.ErrorTransformer, scope promutils.Scope) interfaces.LaunchPlanRepoInterface {
	metrics := newMetrics(scope)
	launchPlanMetrics := launchPlanMetrics{
		SetActiveDuration: scope.MustNewStopWatch(
			"set_active",
			"time taken to set a launch plan to active (and disable the currently active version)", time.Millisecond),
	}

	return &LaunchPlanRepo{
		db:                db,
		errorTransformer:  errorTransformer,
		metrics:           metrics,
		launchPlanMetrics: launchPlanMetrics,
	}
}
