package register

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"testing"

	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	u "github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flyteidl/clients/go/admin/mocks"

	"github.com/stretchr/testify/assert"
)

var (
	ctx             context.Context
	mockAdminClient *mocks.AdminServiceClient
	cmdCtx          cmdCore.CommandContext
	args            []string
	GetDoFunc       func(req *http.Request) (*http.Response, error)
)

var setup = u.Setup

func TestRegisterCommand(t *testing.T) {
	registerCommand := RemoteRegisterCommand()
	assert.Equal(t, registerCommand.Use, "register")
	assert.Equal(t, registerCommand.Short, "Registers tasks/workflows/launchplans from list of generated serialized files.")
	fmt.Println(registerCommand.Commands())
	assert.Equal(t, len(registerCommand.Commands()), 1)
	cmdNouns := registerCommand.Commands()
	// Sort by Use value.
	sort.Slice(cmdNouns, func(i, j int) bool {
		return cmdNouns[i].Use < cmdNouns[j].Use
	})
	assert.Equal(t, cmdNouns[0].Use, "files")
	assert.Equal(t, cmdNouns[0].Aliases, []string{"file"})
	assert.Equal(t, cmdNouns[0].Short, "Registers file resources")
}
