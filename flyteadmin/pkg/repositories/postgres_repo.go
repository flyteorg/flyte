package repositories

import (
	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/gormimpl"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	schedulerGormImpl "github.com/flyteorg/flyteadmin/scheduler/repositories/gormimpl"
	schedulerInterfaces "github.com/flyteorg/flyteadmin/scheduler/repositories/interfaces"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/jinzhu/gorm"
)

type PostgresRepo struct {
	executionRepo                interfaces.ExecutionRepoInterface
	executionEventRepo           interfaces.ExecutionEventRepoInterface
	namedEntityRepo              interfaces.NamedEntityRepoInterface
	launchPlanRepo               interfaces.LaunchPlanRepoInterface
	projectRepo                  interfaces.ProjectRepoInterface
	nodeExecutionRepo            interfaces.NodeExecutionRepoInterface
	nodeExecutionEventRepo       interfaces.NodeExecutionEventRepoInterface
	taskRepo                     interfaces.TaskRepoInterface
	taskExecutionRepo            interfaces.TaskExecutionRepoInterface
	workflowRepo                 interfaces.WorkflowRepoInterface
	resourceRepo                 interfaces.ResourceRepoInterface
	schedulableEntityRepo        schedulerInterfaces.SchedulableEntityRepoInterface
	scheduleEntitiesSnapshotRepo schedulerInterfaces.ScheduleEntitiesSnapShotRepoInterface
}

func (p *PostgresRepo) ExecutionRepo() interfaces.ExecutionRepoInterface {
	return p.executionRepo
}

func (p *PostgresRepo) ExecutionEventRepo() interfaces.ExecutionEventRepoInterface {
	return p.executionEventRepo
}

func (p *PostgresRepo) LaunchPlanRepo() interfaces.LaunchPlanRepoInterface {
	return p.launchPlanRepo
}

func (p *PostgresRepo) NamedEntityRepo() interfaces.NamedEntityRepoInterface {
	return p.namedEntityRepo
}

func (p *PostgresRepo) ProjectRepo() interfaces.ProjectRepoInterface {
	return p.projectRepo
}

func (p *PostgresRepo) NodeExecutionRepo() interfaces.NodeExecutionRepoInterface {
	return p.nodeExecutionRepo
}

func (p *PostgresRepo) NodeExecutionEventRepo() interfaces.NodeExecutionEventRepoInterface {
	return p.nodeExecutionEventRepo
}

func (p *PostgresRepo) TaskRepo() interfaces.TaskRepoInterface {
	return p.taskRepo
}

func (p *PostgresRepo) TaskExecutionRepo() interfaces.TaskExecutionRepoInterface {
	return p.taskExecutionRepo
}

func (p *PostgresRepo) WorkflowRepo() interfaces.WorkflowRepoInterface {
	return p.workflowRepo
}

func (p *PostgresRepo) ResourceRepo() interfaces.ResourceRepoInterface {
	return p.resourceRepo
}

func (p *PostgresRepo) SchedulableEntityRepo() schedulerInterfaces.SchedulableEntityRepoInterface {
	return p.schedulableEntityRepo
}

func (p *PostgresRepo) ScheduleEntitiesSnapshotRepo() schedulerInterfaces.ScheduleEntitiesSnapShotRepoInterface {
	return p.scheduleEntitiesSnapshotRepo
}

func NewPostgresRepo(db *gorm.DB, errorTransformer errors.ErrorTransformer, scope promutils.Scope) RepositoryInterface {
	return &PostgresRepo{
		executionRepo:                gormimpl.NewExecutionRepo(db, errorTransformer, scope.NewSubScope("executions")),
		executionEventRepo:           gormimpl.NewExecutionEventRepo(db, errorTransformer, scope.NewSubScope("execution_events")),
		launchPlanRepo:               gormimpl.NewLaunchPlanRepo(db, errorTransformer, scope.NewSubScope("launch_plans")),
		projectRepo:                  gormimpl.NewProjectRepo(db, errorTransformer, scope.NewSubScope("project")),
		namedEntityRepo:              gormimpl.NewNamedEntityRepo(db, errorTransformer, scope.NewSubScope("named_entity")),
		nodeExecutionRepo:            gormimpl.NewNodeExecutionRepo(db, errorTransformer, scope.NewSubScope("node_executions")),
		nodeExecutionEventRepo:       gormimpl.NewNodeExecutionEventRepo(db, errorTransformer, scope.NewSubScope("node_execution_events")),
		taskRepo:                     gormimpl.NewTaskRepo(db, errorTransformer, scope.NewSubScope("tasks")),
		taskExecutionRepo:            gormimpl.NewTaskExecutionRepo(db, errorTransformer, scope.NewSubScope("task_executions")),
		workflowRepo:                 gormimpl.NewWorkflowRepo(db, errorTransformer, scope.NewSubScope("workflows")),
		resourceRepo:                 gormimpl.NewResourceRepo(db, errorTransformer, scope.NewSubScope("resources")),
		schedulableEntityRepo:        schedulerGormImpl.NewSchedulableEntityRepo(db, errorTransformer, scope.NewSubScope("schedulable_entity")),
		scheduleEntitiesSnapshotRepo: schedulerGormImpl.NewScheduleEntitiesSnapshotRepo(db, errorTransformer, scope.NewSubScope("schedule_entities_snapshot")),
	}
}
