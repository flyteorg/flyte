package update

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/config/subcommand/execution"
	"github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExecutionUpdate(t *testing.T) {
	s := testutils.Setup()
	args := []string{"execution1"}
	// Activate
	execution.UConfig.Activate = true
	s.MockAdminClient.OnUpdateExecutionMatch(mock.Anything, mock.Anything).Return(&admin.ExecutionUpdateResponse{}, nil)
	assert.Nil(t, updateExecutionFunc(s.Ctx, args, s.CmdCtx))
	// Archive
	execution.UConfig.Activate = false
	execution.UConfig.Archive = true
	assert.Nil(t, updateExecutionFunc(s.Ctx, args, s.CmdCtx))
	// Reset
	execution.UConfig.Activate = false
	execution.UConfig.Archive = false

	// Dry run
	execution.UConfig.DryRun = true
	assert.Nil(t, updateExecutionFunc(s.Ctx, args, s.CmdCtx))
	s.MockAdminClient.AssertNotCalled(t, "UpdateExecution", mock.Anything)

	// Reset
	execution.UConfig.DryRun = false
}

func TestExecutionUpdateValidationFailure(t *testing.T) {
	s := testutils.Setup()
	args := []string{"execution1"}
	execution.UConfig.Activate = true
	execution.UConfig.Archive = true
	assert.NotNil(t, updateExecutionFunc(s.Ctx, args, s.CmdCtx))
	// Reset
	execution.UConfig.Activate = false
	execution.UConfig.Archive = false
}

func TestExecutionUpdateFail(t *testing.T) {
	s := testutils.Setup()
	args := []string{"execution1"}
	s.MockAdminClient.OnUpdateExecutionMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed to update"))
	assert.NotNil(t, updateExecutionFunc(s.Ctx, args, s.CmdCtx))
}

func TestExecutionUpdateInvalidArgs(t *testing.T) {
	s := testutils.Setup()
	args := []string{}
	assert.NotNil(t, updateExecutionFunc(s.Ctx, args, s.CmdCtx))
}
