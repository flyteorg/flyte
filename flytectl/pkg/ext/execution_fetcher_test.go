package ext

import (
	"context"
	"fmt"
	"testing"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	executionResponse *admin.Execution
)

func getExecutionFetcherSetup() {
	ctx = context.Background()
	adminClient = new(mocks.AdminServiceClient)
	adminFetcherExt = AdminFetcherExtClient{AdminClient: adminClient}
	projectValue := "dummyProject"
	domainValue := "domainValue"
	executionNameValue := "execName"
	launchPlanNameValue := "launchPlanNameValue"
	launchPlanVersionValue := "launchPlanVersionValue"
	workflowNameValue := "workflowNameValue"
	workflowVersionValue := "workflowVersionValue"
	executionResponse = &admin.Execution{
		Id: &core.WorkflowExecutionIdentifier{
			Project: projectValue,
			Domain:  domainValue,
			Name:    executionNameValue,
		},
		Spec: &admin.ExecutionSpec{
			LaunchPlan: &core.Identifier{
				Project: projectValue,
				Domain:  domainValue,
				Name:    launchPlanNameValue,
				Version: launchPlanVersionValue,
			},
		},
		Closure: &admin.ExecutionClosure{
			WorkflowId: &core.Identifier{
				Project: projectValue,
				Domain:  domainValue,
				Name:    workflowNameValue,
				Version: workflowVersionValue,
			},
			Phase: core.WorkflowExecution_SUCCEEDED,
		},
	}
}

func TestFetchExecutionVersion(t *testing.T) {
	getExecutionFetcherSetup()
	adminClient.EXPECT().GetExecution(mock.Anything, mock.Anything).Return(executionResponse, nil)
	_, err := adminFetcherExt.FetchExecution(ctx, "execName", "dummyProject", "domainValue")
	assert.Nil(t, err)
}

func TestFetchExecutionError(t *testing.T) {
	getExecutionFetcherSetup()
	adminClient.EXPECT().GetExecution(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
	_, err := adminFetcherExt.FetchExecution(ctx, "execName", "dummyProject", "domainValue")
	assert.Equal(t, fmt.Errorf("failed"), err)
}

func TestFetchNodeExecutionDetails(t *testing.T) {
	getExecutionFetcherSetup()
	first := &admin.NodeExecution{Id: &core.NodeExecutionIdentifier{NodeId: "first"}}
	last := &admin.NodeExecution{Id: &core.NodeExecutionIdentifier{NodeId: "last"}}
	adminClient.EXPECT().ListNodeExecutions(mock.Anything, mock.Anything).
		Return(&admin.NodeExecutionList{NodeExecutions: []*admin.NodeExecution{first}, Token: "next"}, nil).Once()
	adminClient.EXPECT().ListNodeExecutions(mock.Anything, mock.MatchedBy(func(r *admin.NodeExecutionListRequest) bool {
		return r.GetToken() == "next" && r.GetUniqueParentId() == "parent" && r.GetWorkflowExecutionId().GetName() == "execName"
	})).Return(&admin.NodeExecutionList{NodeExecutions: []*admin.NodeExecution{last}}, nil).Once()

	nodes, err := adminFetcherExt.FetchNodeExecutionDetails(ctx, "execName", "dummyProject", "domainValue", "parent")
	require.NoError(t, err)
	assert.Equal(t, []*admin.NodeExecution{first, last}, nodes.GetNodeExecutions())
	assert.Empty(t, nodes.GetToken())
}

func TestFetchNodeExecutionDetailsError(t *testing.T) {
	getExecutionFetcherSetup()
	adminClient.EXPECT().ListNodeExecutions(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
	_, err := adminFetcherExt.FetchNodeExecutionDetails(ctx, "execName", "dummyProject", "domainValue", "")
	assert.Equal(t, fmt.Errorf("failed"), err)
}

func TestFetchTaskExecOnNode(t *testing.T) {
	getExecutionFetcherSetup()
	first := &admin.TaskExecution{Id: &core.TaskExecutionIdentifier{RetryAttempt: 0}}
	last := &admin.TaskExecution{Id: &core.TaskExecutionIdentifier{RetryAttempt: 1}}
	adminClient.EXPECT().ListTaskExecutions(mock.Anything, mock.Anything).
		Return(&admin.TaskExecutionList{TaskExecutions: []*admin.TaskExecution{first}, Token: "next"}, nil).Once()
	adminClient.EXPECT().ListTaskExecutions(mock.Anything, mock.MatchedBy(func(r *admin.TaskExecutionListRequest) bool {
		return r.GetToken() == "next" && r.GetNodeExecutionId().GetNodeId() == "nodeId"
	})).Return(&admin.TaskExecutionList{TaskExecutions: []*admin.TaskExecution{last}}, nil).Once()

	tasks, err := adminFetcherExt.FetchTaskExecutionsOnNode(ctx, "nodeId", "execName", "dummyProject", "domainValue")
	require.NoError(t, err)
	assert.Equal(t, []*admin.TaskExecution{first, last}, tasks.GetTaskExecutions())
	assert.Empty(t, tasks.GetToken())
}

func TestFetchTaskExecOnNodeError(t *testing.T) {
	getExecutionFetcherSetup()
	adminClient.EXPECT().ListTaskExecutions(mock.Anything, mock.Anything).Return(&admin.TaskExecutionList{Token: "next"}, nil).Once()
	adminClient.EXPECT().ListTaskExecutions(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
	tasks, err := adminFetcherExt.FetchTaskExecutionsOnNode(ctx, "nodeId", "execName", "dummyProject", "domainValue")
	assert.Equal(t, fmt.Errorf("failed"), err)
	assert.Nil(t, tasks)
}

func TestFetchNodeData(t *testing.T) {
	getExecutionFetcherSetup()
	adminClient.EXPECT().GetNodeExecutionData(mock.Anything, mock.Anything).Return(&admin.NodeExecutionGetDataResponse{}, nil)
	_, err := adminFetcherExt.FetchNodeExecutionData(ctx, "nodeId", "execName", "dummyProject", "domainValue")
	assert.Nil(t, err)
}

func TestFetchNodeDataError(t *testing.T) {
	getExecutionFetcherSetup()
	adminClient.EXPECT().GetNodeExecutionData(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
	_, err := adminFetcherExt.FetchNodeExecutionData(ctx, "nodeId", "execName", "dummyProject", "domainValue")
	assert.Equal(t, fmt.Errorf("failed"), err)
}
