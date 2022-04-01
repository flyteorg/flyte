package update

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/config/subcommand/launchplan"
	"github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLPUpdate(t *testing.T) {
	s := testutils.Setup()
	launchplan.UConfig = &launchplan.UpdateConfig{Version: "v1", Archive: true}
	args := []string{"lp1"}
	s.MockAdminClient.OnUpdateLaunchPlanMatch(mock.Anything, mock.Anything).Return(&admin.LaunchPlanUpdateResponse{}, nil)
	assert.Nil(t, updateLPFunc(s.Ctx, args, s.CmdCtx))
}

func TestLPUpdateFail(t *testing.T) {
	s := testutils.Setup()
	launchplan.UConfig = &launchplan.UpdateConfig{Version: "v1", Archive: true}
	args := []string{"task1"}
	s.MockAdminClient.OnUpdateLaunchPlanMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed to update"))
	assert.NotNil(t, updateLPFunc(s.Ctx, args, s.CmdCtx))
}

func TestLPUpdateInvalidArgs(t *testing.T) {
	s := testutils.Setup()
	launchplan.UConfig = &launchplan.UpdateConfig{Version: "v1", Archive: true, Activate: true}
	args := []string{}
	assert.NotNil(t, updateLPFunc(s.Ctx, args, s.CmdCtx))
}
