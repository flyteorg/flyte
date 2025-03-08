package gorm

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"

	"gorm.io/gorm"

	"github.com/flyteorg/flyte/flyteadmin/concurrency/repositories/interfaces"
	adminRepoInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

// GormConcurrencyRepo implements the ConcurrencyRepoInterface using GORM
type GormConcurrencyRepo struct {
	db    *gorm.DB
	scope promutils.Scope
	repo  adminRepoInterfaces.Repository
}

// NewGormConcurrencyRepo creates a new GORM-based concurrency repository
func NewGormConcurrencyRepo(
	db *gorm.DB,
	repo adminRepoInterfaces.Repository,
	scope promutils.Scope,
) interfaces.ConcurrencyRepoInterface {
	return &GormConcurrencyRepo{
		db:    db,
		repo:  repo,
		scope: scope,
	}
}

// GetPendingExecutions retrieves all pending executions ordered by creation time
func (r *GormConcurrencyRepo) GetPendingExecutions(ctx context.Context) ([]models.Execution, error) {
	var executions []models.Execution

	// Query pending executions in FIFO order
	tx := r.db.Where("phase = ?", core.WorkflowExecution_PENDING.String()).
		Order("created_at ASC").
		Find(&executions)

	if tx.Error != nil {
		return nil, fmt.Errorf("failed to get pending executions: %w", tx.Error)
	}

	return executions, nil
}

// GetRunningExecutionsForLaunchPlan retrieves all non-terminal executions for a launch plan
func (r *GormConcurrencyRepo) GetRunningExecutionsForLaunchPlan(
	ctx context.Context,
	launchPlanID core.Identifier,
) ([]models.Execution, error) {
	var executions []models.Execution

	// Get non-terminal executions for the given launch plan
	tx := r.db.Where("launch_plan_project = ? AND launch_plan_domain = ? AND launch_plan_name = ? AND phase NOT IN (?, ?, ?, ?)",
		launchPlanID.Project,
		launchPlanID.Domain,
		launchPlanID.Name,
		core.WorkflowExecution_SUCCEEDED.String(),
		core.WorkflowExecution_FAILED.String(),
		core.WorkflowExecution_ABORTED.String(),
		core.WorkflowExecution_TIMED_OUT.String(),
	).Find(&executions)

	if tx.Error != nil {
		return nil, fmt.Errorf("failed to get running executions for launch plan: %w", tx.Error)
	}

	return executions, nil
}

// GetRunningExecutionsCount gets the count of non-terminal executions for a launch plan
func (r *GormConcurrencyRepo) GetRunningExecutionsCount(
	ctx context.Context,
	launchPlanID core.Identifier,
) (int, error) {
	var count int64

	// Count non-terminal executions for the given launch plan
	tx := r.db.Model(&models.Execution{}).
		Where("launch_plan_project = ? AND launch_plan_domain = ? AND launch_plan_name = ? AND phase NOT IN (?, ?, ?, ?)",
			launchPlanID.Project,
			launchPlanID.Domain,
			launchPlanID.Name,
			core.WorkflowExecution_SUCCEEDED.String(),
			core.WorkflowExecution_FAILED.String(),
			core.WorkflowExecution_ABORTED.String(),
			core.WorkflowExecution_TIMED_OUT.String(),
		).Count(&count)

	if tx.Error != nil {
		return 0, fmt.Errorf("failed to count running executions for launch plan: %w", tx.Error)
	}

	return int(count), nil
}

// GetActiveLaunchPlanWithConcurrency gets a launch plan with concurrency settings
func (r *GormConcurrencyRepo) GetActiveLaunchPlanWithConcurrency(
	ctx context.Context,
	launchPlanID core.Identifier,
) (*models.LaunchPlan, error) {
	var launchPlan models.LaunchPlan

	// Get the active launch plan with its concurrency settings
	tx := r.db.Where("project = ? AND domain = ? AND name = ? AND state = ? AND scheduler_policy IS NOT NULL",
		launchPlanID.Project,
		launchPlanID.Domain,
		launchPlanID.Name,
		admin.LaunchPlanState_ACTIVE.String(),
	).First(&launchPlan)

	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return nil, nil // No active launch plan with concurrency
		}
		return nil, fmt.Errorf("failed to get active launch plan: %w", tx.Error)
	}

	return &launchPlan, nil
}

// UpdateExecutionPhase updates the phase of an execution
func (r *GormConcurrencyRepo) UpdateExecutionPhase(
	ctx context.Context,
	executionID core.WorkflowExecutionIdentifier,
	phase core.WorkflowExecution_Phase,
) error {
	tx := r.db.Model(&models.Execution{}).
		Where("project = ? AND domain = ? AND name = ?",
			executionID.Project,
			executionID.Domain,
			executionID.Name,
		).
		Update("phase", phase.String())

	if tx.Error != nil {
		return fmt.Errorf("failed to update execution phase: %w", tx.Error)
	}

	return nil
}

// GetOldestRunningExecution gets the oldest running execution for a launch plan
func (r *GormConcurrencyRepo) GetOldestRunningExecution(
	ctx context.Context,
	launchPlanID core.Identifier,
) (*models.Execution, error) {
	var execution models.Execution

	// Get the oldest running execution for the given launch plan
	tx := r.db.Where("launch_plan_project = ? AND launch_plan_domain = ? AND launch_plan_name = ? AND phase NOT IN (?, ?, ?, ?)",
		launchPlanID.Project,
		launchPlanID.Domain,
		launchPlanID.Name,
		core.WorkflowExecution_SUCCEEDED.String(),
		core.WorkflowExecution_FAILED.String(),
		core.WorkflowExecution_ABORTED.String(),
		core.WorkflowExecution_TIMED_OUT.String(),
	).
		Order("created_at ASC").
		First(&execution)

	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return nil, nil // No running executions
		}
		return nil, fmt.Errorf("failed to get oldest running execution: %w", tx.Error)
	}

	return &execution, nil
}

// GetAllLaunchPlansWithConcurrency gets all launch plans with concurrency settings
func (r *GormConcurrencyRepo) GetAllLaunchPlansWithConcurrency(ctx context.Context) ([]models.LaunchPlan, error) {
	var launchPlans []models.LaunchPlan

	// Get all launch plans with concurrency settings
	tx := r.db.Where("scheduler_policy IS NOT NULL").Find(&launchPlans)

	if tx.Error != nil {
		return nil, fmt.Errorf("failed to get launch plans with concurrency: %w", tx.Error)
	}

	return launchPlans, nil
}

// GetExecutionByID retrieves an execution by its identifier
func (r *GormConcurrencyRepo) GetExecutionByID(ctx context.Context, executionID core.WorkflowExecutionIdentifier) (*models.Execution, error) {
	var execution models.Execution

	tx := r.db.Where("project = ? AND domain = ? AND name = ?",
		executionID.Project, executionID.Domain, executionID.Name).
		First(&execution)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return &execution, nil
}
