package update

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWorkflowUpdate(t *testing.T) {
	s := testutils.Setup()
	args := []string{"workflow1"}
	s.MockAdminClient.OnUpdateNamedEntityMatch(mock.Anything, mock.Anything).Return(&admin.NamedEntityUpdateResponse{}, nil)
	assert.Nil(t, getUpdateWorkflowFunc(&NamedEntityConfig{})(s.Ctx, args, s.CmdCtx))
}

func TestWorkflowUpdateFail(t *testing.T) {
	s := testutils.Setup()
	args := []string{"workflow1"}
	s.MockAdminClient.OnUpdateNamedEntityMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed to update"))
	assert.NotNil(t, getUpdateWorkflowFunc(&NamedEntityConfig{})(s.Ctx, args, s.CmdCtx))
}

func TestWorkflowUpdateInvalidArgs(t *testing.T) {
	s := testutils.Setup()
	assert.NotNil(t, getUpdateWorkflowFunc(&NamedEntityConfig{})(s.Ctx, []string{}, s.CmdCtx))
}
