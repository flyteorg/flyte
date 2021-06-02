package register

import (
	"testing"

	rconfig "github.com/flyteorg/flytectl/cmd/config/subcommand/register"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterFromFiles(t *testing.T) {
	t.Run("Valid registration", func(t *testing.T) {
		setup()
		registerFilesSetup()
		rconfig.DefaultFilesConfig.Archive = true
		args = []string{"testdata/valid-parent-folder-register.tar"}
		mockAdminClient.OnCreateTaskMatch(mock.Anything, mock.Anything).Return(nil, nil)
		mockAdminClient.OnCreateWorkflowMatch(mock.Anything, mock.Anything).Return(nil, nil)
		mockAdminClient.OnCreateLaunchPlanMatch(mock.Anything, mock.Anything).Return(nil, nil)
		err := registerFromFilesFunc(ctx, args, cmdCtx)
		assert.Nil(t, err)
	})
	t.Run("Invalid registration file", func(t *testing.T) {
		setup()
		registerFilesSetup()
		rconfig.DefaultFilesConfig.Archive = true
		args = []string{"testdata/invalid.tar"}
		err := registerFromFilesFunc(ctx, args, cmdCtx)
		assert.NotNil(t, err)
	})
}
