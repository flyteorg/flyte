package create

import (
	"errors"
	"testing"

	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/stretchr/testify/assert"
)

var (
	executionCreateResponse *admin.ExecutionCreateResponse
	relaunchRequest         *admin.ExecutionRelaunchRequest
	recoverRequest          *admin.ExecutionRecoverRequest
)

// This function needs to be called after testutils.Steup()
func createExecutionUtilSetup() {
	executionCreateResponse = &admin.ExecutionCreateResponse{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "flytesnacks",
			Domain:  "development",
			Name:    "f652ea3596e7f4d80a0e",
		},
	}
	relaunchRequest = &admin.ExecutionRelaunchRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Name:    "execName",
			Project: config.GetConfig().Project,
			Domain:  config.GetConfig().Domain,
		},
	}
	recoverRequest = &admin.ExecutionRecoverRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Name:    "execName",
			Project: config.GetConfig().Project,
			Domain:  config.GetConfig().Domain,
		},
	}
}

func TestCreateExecutionForRelaunch(t *testing.T) {
	s := setup()
	createExecutionUtilSetup()
	s.MockAdminClient.OnRelaunchExecutionMatch(s.Ctx, relaunchRequest).Return(executionCreateResponse, nil)
	err := relaunchExecution(s.Ctx, "execName", config.GetConfig().Project, config.GetConfig().Domain, s.CmdCtx)
	assert.Nil(t, err)
}

func TestCreateExecutionForRelaunchNotFound(t *testing.T) {
	s := setup()
	createExecutionUtilSetup()
	s.MockAdminClient.OnRelaunchExecutionMatch(s.Ctx, relaunchRequest).Return(nil, errors.New("unknown execution"))
	err := relaunchExecution(s.Ctx, "execName", config.GetConfig().Project, config.GetConfig().Domain, s.CmdCtx)
	assert.NotNil(t, err)
	assert.Equal(t, err, errors.New("unknown execution"))
}

func TestCreateExecutionForRecovery(t *testing.T) {
	s := setup()
	createExecutionUtilSetup()
	s.MockAdminClient.OnRecoverExecutionMatch(s.Ctx, recoverRequest).Return(executionCreateResponse, nil)
	err := recoverExecution(s.Ctx, "execName", config.GetConfig().Project, config.GetConfig().Domain, s.CmdCtx)
	assert.Nil(t, err)
}

func TestCreateExecutionForRecoveryNotFound(t *testing.T) {
	s := setup()
	createExecutionUtilSetup()
	s.MockAdminClient.OnRecoverExecutionMatch(s.Ctx, recoverRequest).Return(nil, errors.New("unknown execution"))
	err := recoverExecution(s.Ctx, "execName", config.GetConfig().Project, config.GetConfig().Domain, s.CmdCtx)
	assert.NotNil(t, err)
	assert.Equal(t, err, errors.New("unknown execution"))
}
