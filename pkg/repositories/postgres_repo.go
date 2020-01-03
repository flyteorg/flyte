package repositories

import (
	"github.com/jinzhu/gorm"
	"github.com/lyft/flyteadmin/pkg/repositories/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/gormimpl"
	"github.com/lyft/flyteadmin/pkg/repositories/interfaces"
	"github.com/lyft/flytestdlib/promutils"
)

type PostgresRepo struct {
	executionRepo               interfaces.ExecutionRepoInterface
	namedEntityRepo             interfaces.NamedEntityRepoInterface
	launchPlanRepo              interfaces.LaunchPlanRepoInterface
	projectRepo                 interfaces.ProjectRepoInterface
	projectAttributesRepo       interfaces.ProjectAttributesRepoInterface
	projectDomainAttributesRepo interfaces.ProjectDomainAttributesRepoInterface
	nodeExecutionRepo           interfaces.NodeExecutionRepoInterface
	taskRepo                    interfaces.TaskRepoInterface
	taskExecutionRepo           interfaces.TaskExecutionRepoInterface
	workflowRepo                interfaces.WorkflowRepoInterface
	workflowAttributesRepo      interfaces.WorkflowAttributesRepoInterface
}

func (p *PostgresRepo) ExecutionRepo() interfaces.ExecutionRepoInterface {
	return p.executionRepo
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

func (p *PostgresRepo) ProjectAttributesRepo() interfaces.ProjectAttributesRepoInterface {
	return p.projectAttributesRepo
}

func (p *PostgresRepo) ProjectDomainAttributesRepo() interfaces.ProjectDomainAttributesRepoInterface {
	return p.projectDomainAttributesRepo
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

func (p *PostgresRepo) WorkflowAttributesRepo() interfaces.WorkflowAttributesRepoInterface {
	return p.workflowAttributesRepo
}

func NewPostgresRepo(db *gorm.DB, errorTransformer errors.ErrorTransformer, scope promutils.Scope) RepositoryInterface {
	return &PostgresRepo{
		executionRepo:               gormimpl.NewExecutionRepo(db, errorTransformer, scope.NewSubScope("executions")),
		launchPlanRepo:              gormimpl.NewLaunchPlanRepo(db, errorTransformer, scope.NewSubScope("launch_plans")),
		projectRepo:                 gormimpl.NewProjectRepo(db, errorTransformer, scope.NewSubScope("project")),
		projectAttributesRepo:       gormimpl.NewProjectAttributesRepo(db, errorTransformer, scope.NewSubScope("project_attrs")),
		projectDomainAttributesRepo: gormimpl.NewProjectDomainAttributesRepo(db, errorTransformer, scope.NewSubScope("project_domain_attrs")),
		namedEntityRepo:             gormimpl.NewNamedEntityRepo(db, errorTransformer, scope.NewSubScope("named_entity")),
		nodeExecutionRepo:           gormimpl.NewNodeExecutionRepo(db, errorTransformer, scope.NewSubScope("node_executions")),
		taskRepo:                    gormimpl.NewTaskRepo(db, errorTransformer, scope.NewSubScope("tasks")),
		taskExecutionRepo:           gormimpl.NewTaskExecutionRepo(db, errorTransformer, scope.NewSubScope("task_executions")),
		workflowRepo:                gormimpl.NewWorkflowRepo(db, errorTransformer, scope.NewSubScope("workflows")),
		workflowAttributesRepo:      gormimpl.NewWorkflowAttributesRepo(db, errorTransformer, scope.NewSubScope("workflow_attrs")),
	}
}
