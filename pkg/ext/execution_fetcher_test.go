package ext

import (
	"context"
	"fmt"
	"testing"

	"github.com/flyteorg/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	adminClient.OnGetExecutionMatch(mock.Anything, mock.Anything).Return(executionResponse, nil)
	_, err := adminFetcherExt.FetchExecution(ctx, "execName", "dummyProject", "domainValue")
	assert.Nil(t, err)
}

func TestFetchExecutionError(t *testing.T) {
	getExecutionFetcherSetup()
	adminClient.OnGetExecutionMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
	_, err := adminFetcherExt.FetchExecution(ctx, "execName", "dummyProject", "domainValue")
	assert.Equal(t, fmt.Errorf("failed"), err)
}

func TestFetchNodeExecutionDetails(t *testing.T) {
	getExecutionFetcherSetup()
	adminClient.OnListNodeExecutionsMatch(mock.Anything, mock.Anything).Return(&admin.NodeExecutionList{}, nil)
	_, err := adminFetcherExt.FetchNodeExecutionDetails(ctx, "execName", "dummyProject", "domainValue")
	assert.Nil(t, err)
}

func TestFetchNodeExecutionDetailsError(t *testing.T) {
	getExecutionFetcherSetup()
	adminClient.OnListNodeExecutionsMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
	_, err := adminFetcherExt.FetchNodeExecutionDetails(ctx, "execName", "dummyProject", "domainValue")
	assert.Equal(t, fmt.Errorf("failed"), err)
}

func TestFetchTaskExecOnNode(t *testing.T) {
	getExecutionFetcherSetup()
	adminClient.OnListTaskExecutionsMatch(mock.Anything, mock.Anything).Return(&admin.TaskExecutionList{}, nil)
	_, err := adminFetcherExt.FetchTaskExecutionsOnNode(ctx, "nodeId", "execName", "dummyProject", "domainValue")
	assert.Nil(t, err)
}

func TestFetchTaskExecOnNodeError(t *testing.T) {
	getExecutionFetcherSetup()
	adminClient.OnListTaskExecutionsMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
	_, err := adminFetcherExt.FetchTaskExecutionsOnNode(ctx, "nodeId", "execName", "dummyProject", "domainValue")
	assert.Equal(t, fmt.Errorf("failed"), err)
}
