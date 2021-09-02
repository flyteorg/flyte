package update

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func UpdateLPMetaSetup() {
	ctx = testutils.Ctx
	cmdCtx = testutils.CmdCtx
	mockClient = testutils.MockClient
}

func TestLPMetaUpdate(t *testing.T) {
	testutils.Setup()
	UpdateLPMetaSetup()
	namedEntityConfig = &NamedEntityConfig{}
	args = []string{"task1"}
	mockClient.OnUpdateNamedEntityMatch(mock.Anything, mock.Anything).Return(&admin.NamedEntityUpdateResponse{}, nil)
	assert.Nil(t, updateLPMetaFunc(ctx, args, cmdCtx))
}

func TestLPMetaUpdateFail(t *testing.T) {
	testutils.Setup()
	UpdateLPMetaSetup()
	namedEntityConfig = &NamedEntityConfig{}
	args = []string{"task1"}
	mockClient.OnUpdateNamedEntityMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed to update"))
	assert.NotNil(t, updateTaskFunc(ctx, args, cmdCtx))
}

func TestLPMetaUpdateInvalidArgs(t *testing.T) {
	testutils.Setup()
	UpdateLPMetaSetup()
	namedEntityConfig = &NamedEntityConfig{}
	args = []string{}
	assert.NotNil(t, updateTaskFunc(ctx, args, cmdCtx))
}
