package mocks

import (
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	sIface "github.com/flyteorg/flyteadmin/scheduler/repositories/interfaces"
	sMocks "github.com/flyteorg/flyteadmin/scheduler/repositories/mocks"
	"gorm.io/gorm"
)

type MockRepository struct {
	taskRepo                      interfaces.TaskRepoInterface
	workflowRepo                  interfaces.WorkflowRepoInterface
	launchPlanRepo                interfaces.LaunchPlanRepoInterface
	executionRepo                 interfaces.ExecutionRepoInterface
	ExecutionEventRepoIface       interfaces.ExecutionEventRepoInterface
	nodeExecutionRepo             interfaces.NodeExecutionRepoInterface
	NodeExecutionEventRepoIface   interfaces.NodeExecutionEventRepoInterface
	ProjectRepoIface              interfaces.ProjectRepoInterface
	resourceRepo                  interfaces.ResourceRepoInterface
	taskExecutionRepo             interfaces.TaskExecutionRepoInterface
	namedEntityRepo               interfaces.NamedEntityRepoInterface
	descriptionEntityRepo         interfaces.DescriptionEntityRepoInterface
	schedulableEntityRepo         sIface.SchedulableEntityRepoInterface
	schedulableEntitySnapshotRepo sIface.ScheduleEntitiesSnapShotRepoInterface
	signalRepo                    interfaces.SignalRepoInterface
}

func (r *MockRepository) GetGormDB() *gorm.DB {
	return nil
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
	return r.ProjectRepoIface
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

func (r *MockRepository) DescriptionEntityRepo() interfaces.DescriptionEntityRepoInterface {
	return r.descriptionEntityRepo
}

func (r *MockRepository) SignalRepo() interfaces.SignalRepoInterface {
	return r.signalRepo
}

func NewMockRepository() interfaces.Repository {
	return &MockRepository{
		taskRepo:                      NewMockTaskRepo(),
		workflowRepo:                  NewMockWorkflowRepo(),
		launchPlanRepo:                NewMockLaunchPlanRepo(),
		executionRepo:                 NewMockExecutionRepo(),
		nodeExecutionRepo:             NewMockNodeExecutionRepo(),
		ProjectRepoIface:              NewMockProjectRepo(),
		resourceRepo:                  NewMockResourceRepo(),
		taskExecutionRepo:             NewMockTaskExecutionRepo(),
		namedEntityRepo:               NewMockNamedEntityRepo(),
		descriptionEntityRepo:         NewMockDescriptionEntityRepo(),
		ExecutionEventRepoIface:       &ExecutionEventRepoInterface{},
		NodeExecutionEventRepoIface:   &NodeExecutionEventRepoInterface{},
		schedulableEntityRepo:         &sMocks.SchedulableEntityRepoInterface{},
		schedulableEntitySnapshotRepo: &sMocks.ScheduleEntitiesSnapShotRepoInterface{},
		signalRepo:                    &SignalRepoInterface{},
	}
}
