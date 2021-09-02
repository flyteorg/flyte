package update

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func UpdateTaskSetup() {
	ctx = testutils.Ctx
	cmdCtx = testutils.CmdCtx
	mockClient = testutils.MockClient
}

func TestTaskUpdate(t *testing.T) {
	testutils.Setup()
	UpdateTaskSetup()
	namedEntityConfig = &NamedEntityConfig{}
	args = []string{"task1"}
	mockClient.OnUpdateNamedEntityMatch(mock.Anything, mock.Anything).Return(&admin.NamedEntityUpdateResponse{}, nil)
	assert.Nil(t, updateTaskFunc(ctx, args, cmdCtx))
}

func TestTaskUpdateFail(t *testing.T) {
	testutils.Setup()
	UpdateWorkflowSetup()
	namedEntityConfig = &NamedEntityConfig{}
	args = []string{"workflow1"}
	mockClient.OnUpdateNamedEntityMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed to update"))
	assert.NotNil(t, updateTaskFunc(ctx, args, cmdCtx))
}

func TestTaskUpdateInvalidArgs(t *testing.T) {
	testutils.Setup()
	UpdateWorkflowSetup()
	namedEntityConfig = &NamedEntityConfig{}
	args = []string{}
	assert.NotNil(t, updateTaskFunc(ctx, args, cmdCtx))
}
