package mocks

import (
	"github.com/lyft/flyteadmin/pkg/repositories"
	"github.com/lyft/flyteadmin/pkg/repositories/interfaces"
)

type MockRepository struct {
	taskRepo                    interfaces.TaskRepoInterface
	workflowRepo                interfaces.WorkflowRepoInterface
	launchPlanRepo              interfaces.LaunchPlanRepoInterface
	executionRepo               interfaces.ExecutionRepoInterface
	nodeExecutionRepo           interfaces.NodeExecutionRepoInterface
	projectRepo                 interfaces.ProjectRepoInterface
	projectAttributesRepo       interfaces.ProjectAttributesRepoInterface
	projectDomainAttributesRepo interfaces.ProjectDomainAttributesRepoInterface
	workflowAttributesRepo      interfaces.WorkflowAttributesRepoInterface
	taskExecutionRepo           interfaces.TaskExecutionRepoInterface
	namedEntityRepo             interfaces.NamedEntityRepoInterface
}

func (r *MockRepository) TaskRepo() interfaces.TaskRepoInterface {
	return r.taskRepo
}

func (r *MockRepository) WorkflowRepo() interfaces.WorkflowRepoInterface {
	return r.workflowRepo
}

func (r *MockRepository) LaunchPlanRepo() interfaces.LaunchPlanRepoInterface {
	return r.launchPlanRepo
}

func (r *MockRepository) ExecutionRepo() interfaces.ExecutionRepoInterface {
	return r.executionRepo
}

func (r *MockRepository) NodeExecutionRepo() interfaces.NodeExecutionRepoInterface {
	return r.nodeExecutionRepo
}

func (r *MockRepository) ProjectRepo() interfaces.ProjectRepoInterface {
	return r.projectRepo
}

func (r *MockRepository) ProjectDomainAttributesRepo() interfaces.ProjectDomainAttributesRepoInterface {
	return r.projectDomainAttributesRepo
}

func (r *MockRepository) WorkflowAttributesRepo() interfaces.WorkflowAttributesRepoInterface {
	return r.workflowAttributesRepo
}

func (r *MockRepository) ProjectAttributesRepo() interfaces.ProjectAttributesRepoInterface {
	return r.projectAttributesRepo
}

func (r *MockRepository) TaskExecutionRepo() interfaces.TaskExecutionRepoInterface {
	return r.taskExecutionRepo
}

func (r *MockRepository) NamedEntityRepo() interfaces.NamedEntityRepoInterface {
	return r.namedEntityRepo
}

func NewMockRepository() repositories.RepositoryInterface {
	return &MockRepository{
		taskRepo:                    NewMockTaskRepo(),
		workflowRepo:                NewMockWorkflowRepo(),
		launchPlanRepo:              NewMockLaunchPlanRepo(),
		executionRepo:               NewMockExecutionRepo(),
		nodeExecutionRepo:           NewMockNodeExecutionRepo(),
		projectRepo:                 NewMockProjectRepo(),
		projectAttributesRepo:       NewMockProjectAttributesRepo(),
		projectDomainAttributesRepo: NewMockProjectDomainAttributesRepo(),
		workflowAttributesRepo:      NewMockWorkflowAttributesRepo(),
		taskExecutionRepo:           NewMockTaskExecutionRepo(),
		namedEntityRepo:             NewMockNamedEntityRepo(),
	}
}
