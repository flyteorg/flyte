package update

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLPMetaUpdate(t *testing.T) {
	s := testutils.Setup()
	namedEntityConfig := &NamedEntityConfig{}
	args := []string{"task1"}
	s.MockAdminClient.OnUpdateNamedEntityMatch(mock.Anything, mock.Anything).Return(&admin.NamedEntityUpdateResponse{}, nil)
	assert.Nil(t, getUpdateLPMetaFunc(namedEntityConfig)(s.Ctx, args, s.CmdCtx))
}

func TestLPMetaUpdateFail(t *testing.T) {
	s := testutils.Setup()
	namedEntityConfig := &NamedEntityConfig{}
	args := []string{"task1"}
	s.MockAdminClient.OnUpdateNamedEntityMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed to update"))
	assert.NotNil(t, getUpdateTaskFunc(namedEntityConfig)(s.Ctx, args, s.CmdCtx))
}

func TestLPMetaUpdateInvalidArgs(t *testing.T) {
	s := testutils.Setup()
	namedEntityConfig := &NamedEntityConfig{}
	args := []string{}
	assert.NotNil(t, getUpdateTaskFunc(namedEntityConfig)(s.Ctx, args, s.CmdCtx))
}
