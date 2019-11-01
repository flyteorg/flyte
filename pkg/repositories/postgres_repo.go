package repositories

import (
	"github.com/jinzhu/gorm"
	"github.com/lyft/flytestdlib/promutils"

	"github.com/lyft/flyteadmin/pkg/repositories/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/gormimpl"
	"github.com/lyft/flyteadmin/pkg/repositories/interfaces"
)

type PostgresRepo struct {
	executionRepo     interfaces.ExecutionRepoInterface
	launchPlanRepo    interfaces.LaunchPlanRepoInterface
	projectRepo       interfaces.ProjectRepoInterface
	projectDomainRepo interfaces.ProjectDomainRepoInterface
	nodeExecutionRepo interfaces.NodeExecutionRepoInterface
	taskRepo          interfaces.TaskRepoInterface
	taskExecutionRepo interfaces.TaskExecutionRepoInterface
	workflowRepo      interfaces.WorkflowRepoInterface
}

func (p *PostgresRepo) ExecutionRepo() interfaces.ExecutionRepoInterface {
	return p.executionRepo
}

func (p *PostgresRepo) LaunchPlanRepo() interfaces.LaunchPlanRepoInterface {
	return p.launchPlanRepo
}

func (p *PostgresRepo) ProjectRepo() interfaces.ProjectRepoInterface {
	return p.projectRepo
}

func (p *PostgresRepo) ProjectDomainRepo() interfaces.ProjectDomainRepoInterface {
	return p.projectDomainRepo
}

func (p *PostgresRepo) NodeExecutionRepo() interfaces.NodeExecutionRepoInterface {
	return p.nodeExecutionRepo
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

func NewPostgresRepo(db *gorm.DB, errorTransformer errors.ErrorTransformer, scope promutils.Scope) RepositoryInterface {
	return &PostgresRepo{
		executionRepo:     gormimpl.NewExecutionRepo(db, errorTransformer, scope.NewSubScope("executions")),
		launchPlanRepo:    gormimpl.NewLaunchPlanRepo(db, errorTransformer, scope.NewSubScope("launch_plans")),
		projectRepo:       gormimpl.NewProjectRepo(db, errorTransformer, scope.NewSubScope("project")),
		projectDomainRepo: gormimpl.NewProjectDomainRepo(db, errorTransformer, scope.NewSubScope("project_domain")),
		nodeExecutionRepo: gormimpl.NewNodeExecutionRepo(db, errorTransformer, scope.NewSubScope("node_executions")),
		taskRepo:          gormimpl.NewTaskRepo(db, errorTransformer, scope.NewSubScope("tasks")),
		taskExecutionRepo: gormimpl.NewTaskExecutionRepo(db, errorTransformer, scope.NewSubScope("task_executions")),
		workflowRepo:      gormimpl.NewWorkflowRepo(db, errorTransformer, scope.NewSubScope("workflows")),
	}
}
