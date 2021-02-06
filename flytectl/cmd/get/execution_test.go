package get

import (
	"context"
	"errors"
	"github.com/lyft/flytectl/cmd/config"
	cmdCore "github.com/lyft/flytectl/cmd/core"
	"github.com/lyft/flyteidl/clients/go/admin/mocks"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

const projectValue = "dummyProject"
const domainValue = "dummyDomain"
const executionNameValue = "e124"
const launchPlanNameValue = "lp_name"
const launchPlanVersionValue = "lp_version"
const workflowNameValue = "wf_name"
const workflowVersionValue = "wf_version"

func TestListExecutionFunc(t *testing.T) {
	ctx := context.Background()
	config.GetConfig().Project = projectValue
	config.GetConfig().Domain = domainValue
	config.GetConfig().Output = "json"
	var args []string
	mockClient := new(mocks.AdminServiceClient)
	mockOutStream := new(io.Writer)
	cmdCtx := cmdCore.NewCommandContext(mockClient, *mockOutStream)
	execListRequest := &admin.ResourceListRequest{
		Limit: 100,
		Id: &admin.NamedEntityIdentifier{
			Project: projectValue,
			Domain:  domainValue,
		},
	}
	executionResponse := &admin.Execution{
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
	var executions []*admin.Execution
	executions = append(executions, executionResponse)
	executionList := &admin.ExecutionList{
		Executions: executions,
	}
	mockClient.OnListExecutionsMatch(ctx, execListRequest).Return(executionList, nil)
	err := getExecutionFunc(ctx, args, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertCalled(t, "ListExecutions", ctx, execListRequest)
}

func TestListExecutionFuncWithError(t *testing.T) {
	ctx := context.Background()
	config.GetConfig().Project = projectValue
	config.GetConfig().Domain = domainValue
	config.GetConfig().Output = "json"
	var args []string
	mockClient := new(mocks.AdminServiceClient)
	mockOutStream := new(io.Writer)
	cmdCtx := cmdCore.NewCommandContext(mockClient, *mockOutStream)
	execListRequest := &admin.ResourceListRequest{
		Limit: 100,
		Id: &admin.NamedEntityIdentifier{
			Project: projectValue,
			Domain:  domainValue,
		},
	}
	executionResponse := &admin.Execution{
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
	var executions []*admin.Execution
	executions = append(executions, executionResponse)
	mockClient.OnListExecutionsMatch(ctx, execListRequest).Return(nil, errors.New("Executions NotFound."))
	err := getExecutionFunc(ctx, args, cmdCtx)
	assert.NotNil(t, err)
	assert.Equal(t, err, errors.New("Executions NotFound."))
	mockClient.AssertCalled(t, "ListExecutions", ctx, execListRequest)
}

func TestGetExecutionFunc(t *testing.T) {
	ctx := context.Background()
	config.GetConfig().Project = projectValue
	config.GetConfig().Domain = domainValue
	config.GetConfig().Output = "json"
	mockClient := new(mocks.AdminServiceClient)
	mockOutStream := new(io.Writer)
	cmdCtx := cmdCore.NewCommandContext(mockClient, *mockOutStream)
	execGetRequest := &admin.WorkflowExecutionGetRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Project: projectValue,
			Domain:  domainValue,
			Name:    executionNameValue,
		},
	}
	executionResponse := &admin.Execution{
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
	var executions []*admin.Execution
	executions = append(executions, executionResponse)
	args := []string{executionNameValue}
	mockClient.OnGetExecutionMatch(ctx, execGetRequest).Return(executionResponse, nil)
	err := getExecutionFunc(ctx, args, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertCalled(t, "GetExecution", ctx, execGetRequest)
}

func TestGetExecutionFuncWithError(t *testing.T) {
	ctx := context.Background()
	config.GetConfig().Project = projectValue
	config.GetConfig().Domain = domainValue
	config.GetConfig().Output = "json"
	mockClient := new(mocks.AdminServiceClient)
	mockOutStream := new(io.Writer)
	cmdCtx := cmdCore.NewCommandContext(mockClient, *mockOutStream)
	execGetRequest := &admin.WorkflowExecutionGetRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Project: projectValue,
			Domain:  domainValue,
			Name:    executionNameValue,
		},
	}
	executionResponse := &admin.Execution{
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
	var executions []*admin.Execution
	executions = append(executions, executionResponse)
	args := []string{executionNameValue}
	mockClient.OnGetExecutionMatch(ctx, execGetRequest).Return(nil, errors.New("Execution NotFound."))
	err := getExecutionFunc(ctx, args, cmdCtx)
	assert.NotNil(t, err)
	assert.Equal(t, err, errors.New("Execution NotFound."))
	mockClient.AssertCalled(t, "GetExecution", ctx, execGetRequest)
}
