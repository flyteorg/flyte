package delete

import (
	"context"
	"sort"
	"testing"

	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flyteidl/clients/go/admin/mocks"

	"github.com/stretchr/testify/assert"
)

var (
	err        error
	ctx        context.Context
	mockClient *mocks.AdminServiceClient
	cmdCtx     cmdCore.CommandContext
)
var setup = testutils.Setup
var tearDownAndVerify = testutils.TearDownAndVerify

func TestDeleteCommand(t *testing.T) {
	deleteCommand := RemoteDeleteCommand()
	assert.Equal(t, deleteCommand.Use, "delete")
	assert.Equal(t, deleteCommand.Short, deleteCmdShort)
	assert.Equal(t, deleteCommand.Long, deleteCmdLong)
	assert.Equal(t, len(deleteCommand.Commands()), 2)
	cmdNouns := deleteCommand.Commands()
	// Sort by Use value.
	sort.Slice(cmdNouns, func(i, j int) bool {
		return cmdNouns[i].Use < cmdNouns[j].Use
	})
	assert.Equal(t, cmdNouns[0].Use, "execution")
	assert.Equal(t, cmdNouns[0].Aliases, []string{"executions"})
	assert.Equal(t, cmdNouns[0].Short, execCmdShort)
	assert.Equal(t, cmdNouns[0].Long, execCmdLong)
	assert.Equal(t, cmdNouns[1].Use, "task-resource-attribute")
	assert.Equal(t, cmdNouns[1].Aliases, []string{"task-resource-attributes"})
	assert.Equal(t, cmdNouns[1].Short, taskResourceAttributesShort)
	assert.Equal(t, cmdNouns[1].Long, taskResourceAttributesLong)
}
