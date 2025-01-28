package tests

import (
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/rpc/adminservice"
	mockScope "github.com/flyteorg/flyte/flytestdlib/promutils"
)

type NewMockAdminServerInput struct {
	executionManager     *mocks.ExecutionInterface
	launchPlanManager    *mocks.LaunchPlanInterface
	nodeExecutionManager *mocks.NodeExecutionInterface
	projectManager       *mocks.ProjectInterface
	resourceManager      *mocks.ResourceInterface
	taskManager          *mocks.TaskInterface
	workflowManager      *mocks.WorkflowInterface
	taskExecutionManager *mocks.TaskExecutionInterface
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
