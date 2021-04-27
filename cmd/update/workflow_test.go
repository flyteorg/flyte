package update

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func UpdateWorkflowSetup() {
	ctx = testutils.Ctx
	cmdCtx = testutils.CmdCtx
	mockClient = testutils.MockClient
}

func TestWorkflowUpdate(t *testing.T) {
	testutils.Setup()
	UpdateWorkflowSetup()
	namedEntityConfig = &NamedEntityConfig{}
	args = []string{"workflow1"}
	mockClient.OnUpdateNamedEntityMatch(mock.Anything, mock.Anything).Return(&admin.NamedEntityUpdateResponse{}, nil)
	assert.Nil(t, updateWorkflowFunc(ctx, args, cmdCtx))
}

func TestWorkflowUpdateFail(t *testing.T) {
	testutils.Setup()
	UpdateWorkflowSetup()
	namedEntityConfig = &NamedEntityConfig{}
	args = []string{"workflow1"}
	mockClient.OnUpdateNamedEntityMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed to update"))
	assert.NotNil(t, updateWorkflowFunc(ctx, args, cmdCtx))
}

func TestWorkflowUpdateInvalidArgs(t *testing.T) {
	testutils.Setup()
	UpdateWorkflowSetup()
	namedEntityConfig = &NamedEntityConfig{}
	args = []string{}
	assert.NotNil(t, updateWorkflowFunc(ctx, args, cmdCtx))
}
