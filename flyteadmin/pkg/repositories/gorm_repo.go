package repositories

import (
	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/gormimpl"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	schedulerGormImpl "github.com/flyteorg/flyteadmin/scheduler/repositories/gormimpl"
	schedulerInterfaces "github.com/flyteorg/flyteadmin/scheduler/repositories/interfaces"
	"github.com/flyteorg/flytestdlib/promutils"
	"gorm.io/gorm"
)

type GormRepo struct {
	db                           *gorm.DB
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
	descriptionEntityRepo        interfaces.DescriptionEntityRepoInterface
	schedulableEntityRepo        schedulerInterfaces.SchedulableEntityRepoInterface
	scheduleEntitiesSnapshotRepo schedulerInterfaces.ScheduleEntitiesSnapShotRepoInterface
	signalRepo                   interfaces.SignalRepoInterface
}

func (r *GormRepo) ExecutionRepo() interfaces.ExecutionRepoInterface {
	return r.executionRepo
}

func (r *GormRepo) ExecutionEventRepo() interfaces.ExecutionEventRepoInterface {
	return r.executionEventRepo
}

func (r *GormRepo) LaunchPlanRepo() interfaces.LaunchPlanRepoInterface {
	return r.launchPlanRepo
}

func (r *GormRepo) NamedEntityRepo() interfaces.NamedEntityRepoInterface {
	return r.namedEntityRepo
}

func (r *GormRepo) ProjectRepo() interfaces.ProjectRepoInterface {
	return r.projectRepo
}

func (r *GormRepo) NodeExecutionRepo() interfaces.NodeExecutionRepoInterface {
	return r.nodeExecutionRepo
}

func (r *GormRepo) NodeExecutionEventRepo() interfaces.NodeExecutionEventRepoInterface {
	return r.nodeExecutionEventRepo
}

func (r *GormRepo) TaskRepo() interfaces.TaskRepoInterface {
	return r.taskRepo
}

func (r *GormRepo) TaskExecutionRepo() interfaces.TaskExecutionRepoInterface {
	return r.taskExecutionRepo
}

func (r *GormRepo) WorkflowRepo() interfaces.WorkflowRepoInterface {
	return r.workflowRepo
}

func (r *GormRepo) ResourceRepo() interfaces.ResourceRepoInterface {
	return r.resourceRepo
}

func (r *GormRepo) DescriptionEntityRepo() interfaces.DescriptionEntityRepoInterface {
	return r.descriptionEntityRepo
}

func (r *GormRepo) SchedulableEntityRepo() schedulerInterfaces.SchedulableEntityRepoInterface {
	return r.schedulableEntityRepo
}

func (r *GormRepo) ScheduleEntitiesSnapshotRepo() schedulerInterfaces.ScheduleEntitiesSnapShotRepoInterface {
	return r.scheduleEntitiesSnapshotRepo
}

func (r *GormRepo) SignalRepo() interfaces.SignalRepoInterface {
	return r.signalRepo
}

func (r *GormRepo) GetGormDB() *gorm.DB {
	return r.db
}

func NewGormRepo(db *gorm.DB, errorTransformer errors.ErrorTransformer, scope promutils.Scope) interfaces.Repository {
	return &GormRepo{
		db:                           db,
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
		descriptionEntityRepo:        gormimpl.NewDescriptionEntityRepo(db, errorTransformer, scope.NewSubScope("description_entities")),
		schedulableEntityRepo:        schedulerGormImpl.NewSchedulableEntityRepo(db, errorTransformer, scope.NewSubScope("schedulable_entity")),
		scheduleEntitiesSnapshotRepo: schedulerGormImpl.NewScheduleEntitiesSnapshotRepo(db, errorTransformer, scope.NewSubScope("schedule_entities_snapshot")),
		signalRepo:                   gormimpl.NewSignalRepo(db, errorTransformer, scope.NewSubScope("signals")),
	}
}
