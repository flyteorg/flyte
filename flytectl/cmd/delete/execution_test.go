package delete

import (
	"errors"
	"testing"

	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/stretchr/testify/assert"
)

var (
	args                  []string
	terminateExecRequests []*admin.ExecutionTerminateRequest
)

func terminateExecutionSetup() {
	args = append(args, "exec1", "exec2")
	terminateExecRequests = []*admin.ExecutionTerminateRequest{
		{Id: &core.WorkflowExecutionIdentifier{
			Name:    "exec1",
			Project: config.GetConfig().Project,
			Domain:  config.GetConfig().Domain,
		}},
		{Id: &core.WorkflowExecutionIdentifier{
			Name:    "exec2",
			Project: config.GetConfig().Project,
			Domain:  config.GetConfig().Domain,
		}},
	}
}

func TestTerminateExecutionFunc(t *testing.T) {
	s := setup()
	terminateExecutionSetup()
	terminateExecResponse := &admin.ExecutionTerminateResponse{}
	s.MockAdminClient.OnTerminateExecutionMatch(s.Ctx, terminateExecRequests[0]).Return(terminateExecResponse, nil)
	s.MockAdminClient.OnTerminateExecutionMatch(s.Ctx, terminateExecRequests[1]).Return(terminateExecResponse, nil)
	err := terminateExecutionFunc(s.Ctx, args, s.CmdCtx)
	assert.Nil(t, err)
	s.MockAdminClient.AssertCalled(t, "TerminateExecution", s.Ctx, terminateExecRequests[0])
	s.MockAdminClient.AssertCalled(t, "TerminateExecution", s.Ctx, terminateExecRequests[1])
	tearDownAndVerify(t, s.Writer, "")
}

func TestTerminateExecutionFuncWithError(t *testing.T) {
	s := setup()
	terminateExecutionSetup()
	terminateExecResponse := &admin.ExecutionTerminateResponse{}
	s.MockAdminClient.OnTerminateExecutionMatch(s.Ctx, terminateExecRequests[0]).Return(nil, errors.New("failed to terminate"))
	s.MockAdminClient.OnTerminateExecutionMatch(s.Ctx, terminateExecRequests[1]).Return(terminateExecResponse, nil)
	err := terminateExecutionFunc(s.Ctx, args, s.CmdCtx)
	assert.Equal(t, errors.New("failed to terminate"), err)
	s.MockAdminClient.AssertCalled(t, "TerminateExecution", s.Ctx, terminateExecRequests[0])
	s.MockAdminClient.AssertNotCalled(t, "TerminateExecution", s.Ctx, terminateExecRequests[1])
	tearDownAndVerify(t, s.Writer, "")
}

func TestTerminateExecutionFuncWithPartialSuccess(t *testing.T) {
	s := setup()
	terminateExecutionSetup()
	terminateExecResponse := &admin.ExecutionTerminateResponse{}
	s.MockAdminClient.OnTerminateExecutionMatch(s.Ctx, terminateExecRequests[0]).Return(terminateExecResponse, nil)
	s.MockAdminClient.OnTerminateExecutionMatch(s.Ctx, terminateExecRequests[1]).Return(nil, errors.New("failed to terminate"))
	err := terminateExecutionFunc(s.Ctx, args, s.CmdCtx)
	assert.Equal(t, errors.New("failed to terminate"), err)
	s.MockAdminClient.AssertCalled(t, "TerminateExecution", s.Ctx, terminateExecRequests[0])
	s.MockAdminClient.AssertCalled(t, "TerminateExecution", s.Ctx, terminateExecRequests[1])
	tearDownAndVerify(t, s.Writer, "")
}
