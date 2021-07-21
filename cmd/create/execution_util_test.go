package create

import (
	"errors"
	"testing"

	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/cmd/testutils"
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
	ctx = testutils.Ctx
	cmdCtx = testutils.CmdCtx
	mockClient = testutils.MockClient
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
	setup()
	createExecutionUtilSetup()
	mockClient.OnRelaunchExecutionMatch(ctx, relaunchRequest).Return(executionCreateResponse, nil)
	err = relaunchExecution(ctx, "execName", config.GetConfig().Project, config.GetConfig().Domain, cmdCtx)
	assert.Nil(t, err)
}

func TestCreateExecutionForRelaunchNotFound(t *testing.T) {
	setup()
	createExecutionUtilSetup()
	mockClient.OnRelaunchExecutionMatch(ctx, relaunchRequest).Return(nil, errors.New("unknown execution"))
	err = relaunchExecution(ctx, "execName", config.GetConfig().Project, config.GetConfig().Domain, cmdCtx)
	assert.NotNil(t, err)
	assert.Equal(t, err, errors.New("unknown execution"))
}

func TestCreateExecutionForRecovery(t *testing.T) {
	setup()
	createExecutionUtilSetup()
	mockClient.OnRecoverExecutionMatch(ctx, recoverRequest).Return(executionCreateResponse, nil)
	err = recoverExecution(ctx, "execName", config.GetConfig().Project, config.GetConfig().Domain, cmdCtx)
	assert.Nil(t, err)
}

func TestCreateExecutionForRecoveryNotFound(t *testing.T) {
	setup()
	createExecutionUtilSetup()
	mockClient.OnRecoverExecutionMatch(ctx, recoverRequest).Return(nil, errors.New("unknown execution"))
	err = recoverExecution(ctx, "execName", config.GetConfig().Project, config.GetConfig().Domain, cmdCtx)
	assert.NotNil(t, err)
	assert.Equal(t, err, errors.New("unknown execution"))
}
