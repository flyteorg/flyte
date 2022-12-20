package tests

import (
	"github.com/flyteorg/flyteadmin/pkg/manager/mocks"
	"github.com/flyteorg/flyteadmin/pkg/rpc/adminservice"
	mockScope "github.com/flyteorg/flytestdlib/promutils"
)

type NewMockAdminServerInput struct {
	executionManager     *mocks.MockExecutionManager
	launchPlanManager    *mocks.MockLaunchPlanManager
	nodeExecutionManager *mocks.MockNodeExecutionManager
	projectManager       *mocks.MockProjectManager
	resourceManager      *mocks.MockResourceManager
	taskManager          *mocks.MockTaskManager
	workflowManager      *mocks.MockWorkflowManager
	taskExecutionManager *mocks.MockTaskExecutionManager
}

func NewMockAdminServer(input NewMockAdminServerInput) *adminservice.AdminService {
	var testScope = mockScope.NewTestScope()
	return &adminservice.AdminService{
		ExecutionManager:     input.executionManager,
		LaunchPlanManager:    input.launchPlanManager,
		NodeExecutionManager: input.nodeExecutionManager,
		TaskManager:          input.taskManager,
		ProjectManager:       input.projectManager,
		ResourceManager:      input.resourceManager,
		WorkflowManager:      input.workflowManager,
		TaskExecutionManager: input.taskExecutionManager,
		Metrics:              adminservice.InitMetrics(testScope),
	}
}
