package mocks

import (
	"github.com/lyft/flyteadmin/pkg/repositories"
	"github.com/lyft/flyteadmin/pkg/repositories/interfaces"
)

type MockRepository struct {
	taskRepo          interfaces.TaskRepoInterface
	workflowRepo      interfaces.WorkflowRepoInterface
	launchPlanRepo    interfaces.LaunchPlanRepoInterface
	executionRepo     interfaces.ExecutionRepoInterface
	nodeExecutionRepo interfaces.NodeExecutionRepoInterface
	projectRepo       interfaces.ProjectRepoInterface
	projectDomainRepo interfaces.ProjectDomainRepoInterface
	taskExecutionRepo interfaces.TaskExecutionRepoInterface
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

func (r *MockRepository) ProjectDomainRepo() interfaces.ProjectDomainRepoInterface {
	return r.projectDomainRepo
}

func (r *MockRepository) TaskExecutionRepo() interfaces.TaskExecutionRepoInterface {
	return r.taskExecutionRepo
}

func NewMockRepository() repositories.RepositoryInterface {
	return &MockRepository{
		taskRepo:          NewMockTaskRepo(),
		workflowRepo:      NewMockWorkflowRepo(),
		launchPlanRepo:    NewMockLaunchPlanRepo(),
		executionRepo:     NewMockExecutionRepo(),
		nodeExecutionRepo: NewMockNodeExecutionRepo(),
		projectRepo:       NewMockProjectRepo(),
		projectDomainRepo: NewMockProjectDomainRepo(),
		taskExecutionRepo: NewMockTaskExecutionRepo(),
	}
}
