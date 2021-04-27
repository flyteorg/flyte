package update

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func UpdateLPSetup() {
	ctx = testutils.Ctx
	cmdCtx = testutils.CmdCtx
	mockClient = testutils.MockClient
}

func TestLPUpdate(t *testing.T) {
	testutils.Setup()
	UpdateLPSetup()
	namedEntityConfig = &NamedEntityConfig{}
	args = []string{"task1"}
	mockClient.OnUpdateNamedEntityMatch(mock.Anything, mock.Anything).Return(&admin.NamedEntityUpdateResponse{}, nil)
	assert.Nil(t, updateLPFunc(ctx, args, cmdCtx))
}

func TestLPUpdateFail(t *testing.T) {
	testutils.Setup()
	UpdateLPSetup()
	namedEntityConfig = &NamedEntityConfig{}
	args = []string{"task1"}
	mockClient.OnUpdateNamedEntityMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed to update"))
	assert.NotNil(t, updateTaskFunc(ctx, args, cmdCtx))
}

func TestLPUpdateInvalidArgs(t *testing.T) {
	testutils.Setup()
	UpdateLPSetup()
	namedEntityConfig = &NamedEntityConfig{}
	args = []string{}
	assert.NotNil(t, updateTaskFunc(ctx, args, cmdCtx))
}
