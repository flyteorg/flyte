package delete

import (
	"context"
	"errors"
	"io"
	"testing"

	cmdCore "github.com/lyft/flytectl/cmd/core"
	"github.com/lyft/flyteidl/clients/go/admin/mocks"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/stretchr/testify/assert"
)

var (
	ctx  context.Context
	args []string
)

func setup() {
	ctx = context.Background()
	args = []string{}
}

func TestTerminateExecutionFunc(t *testing.T) {
	setup()
	args = append(args, "exec1", "exec2")
	mockClient := new(mocks.AdminServiceClient)
	mockOutStream := new(io.Writer)
	cmdCtx := cmdCore.NewCommandContext(mockClient, *mockOutStream)
	terminateExecRequests := []*admin.ExecutionTerminateRequest{
		{Id: &core.WorkflowExecutionIdentifier{Name: "exec1"}},
		{Id: &core.WorkflowExecutionIdentifier{Name: "exec2"}},
	}
	terminateExecResponse := &admin.ExecutionTerminateResponse{}
	mockClient.OnTerminateExecutionMatch(ctx, terminateExecRequests[0]).Return(terminateExecResponse, nil)
	mockClient.OnTerminateExecutionMatch(ctx, terminateExecRequests[1]).Return(terminateExecResponse, nil)
	err := terminateExecutionFunc(ctx, args, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertCalled(t, "TerminateExecution", ctx, terminateExecRequests[0])
	mockClient.AssertCalled(t, "TerminateExecution", ctx, terminateExecRequests[1])
}

func TestTerminateExecutionFuncWithError(t *testing.T) {
	setup()
	args = append(args, "exec1", "exec2")
	mockClient := new(mocks.AdminServiceClient)
	mockOutStream := new(io.Writer)
	cmdCtx := cmdCore.NewCommandContext(mockClient, *mockOutStream)
	terminateExecRequests := []*admin.ExecutionTerminateRequest{
		{Id: &core.WorkflowExecutionIdentifier{Name: "exec1"}},
		{Id: &core.WorkflowExecutionIdentifier{Name: "exec2"}},
	}
	terminateExecResponse := &admin.ExecutionTerminateResponse{}
	mockClient.OnTerminateExecutionMatch(ctx, terminateExecRequests[0]).Return(nil, errors.New("failed to terminate"))
	mockClient.OnTerminateExecutionMatch(ctx, terminateExecRequests[1]).Return(terminateExecResponse, nil)
	err := terminateExecutionFunc(ctx, args, cmdCtx)
	assert.Equal(t, errors.New("failed to terminate"), err)
	mockClient.AssertCalled(t, "TerminateExecution", ctx, terminateExecRequests[0])
	mockClient.AssertNotCalled(t, "TerminateExecution", ctx, terminateExecRequests[1])
}

func TestTerminateExecutionFuncWithPartialSuccess(t *testing.T) {
	setup()
	args = append(args, "exec1", "exec2")
	mockClient := new(mocks.AdminServiceClient)
	mockOutStream := new(io.Writer)
	cmdCtx := cmdCore.NewCommandContext(mockClient, *mockOutStream)
	terminateExecRequests := []*admin.ExecutionTerminateRequest{
		{Id: &core.WorkflowExecutionIdentifier{Name: "exec1"}},
		{Id: &core.WorkflowExecutionIdentifier{Name: "exec2"}},
	}
	terminateExecResponse := &admin.ExecutionTerminateResponse{}
	mockClient.OnTerminateExecutionMatch(ctx, terminateExecRequests[0]).Return(terminateExecResponse, nil)
	mockClient.OnTerminateExecutionMatch(ctx, terminateExecRequests[1]).Return(nil, errors.New("failed to terminate"))
	err := terminateExecutionFunc(ctx, args, cmdCtx)
	assert.Equal(t, errors.New("failed to terminate"), err)
	mockClient.AssertCalled(t, "TerminateExecution", ctx, terminateExecRequests[0])
	mockClient.AssertCalled(t, "TerminateExecution", ctx, terminateExecRequests[1])
}
