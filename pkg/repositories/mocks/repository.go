package mocks

import (
	"github.com/flyteorg/flyteadmin/pkg/repositories"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	sIface "github.com/flyteorg/flyteadmin/scheduler/repositories/interfaces"
	sMocks "github.com/flyteorg/flyteadmin/scheduler/repositories/mocks"
)

type MockRepository struct {
	taskRepo                      interfaces.TaskRepoInterface
	workflowRepo                  interfaces.WorkflowRepoInterface
	launchPlanRepo                interfaces.LaunchPlanRepoInterface
	executionRepo                 interfaces.ExecutionRepoInterface
	ExecutionEventRepoIface       interfaces.ExecutionEventRepoInterface
	nodeExecutionRepo             interfaces.NodeExecutionRepoInterface
	NodeExecutionEventRepoIface   interfaces.NodeExecutionEventRepoInterface
	projectRepo                   interfaces.ProjectRepoInterface
	resourceRepo                  interfaces.ResourceRepoInterface
	taskExecutionRepo             interfaces.TaskExecutionRepoInterface
	namedEntityRepo               interfaces.NamedEntityRepoInterface
	schedulableEntityRepo         sIface.SchedulableEntityRepoInterface
	schedulableEntitySnapshotRepo sIface.ScheduleEntitiesSnapShotRepoInterface
}

func (r *MockRepository) SchedulableEntityRepo() sIface.SchedulableEntityRepoInterface {
	return r.schedulableEntityRepo
}

func (r *MockRepository) ScheduleEntitiesSnapshotRepo() sIface.ScheduleEntitiesSnapShotRepoInterface {
	return r.schedulableEntitySnapshotRepo
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

func (r *MockRepository) ExecutionEventRepo() interfaces.ExecutionEventRepoInterface {
	return r.ExecutionEventRepoIface
}

func (r *MockRepository) NodeExecutionRepo() interfaces.NodeExecutionRepoInterface {
	return r.nodeExecutionRepo
}

func (r *MockRepository) NodeExecutionEventRepo() interfaces.NodeExecutionEventRepoInterface {
	return r.NodeExecutionEventRepoIface
}

func (r *MockRepository) ProjectRepo() interfaces.ProjectRepoInterface {
	return r.projectRepo
}

func (r *MockRepository) ResourceRepo() interfaces.ResourceRepoInterface {
	return r.resourceRepo
}

func (r *MockRepository) TaskExecutionRepo() interfaces.TaskExecutionRepoInterface {
	return r.taskExecutionRepo
}

func (r *MockRepository) NamedEntityRepo() interfaces.NamedEntityRepoInterface {
	return r.namedEntityRepo
}

func NewMockRepository() repositories.RepositoryInterface {
	return &MockRepository{
		taskRepo:                      NewMockTaskRepo(),
		workflowRepo:                  NewMockWorkflowRepo(),
		launchPlanRepo:                NewMockLaunchPlanRepo(),
		executionRepo:                 NewMockExecutionRepo(),
		nodeExecutionRepo:             NewMockNodeExecutionRepo(),
		projectRepo:                   NewMockProjectRepo(),
		resourceRepo:                  NewMockResourceRepo(),
		taskExecutionRepo:             NewMockTaskExecutionRepo(),
		namedEntityRepo:               NewMockNamedEntityRepo(),
		ExecutionEventRepoIface:       &ExecutionEventRepoInterface{},
		NodeExecutionEventRepoIface:   &NodeExecutionEventRepoInterface{},
		schedulableEntityRepo:         &sMocks.SchedulableEntityRepoInterface{},
		schedulableEntitySnapshotRepo: &sMocks.ScheduleEntitiesSnapShotRepoInterface{},
	}
}
