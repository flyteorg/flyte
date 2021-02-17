package delete

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteCommand(t *testing.T) {
	deleteCommand := RemoteDeleteCommand()
	assert.Equal(t, deleteCommand.Use, "delete")
	assert.Equal(t, deleteCommand.Short, deleteCmdShort)
	assert.Equal(t, deleteCommand.Long, deleteCmdLong)
	assert.Equal(t, len(deleteCommand.Commands()), 1)
	cmdNouns := deleteCommand.Commands()
	// Sort by Use value.
	sort.Slice(cmdNouns, func(i, j int) bool {
		return cmdNouns[i].Use < cmdNouns[j].Use
	})
	assert.Equal(t, cmdNouns[0].Use, "execution")
	assert.Equal(t, cmdNouns[0].Aliases, []string{"executions"})
	assert.Equal(t, cmdNouns[0].Short, execCmdShort)
	assert.Equal(t, cmdNouns[0].Long, execCmdLong)
}
